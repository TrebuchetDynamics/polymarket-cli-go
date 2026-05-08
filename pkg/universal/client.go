// Package universal is a single-stop client for all Polymarket data and
// authenticated trading operations.
//
// Use universal when you need to query markets, order books, prices, events,
// volume, leaderboards, and streaming data without managing multiple API
// clients. The client also exposes the authenticated CLOB surface — API-key
// minting, balance/allowance, order placement and cancellation — so a Go
// SDK consumer can run the headless onboarding and trading flow described
// in docs/BUILDER-AUTO.md without dropping into internal packages.
//
// When not to use this package:
//   - For deposit wallet lifecycle (deploy / proxy / approvals) — that
//     surface is gated to builder credentials and lives in internal/relayer.
//
// Stability: Client, NewClient, DefaultConfig, and every method on Client
// are part of the polygolem public SDK and follow semver. Some method
// signatures still expose shared protocol types from internal packages;
// those types should be promoted or re-exported before the SDK is considered
// clean for external modules.
package universal

import (
	"context"
	"fmt"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/marketdiscovery"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/stream"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
	sdkdata "github.com/TrebuchetDynamics/polygolem/pkg/data"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
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
	data      *sdkdata.Client
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
	dc := sdkdata.NewClient(sdkdata.Config{BaseURL: cfg.DataBaseURL})

	return &Client{
		gamma:     gc,
		clob:      cc,
		data:      dc,
		discovery: marketdiscovery.New(gc, cc),
	}
}

// ActiveMarkets returns markets currently flagged active by Gamma.
func (c *Client) ActiveMarkets(ctx context.Context) ([]types.Market, error) {
	return c.gamma.ActiveMarkets(ctx)
}

// Markets lists markets matching the given filter parameters.
func (c *Client) Markets(ctx context.Context, params *types.GetMarketsParams) ([]types.Market, error) {
	return c.gamma.Markets(ctx, params)
}

// MarketByID fetches a single market by its Gamma ID.
func (c *Client) MarketByID(ctx context.Context, id string) (*types.Market, error) {
	return c.gamma.MarketByID(ctx, id)
}

// MarketBySlug fetches a single market by slug.
func (c *Client) MarketBySlug(ctx context.Context, slug string) (*types.Market, error) {
	return c.gamma.MarketBySlug(ctx, slug)
}

// Events lists events matching the given filter parameters.
func (c *Client) Events(ctx context.Context, params *types.GetEventsParams) ([]types.Event, error) {
	return c.gamma.Events(ctx, params)
}

// EventByID fetches a single event by its Gamma ID.
func (c *Client) EventByID(ctx context.Context, id string) (*types.Event, error) {
	return c.gamma.EventByID(ctx, id)
}

// EventBySlug fetches a single event by slug.
func (c *Client) EventBySlug(ctx context.Context, slug string) (*types.Event, error) {
	return c.gamma.EventBySlug(ctx, slug)
}

// Search performs Gamma's public search across events, markets, tags, and profiles.
func (c *Client) Search(ctx context.Context, params *types.SearchParams) (*types.SearchResponse, error) {
	return c.gamma.Search(ctx, params)
}

// Series lists market series matching the given filter parameters.
func (c *Client) Series(ctx context.Context, params *types.GetSeriesParams) ([]types.Series, error) {
	return c.gamma.Series(ctx, params)
}

// Tags lists tags matching the given filter parameters.
func (c *Client) Tags(ctx context.Context, params *types.GetTagsParams) ([]types.Tag, error) {
	return c.gamma.Tags(ctx, params)
}

// SportsMetadata returns the current sports metadata catalog.
func (c *Client) SportsMetadata(ctx context.Context) ([]types.SportMetadata, error) {
	return c.gamma.SportsMetadata(ctx)
}

// Comments returns comments matching the given query.
func (c *Client) Comments(ctx context.Context, params *types.CommentQuery) ([]types.Comment, error) {
	return c.gamma.Comments(ctx, params)
}

// MarketBySlug fetches a single market by slug.
func (c *Client) GammaMarketBySlug(ctx context.Context, slug string) (*types.Market, error) {
	return c.gamma.MarketBySlug(ctx, slug)
}

// EventBySlug fetches a single event by slug.
func (c *Client) GammaEventBySlug(ctx context.Context, slug string) (*types.Event, error) {
	return c.gamma.EventBySlug(ctx, slug)
}

// SeriesByID fetches a single series by its ID.
func (c *Client) SeriesByID(ctx context.Context, id string) (*types.Series, error) {
	return c.gamma.SeriesByID(ctx, id)
}

