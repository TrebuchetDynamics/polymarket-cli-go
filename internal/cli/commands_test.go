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
	for _, want := range []string{"preflight", "markets", "orderbook", "prices", "paper", "auth", "live"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
}
