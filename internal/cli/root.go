package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
	Version string
	Stdout  io.Writer
	Stderr  io.Writer
}

func NewRootCommand(opts Options) *cobra.Command {
	if opts.Version == "" {
		opts.Version = "dev"
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	root := &cobra.Command{
		Use:           "polymarket",
		Short:         "Safe Polymarket CLI for research and automation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "polymarket %s\n", opts.Version)
			return err
		},
	})

	return root
}
