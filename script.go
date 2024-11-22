package yamon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/b1naryth1ef/yamon/common"
)

type ScriptMetric struct {
	Type  string  `json:"type"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`

	Time int64             `json:"time,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

type ScriptResult struct {
	Metrics []ScriptMetric `json:"metrics,omitempty"`
}

type Script struct {
	interval time.Duration
	path     string
	args     []string
	env      []string
}

func NewScript(scriptConfig common.DaemonScriptConfig) (*Script, error) {
	if scriptConfig.Interval == "" {
		scriptConfig.Interval = "1m"
	}
	interval, err := time.ParseDuration(scriptConfig.Interval)
	if err != nil {
		return nil, err
	}

	if scriptConfig.Args == nil {
		scriptConfig.Args = []string{}
	}

	env := []string{}
	if scriptConfig.Env != nil {
		for k, v := range scriptConfig.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return &Script{
		interval: interval,
		path:     scriptConfig.Path,
		args:     scriptConfig.Args,
		env:      env,
	}, nil
}

func (s *Script) Execute(sink common.Sink) error {
	cmd := exec.Command(s.path, s.args...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), s.env...)

	err := cmd.Run()
	if err != nil {
		return err
	}

	var result ScriptResult
	err = json.NewDecoder(&stdout).Decode(&result)
	if err != nil {
		return err
	}

	if result.Metrics != nil {
		for _, metric := range result.Metrics {
			var entry *common.Metric
			if metric.Type == common.MetricTypeGauge {
				entry = common.NewGauge(metric.Name, metric.Value, metric.Tags)
			} else if metric.Type == common.MetricTypeCounter {
				entry = common.NewCounter(metric.Name, metric.Value, metric.Tags)
			}
			if metric.Time > 0 {
				entry.Time = time.Unix(metric.Time, 0)
			}
			sink.WriteMetric(entry)
		}
	}

	return nil
}

func (s *Script) Run(sink common.Sink) {
	timer := time.NewTimer(s.interval)
	for {
		err := s.Execute(sink)
		if err != nil {
			slog.Error("script: failed to execute", slog.String("path", s.path), slog.Any("error", err))
		}
		<-timer.C
	}
}
