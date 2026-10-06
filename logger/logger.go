// Package logger is a thin, opinionated wrapper around go.uber.org/zap.
//
// It writes human-readable colored output to the console and rotating JSON
// files to disk, and ships a ready-to-use package-level logger:
//
//	logger.Info("server started", zap.String("addr", ":8080"))
//
// Without any setup the package-level logger writes to the console only. Call
// Init (or SetDefault) once at startup to enable file output.
package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps a zap.Logger. It is safe for concurrent use.
type Logger struct {
	raw      *zap.Logger        // no wrapper skip; what Zap() hands out
	internal *zap.Logger        // raw + frames skipped for this package's methods
	sugar    *zap.SugaredLogger // derived from internal
	level    zap.AtomicLevel    // shared by every logger derived from this one
	skip     int                // extra frames skipped on top of the method wrapper
	closeFn  func() error       // releases file handles; nil if there is no file output
}

// New builds a Logger from cfg. A nil cfg is equivalent to DefaultConfig().
// The cfg value is never modified.
func New(cfg *Config) (*Logger, error) {
	c := *DefaultConfig()
	if cfg != nil {
		c = cfg.withDefaults()
	}

	if c.DisableConsole && c.DisableFile {
		return nil, errors.New("logger: at least one output (console or file) must be enabled")
	}
	lvl, err := ParseLevel(c.Level)
	if err != nil {
		return nil, err
	}
	stackLvl, err := ParseLevel(c.StacktraceLevel)
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(c.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("logger: invalid time zone %q: %w", c.TimeZone, err)
	}

	atomicLevel := zap.NewAtomicLevelAt(lvl)
	var (
		cores   []zapcore.Core
		closeFn func() error
	)

	if !c.DisableFile {
		if err := os.MkdirAll(c.Dir, 0o755); err != nil {
			return nil, fmt.Errorf("logger: create directory %q: %w", c.Dir, err)
		}
		w := &lumberjack.Logger{
			Filename:   filepath.Join(c.Dir, c.Filename),
			MaxSize:    c.MaxSize,
			MaxAge:     c.MaxAge,
			MaxBackups: c.MaxBackups,
			Compress:   c.Compress,
			LocalTime:  true,
		}
		enc := encoderConfig(zapcore.CapitalLevelEncoder, timeEncoder(loc, fileTimeLayout))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(enc), zapcore.AddSync(w), atomicLevel))
		closeFn = w.Close
	}

	if !c.DisableConsole {
		enc := encoderConfig(consoleLevelEncoder(supportsColor(os.Stdout)), timeEncoder(loc, consoleTimeLayout))
		cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(enc), zapcore.Lock(os.Stdout), atomicLevel))
	}

	raw := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddCallerSkip(c.CallerSkip),
		zap.AddStacktrace(stackLvl),
	)
	return newLogger(raw, atomicLevel, 0, closeFn), nil
}

// newLogger assembles a Logger. skip is the number of frames, beyond the
// method wrapper itself, between the user's call site and this package.
func newLogger(raw *zap.Logger, level zap.AtomicLevel, skip int, closeFn func() error) *Logger {
	internal := raw.WithOptions(zap.AddCallerSkip(1 + skip))
	return &Logger{
		raw:      raw,
		internal: internal,
		sugar:    internal.Sugar(),
		level:    level,
		skip:     skip,
		closeFn:  closeFn,
	}
}

// withSkip returns a Logger that skips n additional frames (n may be negative).
// It shares the output, level and Close behavior of l.
func (l *Logger) withSkip(n int) *Logger {
	return newLogger(l.raw, l.level, l.skip+n, l.closeFn)
}

// ==================== Lifecycle ====================

// Sync flushes any buffered entries. Errors caused by stdout not supporting
// fsync (a terminal or pipe) are ignored. Call it before the program exits.
func (l *Logger) Sync() error {
	err := l.raw.Sync()
	if errors.Is(err, syscall.ENOTTY) || errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.EBADF) {
		return nil
	}
	return err
}

