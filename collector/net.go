package collector

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/util"
)

type NetworkDeviceStats struct {
	Bytes   uint64
	Packets uint64
	Errors  uint64
	Drop    uint64
}

type NetworkDevice struct {
	Name string
	Rx   NetworkDeviceStats
	Tx   NetworkDeviceStats
}

func getNetworkDevices() ([]NetworkDevice, error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, err
	}

	var result []NetworkDevice

	lines := bytes.Split(data, []byte{'\n'})
	for _, line := range lines[2:] {
		parts := util.FilterRepeatingSpaces(strings.Split(string(line), " "))

		if len(parts) == 0 {
			continue
		}

		device := NetworkDevice{
			Name: parts[0][:len(parts[0])-1],
		}

		device.Rx.Bytes = util.ParseNumber(parts[1])
		device.Rx.Packets = util.ParseNumber(parts[2])
		device.Rx.Errors = util.ParseNumber(parts[3])
		device.Rx.Drop = util.ParseNumber(parts[4])
		device.Tx.Bytes = util.ParseNumber(parts[9])
		device.Tx.Packets = util.ParseNumber(parts[10])
		device.Tx.Errors = util.ParseNumber(parts[11])
		device.Rx.Drop = util.ParseNumber(parts[12])
		result = append(result, device)
	}

	return result, nil
}

var networkCollector = Simple("net", func(ctx context.Context, sink common.Sink) error {
	devices, err := getNetworkDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		if strings.HasPrefix(device.Name, "veth") || strings.HasPrefix(device.Name, "br-") {
			continue
		}

		tags := tags("iface", device.Name)
		sink.WriteMetric(common.NewCounter(
			"net.rx.bytes", device.Rx.Bytes, tags,
		))
		sink.WriteMetric(common.NewCounter(
			"net.rx.packets", device.Rx.Packets, tags,
		))
		sink.WriteMetric(common.NewCounter(
			"net.tx.bytes", device.Tx.Bytes, tags,
		))
		sink.WriteMetric(common.NewCounter(
			"net.tx.packets", device.Tx.Packets, tags,
		))
	}

	return nil
})
