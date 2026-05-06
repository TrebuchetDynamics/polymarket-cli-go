package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/TrebuchetDynamics/polymarket-cli-go/internal/output"
	"github.com/TrebuchetDynamics/polymarket-cli-go/internal/preflight"
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

	var jsonOutput bool

	root := &cobra.Command{
		Use:           "polymarket",
		Short:         "Safe Polymarket CLI for research and automation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit JSON output")

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return output.WriteJSON(cmd.OutOrStdout(), map[string]string{"version": opts.Version})
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "polymarket %s\n", opts.Version)
			return err
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "preflight",
		Short: "Inspect local CLI readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result := runLocalPreflight(cmd.Context(), opts.Version)
			if jsonOutput {
				return output.WriteJSON(cmd.OutOrStdout(), result)
			}
			return writePreflight(cmd.OutOrStdout(), result)
		},
	})
	root.AddCommand(commandGroup("markets", "Read market data",
		safeSkeleton("search", "Search markets"),
		safeSkeleton("get", "Get one market"),
		safeSkeleton("active", "List active markets"),
	))
	root.AddCommand(commandGroup("orderbook", "Read order books",
		safeSkeleton("get", "Get an order book"),
	))
	root.AddCommand(commandGroup("prices", "Read prices",
		safeSkeleton("get", "Get token prices"),
	))
	root.AddCommand(commandGroup("paper", "Inspect local paper trading state",
		safeSkeleton("buy", "Simulate a local paper buy"),
		safeSkeleton("sell", "Simulate a local paper sell"),
		safeSkeleton("positions", "List local paper positions"),
		safeSkeleton("reset", "Reset local paper state"),
	))
	root.AddCommand(commandGroup("auth", "Inspect authentication readiness",
		safeSkeleton("status", "Show authentication status"),
	))
	root.AddCommand(commandGroup("live", "Inspect live gate status",
		safeSkeleton("status", "Show live gate status"),
	))

	return root
}

func commandGroup(use, short string, children ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(children...)
	return cmd
}

func safeSkeleton(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: not implemented\n", cmd.CommandPath())
			return err
		},
	}
}

func runLocalPreflight(ctx context.Context, version string) preflight.Result {
	return preflight.Run(ctx, []preflight.Check{
		{
			Name: "version",
			Probe: func(context.Context) error {
				if version == "" {
					return fmt.Errorf("version is empty")
				}
				return nil
			},
		},
		{
			Name: "output",
			Probe: func(context.Context) error {
				return nil
			},
		},
	})
}

func writePreflight(w io.Writer, result preflight.Result) error {
	status := "ok"
	if !result.OK {
		status = "failed"
	}
	if _, err := fmt.Fprintf(w, "preflight: %s\n", status); err != nil {
		return err
	}
	for _, check := range result.Checks {
		if check.Message == "" {
			if _, err := fmt.Fprintf(w, "- %s: %s\n", check.Name, check.Status); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "- %s: %s (%s)\n", check.Name, check.Status, check.Message); err != nil {
			return err
		}
	}
	return nil
}
