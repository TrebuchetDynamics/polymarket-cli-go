package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandsReferenceMatchesGeneratedOutput(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	want := readRepoFileForTest(t, "docs/COMMANDS.md")
	got := GenerateCommandsMarkdown(root)

	if got != want {
		t.Fatalf("docs/COMMANDS.md is stale; run go run ./cmd/polygolem_docs\n%s", firstDiffForTest(want, got))
	}
}

func TestAstroCLIReferenceMatchesGeneratedOutput(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	want := readRepoFileForTest(t, "docs-site/src/content/docs/reference/cli.mdx")
	got := GenerateCLIReferenceMDX(root)

	if got != want {
		t.Fatalf("docs-site CLI reference is stale; run go run ./cmd/polygolem_docs\n%s", firstDiffForTest(want, got))
	}
}

func TestGeneratedCommandReferenceContainsEveryCommandPath(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	got := GenerateCommandsMarkdown(root)

	for _, path := range commandPathsForTest(root) {
		want := "### " + path + "\n"
		if !strings.Contains(got, want) {
			t.Fatalf("generated command reference missing section %q", want)
		}
	}
}

func readRepoFileForTest(t *testing.T, rel string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	body, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func commandPathsForTest(root *cobra.Command) []string {
	var paths []string
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.Hidden {
			return
		}
		paths = append(paths, cmd.CommandPath())
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
	return paths
}

func firstDiffForTest(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	limit := len(wantLines)
	if len(gotLines) < limit {
		limit = len(gotLines)
	}
	for i := 0; i < limit; i++ {
		if wantLines[i] != gotLines[i] {
			return "first diff line " + strconv.Itoa(i+1) + "\nwant: " + wantLines[i] + "\n got: " + gotLines[i]
		}
	}
	if len(wantLines) != len(gotLines) {
		return "line count differs: want " + strconv.Itoa(len(wantLines)) + ", got " + strconv.Itoa(len(gotLines))
	}
	return "contents differ"
}
