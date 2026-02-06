package main

import (
	"log/slog"

	"github.com/godeh/sloggergo"
	"github.com/godeh/sloggergo/formatter"
	"github.com/godeh/sloggergo/sink"
)

func main() {
	// Create JSON formatter (pretty print for demo purposes)
	jsonFmt := formatter.NewJSONPretty()

	// Create stdout sink with JSON formatter
	jsonSink := sink.NewStdout(sink.WithFormatter(jsonFmt))

	log := sloggergo.New(
		sloggergo.WithLevel(sloggergo.InfoLevel),
		sloggergo.WithSink(jsonSink),
	)
	defer log.Close()

	// Structured logs will be formatted as JSON objects
	log.Info("User logged in",
		slog.Int("user_id", 42),
		slog.String("username", "gopher"),
		slog.Any("roles", []string{"admin", "editor"}),
	)
}
