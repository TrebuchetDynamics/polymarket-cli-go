package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type OrderBook struct {
	Market string  `json:"market"`
	Bids   []Level `json:"bids"`
	Asks   []Level `json:"asks"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, http: httpClient}
}

func (c *Client) OrderBook(ctx context.Context, tokenID string) (OrderBook, error) {
	u, err := url.Parse(c.baseURL + "/book")
	if err != nil {
		return OrderBook{}, err
	}
	q := u.Query()
	q.Set("token_id", tokenID)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return OrderBook{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return OrderBook{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return OrderBook{}, fmt.Errorf("clob status %d", resp.StatusCode)
	}
	var book OrderBook
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return OrderBook{}, err
	}
	return book, nil
}
