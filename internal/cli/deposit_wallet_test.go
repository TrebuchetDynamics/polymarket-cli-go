package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// relayerClientFromEnv prefers the V2 RELAYER_API_KEY scheme when both
// RELAYER_API_KEY and RELAYER_API_KEY_ADDRESS are set. The legacy
// POLY_BUILDER_* HMAC scheme is the fallback.

func TestRelayerClientFromEnvPrefersV2(t *testing.T) {
	t.Setenv("POLYMARKET_RELAYER_URL", "https://relayer-v2.example.com")
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")
	// Clear legacy creds.
	t.Setenv("POLYMARKET_BUILDER_API_KEY", "")
	t.Setenv("POLYMARKET_BUILDER_SECRET", "")
	t.Setenv("POLYMARKET_BUILDER_PASSPHRASE", "")
	t.Setenv("BUILDER_API_KEY", "")
	t.Setenv("BUILDER_SECRET", "")
	t.Setenv("BUILDER_PASS_PHRASE", "")

	rc, err := relayerClientFromEnv()
	if err != nil {
		t.Fatalf("relayerClientFromEnv returned error with V2 creds set: %v", err)
	}
	if rc == nil {
		t.Fatal("returned nil client")
	}
}

func TestRelayerClientFromEnvFallsBackToLegacy(t *testing.T) {
	t.Setenv("POLYMARKET_RELAYER_URL", "https://relayer-v2.example.com")
	t.Setenv("RELAYER_API_KEY", "")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "")
	t.Setenv("POLYMARKET_BUILDER_API_KEY", "legacy-key")
	t.Setenv("POLYMARKET_BUILDER_SECRET", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	t.Setenv("POLYMARKET_BUILDER_PASSPHRASE", "legacy-pass")

	rc, err := relayerClientFromEnv()
	if err != nil {
		t.Fatalf("relayerClientFromEnv with legacy creds: %v", err)
	}
	if rc == nil {
		t.Fatal("returned nil client")
	}
}

func TestRelayerClientFromEnvErrorsWhenNoCredsAtAll(t *testing.T) {
	t.Setenv("RELAYER_API_KEY", "")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "")
	t.Setenv("POLYMARKET_BUILDER_API_KEY", "")
	t.Setenv("POLYMARKET_BUILDER_SECRET", "")
	t.Setenv("POLYMARKET_BUILDER_PASSPHRASE", "")
	t.Setenv("BUILDER_API_KEY", "")
	t.Setenv("BUILDER_SECRET", "")
	t.Setenv("BUILDER_PASS_PHRASE", "")

	if _, err := relayerClientFromEnv(); err == nil {
		t.Fatal("expected error when no creds are set")
	}
}

func TestRelayerClientFromEnvIgnoresPartialV2(t *testing.T) {
	// Only RELAYER_API_KEY set without ADDRESS — should fall through to
	// legacy. (Legacy creds also missing here, so we expect an error.)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "")
	t.Setenv("POLYMARKET_BUILDER_API_KEY", "")
	t.Setenv("POLYMARKET_BUILDER_SECRET", "")
	t.Setenv("POLYMARKET_BUILDER_PASSPHRASE", "")
	t.Setenv("BUILDER_API_KEY", "")
	t.Setenv("BUILDER_SECRET", "")
	t.Setenv("BUILDER_PASS_PHRASE", "")

	if _, err := relayerClientFromEnv(); err == nil {
		t.Fatal("expected error when V2 is partial and legacy is missing")
	}
}

func TestDepositWalletStatusOutputsWalletNonce(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/deployed":
			_ = json.NewEncoder(w).Encode(map[string]any{"deployed": true})
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "7"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", server.URL)
	t.Setenv("RELAYER_API_KEY", "relayer-key")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "status")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false, want true\nenvelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v\n%s", err, got.Data)
	}
	if data["walletNonce"] != "7" {
		t.Fatalf("walletNonce=%v want 7; data=%v", data["walletNonce"], data)
	}
	if _, ok := data["wallerNonce"]; ok {
		t.Fatalf("deprecated wallerNonce key present: %v", data)
	}
}
