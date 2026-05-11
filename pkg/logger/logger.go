package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/university/sports-event-planner-platform/pkg/config"
)

func New(cfg config.AppConfig) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With(
		"service", cfg.Name,
		"env", cfg.Environment,
	)
}
