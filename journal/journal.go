package journal

import (
	"log/slog"

	"github.com/b1naryth1ef/yamon/common"
)

func Run(config *common.DaemonJournalConfig, sink common.Sink) error {
	var tracker JournalTracker
	var err error
	if config.CursorPath != "" {
		tracker, err = NewFileBasedJournalTracker(config.CursorPath, config.CursorSync)
		if err != nil {
			return err
		}
	} else {
		tracker = &NoopJournalTracker{}
	}

	client := NewJournalClient(sink, tracker, config.IgnoredServices)

	go func() {
		err := client.Run()
		if err != nil {
			slog.Error("error running journal reader", slog.Any("error", err))
		}
	}()

	return nil
}
