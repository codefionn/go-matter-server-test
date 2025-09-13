package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"Trace", TraceLevel, "TRACE"},
		{"Debug", DebugLevel, "DEBUG"},
		{"Info", InfoLevel, "INFO"},
		{"Warn", WarnLevel, "WARN"},
		{"Error", ErrorLevel, "ERROR"},
		{"Fatal", FatalLevel, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if levelNames[tt.level] != tt.expected {
				t.Errorf("Expected level name %s, got %s", tt.expected, levelNames[tt.level])
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
		hasError bool
	}{
		{"trace", TraceLevel, false},
		{"debug", DebugLevel, false},
		{"info", InfoLevel, false},
		{"warn", WarnLevel, false},
		{"warning", WarnLevel, false},
		{"error", ErrorLevel, false},
		{"fatal", FatalLevel, false},
		{"invalid", InfoLevel, true},
		{"", InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLogLevel(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if level != tt.expected {
					t.Errorf("Expected level %v, got %v", tt.expected, level)
				}
			}
		})
	}
}

func TestLoggerCreation(t *testing.T) {
	var buf bytes.Buffer

	logger := New(Config{
		Level:  DebugLevel,
		Format: ConsoleFormat,
		Output: &buf,
	})

	if logger.level != DebugLevel {
		t.Errorf("Expected level %v, got %v", DebugLevel, logger.level)
	}
	if logger.format != ConsoleFormat {
		t.Errorf("Expected format %v, got %v", ConsoleFormat, logger.format)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	logger := New(Config{
		Level:  InfoLevel,
		Format: ConsoleFormat,
		Output: &buf,
	})

	loggerWithFields := logger.With(
		String("key1", "value1"),
		Int("key2", 42),
	)

	if len(loggerWithFields.fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(loggerWithFields.fields))
	}

	if loggerWithFields.fields[0].Key != "key1" || loggerWithFields.fields[0].Value != "value1" {
		t.Error("First field not set correctly")
	}

	if loggerWithFields.fields[1].Key != "key2" || loggerWithFields.fields[1].Value != 42 {
		t.Error("Second field not set correctly")
	}
}

func TestLoggerWithName(t *testing.T) {
	var buf bytes.Buffer

	logger := New(Config{
		Level:  InfoLevel,
		Format: ConsoleFormat,
		Output: &buf,
	})

	namedLogger := logger.WithName("test-logger")

	if namedLogger.name != "test-logger" {
		t.Errorf("Expected name 'test-logger', got '%s'", namedLogger.name)
	}
}

func TestLogLevel(t *testing.T) {
	var buf bytes.Buffer

	logger := New(Config{
		Level:  WarnLevel,
		Format: ConsoleFormat,
		Output: &buf,
	})

	// Should log (level >= WarnLevel)
	logger.Warn("warning message")
	logger.Error("error message")

	// Should not log (level < WarnLevel)
	logger.Info("info message")
	logger.Debug("debug message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have exactly 2 lines (warn and error)
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d: %v", len(lines), lines)
	}

	if !strings.Contains(lines[0], "warning message") {
		t.Error("Warning message not found in output")
	}
	if !strings.Contains(lines[1], "error message") {
		t.Error("Error message not found in output")
	}
}

func TestIsEnabled(t *testing.T) {
	logger := New(Config{
		Level: WarnLevel,
	})

	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{TraceLevel, false},
		{DebugLevel, false},
		{InfoLevel, false},
		{WarnLevel, true},
		{ErrorLevel, true},
		{FatalLevel, true},
	}

	for _, tt := range tests {
		t.Run(levelNames[tt.level], func(t *testing.T) {
			if logger.IsEnabled(tt.level) != tt.expected {
				t.Errorf("IsEnabled(%v) = %v, expected %v", tt.level, logger.IsEnabled(tt.level), tt.expected)
			}
		})
	}
}

func TestConsoleFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := New(Config{
		Level:      InfoLevel,
		Format:     ConsoleFormat,
		Output:     &buf,
		UseColors:  false,
		TimeFormat: "2006-01-02 15:04:05.000",
	})

	logger.Info("test message", String("key", "value"))

	output := buf.String()

	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level in output")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Expected message in output")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Expected field in output")
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: &buf,
	})

	logger.Info("test message", String("key", "value"))

	output := strings.TrimSpace(buf.String())

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", logEntry["level"])
	}
	if logEntry["message"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", logEntry["message"])
	}
	if logEntry["key"] != "value" {
		t.Errorf("Expected key 'value', got %v", logEntry["key"])
	}
}

func TestFieldHelpers(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected interface{}
	}{
		{"String", String("test", "value"), "value"},
		{"Int", Int("test", 42), 42},
		{"Int64", Int64("test", int64(42)), int64(42)},
		{"Float64", Float64("test", 3.14), 3.14},
		{"Bool", Bool("test", true), true},
		{"Duration", Duration("test", time.Second), "1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Key != "test" {
				t.Errorf("Expected key 'test', got '%s'", tt.field.Key)
			}
			if tt.field.Value != tt.expected {
				t.Errorf("Expected value %v, got %v", tt.expected, tt.field.Value)
			}
		})
	}
}

func TestErrorField(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected interface{}
	}{
		{"nil error", nil, nil},
		{"actual error", errors.New("test error"), "test error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := ErrorField(tt.err)
			if field.Key != "error" {
				t.Errorf("Expected key 'error', got '%s'", field.Key)
			}
			if tt.err == nil {
				if field.Value != nil {
					t.Errorf("Expected nil value for nil error, got %v", field.Value)
				}
			} else {
				if field.Value != tt.err.Error() {
					t.Errorf("Expected error message '%s', got %v", tt.err.Error(), field.Value)
				}
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	logger := New(Config{
		Level: InfoLevel,
	})

	if logger.GetLevel() != InfoLevel {
		t.Errorf("Initial level should be Info, got %v", logger.GetLevel())
	}

	logger.SetLevel(DebugLevel)

	if logger.GetLevel() != DebugLevel {
		t.Errorf("Level should be Debug after SetLevel, got %v", logger.GetLevel())
	}
}

func TestFormatJSONValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", "test", `"test"`},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool", true, "true"},
		{"complex", map[string]int{"key": 1}, `"map[key:1]"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatJSONValue(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEscapeJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `hello`},
		{`hello"world`, `hello\"world`},
		{"hello\nworld", `hello\nworld`},
		{"hello\tworld", `hello\tworld`},
		{`hello\world`, `hello\\world`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeJSON(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test that global logger functions don't panic
	SetLevel(DebugLevel)
	SetFormat(JSONFormat)

	// These should not panic
	Trace("trace message")
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	Tracef("trace %s", "formatted")
	Debugf("debug %s", "formatted")
	Infof("info %s", "formatted")
	Warnf("warn %s", "formatted")
	Errorf("error %s", "formatted")
}
