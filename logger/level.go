package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

// ParseLevel converts a case-insensitive level name into a zapcore.Level.
// An empty string yields InfoLevel; unknown names return an error instead of
// silently falling back, so configuration typos surface at startup.
func ParseLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "", "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "dpanic":
		return zapcore.DPanicLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("logger: unknown level %q", level)
	}
}
