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
	root.AddCommand(&cobra.Command{
		Use:   "preflight",
		Short: "Inspect local CLI readiness",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "preflight: not configured")
			return err
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "markets",
		Short: "Read market data",
	})
	root.AddCommand(&cobra.Command{
		Use:   "orderbook",
		Short: "Read order books",
	})
	root.AddCommand(&cobra.Command{
		Use:   "prices",
		Short: "Read prices",
	})
	root.AddCommand(&cobra.Command{
		Use:   "paper",
		Short: "Inspect local paper trading state",
	})
	root.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Inspect authentication readiness",
	})
	root.AddCommand(&cobra.Command{
		Use:   "live",
		Short: "Inspect live gate status",
	})

	return root
}
