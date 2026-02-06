package sink

import (
	"io"
	"os"
	"sync"

	"github.com/godeh/sloggergo/formatter"
)

// Sink defines the interface for log output destinations.
type Sink interface {
	Write(entry *formatter.Entry) error
	Close() error
}

// StdoutSink writes log entries to stdout.
type StdoutSink struct {
	mu        sync.Mutex
	writer    io.Writer
	formatter formatter.Formatter
}

// StdoutOption configures a StdoutSink.
type StdoutOption func(*StdoutSink)

// WithWriter sets a custom writer (useful for testing).
func WithWriter(w io.Writer) StdoutOption {
	return func(s *StdoutSink) {
		s.writer = w
	}
}

// WithFormatter sets the formatter for the sink.
func WithFormatter(f formatter.Formatter) StdoutOption {
	return func(s *StdoutSink) {
		s.formatter = f
	}
}

// NewStdout creates a new stdout sink.
func NewStdout(opts ...StdoutOption) *StdoutSink {
	s := &StdoutSink{
		writer:    os.Stdout,
		formatter: formatter.NewText(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Write writes the entry to stdout.
func (s *StdoutSink) Write(entry *formatter.Entry) error {
	data, err := s.formatter.Format(entry)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.writer.Write(data)
	return err
}

// Close is a no-op for stdout.
func (s *StdoutSink) Close() error {
	return nil
}
