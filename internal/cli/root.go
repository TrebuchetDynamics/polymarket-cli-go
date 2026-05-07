package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/marketdiscovery"
	"github.com/TrebuchetDynamics/polygolem/internal/output"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/preflight"
	"github.com/TrebuchetDynamics/polygolem/pkg/bridge"
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

func newWire(jsonOut bool) *wire {
	return &wire{
		gamma:    gamma.NewClient(gammaBaseURL, nil),
		clob:     clob.NewClient(clobBaseURL, nil),
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
				return output.WriteJSON(cmd.OutOrStdout(), map[string]string{"version": opts.Version})
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
				return output.WriteJSON(cmd.OutOrStdout(), result)
			}
			return writePreflight(cmd.OutOrStdout(), result)
		},
	})

	root.AddCommand(discoverCmd(jsonOutput))
	root.AddCommand(orderbookCmd(jsonOutput))
	root.AddCommand(healthCmd(jsonOutput))
	root.AddCommand(eventsCmd(jsonOutput))
	root.AddCommand(bridgeCmd(jsonOutput))
	root.AddCommand(commandGroup("paper", "Inspect local paper trading state",
		skeleton("buy"), skeleton("sell"), skeleton("positions"), skeleton("reset"),
	))
	root.AddCommand(commandGroup("auth", "Inspect authentication readiness",
		skeleton("status"),
	))
	root.AddCommand(commandGroup("live", "Inspect live gate status",
		skeleton("status"),
	))
	return root
}

func discoverCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var query, marketID, marketSlug string
	var limit int

	cmd := commandGroup("discover", "Market discovery via Polymarket Gamma API")

	searchCmd := &cobra.Command{
		Use: "search", Short: "Search markets and events", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if query != "" {
				resp, err := w.gamma.Search(cmd.Context(), &polytypes.SearchParams{Q: query, LimitPerType: &limit})
				if err != nil {
					return err
				}
				return w.printJSON(cmd, resp)
			}
			active, closed := true, false
			markets, err := w.gamma.Markets(cmd.Context(), &polytypes.GetMarketsParams{Active: &active, Closed: &closed, Limit: limit})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, markets)
		},
	}
	searchCmd.Flags().StringVar(&query, "query", "", "text query")
	searchCmd.Flags().IntVar(&limit, "limit", 10, "max results")
	cmd.AddCommand(searchCmd)

	marketCmd := &cobra.Command{
		Use: "market", Short: "Get market details", Args: cobra.NoArgs,
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
		Use: "enrich", Short: "Enrich market with CLOB data", Args: cobra.NoArgs,
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

	for _, spec := range []struct{ use, short string; fn func(context.Context, string) (interface{}, error) }{
		{"get", "Get L2 order book", func(ctx context.Context, tid string) (interface{}, error) { return w.clob.OrderBook(ctx, tid) }},
		{"price", "Get best price (BUY side)", func(ctx context.Context, tid string) (interface{}, error) {
			p, err := w.clob.Price(ctx, tid, "BUY")
			return map[string]string{"token_id": tid, "price": p}, err
		}},
		{"midpoint", "Get midpoint price", func(ctx context.Context, tid string) (interface{}, error) {
			m, err := w.clob.Midpoint(ctx, tid)
			return map[string]string{"token_id": tid, "midpoint": m}, err
		}},
		{"spread", "Get bid-ask spread", func(ctx context.Context, tid string) (interface{}, error) {
			s, err := w.clob.Spread(ctx, tid)
			return map[string]string{"token_id": tid, "spread": s}, err
		}},
		{"tick-size", "Get minimum tick size", func(ctx context.Context, tid string) (interface{}, error) { return w.clob.TickSize(ctx, tid) }},
		{"fee-rate", "Get fee rate in bps", func(ctx context.Context, tid string) (interface{}, error) {
			f, err := w.clob.FeeRateBps(ctx, tid)
			return map[string]string{"token_id": tid, "fee_rate_bps": strconv.Itoa(f)}, err
		}},
		{"last-trade", "Get last trade price", func(ctx context.Context, tid string) (interface{}, error) {
			p, err := w.clob.LastTradePrice(ctx, tid)
			return map[string]string{"token_id": tid, "price": p}, err
		}},
	} {
		sub := spec
		c := &cobra.Command{Use: sub.use, Short: sub.short, Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				if tokenID == "" {
					return fmt.Errorf("--token-id required")
				}
				result, err := sub.fn(cmd.Context(), tokenID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, result)
			},
		}
		c.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
		cmd.AddCommand(c)
	}
	return cmd
}

func healthCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	return &cobra.Command{
		Use: "health", Short: "Check Gamma and CLOB API reachability", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			status := map[string]string{"gamma": "ok", "clob": "ok"}
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

func eventsCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var limit int
	cmd := commandGroup("events", "List Polymarket events")
	cmd.AddCommand(&cobra.Command{
		Use: "list", Short: "List events", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			events, err := w.gamma.Events(cmd.Context(), &polytypes.GetEventsParams{Limit: limit})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, events)
		},
	})
	return cmd
}

func bridgeCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("bridge", "Polymarket Bridge API")
	cmd.AddCommand(&cobra.Command{
		Use: "assets", Short: "List supported bridge assets", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bc := bridge.NewClient("", nil)
			a, err := bc.GetSupportedAssets(cmd.Context())
			if err != nil {
				return err
			}
			return w.printJSON(cmd, a)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use: "deposit <address>", Short: "Create deposit addresses", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bc := bridge.NewClient("", nil)
			d, err := bc.CreateDepositAddress(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, d)
		},
	})
	return cmd
}

func commandGroup(use, short string, children ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}
	cmd.AddCommand(children...)
	return cmd
}

func skeleton(use string) *cobra.Command {
	return &cobra.Command{Use: use, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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