// Close flushes buffered entries and releases the log file. Loggers derived
// through With, WithValues or Named share the same file, so close only the
// root Logger, and only once you are done with all of them.
func (l *Logger) Close() error {
	err := l.Sync()
	if l.closeFn != nil {
		err = errors.Join(err, l.closeFn())
	}
	return err
}

// ==================== Level ====================

// Level returns the current minimum level.
func (l *Logger) Level() zapcore.Level { return l.level.Level() }

// SetLevel changes the minimum level at runtime. It affects this Logger and
// every Logger derived from the same root.
func (l *Logger) SetLevel(level string) error {
	lvl, err := ParseLevel(level)
	if err != nil {
		return err
	}
	l.level.SetLevel(lvl)
	return nil
}

// ==================== Underlying loggers ====================

// Zap returns the underlying *zap.Logger, with a correct caller. Use it to
// integrate with frameworks that accept a *zap.Logger (Gin, gRPC, ...).
func (l *Logger) Zap() *zap.Logger { return l.raw }

// Sugar returns the underlying *zap.SugaredLogger, with a correct caller.
func (l *Logger) Sugar() *zap.SugaredLogger { return l.raw.Sugar() }

// ==================== Structured logging ====================

func (l *Logger) Debug(msg string, fields ...zap.Field)  { l.internal.Debug(msg, fields...) }
func (l *Logger) Info(msg string, fields ...zap.Field)   { l.internal.Info(msg, fields...) }
func (l *Logger) Warn(msg string, fields ...zap.Field)   { l.internal.Warn(msg, fields...) }
func (l *Logger) Error(msg string, fields ...zap.Field)  { l.internal.Error(msg, fields...) }
func (l *Logger) DPanic(msg string, fields ...zap.Field) { l.internal.DPanic(msg, fields...) }
func (l *Logger) Panic(msg string, fields ...zap.Field)  { l.internal.Panic(msg, fields...) }
func (l *Logger) Fatal(msg string, fields ...zap.Field)  { l.internal.Fatal(msg, fields...) }

// ==================== Printf-style logging ====================

func (l *Logger) Debugf(template string, args ...any)  { l.sugar.Debugf(template, args...) }
func (l *Logger) Infof(template string, args ...any)   { l.sugar.Infof(template, args...) }
func (l *Logger) Warnf(template string, args ...any)   { l.sugar.Warnf(template, args...) }
func (l *Logger) Errorf(template string, args ...any)  { l.sugar.Errorf(template, args...) }
func (l *Logger) DPanicf(template string, args ...any) { l.sugar.DPanicf(template, args...) }
func (l *Logger) Panicf(template string, args ...any)  { l.sugar.Panicf(template, args...) }
func (l *Logger) Fatalf(template string, args ...any)  { l.sugar.Fatalf(template, args...) }

// ==================== Key-value logging ====================

func (l *Logger) Debugw(msg string, kv ...any)  { l.sugar.Debugw(msg, kv...) }
func (l *Logger) Infow(msg string, kv ...any)   { l.sugar.Infow(msg, kv...) }
func (l *Logger) Warnw(msg string, kv ...any)   { l.sugar.Warnw(msg, kv...) }
func (l *Logger) Errorw(msg string, kv ...any)  { l.sugar.Errorw(msg, kv...) }
func (l *Logger) DPanicw(msg string, kv ...any) { l.sugar.DPanicw(msg, kv...) }
func (l *Logger) Panicw(msg string, kv ...any)  { l.sugar.Panicw(msg, kv...) }
func (l *Logger) Fatalw(msg string, kv ...any)  { l.sugar.Fatalw(msg, kv...) }

// ==================== Child loggers ====================

// With returns a child Logger that adds fields to every entry.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return newLogger(l.raw.With(fields...), l.level, l.skip, l.closeFn)
}

// WithValues is like With but takes loosely typed key-value pairs.
func (l *Logger) WithValues(kv ...any) *Logger {
	raw := l.raw.Sugar().With(kv...).Desugar()
	return newLogger(raw, l.level, l.skip, l.closeFn)
}

// Named returns a child Logger with name appended to the logger name.
func (l *Logger) Named(name string) *Logger {
	return newLogger(l.raw.Named(name), l.level, l.skip, l.closeFn)
}
