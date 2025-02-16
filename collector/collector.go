package collector

import (
	"context"

	"github.com/b1naryth1ef/yamon/collector/cgroup"
	"github.com/b1naryth1ef/yamon/common"
)

type registry map[string]Collector

var Registry = registry{}

func (r registry) Add(name string, collector Collector) {
	r[name] = collector
}

func (r registry) Get(name string) Collector {
	return r[name]
}

type Collector interface {
	Collect(context.Context, common.Sink) error
}

type simpleCollector struct {
	fn func(ctx context.Context, sink common.Sink) error
}

func (s *simpleCollector) Collect(ctx context.Context, sink common.Sink) error {
	return s.fn(ctx, sink)
}

func Simple(name string, fn func(ctx context.Context, sink common.Sink) error) Collector {
	collector := &simpleCollector{fn: fn}
	Registry.Add(name, collector)
	return collector
}

func tags(fields ...string) map[string]string {
	result := map[string]string{}

	for i := 0; i < len(fields); i += 2 {
		result[fields[i]] = fields[i+1]
	}

	return result
}

func init() {
	Registry.Add("cgroup", cgroup.NewCGroupCollector())
}
