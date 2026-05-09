package settlement

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	"github.com/TrebuchetDynamics/polygolem/pkg/ctf"
	"github.com/TrebuchetDynamics/polygolem/pkg/data"
	"github.com/TrebuchetDynamics/polygolem/pkg/relayer"
)

// testPrivateKey matches the relayer test fixture (deterministic).
const testPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

func TestFindRedeemableFiltersNonRedeemable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"asset": "tok-1", "conditionId": "0xc1", "redeemable": false, "size": 1.0, "outcome": "Up"},
			{"asset": "tok-2", "conditionId": "0xc2", "redeemable": true, "size": 4.0784, "outcome": "Up", "title": "ETH winner", "slug": "eth-updown-5m-1778316000"},
			{"asset": "tok-3", "conditionId": "0xc3", "redeemable": false, "size": 0.5, "outcome": "Down"},
		})
	}))
	defer server.Close()

	client := data.NewClient(data.Config{BaseURL: server.URL})
	rows, err := FindRedeemable(context.Background(), client, "0x21999a074344610057c9b2B362332388a44502D4")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows=%d want 1; got %+v", len(rows), rows)
	}
	if rows[0].TokenID != "tok-2" || rows[0].ConditionID != "0xc2" || rows[0].Outcome != "Up" {
		t.Errorf("filtered position fields wrong: %+v", rows[0])
	}
	if rows[0].Title != "ETH winner" || rows[0].Slug != "eth-updown-5m-1778316000" {
		t.Errorf("operator-display fields not threaded: %+v", rows[0])
	}
}

func TestCheckReadinessBlocksMissingAdapterApproval(t *testing.T) {
	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"asset":        "tok-1",
			"conditionId":  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"redeemable":   true,
			"negativeRisk": false,
			"size":         2.5,
			"outcome":      "Up",
		}})
	}))
	defer dataSrv.Close()
	rpcSrv := settlementRPCServer(t, "0x60016000", false)
	defer rpcSrv.Close()

	status, err := CheckReadiness(context.Background(), data.NewClient(data.Config{BaseURL: dataSrv.URL}), "0xowner", "0x21999a074344610057c9b2B362332388a44502D4", ReadinessOptions{
		RPCURL:            rpcSrv.URL,
		RelayerConfigured: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if status.Ready {
		t.Fatalf("ready=true want false: %+v", status)
	}
	if status.Status != StatusMissingAdapterApproval {
		t.Fatalf("status=%q want %q: %+v", status.Status, StatusMissingAdapterApproval, status)
	}
	if len(status.MissingApprovals) != 2 {
		t.Fatalf("missingApprovals=%v want both adapters", status.MissingApprovals)
	}
	if status.RedeemableCount != 1 {
		t.Fatalf("redeemableCount=%d want 1", status.RedeemableCount)
	}
}

func TestCheckReadinessReadyWhenDeployedRelayerAndAdaptersApproved(t *testing.T) {
	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer dataSrv.Close()
	rpcSrv := settlementRPCServer(t, "0x60016000", true)
	defer rpcSrv.Close()

	status, err := CheckReadiness(context.Background(), data.NewClient(data.Config{BaseURL: dataSrv.URL}), "0xowner", "0x21999a074344610057c9b2B362332388a44502D4", ReadinessOptions{
		RPCURL:            rpcSrv.URL,
		RelayerConfigured: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ready {
		t.Fatalf("ready=false want true: %+v", status)
	}
	if status.Status != StatusReady {
		t.Fatalf("status=%q want %q", status.Status, StatusReady)
	}
	if len(status.AdapterApprovals) != 2 {
		t.Fatalf("adapterApprovals=%v want two adapter checks", status.AdapterApprovals)
	}
}

func settlementRPCServer(t *testing.T, code string, approved bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode rpc request: %v", err)
		}
		switch body.Method {
		case "eth_getCode":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": code})
		case "eth_chainId":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x89"})
		case "eth_call":
			result := "0x0000000000000000000000000000000000000000000000000000000000000000"
			if approved {
				result = "0x0000000000000000000000000000000000000000000000000000000000000001"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": result})
		default:
			t.Fatalf("unexpected rpc method %q", body.Method)
		}
	}))
}

