// Package universal is a single-stop client for all Polymarket market data.
//
// Use universal when you need to query markets, order books, prices, events,
// volume, leaderboards, and streaming data without managing multiple API
// clients. The client delegates to the internal gamma, clob, dataapi, and
// stream packages while presenting one typed surface.
//
// When not to use this package:
//   - For authenticated trading operations (order placement, cancellation) —
//     those require signing and are not part of the read-only universal SDK.
//   - For deposit wallet lifecycle — use internal/wallet or the CLI directly.
//
// Stability: Client, NewClient, DefaultConfig, and every method on Client
// are part of the polygolem public SDK and follow semver.
package universal

import (
	"context"
	"fmt"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/dataapi"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/marketdiscovery"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/stream"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

const (
	defaultGammaBaseURL = "https://gamma-api.polymarket.com"
	defaultCLOBBaseURL  = "https://clob.polymarket.com"
	defaultDataBaseURL  = "https://data-api.polymarket.com"
	defaultStreamURL    = "wss://ws-subscriptions-clob.polymarket.com/ws/"
)

// Client queries all Polymarket public data APIs through one surface.
// Methods are safe for concurrent use; each call is independent.
type Client struct {
	gamma     *gamma.Client
	clob      *clob.Client
	data      *dataapi.Client
	discovery *marketdiscovery.Service
}

// Config holds base URLs for Polymarket data endpoints.
// Zero values fall back to production URLs.
type Config struct {
	GammaBaseURL string
	CLOBBaseURL  string
	DataBaseURL  string
}

// DefaultConfig returns production defaults.
func DefaultConfig() Config {
	return Config{
		GammaBaseURL: defaultGammaBaseURL,
		CLOBBaseURL:  defaultCLOBBaseURL,
		DataBaseURL:  defaultDataBaseURL,
	}
}

// NewClient creates a universal client with production defaults.
// Pass an explicit Config to override endpoints.
func NewClient(cfg Config) *Client {
	if cfg.GammaBaseURL == "" {
		cfg.GammaBaseURL = defaultGammaBaseURL
	}
	if cfg.CLOBBaseURL == "" {
		cfg.CLOBBaseURL = defaultCLOBBaseURL
	}
	if cfg.DataBaseURL == "" {
		cfg.DataBaseURL = defaultDataBaseURL
	}

	gc := gamma.NewClient(cfg.GammaBaseURL, nil)
	cc := clob.NewClient(cfg.CLOBBaseURL, nil)
	dc := dataapi.NewClient(cfg.DataBaseURL, nil)

	return &Client{
		gamma:     gc,
		clob:      cc,
		data:      dc,
		discovery: marketdiscovery.New(gc, cc),
	}
}

// ActiveMarkets returns markets currently flagged active by Gamma.
func (c *Client) ActiveMarkets(ctx context.Context) ([]polytypes.Market, error) {
	return c.gamma.ActiveMarkets(ctx)
}

// Markets lists markets matching the given filter parameters.
func (c *Client) Markets(ctx context.Context, params *polytypes.GetMarketsParams) ([]polytypes.Market, error) {
	return c.gamma.Markets(ctx, params)
}

// MarketByID fetches a single market by its Gamma ID.
func (c *Client) MarketByID(ctx context.Context, id string) (*polytypes.Market, error) {
	return c.gamma.MarketByID(ctx, id)
}

// MarketBySlug fetches a single market by slug.
func (c *Client) MarketBySlug(ctx context.Context, slug string) (*polytypes.Market, error) {
	return c.gamma.MarketBySlug(ctx, slug)
}

// Events lists events matching the given filter parameters.
func (c *Client) Events(ctx context.Context, params *polytypes.GetEventsParams) ([]polytypes.Event, error) {
	return c.gamma.Events(ctx, params)
}

// EventByID fetches a single event by its Gamma ID.
func (c *Client) EventByID(ctx context.Context, id string) (*polytypes.Event, error) {
	return c.gamma.EventByID(ctx, id)
}

// EventBySlug fetches a single event by slug.
func (c *Client) EventBySlug(ctx context.Context, slug string) (*polytypes.Event, error) {
	return c.gamma.EventBySlug(ctx, slug)
}

// Search performs Gamma's public search across events, markets, tags, and profiles.
func (c *Client) Search(ctx context.Context, params *polytypes.SearchParams) (*polytypes.SearchResponse, error) {
	return c.gamma.Search(ctx, params)
}

// Series lists market series matching the given filter parameters.
func (c *Client) Series(ctx context.Context, params *polytypes.GetSeriesParams) ([]polytypes.Series, error) {
	return c.gamma.Series(ctx, params)
}

// Tags lists tags matching the given filter parameters.
func (c *Client) Tags(ctx context.Context, params *polytypes.GetTagsParams) ([]polytypes.Tag, error) {
	return c.gamma.Tags(ctx, params)
}

// SportsMetadata returns the current sports metadata catalog.
func (c *Client) SportsMetadata(ctx context.Context) ([]polytypes.SportMetadata, error) {
	return c.gamma.SportsMetadata(ctx)
}

// Comments returns comments matching the given query.
func (c *Client) Comments(ctx context.Context, params *polytypes.CommentQuery) ([]polytypes.Comment, error) {
	return c.gamma.Comments(ctx, params)
}

// OrderBook returns L2 order book depth for a token.
func (c *Client) OrderBook(ctx context.Context, tokenID string) (*polytypes.OrderBook, error) {
	return c.clob.OrderBook(ctx, tokenID)
}

// OrderBooks returns order books for multiple tokens.
func (c *Client) OrderBooks(ctx context.Context, params []polytypes.BookParams) ([]polytypes.OrderBook, error) {
	return c.clob.OrderBooks(ctx, params)
}

// Price returns the best bid or ask for a token.
func (c *Client) Price(ctx context.Context, tokenID, side string) (string, error) {
	return c.clob.Price(ctx, tokenID, side)
}

// Prices returns best prices for multiple tokens.
func (c *Client) Prices(ctx context.Context, params []polytypes.BookParams) (map[string]string, error) {
	return c.clob.Prices(ctx, params)
}

// Midpoint returns the midpoint price for a token.
func (c *Client) Midpoint(ctx context.Context, tokenID string) (string, error) {
	return c.clob.Midpoint(ctx, tokenID)
}

// Midpoints returns midpoint prices for multiple tokens.
func (c *Client) Midpoints(ctx context.Context, params []polytypes.BookParams) (map[string]string, error) {
	return c.clob.Midpoints(ctx, params)
}

// Spread returns the spread for a token.
func (c *Client) Spread(ctx context.Context, tokenID string) (string, error) {
	return c.clob.Spread(ctx, tokenID)
}

// TickSize returns the tick size for a token.
func (c *Client) TickSize(ctx context.Context, tokenID string) (*polytypes.TickSize, error) {
	return c.clob.TickSize(ctx, tokenID)
}

// NegRisk returns neg risk info for a token.
func (c *Client) NegRisk(ctx context.Context, tokenID string) (*polytypes.NegRiskInfo, error) {
	return c.clob.NegRisk(ctx, tokenID)
}

// FeeRateBps returns the fee rate in basis points for a token.
func (c *Client) FeeRateBps(ctx context.Context, tokenID string) (int, error) {
	return c.clob.FeeRateBps(ctx, tokenID)
}

// LastTradePrice returns the last trade price for a token.
func (c *Client) LastTradePrice(ctx context.Context, tokenID string) (string, error) {
	return c.clob.LastTradePrice(ctx, tokenID)
}

// LastTradesPrices returns last trade prices for multiple tokens.
func (c *Client) LastTradesPrices(ctx context.Context, params []polytypes.BookParams) (map[string]string, error) {
	return c.clob.LastTradesPrices(ctx, params)
}

// PricesHistory returns OHLCV price history.
func (c *Client) PricesHistory(ctx context.Context, params *polytypes.PriceHistoryParams) (*polytypes.PriceHistory, error) {
	return c.clob.PricesHistory(ctx, params)
}

// CLOBMarkets lists CLOB markets with cursor pagination.
func (c *Client) CLOBMarkets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	return c.clob.Markets(ctx, nextCursor)
}

