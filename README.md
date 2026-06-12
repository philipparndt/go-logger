# go-logger

Shared logger configuration for my Go projects. Built on `log/slog` with colored output, multiple styles, and optional OpenTelemetry support.

## Install

```
go get github.com/philipparndt/go-logger
```

## Usage

```go
import logger "github.com/philipparndt/go-logger"

func main() {
    // Initialize with a style: Logger(), Slog(), Slim(), CLI(), or CLICompact()
    logger.Init("info", logger.Logger())

    // Simple logging
    logger.Info("application started", "version", "1.0.0")
    logger.Debug("processing", "items", 42)
    logger.Warn("deprecated endpoint used")
    logger.Error("connection failed", "host", "db.example.com")

    // With context
    ctx := context.Background()
    logger.Info(ctx, "request handled", "method", "GET")

    // Change level at runtime
    logger.SetLevel("debug")

    // Check if level is enabled (avoid expensive computations)
    if logger.IsLevelEnabled(slog.LevelDebug) {
        logger.Debug("expensive debug info", "data", computeDebugData())
    }
}
```

## Styles

- **`Logger()`** — Default format with timestamp, uptime in seconds, level,
  goroutine ID, message and key=value pairs in gray:
  ```
  2024-01-15T15:04:05Z [  12] INFO [  1] application started version="1.0.0"
  ```

- **`LoggerWithoutUptime()`** — Same as `Logger()` but without the uptime field:
  ```
  2024-01-15T15:04:05Z INFO [  1] application started version="1.0.0"
  ```

- **`Slog()`** — Structured slog-style format:
  ```
  time=2024-01-15T15:04:05Z level=INFO msg="application started" version="1.0.0" gid=1
  ```

- **`Slim()`** — Compact with args as array:
  ```
  2024-01-15T15:04:05Z [INFO] application started [version 1.0.0 gid 1]
  ```

- **`CLICompact()`** — For CLI tools: uptime in seconds, message in the color
  of the level (level name hidden), key=value pairs in gray:
  ```
  [   0] application started version="1.0.0"
  ```

- **`CLI()`** — Same as `CLICompact()`, but also shows the log level:
  ```
  [   0] INFO  application started version="1.0.0"
  ```

## Timezone

Log timestamps default to `Europe/Berlin`. Override with `SetTimezone` using
any IANA name:

```go
if err := logger.SetTimezone("UTC"); err != nil {
    // invalid zone
}
```

## Chi Middleware

```
go get github.com/philipparndt/go-logger/chi
```

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    loggerchi "github.com/philipparndt/go-logger/chi"
)

r := chi.NewRouter()
middleware.DefaultLogger = loggerchi.Logger()
r.Use(middleware.Logger)
```

## OpenTelemetry Support

```
go get github.com/philipparndt/go-logger/otel
```

```go
import _ "github.com/philipparndt/go-logger/otel"
```

Import the package to automatically add trace IDs from context to log output.
