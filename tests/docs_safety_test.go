package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentationSafety(t *testing.T) {
	root := repositoryRoot(t)

	requiredDocs := []string{
		"docs/ARCHITECTURE.md",
		"docs/COMMANDS.md",
		"docs/SAFETY.md",
		"docs/REFERENCE-RUST-CLI.md",
	}
	for _, requiredDoc := range requiredDocs {
		path := filepath.Join(root, filepath.FromSlash(requiredDoc))
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected required documentation at %s: %v", requiredDoc, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected required documentation at %s to be non-empty", requiredDoc)
		}
	}

	readme := readRepositoryFile(t, root, "README.md")
	if !strings.Contains(readme, "polymarket-cli-go") && !strings.Contains(readme, "Go Phase 1") {
		t.Fatalf("README.md must identify this repository as polymarket-cli-go or Go Phase 1")
	}

	blockedPhrases := []string{
		"Rust CLI for Polymarket",
		"cargo install --path .",
		"brew install polymarket",
		"polymarket setup",
		"polymarket wallet create",
		"clob create-order",
		"clob market-order",
	}
	activeUserDocs := []string{
		"README.md",
		"docs/ARCHITECTURE.md",
		"docs/COMMANDS.md",
		"docs/SAFETY.md",
	}
	for _, relativePath := range activeUserDocs {
		content := readRepositoryFile(t, root, relativePath)
		for _, blockedPhrase := range blockedPhrases {
			if strings.Contains(content, blockedPhrase) {
				t.Fatalf("%s contains unsupported Phase 1 phrase %q", relativePath, blockedPhrase)
			}
		}
	}

	reference := readRepositoryFile(t, root, "docs/REFERENCE-RUST-CLI.md")
	expectedReferenceText := "- `market-order`: builds, signs, and posts a market order through `post_order`."
	if !strings.Contains(reference, expectedReferenceText) {
		t.Fatalf("docs/REFERENCE-RUST-CLI.md must preserve exact upstream audit text %q", expectedReferenceText)
	}

	plan := readRepositoryFile(t, root, "docs/superpowers/plans/2026-05-06-polymarket-go-cli-phase-1.md")
	expectedPlanSnippet := `rg -n "live trading works|place live|create live order|market-order" docs README.md`
	if !strings.Contains(plan, expectedPlanSnippet) {
		t.Fatalf("phase 1 plan must preserve exact verification snippet %q", expectedPlanSnippet)
	}
	expectedPlanText := "Expected: no claim that live trading works in Phase 1."
	if !strings.Contains(plan, expectedPlanText) {
		t.Fatalf("phase 1 plan must preserve expected wording %q", expectedPlanText)
	}

	safety := readRepositoryFile(t, root, "docs/SAFETY.md")
	for _, gate := range []string{
		"POLYMARKET_LIVE_PROFILE=on",
		"live_trading_enabled: true",
		"--confirm-live",
		"preflight",
	} {
		if !strings.Contains(safety, gate) {
			t.Fatalf("docs/SAFETY.md must document live gate %q", gate)
		}
	}
}

func readRepositoryFile(t *testing.T, root, relativePath string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatalf("expected to read %s: %v", relativePath, err)
	}
	return string(content)
}
