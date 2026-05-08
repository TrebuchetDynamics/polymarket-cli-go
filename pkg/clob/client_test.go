package clob

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

func TestClientOrderBookReturnsPublicDTO(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/book" || r.URL.Query().Get("token_id") != "token-1" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"market":"condition-1",
			"asset_id":"token-1",
			"timestamp":"1710000000",
			"hash":"book-hash",
			"bids":[{"price":"0.44","size":"10"}],
			"asks":[{"price":"0.46","size":"11"}],
			"min_order_size":"5",
			"tick_size":"0.01",
			"neg_risk":true,
			"last_trade_price":"0.45"
		}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	book, err := client.OrderBook(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("OrderBook returned error: %v", err)
	}

	var publicBook *types.CLOBOrderBook = book
	if publicBook.Market != "condition-1" || publicBook.AssetID != "token-1" {
		t.Fatalf("unexpected order book identity: %+v", publicBook)
	}
	if publicBook.TickSize != "0.01" || publicBook.NegRisk != true || publicBook.LastTradePrice != "0.45" {
		t.Fatalf("missing CLOB book metadata: %+v", publicBook)
	}
	if got := publicBook.Bids[0]; got.Price != "0.44" || got.Size != "10" {
		t.Fatalf("unexpected bid level: %+v", got)
	}
}

func TestClientMarketReturnsPublicDTO(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/condition-1" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"condition_id":"condition-1",
			"question_id":"question-1",
			"tokens":[{"token_id":"token-yes","outcome":"Yes","price":0.51,"winner":false}],
			"enable_order_book":true,
			"accepting_orders":true,
			"order_price_min_tick_size":0.01,
			"order_min_size":5,
			"maker_base_fee":0,
			"taker_base_fee":0
		}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	market, err := client.Market(context.Background(), "condition-1")
	if err != nil {
		t.Fatalf("Market returned error: %v", err)
	}

	var publicMarket *types.CLOBMarket = market
	if publicMarket.ConditionID != "condition-1" || len(publicMarket.Tokens) != 1 {
		t.Fatalf("unexpected market: %+v", publicMarket)
	}
	if got := publicMarket.Tokens[0]; got.TokenID != "token-yes" || got.Price != "0.51" {
		t.Fatalf("unexpected token conversion: %+v", got)
	}
}
