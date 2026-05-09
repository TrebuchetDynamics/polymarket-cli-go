// Package data exposes the public, read-only Polymarket Data API SDK surface.
package data

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/internal/dataapi"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

const defaultBaseURL = "https://data-api.polymarket.com"

// Config holds Data API client settings.
type Config struct {
	BaseURL string
}

// DefaultConfig returns production Data API defaults.
func DefaultConfig() Config {
	return Config{BaseURL: defaultBaseURL}
}

// Client is a read-only Polymarket Data API client.
type Client struct {
	inner *dataapi.Client
}

// NewClient creates a Data API client. Zero-valued config uses production.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{inner: dataapi.NewClient(cfg.BaseURL, nil)}
}

// Health checks Data API availability.
func (c *Client) Health(ctx context.Context) error {
	return c.inner.Health(ctx)
}

// CurrentPositions returns current open positions for a user.
func (c *Client) CurrentPositions(ctx context.Context, user string) ([]types.Position, error) {
	return c.CurrentPositionsWithLimit(ctx, user, 0)
}

// CurrentPositionsWithLimit returns current open positions for a user with a row limit.
func (c *Client) CurrentPositionsWithLimit(ctx context.Context, user string, limit int) ([]types.Position, error) {
	rows, err := c.inner.CurrentPositionsWithLimit(ctx, user, limit)
	if err != nil {
		return nil, err
	}
	return positionsFromInternal(rows), nil
}

// ClosedPositions returns closed positions for a user.
func (c *Client) ClosedPositions(ctx context.Context, user string) ([]types.ClosedPosition, error) {
	return c.ClosedPositionsWithLimit(ctx, user, 0)
}

// ClosedPositionsWithLimit returns closed positions for a user with a row limit.
func (c *Client) ClosedPositionsWithLimit(ctx context.Context, user string, limit int) ([]types.ClosedPosition, error) {
	rows, err := c.inner.ClosedPositionsWithLimit(ctx, user, limit)
	if err != nil {
		return nil, err
	}
	return closedPositionsFromInternal(rows), nil
}

// Trades returns public trades for a user.
func (c *Client) Trades(ctx context.Context, user string, limit int) ([]types.Trade, error) {
	rows, err := c.inner.Trades(ctx, user, limit)
	if err != nil {
		return nil, err
	}
	return tradesFromInternal(rows), nil
}

// Activity returns public activity for a user.
func (c *Client) Activity(ctx context.Context, user string, limit int) ([]types.Activity, error) {
	rows, err := c.inner.Activity(ctx, user, limit)
	if err != nil {
		return nil, err
	}
	return activitiesFromInternal(rows), nil
}

// TopHolders returns top holders for a market condition hash.
func (c *Client) TopHolders(ctx context.Context, market string, limit int) ([]types.Holder, error) {
	rows, err := c.inner.TopHolders(ctx, market, limit)
	if err != nil {
		return nil, err
	}
	return holdersFromInternal(rows), nil
}

// TotalValue returns total portfolio value for a user.
func (c *Client) TotalValue(ctx context.Context, user string) (*types.PortfolioValue, error) {
	row, err := c.inner.TotalValue(ctx, user)
	if err != nil {
		return nil, err
	}
	return portfolioValueFromInternal(row), nil
}

// MarketsTraded returns the count of markets traded by a user.
func (c *Client) MarketsTraded(ctx context.Context, user string) (*types.TotalMarketsTraded, error) {
	row, err := c.inner.MarketsTraded(ctx, user)
	if err != nil {
		return nil, err
	}
	return totalMarketsTradedFromInternal(row), nil
}

// OpenInterest returns open interest for a market condition hash.
func (c *Client) OpenInterest(ctx context.Context, market string) (*types.OpenInterest, error) {
	row, err := c.inner.OpenInterest(ctx, market)
	if err != nil {
		return nil, err
	}
	return openInterestFromInternal(row), nil
}

// TraderLeaderboard returns the trader leaderboard.
func (c *Client) TraderLeaderboard(ctx context.Context, limit int) ([]types.LeaderboardRow, error) {
	rows, err := c.inner.TraderLeaderboard(ctx, limit)
	if err != nil {
		return nil, err
	}
	return leaderboardRowsFromInternal(rows), nil
}

