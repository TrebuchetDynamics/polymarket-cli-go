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
	for _, relativePath := range append([]string{"README.md"}, requiredDocs...) {
		content := readRepositoryFile(t, root, relativePath)
		for _, blockedPhrase := range blockedPhrases {
			if strings.Contains(content, blockedPhrase) {
				t.Fatalf("%s contains unsupported Phase 1 phrase %q", relativePath, blockedPhrase)
			}
		}
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
