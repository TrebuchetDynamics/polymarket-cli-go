package cli

import (
	"fmt"
	"time"

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
	return cmd
}
