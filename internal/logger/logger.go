package logger

import (
	"io"
	"log"
	"os"
	"runtime"
)

const (
	LogLevelDebug = 0
	LogLevelInfo  = 1
	LogLevelWarn  = 2
	LogLevelError = 3
)

var (
	logLevel           = LogLevelInfo
	logFile  io.Writer = os.Stdout
)

var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
)

func SetLogLevel(level string) {
	switch level {
	case "debug":
		logLevel = LogLevelDebug
	case "info":
		logLevel = LogLevelInfo
	case "warn":
		logLevel = LogLevelWarn
	case "error":
		logLevel = LogLevelError
	default:
		logLevel = LogLevelInfo
	}
}

func SetLogFile(path string) {
	if path == "" {
		logFile = os.Stdout
		log.SetOutput(logFile)
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("[WARN] Failed to open log file: %v, fallback to stdout", err)
		logFile = os.Stdout
		log.SetOutput(logFile)
		return
	}
	logFile = f
	log.SetOutput(logFile)
}

func isatty(w io.Writer) bool {
	if runtime.GOOS == "windows" {
		return w == os.Stdout || w == os.Stderr
	}
	return w == os.Stdout || w == os.Stderr
}

func colorize(level, msg string) string {
	// 只有输出到控制台才加颜色，输出到文件不加颜色
	if !isatty(logFile) {
		return msg
	}
	switch level {
	case "DEBUG":
		return colorGray + msg + colorReset
	case "INFO":
		return msg // INFO 不加颜色
	case "WARN":
		return colorYellow + msg + colorReset
	case "ERROR":
		return colorRed + msg + colorReset
	default:
		return msg
	}
}

func Debug(format string, a ...interface{}) {
	if logLevel <= LogLevelDebug {
		log.SetOutput(logFile)
		msg := colorize("DEBUG", "[DEBUG] "+format)
		log.Printf(msg, a...)
	}
}
func Info(format string, a ...interface{}) {
	if logLevel <= LogLevelInfo {
		log.SetOutput(logFile)
		msg := colorize("INFO", "[INFO] "+format)
		log.Printf(msg, a...)
	}
}
func Warn(format string, a ...interface{}) {
	if logLevel <= LogLevelWarn {
		log.SetOutput(logFile)
		msg := colorize("WARN", "[WARN] "+format)
		log.Printf(msg, a...)
	}
}
func Error(format string, a ...interface{}) {
	if logLevel <= LogLevelError {
		log.SetOutput(logFile)
		msg := colorize("ERROR", "[ERROR] "+format)
		log.Printf(msg, a...)
	}
}
