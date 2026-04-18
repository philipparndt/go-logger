// Package chi provides a chi middleware logger that routes through go-logger.
//
// Usage:
//
//	import (
//	    "github.com/go-chi/chi/v5"
//	    "github.com/go-chi/chi/v5/middleware"
//	    loggerchi "github.com/philipparndt/go-logger/chi"
//	)
//
//	r := chi.NewRouter()
//	middleware.DefaultLogger = loggerchi.Logger()
//	r.Use(middleware.Logger)
package chi

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	logger "github.com/philipparndt/go-logger"
)

// Logger returns a chi-compatible middleware that logs requests at info level.
func Logger() func(next http.Handler) http.Handler {
	return LoggerWithLevel(slog.LevelInfo)
}

// LoggerWithLevel returns a chi-compatible middleware that logs requests at
// the specified level.
func LoggerWithLevel(level slog.Level) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&logFormatter{level: level})
}

type logFormatter struct {
	level slog.Level
}

func (l *logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &logEntry{request: r, level: l.level}
}

type logEntry struct {
	request *http.Request
	level   slog.Level
}

func (e *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	pattern := routePattern(e.request)

	slog.Default().Log(e.request.Context(), e.level, "http request",
		"method", e.request.Method,
		"path", pattern,
		"status", status,
		"bytes", bytes,
		"duration", elapsed.String(),
		"from", e.request.RemoteAddr,
	)
}

func (e *logEntry) Panic(v interface{}, stack []byte) {
	logger.Error(e.request.Context(), "http panic",
		"error", fmt.Sprintf("%v", v),
		"stack", string(stack),
	)
}

// routePattern returns the chi route pattern if available, falling back to the
// raw request URI.
func routePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx != nil {
		if pattern := rctx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.RequestURI
}
