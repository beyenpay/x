package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// readEntries returns the JSON log entries written to dir/app.log.
func readEntries(t *testing.T, dir string) []map[string]any {
	t.Helper()
	f, err := os.Open(filepath.Join(dir, "app.log"))
	if err != nil {
		t.Fatalf("open log file: %v", err)
	}
	defer f.Close()

	var out []map[string]any
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("invalid JSON line %q: %v", sc.Text(), err)
		}
		out = append(out, m)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan log file: %v", err)
	}
	return out
}

func newFileLogger(t *testing.T, cfg Config) (*Logger, string) {
	t.Helper()
	dir := t.TempDir()
	cfg.Dir = dir
	cfg.DisableConsole = true
	l, err := New(&cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l, dir
}

func TestParseLevel(t *testing.T) {
	for _, in := range []string{"debug", "INFO", " warn ", "warning", "error", "dpanic", "panic", "fatal", ""} {
		if _, err := ParseLevel(in); err != nil {
			t.Errorf("ParseLevel(%q) unexpected error: %v", in, err)
		}
	}
	if _, err := ParseLevel("verbose"); err == nil {
		t.Error("ParseLevel(\"verbose\") should fail")
	}
}

func TestNewValidation(t *testing.T) {
	cases := map[string]Config{
		"no output":      {DisableConsole: true, DisableFile: true},
		"invalid level":  {Level: "loud", DisableFile: true},
		"invalid stack":  {StacktraceLevel: "loud", DisableFile: true},
		"invalid zone":   {TimeZone: "Mars/Base", DisableFile: true},
		"unwritable dir": {Dir: "/dev/null/x", DisableConsole: true},
	}
	for name, cfg := range cases {
		if _, err := New(&cfg); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestNewDoesNotMutateConfig(t *testing.T) {
	cfg := &Config{DisableFile: true}
	l, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if (*cfg != Config{DisableFile: true}) {
		t.Errorf("config was mutated: %+v", *cfg)
	}
}

func TestFileOutputIsJSONWithFields(t *testing.T) {
	l, dir := newFileLogger(t, Config{TimeZone: "UTC"})
	l.Info("hello", zap.String("k", "v"))
	l.With(zap.String("rid", "r1")).Infow("child", "n", 1)

	entries := readEntries(t, dir)
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0]["msg"] != "hello" || entries[0]["k"] != "v" || entries[0]["level"] != "INFO" {
		t.Errorf("unexpected entry: %v", entries[0])
	}
	if entries[1]["rid"] != "r1" || entries[1]["n"] != float64(1) {
		t.Errorf("unexpected child entry: %v", entries[1])
	}
	if ts, _ := entries[0]["time"].(string); !strings.HasSuffix(ts, "Z") {
		t.Errorf("time %q is not in UTC", ts)
	}
}

// Every way of logging must report this test file as the caller.
func TestCallerPointsToCaller(t *testing.T) {
	l, dir := newFileLogger(t, Config{})
	SetDefault(l)

	l.Info("method")
	l.Infof("sugar %d", 1)
	l.Infow("kv")
	l.With(zap.Int("a", 1)).Info("child")
	l.WithValues("a", 1).Info("child-kv")
	l.Zap().Info("raw-zap")
	l.Sugar().Infow("raw-sugar")
	Info("global")
	Infof("global-f")
	Infow("global-w")
	With(zap.Int("a", 1)).Info("global-child")
	Default().Info("global-default")
	Zap().Info("global-zap")

	entries := readEntries(t, dir)
	if len(entries) != 13 {
		t.Fatalf("got %d entries, want 13", len(entries))
	}
	for _, e := range entries {
		if c, _ := e["caller"].(string); !strings.HasPrefix(c, "logger/logger_test.go:") {
			t.Errorf("entry %q: caller = %q, want logger/logger_test.go", e["msg"], c)
		}
	}
}

func TestCallerSkipConfig(t *testing.T) {
	l, dir := newFileLogger(t, Config{CallerSkip: 1})
	helper := func() { l.Info("via helper") }
	helper() // with CallerSkip 1 the caller is this line, not the helper body

	entries := readEntries(t, dir)
	if c, _ := entries[0]["caller"].(string); !strings.HasPrefix(c, "logger/logger_test.go:") {
		t.Errorf("caller = %q", c)
	}
}

func TestSetLevelAffectsChildren(t *testing.T) {
	l, dir := newFileLogger(t, Config{Level: "warn"})
	child := l.With(zap.Int("a", 1))

	l.Info("dropped")
	if err := l.SetLevel("debug"); err != nil {
		t.Fatal(err)
	}
	child.Debug("kept")

	if err := l.SetLevel("nope"); err == nil {
		t.Error("SetLevel(\"nope\") should fail")
	}

	entries := readEntries(t, dir)
	if len(entries) != 1 || entries[0]["msg"] != "kept" {
		t.Errorf("unexpected entries: %v", entries)
	}
}

func TestStacktraceLevel(t *testing.T) {
	l, dir := newFileLogger(t, Config{StacktraceLevel: "error"})
	l.Warn("no stack")
	l.Error("with stack")

	entries := readEntries(t, dir)
	if _, ok := entries[0]["stacktrace"]; ok {
		t.Error("warn entry should not carry a stacktrace")
	}
	if _, ok := entries[1]["stacktrace"]; !ok {
		t.Error("error entry should carry a stacktrace")
	}
}

func TestDefaultLoggerIsUsableWithoutInit(t *testing.T) {
	// Must not panic and must not create files.
	Info("console only")
	if err := Sync(); err != nil {
		t.Errorf("Sync: %v", err)
	}
}
