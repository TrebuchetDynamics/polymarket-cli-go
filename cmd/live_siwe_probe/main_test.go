package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPollDepositWalletCodeUsesProvidedRPC(t *testing.T) {
	var gotMethod string
	var gotAddr string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		var body struct {
			Method string   `json:"method"`
			Params []string `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Method != "eth_getCode" || len(body.Params) != 2 {
			t.Fatalf("unexpected RPC body: %+v", body)
		}
		gotAddr = body.Params[0]
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x60016000"}`))
	}))
	defer server.Close()

	addr := "0x19bE70b1e4F59C0663a999C0dC6f5b3C68CFCaF3"
	hasCode, err := pollDepositWalletCode(context.Background(), server.Client(), server.URL, addr)
	if err != nil {
		t.Fatalf("pollDepositWalletCode returned error: %v", err)
	}
	if !hasCode {
		t.Fatal("expected deployed code")
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %s, want POST", gotMethod)
	}
	if gotAddr != addr {
		t.Fatalf("addr = %s, want %s", gotAddr, addr)
	}
}

func TestPollDepositWalletCodeTreatsEmptyCodeAsNotDeployed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x"}`))
	}))
	defer server.Close()

	hasCode, err := pollDepositWalletCode(context.Background(), server.Client(), server.URL, "0x19bE70b1e4F59C0663a999C0dC6f5b3C68CFCaF3")
	if err != nil {
		t.Fatalf("pollDepositWalletCode returned error: %v", err)
	}
	if hasCode {
		t.Fatal("empty code should not be deployed")
	}
}
