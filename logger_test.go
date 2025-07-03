package logger

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	Info("Test")
	Info("Test", "with", "data")
}

func TestTrace(t *testing.T) {
	SetLevel("trace")
	Trace("Test")
	Trace("Test", "with", "data")
}

func TestDebug(t *testing.T) {
	SetLevel("debug")
	Debug("Test")
	Debug("Test", "with", "data")
}

func TestWarn(t *testing.T) {
	Warn("Test")
	Warn("Test", "with", "data")
}

func TestError(t *testing.T) {
	Error("Test")
	Error("Test", "with", "data")
}

func TestPanic(t *testing.T) {
	assert.Panics(t, func() { Panic("Test") })
	assert.Panics(t, func() { Panic("Test", "with", "data") })
}

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

func TestLogTo(t *testing.T) {
	defer LogTo(nil)

	var buf strings.Builder
	LogTo(&buf)
	Info("Test log to writer")
	assert.Contains(t, buf.String(), "Test log to writer")
}