// LiveVolume returns live volume for an event ID.
func (c *Client) LiveVolume(ctx context.Context, eventID int) (*types.LiveVolumeResponse, error) {
	row, err := c.inner.LiveVolume(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return liveVolumeFromInternal(row), nil
}

func positionsFromInternal(rows []dataapi.Position) []types.Position {
	out := make([]types.Position, len(rows))
	for i, row := range rows {
		out[i] = types.Position{
			TokenID:       row.TokenID,
			ConditionID:   row.ConditionID,
			MarketID:      row.MarketID,
			Side:          row.Side,
			AvgPrice:      row.AvgPrice,
			Size:          row.Size,
			CurrentPrice:  row.CurrentPrice,
			UnrealizedPnl: row.UnrealizedPnl,
		}
	}
	return out
}

func closedPositionsFromInternal(rows []dataapi.ClosedPosition) []types.ClosedPosition {
	out := make([]types.ClosedPosition, len(rows))
	for i, row := range rows {
		out[i] = types.ClosedPosition{
			TokenID:      row.TokenID,
			ConditionID:  row.ConditionID,
			MarketID:     row.MarketID,
			Side:         row.Side,
			AvgPriceBuy:  row.AvgPriceBuy,
			AvgPriceSell: row.AvgPriceSell,
			Size:         row.Size,
			RealizedPnl:  row.RealizedPnl,
		}
	}
	return out
}

func tradesFromInternal(rows []dataapi.Trade) []types.Trade {
	out := make([]types.Trade, len(rows))
	for i, row := range rows {
		out[i] = types.Trade{
			ID:         row.ID,
			Market:     row.Market,
			AssetID:    row.AssetID,
			Side:       row.Side,
			Price:      row.Price,
			Size:       row.Size,
			FeeRateBps: row.FeeRateBps,
			CreatedAt:  row.CreatedAt,
		}
	}
	return out
}

func activitiesFromInternal(rows []dataapi.Activity) []types.Activity {
	out := make([]types.Activity, len(rows))
	for i, row := range rows {
		out[i] = types.Activity{
			Type:      row.Type,
			Market:    row.Market,
			AssetID:   row.AssetID,
			Side:      row.Side,
			Price:     row.Price,
			Size:      row.Size,
			Timestamp: row.Timestamp,
		}
	}
	return out
}

func holdersFromInternal(rows []dataapi.MetaHolder) []types.Holder {
	out := make([]types.Holder, len(rows))
	for i, row := range rows {
		out[i] = types.Holder{
			Address: row.Address,
			Shares:  row.Shares,
			Pnl:     row.Pnl,
			Volume:  row.Volume,
		}
	}
	return out
}

func portfolioValueFromInternal(row *dataapi.TotalValue) *types.PortfolioValue {
	if row == nil {
		return nil
	}
	return &types.PortfolioValue{
		User:      row.User,
		Value:     row.Value,
		Timestamp: row.Timestamp,
	}
}

func totalMarketsTradedFromInternal(row *dataapi.TotalMarketsTraded) *types.TotalMarketsTraded {
	if row == nil {
		return nil
	}
	return &types.TotalMarketsTraded{
		User:          row.User,
		MarketsTraded: row.MarketsTraded,
	}
}

func openInterestFromInternal(row *dataapi.OpenInterest) *types.OpenInterest {
	if row == nil {
		return nil
	}
	return &types.OpenInterest{
		Market:    row.Market,
		AssetID:   row.AssetID,
		OpenValue: row.OpenValue,
	}
}

func leaderboardRowsFromInternal(rows []dataapi.TraderLeaderboardEntry) []types.LeaderboardRow {
	out := make([]types.LeaderboardRow, len(rows))
	for i, row := range rows {
		out[i] = types.LeaderboardRow{
			Rank:   row.Rank,
			User:   row.User,
			Volume: row.Volume,
			Pnl:    row.Pnl,
			ROI:    row.ROI,
		}
	}
	return out
}

func liveVolumeFromInternal(row *dataapi.LiveVolumeResponse) *types.LiveVolumeResponse {
	if row == nil {
		return nil
	}
	out := &types.LiveVolumeResponse{
		Total:   row.Total,
		Markets: make([]types.LiveVolumeMarket, len(row.Markets)),
		Events:  make([]types.LiveVolumeRow, len(row.Events)),
	}
	for i, market := range row.Markets {
		out.Markets[i] = types.LiveVolumeMarket{
			Market: market.Market,
			Value:  market.Value,
		}
	}
	for i, event := range row.Events {
		out.Events[i] = types.LiveVolumeRow{
			EventID:   event.EventID,
			EventSlug: event.EventSlug,
			Title:     event.Title,
			Volume:    event.Volume,
		}
	}
	return out
}
