package yamon

import (
	"log"
	"log/slog"
	"os"
)

func parseLevel(s string) slog.Level {
	var level slog.Level
	err := level.UnmarshalText([]byte(s))
	if err != nil {
		log.Panicf("Failed to parse logging level '%s': %v\n", s, err)
	}
	return level
}

func SetupLogging(level string) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	}))
	slog.SetDefault(logger)
}
