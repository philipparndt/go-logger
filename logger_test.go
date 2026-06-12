package logger

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// contextKey is a custom type for context keys.
type contextKey string

// resetForTest redirects slog to a buffer and re-initialises the global
// logger, returning the buffer.
func resetForTest(level string, style Style) *bytes.Buffer {
	noColor = true
	var buf bytes.Buffer
	currentLevel.Set(parseLevel(level))
	var handler slog.Handler
	switch style {
	case styleLogger:
		handler = NewColoredLogHandler(&currentLevel, &buf)
	case styleLoggerNoUptime:
		handler = NewColoredLogHandlerNoUptime(&currentLevel, &buf)
	case styleSlog:
		handler = NewColoredSlogHandler(&currentLevel, &buf)
	case styleSlim:
		handler = NewColoredSlimHandler(&currentLevel, &buf)
	case styleCLICompact:
		handler = NewColoredCLIHandler(&currentLevel, &buf, false)
	case styleCLI:
		handler = NewColoredCLIHandler(&currentLevel, &buf, true)
	}
	slog.SetDefault(slog.New(handler))
	return &buf
}

// --- Level filtering ---

//nolint:paralleltest
func TestInit_SlogStyle_DebugVisible(t *testing.T) {
	buf := resetForTest("debug", styleSlog)
	Debug("hello debug")
	assert.Contains(t, buf.String(), "hello debug")
	assert.Contains(t, buf.String(), "DEBUG")
}

//nolint:paralleltest
func TestInit_LoggerStyle_DebugFiltered(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	Debug("should not appear")
	assert.Empty(t, buf.String())
}

//nolint:paralleltest
func TestInit_LoggerStyle_InfoVisible(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	Info("hello info")
	assert.Contains(t, buf.String(), "hello info")
	assert.Contains(t, buf.String(), "INFO")
}

// --- Timestamp format ---

//nolint:paralleltest
func TestTimestamp_IsRFC3339_LoggerStyle(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	Info("ts check")
	line := buf.String()
	parts := strings.SplitN(line, " ", 2)
	_, err := time.Parse(time.RFC3339, parts[0])
	assert.NoError(t, err, "timestamp %q is not RFC3339", parts[0])
}

//nolint:paralleltest
func TestTimestamp_IsRFC3339_SlogStyle(t *testing.T) {
	buf := resetForTest("info", styleSlog)
	Info("ts check")
	line := buf.String()
	before, after, found := strings.Cut(line, "time=")
	_ = before
	assert.True(t, found)
	ts := strings.SplitN(after, " ", 2)[0]
	_, err := time.Parse(time.RFC3339, ts)
	assert.NoError(t, err, "timestamp %q is not RFC3339", ts)
}

// --- Key/value args ---

//nolint:paralleltest
func TestKeyValue_LoggerStyle(t *testing.T) {
	buf := resetForTest("debug", styleLogger)
	Debug("msg", "mykey", "myval")
	line := buf.String()
	assert.Contains(t, line, "mykey=")
	assert.Contains(t, line, "myval")
}

//nolint:paralleltest
func TestKeyValue_SlogStyle(t *testing.T) {
	buf := resetForTest("debug", styleSlog)
	Debug("msg", "mykey", "myval")
	line := buf.String()
	assert.Contains(t, line, "mykey=")
	assert.Contains(t, line, "myval")
}

// --- NO_COLOR env var ---

//nolint:paralleltest
func TestNoColor_DisablesANSI(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	noColor = true
	assert.NotContains(t, levelColor(slog.LevelError), "\033[")
}

//nolint:paralleltest
func TestNoColor_ColorsPresentWhenNotSet(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	noColor = false
	assert.Contains(t, levelColor(slog.LevelError), "\033[")
}

// --- Panic ---

//nolint:paralleltest
func TestPanic_LogsThenPanics(t *testing.T) {
	buf := resetForTest("error", styleLogger)
	assert.Panics(t, func() {
		Panic("boom")
	})
	assert.Contains(t, buf.String(), "boom")
	assert.Contains(t, buf.String(), "ERROR")
}

