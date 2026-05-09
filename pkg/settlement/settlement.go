// Package settlement turns redeemable Polymarket V2 positions into
// deposit-wallet WALLET batches that route through the V2 collateral
// adapters and return pUSD to the wallet.
//
// V2 redemption flow:
//
//  1. Find positions where Data API redeemable=true.
//  2. Group by conditionId (a binary market with both YES and NO open
//     shows two rows but redeems with one call; neg-risk markets share
//     a conditionId across questions).
//  3. For each unique conditionId, encode redeemPositions(address(0),
//     bytes32(0), conditionId, []) calldata. The adapter ignores the
//     first three args and the indexSets array.
//  4. Target CtfCollateralAdapter for binary markets, or
//     NegRiskCtfCollateralAdapter for neg-risk markets.
//  5. Sign a single WALLET batch and submit via the relayer.
//
// Stability: FindRedeemable, BuildRedeemCall, SubmitRedeem,
// RedeemablePosition, and RedeemResult are part of the polygolem public
// SDK and follow semver.
//
// Live-money safety: this package builds and submits Polygon mainnet
// transactions. Callers must enforce their own gates (operator confirm,
// dry-run review, etc.) before calling SubmitRedeem.
package settlement

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	"github.com/TrebuchetDynamics/polygolem/pkg/ctf"
	"github.com/TrebuchetDynamics/polygolem/pkg/data"
	"github.com/TrebuchetDynamics/polygolem/pkg/relayer"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

// DefaultBatchLimit caps the number of redeem calls per WALLET batch.
// Empirical safety bound; chunk callers when more positions are needed.
const DefaultBatchLimit = 10

