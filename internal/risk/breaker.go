package risk

import (
	"sync"
	"time"
)

// TripReason explains why the breaker tripped.
type TripReason int

const (
	ReasonConsecutiveErrors TripReason = iota
	ReasonDailyLossLimit
	ReasonPositionPerMarket
	ReasonTotalPosition
	ReasonManualHalt
)

func (r TripReason) String() string {
	switch r {
	case ReasonConsecutiveErrors:
		return "consecutive_errors"
	case ReasonDailyLossLimit:
		return "daily_loss_limit"
	case ReasonPositionPerMarket:
		return "position_per_market"
	case ReasonTotalPosition:
		return "total_position"
	case ReasonManualHalt:
		return "manual_halt"
	default:
		return "unknown"
	}
}

// Policy defines risk limits for trading.
type Policy struct {
	MaxOrderUSD          float64 `json:"max_order_usd"`
	MaxOpenOrders        int     `json:"max_open_orders"`
	DailyLossLimitUSD    float64 `json:"daily_loss_limit_usd"`
	DailyPnLResetHour    int     `json:"daily_pnl_reset_hour"`
	MaxConsecutiveErrs   int     `json:"max_consecutive_errors"`
	CoolDownSecs         int     `json:"cooldown_secs"`
	MaxPositionPerMarket float64 `json:"max_position_per_market"`
	MaxTotalPosition     float64 `json:"max_total_position"`
}

// DefaultPolicy returns conservative defaults.
func DefaultPolicy() Policy {
	return Policy{
		MaxOrderUSD:          10.0,
		MaxOpenOrders:        5,
		DailyLossLimitUSD:    100.0,
		DailyPnLResetHour:    0,
		MaxConsecutiveErrs:   5,
		CoolDownSecs:         300,
		MaxPositionPerMarket: 50.0,
		MaxTotalPosition:     200.0,
	}
}

// Status is a snapshot of the breaker's current state.
type Status struct {
	Halted          bool               `json:"halted"`
	TripReason      TripReason         `json:"trip_reason"`
	TripReasonMsg   string             `json:"trip_reason_message"`
	LastBreak       time.Time          `json:"last_break"`
	ConsecutiveErrs int                `json:"consecutive_errors"`
	DailyLossUSD    float64            `json:"daily_loss_usd"`
	TotalPosition   float64            `json:"total_position_usd"`
	Positions       map[string]float64 `json:"positions"`
	CoolDownReady   bool               `json:"cooldown_ready"`
}

// Breaker tracks violations and can halt trading.
type Breaker struct {
	policy          Policy
	mu              sync.Mutex
	consecutiveErrs int
	dailyLoss       float64
	dailyLossReset  time.Time
	lastBreak       time.Time
	halted          bool
	tripReason      TripReason
	positions       map[string]float64
}

// NewBreaker creates a risk circuit breaker.
func NewBreaker(policy Policy) *Breaker {
	return &Breaker{
		policy:    policy,
		positions: make(map[string]float64),
	}
}

// RecordError increments the error counter and returns true if we should break.
func (b *Breaker) RecordError() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.checkDailyResetLocked()
	b.consecutiveErrs++
	if b.consecutiveErrs >= b.policy.MaxConsecutiveErrs {
		b.halted = true
		b.tripReason = ReasonConsecutiveErrors
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
	b.checkDailyResetLocked()
	b.dailyLoss += amount
	if b.dailyLoss >= b.policy.DailyLossLimitUSD {
		b.halted = true
		b.tripReason = ReasonDailyLossLimit
		b.lastBreak = time.Now()
		return true
	}
	return false
}

// RecordPosition updates the position for a token and checks limits.
// Returns true if a limit was breached and the breaker tripped.
func (b *Breaker) RecordPosition(tokenID string, size float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.positions[tokenID] = size
	total := 0.0
	for _, v := range b.positions {
		total += abs(v)
	}
	if b.policy.MaxPositionPerMarket > 0 && abs(size) > b.policy.MaxPositionPerMarket {
		b.halted = true
		b.tripReason = ReasonPositionPerMarket
		b.lastBreak = time.Now()
		return true
	}
	if b.policy.MaxTotalPosition > 0 && total > b.policy.MaxTotalPosition {
		b.halted = true
		b.tripReason = ReasonTotalPosition
		b.lastBreak = time.Now()
		return true
	}
	return false
}

// Halt manually trips the breaker.
func (b *Breaker) Halt() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.halted = true
	b.tripReason = ReasonManualHalt
	b.lastBreak = time.Now()
}

// Status returns a snapshot of the current breaker state.
func (b *Breaker) Status() Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.checkDailyResetLocked()
	coolDownReady := false
	if b.halted && b.policy.CoolDownSecs > 0 {
		coolDownReady = time.Since(b.lastBreak) > time.Duration(b.policy.CoolDownSecs)*time.Second
	}
	total := 0.0
	for _, v := range b.positions {
		total += abs(v)
	}
	posCopy := make(map[string]float64, len(b.positions))
	for k, v := range b.positions {
		posCopy[k] = v
	}
	return Status{
		Halted:          b.halted,
		TripReason:      b.tripReason,
		TripReasonMsg:   b.tripReason.String(),
		LastBreak:       b.lastBreak,
		ConsecutiveErrs: b.consecutiveErrs,
		DailyLossUSD:    b.dailyLoss,
		TotalPosition:   total,
		Positions:       posCopy,
		CoolDownReady:   coolDownReady,
	}
}

// CanProceed returns true if trading is allowed.
func (b *Breaker) CanProceed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.checkDailyResetLocked()
	if b.halted {
		if b.policy.CoolDownSecs > 0 {
			if time.Since(b.lastBreak) > time.Duration(b.policy.CoolDownSecs)*time.Second {
				b.halted = false
				b.consecutiveErrs = 0
				b.tripReason = 0
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
	b.tripReason = 0
	b.consecutiveErrs = 0
	b.dailyLoss = 0
	b.positions = make(map[string]float64)
}

// checkDailyResetLocked resets daily loss if we've crossed the configured UTC reset hour.
// Must be called with b.mu held.
func (b *Breaker) checkDailyResetLocked() {
	if b.policy.DailyLossLimitUSD <= 0 {
		return
	}
	now := time.Now().UTC()
	if b.dailyLossReset.IsZero() {
		b.dailyLossReset = now
		return
	}
	lastReset := b.dailyLossReset
	if now.Day() != lastReset.Day() || now.Month() != lastReset.Month() || now.Year() != lastReset.Year() {
		if now.Hour() >= b.policy.DailyPnLResetHour {
			b.dailyLoss = 0
			b.dailyLossReset = now
		}
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
