package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpListsPhaseOneCommands(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	help := stdout.String()
	for _, want := range []string{"preflight", "discover", "orderbook", "health", "paper", "auth", "live"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
}

func TestDataCommandExposesOrderResultsAudit(t *testing.T) {
	root := NewRootCommand(Options{Version: "test", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"data", "order-results"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if cmd == nil || cmd.Name() != "order-results" {
		t.Fatalf("data order-results command missing, got %v", cmd)
	}
}

func TestDepositWalletCommandExposesSettlementStatus(t *testing.T) {
	root := NewRootCommand(Options{Version: "test", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"deposit-wallet", "settlement-status"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if cmd == nil || cmd.Name() != "settlement-status" {
		t.Fatalf("deposit-wallet settlement-status command missing, got %v", cmd)
	}
}
