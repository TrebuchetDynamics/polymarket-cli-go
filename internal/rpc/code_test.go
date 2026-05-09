package rpc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHasCodeTrueWhenRPCReturnsBytecode(t *testing.T) {
	server := codeServer(t, "0x60016000")
	defer server.Close()

	deployed, err := HasCode(t.Context(), "0x21999a074344610057c9b2B362332388a44502D4", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !deployed {
		t.Fatal("expected deployed code")
	}
}

func TestHasCodeFalseWhenRPCReturnsEmptyCode(t *testing.T) {
	server := codeServer(t, "0x")
	defer server.Close()

	deployed, err := HasCode(t.Context(), "0x21999a074344610057c9b2B362332388a44502D4", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if deployed {
		t.Fatal("expected no deployed code")
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

func TestIsApprovedForAllTrue(t *testing.T) {
	server := isApprovedForAllServer(t,
		"0x4d97dcd97ec945f40cf65f87097ace5ea0476045",
		"0x21999a074344610057c9b2b362332388a44502d4",
		"0xada100db00ca00073811820692005400218fce1f",
		true,
	)
	defer server.Close()

	got, err := IsApprovedForAll(t.Context(),
		"0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
		"0x21999a074344610057c9b2B362332388a44502D4",
		"0xAdA100Db00Ca00073811820692005400218FcE1f",
		server.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Fatal("expected approved=true")
	}
}

func TestIsApprovedForAllFalse(t *testing.T) {
	server := isApprovedForAllServer(t,
		"0x4d97dcd97ec945f40cf65f87097ace5ea0476045",
		"0x21999a074344610057c9b2b362332388a44502d4",
		"0xada100db00ca00073811820692005400218fce1f",
		false,
	)
	defer server.Close()

	got, err := IsApprovedForAll(t.Context(),
		"0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
		"0x21999a074344610057c9b2B362332388a44502D4",
		"0xAdA100Db00Ca00073811820692005400218FcE1f",
		server.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Fatal("expected approved=false")
	}
}

func isApprovedForAllServer(t *testing.T, expectTo, expectOwner, expectOperator string, approved bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
			Params []any  `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode rpc request: %v", err)
		}
		// ethclient pings eth_chainId on dial; reply 0x89 (Polygon) and skip.
		if body.Method == "eth_chainId" {
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x89"})
			return
		}
		if body.Method != "eth_call" {
			t.Fatalf("method=%q want eth_call", body.Method)
		}
		call, _ := body.Params[0].(map[string]any)
		if to, _ := call["to"].(string); to != expectTo {
			t.Errorf("to=%s want %s", to, expectTo)
		}
		// ethclient uses "input" (per EIP-1474); some legacy nodes also accept "data".
		data, _ := call["input"].(string)
		if data == "" {
			data, _ = call["data"].(string)
		}
		if len(data) != 2+8+64+64 {
			t.Fatalf("data len=%d (%s)", len(data), data)
		}
		if data[:10] != "0xe985e9c5" {
			t.Errorf("selector=%s want 0xe985e9c5 (isApprovedForAll)", data[:10])
		}
		// owner is in bytes 10..73; operator in 74..137 (0x-prefixed indexing).
		ownerHex := "0x" + data[10+24:10+64]
		operatorHex := "0x" + data[10+64+24:10+64+64]
		if ownerHex != expectOwner {
			t.Errorf("owner=%s want %s", ownerHex, expectOwner)
		}
		if operatorHex != expectOperator {
			t.Errorf("operator=%s want %s", operatorHex, expectOperator)
		}
		result := "0x" + "00000000000000000000000000000000000000000000000000000000000000"
		if approved {
			result += "01"
		} else {
			result += "00"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  result,
		})
	}))
}
