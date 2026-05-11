// Package orders provides an experimental public API for building and
// validating Polymarket V2 orders.
//
// WARNING: This package is experimental. APIs may change without notice
// in patch releases. Do not depend on it for stable production code.
// Track github.com/TrebuchetDynamics/polygolem/issues for stabilization.
package orders

import "fmt"

// Side represents the order side.
type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// OrderType represents the order type.
type OrderType string

const (
	OrderTypeGTC OrderType = "GTC"
	OrderTypeFOK OrderType = "FOK"
	OrderTypeFAK OrderType = "FAK"
)

// OrderIntent represents a user's order before building/signing.
type OrderIntent struct {
	TokenID    string
	Side       Side
	Price      string
	Size       string
	AmountUSDC string
	OrderType  OrderType
	TickSize   string
	NegRisk    bool
	FeeRateBps int
	Expiration int64
	Funder     string
	PostOnly   bool
}

// Validate checks that the order intent is well-formed.
func (oi *OrderIntent) Validate() error {
	if oi.TokenID == "" {
		return fmt.Errorf("token_id is required")
	}
	switch oi.Side {
	case SideBuy, SideSell:
	default:
		return fmt.Errorf("side must be BUY or SELL")
	}
	if oi.Price == "" && oi.AmountUSDC == "" {
		return fmt.Errorf("price or amount_usdc required")
	}
	if oi.Size == "" && oi.AmountUSDC == "" {
		return fmt.Errorf("size or amount_usdc required")
	}
	if oi.TickSize == "" || oi.TickSize == "0" {
		return fmt.Errorf("valid tick_size required")
	}
	if oi.FeeRateBps < 0 {
		return fmt.Errorf("invalid fee_rate_bps")
	}
	return nil
}
