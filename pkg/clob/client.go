// Package clob exposes the public Polymarket CLOB SDK surface.
//
// Use clob when you need typed access to CLOB market data: market lists,
// order books, prices, spreads, tick sizes, negative-risk metadata, last trade
// prices, and price history. Authenticated methods also expose CLOB L2 API-key
// derivation, balance/allowance reads, order reads, order placement, and
// cancellation. They require an explicit private key argument; callers must
// enforce live-mode gates before invoking mutating methods.
package clob

import (
	"context"

	internalclob "github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

const defaultBaseURL = "https://clob.polymarket.com"

// Config holds CLOB client settings.
type Config struct {
	BaseURL string
	// BuilderCode is the optional V2 order builder attribution bytes32.
	// Empty values sign orders with the zero bytes32 builder code.
	BuilderCode string
	// Credentials are pre-provisioned CLOB L2 HMAC credentials. When set,
	// authenticated deposit-wallet calls use them instead of deriving a key
	// through /auth/derive-api-key.
	Credentials APIKey
}

// DefaultConfig returns production CLOB defaults.
func DefaultConfig() Config {
	return Config{BaseURL: defaultBaseURL}
}

// Client is a Polymarket CLOB client.
type Client struct {
	inner *internalclob.Client
}

// NewClient creates a CLOB client. A zero-valued config uses production.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	inner := internalclob.NewClient(cfg.BaseURL, nil)
	inner.SetBuilderCode(cfg.BuilderCode)
	if apiKeyConfigured(cfg.Credentials) {
		inner.SetL2Credentials(apiKeyToInternal(cfg.Credentials))
	}
	return &Client{inner: inner}
}

// Health checks CLOB API availability.
func (c *Client) Health(ctx context.Context) error {
	return c.inner.Health(ctx)
}

// ServerTime returns the CLOB server time.
func (c *Client) ServerTime(ctx context.Context) (*types.CLOBServerTime, error) {
	row, err := c.inner.ServerTime(ctx)
	if err != nil {
		return nil, err
	}
	return serverTimeFromInternal(row), nil
}

// Markets lists CLOB markets with cursor pagination.
func (c *Client) Markets(ctx context.Context, nextCursor string) (*types.CLOBPaginatedMarkets, error) {
	row, err := c.inner.Markets(ctx, nextCursor)
	if err != nil {
		return nil, err
	}
	return paginatedMarketsFromInternal(row), nil
}

// Market returns one CLOB market by condition ID.
func (c *Client) Market(ctx context.Context, conditionID string) (*types.CLOBMarket, error) {
	row, err := c.inner.Market(ctx, conditionID)
	if err != nil {
		return nil, err
	}
	return marketFromInternal(row), nil
}

// MarketByToken resolves a token ID to its parent CLOB market IDs.
func (c *Client) MarketByToken(ctx context.Context, tokenID string) (*types.CLOBMarketByTokenResponse, error) {
	row, err := c.inner.MarketByToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	return marketByTokenFromInternal(row), nil
}

