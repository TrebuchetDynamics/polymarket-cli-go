package marketresolver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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
						"resolutionSource": "https://data.chain.link/streams/btc-usd",
						"outcomes": ["Up", "Down"],
						"active": true,
						"closed": false,
						"enableOrderBook": true,
						"acceptingOrders": true,
						"orderMinSize": 5,
						"orderPriceMinTickSize": 0.01,
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
	if got.ResolutionSource != "https://data.chain.link/streams/btc-usd" || got.Question == "" || got.Slug != "btc-updown-5m-1778114700" {
		t.Fatalf("market metadata was not preserved: %#v", got)
	}
	if got.MinOrderSize != 5 || got.TickSize != 0.01 {
		t.Fatalf("market execution metadata was not preserved: %#v", got)
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
					"resolutionSource": "https://data.chain.link/streams/btc-usd",
					"outcomes": ["Up", "Down"],
					"active": true,
					"closed": false,
					"enableOrderBook": true,
					"acceptingOrders": true,
					"orderMinSize": 5,
					"orderPriceMinTickSize": 0.01,
					"clobTokenIds": "[\"up-token\", \"down-token\"]",
					"startDate": "2026-05-07T00:45:00Z",
					"endDate":   "2026-05-07T00:50:00Z"
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
	if got.ResolutionSource != "https://data.chain.link/streams/btc-usd" || got.Question == "" || got.Slug != "btc-updown-5m-1778114700" {
		t.Fatalf("market metadata was not preserved: %#v", got)
	}
	if got.MinOrderSize != 5 || got.TickSize != 0.01 {
		t.Fatalf("market execution metadata was not preserved: %#v", got)
	}
	if searchCalled {
		t.Fatal("resolver should try the deterministic slug before public search")
	}
}

func eventResponseWithStartDate(slug, conditionID, up, down, startDateISO, endDateISO string) string {
	return `[{
		"id": "event-1",
		"slug": "` + slug + `",
		"title": "test market",
		"active": true,
		"closed": false,
		"markets": [{
			"id": "market-1",
			"question": "test market",
			"conditionId": "` + conditionID + `",
			"slug": "` + slug + `",
			"outcomes": ["Up", "Down"],
			"active": true,
			"closed": false,
			"enableOrderBook": true,
			"acceptingOrders": true,
			"clobTokenIds": "[\"` + up + `\", \"` + down + `\"]",
			"startDate": "` + startDateISO + `",
			"endDate":   "` + endDateISO + `"
		}]
	}]`
}

func eventResponseWithEventStartTime(slug, conditionID, up, down, startDateISO, eventStartTimeISO, endDateISO string) string {
	return `[{
		"id": "event-1",
		"slug": "` + slug + `",
		"title": "test market",
		"active": true,
		"closed": false,
		"markets": [{
			"id": "market-1",
			"question": "test market",
			"conditionId": "` + conditionID + `",
			"slug": "` + slug + `",
			"outcomes": ["Up", "Down"],
			"active": true,
			"closed": false,
			"enableOrderBook": true,
			"acceptingOrders": true,
			"clobTokenIds": "[\"` + up + `\", \"` + down + `\"]",
			"startDate": "` + startDateISO + `",
			"eventStartTime": "` + eventStartTimeISO + `",
			"endDate":   "` + endDateISO + `"
		}]
	}]`
}

func TestResolveTokenIDsForWindow_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/events":
			if r.URL.Query().Get("slug") != "btc-updown-5m-1778114700" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(eventResponseWithStartDate(
				"btc-updown-5m-1778114700", "cid", "up", "down",
				"2026-05-07T00:45:00Z", "2026-05-07T00:50:00Z",
			)))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.URL)
	got := resolver.ResolveTokenIDsForWindow(context.Background(), "BTC", "5m", time.Unix(1778114700, 0).UTC())
	if got.Status != StatusAvailable {
		t.Fatalf("status=%q source=%q", got.Status, got.Source)
	}
	if got.UpTokenID != "up" || got.DownTokenID != "down" || got.ConditionID != "cid" {
		t.Fatalf("tokens=%#v", got)
	}
	if got.Source != "gamma:event_slug_strict:btc-updown-5m-1778114700" {
		t.Fatalf("source=%q", got.Source)
	}
	if !got.StartDate.Equal(time.Unix(1778114700, 0).UTC()) {
		t.Fatalf("startDate=%v want %v", got.StartDate, time.Unix(1778114700, 0).UTC())
	}
}

