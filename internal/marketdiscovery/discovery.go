// Package marketdiscovery provides market enrichment by joining Gamma metadata with CLOB details.
package marketdiscovery

import (
	"context"
	"fmt"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

// Service enriches Gamma markets with CLOB data.
type Service struct {
	gamma *gamma.Client
	clob  *clob.Client
}

// New creates a market discovery service.
func New(gammaClient *gamma.Client, clobClient *clob.Client) *Service {
	return &Service{gamma: gammaClient, clob: clobClient}
}

// EnrichMarket fetches CLOB details for a Gamma market and returns an EnrichedMarket.
func (s *Service) EnrichMarket(ctx context.Context, market polytypes.Market) (*polytypes.EnrichedMarket, error) {
	tokenIDs := extractTokenIDs(market.ClobTokenIDs)
	if len(tokenIDs) == 0 {
		return nil, fmt.Errorf("market %s has no CLOB token IDs", market.ID)
	}

	// Use the first token ID for CLOB queries
	tokenID := tokenIDs[0]

	enriched := &polytypes.EnrichedMarket{Market: market}

	// Fetch tick size
	tickSize, err := s.clob.TickSize(ctx, tokenID)
	if err != nil {
		return enriched, fmt.Errorf("tick size for %s: %w", tokenID, err)
	}
	enriched.TickSize = *tickSize

	// Fetch neg risk
	negRisk, err := s.clob.NegRisk(ctx, tokenID)
	if err != nil {
		return enriched, fmt.Errorf("neg risk for %s: %w", tokenID, err)
	}
	enriched.NegRisk = negRisk.NegRisk

	// Fetch fee rate
	feeRate, err := s.clob.FeeRateBps(ctx, tokenID)
	if err != nil {
		return enriched, fmt.Errorf("fee rate for %s: %w", tokenID, err)
	}
	enriched.FeeRateBps = feeRate

	// Optional: fetch order book
	if ob, err := s.clob.OrderBook(ctx, tokenID); err == nil {
		enriched.OrderBook = ob
	}

	// Optional: fetch last price
	if price, err := s.clob.LastTradePrice(ctx, tokenID); err == nil {
		enriched.LastPrice = price
	}

	// Optional: fetch midpoint
	if mid, err := s.clob.Midpoint(ctx, tokenID); err == nil {
		enriched.Midpoint = mid
	}

	// Optional: fetch spread
	if spread, err := s.clob.Spread(ctx, tokenID); err == nil {
		enriched.Spread = spread
	}

	return enriched, nil
}

// EnrichedMarkets filters active Gamma markets and enriches them with CLOB data.
func (s *Service) EnrichedMarkets(ctx context.Context, limit int) ([]polytypes.EnrichedMarket, error) {
	active := true
	closed := false
	markets, err := s.gamma.Markets(ctx, &polytypes.GetMarketsParams{
		Active: &active,
		Closed: &closed,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch gamma markets: %w", err)
	}

	var enriched []polytypes.EnrichedMarket
	for _, m := range markets {
		em, err := s.EnrichMarket(ctx, m)
		if err != nil {
			// Non-fatal: skip markets that can't be enriched
			continue
		}
		enriched = append(enriched, *em)
	}
	return enriched, nil
}

// SearchAndEnrich searches Gamma and enriches the results.
func (s *Service) SearchAndEnrich(ctx context.Context, query string, limit int) ([]polytypes.EnrichedMarket, error) {
	searchResp, err := s.gamma.Search(ctx, &polytypes.SearchParams{
		Q:            query,
		LimitPerType: &limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search gamma: %w", err)
	}

	var enriched []polytypes.EnrichedMarket
	for _, evt := range searchResp.Events {
		// Events in search response don't have full market data.
		// We need to fetch the full event to get markets.
		fullEvent, err := s.gamma.EventBySlug(ctx, evt.Slug)
		if err != nil {
			continue
		}
		for _, m := range fullEvent.Markets {
			if !m.Active || !m.EnableOrderBook {
				continue
			}
			em, err := s.EnrichMarket(ctx, m)
			if err != nil {
				continue
			}
			enriched = append(enriched, *em)
		}
	}
	return enriched, nil
}

// extractTokenIDs parses the JSON-encoded CLOB token IDs string from Gamma.
func extractTokenIDs(raw string) []string {
	if raw == "" || raw == "[]" {
		return nil
	}
	// Gamma stores clobTokenIds as a JSON string like "[\"id1\",\"id2\"]"
	// Simple parsing: strip brackets and quotes
	s := raw
	s = trimPrefix(s, "[")
	s = trimSuffix(s, "]")
	if s == "" {
		return nil
	}
	parts := splitQuoted(s)
	return parts
}

func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func trimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}

func splitQuoted(s string) []string {
	var parts []string
	current := ""
	inQuote := false
	for _, ch := range s {
		switch ch {
		case '"':
			inQuote = !inQuote
			if !inQuote && current != "" {
				parts = append(parts, current)
				current = ""
			}
		case ',':
			if !inQuote {
				continue
			}
			current += string(ch)
		default:
			if inQuote {
				current += string(ch)
			}
		}
	}
	return parts
}
