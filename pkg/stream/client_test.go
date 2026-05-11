package stream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestMarketClientSubscribeAndDispatchesPublicDTOs(t *testing.T) {
	upgrader := websocket.Upgrader{}
	receivedSubscribe := make(chan map[string]interface{}, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		defer conn.Close()

		var sub map[string]interface{}
		if err := conn.ReadJSON(&sub); err != nil {
			t.Errorf("read subscribe: %v", err)
			return
		}
		receivedSubscribe <- sub

		_ = conn.WriteJSON(map[string]interface{}{
			"event_type": "price_change",
			"market":     "condition-1",
			"timestamp":  "1757908892351",
			"price_changes": []map[string]string{{
				"asset_id": "token-1",
				"price":    "0.5",
				"size":     "200",
				"side":     "BUY",
				"hash":     "hash-1",
				"best_bid": "0.5",
				"best_ask": "1",
			}},
		})
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewMarketClient(Config{
		URL:          wsURL,
		PingInterval: time.Hour,
		PongTimeout:  time.Second,
		Reconnect:    false,
	})

	gotPriceChange := make(chan PriceChangeMessage, 1)
	client.OnPriceChange = func(msg PriceChangeMessage) {
		gotPriceChange <- msg
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	defer client.Close()
	if err := client.SubscribeAssets(ctx, []string{"token-1"}); err != nil {
		t.Fatalf("SubscribeAssets returned error: %v", err)
	}

	select {
	case sub := <-receivedSubscribe:
		if sub["type"] != "market" {
			t.Fatalf("type = %v, want market", sub["type"])
		}
		assets, ok := sub["assets_ids"].([]interface{})
		if !ok || len(assets) != 1 || assets[0] != "token-1" {
			t.Fatalf("assets_ids = %#v", sub["assets_ids"])
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for subscribe payload")
	}

	select {
	case msg := <-gotPriceChange:
		if msg.Market != "condition-1" || len(msg.PriceChanges) != 1 {
			t.Fatalf("unexpected price change: %+v", msg)
		}
		if change := msg.PriceChanges[0]; change.BestBid != "0.5" || change.BestAsk != "1" {
			t.Fatalf("missing best bid/ask: %+v", change)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for price change")
	}
}

func TestMarketClientSubscribeWithCustomFeaturesAndDispatchesV2MarketEvents(t *testing.T) {
	upgrader := websocket.Upgrader{}
	receivedSubscribe := make(chan map[string]interface{}, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		defer conn.Close()

		var sub map[string]interface{}
		if err := conn.ReadJSON(&sub); err != nil {
			t.Errorf("read subscribe: %v", err)
			return
		}
		receivedSubscribe <- sub

		_ = conn.WriteJSON(map[string]interface{}{
			"event_type":    "tick_size_change",
			"asset_id":      "token-1",
			"market":        "condition-1",
			"old_tick_size": "0.01",
			"new_tick_size": "0.001",
			"timestamp":     "1757908892351",
		})
		_ = conn.WriteJSON(map[string]interface{}{
			"event_type": "best_bid_ask",
			"asset_id":   "token-1",
			"market":     "condition-1",
			"best_bid":   "0.73",
			"best_ask":   "0.77",
			"spread":     "0.04",
			"timestamp":  "1766789469958",
		})
		_ = conn.WriteJSON(map[string]interface{}{
			"event_type": "market_resolved",
			"id":         "1031769",
			"market":     "condition-1",
			"assets_ids": []string{"token-yes", "token-no"},
			"timestamp":  "1766790415550",
			"tags":       []string{"stocks"},
		})
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewMarketClient(Config{
		URL:                  wsURL,
		PingInterval:         time.Hour,
		PongTimeout:          time.Second,
		Reconnect:            false,
		CustomFeatureEnabled: true,
		Level:                2,
	})

	gotTick := make(chan TickSizeChangeMessage, 1)
	gotBest := make(chan BestBidAskMessage, 1)
	gotResolved := make(chan MarketResolvedMessage, 1)
	client.OnTickSizeChange = func(msg TickSizeChangeMessage) { gotTick <- msg }
	client.OnBestBidAsk = func(msg BestBidAskMessage) { gotBest <- msg }
	client.OnMarketResolved = func(msg MarketResolvedMessage) { gotResolved <- msg }

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	defer client.Close()
	if err := client.SubscribeAssets(ctx, []string{"token-1"}); err != nil {
		t.Fatalf("SubscribeAssets returned error: %v", err)
	}

	select {
	case sub := <-receivedSubscribe:
		if sub["type"] != "market" {
			t.Fatalf("type = %v, want market", sub["type"])
		}
		if sub["custom_feature_enabled"] != true {
			t.Fatalf("custom_feature_enabled = %v, want true", sub["custom_feature_enabled"])
		}
		if sub["level"] != float64(2) {
			t.Fatalf("level = %v, want 2", sub["level"])
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for subscribe payload")
	}

	select {
	case msg := <-gotTick:
		if msg.NewTickSize != "0.001" || msg.OldTickSize != "0.01" {
			t.Fatalf("unexpected tick-size event: %+v", msg)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for tick-size event")
	}
	select {
	case msg := <-gotBest:
		if msg.BestBid != "0.73" || msg.BestAsk != "0.77" || msg.Spread != "0.04" {
			t.Fatalf("unexpected best-bid-ask event: %+v", msg)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for best-bid-ask event")
	}
	select {
	case msg := <-gotResolved:
		if msg.ID != "1031769" || len(msg.AssetIDs) != 2 || msg.Tags[0] != "stocks" {
			t.Fatalf("unexpected market-resolved event: %+v", msg)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for market-resolved event")
	}
}

func TestMarketStreamDTOsUnmarshalCurrentFields(t *testing.T) {
	var trade LastTradeMessage
	if err := json.Unmarshal([]byte(`{
		"event_type":"last_trade_price",
		"asset_id":"token-1",
		"market":"condition-1",
		"price":"0.5",
		"size":"10",
		"fee_rate_bps":"10",
		"side":"BUY",
		"timestamp":"1757908892351",
		"transaction_hash":"0xabc123"
	}`), &trade); err != nil {
		t.Fatalf("unmarshal last trade: %v", err)
	}
	if trade.TransactionHash != "0xabc123" {
		t.Fatalf("TransactionHash = %q, want 0xabc123", trade.TransactionHash)
	}

	var priceChange PriceChangeMessage
	if err := json.Unmarshal([]byte(`{
		"event_type":"price_change",
		"market":"condition-1",
		"price_changes":[{"asset_id":"token-1","best_bid":"0.5","best_ask":"1"}],
		"timestamp":"1757908892351"
	}`), &priceChange); err != nil {
		t.Fatalf("unmarshal price change: %v", err)
	}
	if priceChange.PriceChanges[0].BestBid != "0.5" || priceChange.PriceChanges[0].BestAsk != "1" {
		t.Fatalf("price change = %+v", priceChange.PriceChanges[0])
	}
}

func TestSubscribeAssetsBeforeConnectReturnsError(t *testing.T) {
	client := NewMarketClient(Config{})
	if err := client.SubscribeAssets(context.Background(), []string{"token-1"}); err == nil {
		t.Fatal("expected error before connect")
	}
}

func TestMarketClientReconnectResubscribesAssets(t *testing.T) {
	upgrader := websocket.Upgrader{}
	subscriptions := make(chan map[string]interface{}, 2)
	var connections atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		defer conn.Close()

		var sub map[string]interface{}
		if err := conn.ReadJSON(&sub); err != nil {
			t.Errorf("read subscribe: %v", err)
			return
		}
		subscriptions <- sub

		if connections.Add(1) == 1 {
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "reconnect"))
		} else {
			<-r.Context().Done()
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewMarketClient(Config{
		URL:               wsURL,
		PingInterval:      time.Hour,
		PongTimeout:       time.Second,
		Reconnect:         true,
		ReconnectDelay:    time.Millisecond,
		ReconnectMaxDelay: time.Millisecond,
		ReconnectMax:      1,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	defer client.Close()
	if err := client.SubscribeAssets(ctx, []string{"token-1"}); err != nil {
		t.Fatalf("SubscribeAssets returned error: %v", err)
	}

	for i := 0; i < 2; i++ {
		select {
		case sub := <-subscriptions:
			assets, ok := sub["assets_ids"].([]interface{})
			if !ok || len(assets) != 1 || assets[0] != "token-1" {
				t.Fatalf("assets_ids = %#v", sub["assets_ids"])
			}
		case <-ctx.Done():
			t.Fatalf("timed out waiting for subscription %d", i+1)
		}
	}
}