// RedeemablePosition is the redemption-relevant subset of a Data API
// position. Built from types.Position; sized to what BuildRedeemCall
// and operator output need, no more.
type RedeemablePosition struct {
	TokenID      string  `json:"tokenID"`
	ConditionID  string  `json:"conditionID"`
	Size         float64 `json:"size"`
	Outcome      string  `json:"outcome"`
	NegativeRisk bool    `json:"negativeRisk"`
	EndDate      string  `json:"endDate"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
}

// RedeemResult summarizes the relayer response for a redeem batch.
type RedeemResult struct {
	TransactionID string               `json:"transactionID"`
	State         string               `json:"state"`
	Wallet        string               `json:"wallet"`
	Nonce         string               `json:"nonce"`
	Deadline      string               `json:"deadline"`
	CallCount     int                  `json:"callCount"`
	Redeemed      []RedeemablePosition `json:"redeemed"`
}

// FindRedeemable returns positions with redeemable=true for owner via
// the Data API. owner must be the deposit wallet address (positions
// live in the wallet, not the EOA). Does not call Gamma; NegativeRisk
// is taken from the Data API row.
func FindRedeemable(ctx context.Context, dataClient *data.Client, owner string) ([]RedeemablePosition, error) {
	if dataClient == nil {
		return nil, fmt.Errorf("settlement: data client is required")
	}
	rows, err := dataClient.CurrentPositions(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("settlement: positions: %w", err)
	}
	out := make([]RedeemablePosition, 0, len(rows))
	for _, p := range rows {
		if !p.Redeemable {
			continue
		}
		out = append(out, fromTypesPosition(p))
	}
	return out, nil
}

func fromTypesPosition(p types.Position) RedeemablePosition {
	return RedeemablePosition{
		TokenID:      p.TokenID,
		ConditionID:  p.ConditionID,
		Size:         p.Size,
		Outcome:      p.Outcome,
		NegativeRisk: p.NegativeRisk,
		EndDate:      p.EndDate,
		Title:        p.Title,
		Slug:         p.Slug,
	}
}

// BuildRedeemCall encodes redeemPositions(address(0), bytes32(0),
// conditionId, []) for the V2 collateral adapter that matches the
// position's market kind. Calldata reuses pkg/ctf.RedeemPositionsData;
// only the call target switches between CtfCollateralAdapter and
// NegRiskCtfCollateralAdapter.
//
// The adapter ignores collateralToken, parentCollectionId, and
// indexSets internally — it reads the wallet's CTF balances on the
// derived position IDs and uses CTFHelpers.partition() = [1, 2] for
// the underlying redeem. We pass zero values to keep calldata minimal.
func BuildRedeemCall(p RedeemablePosition) (relayer.DepositWalletCall, error) {
	if p.ConditionID == "" {
		return relayer.DepositWalletCall{}, fmt.Errorf("settlement: empty conditionID")
	}
	cid := common.HexToHash(p.ConditionID)
	calldata, err := ctf.RedeemPositionsData(common.Address{}, common.Hash{}, cid, []*big.Int{})
	if err != nil {
		return relayer.DepositWalletCall{}, fmt.Errorf("settlement: encode redeem: %w", err)
	}
	return relayer.DepositWalletCall{
		Target: contracts.RedeemAdapterFor(p.NegativeRisk),
		Value:  "0",
		Data:   "0x" + hex.EncodeToString(calldata),
	}, nil
}

// SubmitRedeem groups positions by conditionID, builds one
// DepositWalletCall per unique condition, signs a single WALLET batch
// (capped at limit calls), and submits via the relayer.
//
// Idempotent: re-running on a wallet whose CTF balance for a condition
// is already zero is a contract-level no-op (the adapter zero-pays).
//
// limit defaults to DefaultBatchLimit when <= 0. If len(positions)
// exceeds limit after deduplication, only the first limit conditions
// are submitted; the caller must split into multiple batches.
func SubmitRedeem(
	ctx context.Context,
	rc *relayer.Client,
	privateKey string,
	positions []RedeemablePosition,
	limit int,
) (*RedeemResult, error) {
	if rc == nil {
		return nil, fmt.Errorf("settlement: relayer client is required")
	}
	if len(positions) == 0 {
		return nil, fmt.Errorf("settlement: no positions to redeem")
	}
	if limit <= 0 {
		limit = DefaultBatchLimit
	}

	signer, err := relayer.NewSigner(privateKey, 137)
	if err != nil {
		return nil, fmt.Errorf("settlement: init signer: %w", err)
	}
	owner, wallet, err := relayer.DepositWalletAddress(privateKey)
	if err != nil {
		return nil, err
	}

	deduped := dedupeByCondition(positions)
	if len(deduped) > limit {
		deduped = deduped[:limit]
	}

	calls := make([]relayer.DepositWalletCall, 0, len(deduped))
	for _, p := range deduped {
		call, err := BuildRedeemCall(p)
		if err != nil {
			return nil, fmt.Errorf("settlement: build call for %s: %w", p.ConditionID, err)
		}
		calls = append(calls, call)
	}

	nonce, err := rc.GetNonce(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("settlement: fetch nonce: %w", err)
	}
	deadline := relayer.BuildDeadline(240)
	sig, err := relayer.SignWalletBatch(signer, wallet, nonce, deadline, calls)
	if err != nil {
		return nil, fmt.Errorf("settlement: sign batch: %w", err)
	}
	tx, err := rc.SubmitWalletBatch(ctx, owner, wallet, nonce, sig, deadline, calls)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not in the allowed list") {
			return nil, fmt.Errorf("settlement: upstream relayer allowlist blocker: submit redeem batch: %w", err)
		}
		return nil, fmt.Errorf("settlement: submit redeem batch: %w", err)
	}
	return &RedeemResult{
		TransactionID: tx.TransactionID,
		State:         tx.State,
		Wallet:        wallet,
		Nonce:         nonce,
		Deadline:      deadline,
		CallCount:     len(calls),
		Redeemed:      deduped,
	}, nil
}

// dedupeByCondition collapses YES/NO duplicates that share a conditionId.
// Order-preserving on first appearance; later duplicates dropped.
func dedupeByCondition(rows []RedeemablePosition) []RedeemablePosition {
	seen := make(map[string]struct{}, len(rows))
	out := make([]RedeemablePosition, 0, len(rows))
	for _, p := range rows {
		if p.ConditionID == "" {
			continue
		}
		if _, ok := seen[p.ConditionID]; ok {
			continue
		}
		seen[p.ConditionID] = struct{}{}
		out = append(out, p)
	}
	return out
}
