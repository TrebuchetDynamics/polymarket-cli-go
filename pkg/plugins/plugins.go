// Package plugins defines extension points for third-party consumers.
package plugins

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

// MarketDataPlugin allows custom market resolution and filtering.
type MarketDataPlugin interface {
	// Resolve takes an asset and timeframe and returns the best-matching market.
	Resolve(ctx context.Context, asset, timeframe string) (*types.Market, error)
	// Filter returns true if the market passes the plugin's criteria.
	Filter(ctx context.Context, market *types.Market) (bool, error)
}

// RiskPlugin allows custom pre-trade risk checks.
type RiskPlugin interface {
	// CheckOrder evaluates an order before it is signed and submitted.
	// Returns nil if the order passes; an error blocks the order.
	CheckOrder(ctx context.Context, order Order) error
}

// Order is the minimal order representation passed to risk plugins.
type Order struct {
	TokenID   string
	Side      string
	Price     string
	Size      string
	OrderType string
}
