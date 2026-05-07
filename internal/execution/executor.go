package execution

import (
	"context"
	"fmt"

	"github.com/TrebuchetDynamics/polygolem/internal/orders"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

// Executor defines the order execution contract.
// Paper and live implementations share this interface per PRD R6.
type Executor interface {
	// Place submits one order.
	Place(ctx context.Context, intent *orders.OrderIntent) (*orders.OrderResponse, error)

	// Cancel cancels an order by ID.
	Cancel(ctx context.Context, orderID string) error

	// GetOrder fetches a single order by ID.
	GetOrder(ctx context.Context, orderID string) (*orders.OrderResponse, error)

	// ListOrders lists open orders.
	ListOrders(ctx context.Context) ([]orders.OrderResponse, error)
}

// PaperExecutor simulates order execution locally.
// Never touches the network or authenticated endpoints.
type PaperExecutor struct {
	positions *PaperPositions
}

// PaperPositions tracks local paper trading state.
type PaperPositions struct {
	Cash    polytypes.Decimal        `json:"cash"`
	Orders  []PaperOrder             `json:"orders"`
	Fills   []PaperFill              `json:"fills"`
}

// PaperOrder represents a simulated paper order.
type PaperOrder struct {
	OrderID    string              `json:"order_id"`
	TokenID    string              `json:"token_id"`
	Side       string              `json:"side"`
	Price      string              `json:"price"`
	Size       string              `json:"size"`
	Status     orders.LifecycleState `json:"status"`
	FilledSize string              `json:"filled_size"`
}

// PaperFill represents a simulated paper fill.
type PaperFill struct {
	OrderID  string `json:"order_id"`
	TokenID  string `json:"token_id"`
	Side     string `json:"side"`
	Price    string `json:"price"`
	Size     string `json:"size"`
}

// NewPaperExecutor creates a paper executor with initial state.
func NewPaperExecutor(initialCash string) *PaperExecutor {
	return &PaperExecutor{
		positions: &PaperPositions{
			Cash:   polytypes.MustDecimal(initialCash),
			Orders: []PaperOrder{},
			Fills:  []PaperFill{},
		},
	}
}

func (pe *PaperExecutor) Place(ctx context.Context, intent *orders.OrderIntent) (*orders.OrderResponse, error) {
	if err := intent.Validate(); err != nil {
		return nil, err
	}
	orderID := fmt.Sprintf("paper-%d", len(pe.positions.Orders)+1)
	pe.positions.Orders = append(pe.positions.Orders, PaperOrder{
		OrderID:    orderID,
		TokenID:    intent.TokenID,
		Side:       intent.Side.String(),
		Price:      intent.Price.String(),
		Size:       intent.Size.String(),
		Status:     orders.StateAccepted,
		FilledSize: "0",
	})
	// Simulate fill
	pe.positions.Orders[len(pe.positions.Orders)-1].Status = orders.StateMatched
	pe.positions.Orders[len(pe.positions.Orders)-1].FilledSize = intent.Size.String()
	pe.positions.Fills = append(pe.positions.Fills, PaperFill{
		OrderID: orderID,
		TokenID: intent.TokenID,
		Side:    intent.Side.String(),
		Price:   intent.Price.String(),
		Size:    intent.Size.String(),
	})
	return &orders.OrderResponse{
		Success: true,
		OrderID: orderID,
		Status:  string(orders.StateMatched),
	}, nil
}

func (pe *PaperExecutor) Cancel(ctx context.Context, orderID string) error {
	for i, o := range pe.positions.Orders {
		if o.OrderID == orderID {
			pe.positions.Orders[i].Status = orders.StateCanceled
			return nil
		}
	}
	return fmt.Errorf("order %s not found", orderID)
}

func (pe *PaperExecutor) GetOrder(ctx context.Context, orderID string) (*orders.OrderResponse, error) {
	for _, o := range pe.positions.Orders {
		if o.OrderID == orderID {
			return &orders.OrderResponse{
				Success: true,
				OrderID: o.OrderID,
				Status:  string(o.Status),
			}, nil
		}
	}
	return nil, fmt.Errorf("order %s not found", orderID)
}

func (pe *PaperExecutor) ListOrders(ctx context.Context) ([]orders.OrderResponse, error) {
	result := make([]orders.OrderResponse, len(pe.positions.Orders))
	for i, o := range pe.positions.Orders {
		result[i] = orders.OrderResponse{
			Success: true,
			OrderID: o.OrderID,
			Status:  string(o.Status),
		}
	}
	return result, nil
}

func (pe *PaperExecutor) Positions() *PaperPositions {
	return pe.positions
}
