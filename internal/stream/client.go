package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Config holds WebSocket connection configuration.
type Config struct {
	URL                string
	PingInterval       time.Duration
	PongTimeout        time.Duration
	Reconnect          bool
	ReconnectDelay     time.Duration
	ReconnectMaxDelay  time.Duration
	ReconnectMax       int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig(url string) Config {
	return Config{
		URL:               url,
		PingInterval:      10 * time.Second,
		PongTimeout:       30 * time.Second,
		Reconnect:         true,
		ReconnectDelay:    2 * time.Second,
		ReconnectMaxDelay: 30 * time.Second,
		ReconnectMax:      5,
	}
}

// BookMessage is a WebSocket order book event.
type BookMessage struct {
	EventType string       `json:"event_type"`
	AssetID   string       `json:"asset_id"`
	Market    string       `json:"market"`
	Timestamp string       `json:"timestamp"`
	Hash      string       `json:"hash"`
	Bids      []PriceLevel `json:"bids"`
	Asks      []PriceLevel `json:"asks"`
}

// PriceLevel is a single price level in a WebSocket book message.
type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// PriceChangeMessage is a WebSocket price change event.
type PriceChangeMessage struct {
	EventType string              `json:"event_type"`
	Market    string              `json:"market"`
	Changes   []PriceChangeEntry  `json:"price_changes"`
	Timestamp string              `json:"timestamp"`
}

// PriceChangeEntry is a single price change.
type PriceChangeEntry struct {
	AssetID string `json:"asset_id"`
	Price   string `json:"price"`
	Side    string `json:"side"`
	Size    string `json:"size"`
	Hash    string `json:"hash"`
}

// LastTradeMessage is a WebSocket last trade event.
type LastTradeMessage struct {
	EventType  string `json:"event_type"`
	AssetID    string `json:"asset_id"`
	Market     string `json:"market"`
	Price      string `json:"price"`
	Side       string `json:"side"`
	Size       string `json:"size"`
	FeeRateBps string `json:"fee_rate_bps"`
	Timestamp  string `json:"timestamp"`
}

// MarketClient manages a public market WebSocket connection.
type MarketClient struct {
	config     Config
	conn       *websocket.Conn
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	connected  atomic.Bool
	reconnects int32

	// Callbacks
	OnBook         func(BookMessage)
	OnPriceChange  func(PriceChangeMessage)
	OnLastTrade    func(LastTradeMessage)
	OnError        func(error)
}

// NewMarketClient creates a public market WebSocket client.
func NewMarketClient(config Config) *MarketClient {
	return &MarketClient{config: config}
}

// Connect establishes the WebSocket connection.
func (mc *MarketClient) Connect(ctx context.Context) error {
	mc.ctx, mc.cancel = context.WithCancel(ctx)
	return mc.dial()
}

func (mc *MarketClient) dial() error {
	conn, _, err := websocket.DefaultDialer.Dial(mc.config.URL, nil)
	if err != nil {
		return fmt.Errorf("ws dial: %w", err)
	}
	mc.conn = conn
	mc.connected.Store(true)
	mc.conn.SetPongHandler(func(string) error {
		mc.conn.SetReadDeadline(time.Now().Add(mc.config.PongTimeout))
		return nil
	})
	go mc.readLoop()
	go mc.pingLoop()
	return nil
}

func (mc *MarketClient) readLoop() {
	defer mc.conn.Close()
	for {
		select {
		case <-mc.ctx.Done():
			return
		default:
		}
		_, msg, err := mc.conn.ReadMessage()
		if err != nil {
			if mc.OnError != nil {
				mc.OnError(fmt.Errorf("ws read: %w", err))
			}
			mc.connected.Store(false)
			mc.reconnect()
			return
		}
		mc.dispatch(msg)
	}
}

func (mc *MarketClient) dispatch(msg []byte) {
	// Try parsing as each event type
	if mc.OnBook != nil {
		var book BookMessage
		if json.Unmarshal(msg, &book) == nil && book.EventType == "book" {
			mc.OnBook(book)
			return
		}
	}
	if mc.OnPriceChange != nil {
		var pc PriceChangeMessage
		if json.Unmarshal(msg, &pc) == nil && pc.EventType == "price_change" {
			mc.OnPriceChange(pc)
			return
		}
	}
	if mc.OnLastTrade != nil {
		var lt LastTradeMessage
		if json.Unmarshal(msg, &lt) == nil && lt.EventType == "last_trade_price" {
			mc.OnLastTrade(lt)
			return
		}
	}
}

func (mc *MarketClient) pingLoop() {
	ticker := time.NewTicker(mc.config.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.mu.Lock()
			if mc.conn != nil {
				mc.conn.WriteMessage(websocket.PingMessage, nil)
			}
			mc.mu.Unlock()
		}
	}
}

func (mc *MarketClient) reconnect() {
	if !mc.config.Reconnect || atomic.LoadInt32(&mc.reconnects) >= int32(mc.config.ReconnectMax) {
		return
	}
	atomic.AddInt32(&mc.reconnects, 1)
	delay := mc.config.ReconnectDelay
	for i := int32(0); i < atomic.LoadInt32(&mc.reconnects); i++ {
		delay *= 2
		if delay > mc.config.ReconnectMaxDelay {
			delay = mc.config.ReconnectMaxDelay
		}
	}
	time.Sleep(delay)
	if mc.ctx.Err() != nil {
		return
	}
	mc.dial()
}

// SubscribeAssets subscribes to order book updates for given token IDs.
func (mc *MarketClient) SubscribeAssets(ctx context.Context, assetIDs []string) error {
	msg := map[string]interface{}{
		"type":       "market",
		"assets_ids": assetIDs,
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.conn.WriteJSON(msg)
}

// Close shuts down the WebSocket connection.
func (mc *MarketClient) Close() {
	mc.cancel()
	mc.mu.Lock()
	if mc.conn != nil {
		mc.conn.Close()
	}
	mc.mu.Unlock()
}

// IsConnected returns the current connection state.
func (mc *MarketClient) IsConnected() bool {
	return mc.connected.Load()
}
