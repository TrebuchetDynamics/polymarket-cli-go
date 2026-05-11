// Package marketdata turns raw Polymarket market-stream events into
// per-token orderbook/share-price snapshots.
package marketdata

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/TrebuchetDynamics/polygolem/pkg/stream"
)

// Level is one price level in a tracked orderbook snapshot.
type Level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// Snapshot is the latest normalized market-data view for one CLOB token.
type Snapshot struct {
	EventType        string  `json:"event_type"`
	AssetID          string  `json:"asset_id"`
	Market           string  `json:"market,omitempty"`
	Timestamp        string  `json:"timestamp,omitempty"`
	BestBid          string  `json:"best_bid,omitempty"`
	BestAsk          string  `json:"best_ask,omitempty"`
	Spread           string  `json:"spread,omitempty"`
	Midpoint         string  `json:"midpoint,omitempty"`
	TickSize         string  `json:"tick_size,omitempty"`
	PreviousTickSize string  `json:"previous_tick_size,omitempty"`
	LastTradePrice   string  `json:"last_trade_price,omitempty"`
	LastTradeSize    string  `json:"last_trade_size,omitempty"`
	LastTradeSide    string  `json:"last_trade_side,omitempty"`
	TransactionHash  string  `json:"transaction_hash,omitempty"`
	UpdateHash       string  `json:"update_hash,omitempty"`
	Bids             []Level `json:"bids,omitempty"`
	Asks             []Level `json:"asks,omitempty"`
}

// Tracker keeps an in-memory latest snapshot per asset ID.
type Tracker struct {
	mu        sync.RWMutex
	snapshots map[string]Snapshot
}

// NewTracker creates an empty market-data tracker.
func NewTracker() *Tracker {
	return &Tracker{snapshots: make(map[string]Snapshot)}
}

// ApplyBook records a full book snapshot and returns the normalized latest view.
func (t *Tracker) ApplyBook(msg stream.BookMessage) Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	snapshot := t.snapshotFor(msg.AssetID)
	snapshot.EventType = firstNonEmpty(msg.EventType, "book")
	snapshot.AssetID = msg.AssetID
	snapshot.Market = firstNonEmpty(msg.Market, snapshot.Market)
	snapshot.Timestamp = msg.Timestamp
	snapshot.UpdateHash = msg.Hash
	snapshot.Bids = levelsFromStream(msg.Bids)
	snapshot.Asks = levelsFromStream(msg.Asks)
	sortLevels(snapshot.Bids, true)
	sortLevels(snapshot.Asks, false)
	refreshPrices(&snapshot)
	t.snapshots[msg.AssetID] = snapshot
	return snapshot
}

// ApplyPriceChange applies one price-change message and returns one updated
// snapshot per changed asset.
func (t *Tracker) ApplyPriceChange(msg stream.PriceChangeMessage) []Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := make([]Snapshot, 0, len(msg.PriceChanges))
	for _, change := range msg.PriceChanges {
		if strings.TrimSpace(change.AssetID) == "" {
			continue
		}
		snapshot := t.snapshotFor(change.AssetID)
		snapshot.EventType = firstNonEmpty(msg.EventType, "price_change")
		snapshot.AssetID = change.AssetID
		snapshot.Market = firstNonEmpty(msg.Market, snapshot.Market)
		snapshot.Timestamp = msg.Timestamp
		snapshot.UpdateHash = change.Hash
		applyLevelChange(&snapshot, change)
		useStreamBest := change.BestBid != "" || change.BestAsk != ""
		if change.BestBid != "" {
			snapshot.BestBid = change.BestBid
		}
		if change.BestAsk != "" {
			snapshot.BestAsk = change.BestAsk
		}
		if useStreamBest {
			refreshMidpoint(&snapshot)
		} else {
			refreshPrices(&snapshot)
		}
		t.snapshots[change.AssetID] = snapshot
		out = append(out, snapshot)
	}
	return out
}

// ApplyLastTrade records the latest trade for one asset and returns the
// normalized latest view.
func (t *Tracker) ApplyLastTrade(msg stream.LastTradeMessage) Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	snapshot := t.snapshotFor(msg.AssetID)
	snapshot.EventType = firstNonEmpty(msg.EventType, "last_trade_price")
	snapshot.AssetID = msg.AssetID
	snapshot.Market = firstNonEmpty(msg.Market, snapshot.Market)
	snapshot.Timestamp = msg.Timestamp
	snapshot.LastTradePrice = msg.Price
	snapshot.LastTradeSize = msg.Size
	snapshot.LastTradeSide = msg.Side
	snapshot.TransactionHash = msg.TransactionHash
	refreshPrices(&snapshot)
	t.snapshots[msg.AssetID] = snapshot
	return snapshot
}

