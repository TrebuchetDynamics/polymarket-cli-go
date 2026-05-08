package types

import (
	"encoding/json"
	"strings"
	"time"
)

// NormalizedTime handles the date and timestamp formats returned by Gamma.
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

func (ct NormalizedTime) Time() time.Time { return time.Time(ct) }
func (ct NormalizedTime) IsZero() bool    { return time.Time(ct).IsZero() }
func (ct NormalizedTime) String() string  { return ct.Time().String() }

// StringOrArray handles Gamma fields that may arrive as a string, JSON string,
// one-dimensional array, or nested array.
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
