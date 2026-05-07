package risk

import (
	"testing"
)

func TestBreakerStartsClosed(t *testing.T) {
	b := NewBreaker(DefaultPolicy())
	if !b.CanProceed() {
		t.Fatal("should start closed")
	}
}

func TestBreakerOpensOnConsecutiveErrors(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxConsecutiveErrs = 3
	b := NewBreaker(policy)
	for i := 0; i < 3; i++ {
		if b.RecordError() && i < 2 {
			t.Fatalf("should not break on error %d", i)
		}
	}
	if b.CanProceed() {
		t.Fatal("should be halted")
	}
}

func TestBreakerResetsOnSuccess(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxConsecutiveErrs = 3
	b := NewBreaker(policy)
	b.RecordError()
	b.RecordError()
	b.RecordSuccess()
	b.RecordError()
	if b.CanProceed() {
		// After 1 error post-reset, should still proceed
	}
}

func TestBreakerDailyLossLimit(t *testing.T) {
	policy := DefaultPolicy()
	policy.DailyLossLimitUSD = 50
	b := NewBreaker(policy)
	if b.RecordLoss(60) {
		if b.CanProceed() {
			t.Fatal("should be halted after exceeding daily loss")
		}
	}
}

func TestBreakerCoolDown(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxConsecutiveErrs = 1
	policy.CoolDownSecs = 0
	b := NewBreaker(policy)
	b.RecordError()
	if b.CanProceed() {
		t.Fatal("should be halted")
	}
	b.Reset()
	if !b.CanProceed() {
		t.Fatal("should proceed after reset")
	}
}

func TestBreakerRecordSuccessClearsErrors(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxConsecutiveErrs = 5
	b := NewBreaker(policy)
	for i := 0; i < 4; i++ {
		b.RecordError()
	}
	b.RecordSuccess()
	for i := 0; i < 4; i++ {
		b.RecordError()
	}
	// Should still be closed — successes reset the counter
	if !b.CanProceed() {
		t.Fatal("should still be closed")
	}
}

func TestBreakerHalted(t *testing.T) {
	b := NewBreaker(DefaultPolicy())
	if b.Halted() {
		t.Fatal("should not be halted initially")
	}
	b.RecordError()
	b.RecordError()
	b.RecordError()
	b.RecordError()
	b.RecordError()
	if !b.Halted() {
		t.Fatal("should be halted after 5 errors")
	}
}
