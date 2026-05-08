// Package orderbook is a read-only Polymarket CLOB order-book reader.
//
// Use orderbook when you want top-of-book price discovery for one or more
// token IDs without pulling in the full polygolem CLI. The Reader interface is
// the only public entry point; NewReader returns a production implementation
// backed by the CLOB HTTP API.
//
// When not to use this package:
//   - For authenticated CLOB operations (create or cancel orders) — use
//     pkg/universal or the CLI.
//   - For low-latency streaming — use a WebSocket client instead.
//
// Stability: the Reader interface, OrderBook, Level, and NewReader are part of
// the polygolem public SDK and follow semver. Internal helpers remain
// unexported and may change.
package orderbook

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

// OrderBook is a snapshot of one Polymarket CLOB market.
// Bids are sorted highest-price first and Asks lowest-price first.
// LastTradePrice may be zero if the snapshot does not include a trade
// reference.
type OrderBook struct {
	MarketID       string
	TokenID        string
	Bids           []Level
	Asks           []Level
	LastTradePrice float64
}

// Level is one price level in the order book — a single Price with the total
// Size resting at that price.
type Level struct {
	Price float64
	Size  float64
}

// Reader fetches CLOB order books by ERC-1155 token ID.
// Implementations must be safe for concurrent use by multiple goroutines.
type Reader interface {
	// OrderBook returns the current order-book snapshot for tokenID.
	// The returned OrderBook is sorted best-first on each side.
	OrderBook(ctx context.Context, tokenID string) (OrderBook, error)
}

// NewReader returns a Reader backed by the polygolem CLOB client at
// clobBaseURL. Pass an empty string to use the Polymarket production CLOB URL.
// The returned Reader uses the package's default HTTP transport with retry and
// rate limiting.
func NewReader(clobBaseURL string) Reader {
	return &reader{client: clob.NewClient(clobBaseURL, nil)}
}

type reader struct {
	client *clob.Client
}

func (r *reader) OrderBook(ctx context.Context, tokenID string) (OrderBook, error) {
	pb, err := r.client.OrderBook(ctx, tokenID)
	if err != nil {
		return OrderBook{}, fmt.Errorf("polygolem orderbook: %w", err)
	}
	return convertBook(pb), nil
}

func convertBook(pb *polytypes.OrderBook) OrderBook {
	if pb == nil {
		return OrderBook{}
	}
	bids := convertLevels(pb.Bids)
	asks := convertLevels(pb.Asks)
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].Price > bids[j].Price
	})
	sort.Slice(asks, func(i, j int) bool {
		return asks[i].Price < asks[j].Price
	})
	return OrderBook{
		MarketID: pb.Market,
		TokenID:  pb.AssetID,
		Bids:     bids,
		Asks:     asks,
	}
}

func convertLevels(levels []polytypes.OrderBookLevel) []Level {
	out := make([]Level, 0, len(levels))
	for _, lv := range levels {
		out = append(out, Level{
			Price: parseFloat(lv.Price),
			Size:  parseFloat(lv.Size),
		})
	}
	return out
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
