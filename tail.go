package yamon

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/elastic/go-libaudit/v2"
	"github.com/elastic/go-libaudit/v2/aucoalesce"
	"github.com/elastic/go-libaudit/v2/auparse"
	"github.com/influxdata/tail"
)

type auditHandler struct {
	sink common.Sink
}

func (a *auditHandler) ReassemblyComplete(msgs []*auparse.AuditMessage) {
	event, err := aucoalesce.CoalesceMessages(msgs)
	if err != nil {
		slog.Error("failed to coalesce messages in audit handler", slog.Any("error", err))
		return
	}

	jsonBytes, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal audit event", slog.Any("error", err))
		return
	}
	ev := common.NewEvent(
		"audit."+event.Type.String(),
		string(jsonBytes),
		nil,
	)
	ev.Time = event.Timestamp
	a.sink.WriteEvent(ev)
}

func (a *auditHandler) EventsLost(count int) {
	slog.Warn("detected loss of events", slog.Int("count", count))
}

func RunTail(cfg common.LogFileBlock, sink common.Sink) {
	t, err := tail.TailFile(cfg.Path, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{
		Offset: 0,
		Whence: io.SeekEnd,
	}})
	if err != nil {
		log.Panicf("Failed to tail log file %v", err)
		return
	}

	level := cfg.Level
	service := cfg.Service
	if service == "" {
		service = cfg.Path
	}

	if cfg.Format == "audit" {
		reassembler, err := libaudit.NewReassembler(100, 5*time.Second, &auditHandler{sink: sink})
		if err != nil {
			log.Panicf("failed to create reassembler: %v", err)
		}
		defer reassembler.Close()

		go func() {
			t := time.NewTicker(500 * time.Millisecond)
			defer t.Stop()
			for range t.C {
				if reassembler.Maintain() != nil {
					return
				}
			}
		}()

		for line := range t.Lines {
			auditMsg, err := auparse.ParseLogLine(line.Text)
			if err != nil {
				slog.Error("auparse ParseLogLine error", slog.Any("error", err))
				continue
			}
			reassembler.PushMessage(auditMsg)
		}

	} else if cfg.Format != "" {
		log.Panicf("Invalid log format %v", cfg.Format)
	} else {
		for line := range t.Lines {
			logEntry := common.NewLogEntry(service, line.Text, nil)
			logEntry.Level = level
			sink.WriteLog(logEntry)
		}
	}

}