// CLOBMarket returns a single CLOB market by condition ID.
func (c *Client) CLOBMarket(ctx context.Context, conditionID string) (*polytypes.CLOBMarket, error) {
	return c.clob.Market(ctx, conditionID)
}

// CurrentPositions returns current open positions for a user.
func (c *Client) CurrentPositions(ctx context.Context, user string) ([]dataapi.Position, error) {
	return c.data.CurrentPositions(ctx, user)
}

// Trades returns trades for a user.
func (c *Client) Trades(ctx context.Context, user string, limit int) ([]dataapi.Trade, error) {
	return c.data.Trades(ctx, user, limit)
}

// Activity returns recent activity for a user.
func (c *Client) Activity(ctx context.Context, user string, limit int) ([]dataapi.Activity, error) {
	return c.data.Activity(ctx, user, limit)
}

// TopHolders returns top holders for a token.
func (c *Client) TopHolders(ctx context.Context, tokenID string, limit int) ([]dataapi.MetaHolder, error) {
	return c.data.TopHolders(ctx, tokenID, limit)
}

// TotalValue returns total portfolio value for a user.
func (c *Client) TotalValue(ctx context.Context, user string) (*dataapi.TotalValue, error) {
	return c.data.TotalValue(ctx, user)
}

// OpenInterest returns open interest for a token.
func (c *Client) OpenInterest(ctx context.Context, tokenID string) (*dataapi.OpenInterest, error) {
	return c.data.OpenInterest(ctx, tokenID)
}