// ApplyBestBidAsk records a top-of-book update and returns the normalized
// latest view.
func (t *Tracker) ApplyBestBidAsk(msg stream.BestBidAskMessage) Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	snapshot := t.snapshotFor(msg.AssetID)
	snapshot.EventType = firstNonEmpty(msg.EventType, "best_bid_ask")
	snapshot.AssetID = msg.AssetID
	snapshot.Market = firstNonEmpty(msg.Market, snapshot.Market)
	snapshot.Timestamp = msg.Timestamp
	snapshot.BestBid = firstNonEmpty(msg.BestBid, snapshot.BestBid)
	snapshot.BestAsk = firstNonEmpty(msg.BestAsk, snapshot.BestAsk)
	refreshMidpoint(&snapshot)
	snapshot.Spread = firstNonEmpty(msg.Spread, snapshot.Spread)
	t.snapshots[msg.AssetID] = snapshot
	return snapshot
}

// ApplyTickSizeChange records the latest tick-size metadata for one asset.
func (t *Tracker) ApplyTickSizeChange(msg stream.TickSizeChangeMessage) Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	snapshot := t.snapshotFor(msg.AssetID)
	snapshot.EventType = firstNonEmpty(msg.EventType, "tick_size_change")
	snapshot.AssetID = msg.AssetID
	snapshot.Market = firstNonEmpty(msg.Market, snapshot.Market)
	snapshot.Timestamp = msg.Timestamp
	snapshot.PreviousTickSize = msg.OldTickSize
	snapshot.TickSize = msg.NewTickSize
	t.snapshots[msg.AssetID] = snapshot
	return snapshot
}

// Snapshot returns a copy of the latest snapshot for assetID.
func (t *Tracker) Snapshot(assetID string) (Snapshot, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	snapshot, ok := t.snapshots[assetID]
	return snapshot, ok
}

func (t *Tracker) snapshotFor(assetID string) Snapshot {
	if snapshot, ok := t.snapshots[assetID]; ok {
		return snapshot
	}
	return Snapshot{AssetID: assetID}
}

func levelsFromStream(rows []stream.PriceLevel) []Level {
	out := make([]Level, len(rows))
	for i, row := range rows {
		out[i] = Level{Price: row.Price, Size: row.Size}
	}
	return out
}

func applyLevelChange(snapshot *Snapshot, change stream.PriceChangeEntry) {
	side := strings.ToUpper(strings.TrimSpace(change.Side))
	switch side {
	case "BUY", "BID", "BIDS":
		snapshot.Bids = upsertLevel(snapshot.Bids, change.Price, change.Size)
		sortLevels(snapshot.Bids, true)
	case "SELL", "ASK", "ASKS":
		snapshot.Asks = upsertLevel(snapshot.Asks, change.Price, change.Size)
		sortLevels(snapshot.Asks, false)
	}
}

func upsertLevel(levels []Level, price, size string) []Level {
	if strings.TrimSpace(price) == "" {
		return levels
	}
	if isZeroSize(size) {
		out := levels[:0]
		for _, level := range levels {
			if level.Price != price {
				out = append(out, level)
			}
		}
		return out
	}
	for i := range levels {
		if levels[i].Price == price {
			levels[i].Size = size
			return levels
		}
	}
	return append(levels, Level{Price: price, Size: size})
}

func isZeroSize(size string) bool {
	value, err := strconv.ParseFloat(strings.TrimSpace(size), 64)
	return err == nil && value == 0
}

func sortLevels(levels []Level, bid bool) {
	sort.SliceStable(levels, func(i, j int) bool {
		left, leftOK := parsePrice(levels[i].Price)
		right, rightOK := parsePrice(levels[j].Price)
		if !leftOK || !rightOK {
			return leftOK
		}
		if bid {
			return left > right
		}
		return left < right
	})
}

func refreshPrices(snapshot *Snapshot) {
	if len(snapshot.Bids) > 0 {
		snapshot.BestBid = snapshot.Bids[0].Price
	}
	if len(snapshot.Asks) > 0 {
		snapshot.BestAsk = snapshot.Asks[0].Price
	}
	refreshMidpoint(snapshot)
}

func refreshMidpoint(snapshot *Snapshot) {
	if midpoint, ok := midpoint(snapshot.BestBid, snapshot.BestAsk); ok {
		snapshot.Midpoint = midpoint
	}
	if spread, ok := spread(snapshot.BestBid, snapshot.BestAsk); ok {
		snapshot.Spread = spread
	}
}

func midpoint(bid, ask string) (string, bool) {
	bidValue, bidOK := parsePrice(bid)
	askValue, askOK := parsePrice(ask)
	if !bidOK || !askOK {
		return "", false
	}
	return strconv.FormatFloat((bidValue+askValue)/2, 'f', -1, 64), true
}

func spread(bid, ask string) (string, bool) {
	bidValue, bidOK := parsePrice(bid)
	askValue, askOK := parsePrice(ask)
	if !bidOK || !askOK {
		return "", false
	}
	return strconv.FormatFloat(askValue-bidValue, 'f', -1, 64), true
}

func parsePrice(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return parsed, err == nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
