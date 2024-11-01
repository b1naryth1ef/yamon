package journal

import (
	"log/slog"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/b1naryth1ef/yamon/journal/journalctl"
)

type JournalClient struct {
	sink            common.LogSink
	tracker         JournalTracker
	ignoredServices map[string]struct{}
}

func NewJournalClient(sink common.LogSink, tracker JournalTracker, ignoredServices []string) *JournalClient {
	ignored := map[string]struct{}{}
	if ignoredServices != nil {
		for _, service := range ignoredServices {
			ignored[service] = struct{}{}
		}
	}
	return &JournalClient{
		sink:            sink,
		tracker:         tracker,
		ignoredServices: ignored,
	}
}

func (j *JournalClient) Run() error {
	lines := 0
	instance, err := journalctl.New(&journalctl.Opts{
		Output: "json",
		Follow: true,
		Lines:  &lines,
		OnInvalidJSON: func(data []byte, err error) {
			slog.Warn("journalctl json parse error", slog.String("data", string(data)), slog.Any("error", err))
		},
	})
	if err != nil {
		return err
	}

	entries := instance.Entries()
	for {
		entry := <-entries

		service := entry["SYSLOG_IDENTIFIER"]
		delete(entry, "SYSLOG_IDENTIFIER")
		if _, ok := j.ignoredServices[service]; ok {
			continue
		}

		message := entry["MESSAGE"]
		delete(entry, "MESSAGE")
		delete(entry, "_HOSTNAME")
		delete(entry, "_SYSTEMD_INVOCATION_ID")
		delete(entry, "_STREAM_ID")
		delete(entry, "__MONOTONIC_TIMESTAMP")

		realtimeTimestamp := entry.RealtimeTimestamp()
		delete(entry, "__REALTIME_TIMESTAMP")

		cursor := entry["__CURSOR"]
		delete(entry, "__CURSOR")

		logEntry := common.NewLogEntry(
			service,
			message,
			entry,
		)

		logEntry.Time = realtimeTimestamp
		logEntry.Level = getLevelName(entry["PRIORITY"])
		j.sink.WriteLog(logEntry)

		err = j.tracker.CommitCursor(cursor)
		if err != nil {
			return err
		}
	}
}

func getLevelName(level string) string {
	switch level {
	case "0", "1", "2":
		return "critical"
	case "3":
		return "error"
	case "4":
		return "warning"
	case "5", "6":
		return "info"
	case "7":
		return "debug"
	default:
		return ""
	}
}
