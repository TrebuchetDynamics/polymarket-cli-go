package polytypes

import "github.com/TrebuchetDynamics/polygolem/pkg/types"

type Comment = types.Comment
type CommentUser = types.CommentUser
type CommentQuery = types.CommentQuery
type CommentByIDQuery = types.CommentByIDQuery
type CommentsByUserQuery = types.CommentsByUserQuery

// --- Rewards (CLOB) ---

// RewardsConfig represents active rewards configuration.
type RewardsConfig struct {
	Market           string  `json:"market"`
	AssetAddress     string  `json:"asset_address"`
	RewardsMinSize   float64 `json:"rewards_min_size"`
	RewardsMaxSpread float64 `json:"rewards_max_spread"`
	Active           bool    `json:"active"`
}

// RawRewards represents raw rewards for a market.
type RawRewards struct {
	Market      string  `json:"market"`
	Date        string  `json:"date"`
	RewardsPaid float64 `json:"rewards_paid"`
	Volume      float64 `json:"volume"`
}

// UserEarnings represents earnings for a user.
type UserEarnings struct {
	Date     string  `json:"date"`
	Earnings float64 `json:"earnings"`
	Market   string  `json:"market,omitempty"`
}

// TotalEarnings represents total earnings.
type TotalEarnings struct {
	Date     string  `json:"date"`
	Earnings float64 `json:"earnings"`
}

// RewardPercentages represents reward percentages.
type RewardPercentages struct {
	Market           string  `json:"market"`
	RewardPercentage float64 `json:"reward_percentage"`
}

// UserRewardsMarket represents user rewards by market.
type UserRewardsMarket struct {
	Market           string  `json:"market"`
	TotalRewards     float64 `json:"total_rewards"`
	RewardPercentage float64 `json:"reward_percentage"`
}

// UserRewardsByMarketRequest represents query params.
type UserRewardsByMarketRequest struct {
	Date          string `json:"date,omitempty"`
	OrderBy       string `json:"order_by,omitempty"`
	NoCompetition bool   `json:"no_competition,omitempty"`
}

// RebatedFees represents current rebated fees for a maker.
type RebatedFees struct {
	MakerAddress string  `json:"maker_address"`
	Market       string  `json:"market,omitempty"`
	TotalRebated float64 `json:"total_rebated"`
	Date         string  `json:"date"`
}

type SportsMarketType = types.SportsMarketType
type KeysetParams = types.KeysetParams
type KeysetResponse[T any] = types.KeysetResponse[T]
type MarketByTokenResponse = types.MarketByTokenResponse
