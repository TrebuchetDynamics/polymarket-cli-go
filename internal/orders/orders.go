// Package orders provides order intent, building, validation, and lifecycle models.
// Based on patterns from polymarket-go-sdk and rs-clob-client.
package orders

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/errors"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

// OrderIntent represents a user's order before building/signing.
type OrderIntent struct {
	TokenID      string              `json:"token_id"`
	Side         polytypes.Side      `json:"side"`
	Price        polytypes.Decimal   `json:"price"`
	Size         polytypes.Decimal   `json:"size"`
	AmountUSDC   polytypes.Decimal   `json:"amount_usdc,omitempty"`
	OrderType    polytypes.OrderType `json:"order_type"`
	SigType      polytypes.SignatureType `json:"signature_type"`
	TickSize     polytypes.TickSize  `json:"tick_size"`
	NegRisk      bool                `json:"neg_risk"`
	FeeRateBps   int                 `json:"fee_rate_bps"`
	Nonce        uint64              `json:"nonce,omitempty"`
	Expiration   int64               `json:"expiration,omitempty"`
	Funder       string              `json:"funder,omitempty"`
	PostOnly     bool                `json:"post_only,omitempty"`
}

// Validate checks that the order intent is well-formed.
func (oi *OrderIntent) Validate() error {
	if oi.TokenID == "" {
		return errors.New(errors.CodeMissingField, "token_id is required")
	}
	switch oi.Side {
	case polytypes.SideBuy, polytypes.SideSell:
	default:
		return errors.New(errors.CodeInvalidValue, "side must be BUY or SELL")
	}
	if oi.Price.IsZero() && oi.AmountUSDC.IsZero() {
		return errors.New(errors.CodeMissingField, "price or amount_usdc required")
	}
	if oi.Size.IsZero() && oi.AmountUSDC.IsZero() {
		return errors.New(errors.CodeMissingField, "size or amount_usdc required")
	}
	if oi.TickSize.TickSize == "" || oi.TickSize.TickSize == "0" {
		return errors.New(errors.CodeInvalidValue, "valid tick_size required")
	}
	if oi.FeeRateBps < 0 {
		return errors.New(errors.CodeInvalidValue, "invalid fee_rate_bps")
	}
	return nil
}

// LifecycleState models order status through its lifecycle.
type LifecycleState string

const (
	StateCreated  LifecycleState = "created"
	StateAccepted LifecycleState = "accepted"
	StateLive     LifecycleState = "live"
	StatePartial  LifecycleState = "partial"
	StateMatched  LifecycleState = "matched"
	StateCanceled LifecycleState = "canceled"
	StateRejected LifecycleState = "rejected"
	StateFailed   LifecycleState = "failed"
	StateMined    LifecycleState = "mined"
	StateConfirmed LifecycleState = "confirmed"
)

// SignedOrder represents a signed but unposted order.
type SignedOrder struct {
	Order     OrderData  `json:"order"`
	Signature string     `json:"signature"`
	Owner     string     `json:"owner"`
	OrderType polytypes.OrderType `json:"order_type"`
}

// OrderData contains the core order parameters for signing.
type OrderData struct {
	TokenID    string `json:"token_id"`
	Price      string `json:"price"`
	Size       string `json:"size"`
	Side       string `json:"side"`
	FeeRateBps string `json:"fee_rate_bps"`
	Nonce      string `json:"nonce"`
	Expiration string `json:"expiration"`
	Taker      string `json:"taker"`
	Maker      string `json:"maker"`
	Salt       string `json:"salt"`
	Signer     string `json:"signer"`
}

// OrderResponse represents the CLOB response from posting an order.
type OrderResponse struct {
	Success    bool   `json:"success"`
	OrderID    string `json:"order_id"`
	Status     string `json:"status"`
	ErrorMsg   string `json:"error_msg,omitempty"`
	TxHash     string `json:"transaction_hash,omitempty"`
}

// ComputeAmounts calculates maker/taker amounts for an order.
// BUY: makerAmount = size * price (spend USDC), takerAmount = size (receive tokens)
// SELL: makerAmount = size (provide tokens), takerAmount = size * price (receive USDC)
func ComputeAmounts(intent *OrderIntent) (makerAmount, takerAmount *big.Rat) {
	size := intent.Size.Rat()
	price := intent.Price.Rat()

	maker := new(big.Rat)
	taker := new(big.Rat)

	if intent.Side == polytypes.SideBuy {
		maker.Mul(size, price)
		taker.Set(size)
	} else {
		maker.Set(size)
		taker.Mul(size, price)
	}
	return maker, taker
}

// RoundToTick rounds a value to the nearest tick size multiple.
func RoundToTick(value *big.Rat, tickSize *big.Rat) *big.Rat {
	q := new(big.Rat).Quo(value, tickSize)
	round := new(big.Rat).SetInt(new(big.Int).Div(q.Num(), q.Denom()))
	return new(big.Rat).Mul(round, tickSize)
}

// ValidatePriceAgainstTick checks that price is within [tickSize, 1-tickSize].
func ValidatePriceAgainstTick(price, tickSize *big.Rat) error {
	one := new(big.Rat).SetInt64(1)
	min := new(big.Rat).Set(tickSize)
	max := new(big.Rat).Sub(one, tickSize)
	if price.Cmp(min) < 0 || price.Cmp(max) > 0 {
		return errors.New(errors.CodeInvalidValue,
			fmt.Sprintf("price %s must be in [%s, %s]", price.FloatString(6), min.FloatString(6), max.FloatString(6)))
	}
	return nil
}

// BuildSalt generates a deterministic salt for order signing.
func BuildSalt(seed uint64) string {
	return strconv.FormatUint(seed, 10)
}

// DefaultExpiration returns the default order expiration (now + 365 days).
func DefaultExpiration() int64 {
	return time.Now().Add(365 * 24 * time.Hour).Unix()
}
