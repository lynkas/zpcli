package logx

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	initOnce sync.Once
	enabled  bool
)

func Init(verbosity int, format string) {
	initOnce.Do(func() {
		enabled = verbosity > 0
		level := parseVerbosity(verbosity)
		opts := &slog.HandlerOptions{Level: level}

		var handler slog.Handler
		writer := io.Discard
		if enabled {
			writer = os.Stdout
		}

		if strings.EqualFold(format, "json") {
			handler = slog.NewJSONHandler(writer, opts)
		} else {
			handler = slog.NewTextHandler(writer, opts)
		}

		slog.SetDefault(slog.New(handler))
	})
}

func parseVerbosity(verbosity int) slog.Level {
	switch {
	case verbosity >= 2:
		return slog.LevelDebug
	case verbosity == 1:
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func Enabled() bool {
	return enabled
}

func Logger(component string) *slog.Logger {
	return slog.Default().With("component", component)
}
