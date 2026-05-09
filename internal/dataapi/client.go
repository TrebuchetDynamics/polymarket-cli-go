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

// Position is a current Data API position for a user (the proxy/deposit
// wallet, not the EOA). Field names follow Polymarket's documented camelCase
// schema; see https://docs.polymarket.com/api-reference/core/get-current-positions-for-a-user.md.
type Position struct {
	TokenID         string  `json:"asset"`
	ConditionID     string  `json:"conditionId"`
	EventID         string  `json:"eventId"`
	ProxyWallet     string  `json:"proxyWallet"`
	Size            float64 `json:"size"`
	AvgPrice        float64 `json:"avgPrice"`
	InitialValue    float64 `json:"initialValue"`
	CurrentValue    float64 `json:"currentValue"`
	CurrentPrice    float64 `json:"curPrice"`
	CashPnl         float64 `json:"cashPnl"`
	PercentPnl      float64 `json:"percentPnl"`
	TotalBought     float64 `json:"totalBought"`
	RealizedPnl     float64 `json:"realizedPnl"`
	PercentRealized float64 `json:"percentRealizedPnl"`
	// V2 redemption-relevant fields.
	Redeemable      bool   `json:"redeemable"`
	Mergeable       bool   `json:"mergeable"`
	NegativeRisk    bool   `json:"negativeRisk"`
	Outcome         string `json:"outcome"`
	OutcomeIndex    int    `json:"outcomeIndex"`
	OppositeOutcome string `json:"oppositeOutcome"`
	OppositeAsset   string `json:"oppositeAsset"`
	EndDate         string `json:"endDate"`
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	EventSlug       string `json:"eventSlug"`
	Icon            string `json:"icon"`
}

type ClosedPosition struct {
	TokenID         string  `json:"asset"`
	ConditionID     string  `json:"conditionId"`
	ProxyWallet     string  `json:"proxyWallet,omitempty"`
	MarketID        string  `json:"market_id,omitempty"`
	Side            string  `json:"side,omitempty"`
	AvgPrice        float64 `json:"avgPrice"`
	AvgPriceBuy     float64 `json:"avg_price_buy,omitempty"`
	AvgPriceSell    float64 `json:"avg_price_sell,omitempty"`
	Size            float64 `json:"size"`
	TotalBought     float64 `json:"totalBought,omitempty"`
	RealizedPnl     float64 `json:"realizedPnl"`
	CurrentPrice    float64 `json:"curPrice,omitempty"`
	Timestamp       string  `json:"timestamp,omitempty"`
	Title           string  `json:"title,omitempty"`
	Slug            string  `json:"slug,omitempty"`
	Icon            string  `json:"icon,omitempty"`
	EventSlug       string  `json:"eventSlug,omitempty"`
	Outcome         string  `json:"outcome,omitempty"`
	OutcomeIndex    int     `json:"outcomeIndex,omitempty"`
	OppositeOutcome string  `json:"oppositeOutcome,omitempty"`
	OppositeAsset   string  `json:"oppositeAsset,omitempty"`
	EndDate         string  `json:"endDate,omitempty"`
}

