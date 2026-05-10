package enabletrading

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
	gethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	PolygonChainID             int64  = 137
	ClobAuthAttestationMessage string = "This message attests that I control the given wallet"

	observedUSDCAddress         = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	observedPUSDAddress         = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	observedCTFSpender          = "0x4d97dcd97ec945f40cf65f87097ace5ea0476045"
	observedUSDCSpender         = "0x93070a847efef7f70739046a929d47a521f5b8ee"
	erc20ApproveSelector        = "095ea7b3"
	maxUint256Hex               = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	defaultApprovalDeadlineSecs = int64(240)
)

type Domain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainID           int64  `json:"chainId"`
	VerifyingContract string `json:"verifyingContract,omitempty"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TypedData struct {
	Types       map[string][]Field     `json:"types"`
	PrimaryType string                 `json:"primaryType"`
	Domain      Domain                 `json:"domain"`
	Message     map[string]interface{} `json:"message"`
}

type RelayerClient interface {
	IsDeployed(ctx context.Context, ownerAddress string) (bool, error)
	SubmitWalletCreate(ctx context.Context, ownerAddress string) (*relayer.RelayerTransaction, error)
	GetNonce(ctx context.Context, ownerAddress string) (string, error)
	SubmitWalletBatch(ctx context.Context, ownerAddress, walletAddress, nonce, signature, deadline string, calls []relayer.DepositWalletCall) (*relayer.RelayerTransaction, error)
	ApprovalsReady(ctx context.Context, walletAddress string) (bool, error)
}

type CLOBClient interface {
	CreateOrDeriveAPIKey(ctx context.Context, privateKey string) (auth.APIKey, error)
	VerifyTradingEnabled(ctx context.Context, privateKey string) (bool, error)
}

type Options struct {
	Signer               *auth.PrivateKeySigner
	OwnerPrivateKey      string
	DepositWalletAddress string
	DeployIfNeeded       bool
	CreateOrDeriveAPIKey bool
	ApproveTokens        bool
	MaxApproval          bool
	DryRun               bool
	DeadlineSeconds      int64
	Relayer              RelayerClient
	CLOB                 CLOBClient
}

type Result struct {
	DepositWalletAddress    string      `json:"depositWalletAddress"`
	Deployed                bool        `json:"deployed"`
	ClobAuthSigned          bool        `json:"clobAuthSigned"`
	APIKeysCreated          bool        `json:"apiKeysCreated"`
	TokenApprovalsSigned    bool        `json:"tokenApprovalsSigned"`
	TokenApprovalsSubmitted bool        `json:"tokenApprovalsSubmitted"`
	ReadyToTrade            bool        `json:"readyToTrade"`
	TxHashes                []string    `json:"txHashes"`
	Warnings                []string    `json:"warnings"`
	APIKey                  auth.APIKey `json:"apiKey,omitempty"`
	PlannedClobAuth         *TypedData  `json:"plannedClobAuth,omitempty"`
	PlannedApprovalBatch    *TypedData  `json:"plannedApprovalBatch,omitempty"`
}

func BuildClobAuthTypedData(address, timestamp string, nonce uint64) TypedData {
	return TypedData{
		Types: map[string][]Field{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain:      Domain{Name: "ClobAuthDomain", Version: "1", ChainID: PolygonChainID},
		Message: map[string]interface{}{
			"address":   address,
			"timestamp": timestamp,
			"nonce":     strconv.FormatUint(nonce, 10),
			"message":   ClobAuthAttestationMessage,
		},
	}
}

func SignClobAuthTypedData(signer *auth.PrivateKeySigner, address, timestamp string, nonce uint64) (string, error) {
	if signer == nil {
		return "", fmt.Errorf("signer is required")
	}
	if signer.ChainID() != PolygonChainID {
		return "", fmt.Errorf("wrong chain: expected chainId 137, got %d", signer.ChainID())
	}
	typed := BuildClobAuthTypedData(address, timestamp, nonce)
	if typed.Message["message"] != ClobAuthAttestationMessage {
		return "", fmt.Errorf("invalid ClobAuth attestation message")
	}
	sig, err := signer.SignEIP712(toAPIType(typed))
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(sig), nil
}

func ObservedEnableTradingApprovalCalls() []relayer.DepositWalletCall {
	return []relayer.DepositWalletCall{
		buildApproveCall(observedPUSDAddress, observedCTFSpender),
		buildApproveCall(observedUSDCAddress, observedUSDCSpender),
	}
}

func BuildDepositWalletApprovalBatchTypedData(walletAddress, nonce, deadline string, calls []relayer.DepositWalletCall) (TypedData, error) {
	walletAddress = strings.TrimSpace(walletAddress)
	if walletAddress == "" {
		return TypedData{}, fmt.Errorf("deposit wallet address is required")
	}
	if strings.TrimSpace(nonce) == "" {
		return TypedData{}, fmt.Errorf("nonce is required")
	}
	if strings.TrimSpace(deadline) == "" {
		return TypedData{}, fmt.Errorf("deadline is required")
	}
	if len(calls) == 0 {
		return TypedData{}, fmt.Errorf("at least one approval call is required")
	}
	callMaps := make([]map[string]string, len(calls))
	for i, call := range calls {
		if err := validateApprovalCall(call); err != nil {
			return TypedData{}, fmt.Errorf("call %d: %w", i, err)
		}
		callMaps[i] = map[string]string{"target": call.Target, "value": call.Value, "data": call.Data}
	}
	return TypedData{
		Types: map[string][]Field{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Call": {
				{Name: "target", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "data", Type: "bytes"},
			},
			"Batch": {
				{Name: "wallet", Type: "address"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
				{Name: "calls", Type: "Call[]"},
			},
		},
		PrimaryType: "Batch",
		Domain:      Domain{Name: "DepositWallet", Version: "1", ChainID: PolygonChainID, VerifyingContract: walletAddress},
		Message: map[string]interface{}{
			"wallet":   walletAddress,
			"nonce":    nonce,
			"deadline": deadline,
			"calls":    callMaps,
		},
	}, nil
}

func SignDepositWalletApprovalBatch(signer *auth.PrivateKeySigner, walletAddress, nonce, deadline string, calls []relayer.DepositWalletCall) (string, error) {
	if signer == nil {
		return "", fmt.Errorf("signer is required")
	}
	if signer.ChainID() != PolygonChainID {
		return "", fmt.Errorf("wrong chain: expected chainId 137, got %d", signer.ChainID())
	}
	deadlineInt, err := strconv.ParseInt(strings.TrimSpace(deadline), 10, 64)
	if err != nil {
		return "", fmt.Errorf("deadline must be unix seconds: %w", err)
	}
	if deadlineInt <= time.Now().Unix() {
		return "", fmt.Errorf("deadline is expired")
	}
	typed, err := BuildDepositWalletApprovalBatchTypedData(walletAddress, nonce, deadline, calls)
	if err != nil {
		return "", err
	}
	sig, err := signer.SignEIP712(toAPIType(typed))
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(sig), nil
}

func EnableTradingHeadless(ctx context.Context, opts Options) (Result, error) {
	if opts.Signer == nil {
		return Result{}, fmt.Errorf("signer is required")
	}
	if opts.Signer.ChainID() != PolygonChainID {
		return Result{}, fmt.Errorf("wrong chain: expected chainId 137, got %d", opts.Signer.ChainID())
	}
	wallet := strings.TrimSpace(opts.DepositWalletAddress)
	if wallet == "" {
		return Result{}, fmt.Errorf("deposit wallet address is required")
	}
	result := Result{DepositWalletAddress: wallet}
	owner := opts.Signer.Address()

	deployed := false
	if opts.Relayer != nil {
		isDeployed, err := opts.Relayer.IsDeployed(ctx, owner)
		if err != nil {
			return result, fmt.Errorf("check deposit wallet deployment: %w", err)
		}
		deployed = isDeployed
	}
	if opts.DeployIfNeeded && !deployed {
		if opts.DryRun {
			result.Warnings = append(result.Warnings, "dry-run: deposit wallet deployment would be submitted")
		} else {
			if opts.Relayer == nil {
				return result, fmt.Errorf("relayer client is required to deploy wallet")
			}
			tx, err := opts.Relayer.SubmitWalletCreate(ctx, owner)
			if err != nil {
				return result, err
			}
			result.Deployed = true
			appendTxHash(&result, tx)
		}
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	clobTyped := BuildClobAuthTypedData(owner, timestamp, 0)
	result.PlannedClobAuth = &clobTyped
	if opts.CreateOrDeriveAPIKey {
		if opts.DryRun {
			result.Warnings = append(result.Warnings, "dry-run: CLOB API key creation/derivation would sign ClobAuth")
		} else {
			if _, err := SignClobAuthTypedData(opts.Signer, owner, timestamp, 0); err != nil {
				return result, err
			}
			result.ClobAuthSigned = true
			if opts.CLOB == nil {
				return result, fmt.Errorf("CLOB client is required to create or derive API keys")
			}
			key, err := opts.CLOB.CreateOrDeriveAPIKey(ctx, opts.OwnerPrivateKey)
			if err != nil {
				return result, err
			}
			result.APIKey = key
			result.APIKeysCreated = true
		}
	}

	calls := ObservedEnableTradingApprovalCalls()
	deadline := strconv.FormatInt(time.Now().Unix()+deadlineSeconds(opts.DeadlineSeconds), 10)
	nonce := "0"
	if opts.Relayer != nil {
		fetched, err := opts.Relayer.GetNonce(ctx, owner)
		if err != nil {
			return result, fmt.Errorf("get WALLET nonce: %w", err)
		}
		nonce = fetched
	}
	approvalTyped, err := BuildDepositWalletApprovalBatchTypedData(wallet, nonce, deadline, calls)
	if err != nil {
		return result, err
	}
	result.PlannedApprovalBatch = &approvalTyped
	if opts.ApproveTokens {
		if !opts.MaxApproval {
			return result, fmt.Errorf("max approval requires explicit MaxApproval=true")
		}
		ready := false
		if opts.Relayer != nil {
			var err error
			ready, err = opts.Relayer.ApprovalsReady(ctx, wallet)
			if err != nil {
				return result, fmt.Errorf("check approvals: %w", err)
			}
		}
		if ready {
			result.Warnings = append(result.Warnings, "token approvals already sufficient; skipped")
		} else if opts.DryRun {
			result.Warnings = append(result.Warnings, "dry-run: token approval batch would be signed/submitted")
		} else {
			if opts.Relayer == nil {
				return result, fmt.Errorf("relayer client is required to submit approvals")
			}
			sig, err := SignDepositWalletApprovalBatch(opts.Signer, wallet, nonce, deadline, calls)
			if err != nil {
				return result, err
			}
			result.TokenApprovalsSigned = true
			tx, err := opts.Relayer.SubmitWalletBatch(ctx, owner, wallet, nonce, sig, deadline, calls)
			if err != nil {
				return result, err
			}
			result.TokenApprovalsSubmitted = true
			appendTxHash(&result, tx)
		}
	}

	if opts.CLOB != nil && !opts.DryRun {
		ready, err := opts.CLOB.VerifyTradingEnabled(ctx, opts.OwnerPrivateKey)
		if err != nil {
			return result, err
		}
		result.ReadyToTrade = ready
	}
	return result, nil
}

func (r Result) Redacted() Result {
	r.APIKey = r.APIKey.Redacted()
	return r
}

func (r Result) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func buildApproveCall(tokenAddress, spenderAddress string) relayer.DepositWalletCall {
	spender := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(spenderAddress), "0x"))
	return relayer.DepositWalletCall{
		Target: tokenAddress,
		Value:  "0",
		Data:   "0x" + erc20ApproveSelector + pad32(spender) + maxUint256Hex,
	}
}

func validateApprovalCall(call relayer.DepositWalletCall) error {
	if strings.TrimSpace(call.Target) == "" {
		return fmt.Errorf("target is required")
	}
	if call.Value != "0" {
		return fmt.Errorf("approval call value must be 0")
	}
	data := strings.ToLower(strings.TrimSpace(call.Data))
	if !strings.HasPrefix(data, "0x"+erc20ApproveSelector) {
		return fmt.Errorf("only ERC20 approve calls are allowed")
	}
	if !strings.HasSuffix(data, maxUint256Hex) {
		return fmt.Errorf("approval amount must be maxUint256")
	}
	return nil
}

func pad32(value string) string {
	return strings.Repeat("0", 64-len(value)) + value
}

func deadlineSeconds(seconds int64) int64 {
	if seconds <= 0 {
		return defaultApprovalDeadlineSecs
	}
	return seconds
}

func appendTxHash(result *Result, tx *relayer.RelayerTransaction) {
	if tx != nil && tx.TransactionHash != "" {
		result.TxHashes = append(result.TxHashes, tx.TransactionHash)
	}
}

func toAPIType(td TypedData) apitypes.TypedData {
	types := apitypes.Types{}
	for name, fields := range td.Types {
		converted := make([]apitypes.Type, len(fields))
		for i, field := range fields {
			converted[i] = apitypes.Type{Name: field.Name, Type: field.Type}
		}
		types[name] = converted
	}
	message := apitypes.TypedDataMessage{}
	for key, value := range td.Message {
		switch key {
		case "nonce", "deadline":
			message[key] = toHexDecimal(value)
		case "calls":
			message[key] = toAPICalls(value)
		default:
			message[key] = value
		}
	}
	return apitypes.TypedData{
		Types:       types,
		PrimaryType: td.PrimaryType,
		Domain: apitypes.TypedDataDomain{
			Name:              td.Domain.Name,
			Version:           td.Domain.Version,
			ChainId:           (*gethmath.HexOrDecimal256)(big.NewInt(td.Domain.ChainID)),
			VerifyingContract: td.Domain.VerifyingContract,
		},
		Message: message,
	}
}

func toHexDecimal(value interface{}) *gethmath.HexOrDecimal256 {
	s := fmt.Sprintf("%v", value)
	bi, ok := new(big.Int).SetString(s, 10)
	if !ok {
		bi = big.NewInt(0)
	}
	return (*gethmath.HexOrDecimal256)(bi)
}

func toAPICalls(value interface{}) interface{} {
	calls, ok := value.([]map[string]string)
	if !ok {
		return value
	}
	out := make([]interface{}, len(calls))
	for i, call := range calls {
		out[i] = map[string]interface{}{
			"target": call["target"],
			"value":  toHexDecimal(call["value"]),
			"data":   call["data"],
		}
	}
	return out
}
