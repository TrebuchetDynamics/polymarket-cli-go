package dataapi

import (
	"context"
	"net/url"
	"strconv"

	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

// Client provides read-only Data API access.
type Client struct {
	transport *transport.Client
}

func NewClient(baseURL string, tc *transport.Client) *Client {
	if tc == nil {
		tc = transport.New(nil, transport.DefaultConfig(baseURL))
	}
	return &Client{transport: tc}
}

// --- Types ---

type Position struct {
	TokenID       string  `json:"token_id"`
	ConditionID   string  `json:"condition_id"`
	MarketID      string  `json:"market_id"`
	Side          string  `json:"side"`
	AvgPrice      float64 `json:"avg_price"`
	Size          float64 `json:"size"`
	CurrentPrice  float64 `json:"current_price"`
	UnrealizedPnl float64 `json:"unrealized_pnl"`
}

type ClosedPosition struct {
	TokenID       string  `json:"token_id"`
	ConditionID   string  `json:"condition_id"`
	MarketID      string  `json:"market_id"`
	Side          string  `json:"side"`
	AvgPriceBuy   float64 `json:"avg_price_buy"`
	AvgPriceSell  float64 `json:"avg_price_sell"`
	Size          float64 `json:"size"`
	RealizedPnl   float64 `json:"realized_pnl"`
}

type Trade struct {
	ID        string  `json:"id"`
	Market    string  `json:"market"`
	AssetID   string  `json:"asset_id"`
	Side      string  `json:"side"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	FeeRateBps int     `json:"fee_rate_bps"`
	CreatedAt string  `json:"created_at"`
}

type Activity struct {
	Type      string `json:"type"`
	Market    string `json:"market"`
	AssetID   string `json:"asset_id"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Size      string `json:"size"`
	Timestamp string `json:"timestamp"`
}

type MetaHolder struct {
	Address string  `json:"address"`
	Shares  float64 `json:"shares"`
	Pnl     float64 `json:"pnl"`
	Volume  float64 `json:"volume"`
}

type TotalValue struct {
	User      string  `json:"user"`
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

type TotalMarketsTraded struct {
	User          string `json:"user"`
	MarketsTraded int    `json:"markets_traded"`
}

type OpenInterest struct {
	Market    string  `json:"market"`
	AssetID   string  `json:"asset_id"`
	OpenValue float64 `json:"open_value"`
}

type TraderLeaderboardEntry struct {
	Rank    int     `json:"rank"`
	User    string  `json:"user"`
	Volume  float64 `json:"volume"`
	Pnl     float64 `json:"pnl"`
	ROI     float64 `json:"roi"`
}

type LiveVolumeEntry struct {
	EventID   string  `json:"event_id"`
	EventSlug string  `json:"event_slug"`
	Title     string  `json:"title"`
	Volume    float64 `json:"volume"`
}

type LiveVolumeResponse struct {
	Total  int                `json:"total"`
	Events []LiveVolumeEntry  `json:"events"`
}

// --- Methods ---

func (c *Client) Health(ctx context.Context) error {
	return c.transport.Get(ctx, "/", nil)
}

func (c *Client) CurrentPositions(ctx context.Context, user string) ([]Position, error) {
	path := buildPath("/positions", map[string]string{"user": user})
	var result []Position
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ClosedPositions(ctx context.Context, user string) ([]ClosedPosition, error) {
	path := buildPath("/closed-positions", map[string]string{"user": user})
	var result []ClosedPosition
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Trades(ctx context.Context, user string, limit int) ([]Trade, error) {
	path := buildPath("/trades", map[string]string{
		"user":  user,
		"limit": strconv.Itoa(limit),
	})
	var result []Trade
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Activity(ctx context.Context, user string, limit int) ([]Activity, error) {
	path := buildPath("/activity", map[string]string{
		"user":  user,
		"limit": strconv.Itoa(limit),
	})
	var result []Activity
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) TopHolders(ctx context.Context, tokenID string, limit int) ([]MetaHolder, error) {
	path := buildPath("/top-holders", map[string]string{
		"token_id": tokenID,
		"limit":    strconv.Itoa(limit),
	})
	var result []MetaHolder
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) TotalValue(ctx context.Context, user string) (*TotalValue, error) {
	path := buildPath("/total-value", map[string]string{"user": user})
	var result TotalValue
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) MarketsTraded(ctx context.Context, user string) (*TotalMarketsTraded, error) {
	path := buildPath("/total-markets-traded", map[string]string{"user": user})
	var result TotalMarketsTraded
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) OpenInterest(ctx context.Context, tokenID string) (*OpenInterest, error) {
	path := buildPath("/open-interest", map[string]string{"token_id": tokenID})
	var result OpenInterest
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) TraderLeaderboard(ctx context.Context, limit int) ([]TraderLeaderboardEntry, error) {
	path := buildPath("/trader-leaderboard", map[string]string{"limit": strconv.Itoa(limit)})
	var result []TraderLeaderboardEntry
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) LiveVolume(ctx context.Context, limit int) (*LiveVolumeResponse, error) {
	path := buildPath("/live-volume", map[string]string{"limit": strconv.Itoa(limit)})
	var result LiveVolumeResponse
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func buildPath(base string, params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	if len(q) > 0 {
		return base + "?" + q.Encode()
	}
	return base
}
