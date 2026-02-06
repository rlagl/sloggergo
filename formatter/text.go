package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// TextFormatter formats log entries as human-readable text.
type TextFormatter struct {
	// DisableColors disables ANSI color output.
	DisableColors bool

	// DisableTimestamp hides the timestamp.
	DisableTimestamp bool

	// DisableCaller hides the caller information.
	DisableCaller bool

	// TimestampFormat is the format for timestamps.
	TimestampFormat string

	// PrettyPrint enables multi-line output for easier reading in development.
	PrettyPrint bool
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// Format formats the entry as text.
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var buf bytes.Buffer

	// Timestamp
	if !f.DisableTimestamp && entry.Time != "" {
		buf.WriteString(f.colorize(colorGray, entry.Time))
		buf.WriteString(" ")
	}

	// Level
	levelColor := f.getLevelColor(entry.Level)
	buf.WriteString(f.colorize(levelColor, fmt.Sprintf("%-5s", entry.Level)))
	buf.WriteString(" ")

	// Caller
	if !f.DisableCaller && entry.Caller != "" {
		buf.WriteString(f.colorize(colorCyan, entry.Caller))
		buf.WriteString(" ")
	}

	// Message
	buf.WriteString(entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		if f.PrettyPrint {
			buf.WriteString("\n")
			// Sort keys for consistent output
			keys := make([]string, 0, len(entry.Fields))
			for k := range entry.Fields {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				v := entry.Fields[k]
				buf.WriteString(strings.Repeat(" ", 4)) // Indent
				buf.WriteString(f.colorize(colorBlue, k))
				buf.WriteString(": ")

				// Pretty print complex values
				jsonBytes, err := json.MarshalIndent(v, strings.Repeat(" ", 4), "  ")
				if err == nil && (strings.HasPrefix(string(jsonBytes), "{") || strings.HasPrefix(string(jsonBytes), "[")) {
					buf.WriteString(string(jsonBytes))
				} else {
					fmt.Fprintf(&buf, "%v", v)
				}
				buf.WriteString("\n")
			}
		} else {
			buf.WriteString(" ")
			first := true
			// To ensure deterministic order for tests/sanity, let's sort keys even in non-pretty mode if it's cheap,
			// or just iterate map. Map iteration is random.
			// Let's stick to map iteration for perf unless sorted is needed?
			// Standard behavior is usually random or use specific marshaler.
			// Standard `text` formatter usually just loops.
			for k, v := range entry.Fields {
				if !first {
					buf.WriteString(" ")
				}
				buf.WriteString(f.colorize(colorBlue, k))
				buf.WriteString("=")
				fmt.Fprintf(&buf, "%v", v)
				first = false
			}
		}
	}

	buf.WriteString("\n")
	return buf.Bytes(), nil
}

func (f *TextFormatter) colorize(color, text string) string {
	if f.DisableColors {
		return text
	}
	return color + text + colorReset
}

func (f *TextFormatter) getLevelColor(level string) string {
	switch level {
	case "DEBUG":
		return colorGray
	case "INFO":
		return colorGreen
	case "WARN":
		return colorYellow
	case "ERROR":
		return colorRed
	case "FATAL":
		return colorPurple
	default:
		return colorReset
	}
}

// NewText creates a new text formatter.
func NewText() *TextFormatter {
	return &TextFormatter{}
}

// NewTextNoColor creates a new text formatter without colors.
func NewTextNoColor() *TextFormatter {
	return &TextFormatter{DisableColors: true}
}

// NewTextPretty creates a new text formatter with pretty printing enabled.
func NewTextPretty() *TextFormatter {
	return &TextFormatter{PrettyPrint: true}
}
