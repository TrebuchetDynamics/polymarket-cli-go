package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler  { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler       { return h }

func TestRequestLogsError(t *testing.T) {
	h := &captureHandler{}
	log := New(slog.New(h))
	ctx := context.Background()
	log.Request(ctx, "GET", "/book", 500, 100*time.Millisecond, errors.New("boom"))
	if len(h.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(h.records))
	}
	r := h.records[0]
	if r.Level != slog.LevelError {
		t.Fatalf("level = %v, want error", r.Level)
	}
	if !strings.Contains(r.Message, "failed") {
		t.Fatalf("message = %s, want 'failed'", r.Message)
	}
}

func TestRequestLogsSuccess(t *testing.T) {
	h := &captureHandler{}
	log := New(slog.New(h))
	ctx := context.Background()
	log.Request(ctx, "GET", "/book", 200, 50*time.Millisecond, nil)
	if len(h.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(h.records))
	}
	r := h.records[0]
	if r.Level != slog.LevelInfo {
		t.Fatalf("level = %v, want info", r.Level)
	}
}

func TestRetryLogsWarning(t *testing.T) {
	h := &captureHandler{}
	log := New(slog.New(h))
	ctx := context.Background()
	log.Retry(ctx, "GET", "/book", 2, errors.New("timeout"))
	if len(h.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(h.records))
	}
	r := h.records[0]
	if r.Level != slog.LevelWarn {
		t.Fatalf("level = %v, want warn", r.Level)
	}
}

func TestNilLoggerIsNoOp(t *testing.T) {
	log := New(nil)
	ctx := context.Background()
	log.Request(ctx, "GET", "/book", 200, 10*time.Millisecond, nil)
	log.Retry(ctx, "GET", "/book", 1, errors.New("timeout"))
	log.RateLimited(ctx, "GET", "/book", 100*time.Millisecond)
	log.CircuitOpen(ctx, "GET", "/book")
}

func TestRedactableValue(t *testing.T) {
	v := RedactableValue("supersecret123")
	got := v.LogValue().String()
	want := "supe...t123"
	if got != want {
		t.Fatalf("redacted = %s, want %s", got, want)
	}
}
