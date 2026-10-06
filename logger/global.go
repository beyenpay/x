package logger

import (
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// std holds the package-level logger. It is stored with one extra skipped
// frame because the functions below sit between the caller and Logger.
var std atomic.Pointer[Logger]

func init() {
	// Console-only until the application calls Init or SetDefault, so that
	// importing this package never touches the file system.
	l, err := New(&Config{DisableFile: true})
	if err != nil {
		l = newLogger(zap.NewNop(), zap.NewAtomicLevel(), 0, nil)
	}
	SetDefault(l)
}

// Init builds a Logger from cfg and installs it as the package-level logger.
// Call it once at startup.
func Init(cfg *Config) error {
	l, err := New(cfg)
	if err != nil {
		return err
	}
	SetDefault(l)
	return nil
}

// SetDefault installs l as the package-level logger. A nil l is ignored.
func SetDefault(l *Logger) {
	if l == nil {
		return
	}
	std.Store(l.withSkip(1))
}

// Default returns the package-level logger for use as a value, e.g. to pass
// it to a constructor.
func Default() *Logger { return std.Load().withSkip(-1) }

// Sync flushes the package-level logger.
func Sync() error { return std.Load().Sync() }

// Close flushes and releases the package-level logger's file.
func Close() error { return std.Load().Close() }

// SetLevel changes the package-level logger's minimum level at runtime.
func SetLevel(level string) error { return std.Load().SetLevel(level) }

// Level returns the package-level logger's minimum level.
func Level() zapcore.Level { return std.Load().Level() }

// Zap returns the package-level logger's underlying *zap.Logger.
func Zap() *zap.Logger { return std.Load().Zap() }

// Sugar returns the package-level logger's underlying *zap.SugaredLogger.
func Sugar() *zap.SugaredLogger { return std.Load().Sugar() }

// With returns a child of the package-level logger carrying fields.
func With(fields ...zap.Field) *Logger { return Default().With(fields...) }

// WithValues returns a child of the package-level logger carrying kv pairs.
func WithValues(kv ...any) *Logger { return Default().WithValues(kv...) }

// Named returns a child of the package-level logger with the given name.
func Named(name string) *Logger { return Default().Named(name) }

// ==================== Structured logging ====================

func Debug(msg string, fields ...zap.Field)  { std.Load().Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)   { std.Load().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)   { std.Load().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field)  { std.Load().Error(msg, fields...) }
func DPanic(msg string, fields ...zap.Field) { std.Load().DPanic(msg, fields...) }
func Panic(msg string, fields ...zap.Field)  { std.Load().Panic(msg, fields...) }
func Fatal(msg string, fields ...zap.Field)  { std.Load().Fatal(msg, fields...) }

// ==================== Printf-style logging ====================

func Debugf(template string, args ...any)  { std.Load().Debugf(template, args...) }
func Infof(template string, args ...any)   { std.Load().Infof(template, args...) }
func Warnf(template string, args ...any)   { std.Load().Warnf(template, args...) }
func Errorf(template string, args ...any)  { std.Load().Errorf(template, args...) }
func DPanicf(template string, args ...any) { std.Load().DPanicf(template, args...) }
func Panicf(template string, args ...any)  { std.Load().Panicf(template, args...) }
func Fatalf(template string, args ...any)  { std.Load().Fatalf(template, args...) }

// ==================== Key-value logging ====================

func Debugw(msg string, kv ...any)  { std.Load().Debugw(msg, kv...) }
func Infow(msg string, kv ...any)   { std.Load().Infow(msg, kv...) }
func Warnw(msg string, kv ...any)   { std.Load().Warnw(msg, kv...) }
func Errorw(msg string, kv ...any)  { std.Load().Errorw(msg, kv...) }
func DPanicw(msg string, kv ...any) { std.Load().DPanicw(msg, kv...) }
func Panicw(msg string, kv ...any)  { std.Load().Panicw(msg, kv...) }
func Fatalw(msg string, kv ...any)  { std.Load().Fatalw(msg, kv...) }
