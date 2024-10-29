package yamon

import (
	"context"
	"log/slog"
	"time"

	"github.com/b1naryth1ef/yamon/collector"
	"github.com/b1naryth1ef/yamon/common"
)

type Producer struct {
	sink       common.Sink
	collectors []collector.Collector
}

func NewProducer(sink common.Sink, collectors ...collector.Collector) *Producer {
	return &Producer{
		sink:       sink,
		collectors: collectors,
	}
}

func (p *Producer) produce() {
	results := make(chan struct{}, len(p.collectors))

	for _, collector := range p.collectors {
		collector := collector
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			err := collector.Collect(ctx, p.sink)
			if err != nil {
				slog.Warn("producer.collector.failed", "collector", collector, "error", err)
			}

			results <- struct{}{}
		}()
	}

	for range p.collectors {
		<-results
	}
}

func (p *Producer) Run(interval time.Duration) {
	for {
		p.produce()
		time.Sleep(interval)
	}
}
