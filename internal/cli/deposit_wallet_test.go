package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
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
	t.Setenv("POLYGOLEM_RELAYER_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))
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
	t.Setenv("POLYGOLEM_RELAYER_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))
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
	t.Setenv("POLYGOLEM_RELAYER_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))
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

func TestRelayerClientFromEnvLoadsRelayerKeyFromEnvFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env.relayer-v2")
	if err := os.WriteFile(envFile, []byte("RELAYER_API_KEY=file-key\nRELAYER_API_KEY_ADDRESS=0xabc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("POLYGOLEM_RELAYER_ENV_FILE", envFile)
	t.Setenv("RELAYER_API_KEY", "")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "")
	t.Setenv("POLYMARKET_BUILDER_API_KEY", "")
	t.Setenv("POLYMARKET_BUILDER_SECRET", "")
	t.Setenv("POLYMARKET_BUILDER_PASSPHRASE", "")

	rc, err := relayerClientFromEnv()
	if err != nil {
		t.Fatalf("relayerClientFromEnv with env file: %v", err)
	}
	if rc == nil {
		t.Fatal("returned nil client")
	}
}

func TestParsePUSDAmountUsesExactSixDecimalBaseUnits(t *testing.T) {
	tests := map[string]string{
		"3.053937":  "3053937",
		"3.0539370": "3053937",
		"0.000001":  "1",
		"42":        "42000000",
	}
	for input, want := range tests {
		got, err := parsePUSDAmount(input)
		if err != nil {
			t.Fatalf("parsePUSDAmount(%q): %v", input, err)
		}
		if got.String() != want {
			t.Fatalf("parsePUSDAmount(%q)=%s, want %s", input, got, want)
		}
	}
}

func TestParsePUSDAmountRejectsTooManyNonZeroDecimals(t *testing.T) {
	if _, err := parsePUSDAmount("0.0000001"); err == nil {
		t.Fatal("expected error for sub-micro pUSD amount")
	}
}

func mustReadFileForTest(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
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

func TestDepositWalletOnboardIncludesEnableTradingSigns(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	submitCallCounts := make([]int, 0, 2)
	nonce := 40
	relayerSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			nonce++
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": fmt.Sprintf("%d", nonce)})
		case "/submit":
			var body struct {
				DepositWalletParams struct {
					Calls []map[string]string `json:"calls"`
				} `json:"depositWalletParams"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode submit body: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			submitCallCounts = append(submitCallCounts, len(body.DepositWalletParams.Calls))
			switch len(submitCallCounts) {
			case 1:
				if len(body.DepositWalletParams.Calls) != 10 {
					t.Errorf("first batch call count=%d want 10", len(body.DepositWalletParams.Calls))
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"transactionID": "tx-standard-approve", "state": "STATE_NEW"})
			case 2:
				if len(body.DepositWalletParams.Calls) != 2 {
					t.Errorf("second batch call count=%d want 2", len(body.DepositWalletParams.Calls))
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"transactionID": "tx-enable-trading", "state": "STATE_NEW"})
			default:
				t.Errorf("unexpected extra submit call %d", len(submitCallCounts))
				http.Error(w, "unexpected submit", http.StatusTeapot)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer relayerSrv.Close()

	clobSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/api-key":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "clob-key",
				"secret":     "clob-secret",
				"passphrase": "clob-pass",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer clobSrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", relayerSrv.URL)
	t.Setenv("POLYMARKET_CLOB_URL", clobSrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "onboard", "--skip-deploy")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false envelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v\n%s", err, got.Data)
	}
	enableTrading, ok := data["enableTrading"].(map[string]any)
	if !ok {
		t.Fatalf("enableTrading result missing: %v", data)
	}
	if enableTrading["clobAuthSigned"] != true || enableTrading["apiKeysCreatedOrDerived"] != true || enableTrading["tokenApprovalsSigned"] != true || enableTrading["tokenApprovalsSubmitted"] != true {
		t.Fatalf("enableTrading flags unexpected: %v", enableTrading)
	}
	if enableTrading["callCount"] != float64(2) {
		t.Fatalf("enableTrading callCount=%v want 2", enableTrading["callCount"])
	}
	if len(submitCallCounts) != 2 || submitCallCounts[0] != 10 || submitCallCounts[1] != 2 {
		t.Fatalf("submit call counts=%v want [10 2]", submitCallCounts)
	}
}

func TestDepositWalletEnableTradingSubmitsSigns(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	var submitCalls int
	relayerSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "51"})
		case "/submit":
			submitCalls++
			var body struct {
				DepositWalletParams struct {
					Calls []map[string]string `json:"calls"`
				} `json:"depositWalletParams"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode submit body: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if len(body.DepositWalletParams.Calls) != 2 {
				t.Errorf("enable-trading call count=%d want 2", len(body.DepositWalletParams.Calls))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"transactionID": "tx-enable-trading", "state": "STATE_NEW"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer relayerSrv.Close()

	clobSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/api-key":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "clob-key",
				"secret":     "clob-secret",
				"passphrase": "clob-pass",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer clobSrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", relayerSrv.URL)
	t.Setenv("POLYMARKET_CLOB_URL", clobSrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "enable-trading")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false envelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v\n%s", err, got.Data)
	}
	if data["clobAuthSigned"] != true || data["apiKeysCreatedOrDerived"] != true || data["tokenApprovalsSubmitted"] != true {
		t.Fatalf("enable-trading flags unexpected: %v", data)
	}
	if data["callCount"] != float64(2) {
		t.Fatalf("callCount=%v want 2", data["callCount"])
	}
	if submitCalls != 1 {
		t.Fatalf("submitCalls=%d want 1", submitCalls)
	}
}

func TestDepositWalletEnableTradingDryRunDoesNotSubmit(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	var submitCalls int
	relayerSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "52"})
		case "/submit":
			submitCalls++
			http.Error(w, "unexpected submit", http.StatusTeapot)
		default:
			http.NotFound(w, r)
		}
	}))
	defer relayerSrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", relayerSrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "enable-trading", "--dry-run")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false envelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v\n%s", err, got.Data)
	}
	if data["dryRun"] != true || data["wouldSubmitApprovals"] != true || data["approvalBatchSignable"] != true {
		t.Fatalf("dry-run data unexpected: %v", data)
	}
	if submitCalls != 0 {
		t.Fatalf("submitCalls=%d want 0", submitCalls)
	}
}

