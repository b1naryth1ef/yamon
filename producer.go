package yamon

import (
	"log"
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
	for _, collector := range p.collectors {
		err := collector.Collect(p.sink)
		if err != nil {
			log.Printf("Collector %v failed: %v", collector, err)
			continue
		}
	}
}

func (p *Producer) Run(interval time.Duration) {
	for {
		p.produce()
		time.Sleep(interval)
	}
}