// TagByID fetches a single tag by its ID.
func (c *Client) TagByID(ctx context.Context, id string) (*types.Tag, error) {
	return c.gamma.TagByID(ctx, id)
}

// TagBySlug fetches a single tag by slug.
func (c *Client) TagBySlug(ctx context.Context, slug string) (*types.Tag, error) {
	return c.gamma.TagBySlug(ctx, slug)
}

// RelatedTagsByID returns tags related to the given tag ID.
func (c *Client) RelatedTagsByID(ctx context.Context, tagID string) ([]types.TagRelationship, error) {
	return c.gamma.RelatedTagsByID(ctx, tagID)
}

// RelatedTagsBySlug returns tags related to the given tag slug.
func (c *Client) RelatedTagsBySlug(ctx context.Context, slug string) ([]types.TagRelationship, error) {
	return c.gamma.RelatedTagsBySlug(ctx, slug)
}

// Teams lists sports teams matching the given filter parameters.
func (c *Client) Teams(ctx context.Context, params *types.GetTeamsParams) ([]types.Team, error) {
	return c.gamma.Teams(ctx, params)
}

// CommentByID fetches a single comment by its ID.
func (c *Client) CommentByID(ctx context.Context, id string) (*types.Comment, error) {
	return c.gamma.CommentByID(ctx, id)
}

// CommentsByUser fetches comments by a specific user address.
func (c *Client) CommentsByUser(ctx context.Context, userAddress string, limit int) ([]types.Comment, error) {
	return c.gamma.CommentsByUser(ctx, userAddress, limit)
}

// PublicProfile fetches a public profile by wallet address.
func (c *Client) PublicProfile(ctx context.Context, walletAddress string) (*types.Profile, error) {
	return c.gamma.PublicProfile(ctx, walletAddress)
}

// SportsMarketTypes returns the current sports market types catalog.
func (c *Client) SportsMarketTypes(ctx context.Context) ([]types.SportsMarketType, error) {
	return c.gamma.SportsMarketTypes(ctx)
}

// MarketByToken fetches market metadata by CLOB token ID.
func (c *Client) MarketByToken(ctx context.Context, tokenID string) (*types.MarketByTokenResponse, error) {
	return c.gamma.MarketByToken(ctx, tokenID)
}

// EventsKeyset returns events with keyset pagination.
func (c *Client) EventsKeyset(ctx context.Context, params *types.KeysetParams) ([]types.Event, string, error) {
	return c.gamma.EventsKeyset(ctx, params)
}

// MarketsKeyset returns markets with keyset pagination.
func (c *Client) MarketsKeyset(ctx context.Context, params *types.KeysetParams) ([]types.Market, string, error) {
	return c.gamma.MarketsKeyset(ctx, params)
}

// --- CLOB: Authenticated Orders & Cancellation ---

// ListOrders returns the authenticated user's open orders.
func (c *Client) ListOrders(ctx context.Context, privateKey string) ([]clob.OrderRecord, error) {
	return c.clob.ListOrders(ctx, privateKey)
}

// Order returns one authenticated order by order ID.
func (c *Client) Order(ctx context.Context, privateKey, orderID string) (*clob.OrderRecord, error) {
	return c.clob.Order(ctx, privateKey, orderID)
}

// ListTrades returns the authenticated user's trade history.
func (c *Client) ListTrades(ctx context.Context, privateKey string) ([]clob.TradeRecord, error) {
	return c.clob.ListTrades(ctx, privateKey)
}

// CancelOrder cancels a single open CLOB order.
func (c *Client) CancelOrder(ctx context.Context, privateKey, orderID string) (*clob.CancelOrdersResponse, error) {
	return c.clob.CancelOrder(ctx, privateKey, orderID)
}

// CancelOrders cancels multiple open CLOB orders by order ID.
func (c *Client) CancelOrders(ctx context.Context, privateKey string, orderIDs []string) (*clob.CancelOrdersResponse, error) {
	return c.clob.CancelOrders(ctx, privateKey, orderIDs)
}

// CancelAll cancels all open CLOB orders for the authenticated user.
func (c *Client) CancelAll(ctx context.Context, privateKey string) (*clob.CancelOrdersResponse, error) {
	return c.clob.CancelAll(ctx, privateKey)
}

// CancelMarket cancels open CLOB orders matching a market or asset filter.
func (c *Client) CancelMarket(ctx context.Context, privateKey string, params clob.CancelMarketParams) (*clob.CancelOrdersResponse, error) {
	return c.clob.CancelMarket(ctx, privateKey, params)
}

