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
		"actions/setup-go@v5",
		"go-version-file: go.mod",
		"go vet ./...",
		"go test ./...",
		"gofmt -w .",
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

	// pkg/ is the approved public SDK boundary (Phase 0 bookreader)
	if _, err := os.Stat(filepath.Join(root, "pkg/bookreader")); err != nil {
		t.Fatalf("pkg/bookreader public boundary is missing: %v", err)
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
