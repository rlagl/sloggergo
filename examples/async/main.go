package main

import (
	"log/slog"
	"time"

	"github.com/godeh/sloggergo"
	"github.com/godeh/sloggergo/sink"
)

func main() {
	baseLogger := sloggergo.New(
		sloggergo.WithLevel(sloggergo.InfoLevel),
		sloggergo.WithSink(sink.NewStdout()),
	)

	// Wrap with AsyncLogger
	// This will not block the main execution thread for log writes
	asyncLog := sloggergo.NewAsync(
		baseLogger,
		sloggergo.WithBufferSize(1000),
		sloggergo.WithWorkers(2),
	)
	// Important: flush and close to ensure all logs are written before exit
	defer asyncLog.Close()

	asyncLog.Info("Starting high throughput logging...")

	start := time.Now()
	for i := range 100 {
		asyncLog.Info("Processing item", slog.Int("index", i))
	}

	asyncLog.Info("Finished processing", slog.Duration("duration", time.Since(start)))

	// Force flush just to be sure (Close also flushes)
	asyncLog.Flush()
}
