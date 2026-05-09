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

// --- redeemable + redeem tests ---

const redeemTestPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

// dataAPIWithOnePosition serves a single redeemable=true position fixture.
func dataAPIWithOnePosition(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"asset":        "10203228750887270363579341300435494148775390248158812958841180330451031762744",
			"conditionId":  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"size":         4.0784,
			"avgPrice":     0.5099,
			"redeemable":   true,
			"mergeable":    false,
			"negativeRisk": false,
			"outcome":      "Up",
			"slug":         "eth-updown-5m-1778316000",
			"title":        "Ethereum Up or Down - May 9, 4:40AM-4:45AM ET",
		}})
	}))
}

func TestDepositWalletRedeemableJSON(t *testing.T) {
	dataSrv := dataAPIWithOnePosition(t)
	defer dataSrv.Close()
	t.Setenv("POLYMARKET_PRIVATE_KEY", redeemTestPrivateKey)
	t.Setenv("POLYMARKET_DATA_API_URL", dataSrv.URL)

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "redeemable")
	if err != nil {
		t.Fatalf("Execute error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false: %s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatal(err)
	}
	if v, _ := data["count"].(float64); v != 1 {
		t.Fatalf("count=%v want 1; data=%v", data["count"], data)
	}
	if data["depositWallet"] == nil {
		t.Fatal("depositWallet missing")
	}
}

func TestDepositWalletRedeemDryRunPrintsCalls(t *testing.T) {
	dataSrv := dataAPIWithOnePosition(t)
	defer dataSrv.Close()
	t.Setenv("POLYMARKET_PRIVATE_KEY", redeemTestPrivateKey)
	t.Setenv("POLYMARKET_DATA_API_URL", dataSrv.URL)

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "redeem")
	if err != nil {
		t.Fatalf("Execute error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false: %s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatal(err)
	}
	calls, ok := data["calls"].([]any)
	if !ok || len(calls) != 1 {
		t.Fatalf("dry-run calls=%v want 1 entry", data["calls"])
	}
	note, _ := data["note"].(string)
	if !strings.Contains(note, "REDEEM_WINNERS") {
		t.Errorf("note must mention REDEEM_WINNERS confirm token: %q", note)
	}
}

func TestDepositWalletRedeemRequiresConfirm(t *testing.T) {
	dataSrv := dataAPIWithOnePosition(t)
	defer dataSrv.Close()
	t.Setenv("POLYMARKET_PRIVATE_KEY", redeemTestPrivateKey)
	t.Setenv("POLYMARKET_DATA_API_URL", dataSrv.URL)

	_, stderr, err := executeRootForTest("--json", "deposit-wallet", "redeem", "--submit")
	if err == nil {
		t.Fatalf("expected error when --submit set without --confirm; stderr=%s", stderr)
	}
	if !strings.Contains(err.Error(), "REDEEM_WINNERS") {
		t.Fatalf("error must mention REDEEM_WINNERS: %v", err)
	}
}

func TestDepositWalletRedeemRefusesWithoutAdapterApproval(t *testing.T) {
	dataSrv := dataAPIWithOnePosition(t)
	defer dataSrv.Close()
	// RPC server returns isApprovedForAll=false.
	rpcSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body.Method == "eth_chainId" {
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x89"})
			return
		}
		// Any eth_call → return false (32 bytes of zero).
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0", "id": 1,
			"result": "0x0000000000000000000000000000000000000000000000000000000000000000",
		})
	}))
	defer rpcSrv.Close()
	relaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/submit" {
			t.Error("relayer /submit must not be called when adapter approval is missing")
		}
		http.NotFound(w, r)
	}))
	defer relaySrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", redeemTestPrivateKey)
	t.Setenv("POLYMARKET_DATA_API_URL", dataSrv.URL)
	t.Setenv("POLYMARKET_RELAYER_URL", relaySrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")
	t.Setenv("POLYGON_RPC_URL", rpcSrv.URL)

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "redeem", "--submit", "--confirm", "REDEEM_WINNERS", "--rpc-url", rpcSrv.URL)
	if err != nil {
		t.Fatalf("Execute error: %v\nstderr:\n%s", err, stderr)
	}
	// The CLI returns ok envelope with an inner ok=false JSON; assert
	// that the body called out the missing approval.
	if !strings.Contains(stdout, "missingApprovals") || !strings.Contains(stdout, "approve-adapters") {
		t.Fatalf("redeem must point to approve-adapters when isApprovedForAll=false; stdout=%s", stdout)
	}
}

func TestDepositWalletRedeemHappyPathSubmits(t *testing.T) {
	dataSrv := dataAPIWithOnePosition(t)
	defer dataSrv.Close()
	rpcSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Method == "eth_chainId" {
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x89"})
			return
		}
		// eth_call → return approved=true.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0", "id": 1,
			"result": "0x0000000000000000000000000000000000000000000000000000000000000001",
		})
	}))
	defer rpcSrv.Close()

	var submitCalls int
	relaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "9"})
		case "/submit":
			submitCalls++
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["type"] != "WALLET" {
				t.Errorf("submit type=%v want WALLET", body["type"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"transactionID": "redeem-tx-1",
				"state":         "STATE_NEW",
				"type":          "WALLET",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer relaySrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", redeemTestPrivateKey)
	t.Setenv("POLYMARKET_DATA_API_URL", dataSrv.URL)
	t.Setenv("POLYMARKET_RELAYER_URL", relaySrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")
	t.Setenv("POLYGON_RPC_URL", rpcSrv.URL)

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "redeem", "--submit", "--confirm", "REDEEM_WINNERS", "--rpc-url", rpcSrv.URL)
	if err != nil {
		t.Fatalf("Execute error: %v\nstderr:\n%s", err, stderr)
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
		t.Fatal(err)
	}
	if data["transactionID"] != "redeem-tx-1" {
		t.Fatalf("transactionID=%v", data["transactionID"])
	}
}