func TestDepositWalletEnableTradingAutoMintsRelayerKeyWhenMissing(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	envFile := filepath.Join(t.TempDir(), ".env.relayer-v2")
	var authMinted bool
	var submitCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			if r.URL.Query().Get("type") == "WALLET" {
				_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "61"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"nonce": "siwe-nonce"})
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "polymarket_session", Value: "session"})
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/profiles":
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "profile-1", "proxyWallet": "0xproxy"})
		case "/relayer/api/auth":
			authMinted = true
			_ = json.NewEncoder(w).Encode(map[string]string{
				"apiKey":    "auto-relayer-key",
				"address":   "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23",
				"createdAt": "2026-05-10T00:00:00Z",
			})
		case "/submit":
			submitCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{"transactionID": "tx-enable-trading", "state": "STATE_NEW"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	clobSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/api-key":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "clob-key",
				"secret":     "clob-secret",
				"passphrase": "clob-pass",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer clobSrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYGOLEM_RELAYER_ENV_FILE", envFile)
	t.Setenv("POLYMARKET_GAMMA_URL", server.URL)
	t.Setenv("POLYMARKET_RELAYER_URL", server.URL)
	t.Setenv("POLYMARKET_CLOB_URL", clobSrv.URL)
	t.Setenv("RELAYER_API_KEY", "")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "")
	t.Setenv("POLYMARKET_BUILDER_API_KEY", "")
	t.Setenv("POLYMARKET_BUILDER_SECRET", "")
	t.Setenv("POLYMARKET_BUILDER_PASSPHRASE", "")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "enable-trading")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false envelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v\n%s", err, got.Data)
	}
	if !authMinted {
		t.Fatal("expected automatic relayer auth mint")
	}
	if submitCalls != 1 {
		t.Fatalf("submitCalls=%d want 1", submitCalls)
	}
	if !strings.Contains(mustReadFileForTest(t, envFile), "RELAYER_API_KEY=auto-relayer-key") {
		t.Fatalf("auto-minted relayer key was not persisted to %s", envFile)
	}
}

