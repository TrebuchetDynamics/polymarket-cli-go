package universal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

func TestNewClientUsesDefaults(t *testing.T) {
	c := NewClient(Config{})
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.gamma == nil {
		t.Error("gamma client is nil")
	}
	if c.clob == nil {
		t.Error("clob client is nil")
	}
	if c.data == nil {
		t.Error("data client is nil")
	}
}

func TestActiveMarkets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]types.Market{
			{ID: "m1", Question: "Will it rain?"},
		})
	}))
	defer srv.Close()

	c := NewClient(Config{GammaBaseURL: srv.URL})
	markets, err := c.ActiveMarkets(context.Background())
	if err != nil {
		t.Fatalf("ActiveMarkets error: %v", err)
	}
	if len(markets) != 1 {
		t.Fatalf("expected 1 market, got %d", len(markets))
	}
}

func TestSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/public-search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(types.SearchResponse{
			Events: []types.Event{{ID: "e1", Title: "Bitcoin"}},
		})
	}))
	defer srv.Close()

	c := NewClient(Config{GammaBaseURL: srv.URL})
	resp, err := c.Search(context.Background(), &types.SearchParams{Q: "btc"})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Events))
	}
}

func TestOrderBook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/book" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(polytypes.OrderBook{
			AssetID: "token-1",
			Bids:    []polytypes.OrderBookLevel{{Price: "0.50", Size: "100"}},
			Asks:    []polytypes.OrderBookLevel{{Price: "0.51", Size: "200"}},
		})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	book, err := c.OrderBook(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("OrderBook error: %v", err)
	}
	if book.AssetID != "token-1" {
		t.Errorf("expected asset token-1, got %s", book.AssetID)
	}
}

func TestLiveVolume(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/live-volume" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total":  1,
			"events": []map[string]interface{}{{"event_id": "e1", "volume": 1000.0}},
		})
	}))
	defer srv.Close()

	c := NewClient(Config{DataBaseURL: srv.URL})
	vol, err := c.LiveVolume(context.Background(), 10)
	if err != nil {
		t.Fatalf("LiveVolume error: %v", err)
	}
	if vol.Total != 1 {
		t.Errorf("expected total 1, got %d", vol.Total)
	}
}

func TestCurrentPositionsWithLimitRoutes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "0xuser" || r.URL.Query().Get("limit") != "5" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode([]map[string]string{{"token_id": "token-1"}})
	}))
	defer srv.Close()

	c := NewClient(Config{DataBaseURL: srv.URL})
	rows, err := c.CurrentPositionsWithLimit(context.Background(), "0xuser", 5)
	if err != nil {
		t.Fatalf("CurrentPositionsWithLimit error: %v", err)
	}
	if len(rows) != 1 || rows[0].TokenID != "token-1" {
		t.Errorf("unexpected rows: %+v", rows)
	}
}

func TestHealthCheckPartial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": "ok"})
	}))
	defer srv.Close()

	c := NewClient(Config{GammaBaseURL: srv.URL})
	resp, err := c.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("HealthCheck should not error on partial failure: %v", err)
	}
	if !resp.GammaOK {
		t.Error("expected GammaOK to be true")
	}
}

func TestHealthCheckAllFailure(t *testing.T) {
	c := NewClient(Config{
		GammaBaseURL: "http://localhost:1",
		CLOBBaseURL:  "http://localhost:1",
		DataBaseURL:  "http://localhost:1",
	})
	_, err := c.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error when all APIs are unreachable")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.GammaBaseURL != defaultGammaBaseURL {
		t.Errorf("gamma: %s", cfg.GammaBaseURL)
	}
	if cfg.CLOBBaseURL != defaultCLOBBaseURL {
		t.Errorf("clob: %s", cfg.CLOBBaseURL)
	}
	if cfg.DataBaseURL != defaultDataBaseURL {
		t.Errorf("data: %s", cfg.DataBaseURL)
	}
}

func TestStreamClient(t *testing.T) {
	c := NewClient(Config{})
	sc := c.StreamClient()
	if sc == nil {
		t.Fatal("StreamClient returned nil")
	}
}

// Deterministic test EOA — same key the internal/clob tests use.
const testPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
const testBuilderCode = "0x1111111111111111111111111111111111111111111111111111111111111111"
const testDepositWallet = "0x19bE70b1e4F59C0663a999C0dC6f5b3C68CFCaF3"

func TestCreateOrDeriveAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/api-key" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"apiKey":     "k1-uuid-shape-1234",
			"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			"passphrase": "pp1",
		})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	key, err := c.CreateOrDeriveAPIKey(context.Background(), testPrivateKey)
	if err != nil {
		t.Fatalf("CreateOrDeriveAPIKey error: %v", err)
	}
	if key.Key != "k1-uuid-shape-1234" {
		t.Errorf("expected api key, got %q", key.Key)
	}
	if key.Passphrase != "pp1" {
		t.Errorf("expected passphrase pp1, got %q", key.Passphrase)
	}
}

