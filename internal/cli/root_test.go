package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommandPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer

	root := NewRootCommand(Options{
		Version: "test-version",
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "test-version") {
		t.Fatalf("version output %q does not include test-version", got)
	}
}