func TestDepositWalletStatusCheckEnableTradingValidatesSignaturesAndAllowances(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	relayerSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/deployed":
			_ = json.NewEncoder(w).Encode(map[string]any{"deployed": true})
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "9"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer relayerSrv.Close()

	rpcSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode rpc body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body.Method != "eth_call" {
			t.Errorf("method=%q want eth_call", body.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x" + strings.Repeat("f", 64),
		})
	}))
	defer rpcSrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", relayerSrv.URL)
	t.Setenv("POLYGON_RPC_URL", rpcSrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")
	t.Setenv("POLYMARKET_CLOB_API_KEY", "configured-clob-key")
	t.Setenv("POLYMARKET_CLOB_SECRET", "configured-clob-secret")
	t.Setenv("POLYMARKET_CLOB_PASSPHRASE", "configured-clob-pass")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "status", "--check-enable-trading")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false envelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v\n%s", err, got.Data)
	}
	enableTrading, ok := data["enableTrading"].(map[string]any)
	if !ok {
		t.Fatalf("enableTrading validation missing: %v", data)
	}
	if enableTrading["clobAuthSignable"] != true || enableTrading["approvalBatchSignable"] != true || enableTrading["tokenApprovalsReady"] != true || enableTrading["ready"] != true {
		t.Fatalf("enableTrading validation unexpected: %v", enableTrading)
	}
	checks, ok := enableTrading["approvalChecks"].([]any)
	if !ok || len(checks) != 2 {
		t.Fatalf("approvalChecks=%v want 2 checks", enableTrading["approvalChecks"])
	}
}

func TestDepositWalletStatusCheckEnableTradingTreatsDerivableCLOBKeyAsReady(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	relayerSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/deployed":
			_ = json.NewEncoder(w).Encode(map[string]any{"deployed": true})
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "10"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer relayerSrv.Close()

	rpcSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x" + strings.Repeat("f", 64),
		})
	}))
	defer rpcSrv.Close()

	clobSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "derived-clob-key",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "derived-pass",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer clobSrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)
	t.Setenv("POLYMARKET_RELAYER_URL", relayerSrv.URL)
	t.Setenv("POLYGON_RPC_URL", rpcSrv.URL)
	t.Setenv("POLYMARKET_CLOB_URL", clobSrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")
	t.Setenv("POLYMARKET_CLOB_API_KEY", "")
	t.Setenv("POLYMARKET_CLOB_SECRET", "")
	t.Setenv("POLYMARKET_CLOB_PASSPHRASE", "")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "status", "--check-enable-trading")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false envelope=%s", stdout)
	}
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data decode: %v\n%s", err, got.Data)
	}
	enableTrading, ok := data["enableTrading"].(map[string]any)
	if !ok {
		t.Fatalf("enableTrading validation missing: %v", data)
	}
	if enableTrading["clobCredentialsReady"] != true || enableTrading["clobCredentialsSource"] != "derived" || enableTrading["ready"] != true {
		t.Fatalf("enableTrading validation unexpected: %v", enableTrading)
	}
}

// --- redeemable + redeem tests ---

const redeemTestPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

func TestDepositWalletRedeemHelpRejectsDirectCTFFallback(t *testing.T) {
	stdout, stderr, err := executeRootForTest("deposit-wallet", "redeem", "--help")
	if err != nil {
		t.Fatalf("Execute error: %v\nstderr:\n%s", err, stderr)
	}
	if strings.Contains(stdout, "via-ctf") {
		t.Fatalf("redeem help must not advertise a raw CTF fallback:\n%s", stdout)
	}
	for _, forbidden := range []string{"--via-eoa", "via-eoa", "EOA pays POL"} {
		if strings.Contains(stdout, forbidden) {
			t.Fatalf("redeem help must not advertise an EOA submission path %q:\n%s", forbidden, stdout)
		}
	}
	for _, want := range []string{
		"EIP-712 WALLET batch",
		"no direct EOA bypass",
		"ConditionalTokens fallback",
		"no SAFE/PROXY shortcut",
		"CtfCollateralAdapter",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("redeem help missing %q:\n%s", want, stdout)
		}
	}
}

