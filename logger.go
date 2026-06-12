package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

// defaultTimezone is the timezone used when SetTimezone has not been called.
// Set to Germany (Europe/Berlin) per project default.
const defaultTimezone = "Europe/Berlin"

// Style selects the output format at Init time.
type Style int

const (
	styleLogger Style = iota
	styleLoggerNoUptime
	styleSlog
	styleSlim
	styleCLICompact
	styleCLI
)

// Logger returns the StyleLogger constant for use with Init.
// Default format: timestamp, uptime in seconds, level, goroutine ID,
// message, and key=value pairs in gray.
//
//	2024-01-15T15:04:05Z [  12] INFO [  1] application started version="1.0.0"
func Logger() Style { return styleLogger }

// LoggerWithoutUptime returns the legacy logger style without the uptime field.
//
//	2024-01-15T15:04:05Z INFO [  1] application started version="1.0.0"
func LoggerWithoutUptime() Style { return styleLoggerNoUptime }

// Slog returns the StyleSlog constant for use with Init (structured key=value).
func Slog() Style { return styleSlog }

// Slim returns the StyleSlim constant for use with Init (args as array in gray).
func Slim() Style { return styleSlim }

// CLICompact returns the StyleCLICompact constant for use with Init.
// Compact format for CLI tools: uptime in seconds followed by the message in
// the color of the level, without showing the level name. Args in gray.
//
//	[   0] application started version="1.0.0"
func CLICompact() Style { return styleCLICompact }

// CLI returns the StyleCLI constant for use with Init.
// Same as CLICompact, but additionally shows the log level:
//
//	[   0] INFO  application started version="1.0.0"
func CLI() Style { return styleCLI }

// LevelTrace is a custom slog level below Debug.
var LevelTrace = slog.Level(-8)

// LevelPanic is a custom slog level above Error for panic.
var LevelPanic = slog.Level(12)

// noColor returns true when the NO_COLOR or CI environment variable is set.
var noColor bool

// currentLevel is a mutable level that can be changed with SetLevel().
var currentLevel slog.LevelVar

// showThreads controls whether goroutine IDs are displayed in logs.
var showThreads = true

// contextEnricher is an optional function that extracts additional key/value
// pairs from the context (e.g. trace IDs). Set via SetContextEnricher.
var contextEnricher func(ctx context.Context) (string, string)

// currentOut is the current output writer, defaults to os.Stdout.
var currentOut io.Writer = os.Stdout

// currentStyle stores the current style for re-initialization via LogTo.
var currentStyle = styleLogger

// startTime is captured at package load and used to compute uptime in the
// default Logger format.
var startTime = time.Now()

// currentLocation is the timezone used when formatting log timestamps.
// Defaults to defaultTimezone; falls back to UTC if the zone fails to load.
var currentLocation = mustLoadDefaultLocation()

func mustLoadDefaultLocation() *time.Location {
	loc, err := time.LoadLocation(defaultTimezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func init() {
	noColor = os.Getenv("NO_COLOR") != "" || os.Getenv("CI") != ""
	currentLevel.Set(slog.LevelInfo)
}

// SetTimezone changes the timezone used to format log timestamps.
// Pass an IANA name like "Europe/Berlin", "UTC", or "America/New_York".
func SetTimezone(tz string) error {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("load timezone %q: %w", tz, err)
	}
	currentLocation = loc
	return nil
}

// formatTime formats a record timestamp in the configured timezone.
func formatTime(t time.Time) string {
	return t.In(currentLocation).Format(time.RFC3339)
}

// Init configures the global slog logger.
// Call once at program startup, e.g. logger.Init("info", logger.Logger()).
func Init(level string, style Style) {
	noColor = os.Getenv("NO_COLOR") != "" || os.Getenv("CI") != ""
	currentLevel.Set(parseLevel(level))
	currentStyle = style
	initHandler(style, currentOut)
	initKlog()
}

func initHandler(style Style, out io.Writer) {
	var handler slog.Handler
	switch style {
	case styleLogger:
		handler = NewColoredLogHandler(&currentLevel, out)
	case styleLoggerNoUptime:
		handler = NewColoredLogHandlerNoUptime(&currentLevel, out)
	case styleSlog:
		handler = NewColoredSlogHandler(&currentLevel, out)
	case styleSlim:
		handler = NewColoredSlimHandler(&currentLevel, out)
	case styleCLICompact:
		handler = NewColoredCLIHandler(&currentLevel, out, false)
	case styleCLI:
		handler = NewColoredCLIHandler(&currentLevel, out, true)
	}
	slog.SetDefault(slog.New(handler))
}

// SetLevel changes the global log level.
// Can be called after Init() to adjust verbosity at runtime.
// level must be one of: "trace", "debug", "info" (default), "warn", "error" (case-insensitive).
func SetLevel(level string) {
	currentLevel.Set(parseLevel(level))
}

// ShowThreads enables or disables displaying goroutine IDs in logs.
// Enabled by default.
func ShowThreads(show bool) {
	showThreads = show
}

// SetContextEnricher sets a function that extracts additional key/value pairs
// from the context (e.g. OpenTelemetry trace IDs). The function should return
// an empty key to indicate no value is available.
func SetContextEnricher(fn func(ctx context.Context) (string, string)) {
	contextEnricher = fn
}

// enrichFromContext calls the context enricher if set and returns key/value.
func enrichFromContext(ctx context.Context) (string, string) {
	if contextEnricher != nil {
		return contextEnricher(ctx)
	}
	return "", ""
}

// IsLevelEnabled reports whether the given level string will be logged.
func IsLevelEnabled(level string) bool {
	return parseLevel(level) >= currentLevel.Level()
}

// LogTo redirects log output to the given writer. Pass nil to reset to os.Stdout.
func LogTo(out io.Writer) {
	if out == nil {
		currentOut = os.Stdout
	} else {
		currentOut = out
	}
	initHandler(currentStyle, currentOut)
}

// extractArgs splits the variadic args into an optional leading context,
// a required string message, and the remaining key/value pairs.
//
// Supported call shapes:
//
//	logger.Info("msg")
//	logger.Info("msg", "key", val)
//	logger.Info(ctx, "msg")
//	logger.Info(ctx, "msg", "key", val)
func extractArgs(args ...any) (context.Context, string, []any) {
	if len(args) == 0 {
		return context.Background(), "", nil
	}
	ctx := context.Background()
	idx := 0
	if c, ok := args[0].(context.Context); ok {
		ctx = c
		idx = 1
	}
	if idx >= len(args) {
		return ctx, "", nil
	}
	msg, _ := args[idx].(string)
	kv := args[idx+1:]
	return ctx, msg, kv
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "panic":
		return LevelPanic
	default:
		return slog.LevelInfo
	}
}

