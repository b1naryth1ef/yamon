package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/b1naryth1ef/yamon/collector"
	"github.com/b1naryth1ef/yamon/common"
)

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

func main() {
	var args struct {
		Collector string `arg:"positional,required"`
		LogLevel  string `arg:"env:LOG_LEVEL,-l,--log-level" default:"info"`
	}

	arg.MustParse(&args)

	collector := collector.Registry.Get(args.Collector)
	if collector == nil {
		fmt.Printf("No collector with name '%s'\n", os.Args[1])
		return
	}

	sink := &FakeSink{}
	err := collector.Collect(context.Background(), sink)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

}
