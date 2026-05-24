package logger

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestAsyncLogger_Basic(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{
		BufferSize: 10,
		Output:     &buf,
		TimeFormat: "2006-01-02 15:04:05",
		Level:      DebugLevel,
	})

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Debug("Debug message", nil)
	logger.Info("Info message", nil)
	logger.Warn("Warn message", nil)
	logger.Error("Error message", nil)

	// Give time for async processing
	time.Sleep(100 * time.Millisecond)

	output := buf.String()
	if output == "" {
		t.Error("Expected log output, got empty")
	}

	// Check if all levels are present
	if !strings.Contains(output, "DEBUG") {
		t.Error("Expected DEBUG level in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level in output")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Expected WARN level in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected ERROR level in output")
	}
}

func TestAsyncLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer

	logger, _ := New(Config{
		BufferSize: 10,
		Output:     &buf,
	})
	defer logger.Close()

	userLogger := logger.WithFields(Fields{
		"user_id": "123",
		"role":    "admin",
	})

	userLogger.Info("User logged in", Fields{
		"ip": "192.168.1.1",
	})

	time.Sleep(100 * time.Millisecond)

	output := buf.String()

	// Check if fields are present
	if !strings.Contains(output, "user_id=123") {
		t.Error("Expected user_id field in output")
	}
	if !strings.Contains(output, "role=admin") {
		t.Error("Expected role field in output")
	}
	if !strings.Contains(output, "ip=192.168.1.1") {
		t.Error("Expected ip field in output")
	}
}

func TestAsyncLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	logger, _ := New(Config{
		BufferSize: 10,
		Output:     &buf,
		Level:      WarnLevel, // Only WARN and above
	})
	defer logger.Close()

	logger.Debug("Debug message", nil)
	logger.Info("Info message", nil)
	logger.Warn("Warn message", nil)
	logger.Error("Error message", nil)

	time.Sleep(100 * time.Millisecond)

	output := buf.String()

	// Debug and Info should be filtered out
	if strings.Contains(output, "DEBUG") {
		t.Error("DEBUG message should be filtered out")
	}
	if strings.Contains(output, "INFO") {
		t.Error("INFO message should be filtered out")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("WARN message should be present")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("ERROR message should be present")
	}
}

func TestAsyncLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer

	logger, _ := New(Config{
		BufferSize: 10,
		Output:     &buf,
		JSONFormat: true,
	})
	defer logger.Close()

	logger.Info("Test message", Fields{
		"key":    "value",
		"number": 42,
	})

	time.Sleep(100 * time.Millisecond)

	output := buf.String()

	// Check if output is valid JSON
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Error("Expected JSON format output")
	}
	if !strings.Contains(output, `"message":"Test message"`) {
		t.Error("Expected message field in JSON")
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Error("Expected key field in JSON")
	}
}

func TestAsyncLogger_IncludeCaller(t *testing.T) {
	var buf bytes.Buffer

	logger, _ := New(Config{
		BufferSize:    10,
		Output:        &buf,
		IncludeCaller: true,
	})
	defer logger.Close()

	logger.Info("Test with caller", nil)

	time.Sleep(100 * time.Millisecond)

	output := buf.String()

	// Check if caller info is present (should contain .go file)
	if !strings.Contains(output, ".go:") {
		t.Errorf("Expected caller info in output, got: %s", output)
	}
}

func TestAsyncLogger_Stats(t *testing.T) {
	var buf bytes.Buffer

	logger, _ := New(Config{
		BufferSize: 1, // Small buffer to test dropping
		Output:     &buf,
	})
	defer logger.Close()

	// Send many messages quickly
	for i := 0; i < 100; i++ {
		logger.Info("Test message", Fields{"index": i})
	}

	time.Sleep(100 * time.Millisecond)

	stats := logger.GetStats()

	if stats.TotalEntries == 0 {
		t.Error("Expected some total entries")
	}

	// Some entries might be dropped due to small buffer
	t.Logf("Stats: Total=%d, Dropped=%d, Errors=%d",
		stats.TotalEntries, stats.DroppedEntries, stats.Errors)
}

func TestAsyncLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer

	logger, _ := New(Config{
		BufferSize: 10,
		Output:     &buf,
		Level:      DebugLevel,
	})
	defer logger.Close()

	logger.Debug("First debug", nil)

	// Change level to Error
	logger.SetLevel(ErrorLevel)

	logger.Debug("Second debug", nil)
	logger.Info("Info message", nil)
	logger.Error("Error message", nil)

	time.Sleep(100 * time.Millisecond)

	output := buf.String()

	// First debug should be present
	if !strings.Contains(output, "First debug") {
		t.Error("First debug message should be present")
	}

	// Second debug should be filtered
	if strings.Contains(output, "Second debug") {
		t.Error("Second debug message should be filtered out")
	}

	// Info should be filtered
	if strings.Contains(output, "Info message") {
		t.Error("Info message should be filtered out")
	}

	// Error should be present
	if !strings.Contains(output, "Error message") {
		t.Error("Error message should be present")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"DEBUG", DebugLevel},
		{"INFO", InfoLevel},
		{"WARN", WarnLevel},
		{"WARNING", WarnLevel},
		{"ERROR", ErrorLevel},
		{"FATAL", FatalLevel},
		{"UNKNOWN", InfoLevel},
	}

	for _, tt := range tests {
		result := ParseLevel(tt.input)
		if result != tt.expected {
			t.Errorf("ParseLevel(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func BenchmarkAsyncLogger(b *testing.B) {
	var buf bytes.Buffer

	logger, _ := New(Config{
		BufferSize: 10000,
		Output:     &buf,
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message", Fields{
			"iteration": i,
			"test":      "value",
		})
	}

	// Wait for all messages to be processed
	time.Sleep(100 * time.Millisecond)
}

func BenchmarkSyncLogger(b *testing.B) {
	var buf bytes.Buffer

	// Create a synchronous logger with small buffer
	logger, _ := New(Config{
		BufferSize: 1,
		Output:     &buf,
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message", Fields{
			"iteration": i,
			"test":      "value",
		})
	}

	time.Sleep(100 * time.Millisecond)
}
