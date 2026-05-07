// Package polytypes — CLOB types stolen from polymarket-go-sdk and rs-clob-client.
package polytypes

// OrderBookLevel is a single price level in the order book.
type OrderBookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// OrderBook represents L2 order book depth for a token.
type OrderBook struct {
	Market    string            `json:"market"`
	AssetID   string            `json:"asset_id"`
	Timestamp string            `json:"timestamp"`
	Hash      string            `json:"hash"`
	Bids      []OrderBookLevel  `json:"bids"`
	Asks      []OrderBookLevel  `json:"asks"`
}

// TickSize represents the minimum tick size for a market.
type TickSize struct {
	MinimumTickSize  string `json:"minimum_tick_size"`
	MinimumOrderSize string `json:"minimum_order_size"`
	TickSize         string `json:"tick_size"`
}

// NegRiskInfo represents negative risk market info.
type NegRiskInfo struct {
	NegRisk         bool   `json:"neg_risk"`
	NegRiskMarketID string `json:"neg_risk_market_id,omitempty"`
	NegRiskFeeBips  int    `json:"neg_risk_fee_bips,omitempty"`
}

// FeeRate represents the fee rate in basis points.
type FeeRate struct {
	FeeRateBps int `json:"fee_rate_bps"`
}

// PricePoint represents a single price history data point.
type PricePoint struct {
	T        string `json:"t"` // timestamp
	P        string `json:"p"` // price
	Volume   string `json:"v,omitempty"`
	Interval string `json:"interval,omitempty"`
}

// PriceHistory represents OHLCV price history.
type PriceHistory struct {
	History []PricePoint `json:"history"`
}

// CLOBMarket represents a market from the CLOB API.
type CLOBMarket struct {
	ConditionID            string   `json:"condition_id"`
	QuestionID             string   `json:"question_id"`
	Tokens                 []Token  `json:"tokens"`
	RewardsMinSize         float64  `json:"rewards_min_size"`
	RewardsMaxSpread       float64  `json:"rewards_max_spread"`
	Spread                 float64  `json:"spread"`
	EnableOrderBook        bool     `json:"enable_order_book"`
	OrderPriceMinTickSize  float64  `json:"order_price_min_tick_size"`
	OrderMinSize           float64  `json:"order_min_size"`
	Closed                 bool     `json:"closed"`
	Archived               bool     `json:"archived"`
	AcceptingOrders        bool     `json:"accepting_orders"`
	NegRisk                bool     `json:"neg_risk"`
	NegRiskMarketID        string   `json:"neg_risk_market_id,omitempty"`
	NegRiskRequestID       string   `json:"neg_risk_request_id,omitempty"`
	MakerBaseFee           int      `json:"maker_base_fee"`
	TakerBaseFee           int      `json:"taker_base_fee"`
	NotificationsEnabled   bool     `json:"notifications_enabled"`
}

// Token represents a CLOB outcome token.
type Token struct {
	TokenID string `json:"token_id"`
	Outcome string `json:"outcome"`
	Price   string `json:"price"`
	Winner  bool   `json:"winner"`
}

// CLOBPaginatedMarkets represents cursor-paginated CLOB markets.
type CLOBPaginatedMarkets struct {
	Limit       int          `json:"limit"`
	Count       int          `json:"count"`
	NextCursor  string       `json:"next_cursor"`
	Data        []CLOBMarket `json:"data"`
}

// BookParams represents parameters for batch order book requests.
type BookParams struct {
	TokenID string `json:"token_id"`
	Side    string `json:"side,omitempty"` // BUY or SELL (for price requests)
}

// PriceHistoryParams represents parameters for price history requests.
type PriceHistoryParams struct {
	Market      string `json:"market,omitempty"`
	Interval    string `json:"interval,omitempty"` // 1m, 1h, 6h, 1d, 1w, max
	Fidelity    int    `json:"fidelity,omitempty"`
	StartTS     int64  `json:"start_ts,omitempty"`
	EndTS       int64  `json:"end_ts,omitempty"`
}

// MidpointResponse represents a midpoint price.
type MidpointResponse struct {
	Midpoint string `json:"mid"`
}

// PriceResponse represents a price/spread response.
type PriceResponse struct {
	Price  string `json:"price"`
	Spread string `json:"spread,omitempty"`
}

// ServerTime represents the server time response.
type ServerTime struct {
	Timestamp string `json:"timestamp"`
	ISO       string `json:"iso"`
}

// EnrichedMarket joins Gamma metadata with CLOB details.
type EnrichedMarket struct {
	// From Gamma
	Market      Market  `json:"market"`
	// From CLOB
	TickSize    TickSize `json:"tick_size"`
	NegRisk     bool     `json:"neg_risk"`
	FeeRateBps  int      `json:"fee_rate_bps"`
	// Optional
	OrderBook   *OrderBook `json:"order_book,omitempty"`
	LastPrice   string     `json:"last_price,omitempty"`
	Midpoint    string     `json:"midpoint,omitempty"`
	Spread      string     `json:"spread,omitempty"`
}
