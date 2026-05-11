package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

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
