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

// TestCurrentPositionsDecodesV2Schema fixtures the Polymarket Data API
// camelCase response shape and verifies every redemption-relevant field
// decodes onto the typed struct. The fixture values mirror the live
// 2026-05-09 response from deposit wallet 0x21999a07...02D4 for the
// resolved ETH-Up market.
func TestCurrentPositionsDecodesV2Schema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"asset":           "10203228750887270363579341300435494148775390248158812958841180330451031762744",
			"conditionId":     "0xcondition",
			"eventId":         "event-eth",
			"proxyWallet":     "0x21999a074344610057c9b2B362332388a44502D4",
			"size":            4.0784,
			"avgPrice":        0.5099,
			"curPrice":        0.135,
			"cashPnl":         -1.5294,
			"percentPnl":      -73.5,
			"redeemable":      true,
			"mergeable":       false,
			"negativeRisk":    false,
			"outcome":         "Up",
			"outcomeIndex":    0,
			"oppositeOutcome": "Down",
			"oppositeAsset":   "10203228750887270363579341300435494148775390248158812958841180330451031762745",
			"endDate":         "2026-05-09",
			"title":           "Ethereum Up or Down - May 9, 4:40AM-4:45AM ET",
			"slug":            "eth-updown-5m-1778316000",
			"eventSlug":       "eth-updown-5m-1778316000",
		}})
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	rows, err := client.CurrentPositions(context.Background(), "0x21999a074344610057c9b2B362332388a44502D4")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows=%d want 1", len(rows))
	}
	p := rows[0]
	if !p.Redeemable {
		t.Errorf("redeemable=false, want true")
	}
	if p.Mergeable {
		t.Errorf("mergeable=true, want false")
	}
	if p.NegativeRisk {
		t.Errorf("negativeRisk=true, want false")
	}
	if p.Outcome != "Up" || p.OutcomeIndex != 0 || p.OppositeOutcome != "Down" {
		t.Errorf("outcome fields=%+v", p)
	}
	if p.AvgPrice != 0.5099 || p.CurrentPrice != 0.135 {
		t.Errorf("price fields avg=%v cur=%v", p.AvgPrice, p.CurrentPrice)
	}
	if p.Size != 4.0784 {
		t.Errorf("size=%v want 4.0784", p.Size)
	}
	if p.EndDate != "2026-05-09" {
		t.Errorf("endDate=%q", p.EndDate)
	}
	if p.Slug != "eth-updown-5m-1778316000" {
		t.Errorf("slug=%q", p.Slug)
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

func TestTradesDecodeCurrentDataAPIShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Fatalf("path=%s want /trades", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{
			"proxyWallet":"0xwallet",
			"side":"BUY",
			"asset":"token-sol-up",
			"conditionId":"0xsol",
			"size":2.862744,
			"price":0.5099998463013109,
			"timestamp":1778314880,
			"title":"Solana Up or Down",
			"slug":"sol-updown-5m-1778329200",
			"outcome":"Up",
			"transactionHash":"0xsoltx"
		}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, transport.New(server.Client(), transport.DefaultConfig(server.URL)))
	rows, err := client.Trades(context.Background(), "0xwallet", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows=%d want 1", len(rows))
	}
	row := rows[0]
	if row.Market != "0xsol" || row.AssetID != "token-sol-up" || row.TransactionHash != "0xsoltx" {
		t.Fatalf("row identifiers not decoded: %+v", row)
	}
	if row.CreatedAt != "1778314880" || row.Outcome != "Up" {
		t.Fatalf("row metadata not decoded: %+v", row)
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
