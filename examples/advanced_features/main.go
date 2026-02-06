package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/godeh/sloggergo"
	"github.com/godeh/sloggergo/formatter"
	"github.com/godeh/sloggergo/sink"
)

// Mock Tracing Context
type contextKey string

const traceIDKey contextKey = "trace_id"

// Sanitization Hook (PII Masking)
func piiSanitizerHook(ctx context.Context, entry *formatter.Entry) error {
	if email, ok := entry.Fields["email"]; ok {
		if emailStr, ok := email.(string); ok {
			parts := strings.Split(emailStr, "@")
			if len(parts) == 2 {
				entry.Fields["email"] = parts[0][:1] + "***@" + parts[1]
			}
		}
	}
	return nil
}

// Error Dropping Hook (Drop logs containing "ignore_me")
func errorDropperHook(ctx context.Context, entry *formatter.Entry) error {
	if strings.Contains(entry.Message, "ignore_me") {
		return errors.New("log dropped by hook")
	}
	return nil
}

// Context Extractor
func traceExtractor(ctx context.Context) []slog.Attr {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return []slog.Attr{
			slog.String("trace_id", traceID),
		}
	}
	return nil
}

func main() {
	// 1. Create Pretty Printer
	prettyFormatter := formatter.NewTextPretty()

	stdoutSink := sink.NewStdout(sink.WithFormatter(prettyFormatter))

	// 2. Initialize Logger with Features
	log := sloggergo.New(
		sloggergo.WithSink(stdoutSink),
		sloggergo.WithLevel(sloggergo.InfoLevel),
		sloggergo.WithContextExtractor(traceExtractor),
		sloggergo.WithHook(piiSanitizerHook),
		sloggergo.WithHook(errorDropperHook),
	)
	defer log.Close()

	ctx := context.WithValue(context.Background(), traceIDKey, "abc-123-xyz")

	fmt.Println("=== 1. Context Propagation ===")
	// Should show trace_id=abc-123-xyz
	log.InfoContext(ctx, "Processing request")

	fmt.Println("\n=== 2. Pretty Printing & Structs ===")
	// Should show pretty printed struct
	user := map[string]any{
		"id":   101,
		"role": "admin",
		"meta": map[string]string{"region": "us-east"},
	}
	log.Info("User details", slog.Any("user", user))

	fmt.Println("\n=== 3. Hooks: PII Sanitization ===")
	// Should mask email
	log.Info("User login", slog.String("email", "john.doe@example.com"))

	fmt.Println("\n=== 4. Hooks: Dropping Logs ===")
	// Should NOT be printed
	log.Info("This message contains ignore_me and should contain it")
}