// --- CLOB: Headless Onboarding & Trading (L1/L2 auth) ---

// CreateOrDeriveAPIKey signs the canonical ClobAuth EIP-712 payload with
// privateKey, posts it to /auth/api-key, and falls back to the deterministic
// /auth/derive-api-key on conflict. First call for a new EOA lazy-creates
// the account, builder profile, and bytes32 builder code — see
// docs/BUILDER-AUTO.md for the empirical flow.
func (c *Client) CreateOrDeriveAPIKey(ctx context.Context, privateKey string) (auth.APIKey, error) {
	return c.clob.CreateOrDeriveAPIKey(ctx, privateKey)
}

// CreateBuilderFeeKey mints a builder fee key via L2 auth.
// Used for V2 order builder field attribution. Requires existing L2 credentials.
// Fully headless — no cookie, no browser, no relayer.
func (c *Client) CreateBuilderFeeKey(ctx context.Context, privateKey string) (auth.APIKey, error) {
	return c.clob.CreateBuilderFeeKey(ctx, privateKey)
}

// DeriveAPIKey returns the deterministic L2 credentials for an existing
// account via GET /auth/derive-api-key. Use CreateOrDeriveAPIKey when the
// caller is unsure whether an account has been provisioned yet.
func (c *Client) DeriveAPIKey(ctx context.Context, privateKey string) (auth.APIKey, error) {
	return c.clob.DeriveAPIKey(ctx, privateKey)
}

// BalanceAllowance returns the authenticated CLOB collateral or conditional
// token balance plus allowances against the V2 exchange spenders.
func (c *Client) BalanceAllowance(ctx context.Context, privateKey string, params clob.BalanceAllowanceParams) (*clob.BalanceAllowanceResponse, error) {
	return c.clob.BalanceAllowance(ctx, privateKey, params)
}

// UpdateBalanceAllowance forces the CLOB to refresh its on-chain
// balance/allowance cache for the authenticated user.
func (c *Client) UpdateBalanceAllowance(ctx context.Context, privateKey string, params clob.BalanceAllowanceParams) (*clob.BalanceAllowanceResponse, error) {
	return c.clob.UpdateBalanceAllowance(ctx, privateKey, params)
}

// CreateLimitOrder signs and submits a V2 limit order. The privateKey signs
// both the ClobAuth (for the API-key derivation) and the EIP-712 order
// itself. Returns the placement response with order id and matched amounts.
func (c *Client) CreateLimitOrder(ctx context.Context, privateKey string, params clob.CreateOrderParams) (*clob.OrderPlacementResponse, error) {
	return c.clob.CreateLimitOrder(ctx, privateKey, params)
}

// CreateMarketOrder signs and submits a V2 market order. Use Amount instead
// of Size on the params to express a fill-this-much budget.
func (c *Client) CreateMarketOrder(ctx context.Context, privateKey string, params clob.MarketOrderParams) (*clob.OrderPlacementResponse, error) {
	return c.clob.CreateMarketOrder(ctx, privateKey, params)
}

// --- CLOB: Metadata & Scoring ---

// CLOBServerTime returns the CLOB's current server time. Useful for
// signing payloads that embed a timestamp the backend must accept.
func (c *Client) CLOBServerTime(ctx context.Context) (*polytypes.ServerTime, error) {
	return c.clob.ServerTime(ctx)
}

// OrderScoring reports whether a single order id is currently scoring
// (eligible for liquidity rewards).
func (c *Client) OrderScoring(ctx context.Context, orderID string) (bool, error) {
	return c.clob.OrderScoring(ctx, orderID)
}

// OrdersScoring reports the scoring eligibility for a batch of order ids.
// Returns one boolean per id, in the order supplied.
func (c *Client) OrdersScoring(ctx context.Context, orderIDs []string) ([]bool, error) {
	return c.clob.OrdersScoring(ctx, orderIDs)
}

// --- CLOB: Rewards ---

// RewardsConfig returns the current liquidity-reward configuration entries.
func (c *Client) RewardsConfig(ctx context.Context) ([]polytypes.RewardsConfig, error) {
	return c.clob.RewardsConfig(ctx)
}

// RawRewards returns the per-maker raw reward breakdown for a market.
func (c *Client) RawRewards(ctx context.Context, market string) ([]polytypes.RawRewards, error) {
	return c.clob.RawRewards(ctx, market)
}

