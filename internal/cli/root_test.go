package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCommandPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "test-version") {
		t.Fatalf("version output %q does not include test-version", got)
	}
}

func TestJSONFlagIsAcceptedAndPreflightEmitsJSON(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"--json", "preflight"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var got struct {
		OK     bool `json:"ok"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("preflight stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got.Checks) == 0 {
		t.Fatalf("preflight JSON checks is empty: %s", stdout.String())
	}
}

func TestDocumentedSubcommandsAreRegistered(t *testing.T) {
	for _, args := range [][]string{
		{"discover", "search"},
		{"discover", "market"},
		{"discover", "enrich"},
		{"orderbook", "get"},
		{"orderbook", "price"},
		{"orderbook", "midpoint"},
		{"orderbook", "spread"},
		{"orderbook", "tick-size"},
		{"orderbook", "fee-rate"},
		{"orderbook", "last-trade"},
		{"events", "list"},
		{"bridge", "assets"},
		{"bridge", "deposit"},
		{"health"},
		{"paper", "buy"},
		{"paper", "sell"},
		{"paper", "positions"},
		{"paper", "reset"},
		{"auth", "status"},
		{"live", "status"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
			root.SetArgs(append(args, "--help"))
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			wantPath := "polygolem " + strings.Join(args, " ")
			if !strings.Contains(stdout.String(), wantPath) {
				t.Fatalf("help output does not identify exact command path %q:\n%s", wantPath, stdout.String())
			}
		})
	}
}

func TestDocumentedSubcommandArgsAreNotHandledByParentOnly(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"discover", "search", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "polygolem discover search") {
		t.Fatalf("discover search was not handled by its own command:\n%s", stdout.String())
	}
}
