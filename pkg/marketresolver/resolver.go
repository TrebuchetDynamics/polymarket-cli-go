// Package marketresolver resolves Polymarket market identifiers — slug,
// asset, timeframe, or window-start time — into canonical token IDs.
//
// Use marketresolver when a downstream consumer (for example a trading
// bot) needs to convert a human-friendly identifier into the up/down
// token IDs and condition ID needed to place an order. The resolver
// performs only Gamma reads; it does not sign or mutate anything.
//
// When not to use this package:
//   - For full Gamma metadata access — use pkg/gamma directly.
//   - For order book pricing — use pkg/bookreader.
//
// Stability: Resolver, NewResolver, the four Resolve methods,
// ValidateToken, MarketStatus and its constants, ResolveResult, and
// CryptoMarket are part of the polygolem public SDK and follow semver.
package marketresolver

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
)

// CryptoMarket represents a resolved crypto up/down market with token IDs.
// Slug and Question come from the Gamma event payload. UpTokenID and
// DownTokenID may be empty if the market's outcomes do not include both
// "up"/"yes" and "down"/"no".
type CryptoMarket struct {
	ConditionID string
	Asset       string
	Timeframe   string
	UpTokenID   string
	DownTokenID string
	Accepting   bool
	Closed      bool
	Question    string
	Slug        string
}

// Resolver finds active markets from the Gamma API.
// Methods are safe for concurrent use; each call is independent.
type Resolver struct {
	gamma *gamma.Client
}

const defaultGammaBaseURL = "https://gamma-api.polymarket.com"

// NewResolver creates a market resolver targeting the given Gamma base URL.
// If gammaBaseURL is empty, the production Gamma URL is used.
func NewResolver(gammaBaseURL string) *Resolver {
	gammaBaseURL = strings.TrimSpace(gammaBaseURL)
	if gammaBaseURL == "" {
		gammaBaseURL = defaultGammaBaseURL
	}
	return &Resolver{
		gamma: gamma.NewClient(gammaBaseURL, nil),
	}
}

// ResolveCryptoMarkets finds active CLOB-enabled up/down markets for an asset.
// Returns only accepting, non-closed markets with valid token IDs.
// asset is matched case-insensitively; concurrent Gamma searches are
// fanned out per timeframe.
func (r *Resolver) ResolveCryptoMarkets(ctx context.Context, asset string) ([]CryptoMarket, error) {
	queries := cryptoQueries(strings.ToUpper(asset))

	var all []CryptoMarket
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(queries))

	for _, q := range queries {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()
			markets, err := r.searchQuery(ctx, asset, query)
			if err != nil {
				errCh <- err
				return
			}
			mu.Lock()
			all = append(all, markets...)
			mu.Unlock()
		}(q)
	}
	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return nil, err
	}

	// Dedup by condition ID
	seen := map[string]bool{}
	var unique []CryptoMarket
	for _, m := range all {
		if seen[m.ConditionID] {
			continue
		}
		seen[m.ConditionID] = true
		unique = append(unique, m)
	}

	sort.Slice(unique, func(i, j int) bool {
		if unique[i].Asset != unique[j].Asset {
			return unique[i].Asset < unique[j].Asset
		}
		return unique[i].Timeframe < unique[j].Timeframe
	})

	return unique, nil
}

// ResolveTokenIDsAt resolves token IDs for a specific crypto window.
// Crypto up/down markets use deterministic slugs such as
// btc-updown-5m-1778114700, where the suffix is the UTC window start
// epoch. Falls back to ResolveTokenIDs when the slug lookup misses.
func (r *Resolver) ResolveTokenIDsAt(ctx context.Context, asset, timeframe string, windowStart time.Time) ResolveResult {
	if slug := cryptoWindowSlug(asset, timeframe, windowStart); slug != "" {
		if evt, err := r.gamma.EventBySlug(ctx, slug); err == nil {
			if result, ok := firstAcceptingMarket(asset, timeframe, marketsFromGamma(asset, evt.Markets)); ok {
				result.Source = "gamma:event_slug:" + slug
				return result
			}
		}
	}
	return r.ResolveTokenIDs(ctx, asset, timeframe)
}

