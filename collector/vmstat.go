package collector

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/b1naryth1ef/yamon/common"
)

var vmStatCollector = Simple("vmstat", func(ctx context.Context, sink common.Sink) error {
	vmStatRaw, err := os.ReadFile("/proc/vmstat")
	if err != nil {
		return err
	}

	for idx, line := range bytes.Split(vmStatRaw, []byte{'\n'}) {
		if idx < 2 {
			continue
		}

		parts := bytes.Split(line, []byte{' '})
		if len(parts) != 2 {
			continue
		}

		value, err := strconv.Atoi(string(parts[1]))
		if err != nil {
			slog.Warn("invalid or corruppted vmstat line", slog.String("line", string(line)))
			continue
		}

		sink.WriteMetric(common.NewCounter(fmt.Sprintf("vmstat.%s", string(parts[0])), value, nil))
	}

	return nil
})
