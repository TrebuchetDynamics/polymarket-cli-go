package universal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
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
		json.NewEncoder(w).Encode([]polytypes.Market{
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
		json.NewEncoder(w).Encode(polytypes.SearchResponse{
			Events: []polytypes.Event{{ID: "e1", Title: "Bitcoin"}},
		})
	}))
	defer srv.Close()

	c := NewClient(Config{GammaBaseURL: srv.URL})
	resp, err := c.Search(context.Background(), &polytypes.SearchParams{Q: "btc"})
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
	resp, err := c.BalanceAllowance(context.Background(), testPrivateKey, clob.BalanceAllowanceParams{
		AssetType:     "COLLATERAL",
		SignatureType: 0,
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
	resp, err := c.UpdateBalanceAllowance(context.Background(), testPrivateKey, clob.BalanceAllowanceParams{
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
	resp, err := c.CreateLimitOrder(context.Background(), testPrivateKey, clob.CreateOrderParams{
		TokenID:       "1234567890",
		Side:          "BUY",
		Price:         "0.50",
		Size:          "10",
		OrderType:     "GTC",
		SignatureType: 0,
	})
	if err != nil {
		t.Fatalf("CreateLimitOrder error: %v", err)
	}
	if resp.OrderID != "ord-1" {
		t.Errorf("expected ord-1, got %q", resp.OrderID)
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
	resp, err := c.CreateMarketOrder(context.Background(), testPrivateKey, clob.MarketOrderParams{
		TokenID:       "1234567890",
		Side:          "BUY",
		Amount:        "5",
		OrderType:     "FOK",
		SignatureType: 0,
	})
	if err != nil {
		t.Fatalf("CreateMarketOrder error: %v", err)
	}
	if resp.OrderID != "ord-mkt-1" {
		t.Errorf("expected ord-mkt-1, got %q", resp.OrderID)
	}
}