type Trade struct {
	ID              string  `json:"id"`
	Market          string  `json:"market"`
	AssetID         string  `json:"asset_id"`
	ProxyWallet     string  `json:"proxyWallet,omitempty"`
	Side            string  `json:"side"`
	Price           float64 `json:"price"`
	Size            float64 `json:"size"`
	FeeRateBps      int     `json:"fee_rate_bps"`
	Outcome         string  `json:"outcome,omitempty"`
	OutcomeIndex    int     `json:"outcomeIndex,omitempty"`
	Title           string  `json:"title,omitempty"`
	Slug            string  `json:"slug,omitempty"`
	EventSlug       string  `json:"eventSlug,omitempty"`
	Icon            string  `json:"icon,omitempty"`
	Status          string  `json:"status,omitempty"`
	TransactionHash string  `json:"transaction_hash,omitempty"`
	TakerOrderID    string  `json:"taker_order_id,omitempty"`
	TraderSide      string  `json:"trader_side,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

func (p *ClosedPosition) UnmarshalJSON(data []byte) error {
	var aux struct {
		TokenID          string          `json:"asset"`
		TokenIDSnake     string          `json:"token_id"`
		ConditionID      string          `json:"conditionId"`
		ConditionIDSnake string          `json:"condition_id"`
		ProxyWallet      string          `json:"proxyWallet"`
		MarketID         string          `json:"market_id"`
		Side             string          `json:"side"`
		AvgPrice         json.RawMessage `json:"avgPrice"`
		AvgPriceBuy      json.RawMessage `json:"avg_price_buy"`
		AvgPriceSell     json.RawMessage `json:"avg_price_sell"`
		Size             json.RawMessage `json:"size"`
		TotalBought      json.RawMessage `json:"totalBought"`
		RealizedPnl      json.RawMessage `json:"realizedPnl"`
		RealizedPnlSnake json.RawMessage `json:"realized_pnl"`
		CurrentPrice     json.RawMessage `json:"curPrice"`
		Timestamp        json.RawMessage `json:"timestamp"`
		Title            string          `json:"title"`
		Slug             string          `json:"slug"`
		Icon             string          `json:"icon"`
		EventSlug        string          `json:"eventSlug"`
		Outcome          string          `json:"outcome"`
		OutcomeIndex     int             `json:"outcomeIndex"`
		OppositeOutcome  string          `json:"oppositeOutcome"`
		OppositeAsset    string          `json:"oppositeAsset"`
		EndDate          string          `json:"endDate"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	avgPrice, err := jsonFloatOrZero(aux.AvgPrice)
	if err != nil {
		return fmt.Errorf("decode closed position avgPrice: %w", err)
	}
	avgPriceBuy, err := jsonFloatOrZero(aux.AvgPriceBuy)
	if err != nil {
		return fmt.Errorf("decode closed position avg_price_buy: %w", err)
	}
	avgPriceSell, err := jsonFloatOrZero(aux.AvgPriceSell)
	if err != nil {
		return fmt.Errorf("decode closed position avg_price_sell: %w", err)
	}
	size, err := jsonFloatOrZero(aux.Size)
	if err != nil {
		return fmt.Errorf("decode closed position size: %w", err)
	}
	totalBought, err := jsonFloatOrZero(aux.TotalBought)
	if err != nil {
		return fmt.Errorf("decode closed position totalBought: %w", err)
	}
	realizedPnl, err := jsonFloatOrZero(firstRaw(aux.RealizedPnl, aux.RealizedPnlSnake))
	if err != nil {
		return fmt.Errorf("decode closed position realizedPnl: %w", err)
	}
	currentPrice, err := jsonFloatOrZero(aux.CurrentPrice)
	if err != nil {
		return fmt.Errorf("decode closed position curPrice: %w", err)
	}
	if avgPrice == 0 {
		avgPrice = avgPriceBuy
	}
	if size == 0 {
		size = totalBought
	}
	*p = ClosedPosition{
		TokenID:         firstNonEmpty(aux.TokenID, aux.TokenIDSnake),
		ConditionID:     firstNonEmpty(aux.ConditionID, aux.ConditionIDSnake),
		ProxyWallet:     aux.ProxyWallet,
		MarketID:        aux.MarketID,
		Side:            aux.Side,
		AvgPrice:        avgPrice,
		AvgPriceBuy:     avgPriceBuy,
		AvgPriceSell:    avgPriceSell,
		Size:            size,
		TotalBought:     totalBought,
		RealizedPnl:     realizedPnl,
		CurrentPrice:    currentPrice,
		Timestamp:       jsonStringOrNumber(aux.Timestamp),
		Title:           aux.Title,
		Slug:            aux.Slug,
		Icon:            aux.Icon,
		EventSlug:       aux.EventSlug,
		Outcome:         aux.Outcome,
		OutcomeIndex:    aux.OutcomeIndex,
		OppositeOutcome: aux.OppositeOutcome,
		OppositeAsset:   aux.OppositeAsset,
		EndDate:         aux.EndDate,
	}
	return nil
}

