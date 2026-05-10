// Package enabletrading builds and signs the headless equivalent of
// polymarket.com's V2 "Enable Trading" flow.
package enabletrading

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	sdkrelayer "github.com/TrebuchetDynamics/polygolem/pkg/relayer"
	gethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	PolygonChainID          = contracts.PolygonChainID
	ClobAuthControlMessage  = "This message attests that I control the given wallet"
	clobAuthDomainName      = "ClobAuthDomain"
	clobAuthDomainVersion   = "1"
	depositWalletDomainName = "DepositWallet"
	depositWalletVersion    = "1"
)

// DepositWalletCall is one call inside a DepositWallet Batch payload.
type DepositWalletCall = sdkrelayer.DepositWalletCall

// TypedDataField is one EIP-712 field descriptor.
type TypedDataField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ClobAuthDomain is the EIP-712 domain used by Polymarket CLOB L1 auth.
type ClobAuthDomain struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	ChainID int64  `json:"chainId"`
}

// ClobAuthMessage is the EIP-712 message used by Polymarket CLOB L1 auth.
type ClobAuthMessage struct {
	Address   string `json:"address"`
	Timestamp string `json:"timestamp"`
	Nonce     uint64 `json:"nonce"`
	Message   string `json:"message"`
}

// ClobAuthTypedData is the JSON shape displayed by polymarket.com for
// Enable Trading -> generate API keys.
type ClobAuthTypedData struct {
	Types       map[string][]TypedDataField `json:"types"`
	PrimaryType string                      `json:"primaryType"`
	Domain      ClobAuthDomain              `json:"domain"`
	Message     ClobAuthMessage             `json:"message"`
}

// ApprovalBatchDomain is the EIP-712 domain for DepositWallet.Batch.
type ApprovalBatchDomain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainID           int64  `json:"chainId"`
	VerifyingContract string `json:"verifyingContract"`
}

// ApprovalBatchMessage is the DepositWallet.Batch message displayed by
// polymarket.com for Enable Trading -> Approve Tokens.
type ApprovalBatchMessage struct {
	Wallet   string              `json:"wallet"`
	Nonce    string              `json:"nonce"`
	Deadline string              `json:"deadline"`
	Calls    []DepositWalletCall `json:"calls"`
}

// DepositWalletBatchTypedData is the JSON shape for DepositWallet.Batch.
type DepositWalletBatchTypedData struct {
	Types       map[string][]TypedDataField `json:"types"`
	PrimaryType string                      `json:"primaryType"`
	Domain      ApprovalBatchDomain         `json:"domain"`
	Message     ApprovalBatchMessage        `json:"message"`
}

type ClobAuthParams struct {
	Address   string
	ChainID   int64
	Timestamp string
	Nonce     uint64
}

type ApprovalBatchParams struct {
	DepositWallet string
	ChainID       int64
	Nonce         string
	Deadline      string
	Calls         []DepositWalletCall
}

// CLOBKeyClient is the narrow CLOB credential interface used by the high-level
// enable-trading helper.
type CLOBKeyClient interface {
	CreateOrDeriveAPIKey(ctx context.Context, privateKey string) (sdkclob.APIKey, error)
}

// WalletRelayer is the narrow relayer interface used by the high-level
// enable-trading helper.
type WalletRelayer interface {
	IsDeployed(ctx context.Context, ownerAddress string) (bool, error)
	SubmitWalletCreate(ctx context.Context, ownerAddress string) (*sdkrelayer.RelayerTransaction, error)
	PollTransaction(ctx context.Context, txID string, maxAttempts int, interval time.Duration) (*sdkrelayer.RelayerTransaction, error)
	GetNonce(ctx context.Context, ownerAddress string) (string, error)
	SubmitWalletBatch(ctx context.Context, ownerAddress, walletAddress, nonce, signature, deadline string, calls []DepositWalletCall) (*sdkrelayer.RelayerTransaction, error)
}

