package polytypes

import (
	"encoding/json"
	"math/big"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

type NormalizedTime = types.NormalizedTime
type StringOrArray = types.StringOrArray

// Decimal is a string-backed decimal for Polymarket prices/amounts.
// Avoids float64 precision loss. Backed by math/big.Rat.
type Decimal struct {
	rat *big.Rat
}

func NewDecimal(s string) (Decimal, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return Decimal{}, &InvalidDecimalError{Input: s}
	}
	return Decimal{rat: r}, nil
}

func MustDecimal(s string) Decimal {
	d, err := NewDecimal(s)
	if err != nil {
		panic(err)
	}
	return d
}

func DecimalFromInt64(n int64) Decimal {
	return Decimal{rat: new(big.Rat).SetInt64(n)}
}

func (d Decimal) String() string {
	if d.rat == nil {
		return "0"
	}
	return d.rat.FloatString(6)
}

func (d Decimal) Float64() float64 {
	if d.rat == nil {
		return 0
	}
	f, _ := d.rat.Float64()
	return f
}

func (d Decimal) Rat() *big.Rat {
	if d.rat == nil {
		return new(big.Rat)
	}
	return new(big.Rat).Set(d.rat)
}

func (d Decimal) IsZero() bool {
	return d.rat == nil || d.rat.Sign() == 0
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Decimal) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := NewDecimal(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

type InvalidDecimalError struct{ Input string }

func (e *InvalidDecimalError) Error() string { return "invalid decimal: " + e.Input }

type Side int

const (
	SideBuy  Side = 0
	SideSell Side = 1
)

func (s Side) String() string {
	switch s {
	case SideBuy:
		return "BUY"
	case SideSell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

type OrderType string

const (
	OrderTypeGTC OrderType = "GTC"
	OrderTypeFOK OrderType = "FOK"
	OrderTypeGTD OrderType = "GTD"
	OrderTypeFAK OrderType = "FAK"
)

// SignatureType is the on-wire `signatureType` enum. Polymarket V2
// (2026-04-28 cutover) accepts only sigtype 3 (deposit wallet, POLY_1271).
// Sigtypes 0/1/2 (EOA / proxy / Gnosis Safe) are dead — `clob/order` rejects
// them with "maker address not allowed, please use the deposit wallet flow".
type SignatureType int

const (
	SignaturePoly1271 SignatureType = 3
)

func (st SignatureType) String() string {
	switch st {
	case SignaturePoly1271:
		return "POLY_1271"
	default:
		return "UNKNOWN"
	}
}
