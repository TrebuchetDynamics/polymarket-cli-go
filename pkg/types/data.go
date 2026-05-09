// Package types contains public DTOs shared by polygolem SDK packages.
package types

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
	AssetID   string  `json:"asset_id,omitempty"`
	OpenValue float64 `json:"value"`
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

// LiveVolumeMarket is one per-market row in an event live-volume response.
type LiveVolumeMarket struct {
	Market string  `json:"market"`
	Value  float64 `json:"value"`
}

// LiveVolumeResponse is the Data API live-volume response.
type LiveVolumeResponse struct {
	Total   float64            `json:"total"`
	Markets []LiveVolumeMarket `json:"markets,omitempty"`
	Events  []LiveVolumeRow    `json:"events,omitempty"`
}
