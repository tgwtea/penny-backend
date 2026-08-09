package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// New creates the structured JSON logger shared by every executable.
func New(output io.Writer, level string) (*slog.Logger, error) {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("unsupported PENNY_LOG_LEVEL %q", level)
	}

	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slogLevel})), nil
}