func TestBuildRedeemCallBinaryTargetsCtfCollateralAdapter(t *testing.T) {
	cid := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	p := RedeemablePosition{ConditionID: cid.Hex(), NegativeRisk: false}
	call, err := BuildRedeemCall(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(call.Target, contracts.CtfCollateralAdapter) {
		t.Errorf("target=%s want CtfCollateralAdapter %s", call.Target, contracts.CtfCollateralAdapter)
	}
	if call.Value != "0" {
		t.Errorf("value=%s want 0", call.Value)
	}
	expected, err := ctf.RedeemPositionsData(common.Address{}, common.Hash{}, cid, []*big.Int{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := hex.DecodeString(strings.TrimPrefix(call.Data, "0x"))
	if err != nil {
		t.Fatalf("decode call data: %v", err)
	}
	if !bytes.Equal(got, expected) {
		t.Errorf("calldata mismatch:\n got=%x\nwant=%x", got, expected)
	}
}

func TestBuildRedeemCallNegRiskTargetsAdapter(t *testing.T) {
	cid := common.HexToHash("0xdeadbeef00000000000000000000000000000000000000000000000000000000")
	p := RedeemablePosition{ConditionID: cid.Hex(), NegativeRisk: true}
	call, err := BuildRedeemCall(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(call.Target, contracts.NegRiskCtfCollateralAdapter) {
		t.Errorf("target=%s want NegRiskCtfCollateralAdapter %s", call.Target, contracts.NegRiskCtfCollateralAdapter)
	}
}

func TestBuildRedeemCallEmptyConditionRejected(t *testing.T) {
	if _, err := BuildRedeemCall(RedeemablePosition{ConditionID: ""}); err == nil {
		t.Fatal("expected error on empty conditionID")
	}
}

func TestDedupeByConditionCollapsesYesNoSplit(t *testing.T) {
	rows := []RedeemablePosition{
		{TokenID: "yes-1", ConditionID: "0xa", Outcome: "Up"},
		{TokenID: "no-1", ConditionID: "0xa", Outcome: "Down"},
		{TokenID: "yes-2", ConditionID: "0xb", Outcome: "Up"},
		{TokenID: "skip", ConditionID: ""},
	}
	out := dedupeByCondition(rows)
	if len(out) != 2 {
		t.Fatalf("len=%d want 2", len(out))
	}
	if out[0].ConditionID != "0xa" || out[0].TokenID != "yes-1" {
		t.Errorf("first row=%+v want conditionID 0xa first-seen yes-1", out[0])
	}
	if out[1].ConditionID != "0xb" {
		t.Errorf("second row=%+v want conditionID 0xb", out[1])
	}
}

func TestSubmitRedeemHappyPath(t *testing.T) {
	var submittedCalls []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "12"})
		case "/submit":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode body: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if body["type"] != "WALLET" {
				t.Errorf("type=%v want WALLET", body["type"])
			}
			params, _ := body["depositWalletParams"].(map[string]any)
			rawCalls, _ := params["calls"].([]any)
			for _, c := range rawCalls {
				if m, ok := c.(map[string]any); ok {
					submittedCalls = append(submittedCalls, m)
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"transactionID": "redeem-1",
				"state":         "STATE_NEW",
				"type":          "WALLET",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rc, err := relayer.NewV2(server.URL, relayer.V2APIKey{Key: "v2", Address: "0xabc"}, 137)
	if err != nil {
		t.Fatal(err)
	}
	positions := []RedeemablePosition{
		{TokenID: "yes-1", ConditionID: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", Outcome: "Up"},
		{TokenID: "yes-2", ConditionID: "0xdeadbeef00000000000000000000000000000000000000000000000000000000", Outcome: "Up", NegativeRisk: true},
	}
	got, err := SubmitRedeem(context.Background(), rc, testPrivateKey, positions, 0)
	if err != nil {
		t.Fatalf("SubmitRedeem error: %v", err)
	}
	if got.TransactionID != "redeem-1" || got.State != "STATE_NEW" {
		t.Errorf("result=%+v", got)
	}
	if got.CallCount != 2 {
		t.Errorf("callCount=%d want 2", got.CallCount)
	}
	if len(submittedCalls) != 2 {
		t.Fatalf("relayer saw %d calls, want 2", len(submittedCalls))
	}
	// Call 0 routes to the binary adapter; call 1 to neg-risk adapter.
	if !strings.EqualFold(asString(submittedCalls[0]["target"]), contracts.CtfCollateralAdapter) {
		t.Errorf("call0 target=%v want %s", submittedCalls[0]["target"], contracts.CtfCollateralAdapter)
	}
	if !strings.EqualFold(asString(submittedCalls[1]["target"]), contracts.NegRiskCtfCollateralAdapter) {
		t.Errorf("call1 target=%v want %s", submittedCalls[1]["target"], contracts.NegRiskCtfCollateralAdapter)
	}
}

func TestSubmitRedeemRespectsLimit(t *testing.T) {
	var submittedCallCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nonce":
			_ = json.NewEncoder(w).Encode(map[string]any{"nonce": "1"})
		case "/submit":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			params, _ := body["depositWalletParams"].(map[string]any)
			rawCalls, _ := params["calls"].([]any)
			submittedCallCount = len(rawCalls)
			_ = json.NewEncoder(w).Encode(map[string]any{"transactionID": "x", "state": "STATE_NEW", "type": "WALLET"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rc, _ := relayer.NewV2(server.URL, relayer.V2APIKey{Key: "v2", Address: "0xabc"}, 137)
	positions := make([]RedeemablePosition, 15)
	for i := range positions {
		positions[i] = RedeemablePosition{
			ConditionID: padHexCondition(i),
			Outcome:     "Up",
		}
	}
	got, err := SubmitRedeem(context.Background(), rc, testPrivateKey, positions, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got.CallCount != 10 {
		t.Errorf("CallCount=%d want 10", got.CallCount)
	}
	if submittedCallCount != 10 {
		t.Errorf("relayer received %d calls, want 10", submittedCallCount)
	}
}

func TestSubmitRedeemRefusesEmpty(t *testing.T) {
	rc, _ := relayer.NewV2("https://relayer.example", relayer.V2APIKey{Key: "v2", Address: "0xabc"}, 137)
	if _, err := SubmitRedeem(context.Background(), rc, testPrivateKey, nil, 0); err == nil {
		t.Fatal("expected error on empty positions slice")
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func padHexCondition(i int) string {
	const zero = "0000000000000000000000000000000000000000000000000000000000000000"
	suffix := common.LeftPadBytes(big.NewInt(int64(i+1)).Bytes(), 32)
	_ = zero
	return "0x" + hex.EncodeToString(suffix)
}
