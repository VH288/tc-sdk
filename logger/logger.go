package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	ServiceName string
	LogLevel    string
}

func New(cfg Config) *slog.Logger {
	return NewWithWriter(cfg, os.Stdout)
}

func NewWithWriter(cfg Config, w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: ParseLevel(cfg.LogLevel),
	})

	logger := slog.New(handler)
	if cfg.ServiceName != "" {
		logger = logger.With("service", cfg.ServiceName)
	}

	return logger
}

func ParseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
