package yamon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (s *ScriptMetric) Write(sink common.MetricSink) {
	var entry *common.Metric
	if s.Type == common.MetricTypeGauge {
		entry = common.NewGauge(s.Name, s.Value, s.Tags)
	} else if s.Type == common.MetricTypeCounter {
		entry = common.NewCounter(s.Name, s.Value, s.Tags)
	}
	if s.Time > 0 {
		entry.Time = time.Unix(s.Time, 0)
	}
	sink.WriteMetric(entry)
}

type ScriptLogEntry struct {
	Service string `json:"service"`
	Level   string `json:"level"`
	Data    string `json:"data"`

	Time int64             `json:"time,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

func (l *ScriptLogEntry) Write(sink common.LogSink) {
	entry := common.NewLogEntry(l.Service, l.Data, l.Tags)

	entry.Level = l.Level
	if l.Time > 0 {
		entry.Time = time.Unix(l.Time, 0)
	}

	sink.WriteLog(entry)
}

type ScriptEvent struct {
	Type string `json:"type"`
	Data string `json:"data"`

	Time int64             `json:"time,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

func (e *ScriptEvent) Write(sink common.EventSink) {
	entry := common.NewEvent(e.Type, e.Data, e.Tags)
	if e.Time > 0 {
		entry.Time = time.Unix(e.Time, 0)
	}
	sink.WriteEvent(entry)
}

type ScriptResult struct {
	Metrics []ScriptMetric `json:"metrics,omitempty"`
	Metric  *ScriptMetric  `json:"metric,omitempty"`

	Logs []ScriptLogEntry `json:"logs,omitempty"`
	Log  *ScriptLogEntry  `json:"log,omitempty"`

	Events []ScriptEvent `json:"events,omitempty"`
	Event  *ScriptEvent  `json:"event,omitempty"`
}

func (s *ScriptResult) Write(sink common.Sink) {
	if s.Metric != nil {
		s.Metric.Write(sink)
	}

	if s.Metrics != nil {
		for _, metric := range s.Metrics {
			metric.Write(sink)
		}
	}

	if s.Log != nil {
		s.Log.Write(sink)
	}

	if s.Logs != nil {
		for _, log := range s.Logs {
			log.Write(sink)
		}
	}

	if s.Event != nil {
		s.Event.Write(sink)
	}

	if s.Events != nil {
		for _, event := range s.Events {
			event.Write(sink)
		}
	}
}

type Script struct {
	interval  time.Duration
	path      string
	args      []string
	env       []string
	timeout   time.Duration
	streaming bool
}

func NewScript(scriptConfig common.DaemonScriptConfig) (*Script, error) {
	if scriptConfig.Interval == "" {
		scriptConfig.Interval = "1m"
	}
	interval, err := time.ParseDuration(scriptConfig.Interval)
	if err != nil {
		return nil, err
	}

	if scriptConfig.Timeout == "" {
		scriptConfig.Timeout = "15s"
	}
	timeout, err := time.ParseDuration(scriptConfig.Timeout)
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
		interval:  interval,
		path:      scriptConfig.Path,
		args:      scriptConfig.Args,
		env:       env,
		timeout:   timeout,
		streaming: scriptConfig.Streaming,
	}, nil
}

var ErrStreamingScriptExited = errors.New("streaming script exited")

func (s *Script) Execute(sink common.Sink) error {
	var cmd *exec.Cmd
	var stdout bytes.Buffer

	if s.streaming {
		cmd = exec.Command(s.path, s.args...)
		r, w := io.Pipe()
		cmd.Stdout = w

		go func() {
			lines := bufio.NewScanner(r)
			for lines.Scan() {
				var result ScriptResult
				line := lines.Bytes()

				err := json.Unmarshal(line, &result)
				if err != nil {
					slog.Warn("script: failed to parse streaming result", slog.Any("error", err), slog.String("data", string(line)))
					continue
				}
			}
		}()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()

		cmd = exec.CommandContext(ctx, s.path, s.args...)
		cmd.Stdout = &stdout
	}

	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), s.env...)

	err := cmd.Run()
	if err != nil {
		return err
	}

	if s.streaming {
		slog.Error("script: exited", slog.String("path", s.path))
		return ErrStreamingScriptExited
	}

	var result ScriptResult
	err = json.NewDecoder(&stdout).Decode(&result)
	if err != nil {
		return err
	}
	result.Write(sink)
	return nil
}

func (s *Script) Run(sink common.Sink) {
	if s.streaming {
		go func() {
			err := s.Execute(sink)
			if err != nil {
				slog.Error("script: failed to streaming-execute", slog.String("path", s.path), slog.Any("error", err))
			}
		}()
	} else {
		ticker := time.NewTicker(s.interval)
		for {
			err := s.Execute(sink)
			if err != nil {
				slog.Error("script: failed to execute", slog.String("path", s.path), slog.Any("error", err))
			}
			time.Sleep(s.interval)
			<-ticker.C
		}
	}
}
