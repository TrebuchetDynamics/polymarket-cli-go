package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/stream"
	"github.com/spf13/cobra"
)

func streamCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("stream", "Polymarket WebSocket streams")

	var assetsRaw string
	var url string
	var maxMessages int
	var customFeatures bool
	var level int
	marketCmd := &cobra.Command{Use: "market", Short: "Stream public CLOB market events", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assetIDs := splitCSV(assetsRaw)
			if len(assetIDs) == 0 {
				return fmt.Errorf("--asset-ids required")
			}
			cfg := stream.DefaultConfig(url)
			cfg.PingInterval = 10 * time.Second
			cfg.CustomFeatureEnabled = customFeatures
			cfg.Level = level
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
			client.OnTickSizeChange = func(msg stream.TickSizeChangeMessage) { emit(msg) }
			client.OnBestBidAsk = func(msg stream.BestBidAskMessage) { emit(msg) }
			client.OnNewMarket = func(msg stream.NewMarketMessage) { emit(msg) }
			client.OnMarketResolved = func(msg stream.MarketResolvedMessage) { emit(msg) }
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
	marketCmd.Flags().BoolVar(&customFeatures, "custom-features", false, "request best-bid-ask and market lifecycle events")
	marketCmd.Flags().IntVar(&level, "level", 0, "optional Polymarket market-stream subscription level")
	cmd.AddCommand(marketCmd)

	var cryptoStreamAsset string
	var cryptoStreamInterval string
	var cryptoStreamMaxMsgs int
	cryptoCmd := &cobra.Command{
		Use:   "crypto",
		Short: "Stream live crypto market events",
		Long: `Discover active crypto markets and stream their WebSocket events in real-time.

Auto-discovers crypto markets by asset and interval, extracts token IDs, and
subscribes to the CLOB market stream for live order book and price updates.

Examples:
  polygolem stream crypto --asset BTC --interval 5m          # Stream BTC 5m markets
  polygolem stream crypto --asset ETH --max-messages 100     # Stream ETH markets
  polygolem stream crypto --asset SOL --custom-features      # With best-bid-ask events`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			searchQuery := cryptoStreamAsset
			if cryptoStreamInterval != "" {
				if searchQuery != "" {
					searchQuery += " "
				}
				searchQuery += cryptoStreamInterval
			}
			if searchQuery == "" {
				searchQuery = "crypto"
			}
			searchLimit := 50
			resp, err := w.gamma.Search(cmd.Context(), &polytypes.SearchParams{
				Q:            searchQuery,
				LimitPerType: &searchLimit,
			})
			if err != nil {
				return err
			}

			var tokenIDs []string
			for _, event := range resp.Events {
				if !event.Active || event.Closed {
					continue
				}
				for _, market := range event.Markets {
					if !market.Active || market.Closed {
						continue
					}
					if cryptoStreamAsset != "" &&
						!strings.Contains(strings.ToUpper(market.Question), strings.ToUpper(cryptoStreamAsset)) &&
						!strings.Contains(strings.ToUpper(event.Title), strings.ToUpper(cryptoStreamAsset)) {
						continue
					}
					if cryptoStreamInterval != "" &&
						!strings.Contains(strings.ToLower(event.Title), strings.ToLower(cryptoStreamInterval)) &&
						!strings.Contains(strings.ToLower(market.Question), strings.ToLower(cryptoStreamInterval)) {
						continue
					}
					tokenIDs = append(tokenIDs, parseClobTokenIDs(market.ClobTokenIDs)...)
				}
			}

			if len(tokenIDs) == 0 {
				return fmt.Errorf("no active crypto markets found for asset=%s interval=%s", cryptoStreamAsset, cryptoStreamInterval)
			}

			cfg := stream.DefaultConfig(url)
			cfg.PingInterval = 10 * time.Second
			cfg.CustomFeatureEnabled = customFeatures
			client := stream.NewMarketClient(cfg)
			done := make(chan struct{})
			count := 0
			emit := func(v interface{}) {
				if cryptoStreamMaxMsgs > 0 && count >= cryptoStreamMaxMsgs {
					return
				}
				count++
				_ = w.printJSON(cmd, v)
				if cryptoStreamMaxMsgs > 0 && count >= cryptoStreamMaxMsgs {
					close(done)
				}
			}
			client.OnBook = func(msg stream.BookMessage) { emit(msg) }
			client.OnPriceChange = func(msg stream.PriceChangeMessage) { emit(msg) }
			client.OnLastTrade = func(msg stream.LastTradeMessage) { emit(msg) }
			client.OnTickSizeChange = func(msg stream.TickSizeChangeMessage) { emit(msg) }
			client.OnBestBidAsk = func(msg stream.BestBidAskMessage) { emit(msg) }
			client.OnNewMarket = func(msg stream.NewMarketMessage) { emit(msg) }
			client.OnMarketResolved = func(msg stream.MarketResolvedMessage) { emit(msg) }
			client.OnError = func(err error) {
				if cryptoStreamMaxMsgs == 0 {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "stream error: %v\n", err)
				}
			}

			_ = w.printJSON(cmd, map[string]interface{}{
				"status":    "connecting",
				"markets":   len(tokenIDs),
				"asset":     cryptoStreamAsset,
				"interval":  cryptoStreamInterval,
				"token_ids": tokenIDs,
			})

			if err := client.Connect(cmd.Context()); err != nil {
				return err
			}
			defer client.Close()
			if err := client.SubscribeAssets(cmd.Context(), tokenIDs); err != nil {
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
	cryptoCmd.Flags().StringVar(&cryptoStreamAsset, "asset", "", "crypto asset filter (BTC, ETH, SOL, XRP, DOGE, BNB, HYPE)")
	cryptoCmd.Flags().StringVar(&cryptoStreamInterval, "interval", "", "interval filter (5m, 15m, 1h)")
	cryptoCmd.Flags().IntVar(&cryptoStreamMaxMsgs, "max-messages", 0, "stop after this many messages; 0 streams until interrupted")
	cryptoCmd.Flags().BoolVar(&customFeatures, "custom-features", false, "request best-bid-ask and market lifecycle events")
	cmd.AddCommand(cryptoCmd)

	return cmd
}