// UserEarnings returns the authenticated-day earnings record for date
// (YYYY-MM-DD).
func (c *Client) UserEarnings(ctx context.Context, date string) ([]polytypes.UserEarnings, error) {
	return c.clob.UserEarnings(ctx, date)
}

// TotalEarnings returns aggregate platform earnings for date (YYYY-MM-DD).
func (c *Client) TotalEarnings(ctx context.Context, date string) (*polytypes.TotalEarnings, error) {
	return c.clob.TotalEarnings(ctx, date)
}

// RewardPercentages returns the current reward-share percentages.
func (c *Client) RewardPercentages(ctx context.Context) ([]polytypes.RewardPercentages, error) {
	return c.clob.RewardPercentages(ctx)
}

// UserRewardsByMarket returns the authenticated user's rewards grouped by
// market.
func (c *Client) UserRewardsByMarket(ctx context.Context, params *polytypes.UserRewardsByMarketRequest) ([]polytypes.UserRewardsMarket, error) {
	return c.clob.UserRewardsByMarket(ctx, params)
}

// RebatedFees returns the rebated-fee schedule.
func (c *Client) RebatedFees(ctx context.Context) ([]polytypes.RebatedFees, error) {
	return c.clob.RebatedFees(ctx)
}

// --- CLOB: Extended Market Lists ---

// SimplifiedMarkets returns simplified CLOB markets.
func (c *Client) SimplifiedMarkets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	return c.clob.SimplifiedMarkets(ctx, nextCursor)
}

// SamplingMarkets returns sampling CLOB markets.
func (c *Client) SamplingMarkets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	return c.clob.SamplingMarkets(ctx, nextCursor)
}

// SamplingSimplifiedMarkets returns sampling simplified CLOB markets.
func (c *Client) SamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	return c.clob.SamplingSimplifiedMarkets(ctx, nextCursor)
}

// --- Data API: Extended ---

// ClosedPositions returns closed positions for a user.
func (c *Client) ClosedPositions(ctx context.Context, user string) ([]types.ClosedPosition, error) {
	return c.data.ClosedPositions(ctx, user)
}

// ClosedPositionsWithLimit returns closed positions for a user with a row limit.
func (c *Client) ClosedPositionsWithLimit(ctx context.Context, user string, limit int) ([]types.ClosedPosition, error) {
	return c.data.ClosedPositionsWithLimit(ctx, user, limit)
}

// MarketsTraded returns the count of markets traded by a user.
func (c *Client) MarketsTraded(ctx context.Context, user string) (*types.TotalMarketsTraded, error) {
	return c.data.MarketsTraded(ctx, user)
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
func (c *Client) CurrentPositions(ctx context.Context, user string) ([]types.Position, error) {
	return c.data.CurrentPositions(ctx, user)
}

// CurrentPositionsWithLimit returns current open positions for a user with a row limit.
func (c *Client) CurrentPositionsWithLimit(ctx context.Context, user string, limit int) ([]types.Position, error) {
	return c.data.CurrentPositionsWithLimit(ctx, user, limit)
}

// Trades returns trades for a user.
func (c *Client) Trades(ctx context.Context, user string, limit int) ([]types.Trade, error) {
	return c.data.Trades(ctx, user, limit)
}

// Activity returns recent activity for a user.
func (c *Client) Activity(ctx context.Context, user string, limit int) ([]types.Activity, error) {
	return c.data.Activity(ctx, user, limit)
}

// TopHolders returns top holders for a token.
func (c *Client) TopHolders(ctx context.Context, tokenID string, limit int) ([]types.Holder, error) {
	return c.data.TopHolders(ctx, tokenID, limit)
}

// TotalValue returns total portfolio value for a user.
func (c *Client) TotalValue(ctx context.Context, user string) (*types.PortfolioValue, error) {
	return c.data.TotalValue(ctx, user)
}

// OpenInterest returns open interest for a token.
func (c *Client) OpenInterest(ctx context.Context, tokenID string) (*types.OpenInterest, error) {
	return c.data.OpenInterest(ctx, tokenID)
}

// TraderLeaderboard returns the trader leaderboard.
func (c *Client) TraderLeaderboard(ctx context.Context, limit int) ([]types.LeaderboardRow, error) {
	return c.data.TraderLeaderboard(ctx, limit)
}

// LiveVolume returns live volume data.
func (c *Client) LiveVolume(ctx context.Context, limit int) (*types.LiveVolumeResponse, error) {
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
func (c *Client) EnrichMarket(ctx context.Context, market types.Market) (*polytypes.EnrichedMarket, error) {
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
