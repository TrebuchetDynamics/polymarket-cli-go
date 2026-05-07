package transport

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
// Stolen from polymarket-go-sdk/pkg/transport/ratelimit.go.
type RateLimiter struct {
	mu           sync.Mutex
	capacity     int
	tokensPerSec float64
	tokens       float64
	lastRefill   time.Time
	stopped      bool
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10
	}
	return &RateLimiter{
		capacity:     requestsPerSecond,
		tokensPerSec: float64(requestsPerSecond),
		tokens:       float64(requestsPerSecond),
		lastRefill:   time.Now(),
	}
}

func (rl *RateLimiter) Start() {}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		if rl.stopped {
			rl.mu.Unlock()
			return context.Canceled
		}
		rl.refillTokens()
		if rl.tokens >= 1.0 {
			rl.tokens -= 1.0
			rl.mu.Unlock()
			return nil
		}
		tokensNeeded := 1.0 - rl.tokens
		waitDuration := time.Duration(float64(time.Second) * tokensNeeded / rl.tokensPerSec)
		rl.mu.Unlock()

		timer := time.NewTimer(waitDuration)
		select {
		case <-timer.C:
			timer.Stop()
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

func (rl *RateLimiter) TryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refillTokens()
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}
	return false
}

func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.stopped = true
}

func (rl *RateLimiter) refillTokens() {
	if rl.stopped {
		return
	}
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := elapsed.Seconds() * rl.tokensPerSec
	rl.tokens += tokensToAdd
	if rl.tokens > float64(rl.capacity) {
		rl.tokens = float64(rl.capacity)
	}
	rl.lastRefill = now
}

func (rl *RateLimiter) Capacity() int     { return rl.capacity }
func (rl *RateLimiter) Available() int    { rl.mu.Lock(); defer rl.mu.Unlock(); rl.refillTokens(); return int(rl.tokens) }
