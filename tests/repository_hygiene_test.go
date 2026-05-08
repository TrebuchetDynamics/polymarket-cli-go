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
		"scripts",
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

	// pkg/ is the approved public SDK boundary.
	if _, err := os.Stat(filepath.Join(root, "pkg/clob")); err != nil {
		t.Fatalf("pkg/clob public boundary is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pkg/orderbook")); err != nil {
		t.Fatalf("pkg/orderbook public boundary is missing: %v", err)
	}
	// Deprecated compatibility package retained for existing consumers.
	if _, err := os.Stat(filepath.Join(root, "pkg/bookreader")); err != nil {
		t.Fatalf("pkg/bookreader compatibility boundary is missing: %v", err)
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
