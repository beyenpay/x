# x

Shared Go packages for Beyen backend services. Each directory is an independent tool with its own README.

```bash
go get github.com/beyenpay/x/<tool>
```

## [logger](logger/README.md)

Zap-based logging with a colored console, rotating JSON files, runtime level control and a zero-setup package-level logger.

## [uid](uid/README.md)

Unique ID generation: compact 32-character hex IDs (random or time-ordered) with optional entity prefixes such as `ord_`.
