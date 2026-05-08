package clob

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	internalclob "github.com/TrebuchetDynamics/polygolem/internal/clob"
)

// APIKey is a Polymarket CLOB L2 credential.
type APIKey struct {
	Key        string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// BalanceAllowanceParams filters CLOB collateral or conditional-token balance
// and allowance reads. V2 calls always use deposit-wallet signature type 3.
type BalanceAllowanceParams struct {
	Asset     string
	AssetType string
	TokenID   string
}

// BalanceAllowanceResponse is the authenticated CLOB balance/allowance state.
type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances,omitempty"`
	Allowance  string            `json:"allowance,omitempty"`
}

// CreateOrderParams is the public input to CreateLimitOrder.
type CreateOrderParams struct {
	TokenID    string
	Side       string
	Price      string
	Size       string
	OrderType  string
	Expiration string
}

// MarketOrderParams is the public input to CreateMarketOrder.
type MarketOrderParams struct {
	TokenID   string
	Side      string
	Amount    string
	Price     string
	OrderType string
}

// OrderPlacementResponse is the CLOB response after posting a signed order.
type OrderPlacementResponse struct {
	Success            bool     `json:"success"`
	OrderID            string   `json:"orderID"`
	Status             string   `json:"status"`
	MakingAmount       string   `json:"makingAmount,omitempty"`
	TakingAmount       string   `json:"takingAmount,omitempty"`
	ErrorMsg           string   `json:"errorMsg,omitempty"`
	TransactionsHashes []string `json:"transactionsHashes,omitempty"`
	TradeIDs           []string `json:"tradeIDs,omitempty"`
}

// CancelOrdersResponse is returned by CLOB cancellation endpoints.
type CancelOrdersResponse struct {
	Canceled    []string          `json:"canceled"`
	NotCanceled map[string]string `json:"not_canceled"`
}

// CancelMarketParams filters cancel-market requests by condition ID or token ID.
type CancelMarketParams struct {
	Market string
	Asset  string
}

// OrderRecord is an authenticated CLOB order record.
type OrderRecord struct {
	ID              string   `json:"id"`
	Status          string   `json:"status"`
	Owner           string   `json:"owner"`
	Market          string   `json:"market"`
	AssetID         string   `json:"asset_id"`
	Side            string   `json:"side"`
	OriginalSize    string   `json:"original_size"`
	SizeMatched     string   `json:"size_matched"`
	Price           string   `json:"price"`
	Outcome         string   `json:"outcome"`
	OrderType       string   `json:"order_type,omitempty"`
	SignatureType   int      `json:"signature_type,omitempty"`
	CreatedAt       string   `json:"created_at"`
	Expiration      string   `json:"expiration"`
	MakerAddress    string   `json:"maker_address"`
	AssociateTrades []string `json:"associate_trades,omitempty"`
}

// TradeRecord is an authenticated CLOB trade record.
type TradeRecord struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Market          string `json:"market"`
	AssetID         string `json:"asset_id"`
	Side            string `json:"side"`
	Price           string `json:"price"`
	Size            string `json:"size"`
	FeeRateBps      string `json:"fee_rate_bps"`
	Outcome         string `json:"outcome"`
	Owner           string `json:"owner"`
	Builder         string `json:"builder"`
	MatchedAmount   string `json:"matched_amount"`
	TransactionHash string `json:"transaction_hash"`
	CreatedAt       string `json:"created_at"`
	LastUpdated     string `json:"last_updated"`
}

