package errors

import (
	goerrors "errors"
	"fmt"
	"testing"
)

func TestNew_ReturnsStructuredError(t *testing.T) {
	err := New(CodeMissingField, "field is required")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != CodeMissingField {
		t.Fatalf("code=%q", err.Code)
	}
}

func TestNewf_FormatsMessage(t *testing.T) {
	err := Newf(CodeInvalidValue, "value %d is invalid", 42)
	if err.Message != "value 42 is invalid" {
		t.Fatalf("message=%q", err.Message)
	}
}

func TestWrap_WrapsUnderlying(t *testing.T) {
	base := fmt.Errorf("underlying error")
	err := Wrap(CodeConnectionFailed, "connection lost", base)
	if !goerrors.Is(err, base) {
		t.Fatal("errors.Is should find underlying")
	}
}

func TestWithHTTP_SetsHTTPCode(t *testing.T) {
	err := WithHTTP(CodeUnauthorized, "not allowed", 401)
	if err.HTTPCode != 401 {
		t.Fatalf("HTTPCode=%d", err.HTTPCode)
	}
}

func TestCodeConstants_AreUnique(t *testing.T) {
	seen := map[Code]bool{}
	for _, c := range []Code{
		CodeTimeout, CodeConnectionFailed, CodeRateLimited, CodeCircuitOpen,
		CodeMissingSigner, CodeMissingCreds, CodeInvalidSignature, CodeUnauthorized,
		CodeOrderNotFound, CodeInsufficientFunds, CodeInvalidOrder, CodeInvalidTokenID,
		CodeMissingField, CodeInvalidValue, CodeBatchSizeExceed,
		CodeLiveDisabled, CodePreflightFailed, CodeNotAuthorized,
		CodeMarketNotFound, CodeEventNotFound,
	} {
		if seen[c] {
			t.Fatalf("duplicate code: %s", c)
		}
		seen[c] = true
	}
	if len(seen) != 20 {
		t.Fatalf("expected 20 codes, got %d", len(seen))
	}
}

func TestError_ErrorFormat(t *testing.T) {
	err := New(CodeTimeout, "request timed out")
	s := err.Error()
	if s == "" {
		t.Fatal("Error() returned empty string")
	}
}
