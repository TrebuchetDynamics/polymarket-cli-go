package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

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

	var createKeyForAddressOutput, createKeyForAddressOwner string
	createKeyForAddressCmd := &cobra.Command{Use: "create-api-key-for-address", Short: "Create CLOB API credentials while reporting a maker address", Args: cobra.NoArgs,
		Long: `Creates CLOB L2 credentials using EOA L1 auth and echoes the
configured maker address. Polymarket login and CLOB HTTP authentication sign
with the EOA; the deposit wallet remains the POLY_1271 trading wallet inside
orders, balances, approvals, and settlement.

The --owner flag is retained for source compatibility with older automation
and is returned in the JSON output. It is not used as POLY_ADDRESS for the
CLOB L1 auth headers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(createKeyForAddressOutput); err != nil {
				return err
			}
			owner := strings.TrimSpace(createKeyForAddressOwner)
			if owner == "" {
				return fmt.Errorf("--owner is required")
			}
			if !common.IsHexAddress(owner) {
				return fmt.Errorf("--owner must be an Ethereum address")
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			apiKey, err := w.clob.CreateAPIKeyForAddress(cmd.Context(), key, owner)
			if err != nil {
				return fmt.Errorf("%w\n\nNote: Polymarket login and CLOB HTTP auth sign with the EOA; the deposit wallet remains the POLY_1271 trading wallet. Run `polygolem auth login` first if this EOA has not been profiled. See docs/ONBOARDING.md", err)
			}
			return w.printJSON(cmd, map[string]string{"api_key": apiKey.Key, "owner": common.HexToAddress(owner).Hex()})
		},
	}
	addOutput(createKeyForAddressCmd, &createKeyForAddressOutput)
	createKeyForAddressCmd.Flags().StringVar(&createKeyForAddressOwner, "owner", "", "deposit wallet owner address")
	cmd.AddCommand(createKeyForAddressCmd)

	var createBuilderFeeKeyOutput string
	createBuilderFeeKeyCmd := &cobra.Command{
		Use:   "create-builder-fee-key",
		Short: "Mint a CLOB builder fee key (POST /auth/builder-api-key)",
		Long: `Mints a builder fee key by signing an L2 HMAC-authenticated
POST to /auth/builder-api-key. The returned triple is the fee
attribution key — attach its 'key' to the 'builder' bytes32 field of V2
orders to claim integrator fees.

This is a different credential from the L2 trading triple minted by
'create-api-key'; both are needed for full V2 integrator setup. See
docs/HEADLESS-BUILDER-KEYS-INVESTIGATION.md.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(createBuilderFeeKeyOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			feeKey, err := w.clob.CreateBuilderFeeKey(cmd.Context(), key)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]string{"builder_fee_key": feeKey.Key})
		},
	}
	addOutput(createBuilderFeeKeyCmd, &createBuilderFeeKeyOutput)
	cmd.AddCommand(createBuilderFeeKeyCmd)

	var listBuilderFeeKeysOutput string
	listBuilderFeeKeysCmd := &cobra.Command{
		Use:   "list-builder-fee-keys",
		Short: "List builder fee keys (GET /auth/builder-api-keys)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(listBuilderFeeKeysOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			records, err := w.clob.ListBuilderFeeKeys(cmd.Context(), key)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, records)
		},
	}
	addOutput(listBuilderFeeKeysCmd, &listBuilderFeeKeysOutput)
	cmd.AddCommand(listBuilderFeeKeysCmd)

	var revokeBuilderFeeKeyOutput, revokeBuilderFeeKey string
	revokeBuilderFeeKeyCmd := &cobra.Command{
		Use:   "revoke-builder-fee-key",
		Short: "Revoke a builder fee key (DELETE /auth/builder-api-key/{key})",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(revokeBuilderFeeKeyOutput); err != nil {
				return err
			}
			if strings.TrimSpace(revokeBuilderFeeKey) == "" {
				return fmt.Errorf("--key is required")
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			if err := w.clob.RevokeBuilderFeeKey(cmd.Context(), key, revokeBuilderFeeKey); err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]string{"revoked": revokeBuilderFeeKey})
		},
	}
	addOutput(revokeBuilderFeeKeyCmd, &revokeBuilderFeeKeyOutput)
	revokeBuilderFeeKeyCmd.Flags().StringVar(&revokeBuilderFeeKey, "key", "", "builder fee key to revoke")
	cmd.AddCommand(revokeBuilderFeeKeyCmd)

	var balanceOutput, balanceAssetType, balanceTokenID string
	balanceCmd := &cobra.Command{Use: "balance", Short: "Get CLOB balance and allowances", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(balanceOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			res, err := w.clob.BalanceAllowance(cmd.Context(), key, clob.BalanceAllowanceParams{
				AssetType: balanceAssetType,
				TokenID:   balanceTokenID,
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
	cmd.AddCommand(balanceCmd)

	var updateBalanceOutput, updateBalanceAssetType, updateBalanceTokenID string
	updateBalanceCmd := &cobra.Command{Use: "update-balance", Short: "Refresh CLOB balance and allowances", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(updateBalanceOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			res, err := w.clob.UpdateBalanceAllowance(cmd.Context(), key, clob.BalanceAllowanceParams{
				AssetType: updateBalanceAssetType,
				TokenID:   updateBalanceTokenID,
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

	var probeOutput, probeMarket, probeAssetID, probeCursor string
	probeCmd := &cobra.Command{
		Use:   "market-trades-probe",
		Short: "Probe CLOB trade scope for one market or token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(probeOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			res, err := w.clob.MarketTradesProbe(cmd.Context(), key, clob.MarketTradesProbeRequest{
				Market:     probeMarket,
				AssetID:    probeAssetID,
				NextCursor: probeCursor,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, res)
		},
	}
	addOutput(probeCmd, &probeOutput)
	probeCmd.Flags().StringVar(&probeMarket, "market", "", "market condition ID")
	probeCmd.Flags().StringVar(&probeAssetID, "asset-id", "", "CLOB token ID")
	probeCmd.Flags().StringVar(&probeCursor, "cursor", "", "optional next_cursor for diagnostics")
	cmd.AddCommand(probeCmd)

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

	var createOrderOutput, createOrderToken, createOrderSide, createOrderPrice, createOrderSize, createOrderType, createOrderExpiration, createOrderBuilderCode string
	var createOrderPostOnly bool
	createOrderCmd := &cobra.Command{Use: "create-order", Short: "Create a signed CLOB limit order", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(createOrderOutput); err != nil {
				return err
			}
			builderCode := builderCodeFromFlagOrEnv(createOrderBuilderCode)
			if err := validateBuilderCodeForCLI(builderCode); err != nil {
				return err
			}
			w.clob.SetBuilderCode(builderCode)
			key, err := privateKey()
			if err != nil {
				return err
			}
			warnIfNoDepositKey(cmd.Context(), cmd.ErrOrStderr(), key)

			res, err := w.clob.CreateLimitOrder(cmd.Context(), key, clob.CreateOrderParams{
				TokenID:    createOrderToken,
				Side:       createOrderSide,
				Price:      createOrderPrice,
				Size:       createOrderSize,
				OrderType:  createOrderType,
				Expiration: createOrderExpiration,
				PostOnly:   createOrderPostOnly,
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
	createOrderCmd.Flags().StringVar(&createOrderExpiration, "expiration", "0", "unix timestamp for GTD orders (0 = no expiration)")
	createOrderCmd.Flags().StringVar(&createOrderBuilderCode, "builder-code", "", "0x-prefixed bytes32 builder attribution code")
	createOrderCmd.Flags().BoolVar(&createOrderPostOnly, "post-only", false, "post-only order (maker-only, rejected if it would take)")
	cmd.AddCommand(createOrderCmd)

	var batchOrdersOutput, batchOrdersFile, batchOrdersBuilderCode string
	batchOrdersCmd := &cobra.Command{Use: "batch-orders", Short: "Create multiple signed CLOB limit orders", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(batchOrdersOutput); err != nil {
				return err
			}
			builderCode := builderCodeFromFlagOrEnv(batchOrdersBuilderCode)
			if err := validateBuilderCodeForCLI(builderCode); err != nil {
				return err
			}
			if strings.TrimSpace(batchOrdersFile) == "" {
				return fmt.Errorf("--orders-file is required")
			}
			reader, closeReader, err := openBatchOrdersInput(cmd, batchOrdersFile)
			if err != nil {
				return err
			}
			if closeReader != nil {
				defer closeReader()
			}
			orders, err := parseBatchOrderParams(reader)
			if err != nil {
				return err
			}
			w.clob.SetBuilderCode(builderCode)
			key, err := privateKey()
			if err != nil {
				return err
			}
			warnIfNoDepositKey(cmd.Context(), cmd.ErrOrStderr(), key)

			res, err := w.clob.CreateBatchOrders(cmd.Context(), key, orders)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, res)
		},
	}
	addOutput(batchOrdersCmd, &batchOrdersOutput)
	batchOrdersCmd.Flags().StringVar(&batchOrdersFile, "orders-file", "", "JSON array of limit orders, or '-' for stdin")
	batchOrdersCmd.Flags().StringVar(&batchOrdersBuilderCode, "builder-code", "", "0x-prefixed bytes32 builder attribution code")
	cmd.AddCommand(batchOrdersCmd)

	var marketOrderOutput, marketOrderToken, marketOrderSide, marketOrderAmount, marketOrderPrice, marketOrderType, marketOrderBuilderCode string
	marketOrderCmd := &cobra.Command{Use: "market-order", Short: "Create a signed CLOB market/FOK order", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(marketOrderOutput); err != nil {
				return err
			}
			builderCode := builderCodeFromFlagOrEnv(marketOrderBuilderCode)
			if err := validateBuilderCodeForCLI(builderCode); err != nil {
				return err
			}
			w.clob.SetBuilderCode(builderCode)
			key, err := privateKey()
			if err != nil {
				return err
			}
			warnIfNoDepositKey(cmd.Context(), cmd.ErrOrStderr(), key)

			res, err := w.clob.CreateMarketOrder(cmd.Context(), key, clob.MarketOrderParams{
				TokenID:   marketOrderToken,
				Side:      marketOrderSide,
				Amount:    marketOrderAmount,
				Price:     marketOrderPrice,
				OrderType: marketOrderType,
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
	marketOrderCmd.Flags().StringVar(&marketOrderBuilderCode, "builder-code", "", "0x-prefixed bytes32 builder attribution code")
	cmd.AddCommand(marketOrderCmd)

	var heartbeatOutput, heartbeatID string
	heartbeatCmd := &cobra.Command{Use: "heartbeat", Short: "Send one CLOB heartbeat ping", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(heartbeatOutput); err != nil {
				return err
			}
			key, err := privateKey()
			if err != nil {
				return err
			}
			if err := w.clob.Heartbeat(cmd.Context(), key, heartbeatID); err != nil {
				return err
			}
			return w.printJSON(cmd, map[string]interface{}{
				"ok":           true,
				"heartbeat_id": heartbeatID,
			})
		},
	}
	addOutput(heartbeatCmd, &heartbeatOutput)
	heartbeatCmd.Flags().StringVar(&heartbeatID, "id", "", "optional heartbeat id")
	cmd.AddCommand(heartbeatCmd)

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

	var marketByTokenOutput string
	marketByTokenCmd := &cobra.Command{Use: "market-by-token <token-id>", Short: "Resolve CLOB market by token ID", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(marketByTokenOutput); err != nil {
				return err
			}
			market, err := w.clob.MarketByToken(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, market)
		},
	}
	addOutput(marketByTokenCmd, &marketByTokenOutput)
	cmd.AddCommand(marketByTokenCmd)

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