// BuilderFeeKeyRecord represents one CLOB builder-fee key.
type BuilderFeeKeyRecord struct {
	Key        string `json:"key"`
	Secret     string `json:"secret,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

// CreateOrDeriveAPIKey creates new CLOB L2 credentials, falling back to
// deterministic derivation when a key already exists.
func (c *Client) CreateOrDeriveAPIKey(ctx context.Context, privateKey string) (APIKey, error) {
	key, err := c.inner.CreateOrDeriveAPIKey(ctx, privateKey)
	if err != nil {
		return APIKey{}, err
	}
	return apiKeyFromInternal(key), nil
}

// DeriveAPIKey derives existing CLOB L2 credentials.
func (c *Client) DeriveAPIKey(ctx context.Context, privateKey string) (APIKey, error) {
	key, err := c.inner.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return APIKey{}, err
	}
	return apiKeyFromInternal(key), nil
}

// CreateBuilderFeeKey mints a CLOB builder-fee key via L2 auth.
func (c *Client) CreateBuilderFeeKey(ctx context.Context, privateKey string) (APIKey, error) {
	key, err := c.inner.CreateBuilderFeeKey(ctx, privateKey)
	if err != nil {
		return APIKey{}, err
	}
	return apiKeyFromInternal(key), nil
}

// ListBuilderFeeKeys lists CLOB builder-fee keys for the authenticated wallet.
func (c *Client) ListBuilderFeeKeys(ctx context.Context, privateKey string) ([]BuilderFeeKeyRecord, error) {
	rows, err := c.inner.ListBuilderFeeKeys(ctx, privateKey)
	if err != nil {
		return nil, err
	}
	out := make([]BuilderFeeKeyRecord, len(rows))
	for i, row := range rows {
		out[i] = BuilderFeeKeyRecord{
			Key:        row.Key,
			Secret:     row.Secret,
			Passphrase: row.Passphrase,
			CreatedAt:  row.CreatedAt,
		}
	}
	return out, nil
}

// RevokeBuilderFeeKey deletes a CLOB builder-fee key.
func (c *Client) RevokeBuilderFeeKey(ctx context.Context, privateKey, builderKey string) error {
	return c.inner.RevokeBuilderFeeKey(ctx, privateKey, builderKey)
}

// BalanceAllowance returns CLOB collateral or conditional token balance and allowances.
func (c *Client) BalanceAllowance(ctx context.Context, privateKey string, params BalanceAllowanceParams) (*BalanceAllowanceResponse, error) {
	row, err := c.inner.BalanceAllowance(ctx, privateKey, balanceAllowanceParamsToInternal(params))
	if err != nil {
		return nil, err
	}
	return balanceAllowanceFromInternal(row), nil
}

// UpdateBalanceAllowance refreshes the CLOB balance/allowance cache.
func (c *Client) UpdateBalanceAllowance(ctx context.Context, privateKey string, params BalanceAllowanceParams) (*BalanceAllowanceResponse, error) {
	row, err := c.inner.UpdateBalanceAllowance(ctx, privateKey, balanceAllowanceParamsToInternal(params))
	if err != nil {
		return nil, err
	}
	return balanceAllowanceFromInternal(row), nil
}

// ListOrders returns the authenticated user's open CLOB orders.
func (c *Client) ListOrders(ctx context.Context, privateKey string) ([]OrderRecord, error) {
	rows, err := c.inner.ListOrders(ctx, privateKey)
	if err != nil {
		return nil, err
	}
	return orderRecordsFromInternal(rows), nil
}

// Order returns one authenticated CLOB order by order ID.
func (c *Client) Order(ctx context.Context, privateKey, orderID string) (*OrderRecord, error) {
	row, err := c.inner.Order(ctx, privateKey, orderID)
	if err != nil {
		return nil, err
	}
	return orderRecordFromInternal(row), nil
}

// ListTrades returns the authenticated user's CLOB trade history.
func (c *Client) ListTrades(ctx context.Context, privateKey string) ([]TradeRecord, error) {
	rows, err := c.inner.ListTrades(ctx, privateKey)
	if err != nil {
		return nil, err
	}
	return tradeRecordsFromInternal(rows), nil
}

// CancelOrder cancels a single open CLOB order.
func (c *Client) CancelOrder(ctx context.Context, privateKey, orderID string) (*CancelOrdersResponse, error) {
	row, err := c.inner.CancelOrder(ctx, privateKey, orderID)
	if err != nil {
		return nil, err
	}
	return cancelOrdersFromInternal(row), nil
}

// CancelOrders cancels multiple open CLOB orders by order ID.
func (c *Client) CancelOrders(ctx context.Context, privateKey string, orderIDs []string) (*CancelOrdersResponse, error) {
	row, err := c.inner.CancelOrders(ctx, privateKey, orderIDs)
	if err != nil {
		return nil, err
	}
	return cancelOrdersFromInternal(row), nil
}

// CancelAll cancels all open CLOB orders for the authenticated user.
func (c *Client) CancelAll(ctx context.Context, privateKey string) (*CancelOrdersResponse, error) {
	row, err := c.inner.CancelAll(ctx, privateKey)
	if err != nil {
		return nil, err
	}
	return cancelOrdersFromInternal(row), nil
}

// CancelMarket cancels open CLOB orders matching a market or asset filter.
func (c *Client) CancelMarket(ctx context.Context, privateKey string, params CancelMarketParams) (*CancelOrdersResponse, error) {
	row, err := c.inner.CancelMarket(ctx, privateKey, cancelMarketParamsToInternal(params))
	if err != nil {
		return nil, err
	}
	return cancelOrdersFromInternal(row), nil
}

// CreateLimitOrder signs and submits a V2 limit order.
func (c *Client) CreateLimitOrder(ctx context.Context, privateKey string, params CreateOrderParams) (*OrderPlacementResponse, error) {
	row, err := c.inner.CreateLimitOrder(ctx, privateKey, createOrderParamsToInternal(params))
	if err != nil {
		return nil, err
	}
	return orderPlacementFromInternal(row), nil
}

// CreateMarketOrder signs and submits a V2 buy-side market order.
func (c *Client) CreateMarketOrder(ctx context.Context, privateKey string, params MarketOrderParams) (*OrderPlacementResponse, error) {
	row, err := c.inner.CreateMarketOrder(ctx, privateKey, marketOrderParamsToInternal(params))
	if err != nil {
		return nil, err
	}
	return orderPlacementFromInternal(row), nil
}

func apiKeyFromInternal(row auth.APIKey) APIKey {
	return APIKey{
		Key:        row.Key,
		Secret:     row.Secret,
		Passphrase: row.Passphrase,
	}
}

func balanceAllowanceParamsToInternal(params BalanceAllowanceParams) internalclob.BalanceAllowanceParams {
	return internalclob.BalanceAllowanceParams{
		Asset:     params.Asset,
		AssetType: params.AssetType,
		TokenID:   params.TokenID,
	}
}

func balanceAllowanceFromInternal(row *internalclob.BalanceAllowanceResponse) *BalanceAllowanceResponse {
	if row == nil {
		return nil
	}
	return &BalanceAllowanceResponse{
		Balance:    row.Balance,
		Allowances: row.Allowances,
		Allowance:  row.Allowance,
	}
}

func orderRecordsFromInternal(rows []internalclob.OrderRecord) []OrderRecord {
	out := make([]OrderRecord, len(rows))
	for i, row := range rows {
		out[i] = orderRecordValueFromInternal(row)
	}
	return out
}

func orderRecordFromInternal(row *internalclob.OrderRecord) *OrderRecord {
	if row == nil {
		return nil
	}
	out := orderRecordValueFromInternal(*row)
	return &out
}

func orderRecordValueFromInternal(row internalclob.OrderRecord) OrderRecord {
	return OrderRecord{
		ID:              row.ID,
		Status:          row.Status,
		Owner:           row.Owner,
		Market:          row.Market,
		AssetID:         row.AssetID,
		Side:            row.Side,
		OriginalSize:    row.OriginalSize,
		SizeMatched:     row.SizeMatched,
		Price:           row.Price,
		Outcome:         row.Outcome,
		OrderType:       firstNonEmptyString(row.OrderType, row.Type),
		SignatureType:   row.SignatureType,
		CreatedAt:       row.CreatedAt,
		Expiration:      row.Expiration,
		MakerAddress:    row.MakerAddress,
		AssociateTrades: row.AssociateTrades,
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func tradeRecordsFromInternal(rows []internalclob.TradeRecord) []TradeRecord {
	out := make([]TradeRecord, len(rows))
	for i, row := range rows {
		out[i] = TradeRecord{
			ID:              row.ID,
			Status:          row.Status,
			Market:          row.Market,
			AssetID:         row.AssetID,
			Side:            row.Side,
			Price:           row.Price,
			Size:            row.Size,
			FeeRateBps:      row.FeeRateBps,
			Outcome:         row.Outcome,
			Owner:           row.Owner,
			Builder:         row.Builder,
			MatchedAmount:   row.MatchedAmount,
			TransactionHash: row.TransactionHash,
			CreatedAt:       row.CreatedAt,
			LastUpdated:     row.LastUpdated,
		}
	}
	return out
}

func cancelOrdersFromInternal(row *internalclob.CancelOrdersResponse) *CancelOrdersResponse {
	if row == nil {
		return nil
	}
	return &CancelOrdersResponse{
		Canceled:    row.Canceled,
		NotCanceled: row.NotCanceled,
	}
}

func cancelMarketParamsToInternal(params CancelMarketParams) internalclob.CancelMarketParams {
	return internalclob.CancelMarketParams{
		Market: params.Market,
		Asset:  params.Asset,
	}
}

func createOrderParamsToInternal(params CreateOrderParams) internalclob.CreateOrderParams {
	return internalclob.CreateOrderParams{
		TokenID:    params.TokenID,
		Side:       params.Side,
		Price:      params.Price,
		Size:       params.Size,
		OrderType:  params.OrderType,
		Expiration: params.Expiration,
	}
}

func marketOrderParamsToInternal(params MarketOrderParams) internalclob.MarketOrderParams {
	return internalclob.MarketOrderParams{
		TokenID:   params.TokenID,
		Side:      params.Side,
		Amount:    params.Amount,
		Price:     params.Price,
		OrderType: params.OrderType,
	}
}

func orderPlacementFromInternal(row *internalclob.OrderPlacementResponse) *OrderPlacementResponse {
	if row == nil {
		return nil
	}
	return &OrderPlacementResponse{
		Success:            row.Success,
		OrderID:            row.OrderID,
		Status:             row.Status,
		MakingAmount:       row.MakingAmount,
		TakingAmount:       row.TakingAmount,
		ErrorMsg:           row.ErrorMsg,
		TransactionsHashes: row.TransactionsHashes,
		TradeIDs:           row.TradeIDs,
	}
}
