package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/util"
)

var statKeys = []string{
	"reads_completed",
	"reads_merged",
	"sectors_read",
	"time_spent_reading",
	"writes_completed",
	"writes_merged",
	"sectors_written",
	"time_spent_writing",
	"io_in_progress",
	"time_spent_doing_io",
	"weighted_time_spent_doing_io",
	"discards_completed",
	"discards_merged",
	"sectors_discarded",
	"time_spend_discarding",
	"flush_requests_completed",
	"time_spent_flushing",
}

var diskIOCollector = Simple("disk_io", func(ctx context.Context, sink common.Sink) error {
	data, err := os.ReadFile("/proc/diskstats")
	if err != nil {
		return err
	}

	for _, line := range bytes.Split(data, []byte{'\n'}) {
		parts := util.FilterRepeatingSpaces(strings.Split(string(line), " "))
		if len(parts) == 0 {
			continue
		}

		// ignore loop devices
		if strings.HasPrefix(parts[2], "loop") {
			continue
		}

		tags := map[string]string{
			"device": parts[2],
		}
		for idx, valueStr := range parts[3:] {
			value, _ := strconv.Atoi(valueStr)
			sink.WriteMetric(common.NewCounter(fmt.Sprintf("disk.%s", statKeys[idx]), value, tags))
		}
	}
	return nil
})

type DiskUsage struct {
	FileSystemPath string
	MountPath      string
	Type           string
	Inodes         uint64
	IFree          uint64
	IUsed          uint64
	Avail          uint64
	Used           uint64
}

func getDiskUsage(ctx context.Context) ([]DiskUsage, error) {
	cmd := exec.CommandContext(ctx, "/bin/df", "--output=source,target,fstype,file,itotal,iavail,iused,ipcent,size,avail,used,pcent")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	var results []DiskUsage
	for {
		line, err := stdout.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		parts := util.FilterRepeatingSpaces(strings.Split(line, " "))
		if parts[0] == "Filesystem" {
			continue
		}

		usage := DiskUsage{
			FileSystemPath: parts[0],
			MountPath:      parts[1],
			Type:           parts[2],
			Inodes:         util.ParseNumber(parts[4]),
			IFree:          util.ParseNumber(parts[5]),
			IUsed:          util.ParseNumber(parts[6]),
			Avail:          util.ParseNumber(parts[9]),
			Used:           util.ParseNumber(parts[10]),
		}
		results = append(results, usage)
	}

	return results, err
}

var diskUsageCollector = Simple("disk_usage", func(ctx context.Context, sink common.Sink) error {
	usage, err := getDiskUsage(ctx)
	if err != nil {
		return err
	}

	for _, disk := range usage {
		if disk.Type == "tmpfs" || disk.Type == "sysfs" || disk.Type == "proc" || (disk.Inodes == 0 && disk.Used == 0 && disk.Avail == 0) {
			continue
		}

		if strings.Contains(disk.MountPath, "overlay2") {
			continue
		}

		tags := tags(
			"path", disk.FileSystemPath,
			"mount", disk.MountPath,
			"type", disk.Type,
		)
		sink.WriteMetric(common.NewGauge("disk.free", disk.Avail, tags))
		sink.WriteMetric(common.NewGauge("disk.used", disk.Used, tags))
	}

	return nil
})
