package paper

import "testing"

func TestBuyUpdatesLocalPositionWithoutExternalExecution(t *testing.T) {
	state := NewState("USD", 100)
	fill, err := state.Buy(Order{
		MarketID: "market-1",
		TokenID:  "yes-token",
		Price:    0.25,
		Size:     10,
	})
	if err != nil {
		t.Fatalf("Buy returned error: %v", err)
	}
	if fill.Live {
		t.Fatal("paper fill must not be live")
	}
	if state.Cash != 97.5 {
		t.Fatalf("Cash = %v, want 97.5", state.Cash)
	}
	if state.Positions["yes-token"].Size != 10 {
		t.Fatalf("position size = %v", state.Positions["yes-token"].Size)
	}
}
