package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var customLogger *log.Logger

const (
	debugLevel int = 4
	infoLevel  int = 3
	warnLevel  int = 2
	errorLevel int = 1
	panicLevel int = 0
)

var logLevel = infoLevel

var (
	red    = "\033[31m"
	yellow = "\033[33m"
	nc     = "\033[0m"
	purple = "\033[35m"
	gray   = "\033[37m"
)

func init() {
	customLogger = log.New(os.Stdout, "", 0)

	if os.Getenv("NO_COLOR") != "" {
		red = ""
		yellow = ""
		nc = ""
		purple = ""
		gray = ""
	}
}

func logMessage(severity string, color string, message string, a ...any) {
	var level = fmt.Sprintf("%s[%s]%s", color, strings.ToUpper(severity), nc)
	var timedate = time.Now().Format("2006-01-02T15:04:05 MST")
	var data = ""
	if len(a) > 0 {
		data = fmt.Sprintf(" %s%s%s", gray, a, nc)
	}

	customLogger.Printf("%s %s %s%s\n", timedate, level, message, data)
}

func Debug(message string, a ...any) {
	if logLevel >= debugLevel {
		logMessage("debug", purple, message, a...)
	}
}

func Info(message string, a ...any) {
	if logLevel >= infoLevel {
		logMessage("info", nc, message, a...)
	}
}

func Warn(message string, a ...any) {
	if logLevel >= warnLevel {
		logMessage("warn", yellow, message, a...)
	}
}

func Error(message string, a ...any) {
	logMessage("error", red, message, a...)
}

func Panic(message string, a ...any) {
	logMessage("panic", red, message, a...)
	panic(message)
}

func SetLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		logLevel = debugLevel
	case "info":
		logLevel = infoLevel
	case "warn":
		logLevel = warnLevel
	case "error":
		logLevel = errorLevel
	case "panic":
		logLevel = panicLevel
	}
}