// --- Trace level ---

//nolint:paralleltest
func TestTrace_VisibleAtTraceLevel(t *testing.T) {
	buf := resetForTest("trace", styleLogger)
	Trace("trace msg")
	output := buf.String()
	assert.Contains(t, output, "TRACE")
	assert.Contains(t, output, "trace msg")
}

//nolint:paralleltest
func TestTrace_HiddenAtInfoLevel(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	Trace("should not appear")
	assert.Empty(t, buf.String())
}

// --- Typed context functions ---

//nolint:paralleltest
func TestInfoCtx_WithContext(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	ctx := context.WithValue(t.Context(), contextKey("k"), "v")
	InfoCtx(ctx, "message with context", "key", "value")
	output := buf.String()
	assert.Contains(t, output, "message with context")
	assert.Contains(t, output, "key")
}

//nolint:paralleltest
func TestDebugCtx_WithContext(t *testing.T) {
	buf := resetForTest("debug", styleLogger)
	ctx := context.WithValue(t.Context(), contextKey("user"), "alice")
	DebugCtx(ctx, "debug message")
	output := buf.String()
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "debug message")
}

//nolint:paralleltest
func TestTraceCtx_WithContext(t *testing.T) {
	buf := resetForTest("trace", styleLogger)
	ctx := t.Context()
	TraceCtx(ctx, "trace with explicit context")
	output := buf.String()
	assert.Contains(t, output, "TRACE")
	assert.Contains(t, output, "trace with explicit context")
}

//nolint:paralleltest
func TestWarnCtx_WithContext(t *testing.T) {
	buf := resetForTest("warn", styleLogger)
	ctx := t.Context()
	WarnCtx(ctx, "warning message", "severity", "high")
	output := buf.String()
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "warning message")
}

//nolint:paralleltest
func TestErrorCtx_WithContext(t *testing.T) {
	buf := resetForTest("error", styleLogger)
	ctx := t.Context()
	ErrorCtx(ctx, "error message", "code", 500)
	output := buf.String()
	assert.Contains(t, output, "ERROR")
	assert.Contains(t, output, "error message")
}

//nolint:paralleltest
func TestPanicCtx_LogsThenPanics(t *testing.T) {
	buf := resetForTest("error", styleLogger)
	ctx := t.Context()
	assert.Panics(t, func() {
		PanicCtx(ctx, "panic with context")
	})
	output := buf.String()
	assert.Contains(t, output, "ERROR")
	assert.Contains(t, output, "panic with context")
}

// --- extractArgs ---

func TestExtractArgs_MsgOnly(t *testing.T) {
	t.Parallel()
	ctx, msg, kv := extractArgs("hello")
	assert.NotNil(t, ctx)
	assert.Equal(t, "hello", msg)
	assert.Empty(t, kv)
}

func TestExtractArgs_CtxAndMsg(t *testing.T) {
	t.Parallel()
	expected := context.WithValue(t.Context(), contextKey("k"), "v")
	ctx, msg, kv := extractArgs(expected, "hello", "k", 1)
	assert.Equal(t, expected, ctx)
	assert.Equal(t, "hello", msg)
	assert.Equal(t, []any{"k", 1}, kv)
}

func TestExtractArgs_NoArgs(t *testing.T) {
	t.Parallel()
	ctx, msg, kv := extractArgs()
	assert.NotNil(t, ctx)
	assert.Equal(t, "", msg)
	assert.Empty(t, kv)
}

// --- parseLevel ---

func TestParseLevel_CaseInsensitive(t *testing.T) {
	t.Parallel()
	assert.Equal(t, LevelTrace, parseLevel("TRACE"))
	assert.Equal(t, LevelTrace, parseLevel("trace"))
	assert.Equal(t, slog.LevelDebug, parseLevel("DEBUG"))
	assert.Equal(t, slog.LevelInfo, parseLevel("INFO"))
	assert.Equal(t, slog.LevelWarn, parseLevel("WARN"))
	assert.Equal(t, slog.LevelError, parseLevel("ERROR"))
	assert.Equal(t, LevelPanic, parseLevel("panic"))
	assert.Equal(t, slog.LevelInfo, parseLevel("unknown"))
}

