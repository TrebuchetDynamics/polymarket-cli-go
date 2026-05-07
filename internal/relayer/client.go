package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

const depositWalletFactoryAddr = "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"

// Client provides access to the Polymarket builder relayer API.
type Client struct {
	transport *transport.Client
	auth      auth.BuilderConfig
	chainID   int64
}

// New creates a relayer Client. bc must contain valid builder credentials
// (BUILDER_API_KEY / BUILDER_SECRET / BUILDER_PASSPHRASE).
func New(baseURL string, bc auth.BuilderConfig, chainID int64) (*Client, error) {
	if !bc.Valid() {
		return nil, fmt.Errorf("relayer: builder credentials are required (key, secret, passphrase)")
	}
	if chainID == 0 {
		chainID = 137
	}
	c := &Client{
		transport: transport.New(nil, transport.DefaultConfig(baseURL)),
		auth:      bc,
		chainID:   chainID,
	}
	return c, nil
}

func (c *Client) buildAuthHeaders(method, path string, body *string) (map[string]string, error) {
	return auth.BuildBuilderHeaders(&c.auth, time.Now().Unix(), method, path, body)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("relayer: marshal request: %w", err)
	}
	compactBody := auth.CompactJSON(string(bodyBytes))
	headers, err := c.buildAuthHeaders(http.MethodPost, path, &compactBody)
	if err != nil {
		return fmt.Errorf("relayer: build auth headers: %w", err)
	}
	return c.transport.PostWithHeaders(ctx, path, body, headers, result)
}

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	headers, err := c.buildAuthHeaders(http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("relayer: build auth headers: %w", err)
	}
	return c.transport.GetWithHeaders(ctx, path, headers, result)
}

// SubmitWalletCreate deploys a new deposit wallet via WALLET-CREATE.
func (c *Client) SubmitWalletCreate(ctx context.Context, ownerAddress string) (*RelayerTransaction, error) {
	ownerAddress = strings.TrimSpace(ownerAddress)
	if ownerAddress == "" {
		return nil, fmt.Errorf("relayer: owner address is required for WALLET-CREATE")
	}
	req := WalletCreateRequest{
		Type: "WALLET-CREATE",
		From: ownerAddress,
		To:   depositWalletFactoryAddr,
	}
	var tx RelayerTransaction
	if err := c.post(ctx, "/submit", req, &tx); err != nil {
		return nil, fmt.Errorf("relayer: WALLET-CREATE: %w", err)
	}
	return &tx, nil
}

// SubmitWalletBatch submits a signed deposit wallet batch to the relayer.
func (c *Client) SubmitWalletBatch(ctx context.Context, ownerAddress, walletAddress, nonce, signature, deadline string, calls []DepositWalletCall) (*RelayerTransaction, error) {
	ownerAddress = strings.TrimSpace(ownerAddress)
	walletAddress = strings.TrimSpace(walletAddress)
	if ownerAddress == "" || walletAddress == "" {
		return nil, fmt.Errorf("relayer: owner and wallet addresses are required")
	}
	if len(calls) == 0 {
		return nil, fmt.Errorf("relayer: at least one call is required")
	}
	req := WalletBatchRequest{
		Type:      "WALLET",
		From:      ownerAddress,
		To:        depositWalletFactoryAddr,
		Nonce:     nonce,
		Signature: signature,
		DepositWalletParams: depositWalletParams{
			DepositWallet: walletAddress,
			Deadline:      deadline,
			Calls:         calls,
		},
	}
	var tx RelayerTransaction
	if err := c.post(ctx, "/submit", req, &tx); err != nil {
		return nil, fmt.Errorf("relayer: WALLET batch: %w", err)
	}
	return &tx, nil
}

// GetNonce fetches the current WALLET nonce for the owner address.
func (c *Client) GetNonce(ctx context.Context, ownerAddress string) (string, error) {
	ownerAddress = strings.TrimSpace(ownerAddress)
	if ownerAddress == "" {
		return "", fmt.Errorf("relayer: owner address is required for nonce")
	}
	path := fmt.Sprintf("/nonce?address=%s&type=WALLET", ownerAddress)
	var resp NonceResponse
	if err := c.get(ctx, path, &resp); err != nil {
		return "", fmt.Errorf("relayer: get nonce: %w", err)
	}
	if strings.TrimSpace(resp.Nonce) == "" {
		return "", fmt.Errorf("relayer: empty nonce response")
	}
	return resp.Nonce, nil
}

// GetTransaction polls for a single transaction by ID.
func (c *Client) GetTransaction(ctx context.Context, txID string) (*RelayerTransaction, error) {
	txID = strings.TrimSpace(txID)
	if txID == "" {
		return nil, fmt.Errorf("relayer: transaction ID is required")
	}
	path := fmt.Sprintf("/transaction?id=%s", txID)
	var tx RelayerTransaction
	if err := c.get(ctx, path, &tx); err != nil {
		return nil, fmt.Errorf("relayer: get transaction: %w", err)
	}
	return &tx, nil
}

// PollTransaction polls the relayer until the transaction reaches a
// terminal state or ctx is cancelled. maxAttempts * interval controls the
// polling window.
func (c *Client) PollTransaction(ctx context.Context, txID string, maxAttempts int, interval time.Duration) (*RelayerTransaction, error) {
	if maxAttempts <= 0 {
		maxAttempts = 50
	}
	if interval <= 0 {
		interval = 2 * time.Second
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		tx, err := c.GetTransaction(ctx, txID)
		if err != nil {
			return nil, err
		}
		state := RelayerTransactionState(tx.State)
		if state.IsTerminal() {
			if !state.IsSuccess() {
				return tx, fmt.Errorf("relayer: transaction %s reached terminal state %s", txID, tx.State)
			}
			return tx, nil
		}
		select {
		case <-ctx.Done():
			return tx, ctx.Err()
		case <-time.After(interval):
		}
	}
	return nil, fmt.Errorf("relayer: timed out waiting for transaction %s after %d attempts", txID, maxAttempts)
}

// IsDeployed checks whether a deposit wallet has been deployed for the
// given owner address.
func (c *Client) IsDeployed(ctx context.Context, ownerAddress string) (bool, error) {
	ownerAddress = strings.TrimSpace(ownerAddress)
	if ownerAddress == "" {
		return false, fmt.Errorf("relayer: owner address is required")
	}
	path := fmt.Sprintf("/deployed?address=%s", ownerAddress)
	var resp DeployedResponse
	if err := c.get(ctx, path, &resp); err != nil {
		return false, fmt.Errorf("relayer: deployed check: %w", err)
	}
	return resp.Deployed, nil
}
