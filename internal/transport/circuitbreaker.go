package transport

import (
	"sync"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/errors"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half-open"
)

// CircuitBreakerConfig holds configuration.
type CircuitBreakerConfig struct {
	MaxFailures     int
	ResetTimeout    time.Duration
	HalfOpenMaxReqs int
}

func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:     5,
		ResetTimeout:    60 * time.Second,
		HalfOpenMaxReqs: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern.
// Stolen from polymarket-go-sdk/pkg/transport/circuitbreaker.go.
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenMaxReqs int

	mu              sync.RWMutex
	state           CircuitState
	failures        int
	lastFailTime    time.Time
	halfOpenReqs    int
	halfOpenSuccess int
	halfOpenFailure int
}

// FailurePredicate determines whether an error counts as a failure.
type FailurePredicate func(error) bool

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 5
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 60 * time.Second
	}
	if config.HalfOpenMaxReqs <= 0 {
		config.HalfOpenMaxReqs = 3
	}
	return &CircuitBreaker{
		maxFailures:     config.MaxFailures,
		resetTimeout:    config.ResetTimeout,
		halfOpenMaxReqs: config.HalfOpenMaxReqs,
		state:           StateClosed,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	return cb.CallWithPredicate(fn, nil)
}

func (cb *CircuitBreaker) CallWithPredicate(fn func() error, shouldCount FailurePredicate) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}
	err := fn()
	count := err != nil
	if count && shouldCount != nil {
		count = shouldCount(err)
	}
	cb.afterRequest(err, count)
	return err
}

func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailTime) <= cb.resetTimeout {
			return errors.New(errors.CodeCircuitOpen, "circuit breaker is open")
		}
		cb.state = StateHalfOpen
		cb.halfOpenReqs = 0
		cb.halfOpenSuccess = 0
		cb.halfOpenFailure = 0
		fallthrough
	case StateHalfOpen:
		if cb.halfOpenReqs >= cb.halfOpenMaxReqs {
			return errors.New(errors.CodeRateLimited, "circuit half-open: too many requests")
		}
		cb.halfOpenReqs++
		return nil
	}
	return nil
}

func (cb *CircuitBreaker) afterRequest(err error, countFailure bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err != nil && countFailure {
		cb.recordFailure()
		return
	}
	cb.recordSuccess()
}

func (cb *CircuitBreaker) recordFailure() {
	cb.lastFailTime = time.Now()
	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.failures = cb.maxFailures
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.halfOpenSuccess++
		if cb.halfOpenSuccess >= cb.halfOpenMaxReqs {
			cb.state = StateClosed
			cb.failures = 0
		}
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenReqs = 0
	cb.halfOpenSuccess = 0
	cb.halfOpenFailure = 0
}

// RecordResult records a request result for manual beforeRequest/afterRequest flows.
func (cb *CircuitBreaker) RecordResult(err error) {
	cb.afterRequest(err, err != nil)
}

type CircuitBreakerStats struct {
	State           CircuitState
	Failures        int
	HalfOpenReqs    int
	HalfOpenSuccess int
	HalfOpenFailure int
	LastFailTime    time.Time
}

func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return CircuitBreakerStats{
		State:           cb.state,
		Failures:        cb.failures,
		HalfOpenReqs:    cb.halfOpenReqs,
		HalfOpenSuccess: cb.halfOpenSuccess,
		HalfOpenFailure: cb.halfOpenFailure,
		LastFailTime:    cb.lastFailTime,
	}
}
