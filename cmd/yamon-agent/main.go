package main

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/b1naryth1ef/yamon"
	"github.com/b1naryth1ef/yamon/collector"
	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/journal"
	"github.com/b1naryth1ef/yamon/prom"
	"github.com/influxdata/tail"
	flag "github.com/spf13/pflag"
)

var configPath = flag.StringP("config", "c", "config.hcl", "path to configuration file")

func main() {
	flag.Parse()

	configPathStr := *configPath
	if configPathStr == "" {
		configPathStr = os.Getenv("CONFIG_PATH")
	}

	config, err := common.LoadDaemonConfig(configPathStr)
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
		MetricThreshold: 1000,
		LogThreshold:    1000,
		Interval:        time.Second * 5,
	})

	sink := yamon.NewSinkMetadataFilter(hostname, nil, forwardClientSink)

	if config.Journal != nil {
		err = journal.Run(config.Journal, sink)
		if err != nil {
			log.Panicf("Failed to start journal: %v", err)
			return
		}
	}

	for _, logFile := range config.LogFile {
		t, err := tail.TailFile(logFile.Path, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		}})
		if err != nil {
			log.Panicf("Failed to tail log file %v", err)
			return
		}

		level := logFile.Level
		service := logFile.Service
		if service == "" {
			service = logFile.Path
		}

		go func() {
			for line := range t.Lines {
				logEntry := common.NewLogEntry(service, line.Text, nil)
				logEntry.Level = level
				sink.WriteLog(logEntry)
			}
		}()
	}

	for _, promCfg := range config.Prometheus {
		interval, err := time.ParseDuration(promCfg.Interval)
		if err != nil {
			log.Panicf("Failed: %v (%v)", err, promCfg.Interval)
		}
		scraper := prom.NewScraper(promCfg.URL, interval, sink)
		go scraper.Run()
	}

	var collectors []collector.Collector
	if config.Collectors == "*" || config.Collectors == "" {
		collectors = collector.Registry.All()
	} else {
		names := strings.Split(config.Collectors, ",")
		for _, name := range names {
			c := collector.Registry.Get(name)
			if c == nil {
				log.Printf("WARNING: Collector with name '%s' does not exist, skipping", name)
				continue
			}
			collectors = append(collectors, c)
		}
	}

	producer := yamon.NewProducer(
		sink,
		collectors...,
	)
	producer.Run(time.Second * 5)
}
