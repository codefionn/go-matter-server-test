package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	TraceLevel LogLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var levelNames = map[LogLevel]string{
	TraceLevel: "TRACE",
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

var levelColors = map[LogLevel]string{
	TraceLevel: "\033[36m", // Cyan
	DebugLevel: "\033[35m", // Magenta
	InfoLevel:  "\033[32m", // Green
	WarnLevel:  "\033[33m", // Yellow
	ErrorLevel: "\033[31m", // Red
	FatalLevel: "\033[91m", // Bright Red
}

const colorReset = "\033[0m"

// LogFormat represents the output format for logs
type LogFormat int

const (
	ConsoleFormat LogFormat = iota
	JSONFormat
)

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// Logger provides structured logging functionality
type Logger struct {
	level      LogLevel
	format     LogFormat
	writer     io.Writer
	name       string
	fields     []Field
	useColors  bool
	mu         sync.Mutex
	timeFormat string
}

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	Format     LogFormat
	Output     io.Writer
	UseColors  bool
	TimeFormat string
}

// New creates a new logger instance
func New(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}

	if config.TimeFormat == "" {
		config.TimeFormat = "2006-01-02 15:04:05.000"
	}

	return &Logger{
		level:      config.Level,
		format:     config.Format,
		writer:     config.Output,
		useColors:  config.UseColors,
		timeFormat: config.TimeFormat,
	}
}

// NewConsoleLogger creates a new console logger
func NewConsoleLogger(level LogLevel) *Logger {
	return New(Config{
		Level:     level,
		Format:    ConsoleFormat,
		UseColors: true,
	})
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger(level LogLevel) *Logger {
	return New(Config{
		Level:  level,
		Format: JSONFormat,
	})
}

// With creates a new logger with additional fields
func (l *Logger) With(fields ...Field) *Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &Logger{
		level:      l.level,
		format:     l.format,
		writer:     l.writer,
		name:       l.name,
		fields:     newFields,
		useColors:  l.useColors,
		timeFormat: l.timeFormat,
	}
}

