package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Field represents a structured logging field
type Field struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// StructuredLogger implements the Logger interface
type StructuredLogger struct {
	level      LogLevel
	format     string // "json" or "text"
	writer     io.Writer
	baseFields map[string]interface{}
}

// NewLogger creates a new structured logger
func NewLogger() Logger {
	level := InfoLevel
	format := "text"
	
	// Check environment variables
	if levelStr := os.Getenv("LINCTL_LOG_LEVEL"); levelStr != "" {
		switch strings.ToLower(levelStr) {
		case "debug":
			level = DebugLevel
		case "info":
			level = InfoLevel
		case "warn", "warning":
			level = WarnLevel
		case "error":
			level = ErrorLevel
		}
	}
	
	if formatStr := os.Getenv("LINCTL_LOG_FORMAT"); formatStr != "" {
		if strings.ToLower(formatStr) == "json" {
			format = "json"
		}
	}
	
	return &StructuredLogger{
		level:      level,
		format:     format,
		writer:     os.Stderr,
		baseFields: make(map[string]interface{}),
	}
}

// NewLoggerWithConfig creates a logger with specific configuration
func NewLoggerWithConfig(level LogLevel, format string, writer io.Writer) Logger {
	return &StructuredLogger{
		level:      level,
		format:     format,
		writer:     writer,
		baseFields: make(map[string]interface{}),
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, fields ...Field) {
	if l.level <= DebugLevel {
		l.log(DebugLevel, msg, fields...)
	}
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, fields ...Field) {
	if l.level <= InfoLevel {
		l.log(InfoLevel, msg, fields...)
	}
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, fields ...Field) {
	if l.level <= WarnLevel {
		l.log(WarnLevel, msg, fields...)
	}
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, fields ...Field) {
	if l.level <= ErrorLevel {
		l.log(ErrorLevel, msg, fields...)
	}
}

// With creates a new logger with additional base fields
func (l *StructuredLogger) With(fields ...Field) Logger {
	newFields := make(map[string]interface{})
	
	// Copy existing base fields
	for k, v := range l.baseFields {
		newFields[k] = v
	}
	
	// Add new fields
	for _, field := range fields {
		newFields[field.Key] = field.Value
	}
	
	return &StructuredLogger{
		level:      l.level,
		format:     l.format,
		writer:     l.writer,
		baseFields: newFields,
	}
}

// log performs the actual logging
func (l *StructuredLogger) log(level LogLevel, msg string, fields ...Field) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level.String(),
		Message:   msg,
		Fields:    make(map[string]interface{}),
	}
	
	// Add base fields
	for k, v := range l.baseFields {
		entry.Fields[k] = v
	}
	
	// Add message fields
	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}
	
	// Remove fields if empty
	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}
	
	if l.format == "json" {
		l.logJSON(entry)
	} else {
		l.logText(entry)
	}
}

// logJSON outputs the log entry as JSON
func (l *StructuredLogger) logJSON(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple text logging if JSON marshaling fails
		fmt.Fprintf(l.writer, "[%s] %s %s\n", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
		return
	}
	
	fmt.Fprintln(l.writer, string(data))
}

// logText outputs the log entry as human-readable text
func (l *StructuredLogger) logText(entry LogEntry) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	
	if entry.Fields == nil || len(entry.Fields) == 0 {
		fmt.Fprintf(l.writer, "[%s] %s %s\n", timestamp, entry.Level, entry.Message)
		return
	}
	
	// Format fields as key=value pairs
	var fieldStrs []string
	for k, v := range entry.Fields {
		fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
	}
	
	fmt.Fprintf(l.writer, "[%s] %s %s %s\n", timestamp, entry.Level, entry.Message, strings.Join(fieldStrs, " "))
}

// NoOpLogger is a logger that does nothing (for testing)
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(msg string, fields ...Field) {}
func (n *NoOpLogger) Info(msg string, fields ...Field)  {}
func (n *NoOpLogger) Warn(msg string, fields ...Field)  {}
func (n *NoOpLogger) Error(msg string, fields ...Field) {}
func (n *NoOpLogger) With(fields ...Field) Logger       { return n }

// NewNoOpLogger creates a no-op logger for testing
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

// Helper functions for creating fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

func Error(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: "<nil>"}
	}
	return Field{Key: "error", Value: err.Error()}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}