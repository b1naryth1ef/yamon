package collector

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/util"
)

var procStatCPUKeys = []string{
	"user", "nice", "system", "idle", "iowait", "irq", "softirq",
}

var cpuCollector = Simple("cpu", func(ctx context.Context, sink common.Sink) error {
	stat, err := os.ReadFile("/proc/stat")
	if err != nil {
		return err
	}

	for _, line := range bytes.Split(stat, []byte{'\n'}) {
		parts := util.FilterRepeatingSpaces(strings.Split(string(line), " "))
		if len(parts) < 1 {
			continue
		}

		if strings.HasPrefix(parts[0], "cpu") {
			if parts[0] == "cpu" {
				continue
			}
			id := parts[0][3:]

			for idx, key := range procStatCPUKeys {
				value := util.ParseNumber(parts[idx+1])
				sink.WriteMetric(common.NewCounter(fmt.Sprintf("cpu.%s", key), value, map[string]string{"cpu": id}))
			}
		} else if parts[0] == "ctxt" {
			sink.WriteMetric(common.NewCounter("cpu.ctxt", util.ParseNumber(parts[1]), nil))
		}
	}

	return nil
})
