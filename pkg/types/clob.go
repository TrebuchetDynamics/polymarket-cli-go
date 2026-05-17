package types

import "encoding/json"

// CLOBServerTime is the CLOB server-time response.
type CLOBServerTime struct {
	Timestamp string `json:"timestamp"`
	ISO       string `json:"iso"`
}

// CLOBOrderBookLevel is one price level in a CLOB order book.
type CLOBOrderBookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// CLOBOrderBook is a public CLOB order-book snapshot for one outcome token.
type CLOBOrderBook struct {
	Market         string               `json:"market"`
	AssetID        string               `json:"asset_id"`
	Timestamp      string               `json:"timestamp"`
	Hash           string               `json:"hash"`
	Bids           []CLOBOrderBookLevel `json:"bids"`
	Asks           []CLOBOrderBookLevel `json:"asks"`
	MinOrderSize   string               `json:"min_order_size,omitempty"`
	TickSize       string               `json:"tick_size,omitempty"`
	NegRisk        bool                 `json:"neg_risk,omitempty"`
	LastTradePrice string               `json:"last_trade_price,omitempty"`
}

// CLOBTickSize is the minimum size and price increment metadata for a token.
type CLOBTickSize struct {
	MinimumTickSize  string `json:"minimum_tick_size"`
	MinimumOrderSize string `json:"minimum_order_size"`
	TickSize         string `json:"tick_size"`
}

// CLOBNegRiskInfo is negative-risk metadata for a token.
type CLOBNegRiskInfo struct {
	NegRisk         bool   `json:"neg_risk"`
	NegRiskMarketID string `json:"neg_risk_market_id,omitempty"`
	NegRiskFeeBips  int    `json:"neg_risk_fee_bips,omitempty"`
}

// CLOBFeeDetails is the fee curve metadata returned by CLOB market info.
type CLOBFeeDetails struct {
	Rate      float64 `json:"rate,omitempty"`
	Exponent  float64 `json:"exponent,omitempty"`
	TakerOnly bool    `json:"taker_only,omitempty"`
}

// CLOBMarket is a market from the CLOB API.
type CLOBMarket struct {
	ConditionID           string         `json:"condition_id"`
	QuestionID            string         `json:"question_id"`
	Tokens                []CLOBToken    `json:"tokens"`
	GameStartTime         string         `json:"game_start_time,omitempty"`
	RewardsMinSize        float64        `json:"rewards_min_size"`
	RewardsMaxSpread      float64        `json:"rewards_max_spread"`
	Spread                float64        `json:"spread"`
	EnableOrderBook       bool           `json:"enable_order_book"`
	OrderPriceMinTickSize float64        `json:"order_price_min_tick_size"`
	OrderMinSize          float64        `json:"order_min_size"`
	Closed                bool           `json:"closed"`
	Archived              bool           `json:"archived"`
	AcceptingOrders       bool           `json:"accepting_orders"`
	NegRisk               bool           `json:"neg_risk"`
	NegRiskMarketID       string         `json:"neg_risk_market_id,omitempty"`
	NegRiskRequestID      string         `json:"neg_risk_request_id,omitempty"`
	MakerBaseFee          int            `json:"maker_base_fee"`
	TakerBaseFee          int            `json:"taker_base_fee"`
	NotificationsEnabled  bool           `json:"notifications_enabled"`
	RFQEnabled            bool           `json:"rfq_enabled,omitempty"`
	TakerOrderDelay       bool           `json:"taker_order_delay,omitempty"`
	BlockaidCheckEnabled  bool           `json:"blockaid_check_enabled,omitempty"`
	FeeDetails            CLOBFeeDetails `json:"fee_details,omitempty"`
	MinimumOrderAge       int            `json:"minimum_order_age,omitempty"`
}

