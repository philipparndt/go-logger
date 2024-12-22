package logger

import (
	"github.com/stretchr/testify/assert"
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
