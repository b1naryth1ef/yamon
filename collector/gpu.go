package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/b1naryth1ef/yamon/common"
)

// source: https://github.com/henrygd/beszel/blob/8e531e6b3c9e5baa8fc328394744959f479acb59/beszel/internal/agent/gpu.go#L25-L34
type rocmSmiJson struct {
	ID          string `json:"GUID"`
	Name        string `json:"Card series"`
	Temperature string `json:"Temperature (Sensor edge) (C)"`
	MemoryUsed  string `json:"VRAM Total Used Memory (B)"`
	MemoryTotal string `json:"VRAM Total Memory (B)"`
	Usage       string `json:"GPU use (%)"`
	Power       string `json:"Current Socket Graphics Package Power (W)"`
}

func collectNVIDIA(ctx context.Context, sink common.Sink) error {
	cmd := exec.CommandContext(
		ctx,
		"nvidia-smi",
		"--query-gpu=index,name,temperature.gpu,memory.used,memory.total,utilization.gpu,power.draw",
		"--format=csv,noheader,nounits",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	for {
		line, err := stdout.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		parts := strings.Split(line, ", ")
		tags := map[string]string{
			"device": parts[1],
		}

		temp, err := strconv.Atoi(parts[2])
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.nvidia.%s.temperature", parts[0]),
				temp,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse nvidia temp", slog.String("value", parts[2]))
		}

		memoryUsed, err := strconv.Atoi(parts[3])
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.nvidia.%s.memory.used", parts[0]),
				memoryUsed,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse memory used", slog.String("value", parts[3]))
		}

		memoryTotal, err := strconv.Atoi(parts[4])
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.nvidia.%s.memory.total", parts[0]),
				memoryTotal,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse memory total", slog.String("value", parts[4]))
		}

		util, err := strconv.Atoi(parts[5])
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.nvidia.%s.utilization", parts[0]),
				util,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse utilization", slog.String("value", parts[5]))
		}

		powerDraw, err := strconv.ParseFloat(parts[6], 64)
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.nvidia.%s.powerdraw", parts[0]),
				powerDraw,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse power draw", slog.String("value", parts[6]))
		}
	}

	return nil
}

func collectAMD(ctx context.Context, sink common.Sink) error {
	cmd := exec.CommandContext(
		ctx, "rocm-smi", "--showid", "--showtemp", "--showuse", "--showpower",
		"--showproductname", "--showmeminfo", "vram", "--json",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	var data map[string]rocmSmiJson
	err = json.Unmarshal(stdout.Bytes(), &data)
	if err != nil {
		return err
	}

	for cardId, cardData := range data {
		tags := map[string]string{
			"guid":   cardData.ID,
			"device": cardData.Name,
		}

		temperature, err := strconv.ParseFloat(cardData.Temperature, 64)
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.amd.%s.temperature", cardId),
				temperature,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse temperature", slog.String("value", cardData.Temperature))
		}

		memoryUsed, err := strconv.Atoi(cardData.MemoryUsed)
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.amd.%s.memory.used", cardId),
				memoryUsed,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse memory used", slog.String("value", cardData.MemoryUsed))
		}

		memoryTotal, err := strconv.Atoi(cardData.MemoryTotal)
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.amd.%s.memory.total", cardId),
				memoryTotal,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse memory total", slog.String("value", cardData.MemoryTotal))
		}

		usage, err := strconv.Atoi(cardData.Usage)
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.amd.%s.utilization", cardId),
				usage,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse usage", slog.String("value", cardData.Usage))
		}

		power, err := strconv.Atoi(cardData.Power)
		if err == nil {
			sink.WriteMetric(common.NewGauge(
				fmt.Sprintf("gpu.amd.%s.powerdraw", cardId),
				power,
				tags,
			))
		} else {
			slog.Warn("gpu: failed to parse power", slog.String("value", cardData.Power))
		}
	}

	return nil
}

var amdOk bool = false
var amdLastChecked time.Time
var nvidiaOk bool = false
var nvidiaLastChecked time.Time

var gpuCollector = Simple("gpu", func(ctx context.Context, sink common.Sink) error {
	if nvidiaLastChecked.IsZero() || time.Now().Sub(nvidiaLastChecked) > time.Minute {
		_, err := os.Stat("/proc/driver/nvidia")
		nvidiaOk = err == nil
	}

	if nvidiaOk {
		err := collectNVIDIA(ctx, sink)
		if err != nil {
			return err
		}
	}

	if amdLastChecked.IsZero() || time.Now().Sub(amdLastChecked) > time.Minute {
		_, err := os.Stat("/sys/module/amdgpu/initstate")
		amdOk = err == nil
	}

	if amdOk {
		err := collectAMD(ctx, sink)
		if err != nil {
			return err
		}
	}

	return nil
})