func (m *CLOBMarket) UnmarshalJSON(b []byte) error {
	type alias CLOBMarket
	var raw struct {
		alias
		ConditionIDShort   string `json:"c"`
		GameStartTimeShort string `json:"gst"`
		TokensShort        []struct {
			TokenID string `json:"t"`
			Outcome string `json:"o"`
			Price   string `json:"p"`
			Winner  bool   `json:"w"`
		} `json:"t"`
		RewardsShort *struct {
			MinSize         *float64 `json:"mi"`
			MaxSpread       *float64 `json:"ma"`
			MinimumOrderAge *int     `json:"moas"`
		} `json:"r"`
		OrderMinSizeShort          *float64 `json:"mos"`
		OrderPriceMinTickSizeShort *float64 `json:"mts"`
		MakerBaseFeeShort          *int     `json:"mbf"`
		TakerBaseFeeShort          *int     `json:"tbf"`
		AcceptingOrdersShort       *bool    `json:"ao"`
		EnableOrderBookShort       *bool    `json:"cbos"`
		NegRiskShort               *bool    `json:"nr"`
		RFQEnabledShort            *bool    `json:"rfqe"`
		TakerOrderDelayShort       *bool    `json:"itode"`
		BlockaidCheckEnabledShort  *bool    `json:"ibce"`
		FeeDetailsShort            *struct {
			Rate      *float64 `json:"r"`
			Exponent  *float64 `json:"e"`
			TakerOnly *bool    `json:"to"`
		} `json:"fd"`
		MinimumOrderAgeShort *int `json:"oas"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*m = CLOBMarket(raw.alias)
	if m.ConditionID == "" {
		m.ConditionID = raw.ConditionIDShort
	}
	if m.GameStartTime == "" {
		m.GameStartTime = raw.GameStartTimeShort
	}
	if len(m.Tokens) == 0 && len(raw.TokensShort) > 0 {
		m.Tokens = make([]CLOBToken, len(raw.TokensShort))
		for i, token := range raw.TokensShort {
			m.Tokens[i] = CLOBToken{
				TokenID: token.TokenID,
				Outcome: token.Outcome,
				Price:   token.Price,
				Winner:  token.Winner,
			}
		}
	}
	if raw.RewardsShort != nil {
		if raw.RewardsShort.MinSize != nil {
			m.RewardsMinSize = *raw.RewardsShort.MinSize
		}
		if raw.RewardsShort.MaxSpread != nil {
			m.RewardsMaxSpread = *raw.RewardsShort.MaxSpread
		}
		if raw.RewardsShort.MinimumOrderAge != nil {
			m.MinimumOrderAge = *raw.RewardsShort.MinimumOrderAge
		}
	}
	if raw.OrderMinSizeShort != nil {
		m.OrderMinSize = *raw.OrderMinSizeShort
	}
	if raw.OrderPriceMinTickSizeShort != nil {
		m.OrderPriceMinTickSize = *raw.OrderPriceMinTickSizeShort
	}
	if raw.MakerBaseFeeShort != nil {
		m.MakerBaseFee = *raw.MakerBaseFeeShort
	}
	if raw.TakerBaseFeeShort != nil {
		m.TakerBaseFee = *raw.TakerBaseFeeShort
	}
	if raw.AcceptingOrdersShort != nil {
		m.AcceptingOrders = *raw.AcceptingOrdersShort
	}
	if raw.EnableOrderBookShort != nil {
		m.EnableOrderBook = *raw.EnableOrderBookShort
	}
	if raw.NegRiskShort != nil {
		m.NegRisk = *raw.NegRiskShort
	}
	if raw.RFQEnabledShort != nil {
		m.RFQEnabled = *raw.RFQEnabledShort
	}
	if raw.TakerOrderDelayShort != nil {
		m.TakerOrderDelay = *raw.TakerOrderDelayShort
	}
	if raw.BlockaidCheckEnabledShort != nil {
		m.BlockaidCheckEnabled = *raw.BlockaidCheckEnabledShort
	}
	if raw.FeeDetailsShort != nil {
		if raw.FeeDetailsShort.Rate != nil {
			m.FeeDetails.Rate = *raw.FeeDetailsShort.Rate
		}
		if raw.FeeDetailsShort.Exponent != nil {
			m.FeeDetails.Exponent = *raw.FeeDetailsShort.Exponent
		}
		if raw.FeeDetailsShort.TakerOnly != nil {
			m.FeeDetails.TakerOnly = *raw.FeeDetailsShort.TakerOnly
		}
	}
	if raw.MinimumOrderAgeShort != nil {
		m.MinimumOrderAge = *raw.MinimumOrderAgeShort
	}
	return nil
}

// CLOBMarketByTokenResponse resolves a CLOB token ID to its parent market.
type CLOBMarketByTokenResponse struct {
	ConditionID      string `json:"condition_id"`
	PrimaryTokenID   string `json:"primary_token_id"`
	SecondaryTokenID string `json:"secondary_token_id"`
}

// CLOBToken is an outcome token listed on a CLOB market.
type CLOBToken struct {
	TokenID string `json:"token_id"`
	Outcome string `json:"outcome"`
	Price   string `json:"price"`
	Winner  bool   `json:"winner"`
}

// CLOBPaginatedMarkets is the cursor-paginated CLOB market-list response.
type CLOBPaginatedMarkets struct {
	Limit      int          `json:"limit"`
	Count      int          `json:"count"`
	NextCursor string       `json:"next_cursor"`
	Data       []CLOBMarket `json:"data"`
}

// CLOBBookParams identifies one token, and optionally a side, for batch CLOB
// book and price requests.
type CLOBBookParams struct {
	TokenID string `json:"token_id"`
	Side    string `json:"side,omitempty"`
}

// CLOBPricePoint is one price-history point.
type CLOBPricePoint struct {
	T        string `json:"t"`
	P        string `json:"p"`
	Volume   string `json:"v,omitempty"`
	Interval string `json:"interval,omitempty"`
}

// CLOBPriceHistory is the CLOB price-history response.
type CLOBPriceHistory struct {
	History []CLOBPricePoint `json:"history"`
}

// CLOBPriceHistoryParams filters CLOB price-history requests.
type CLOBPriceHistoryParams struct {
	Market   string `json:"market,omitempty"`
	Interval string `json:"interval,omitempty"`
	Fidelity int    `json:"fidelity,omitempty"`
	StartTS  int64  `json:"start_ts,omitempty"`
	EndTS    int64  `json:"end_ts,omitempty"`
}