// WithName creates a new logger with a name
func (l *Logger) WithName(name string) *Logger {
	newLogger := *l
	newLogger.name = name
	return &newLogger
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// IsEnabled returns true if the given level would be logged
func (l *Logger) IsEnabled(level LogLevel) bool {
	return level >= l.GetLevel()
}

// Log outputs a log entry at the specified level
func (l *Logger) Log(level LogLevel, msg string, fields ...Field) {
	if !l.IsEnabled(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Logger:    l.name,
		Fields:    append(l.fields, fields...),
	}

	// Add caller information for error and fatal levels
	if level >= ErrorLevel {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry.Caller = &CallerInfo{
				PC:       pc,
				File:     file,
				Line:     line,
				Function: runtime.FuncForPC(pc).Name(),
			}
		}
	}

	l.writeEntry(entry)

	// Exit on fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Trace logs a trace message
func (l *Logger) Trace(msg string, fields ...Field) {
	l.Log(TraceLevel, msg, fields...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Field) {
	l.Log(DebugLevel, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Field) {
	l.Log(InfoLevel, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Field) {
	l.Log(WarnLevel, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Field) {
	l.Log(ErrorLevel, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.Log(FatalLevel, msg, fields...)
}

// Printf-style logging methods

// Tracef logs a trace message with printf-style formatting
func (l *Logger) Tracef(format string, args ...interface{}) {
	if l.IsEnabled(TraceLevel) {
		l.Log(TraceLevel, fmt.Sprintf(format, args...))
	}
}

// Debugf logs a debug message with printf-style formatting
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.IsEnabled(DebugLevel) {
		l.Log(DebugLevel, fmt.Sprintf(format, args...))
	}
}

// Infof logs an info message with printf-style formatting
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.IsEnabled(InfoLevel) {
		l.Log(InfoLevel, fmt.Sprintf(format, args...))
	}
}

// Warnf logs a warning message with printf-style formatting
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.IsEnabled(WarnLevel) {
		l.Log(WarnLevel, fmt.Sprintf(format, args...))
	}
}

// Errorf logs an error message with printf-style formatting
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.IsEnabled(ErrorLevel) {
		l.Log(ErrorLevel, fmt.Sprintf(format, args...))
	}
}

// Fatalf logs a fatal message with printf-style formatting and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Log(FatalLevel, fmt.Sprintf(format, args...))
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Logger    string
	Fields    []Field
	Caller    *CallerInfo
}

// CallerInfo holds information about the calling code
type CallerInfo struct {
	PC       uintptr
	File     string
	Line     int
	Function string
}

func (l *Logger) writeEntry(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var output string

	switch l.format {
	case ConsoleFormat:
		output = l.formatConsole(entry)
	case JSONFormat:
		output = l.formatJSON(entry)
	default:
		output = l.formatConsole(entry)
	}

	fmt.Fprintln(l.writer, output)
}

func (l *Logger) formatConsole(entry LogEntry) string {
	var b strings.Builder

	// Timestamp
	b.WriteString(entry.Timestamp.Format(l.timeFormat))
	b.WriteString(" ")

	// Level with color
	levelName := levelNames[entry.Level]
	if l.useColors {
		color := levelColors[entry.Level]
		b.WriteString(color)
		b.WriteString(fmt.Sprintf("%-5s", levelName))
		b.WriteString(colorReset)
	} else {
		b.WriteString(fmt.Sprintf("%-5s", levelName))
	}
	b.WriteString(" ")

	// Logger name
	if entry.Logger != "" {
		b.WriteString("[")
		b.WriteString(entry.Logger)
		b.WriteString("] ")
	}

	// Message
	b.WriteString(entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		b.WriteString(" {")
		for i, field := range entry.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%s=%v", field.Key, field.Value))
		}
		b.WriteString("}")
	}

	// Caller info
	if entry.Caller != nil {
		parts := strings.Split(entry.Caller.File, "/")
		file := parts[len(parts)-1]
		b.WriteString(fmt.Sprintf(" (%s:%d)", file, entry.Caller.Line))
	}

	return b.String()
}

func (l *Logger) formatJSON(entry LogEntry) string {
	var b strings.Builder
	b.WriteString("{")

	// Timestamp
	b.WriteString(fmt.Sprintf(`"timestamp":"%s"`, entry.Timestamp.Format(time.RFC3339Nano)))

	// Level
	b.WriteString(fmt.Sprintf(`,"level":"%s"`, levelNames[entry.Level]))

	// Logger name
	if entry.Logger != "" {
		b.WriteString(fmt.Sprintf(`,"logger":"%s"`, entry.Logger))
	}

	// Message
	b.WriteString(fmt.Sprintf(`,"message":"%s"`, escapeJSON(entry.Message)))

	// Fields
	for _, field := range entry.Fields {
		b.WriteString(fmt.Sprintf(`,"%s":%s`, escapeJSON(field.Key), formatJSONValue(field.Value)))
	}

	// Caller info
	if entry.Caller != nil {
		b.WriteString(fmt.Sprintf(`,"caller":{"file":"%s","line":%d,"function":"%s"}`,
			escapeJSON(entry.Caller.File),
			entry.Caller.Line,
			escapeJSON(entry.Caller.Function)))
	}

	b.WriteString("}")
	return b.String()
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func formatJSONValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, escapeJSON(val))
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf(`"%s"`, escapeJSON(fmt.Sprintf("%v", val)))
	}
}

// Helper functions for creating fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

func ErrorField(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// ParseLogLevel parses a string log level
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "trace":
		return TraceLevel, nil
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// Global logger instance
var defaultLogger = NewConsoleLogger(InfoLevel)

// Global logging functions
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func SetFormat(format LogFormat) {
	defaultLogger.format = format
}

func Trace(msg string, fields ...Field) {
	defaultLogger.Trace(msg, fields...)
}

func Debug(msg string, fields ...Field) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	defaultLogger.Fatal(msg, fields...)
}

func Tracef(format string, args ...interface{}) {
	defaultLogger.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}
