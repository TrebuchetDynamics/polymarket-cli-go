package gamma

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActiveMarketsUsesContextAndParsesMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Fatalf("path = %q, want /markets", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1","slug":"m-1","question":"Will it rain?","active":true,"closed":false}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	markets, err := client.ActiveMarkets(context.Background())
	if err != nil {
		t.Fatalf("ActiveMarkets returned error: %v", err)
	}
	if len(markets) != 1 || markets[0].Slug != "m-1" {
		t.Fatalf("unexpected markets: %#v", markets)
	}
}
