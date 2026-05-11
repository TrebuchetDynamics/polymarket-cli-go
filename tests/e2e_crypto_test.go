package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

// TestCryptoMarketDiscovery verifies the crypto market discovery pipeline
// against live Polymarket Gamma and CLOB APIs.
// Skipped under -short; requires network access.
func TestCryptoMarketDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live crypto discovery in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := gamma.NewClient("https://gamma-api.polymarket.com", nil)

	searchLimit := 20
	resp, err := client.Search(ctx, &polytypes.SearchParams{
		Q:            "bitcoin",
		LimitPerType: &searchLimit,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(resp.Events) == 0 {
		t.Fatal("no events found for bitcoin search")
	}

	var marketCount int
	var tokenCount int
	for _, event := range resp.Events {
		if !event.Active {
			continue
		}
		for _, market := range event.Markets {
			if !market.Active || market.Closed {
				continue
			}
			marketCount++
			if market.ClobTokenIDs != "" {
				tokenCount++
			}
		}
	}

	if marketCount == 0 {
		t.Fatal("no active markets found in bitcoin events")
	}
	if tokenCount == 0 {
		t.Fatal("no markets with CLOB token IDs found")
	}

	t.Logf("found %d events, %d active markets, %d with token IDs",
		len(resp.Events), marketCount, tokenCount)
}

// TestCryptoMarketDiscoveryETH verifies ETH market discovery.
func TestCryptoMarketDiscoveryETH(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live crypto discovery in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := gamma.NewClient("https://gamma-api.polymarket.com", nil)

	searchLimit := 10
	resp, err := client.Search(ctx, &polytypes.SearchParams{
		Q:            "ethereum",
		LimitPerType: &searchLimit,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(resp.Events) == 0 {
		t.Fatal("no events found for ethereum search")
	}

	t.Logf("found %d ethereum events", len(resp.Events))
}

func TestCryptoMarketDiscoveryWithInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live crypto discovery in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := gamma.NewClient("https://gamma-api.polymarket.com", nil)

	searchLimit := 50
	resp, err := client.Search(ctx, &polytypes.SearchParams{
		Q:            "crypto",
		LimitPerType: &searchLimit,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	var shortTermCount int
	for _, event := range resp.Events {
		if !event.Active {
			continue
		}
		for _, market := range event.Markets {
			if !market.Active || market.Closed {
				continue
			}
			if containsShortTerm(market.Question) || containsShortTerm(event.Title) {
				shortTermCount++
			}
		}
	}

	t.Logf("found %d short-term crypto markets", shortTermCount)
}

func containsShortTerm(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "5m") ||
		strings.Contains(s, "5 min") ||
		strings.Contains(s, "15m") ||
		strings.Contains(s, "15 min") ||
		strings.Contains(s, "1h") ||
		strings.Contains(s, "1 hour")
}
