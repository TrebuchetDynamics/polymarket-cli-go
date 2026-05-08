// Package gamma is a read-only client for the Polymarket Gamma API
// surfaced for embedded use by downstream Go consumers.
//
// Use gamma when you need typed access to Polymarket markets, events,
// search, tags, series, sports metadata, or comments without pulling in
// the full polygolem CLI. The client performs no signing and is safe in
// read-only contexts.
//
// When not to use this package:
//   - For order book reads — use pkg/bookreader.
//   - For order placement or cancellation — Gamma does not host the
//     mutating CLOB surface.
//
// Stability: Client, NewClient, DefaultConfig, and every method on Client
// are part of the polygolem public SDK and follow semver. Response and query
// DTOs are exported from pkg/types.
package gamma

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

// Client is the public read-only Gamma API client.
// Construct via NewClient. Methods are safe for concurrent use.
type Client struct {
	inner *gamma.Client
}

// NewClient returns a Gamma client targeting baseURL.
// If baseURL is empty, the production Gamma URL is used. The client uses
// the package default HTTP transport with retry and rate limiting.
func NewClient(baseURL string) *Client {
	return &Client{inner: gamma.NewClient(baseURL, nil)}
}

// HealthCheck pings the Gamma /health endpoint and returns the parsed
// response. Use this for readiness probes; it does not validate auth.
func (c *Client) HealthCheck(ctx context.Context) (*types.HealthResponse, error) {
	return c.inner.HealthCheck(ctx)
}

// ActiveMarkets returns markets currently flagged active by Gamma.
// Equivalent to Markets with the active filter set.
func (c *Client) ActiveMarkets(ctx context.Context) ([]types.Market, error) {
	return c.inner.ActiveMarkets(ctx)
}

// Markets lists markets matching the given filter parameters.
// Pass nil for default behavior (server-defined defaults).
func (c *Client) Markets(ctx context.Context, params *types.GetMarketsParams) ([]types.Market, error) {
	return c.inner.Markets(ctx, params)
}

// MarketByID fetches a single market by its Gamma ID.
func (c *Client) MarketByID(ctx context.Context, id string) (*types.Market, error) {
	return c.inner.MarketByID(ctx, id)
}

// Events lists events matching the given filter parameters.
// Pass nil for default behavior.
func (c *Client) Events(ctx context.Context, params *types.GetEventsParams) ([]types.Event, error) {
	return c.inner.Events(ctx, params)
}

// EventByID fetches a single event by its Gamma ID.
func (c *Client) EventByID(ctx context.Context, id string) (*types.Event, error) {
	return c.inner.EventByID(ctx, id)
}

// Series lists market series matching the given filter parameters.
// Pass nil for default behavior.
func (c *Client) Series(ctx context.Context, params *types.GetSeriesParams) ([]types.Series, error) {
	return c.inner.Series(ctx, params)
}

// Search performs Gamma's public search across events, markets, tags, and
// profiles. Pass non-nil params; an empty Q returns server defaults.
func (c *Client) Search(ctx context.Context, params *types.SearchParams) (*types.SearchResponse, error) {
	return c.inner.Search(ctx, params)
}

// Tags lists tags matching the given filter parameters.
// Pass nil for default behavior.
func (c *Client) Tags(ctx context.Context, params *types.GetTagsParams) ([]types.Tag, error) {
	return c.inner.Tags(ctx, params)
}

// SportsMetadata returns the current sports metadata catalog used by
// sports-event markets.
func (c *Client) SportsMetadata(ctx context.Context) ([]types.SportMetadata, error) {
	return c.inner.SportsMetadata(ctx)
}

// Comments returns comments matching the given query — by parent entity
// or author, with optional pagination via params.
func (c *Client) Comments(ctx context.Context, params *types.CommentQuery) ([]types.Comment, error) {
	return c.inner.Comments(ctx, params)
}

// MarketBySlug fetches a single market by its Gamma slug.
func (c *Client) MarketBySlug(ctx context.Context, slug string) (*types.Market, error) {
	return c.inner.MarketBySlug(ctx, slug)
}

// EventBySlug fetches a single event by slug.
func (c *Client) EventBySlug(ctx context.Context, slug string) (*types.Event, error) {
	return c.inner.EventBySlug(ctx, slug)
}

// SeriesByID fetches a single series by its Gamma ID.
func (c *Client) SeriesByID(ctx context.Context, id string) (*types.Series, error) {
	return c.inner.SeriesByID(ctx, id)
}

// TagByID fetches a single tag by its Gamma ID.
func (c *Client) TagByID(ctx context.Context, id string) (*types.Tag, error) {
	return c.inner.TagByID(ctx, id)
}

// TagBySlug fetches a single tag by slug.
func (c *Client) TagBySlug(ctx context.Context, slug string) (*types.Tag, error) {
	return c.inner.TagBySlug(ctx, slug)
}

// RelatedTagsByID returns tags related to the given tag ID.
func (c *Client) RelatedTagsByID(ctx context.Context, tagID string) ([]types.TagRelationship, error) {
	return c.inner.RelatedTagsByID(ctx, tagID)
}

// RelatedTagsBySlug returns tags related to the given tag slug.
func (c *Client) RelatedTagsBySlug(ctx context.Context, slug string) ([]types.TagRelationship, error) {
	return c.inner.RelatedTagsBySlug(ctx, slug)
}

// Teams lists sports teams matching the given filter parameters.
func (c *Client) Teams(ctx context.Context, params *types.GetTeamsParams) ([]types.Team, error) {
	return c.inner.Teams(ctx, params)
}

// CommentByID fetches a single comment by its Gamma ID.
func (c *Client) CommentByID(ctx context.Context, id string) (*types.Comment, error) {
	return c.inner.CommentByID(ctx, id)
}

// CommentsByUser fetches comments by a specific user address.
func (c *Client) CommentsByUser(ctx context.Context, userAddress string, limit int) ([]types.Comment, error) {
	return c.inner.CommentsByUser(ctx, userAddress, limit)
}

// PublicProfile fetches a public profile by wallet address.
func (c *Client) PublicProfile(ctx context.Context, walletAddress string) (*types.Profile, error) {
	return c.inner.PublicProfile(ctx, walletAddress)
}

// SportsMarketTypes returns the current sports market types catalog.
func (c *Client) SportsMarketTypes(ctx context.Context) ([]types.SportsMarketType, error) {
	return c.inner.SportsMarketTypes(ctx)
}

// MarketByToken fetches market metadata by CLOB token ID.
func (c *Client) MarketByToken(ctx context.Context, tokenID string) (*types.MarketByTokenResponse, error) {
	return c.inner.MarketByToken(ctx, tokenID)
}

// EventsKeyset returns events with keyset pagination. Returns the next cursor
// as the second return value.
func (c *Client) EventsKeyset(ctx context.Context, params *types.KeysetParams) ([]types.Event, string, error) {
	return c.inner.EventsKeyset(ctx, params)
}

// MarketsKeyset returns markets with keyset pagination. Returns the next cursor
// as the second return value.
func (c *Client) MarketsKeyset(ctx context.Context, params *types.KeysetParams) ([]types.Market, string, error) {
	return c.inner.MarketsKeyset(ctx, params)
}

// DefaultConfig returns the transport config the Gamma client uses by
// default for baseURL — exposed for callers that want to inspect or
// extend the retry, timeout, and rate-limit defaults.
func DefaultConfig(baseURL string) transport.Config {
	return transport.DefaultConfig(baseURL)
}
