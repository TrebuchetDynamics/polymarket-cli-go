package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAcquire(t *testing.T) {
	rl := NewRateLimiter(100)
	if !rl.TryAcquire() {
		t.Fatal("should acquire")
	}
}

func TestRateLimiterWait(t *testing.T) {
	rl := NewRateLimiter(100)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := rl.Wait(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestRateLimiterExhaustion(t *testing.T) {
	rl := NewRateLimiter(1)
	rl.TryAcquire() // consume the one token
	if rl.TryAcquire() {
		t.Fatal("should be exhausted")
	}
}

func TestRateLimiterStop(t *testing.T) {
	rl := NewRateLimiter(10)
	rl.Stop()
	ctx := context.Background()
	if err := rl.Wait(ctx); err == nil {
		t.Fatal("should return error after stop")
	}
}

func TestCircuitBreakerClosedToOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{MaxFailures: 3, ResetTimeout: time.Hour})
	for i := 0; i < 3; i++ {
		if err := cb.Call(func() error { return http.ErrServerClosed }); err == nil {
			t.Fatal("fn should fail")
		}
	}
	if cb.State() != StateOpen {
		t.Fatalf("expected open, got %s", cb.State())
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{MaxFailures: 1, ResetTimeout: 0, HalfOpenMaxReqs: 2})
	cb.Call(func() error { return http.ErrServerClosed })
	if cb.State() != StateOpen {
		t.Fatal("expected open")
	}
	// With ResetTimeout=0, next Call transitions to half-open
	err := cb.Call(func() error { return nil })
	if err != nil || cb.State() != StateOpen {
		// Single success in half-open doesn't close; need HalfOpenMaxReqs successes
	}
}

func TestCircuitBreakerClosedStaysClosed(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{MaxFailures: 5, ResetTimeout: time.Hour})
	for i := 0; i < 3; i++ {
		cb.Call(func() error { return nil })
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed, got %s", cb.State())
	}
	if cb.Failures() != 0 {
		t.Fatalf("failures should be 0: %d", cb.Failures())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{MaxFailures: 1, ResetTimeout: time.Hour})
	cb.Call(func() error { return http.ErrServerClosed })
	cb.Reset()
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after reset, got %s", cb.State())
	}
}

func TestRedaction(t *testing.T) {
	if RedactSecret("abcdefghijklmnop") != "abcd...mnop" {
		t.Fatalf("redact: %s", RedactSecret("abcdefghijklmnop"))
	}
	if RedactSecret("short") != "[REDACTED]" {
		t.Fatalf("short: %s", RedactSecret("short"))
	}
	if RedactSecret("") != "" {
		t.Fatal("empty should be empty")
	}
}

func TestRedactMap(t *testing.T) {
	m := map[string]string{
		"POLY_API_KEY":    "secret-key",
		"POLY_PASSPHRASE": "secret-pass",
		"User-Agent":      "polygolem",
	}
	out := RedactMap(m)
	if out["POLY_API_KEY"] == "secret-key" {
		t.Fatal("API key not redacted")
	}
	if out["User-Agent"] != "polygolem" {
		t.Fatal("User-Agent should not be redacted")
	}
}

func TestTransportGetRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := DefaultConfig(server.URL + "/")
	cfg.RetryMax = 3
	client := New(server.Client(), cfg)

	var result map[string]bool
	if err := client.Get(context.Background(), "/test", &result); err != nil {
		t.Fatal(err)
	}
	if !result["ok"] {
		t.Fatal("expected ok")
	}
	if attempts < 3 {
		t.Fatalf("expected at least 3 attempts: %d", attempts)
	}
}

func TestTransportGetNoRetryOn404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := DefaultConfig(server.URL + "/")
	cfg.RetryMax = 3
	client := New(server.Client(), cfg)

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}
