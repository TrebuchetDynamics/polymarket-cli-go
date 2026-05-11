package cli

import (
	"fmt"
	"time"

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

	return cmd
}