func (t *Trade) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID                   string          `json:"id"`
		Market               string          `json:"market"`
		ConditionID          string          `json:"conditionId"`
		AssetID              string          `json:"asset_id"`
		AssetIDCamel         string          `json:"assetId"`
		Asset                string          `json:"asset"`
		ProxyWallet          string          `json:"proxyWallet"`
		Side                 string          `json:"side"`
		Price                json.RawMessage `json:"price"`
		Size                 json.RawMessage `json:"size"`
		FeeRateBps           json.RawMessage `json:"fee_rate_bps"`
		FeeRateBpsCamel      json.RawMessage `json:"feeRateBps"`
		Outcome              string          `json:"outcome"`
		OutcomeIndex         int             `json:"outcomeIndex"`
		Title                string          `json:"title"`
		Slug                 string          `json:"slug"`
		EventSlug            string          `json:"eventSlug"`
		Icon                 string          `json:"icon"`
		Status               string          `json:"status"`
		TransactionHash      string          `json:"transaction_hash"`
		TransactionHashCamel string          `json:"transactionHash"`
		TakerOrderID         string          `json:"taker_order_id"`
		TakerOrderIDCamel    string          `json:"takerOrderId"`
		TraderSide           string          `json:"trader_side"`
		TraderSideCamel      string          `json:"traderSide"`
		CreatedAt            json.RawMessage `json:"created_at"`
		Timestamp            json.RawMessage `json:"timestamp"`
		MatchTime            json.RawMessage `json:"match_time"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	price, err := jsonFloatOrZero(aux.Price)
	if err != nil {
		return fmt.Errorf("decode trade price: %w", err)
	}
	size, err := jsonFloatOrZero(aux.Size)
	if err != nil {
		return fmt.Errorf("decode trade size: %w", err)
	}
	feeRateBps, err := jsonIntOrZero(firstRaw(aux.FeeRateBps, aux.FeeRateBpsCamel))
	if err != nil {
		return fmt.Errorf("decode trade fee_rate_bps: %w", err)
	}
	*t = Trade{
		ID:              aux.ID,
		Market:          firstNonEmpty(aux.Market, aux.ConditionID),
		AssetID:         firstNonEmpty(aux.AssetID, aux.AssetIDCamel, aux.Asset),
		ProxyWallet:     aux.ProxyWallet,
		Side:            aux.Side,
		Price:           price,
		Size:            size,
		FeeRateBps:      feeRateBps,
		Outcome:         aux.Outcome,
		OutcomeIndex:    aux.OutcomeIndex,
		Title:           aux.Title,
		Slug:            aux.Slug,
		EventSlug:       aux.EventSlug,
		Icon:            aux.Icon,
		Status:          aux.Status,
		TransactionHash: firstNonEmpty(aux.TransactionHash, aux.TransactionHashCamel),
		TakerOrderID:    firstNonEmpty(aux.TakerOrderID, aux.TakerOrderIDCamel),
		TraderSide:      firstNonEmpty(aux.TraderSide, aux.TraderSideCamel),
		CreatedAt:       firstNonEmpty(jsonStringOrNumber(aux.CreatedAt), jsonStringOrNumber(aux.Timestamp), jsonStringOrNumber(aux.MatchTime)),
	}
	return nil
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

func firstRaw(values ...json.RawMessage) json.RawMessage {
	for _, value := range values {
		if len(value) == 0 {
			continue
		}
		if strings.TrimSpace(string(value)) == "" || strings.TrimSpace(string(value)) == "null" {
			continue
		}
		return value
	}
	return nil
}

func jsonFloatOrZero(raw json.RawMessage) (float64, error) {
	value := jsonStringOrNumber(raw)
	if value == "" {
		return 0, nil
	}
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func jsonIntOrZero(raw json.RawMessage) (int, error) {
	value := jsonStringOrNumber(raw)
	if value == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return n, nil
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
