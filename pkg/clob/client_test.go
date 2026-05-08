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
		if r.URL.Path != "/clob-markets/condition-1" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"gst":"2026-01-01T00:00:00Z",
			"t":[{"t":"token-yes","o":"Yes"}],
			"mos":5,
			"mts":0.01,
			"mbf":0,
			"tbf":0,
			"rfqe":true,
			"itode":true,
			"ibce":true,
			"fd":{"r":0.02,"e":2,"to":true},
			"oas":123
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
	if got := publicMarket.Tokens[0]; got.TokenID != "token-yes" || got.Outcome != "Yes" {
		t.Fatalf("unexpected token conversion: %+v", got)
	}
	if publicMarket.OrderMinSize != 5 || publicMarket.OrderPriceMinTickSize != 0.01 {
		t.Fatalf("unexpected market order constraints: %+v", publicMarket)
	}
	if !publicMarket.RFQEnabled || !publicMarket.TakerOrderDelay || !publicMarket.BlockaidCheckEnabled {
		t.Fatalf("missing current CLOB market flags: %+v", publicMarket)
	}
	if publicMarket.FeeDetails.Rate != 0.02 || publicMarket.FeeDetails.Exponent != 2 || !publicMarket.FeeDetails.TakerOnly {
		t.Fatalf("unexpected fee details: %+v", publicMarket.FeeDetails)
	}
}

func TestClientScalarMarketDataParsesCurrentNumericDTOs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/price":
			_, _ = w.Write([]byte(`{"price":0.45}`))
		case "/midpoint":
			_, _ = w.Write([]byte(`{"mid_price":0.5}`))
		case "/fee-rate":
			_, _ = w.Write([]byte(`{"base_fee":30}`))
		case "/prices-history":
			_, _ = w.Write([]byte(`{"history":[{"t":123,"p":0.45}]}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	price, err := client.Price(context.Background(), "token-1", "BUY")
	if err != nil {
		t.Fatalf("Price returned error: %v", err)
	}
	if price != "0.45" {
		t.Fatalf("Price = %q, want 0.45", price)
	}
	midpoint, err := client.Midpoint(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("Midpoint returned error: %v", err)
	}
	if midpoint != "0.5" {
		t.Fatalf("Midpoint = %q, want 0.5", midpoint)
	}
	fee, err := client.FeeRateBps(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("FeeRateBps returned error: %v", err)
	}
	if fee != 30 {
		t.Fatalf("FeeRateBps = %d, want 30", fee)
	}
	history, err := client.PricesHistory(context.Background(), &types.CLOBPriceHistoryParams{Market: "token-1"})
	if err != nil {
		t.Fatalf("PricesHistory returned error: %v", err)
	}
	if got := history.History[0]; got.T != "123" || got.P != "0.45" {
		t.Fatalf("unexpected history point: %+v", got)
	}
}

func TestClientBatchMarketDataParsesCurrentDTOs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/prices":
			_, _ = w.Write([]byte(`{"token-1":{"BUY":0.45},"token-2":{"SELL":0.52}}`))
		case "/midpoints":
			_, _ = w.Write([]byte(`{"token-1":0.5,"token-2":"0.51"}`))
		case "/last-trades-prices":
			_, _ = w.Write([]byte(`[{"token_id":"token-1","price":"0.44","side":"BUY"},{"token_id":"token-2","price":"0.53","side":"SELL"}]`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	params := []types.CLOBBookParams{
		{TokenID: "token-1", Side: "BUY"},
		{TokenID: "token-2", Side: "SELL"},
	}
	prices, err := client.Prices(context.Background(), params)
	if err != nil {
		t.Fatalf("Prices returned error: %v", err)
	}
	if prices["token-1"] != "0.45" || prices["token-2"] != "0.52" {
		t.Fatalf("unexpected prices: %+v", prices)
	}
	midpoints, err := client.Midpoints(context.Background(), params)
	if err != nil {
		t.Fatalf("Midpoints returned error: %v", err)
	}
	if midpoints["token-1"] != "0.5" || midpoints["token-2"] != "0.51" {
		t.Fatalf("unexpected midpoints: %+v", midpoints)
	}
	lastTrades, err := client.LastTradesPrices(context.Background(), params)
	if err != nil {
		t.Fatalf("LastTradesPrices returned error: %v", err)
	}
	if lastTrades["token-1"] != "0.44" || lastTrades["token-2"] != "0.53" {
		t.Fatalf("unexpected last trades: %+v", lastTrades)
	}
}