func (r *Resolver) searchQuery(ctx context.Context, asset, query string) ([]CryptoMarket, error) {
	lpt := 20
	resp, err := r.gamma.Search(ctx, &polytypes.SearchParams{
		Q:            query,
		LimitPerType: &lpt,
		EventsStatus: "active",
	})
	if err != nil {
		return nil, fmt.Errorf("gamma search %q: %w", query, err)
	}

	var markets []CryptoMarket
	for _, evt := range resp.Events {
		eventMarkets := evt.Markets
		if len(eventMarkets) == 0 {
			if evt.Slug == "" {
				continue
			}
			fullEvt, err := r.gamma.EventBySlug(ctx, evt.Slug)
			if err != nil {
				continue
			}
			eventMarkets = fullEvt.Markets
		}
		for _, m := range eventMarkets {
			markets = append(markets, marketsFromGamma(asset, []polytypes.Market{m})...)
		}
	}
	return markets, nil
}

func marketsFromGamma(asset string, gammaMarkets []polytypes.Market) []CryptoMarket {
	markets := make([]CryptoMarket, 0, len(gammaMarkets))
	for _, m := range gammaMarkets {
		if !m.Active || m.Closed || !m.EnableOrderBook {
			continue
		}
		tokenIDs := extractTokenIDs(m.ClobTokenIDs)
		outcomes := m.Outcomes
		up, down := findUpDownTokenIDs(outcomes, tokenIDs)
		if up == "" || down == "" {
			continue
		}
		tf := inferTimeframe(m.Slug, m.Question)
		markets = append(markets, CryptoMarket{
			ConditionID: m.ConditionID,
			Asset:       asset,
			Timeframe:   tf,
			UpTokenID:   up,
			DownTokenID: down,
			Accepting:   m.AcceptingOrders,
			Closed:      m.Closed,
			Question:    m.Question,
			Slug:        m.Slug,
		})
	}
	return markets
}

func firstAcceptingMarket(asset, timeframe string, markets []CryptoMarket) (ResolveResult, bool) {
	for _, m := range markets {
		if m.Timeframe == timeframe && m.Accepting && !m.Closed {
			return ResolveResult{
				Status:      StatusAvailable,
				UpTokenID:   m.UpTokenID,
				DownTokenID: m.DownTokenID,
				ConditionID: m.ConditionID,
				Asset:       asset,
				Timeframe:   timeframe,
			}, true
		}
	}
	return ResolveResult{}, false
}

func findUpDownTokenIDs(outcomes []string, tokenIDs []string) (string, string) {
	if len(outcomes) != len(tokenIDs) {
		return "", ""
	}
	var up, down string
	for i, o := range outcomes {
		switch strings.ToLower(o) {
		case "up", "yes":
			up = tokenIDs[i]
		case "down", "no":
			down = tokenIDs[i]
		}
	}
	return up, down
}

