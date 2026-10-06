package logger

import "strings"

// Config describes how a Logger is built.
//
// The zero value of every field means "use the default", so a partially
// filled Config is always valid. Start from DefaultConfig when you need to
// see (or tweak) the effective defaults.
type Config struct {
	// Level is the minimum enabled level: debug, info, warn, error, dpanic,
	// panic or fatal. Default: "info".
	Level string `json:"level" yaml:"level" mapstructure:"level"`
	// StacktraceLevel is the minimum level that attaches a stack trace.
	// Default: "dpanic".
	StacktraceLevel string `json:"stacktrace_level" yaml:"stacktrace_level" mapstructure:"stacktrace_level"`

	// Dir is the log directory (relative or absolute). Default: "logs".
	Dir string `json:"dir" yaml:"dir" mapstructure:"dir"`
	// Filename is the active log file name inside Dir. Default: "app.log".
	Filename string `json:"filename" yaml:"filename" mapstructure:"filename"`
	// MaxSize is the maximum size in megabytes before rotation. Default: 100.
	MaxSize int `json:"max_size" yaml:"max_size" mapstructure:"max_size"`
	// MaxAge is the maximum number of days to keep rotated files. Default: 30.
	MaxAge int `json:"max_age" yaml:"max_age" mapstructure:"max_age"`
	// MaxBackups is the maximum number of rotated files to keep. Default: 10.
	MaxBackups int `json:"max_backups" yaml:"max_backups" mapstructure:"max_backups"`
	// Compress gzips rotated files. DefaultConfig enables it; a hand-built
	// Config{} leaves it off.
	Compress bool `json:"compress" yaml:"compress" mapstructure:"compress"`

	// TimeZone is an IANA name such as "UTC" or "Asia/Shanghai", or "Local".
	// Default: "Local".
	TimeZone string `json:"time_zone" yaml:"time_zone" mapstructure:"time_zone"`

	// DisableConsole turns off the colored console (stdout) output.
	DisableConsole bool `json:"disable_console" yaml:"disable_console" mapstructure:"disable_console"`
	// DisableFile turns off the rotating JSON file output.
	DisableFile bool `json:"disable_file" yaml:"disable_file" mapstructure:"disable_file"`

	// CallerSkip is the number of additional stack frames to skip when
	// reporting the caller. Use it when you wrap this package in your own
	// helper functions. Default: 0.
	CallerSkip int `json:"caller_skip" yaml:"caller_skip" mapstructure:"caller_skip"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Level:           "info",
		StacktraceLevel: "dpanic",
		Dir:             "logs",
		Filename:        "app.log",
		MaxSize:         100,
		MaxAge:          30,
		MaxBackups:      10,
		Compress:        true,
		TimeZone:        "Local",
	}
}

// withDefaults returns a copy of c with empty fields filled in.
// The receiver is a value on purpose: the caller's Config is never mutated.
func (c Config) withDefaults() Config {
	def := DefaultConfig()
	if strings.TrimSpace(c.Level) == "" {
		c.Level = def.Level
	}
	if strings.TrimSpace(c.StacktraceLevel) == "" {
		c.StacktraceLevel = def.StacktraceLevel
	}
	if strings.TrimSpace(c.Dir) == "" {
		c.Dir = def.Dir
	}
	if strings.TrimSpace(c.Filename) == "" {
		c.Filename = def.Filename
	}
	if c.MaxSize <= 0 {
		c.MaxSize = def.MaxSize
	}
	if c.MaxAge <= 0 {
		c.MaxAge = def.MaxAge
	}
	if c.MaxBackups <= 0 {
		c.MaxBackups = def.MaxBackups
	}
	if strings.TrimSpace(c.TimeZone) == "" {
		c.TimeZone = def.TimeZone
	}
	if c.CallerSkip < 0 {
		c.CallerSkip = 0
	}
	return c
}
