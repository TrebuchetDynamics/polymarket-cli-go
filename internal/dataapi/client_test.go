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
