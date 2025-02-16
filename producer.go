package yamon

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/b1naryth1ef/yamon/collector"
	"github.com/b1naryth1ef/yamon/common"
)

type Producer struct {
	sink       common.Sink
	collectors []common.CollectorConfig
}

func NewProducer(sink common.Sink, collectors []common.CollectorConfig) *Producer {
	return &Producer{
		sink:       sink,
		collectors: collectors,
	}
}

func (p *Producer) runCollector(col collector.Collector, interval, timeout time.Duration) {
	collect := func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err := col.Collect(ctx, p.sink)
		if err != nil {
			slog.Warn("producer.collector.failed", "collector", col, "error", err)
		}
	}

	for {
		collect()

		time.Sleep(interval)
	}
}

func (p *Producer) Start() error {
	for _, col := range p.collectors {
		if col.Disabled {
			continue
		}

		interval := time.Second * 5
		if col.Interval != "" {
			v, err := time.ParseDuration(col.Interval)
			if err != nil {
				return err
			}
			interval = v
		}

		timeout := time.Second * 5
		if col.Timeout != "" {
			v, err := time.ParseDuration(col.Timeout)
			if err != nil {
				return err
			}
			timeout = v
		}

		inst := collector.Registry.Get(col.Name)
		if inst == nil {
			return fmt.Errorf("no such collector '%s'", col.Name)
		}

		go p.runCollector(inst, interval, timeout)
	}

	return nil
}
