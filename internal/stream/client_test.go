package stream

import (
	"encoding/json"
	"testing"
)

func TestBookMessage_Unmarshal(t *testing.T) {
	raw := `{"event_type":"book","asset_id":"123","market":"0xabc","timestamp":"1234","hash":"h1","bids":[{"price":"0.5","size":"100"}],"asks":[{"price":"0.51","size":"200"}]}`
	var msg BookMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventType != "book" {
		t.Fatalf("event_type=%s", msg.EventType)
	}
	if len(msg.Bids) != 1 || msg.Bids[0].Price != "0.5" {
		t.Fatalf("bids=%+v", msg.Bids)
	}
}

func TestPriceChangeMessage_Unmarshal(t *testing.T) {
	raw := `{"event_type":"price_change","market":"0xabc","changes":[{"asset_id":"123","price":"0.55","side":"BUY","size":"50","hash":"h2"}],"timestamp":"5678"}`
	var msg PriceChangeMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventType != "price_change" {
		t.Fatalf("event_type=%s", msg.EventType)
	}
}

func TestLastTradeMessage_Unmarshal(t *testing.T) {
	raw := `{"event_type":"last_trade_price","asset_id":"123","market":"0xabc","price":"0.52","side":"BUY","size":"30","fee_rate_bps":"0","timestamp":"9012"}`
	var msg LastTradeMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Price != "0.52" {
		t.Fatalf("price=%s", msg.Price)
	}
}

func TestDefaultConfig_UsesURL(t *testing.T) {
	cfg := DefaultConfig("wss://example.com/ws")
	if cfg.URL != "wss://example.com/ws" {
		t.Fatalf("URL=%s", cfg.URL)
	}
	if cfg.PingInterval == 0 {
		t.Fatal("PingInterval is zero")
	}
}

func TestNewMarketClient_ReturnsClient(t *testing.T) {
	cfg := DefaultConfig("wss://example.com/ws")
	client := NewMarketClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
