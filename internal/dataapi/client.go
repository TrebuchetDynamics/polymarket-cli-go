package dataapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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
	TokenID      string  `json:"token_id"`
	ConditionID  string  `json:"condition_id"`
	MarketID     string  `json:"market_id"`
	Side         string  `json:"side"`
	AvgPriceBuy  float64 `json:"avg_price_buy"`
	AvgPriceSell float64 `json:"avg_price_sell"`
	Size         float64 `json:"size"`
	RealizedPnl  float64 `json:"realized_pnl"`
}

type Trade struct {
	ID         string  `json:"id"`
	Market     string  `json:"market"`
	AssetID    string  `json:"asset_id"`
	Side       string  `json:"side"`
	Price      float64 `json:"price"`
	Size       float64 `json:"size"`
	FeeRateBps int     `json:"fee_rate_bps"`
	CreatedAt  string  `json:"created_at"`
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

func (a *Activity) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type      string          `json:"type"`
		Market    string          `json:"market"`
		AssetID   string          `json:"asset_id"`
		Side      string          `json:"side"`
		Price     json.RawMessage `json:"price"`
		Size      json.RawMessage `json:"size"`
		Timestamp json.RawMessage `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	a.Type = aux.Type
	a.Market = aux.Market
	a.AssetID = aux.AssetID
	a.Side = aux.Side
	a.Price = jsonStringOrNumber(aux.Price)
	a.Size = jsonStringOrNumber(aux.Size)
	a.Timestamp = jsonStringOrNumber(aux.Timestamp)
	return nil
}

type MetaHolder struct {
	Address     string  `json:"address"`
	ProxyWallet string  `json:"proxyWallet"`
	Shares      float64 `json:"shares"`
	Amount      float64 `json:"amount"`
	Pnl         float64 `json:"pnl"`
	Volume      float64 `json:"volume"`
}

type holdersByToken struct {
	Token   string       `json:"token"`
	Holders []MetaHolder `json:"holders"`
}

type TotalValue struct {
	User      string  `json:"user"`
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

type TotalMarketsTraded struct {
	User          string `json:"user"`
	MarketsTraded int    `json:"markets_traded"`
	Traded        int    `json:"traded,omitempty"`
}

type OpenInterest struct {
	Market    string  `json:"market"`
	AssetID   string  `json:"asset_id,omitempty"`
	OpenValue float64 `json:"value"`
}

type TraderLeaderboardEntry struct {
	Rank   int     `json:"rank"`
	User   string  `json:"user"`
	Volume float64 `json:"volume"`
	Pnl    float64 `json:"pnl"`
	ROI    float64 `json:"roi"`
}

type LiveVolumeEntry struct {
	EventID   string  `json:"event_id"`
	EventSlug string  `json:"event_slug"`
	Title     string  `json:"title"`
	Volume    float64 `json:"volume"`
}

type LiveVolumeMarket struct {
	Market string  `json:"market"`
	Value  float64 `json:"value"`
}

type LiveVolumeResponse struct {
	Total   float64            `json:"total"`
	Markets []LiveVolumeMarket `json:"markets,omitempty"`
	Events  []LiveVolumeEntry  `json:"events,omitempty"`
}

// --- Methods ---

func (c *Client) Health(ctx context.Context) error {
	return c.transport.Get(ctx, "/", nil)
}

func (c *Client) CurrentPositions(ctx context.Context, user string) ([]Position, error) {
	return c.CurrentPositionsWithLimit(ctx, user, 0)
}

func (c *Client) CurrentPositionsWithLimit(ctx context.Context, user string, limit int) ([]Position, error) {
	params := map[string]string{"user": user}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	path := buildPath("/positions", params)
	var result []Position
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ClosedPositions(ctx context.Context, user string) ([]ClosedPosition, error) {
	return c.ClosedPositionsWithLimit(ctx, user, 0)
}

func (c *Client) ClosedPositionsWithLimit(ctx context.Context, user string, limit int) ([]ClosedPosition, error) {
	params := map[string]string{"user": user}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	path := buildPath("/closed-positions", params)
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

func (c *Client) TopHolders(ctx context.Context, market string, limit int) ([]MetaHolder, error) {
	path := buildPath("/holders", map[string]string{
		"market": market,
		"limit":  strconv.Itoa(limit),
	})
	var result []holdersByToken
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	var holders []MetaHolder
	for _, token := range result {
		for _, holder := range token.Holders {
			if holder.Address == "" {
				holder.Address = holder.ProxyWallet
			}
			if holder.Shares == 0 {
				holder.Shares = holder.Amount
			}
			holders = append(holders, holder)
		}
	}
	return holders, nil
}

func (c *Client) TotalValue(ctx context.Context, user string) (*TotalValue, error) {
	path := buildPath("/value", map[string]string{"user": user})
	raw, err := c.transport.GetRaw(ctx, path)
	if err != nil {
		return nil, err
	}
	var result TotalValue
	if err := json.Unmarshal(raw, &result); err == nil {
		return &result, nil
	}
	var rows []TotalValue
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &TotalValue{User: user}, nil
	}
	result = rows[0]
	return &result, nil
}

func (c *Client) MarketsTraded(ctx context.Context, user string) (*TotalMarketsTraded, error) {
	path := buildPath("/traded", map[string]string{"user": user})
	var result TotalMarketsTraded
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.MarketsTraded == 0 {
		result.MarketsTraded = result.Traded
	}
	return &result, nil
}

func (c *Client) OpenInterest(ctx context.Context, market string) (*OpenInterest, error) {
	path := buildPath("/oi", map[string]string{"market": market})
	var result []OpenInterest
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return &OpenInterest{Market: market}, nil
	}
	return &result[0], nil
}

func (c *Client) TraderLeaderboard(ctx context.Context, limit int) ([]TraderLeaderboardEntry, error) {
	path := buildPath("/v1/leaderboard", map[string]string{"limit": strconv.Itoa(limit)})
	raw, err := c.transport.GetRaw(ctx, path)
	if err != nil {
		return nil, err
	}
	var rows []leaderboardWire
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	out := make([]TraderLeaderboardEntry, len(rows))
	for i, row := range rows {
		rank, err := parseLeaderboardRank(row.Rank)
		if err != nil {
			return nil, err
		}
		volume := row.Volume
		if volume == 0 {
			volume = row.Vol
		}
		out[i] = TraderLeaderboardEntry{
			Rank:   rank,
			User:   firstNonEmpty(row.User, row.ProxyWallet, row.UserName),
			Volume: volume,
			Pnl:    row.Pnl,
			ROI:    row.ROI,
		}
	}
	return out, nil
}

type leaderboardWire struct {
	Rank        json.RawMessage `json:"rank"`
	User        string          `json:"user"`
	ProxyWallet string          `json:"proxyWallet"`
	UserName    string          `json:"userName"`
	Volume      float64         `json:"volume"`
	Vol         float64         `json:"vol"`
	Pnl         float64         `json:"pnl"`
	ROI         float64         `json:"roi"`
}

func parseLeaderboardRank(raw json.RawMessage) (int, error) {
	if len(raw) == 0 {
		return 0, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, fmt.Errorf("decode leaderboard rank: %w", err)
	}
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("decode leaderboard rank %q: %w", s, err)
	}
	return n, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func jsonStringOrNumber(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return ""
	}
	if s[0] == '"' {
		var v string
		if err := json.Unmarshal(raw, &v); err == nil {
			return v
		}
	}
	return s
}

func (c *Client) LiveVolume(ctx context.Context, eventID int) (*LiveVolumeResponse, error) {
	path := buildPath("/live-volume", map[string]string{"id": strconv.Itoa(eventID)})
	raw, err := c.transport.GetRaw(ctx, path)
	if err != nil {
		return nil, err
	}
	var result LiveVolumeResponse
	if err := json.Unmarshal(raw, &result); err == nil {
		return &result, nil
	}
	var rows []LiveVolumeResponse
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &LiveVolumeResponse{}, nil
	}
	result = rows[0]
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
