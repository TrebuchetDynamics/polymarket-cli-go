package execution

import (
	"context"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/orders"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

func TestPaperExecutorPlaceBuy(t *testing.T) {
	pe := NewPaperExecutor("1000")
	intent, _ := orders.NewBuilder("tok-1", polytypes.SideBuy).
		Price("0.55").Size("10").TickSize("0.01").FeeRateBps(0).Build()
	resp, err := pe.Place(context.Background(), &intent)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Fatal("expected success")
	}
	if resp.OrderID == "" {
		t.Fatal("expected order ID")
	}
	pos := pe.Positions()
	if len(pos.Orders) != 1 {
		t.Fatalf("expected 1 order: %d", len(pos.Orders))
	}
	if len(pos.Fills) != 1 {
		t.Fatalf("expected 1 fill: %d", len(pos.Fills))
	}
}

func TestPaperExecutorCancel(t *testing.T) {
	pe := NewPaperExecutor("1000")
	intent, _ := orders.NewBuilder("tok-1", polytypes.SideBuy).
		Price("0.55").Size("10").TickSize("0.01").FeeRateBps(0).Build()
	resp, _ := pe.Place(context.Background(), &intent)
	if err := pe.Cancel(context.Background(), resp.OrderID); err != nil {
		t.Fatal(err)
	}
}

func TestPaperExecutorCancelNotFound(t *testing.T) {
	pe := NewPaperExecutor("1000")
	if err := pe.Cancel(context.Background(), "nonexistent"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPaperExecutorGetOrder(t *testing.T) {
	pe := NewPaperExecutor("1000")
	intent, _ := orders.NewBuilder("tok-1", polytypes.SideBuy).
		Price("0.55").Size("10").TickSize("0.01").FeeRateBps(0).Build()
	resp, _ := pe.Place(context.Background(), &intent)
	order, err := pe.GetOrder(context.Background(), resp.OrderID)
	if err != nil {
		t.Fatal(err)
	}
	if order.OrderID != resp.OrderID {
		t.Fatalf("order ID mismatch: %s vs %s", order.OrderID, resp.OrderID)
	}
}

func TestPaperExecutorListOrders(t *testing.T) {
	pe := NewPaperExecutor("1000")
	intent, _ := orders.NewBuilder("tok-1", polytypes.SideBuy).
		Price("0.55").Size("10").TickSize("0.01").FeeRateBps(0).Build()
	pe.Place(context.Background(), &intent)
	pe.Place(context.Background(), &intent)
	orders, err := pe.ListOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders: %d", len(orders))
	}
}

func TestPaperExecutorRejectsInvalidIntent(t *testing.T) {
	pe := NewPaperExecutor("1000")
	_, err := pe.Place(context.Background(), &orders.OrderIntent{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
