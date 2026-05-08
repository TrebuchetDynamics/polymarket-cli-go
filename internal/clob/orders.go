package clob

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/ethereum/go-ethereum/common"
	gethmath "github.com/ethereum/go-ethereum/common/math"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	clobExchangeAddress    = "0xE111180000d2663C0091e4f400237545B87B996B" // V2 regular
	negRiskExchangeAddress = "0xe2222d279d744050d28e00520010520000310F59" // V2 neg-risk
	zeroAddress            = "0x0000000000000000000000000000000000000000"
	bytes32Zero            = "0x0000000000000000000000000000000000000000000000000000000000000000"
	signatureTypePoly1271  = 3
)

// Test seams: tests override these to make salt and timestamp deterministic.
var (
	orderSalt = generateOrderSalt
	orderNow  = time.Now
)

type CreateOrderParams struct {
	TokenID       string
	Side          string
	Price         string
	Size          string
	OrderType     string
	SignatureType int
}

type MarketOrderParams struct {
	TokenID       string
	Side          string
	Amount        string
	Price         string
	OrderType     string
	SignatureType int
}

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

// CancelOrdersResponse is returned by cancel order and cancel-all endpoints.
type CancelOrdersResponse struct {
	Canceled    []string          `json:"canceled"`
	NotCanceled map[string]string `json:"not_canceled"`
}

// OrderRecord is a single order as returned by ListOrders.
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
	Type            string   `json:"type"`
	SignatureType   int      `json:"signature_type"`
	CreatedAt       string   `json:"created_at"`
	Expiration      string   `json:"expiration"`
	MakerAddress    string   `json:"maker_address"`
	AssociateTrades []string `json:"associate_trades,omitempty"`
}

// TradeRecord is a single trade as returned by ListTrades.
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

