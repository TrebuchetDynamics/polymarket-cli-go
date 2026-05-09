package types

// Comment is a Gamma API comment.
type Comment struct {
	ID        string         `json:"id"`
	Body      string         `json:"body"`
	User      CommentUser    `json:"user"`
	CreatedAt NormalizedTime `json:"createdAt"`
	UpdatedAt NormalizedTime `json:"updatedAt"`
	ParentID  *int           `json:"parentId,omitempty"`
	Replies   []Comment      `json:"replies,omitempty"`
}

// CommentUser is the public comment author payload.
type CommentUser struct {
	Address      string `json:"address"`
	Pseudonym    string `json:"pseudonym"`
	ProfileImage string `json:"profileImage"`
}

// CommentQuery is the public query shape for listing comments.
type CommentQuery struct {
	EntityID   *int    `json:"parent_entity_id,omitempty"`
	EntityType *string `json:"parent_entity_type,omitempty"`
	Limit      int     `json:"limit,omitempty"`
	Offset     int     `json:"offset,omitempty"`
}

// CommentByIDQuery is the public query shape for fetching a single comment.
type CommentByIDQuery struct {
	IncludeReplies *bool `json:"include_replies,omitempty"`
}

// CommentsByUserQuery is the public query shape for user comments.
type CommentsByUserQuery struct {
	UserAddress string `json:"user_address"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// SportsMarketType is a Gamma sports market type catalog row.
type SportsMarketType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// KeysetParams is the public keyset pagination query shape.
type KeysetParams struct {
	Limit     int    `json:"limit,omitempty"`
	KeysetID  string `json:"keyset_id,omitempty"`
	Ascending *bool  `json:"ascending,omitempty"`
	Active    *bool  `json:"active,omitempty"`
	Closed    *bool  `json:"closed,omitempty"`
	Order     string `json:"order,omitempty"`
}

// KeysetResponse wraps a keyset-paginated Gamma response.
type KeysetResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

// MarketByTokenResponse resolves a CLOB token ID back to Gamma metadata.
type MarketByTokenResponse struct {
	Market  Market `json:"market"`
	TokenID string `json:"token_id"`
	Outcome string `json:"outcome"`
}