type EnableTradingParams struct {
	OwnerPrivateKey       string
	DepositWalletAddress  string
	DeployIfNeeded        bool
	CreateOrDeriveCLOBKey bool
	ApproveTokens         bool
	MaxApproval           bool
	DryRun                bool

	ChainID           int64
	ClobAuthTimestamp string
	ClobAuthNonce     uint64
	WalletNonce       string
	ApprovalDeadline  string

	CLOB    CLOBKeyClient
	Relayer WalletRelayer
}

type EnableTradingResult struct {
	EOAAddress              string
	DepositWalletAddress    string
	Deployed                bool
	CLOBAuthSigned          bool
	APIKeysCreatedOrDerived bool
	TokenApprovalsBuilt     bool
	TokenApprovalsSigned    bool
	TokenApprovalsSubmitted bool
	ReadyToTrade            bool
	TxHashes                []string
	Warnings                []string
	PlannedActions          []string

	ClobAuthTypedData      *ClobAuthTypedData           `json:"clobAuthTypedData,omitempty"`
	ApprovalBatchTypedData *DepositWalletBatchTypedData `json:"approvalBatchTypedData,omitempty"`
}

func BuildClobAuthTypedData(params ClobAuthParams) (*ClobAuthTypedData, error) {
	if err := requirePolygon(params.ChainID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.Address) == "" {
		return nil, fmt.Errorf("enabletrading: ClobAuth address is required")
	}
	if strings.TrimSpace(params.Timestamp) == "" {
		return nil, fmt.Errorf("enabletrading: ClobAuth timestamp is required")
	}
	return &ClobAuthTypedData{
		Types: map[string][]TypedDataField{
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
		Domain: ClobAuthDomain{
			Name:    clobAuthDomainName,
			Version: clobAuthDomainVersion,
			ChainID: params.ChainID,
		},
		Message: ClobAuthMessage{
			Address:   strings.TrimSpace(params.Address),
			Timestamp: strings.TrimSpace(params.Timestamp),
			Nonce:     params.Nonce,
			Message:   ClobAuthControlMessage,
		},
	}, nil
}

func HashClobAuthTypedData(td *ClobAuthTypedData) ([]byte, error) {
	if err := validateClobAuthTypedData(td); err != nil {
		return nil, err
	}
	hash, _, err := apitypes.TypedDataAndHash(clobAuthToGeth(td))
	if err != nil {
		return nil, fmt.Errorf("enabletrading: hash ClobAuth typed data: %w", err)
	}
	return hash, nil
}

func SignClobAuthTypedData(privateKey string, td *ClobAuthTypedData) (string, error) {
	if err := validateClobAuthTypedData(td); err != nil {
		return "", err
	}
	signer, err := auth.NewPrivateKeySigner(privateKey, td.Domain.ChainID)
	if err != nil {
		return "", fmt.Errorf("enabletrading: init signer: %w", err)
	}
	if !strings.EqualFold(td.Message.Address, signer.Address()) {
		return "", fmt.Errorf("enabletrading: ClobAuth address %s does not match signer EOA %s", td.Message.Address, signer.Address())
	}
	sig, err := signer.SignEIP712(clobAuthToGeth(td))
	if err != nil {
		return "", fmt.Errorf("enabletrading: sign ClobAuth typed data: %w", err)
	}
	return "0x" + hex.EncodeToString(sig), nil
}

func BuildEnableTradingApprovalCalls() []DepositWalletCall {
	return sdkrelayer.BuildEnableTradingApprovalCalls()
}

func BuildEnableTradingApprovalBatchTypedData(params ApprovalBatchParams) (*DepositWalletBatchTypedData, error) {
	if err := requirePolygon(params.ChainID); err != nil {
		return nil, err
	}
	wallet := strings.TrimSpace(params.DepositWallet)
	if wallet == "" {
		return nil, fmt.Errorf("enabletrading: deposit wallet is required")
	}
	if strings.TrimSpace(params.Nonce) == "" {
		return nil, fmt.Errorf("enabletrading: wallet nonce is required")
	}
	if strings.TrimSpace(params.Deadline) == "" {
		return nil, fmt.Errorf("enabletrading: approval deadline is required")
	}
	if len(params.Calls) == 0 {
		return nil, fmt.Errorf("enabletrading: at least one approval call is required")
	}
	if err := ValidateEnableTradingApprovalCalls(params.Calls); err != nil {
		return nil, err
	}
	return &DepositWalletBatchTypedData{
		Types: map[string][]TypedDataField{
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
		Domain: ApprovalBatchDomain{
			Name:              depositWalletDomainName,
			Version:           depositWalletVersion,
			ChainID:           params.ChainID,
			VerifyingContract: wallet,
		},
		Message: ApprovalBatchMessage{
			Wallet:   wallet,
			Nonce:    strings.TrimSpace(params.Nonce),
			Deadline: strings.TrimSpace(params.Deadline),
			Calls:    append([]DepositWalletCall(nil), params.Calls...),
		},
	}, nil
}

func SignDepositWalletApprovalBatch(privateKey string, td *DepositWalletBatchTypedData) (string, error) {
	if err := validateApprovalBatchTypedData(td); err != nil {
		return "", err
	}
	signer, err := auth.NewPrivateKeySigner(privateKey, td.Domain.ChainID)
	if err != nil {
		return "", fmt.Errorf("enabletrading: init signer: %w", err)
	}
	sig, err := signer.SignEIP712(approvalBatchToGeth(td))
	if err != nil {
		return "", fmt.Errorf("enabletrading: sign DepositWallet Batch: %w", err)
	}
	return "0x" + hex.EncodeToString(sig), nil
}

func EnableTradingHeadless(ctx context.Context, params EnableTradingParams) (*EnableTradingResult, error) {
	chainID := params.ChainID
	if chainID == 0 {
		chainID = PolygonChainID
	}
	if err := requirePolygon(chainID); err != nil {
		return nil, err
	}
	signer, err := auth.NewPrivateKeySigner(params.OwnerPrivateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("enabletrading: init signer: %w", err)
	}
	depositWallet := strings.TrimSpace(params.DepositWalletAddress)
	if depositWallet == "" {
		depositWallet, err = auth.MakerAddressForSignatureType(signer.Address(), chainID, 3)
		if err != nil {
			return nil, fmt.Errorf("enabletrading: derive deposit wallet: %w", err)
		}
	}

	result := &EnableTradingResult{
		EOAAddress:           signer.Address(),
		DepositWalletAddress: depositWallet,
	}

	if params.DeployIfNeeded {
		result.PlannedActions = append(result.PlannedActions, "check_or_deploy_deposit_wallet")
		if !params.DryRun && params.Relayer != nil {
			deployed, err := params.Relayer.IsDeployed(ctx, signer.Address())
			if err != nil {
				return nil, fmt.Errorf("enabletrading: check deployed: %w", err)
			}
			result.Deployed = deployed
			if !deployed {
				tx, err := params.Relayer.SubmitWalletCreate(ctx, signer.Address())
				if err != nil {
					return nil, fmt.Errorf("enabletrading: deploy wallet: %w", err)
				}
				result.TxHashes = appendTx(result.TxHashes, tx)
				result.Deployed = true
			}
		}
	}

	timestamp := strings.TrimSpace(params.ClobAuthTimestamp)
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}
	clobTD, err := BuildClobAuthTypedData(ClobAuthParams{
		Address:   signer.Address(),
		ChainID:   chainID,
		Timestamp: timestamp,
		Nonce:     params.ClobAuthNonce,
	})
	if err != nil {
		return nil, err
	}
	result.ClobAuthTypedData = clobTD
	if params.CreateOrDeriveCLOBKey {
		result.PlannedActions = append(result.PlannedActions, "create_or_derive_clob_api_key")
		if !params.DryRun {
			if _, err := SignClobAuthTypedData(params.OwnerPrivateKey, clobTD); err != nil {
				return nil, err
			}
			result.CLOBAuthSigned = true
			if params.CLOB != nil {
				if _, err := params.CLOB.CreateOrDeriveAPIKey(ctx, params.OwnerPrivateKey); err != nil {
					return nil, fmt.Errorf("enabletrading: create or derive CLOB API key: %w", err)
				}
				result.APIKeysCreatedOrDerived = true
			} else {
				result.Warnings = append(result.Warnings, "CLOB client not provided; API key was not created or derived")
			}
		}
	}

	if params.ApproveTokens {
		if !params.MaxApproval {
			return nil, fmt.Errorf("enabletrading: max approval requires explicit MaxApproval=true")
		}
		calls := BuildEnableTradingApprovalCalls()
		nonce := strings.TrimSpace(params.WalletNonce)
		if nonce == "" && !params.DryRun && params.Relayer != nil {
			nonce, err = params.Relayer.GetNonce(ctx, signer.Address())
			if err != nil {
				return nil, fmt.Errorf("enabletrading: fetch wallet nonce: %w", err)
			}
		}
		if nonce == "" {
			nonce = "0"
		}
		deadline := strings.TrimSpace(params.ApprovalDeadline)
		if deadline == "" {
			deadline = sdkrelayer.BuildDeadline(240)
		}
		batchTD, err := BuildEnableTradingApprovalBatchTypedData(ApprovalBatchParams{
			DepositWallet: depositWallet,
			ChainID:       chainID,
			Nonce:         nonce,
			Deadline:      deadline,
			Calls:         calls,
		})
		if err != nil {
			return nil, err
		}
		result.ApprovalBatchTypedData = batchTD
		result.TokenApprovalsBuilt = true
		result.PlannedActions = append(result.PlannedActions, "approve_enable_trading_tokens")
		if !params.DryRun {
			sig, err := SignDepositWalletApprovalBatch(params.OwnerPrivateKey, batchTD)
			if err != nil {
				return nil, err
			}
			result.TokenApprovalsSigned = true
			if params.Relayer != nil {
				tx, err := params.Relayer.SubmitWalletBatch(ctx, signer.Address(), depositWallet, nonce, sig, deadline, calls)
				if err != nil {
					return nil, fmt.Errorf("enabletrading: submit approval batch: %w", err)
				}
				result.TxHashes = appendTx(result.TxHashes, tx)
				result.TokenApprovalsSubmitted = true
			} else {
				result.Warnings = append(result.Warnings, "relayer client not provided; approval batch was signed but not submitted")
			}
		}
	}

	result.ReadyToTrade = result.Deployed && result.APIKeysCreatedOrDerived && (!params.ApproveTokens || result.TokenApprovalsSubmitted)
	return result, nil
}

func ValidateEnableTradingApprovalCalls(calls []DepositWalletCall) error {
	if len(calls) != 2 {
		return fmt.Errorf("enabletrading: expected 2 approval calls, got %d", len(calls))
	}
	expected := []struct {
		target  string
		spender string
	}{
		{contracts.PUSD, contracts.CTF},
		{contracts.USDCE, contracts.CollateralOnramp},
	}
	for i, call := range calls {
		if !strings.EqualFold(call.Target, expected[i].target) {
			return fmt.Errorf("enabletrading: approval call %d target %s is not allowed", i, call.Target)
		}
		if strings.TrimSpace(call.Value) != "0" {
			return fmt.Errorf("enabletrading: approval call %d value must be 0", i)
		}
		if err := validateApproveCalldata(call.Data, expected[i].spender); err != nil {
			return fmt.Errorf("enabletrading: approval call %d: %w", i, err)
		}
	}
	return nil
}

func validateClobAuthTypedData(td *ClobAuthTypedData) error {
	if td == nil {
		return fmt.Errorf("enabletrading: ClobAuth typed data is nil")
	}
	if err := requirePolygon(td.Domain.ChainID); err != nil {
		return err
	}
	if td.Domain.Name != clobAuthDomainName || td.Domain.Version != clobAuthDomainVersion {
		return fmt.Errorf("enabletrading: unexpected ClobAuth domain")
	}
	if td.PrimaryType != "ClobAuth" {
		return fmt.Errorf("enabletrading: unexpected ClobAuth primary type %q", td.PrimaryType)
	}
	if strings.TrimSpace(td.Message.Address) == "" {
		return fmt.Errorf("enabletrading: ClobAuth address is required")
	}
	if td.Message.Message != ClobAuthControlMessage {
		return fmt.Errorf("enabletrading: unexpected ClobAuth control message")
	}
	return nil
}

func validateApprovalBatchTypedData(td *DepositWalletBatchTypedData) error {
	if td == nil {
		return fmt.Errorf("enabletrading: approval batch typed data is nil")
	}
	if err := requirePolygon(td.Domain.ChainID); err != nil {
		return err
	}
	if td.Domain.Name != depositWalletDomainName || td.Domain.Version != depositWalletVersion {
		return fmt.Errorf("enabletrading: unexpected DepositWallet domain")
	}
	if td.PrimaryType != "Batch" {
		return fmt.Errorf("enabletrading: unexpected batch primary type %q", td.PrimaryType)
	}
	if !strings.EqualFold(td.Domain.VerifyingContract, td.Message.Wallet) {
		return fmt.Errorf("enabletrading: verifyingContract must equal message wallet")
	}
	return ValidateEnableTradingApprovalCalls(td.Message.Calls)
}

func requirePolygon(chainID int64) error {
	if chainID != PolygonChainID {
		return fmt.Errorf("enabletrading: chainID must be 137, got %d", chainID)
	}
	return nil
}

func validateApproveCalldata(data, expectedSpender string) error {
	clean := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(data), "0x"))
	if len(clean) != 8+64+64 {
		return fmt.Errorf("approve calldata has length %d", len(clean))
	}
	if !strings.HasPrefix(clean, "095ea7b3") {
		return fmt.Errorf("calldata is not ERC20 approve")
	}
	spenderWord := clean[8 : 8+64]
	spender := "0x" + spenderWord[24:]
	if !strings.EqualFold(spender, expectedSpender) {
		return fmt.Errorf("spender %s is not allowed", spender)
	}
	if clean[8+64:] != strings.Repeat("f", 64) {
		return fmt.Errorf("approval amount is not max uint256")
	}
	return nil
}