// signedOrderPayload is the CLOB V2 order wire format.
// Differs from V1: taker/nonce/feeRateBps removed, timestamp/metadata/builder added.
// Expiration is in the POST body but NOT in the V2 EIP-712 signed struct.
type signedOrderPayload struct {
	Salt          uint64 `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Side          string `json:"side"`
	Expiration    string `json:"expiration"`
	SignatureType int    `json:"signatureType"`
	Timestamp     string `json:"timestamp"`
	Metadata      string `json:"metadata"`
	Builder       string `json:"builder"`
	Signature     string `json:"signature"`
}

type sendOrderPayload struct {
	Order     signedOrderPayload `json:"order"`
	Owner     string             `json:"owner"`
	OrderType string             `json:"orderType"`
	PostOnly  bool               `json:"postOnly"`
	DeferExec bool               `json:"deferExec"`
}

type orderDraft struct {
	tokenID       *big.Int
	side          string
	makerAmount   string
	takerAmount   string
	signatureType int
	orderType     string
}

func (c *Client) CreateLimitOrder(ctx context.Context, privateKey string, params CreateOrderParams) (*OrderPlacementResponse, error) {
	side, err := normalizeOrderSide(params.Side)
	if err != nil {
		return nil, err
	}
	tokenID, err := parseTokenID(params.TokenID)
	if err != nil {
		return nil, err
	}
	price, err := parseRat(params.Price, "price")
	if err != nil {
		return nil, err
	}
	size, err := parseRat(params.Size, "size")
	if err != nil {
		return nil, err
	}
	if price.Sign() <= 0 || size.Sign() <= 0 {
		return nil, fmt.Errorf("price and size must be positive")
	}
	tick, err := c.TickSize(ctx, params.TokenID)
	if err != nil {
		return nil, fmt.Errorf("tick size lookup failed: %w", err)
	}
	if err := validatePriceScale(price, params.Price, tick); err != nil {
		return nil, err
	}

	makerAmount, takerAmount := limitFixedAmounts(side, price, size)
	draft := orderDraft{
		tokenID:       tokenID,
		side:          side,
		makerAmount:   makerAmount,
		takerAmount:   takerAmount,
		signatureType: params.SignatureType,
		orderType:     normalizeOrderType(params.OrderType, "GTC"),
	}
	return c.signAndPostOrder(ctx, privateKey, draft)
}

func (c *Client) ListOrders(ctx context.Context, privateKey string) ([]OrderRecord, error) {
	var result []OrderRecord
	if err := c.authenticatedL2GET(ctx, privateKey, "/data/orders", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListTrades(ctx context.Context, privateKey string) ([]TradeRecord, error) {
	var result []TradeRecord
	if err := c.authenticatedL2GET(ctx, privateKey, "/data/trades", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CancelOrder(ctx context.Context, privateKey, orderID string) (*CancelOrdersResponse, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	key, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("derive api key: %w", err)
	}
	body := map[string]string{"orderID": orderID}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	compactBody := auth.CompactJSON(string(bodyBytes))
	headers, err := c.l2Headers(privateKey, &key, http.MethodDelete, "/order", &compactBody)
	if err != nil {
		return nil, err
	}
	var result CancelOrdersResponse
	if err := c.transport.DeleteWithHeaders(ctx, "/order", body, headers, &result); err != nil {
		return nil, fmt.Errorf("cancel order: %w", err)
	}
	return &result, nil
}

func (c *Client) CancelAll(ctx context.Context, privateKey string) (*CancelOrdersResponse, error) {
	key, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("derive api key: %w", err)
	}
	headers, err := c.l2Headers(privateKey, &key, http.MethodDelete, "/cancel-all", nil)
	if err != nil {
		return nil, err
	}
	var result CancelOrdersResponse
	if err := c.transport.DeleteWithHeaders(ctx, "/cancel-all", nil, headers, &result); err != nil {
		return nil, fmt.Errorf("cancel all: %w", err)
	}
	return &result, nil
}

func (c *Client) authenticatedRawGET(ctx context.Context, privateKey string, path string) (json.RawMessage, error) {
	key, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("derive api key: %w", err)
	}
	headers, err := c.l2Headers(privateKey, &key, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := c.transport.GetWithHeaders(ctx, path, headers, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) authenticatedL2GET(ctx context.Context, privateKey string, path string, result interface{}) error {
	key, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return fmt.Errorf("derive api key: %w", err)
	}
	headers, err := c.l2Headers(privateKey, &key, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return c.transport.GetWithHeaders(ctx, path, headers, result)
}

func (c *Client) CreateMarketOrder(ctx context.Context, privateKey string, params MarketOrderParams) (*OrderPlacementResponse, error) {
	side, err := normalizeOrderSide(params.Side)
	if err != nil {
		return nil, err
	}
	if side != "BUY" {
		return nil, fmt.Errorf("market-order amount is currently supported for BUY only")
	}
	tokenID, err := parseTokenID(params.TokenID)
	if err != nil {
		return nil, err
	}
	amount, err := parseRat(params.Amount, "amount")
	if err != nil {
		return nil, err
	}
	if amount.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	tick, err := c.TickSize(ctx, params.TokenID)
	if err != nil {
		return nil, fmt.Errorf("tick size lookup failed: %w", err)
	}
	tickScale := decimalPlaces(firstNonEmpty(tick.TickSize, tick.MinimumTickSize))
	var price *big.Rat
	if strings.TrimSpace(params.Price) != "" {
		price, err = parseRat(params.Price, "price")
		if err != nil {
			return nil, err
		}
	} else {
		price, err = c.marketOrderPrice(ctx, params.TokenID, side, amount, normalizeOrderType(params.OrderType, "FOK"))
		if err != nil {
			return nil, err
		}
	}
	if price.Sign() <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}

	taker := new(big.Rat).Quo(amount, price)
	taker = truncateRat(taker, tickScale+2)
	if err := validateMinimumOrderSize(taker, tick); err != nil {
		return nil, err
	}
	draft := orderDraft{
		tokenID:       tokenID,
		side:          side,
		makerAmount:   fixedDecimal(amount, 6),
		takerAmount:   fixedDecimal(taker, 6),
		signatureType: params.SignatureType,
		orderType:     normalizeOrderType(params.OrderType, "FOK"),
	}
	return c.signAndPostOrder(ctx, privateKey, draft)
}

func (c *Client) signAndPostOrder(ctx context.Context, privateKey string, draft orderDraft) (*OrderPlacementResponse, error) {
	signer, err := auth.NewPrivateKeySigner(privateKey, polygonChainID)
	if err != nil {
		return nil, err
	}
	key, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("derive api key: %w", err)
	}
	nr, err := c.NegRisk(ctx, draft.tokenID.String())
	if err != nil {
		return nil, fmt.Errorf("neg-risk lookup: %w", err)
	}
	unsigned, err := buildSignedOrderPayload(signer, draft, orderNow(), nr.NegRisk)
	if err != nil {
		return nil, err
	}
	payload := sendOrderPayload{
		Order:     unsigned,
		Owner:     key.Key,
		OrderType: draft.orderType,
		PostOnly:  false,
		DeferExec: false,
	}
	return c.postOrder(ctx, privateKey, &key, payload, draft.orderType)
}

func (c *Client) postOrder(ctx context.Context, privateKey string, key *auth.APIKey, payload interface{}, orderType string) (*OrderPlacementResponse, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)
	headers, err := c.l2Headers(privateKey, key, http.MethodPost, "/order", &body)
	if err != nil {
		return nil, err
	}
	var result OrderPlacementResponse
	if err := c.transport.PostWithHeaders(ctx, "/order", payload, headers, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func signCLOBOrder(signer *auth.PrivateKeySigner, order signedOrderPayload, negRisk bool) (string, error) {
	typed := buildOrderTypedData(order, negRisk)
	sig, err := signer.SignEIP712(typed)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("0x%x", sig), nil
}

// buildOrderTypedData builds the apitypes.TypedData for a V2 order.
// Shared by signCLOBOrder and wrapPOLY1271Signature.
func buildOrderTypedData(order signedOrderPayload, negRisk bool) apitypes.TypedData {
	sideInt := int64(0)
	if order.Side == "SELL" {
		sideInt = 1
	}
	verifyingContract := clobExchangeAddress
	if negRisk {
		verifyingContract = negRiskExchangeAddress
	}
	tokenID, _ := new(big.Int).SetString(order.TokenID, 10)
	makerAmount, _ := new(big.Int).SetString(order.MakerAmount, 10)
	takerAmount, _ := new(big.Int).SetString(order.TakerAmount, 10)
	timestamp, _ := new(big.Int).SetString(order.Timestamp, 10)
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
				{Name: "timestamp", Type: "uint256"},
				{Name: "metadata", Type: "bytes32"},
				{Name: "builder", Type: "bytes32"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "Polymarket CTF Exchange",
			Version:           "2",
			ChainId:           auth.EIP712ChainID(polygonChainID),
			VerifyingContract: verifyingContract,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          (*gethmath.HexOrDecimal256)(new(big.Int).SetUint64(order.Salt)),
			"maker":         common.HexToAddress(order.Maker).Hex(),
			"signer":        common.HexToAddress(order.Signer).Hex(),
			"tokenId":       (*gethmath.HexOrDecimal256)(tokenID),
			"makerAmount":   (*gethmath.HexOrDecimal256)(makerAmount),
			"takerAmount":   (*gethmath.HexOrDecimal256)(takerAmount),
			"side":          (*gethmath.HexOrDecimal256)(big.NewInt(sideInt)),
			"signatureType": (*gethmath.HexOrDecimal256)(big.NewInt(int64(order.SignatureType))),
			"timestamp":     (*gethmath.HexOrDecimal256)(timestamp),
			"metadata":      common.HexToHash(order.Metadata).Hex(),
			"builder":       common.HexToHash(order.Builder).Hex(),
		},
	}
}

// wrapPOLY1271Signature produces the 636-char ERC-7739 TypedDataSign wrapped
// signature used when signatureType=3 for Polymarket V2 orders.
//
// Layout: innerSig(65) || appDomainSep(32) || contents(32) || contentsType(186) || uint16BE(186)
// = 317 bytes = 634 hex chars + "0x" = 636 chars total.
func wrapPOLY1271Signature(signer *auth.PrivateKeySigner, depositWallet string, orderTypedData apitypes.TypedData) (string, error) {
	// contentsType is the V2 Order type string — must be exactly 186 bytes.
	const contentsType = "Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"
	if len(contentsType) != 186 {
		return "", fmt.Errorf("internal: contentsType length %d != 186", len(contentsType))
	}

	// 1. Extract appDomainSep and contents from the V2 order typed data.
	//    TypedDataAndHash returns (hash, rawData, err) where rawData is a string:
	//    \x19\x01 (2B) || domainSep (32B) || structHash (32B).
	//
	//    Per docs.polymarket.com/trading/deposit-wallets, the user signs a nested
	//    TypedDataSign payload UNDER the CTF Exchange V2 domain — i.e. the OUTER
	//    EIP-712 domain in the keccak256(0x1901 || domSep || hashStruct(...)) is
	//    the Exchange domain (regular or neg-risk per market). The Exchange
	//    domSep is exactly what TypedDataAndHash extracts at rawData[2:34].
	_, rawDataStr, err := apitypes.TypedDataAndHash(orderTypedData)
	if err != nil {
		return "", fmt.Errorf("hash order typed data: %w", err)
	}
	rawData := []byte(rawDataStr)
	if len(rawData) != 66 {
		return "", fmt.Errorf("unexpected rawData length %d", len(rawData))
	}
	appDomainSep := rawData[2:34] // CTF Exchange V2 domain separator — used as OUTER
	contents := rawData[34:66]    // hashStruct(Order)

	// 2. TypedDataSign typehash.
	typeHashStr := "TypedDataSign(Order contents,string name,string version,uint256 chainId,address verifyingContract,bytes32 salt)" + contentsType
	typedDataSignTypehash := ethcrypto.Keccak256([]byte(typeHashStr))

	// 3. hashStruct(TypedDataSign{contents, DepositWallet inline domain values}).
	//    Per docs.polymarket.com/trading/deposit-wallets, the INNER struct's domain
	//    fields describe the WALLET (the contract that will validate via
	//    isValidSignature), NOT the app domain. The app/Exchange identity is
	//    encoded by the OUTER domSep below.
	//
	//      name              = "DepositWallet"
	//      version           = "1"
	//      chainId           = 137 (Polygon mainnet)
	//      verifyingContract = the deposit wallet address
	//      salt              = bytes32(0)
	dwNameHash := ethcrypto.Keccak256([]byte("DepositWallet"))
	dwVerHash := ethcrypto.Keccak256([]byte("1"))
	dwChainIDBytes := common.LeftPadBytes(big.NewInt(polygonChainID).Bytes(), 32)
	dwAddrBytes := common.LeftPadBytes(common.HexToAddress(depositWallet).Bytes(), 32)
	dwSaltBytes := make([]byte, 32) // zeros

	tdsStruct := ethcrypto.Keccak256(
		typedDataSignTypehash,
		contents,
		dwNameHash,
		dwVerHash,
		dwChainIDBytes,
		dwAddrBytes,
		dwSaltBytes,
	)

	// 4. finalHash = keccak256(0x1901 || appDomainSep || tdsStruct).
	//    Outer is the CTF Exchange V2 domain — the app the user is authorizing.
	finalHashInput := make([]byte, 0, 66)
	finalHashInput = append(finalHashInput, 0x19, 0x01)
	finalHashInput = append(finalHashInput, appDomainSep...)
	finalHashInput = append(finalHashInput, tdsStruct...)
	finalHashSum := ethcrypto.Keccak256(finalHashInput)
	var finalHash [32]byte
	copy(finalHash[:], finalHashSum)

	// 5. ECDSA-sign the finalHash with the EOA private key.
	innerSig, err := signer.SignRaw(finalHash)
	if err != nil {
		return "", fmt.Errorf("sign inner: %w", err)
	}

	// 6. Assemble: innerSig(65) || appDomainSep(32) || contents(32) || contentsType(186) || uint16BE(186).
	var lenBuf [2]byte
	binary.BigEndian.PutUint16(lenBuf[:], uint16(len(contentsType)))
	sig := make([]byte, 0, 317)
	sig = append(sig, innerSig...)
	sig = append(sig, appDomainSep...)
	sig = append(sig, contents...)
	sig = append(sig, []byte(contentsType)...)
	sig = append(sig, lenBuf[:]...)
	return "0x" + hex.EncodeToString(sig), nil
}

// buildSignedOrderPayload constructs a signed V2 order payload from a draft.
func buildSignedOrderPayload(signer *auth.PrivateKeySigner, draft orderDraft, ts time.Time, negRisk bool) (signedOrderPayload, error) {
	salt, err := orderSalt()
	if err != nil {
		return signedOrderPayload{}, err
	}
	maker, err := auth.MakerAddressForSignatureType(signer.Address(), polygonChainID, draft.signatureType)
	if err != nil {
		return signedOrderPayload{}, err
	}
	orderSigner := signer.Address()
	if draft.signatureType == signatureTypePoly1271 {
		orderSigner = maker
	}
	payload := signedOrderPayload{
		Salt:          salt,
		Maker:         maker,
		Signer:        orderSigner,
		TokenID:       draft.tokenID.String(),
		MakerAmount:   draft.makerAmount,
		TakerAmount:   draft.takerAmount,
		Side:          draft.side,
		Expiration:    "0",
		SignatureType: draft.signatureType,
		Timestamp:     fmt.Sprintf("%d", ts.UnixMilli()),
		Metadata:      bytes32Zero,
		Builder:       bytes32Zero,
	}
	sig, err := signCLOBOrder(signer, payload, negRisk)
	if err != nil {
		return signedOrderPayload{}, err
	}
	if draft.signatureType == signatureTypePoly1271 {
		typedData := buildOrderTypedData(payload, negRisk)
		sig, err = wrapPOLY1271Signature(signer, maker, typedData)
		if err != nil {
			return signedOrderPayload{}, err
		}
	}
	payload.Signature = sig
	return payload, nil
}

func (c *Client) marketOrderPrice(ctx context.Context, tokenID, side string, amount *big.Rat, orderType string) (*big.Rat, error) {
	book, err := c.OrderBook(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("book lookup failed: %w", err)
	}
	var levels []polytypes.OrderBookLevel
	if side == "BUY" {
		levels = book.Asks
	} else {
		levels = book.Bids
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("no opposing orders")
	}
	sum := new(big.Rat)
	for _, level := range levels {
		price, err := parseRat(level.Price, "level price")
		if err != nil {
			return nil, err
		}
		size, err := parseRat(level.Size, "level size")
		if err != nil {
			return nil, err
		}
		sum.Add(sum, new(big.Rat).Mul(size, price))
		if sum.Cmp(amount) >= 0 {
			return price, nil
		}
	}
	if orderType == "FOK" {
		return nil, fmt.Errorf("insufficient liquidity to fill order")
	}
	return parseRat(levels[0].Price, "level price")
}

func limitFixedAmounts(side string, price, size *big.Rat) (makerAmount, takerAmount string) {
	notional := new(big.Rat).Mul(size, price)
	if side == "BUY" {
		return fixedDecimal(notional, 6), fixedDecimal(size, 6)
	}
	return fixedDecimal(size, 6), fixedDecimal(notional, 6)
}

func fixedDecimal(value *big.Rat, decimals int) string {
	scaled := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	num := new(big.Int).Mul(value.Num(), scaled)
	return new(big.Int).Quo(num, value.Denom()).String()
}

func truncateRat(value *big.Rat, decimals int) *big.Rat {
	if decimals < 0 {
		return new(big.Rat).Set(value)
	}
	scaled := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	num := new(big.Int).Mul(value.Num(), scaled)
	truncated := new(big.Int).Quo(num, value.Denom())
	return new(big.Rat).SetFrac(truncated, scaled)
}

func validateMinimumOrderSize(size *big.Rat, tick *polytypes.TickSize) error {
	minRaw := strings.TrimSpace(tick.MinimumOrderSize)
	if minRaw == "" {
		return nil
	}
	min, err := parseRat(minRaw, "minimum order size")
	if err != nil {
		return err
	}
	if min.Sign() > 0 && size.Cmp(min) < 0 {
		return fmt.Errorf("order size %s is below minimum %s", size.FloatString(6), min.FloatString(6))
	}
	return nil
}

func validatePriceScale(price *big.Rat, raw string, tick *polytypes.TickSize) error {
	tickRaw := firstNonEmpty(tick.TickSize, tick.MinimumTickSize)
	if tickRaw == "" {
		return fmt.Errorf("tick size response missing tick size")
	}
	tickRat, err := parseRat(tickRaw, "tick size")
	if err != nil {
		return err
	}
	if price.Cmp(tickRat) < 0 || price.Cmp(new(big.Rat).Sub(big.NewRat(1, 1), tickRat)) > 0 {
		return fmt.Errorf("price %s is out of bounds for tick size %s", raw, tickRaw)
	}
	if decimalPlaces(raw) > decimalPlaces(tickRaw) {
		return fmt.Errorf("price has too many decimal places for tick size %s", tickRaw)
	}
	return nil
}

func decimalPlaces(raw string) int {
	raw = strings.TrimSpace(raw)
	if idx := strings.IndexByte(raw, '.'); idx >= 0 {
		return len(strings.TrimRight(raw[idx+1:], "0"))
	}
	return 0
}

func parseRat(raw string, field string) (*big.Rat, error) {
	raw = strings.TrimSpace(raw)
	value := new(big.Rat)
	if raw == "" {
		return nil, fmt.Errorf("%s is required", field)
	}
	if _, ok := value.SetString(raw); !ok {
		return nil, fmt.Errorf("invalid %s %q", field, raw)
	}
	return value, nil
}

func parseTokenID(raw string) (*big.Int, error) {
	raw = strings.TrimSpace(raw)
	value, ok := new(big.Int).SetString(raw, 10)
	if raw == "" || !ok || value.Sign() <= 0 {
		return nil, fmt.Errorf("invalid token_id %q", raw)
	}
	return value, nil
}

func normalizeOrderSide(raw string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "BUY", "B":
		return "BUY", nil
	case "SELL", "S":
		return "SELL", nil
	default:
		return "", fmt.Errorf("side must be buy or sell")
	}
}

func normalizeOrderType(raw string, fallback string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "" {
		value = fallback
	}
	switch value {
	case "GTC", "GTD", "FAK", "FOK":
		return value
	default:
		return fallback
	}
}

func generateOrderSalt() (uint64, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, fmt.Errorf("generate salt: %w", err)
	}
	return binary.BigEndian.Uint64(buf[:]) & ((1 << 53) - 1), nil
}