func TestDepositWalletSettlementStatusReportsMissingAdapterApproval(t *testing.T) {
	dataSrv := dataAPIWithOnePosition(t)
	defer dataSrv.Close()
	rpcSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch body.Method {
		case "eth_getCode":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x60016000"})
		case "eth_chainId":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x89"})
		case "eth_call":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x0000000000000000000000000000000000000000000000000000000000000000",
			})
		default:
			t.Errorf("unexpected rpc method=%q", body.Method)
			http.NotFound(w, r)
		}
	}))
	defer rpcSrv.Close()

	t.Setenv("POLYMARKET_PRIVATE_KEY", redeemTestPrivateKey)
	t.Setenv("POLYMARKET_DATA_API_URL", dataSrv.URL)
	t.Setenv("POLYGON_RPC_URL", rpcSrv.URL)
	t.Setenv("RELAYER_API_KEY", "v2-uuid")
	t.Setenv("RELAYER_API_KEY_ADDRESS", "0xabc")

	stdout, stderr, err := executeRootForTest("--json", "deposit-wallet", "settlement-status")
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
	if data["ready"] != false {
		t.Fatalf("ready=%v want false; data=%v", data["ready"], data)
	}
	if data["status"] != "missing_adapter_approval" {
		t.Fatalf("status=%v want missing_adapter_approval; data=%v", data["status"], data)
	}
	if data["nextAction"] == "" || !strings.Contains(data["nextAction"].(string), "approve-adapters") {
		t.Fatalf("nextAction must point to approve-adapters; data=%v", data)
	}
	missing, ok := data["missingApprovals"].([]any)
	if !ok || len(missing) != 2 {
		t.Fatalf("missingApprovals=%v want 2 adapter addresses", data["missingApprovals"])
	}
	if v, _ := data["redeemableCount"].(float64); v != 1 {
		t.Fatalf("redeemableCount=%v want 1", data["redeemableCount"])
	}
}

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
	call, ok := calls[0].(map[string]any)
	if !ok {
		t.Fatalf("dry-run call shape=%T want object", calls[0])
	}
	target, _ := call["target"].(string)
	if !strings.EqualFold(target, contracts.CtfCollateralAdapter) {
		t.Fatalf("redeem dry-run target=%q want V2 collateral adapter %s", target, contracts.CtfCollateralAdapter)
	}
	if strings.EqualFold(target, contracts.CTF) {
		t.Fatalf("redeem dry-run must not target raw ConditionalTokens")
	}
	if data["path"] != "relayer-adapter" {
		t.Fatalf("path=%v want relayer-adapter", data["path"])
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
	if data["path"] != "relayer-adapter" || data["proceedsToken"] != "pUSD" {
		t.Fatalf("redeem path/proceeds=%v/%v want relayer-adapter/pUSD", data["path"], data["proceedsToken"])
	}
}

// TestDepositWalletRedeemSurfacesUpstreamAllowlistBlock asserts that when
// the relayer rejects the WALLET batch with an allowlist error, the CLI
// emits a structured RELAYER_ALLOWLIST_BLOCKED response and stops — not
// a generic submit error and not any kind of fallback path.
func TestDepositWalletRedeemSurfacesUpstreamAllowlistBlock(t *testing.T) {
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0", "id": 1,
			"result": "0x0000000000000000000000000000000000000000000000000000000000000001",
		})
	}))
	defer rpcSrv.Close()

	relaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "9"})
		case "/submit":
			http.Error(w, `{"error":"call blocked: calls to 0xAdA100Db00Ca00073811820692005400218FcE1f are not permitted"}`, http.StatusBadRequest)
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
	got := parseJSONEnvelopeForTest(t, stdout)
	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data["ok"] != false {
		t.Fatalf("ok=%v want false; stdout=%s", data["ok"], stdout)
	}
	if data["command"] != "redeem" {
		t.Fatalf("command=%v want redeem", data["command"])
	}
	innerErr, ok := data["error"].(map[string]any)
	if !ok {
		t.Fatalf("error=%T want map", data["error"])
	}
	if innerErr["code"] != "RELAYER_ALLOWLIST_BLOCKED" {
		t.Fatalf("error.code=%v want RELAYER_ALLOWLIST_BLOCKED", innerErr["code"])
	}
	if innerErr["action"] != "stop" {
		t.Fatalf("error.action=%v want stop", innerErr["action"])
	}
	upstream, ok := innerErr["upstream"].(map[string]any)
	if !ok {
		t.Fatalf("error.upstream=%T want map", innerErr["upstream"])
	}
	if upstream["state"] != "allowlist-rejected" {
		t.Fatalf("error.upstream.state=%v want allowlist-rejected", upstream["state"])
	}
	if _, ok := upstream["tracker"]; ok {
		t.Fatalf("deprecated upstream tracker present: %v", upstream)
	}
}