// --- IsLevelEnabled ---

//nolint:paralleltest
func TestIsLevelEnabled_TrueWhenLevelMeetsThreshold(t *testing.T) {
	resetForTest("debug", styleLogger)
	assert.True(t, IsLevelEnabled("debug"))
	assert.True(t, IsLevelEnabled("info"))
	assert.True(t, IsLevelEnabled("warn"))
	assert.True(t, IsLevelEnabled("error"))
}

//nolint:paralleltest
func TestIsLevelEnabled_FalseWhenLevelBelowThreshold(t *testing.T) {
	resetForTest("info", styleLogger)
	assert.False(t, IsLevelEnabled("trace"))
	assert.False(t, IsLevelEnabled("debug"))
	assert.True(t, IsLevelEnabled("info"))
}

//nolint:paralleltest
func TestIsLevelEnabled_ReflectsSetLevel(t *testing.T) {
	resetForTest("info", styleLogger)
	assert.False(t, IsLevelEnabled("debug"))

	SetLevel("debug")
	assert.True(t, IsLevelEnabled("debug"))

	SetLevel("warn")
	assert.False(t, IsLevelEnabled("debug"))
	assert.True(t, IsLevelEnabled("warn"))
}

//nolint:paralleltest
func TestIsLevelEnabled(t *testing.T) {
	SetLevel("error")
	assert.False(t, IsLevelEnabled("trace"))
	assert.False(t, IsLevelEnabled("debug"))
	assert.False(t, IsLevelEnabled("info"))
	assert.False(t, IsLevelEnabled("warn"))
	assert.True(t, IsLevelEnabled("error"))
	assert.True(t, IsLevelEnabled("panic"))

	SetLevel("debug")
	assert.False(t, IsLevelEnabled("trace"))
	assert.True(t, IsLevelEnabled("debug"))
	assert.True(t, IsLevelEnabled("info"))
	assert.True(t, IsLevelEnabled("warn"))
	assert.True(t, IsLevelEnabled("error"))
	assert.True(t, IsLevelEnabled("panic"))
}

// --- levelName ---

func TestLevelName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "TRACE", levelName(LevelTrace))
	assert.Equal(t, "DEBUG", levelName(slog.LevelDebug))
	assert.Equal(t, "INFO", levelName(slog.LevelInfo))
	assert.Equal(t, "WARN", levelName(slog.LevelWarn))
	assert.Equal(t, "ERROR", levelName(slog.LevelError))
}

// --- Format output shapes ---

//nolint:paralleltest
func TestLoggerFormat_HasBrackets(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	Info("message")
	output := buf.String()
	assert.Contains(t, output, "INFO")
	assert.Regexp(t, `\[\s*\d+\]`, output) // goroutine ID in brackets
	assert.NotContains(t, output, "level=")
}

//nolint:paralleltest
func TestLoggerFormat_HasUptime(t *testing.T) {
	startTime = time.Now().Add(-12 * time.Second)
	buf := resetForTest("info", styleLogger)
	Info("message")
	output := buf.String()
	// Uptime bracket appears before the level
	assert.Contains(t, output, "[  12]")
	assert.Regexp(t, `\[\s*\d+\].*INFO.*\[\s*\d+\]`, output)
}

//nolint:paralleltest
func TestLoggerWithoutUptime_NoUptimeField(t *testing.T) {
	buf := resetForTest("info", styleLoggerNoUptime)
	Info("message")
	output := buf.String()
	assert.Contains(t, output, "INFO")
	// Only one bracketed number (the goroutine ID)
	parts := strings.SplitN(output, "INFO", 2)
	assert.NotRegexp(t, `\[\s*\d+\]`, parts[0])
	assert.Regexp(t, `\[\s*\d+\]`, parts[1])
}

