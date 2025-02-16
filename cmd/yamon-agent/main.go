package main

import (
	"log"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/b1naryth1ef/yamon"
	"github.com/b1naryth1ef/yamon/agent"
	"github.com/b1naryth1ef/yamon/collector"
	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/journal"
	"github.com/b1naryth1ef/yamon/prom"
)

func main() {
	var args struct {
		ConfigPath string `arg:"env:CONFIG_PATH,-c,--config-path" default:"config.hcl"`
		LogLevel   string `arg:"env:LOG_LEVEL,-l,--log-level" default:"info"`
	}
	arg.MustParse(&args)

	yamon.SetupLogging(args.LogLevel)

	config, err := common.LoadDaemonConfig(args.ConfigPath)
	if err != nil {
		log.Panicf("Failed to load configuration: %v", err)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Panicf("err = %v", err)
		return
	}

	forwardClient, err := yamon.NewForwardClient(config.Target)
	if err != nil {
		log.Panicf("err = %v", err)
		return
	}

	forwardClientSink := yamon.NewForwardClientSink(forwardClient, yamon.ForwardClientSinkFlushConfig{
		MetricThreshold: 4000,
		LogThreshold:    2500,
		Interval:        time.Second * 5,
	})

	sink := yamon.NewSinkMetadataFilter(hostname, nil, forwardClientSink)

	if config.Journal != nil && config.Journal.Enabled {
		err = journal.Run(config.Journal, sink)
		if err != nil {
			log.Panicf("Failed to start journal: %v", err)
			return
		}
	}

	if config.HTTP != nil {
		httpServer := agent.NewAgentHTTPServer(sink)
		go httpServer.Run(config.HTTP.Bind)
	}

	for _, logFile := range config.LogFile {
		go yamon.RunTail(logFile, sink)
	}

	for _, scriptConfig := range config.Scripts {
		script, err := yamon.NewScript(scriptConfig)
		if err != nil {
			log.Panicf("Failed to setup script for path %v: %v", scriptConfig.Path, err)
		}
		go script.Run(sink)
	}

	for _, promCfg := range config.Prometheus {
		scraper, err := prom.NewScraper(promCfg)
		if err != nil {
			log.Panicf("Failed to setup prometheus scraper for url %v: %v", promCfg.URL, err)
		}
		go scraper.Run(sink)
	}

	collectors := make(map[string]common.CollectorConfig)
	for _, userCollector := range config.Collectors {
		collectors[userCollector.Name] = userCollector
	}
	for name := range collector.Registry {
		if _, ok := collectors[name]; !ok {
			collectors[name] = common.CollectorConfig{
				Name: name,
			}
		}
	}

	producer := yamon.NewProducer(
		sink,
		slices.Collect(maps.Values(collectors)),
	)
	producer.Start()

	// TODO: catch signals + support "graceful" shutdown
	for {
		time.Sleep(time.Minute * 5)
	}
}
