package sloggergo

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/godeh/sloggergo/formatter"
)

// AsyncLogger wraps a Logger with async capabilities.
type AsyncLogger struct {
	*Logger
	buffer          chan *formatter.Entry
	wg              sync.WaitGroup
	closed          bool
	closeMu         sync.Mutex
	bufferSize      int
	workers         int
	sampling        *SamplingConfig
	shutdownTimeout time.Duration
}

// SamplingConfig configures log sampling.
type SamplingConfig struct {
	Initial    int           // Log first N entries per interval
	Thereafter int           // Then log every N-th entry
	Interval   time.Duration // Sampling interval
}

// AsyncOption configures an AsyncLogger.
type AsyncOption func(*AsyncLogger)

// WithBufferSize sets the buffer size for async logging.
func WithBufferSize(size int) AsyncOption {
	return func(a *AsyncLogger) {
		a.bufferSize = size
	}
}

// WithWorkers sets the number of async workers.
func WithWorkers(n int) AsyncOption {
	return func(a *AsyncLogger) {
		a.workers = n
	}
}

// WithSampling enables log sampling.
func WithSampling(config *SamplingConfig) AsyncOption {
	return func(a *AsyncLogger) {
		a.sampling = config
	}
}

// WithShutdownTimeout sets the timeout for graceful shutdown.
func WithShutdownTimeout(d time.Duration) AsyncOption {
	return func(a *AsyncLogger) {
		a.shutdownTimeout = d
	}
}

// NewAsync creates a new async logger.
func NewAsync(logger *Logger, opts ...AsyncOption) *AsyncLogger {
	a := &AsyncLogger{
		Logger:          logger,
		bufferSize:      1000,
		workers:         2,
		shutdownTimeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(a)
	}

	a.buffer = make(chan *formatter.Entry, a.bufferSize)

	// Start workers
	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)
		go a.worker()
	}

	return a
}

func (a *AsyncLogger) worker() {
	defer a.wg.Done()

	for entry := range a.buffer {
		a.Logger.mu.RLock()
		sinks := a.Logger.sinks
		errorHandler := a.Logger.errorHandler
		a.Logger.mu.RUnlock()

		for _, s := range sinks {
			if err := s.Write(entry); err != nil {
				if errorHandler != nil {
					errorHandler(err)
				}
			}
		}
	}
}

// logAsync sends log entry to buffer without blocking.
func (a *AsyncLogger) logAsync(ctx context.Context, level Level, msg string, keyvals ...slog.Attr) {
	a.Logger.mu.RLock()
	if level < a.Logger.level {
		a.Logger.mu.RUnlock()
		return
	}
	a.Logger.mu.RUnlock()

	// Build entry
	fields := make(map[string]any)
	a.Logger.mu.RLock()
	for k, v := range a.Logger.fields {
		fields[k] = v
	}
	a.Logger.mu.RUnlock()

	for _, val := range keyvals {
		fields[val.Key] = val.Value.Any()
	}

	caller := ""
	if a.Logger.addCaller {
		caller = getCaller(3)
	}

	entry := &formatter.Entry{
		Time:    time.Now().Format(time.RFC3339Nano),
		Level:   level.String(),
		Message: msg,
		Fields:  fields,
		Caller:  caller,
		Context: ctx,
	}

	// Non-blocking send
	select {
	case a.buffer <- entry:
	default:
		// Buffer full, drop log (or could count dropped)
	}
}

// Debug logs a debug message asynchronously.
func (a *AsyncLogger) Debug(msg string, keyvals ...slog.Attr) {
	a.logAsync(context.Background(), DebugLevel, msg, keyvals...)
}

// DebugContext logs a debug message asynchronously with context.
func (a *AsyncLogger) DebugContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	a.logAsync(ctx, DebugLevel, msg, keyvals...)
}

// Info logs an info message asynchronously.
func (a *AsyncLogger) Info(msg string, keyvals ...slog.Attr) {
	a.logAsync(context.Background(), InfoLevel, msg, keyvals...)
}

// InfoContext logs an info message asynchronously with context.
func (a *AsyncLogger) InfoContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	a.logAsync(ctx, InfoLevel, msg, keyvals...)
}

