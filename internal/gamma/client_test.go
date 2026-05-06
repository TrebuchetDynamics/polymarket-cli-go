package gamma

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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

	client := NewClient(server.URL+"/", server.Client())
	markets, err := client.ActiveMarkets(context.Background())
	if err != nil {
		t.Fatalf("ActiveMarkets returned error: %v", err)
	}
	if len(markets) != 1 || markets[0].Slug != "m-1" {
		t.Fatalf("unexpected markets: %#v", markets)
	}
}
