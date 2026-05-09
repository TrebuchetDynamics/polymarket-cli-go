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

func TestPolygonMainnetIncludesV2Adapters(t *testing.T) {
	r := PolygonMainnet()
	cases := map[string]struct{ got, want string }{
		"CtfCollateralAdapter":        {r.CtfCollateralAdapter, "0xADa100874d00e3331D00F2007a9c336a65009718"},
		"NegRiskCtfCollateralAdapter": {r.NegRiskCtfCollateralAdapter, "0xAdA200001000ef00D07553cEE7006808F895c6F1"},
		"CollateralOnramp":            {r.CollateralOnramp, "0x93070a847efEf7F70739046A929D47a521F5B8ee"},
		"CollateralOfframp":           {r.CollateralOfframp, "0x2957922Eb93258b93368531d39fAcCA3B4dC5854"},
		"PermissionedRamp":            {r.PermissionedRamp, "0xebC2459Ec962869ca4c0bd1E06368272732BCb08"},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", name, c.got, c.want)
		}
	}
}

func TestRedeemAdapterFor(t *testing.T) {
	if got := RedeemAdapterFor(false); got != CtfCollateralAdapter {
		t.Errorf("RedeemAdapterFor(false) = %q, want %q", got, CtfCollateralAdapter)
	}
	if got := RedeemAdapterFor(true); got != NegRiskCtfCollateralAdapter {
		t.Errorf("RedeemAdapterFor(true) = %q, want %q", got, NegRiskCtfCollateralAdapter)
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
