package polytypes

// --- Comments (Gamma) ---

// Comment represents a Polymarket comment.
type Comment struct {
	ID        string         `json:"id"`
	Body      string         `json:"body"`
	User      CommentUser    `json:"user"`
	CreatedAt NormalizedTime `json:"createdAt"`
	UpdatedAt NormalizedTime `json:"updatedAt"`
	ParentID  *int           `json:"parentId,omitempty"`
	Replies   []Comment      `json:"replies,omitempty"`
}

// CommentUser represents a comment author.
type CommentUser struct {
	Address      string `json:"address"`
	Pseudonym    string `json:"pseudonym"`
	ProfileImage string `json:"profileImage"`
}

// CommentQuery represents query parameters for listing comments.
type CommentQuery struct {
	EntityID   *int    `json:"entity_id,omitempty"`
	EntityType *string `json:"entity_type,omitempty"`
	Limit      int     `json:"limit,omitempty"`
	Offset     int     `json:"offset,omitempty"`
}

// CommentByIDQuery represents query parameters for a single comment.
type CommentByIDQuery struct {
	IncludeReplies *bool `json:"include_replies,omitempty"`
}

// CommentsByUserQuery represents query parameters for user comments.
type CommentsByUserQuery struct {
	UserAddress string `json:"user_address"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

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

// --- Sports Market Types (Gamma) ---

// SportsMarketType represents a valid sports market type.
type SportsMarketType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// --- Keyset Pagination (Gamma) ---

// KeysetParams represents keyset pagination parameters.
type KeysetParams struct {
	Limit     int    `json:"limit,omitempty"`
	KeysetID  string `json:"keyset_id,omitempty"`
	Ascending *bool  `json:"ascending,omitempty"`
	Active    *bool  `json:"active,omitempty"`
	Closed    *bool  `json:"closed,omitempty"`
	Order     string `json:"order,omitempty"`
}

// KeysetResponse wraps a keyset-paginated response.
type KeysetResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

// --- Market by Token (Gamma) ---

// MarketByTokenResponse represents a market resolved by CLOB token ID.
type MarketByTokenResponse struct {
	Market  Market `json:"market"`
	TokenID string `json:"token_id"`
	Outcome string `json:"outcome"`
}
