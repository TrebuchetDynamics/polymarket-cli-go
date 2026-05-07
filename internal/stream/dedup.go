package stream

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Deduplicator removes duplicate messages from redundant WebSocket connections.
// Stolen from polymarket-kit/go-client/client/aggregator.go.
type Deduplicator struct {
	mu    sync.Mutex
	seen  map[string]int64
	size  int
	ttlMs int64

	In   atomic.Int64
	Dup  atomic.Int64
	Out  atomic.Int64
}

// NewDeduplicator creates a deduplicator with the given capacity and TTL.
func NewDeduplicator(size int, ttl time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:  make(map[string]int64, size),
		size:  size,
		ttlMs: ttl.Milliseconds(),
	}
}

// Process checks if a message is a duplicate. Returns true if new.
func (d *Deduplicator) Process(data []byte) bool {
	key := extractKey(data)
	if key == "" {
		d.In.Add(1)
		d.Out.Add(1)
		return true
	}

	nowMs := time.Now().UnixMilli()

	d.mu.Lock()
	defer d.mu.Unlock()

	d.In.Add(1)

	if ts, exists := d.seen[key]; exists && (nowMs-ts) < d.ttlMs {
		d.Dup.Add(1)
		return false
	}

	d.seen[key] = nowMs

	if len(d.seen) > d.size {
		d.evictLocked(nowMs)
	}

	d.Out.Add(1)
	return true
}

func (d *Deduplicator) evictLocked(nowMs int64) {
	for k, ts := range d.seen {
		if (nowMs - ts) >= d.ttlMs {
			delete(d.seen, k)
		}
	}
}

func (d *Deduplicator) Reset() {
	d.mu.Lock()
	d.seen = make(map[string]int64, d.size)
	d.mu.Unlock()
}

// Stats returns deduplication counters.
func (d *Deduplicator) Stats() (in, dup, out int64) {
	return d.In.Load(), d.Dup.Load(), d.Out.Load()
}

func extractKey(data []byte) string {
	var probe struct {
		EventType string `json:"event_type"`
		Hash      string `json:"hash"`
		AssetID   string `json:"asset_id"`
		Price     string `json:"price"`
		Size      string `json:"size"`
		Market    string `json:"market"`
		Timestamp string `json:"timestamp"`
	}

	if json.Unmarshal(data, &probe) != nil {
		return ""
	}

	switch probe.EventType {
	case "book", "tick_size_change":
		if probe.Hash != "" {
			return probe.EventType + ":" + probe.Hash
		}
	case "price_change":
		if probe.Hash != "" {
			return "pc:" + probe.Hash
		}
		if probe.Market != "" {
			return fmt.Sprintf("pc:%s:%s", probe.Market, probe.Timestamp)
		}
	case "last_trade_price":
		if probe.AssetID != "" && probe.Price != "" {
			return fmt.Sprintf("ltp:%s:%s:%s", probe.AssetID, probe.Price, probe.Size)
		}
	}

	return ""
}

// SplitArray splits a JSON array of messages into individual messages.
func SplitArray(data []byte) []json.RawMessage {
	if len(data) == 0 || data[0] != '[' {
		return nil
	}
	var msgs []json.RawMessage
	if json.Unmarshal(data, &msgs) != nil {
		return nil
	}
	return msgs
}
