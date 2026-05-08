package universal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
