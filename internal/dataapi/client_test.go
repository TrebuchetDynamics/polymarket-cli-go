package dataapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

func TestCurrentPositionsWithLimitUsesQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			t.Fatalf("path=%s want /positions", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "0xuser" || r.URL.Query().Get("limit") != "7" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]Position{{TokenID: "token-1"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	rows, err := client.CurrentPositionsWithLimit(context.Background(), "0xuser", 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].TokenID != "token-1" {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestClosedPositionsWithLimitUsesQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/closed-positions" {
			t.Fatalf("path=%s want /closed-positions", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "0xuser" || r.URL.Query().Get("limit") != "3" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]ClosedPosition{{TokenID: "token-2"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	rows, err := client.ClosedPositionsWithLimit(context.Background(), "0xuser", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].TokenID != "token-2" {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestActivityAcceptsNumericMarketFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activity" {
			t.Fatalf("path=%s want /activity", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"type":"TRADE","price":0.45,"size":2,"timestamp":1710000000}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	rows, err := client.Activity(context.Background(), "0xuser", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Price != "0.45" || rows[0].Size != "2" || rows[0].Timestamp != "1710000000" {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestOpenInterestUsesCurrentOIEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oi" {
			t.Fatalf("path=%s want /oi", r.URL.Path)
		}
		if r.URL.Query().Get("market") != "0xcondition" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]OpenInterest{{Market: "0xcondition", OpenValue: 25}})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	row, err := client.OpenInterest(context.Background(), "0xcondition")
	if err != nil {
		t.Fatal(err)
	}
	if row.Market != "0xcondition" || row.OpenValue != 25 {
		t.Fatalf("row=%+v", row)
	}
}

func TestTopHoldersUsesCurrentHoldersEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/holders" {
			t.Fatalf("path=%s want /holders", r.URL.Path)
		}
		if r.URL.Query().Get("market") != "0xcondition" || r.URL.Query().Get("limit") != "2" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"token": "token-1",
			"holders": []map[string]any{{
				"proxyWallet": "0xholder",
				"amount":      7.5,
			}},
		}})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	rows, err := client.TopHolders(context.Background(), "0xcondition", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Address != "0xholder" || rows[0].Shares != 7.5 {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestTotalValueUsesCurrentValueEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/value" {
			t.Fatalf("path=%s want /value", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "0xuser" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]TotalValue{{User: "0xuser", Value: 42}})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	row, err := client.TotalValue(context.Background(), "0xuser")
	if err != nil {
		t.Fatal(err)
	}
	if row.User != "0xuser" || row.Value != 42 {
		t.Fatalf("row=%+v", row)
	}
}

func TestMarketsTradedUsesCurrentTradedEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/traded" {
			t.Fatalf("path=%s want /traded", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "0xuser" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"user": "0xuser", "traded": 3})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	row, err := client.MarketsTraded(context.Background(), "0xuser")
	if err != nil {
		t.Fatal(err)
	}
	if row.User != "0xuser" || row.MarketsTraded != 3 {
		t.Fatalf("row=%+v", row)
	}
}

func TestLiveVolumeUsesEventID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/live-volume" {
			t.Fatalf("path=%s want /live-volume", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "2144505" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(LiveVolumeResponse{
			Total: 42,
			Markets: []LiveVolumeMarket{{
				Market: "0xcondition",
				Value:  42,
			}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	row, err := client.LiveVolume(context.Background(), 2144505)
	if err != nil {
		t.Fatal(err)
	}
	if row.Total != 42 || len(row.Markets) != 1 || row.Markets[0].Market != "0xcondition" {
		t.Fatalf("row=%+v", row)
	}
}

func TestTraderLeaderboardUsesCurrentV1Endpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/leaderboard" {
			t.Fatalf("path=%s want /v1/leaderboard", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"rank":        "1",
			"proxyWallet": "0xleader",
			"vol":         123.45,
			"pnl":         6.7,
		}})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	rows, err := client.TraderLeaderboard(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Rank != 1 || rows[0].User != "0xleader" || rows[0].Volume != 123.45 {
		t.Fatalf("rows=%+v", rows)
	}
}
