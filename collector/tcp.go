package collector

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
)

var tcpCollector = Simple("tcp", func(ctx context.Context, sink common.Sink) error {
	data, err := os.ReadFile("/proc/net/netstat")
	if err != nil {
		return err
	}

	var key string
	var keys []string

	for _, line := range bytes.Split(data, []byte{'\n'}) {
		lineParts := strings.Split(string(line), ": ")

		if len(lineParts) < 2 {
			continue
		}

		if key == "" {
			key = lineParts[0]
			keys = strings.Split(lineParts[1], " ")
			continue
		}

		if lineParts[0] != key {
			return fmt.Errorf("invalid netstat parse (order issue?)")
		}

		modKey := strings.ToLower(lineParts[0][:len(lineParts[0])-3])
		for idx, value := range strings.Split(string(lineParts[1]), " ") {
			v, err := strconv.Atoi(value)
			if err != nil {
				return err
			}

			sink.WriteMetric(common.NewCounter(fmt.Sprintf("%s.%s", modKey, keys[idx]), v, nil))
		}

		key = ""
	}

	return nil
})
