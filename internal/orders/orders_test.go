package orders

import (
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

func TestOrderIntentValidate(t *testing.T) {
	oi := OrderIntent{
		TokenID:  "123",
		Side:     polytypes.SideBuy,
		Price:    polytypes.MustDecimal("0.55"),
		Size:     polytypes.MustDecimal("10"),
		TickSize: polytypes.TickSize{TickSize: "0.01"},
	}
	if err := oi.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestOrderIntentMissingTokenID(t *testing.T) {
	oi := OrderIntent{Side: polytypes.SideBuy}
	if err := oi.Validate(); err == nil {
		t.Fatal("expected error for missing token_id")
	}
}

func TestOrderIntentInvalidSide(t *testing.T) {
	oi := OrderIntent{TokenID: "123", Side: polytypes.Side(99)}
	if err := oi.Validate(); err == nil {
		t.Fatal("expected error for invalid side")
	}
}

func TestOrderIntentMissingPriceAndAmount(t *testing.T) {
	oi := OrderIntent{TokenID: "123", Side: polytypes.SideBuy}
	if err := oi.Validate(); err == nil {
		t.Fatal("expected error for missing price/amount")
	}
}

func TestOrderIntentMissingTickSize(t *testing.T) {
	oi := OrderIntent{TokenID: "123", Side: polytypes.SideBuy, Price: polytypes.MustDecimal("0.55")}
	if err := oi.Validate(); err == nil {
		t.Fatal("expected error for missing tick size")
	}
}

func TestComputeAmountsBuy(t *testing.T) {
	oi := OrderIntent{
		Side:  polytypes.SideBuy,
		Price: polytypes.MustDecimal("0.55"),
		Size:  polytypes.MustDecimal("10"),
	}
	maker, taker := ComputeAmounts(&oi)
	if maker == nil || taker == nil {
		t.Fatal("amounts are nil")
	}
}

func TestComputeAmountsSell(t *testing.T) {
	oi := OrderIntent{
		Side:  polytypes.SideSell,
		Price: polytypes.MustDecimal("0.45"),
		Size:  polytypes.MustDecimal("5"),
	}
	maker, taker := ComputeAmounts(&oi)
	if maker == nil || taker == nil {
		t.Fatal("amounts are nil")
	}
}

func TestFluentBuilder(t *testing.T) {
	intent, err := NewBuilder("tok-123", polytypes.SideBuy).
		Price("0.55").Size("10").TickSize("0.01").FeeRateBps(0).
		OrderType(polytypes.OrderTypeGTC).SigType(polytypes.SignatureEOA).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if intent.TokenID != "tok-123" {
		t.Fatalf("token_id: %s", intent.TokenID)
	}
	if intent.Side != polytypes.SideBuy {
		t.Fatal("side mismatch")
	}
	if intent.Price.String() != "0.550000" {
		t.Fatalf("price: %s", intent.Price.String())
	}
}

func TestFluentBuilderMustBuildPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewBuilder("", polytypes.SideBuy).MustBuild()
}

func TestValidatePriceAgainstTick(t *testing.T) {
	price := polytypes.MustDecimal("0.55").Rat()
	tick := polytypes.MustDecimal("0.01").Rat()
	if err := ValidatePriceAgainstTick(price, tick); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePriceBelowTick(t *testing.T) {
	price := polytypes.MustDecimal("0.005").Rat()
	tick := polytypes.MustDecimal("0.01").Rat()
	if err := ValidatePriceAgainstTick(price, tick); err == nil {
		t.Fatal("expected error for price below tick size")
	}
}

func TestValidatePriceAboveMax(t *testing.T) {
	price := polytypes.MustDecimal("0.995").Rat()
	tick := polytypes.MustDecimal("0.01").Rat()
	if err := ValidatePriceAgainstTick(price, tick); err == nil {
		t.Fatal("expected error for price above 1-tick")
	}
}