// levelName maps an slog.Level to its display name.
func levelName(level slog.Level) string {
	switch {
	case level >= LevelPanic:
		return "PANIC"
	case level == LevelTrace:
		return "TRACE"
	case level >= slog.LevelError:
		return "ERROR"
	case level >= slog.LevelWarn:
		return "WARN"
	case level >= slog.LevelInfo:
		return "INFO"
	case level >= slog.LevelDebug:
		return "DEBUG"
	default:
		return "TRACE"
	}
}

// ---- Color helpers ----

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorPurple = "\033[35m"
	colorGray   = "\033[37m"
)

func levelColor(level slog.Level) string {
	if noColor {
		return ""
	}
	switch {
	case level >= slog.LevelError:
		return colorRed
	case level >= slog.LevelWarn:
		return colorYellow
	case level >= slog.LevelInfo:
		return colorReset
	case level >= slog.LevelDebug:
		return colorPurple
	default:
		return colorGray // trace
	}
}

func resetColor() string {
	if noColor {
		return ""
	}
	return colorReset
}

// ---- Public log functions ----

// Trace logs at trace level with optional context (via type assertion).
func Trace(args ...any) {
	ctx, msg, kv := extractArgs(args...)
	slog.Default().Log(ctx, LevelTrace, msg, kv...)
}

// TraceCtx logs at trace level with explicit context.
func TraceCtx(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, LevelTrace, msg, args...)
}

// Debug logs at debug level with optional context (via type assertion).
func Debug(args ...any) {
	ctx, msg, kv := extractArgs(args...)
	slog.Default().Log(ctx, slog.LevelDebug, msg, kv...)
}

// DebugCtx logs at debug level with explicit context.
func DebugCtx(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, slog.LevelDebug, msg, args...)
}

// Info logs at info level with optional context (via type assertion).
func Info(args ...any) {
	ctx, msg, kv := extractArgs(args...)
	slog.Default().Log(ctx, slog.LevelInfo, msg, kv...)
}

// InfoCtx logs at info level with explicit context.
func InfoCtx(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, slog.LevelInfo, msg, args...)
}

// Warn logs at warn level with optional context (via type assertion).
func Warn(args ...any) {
	ctx, msg, kv := extractArgs(args...)
	slog.Default().Log(ctx, slog.LevelWarn, msg, kv...)
}

// WarnCtx logs at warn level with explicit context.
func WarnCtx(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, slog.LevelWarn, msg, args...)
}

// Error logs at error level with optional context (via type assertion).
func Error(args ...any) {
	ctx, msg, kv := extractArgs(args...)
	slog.Default().Log(ctx, slog.LevelError, msg, kv...)
}

// ErrorCtx logs at error level with explicit context.
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, slog.LevelError, msg, args...)
}

// Panic logs at error level then panics with the message string (optional context via type assertion).
func Panic(args ...any) {
	ctx, msg, kv := extractArgs(args...)
	slog.Default().Log(ctx, slog.LevelError, msg, kv...)
	panic(msg)
}

// PanicCtx logs at error level with explicit context, then panics.
func PanicCtx(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, slog.LevelError, msg, args...)
	panic(msg)
}

// Log logs a message at the given level string.
func Log(level string, args ...any) {
	ctx, msg, kv := extractArgs(args...)
	slog.Default().Log(ctx, parseLevel(level), msg, kv...)
}

// Logr returns a logr.Logger backed by the current slog handler.
// This allows libraries that use logr to log through go-logger with the configured style.
// Must be called after Init().
func Logr() logr.Logger {
	return logr.FromSlogHandler(slog.Default().Handler())
}

// initKlog bridges klog output through the current slog.Default() handler.
func initKlog() {
	klog.SetLogger(logr.FromSlogHandler(slog.Default().Handler()))
}
