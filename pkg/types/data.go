// Package types contains public DTOs shared by polygolem SDK packages.
package types

// Position is a current Data API position for a user.
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

// ClosedPosition is a closed Data API position for a user.
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

// Trade is a Data API trade row.
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

// Activity is a Data API user activity row.
type Activity struct {
	Type      string `json:"type"`
	Market    string `json:"market"`
	AssetID   string `json:"asset_id"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Size      string `json:"size"`
	Timestamp string `json:"timestamp"`
}

// Holder is a top-holder Data API row.
type Holder struct {
	Address string  `json:"address"`
	Shares  float64 `json:"shares"`
	Pnl     float64 `json:"pnl"`
	Volume  float64 `json:"volume"`
}

// PortfolioValue is a user's total portfolio value.
type PortfolioValue struct {
	User      string  `json:"user"`
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

// TotalMarketsTraded is the count of markets a user has traded.
type TotalMarketsTraded struct {
	User          string `json:"user"`
	MarketsTraded int    `json:"markets_traded"`
}

// OpenInterest is an open-interest row.
type OpenInterest struct {
	Market    string  `json:"market"`
	AssetID   string  `json:"asset_id"`
	OpenValue float64 `json:"open_value"`
}

// LeaderboardRow is a trader leaderboard row.
type LeaderboardRow struct {
	Rank   int     `json:"rank"`
	User   string  `json:"user"`
	Volume float64 `json:"volume"`
	Pnl    float64 `json:"pnl"`
	ROI    float64 `json:"roi"`
}

// LiveVolumeRow is a live-volume event row.
type LiveVolumeRow struct {
	EventID   string  `json:"event_id"`
	EventSlug string  `json:"event_slug"`
	Title     string  `json:"title"`
	Volume    float64 `json:"volume"`
}

// LiveVolumeResponse is the Data API live-volume response.
type LiveVolumeResponse struct {
	Total  int             `json:"total"`
	Events []LiveVolumeRow `json:"events"`
}