//nolint:paralleltest
func TestSlogFormat_HasLevelEquals(t *testing.T) {
	buf := resetForTest("info", styleSlog)
	Info("message")
	output := buf.String()
	assert.Contains(t, output, "level=INFO")
	assert.NotContains(t, output, "[INFO]")
	assert.Contains(t, output, "msg=")
}

//nolint:paralleltest
func TestLoggerStyle_AllLevels(t *testing.T) {
	buf := resetForTest("trace", styleLogger)
	Trace("trace message")
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")
	output := buf.String()
	assert.Contains(t, output, "TRACE")
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "ERROR")
	assert.Regexp(t, `\[\s*\d+\]`, output)
}

//nolint:paralleltest
func TestSlogStyle_AllLevels(t *testing.T) {
	buf := resetForTest("trace", styleSlog)
	Trace("trace message")
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")
	output := buf.String()
	assert.Contains(t, output, "level=TRACE")
	assert.Contains(t, output, "level=DEBUG")
	assert.Contains(t, output, "level=INFO")
	assert.Contains(t, output, "level=WARN")
	assert.Contains(t, output, "level=ERROR")
}

// --- SetLevel after Init ---

//nolint:paralleltest
func TestSetLevel_ChangesLevelAfterInit(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	Debug("this should not appear")
	assert.Empty(t, buf.String())

	SetLevel("debug")
	Debug("this should appear")
	assert.Contains(t, buf.String(), "this should appear")
}

//nolint:paralleltest
func TestSetLevel_CaseInsensitive(t *testing.T) {
	buf := resetForTest("info", styleLogger)

	SetLevel("DEBUG")
	Debug("debug message")
	assert.Contains(t, buf.String(), "debug message")

	buf.Reset()
	SetLevel("TRACE")
	Trace("trace message")
	assert.Contains(t, buf.String(), "trace message")
}

//nolint:paralleltest
func TestSetLevel_CanReduceVerbosity(t *testing.T) {
	buf := resetForTest("debug", styleLogger)
	Debug("should appear")
	assert.Contains(t, buf.String(), "should appear")

	buf.Reset()
	SetLevel("info")
	Debug("should not appear")
	assert.Empty(t, buf.String())
}

// --- Logr interface ---

//nolint:paralleltest
func TestLogr_ReturnsValidLogger(t *testing.T) {
	resetForTest("info", styleLogger)
	l := Logr()
	assert.NotNil(t, l)
	// Should be able to call Info without panicking
	l.Info("test message", "key", "value")
}

//nolint:paralleltest
func TestLogr_LogsWithCorrectStyle(t *testing.T) {
	buf := resetForTest("info", styleLogger)
	l := Logr()
	l.Info("message from logr", "key", "value")
	output := buf.String()
	assert.Contains(t, output, "INFO")
	assert.Regexp(t, `\[\s*\d+\]`, output)
	assert.Contains(t, output, "message from logr")
}

// --- Slim style ---

//nolint:paralleltest
func TestSlimStyle_HasBrackets(t *testing.T) {
	buf := resetForTest("info", styleSlim)
	Info("message")
	output := buf.String()
	assert.Contains(t, output, "[INFO]")
	assert.NotContains(t, output, "level=")
	assert.NotContains(t, output, "msg=")
}

//nolint:paralleltest
func TestSlimStyle_ArgsAsArray(t *testing.T) {
	buf := resetForTest("info", styleSlim)
	Info("message", "key", "value", "num", 42)
	line := buf.String()
	assert.Contains(t, line, "[")
	assert.Contains(t, line, "]")
	assert.Contains(t, line, "key")
	assert.Contains(t, line, "value")
	assert.Contains(t, line, "42")
	assert.Contains(t, line, "gid")
}

//nolint:paralleltest
func TestSlimStyle_NoArgsNoArray(t *testing.T) {
	buf := resetForTest("info", styleSlim)
	showThreads = false
	Info("message only")
	showThreads = true
	line := buf.String()
	assert.NotContains(t, line, "[]")
}

//nolint:paralleltest
func TestSlimStyle_AllLevels(t *testing.T) {
	buf := resetForTest("trace", styleSlim)
	Trace("trace msg")
	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")
	output := buf.String()
	assert.Contains(t, output, "[TRACE]")
	assert.Contains(t, output, "[DEBUG]")
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "[ERROR]")
}

