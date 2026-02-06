package sink

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/godeh/sloggergo/formatter"
)

// FileSink writes log entries to a file.
type FileSink struct {
	mu        sync.Mutex
	file      *os.File
	path      string
	formatter formatter.Formatter
}

// FileOption configures a FileSink.
type FileOption func(*FileSink)

// WithFileFormatter sets the formatter for the file sink.
func WithFileFormatter(f formatter.Formatter) FileOption {
	return func(s *FileSink) {
		s.formatter = f
	}
}

// NewFile creates a new file sink.
func NewFile(path string, opts ...FileOption) (*FileSink, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	// Open file for append
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	s := &FileSink{
		file:      file,
		path:      path,
		formatter: formatter.NewTextNoColor(), // No colors for files
	}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

// Write writes the entry to the file.
func (s *FileSink) Write(entry *formatter.Entry) error {
	data, err := s.formatter.Format(entry)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.file.Write(data)
	return err
}

// Close closes the file.
func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
