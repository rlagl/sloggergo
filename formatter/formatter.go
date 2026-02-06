package formatter

import "context"

// Entry represents a log entry for formatting.
// This is defined here to avoid import cycles.
type Entry struct {
	Time    string
	Level   string
	Message string
	Fields  map[string]any
	Caller  string
	Context context.Context `json:"-"`
}

// Formatter defines the interface for formatting log entries.
type Formatter interface {
	// Format formats a log entry into a byte slice.
	Format(entry *Entry) ([]byte, error)
}
