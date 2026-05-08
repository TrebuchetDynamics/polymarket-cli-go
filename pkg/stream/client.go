// Package stream exposes the public, read-only Polymarket CLOB WebSocket SDK.
//
// Use stream when you need market-channel updates for CLOB token IDs:
// order-book snapshots, price changes, and last-trade events. This package is
// public market data only. It does not implement authenticated user streams or
// request L2 WebSocket credentials.
package stream

import (
	"context"
	"time"

	internalstream "github.com/TrebuchetDynamics/polygolem/internal/stream"
)

const defaultMarketURL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"

// Config holds WebSocket connection configuration.
type Config struct {
	URL               string
	PingInterval      time.Duration
	PongTimeout       time.Duration
	Reconnect         bool
	ReconnectDelay    time.Duration
	ReconnectMaxDelay time.Duration
	ReconnectMax      int
}

// DefaultConfig returns production market-stream defaults. Pass an empty URL
// to use the Polymarket production market-channel endpoint.
func DefaultConfig(url string) Config {
	if url == "" {
		url = defaultMarketURL
	}
	cfg := internalstream.DefaultConfig(url)
	return configFromInternal(cfg)
}

// MarketClient manages one public market WebSocket connection.
type MarketClient struct {
	inner *internalstream.MarketClient

	OnBook        func(BookMessage)
	OnPriceChange func(PriceChangeMessage)
	OnLastTrade   func(LastTradeMessage)
	OnError       func(error)
}

// NewMarketClient creates a public market WebSocket client. A zero-valued
// Config uses production defaults.
func NewMarketClient(cfg Config) *MarketClient {
	if cfg.URL == "" {
		cfg = DefaultConfig("")
	}
	client := &MarketClient{}
	inner := internalstream.NewMarketClient(configToInternal(cfg))
	inner.OnBook = func(msg internalstream.BookMessage) {
		if client.OnBook != nil {
			client.OnBook(bookFromInternal(msg))
		}
	}
	inner.OnPriceChange = func(msg internalstream.PriceChangeMessage) {
		if client.OnPriceChange != nil {
			client.OnPriceChange(priceChangeFromInternal(msg))
		}
	}
	inner.OnLastTrade = func(msg internalstream.LastTradeMessage) {
		if client.OnLastTrade != nil {
			client.OnLastTrade(lastTradeFromInternal(msg))
		}
	}
	inner.OnError = func(err error) {
		if client.OnError != nil {
			client.OnError(err)
		}
	}
	client.inner = inner
	return client
}

// Connect establishes the WebSocket connection.
func (c *MarketClient) Connect(ctx context.Context) error {
	return c.inner.Connect(ctx)
}

// SubscribeAssets subscribes to public market events for CLOB token IDs.
func (c *MarketClient) SubscribeAssets(ctx context.Context, assetIDs []string) error {
	return c.inner.SubscribeAssets(ctx, assetIDs)
}

// Close shuts down the WebSocket connection.
func (c *MarketClient) Close() {
	c.inner.Close()
}

// IsConnected returns the current connection state.
func (c *MarketClient) IsConnected() bool {
	return c.inner.IsConnected()
}

// BookMessage is a WebSocket order-book snapshot event.
type BookMessage struct {
	EventType string       `json:"event_type"`
	AssetID   string       `json:"asset_id"`
	Market    string       `json:"market"`
	Timestamp string       `json:"timestamp"`
	Hash      string       `json:"hash"`
	Bids      []PriceLevel `json:"bids"`
	Asks      []PriceLevel `json:"asks"`
}

// PriceLevel is a single price level in a book event.
type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// PriceChangeMessage is a WebSocket order-book price-level update event.
type PriceChangeMessage struct {
	EventType    string             `json:"event_type"`
	Market       string             `json:"market"`
	PriceChanges []PriceChangeEntry `json:"price_changes"`
	Timestamp    string             `json:"timestamp"`
}

// PriceChangeEntry is one price-level update.
type PriceChangeEntry struct {
	AssetID string `json:"asset_id"`
	Price   string `json:"price"`
	Side    string `json:"side"`
	Size    string `json:"size"`
	Hash    string `json:"hash"`
	BestBid string `json:"best_bid,omitempty"`
	BestAsk string `json:"best_ask,omitempty"`
}

