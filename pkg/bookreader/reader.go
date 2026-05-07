// Package bookreader provides a public BookReader interface and Polygolem implementation.
// This is the Phase 0 boundary between go-bot and polygolem — replaces direct CLOB clients.
package bookreader

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

// OrderBook is the public order book type usable by go-bot.
type OrderBook struct {
	MarketID       string
	TokenID        string
	Bids           []Level
	Asks           []Level
	LastTradePrice float64
}

// Level is a single price level.
type Level struct {
	Price float64
	Size  float64
}

// Reader fetches CLOB order books.
type Reader interface {
	OrderBook(ctx context.Context, tokenID string) (OrderBook, error)
}

// NewReader creates a BookReader backed by the Polygolem CLOB client.
func NewReader(clobBaseURL string) Reader {
	return &reader{client: clob.NewClient(clobBaseURL, nil)}
}

type reader struct {
	client *clob.Client
}

func (r *reader) OrderBook(ctx context.Context, tokenID string) (OrderBook, error) {
	pb, err := r.client.OrderBook(ctx, tokenID)
	if err != nil {
		return OrderBook{}, fmt.Errorf("polygolem book: %w", err)
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
