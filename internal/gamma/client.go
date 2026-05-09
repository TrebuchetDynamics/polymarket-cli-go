package gamma

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

// Client provides read-only access to the Polymarket Gamma API.
// Base URL: https://gamma-api.polymarket.com
type Client struct {
	transport *transport.Client
}

// NewClient creates a Gamma API client.
func NewClient(baseURL string, tc *transport.Client) *Client {
	if tc == nil {
		tc = transport.New(nil, transport.DefaultConfig(baseURL))
	}
	return &Client{transport: tc}
}

// HealthCheck verifies the Gamma API is reachable.
func (c *Client) HealthCheck(ctx context.Context) (*polytypes.HealthResponse, error) {
	if _, err := c.transport.GetRaw(ctx, "/"); err != nil {
		return nil, err
	}
	return &polytypes.HealthResponse{Data: "ok"}, nil
}

// ActiveMarkets returns active, non-closed markets.
func (c *Client) ActiveMarkets(ctx context.Context) ([]polytypes.Market, error) {
	active := true
	closed := false
	return c.Markets(ctx, &polytypes.GetMarketsParams{
		Active: &active,
		Closed: &closed,
	})
}

// Markets lists markets with optional filters.
func (c *Client) Markets(ctx context.Context, params *polytypes.GetMarketsParams) ([]polytypes.Market, error) {
	path, err := buildQueryPath("/markets", params)
	if err != nil {
		return nil, err
	}
	var result []polytypes.Market
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MarketByID returns a single market by Gamma ID.
func (c *Client) MarketByID(ctx context.Context, id string) (*polytypes.Market, error) {
	var result *polytypes.Market
	if err := c.transport.Get(ctx, "/markets/"+id, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MarketBySlug returns a single market by slug.
func (c *Client) MarketBySlug(ctx context.Context, slug string) (*polytypes.Market, error) {
	var result *polytypes.Market
	if err := c.transport.Get(ctx, "/markets/"+slug, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Events lists events with optional filters.
func (c *Client) Events(ctx context.Context, params *polytypes.GetEventsParams) ([]polytypes.Event, error) {
	path, err := buildQueryPath("/events", params)
	if err != nil {
		return nil, err
	}
	var result []polytypes.Event
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// EventByID returns a single event by ID.
func (c *Client) EventByID(ctx context.Context, id string) (*polytypes.Event, error) {
	var result *polytypes.Event
	if err := c.transport.Get(ctx, "/events/"+id, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// EventBySlug returns a single event by slug.
func (c *Client) EventBySlug(ctx context.Context, slug string) (*polytypes.Event, error) {
	events, err := c.Events(ctx, &polytypes.GetEventsParams{
		Limit: 1,
		Slug:  []string{slug},
	})
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("event slug %q not found", slug)
	}
	return &events[0], nil
}

// Series lists series with optional filters.
func (c *Client) Series(ctx context.Context, params *polytypes.GetSeriesParams) ([]polytypes.Series, error) {
	path, err := buildQueryPath("/series", params)
	if err != nil {
		return nil, err
	}
	var result []polytypes.Series
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SeriesByID returns a single series by ID.
func (c *Client) SeriesByID(ctx context.Context, id string) (*polytypes.Series, error) {
	var result *polytypes.Series
	if err := c.transport.Get(ctx, "/series/"+id, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Search performs a cross-entity search.
func (c *Client) Search(ctx context.Context, params *polytypes.SearchParams) (*polytypes.SearchResponse, error) {
	path, err := buildQueryPath("/public-search", params)
	if err != nil {
		return nil, err
	}
	var result polytypes.SearchResponse
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Tags lists tags with optional filters.
func (c *Client) Tags(ctx context.Context, params *polytypes.GetTagsParams) ([]polytypes.Tag, error) {
	path, err := buildQueryPath("/tags", params)
	if err != nil {
		return nil, err
	}
	var result []polytypes.Tag
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TagByID returns a single tag by ID.
func (c *Client) TagByID(ctx context.Context, id string) (*polytypes.Tag, error) {
	var result *polytypes.Tag
	if err := c.transport.Get(ctx, "/tags/"+id, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TagBySlug returns a single tag by slug.
func (c *Client) TagBySlug(ctx context.Context, slug string) (*polytypes.Tag, error) {
	var result *polytypes.Tag
	if err := c.transport.Get(ctx, "/tags/"+slug, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RelatedTagsByID returns related tags for a tag ID.
func (c *Client) RelatedTagsByID(ctx context.Context, tagID string) ([]polytypes.TagRelationship, error) {
	var result []polytypes.TagRelationship
	if err := c.transport.Get(ctx, "/tags/"+tagID+"/related", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RelatedTagsBySlug returns related tags for a tag slug.
func (c *Client) RelatedTagsBySlug(ctx context.Context, slug string) ([]polytypes.TagRelationship, error) {
	var result []polytypes.TagRelationship
	if err := c.transport.Get(ctx, "/tags/"+slug+"/related", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Teams lists sports teams with optional filters.
func (c *Client) Teams(ctx context.Context, params *polytypes.GetTeamsParams) ([]polytypes.Team, error) {
	path, err := buildQueryPath("/teams", params)
	if err != nil {
		return nil, err
	}
	var result []polytypes.Team
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SportsMetadata returns sports metadata.
func (c *Client) SportsMetadata(ctx context.Context) ([]polytypes.SportMetadata, error) {
	var result []polytypes.SportMetadata
	if err := c.transport.Get(ctx, "/sports-metadata", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// buildQueryPath serializes a params struct to URL query parameters using reflection.
// For simple cases, the caller builds the path directly.
// This version handles the most common field types.
func buildQueryPath(basePath string, params interface{}) (string, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid base path %q: %w", basePath, err)
	}

	q := u.Query()

	switch p := params.(type) {
	case *polytypes.GetMarketsParams:
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
		if p.Offset > 0 {
			q.Set("offset", strconv.Itoa(p.Offset))
		}
		if p.Closed != nil {
			q.Set("closed", strconv.FormatBool(*p.Closed))
		}
		if p.Active != nil {
			q.Set("active", strconv.FormatBool(*p.Active))
		}
		if p.TagID != nil {
			q.Set("tag_id", strconv.Itoa(*p.TagID))
		}
		if p.Order != "" {
			q.Set("order", p.Order)
		}
		if p.Ascending != nil {
			q.Set("ascending", strconv.FormatBool(*p.Ascending))
		}
		for _, s := range p.Slug {
			q.Add("slug", s)
		}
		for _, s := range p.ClobTokenIDs {
			q.Add("clob_token_ids", s)
		}
		for _, s := range p.ConditionIDs {
			q.Add("condition_ids", s)
		}
		if p.LiquidityNumMin != nil {
			q.Set("liquidity_num_min", strconv.FormatFloat(*p.LiquidityNumMin, 'f', -1, 64))
		}
		if p.LiquidityNumMax != nil {
			q.Set("liquidity_num_max", strconv.FormatFloat(*p.LiquidityNumMax, 'f', -1, 64))
		}
		if p.VolumeNumMin != nil {
			q.Set("volume_num_min", strconv.FormatFloat(*p.VolumeNumMin, 'f', -1, 64))
		}
		if p.VolumeNumMax != nil {
			q.Set("volume_num_max", strconv.FormatFloat(*p.VolumeNumMax, 'f', -1, 64))
		}
		for _, smt := range p.SportsMarketTypes {
			q.Add("sports_market_types", smt)
		}

	case *polytypes.GetEventsParams:
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
		if p.Offset > 0 {
			q.Set("offset", strconv.Itoa(p.Offset))
		}
		if p.Closed != nil {
			q.Set("closed", strconv.FormatBool(*p.Closed))
		}
		if p.TagID != nil {
			q.Set("tag_id", strconv.Itoa(*p.TagID))
		}
		if p.Order != "" {
			q.Set("order", p.Order)
		}
		if p.Ascending != nil {
			q.Set("ascending", strconv.FormatBool(*p.Ascending))
		}
		for _, s := range p.Slug {
			q.Add("slug", s)
		}

	case *polytypes.GetSeriesParams:
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
		if p.Offset > 0 {
			q.Set("offset", strconv.Itoa(p.Offset))
		}
		if p.Closed != nil {
			q.Set("closed", strconv.FormatBool(*p.Closed))
		}
		if p.Order != "" {
			q.Set("order", p.Order)
		}
		if p.Ascending != nil {
			q.Set("ascending", strconv.FormatBool(*p.Ascending))
		}

	case *polytypes.SearchParams:
		q.Set("q", p.Q)
		if p.LimitPerType != nil {
			q.Set("limit_per_type", strconv.Itoa(*p.LimitPerType))
		}
		if p.Page != nil {
			q.Set("page", strconv.Itoa(*p.Page))
		}
		if p.EventsStatus != "" {
			q.Set("events_status", p.EventsStatus)
		}
		if p.Ascending != nil {
			q.Set("ascending", strconv.FormatBool(*p.Ascending))
		}
		if p.Sort != "" {
			q.Set("sort", p.Sort)
		}
		for _, tag := range p.EventsTag {
			q.Add("events_tag", tag)
		}

	case *polytypes.GetTagsParams:
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
		if p.Offset > 0 {
			q.Set("offset", strconv.Itoa(p.Offset))
		}
		if p.Order != "" {
			q.Set("order", p.Order)
		}
		if p.Ascending != nil {
			q.Set("ascending", strconv.FormatBool(*p.Ascending))
		}

	case *polytypes.GetTeamsParams:
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
		if p.Offset > 0 {
			q.Set("offset", strconv.Itoa(p.Offset))
		}
		if p.Order != "" {
			q.Set("order", p.Order)
		}
		if p.Ascending != nil {
			q.Set("ascending", strconv.FormatBool(*p.Ascending))
		}
		for _, l := range p.League {
			q.Add("league", l)
		}
		for _, n := range p.Name {
			q.Add("name", n)
		}

	default:
		return basePath, nil
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// --- Comments ---

func (c *Client) Comments(ctx context.Context, params *polytypes.CommentQuery) ([]polytypes.Comment, error) {
	path, err := buildCommentPath("/comments", params)
	if err != nil {
		return nil, err
	}
	var result []polytypes.Comment
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CommentByID(ctx context.Context, id string) (*polytypes.Comment, error) {
	var result *polytypes.Comment
	if err := c.transport.Get(ctx, "/comments/"+id, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CommentsByUser(ctx context.Context, userAddress string, limit int) ([]polytypes.Comment, error) {
	path := fmt.Sprintf("/comments?user_address=%s&limit=%d", userAddress, limit)
	var result []polytypes.Comment
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Profiles ---

func (c *Client) PublicProfile(ctx context.Context, walletAddress string) (*polytypes.Profile, error) {
	var result *polytypes.Profile
	if err := c.transport.Get(ctx, "/profiles/"+walletAddress, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Sports ---

func (c *Client) SportsMarketTypes(ctx context.Context) ([]polytypes.SportsMarketType, error) {
	var result []polytypes.SportsMarketType
	if err := c.transport.Get(ctx, "/sports-market-types", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Market by Token ---

func (c *Client) MarketByToken(ctx context.Context, tokenID string) (*polytypes.MarketByTokenResponse, error) {
	var result *polytypes.MarketByTokenResponse
	if err := c.transport.Get(ctx, "/markets/token/"+tokenID, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Keyset Pagination ---

func (c *Client) EventsKeyset(ctx context.Context, params *polytypes.KeysetParams) ([]polytypes.Event, string, error) {
	path, err := buildKeysetPath("/events-keyset", params)
	if err != nil {
		return nil, "", err
	}
	var result polytypes.KeysetResponse[polytypes.Event]
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, "", err
	}
	return result.Data, result.NextCursor, nil
}

func (c *Client) MarketsKeyset(ctx context.Context, params *polytypes.KeysetParams) ([]polytypes.Market, string, error) {
	path, err := buildKeysetPath("/markets-keyset", params)
	if err != nil {
		return nil, "", err
	}
	var result polytypes.KeysetResponse[polytypes.Market]
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, "", err
	}
	return result.Data, result.NextCursor, nil
}

func buildCommentPath(basePath string, params *polytypes.CommentQuery) (string, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	if params.EntityID != nil {
		q.Set("parent_entity_id", strconv.Itoa(*params.EntityID))
	}
	if params.EntityType != nil {
		q.Set("parent_entity_type", *params.EntityType)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func buildKeysetPath(basePath string, params *polytypes.KeysetParams) (string, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.KeysetID != "" {
		q.Set("keyset_id", params.KeysetID)
	}
	if params.Ascending != nil {
		q.Set("ascending", strconv.FormatBool(*params.Ascending))
	}
	if params.Active != nil {
		q.Set("active", strconv.FormatBool(*params.Active))
	}
	if params.Closed != nil {
		q.Set("closed", strconv.FormatBool(*params.Closed))
	}
	if params.Order != "" {
		q.Set("order", params.Order)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
