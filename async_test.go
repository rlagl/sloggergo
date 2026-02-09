package sloggergo

import (
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godeh/sloggergo/formatter"
)

func TestAsyncLoggerAppliesHooksAndContext(t *testing.T) {
	mock := &mockSink{}

	type traceKeyType string
	const traceKey traceKeyType = "trace_id"

	extractor := func(ctx context.Context) []slog.Attr {
		if v, ok := ctx.Value(traceKey).(string); ok {
			return []slog.Attr{slog.String("trace_id", v)}
		}
		return nil
	}

	hook := func(_ context.Context, entry *formatter.Entry) error {
		entry.Fields["hooked"] = "yes"
		return nil
	}

	base := New(
		WithSink(mock),
		WithContextExtractor(extractor),
		WithHook(hook),
		WithTimeFormat("2006-01-02"),
	)

	async := NewAsync(base)
	ctx := context.WithValue(context.Background(), traceKey, "abc-123")
	async.InfoContext(ctx, "hello")
	_ = async.Close()

	if mock.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", mock.Len())
	}

	entry := mock.entries[0]
	if entry.Fields["trace_id"] != "abc-123" {
		t.Fatalf("expected trace_id=abc-123, got %v", entry.Fields["trace_id"])
	}
	if entry.Fields["hooked"] != "yes" {
		t.Fatalf("expected hooked=yes, got %v", entry.Fields["hooked"])
	}
	if len(entry.Time) != len("2006-01-02") || strings.Count(entry.Time, "-") != 2 {
		t.Fatalf("expected time format YYYY-MM-DD, got %q", entry.Time)
	}
}

func TestAsyncLoggerCloseDuringLogging(t *testing.T) {
	mock := &mockSink{}
	base := New(WithSink(mock))
	async := NewAsync(base, WithBufferSize(10), WithWorkers(1))

	var panicked atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				panicked.Store(true)
			}
		}()
		for i := 0; i < 1000; i++ {
			async.Info("msg")
		}
	}()

	time.Sleep(1 * time.Millisecond)
	_ = async.Close()
	<-done

	if panicked.Load() {
		t.Fatalf("async logging panicked during Close")
	}
}
