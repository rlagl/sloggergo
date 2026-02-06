package main

import (
	"fmt"
	"os"

	"github.com/godeh/sloggergo"
	"github.com/godeh/sloggergo/sink"
)

func main() {
	logFile := "app.log"

	// Ensure cleanup
	defer os.Remove(logFile)

	// Create file sink
	fileSink, err := sink.NewFile(logFile)
	if err != nil {
		panic(err)
	}

	// Create logger with multiple sinks (file + stdout)
	log := sloggergo.New(
		sloggergo.WithLevel(sloggergo.InfoLevel),
		sloggergo.WithSink(fileSink),
		sloggergo.WithSink(sink.NewStdout()),
	)
	defer log.Close()

	log.Info("Logging to both file and stdout")
	log.Warn("This is a warning")

	fmt.Printf("Check %s for log output\n", logFile)
}
