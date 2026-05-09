package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestDepositWalletStatusUsesOnchainCodeWhenRelayerFalse(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/deployed":
			_ = json.NewEncoder(w).Encode(map[string]any{"deployed": false})
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "7"})
		case "/":
			var body struct {
				Method string `json:"method"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode rpc request: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if body.Method != "eth_getCode" {
				t.Errorf("rpc method=%q want eth_getCode", body.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x60016000",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", server.URL)
	t.Setenv("POLYGON_RPC_URL", server.URL)
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
	if data["deployed"] != true || data["relayerDeployed"] != false || data["onchainCodeDeployed"] != true {
		t.Fatalf("deployment fields=%v", data)
	}
	if data["deploymentStatusSource"] != "polygon_code" {
		t.Fatalf("deploymentStatusSource=%v", data["deploymentStatusSource"])
	}
}

func TestDepositWalletDeploySkipsWalletCreateWhenCodeExists(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			var body struct {
				Method string `json:"method"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode rpc request: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if body.Method != "eth_getCode" {
				t.Errorf("rpc method=%q want eth_getCode", body.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x60016000",
			})
		case "/submit":
			t.Error("deploy should not submit WALLET-CREATE when bytecode exists")
			http.Error(w, "unexpected submit", http.StatusTeapot)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", server.URL)
	t.Setenv("POLYGON_RPC_URL", server.URL)

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "deploy", "--wait")
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
	if data["state"] != "already_deployed" || data["onchainCodeDeployed"] != true {
		t.Fatalf("deploy guard fields=%v", data)
	}
}

func TestDepositWalletApproveAdaptersDryRun(t *testing.T) {
	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "approve-adapters")
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
	calls, ok := data["calls"].([]any)
	if !ok || len(calls) != 4 {
		t.Fatalf("dry-run calls=%v want 4 entries", data["calls"])
	}
	adapters, ok := data["adapters"].([]any)
	if !ok || len(adapters) != 2 {
		t.Fatalf("adapters=%v want 2", data["adapters"])
	}
}

func TestDepositWalletApproveAdaptersRequiresConfirm(t *testing.T) {
	t.Setenv("POLYMARKET_PRIVATE_KEY", "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	_, stderr, err := executeRootForTest("--json", "deposit-wallet", "approve-adapters", "--submit")
	if err == nil {
		t.Fatalf("expected error when --submit is set without --confirm; stderr=%s", stderr)
	}
	if !strings.Contains(err.Error(), "APPROVE_ADAPTERS") {
		t.Fatalf("error must mention APPROVE_ADAPTERS confirm token: %v", err)
	}
}

func TestDepositWalletApproveAdaptersSubmitsBatch(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	var submitCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "42"})
		case "/submit":
			submitCalls++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode submit body: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if body["type"] != "WALLET" {
				t.Errorf("submit type=%v want WALLET", body["type"])
			}
			params, _ := body["depositWalletParams"].(map[string]any)
			calls, _ := params["calls"].([]any)
			if len(calls) != 4 {
				t.Errorf("submitted call count=%d want 4 (adapter approval batch)", len(calls))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"transactionID": "tx-adapter-approve",
				"state":         "STATE_NEW",
				"type":          "WALLET",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", server.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "approve-adapters", "--submit", "--confirm", "APPROVE_ADAPTERS")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	if submitCalls != 1 {
		t.Fatalf("relayer /submit called %d times, want 1", submitCalls)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false envelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v", err)
	}
	if data["transactionID"] != "tx-adapter-approve" {
		t.Fatalf("transactionID=%v", data["transactionID"])
	}
	if v, _ := data["approvals"].(float64); v != 4 {
		t.Fatalf("approvals=%v want 4", data["approvals"])
	}
}

