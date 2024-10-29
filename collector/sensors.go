package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
)

type sensorSubFeatures map[string]float64
type sensorFeatures map[string]sensorSubFeatures
type sensorChips map[string]sensorFeatures

func sensorStr(value string) string {
	return strings.ReplaceAll(strings.ToLower(strings.Join(strings.Split(value, " "), "_")), ":", "_")
}

var sensorsCollector = Simple("sensors", func(ctx context.Context, sink common.Sink) error {
	cmd := exec.CommandContext(ctx, "sensors", "-j", "-A")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return err
	}

	var data sensorChips
	err = json.Unmarshal(stdout.Bytes(), &data)
	if err != nil {
		return err
	}

	for chipName, chip := range data {
		tags := map[string]string{
			"chip": chipName,
		}

		for feature, subFeatures := range chip {
			tags["feature"] = sensorStr(feature)
			for subFeature, value := range subFeatures {
				parts := strings.SplitN(subFeature, "_", 2)

				sink.WriteMetric(common.NewGauge(
					fmt.Sprintf("sensors.%s.%s", parts[0], parts[1]),
					value,
					tags,
				))

			}
		}
	}

	return nil
})
