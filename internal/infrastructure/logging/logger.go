package logging

import (
	"os"
	"strings"

	"log/slog"
)

// SetupLogger creates and configures a structured logger
func SetupLogger() *slog.Logger {
	logLevel := getLogLevel()

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}

// getLogLevel determines the log level from environment variables
func getLogLevel() slog.Level {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))

	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error", "crit":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
