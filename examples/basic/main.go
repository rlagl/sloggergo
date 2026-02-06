package main

import (
	"log/slog"
	"time"

	"github.com/godeh/sloggergo"
	"github.com/godeh/sloggergo/sink"
)

func main() {
	// Create a simple logger writing to stdout
	log := sloggergo.New(
		sloggergo.WithLevel(sloggergo.DebugLevel),
		sloggergo.WithSink(sink.NewStdout()),
	)
	defer log.Close()

	log.Info("Starting application...")

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	log.Debug("Processing request", slog.Int("id", 123), slog.String("ip", "192.168.1.1"))
	log.Warn("High memory usage detected", slog.Int("usage_mb", 512))
	log.Error("Database connection failed", slog.Int("retry", 3))

	log.Info("Application finished")
}
