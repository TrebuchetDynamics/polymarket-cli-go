package marketdiscovery

import (
	"testing"
)

func TestExtractTokenIDs_ValidJSON(t *testing.T) {
	result := extractTokenIDs(`["123","456"]`)
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0] != "123" || result[1] != "456" {
		t.Fatalf("got %v", result)
	}
}

func TestExtractTokenIDs_EmptyArray(t *testing.T) {
	result := extractTokenIDs("[]")
	if len(result) != 0 {
		t.Fatalf("expected 0 tokens, got %d", len(result))
	}
}

func TestExtractTokenIDs_EmptyString(t *testing.T) {
	result := extractTokenIDs("")
	if len(result) != 0 {
		t.Fatalf("expected 0 tokens, got %d", len(result))
	}
}

func TestSplitQuoted_TwoValues(t *testing.T) {
	result := splitQuoted(`"123","456"`)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(result), result)
	}
	if result[0] != "123" || result[1] != "456" {
		t.Fatalf("got %v", result)
	}
}

func TestSplitQuoted_Empty(t *testing.T) {
	result := splitQuoted("")
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestTrimPrefix_Matches(t *testing.T) {
	result := trimPrefix("[123]", "[")
	if result != "123]" {
		t.Fatalf("got %q", result)
	}
}

func TestTrimPrefix_NoMatch(t *testing.T) {
	result := trimPrefix("hello", "[")
	if result != "hello" {
		t.Fatalf("got %q", result)
	}
}

func TestTrimSuffix_Matches(t *testing.T) {
	result := trimSuffix("[123]", "]")
	if result != "[123" {
		t.Fatalf("got %q", result)
	}
}
