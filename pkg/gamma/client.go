package gamma

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Market struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	ConditionID string `json:"conditionId"`
	Active      bool   `json:"active"`
	Closed      bool   `json:"closed"`
	EndDate     string `json:"endDate"`
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://gamma-api.polymarket.com"
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetActiveMarkets(ctx context.Context) ([]Market, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/events?active=true&closed=false&limit=100", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("markets query failed: %s", string(body))
	}

	var result struct {
		Events []struct {
			Markets []Market `json:"markets"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var markets []Market
	for _, event := range result.Events {
		for _, m := range event.Markets {
			if m.Active && !m.Closed {
				markets = append(markets, m)
			}
		}
	}
	return markets, nil
}
