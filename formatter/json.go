package formatter

import (
	"encoding/json"
)

// JSONFormatter formats log entries as JSON.
type JSONFormatter struct {
	// PrettyPrint enables indented JSON output.
	PrettyPrint bool
}

// jsonEntry is the JSON representation of a log entry.
type jsonEntry struct {
	Time    string         `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Caller  string         `json:"caller,omitempty"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// Format formats the entry as JSON.
func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	je := jsonEntry{
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
		Caller:  entry.Caller,
		Fields:  entry.Fields,
	}

	if f.PrettyPrint {
		data, err := json.MarshalIndent(je, "", "  ")
		if err != nil {
			return nil, err
		}
		return append(data, '\n'), nil
	}
	data, err := json.Marshal(je)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// NewJSON creates a new JSON formatter.
func NewJSON() *JSONFormatter {
	return &JSONFormatter{}
}

// NewJSONPretty creates a new JSON formatter with pretty printing.
func NewJSONPretty() *JSONFormatter {
	return &JSONFormatter{PrettyPrint: true}
}