func TestResolveTokenIDsForWindowUsesEventStartTimeForRecurringCryptoMarkets(t *testing.T) {
	window := time.Unix(1778385900, 0).UTC() // 2026-05-10T04:05:00Z
	slug := "btc-updown-5m-1778385900"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/events":
			if r.URL.Query().Get("slug") != slug {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(eventResponseWithEventStartTime(
				slug, "cid", "up", "down",
				"2026-05-09T04:12:49Z",
				window.Format(time.RFC3339),
				window.Add(5*time.Minute).Format(time.RFC3339),
			)))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.URL)
	got := resolver.ResolveTokenIDsForWindow(context.Background(), "BTC", "5m", window)
	if got.Status != StatusAvailable {
		t.Fatalf("status=%q source=%q start=%v", got.Status, got.Source, got.StartDate)
	}
	if !got.StartDate.Equal(window) || !got.EndDate.Equal(window.Add(5*time.Minute)) {
		t.Fatalf("window=%v-%v want %v-%v", got.StartDate, got.EndDate, window, window.Add(5*time.Minute))
	}
}

func TestResolveTokenIDsForWindow_RejectsWrongWindow(t *testing.T) {
	wantWindow := time.Unix(1778114700, 0).UTC() // 2026-05-07T00:45:00Z
	gotWindow := wantWindow.Add(5 * time.Minute) // 2026-05-07T00:50:00Z

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/events":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(eventResponseWithStartDate(
				"btc-updown-5m-1778114700", "cid", "up", "down",
				gotWindow.Format(time.RFC3339), gotWindow.Add(5*time.Minute).Format(time.RFC3339),
			)))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.URL)
	got := resolver.ResolveTokenIDsForWindow(context.Background(), "BTC", "5m", wantWindow)
	if got.Status != StatusWindowMismatch {
		t.Fatalf("status=%q source=%q", got.Status, got.Source)
	}
	if got.Source == "" || !strings.Contains(got.Source, "slug_hit_window_mismatch") {
		t.Fatalf("source must call out window mismatch: %q", got.Source)
	}
	if !got.StartDate.Equal(gotWindow) {
		t.Fatalf("startDate=%v want %v (the wrong-window market's bound)", got.StartDate, gotWindow)
	}
}

func TestResolveTokenIDsForWindow_NeverFallsThrough(t *testing.T) {
	var searchCalls, eventCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/events":
			eventCalls++
			http.NotFound(w, r) // slug miss
		case "/public-search":
			searchCalls++
			http.Error(w, "strict resolver must not fall through to public-search", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.URL)
	got := resolver.ResolveTokenIDsForWindow(context.Background(), "BTC", "5m", time.Unix(1778114700, 0).UTC())
	if got.Status != StatusUnresolved {
		t.Fatalf("status=%q source=%q", got.Status, got.Source)
	}
	if eventCalls != 1 {
		t.Fatalf("event lookup called %d times, want 1", eventCalls)
	}
	if searchCalls != 0 {
		t.Fatalf("strict resolver fell through to public-search (%d calls)", searchCalls)
	}
}

func TestResolveTokenIDsAt_FailsClosedOnSlugHitMismatch(t *testing.T) {
	// Reproduces the 2026-05-09 SOL trap: signal bar 08:20 UTC tried to buy
	// the 08:20 market, but the slug-hit returned a market starting 12:20 UTC.
	signalWindow := time.Date(2026, 5, 9, 8, 20, 0, 0, time.UTC)
	wrongWindow := time.Date(2026, 5, 9, 12, 20, 0, 0, time.UTC)
	signalSlug := "sol-updown-5m-" + strconv.FormatInt(signalWindow.Unix(), 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/events":
			if r.URL.Query().Get("slug") != signalSlug {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(eventResponseWithStartDate(
				signalSlug, "cid-wrong", "up", "down",
				wrongWindow.Format(time.RFC3339), wrongWindow.Add(5*time.Minute).Format(time.RFC3339),
			)))
		case "/public-search":
			t.Fatal("ResolveTokenIDsAt must not fall through to public-search after a slug-hit window mismatch")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.URL)
	got := resolver.ResolveTokenIDsAt(context.Background(), "SOL", "5m", signalWindow)
	if got.Status != StatusWindowMismatch {
		t.Fatalf("status=%q source=%q (should fail closed on wrong-window slug hit)", got.Status, got.Source)
	}
	if !got.StartDate.Equal(wrongWindow) {
		t.Fatalf("startDate=%v want %v", got.StartDate, wrongWindow)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func intToString(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	s := string(buf[i:])
	if neg {
		s = "-" + s
	}
	return s
}
