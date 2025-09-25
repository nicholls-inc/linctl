package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Test default logger creation
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}

	// Test with environment variables
	os.Setenv("LINCTL_LOG_LEVEL", "debug")
	os.Setenv("LINCTL_LOG_FORMAT", "json")
	defer func() {
		os.Unsetenv("LINCTL_LOG_LEVEL")
		os.Unsetenv("LINCTL_LOG_FORMAT")
	}()

	logger = NewLogger()
	if logger == nil {
		t.Fatal("NewLogger() with env vars returned nil")
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithConfig(DebugLevel, "text", &buf)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	// Check that all levels are present
	if !strings.Contains(output, "DEBUG") {
		t.Error("Debug message not found")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Info message not found")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Warn message not found")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Error message not found")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithConfig(WarnLevel, "text", &buf)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have WARN and ERROR
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines with WARN level, got %d", len(lines))
	}

	if strings.Contains(output, "DEBUG") || strings.Contains(output, "INFO") {
		t.Error("Debug/Info messages should be filtered out at WARN level")
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithConfig(InfoLevel, "json", &buf)

	logger.Info("test message", String("key", "value"), Int("number", 42))

	output := strings.TrimSpace(buf.String())

	// Parse as JSON
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log entry: %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}

	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", entry.Message)
	}

	if entry.Fields["key"] != "value" {
		t.Errorf("Expected field key=value, got %v", entry.Fields["key"])
	}

	if entry.Fields["number"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected field number=42, got %v", entry.Fields["number"])
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithConfig(InfoLevel, "text", &buf)

	logger.Info("test message", String("key", "value"), Int("number", 42))

	output := strings.TrimSpace(buf.String())

	if !strings.Contains(output, "INFO") {
		t.Error("Text format should contain level")
	}

	if !strings.Contains(output, "test message") {
		t.Error("Text format should contain message")
	}

	if !strings.Contains(output, "key=value") {
		t.Error("Text format should contain fields")
	}

	if !strings.Contains(output, "number=42") {
		t.Error("Text format should contain numeric fields")
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithConfig(InfoLevel, "json", &buf)

	// Create logger with base fields
	contextLogger := logger.With(String("service", "linctl"), String("version", "1.0"))

	contextLogger.Info("test message", String("request_id", "123"))

	output := strings.TrimSpace(buf.String())

	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log entry: %v", err)
	}

	// Check base fields
	if entry.Fields["service"] != "linctl" {
		t.Errorf("Expected base field service=linctl, got %v", entry.Fields["service"])
	}

	if entry.Fields["version"] != "1.0" {
		t.Errorf("Expected base field version=1.0, got %v", entry.Fields["version"])
	}

	// Check message field
	if entry.Fields["request_id"] != "123" {
		t.Errorf("Expected message field request_id=123, got %v", entry.Fields["request_id"])
	}
}

func TestFieldHelpers(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithConfig(InfoLevel, "json", &buf)

	err := &testError{"test error"}
	duration := 5 * time.Second

	logger.Info("test message",
		String("str", "value"),
		Int("int", 42),
		Bool("bool", true),
		Duration("duration", duration),
		Error(err),
	)

	output := strings.TrimSpace(buf.String())

	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log entry: %v", err)
	}

	if entry.Fields["str"] != "value" {
		t.Errorf("String field incorrect: %v", entry.Fields["str"])
	}

	if entry.Fields["int"] != float64(42) {
		t.Errorf("Int field incorrect: %v", entry.Fields["int"])
	}

	if entry.Fields["bool"] != true {
		t.Errorf("Bool field incorrect: %v", entry.Fields["bool"])
	}

	if entry.Fields["duration"] != "5s" {
		t.Errorf("Duration field incorrect: %v", entry.Fields["duration"])
	}

	if entry.Fields["error"] != "test error" {
		t.Errorf("Error field incorrect: %v", entry.Fields["error"])
	}
}

func TestNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()

	// These should not panic
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	contextLogger := logger.With(String("key", "value"))
	contextLogger.Info("test")
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("LogLevel(%d).String() = %s, expected %s", test.level, test.level.String(), test.expected)
		}
	}
}

func TestEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithConfig(InfoLevel, "json", &buf)

	logger.Info("test message")

	output := strings.TrimSpace(buf.String())

	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log entry: %v", err)
	}

	// Fields should be nil when empty
	if entry.Fields != nil {
		t.Errorf("Expected nil fields for empty field list, got %v", entry.Fields)
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