// TraderLeaderboard returns the trader leaderboard.
func (c *Client) TraderLeaderboard(ctx context.Context, limit int) ([]dataapi.TraderLeaderboardEntry, error) {
	return c.data.TraderLeaderboard(ctx, limit)
}

// LiveVolume returns live volume data.
func (c *Client) LiveVolume(ctx context.Context, limit int) (*dataapi.LiveVolumeResponse, error) {
	return c.data.LiveVolume(ctx, limit)
}

// EnrichedMarkets returns active Gamma markets enriched with CLOB data.
func (c *Client) EnrichedMarkets(ctx context.Context, limit int) ([]polytypes.EnrichedMarket, error) {
	return c.discovery.EnrichedMarkets(ctx, limit)
}

// SearchAndEnrich searches Gamma and enriches matching markets with CLOB data.
func (c *Client) SearchAndEnrich(ctx context.Context, query string, limit int) ([]polytypes.EnrichedMarket, error) {
	return c.discovery.SearchAndEnrich(ctx, query, limit)
}

// EnrichMarket enriches a single Gamma market with CLOB data.
func (c *Client) EnrichMarket(ctx context.Context, market polytypes.Market) (*polytypes.EnrichedMarket, error) {
	return c.discovery.EnrichMarket(ctx, market)
}

// StreamClient returns a new WebSocket market stream client.
func (c *Client) StreamClient() *stream.MarketClient {
	return stream.NewMarketClient(stream.DefaultConfig(defaultStreamURL))
}

// StreamClientWithConfig returns a WebSocket market stream client with custom config.
func (c *Client) StreamClientWithConfig(cfg stream.Config) *stream.MarketClient {
	return stream.NewMarketClient(cfg)
}

// HealthCheck pings all three HTTP APIs and returns a summary.
// An error is returned only if all three APIs are unreachable.
func (c *Client) HealthCheck(ctx context.Context) (HealthResponse, error) {
	var resp HealthResponse
	var errs []error

	if _, err := c.gamma.HealthCheck(ctx); err != nil {
		errs = append(errs, fmt.Errorf("gamma: %w", err))
	} else {
		resp.GammaOK = true
	}

	if err := c.clob.Health(ctx); err != nil {
		errs = append(errs, fmt.Errorf("clob: %w", err))
	} else {
		resp.CLOBOK = true
	}

	if err := c.data.Health(ctx); err != nil {
		errs = append(errs, fmt.Errorf("data: %w", err))
	} else {
		resp.DataOK = true
	}

	if len(errs) == 3 {
		return resp, fmt.Errorf("all APIs unreachable: %v", errs)
	}
	return resp, nil
}

// HealthResponse reports the reachability of each Polymarket API.
type HealthResponse struct {
	GammaOK bool `json:"gamma_ok"`
	CLOBOK  bool `json:"clob_ok"`
	DataOK  bool `json:"data_ok"`
}

// DefaultTransportConfig returns the transport config the universal client
// uses by default.
func DefaultTransportConfig(baseURL string) transport.Config {
	return transport.DefaultConfig(baseURL)
}
