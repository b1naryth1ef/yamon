package cgroup

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/util"
)

type deviceInfo struct {
	Name string
	Type string
}

type CGroupCollector struct {
	// todo: this can get stale?
	deviceCache map[string]deviceInfo
}

func NewCGroupCollector() *CGroupCollector {
	return &CGroupCollector{
		deviceCache: map[string]deviceInfo{},
	}
}

func (c *CGroupCollector) Collect(ctx context.Context, sink common.Sink) error {
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("cgroup v1 not supported")
	}

	return c.collectGroup("/sys/fs/cgroup", sink)
}

func (c *CGroupCollector) collectGroup(path string, sink common.Sink) error {
	items, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	cgroupPath := strings.TrimPrefix(path, "/sys/fs/cgroup/")
	cgroupPathParts := strings.Split(cgroupPath, "/")
	cgroupName := cgroupPathParts[len(cgroupPathParts)-1]
	tags := map[string]string{
		"cgroup_path": cgroupPath,
		"cgroup_name": cgroupName,
	}

	for _, item := range items {
		itemPath := filepath.Join(path, item.Name())
		if item.IsDir() {
			err = c.collectGroup(itemPath, sink)
			if err != nil {
				return err
			}
		} else {
			var err error
			switch item.Name() {
			case "cpu.stat":
				err = collectFile("cgroup.cpu", itemPath, tags, sink)
			case "memory.stat":
				err = collectFile("cgroup.memory", itemPath, tags, sink)
			case "memory.current":
				err = collectFileStat("cgroup.memory.current", itemPath, tags, sink)
			case "memory.swap.current":
				err = collectFileStat("cgroup.memory.swap.current", itemPath, tags, sink)
			case "io.stat":
				err = c.collectIOStat(itemPath, tags, sink)
			}

			if err != nil {
				return err
			}

		}
	}

	return nil
}

func collectFile(base, path string, tags map[string]string, sink common.Sink) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			return nil
		}
		parts := strings.Split(string(line), " ")
		num := util.ParseNumber(parts[1])
		sink.WriteMetric(common.NewCounter(fmt.Sprintf("%s.%s", base, parts[0]), num, tags))
	}

	return nil
}

func collectFileStat(name, path string, tags map[string]string, sink common.Sink) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	sink.WriteMetric(common.NewGauge(name, util.ParseNumber(string(data)), tags))

	return nil
}

func (c *CGroupCollector) getDeviceInfo(name string) (deviceInfo, error) {
	var result deviceInfo

	data, err := os.ReadFile(fmt.Sprintf("/sys/dev/block/%s/uevent", name))
	if err != nil {
		return result, err
	}

	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			break
		}

		parts := strings.Split(string(line), "=")
		if parts[0] == "DEVNAME" {
			result.Name = parts[1]
		} else if parts[0] == "DEVTYPE" {
			result.Type = parts[1]
		}
	}

	return result, nil
}

// kernel developer is bastard man
// 253:1 rbytes=3190784 wbytes=655360 rios=123 wios=104 dbytes=0 dios=0
// 7:7 7:6 7:5 rbytes=1145856 wbytes=0 rios=135 wios=0 dbytes=0 dios=0
func (c *CGroupCollector) collectIOStat(path string, tags map[string]string, sink common.Sink) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			return nil
		}

		parts := strings.Split(string(line), " ")

		devices := make([]deviceInfo, 0, len(parts))
		for _, part := range parts {
			if strings.Contains(part, "=") || part == "" {
				break
			}

			dev, err := c.getDeviceInfo(part)
			if err != nil {
				slog.Debug(
					"failed to fetch block device info for cgroup iostat",
					slog.String("path", path),
					slog.Any("parts", parts),
				)
				return err
			}
			devices = append(devices, dev)
		}

		for _, dev := range devices {
			tags["device_name"] = dev.Name
			tags["device_type"] = dev.Type

			for i := len(devices); i < len(parts); i++ {
				if parts[i] == "" {
					break
				}

				p := strings.Split(parts[i], "=")
				sink.WriteMetric(common.NewCounter(fmt.Sprintf("cgroup.iostat.%s", p[0]), util.ParseNumber(p[1]), tags))
			}
		}
	}

	return nil
}
