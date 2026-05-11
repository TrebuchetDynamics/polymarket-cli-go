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
		"docs/history/REFERENCE-RUST-CLI.md",
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
	if !strings.Contains(readme, "polygolem") && !strings.Contains(readme, "Go Phase 1") {
		t.Fatalf("README.md must identify this repository as polygolem or Go Phase 1")
	}

	blockedPhrases := []string{
		"Rust CLI for Polymarket",
		"cargo install --path .",
		"brew install polymarket",
		"brew install polygolem",
		"polymarket setup",
		"polygolem setup",
		"polymarket wallet create",
		"polygolem wallet create",
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

	loginDocs := []string{
		"README.md",
		"docs/ONBOARDING.md",
		"docs/BROWSER-SETUP.md",
		"docs-site/src/content/docs/docs/guides/deposit-wallet-lifecycle.mdx",
		"docs-site/src/content/docs/docs/guides/builder-auto.mdx",
	}
	for _, relativePath := range loginDocs {
		content := readRepositoryFile(t, root, relativePath)
		for _, required := range []string{
			"polygolem auth login",
			"Polymarket login signs with the EOA",
			"deposit wallet",
		} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s must document headless auth login wording %q", relativePath, required)
			}
		}
		for _, blocked := range []string{
			"New users need one browser login",
			"browser login is required",
			"Requires browser login for new users",
			"pure headless onboarding is not possible",
			"Deposit-wallet-owned API keys cannot be created headlessly",
		} {
			if strings.Contains(content, blocked) {
				t.Fatalf("%s contains stale browser-first onboarding claim %q", relativePath, blocked)
			}
		}
	}

	reference := readRepositoryFile(t, root, "docs/history/REFERENCE-RUST-CLI.md")
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

	architecture := readRepositoryFile(t, root, "docs/ARCHITECTURE.md")
	for _, required := range []string{
		"Go protocol and automation stack for Polymarket with a\nCobra-based CLI frontend",
		"Command handlers parse flags, call package APIs, and render output via\n`internal/output`",
		"Cobra command handlers must not contain protocol or trading business logic",
		"**Read-only** (default): public market data only",
		"**Paper**: local simulation",
		"**Live**: gated. Requires preflight + risk + funding gates to pass",
	} {
		if !strings.Contains(architecture, required) {
			t.Fatalf("docs/ARCHITECTURE.md must include architecture framing %q", required)
		}
	}

	commands := readRepositoryFile(t, root, "docs/COMMANDS.md")
	for _, required := range []string{
		"--json",
		"polygolem --json version",
		"set -euo pipefail",
		"jq",
	} {
		if !strings.Contains(commands, required) {
			t.Fatalf("docs/COMMANDS.md must include command automation guidance %q", required)
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
	for _, required := range []string{
		"Preflight checks config validity, wallet readiness, auth readiness, network consistency, API health, and chain consistency",
		"Automation must treat any preflight failure as terminal",
		"Dangerous operations include real order submission, payload signing, on-chain transactions, token approvals, private-key handling, and authenticated trading mutations",
		"Phase 1 intentionally contains no code path for those operations",
	} {
		if !strings.Contains(safety, required) {
			t.Fatalf("docs/SAFETY.md must include safety guidance %q", required)
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
