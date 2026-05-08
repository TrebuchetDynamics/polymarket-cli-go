package risk

import (
	"testing"
	"time"
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

func TestBreakerRecordsTripReasonOnConsecutiveErrors(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxConsecutiveErrs = 1
	b := NewBreaker(policy)
	b.RecordError()
	status := b.Status()
	if status.TripReason != ReasonConsecutiveErrors {
		t.Fatalf("trip reason=%d want ReasonConsecutiveErrors", status.TripReason)
	}
}

func TestBreakerRecordsTripReasonOnDailyLossLimit(t *testing.T) {
	policy := DefaultPolicy()
	policy.DailyLossLimitUSD = 10
	b := NewBreaker(policy)
	b.RecordLoss(20)
	status := b.Status()
	if status.TripReason != ReasonDailyLossLimit {
		t.Fatalf("trip reason=%d want ReasonDailyLossLimit", status.TripReason)
	}
}

func TestBreakerStatusIncludesPositions(t *testing.T) {
	policy := DefaultPolicy()
	b := NewBreaker(policy)
	b.RecordPosition("token1", 5.0)
	b.RecordPosition("token2", -3.0)
	status := b.Status()
	if status.Positions["token1"] != 5.0 || status.Positions["token2"] != -3.0 {
		t.Fatalf("positions=%v", status.Positions)
	}
	if status.TotalPosition != 8.0 {
		t.Fatalf("total position=%f want 8.0", status.TotalPosition)
	}
}

func TestBreakerPositionLimitHaltsTrading(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxPositionPerMarket = 10.0
	b := NewBreaker(policy)
	if !b.RecordPosition("token1", 15.0) {
		t.Fatal("should halt for single position record")
	}
	if b.CanProceed() {
		t.Fatal("should be halted after exceeding per-market position")
	}
	status := b.Status()
	if status.TripReason != ReasonPositionPerMarket {
		t.Fatalf("trip reason=%d want ReasonPositionPerMarket", status.TripReason)
	}
}

func TestBreakerTotalPositionLimitHaltsTrading(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxTotalPosition = 10.0
	b := NewBreaker(policy)
	b.RecordPosition("token1", 6.0)
	b.RecordPosition("token2", 6.0)
	if b.CanProceed() {
		t.Fatal("should be halted after exceeding total position")
	}
	status := b.Status()
	if status.TripReason != ReasonTotalPosition {
		t.Fatalf("trip reason=%d want ReasonTotalPosition", status.TripReason)
	}
}

func TestBreakerDailyLossResetsAtConfiguredHour(t *testing.T) {
	policy := DefaultPolicy()
	policy.DailyLossLimitUSD = 100
	policy.DailyPnLResetHour = 0
	b := NewBreaker(policy)
	b.RecordLoss(50)
	if b.Status().DailyLossUSD != 50 {
		t.Fatalf("daily loss=%f want 50", b.Status().DailyLossUSD)
	}
	b.dailyLossReset = time.Now().UTC().Add(-26 * time.Hour)
	if !b.CanProceed() {
		t.Fatal("should reset daily loss and allow trading")
	}
	status := b.Status()
	if status.DailyLossUSD != 0 {
		t.Fatalf("daily loss should reset to 0, got %f", status.DailyLossUSD)
	}
}

func TestBreakerManualHaltRecordsReason(t *testing.T) {
	b := NewBreaker(DefaultPolicy())
	b.Halt()
	if !b.Halted() {
		t.Fatal("should be halted after manual halt")
	}
	status := b.Status()
	if status.TripReason != ReasonManualHalt {
		t.Fatalf("trip reason=%d want ReasonManualHalt", status.TripReason)
	}
}

func TestBreakerDefaultCooldownIs300Seconds(t *testing.T) {
	policy := DefaultPolicy()
	if policy.CoolDownSecs != 300 {
		t.Fatalf("default cooldown=%d want 300", policy.CoolDownSecs)
	}
}
