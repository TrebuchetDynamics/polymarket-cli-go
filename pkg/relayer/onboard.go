package relayer

import (
	"context"
	"fmt"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
)

// OnboardOptions controls the public one-call deposit-wallet onboarding flow.
// By default, OnboardDepositWallet checks deployment, deploys when needed,
// polls the deploy transaction, and submits the standard approval batch.
type OnboardOptions struct {
	// SkipDeploy skips the deployed check and WALLET-CREATE submission.
	SkipDeploy bool
	// SkipApprove skips the standard pUSD + CTF approval batch.
	SkipApprove bool
	// DeployPollMaxAttempts controls deploy polling. Defaults to 50.
	DeployPollMaxAttempts int
	// DeployPollInterval controls deploy polling. Defaults to 2 seconds.
	DeployPollInterval time.Duration
	// ApprovalDeadlineSeconds is passed to BuildDeadline. Values shorter than
	// MinWalletBatchDeadlineSeconds are clamped to the relayer-safe minimum.
	ApprovalDeadlineSeconds int64
}

// OnboardDeployResult describes the deploy phase.
type OnboardDeployResult struct {
	TransactionID   string `json:"transactionID,omitempty"`
	State           string `json:"state,omitempty"`
	Skipped         bool   `json:"skipped,omitempty"`
	AlreadyDeployed bool   `json:"alreadyDeployed,omitempty"`
}

// OnboardApprovalResult describes the approval-batch phase.
type OnboardApprovalResult struct {
	TransactionID string `json:"transactionID,omitempty"`
	State         string `json:"state,omitempty"`
	Nonce         string `json:"nonce,omitempty"`
	Deadline      string `json:"deadline,omitempty"`
	CallCount     int    `json:"callCount,omitempty"`
	Skipped       bool   `json:"skipped,omitempty"`
}

// OnboardResult is returned by OnboardDepositWallet.
type OnboardResult struct {
	Owner         string                 `json:"owner"`
	DepositWallet string                 `json:"depositWallet"`
	Deploy        *OnboardDeployResult   `json:"deploy,omitempty"`
	Approve       *OnboardApprovalResult `json:"approve,omitempty"`
	NextSteps     []string               `json:"nextSteps,omitempty"`
}

// DepositWalletAddress returns the controlling EOA address and deterministic
// Polymarket V2 deposit-wallet address for privateKey.
func DepositWalletAddress(privateKey string) (owner string, depositWallet string, err error) {
	signer, err := NewSigner(privateKey, 137)
	if err != nil {
		return "", "", fmt.Errorf("relayer: init signer: %w", err)
	}
	wallet, err := auth.MakerAddressForSignatureType(signer.Address(), signer.ChainID(), 3)
	if err != nil {
		return "", "", fmt.Errorf("relayer: derive deposit wallet: %w", err)
	}
	return signer.Address(), wallet, nil
}

// OnboardDepositWallet derives the deposit wallet address, deploys it through
// the relayer when needed, and submits the standard V2 pUSD + CTF approval
// batch. It does not transfer pUSD; callers must fund explicitly.
func OnboardDepositWallet(ctx context.Context, client *Client, privateKey string, opts OnboardOptions) (*OnboardResult, error) {
	if client == nil {
		return nil, fmt.Errorf("relayer: client is required")
	}
	signer, err := NewSigner(privateKey, 137)
	if err != nil {
		return nil, fmt.Errorf("relayer: init signer: %w", err)
	}
	owner, wallet, err := DepositWalletAddress(privateKey)
	if err != nil {
		return nil, err
	}
	result := &OnboardResult{
		Owner:         owner,
		DepositWallet: wallet,
		NextSteps: []string{
			"Fund the deposit wallet with pUSD before live trading.",
			"Refresh CLOB collateral balance and allowances before placing orders.",
		},
	}

	if opts.SkipDeploy {
		result.Deploy = &OnboardDeployResult{Skipped: true, State: "skipped"}
	} else {
		deployed, err := client.IsDeployed(ctx, owner)
		if err != nil {
			return nil, fmt.Errorf("relayer: check deployed: %w", err)
		}
		if deployed {
			result.Deploy = &OnboardDeployResult{AlreadyDeployed: true, State: "already_deployed"}
		} else {
			tx, err := client.SubmitWalletCreate(ctx, owner)
			if err != nil {
				return nil, fmt.Errorf("relayer: WALLET-CREATE: %w", err)
			}
			final, err := client.PollTransaction(ctx, tx.TransactionID, opts.DeployPollMaxAttempts, opts.DeployPollInterval)
			if err != nil {
				return nil, fmt.Errorf("relayer: poll WALLET-CREATE: %w", err)
			}
			result.Deploy = &OnboardDeployResult{
				TransactionID: final.TransactionID,
				State:         final.State,
			}
		}
	}

	if opts.SkipApprove {
		result.Approve = &OnboardApprovalResult{Skipped: true, State: "skipped"}
		return result, nil
	}

	// 6-call trading approvals plus 4-call adapter approvals. The latter
	// is required for V2 split/merge/redeem; baking it into the
	// post-deploy batch makes new wallets redeem-ready out of the box
	// and avoids a separate operator step.
	calls := append(BuildApprovalCalls(), BuildAdapterApprovalCalls()...)
	nonce, err := client.GetNonce(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("relayer: fetch nonce: %w", err)
	}
	deadline := BuildDeadline(opts.ApprovalDeadlineSeconds)
	sig, err := SignWalletBatch(signer, wallet, nonce, deadline, calls)
	if err != nil {
		return nil, fmt.Errorf("relayer: sign approval batch: %w", err)
	}
	tx, err := client.SubmitWalletBatch(ctx, owner, wallet, nonce, sig, deadline, calls)
	if err != nil {
		return nil, fmt.Errorf("relayer: submit approval batch: %w", err)
	}
	result.Approve = &OnboardApprovalResult{
		TransactionID: tx.TransactionID,
		State:         tx.State,
		Nonce:         nonce,
		Deadline:      deadline,
		CallCount:     len(calls),
	}
	return result, nil
}