func clobAuthToGeth(td *ClobAuthTypedData) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
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
		Domain: apitypes.TypedDataDomain{
			Name:    td.Domain.Name,
			Version: td.Domain.Version,
			ChainId: (*gethmath.HexOrDecimal256)(big.NewInt(td.Domain.ChainID)),
		},
		Message: apitypes.TypedDataMessage{
			"address":   td.Message.Address,
			"timestamp": td.Message.Timestamp,
			"nonce":     (*gethmath.HexOrDecimal256)(new(big.Int).SetUint64(td.Message.Nonce)),
			"message":   td.Message.Message,
		},
	}
}

func approvalBatchToGeth(td *DepositWalletBatchTypedData) apitypes.TypedData {
	calls := make([]interface{}, len(td.Message.Calls))
	for i, call := range td.Message.Calls {
		calls[i] = map[string]interface{}{
			"target": call.Target,
			"value":  call.Value,
			"data":   call.Data,
		}
	}
	return apitypes.TypedData{
		Types: apitypes.Types{
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
		Domain: apitypes.TypedDataDomain{
			Name:              td.Domain.Name,
			Version:           td.Domain.Version,
			ChainId:           (*gethmath.HexOrDecimal256)(big.NewInt(td.Domain.ChainID)),
			VerifyingContract: td.Domain.VerifyingContract,
		},
		Message: apitypes.TypedDataMessage{
			"wallet":   td.Message.Wallet,
			"nonce":    parseUint256(td.Message.Nonce),
			"deadline": parseUint256(td.Message.Deadline),
			"calls":    calls,
		},
	}
}

func parseUint256(value string) *gethmath.HexOrDecimal256 {
	n, ok := new(big.Int).SetString(strings.TrimSpace(value), 10)
	if !ok {
		n = big.NewInt(0)
	}
	return (*gethmath.HexOrDecimal256)(n)
}

func appendTx(out []string, tx *sdkrelayer.RelayerTransaction) []string {
	if tx == nil {
		return out
	}
	if strings.TrimSpace(tx.TransactionHash) != "" {
		return append(out, tx.TransactionHash)
	}
	if strings.TrimSpace(tx.TransactionID) != "" {
		return append(out, tx.TransactionID)
	}
	return out
}