// LastTradeMessage is a WebSocket trade execution event.
type LastTradeMessage struct {
	EventType       string `json:"event_type"`
	AssetID         string `json:"asset_id"`
	Market          string `json:"market"`
	Price           string `json:"price"`
	Side            string `json:"side"`
	Size            string `json:"size"`
	FeeRateBps      string `json:"fee_rate_bps"`
	Timestamp       string `json:"timestamp"`
	TransactionHash string `json:"transaction_hash,omitempty"`
}

// Deduplicator removes duplicate raw WebSocket messages.
type Deduplicator struct {
	inner *internalstream.Deduplicator
}

// NewDeduplicator creates a raw-message deduplicator with capacity and TTL.
func NewDeduplicator(size int, ttl time.Duration) *Deduplicator {
	return &Deduplicator{inner: internalstream.NewDeduplicator(size, ttl)}
}

// Process returns true when data has not been seen recently.
func (d *Deduplicator) Process(data []byte) bool {
	return d.inner.Process(data)
}

// Reset clears the deduplication cache.
func (d *Deduplicator) Reset() {
	d.inner.Reset()
}

// Stats returns input, duplicate, and output counters.
func (d *Deduplicator) Stats() (in, dup, out int64) {
	return d.inner.Stats()
}

func configToInternal(cfg Config) internalstream.Config {
	return internalstream.Config{
		URL:               cfg.URL,
		PingInterval:      cfg.PingInterval,
		PongTimeout:       cfg.PongTimeout,
		Reconnect:         cfg.Reconnect,
		ReconnectDelay:    cfg.ReconnectDelay,
		ReconnectMaxDelay: cfg.ReconnectMaxDelay,
		ReconnectMax:      cfg.ReconnectMax,
	}
}

func configFromInternal(cfg internalstream.Config) Config {
	return Config{
		URL:               cfg.URL,
		PingInterval:      cfg.PingInterval,
		PongTimeout:       cfg.PongTimeout,
		Reconnect:         cfg.Reconnect,
		ReconnectDelay:    cfg.ReconnectDelay,
		ReconnectMaxDelay: cfg.ReconnectMaxDelay,
		ReconnectMax:      cfg.ReconnectMax,
	}
}

func bookFromInternal(msg internalstream.BookMessage) BookMessage {
	return BookMessage{
		EventType: msg.EventType,
		AssetID:   msg.AssetID,
		Market:    msg.Market,
		Timestamp: msg.Timestamp,
		Hash:      msg.Hash,
		Bids:      levelsFromInternal(msg.Bids),
		Asks:      levelsFromInternal(msg.Asks),
	}
}

func levelsFromInternal(rows []internalstream.PriceLevel) []PriceLevel {
	out := make([]PriceLevel, len(rows))
	for i, row := range rows {
		out[i] = PriceLevel{
			Price: row.Price,
			Size:  row.Size,
		}
	}
	return out
}

func priceChangeFromInternal(msg internalstream.PriceChangeMessage) PriceChangeMessage {
	return PriceChangeMessage{
		EventType:    msg.EventType,
		Market:       msg.Market,
		PriceChanges: priceChangeEntriesFromInternal(msg.Changes),
		Timestamp:    msg.Timestamp,
	}
}

func priceChangeEntriesFromInternal(rows []internalstream.PriceChangeEntry) []PriceChangeEntry {
	out := make([]PriceChangeEntry, len(rows))
	for i, row := range rows {
		out[i] = PriceChangeEntry{
			AssetID: row.AssetID,
			Price:   row.Price,
			Side:    row.Side,
			Size:    row.Size,
			Hash:    row.Hash,
			BestBid: row.BestBid,
			BestAsk: row.BestAsk,
		}
	}
	return out
}

func lastTradeFromInternal(msg internalstream.LastTradeMessage) LastTradeMessage {
	return LastTradeMessage{
		EventType:       msg.EventType,
		AssetID:         msg.AssetID,
		Market:          msg.Market,
		Price:           msg.Price,
		Side:            msg.Side,
		Size:            msg.Size,
		FeeRateBps:      msg.FeeRateBps,
		Timestamp:       msg.Timestamp,
		TransactionHash: msg.TransactionHash,
	}
}
