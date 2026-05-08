package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TrebuchetDynamics/polygolem/internal/cli"
)

type generatedFile struct {
	path string
	body string
}

func main() {
	check := flag.Bool("check", false, "check generated docs without writing files")
	flag.Parse()

	root := cli.NewRootCommand(cli.Options{Version: "dev"})
	files := []generatedFile{
		{path: "docs/COMMANDS.md", body: cli.GenerateCommandsMarkdown(root)},
		{path: "docs-site/src/content/docs/reference/cli.mdx", body: cli.GenerateCLIReferenceMDX(root)},
	}

	if *check {
		if err := checkGeneratedFiles(files); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := writeGeneratedFiles(files); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func checkGeneratedFiles(files []generatedFile) error {
	for _, file := range files {
		current, err := os.ReadFile(file.path)
		if err != nil {
			return err
		}
		if string(current) != file.body {
			return fmt.Errorf("%s is stale; run go run ./cmd/polygolem_docs", file.path)
		}
	}
	return nil
}

func writeGeneratedFiles(files []generatedFile) error {
	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file.path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(file.path, []byte(file.body), 0o644); err != nil {
			return err
		}
	}
	return nil
}
