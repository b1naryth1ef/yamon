package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/util"
)

type kstat struct {
	name   string
	typeId string
	data   string
}

func newKstat(data []byte) []kstat {
	result := []kstat{}
	for idx, line := range bytes.Split(data, []byte{'\n'}) {
		if idx < 2 {
			continue
		}

		parts := util.FilterRepeatingSpaces(strings.Split(string(line), " "))
		if len(parts) != 3 {
			continue
		}
		result = append(result, kstat{
			name:   parts[0],
			typeId: parts[1],
			data:   parts[2],
		})
	}
	return result
}

var zfsCollector = Simple("zfs", func(ctx context.Context, sink common.Sink) error {
	zfetchstats, err := os.ReadFile("/proc/spl/kstat/zfs/zfetchstats")
	if err != nil {
		// ZFS: is not enabled
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	for _, stat := range newKstat(zfetchstats) {
		v, err := strconv.Atoi(stat.data)
		if err != nil {
			return err
		}
		sink.WriteMetric(common.NewCounter(fmt.Sprintf("zfs.zfetch.%s", stat.name), v, nil))
	}

	arcstats, err := os.ReadFile("/proc/spl/kstat/zfs/arcstats")
	if err != nil {
		return err
	}
	for _, stat := range newKstat(arcstats) {
		v, err := strconv.Atoi(stat.data)
		if err != nil {
			return err
		}
		sink.WriteMetric(common.NewCounter(fmt.Sprintf("zfs.arcstats.%s", stat.name), v, nil))
	}

	entries, err := os.ReadDir("/proc/spl/kstat/zfs")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		poolPath := filepath.Join("/proc/spl/kstat/zfs", entry.Name())
		objsets, err := os.ReadDir(poolPath)
		if err != nil {
			return err
		}

		for _, objSet := range objsets {
			if !strings.HasPrefix(objSet.Name(), "objset-") {
				continue
			}

			data, err := os.ReadFile(filepath.Join(poolPath, objSet.Name()))
			if err != nil {
				return err
			}

			var keys = map[string]string{}

			for _, stat := range newKstat(data) {
				if stat.name == "dataset_name" {
					keys["dataset"] = stat.data
				} else if stat.typeId == "4" {
					v, err := strconv.Atoi(stat.data)
					if err != nil {
						return err
					}
					sink.WriteMetric(common.NewCounter(fmt.Sprintf("zfs.dataset.%s", stat.name), v, keys))
				} else {
					slog.Warn("unsupported kstat zfs type", slog.String("name", stat.name), slog.String("typeId", stat.typeId))
					continue
				}
			}
		}
	}

	return nil
})