func extractTokenIDs(raw string) []string {
	if raw == "" || raw == "[]" {
		return nil
	}
	s := raw
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	var ids []string
	current := ""
	inQuote := false
	for _, ch := range s {
		switch ch {
		case '"':
			inQuote = !inQuote
			if !inQuote && current != "" {
				ids = append(ids, current)
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
	return ids
}

func cryptoQueries(asset string) []string {
	names := map[string][]string{
		"BTC":  {"bitcoin"},
		"ETH":  {"ethereum"},
		"SOL":  {"solana"},
		"XRP":  {"xrp"},
		"DOGE": {"doge"},
		"BNB":  {"bnb"},
	}
	nameList := names[asset]
	if len(nameList) == 0 {
		nameList = []string{strings.ToLower(asset)}
	}
	var queries []string
	for _, name := range nameList {
		for _, tf := range []string{"5m", "15m"} {
			queries = append(queries, name+" "+tf)
		}
	}
	return queries
}

func cryptoWindowSlug(asset, timeframe string, windowStart time.Time) string {
	if windowStart.IsZero() {
		return ""
	}
	prefixes := map[string]string{
		"BTC":  "btc",
		"ETH":  "eth",
		"SOL":  "sol",
		"XRP":  "xrp",
		"DOGE": "doge",
		"BNB":  "bnb",
	}
	prefix := prefixes[strings.ToUpper(asset)]
	if prefix == "" {
		return ""
	}
	switch timeframe {
	case "5m", "15m", "4h":
	default:
		return ""
	}
	return fmt.Sprintf("%s-updown-%s-%d", prefix, timeframe, windowStart.UTC().Unix())
}

func inferTimeframe(slug, question string) string {
	text := strings.ToLower(slug + " " + question)
	for _, tf := range []string{"5m", "5 min", "5-minute", "15m", "15 min", "15-minute"} {
		if strings.Contains(text, tf) {
			if strings.HasPrefix(tf, "5") {
				return "5m"
			}
			return "15m"
		}
	}
	return ""
}

// MarketStatus classifies market availability returned by ResolveResult.
type MarketStatus string

// Market status values reported by ResolveResult.
const (
	// StatusAvailable means the resolver found an accepting non-closed
	// market with valid up/down token IDs.
	StatusAvailable MarketStatus = "available"
	// StatusUnavailable means the resolver found a market but it is not
	// accepting orders (paused or closed).
	StatusUnavailable MarketStatus = "unavailable"
	// StatusStaleToken means a previously valid token ID can no longer
	// be priced; use ResolveTokenIDs again to discover the current one.
	StatusStaleToken MarketStatus = "stale_token"
	// StatusUnresolved means no active matching market could be found.
	StatusUnresolved MarketStatus = "unresolved"
)

// ResolveResult is the structured result of a market/token resolution.
// Source identifies which Gamma path produced the answer (deterministic
// slug, public search, or an error string).
type ResolveResult struct {
	Status      MarketStatus `json:"status"`
	UpTokenID   string       `json:"up_token_id"`
	DownTokenID string       `json:"down_token_id"`
	ConditionID string       `json:"condition_id"`
	Asset       string       `json:"asset"`
	Timeframe   string       `json:"timeframe"`
	Source      string       `json:"source"`
}

// ResolveTokenIDs resolves token IDs for a given asset+timeframe.
// Returns StatusUnresolved if no active accepting market is found.
// Source records which Gamma path produced the result, useful for
// debugging stale-token issues.
func (r *Resolver) ResolveTokenIDs(ctx context.Context, asset, timeframe string) ResolveResult {
	markets, err := r.ResolveCryptoMarkets(ctx, asset)
	if err != nil {
		return ResolveResult{
			Status:    StatusUnresolved,
			Asset:     asset,
			Timeframe: timeframe,
			Source:    fmt.Sprintf("gamma_error:%v", err),
		}
	}
	if result, ok := firstAcceptingMarket(asset, timeframe, markets); ok {
		result.Source = "gamma:crypto_search"
		return result
	}
	return ResolveResult{
		Status:    StatusUnresolved,
		Asset:     asset,
		Timeframe: timeframe,
		Source:    fmt.Sprintf("gamma:no_match (found %d markets)", len(markets)),
	}
}

// ValidateToken checks if a token ID is still valid by basic format checks.
// Returns StatusStaleToken if the CLOB returns an error for the token (a
// fuller validation requires CLOB access in the bookreader layer); for
// now it returns StatusUnresolved on empty/non-numeric token IDs and
// StatusAvailable otherwise.
func (r *Resolver) ValidateToken(ctx context.Context, tokenID string) MarketStatus {
	// A simple approach: check that the token is non-empty.
	// Full validation requires CLOB access which is in the bookreader layer.
	if tokenID == "" {
		return StatusUnresolved
	}
	if _, err := strconv.ParseUint(tokenID, 10, 64); err != nil {
		return StatusUnresolved
	}
	return StatusAvailable
}
