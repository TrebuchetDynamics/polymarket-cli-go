package gamma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Market struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Question string `json:"question"`
	Active   bool   `json:"active"`
	Closed   bool   `json:"closed"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: httpClient}
}

func (c *Client) ActiveMarkets(ctx context.Context) ([]Market, error) {
	u, err := url.Parse(c.baseURL + "/markets")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("active", "true")
	q.Set("closed", "false")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("gamma status %d", resp.StatusCode)
	}
	var markets []Market
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, err
	}
	return markets, nil
}
