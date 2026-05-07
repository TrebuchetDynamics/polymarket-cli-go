package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

// Client provides read-only access to the Polymarket CLOB API.
// Base URL: https://clob.polymarket.com
type Client struct {
	transport *transport.Client
}

const polygonChainID = 137

// BalanceAllowanceParams are query parameters for CLOB balance-allowance.
type BalanceAllowanceParams struct {
	Asset         string
	AssetType     string
	TokenID       string
	SignatureType int
}

// BalanceAllowanceResponse is the authenticated CLOB collateral/token balance.
type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances,omitempty"`
	Allowance  string            `json:"allowance,omitempty"`
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

// CreateOrDeriveAPIKey creates new CLOB L2 credentials with L1 auth, falling
// back to deterministic derivation when a key already exists.
func (c *Client) CreateOrDeriveAPIKey(ctx context.Context, privateKey string) (auth.APIKey, error) {
	key, err := c.createAPIKey(ctx, privateKey)
	if err == nil {
		return key, nil
	}
	return c.DeriveAPIKey(ctx, privateKey)
}

// DeriveAPIKey derives existing CLOB L2 credentials with L1 auth.
func (c *Client) DeriveAPIKey(ctx context.Context, privateKey string) (auth.APIKey, error) {
	headers, err := auth.BuildL1HeadersFromPrivateKey(privateKey, polygonChainID, time.Now().Unix(), 0)
	if err != nil {
		return auth.APIKey{}, err
	}
	var raw apiKeyResponse
	if err := c.transport.GetWithHeaders(ctx, "/auth/derive-api-key", headers, &raw); err != nil {
		return auth.APIKey{}, err
	}
	return raw.apiKey(), nil
}

func (c *Client) createAPIKey(ctx context.Context, privateKey string) (auth.APIKey, error) {
	headers, err := auth.BuildL1HeadersFromPrivateKey(privateKey, polygonChainID, time.Now().Unix(), 0)
	if err != nil {
		return auth.APIKey{}, err
	}
	var raw apiKeyResponse
	if err := c.transport.PostWithHeaders(ctx, "/auth/api-key", nil, headers, &raw); err != nil {
		return auth.APIKey{}, err
	}
	return raw.apiKey(), nil
}

type apiKeyResponse struct {
	APIKey         string `json:"apiKey"`
	APIKeySnake    string `json:"api_key"`
	Secret         string `json:"secret"`
	Passphrase     string `json:"passphrase"`
	PassphraseAlt  string `json:"passPhrase"`
	PassphraseAlt2 string `json:"pass_phrase"`
}

func (r apiKeyResponse) apiKey() auth.APIKey {
	passphrase := firstNonEmpty(r.Passphrase, r.PassphraseAlt, r.PassphraseAlt2)
	return auth.APIKey{
		Key:        firstNonEmpty(r.APIKey, r.APIKeySnake),
		Secret:     r.Secret,
		Passphrase: passphrase,
	}
}