// --- CLI styles ---

//nolint:paralleltest
func TestCLICompact_HasUptimeAndMessage(t *testing.T) {
	startTime = time.Now()
	buf := resetForTest("info", styleCLICompact)
	Info("message")
	output := buf.String()
	assert.Contains(t, output, "[   0] message")
}

//nolint:paralleltest
func TestCLICompact_NoLevelNoTimestamp(t *testing.T) {
	buf := resetForTest("info", styleCLICompact)
	Info("message")
	output := buf.String()
	assert.NotContains(t, output, "INFO")
	assert.NotRegexp(t, `\d{4}-\d{2}-\d{2}T`, output)
}

//nolint:paralleltest
func TestCLICompact_UptimeValue(t *testing.T) {
	startTime = time.Now().Add(-12 * time.Second)
	defer func() { startTime = time.Now() }()
	buf := resetForTest("info", styleCLICompact)
	Info("message")
	assert.Contains(t, buf.String(), "[  12]")
}

//nolint:paralleltest
func TestCLICompact_KeyValueArgs(t *testing.T) {
	buf := resetForTest("info", styleCLICompact)
	Info("message", "mykey", "myval")
	line := buf.String()
	assert.Contains(t, line, "mykey=")
	assert.Contains(t, line, "myval")
}

//nolint:paralleltest
func TestCLI_ShowsLevel(t *testing.T) {
	startTime = time.Now()
	buf := resetForTest("info", styleCLI)
	Info("message")
	output := buf.String()
	assert.Contains(t, output, "[   0] INFO  message")
	assert.NotRegexp(t, `\d{4}-\d{2}-\d{2}T`, output)
}

//nolint:paralleltest
func TestCLI_AllLevels(t *testing.T) {
	buf := resetForTest("trace", styleCLI)
	Trace("trace msg")
	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")
	output := buf.String()
	assert.Contains(t, output, "TRACE")
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "ERROR")
}

//nolint:paralleltest
func TestCLICompact_MessageUsesLevelColor(t *testing.T) {
	var buf bytes.Buffer
	noColor = false
	defer func() { noColor = true }()
	currentLevel.Set(parseLevel("error"))
	slog.SetDefault(slog.New(NewColoredCLIHandler(&currentLevel, &buf, false)))
	Error("boom")
	output := buf.String()
	assert.Contains(t, output, colorRed+"boom"+colorReset)
	assert.NotContains(t, output, "ERROR")
}

// --- Timezone ---

//nolint:paralleltest
func TestTimezone_DefaultIsBerlin(t *testing.T) {
	assert.Equal(t, "Europe/Berlin", currentLocation.String())
}

//nolint:paralleltest
func TestSetTimezone_ChangesTimestampZone(t *testing.T) {
	defer func() { _ = SetTimezone("Europe/Berlin") }()

	assert.NoError(t, SetTimezone("UTC"))
	buf := resetForTest("info", styleLogger)
	Info("zone check")
	line := buf.String()
	parts := strings.SplitN(line, " ", 2)
	assert.True(t, strings.HasSuffix(parts[0], "Z"), "expected UTC Z suffix, got %q", parts[0])

	assert.NoError(t, SetTimezone("America/New_York"))
	buf2 := resetForTest("info", styleLogger)
	Info("zone check")
	line2 := buf2.String()
	parts2 := strings.SplitN(line2, " ", 2)
	assert.Regexp(t, `[+-]\d{2}:\d{2}$`, parts2[0])
}

//nolint:paralleltest
func TestSetTimezone_RejectsInvalidZone(t *testing.T) {
	err := SetTimezone("Not/AZone")
	assert.Error(t, err)
}

// --- LogTo ---

//nolint:paralleltest
func TestLogTo(t *testing.T) {
	defer LogTo(nil)

	var buf strings.Builder
	LogTo(&buf)
	Info("Test log to writer")
	assert.Contains(t, buf.String(), "Test log to writer")
}
