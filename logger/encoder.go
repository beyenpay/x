package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	consoleTimeLayout = "2006-01-02 15:04:05.000 -07:00"
	fileTimeLayout    = "2006-01-02T15:04:05.000Z07:00" // RFC 3339 with milliseconds

	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

// encoderConfig builds the encoder settings shared by the console and file
// outputs; only the level and time encoders differ.
func encoderConfig(level zapcore.LevelEncoder, t zapcore.TimeEncoder) zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      level,
		EncodeTime:       t,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " | ",
	}
}

// timeEncoder formats timestamps in loc using layout. The location is
// resolved once at construction time, never per log entry.
func timeEncoder(loc *time.Location, layout string) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(loc).Format(layout))
	}
}

// consoleLevelEncoder pads the level name to a fixed width and, when color is
// true, wraps it in ANSI color codes.
func consoleLevelEncoder(color bool) zapcore.LevelEncoder {
	return func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		s := fmt.Sprintf("%-5s", level.CapitalString())
		if color {
			s = levelColor(level) + s + colorReset
		}
		enc.AppendString(s)
	}
}

func levelColor(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return colorPurple
	case zapcore.InfoLevel:
		return colorCyan
	case zapcore.WarnLevel:
		return colorYellow
	default: // error, dpanic, panic, fatal
		return colorRed
	}
}

// supportsColor reports whether f looks like an interactive terminal that
// should receive ANSI colors. It honors https://no-color.org.
func supportsColor(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
