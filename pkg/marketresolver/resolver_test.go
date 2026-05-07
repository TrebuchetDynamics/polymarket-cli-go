package marketresolver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResolveTokenIDsUsesEmbeddedPublicSearchMarkets(t *testing.T) {
	eventLookupCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/public-search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"events": [{
					"id": "event-1",
					"slug": "btc-updown-5m-1778114700",
					"title": "Bitcoin Up or Down - May 6, 8:45PM-8:50PM ET",
					"active": true,
					"closed": false,
					"markets": [{
						"id": "market-1",
						"question": "Bitcoin Up or Down - May 6, 8:45PM-8:50PM ET",
						"conditionId": "condition-1",
						"slug": "btc-updown-5m-1778114700",
						"outcomes": ["Up", "Down"],
						"active": true,
						"closed": false,
						"enableOrderBook": true,
						"acceptingOrders": true,
						"clobTokenIds": "[\"up-token\", \"down-token\"]"
					}]
				}],
				"tags": [],
				"profiles": [],
				"pagination": {"hasMore": false, "totalResults": 1}
			}`))
		case "/events/btc-updown-5m-1778114700":
			eventLookupCalled = true
			http.Error(w, "event lookup should not be required", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.URL)
	got := resolver.ResolveTokenIDs(context.Background(), "BTC", "5m")

	if got.Status != StatusAvailable {
		t.Fatalf("status=%q source=%q", got.Status, got.Source)
	}
	if got.ConditionID != "condition-1" || got.UpTokenID != "up-token" || got.DownTokenID != "down-token" {
		t.Fatalf("unexpected tokens: %#v", got)
	}
	if eventLookupCalled {
		t.Fatal("resolver should use embedded public-search markets before calling event detail")
	}
}

func TestResolveTokenIDsAtUsesDeterministicCryptoSlug(t *testing.T) {
	searchCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/events":
			if r.URL.Query().Get("slug") != "btc-updown-5m-1778114700" {
				t.Fatalf("slug query=%q", r.URL.Query().Get("slug"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{
				"id": "event-1",
				"slug": "btc-updown-5m-1778114700",
				"title": "Bitcoin Up or Down - May 6, 8:45PM-8:50PM ET",
				"active": true,
				"closed": false,
				"markets": [{
					"id": "market-1",
					"question": "Bitcoin Up or Down - May 6, 8:45PM-8:50PM ET",
					"conditionId": "condition-1",
					"slug": "btc-updown-5m-1778114700",
					"outcomes": ["Up", "Down"],
					"active": true,
					"closed": false,
					"enableOrderBook": true,
					"acceptingOrders": true,
					"clobTokenIds": "[\"up-token\", \"down-token\"]"
				}]
			}]`))
		case "/public-search":
			searchCalled = true
			http.Error(w, "search should not be required for deterministic crypto windows", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.URL)
	got := resolver.ResolveTokenIDsAt(context.Background(), "BTC", "5m", time.Unix(1778114700, 0).UTC())

	if got.Status != StatusAvailable {
		t.Fatalf("status=%q source=%q", got.Status, got.Source)
	}
	if got.ConditionID != "condition-1" || got.UpTokenID != "up-token" || got.DownTokenID != "down-token" {
		t.Fatalf("unexpected tokens: %#v", got)
	}
	if searchCalled {
		t.Fatal("resolver should try the deterministic slug before public search")
	}
}
