// Package bridge is a client for the Polymarket Bridge API — supported
// assets, deposit addresses, deposit-status polling, and quotes.
//
// Use bridge to discover which assets can be bridged into Polymarket and
// to surface a deposit address for an EOA. The client is HTTP-only and
// performs no signing; it is safe to use in read-only mode.
//
// When not to use this package:
//   - For on-chain transfers — use a Polygon RPC client directly.
//   - For order placement — see the polygolem CLOB surface.
//
// Stability: the Client constructor, methods, and request/response types
// are part of the polygolem public SDK and follow semver. Pass a nil
// transport to NewClient to use the package default; advanced callers may
// supply their own.
package bridge

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

const defaultBridgeBaseURL = "https://bridge.polymarket.com"

// Client provides read-only access to the Polymarket Bridge API.
// Construct via NewClient. Methods are safe for concurrent use.
type Client struct {
	transport *transport.Client
}

// NewClient returns a Bridge API client.
// If baseURL is empty, the production Bridge URL is used.
// If tc is nil, a default transport with retry and rate limiting is
// constructed.
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

// DepositAddress carries the per-chain deposit addresses returned by the
// Bridge for a given Polymarket account.
type DepositAddress struct {
	EVM string `json:"evm"`
	SVM string `json:"svm"`
	BTC string `json:"btc"`
}

// CreateDepositAddressResponse is the response shape for POST /deposit.
type CreateDepositAddressResponse struct {
	Address DepositAddress `json:"address"`
	Note    string         `json:"note"`
}

// TokenInfo describes one token (name, symbol, address, decimals) as
// reported by the Bridge.
type TokenInfo struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

// SupportedAsset is one entry in the Bridge's supported-assets list,
// pairing a chain with the token usable as deposit collateral.
type SupportedAsset struct {
	ChainID        string    `json:"chainId"`
	ChainName      string    `json:"chainName"`
	Token          TokenInfo `json:"token"`
	MinCheckoutUsd float64   `json:"minCheckoutUsd"`
}

// SupportedAssetsResponse is the response shape for GET /supported-assets.
type SupportedAssetsResponse struct {
	SupportedAssets []SupportedAsset `json:"supportedAssets"`
}

// DepositTransaction describes one deposit attempt observed by the Bridge.
// Status is a Bridge-defined string; clients should treat unknown values
// as opaque.
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

// DepositStatusResponse is the response shape for GET /status/{address}.
type DepositStatusResponse struct {
	Transactions []DepositTransaction `json:"transactions"`
}

// QuoteRequest is the input to GetQuote — the source token and amount,
// recipient, and target token on the Polymarket side.
type QuoteRequest struct {
	FromAmountBaseUnit string `json:"fromAmountBaseUnit"`
	FromChainID        string `json:"fromChainId"`
	FromTokenAddress   string `json:"fromTokenAddress"`
	RecipientAddress   string `json:"recipientAddress"`
	ToChainID          string `json:"toChainId"`
	ToTokenAddress     string `json:"toTokenAddress"`
}

// FeeBreakdown enumerates the fee components a Bridge quote includes.
// All percent fields are expressed as a fraction (0.01 = 1%).
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

// QuoteResponse is the response shape for POST /quote — estimated input
// and output USD, an estimated time, the fee breakdown, and a quote ID
// that the caller must echo when accepting the quote.
type QuoteResponse struct {
	EstCheckoutTimeMs  int64        `json:"estCheckoutTimeMs"`
	EstFeeBreakdown    FeeBreakdown `json:"estFeeBreakdown"`
	EstInputUsd        float64      `json:"estInputUsd"`
	EstOutputUsd       float64      `json:"estOutputUsd"`
	EstToTokenBaseUnit string       `json:"estToTokenBaseUnit"`
	QuoteID            string       `json:"quoteId"`
}

// --- Methods ---

// CreateDepositAddress requests the Bridge mint a deposit address for the
// given Polymarket-side address. The Bridge returns a per-chain address
// set; only one of EVM/SVM/BTC is typically populated per request.
func (c *Client) CreateDepositAddress(ctx context.Context, address string) (*CreateDepositAddressResponse, error) {
	var result CreateDepositAddressResponse
	if err := c.transport.Post(ctx, "/deposit", map[string]string{"address": address}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSupportedAssets returns the assets the Bridge currently accepts as
// deposit collateral.
func (c *Client) GetSupportedAssets(ctx context.Context) (*SupportedAssetsResponse, error) {
	var result SupportedAssetsResponse
	if err := c.transport.Get(ctx, "/supported-assets", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDepositStatus polls the Bridge for outstanding and recent deposit
// transactions targeting depositAddress.
func (c *Client) GetDepositStatus(ctx context.Context, depositAddress string) (*DepositStatusResponse, error) {
	var result DepositStatusResponse
	if err := c.transport.Get(ctx, "/status/"+depositAddress, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetQuote asks the Bridge to price a deposit move described by req.
// The returned QuoteID is the handle the caller will echo in a follow-up
// accept call.
func (c *Client) GetQuote(ctx context.Context, req QuoteRequest) (*QuoteResponse, error) {
	var result QuoteResponse
	if err := c.transport.Post(ctx, "/quote", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
