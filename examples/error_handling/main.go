package main

import (
	"errors"
	"fmt"

	"github.com/godeh/sloggergo"
	"github.com/godeh/sloggergo/formatter"
)

// FaultySink simulates a sink that always fails
type FaultySink struct{}

func (s *FaultySink) Write(entry *formatter.Entry) error {
	return errors.New("simulated write error")
}

func (s *FaultySink) Close() error { return nil }

func main() {
	// Define an error handler
	errorHandler := func(err error) {
		fmt.Printf("!!! Logger Error Caught: %v\n", err)
	}

	log := sloggergo.New(
		sloggergo.WithSink(&FaultySink{}),
		sloggergo.WithErrorHandler(errorHandler), // Register handler
	)
	defer log.Close()

	log.Info("This log will fail to write")
}
