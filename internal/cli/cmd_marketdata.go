package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/pkg/marketdata"
	sdkstream "github.com/TrebuchetDynamics/polygolem/pkg/stream"
	"github.com/spf13/cobra"
)

func marketDataCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("marketdata", "Live CLOB orderbook and share-price snapshots")

	var assetsRaw string
	var url string
	var maxMessages int
	var customFeatures bool
	var level int
	liveCmd := &cobra.Command{Use: "live", Short: "Stream enriched CLOB market-data snapshots", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assetIDs := splitCSV(assetsRaw)
			if len(assetIDs) == 0 {
				return fmt.Errorf("--asset-ids required")
			}
			cfg := sdkstream.DefaultConfig(url)
			cfg.PingInterval = 10 * time.Second
			cfg.CustomFeatureEnabled = customFeatures
			cfg.Level = level
			client := sdkstream.NewMarketClient(cfg)
			tracker := marketdata.NewTracker()
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
			client.OnBook = func(msg sdkstream.BookMessage) {
				emit(tracker.ApplyBook(msg))
			}
			client.OnPriceChange = func(msg sdkstream.PriceChangeMessage) {
				for _, snapshot := range tracker.ApplyPriceChange(msg) {
					emit(snapshot)
				}
			}
			client.OnLastTrade = func(msg sdkstream.LastTradeMessage) {
				emit(tracker.ApplyLastTrade(msg))
			}
			client.OnBestBidAsk = func(msg sdkstream.BestBidAskMessage) {
				emit(tracker.ApplyBestBidAsk(msg))
			}
			client.OnTickSizeChange = func(msg sdkstream.TickSizeChangeMessage) {
				emit(tracker.ApplyTickSizeChange(msg))
			}
			client.OnError = func(err error) {
				if maxMessages == 0 {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "marketdata stream error: %v\n", err)
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
	liveCmd.Flags().StringVar(&assetsRaw, "asset-ids", "", "comma-separated CLOB token IDs")
	liveCmd.Flags().StringVar(&url, "url", marketStreamBaseURL, "WebSocket URL")
	liveCmd.Flags().IntVar(&maxMessages, "max-messages", 0, "stop after this many snapshots; 0 streams until interrupted")
	liveCmd.Flags().BoolVar(&customFeatures, "custom-features", true, "request best-bid-ask and market lifecycle events")
	liveCmd.Flags().IntVar(&level, "level", 0, "optional Polymarket market-stream subscription level")
	cmd.AddCommand(liveCmd)

	var cryptoAsset string
	var cryptoInterval string
	var cryptoLimit int
	cryptoCmd := &cobra.Command{
		Use:   "crypto",
		Short: "Get live marketdata snapshots for crypto markets",
		Long: `Discover crypto markets and fetch current CLOB snapshots (price, spread,
order book) for each. Returns a single snapshot per market — no continuous stream.

Examples:
  polygolem marketdata crypto --asset BTC --interval 5m    # BTC 5m snapshots
  polygolem marketdata crypto --asset ETH --limit 10       # ETH market snapshots`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			searchQuery := cryptoAsset
			if cryptoInterval != "" {
				if searchQuery != "" {
					searchQuery += " "
				}
				searchQuery += cryptoInterval
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

			type cryptoSnapshot struct {
				EventID         string  `json:"event_id"`
				EventTitle      string  `json:"event_title"`
				MarketID        string  `json:"market_id"`
				Question        string  `json:"question"`
				TokenID         string  `json:"token_id"`
				Outcome         string  `json:"outcome"`
				Price           string  `json:"price"`
				Spread          string  `json:"spread"`
				Midpoint        string  `json:"midpoint"`
				TickSize        string  `json:"tick_size"`
				Volume24hr      float64 `json:"volume_24h"`
				AcceptingOrders bool    `json:"accepting_orders"`
				EndDate         string  `json:"end_date"`
			}

			var results []cryptoSnapshot
			for _, event := range resp.Events {
				if !event.Active || event.Closed {
					continue
				}
				for _, market := range event.Markets {
					if !market.Active || market.Closed {
						continue
					}
					if cryptoAsset != "" &&
						!strings.Contains(strings.ToUpper(market.Question), strings.ToUpper(cryptoAsset)) &&
						!strings.Contains(strings.ToUpper(event.Title), strings.ToUpper(cryptoAsset)) {
						continue
					}
					if cryptoInterval != "" &&
						!strings.Contains(strings.ToLower(event.Title), strings.ToLower(cryptoInterval)) &&
						!strings.Contains(strings.ToLower(market.Question), strings.ToLower(cryptoInterval)) {
						continue
					}

					tokenIDs := parseClobTokenIDs(market.ClobTokenIDs)
					if len(tokenIDs) == 0 {
						continue
					}
					tokenID := tokenIDs[0]

					cs := cryptoSnapshot{
						EventID:         event.ID,
						EventTitle:      event.Title,
						MarketID:        market.ID,
						Question:        market.Question,
						TokenID:         tokenID,
						Outcome:         "",
						Volume24hr:      market.Volume24hr,
						AcceptingOrders: market.AcceptingOrders,
						EndDate:         market.EndDateISO,
					}
					if outcomes := []string(market.Outcomes); len(outcomes) > 0 {
						cs.Outcome = outcomes[0]
					}
					if price, err := w.clob.Price(cmd.Context(), tokenID, "BUY"); err == nil {
						cs.Price = price
					}
					if spread, err := w.clob.Spread(cmd.Context(), tokenID); err == nil {
						cs.Spread = spread
					}
					if midpoint, err := w.clob.Midpoint(cmd.Context(), tokenID); err == nil {
						cs.Midpoint = midpoint
					}
					if tick, err := w.clob.TickSize(cmd.Context(), tokenID); err == nil && tick != nil {
						cs.TickSize = tick.MinimumTickSize
					}

					results = append(results, cs)
					if len(results) >= cryptoLimit {
						break
					}
				}
				if len(results) >= cryptoLimit {
					break
				}
			}

			return w.printJSON(cmd, map[string]interface{}{
				"query":    searchQuery,
				"asset":    cryptoAsset,
				"interval": cryptoInterval,
				"count":    len(results),
				"markets":  results,
			})
		},
	}
	cryptoCmd.Flags().StringVar(&cryptoAsset, "asset", "", "crypto asset filter (BTC, ETH, SOL, XRP, DOGE, BNB, HYPE)")
	cryptoCmd.Flags().StringVar(&cryptoInterval, "interval", "", "interval filter (5m, 15m, 1h)")
	cryptoCmd.Flags().IntVar(&cryptoLimit, "limit", 20, "max markets")
	cmd.AddCommand(cryptoCmd)

	return cmd
}
