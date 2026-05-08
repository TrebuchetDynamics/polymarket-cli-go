package cli

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

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
	root.AddCommand(clobCmd(jsonOutput))
	root.AddCommand(healthCmd(jsonOutput))
	root.AddCommand(eventsCmd(jsonOutput))
	root.AddCommand(bridgeCmd(jsonOutput))
	root.AddCommand(depositWalletCmd(jsonOutput))
	root.AddCommand(newBuilderCommand(jsonOutput))
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

func clobCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("clob", "CLOB market data and authenticated account commands")

	addOutput := func(c *cobra.Command, output *string) {
		c.Flags().StringVar(output, "output", "json", "output format (json)")
	}
	checkOutput := func(output string) error {
		if output != "" && output != "json" {
			return fmt.Errorf("only --output json is supported")
		}
		return nil
	}
	privateKey := func() (string, error) {
		key := strings.TrimSpace(os.Getenv("POLYMARKET_PRIVATE_KEY"))
		if key == "" {
			return "", fmt.Errorf("POLYMARKET_PRIVATE_KEY is required")
		}
		return key, nil
	}

	var bookOutput string
	bookCmd := &cobra.Command{Use: "book <token-id>", Short: "Get L2 order book", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(bookOutput); err != nil {
				return err
			}
			book, err := w.clob.OrderBook(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, book)
		},
	}
	addOutput(bookCmd, &bookOutput)
	cmd.AddCommand(bookCmd)

	var tickOutput string
	tickCmd := &cobra.Command{Use: "tick-size <token-id>", Short: "Get minimum tick size", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(tickOutput); err != nil {
				return err
			}
			tick, err := w.clob.TickSize(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, tick)
		},
	}
	addOutput(tickCmd, &tickOutput)
	cmd.AddCommand(tickCmd)

	var createKeyOutput string
	createKeyCmd := &cobra.Command{Use: "create-api-key", Short: "Create or derive CLOB API credentials", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(createKeyOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			apiKey, err := w.clob.CreateOrDeriveAPIKey(cmd.Context(), key)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]string{"api_key": apiKey.Key})
		},
	}
	addOutput(createKeyCmd, &createKeyOutput)
	cmd.AddCommand(createKeyCmd)

	var balanceOutput, balanceAssetType, balanceTokenID, balanceSignatureType string
	balanceCmd := &cobra.Command{Use: "balance", Short: "Get CLOB balance and allowances", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(balanceOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			sig, err := parseSignatureTypeFlag(balanceSignatureType)
			if err != nil {
				return err
			}
			res, err := w.clob.BalanceAllowance(cmd.Context(), key, clob.BalanceAllowanceParams{
				AssetType:     balanceAssetType,
				TokenID:       balanceTokenID,
				SignatureType: sig,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, normalizeCollateralBalanceResponse(balanceResponseMap(res)))
		},
	}
	addOutput(balanceCmd, &balanceOutput)
	balanceCmd.Flags().StringVar(&balanceAssetType, "asset-type", "collateral", "asset type")
	balanceCmd.Flags().StringVar(&balanceTokenID, "token-id", "", "conditional token id")
	balanceCmd.Flags().StringVar(&balanceSignatureType, "signature-type", "proxy", "signature type: eoa, proxy, safe, deposit")
	cmd.AddCommand(balanceCmd)

	var updateBalanceOutput, updateBalanceAssetType, updateBalanceTokenID, updateBalanceSignatureType string
	updateBalanceCmd := &cobra.Command{Use: "update-balance", Short: "Refresh CLOB balance and allowances", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(updateBalanceOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			sig, err := parseSignatureTypeFlag(updateBalanceSignatureType)
			if err != nil {
				return err
			}
			res, err := w.clob.UpdateBalanceAllowance(cmd.Context(), key, clob.BalanceAllowanceParams{
				AssetType:     updateBalanceAssetType,
				TokenID:       updateBalanceTokenID,
				SignatureType: sig,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, normalizeCollateralBalanceResponse(balanceResponseMap(res)))
		},
	}
	addOutput(updateBalanceCmd, &updateBalanceOutput)
	updateBalanceCmd.Flags().StringVar(&updateBalanceAssetType, "asset-type", "collateral", "asset type")
	updateBalanceCmd.Flags().StringVar(&updateBalanceTokenID, "token-id", "", "conditional token id")
	updateBalanceCmd.Flags().StringVar(&updateBalanceSignatureType, "signature-type", "proxy", "signature type: eoa, proxy, safe, deposit")
	cmd.AddCommand(updateBalanceCmd)

	var ordersOutput string
	ordersCmd := &cobra.Command{Use: "orders", Short: "List authenticated CLOB orders", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(ordersOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			rows, err := w.clob.ListOrders(cmd.Context(), key)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addOutput(ordersCmd, &ordersOutput)
	cmd.AddCommand(ordersCmd)

	var tradesOutput string
	tradesCmd := &cobra.Command{Use: "trades", Short: "List authenticated CLOB trades", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(tradesOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			rows, err := w.clob.ListTrades(cmd.Context(), key)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addOutput(tradesCmd, &tradesOutput)
	cmd.AddCommand(tradesCmd)

	var cancelOutput string
	cancelCmd := &cobra.Command{Use: "cancel <order-id>", Short: "Cancel a single open CLOB order", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(cancelOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			resp, err := w.clob.CancelOrder(cmd.Context(), key, args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, resp)
		},
	}
	addOutput(cancelCmd, &cancelOutput)
	cmd.AddCommand(cancelCmd)

	var cancelAllOutput string
	cancelAllCmd := &cobra.Command{Use: "cancel-all", Short: "Cancel all open CLOB orders", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(cancelAllOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			resp, err := w.clob.CancelAll(cmd.Context(), key)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, resp)
		},
	}
	addOutput(cancelAllCmd, &cancelAllOutput)
	cmd.AddCommand(cancelAllCmd)

	var createOrderOutput, createOrderToken, createOrderSide, createOrderPrice, createOrderSize, createOrderType, createOrderSignatureType string
	createOrderCmd := &cobra.Command{Use: "create-order", Short: "Create a signed CLOB limit order", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(createOrderOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			sig, err := parseSignatureTypeFlag(createOrderSignatureType)
			if err != nil {
				return err
			}
			res, err := w.clob.CreateLimitOrder(cmd.Context(), key, clob.CreateOrderParams{
				TokenID:       createOrderToken,
				Side:          createOrderSide,
				Price:         createOrderPrice,
				Size:          createOrderSize,
				OrderType:     createOrderType,
				SignatureType: sig,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, res)
		},
	}
	addOutput(createOrderCmd, &createOrderOutput)
	createOrderCmd.Flags().StringVar(&createOrderToken, "token", "", "CLOB token id")
	createOrderCmd.Flags().StringVar(&createOrderSide, "side", "buy", "order side")
	createOrderCmd.Flags().StringVar(&createOrderPrice, "price", "", "limit price")
	createOrderCmd.Flags().StringVar(&createOrderSize, "size", "", "order size")
	createOrderCmd.Flags().StringVar(&createOrderType, "order-type", "GTC", "order type")
	createOrderCmd.Flags().StringVar(&createOrderSignatureType, "signature-type", "proxy", "signature type: eoa, proxy, safe, deposit")
	cmd.AddCommand(createOrderCmd)

	var marketOrderOutput, marketOrderToken, marketOrderSide, marketOrderAmount, marketOrderPrice, marketOrderType, marketOrderSignatureType string
	marketOrderCmd := &cobra.Command{Use: "market-order", Short: "Create a signed CLOB market/FOK order", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(marketOrderOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			sig, err := parseSignatureTypeFlag(marketOrderSignatureType)
			if err != nil {
				return err
			}
			res, err := w.clob.CreateMarketOrder(cmd.Context(), key, clob.MarketOrderParams{
				TokenID:       marketOrderToken,
				Side:          marketOrderSide,
				Amount:        marketOrderAmount,
				Price:         marketOrderPrice,
				OrderType:     marketOrderType,
				SignatureType: sig,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, res)
		},
	}
	addOutput(marketOrderCmd, &marketOrderOutput)
	marketOrderCmd.Flags().StringVar(&marketOrderToken, "token", "", "CLOB token id")
	marketOrderCmd.Flags().StringVar(&marketOrderSide, "side", "buy", "order side")
	marketOrderCmd.Flags().StringVar(&marketOrderAmount, "amount", "", "USDC amount")
	marketOrderCmd.Flags().StringVar(&marketOrderPrice, "price", "", "limit price")
	marketOrderCmd.Flags().StringVar(&marketOrderType, "order-type", "FOK", "order type")
	marketOrderCmd.Flags().StringVar(&marketOrderSignatureType, "signature-type", "proxy", "signature type: eoa, proxy, safe, deposit")
	cmd.AddCommand(marketOrderCmd)

	var priceHistoryOutput, priceHistoryInterval string
	priceHistoryCmd := &cobra.Command{Use: "price-history <token-id>", Short: "Get CLOB token price history", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(priceHistoryOutput); err != nil {
				return err
			}
			history, err := w.clob.PricesHistory(cmd.Context(), &polytypes.PriceHistoryParams{
				Market:   args[0],
				Interval: priceHistoryInterval,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, history)
		},
	}
	addOutput(priceHistoryCmd, &priceHistoryOutput)
	priceHistoryCmd.Flags().StringVar(&priceHistoryInterval, "interval", "1m", "history interval")
	cmd.AddCommand(priceHistoryCmd)

	var marketOutput string
	marketCmd := &cobra.Command{Use: "market <condition-id>", Short: "Get CLOB market by condition ID", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(marketOutput); err != nil {
				return err
			}
			market, err := w.clob.Market(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, market)
		},
	}
	addOutput(marketCmd, &marketOutput)
	cmd.AddCommand(marketCmd)

	return cmd
}

func parseSignatureTypeFlag(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "proxy", "1":
		return 1, nil
	case "eoa", "0":
		return 0, nil
	case "safe", "gnosis", "gnosis-safe", "2":
		return 2, nil
	case "deposit", "deposit-wallet", "poly-1271", "3":
		return 3, nil
	default:
		return 0, fmt.Errorf("unsupported signature type %q", value)
	}
}

func balanceResponseMap(res *clob.BalanceAllowanceResponse) map[string]interface{} {
	out := map[string]interface{}{}
	if res == nil {
		return out
	}
	out["balance"] = res.Balance
	if len(res.Allowances) > 0 {
		out["allowances"] = res.Allowances
	}
	if strings.TrimSpace(res.Allowance) != "" {
		out["allowance"] = res.Allowance
	}
	return out
}

func normalizeCollateralBalanceResponse(raw map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		out[key] = value
	}
	balance, ok := out["balance"].(string)
	if !ok {
		return out
	}
	if scaled, ok := scaleBaseUnits(balance, 6); ok {
		out["balance"] = scaled
	}
	return out
}

func scaleBaseUnits(value string, decimals int) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, ".") || decimals <= 0 {
		return value, false
	}
	n := new(big.Int)
	if _, ok := n.SetString(value, 10); !ok {
		return value, false
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	whole := new(big.Int).Quo(new(big.Int).Set(n), scale)
	frac := new(big.Int).Mod(new(big.Int).Set(n), scale).String()
	if len(frac) < decimals {
		frac = strings.Repeat("0", decimals-len(frac)) + frac
	}
	return whole.String() + "." + frac, true
}

func orderbookCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var tokenID string

	cmd := commandGroup("orderbook", "Read CLOB order book data")

	for _, spec := range []struct {
		use, short string
		fn         func(context.Context, string) (interface{}, error)
	}{
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
