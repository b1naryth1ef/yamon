package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/b1naryth1ef/yamon"
	"github.com/b1naryth1ef/yamon/collector"
	"github.com/b1naryth1ef/yamon/common"
)

var CLI struct {
	Collector struct {
		Name string `arg:"" help:"the collector to execute"`
	} `cmd:"" help:"Run a yamon collector and output the results."`

	Script struct {
		Path      string   `arg:"" type:"path" help:"path of script to execute"`
		Env       []string `name:"env" help:"configure env variables"`
		Streaming bool     `name:"streaming" help:"enable streaming mode"`
	} `cmd:"" help:"Run a yamon-compatible script and output the results."`

	LogLevel string `name:"log-level" default:"info"`
}

type FakeSink struct {
}

func (f *FakeSink) Write(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(data))
}

func (f *FakeSink) WriteMetric(metric *common.Metric) {
	f.Write(metric)
}

func (f *FakeSink) WriteLog(entry *common.LogEntry) {
	f.Write(entry)
}

func (f *FakeSink) WriteEvent(event *common.Event) {
	f.Write(event)
}

func commandScript() error {
	env := map[string]string{}
	for _, i := range CLI.Script.Env {
		before, after, ok := strings.Cut(i, "=")
		if !ok {
			return fmt.Errorf("invalid env arg '%v'", i)
		}
		env[before] = after
	}

	script, err := yamon.NewScript(common.DaemonScriptConfig{
		Path:      CLI.Script.Path,
		Env:       env,
		Streaming: CLI.Script.Streaming,
	})
	if err != nil {
		return err
	}

	return script.Execute(&FakeSink{})
}

func commandCollector() error {
	collector := collector.Registry.Get(CLI.Collector.Name)
	if collector == nil {
		return fmt.Errorf("No collector with name '%s'\n", CLI.Collector.Name)
	}

	sink := &FakeSink{}
	err := collector.Collect(context.Background(), sink)
	if err != nil {
		return fmt.Errorf("Error: %v\n", err)
	}

	return nil
}

func main() {
	ctx := kong.Parse(&CLI)

	yamon.SetupLogging(CLI.LogLevel)

	var err error
	switch ctx.Command() {
	case "collector <name>":
		err = commandCollector()
	case "script <path>":
		err = commandScript()
	default:
		log.Panicf("unknown command %s", ctx.Command())
	}
	if err != nil {
		panic(err)
	}
}