// BalanceAllowance returns collateral or conditional token balance and allowances.
func (c *Client) BalanceAllowance(ctx context.Context, privateKey string, params BalanceAllowanceParams) (*BalanceAllowanceResponse, error) {
	key, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("derive api key: %w", err)
	}
	path := balanceAllowancePath("/balance-allowance", params)
	headers, err := c.l2Headers(privateKey, &key, http.MethodGet, "/balance-allowance", nil)
	if err != nil {
		return nil, err
	}
	var result BalanceAllowanceResponse
	if err := c.transport.GetWithHeaders(ctx, path, headers, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateBalanceAllowance refreshes the CLOB balance/allowance cache.
func (c *Client) UpdateBalanceAllowance(ctx context.Context, privateKey string, params BalanceAllowanceParams) (*BalanceAllowanceResponse, error) {
	key, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("derive api key: %w", err)
	}
	path := balanceAllowancePath("/balance-allowance/update", params)
	headers, err := c.l2Headers(privateKey, &key, http.MethodGet, "/balance-allowance/update", nil)
	if err != nil {
		return nil, err
	}
	var result BalanceAllowanceResponse
	if err := c.transport.GetWithHeaders(ctx, path, headers, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) l2Headers(privateKey string, key *auth.APIKey, method, path string, body *string) (map[string]string, error) {
	signer, err := auth.NewPrivateKeySigner(privateKey, polygonChainID)
	if err != nil {
		return nil, err
	}
	headers, err := auth.BuildL2Headers(key, time.Now().Unix(), method, path, body)
	if err != nil {
		return nil, err
	}
	headers["POLY_ADDRESS"] = signer.Address()
	return headers, nil
}

func balanceAllowancePath(base string, params BalanceAllowanceParams) string {
	q := url.Values{}
	if params.Asset != "" {
		q.Set("asset", params.Asset)
	}
	if params.AssetType != "" {
		q.Set("asset_type", strings.ToUpper(params.AssetType))
	}
	if params.TokenID != "" {
		q.Set("token_id", params.TokenID)
	}
	if params.SignatureType >= 0 {
		q.Set("signature_type", strconv.Itoa(params.SignatureType))
	}
	if len(q) == 0 {
		return base
	}
	return base + "?" + q.Encode()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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

// --- Rewards ---

func (c *Client) RewardsConfig(ctx context.Context) ([]polytypes.RewardsConfig, error) {
	var result []polytypes.RewardsConfig
	if err := c.transport.Get(ctx, "/rewards/config", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) RawRewards(ctx context.Context, market string) ([]polytypes.RawRewards, error) {
	path := fmt.Sprintf("/rewards/raw?market=%s", url.QueryEscape(market))
	var result []polytypes.RawRewards
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) UserEarnings(ctx context.Context, date string) ([]polytypes.UserEarnings, error) {
	path := fmt.Sprintf("/rewards/earnings?date=%s", url.QueryEscape(date))
	var result []polytypes.UserEarnings
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) TotalEarnings(ctx context.Context, date string) (*polytypes.TotalEarnings, error) {
	var result polytypes.TotalEarnings
	if err := c.transport.Get(ctx, "/rewards/total-earnings?date="+url.QueryEscape(date), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) RewardPercentages(ctx context.Context) ([]polytypes.RewardPercentages, error) {
	var result []polytypes.RewardPercentages
	if err := c.transport.Get(ctx, "/rewards/percentages", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) UserRewardsByMarket(ctx context.Context, params *polytypes.UserRewardsByMarketRequest) ([]polytypes.UserRewardsMarket, error) {
	var result []polytypes.UserRewardsMarket
	if err := c.transport.Get(ctx, "/rewards/markets", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) RebatedFees(ctx context.Context) ([]polytypes.RebatedFees, error) {
	var result []polytypes.RebatedFees
	if err := c.transport.Get(ctx, "/rebates", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Simplified & Sampling Markets ---

func (c *Client) SimplifiedMarkets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	path := "/simplified-markets"
	if nextCursor != "" {
		path += "?next_cursor=" + url.QueryEscape(nextCursor)
	}
	var result polytypes.CLOBPaginatedMarkets
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) SamplingMarkets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	path := "/sampling-markets"
	if nextCursor != "" {
		path += "?next_cursor=" + url.QueryEscape(nextCursor)
	}
	var result polytypes.CLOBPaginatedMarkets
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) SamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (*polytypes.CLOBPaginatedMarkets, error) {
	path := "/sampling-simplified-markets"
	if nextCursor != "" {
		path += "?next_cursor=" + url.QueryEscape(nextCursor)
	}
	var result polytypes.CLOBPaginatedMarkets
	if err := c.transport.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Order Scoring ---

func (c *Client) OrderScoring(ctx context.Context, orderID string) (bool, error) {
	var wrapper struct {
		Scoring bool `json:"scoring"`
	}
	if err := c.transport.Get(ctx, "/orders/scoring?order_id="+url.QueryEscape(orderID), &wrapper); err != nil {
		return false, err
	}
	return wrapper.Scoring, nil
}

func (c *Client) OrdersScoring(ctx context.Context, orderIDs []string) ([]bool, error) {
	var result []bool
	body := map[string]interface{}{"order_ids": orderIDs}
	if err := c.transport.Post(ctx, "/orders/scoring", body, &result); err != nil {
		return nil, err
	}
	return result, nil
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
