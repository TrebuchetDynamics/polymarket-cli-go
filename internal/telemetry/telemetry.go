package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Logger is a thin structured-log wrapper around log/slog for protocol clients.
type Logger struct {
	slog *slog.Logger
}

// New creates a telemetry logger. If slog is nil, logging is a no-op.
func New(logger *slog.Logger) *Logger {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Logger{slog: logger}
}

// Request logs a completed HTTP request with method, path, status, duration,
// and optional error.
func (l *Logger) Request(ctx context.Context, method, path string, status int, dur time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Duration("dur", dur),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.slog.LogAttrs(ctx, slog.LevelError, "request failed", attrs...)
		return
	}
	level := slog.LevelInfo
	if status >= 500 {
		level = slog.LevelError
	} else if status >= 400 {
		level = slog.LevelWarn
	}
	l.slog.LogAttrs(ctx, level, "request", attrs...)
}

// Retry logs a retry attempt with attempt number and the error that triggered it.
func (l *Logger) Retry(ctx context.Context, method, path string, attempt int, err error) {
	l.slog.LogAttrs(ctx, slog.LevelWarn, "retry",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("attempt", attempt),
		slog.String("error", err.Error()),
	)
}

// RateLimited logs a rate-limit wait event.
func (l *Logger) RateLimited(ctx context.Context, method, path string, wait time.Duration) {
	l.slog.LogAttrs(ctx, slog.LevelWarn, "rate limited",
		slog.String("method", method),
		slog.String("path", path),
		slog.Duration("wait", wait),
	)
}

// CircuitOpen logs a circuit breaker open event.
func (l *Logger) CircuitOpen(ctx context.Context, method, path string) {
	l.slog.LogAttrs(ctx, slog.LevelError, "circuit breaker open",
		slog.String("method", method),
		slog.String("path", path),
	)
}

// RedactableValue wraps a value that may be redacted in logs.
type RedactableValue string

func (v RedactableValue) LogValue() slog.Value {
	if len(v) <= 8 {
		return slog.StringValue("[REDACTED]")
	}
	return slog.StringValue(fmt.Sprintf("%s...%s", v[:4], v[len(v)-4:]))
}