func TestCreateAPIKeyForAddress(t *testing.T) {
	var sawAddress string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/api-key" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		sawAddress = r.Header.Get("POLY_ADDRESS")
		json.NewEncoder(w).Encode(map[string]string{
			"apiKey":     "owner-key",
			"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			"passphrase": "owner-pass",
		})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	key, err := c.CreateAPIKeyForAddress(context.Background(), testPrivateKey, testDepositWallet)
	if err != nil {
		t.Fatalf("CreateAPIKeyForAddress error: %v", err)
	}
	if sawAddress != testDepositWallet {
		t.Fatalf("POLY_ADDRESS = %s, want %s", sawAddress, testDepositWallet)
	}
	if key.Key != "owner-key" {
		t.Errorf("expected owner key, got %q", key.Key)
	}
}

func TestDeriveAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/derive-api-key" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"apiKey":     "k2-derived",
			"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			"passphrase": "pp2",
		})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	key, err := c.DeriveAPIKey(context.Background(), testPrivateKey)
	if err != nil {
		t.Fatalf("DeriveAPIKey error: %v", err)
	}
	if key.Key != "k2-derived" {
		t.Errorf("expected derived key, got %q", key.Key)
	}
}

func TestOrderRoutes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k-order",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp-order",
			})
		case "/order/0xabc":
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "0xabc", "status": "LIVE"})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	order, err := c.Order(context.Background(), testPrivateKey, "0xabc")
	if err != nil {
		t.Fatalf("Order error: %v", err)
	}
	if order.ID != "0xabc" || order.Status != "LIVE" {
		t.Errorf("unexpected order: %+v", order)
	}
}

func TestCancelOrdersRoutes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k-cancel-orders",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp-cancel-orders",
			})
		case "/orders":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"canceled": []string{"0x1", "0x2"}})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	resp, err := c.CancelOrders(context.Background(), testPrivateKey, []string{"0x1", "0x2"})
	if err != nil {
		t.Fatalf("CancelOrders error: %v", err)
	}
	if len(resp.Canceled) != 2 {
		t.Errorf("expected 2 canceled orders, got %+v", resp)
	}
}

func TestCancelMarketRoutes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k-cancel-market",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp-cancel-market",
			})
		case "/cancel-market-orders":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"canceled": []string{"0x1"}})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	resp, err := c.CancelMarket(context.Background(), testPrivateKey, sdkclob.CancelMarketParams{Market: "0xmarket"})
	if err != nil {
		t.Fatalf("CancelMarket error: %v", err)
	}
	if len(resp.Canceled) != 1 {
		t.Errorf("expected 1 canceled order, got %+v", resp)
	}
}

func TestBalanceAllowance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k3",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp3",
			})
		case "/balance-allowance":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"balance":   "1000000",
				"allowance": "999",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	resp, err := c.BalanceAllowance(context.Background(), testPrivateKey, sdkclob.BalanceAllowanceParams{
		AssetType: "COLLATERAL",
	})
	if err != nil {
		t.Fatalf("BalanceAllowance error: %v", err)
	}
	if resp.Balance != "1000000" {
		t.Errorf("expected balance 1000000, got %q", resp.Balance)
	}
}

func TestUpdateBalanceAllowance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k4",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp4",
			})
		case "/balance-allowance/update":
			json.NewEncoder(w).Encode(map[string]string{"balance": "2000000"})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	resp, err := c.UpdateBalanceAllowance(context.Background(), testPrivateKey, sdkclob.BalanceAllowanceParams{
		AssetType: "COLLATERAL",
	})
	if err != nil {
		t.Fatalf("UpdateBalanceAllowance error: %v", err)
	}
	if resp.Balance != "2000000" {
		t.Errorf("expected updated balance 2000000, got %q", resp.Balance)
	}
}

// Limit/market order placement: the internal clob client looks up tick
// size, signs, and POSTs /order. We mock the necessary dependencies and
// only assert that the wrapper routes through.
func TestCreateLimitOrderRoutes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k5",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp5",
			})
		case "/tick-size":
			json.NewEncoder(w).Encode(map[string]string{"minimum_tick_size": "0.01"})
		case "/neg-risk":
			json.NewEncoder(w).Encode(map[string]bool{"neg_risk": false})
		case "/order":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true, "orderID": "ord-1", "status": "matched",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	resp, err := c.CreateLimitOrder(context.Background(), testPrivateKey, sdkclob.CreateOrderParams{
		TokenID:   "1234567890",
		Side:      "BUY",
		Price:     "0.50",
		Size:      "10",
		OrderType: "GTC",
	})
	if err != nil {
		t.Fatalf("CreateLimitOrder error: %v", err)
	}
	if resp.OrderID != "ord-1" {
		t.Errorf("expected ord-1, got %q", resp.OrderID)
	}
}

