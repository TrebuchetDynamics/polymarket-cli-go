package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

// Client provides read-only access to the Polymarket CLOB API.
// Base URL: https://clob.polymarket.com
type Client struct {
	transport *transport.Client
}

// NewClient creates a CLOB API client.
func NewClient(baseURL string, tc *transport.Client) *Client {
	if tc == nil {
		tc = transport.New(nil, transport.DefaultConfig(baseURL))
	}
	return &Client{transport: tc}
}

// Health checks the CLOB API is reachable.
func (c *Client) Health(ctx context.Context) error {
	return c.transport.Get(ctx, "/", nil)
}

// ServerTime returns the current server time.
func (c *Client) ServerTime(ctx context.Context) (*polytypes.ServerTime, error) {
	var result polytypes.ServerTime
	if err := c.transport.Get(ctx, "/time", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Markets lists CLOB markets with cursor pagination.
func (c *Client) Markets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	path := "/markets"
	if nextCursor != "" {
		path += "?next_cursor=" + url.QueryEscape(nextCursor)
	}
	var result polytypes.CLOBPaginatedMarkets
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Market returns a single CLOB market by condition ID.
func (c *Client) Market(ctx context.Context, conditionID string) (*polytypes.CLOBMarket, error) {
	var result polytypes.CLOBMarket
	if err := c.transport.Get(ctx, "/markets/"+conditionID, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OrderBook returns L2 order book depth for a token.
func (c *Client) OrderBook(ctx context.Context, tokenID string) (*polytypes.OrderBook, error) {
	path := fmt.Sprintf("/book?token_id=%s", url.QueryEscape(tokenID))
	var result polytypes.OrderBook
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OrderBooks returns order books for multiple tokens (POST).
func (c *Client) OrderBooks(ctx context.Context, params []polytypes.BookParams) ([]polytypes.OrderBook, error) {
	var result []polytypes.OrderBook
	if err := c.transport.Post(ctx, "/books", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Price returns the best bid/ask for a token.
func (c *Client) Price(ctx context.Context, tokenID, side string) (string, error) {
	path := fmt.Sprintf("/price?token_id=%s&side=%s", url.QueryEscape(tokenID), url.QueryEscape(side))
	var wrapper struct {
		Price string `json:"price"`
	}
	if err := c.transport.Get(ctx, path, &wrapper); err != nil {
		return "", err
	}
	return wrapper.Price, nil
}

// Prices returns prices for multiple tokens (POST).
func (c *Client) Prices(ctx context.Context, params []polytypes.BookParams) (map[string]string, error) {
	var raw map[string]json.RawMessage
	if err := c.transport.Post(ctx, "/prices-post", params, &raw); err != nil {
		// Try legacy endpoint
		if err2 := c.transport.Post(ctx, "/prices", params, &raw); err2 != nil {
			return nil, fmt.Errorf("prices: %w (fallback also failed: %v)", err, err2)
		}
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		var wrapper struct {
			Price string `json:"price"`
		}
		if err := json.Unmarshal(v, &wrapper); err != nil {
			result[k] = string(v)
		} else {
			result[k] = wrapper.Price
		}
	}
	return result, nil
}

// Midpoint returns the midpoint price for a token.
func (c *Client) Midpoint(ctx context.Context, tokenID string) (string, error) {
	path := fmt.Sprintf("/midpoint?token_id=%s", url.QueryEscape(tokenID))
	var wrapper struct {
		Mid string `json:"mid"`
	}
	if err := c.transport.Get(ctx, path, &wrapper); err != nil {
		return "", err
	}
	return wrapper.Mid, nil
}

// Midpoints returns midpoints for multiple tokens (POST).
func (c *Client) Midpoints(ctx context.Context, params []polytypes.BookParams) (map[string]string, error) {
	var raw map[string]json.RawMessage
	if err := c.transport.Post(ctx, "/midpoints", params, &raw); err != nil {
		return nil, err
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		var wrapper struct {
			Mid string `json:"mid"`
		}
		if err := json.Unmarshal(v, &wrapper); err != nil {
			result[k] = string(v)
		} else {
			result[k] = wrapper.Mid
		}
	}
	return result, nil
}

// Spread returns the spread for a token.
func (c *Client) Spread(ctx context.Context, tokenID string) (string, error) {
	path := fmt.Sprintf("/spread?token_id=%s", url.QueryEscape(tokenID))
	var wrapper struct {
		Spread string `json:"spread"`
	}
	if err := c.transport.Get(ctx, path, &wrapper); err != nil {
		return "", err
	}
	return wrapper.Spread, nil
}

// TickSize returns the tick size for a token.
func (c *Client) TickSize(ctx context.Context, tokenID string) (*polytypes.TickSize, error) {
	path := fmt.Sprintf("/tick-size?token_id=%s", url.QueryEscape(tokenID))
	// The CLOB v2 API wraps tick size responses differently.
	// Try the standard structure first, then fall back.
	var result polytypes.TickSize
	if err := c.transport.Get(ctx, path, &result); err != nil {
		// Fallback: some endpoints return minimum_tick_size as top-level
		var raw struct {
			MinimumTickSize  string `json:"minimum_tick_size"`
			MinimumOrderSize string `json:"minimum_order_size"`
			TickSize         string `json:"tick_size"`
		}
		if err2 := c.transport.Get(ctx, path, &raw); err2 != nil {
			return nil, fmt.Errorf("tick-size: %w", err)
		}
		result = polytypes.TickSize{
			MinimumTickSize:  raw.MinimumTickSize,
			MinimumOrderSize: raw.MinimumOrderSize,
			TickSize:         raw.TickSize,
		}
	}
	return &result, nil
}

// NegRisk returns neg risk info for a token.
func (c *Client) NegRisk(ctx context.Context, tokenID string) (*polytypes.NegRiskInfo, error) {
	path := fmt.Sprintf("/neg-risk?token_id=%s", url.QueryEscape(tokenID))
	var result polytypes.NegRiskInfo
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FeeRateBps returns the fee rate in basis points for a token.
func (c *Client) FeeRateBps(ctx context.Context, tokenID string) (int, error) {
	path := fmt.Sprintf("/fee-rate?token_id=%s", url.QueryEscape(tokenID))
	var wrapper struct {
		FeeRateBps int `json:"fee_rate_bps"`
	}
	if err := c.transport.Get(ctx, path, &wrapper); err != nil {
		return 0, err
	}
	return wrapper.FeeRateBps, nil
}

// LastTradePrice returns the last trade price for a token.
func (c *Client) LastTradePrice(ctx context.Context, tokenID string) (string, error) {
	path := fmt.Sprintf("/last-trade-price?token_id=%s", url.QueryEscape(tokenID))
	var wrapper struct {
		Price string `json:"price"`
	}
	if err := c.transport.Get(ctx, path, &wrapper); err != nil {
		return "", err
	}
	return wrapper.Price, nil
}

// LastTradesPrices returns last trade prices for multiple tokens (POST).
func (c *Client) LastTradesPrices(ctx context.Context, params []polytypes.BookParams) (map[string]string, error) {
	var raw map[string]json.RawMessage
	if err := c.transport.Post(ctx, "/last-trades-prices", params, &raw); err != nil {
		return nil, err
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		var wrapper struct {
			Price string `json:"price"`
		}
		if err := json.Unmarshal(v, &wrapper); err != nil {
			result[k] = string(v)
		} else {
			result[k] = wrapper.Price
		}
	}
	return result, nil
}

// PricesHistory returns OHLCV price history.
func (c *Client) PricesHistory(ctx context.Context, params *polytypes.PriceHistoryParams) (*polytypes.PriceHistory, error) {
	q := url.Values{}
	if params.Market != "" {
		q.Set("market", params.Market)
	}
	if params.Interval != "" {
		q.Set("interval", params.Interval)
	}
	if params.Fidelity > 0 {
		q.Set("fidelity", strconv.Itoa(params.Fidelity))
	}
	if params.StartTS > 0 {
		q.Set("startTs", strconv.FormatInt(params.StartTS, 10))
	}
	if params.EndTS > 0 {
		q.Set("endTs", strconv.FormatInt(params.EndTS, 10))
	}
	path := "/prices-history"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var result polytypes.PriceHistory
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
