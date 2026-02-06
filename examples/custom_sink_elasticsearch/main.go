package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/godeh/sloggergo"
	"github.com/godeh/sloggergo/formatter"
)

// ElasticsearchSink is a custom implementation of the sink.Sink interface
type ElasticsearchSink struct {
	client *http.Client
	url    string
	index  string
}

func NewElasticsearchSink(url, index string) *ElasticsearchSink {
	return &ElasticsearchSink{
		client: &http.Client{Timeout: 5 * time.Second},
		url:    url,
		index:  index,
	}
}

// Write implements sink.Sink
func (s *ElasticsearchSink) Write(entry *formatter.Entry) error {
	// 1. Create the document payload
	doc := map[string]any{
		"@timestamp": entry.Time,
		"level":      entry.Level,
		"message":    entry.Message,
		"fields":     entry.Fields,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	// 2. Send to Elasticsearch (simple synchronous implementation)
	// In production, you would want to use batching/buffering here!
	endpoint := fmt.Sprintf("%s/%s/_doc", s.url, s.index)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// Close implements sink.Sink
func (s *ElasticsearchSink) Close() error {
	s.client.CloseIdleConnections()
	return nil
}

func main() {
	// Instantiate our custom sink
	esSink := NewElasticsearchSink("http://localhost:9200", "app-logs")

	// Use it with sloggergo
	log := sloggergo.New(
		sloggergo.WithLevel(sloggergo.InfoLevel),
		sloggergo.WithSink(esSink),
	)
	defer log.Close()

	fmt.Println("Sending logs to custom Elasticsearch sink...")
	log.Info("Hello from custom sink!", slog.Int("retry", 1))
	log.Warn("This is a warning", slog.Int("user_id", 99))
}
