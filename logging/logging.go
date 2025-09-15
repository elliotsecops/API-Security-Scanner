package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a structured logger
type Logger struct {
	level     LogLevel
	format    string // "text" or "json"
	timestamp bool
}

// NewLogger creates a new logger with the specified level and format
func NewLogger(level LogLevel, format string) *Logger {
	return &Logger{
		level:     level,
		format:    format,
		timestamp: true,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetFormat sets the output format ("text" or "json")
func (l *Logger) SetFormat(format string) {
	l.format = format
}

// EnableTimestamp enables or disables timestamp in logs
func (l *Logger) EnableTimestamp(enabled bool) {
	l.timestamp = enabled
}

// Log logs a message with the specified level and fields
func (l *Logger) Log(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	switch l.format {
	case "json":
		l.logJSON(level, message, fields)
	default:
		l.logText(level, message, fields)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.Log(DEBUG, message, fields)
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.Log(INFO, message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.Log(WARN, message, fields)
}

// Error logs an error message
func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.Log(ERROR, message, fields)
}

func (l *Logger) logText(level LogLevel, message string, fields map[string]interface{}) {
	var parts []string

	if l.timestamp {
		parts = append(parts, time.Now().Format("2006-01-02T15:04:05Z07:00"))
	}

	parts = append(parts, fmt.Sprintf("[%s]", level.String()))

	parts = append(parts, message)

	// Add fields
	for key, value := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}

	fmt.Fprintln(os.Stderr, strings.Join(parts, " "))
}

func (l *Logger) logJSON(level LogLevel, message string, fields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"level":   level.String(),
		"message": message,
	}

	if l.timestamp {
		logEntry["timestamp"] = time.Now().Format(time.RFC3339)
	}

	// Add fields
	for key, value := range fields {
		logEntry[key] = value
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		// Fallback to text logging if JSON marshaling fails
		l.logText(level, message, fields)
		return
	}

	fmt.Fprintln(os.Stderr, string(jsonData))
}

// Global logger instance
var globalLogger *Logger

func init() {
	// Initialize with INFO level and text format by default
	globalLogger = NewLogger(INFO, "text")
}

// SetGlobalLevel sets the level for the global logger
func SetGlobalLevel(level LogLevel) {
	globalLogger.SetLevel(level)
}

// SetGlobalFormat sets the format for the global logger
func SetGlobalFormat(format string) {
	globalLogger.SetFormat(format)
}

// Debug logs a debug message using the global logger
func Debug(message string, fields map[string]interface{}) {
	globalLogger.Debug(message, fields)
}

// Info logs an info message using the global logger
func Info(message string, fields map[string]interface{}) {
	globalLogger.Info(message, fields)
}

// Warn logs a warning message using the global logger
func Warn(message string, fields map[string]interface{}) {
	globalLogger.Warn(message, fields)
}

// Error logs an error message using the global logger
func Error(message string, fields map[string]interface{}) {
	globalLogger.Error(message, fields)
}