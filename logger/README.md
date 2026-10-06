# logger

A thin wrapper around [zap](https://github.com/uber-go/zap) with rotating file output ([lumberjack](https://github.com/natefinch/lumberjack)) and a ready-to-use package-level logger.

- **Two outputs**: colored, human-readable console and JSON file, each independently switchable.
- **Rotation**: by size, with retention by age and count, optional gzip.
- **Zero setup**: works out of the box (console only); one `Init` call enables file output.
- **Correct caller**: the reported line always points at your code, including `With`, the package-level functions and `Zap()`.
- **Runtime level**: change the level without restarting.
- **Safe defaults**: invalid config fails fast; colors are disabled for pipes and `NO_COLOR`.

## Install

```bash
go get github.com/beyenpay/x/logger
```

## Usage

```go
package main

import (
	"github.com/beyenpay/x/logger"
	"go.uber.org/zap"
)

func main() {
	// Optional: without Init the logger writes to the console only.
	if err := logger.Init(&logger.Config{
		Level:    "debug",
		Dir:      "logs",
		Filename: "gateway.log",
		TimeZone: "Asia/Shanghai",
	}); err != nil {
		panic(err)
	}
	defer logger.Close() // flush and release the log file

	// Structured, printf-style and key-value APIs.
	logger.Info("server started", zap.String("addr", ":8080"))
	logger.Warnf("slow request: %d ms", 300)
	logger.Errorw("query failed", "table", "users", "err", "timeout")

	// Child logger carrying extra fields.
	reqLog := logger.With(zap.String("request_id", "abc123"))
	reqLog.Info("handling")
}
```

Console:

```
2026-10-06 15:12:52.686 +08:00 | INFO  | main.go:20 | server started | {"addr": ":8080"}
```

File (`logs/gateway.log`):

```json
{"level":"INFO","time":"2026-10-06T15:12:52.686+08:00","caller":"main.go:20","msg":"server started","addr":":8080"}
```

### Instance style

```go
l, err := logger.New(nil) // nil = DefaultConfig()
if err != nil {
	return err
}
defer l.Close()

l.Info("hello")
logger.SetDefault(l) // optionally make it the package-level logger
```

### Framework integration

```go
engine.Use(ginzap.Ginzap(logger.Zap(), time.RFC3339, true))
```

### Change level at runtime

```go
_ = logger.SetLevel("debug")
```

## Configuration

Every field is optional; zero values fall back to the defaults below.

| Field             | Default     | Description                                            |
| ----------------- | ----------- | ------------------------------------------------------ |
| `Level`           | `"info"`    | `debug` `info` `warn` `error` `dpanic` `panic` `fatal` |
| `StacktraceLevel` | `"dpanic"`  | Minimum level that attaches a stack trace              |
| `Dir`             | `"logs"`    | Log directory                                          |
| `Filename`        | `"app.log"` | Active log file name                                   |
| `MaxSize`         | `100`       | Max file size in MB before rotation                    |
| `MaxAge`          | `30`        | Days to keep rotated files                             |
| `MaxBackups`      | `10`        | Rotated files to keep                                  |
| `Compress`        | `true`*     | gzip rotated files                                     |
| `TimeZone`        | `"Local"`   | IANA name, e.g. `UTC`, `Asia/Shanghai`                 |
| `DisableConsole`  | `false`     | Turn off console output                                |
| `DisableFile`     | `false`     | Turn off file output                                   |
| `CallerSkip`      | `0`         | Extra frames to skip when you wrap this package        |

\* `DefaultConfig()` and `New(nil)` enable it; a hand-built `Config{}` leaves it off.

## Notes

- Call `Init` or `SetDefault` once at startup.
- Minimal container images without tzdata need `import _ "time/tzdata"` to use named time zones.
- Close only the root logger; children from `With`/`Named` share its file.

## Layout

```
logger/
├── config.go        # Config, DefaultConfig, defaults
├── encoder.go       # console/file encoders, colors, time zone
├── level.go         # ParseLevel
├── logger.go        # Logger, New, level control, child loggers
├── global.go        # package-level logger and functions
├── logger_test.go
└── README.md
```
