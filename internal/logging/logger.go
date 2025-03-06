package logging

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed debugging information
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARN level for concerning but non-critical issues
	WARN
	// ERROR level for errors that might allow the application to continue
	ERROR
	// FATAL level for critical errors that prevent the application from continuing
	FATAL
)

var (
	// Default logger settings
	defaultLogger = &Logger{
		level:  INFO, // Default log level
		prefix: "",
		flags:  log.LstdFlags,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
)

// Logger represents a custom logger with levels
type Logger struct {
	level  LogLevel
	prefix string
	flags  int
	logger *log.Logger
}

// GetLogger returns the default logger
func GetLogger() *Logger {
	return defaultLogger
}

// NewLogger creates a new logger with custom settings
func NewLogger(level LogLevel, prefix string, flags int) *Logger {
	return &Logger{
		level:  level,
		prefix: prefix,
		flags:  flags,
		logger: log.New(os.Stdout, prefix, flags),
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetPrefix sets the log prefix
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
	l.logger.SetPrefix(prefix)
}

// levelToString converts a LogLevel to its string representation
func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// getCallerInfo returns the filename and line number of the calling function
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}
	// Get just the file name, not the full path
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	return fmt.Sprintf("%s:%d", file, line)
}

// log logs a message at the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information (skip 2 levels: log() and the specific level function)
	caller := getCallerInfo(2)

	// Format the message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelToString(level)
	message := fmt.Sprintf(format, args...)

	logLine := fmt.Sprintf("[%s] [%s] %s - %s", timestamp, levelStr, caller, message)
	l.logger.Println(logLine)

	// If fatal, exit the application
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits the application
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// SetGlobalLevel sets the level of the default logger
func SetGlobalLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetGlobalPrefix sets the prefix of the default logger
func SetGlobalPrefix(prefix string) {
	defaultLogger.SetPrefix(prefix)
}

// Global convenience functions

// Debug logs a debug message to the default logger
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message to the default logger
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message to the default logger
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error logs an error message to the default logger
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal logs a fatal message to the default logger and exits the application
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}
