package sink

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/godeh/sloggergo/formatter"
)

func TestFileSinkRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	s, err := NewFile(path, WithMaxSizeMB(1), WithMaxBackups(2))
	if err != nil {
		t.Fatalf("NewFile() returned error: %v", err)
	}
	defer func() { _ = s.Close() }()

	first := &formatter.Entry{
		Time:    "",
		Level:   "INFO",
		Message: string(bytes.Repeat([]byte("a"), 900*1024)),
	}
	if err := s.Write(first); err != nil {
		t.Fatalf("Write(first) returned error: %v", err)
	}

	second := &formatter.Entry{
		Time:    "",
		Level:   "INFO",
		Message: string(bytes.Repeat([]byte("b"), 200*1024)),
	}
	if err := s.Write(second); err != nil {
		t.Fatalf("Write(second) returned error: %v", err)
	}

	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("expected rotated file to exist: %v", err)
	}
}

func TestFileSinkRotationNoBackups(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	s, err := NewFile(path, WithMaxSizeMB(1))
	if err != nil {
		t.Fatalf("NewFile() returned error: %v", err)
	}
	defer func() { _ = s.Close() }()

	first := &formatter.Entry{
		Time:    "",
		Level:   "INFO",
		Message: string(bytes.Repeat([]byte("a"), 900*1024)),
	}
	if err := s.Write(first); err != nil {
		t.Fatalf("Write(first) returned error: %v", err)
	}

	second := &formatter.Entry{
		Time:    "",
		Level:   "INFO",
		Message: string(bytes.Repeat([]byte("b"), 200*1024)),
	}
	if err := s.Write(second); err != nil {
		t.Fatalf("Write(second) returned error: %v", err)
	}

	if _, err := os.Stat(path + ".1"); err == nil {
		t.Fatalf("did not expect rotated file to be kept when max_backups=0")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected active log file to exist: %v", err)
	}
}
