// Package bookreader is a deprecated compatibility wrapper for pkg/orderbook.
//
// Deprecated: use github.com/TrebuchetDynamics/polygolem/pkg/orderbook.
package bookreader

import "github.com/TrebuchetDynamics/polygolem/pkg/orderbook"

// OrderBook is a snapshot of one Polymarket CLOB market.
//
// Deprecated: use orderbook.OrderBook.
type OrderBook = orderbook.OrderBook

// Level is one price level in the order book.
//
// Deprecated: use orderbook.Level.
type Level = orderbook.Level

// Reader fetches CLOB order books by ERC-1155 token ID.
//
// Deprecated: use orderbook.Reader.
type Reader = orderbook.Reader

// NewReader returns a Reader backed by the polygolem CLOB client.
//
// Deprecated: use orderbook.NewReader.
func NewReader(clobBaseURL string) Reader {
	return orderbook.NewReader(clobBaseURL)
}
