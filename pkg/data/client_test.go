package data

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

func TestClientCurrentPositionsReturnsPublicTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			t.Fatalf("path=%s want /positions", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "0xuser" || r.URL.Query().Get("limit") != "2" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"asset":       "token-1",
			"conditionId": "condition-1",
			"size":        7.5,
			"cashPnl":     1.25,
			"redeemable":  true,
		}})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	rows, err := client.CurrentPositionsWithLimit(context.Background(), "0xuser", 2)
	if err != nil {
		t.Fatal(err)
	}
	var publicRows []types.Position = rows
	if len(publicRows) != 1 || publicRows[0].TokenID != "token-1" || publicRows[0].CashPnl != 1.25 {
		t.Fatalf("rows=%+v", publicRows)
	}
	if !publicRows[0].Redeemable {
		t.Fatalf("redeemable not threaded: %+v", publicRows[0])
	}
}

func TestClientLiveVolumeReturnsPublicTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/live-volume" {
			t.Fatalf("path=%s want /live-volume", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"total": 1,
			"events": []map[string]any{{
				"event_id": "event-1",
				"title":    "Volume event",
				"volume":   42.0,
			}},
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	volume, err := client.LiveVolume(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	var publicVolume *types.LiveVolumeResponse = volume
	if publicVolume.Total != 1 || len(publicVolume.Events) != 1 || publicVolume.Events[0].EventID != "event-1" {
		t.Fatalf("volume=%+v", publicVolume)
	}
}
