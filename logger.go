package sloggergo

import (
	"context"
	"log/slog"
	"maps"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/godeh/sloggergo/formatter"
	"github.com/godeh/sloggergo/sink"
)

// Level represents the severity of a log entry.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string into a Level.
func ParseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return DebugLevel
	case "info", "INFO":
		return InfoLevel
	case "warn", "WARN", "warning", "WARNING":
		return WarnLevel
	case "error", "ERROR":
		return ErrorLevel
	case "fatal", "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Logger is the main logging interface.
type Logger struct {
	mu           sync.RWMutex
	level        Level
	sinks        []sink.Sink
	fields       map[string]any
	addCaller    bool
	timeFormat   string
	errorHandler ErrorHandler

	// Context extraction
	extractor ContextExtractor

	// Hooks
	hooks []Hook
}

// ContextExtractor extracts attributes from a context.
type ContextExtractor func(ctx context.Context) []slog.Attr

// Hook is a function that can intercept and modify log entries.
// It returns an error if the entry should be dropped or if an error occurred.
type Hook func(ctx context.Context, entry *formatter.Entry) error

// ErrorHandler is a function that handles errors from sinks.
type ErrorHandler func(error)

// Option configures a Logger.
type Option func(*Logger)

// WithLevel sets the minimum log level.
func WithLevel(level Level) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// WithSink adds a sink to the logger.
func WithSink(s sink.Sink) Option {
	return func(l *Logger) {
		l.sinks = append(l.sinks, s)
	}
}

// WithFields adds default fields to all log entries.
func WithFields(fields map[string]any) Option {
	return func(l *Logger) {
		l.fields = fields
	}
}

// WithCaller enables caller information in log entries.
func WithCaller(enabled bool) Option {
	return func(l *Logger) {
		l.addCaller = enabled
	}
}

// WithTimeFormat sets the time format for log entries.
func WithTimeFormat(format string) Option {
	return func(l *Logger) {
		l.timeFormat = format
	}
}

// WithContextExtractor sets a context extractor.
func WithContextExtractor(extractor ContextExtractor) Option {
	return func(l *Logger) {
		l.extractor = extractor
	}
}

// WithHook adds a hook to the logger.
func WithHook(hook Hook) Option {
	return func(l *Logger) {
		l.hooks = append(l.hooks, hook)
	}
}

// WithErrorHandler sets the error handler for the logger.
func WithErrorHandler(handler ErrorHandler) Option {
	return func(l *Logger) {
		l.errorHandler = handler
	}
}

// New creates a new logger with the given options.
func New(opts ...Option) *Logger {
	l := &Logger{
		level:      InfoLevel,
		sinks:      make([]sink.Sink, 0),
		fields:     make(map[string]any),
		addCaller:  true,
		timeFormat: time.RFC3339Nano,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Shutdown logs application shutdown
func (l *Logger) Shutdown(reason string) {
	l.Info("Application shutting down",
		slog.String("reason", reason),
	)
}

// With returns a new logger with additional fields.
func (l *Logger) With(keyvals ...any) *Logger {
	fields := make(map[string]any)

	l.mu.RLock()
	maps.Copy(fields, l.fields)
	l.mu.RUnlock()

	for i := 0; i < len(keyvals)-1; i += 2 {
		if key, ok := keyvals[i].(string); ok {
			fields[key] = keyvals[i+1]
		}
	}

	return &Logger{
		level:      l.level,
		sinks:      l.sinks,
		fields:     fields,
		addCaller:  l.addCaller,
		timeFormat: l.timeFormat,
	}
}

// SetLevel changes the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// AddSink adds a new sink to the logger.
func (l *Logger) AddSink(s sink.Sink) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sinks = append(l.sinks, s)
}

// Close closes all sinks.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var lastErr error
	for _, s := range l.sinks {
		if err := s.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// log is the internal logging method.
func (l *Logger) log(ctx context.Context, level Level, msg string, keyvals ...slog.Attr) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	sinks := l.sinks
	timeFormat := l.timeFormat
	l.mu.RUnlock()

	// Add context attributes if valid context and extractor is set
	if ctx != nil && l.extractor != nil {
		ctxAttrs := l.extractor(ctx)
		if len(ctxAttrs) > 0 {
			// Prepend context attributes to avoid overriding explicit keyvals?
			// Or append? Usually specific overrides general.
			// Let's append context attrs then keyvals.
			newKeyvals := make([]slog.Attr, 0, len(ctxAttrs)+len(keyvals))
			newKeyvals = append(newKeyvals, ctxAttrs...)
			newKeyvals = append(newKeyvals, keyvals...)
			keyvals = newKeyvals
		}
	}

	// Merge logger-level fields with call-site fields
	fields := make(map[string]any)
	l.mu.RLock()
	maps.Copy(fields, l.fields)
	l.mu.RUnlock()

	for _, val := range keyvals {
		fields[val.Key] = val.Value.Any()
	}

	// Get caller
	caller := ""
	if l.addCaller {
		caller = getCaller(3)
	}

	// Create formatter entry
	entry := &formatter.Entry{
		Time:    time.Now().Format(timeFormat),
		Level:   level.String(),
		Message: msg,
		Fields:  fields,
		Caller:  caller,
		Context: ctx,
	}

	// Run hooks
	for _, hook := range l.hooks {
		if err := hook(ctx, entry); err != nil {
			// Hook returned error/drop signal.
			// We stop processing this entry.
			return
		}
	}

	for _, s := range sinks {
		if err := s.Write(entry); err != nil {
			if l.errorHandler != nil {
				l.errorHandler(err)
			}
		}
	}

	if level == FatalLevel {
		os.Exit(1)
	}
}

func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}

	return short + ":" + itoa(line)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, keyvals ...slog.Attr) {
	l.log(context.Background(), DebugLevel, msg, keyvals...)
}

// DebugContext logs a debug message with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	l.log(ctx, DebugLevel, msg, keyvals...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, keyvals ...slog.Attr) {
	l.log(context.Background(), InfoLevel, msg, keyvals...)
}

// InfoContext logs an info message with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	l.log(ctx, InfoLevel, msg, keyvals...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, keyvals ...slog.Attr) {
	l.log(context.Background(), WarnLevel, msg, keyvals...)
}

// WarnContext logs a warning message with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	l.log(ctx, WarnLevel, msg, keyvals...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, keyvals ...slog.Attr) {
	l.log(context.Background(), ErrorLevel, msg, keyvals...)
}

// ErrorContext logs an error message with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	l.log(ctx, ErrorLevel, msg, keyvals...)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string, keyvals ...slog.Attr) {
	l.log(context.Background(), FatalLevel, msg, keyvals...)
}

// FatalContext logs a fatal message with context.
func (l *Logger) FatalContext(ctx context.Context, msg string, keyvals ...slog.Attr) {
	l.log(ctx, FatalLevel, msg, keyvals...)
}
