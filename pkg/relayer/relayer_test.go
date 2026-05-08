package relayer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

const testBuilderKey = "f33935da-21f0-179b-925a-cf6773d1eada"

// secret is 32 zero bytes base64 — not a real Polymarket key, just a
// well-formed input so the HMAC code doesn't reject the config.
const testBuilderSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

const testBuilderPass = "passphrase-1"

func validBuilderConfig() BuilderConfig {
	return BuilderConfig{
		Key:        testBuilderKey,
		Secret:     testBuilderSecret,
		Passphrase: testBuilderPass,
	}
}

func TestNewRequiresBuilderConfig(t *testing.T) {
	if _, err := New("", BuilderConfig{}, 137); err == nil {
		t.Fatal("expected error for empty BuilderConfig")
	}
	if _, err := New("https://example.test", validBuilderConfig(), 137); err != nil {
		t.Fatalf("unexpected error for valid config: %v", err)
	}
}

func TestNewSigner(t *testing.T) {
	signer, err := NewSigner(testPrivateKey, 0)
	if err != nil {
		t.Fatalf("NewSigner error: %v", err)
	}
	if signer == nil {
		t.Fatal("signer is nil")
	}
	if !strings.HasPrefix(signer.Address(), "0x") {
		t.Errorf("signer address malformed: %q", signer.Address())
	}
	if signer.ChainID() != 137 {
		t.Errorf("expected default chainID 137, got %d", signer.ChainID())
	}
}

func TestBuildApprovalCallsReturnsSix(t *testing.T) {
	calls := BuildApprovalCalls()
	if len(calls) != 6 {
		t.Fatalf("expected 6 approval calls, got %d", len(calls))
	}
	for i, c := range calls {
		if c.Value != "0" {
			t.Errorf("call %d: expected value=0, got %q", i, c.Value)
		}
		if !strings.HasPrefix(c.Data, "0x") {
			t.Errorf("call %d: data missing 0x prefix: %q", i, c.Data)
		}
	}
}

func TestBuildDeadlineDefault(t *testing.T) {
	d := BuildDeadline(0)
	if d == "" {
		t.Fatal("expected non-empty deadline default")
	}
}

func TestSignWalletBatchProducesHexSignature(t *testing.T) {
	signer, err := NewSigner(testPrivateKey, 137)
	if err != nil {
		t.Fatalf("NewSigner error: %v", err)
	}
	calls := BuildApprovalCalls()
	sig, err := SignWalletBatch(signer, "0x1234567890123456789012345678901234567890", "0", "999999999", calls)
	if err != nil {
		t.Fatalf("SignWalletBatch error: %v", err)
	}
	if !strings.HasPrefix(sig, "0x") || len(sig) != 132 { // 0x + 130 hex = 65 bytes
		t.Errorf("expected 0x-prefixed 65-byte signature, got len=%d", len(sig))
	}
}

func TestClientGetNonceAndDeployedRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Every relayer request must carry the POLY_BUILDER_* set.
		for _, h := range []string{"POLY_BUILDER_API_KEY", "POLY_BUILDER_PASSPHRASE", "POLY_BUILDER_TIMESTAMP", "POLY_BUILDER_SIGNATURE"} {
			if r.Header.Get(h) == "" {
				t.Errorf("missing required header %s", h)
			}
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/nonce"):
			json.NewEncoder(w).Encode(NonceResponse{Nonce: "7"})
		case strings.HasPrefix(r.URL.Path, "/deployed"):
			json.NewEncoder(w).Encode(DeployedResponse{Deployed: true, Address: "0xfeed"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, err := New(srv.URL, validBuilderConfig(), 137)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	nonce, err := c.GetNonce(context.Background(), "0xb72dbe5d44c1b549351bef276ba48a1cca5df662")
	if err != nil {
		t.Fatalf("GetNonce error: %v", err)
	}
	if nonce != "7" {
		t.Errorf("expected nonce 7, got %q", nonce)
	}
	deployed, err := c.IsDeployed(context.Background(), "0xb72dbe5d44c1b549351bef276ba48a1cca5df662")
	if err != nil {
		t.Fatalf("IsDeployed error: %v", err)
	}
	if !deployed {
		t.Error("expected deployed=true")
	}
}

func TestStateConstantsLifecycle(t *testing.T) {
	if !StateMined.IsTerminal() {
		t.Error("StateMined should be terminal")
	}
	if !StateMined.IsSuccess() {
		t.Error("StateMined should be successful")
	}
	if StateNew.IsTerminal() {
		t.Error("StateNew should not be terminal")
	}
	if StateFailed.IsSuccess() {
		t.Error("StateFailed should not be successful")
	}
}
