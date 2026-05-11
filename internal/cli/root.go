package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/dataapi"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/marketdiscovery"
	"github.com/TrebuchetDynamics/polygolem/internal/preflight"
	"github.com/spf13/cobra"
)

const (
	gammaBaseURL        = "https://gamma-api.polymarket.com"
	clobBaseURL         = "https://clob.polymarket.com"
	dataBaseURL         = "https://data-api.polymarket.com"
	marketStreamBaseURL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"
)

type Options struct {
	Version string
	Stdout  io.Writer
	Stderr  io.Writer
}

type wire struct {
	gamma    *gamma.Client
	clob     *clob.Client
	data     *dataapi.Client
	discover *marketdiscovery.Service
	jsonOut  bool
}

func (w *wire) printJSON(cmd *cobra.Command, v interface{}) error {
	return writeCommandJSON(cmd, v)
}

func newWire(jsonOut bool) *wire {
	clobClient := clob.NewClient(clobBaseURL, nil)
	if key, ok := clobL2CredentialsFromEnv(); ok {
		clobClient.SetL2Credentials(key)
	}

	return &wire{
		gamma:    gamma.NewClient(gammaBaseURL, nil),
		clob:     clobClient,
		data:     dataapi.NewClient(dataBaseURL, nil),
		discover: marketdiscovery.New(gamma.NewClient(gammaBaseURL, nil), clob.NewClient(clobBaseURL, nil)),
		jsonOut:  jsonOut,
	}
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
		Use:           "polygolem",
		Short:         "Safe Polymarket SDK and CLI for Go",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit JSON output")

	root.AddCommand(&cobra.Command{
		Use: "version", Short: "Print version", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return writeCommandJSON(cmd, map[string]string{"version": opts.Version})
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "polygolem %s\n", opts.Version)
			return err
		},
	})

	root.AddCommand(&cobra.Command{
		Use: "preflight", Short: "Inspect local CLI readiness", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result := runLocalPreflight(cmd.Context(), opts.Version)
			if jsonOutput {
				return writeCommandJSON(cmd, result)
			}
			return writePreflight(cmd.OutOrStdout(), result)
		},
	})

	root.AddCommand(discoverCmd(jsonOutput))
	root.AddCommand(orderbookCmd(jsonOutput))
	root.AddCommand(clobCmd(jsonOutput))
	root.AddCommand(dataCmd(jsonOutput))
	root.AddCommand(healthCmd(jsonOutput))
	root.AddCommand(eventsCmd(jsonOutput))
	root.AddCommand(bridgeCmd(jsonOutput))
	root.AddCommand(marketDataCmd(jsonOutput))
	root.AddCommand(streamCmd(jsonOutput))
	root.AddCommand(depositWalletCmd(jsonOutput))
	root.AddCommand(newBuilderCommand(jsonOutput))
	root.AddCommand(paperCmd(jsonOutput))
	authCmd := commandGroup("auth", "Inspect authentication readiness",
		newAuthStatusCommand(jsonOutput),
	)
	authCmd.AddCommand(newAuthCLOBProbeCommand(jsonOutput))
	authCmd.AddCommand(newAuthLoginCommand(jsonOutput))
	authCmd.AddCommand(newAuthHeadlessOnboardCommand(jsonOutput))
	authCmd.AddCommand(newAuthExportKeyCommand(jsonOutput))
	root.AddCommand(authCmd)
	root.AddCommand(commandGroup("live", "Inspect live gate status",
		skeleton("status"),
	))
	installJSONContract(root)
	return root
}
func commandGroup(use, short string, children ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{Use: use, Short: short, Args: cobra.NoArgs,
		Annotations: map[string]string{commandGroupAnnotation: "true"},
		RunE:        func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}
	cmd.AddCommand(children...)
	return cmd
}

func skeleton(use string) *cobra.Command {
	return &cobra.Command{Use: use, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonEnabled(cmd) {
				return fmt.Errorf("%s: not implemented", commandName(cmd))
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: not implemented\n", cmd.CommandPath())
			return err
		},
	}
}

func runLocalPreflight(ctx context.Context, version string) preflight.Result {
	return preflight.Run(ctx, []preflight.Check{
		{Name: "version", Probe: func(context.Context) error {
			if version == "" {
				return fmt.Errorf("version is empty")
			}
			return nil
		}},
		{Name: "output", Probe: func(context.Context) error { return nil }},
		{Name: "clob_builder_code", Probe: func(context.Context) error {
			return validateBuilderCodeForCLI(builderCodeFromFlagOrEnv(""))
		}},
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
