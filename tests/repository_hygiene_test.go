package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryHygiene(t *testing.T) {
	root := repositoryRoot(t)

	ciPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	ci, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatalf("expected Go CI workflow at %s: %v", ciPath, err)
	}

	ciContent := string(ci)
	for _, required := range []string{
		"actions/setup-go@",
		"go-version-file: go.mod",
		"git ls-files -z '*.go' ':!:opensource-projects/**'",
		"xargs -0 gofmt -w",
		"go vet ./...",
		"go test ./...",
		"git diff --exit-code",
	} {
		if !strings.Contains(ciContent, required) {
			t.Fatalf("expected CI workflow to contain %q", required)
		}
	}

	for _, unsafePath := range []string{
		"Cargo.toml",
		"Cargo.lock",
		"Formula",
		"install.sh",
		"scripts/install.sh",
		"scripts/release.sh",
		"src",
		".github/workflows/release.yml",
		"cmd/polymarket",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(unsafePath))); err == nil {
			t.Fatalf("unsafe Rust/prototype path still exists: %s", unsafePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("could not inspect %s: %v", unsafePath, err)
		}
	}
	if entries, err := os.ReadDir(filepath.Join(root, "scripts")); err == nil {
		for _, entry := range entries {
			if entry.Name() != "playwright-capture" && entry.Name() != "coverage.sh" {
				t.Fatalf("unexpected scripts path still exists: scripts/%s", entry.Name())
			}
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("could not inspect scripts: %v", err)
	}

	// pkg/ is the approved public SDK boundary.
	if _, err := os.Stat(filepath.Join(root, "pkg/clob")); err != nil {
		t.Fatalf("pkg/clob public boundary is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pkg/contracts")); err != nil {
		t.Fatalf("pkg/contracts public boundary is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pkg/settlement")); err != nil {
		t.Fatalf("pkg/settlement public boundary is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pkg/orderbook")); err != nil {
		t.Fatalf("pkg/orderbook public boundary is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pkg/stream")); err != nil {
		t.Fatalf("pkg/stream public boundary is missing: %v", err)
	}
}

func TestRepositoryDoesNotPublishResolvedRemoteBlocker(t *testing.T) {
	root := repositoryRoot(t)

	todo, err := os.ReadFile(filepath.Join(root, "TODO.md"))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("could not inspect TODO.md: %v", err)
	}

	content := string(todo)
	for _, stale := range []string{
		"TrebuchetDynamics/polygolem.git",
		"[BLOCKED] Push phase-1-tdd to origin",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("TODO.md contains resolved remote blocker %q", stale)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
}
