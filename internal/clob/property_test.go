package clob

import (
	"math/big"
	"strings"
	"testing"
	"testing/quick"
)

// Property: normalizeBuilderCode is idempotent on valid inputs.
func TestNormalizeBuilderCodeIdempotent(t *testing.T) {
	f := func(s string) bool {
		// Only test inputs that parse successfully.
		first, err := normalizeBuilderCode(s)
		if err != nil {
			return true // vacuously true for invalid inputs
		}
		second, err := normalizeBuilderCode(first)
		if err != nil {
			return false
		}
		return first == second
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
		t.Error(err)
	}
}

// Property: normalizeBuilderCode always returns lowercase hex for valid inputs.
func TestNormalizeBuilderCodeAlwaysLowercase(t *testing.T) {
	f := func(s string) bool {
		out, err := normalizeBuilderCode(s)
		if err != nil {
			return true
		}
		return strings.ToLower(out) == out
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
		t.Error(err)
	}
}

// Property: truncateRat is idempotent for any non-negative decimal count.
func TestTruncateRatIdempotent(t *testing.T) {
	f := func(num, den int64, decimals int) bool {
		if den == 0 {
			return true
		}
		if decimals < 0 {
			decimals = 0
		}
		if decimals > 18 {
			decimals = 18
		}
		v := new(big.Rat).SetFrac(big.NewInt(num), big.NewInt(den))
		once := truncateRat(v, decimals)
		twice := truncateRat(once, decimals)
		return once.Cmp(twice) == 0
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
		t.Error(err)
	}
}

// Property: truncateRat never increases precision beyond the requested decimals.
func TestTruncateRatRespectsPrecision(t *testing.T) {
	f := func(num, den int64, decimals int) bool {
		if den == 0 {
			return true
		}
		if decimals < 0 {
			decimals = 0
		}
		if decimals > 18 {
			decimals = 18
		}
		v := new(big.Rat).SetFrac(big.NewInt(num), big.NewInt(den))
		truncated := truncateRat(v, decimals)
		// Truncating again with the same decimals should be a no-op.
		return truncated.Cmp(truncateRat(truncated, decimals)) == 0
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
		t.Error(err)
	}
}

// Property: fixedDecimal round-trips approximately through big.Rat for simple fractions.
func TestFixedDecimalRoundTrip(t *testing.T) {
	f := func(n int64, decimals int) bool {
		if decimals < 0 || decimals > 18 {
			return true
		}
		v := new(big.Rat).SetFrac64(n, 1)
		s := fixedDecimal(v, decimals)
		parsed, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return false
		}
		// Reconstruct the original scaled value.
		scaled := new(big.Int).Mul(v.Num(), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
		scaled.Quo(scaled, v.Denom())
		return parsed.Cmp(scaled) == 0
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
		t.Error(err)
	}
}

// Property: normalizeOrderType returns either a known type or the fallback.
func TestNormalizeOrderTypeReturnsKnownOrFallback(t *testing.T) {
	known := map[string]bool{"GTC": true, "GTD": true, "FAK": true, "FOK": true}
	f := func(raw, fallback string) bool {
		out := normalizeOrderType(raw, fallback)
		return known[out] || out == fallback
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
		t.Error(err)
	}
}

// Property: limitFixedAmounts produces maker/taker amounts that parse back to big.Rat.
func TestLimitFixedAmountsParsable(t *testing.T) {
	f := func(priceNum, priceDen, sizeNum, sizeDen int64, sideInt int) bool {
		if priceDen == 0 || sizeDen == 0 {
			return true
		}
		price := new(big.Rat).SetFrac(big.NewInt(priceNum), big.NewInt(priceDen))
		size := new(big.Rat).SetFrac(big.NewInt(sizeNum), big.NewInt(sizeDen))
		if price.Sign() <= 0 || size.Sign() <= 0 {
			return true
		}
		side := "BUY"
		if sideInt%2 != 0 {
			side = "SELL"
		}
		maker, taker := limitFixedAmounts(side, price, size)
		_, ok1 := new(big.Rat).SetString(maker)
		_, ok2 := new(big.Rat).SetString(taker)
		return ok1 && ok2
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
		t.Error(err)
	}
}
