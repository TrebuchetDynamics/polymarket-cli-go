package clob

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xelhaku/polymarket-cli-go/pkg/auth"
)

type Client struct {
	host      string
	chainID   int
	signer    *auth.Signer
	creds     *Credentials
	httpClient *http.Client
}

type Credentials struct {
	Key        string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

type Order struct {
	TokenID   string  `json:"token_id"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Side      string  `json:"side"`
	OrderType string  `json:"order_type"`
}

type MarketOrder struct {
	TokenID   string  `json:"token_id"`
	Amount    float64 `json:"amount"`
	Side      string  `json:"side"`
	OrderType string  `json:"order_type"`
}

type OrderResponse struct {
	Success bool   `json:"success"`
	OrderID string `json:"orderID"`
	Status  string `json:"status"`
	Error   string `json:"error"`
}

type BalanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances"`
}

type BookResponse struct {
	TokenID string     `json:"token_id"`
	Bids    [][]string `json:"bids"`
	Asks    [][]string `json:"asks"`
}

func NewClient(host string, chainID int, signer *auth.Signer) *Client {
	if host == "" {
		host = "https://clob.polymarket.com"
	}
	return &Client{
		host:       host,
		chainID:    chainID,
		signer:     signer,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) SetCredentials(creds *Credentials) {
	c.creds = creds
}

func (c *Client) CreateOrDeriveAPIKey(ctx context.Context) (*Credentials, error) {
	if c.signer == nil {
		return nil, fmt.Errorf("signer required for API key creation")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timestamp + "\n" + timestamp + "\nPOST\n/auth/api-key\n\n"
	
	sig, err := c.signMessage(message)
	if err != nil {
		return nil, fmt.Errorf("sign message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.host+"/auth/api-key", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("POLY_ADDRESS", c.signer.Address())
	req.Header.Set("POLY_SIGNATURE", sig)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_NONCE", timestamp)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api key creation failed: %s", string(body))
	}

	var creds Credentials
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	c.creds = &creds
	return &creds, nil
}

func (c *Client) GetBalance(ctx context.Context, assetType string) (*BalanceResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", 
		fmt.Sprintf("%s/balance?asset_type=%s", c.host, assetType), nil)
	if err != nil {
		return nil, err
	}

	if err := c.signRequest(req); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("balance query failed: %s", string(body))
	}

	var bal BalanceResponse
	if err := json.Unmarshal(body, &bal); err != nil {
		return nil, err
	}
	return &bal, nil
}

func (c *Client) GetOrderBook(ctx context.Context, tokenID string) (*BookResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/book?token_id=%s", c.host, tokenID), nil)
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
		return nil, fmt.Errorf("book query failed: %s", string(body))
	}

	var book BookResponse
	if err := json.Unmarshal(body, &book); err != nil {
		return nil, err
	}
	return &book, nil
}

func (c *Client) CreateOrder(ctx context.Context, order Order) (*OrderResponse, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/order", c.host), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if err := c.signRequest(req); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("order failed (%d): %s", resp.StatusCode, string(body))
	}

	var result OrderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateMarketOrder(ctx context.Context, order MarketOrder) (*OrderResponse, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/order", c.host), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if err := c.signRequest(req); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("market order failed (%d): %s", resp.StatusCode, string(body))
	}

	var result OrderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) signRequest(req *http.Request) error {
	if c.creds == nil {
		return fmt.Errorf("credentials not set")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timestamp + req.Method + req.URL.Path
	if req.URL.RawQuery != "" {
		message += "?" + req.URL.RawQuery
	}
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(body))
		message += string(body)
	}

	h := hmac.New(sha256.New, []byte(c.creds.Secret))
	h.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	req.Header.Set("POLY_ADDRESS", c.signer.Address())
	req.Header.Set("POLY_SIGNATURE", signature)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_API_KEY", c.creds.Key)
	req.Header.Set("POLY_PASSPHRASE", c.creds.Passphrase)
	return nil
}

func (c *Client) signMessage(message string) (string, error) {
	hash := crypto.Keccak256Hash([]byte(message))
	sig, err := crypto.Sign(hash.Bytes(), c.signer.PrivateKey())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}
