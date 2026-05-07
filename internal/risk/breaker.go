package risk

import (
	"sync"
	"time"
)

// Policy defines risk limits for trading.
type Policy struct {
	MaxOrderUSD        float64 `json:"max_order_usd"`
	MaxOpenOrders      int     `json:"max_open_orders"`
	DailyLossLimitUSD  float64 `json:"daily_loss_limit_usd"`
	MaxConsecutiveErrs int     `json:"max_consecutive_errors"`
	CoolDownSecs       int     `json:"cooldown_secs"`
}

// DefaultPolicy returns conservative defaults.
func DefaultPolicy() Policy {
	return Policy{
		MaxOrderUSD:        10.0,
		MaxOpenOrders:      5,
		DailyLossLimitUSD:  100.0,
		MaxConsecutiveErrs: 5,
		CoolDownSecs:       60,
	}
}

// Breaker tracks violations and can halt trading.
type Breaker struct {
	policy          Policy
	mu              sync.Mutex
	consecutiveErrs int
	dailyLoss       float64
	lastBreak       time.Time
	halted          bool
}

// NewBreaker creates a risk circuit breaker.
func NewBreaker(policy Policy) *Breaker {
	return &Breaker{policy: policy}
}

// RecordError increments the error counter and returns true if we should break.
func (b *Breaker) RecordError() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consecutiveErrs++
	if b.consecutiveErrs >= b.policy.MaxConsecutiveErrs {
		b.halted = true
		b.lastBreak = time.Now()
		return true
	}
	return false
}

// RecordSuccess resets the error counter.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consecutiveErrs = 0
}

// RecordLoss adds to daily PnL. Returns true if daily limit hit.
func (b *Breaker) RecordLoss(amount float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dailyLoss += amount
	if b.dailyLoss >= b.policy.DailyLossLimitUSD {
		b.halted = true
		b.lastBreak = time.Now()
		return true
	}
	return false
}

// CanProceed returns true if trading is allowed.
func (b *Breaker) CanProceed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.halted {
		if b.policy.CoolDownSecs > 0 {
			if time.Since(b.lastBreak) > time.Duration(b.policy.CoolDownSecs)*time.Second {
				b.halted = false
				b.consecutiveErrs = 0
				return true
			}
		}
		return false
	}
	return true
}

// Halted returns whether the breaker is currently tripped.
func (b *Breaker) Halted() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.halted
}

// Reset clears all breaker state.
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.halted = false
	b.consecutiveErrs = 0
	b.dailyLoss = 0
}