// OrderBook returns L2 order-book depth for a token.
func (c *Client) OrderBook(ctx context.Context, tokenID string) (*types.CLOBOrderBook, error) {
	row, err := c.inner.OrderBook(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	return orderBookFromInternal(row), nil
}

// OrderBooks returns order books for multiple tokens.
func (c *Client) OrderBooks(ctx context.Context, params []types.CLOBBookParams) ([]types.CLOBOrderBook, error) {
	rows, err := c.inner.OrderBooks(ctx, bookParamsToInternal(params))
	if err != nil {
		return nil, err
	}
	return orderBooksFromInternal(rows), nil
}

// Price returns the best bid or ask for a token.
func (c *Client) Price(ctx context.Context, tokenID, side string) (string, error) {
	return c.inner.Price(ctx, tokenID, side)
}

// Prices returns best prices for multiple tokens.
func (c *Client) Prices(ctx context.Context, params []types.CLOBBookParams) (map[string]string, error) {
	return c.inner.Prices(ctx, bookParamsToInternal(params))
}

// Midpoint returns the midpoint price for a token.
func (c *Client) Midpoint(ctx context.Context, tokenID string) (string, error) {
	return c.inner.Midpoint(ctx, tokenID)
}

// Midpoints returns midpoint prices for multiple tokens.
func (c *Client) Midpoints(ctx context.Context, params []types.CLOBBookParams) (map[string]string, error) {
	return c.inner.Midpoints(ctx, bookParamsToInternal(params))
}

// Spread returns the spread for a token.
func (c *Client) Spread(ctx context.Context, tokenID string) (string, error) {
	return c.inner.Spread(ctx, tokenID)
}

// TickSize returns size and tick metadata for a token.
func (c *Client) TickSize(ctx context.Context, tokenID string) (*types.CLOBTickSize, error) {
	row, err := c.inner.TickSize(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	return tickSizeFromInternal(row), nil
}

// NegRisk returns negative-risk metadata for a token.
func (c *Client) NegRisk(ctx context.Context, tokenID string) (*types.CLOBNegRiskInfo, error) {
	row, err := c.inner.NegRisk(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	return negRiskFromInternal(row), nil
}

// FeeRateBps returns the fee rate in basis points for a token.
func (c *Client) FeeRateBps(ctx context.Context, tokenID string) (int, error) {
	return c.inner.FeeRateBps(ctx, tokenID)
}

// LastTradePrice returns the last trade price for a token.
func (c *Client) LastTradePrice(ctx context.Context, tokenID string) (string, error) {
	return c.inner.LastTradePrice(ctx, tokenID)
}

// LastTradesPrices returns last trade prices for multiple tokens.
func (c *Client) LastTradesPrices(ctx context.Context, params []types.CLOBBookParams) (map[string]string, error) {
	return c.inner.LastTradesPrices(ctx, bookParamsToInternal(params))
}

// PricesHistory returns price history.
func (c *Client) PricesHistory(ctx context.Context, params *types.CLOBPriceHistoryParams) (*types.CLOBPriceHistory, error) {
	row, err := c.inner.PricesHistory(ctx, priceHistoryParamsToInternal(params))
	if err != nil {
		return nil, err
	}
	return priceHistoryFromInternal(row), nil
}

// SimplifiedMarkets returns simplified CLOB markets.
func (c *Client) SimplifiedMarkets(ctx context.Context, nextCursor string) (*types.CLOBPaginatedMarkets, error) {
	row, err := c.inner.SimplifiedMarkets(ctx, nextCursor)
	if err != nil {
		return nil, err
	}
	return paginatedMarketsFromInternal(row), nil
}

// SamplingMarkets returns sampling CLOB markets.
func (c *Client) SamplingMarkets(ctx context.Context, nextCursor string) (*types.CLOBPaginatedMarkets, error) {
	row, err := c.inner.SamplingMarkets(ctx, nextCursor)
	if err != nil {
		return nil, err
	}
	return paginatedMarketsFromInternal(row), nil
}

// SamplingSimplifiedMarkets returns sampling simplified CLOB markets.
func (c *Client) SamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (*types.CLOBPaginatedMarkets, error) {
	row, err := c.inner.SamplingSimplifiedMarkets(ctx, nextCursor)
	if err != nil {
		return nil, err
	}
	return paginatedMarketsFromInternal(row), nil
}

func serverTimeFromInternal(row *polytypes.ServerTime) *types.CLOBServerTime {
	if row == nil {
		return nil
	}
	return &types.CLOBServerTime{
		Timestamp: row.Timestamp,
		ISO:       row.ISO,
	}
}

func paginatedMarketsFromInternal(row *polytypes.CLOBPaginatedMarkets) *types.CLOBPaginatedMarkets {
	if row == nil {
		return nil
	}
	return &types.CLOBPaginatedMarkets{
		Limit:      row.Limit,
		Count:      row.Count,
		NextCursor: row.NextCursor,
		Data:       marketsFromInternal(row.Data),
	}
}

func marketFromInternal(row *polytypes.CLOBMarket) *types.CLOBMarket {
	if row == nil {
		return nil
	}
	out := marketValueFromInternal(*row)
	return &out
}

func marketByTokenFromInternal(row *polytypes.CLOBMarketByTokenResponse) *types.CLOBMarketByTokenResponse {
	if row == nil {
		return nil
	}
	return &types.CLOBMarketByTokenResponse{
		ConditionID:      row.ConditionID,
		PrimaryTokenID:   row.PrimaryTokenID,
		SecondaryTokenID: row.SecondaryTokenID,
	}
}

func marketsFromInternal(rows []polytypes.CLOBMarket) []types.CLOBMarket {
	out := make([]types.CLOBMarket, len(rows))
	for i, row := range rows {
		out[i] = marketValueFromInternal(row)
	}
	return out
}

func marketValueFromInternal(row polytypes.CLOBMarket) types.CLOBMarket {
	return types.CLOBMarket{
		ConditionID:           row.ConditionID,
		QuestionID:            row.QuestionID,
		Tokens:                tokensFromInternal(row.Tokens),
		GameStartTime:         row.GameStartTime,
		RewardsMinSize:        row.RewardsMinSize,
		RewardsMaxSpread:      row.RewardsMaxSpread,
		Spread:                row.Spread,
		EnableOrderBook:       row.EnableOrderBook,
		OrderPriceMinTickSize: row.OrderPriceMinTickSize,
		OrderMinSize:          row.OrderMinSize,
		Closed:                row.Closed,
		Archived:              row.Archived,
		AcceptingOrders:       row.AcceptingOrders,
		NegRisk:               row.NegRisk,
		NegRiskMarketID:       row.NegRiskMarketID,
		NegRiskRequestID:      row.NegRiskRequestID,
		MakerBaseFee:          row.MakerBaseFee,
		TakerBaseFee:          row.TakerBaseFee,
		NotificationsEnabled:  row.NotificationsEnabled,
		RFQEnabled:            row.RFQEnabled,
		TakerOrderDelay:       row.TakerOrderDelay,
		BlockaidCheckEnabled:  row.BlockaidCheckEnabled,
		FeeDetails: types.CLOBFeeDetails{
			Rate:      row.FeeDetails.Rate,
			Exponent:  row.FeeDetails.Exponent,
			TakerOnly: row.FeeDetails.TakerOnly,
		},
		MinimumOrderAge: row.MinimumOrderAge,
	}
}

func tokensFromInternal(rows []polytypes.Token) []types.CLOBToken {
	out := make([]types.CLOBToken, len(rows))
	for i, row := range rows {
		out[i] = types.CLOBToken{
			TokenID: row.TokenID,
			Outcome: row.Outcome,
			Price:   string(row.Price),
			Winner:  row.Winner,
		}
	}
	return out
}

func orderBookFromInternal(row *polytypes.OrderBook) *types.CLOBOrderBook {
	if row == nil {
		return nil
	}
	out := orderBookValueFromInternal(*row)
	return &out
}

func orderBooksFromInternal(rows []polytypes.OrderBook) []types.CLOBOrderBook {
	out := make([]types.CLOBOrderBook, len(rows))
	for i, row := range rows {
		out[i] = orderBookValueFromInternal(row)
	}
	return out
}

func orderBookValueFromInternal(row polytypes.OrderBook) types.CLOBOrderBook {
	return types.CLOBOrderBook{
		Market:         row.Market,
		AssetID:        row.AssetID,
		Timestamp:      row.Timestamp,
		Hash:           row.Hash,
		Bids:           levelsFromInternal(row.Bids),
		Asks:           levelsFromInternal(row.Asks),
		MinOrderSize:   row.MinOrderSize,
		TickSize:       row.TickSize,
		NegRisk:        row.NegRisk,
		LastTradePrice: row.LastTradePrice,
	}
}

func levelsFromInternal(rows []polytypes.OrderBookLevel) []types.CLOBOrderBookLevel {
	out := make([]types.CLOBOrderBookLevel, len(rows))
	for i, row := range rows {
		out[i] = types.CLOBOrderBookLevel{
			Price: row.Price,
			Size:  row.Size,
		}
	}
	return out
}

func tickSizeFromInternal(row *polytypes.TickSize) *types.CLOBTickSize {
	if row == nil {
		return nil
	}
	return &types.CLOBTickSize{
		MinimumTickSize:  row.MinimumTickSize,
		MinimumOrderSize: row.MinimumOrderSize,
		TickSize:         row.TickSize,
	}
}

func negRiskFromInternal(row *polytypes.NegRiskInfo) *types.CLOBNegRiskInfo {
	if row == nil {
		return nil
	}
	return &types.CLOBNegRiskInfo{
		NegRisk:         row.NegRisk,
		NegRiskMarketID: row.NegRiskMarketID,
		NegRiskFeeBips:  row.NegRiskFeeBips,
	}
}

func priceHistoryFromInternal(row *polytypes.PriceHistory) *types.CLOBPriceHistory {
	if row == nil {
		return nil
	}
	out := make([]types.CLOBPricePoint, len(row.History))
	for i, point := range row.History {
		out[i] = types.CLOBPricePoint{
			T:        point.T,
			P:        point.P,
			Volume:   point.Volume,
			Interval: point.Interval,
		}
	}
	return &types.CLOBPriceHistory{History: out}
}

// OrderScoring checks if a specific order is currently scoring for maker rewards.
func (c *Client) OrderScoring(ctx context.Context, orderID string) (bool, error) {
	return c.inner.OrderScoring(ctx, orderID)
}

// BuilderTrades returns trades attributed to the configured builder code.
func (c *Client) BuilderTrades(ctx context.Context, limit int) ([]internalclob.BuilderTrade, error) {
	return c.inner.BuilderTrades(ctx, limit)
}

func bookParamsToInternal(params []types.CLOBBookParams) []polytypes.BookParams {
	out := make([]polytypes.BookParams, len(params))
	for i, param := range params {
		out[i] = polytypes.BookParams{
			TokenID: param.TokenID,
			Side:    param.Side,
		}
	}
	return out
}

func priceHistoryParamsToInternal(params *types.CLOBPriceHistoryParams) *polytypes.PriceHistoryParams {
	if params == nil {
		return &polytypes.PriceHistoryParams{}
	}
	return &polytypes.PriceHistoryParams{
		Market:   params.Market,
		Interval: params.Interval,
		Fidelity: params.Fidelity,
		StartTS:  params.StartTS,
		EndTS:    params.EndTS,
	}
}
