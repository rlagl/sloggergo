package sink

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/godeh/sloggergo/formatter"
)

// FileSink writes log entries to a file.
type FileSink struct {
	mu         sync.Mutex
	file       *os.File
	path       string
	formatter  formatter.Formatter
	size       int64
	maxSize    int64
	maxBackups int
}

// FileOption configures a FileSink.
type FileOption func(*FileSink)

// WithFileFormatter sets the formatter for the file sink.
func WithFileFormatter(f formatter.Formatter) FileOption {
	return func(s *FileSink) {
		s.formatter = f
	}
}

// WithMaxSizeMB sets the maximum file size in megabytes before rotation.
func WithMaxSizeMB(size int) FileOption {
	return func(s *FileSink) {
		if size > 0 {
			s.maxSize = int64(size) * 1024 * 1024
		}
	}
}

// WithMaxBackups sets the number of rotated files to keep.
func WithMaxBackups(n int) FileOption {
	return func(s *FileSink) {
		if n > 0 {
			s.maxBackups = n
		}
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

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	s := &FileSink{
		file:      file,
		path:      path,
		formatter: formatter.NewTextNoColor(), // No colors for files
		size:      info.Size(),
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

	if err := s.rotateIfNeededLocked(len(data)); err != nil {
		return err
	}

	_, err = s.file.Write(data)
	if err == nil {
		s.size += int64(len(data))
	}
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

func (s *FileSink) rotateIfNeededLocked(nextLen int) error {
	if s.maxSize <= 0 {
		return nil
	}

	if s.size+int64(nextLen) <= s.maxSize {
		return nil
	}

	if s.file != nil {
		if err := s.file.Close(); err != nil {
			return err
		}
	}

	if s.maxBackups > 0 {
		for i := s.maxBackups - 1; i >= 1; i-- {
			oldPath := s.path + "." + strconv.Itoa(i)
			newPath := s.path + "." + strconv.Itoa(i+1)
			_ = os.Rename(oldPath, newPath)
		}
		_ = os.Rename(s.path, s.path+".1")
	} else {
		_ = os.Remove(s.path)
	}

	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	s.file = file
	s.size = 0
	return nil
}
