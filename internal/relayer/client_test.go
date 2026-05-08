package relayer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
)

func TestRelayerTransactionState_IsTerminal(t *testing.T) {
	tests := []struct {
		state    RelayerTransactionState
		terminal bool
		success  bool
	}{
		{StateNew, false, false},
		{StateExecuted, false, false},
		{StateMined, true, true},
		{StateConfirmed, true, true},
		{StateFailed, true, false},
		{StateInvalid, true, false},
	}
	for _, tt := range tests {
		if got := tt.state.IsTerminal(); got != tt.terminal {
			t.Errorf("%s.IsTerminal() = %v, want %v", tt.state, got, tt.terminal)
		}
		if got := tt.state.IsSuccess(); got != tt.success {
			t.Errorf("%s.IsSuccess() = %v, want %v", tt.state, got, tt.success)
		}
	}
}

func TestNew_RequiresBuilderConfig(t *testing.T) {
	_, err := New("https://relayer.example.com", auth.BuilderConfig{}, 137)
	if err == nil {
		t.Fatal("expected error for empty builder config")
	}
}

func TestNew_DefaultChainID(t *testing.T) {
	bc := auth.BuilderConfig{Key: "k", Secret: "s", Passphrase: "p"}
	c, err := New("https://relayer.example.com", bc, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if c.chainID != 137 {
		t.Errorf("expected default chainID 137, got %d", c.chainID)
	}
}

func TestClient_RequiresValidAddress(t *testing.T) {
	bc := auth.BuilderConfig{Key: "k", Secret: "s", Passphrase: "p"}
	c, err := New("https://relayer.example.com", bc, 137)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if _, err := c.SubmitWalletCreate(nil, ""); err == nil {
		t.Fatal("expected error for empty owner address")
	}
	if _, err := c.GetNonce(nil, ""); err == nil {
		t.Fatal("expected error for empty nonce address")
	}
	if _, err := c.GetTransaction(nil, ""); err == nil {
		t.Fatal("expected error for empty tx ID")
	}
	if _, err := c.IsDeployed(nil, ""); err == nil {
		t.Fatal("expected error for empty deployed address")
	}
	if _, err := c.SubmitWalletBatch(nil, "", "", "", "", "", nil); err == nil {
		t.Fatal("expected error for empty wallet batch params")
	}
	if _, err := c.SubmitWalletBatch(nil, "0x1", "0x2", "1", "0xsig", "99999", []DepositWalletCall{}); err == nil {
		t.Fatal("expected error for empty calls")
	}
}

func TestClient_GetTransactionAcceptsArrayResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method=%s, want GET", r.Method)
		}
		if r.URL.Path != "/transaction" {
			t.Errorf("path=%q, want /transaction", r.URL.Path)
		}
		if got := r.URL.Query().Get("id"); got != "tx-1" {
			t.Errorf("id=%q, want tx-1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"transactionID":"tx-1","transactionHash":"0xabc","state":"STATE_CONFIRMED","type":"WALLET-CREATE"}]`))
	}))
	defer srv.Close()

	bc := auth.BuilderConfig{Key: "k", Secret: "s", Passphrase: "p"}
	c, err := New(srv.URL, bc, 137)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tx, err := c.GetTransaction(context.Background(), "tx-1")
	if err != nil {
		t.Fatalf("GetTransaction: %v", err)
	}
	if tx.TransactionID != "tx-1" || tx.TransactionHash != "0xabc" || tx.State != "STATE_CONFIRMED" {
		t.Fatalf("unexpected transaction: %+v", tx)
	}
}

func TestClient_GetTransactionRejectsEmptyArrayResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	bc := auth.BuilderConfig{Key: "k", Secret: "s", Passphrase: "p"}
	c, err := New(srv.URL, bc, 137)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = c.GetTransaction(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("error=%q, want transaction ID", err.Error())
	}
}
