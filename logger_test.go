package sloggergo

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/godeh/sloggergo/formatter"
	"github.com/godeh/sloggergo/sink"
)

// mockSink is a test sink that captures log entries.
type mockSink struct {
	mu      sync.Mutex
	entries []*formatter.Entry
}

func (m *mockSink) Write(entry *formatter.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockSink) Close() error {
	return nil
}

func (m *mockSink) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

func TestLoggerBasic(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock))

	log.Info("hello world")

	if mock.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", mock.Len())
	}

	entry := mock.entries[0]
	if entry.Message != "hello world" {
		t.Errorf("expected message 'hello world', got %q", entry.Message)
	}
	if entry.Level != "INFO" {
		t.Errorf("expected level INFO, got %v", entry.Level)
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(WarnLevel), WithSink(mock))

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	if mock.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", mock.Len())
	}

	if mock.entries[0].Level != "WARN" {
		t.Errorf("expected first entry to be WARN, got %s", mock.entries[0].Level)
	}
	if mock.entries[1].Level != "ERROR" {
		t.Errorf("expected second entry to be ERROR, got %s", mock.entries[1].Level)
	}
}

func TestLoggerWithFields(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock))

	log.Info("message", slog.Int("user_id", 123), slog.String("action", "login"))

	if mock.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", mock.Len())
	}

	fields := mock.entries[0].Fields
	if fields["user_id"] != int64(123) {
		t.Errorf("expected user_id=123, got %v", fields["user_id"])
	}
	if fields["action"] != "login" {
		t.Errorf("expected action=login, got %v", fields["action"])
	}
}

func TestLoggerWith(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock))

	childLog := log.With("request_id", "abc-123")
	childLog.Info("processing")

	if mock.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", mock.Len())
	}

	fields := mock.entries[0].Fields
	if fields["request_id"] != "abc-123" {
		t.Errorf("expected request_id=abc-123, got %v", fields["request_id"])
	}
}

func TestLoggerAllLevels(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock))

	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")

	if mock.Len() != 4 {
		t.Fatalf("expected 4 entries, got %d", mock.Len())
	}

	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for i, expected := range levels {
		if mock.entries[i].Level != expected {
			t.Errorf("entry %d: expected level %v, got %v", i, expected, mock.entries[i].Level)
		}
	}
}

func TestLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"INFO", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"fatal", FatalLevel},
		{"unknown", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := ParseLevel(tt.input)
			if level != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, level, tt.expected)
			}
		})
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("Level.String() = %q, want %q", tt.level.String(), tt.expected)
			}
		})
	}
}

func TestLoggerConcurrency(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				log.Info("message", slog.Int("id", id), slog.Int("j", j))
			}
		}(i)
	}
	wg.Wait()

	if mock.Len() != 1000 {
		t.Errorf("expected 1000 entries, got %d", mock.Len())
	}
}

func TestLoggerClose(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock))

	err := log.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestLoggerSetLevel(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock))

	log.Debug("should appear")
	log.SetLevel(ErrorLevel)
	log.Debug("should not appear")
	log.Error("should appear")

	if mock.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", mock.Len())
	}
}

func TestCallerInfo(t *testing.T) {
	mock := &mockSink{}
	log := New(WithLevel(DebugLevel), WithSink(mock), WithCaller(true))

	log.Info("test message")

	if mock.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", mock.Len())
	}

	caller := mock.entries[0].Caller
	if caller == "" {
		t.Error("expected caller info, got empty string")
	}
	// Just verify we have some caller info (filename:line format)
	if !strings.Contains(caller, ":") {
		t.Errorf("expected caller format with ':', got %q", caller)
	}
}

func TestStdoutSink(t *testing.T) {
	s := sink.NewStdout()
	entry := &formatter.Entry{
		Time:    "2024-01-01T00:00:00Z",
		Level:   "INFO",
		Message: "test",
	}
	err := s.Write(entry)
	if err != nil {
		t.Errorf("Write() returned error: %v", err)
	}
	_ = s.Close()
}

func TestNewFromConfig(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")
	configContent := `{
		"logger": {
			"level": "info",
			"format": "text",
			"stdout": { "enabled": false },
			"file": {
				"enabled": true,
				"path": "` + logPath + `"
			}
		}
	}`
	configFile := filepath.Join(dir, "test_config.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	log, err := NewFromConfig(configFile)
	if err != nil {
		t.Fatalf("NewFromConfig() returned error: %v", err)
	}
	defer func() { _ = log.Close() }()

	log.Info("test from config")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), "test from config") {
		t.Fatalf("expected log file to contain message, got: %s", string(data))
	}
}

func TestNewFromConfigRotation(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")
	configContent := `{
		"logger": {
			"level": "info",
			"format": "text",
			"stdout": { "enabled": false },
			"file": {
				"enabled": true,
				"path": "` + logPath + `",
				"max_size_mb": 1,
				"max_backups": 2
			}
		}
	}`
	configFile := filepath.Join(dir, "test_config.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	log, err := NewFromConfig(configFile)
	if err != nil {
		t.Fatalf("NewFromConfig() returned error: %v", err)
	}
	defer func() { _ = log.Close() }()

	large := strings.Repeat("a", 900*1024)
	log.Info("first", slog.String("payload", large))
	log.Info("second", slog.String("payload", strings.Repeat("b", 200*1024)))

	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Fatalf("expected rotated file to exist: %v", err)
	}
}