func TestCreateLimitOrderUsesConfiguredBuilderCode(t *testing.T) {
	var posted map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k-builder",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp-builder",
			})
		case "/tick-size":
			json.NewEncoder(w).Encode(map[string]string{"minimum_tick_size": "0.001"})
		case "/neg-risk":
			json.NewEncoder(w).Encode(map[string]bool{"neg_risk": false})
		case "/order":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode order body: %v", err)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true, "orderID": "ord-builder", "status": "matched",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL, BuilderCode: testBuilderCode})
	_, err := c.CreateLimitOrder(context.Background(), testPrivateKey, sdkclob.CreateOrderParams{
		TokenID:   "12345",
		Side:      "BUY",
		Price:     "0.500000",
		Size:      "1.400000",
		OrderType: "GTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	order, ok := posted["order"].(map[string]any)
	if !ok {
		t.Fatalf("posted order missing: %#v", posted)
	}
	if order["builder"] != testBuilderCode {
		t.Fatalf("posted builder=%#v want %s", order["builder"], testBuilderCode)
	}
}

func TestCreateMarketOrderRoutes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			json.NewEncoder(w).Encode(map[string]string{
				"apiKey":     "k6",
				"secret":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"passphrase": "pp6",
			})
		case "/tick-size":
			json.NewEncoder(w).Encode(map[string]string{"minimum_tick_size": "0.01"})
		case "/neg-risk":
			json.NewEncoder(w).Encode(map[string]bool{"neg_risk": false})
		case "/midpoint":
			json.NewEncoder(w).Encode(map[string]string{"mid": "0.50"})
		case "/book":
			json.NewEncoder(w).Encode(polytypes.OrderBook{
				AssetID: "1234567890",
				Bids:    []polytypes.OrderBookLevel{{Price: "0.49", Size: "100"}},
				Asks:    []polytypes.OrderBookLevel{{Price: "0.50", Size: "100"}},
			})
		case "/order":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true, "orderID": "ord-mkt-1", "status": "matched",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	resp, err := c.CreateMarketOrder(context.Background(), testPrivateKey, sdkclob.MarketOrderParams{
		TokenID:   "1234567890",
		Side:      "BUY",
		Amount:    "5",
		OrderType: "FOK",
	})
	if err != nil {
		t.Fatalf("CreateMarketOrder error: %v", err)
	}
	if resp.OrderID != "ord-mkt-1" {
		t.Errorf("expected ord-mkt-1, got %q", resp.OrderID)
	}
}

// --- Metadata, scoring, and rewards passthrough tests ---

func TestCLOBServerTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/time" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"server_time": "2026-05-07T22:00:00Z"})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.CLOBServerTime(context.Background()); err != nil {
		t.Fatalf("CLOBServerTime error: %v", err)
	}
}

func TestOrderScoring(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/scoring" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("order_id"); got != "ord-1" {
			t.Errorf("expected order_id=ord-1, got %q", got)
		}
		json.NewEncoder(w).Encode(map[string]bool{"scoring": true})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	scoring, err := c.OrderScoring(context.Background(), "ord-1")
	if err != nil {
		t.Fatalf("OrderScoring error: %v", err)
	}
	if !scoring {
		t.Error("expected scoring=true")
	}
}

func TestOrdersScoring(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/scoring" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode([]bool{true, false, true})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	got, err := c.OrdersScoring(context.Background(), []string{"o1", "o2", "o3"})
	if err != nil {
		t.Fatalf("OrdersScoring error: %v", err)
	}
	if len(got) != 3 || got[0] != true || got[1] != false || got[2] != true {
		t.Errorf("unexpected scoring: %v", got)
	}
}

func TestRewardsConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rewards/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 1}})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.RewardsConfig(context.Background()); err != nil {
		t.Fatalf("RewardsConfig error: %v", err)
	}
}

func TestRawRewards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rewards/raw" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("market"); got != "0xmarket" {
			t.Errorf("expected market query, got %q", got)
		}
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.RawRewards(context.Background(), "0xmarket"); err != nil {
		t.Fatalf("RawRewards error: %v", err)
	}
}

func TestUserEarnings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rewards/earnings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("date"); got != "2026-05-07" {
			t.Errorf("expected date query, got %q", got)
		}
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.UserEarnings(context.Background(), "2026-05-07"); err != nil {
		t.Fatalf("UserEarnings error: %v", err)
	}
}

func TestTotalEarnings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rewards/total-earnings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"total": "100"})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.TotalEarnings(context.Background(), "2026-05-07"); err != nil {
		t.Fatalf("TotalEarnings error: %v", err)
	}
}

func TestRewardPercentages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rewards/percentages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.RewardPercentages(context.Background()); err != nil {
		t.Fatalf("RewardPercentages error: %v", err)
	}
}

func TestUserRewardsByMarket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rewards/markets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.UserRewardsByMarket(context.Background(), nil); err != nil {
		t.Fatalf("UserRewardsByMarket error: %v", err)
	}
}

func TestRebatedFees(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rebates" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer srv.Close()

	c := NewClient(Config{CLOBBaseURL: srv.URL})
	if _, err := c.RebatedFees(context.Background()); err != nil {
		t.Fatalf("RebatedFees error: %v", err)
	}
}
