package gamma

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

func TestActiveMarketsUsesContextAndParsesMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/markets" {
			t.Fatalf("path = %q, want /markets", r.URL.Path)
		}
		if r.URL.Query().Get("active") != "true" {
			t.Fatalf("active query = %q, want true", r.URL.Query().Get("active"))
		}
		if r.URL.Query().Get("closed") != "false" {
			t.Fatalf("closed query = %q, want false", r.URL.Query().Get("closed"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1","slug":"m-1","question":"Will it rain?","active":true,"closed":false}]`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)
	markets, err := client.ActiveMarkets(context.Background())
	if err != nil {
		t.Fatalf("ActiveMarkets returned error: %v", err)
	}
	if len(markets) != 1 || markets[0].Slug != "m-1" {
		t.Fatalf("unexpected markets: %#v", markets)
	}
}

func TestMarketsWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "5" {
			t.Fatalf("limit = %q, want 5", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	active := true
	_, err := client.Markets(context.Background(), &polytypes.GetMarketsParams{
		Limit:  5,
		Active: &active,
	})
	if err != nil {
		t.Fatalf("Markets returned error: %v", err)
	}
}

func TestHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":"ok"}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	resp, err := client.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("HealthCheck returned error: %v", err)
	}
	if resp.Data != "ok" {
		t.Fatalf("HealthCheck Data = %q, want ok", resp.Data)
	}
}

func TestSearchCallsCorrectEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/public-search" {
			t.Fatalf("path = %q, want /public-search", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "btc" {
			t.Fatalf("q = %q, want btc", r.URL.Query().Get("q"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"events":[],"tags":[],"profiles":[],"pagination":{"hasMore":false,"totalResults":0}}`))
	}))
	defer server.Close()

	cfg := transport.DefaultConfig(server.URL + "/")
	cfg.RetryMax = 0
	tc := transport.New(server.Client(), cfg)
	client := NewClient(server.URL+"/", tc)

	resp, err := client.Search(context.Background(), &polytypes.SearchParams{
		Q: "btc",
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if resp.Pagination.TotalResults != 0 {
		t.Fatalf("unexpected pagination: %#v", resp.Pagination)
	}
}

func TestEventBySlugUsesEventsSlugQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events" {
			t.Fatalf("path = %q, want /events", r.URL.Path)
		}
		if r.URL.Query().Get("slug") != "btc-updown-5m-1778115300" {
			t.Fatalf("slug query = %q", r.URL.Query().Get("slug"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id":"event-1",
			"slug":"btc-updown-5m-1778115300",
			"title":"Bitcoin Up or Down - May 6, 8:55PM-9:00PM ET",
			"markets":[{"id":"market-1","slug":"btc-updown-5m-1778115300"}]
		}]`))
	}))
	defer server.Close()

	cfg := transport.DefaultConfig(server.URL + "/")
	cfg.RetryMax = 0
	tc := transport.New(server.Client(), cfg)
	client := NewClient(server.URL+"/", tc)

	event, err := client.EventBySlug(context.Background(), "btc-updown-5m-1778115300")
	if err != nil {
		t.Fatalf("EventBySlug returned error: %v", err)
	}
	if event.ID != "event-1" || len(event.Markets) != 1 {
		t.Fatalf("unexpected event: %#v", event)
	}
}
