package polytypes

import (
	"encoding/json"
	"math/big"
	"strings"
	"time"
)

// --- Time handling ---

// NormalizedTime handles multiple API datetime formats from Gamma.
// Copied from polymarket-go-gamma-client/types.go.
type NormalizedTime time.Time

func (ct *NormalizedTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		*ct = NormalizedTime(time.Time{})
		return nil
	}
	if len(s) >= 3 {
		if strings.Contains(s, " ") || strings.Contains(s, "T") {
			last3 := s[len(s)-3:]
			if (last3[0] == '+' || last3[0] == '-') && last3[1] >= '0' && last3[1] <= '9' && last3[2] >= '0' && last3[2] <= '9' {
				if len(s) < 6 || s[len(s)-6:] != last3+":00" {
					s = s[:len(s)-3] + last3 + ":00"
				}
			}
		}
	}
	formats := []string{
		time.RFC3339, time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05+00:00",
		"2006-01-02",
		"January 2, 2006",
	}
	var err error
	var t time.Time
	for _, f := range formats {
		t, err = time.Parse(f, s)
		if err == nil {
			*ct = NormalizedTime(t)
			return nil
		}
	}
	return err
}

func (ct NormalizedTime) MarshalJSON() ([]byte, error) {
	t := time.Time(ct)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(time.RFC3339))
}

func (ct NormalizedTime) Time() time.Time     { return time.Time(ct) }
func (ct NormalizedTime) IsZero() bool         { return time.Time(ct).IsZero() }
func (ct NormalizedTime) String() string        { return ct.Time().String() }

// StringOrArray handles Gamma fields that may arrive as JSON strings or arrays.
// Copied from polymarket-go-gamma-client/types.go.
type StringOrArray []string

func (sa *StringOrArray) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*sa = StringOrArray([]string{})
		return nil
	}
	var arr []string
	if err := json.Unmarshal(b, &arr); err == nil {
		if arr == nil {
			*sa = StringOrArray([]string{})
			return nil
		}
		var flattened []string
		for _, s := range arr {
			if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
				var inner []string
				if json.Unmarshal([]byte(s), &inner) == nil {
					flattened = append(flattened, inner...)
					continue
				}
			}
			flattened = append(flattened, s)
		}
		*sa = StringOrArray(flattened)
		return nil
	}
	var arr2D [][]string
	if err := json.Unmarshal(b, &arr2D); err == nil {
		var flat []string
		for _, inner := range arr2D {
			flat = append(flat, inner...)
		}
		*sa = StringOrArray(flat)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "" {
			*sa = StringOrArray([]string{})
			return nil
		}
		if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
			var inner []string
			if json.Unmarshal([]byte(s), &inner) == nil {
				*sa = StringOrArray(inner)
				return nil
			}
		}
		*sa = StringOrArray([]string{s})
		return nil
	}
	*sa = StringOrArray([]string{})
	return nil
}

func (sa StringOrArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(sa))
}

// --- Decimal type ---

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

// --- Common enums (from rs-clob-client) ---

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

type SignatureType int

const (
	SignatureEOA        SignatureType = 0
	SignatureProxy      SignatureType = 1
	SignatureGnosisSafe SignatureType = 2
)

func (st SignatureType) String() string {
	switch st {
	case SignatureEOA:
		return "EOA"
	case SignatureProxy:
		return "PROXY"
	case SignatureGnosisSafe:
		return "SAFE"
	default:
		return "UNKNOWN"
	}
}
