package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var levelNames = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

func (l Level) String() string {
	if name, ok := levelNames[l]; ok {
		return name
	}
	return "UNKNOWN"
}

// ParseLevel parses a level string into a Level
func ParseLevel(level string) Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Field represents a log field (key-value pair)
type Field struct {
	Key   string
	Value interface{}
}

// Fields is a map of fields
type Fields map[string]interface{}

// Entry represents a single log entry
type Entry struct {
	Level      Level     `json:"level"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	Fields     Fields    `json:"fields,omitempty"`
	Caller     string    `json:"caller,omitempty"`
	StackTrace string    `json:"stack_trace,omitempty"`
}

// Config holds logger configuration
type Config struct {
	// Buffer size for async logging (default: 1000)
	BufferSize int

	// Output destination (default: os.Stdout)
	Output io.Writer

	// Time format (default: "2006-01-02 15:04:05.000")
	TimeFormat string

	// Include caller information (file:line)
	IncludeCaller bool

	// Log level (default: InfoLevel)
	Level Level

	// Use JSON format instead of text
	JSONFormat bool

	// Color output (only for text format)
	ColorOutput bool

	// Maximum fields per entry (default: 50)
	MaxFields int
}

// AsyncLogger provides asynchronous logging capabilities
type AsyncLogger struct {
	entries chan Entry
	wg      sync.WaitGroup
	stop    chan struct{}
	output  io.Writer
	config  Config
	mu      sync.RWMutex
	closed  bool
	stats   LoggerStats
}

// LoggerStats holds logger statistics
type LoggerStats struct {
	TotalEntries   int64
	DroppedEntries int64
	Errors         int64
	mu             sync.RWMutex
}

// New creates a new async logger
func New(cfg Config) (*AsyncLogger, error) {
	// Set defaults
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 1000
	}

	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = "2006-01-02 15:04:05.000"
	}

	if cfg.MaxFields <= 0 {
		cfg.MaxFields = 50
	}

	logger := &AsyncLogger{
		entries: make(chan Entry, cfg.BufferSize),
		stop:    make(chan struct{}),
		output:  cfg.Output,
		config:  cfg,
	}

	logger.wg.Add(1)
	go logger.process()

	return logger, nil
}

// MustNew creates a new async logger or panics
func MustNew(cfg Config) *AsyncLogger {
	logger, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

// process handles log entries asynchronously
func (l *AsyncLogger) process() {
	defer l.wg.Done()

	for {
		select {
		case entry := <-l.entries:
			l.updateStats(true, false)
			if err := l.write(entry); err != nil {
				l.updateStats(false, true)
				// Fallback to stderr
				fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
			}
		case <-l.stop:
			for len(l.entries) > 0 {
				entry := <-l.entries
				if err := l.write(entry); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to write log during shutdown: %v\n", err)
				}
			}
			return
		}
	}
}

// write writes a single log entry to output
func (l *AsyncLogger) write(entry Entry) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.closed {
		return fmt.Errorf("logger is closed")
	}

	var data []byte
	var err error

	if l.config.JSONFormat {
		data, err = json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry: %w", err)
		}
		data = append(data, '\n')
	} else {
		data = []byte(l.formatText(entry))
	}

	_, err = l.output.Write(data)
	return err
}

// formatText formats a log entry as text
func (l *AsyncLogger) formatText(entry Entry) string {
	var builder strings.Builder

	timestamp := entry.Timestamp.Format(l.config.TimeFormat)
	builder.WriteString("[")
	builder.WriteString(timestamp)
	builder.WriteString("] ")

	levelStr := entry.Level.String()
	if l.config.ColorOutput {
		levelStr = l.colorizeLevel(levelStr, entry.Level)
	}
	builder.WriteString(levelStr)
	builder.WriteString(": ")

	builder.WriteString(entry.Message)

	if l.config.IncludeCaller && entry.Caller != "" {
		builder.WriteString(" (")
		builder.WriteString(entry.Caller)
		builder.WriteString(")")
	}

	if len(entry.Fields) > 0 {
		builder.WriteString(" ")
		builder.WriteString(l.formatFields(entry.Fields))
	}

	if entry.StackTrace != "" {
		builder.WriteString("\n")
		builder.WriteString(entry.StackTrace)
	}

	builder.WriteString("\n")
	return builder.String()
}

func (l *AsyncLogger) formatFields(fields Fields) string {
	var builder strings.Builder
	builder.WriteString("{")

	i := 0
	for k, v := range fields {
		if i >= l.config.MaxFields {
			builder.WriteString(fmt.Sprintf(", ... (%d more)", len(fields)-l.config.MaxFields))
			break
		}
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%s=%v", k, v))
		i++
	}

	builder.WriteString("}")
	return builder.String()
}

// colorizeLevel adds ANSI color codes to level string
func (l *AsyncLogger) colorizeLevel(level string, levelType Level) string {
	colors := map[Level]string{
		DebugLevel: "\033[36m", // Cyan
		InfoLevel:  "\033[32m", // Green
		WarnLevel:  "\033[33m", // Yellow
		ErrorLevel: "\033[31m", // Red
		FatalLevel: "\033[35m", // Magenta
	}

	color, ok := colors[levelType]
	if !ok {
		return level
	}

	return color + level + "\033[0m"
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	parts := strings.Split(file, "/")
	if len(parts) >= 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// Log adds a log entry with specified level
func (l *AsyncLogger) Log(level Level, message string, fields Fields) {
	if level < l.config.Level {
		return
	}

	entry := Entry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Fields:    fields,
	}

	if l.config.IncludeCaller {
		entry.Caller = getCallerInfo(3)
	}

	if level == FatalLevel && entry.StackTrace == "" {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		entry.StackTrace = string(buf[:n])
	}

	select {
	case l.entries <- entry:
	default:
		l.updateStats(false, false)
		if err := l.write(entry); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write log synchronously: %v\n", err)
		}
	}

	if level == FatalLevel {
		l.Close()
		os.Exit(1)
	}
}

// updateStats updates logger statistics
func (l *AsyncLogger) updateStats(total, isError bool) {
	l.stats.mu.Lock()
	defer l.stats.mu.Unlock()

	if total {
		l.stats.TotalEntries++
	} else {
		l.stats.DroppedEntries++
	}

	if isError {
		l.stats.Errors++
	}
}

// Debug logs debug level message
func (l *AsyncLogger) Debug(message string, fields Fields) {
	l.Log(DebugLevel, message, fields)
}

// Info logs info level message
func (l *AsyncLogger) Info(message string, fields Fields) {
	l.Log(InfoLevel, message, fields)
}

// Warn logs warning level message
func (l *AsyncLogger) Warn(message string, fields Fields) {
	l.Log(WarnLevel, message, fields)
}

// Error logs error level message
func (l *AsyncLogger) Error(message string, fields Fields) {
	l.Log(ErrorLevel, message, fields)
}

// Fatal logs fatal level message and exits
func (l *AsyncLogger) Fatal(message string, fields Fields) {
	l.Log(FatalLevel, message, fields)
}

// WithFields returns a new logger with preset fields
func (l *AsyncLogger) WithFields(fields Fields) *LoggerWithFields {
	return &LoggerWithFields{
		logger: l,
		fields: fields,
	}
}

// WithField returns a new logger with a single preset field
func (l *AsyncLogger) WithField(key string, value interface{}) *LoggerWithFields {
	return l.WithFields(Fields{key: value})
}

// Close gracefully closes the logger
func (l *AsyncLogger) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	l.mu.Unlock()

	close(l.stop)
	l.wg.Wait()
	return nil
}

// GetStats returns logger statistics
func (l *AsyncLogger) GetStats() LoggerStats {
	l.stats.mu.RLock()
	defer l.stats.mu.RUnlock()
	return l.stats
}

// SetLevel changes the log level dynamically
func (l *AsyncLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// SetOutput changes the output destination dynamically
func (l *AsyncLogger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// LoggerWithFields is a wrapper with preset fields
type LoggerWithFields struct {
	logger *AsyncLogger
	fields Fields
}

func (l *LoggerWithFields) Debug(message string, extraFields Fields) {
	fields := mergeFields(l.fields, extraFields)
	l.logger.Debug(message, fields)
}

func (l *LoggerWithFields) Info(message string, extraFields Fields) {
	fields := mergeFields(l.fields, extraFields)
	l.logger.Info(message, fields)
}

func (l *LoggerWithFields) Warn(message string, extraFields Fields) {
	fields := mergeFields(l.fields, extraFields)
	l.logger.Warn(message, fields)
}

func (l *LoggerWithFields) Error(message string, extraFields Fields) {
	fields := mergeFields(l.fields, extraFields)
	l.logger.Error(message, fields)
}

func (l *LoggerWithFields) Fatal(message string, extraFields Fields) {
	fields := mergeFields(l.fields, extraFields)
	l.logger.Fatal(message, fields)
}

// Helper functions
func mergeFields(base, extra Fields) Fields {
	if extra == nil {
		return base
	}

	result := make(Fields)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}