// Warn logs a warning message asynchronously.
func (a *AsyncLogger) Warn(msg string, keyvals ...slog.Attr) {
	a.logAsync(context.Background(), WarnLevel, msg, keyvals...)
}

// WarnContext logs a warning message asynchronously with context.
func (a *AsyncLogger) WarnContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	a.logAsync(ctx, WarnLevel, msg, keyvals...)
}

// Error logs an error message asynchronously.
func (a *AsyncLogger) Error(msg string, keyvals ...slog.Attr) {
	a.logAsync(context.Background(), ErrorLevel, msg, keyvals...)
}

// ErrorContext logs an error message asynchronously with context.
func (a *AsyncLogger) ErrorContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	a.logAsync(ctx, ErrorLevel, msg, keyvals...)
}

// Fatal logs a fatal message (runs synchronously for safety).
func (a *AsyncLogger) Fatal(msg string, keyvals ...slog.Attr) {
	// Fatal runs synchronously to ensure it's written
	a.Logger.Fatal(msg, keyvals...)
}

// Flush waits for all buffered logs to be written.
func (a *AsyncLogger) Flush() {
	// Wait for buffer to drain
	for len(a.buffer) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}

// Close flushes and closes the async logger.
func (a *AsyncLogger) Close() error {
	a.closeMu.Lock()
	if a.closed {
		a.closeMu.Unlock()
		return nil
	}
	a.closed = true
	a.closeMu.Unlock()

	close(a.buffer)

	// Wait for workers with timeout
	c := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		// Workers finished
	case <-time.After(a.shutdownTimeout):
		// Timeout
	}

	return a.Logger.Close()
}

// BufferLen returns the current buffer length.
func (a *AsyncLogger) BufferLen() int {
	return len(a.buffer)
}

// IsFull returns true if the buffer is full.
func (a *AsyncLogger) IsFull() bool {
	return len(a.buffer) >= a.bufferSize
}

// --- Sampled Logger ---

// SampledLogger wraps a logger with sampling.
type SampledLogger struct {
	*Logger
	config  *SamplingConfig
	counts  map[string]*sampleCounter
	countMu sync.Mutex
}

type sampleCounter struct {
	count     int
	resetTime time.Time
}

// NewSampled creates a logger with sampling.
func NewSampled(logger *Logger, config *SamplingConfig) *SampledLogger {
	return &SampledLogger{
		Logger: logger,
		config: config,
		counts: make(map[string]*sampleCounter),
	}
}

func (s *SampledLogger) shouldLog(key string) bool {
	s.countMu.Lock()
	defer s.countMu.Unlock()

	now := time.Now()
	counter, exists := s.counts[key]

	if !exists || now.After(counter.resetTime) {
		s.counts[key] = &sampleCounter{
			count:     1,
			resetTime: now.Add(s.config.Interval),
		}
		return true
	}

	counter.count++

	// Log first N entries
	if counter.count <= s.config.Initial {
		return true
	}

	// Then log every N-th entry
	if s.config.Thereafter > 0 && (counter.count-s.config.Initial)%s.config.Thereafter == 0 {
		return true
	}

	return false
}

// Info logs with sampling.
func (s *SampledLogger) Info(msg string, keyvals ...slog.Attr) {
	if s.shouldLog(msg) {
		s.Logger.Info(msg, keyvals...)
	}
}

// Warn logs with sampling.
func (s *SampledLogger) Warn(msg string, keyvals ...slog.Attr) {
	if s.shouldLog(msg) {
		s.Logger.Warn(msg, keyvals...)
	}
}

// Debug logs with sampling.
func (s *SampledLogger) Debug(msg string, keyvals ...slog.Attr) {
	if s.shouldLog(msg) {
		s.Logger.Debug(msg, keyvals...)
	}
}

// Error always logs (no sampling for errors).
func (s *SampledLogger) Error(msg string, keyvals ...slog.Attr) {
	s.Logger.Error(msg, keyvals...)
}

// Fatal always logs (no sampling for fatal).
func (s *SampledLogger) Fatal(msg string, keyvals ...slog.Attr) {
	s.Logger.Fatal(msg, keyvals...)
}
