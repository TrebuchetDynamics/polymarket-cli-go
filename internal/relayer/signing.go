package relayer

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	gethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	depositBatchDomainName    = "DepositWallet"
	depositBatchDomainVersion = "1"
)

func depositBatchDomain(chainID int64, verifyingContract string) apitypes.TypedDataDomain {
	return apitypes.TypedDataDomain{
		Name:              depositBatchDomainName,
		Version:           depositBatchDomainVersion,
		ChainId:           (*gethmath.HexOrDecimal256)(big.NewInt(chainID)),
		VerifyingContract: verifyingContract,
	}
}

func depositBatchTypes() apitypes.Types {
	return apitypes.Types{
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
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
	}
}

func callToMap(call DepositWalletCall) map[string]interface{} {
	return map[string]interface{}{
		"target": call.Target,
		"value":  call.Value,
		"data":   call.Data,
	}
}

// SignWalletBatch builds and signs an EIP-712 DepositWallet.Batch message.
// The signature is a 65-byte ECDSA with 0x prefix, ready for relayer submission.
func SignWalletBatch(signer *auth.PrivateKeySigner, walletAddress string, nonce, deadline string, calls []DepositWalletCall) (string, error) {
	if len(calls) == 0 {
		return "", fmt.Errorf("relayer: at least one call required for batch signing")
	}
	chainID := signer.ChainID()
	callsMaps := make([]interface{}, len(calls))
	for i, call := range calls {
		callsMaps[i] = callToMap(call)
	}
	typed := apitypes.TypedData{
		Types:       depositBatchTypes(),
		PrimaryType: "Batch",
		Domain:      depositBatchDomain(chainID, walletAddress),
		Message: map[string]interface{}{
			"wallet":   walletAddress,
			"nonce":    nonce,
			"deadline": deadline,
			"calls":    callsMaps,
		},
	}
	sig, err := signer.SignEIP712(typed)
	if err != nil {
		return "", fmt.Errorf("relayer: sign WALLET batch: %w", err)
	}
	return "0x" + hex.EncodeToString(sig), nil
}

// MinWalletBatchDeadlineSeconds is the shortest WALLET batch validity window
// we send to the production relayer. Shorter windows can be rejected with
// "deadline too soon"; polymarket.com currently uses roughly this window for
// deposit-wallet approvals.
const MinWalletBatchDeadlineSeconds int64 = 1800

// BuildDeadline returns a deadline string for a WALLET batch, defaulting to
// now + MinWalletBatchDeadlineSeconds and clamping shorter caller-provided
// windows up to that minimum.
func BuildDeadline(secondsFromNow int64) string {
	if secondsFromNow < MinWalletBatchDeadlineSeconds {
		secondsFromNow = MinWalletBatchDeadlineSeconds
	}
	return fmt.Sprintf("%d", time.Now().Unix()+secondsFromNow)
}

// ParseNonce converts a nonce string to a numeric string the relayer accepts.
func ParseNonce(raw string) string {
	return strings.TrimSpace(raw)
}
