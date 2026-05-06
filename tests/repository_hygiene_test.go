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
		"pkg",
		"cmd/polymarket",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(unsafePath))); err == nil {
			t.Fatalf("unsafe Rust/prototype path still exists: %s", unsafePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("could not inspect %s: %v", unsafePath, err)
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
