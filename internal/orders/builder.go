package orders

import "github.com/TrebuchetDynamics/polygolem/internal/polytypes"

// Builder provides a fluent API for constructing OrderIntent.
// Stolen from rs-clob-client's order_builder.rs chainable pattern.
type Builder struct {
	intent OrderIntent
	err    error
}

// NewBuilder starts an order builder with required fields. SigType is fixed
// to POLY_1271 (sigtype 3) — the only signature type accepted by Polymarket
// V2 since the 2026-04-28 cutover.
func NewBuilder(tokenID string, side polytypes.Side) *Builder {
	return &Builder{
		intent: OrderIntent{
			TokenID:   tokenID,
			Side:      side,
			OrderType: polytypes.OrderTypeGTC,
			SigType:   polytypes.SignaturePoly1271,
		},
	}
}

func (b *Builder) Price(s string) *Builder {
	b.intent.Price = polytypes.MustDecimal(s)
	return b
}

func (b *Builder) Size(s string) *Builder {
	b.intent.Size = polytypes.MustDecimal(s)
	return b
}

func (b *Builder) AmountUSDC(s string) *Builder {
	b.intent.AmountUSDC = polytypes.MustDecimal(s)
	return b
}

func (b *Builder) OrderType(ot polytypes.OrderType) *Builder {
	b.intent.OrderType = ot
	return b
}

func (b *Builder) TickSize(value string) *Builder {
	b.intent.TickSize = polytypes.TickSize{
		TickSize:         value,
		MinimumTickSize:  value,
		MinimumOrderSize: "1",
	}
	return b
}

func (b *Builder) NegRisk(v bool) *Builder {
	b.intent.NegRisk = v
	return b
}

func (b *Builder) FeeRateBps(r int) *Builder {
	b.intent.FeeRateBps = r
	return b
}

func (b *Builder) Nonce(n uint64) *Builder {
	b.intent.Nonce = n
	return b
}

func (b *Builder) Expiration(unix int64) *Builder {
	b.intent.Expiration = unix
	return b
}

func (b *Builder) Funder(addr string) *Builder {
	b.intent.Funder = addr
	return b
}

func (b *Builder) PostOnly(v bool) *Builder {
	b.intent.PostOnly = v
	return b
}

// Build validates and returns the OrderIntent.
func (b *Builder) Build() (OrderIntent, error) {
	if err := b.intent.Validate(); err != nil {
		return OrderIntent{}, err
	}
	return b.intent, nil
}

// MustBuild panics if validation fails. Use for tests with known-good inputs.
func (b *Builder) MustBuild() OrderIntent {
	oi, err := b.Build()
	if err != nil {
		panic("order builder: " + err.Error())
	}
	return oi
}
