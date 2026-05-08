package cli

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/dataapi"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/marketdiscovery"
	"github.com/TrebuchetDynamics/polygolem/internal/output"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/preflight"
	"github.com/TrebuchetDynamics/polygolem/internal/stream"
	"github.com/TrebuchetDynamics/polygolem/pkg/bridge"
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
	return output.WriteJSON(cmd.OutOrStdout(), v)
}

func newWire(jsonOut bool) *wire {
	return &wire{
		gamma:    gamma.NewClient(gammaBaseURL, nil),
		clob:     clob.NewClient(clobBaseURL, nil),
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
	root.AddCommand(dataCmd(jsonOutput))
	root.AddCommand(healthCmd(jsonOutput))
	root.AddCommand(eventsCmd(jsonOutput))
	root.AddCommand(bridgeCmd(jsonOutput))
	root.AddCommand(streamCmd(jsonOutput))
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

	var marketsLimit, marketsOffset, marketsTagID int
	var marketsOrder string
	var marketsActive, marketsClosed, marketsAscending bool
	marketsCmd := &cobra.Command{Use: "markets", Short: "List Gamma markets", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &polytypes.GetMarketsParams{
				Limit:     marketsLimit,
				Offset:    marketsOffset,
				Order:     marketsOrder,
				Active:    &marketsActive,
				Closed:    &marketsClosed,
				Ascending: &marketsAscending,
			}
			if marketsTagID > 0 {
				params.TagID = &marketsTagID
			}
			markets, err := w.gamma.Markets(cmd.Context(), params)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, markets)
		},
	}
	marketsCmd.Flags().IntVar(&marketsLimit, "limit", 20, "max markets")
	marketsCmd.Flags().IntVar(&marketsOffset, "offset", 0, "pagination offset")
	marketsCmd.Flags().StringVar(&marketsOrder, "order", "", "Gamma order field")
	marketsCmd.Flags().BoolVar(&marketsActive, "active", true, "filter active markets")
	marketsCmd.Flags().BoolVar(&marketsClosed, "closed", false, "filter closed markets")
	marketsCmd.Flags().BoolVar(&marketsAscending, "ascending", false, "sort ascending")
	marketsCmd.Flags().IntVar(&marketsTagID, "tag-id", 0, "filter by tag id")
	cmd.AddCommand(marketsCmd)

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

	var tagsLimit, tagsOffset int
	var tagID, tagSlug string
	tagsCmd := &cobra.Command{Use: "tags", Short: "List or fetch Gamma tags/categories", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tagID != "" {
				tag, err := w.gamma.TagByID(cmd.Context(), tagID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, tag)
			}
			if tagSlug != "" {
				tag, err := w.gamma.TagBySlug(cmd.Context(), tagSlug)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, tag)
			}
			tags, err := w.gamma.Tags(cmd.Context(), &polytypes.GetTagsParams{
				Limit:  tagsLimit,
				Offset: tagsOffset,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, tags)
		},
	}
	tagsCmd.Flags().StringVar(&tagID, "id", "", "tag ID")
	tagsCmd.Flags().StringVar(&tagSlug, "slug", "", "tag slug")
	tagsCmd.Flags().IntVar(&tagsLimit, "limit", 100, "max tags")
	tagsCmd.Flags().IntVar(&tagsOffset, "offset", 0, "pagination offset")
	cmd.AddCommand(tagsCmd)

	var seriesLimit, seriesOffset int
	var seriesID string
	var seriesClosed bool
	seriesCmd := &cobra.Command{Use: "series", Short: "List or fetch Gamma series", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if seriesID != "" {
				series, err := w.gamma.SeriesByID(cmd.Context(), seriesID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, series)
			}
			series, err := w.gamma.Series(cmd.Context(), &polytypes.GetSeriesParams{
				Limit:  seriesLimit,
				Offset: seriesOffset,
				Closed: &seriesClosed,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, series)
		},
	}
	seriesCmd.Flags().StringVar(&seriesID, "id", "", "series ID")
	seriesCmd.Flags().IntVar(&seriesLimit, "limit", 20, "max series")
	seriesCmd.Flags().IntVar(&seriesOffset, "offset", 0, "pagination offset")
	seriesCmd.Flags().BoolVar(&seriesClosed, "closed", false, "filter closed series")
	cmd.AddCommand(seriesCmd)

	var commentID, commentEntityType, commentUser string
	var commentEntityID, commentLimit, commentOffset int
	commentsCmd := &cobra.Command{Use: "comments", Short: "List or fetch public Gamma comments", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if commentID != "" {
				comment, err := w.gamma.CommentByID(cmd.Context(), commentID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, comment)
			}
			if commentUser != "" {
				comments, err := w.gamma.CommentsByUser(cmd.Context(), commentUser, commentLimit)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, comments)
			}
			params := &polytypes.CommentQuery{Limit: commentLimit, Offset: commentOffset}
			if commentEntityID > 0 {
				params.EntityID = &commentEntityID
			}
			if commentEntityType != "" {
				params.EntityType = &commentEntityType
			}
			comments, err := w.gamma.Comments(cmd.Context(), params)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, comments)
		},
	}
	commentsCmd.Flags().StringVar(&commentID, "id", "", "comment ID")
	commentsCmd.Flags().StringVar(&commentUser, "user", "", "user wallet address")
	commentsCmd.Flags().IntVar(&commentEntityID, "entity-id", 0, "comment parent entity ID")
	commentsCmd.Flags().StringVar(&commentEntityType, "entity-type", "", "comment parent entity type")
	commentsCmd.Flags().IntVar(&commentLimit, "limit", 20, "max comments")
	commentsCmd.Flags().IntVar(&commentOffset, "offset", 0, "pagination offset")
	cmd.AddCommand(commentsCmd)

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
	balanceCmd.Flags().StringVar(&balanceSignatureType, "signature-type", "deposit", "signature type: eoa, proxy, safe, deposit")
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
	updateBalanceCmd.Flags().StringVar(&updateBalanceSignatureType, "signature-type", "deposit", "signature type: eoa, proxy, safe, deposit")
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

	var orderOutput string
	orderCmd := &cobra.Command{Use: "order <order-id>", Short: "Get a single authenticated CLOB order", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(orderOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			row, err := w.clob.Order(cmd.Context(), key, args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, row)
		},
	}
	addOutput(orderCmd, &orderOutput)
	cmd.AddCommand(orderCmd)

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

	var cancelOrdersOutput string
	cancelOrdersCmd := &cobra.Command{Use: "cancel-orders <order-ids>", Short: "Cancel multiple open CLOB orders", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(cancelOrdersOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			ids := splitCSV(args[0])
			resp, err := w.clob.CancelOrders(cmd.Context(), key, ids)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, resp)
		},
	}
	addOutput(cancelOrdersCmd, &cancelOrdersOutput)
	cmd.AddCommand(cancelOrdersCmd)

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

	var cancelMarketOutput, cancelMarketID, cancelMarketAsset string
	cancelMarketCmd := &cobra.Command{Use: "cancel-market", Short: "Cancel open CLOB orders for a market or asset", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(cancelMarketOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			resp, err := w.clob.CancelMarket(cmd.Context(), key, clob.CancelMarketParams{
				Market: cancelMarketID,
				Asset:  cancelMarketAsset,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, resp)
		},
	}
	addOutput(cancelMarketCmd, &cancelMarketOutput)
	cancelMarketCmd.Flags().StringVar(&cancelMarketID, "market", "", "market condition ID")
	cancelMarketCmd.Flags().StringVar(&cancelMarketAsset, "asset", "", "asset/token ID")
	cmd.AddCommand(cancelMarketCmd)

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
	createOrderCmd.Flags().StringVar(&createOrderSignatureType, "signature-type", "deposit", "signature type: eoa, proxy, safe, deposit")
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
	marketOrderCmd.Flags().StringVar(&marketOrderSignatureType, "signature-type", "deposit", "signature type: eoa, proxy, safe, deposit")
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

	var marketsOutput, marketsCursor string
	marketsCmd := &cobra.Command{Use: "markets", Short: "List CLOB markets", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(marketsOutput); err != nil {
				return err
			}
			markets, err := w.clob.Markets(cmd.Context(), marketsCursor)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, markets)
		},
	}
	addOutput(marketsCmd, &marketsOutput)
	marketsCmd.Flags().StringVar(&marketsCursor, "cursor", "", "pagination cursor")
	cmd.AddCommand(marketsCmd)

	return cmd
}

func dataCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("data", "Polymarket Data API analytics")

	var user string
	var tokenID string
	var limit int

	addUser := func(c *cobra.Command) {
		c.Flags().StringVar(&user, "user", "", "user wallet address")
		c.Flags().IntVar(&limit, "limit", 20, "max rows")
	}
	requireUser := func() error {
		if strings.TrimSpace(user) == "" {
			return fmt.Errorf("--user required")
		}
		return nil
	}
	addToken := func(c *cobra.Command) {
		c.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
		c.Flags().IntVar(&limit, "limit", 20, "max rows")
	}
	requireToken := func() error {
		if strings.TrimSpace(tokenID) == "" {
			return fmt.Errorf("--token-id required")
		}
		return nil
	}

	positionsCmd := &cobra.Command{Use: "positions", Short: "List open positions for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.CurrentPositions(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUser(positionsCmd)
	cmd.AddCommand(positionsCmd)

	closedPositionsCmd := &cobra.Command{Use: "closed-positions", Short: "List closed positions for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.ClosedPositions(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUser(closedPositionsCmd)
	cmd.AddCommand(closedPositionsCmd)

	tradesCmd := &cobra.Command{Use: "trades", Short: "List public Data API trades for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.Trades(cmd.Context(), user, limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUser(tradesCmd)
	cmd.AddCommand(tradesCmd)

	activityCmd := &cobra.Command{Use: "activity", Short: "List public activity for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.Activity(cmd.Context(), user, limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUser(activityCmd)
	cmd.AddCommand(activityCmd)

	holdersCmd := &cobra.Command{Use: "holders", Short: "List top holders for a token", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(); err != nil {
				return err
			}
			rows, err := w.data.TopHolders(cmd.Context(), tokenID, limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addToken(holdersCmd)
	cmd.AddCommand(holdersCmd)

	valueCmd := &cobra.Command{Use: "value", Short: "Get total portfolio value for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			value, err := w.data.TotalValue(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	addUser(valueCmd)
	cmd.AddCommand(valueCmd)

	marketsTradedCmd := &cobra.Command{Use: "markets-traded", Short: "Get total markets traded for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			value, err := w.data.MarketsTraded(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	addUser(marketsTradedCmd)
	cmd.AddCommand(marketsTradedCmd)

	openInterestCmd := &cobra.Command{Use: "open-interest", Short: "Get open interest for a token", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(); err != nil {
				return err
			}
			value, err := w.data.OpenInterest(cmd.Context(), tokenID)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	addToken(openInterestCmd)
	cmd.AddCommand(openInterestCmd)

	leaderboardCmd := &cobra.Command{Use: "leaderboard", Short: "List trader leaderboard rows", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := w.data.TraderLeaderboard(cmd.Context(), limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	leaderboardCmd.Flags().IntVar(&limit, "limit", 20, "max rows")
	cmd.AddCommand(leaderboardCmd)

	liveVolumeCmd := &cobra.Command{Use: "live-volume", Short: "Get live volume summary", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := w.data.LiveVolume(cmd.Context(), limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	liveVolumeCmd.Flags().IntVar(&limit, "limit", 20, "max rows")
	cmd.AddCommand(liveVolumeCmd)

	return cmd
}

func streamCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("stream", "Polymarket WebSocket streams")

	var assetsRaw string
	var url string
	var maxMessages int
	marketCmd := &cobra.Command{Use: "market", Short: "Stream public CLOB market events", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assetIDs := splitCSV(assetsRaw)
			if len(assetIDs) == 0 {
				return fmt.Errorf("--asset-ids required")
			}
			cfg := stream.DefaultConfig(url)
			cfg.PingInterval = 10 * time.Second
			client := stream.NewMarketClient(cfg)
			done := make(chan struct{})
			count := 0
			emit := func(v interface{}) {
				if maxMessages > 0 && count >= maxMessages {
					return
				}
				count++
				_ = w.printJSON(cmd, v)
				if maxMessages > 0 && count >= maxMessages {
					close(done)
				}
			}
			client.OnBook = func(msg stream.BookMessage) { emit(msg) }
			client.OnPriceChange = func(msg stream.PriceChangeMessage) { emit(msg) }
			client.OnLastTrade = func(msg stream.LastTradeMessage) { emit(msg) }
			client.OnError = func(err error) {
				if maxMessages == 0 {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "stream error: %v\n", err)
				}
			}
			if err := client.Connect(cmd.Context()); err != nil {
				return err
			}
			defer client.Close()
			if err := client.SubscribeAssets(cmd.Context(), assetIDs); err != nil {
				return err
			}
			select {
			case <-cmd.Context().Done():
				return cmd.Context().Err()
			case <-done:
				return nil
			}
		},
	}
	marketCmd.Flags().StringVar(&assetsRaw, "asset-ids", "", "comma-separated CLOB token IDs")
	marketCmd.Flags().StringVar(&url, "url", marketStreamBaseURL, "WebSocket URL")
	marketCmd.Flags().IntVar(&maxMessages, "max-messages", 0, "stop after this many messages; 0 streams until interrupted")
	cmd.AddCommand(marketCmd)

	return cmd
}

func parseSignatureTypeFlag(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "deposit", "deposit-wallet", "poly-1271", "3":
		return 3, nil
	case "proxy", "1":
		return 1, nil
	case "eoa", "0":
		return 0, nil
	case "safe", "gnosis", "gnosis-safe", "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported signature type %q", value)
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
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
