package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
)

var fsCacheCollector = Simple("fscache", func(ctx context.Context, sink common.Sink) error {
	stats, err := os.ReadFile("/proc/fs/fscache/stats")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	for _, line := range bytes.Split(stats, []byte{'\n'}) {
		if !bytes.Contains(line, []byte{':'}) {
			continue
		}
		parts := bytes.Split(line, []byte{':'})

		rootKey := strings.TrimSpace(strings.ToLower(string(parts[0])))
		for _, part := range bytes.Split(parts[1], []byte{' '}) {
			if len(part) == 0 {
				continue
			}
			kv := bytes.SplitN(part, []byte{'='}, 2)
			v, err := strconv.Atoi(string(kv[1]))
			if err != nil {
				// TODO: logging?
				continue
			}

			sink.WriteMetric(common.NewCounter(fmt.Sprintf("fscache.%s.%s", rootKey, strings.ToLower(string(kv[0]))), v, nil))
		}
	}

	return nil
})
