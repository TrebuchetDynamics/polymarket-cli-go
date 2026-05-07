package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/config"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/marketdiscovery"
	"github.com/TrebuchetDynamics/polygolem/internal/modes"
	"github.com/TrebuchetDynamics/polygolem/internal/output"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/preflight"
	"github.com/spf13/cobra"
)

const (
	gammaBaseURL = "https://gamma-api.polymarket.com"
	clobBaseURL  = "https://clob.polymarket.com"
)

type Options struct {
	Version string
	Stdout  io.Writer
	Stderr  io.Writer
}

type wire struct {
	gamma    *gamma.Client
	clob     *clob.Client
	discover *marketdiscovery.Service
	jsonOut  bool
}

func (w *wire) printJSON(cmd *cobra.Command, v interface{}) error {
	return output.WriteJSON(cmd.OutOrStdout(), v)
}

func (w *wire) printOrJSON(cmd *cobra.Command, v interface{}, plain string) error {
	if w.jsonOut {
		return output.WriteJSON(cmd.OutOrStdout(), v)
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), plain)
	return err
}

func newWire(jsonOut bool) *wire {
	gammaClient := gamma.NewClient(gammaBaseURL, nil)
	clobClient := clob.NewClient(clobBaseURL, nil)
	return &wire{
		gamma:    gammaClient,
		clob:     clobClient,
		discover: marketdiscovery.New(gammaClient, clobClient),
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
		Short:         "Safe Polygolem CLI for Polymarket research and automation",
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
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "polygolem %s\n", opts.Version)
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

	// -- discover command group --
	root.AddCommand(discoverCmd(jsonOutput))
	// -- orderbook command group --
	root.AddCommand(orderbookCmd(jsonOutput))
	// -- health command --
	root.AddCommand(healthCmd(jsonOutput))
	// -- paper (skeleton) --
	root.AddCommand(commandGroup("paper", "Inspect local paper trading state",
		safeSkeleton("buy", "Simulate a local paper buy"),
		safeSkeleton("sell", "Simulate a local paper sell"),
		safeSkeleton("positions", "List local paper positions"),
		safeSkeleton("reset", "Reset local paper state"),
	))
	// -- auth --
	root.AddCommand(commandGroup("auth", "Inspect authentication readiness",
		authStatusCmd(func() bool { return jsonOutput }),
	))
	// -- live --
	root.AddCommand(commandGroup("live", "Inspect live gate status",
		liveStatusCmd(func() bool { return jsonOutput }, opts.Version),
	))

	return root
}

func discoverCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)

	var query string
	var limit int
	var marketID string
	var marketSlug string

	cmd := commandGroup("discover", "Market discovery via Polymarket Gamma API")

	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search markets and events",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			active := true
			closed := false
			markets, err := w.gamma.Markets(cmd.Context(), &polytypes.GetMarketsParams{
				Active: &active,
				Closed: &closed,
				Limit:  limit,
			})
			if err != nil {
				return err
			}
			if query != "" {
				// Fall back to search API for text queries
				searchResp, err := w.gamma.Search(cmd.Context(), &polytypes.SearchParams{
					Q:            query,
					LimitPerType: &limit,
				})
				if err != nil {
					return err
				}
				return w.printJSON(cmd, searchResp)
			}
			return w.printJSON(cmd, markets)
		},
	}
	searchCmd.Flags().StringVar(&query, "query", "", "text query")
	searchCmd.Flags().IntVar(&limit, "limit", 10, "max results")
	cmd.AddCommand(searchCmd)

	marketCmd := &cobra.Command{
		Use:   "market",
		Short: "Get market details",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if marketID != "" {
				m, err := w.gamma.MarketByID(cmd.Context(), marketID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, m)
			}
			if marketSlug != "" {
				m, err := w.gamma.MarketBySlug(cmd.Context(), marketSlug)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, m)
			}
			return fmt.Errorf("--id or --slug required")
		},
	}
	marketCmd.Flags().StringVar(&marketID, "id", "", "market Gamma ID")
	marketCmd.Flags().StringVar(&marketSlug, "slug", "", "market slug")
	cmd.AddCommand(marketCmd)

	enrichCmd := &cobra.Command{
		Use:   "enrich",
		Short: "Enrich market with CLOB data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if marketID == "" {
				return fmt.Errorf("--id required")
			}
			m, err := w.gamma.MarketByID(cmd.Context(), marketID)
			if err != nil {
				return err
			}
			em, err := w.discover.EnrichMarket(cmd.Context(), *m)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, em)
		},
	}
	enrichCmd.Flags().StringVar(&marketID, "id", "", "market Gamma ID")
	cmd.AddCommand(enrichCmd)

	return cmd
}

func orderbookCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)

	var tokenID string

	cmd := commandGroup("orderbook", "Read CLOB order book data")

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get L2 order book for a token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}
			book, err := w.clob.OrderBook(cmd.Context(), tokenID)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, book)
		},
	}
	getCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
	cmd.AddCommand(getCmd)

	priceCmd := &cobra.Command{
		Use:   "price",
		Short: "Get best price for a token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}
			price, err := w.clob.Price(cmd.Context(), tokenID, "BUY")
			if err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]string{"token_id": tokenID, "price": price})
		},
	}
	priceCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
	cmd.AddCommand(priceCmd)

	midpointCmd := &cobra.Command{
		Use:   "midpoint",
		Short: "Get midpoint price for a token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}
			mid, err := w.clob.Midpoint(cmd.Context(), tokenID)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]string{"token_id": tokenID, "midpoint": mid})
		},
	}
	midpointCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
	cmd.AddCommand(midpointCmd)

	spreadCmd := &cobra.Command{
		Use:   "spread",
		Short: "Get spread for a token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}
			spread, err := w.clob.Spread(cmd.Context(), tokenID)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]string{"token_id": tokenID, "spread": spread})
		},
	}
	spreadCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
	cmd.AddCommand(spreadCmd)

	tickCmd := &cobra.Command{
		Use:   "tick-size",
		Short: "Get tick size for a token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}
			ts, err := w.clob.TickSize(cmd.Context(), tokenID)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, ts)
		},
	}
	tickCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
	cmd.AddCommand(tickCmd)

	feeCmd := &cobra.Command{
		Use:   "fee-rate",
		Short: "Get fee rate (bps) for a token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}
			fee, err := w.clob.FeeRateBps(cmd.Context(), tokenID)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]string{"token_id": tokenID, "fee_rate_bps": strconv.Itoa(fee)})
		},
	}
	feeCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
	cmd.AddCommand(feeCmd)

	return cmd
}

func healthCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)

	return &cobra.Command{
		Use:   "health",
		Short: "Check Gamma and CLOB API reachability",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			status := map[string]string{
				"gamma": "ok",
				"clob":  "ok",
			}
			if _, err := w.gamma.HealthCheck(cmd.Context()); err != nil {
				status["gamma"] = err.Error()
			}
			if err := w.clob.Health(cmd.Context()); err != nil {
				status["clob"] = err.Error()
			}
			return w.printJSON(cmd, status)
		},
	}
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

type authStatusOutput struct {
	AccessLevel   string `json:"access_level"`
	HasSigner     bool   `json:"has_signer"`
	HasAPIKey     bool   `json:"has_api_key"`
	HasBuilder    bool   `json:"has_builder"`
	SignatureType string `json:"signature_type"`
}

func authStatusCmd(jsonOut func() bool) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			status := readAuthStatusFromEnv()
			if jsonOut() {
				return output.WriteJSON(cmd.OutOrStdout(), status)
			}
			_, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"auth: level=%s signer=%t api_key=%t signature_type=%s\n",
				status.AccessLevel,
				status.HasSigner,
				status.HasAPIKey,
				status.SignatureType,
			)
			return err
		},
	}
}

func readAuthStatusFromEnv() authStatusOutput {
	hasSigner := strings.TrimSpace(os.Getenv("POLYMARKET_PRIVATE_KEY")) != ""
	hasAPIKey := firstNonEmpty("POLYMARKET_CLOB_API_KEY", "POLY_API_KEY") != "" &&
		firstNonEmpty("POLYMARKET_CLOB_SECRET", "POLY_SECRET") != "" &&
		firstNonEmpty("POLYMARKET_CLOB_PASS_PHRASE", "POLY_PASSPHRASE") != ""
	hasBuilder := firstNonEmpty("POLYMARKET_BUILDER_API_KEY", "POLY_BUILDER_API_KEY") != "" &&
		firstNonEmpty("POLYMARKET_BUILDER_SECRET", "POLY_BUILDER_SECRET") != "" &&
		firstNonEmpty("POLYMARKET_BUILDER_PASS_PHRASE", "POLY_BUILDER_PASSPHRASE") != ""
	level := auth.L0
	if hasSigner {
		level = auth.L1
	}
	if hasAPIKey {
		level = auth.L2
	}
	signatureType := strings.TrimSpace(os.Getenv("POLYMARKET_SIGNATURE_TYPE"))
	if signatureType == "" {
		signatureType = "proxy"
	}
	return authStatusOutput{
		AccessLevel:   level.String(),
		HasSigner:     hasSigner,
		HasAPIKey:     hasAPIKey,
		HasBuilder:    hasBuilder,
		SignatureType: signatureType,
	}
}

type liveStatusOutput struct {
	Allowed       bool            `json:"allowed"`
	EnvEnabled    bool            `json:"env_enabled"`
	ConfigEnabled bool            `json:"config_enabled"`
	ConfirmLive   bool            `json:"confirm_live"`
	PreflightOK   bool            `json:"preflight_ok"`
	Failures      []modes.Failure `json:"failures"`
}

func liveStatusCmd(jsonOut func() bool, version string) *cobra.Command {
	var confirmLive bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show live gate status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.Options{})
			if err != nil {
				return err
			}
			preflightResult := runLocalPreflight(cmd.Context(), version)
			envEnabled := strings.EqualFold(strings.TrimSpace(os.Getenv("POLYMARKET_LIVE_PROFILE")), "on")
			gates := modes.ValidateLiveGates(modes.LiveGateInput{
				EnvEnabled:    envEnabled,
				ConfigEnabled: cfg.LiveTradingEnabled,
				ConfirmLive:   confirmLive,
				PreflightOK:   preflightResult.OK,
			})
			status := liveStatusOutput{
				Allowed:       gates.Allowed,
				EnvEnabled:    envEnabled,
				ConfigEnabled: cfg.LiveTradingEnabled,
				ConfirmLive:   confirmLive,
				PreflightOK:   preflightResult.OK,
				Failures:      gates.Failures,
			}
			if jsonOut() {
				return output.WriteJSON(cmd.OutOrStdout(), status)
			}
			state := "blocked"
			if status.Allowed {
				state = "allowed"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "live: %s\n", state)
			return err
		},
	}
	cmd.Flags().BoolVar(&confirmLive, "confirm-live", false, "include live confirmation gate in status evaluation")
	return cmd
}

func firstNonEmpty(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
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
