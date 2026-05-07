package polytypes

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNormalizedTimeRFC3339(t *testing.T) {
	var nt NormalizedTime
	if err := json.Unmarshal([]byte(`"2024-01-15T10:30:00Z"`), &nt); err != nil {
		t.Fatal(err)
	}
	if nt.Time().Year() != 2024 || nt.Time().Month() != 1 || nt.Time().Day() != 15 {
		t.Fatalf("unexpected date: %s", nt.Time())
	}
}

func TestNormalizedTimeSpaceSeparated(t *testing.T) {
	var nt NormalizedTime
	if err := json.Unmarshal([]byte(`"2020-11-02 16:31:01+00"`), &nt); err != nil {
		t.Fatal(err)
	}
	if nt.IsZero() {
		t.Fatal("expected non-zero time")
	}
}

func TestNormalizedTimeDateOnly(t *testing.T) {
	var nt NormalizedTime
	if err := json.Unmarshal([]byte(`"2024-06-15"`), &nt); err != nil {
		t.Fatal(err)
	}
	if nt.IsZero() {
		t.Fatal("expected non-zero time")
	}
}

func TestNormalizedTimeNull(t *testing.T) {
	var nt NormalizedTime
	if err := json.Unmarshal([]byte(`null`), &nt); err != nil {
		t.Fatal(err)
	}
	if !nt.IsZero() {
		t.Fatal("expected zero time for null")
	}
}

func TestNormalizedTimeEmpty(t *testing.T) {
	var nt NormalizedTime
	if err := json.Unmarshal([]byte(`""`), &nt); err != nil {
		t.Fatal(err)
	}
	if !nt.IsZero() {
		t.Fatal("expected zero time for empty string")
	}
}

func TestNormalizedTimeLongMonth(t *testing.T) {
	var nt NormalizedTime
	if err := json.Unmarshal([]byte(`"November 1, 2022"`), &nt); err != nil {
		t.Fatal(err)
	}
	if nt.Time().Year() != 2022 || nt.Time().Month() != 11 {
		t.Fatalf("unexpected: %s", nt.Time())
	}
}

func TestNormalizedTimeMarshal(t *testing.T) {
	nt := NormalizedTime(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	b, err := json.Marshal(nt)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"2024-01-15T10:30:00Z"` {
		t.Fatalf("unexpected marshal: %s", b)
	}
}

func TestNormalizedTimeMarshalZero(t *testing.T) {
	var nt NormalizedTime
	b, err := json.Marshal(nt)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "null" {
		t.Fatalf("expected null for zero time: %s", b)
	}
}

func TestStringOrArrayFromArray(t *testing.T) {
	var sa StringOrArray
	if err := json.Unmarshal([]byte(`["Yes","No"]`), &sa); err != nil {
		t.Fatal(err)
	}
	if len(sa) != 2 || sa[0] != "Yes" || sa[1] != "No" {
		t.Fatalf("unexpected: %v", sa)
	}
}

func TestStringOrArrayFromString(t *testing.T) {
	var sa StringOrArray
	if err := json.Unmarshal([]byte(`"single"`), &sa); err != nil {
		t.Fatal(err)
	}
	if len(sa) != 1 || sa[0] != "single" {
		t.Fatalf("unexpected: %v", sa)
	}
}

func TestStringOrArrayFromJSONString(t *testing.T) {
	var sa StringOrArray
	if err := json.Unmarshal([]byte(`"[\"Up\",\"Down\"]"`), &sa); err != nil {
		t.Fatal(err)
	}
	if len(sa) != 2 || sa[0] != "Up" || sa[1] != "Down" {
		t.Fatalf("unexpected: %v", sa)
	}
}

func TestStringOrArrayNull(t *testing.T) {
	var sa StringOrArray
	if err := json.Unmarshal([]byte(`null`), &sa); err != nil {
		t.Fatal(err)
	}
	if len(sa) != 0 {
		t.Fatalf("expected empty array for null: %v", sa)
	}
}

func TestDecimalFromString(t *testing.T) {
	d, err := NewDecimal("0.52")
	if err != nil {
		t.Fatal(err)
	}
	if d.String() != "0.520000" {
		t.Fatalf("String() = %q", d.String())
	}
	if d.Float64() < 0.51 || d.Float64() > 0.53 {
		t.Fatalf("Float64() = %f", d.Float64())
	}
}

func TestDecimalFromInt(t *testing.T) {
	d := DecimalFromInt64(100)
	if d.String() != "100.000000" {
		t.Fatalf("String() = %q", d.String())
	}
}

func TestDecimalIsZero(t *testing.T) {
	d := DecimalFromInt64(0)
	if !d.IsZero() {
		t.Fatal("expected zero")
	}
	d2 := MustDecimal("0.01")
	if d2.IsZero() {
		t.Fatal("expected non-zero")
	}
}

func TestDecimalMarshalJSON(t *testing.T) {
	d := MustDecimal("0.555")
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"0.555000"` {
		t.Fatalf("marshal: %s", b)
	}
}

func TestDecimalUnmarshalJSON(t *testing.T) {
	var d Decimal
	if err := json.Unmarshal([]byte(`"0.99"`), &d); err != nil {
		t.Fatal(err)
	}
	if d.Float64() < 0.98 || d.Float64() > 1.0 {
		t.Fatalf("unmarshal: %f", d.Float64())
	}
}

func TestSideEnum(t *testing.T) {
	if SideBuy.String() != "BUY" {
		t.Fatalf("SideBuy = %s", SideBuy)
	}
	if SideSell.String() != "SELL" {
		t.Fatalf("SideSell = %s", SideSell)
	}
}

func TestSignatureTypeEnum(t *testing.T) {
	if SignatureEOA.String() != "EOA" {
		t.Fatalf("EOA = %s", SignatureEOA)
	}
	if SignatureGnosisSafe.String() != "SAFE" {
		t.Fatalf("SAFE = %s", SignatureGnosisSafe)
	}
}
