package main

import (
	"strings"
	"testing"
)

func TestRunRequiresExplicitLiveGate(t *testing.T) {
	t.Setenv("POLYGOLEM_EOA_KEY_SCOUT_LIVE", "")

	err := run()
	if err == nil {
		t.Fatal("expected live gate error")
	}
	if !strings.Contains(err.Error(), "POLYGOLEM_EOA_KEY_SCOUT_LIVE=1") {
		t.Fatalf("unexpected error: %v", err)
	}
}
