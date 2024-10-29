package collector

import (
	"context"
	"os"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/util"
	"github.com/mackerelio/go-osstat/memory"
)

var memoryCollector = Simple("memory", func(ctx context.Context, sink common.Sink) error {
	memory, err := memory.Get()
	if err != nil {
		return err
	}
	sink.WriteMetric(common.NewGauge("memory.total", memory.Total, nil))
	sink.WriteMetric(common.NewGauge("memory.used", memory.Used, nil))
	sink.WriteMetric(common.NewGauge("memory.cached", memory.Cached, nil))
	sink.WriteMetric(common.NewGauge("memory.free", memory.Free, nil))
	sink.WriteMetric(common.NewGauge("memory.available", memory.Available, nil))
	return nil
})

var loadCollector = Simple("load", func(ctx context.Context, sink common.Sink) error {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return err
	}

	parts := strings.Split(string(data), " ")

	avg1 := util.ParseFloat(parts[0])
	avg5 := util.ParseFloat(parts[1])
	avg15 := util.ParseFloat(parts[2])

	sink.WriteMetric(common.NewGauge("load.1", avg1, nil))
	sink.WriteMetric(common.NewGauge("load.5", avg5, nil))
	sink.WriteMetric(common.NewGauge("load.15", avg15, nil))
	return nil
})

var uptimeCollector = Simple("uptime", func(ctx context.Context, sink common.Sink) error {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return err
	}

	uptime := util.ParseFloat(strings.Split(string(data), " ")[0])
	sink.WriteMetric(common.NewGauge("uptime", uptime, nil))
	return nil
})
