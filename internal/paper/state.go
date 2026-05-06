package paper

import "fmt"

type Order struct {
	MarketID string  `json:"market_id"`
	TokenID  string  `json:"token_id"`
	Price    float64 `json:"price"`
	Size     float64 `json:"size"`
}

type Fill struct {
	MarketID string  `json:"market_id"`
	TokenID  string  `json:"token_id"`
	Price    float64 `json:"price"`
	Size     float64 `json:"size"`
	Live     bool    `json:"live"`
}

type Position struct {
	TokenID string  `json:"token_id"`
	Size    float64 `json:"size"`
	Cost    float64 `json:"cost"`
}

type State struct {
	Currency  string              `json:"currency"`
	Cash      float64             `json:"cash"`
	Positions map[string]Position `json:"positions"`
	Fills     []Fill              `json:"fills"`
}

func NewState(currency string, cash float64) *State {
	return &State{Currency: currency, Cash: cash, Positions: map[string]Position{}}
}

func (s *State) Buy(order Order) (Fill, error) {
	cost := order.Price * order.Size
	if cost > s.Cash {
		return Fill{}, fmt.Errorf("insufficient paper cash")
	}
	s.Cash -= cost
	pos := s.Positions[order.TokenID]
	pos.TokenID = order.TokenID
	pos.Size += order.Size
	pos.Cost += cost
	s.Positions[order.TokenID] = pos
	fill := Fill{MarketID: order.MarketID, TokenID: order.TokenID, Price: order.Price, Size: order.Size, Live: false}
	s.Fills = append(s.Fills, fill)
	return fill, nil
}
