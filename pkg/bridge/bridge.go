// Package bridge provides read-only Bridge API readiness checks.
// Stolen from ybina/polymarket-go/client/bridge/bridge.go.
// Base URL: https://bridge.polymarket.com
package bridge

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

const defaultBridgeBaseURL = "https://bridge.polymarket.com"

// Client provides read-only Bridge API access.
type Client struct {
	transport *transport.Client
}

func NewClient(baseURL string, tc *transport.Client) *Client {
	if baseURL == "" {
		baseURL = defaultBridgeBaseURL
	}
	if tc == nil {
		tc = transport.New(nil, transport.DefaultConfig(baseURL))
	}
	return &Client{transport: tc}
}

// --- Types ---

type DepositAddress struct {
	EVM string `json:"evm"`
	SVM string `json:"svm"`
	BTC string `json:"btc"`
}

type CreateDepositAddressResponse struct {
	Address DepositAddress `json:"address"`
	Note    string         `json:"note"`
}

type TokenInfo struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

type SupportedAsset struct {
	ChainID        string    `json:"chainId"`
	ChainName      string    `json:"chainName"`
	Token          TokenInfo `json:"token"`
	MinCheckoutUsd float64   `json:"minCheckoutUsd"`
}

type SupportedAssetsResponse struct {
	SupportedAssets []SupportedAsset `json:"supportedAssets"`
}

type DepositTransaction struct {
	FromChainID        string `json:"fromChainId"`
	FromTokenAddress   string `json:"fromTokenAddress"`
	FromAmountBaseUnit string `json:"fromAmountBaseUnit"`
	ToChainID          string `json:"toChainId"`
	ToTokenAddress     string `json:"toTokenAddress"`
	TxHash             string `json:"txHash,omitempty"`
	CreatedTimeMs      int64  `json:"createdTimeMs,omitempty"`
	Status             string `json:"status"`
}

type DepositStatusResponse struct {
	Transactions []DepositTransaction `json:"transactions"`
}

type QuoteRequest struct {
	FromAmountBaseUnit string `json:"fromAmountBaseUnit"`
	FromChainID        string `json:"fromChainId"`
	FromTokenAddress   string `json:"fromTokenAddress"`
	RecipientAddress   string `json:"recipientAddress"`
	ToChainID          string `json:"toChainId"`
	ToTokenAddress     string `json:"toTokenAddress"`
}

type FeeBreakdown struct {
	AppFeeLabel     string  `json:"appFeeLabel"`
	AppFeePercent   float64 `json:"appFeePercent"`
	AppFeeUsd       float64 `json:"appFeeUsd"`
	FillCostPercent float64 `json:"fillCostPercent"`
	FillCostUsd     float64 `json:"fillCostUsd"`
	GasUsd          float64 `json:"gasUsd"`
	MaxSlippage     float64 `json:"maxSlippage"`
	MinReceived     float64 `json:"minReceived"`
	SwapImpact      float64 `json:"swapImpact"`
	SwapImpactUsd   float64 `json:"swapImpactUsd"`
	TotalImpact     float64 `json:"totalImpact"`
	TotalImpactUsd  float64 `json:"totalImpactUsd"`
}

type QuoteResponse struct {
	EstCheckoutTimeMs  int64        `json:"estCheckoutTimeMs"`
	EstFeeBreakdown    FeeBreakdown `json:"estFeeBreakdown"`
	EstInputUsd        float64      `json:"estInputUsd"`
	EstOutputUsd       float64      `json:"estOutputUsd"`
	EstToTokenBaseUnit string       `json:"estToTokenBaseUnit"`
	QuoteID            string       `json:"quoteId"`
}

// --- Methods ---

func (c *Client) CreateDepositAddress(ctx context.Context, address string) (*CreateDepositAddressResponse, error) {
	var result CreateDepositAddressResponse
	if err := c.transport.Post(ctx, "/deposit", map[string]string{"address": address}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetSupportedAssets(ctx context.Context) (*SupportedAssetsResponse, error) {
	var result SupportedAssetsResponse
	if err := c.transport.Get(ctx, "/supported-assets", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetDepositStatus(ctx context.Context, depositAddress string) (*DepositStatusResponse, error) {
	var result DepositStatusResponse
	if err := c.transport.Get(ctx, "/status/"+depositAddress, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetQuote(ctx context.Context, req QuoteRequest) (*QuoteResponse, error) {
	var result QuoteResponse
	if err := c.transport.Post(ctx, "/quote", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
