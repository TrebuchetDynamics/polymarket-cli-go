package contracts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPolygonMainnetRegistry(t *testing.T) {
	registry := PolygonMainnet()
	if registry.ChainID != 137 {
		t.Fatalf("chain id=%d", registry.ChainID)
	}
	if registry.DepositWalletFactory != DepositWalletFactory {
		t.Fatalf("deposit wallet factory=%q", registry.DepositWalletFactory)
	}
	if registry.PUSD != PUSD {
		t.Fatalf("pusd=%q", registry.PUSD)
	}
}

func TestDepositWalletDeployedUsesEthGetCode(t *testing.T) {
	server := codeServer(t, "0x60016000")
	defer server.Close()

	status, err := DepositWalletDeployed(t.Context(), "0x21999a074344610057c9b2B362332388a44502D4", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Deployed {
		t.Fatal("expected deployed")
	}
	if status.Source != "polygon_eth_getCode" {
		t.Fatalf("source=%q", status.Source)
	}
}

func TestDepositWalletDeployedFalseForEmptyCode(t *testing.T) {
	server := codeServer(t, "0x")
	defer server.Close()

	status, err := DepositWalletDeployed(t.Context(), "0x21999a074344610057c9b2B362332388a44502D4", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if status.Deployed {
		t.Fatal("expected not deployed")
	}
}

func codeServer(t *testing.T, result string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode rpc request: %v", err)
		}
		if body.Method != "eth_getCode" {
			t.Fatalf("method=%q want eth_getCode", body.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  result,
		})
	}))
}
