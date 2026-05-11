package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
	"github.com/TrebuchetDynamics/polygolem/internal/rpc"
	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	"github.com/TrebuchetDynamics/polygolem/pkg/data"
	sdkenabletrading "github.com/TrebuchetDynamics/polygolem/pkg/enabletrading"
	"github.com/TrebuchetDynamics/polygolem/pkg/settlement"
	"github.com/spf13/cobra"
)

// upstreamRelayerBlockJSON is the structured response surfaced when
// Polymarket's relayer rejects a WALLET batch with an allowlist
// rejection (HTTP 400 "not in the allowed list" / "are not permitted"
// / "call blocked"). The operator must first verify that local contract
// constants match Polymarket's current contract reference. If they do,
// the V2 deposit-wallet path is non-negotiable and no fallback is attempted.
func upstreamRelayerBlockJSON(wallet, command string, err error) map[string]interface{} {
	return map[string]interface{}{
		"ok":            false,
		"depositWallet": wallet,
		"command":       command,
		"error": map[string]interface{}{
			"code":    "RELAYER_ALLOWLIST_BLOCKED",
			"message": err.Error(),
			"action":  "stop",
			"reason":  "Polymarket relayer rejected the WALLET batch via its allowlist policy. Verify the local V2 adapter constants against Polymarket's current contract reference; if they match, stop. The V2 deposit wallet path has no EOA bypass, raw CTF fallback, or SAFE/PROXY shortcut.",
			"upstream": map[string]string{
				"state":  "allowlist-rejected",
				"verify": "https://docs.polymarket.com/resources/contracts",
			},
		},
	}
}

const defaultDataAPIURL = "https://data-api.polymarket.com"

const defaultRelayerURL = "https://relayer-v2.polymarket.com"

func depositWalletCmd(jsonOut bool) *cobra.Command {
	cmd := commandGroup("deposit-wallet", "Deposit wallet onboarding (WALLET-CREATE, nonce, batch, status)")

	cmd.AddCommand(depositWalletDeriveCmd(jsonOut))
	cmd.AddCommand(depositWalletDeployCmd(jsonOut))
	cmd.AddCommand(depositWalletNonceCmd(jsonOut))
	cmd.AddCommand(depositWalletStatusCmd(jsonOut))
	cmd.AddCommand(depositWalletBatchCmd(jsonOut))
	cmd.AddCommand(depositWalletApproveCmd(jsonOut))
	cmd.AddCommand(depositWalletApproveAdaptersCmd(jsonOut))
	cmd.AddCommand(depositWalletEnableTradingCmd(jsonOut))
	cmd.AddCommand(depositWalletSettlementStatusCmd(jsonOut))
	cmd.AddCommand(depositWalletRedeemableCmd(jsonOut))
	cmd.AddCommand(depositWalletRedeemCmd(jsonOut))
	cmd.AddCommand(depositWalletFundCmd(jsonOut))
	cmd.AddCommand(depositWalletSwapCmd(jsonOut))
	cmd.AddCommand(depositWalletOnboardCmd(jsonOut))
	return cmd
}

func depositWalletDeriveCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := &cobra.Command{
		Use:   "derive",
		Short: "Derive the deterministic deposit wallet address",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			wallet, err := auth.MakerAddressForSignatureType(signer.Address(), signer.ChainID(), 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			return w.printJSON(cmd, map[string]string{
				"owner":         signer.Address(),
				"depositWallet": wallet,
			})
		},
	}
	return cmd
}

func depositWalletDeployCmd(jsonOut bool) *cobra.Command {
	var wait bool
	var timeout time.Duration
	var rpcURL string
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the deposit wallet via relayer WALLET-CREATE",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			wallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet address: %w", err)
			}
			codeStatus, err := contracts.DepositWalletDeployed(cmd.Context(), wallet, firstNonEmptyCLI(rpcURL, os.Getenv("POLYGON_RPC_URL")))
			if err != nil {
				return fmt.Errorf("check on-chain deposit wallet code before WALLET-CREATE: %w", err)
			}
			if codeStatus.Deployed {
				return printJSON(cmd, map[string]interface{}{
					"state":               "already_deployed",
					"owner":               owner,
					"depositWallet":       wallet,
					"onchainCodeDeployed": true,
					"deploymentSource":    codeStatus.Source,
					"note":                "deposit wallet already has code on Polygon; skipped WALLET-CREATE",
				})
			}
			rc, authResult, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			tx, err := rc.SubmitWalletCreate(cmd.Context(), owner)
			if err != nil {
				return fmt.Errorf("WALLET-CREATE failed: %w", err)
			}
			if !wait {
				return printJSON(cmd, tx)
			}
			if timeout <= 0 {
				timeout = 2 * time.Minute
			}
			maxAttempts := int(timeout / (2 * time.Second))
			if maxAttempts < 1 {
				maxAttempts = 1
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			final, err := rc.PollTransaction(ctx, tx.TransactionID, maxAttempts, 2*time.Second)
			if err != nil {
				return fmt.Errorf("WALLET-CREATE poll: %w", err)
			}
			return printJSON(cmd, map[string]interface{}{
				"transactionID": final.TransactionID,
				"state":         final.State,
				"owner":         owner,
				"depositWallet": wallet,
				"auth":          authResult,
			})
		},
	}
	cmd.Flags().BoolVar(&wait, "wait", false, "poll until transaction reaches terminal state")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "max wait time for --wait")
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Polygon RPC URL (default: public node)")
	return cmd
}

func depositWalletNonceCmd(jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nonce",
		Short: "Get the current WALLET nonce for the owner",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			rc, _, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			nonce, err := rc.GetNonce(cmd.Context(), owner)
			if err != nil {
				return fmt.Errorf("get nonce: %w", err)
			}
			return printJSON(cmd, map[string]string{
				"address": owner,
				"type":    "WALLET",
				"nonce":   nonce,
			})
		},
	}
	return cmd
}

func depositWalletStatusCmd(jsonOut bool) *cobra.Command {
	var txID string
	var checkEnableTrading bool
	var rpcURL string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check deposit wallet deployment status or transaction state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			rc, _, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()

			if txID != "" {
				tx, err := rc.GetTransaction(cmd.Context(), txID)
				if err != nil {
					return fmt.Errorf("get transaction: %w", err)
				}
				return printJSON(cmd, tx)
			}
			relayerDeployed, err := rc.IsDeployed(cmd.Context(), owner)
			if err != nil {
				return fmt.Errorf("deployed check: %w", err)
			}
			wallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet address: %w", err)
			}
			onchainDeployed := relayerDeployed
			if !relayerDeployed {
				codeStatus, err := contracts.DepositWalletDeployed(cmd.Context(), wallet, os.Getenv("POLYGON_RPC_URL"))
				if err != nil {
					return fmt.Errorf("on-chain deposit wallet code check: %w", err)
				}
				onchainDeployed = codeStatus.Deployed
			}
			nonce, err := rc.GetNonce(cmd.Context(), owner)
			if err != nil {
				nonce = "error: " + err.Error()
			}
			result := map[string]interface{}{
				"owner":                  owner,
				"depositWallet":          wallet,
				"deployed":               relayerDeployed || onchainDeployed,
				"relayerDeployed":        relayerDeployed,
				"onchainCodeDeployed":    onchainDeployed,
				"deploymentStatusSource": deploymentStatusSource(relayerDeployed, onchainDeployed),
				"walletNonce":            nonce,
			}
			if checkEnableTrading {
				validation, err := validateEnableTradingReadiness(cmd.Context(), key, wallet, nonce, clobCredentialsReadyForCLI(cmd.Context(), key), firstNonEmptyCLI(rpcURL, os.Getenv("POLYGON_RPC_URL")))
				if err != nil {
					return err
				}
				result["enableTrading"] = validation
			}
			return printJSON(cmd, result)
		},
	}
	cmd.Flags().StringVar(&txID, "tx-id", "", "transaction ID to poll")
	cmd.Flags().BoolVar(&checkEnableTrading, "check-enable-trading", false, "validate ClobAuth signing and UI Enable Trading token approvals")
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Polygon RPC URL for --check-enable-trading allowance checks (default: POLYGON_RPC_URL or public node)")
	return cmd
}

func depositWalletBatchCmd(jsonOut bool) *cobra.Command {
	var callsJSON string
	var walletAddress string
	var nonce string
	var deadline int64
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Sign and submit a deposit wallet WALLET batch",
		Long: `Sign an EIP-712 DepositWallet.Batch message and submit to the relayer.

The --calls-json must be a JSON array of DepositWalletCall objects:
  [{"target":"0x...","value":"0","data":"0x..."}, ...]

Use --auto-approve to build and submit the standard 6-call approval batch
(pUSD + CTF for all 3 V2 exchange spenders).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()

			if strings.TrimSpace(callsJSON) == "" {
				return fmt.Errorf("--calls-json is required (use --auto-approve for standard approval batch)")
			}
			var calls []relayer.DepositWalletCall
			if err := json.Unmarshal([]byte(callsJSON), &calls); err != nil {
				return fmt.Errorf("parse --calls-json: %w", err)
			}
			if len(calls) == 0 {
				return fmt.Errorf("--calls-json must contain at least one call")
			}

			if strings.TrimSpace(walletAddress) == "" {
				var err error
				walletAddress, err = auth.MakerAddressForSignatureType(owner, 137, 3)
				if err != nil {
					return fmt.Errorf("derive deposit wallet: %w", err)
				}
			}

			rc, _, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}

			if strings.TrimSpace(nonce) == "" {
				n, err := rc.GetNonce(cmd.Context(), owner)
				if err != nil {
					return fmt.Errorf("fetch nonce: %w", err)
				}
				nonce = n
			}

			dl := relayer.BuildDeadline(deadline)
			sig, err := relayer.SignWalletBatch(signer, walletAddress, nonce, dl, calls)
			if err != nil {
				return fmt.Errorf("sign batch: %w", err)
			}

			tx, err := rc.SubmitWalletBatch(cmd.Context(), owner, walletAddress, nonce, sig, dl, calls)
			if err != nil {
				return fmt.Errorf("submit WALLET batch: %w", err)
			}
			return printJSON(cmd, map[string]interface{}{
				"transactionID": tx.TransactionID,
				"state":         tx.State,
				"wallet":        walletAddress,
				"nonce":         nonce,
				"callCount":     len(calls),
			})
		},
	}
	cmd.Flags().StringVar(&callsJSON, "calls-json", "", "JSON array of DepositWalletCall objects")
	cmd.Flags().StringVar(&walletAddress, "wallet", "", "deposit wallet address (default: derived from EOA)")
	cmd.Flags().StringVar(&nonce, "nonce", "", "WALLET nonce (default: fetched from relayer)")
	cmd.Flags().Int64Var(&deadline, "deadline", relayer.MinWalletBatchDeadlineSeconds, "deadline seconds from now")
	return cmd
}

func depositWalletApproveCmd(jsonOut bool) *cobra.Command {
	var autoApprove bool
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Build and optionally submit approval calls for the deposit wallet",
		Long: `Build the standard 6-call approval batch (pUSD + CTF for all 3 V2 exchange spenders).

Without --submit, prints the calldata JSON for review.
With --submit, signs and submits the WALLET batch via the relayer.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			callsJSON, err := relayer.BuildApprovalCallsJSON()
			if err != nil {
				return fmt.Errorf("build approval calls: %w", err)
			}
			if !autoApprove {
				raw := json.RawMessage(callsJSON)
				return printJSON(cmd, map[string]interface{}{
					"calls": raw,
					"note":  "review calldata, then run with --submit to sign and send",
				})
			}
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			walletAddress, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			rc, _, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			nonce, err := rc.GetNonce(cmd.Context(), owner)
			if err != nil {
				return fmt.Errorf("fetch nonce: %w", err)
			}
			var calls []relayer.DepositWalletCall
			if err := json.Unmarshal([]byte(callsJSON), &calls); err != nil {
				return fmt.Errorf("parse approval calls: %w", err)
			}
			dl := relayer.BuildDeadline(240)
			sig, err := relayer.SignWalletBatch(signer, walletAddress, nonce, dl, calls)
			if err != nil {
				return fmt.Errorf("sign batch: %w", err)
			}
			tx, err := rc.SubmitWalletBatch(cmd.Context(), owner, walletAddress, nonce, sig, dl, calls)
			if err != nil {
				return fmt.Errorf("submit approval batch: %w", err)
			}
			return printJSON(cmd, map[string]interface{}{
				"transactionID": tx.TransactionID,
				"state":         tx.State,
				"wallet":        walletAddress,
				"approvals":     len(calls),
			})
		},
	}
	cmd.Flags().BoolVar(&autoApprove, "submit", false, "sign and submit the approval batch")
	return cmd
}

// depositWalletApproveAdaptersCmd is the one-shot migration command for
// existing deposit wallets that ran the original 6-call trading-only
// approval batch and therefore cannot redeem on V2. It submits the 4-call
// adapter approval batch (pUSD approve + CTF setApprovalForAll for both
// CtfCollateralAdapter and NegRiskCtfCollateralAdapter). Idempotent.
//
// Live-money safety: dry-run by default (prints calldata only). To submit,
// the operator must pass BOTH --submit AND --confirm APPROVE_ADAPTERS.
func depositWalletApproveAdaptersCmd(jsonOut bool) *cobra.Command {
	var submit bool
	var confirm string
	cmd := &cobra.Command{
		Use:   "approve-adapters",
		Short: "Approve V2 collateral adapters for redeem (one-shot per wallet)",
		Long: `Submits the 4-call adapter approval batch (pUSD approve + CTF setApprovalForAll
for CtfCollateralAdapter and NegRiskCtfCollateralAdapter). Required once per
deposit wallet before V2 redeem will succeed. Idempotent.

Without --submit, prints the calldata JSON for review.
With --submit, the operator must also pass --confirm APPROVE_ADAPTERS to
authorize the live-money WALLET batch.

NOTE: The V2 deposit-wallet path is non-negotiable: the owner signs an EIP-712
WALLET batch, the relayer submits it through the deposit-wallet factory, and
the wallet call targets the V2 collateral adapters. If Polymarket's relayer
allowlist rejects these calls with HTTP 400 "not in the allowed list", first
verify the adapter addresses against Polymarket's current contract reference;
if they match, stop.
The wallet implementation gates execute() behind onlyFactory and the factory
gates proxy() behind onlyOperator, so a direct EOA bypass is not possible.
Do not fall back to raw ConditionalTokens.redeemPositions, SAFE, or PROXY;
V2 deposit-wallet redeem must route through the collateral adapters.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			callsJSON, err := relayer.BuildAdapterApprovalCallsJSON()
			if err != nil {
				return fmt.Errorf("build adapter approval calls: %w", err)
			}
			if !submit {
				raw := json.RawMessage(callsJSON)
				return printJSON(cmd, map[string]interface{}{
					"calls":    raw,
					"adapters": []string{contracts.CtfCollateralAdapter, contracts.NegRiskCtfCollateralAdapter},
					"note":     "review calldata, then run with --submit --confirm APPROVE_ADAPTERS to sign and send",
				})
			}
			if confirm != "APPROVE_ADAPTERS" {
				return fmt.Errorf("--submit requires --confirm APPROVE_ADAPTERS (got %q)", confirm)
			}
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			walletAddress, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			var calls []relayer.DepositWalletCall
			if err := json.Unmarshal([]byte(callsJSON), &calls); err != nil {
				return fmt.Errorf("parse adapter approval calls: %w", err)
			}

			rc, _, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			nonce, err := rc.GetNonce(cmd.Context(), owner)
			if err != nil {
				return fmt.Errorf("fetch nonce: %w", err)
			}
			dl := relayer.BuildDeadline(240)
			sig, err := relayer.SignWalletBatch(signer, walletAddress, nonce, dl, calls)
			if err != nil {
				return fmt.Errorf("sign batch: %w", err)
			}
			tx, err := rc.SubmitWalletBatch(cmd.Context(), owner, walletAddress, nonce, sig, dl, calls)
			if err != nil {
				if errors.Is(err, relayer.ErrRelayerAllowlistBlocked) {
					return printJSON(cmd, upstreamRelayerBlockJSON(walletAddress, "approve-adapters", err))
				}
				return fmt.Errorf("submit adapter approval batch: %w", err)
			}
			return printJSON(cmd, map[string]interface{}{
				"transactionID": tx.TransactionID,
				"state":         tx.State,
				"wallet":        walletAddress,
				"approvals":     len(calls),
				"adapters":      []string{contracts.CtfCollateralAdapter, contracts.NegRiskCtfCollateralAdapter},
				"path":          "relayer",
			})
		},
	}
	cmd.Flags().BoolVar(&submit, "submit", false, "sign and submit the adapter approval batch (requires --confirm APPROVE_ADAPTERS)")
	cmd.Flags().StringVar(&confirm, "confirm", "", "live-money confirmation token; must be 'APPROVE_ADAPTERS' when --submit is set")
	return cmd
}

func depositWalletEnableTradingCmd(jsonOut bool) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "enable-trading",
		Short: "Complete the UI Enable Trading signs for an existing deposit wallet",
		Long: `Signs the same two prompts polymarket.com shows after deposit-wallet deploy:

1. ClobAuth — EOA-signed message to create or derive CLOB API keys.
2. Approve Tokens — DepositWallet.Batch signing for the 2-call UI token
   approval batch: pUSD -> CTF and USDC.e -> CollateralOnramp.

Use this when the wallet is already deployed but the UI still shows
"Enable Trading" or "Approve Tokens". If relayer credentials are missing,
Polygolem signs SIWE locally, registers the profile if needed, mints and
persists the V2 relayer key, then continues automatically.

The browser may still ask for a local ClobAuth signature because
polymarket.com stores browser-local API state; this command prepares
Polygolem's headless trading path and submits the on-chain deposit-wallet
approvals.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			wallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			if dryRun {
				rc, err := relayerClientFromEnv()
				if err != nil {
					return fmt.Errorf("init relayer client: %w", err)
				}
				result, err := dryRunEnableTradingSigns(cmd.Context(), key, owner, wallet, rc)
				if err != nil {
					return err
				}
				return printJSON(cmd, result)
			}
			rc, authResult, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			result, err := submitEnableTradingSigns(cmd.Context(), key, owner, wallet, rc)
			if err != nil {
				return err
			}
			result["owner"] = owner
			result["depositWallet"] = wallet
			result["auth"] = authResult
			return printJSON(cmd, result)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "build and validate typed data without signing, creating API keys, or submitting approvals")
	return cmd
}

func depositWalletFundCmd(jsonOut bool) *cobra.Command {
	var amountPUSD string
	var rpcURL string
	cmd := &cobra.Command{
		Use:   "fund",
		Short: "Transfer pUSD from EOA to the deposit wallet",
		Long: `Send pUSD from the EOA to the deposit wallet address via direct ERC-20 transfer.

--amount is in pUSD (e.g. "0.71" for 0.71 pUSD). Uses 6 decimals internally.
Requires POL for gas on Polygon.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			wallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			if strings.TrimSpace(amountPUSD) == "" {
				return fmt.Errorf("--amount is required (pUSD to transfer, e.g. 0.71)")
			}
			amountFloat, err := parsePUSDAmount(amountPUSD)
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}
			if amountFloat.Sign() <= 0 {
				return fmt.Errorf("amount must be positive")
			}
			txHash, err := rpc.TransferPUSD(cmd.Context(), key, wallet, amountFloat, rpcURL)
			if err != nil {
				return fmt.Errorf("transfer pUSD: %w", err)
			}
			return printJSON(cmd, map[string]string{
				"txHash": txHash,
				"from":   owner,
				"to":     wallet,
				"amount": amountPUSD,
			})
		},
	}
	cmd.Flags().StringVar(&amountPUSD, "amount", "", "pUSD amount to transfer (e.g. 0.71)")
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Polygon RPC URL (default: public node)")
	return cmd
}

// depositWalletSwapCmd swaps native POL on the EOA into pUSD via Uniswap V3
// (multihop WMATIC → USDC.e → pUSD, both legs at 0.05% fee tier). The pUSD
// lands on the EOA; chain `polygolem deposit-wallet fund --amount X` after
// to move it into the deposit wallet.
func depositWalletSwapCmd(jsonOut bool) *cobra.Command {
	var amountPUSDOut string
	var maxPOLIn string
	var rpcURL string
	cmd := &cobra.Command{
		Use:   "swap-pol-pusd",
		Short: "Swap native POL into an exact amount of pUSD via Uniswap V3",
		Long: `Swap native POL on the EOA into exactly --out-pusd of pUSD via Uniswap V3
on Polygon (multihop WMATIC → USDC.e → pUSD, 0.05% fee per leg). Excess POL
is refunded by the router via multicall(refundETH).

The pUSD lands on the EOA. Use 'polygolem deposit-wallet fund --amount X'
afterwards to move pUSD into the deposit wallet.

--out-pusd is the exact pUSD amount to receive (e.g. "0.72" for 0.72 pUSD).
--max-pol-in caps the POL the router may consume (e.g. "10" for 10 POL).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			if strings.TrimSpace(amountPUSDOut) == "" {
				return fmt.Errorf("--out-pusd is required (pUSD to receive, e.g. 0.72)")
			}
			if strings.TrimSpace(maxPOLIn) == "" {
				return fmt.Errorf("--max-pol-in is required (max POL to spend, e.g. 10)")
			}
			outPUSD, err := parsePUSDAmount(amountPUSDOut)
			if err != nil {
				return fmt.Errorf("invalid --out-pusd: %w", err)
			}
			if outPUSD.Sign() <= 0 {
				return fmt.Errorf("--out-pusd must be positive")
			}
			maxPOLWei, err := parsePOLAmount(maxPOLIn)
			if err != nil {
				return fmt.Errorf("invalid --max-pol-in: %w", err)
			}
			if maxPOLWei.Sign() <= 0 {
				return fmt.Errorf("--max-pol-in must be positive")
			}
			txHash, err := rpc.SwapPOLForExactPUSD(cmd.Context(), key, outPUSD, maxPOLWei, rpcURL)
			if err != nil {
				return fmt.Errorf("swap POL→pUSD: %w", err)
			}
			return printJSON(cmd, map[string]string{
				"txHash":        txHash,
				"recipient":     signer.Address(),
				"amountPUSDOut": amountPUSDOut,
				"maxPOLIn":      maxPOLIn,
			})
		},
	}
	cmd.Flags().StringVar(&amountPUSDOut, "out-pusd", "", "exact pUSD amount to receive (e.g. 0.72)")
	cmd.Flags().StringVar(&maxPOLIn, "max-pol-in", "", "max POL the router may consume (e.g. 10)")
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Polygon RPC URL (default: public node)")
	return cmd
}

// parsePOLAmount converts a human POL string (e.g. "10", "0.5") to wei
// (18-decimal *big.Int).
func parsePOLAmount(s string) (*big.Int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty amount")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid POL amount %q", s)
	}
	whole, ok := new(big.Int).SetString(parts[0], 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer part: %s", parts[0])
	}
	weiPerPOL := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	result := new(big.Int).Mul(whole, weiPerPOL)
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 18 {
			frac = frac[:18]
		}
		// pad to 18 digits
		for len(frac) < 18 {
			frac += "0"
		}
		fracInt, ok := new(big.Int).SetString(frac, 10)
		if !ok {
			return nil, fmt.Errorf("invalid fractional part: %s", parts[1])
		}
		result.Add(result, fracInt)
	}
	return result, nil
}

func depositWalletOnboardCmd(jsonOut bool) *cobra.Command {
	var skipDeploy bool
	var skipApprove bool
	var skipEnableTrading bool
	var fundAmount string
	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Full deposit wallet onboarding: deploy + approve + enable trading + fund",
		Long: `Run the complete deposit wallet setup sequence:

1. Derive the deterministic deposit wallet address
2. Deploy via WALLET-CREATE (skip with --skip-deploy if already deployed)
3. Submit the 10-call approval batch for trading and V2 settlement adapters
   (skip with --skip-approve)
4. Sign ClobAuth and submit the 2-call UI Enable Trading approval batch
   (skip with --skip-enable-trading)
5. Transfer pUSD from EOA to deposit wallet (requires --fund-amount)

After onboarding, sync CLOB:
  polygolem clob update-balance --asset-type collateral

If relayer credentials are missing, Polygolem signs SIWE locally, registers
the profile if needed, mints and persists the V2 relayer key, then continues
automatically.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			wallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			rc, authResult, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			result := map[string]interface{}{
				"owner":         owner,
				"depositWallet": wallet,
				"auth":          authResult,
			}

			if !skipDeploy {
				deployed, err := depositWalletDeployed(cmd.Context(), rc, owner, wallet)
				if err != nil {
					return fmt.Errorf("check deployed: %w", err)
				}
				if !deployed {
					tx, err := rc.SubmitWalletCreate(cmd.Context(), owner)
					if err != nil {
						return fmt.Errorf("WALLET-CREATE: %w", err)
					}
					final, err := rc.PollTransaction(cmd.Context(), tx.TransactionID, 50, 2*time.Second)
					if err != nil {
						return fmt.Errorf("poll WALLET-CREATE: %w", err)
					}
					result["deploy"] = map[string]string{
						"transactionID": final.TransactionID,
						"state":         final.State,
					}
				} else {
					result["deploy"] = "already_deployed"
				}
			}

			if !skipApprove {
				calls := append(relayer.BuildApprovalCalls(), relayer.BuildAdapterApprovalCalls()...)
				nonce, err := rc.GetNonce(cmd.Context(), owner)
				if err != nil {
					return fmt.Errorf("fetch nonce: %w", err)
				}
				dl := relayer.BuildDeadline(240)
				sig, err := relayer.SignWalletBatch(signer, wallet, nonce, dl, calls)
				if err != nil {
					return fmt.Errorf("sign batch: %w", err)
				}
				tx, err := rc.SubmitWalletBatch(cmd.Context(), owner, wallet, nonce, sig, dl, calls)
				if err != nil {
					return fmt.Errorf("submit approval batch: %w", err)
				}
				result["approve"] = map[string]interface{}{
					"transactionID": tx.TransactionID,
					"state":         tx.State,
					"callCount":     len(calls),
					"includes":      []string{"trading", "settlement-adapters"},
				}
			}

			if !skipEnableTrading {
				enableResult, err := submitEnableTradingSigns(cmd.Context(), key, owner, wallet, rc)
				if err != nil {
					return fmt.Errorf("enable trading signs: %w", err)
				}
				result["enableTrading"] = enableResult
			}

			if strings.TrimSpace(fundAmount) != "" {
				amountFloat, err := parsePUSDAmount(fundAmount)
				if err != nil {
					return fmt.Errorf("invalid --fund-amount: %w", err)
				}
				txHash, err := rpc.TransferPUSD(cmd.Context(), key, wallet, amountFloat, "")
				if err != nil {
					return fmt.Errorf("fund transfer: %w", err)
				}
				result["fund"] = map[string]string{
					"txHash": txHash,
					"from":   owner,
					"to":     wallet,
					"amount": fundAmount,
				}
			}

			result["nextSteps"] = []string{
				"Run: polygolem clob update-balance --asset-type collateral",
				"Verify: polygolem clob balance --asset-type collateral",
			}

			warnIfNoDepositKey(cmd.Context(), cmd.ErrOrStderr(), key)

			return printJSON(cmd, result)
		},
	}
	cmd.Flags().BoolVar(&skipDeploy, "skip-deploy", false, "skip WALLET-CREATE (wallet already deployed)")
	cmd.Flags().BoolVar(&skipApprove, "skip-approve", false, "skip approval batch")
	cmd.Flags().BoolVar(&skipEnableTrading, "skip-enable-trading", false, "skip ClobAuth and UI Enable Trading token approval signs")
	cmd.Flags().StringVar(&fundAmount, "fund-amount", "", "pUSD amount to transfer from EOA to deposit wallet (e.g. 0.71)")
	return cmd
}

func dryRunEnableTradingSigns(ctx context.Context, privateKey, owner, wallet string, rc *relayer.Client) (map[string]interface{}, error) {
	clobTD, err := sdkenabletrading.BuildClobAuthTypedData(sdkenabletrading.ClobAuthParams{
		Address:   owner,
		ChainID:   sdkenabletrading.PolygonChainID,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
		Nonce:     0,
	})
	if err != nil {
		return nil, err
	}
	if _, err := sdkenabletrading.HashClobAuthTypedData(clobTD); err != nil {
		return nil, err
	}
	nonce, err := rc.GetNonce(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("fetch nonce: %w", err)
	}
	calls := sdkenabletrading.BuildEnableTradingApprovalCalls()
	deadline := relayer.BuildDeadline(240)
	batchTD, err := sdkenabletrading.BuildEnableTradingApprovalBatchTypedData(sdkenabletrading.ApprovalBatchParams{
		DepositWallet: wallet,
		ChainID:       sdkenabletrading.PolygonChainID,
		Nonce:         nonce,
		Deadline:      deadline,
		Calls:         calls,
	})
	if err != nil {
		return nil, err
	}
	if _, err := sdkenabletrading.SignDepositWalletApprovalBatch(privateKey, batchTD); err != nil {
		return nil, fmt.Errorf("validate DepositWallet approval signature: %w", err)
	}
	return map[string]interface{}{
		"owner":                     owner,
		"depositWallet":             wallet,
		"dryRun":                    true,
		"clobAuthBuildable":         true,
		"approvalBatchSignable":     true,
		"tokenApprovalCallCount":    len(calls),
		"tokenApprovalTargets":      []string{contracts.PUSD, contracts.USDCE},
		"tokenApprovalSpenders":     []string{contracts.CTF, contracts.CollateralOnramp},
		"wouldCreateOrDeriveAPIKey": true,
		"wouldSubmitApprovals":      true,
	}, nil
}

func submitEnableTradingSigns(ctx context.Context, privateKey, owner, wallet string, rc *relayer.Client) (map[string]interface{}, error) {
	clobTD, err := sdkenabletrading.BuildClobAuthTypedData(sdkenabletrading.ClobAuthParams{
		Address:   owner,
		ChainID:   sdkenabletrading.PolygonChainID,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
		Nonce:     0,
	})
	if err != nil {
		return nil, err
	}
	if _, err := sdkenabletrading.SignClobAuthTypedData(privateKey, clobTD); err != nil {
		return nil, err
	}
	c := sdkclob.NewClient(sdkclob.Config{BaseURL: clobBaseURLFromEnv()})
	if _, err := c.CreateOrDeriveAPIKey(ctx, privateKey); err != nil {
		return nil, fmt.Errorf("create or derive CLOB API key: %w", err)
	}

	calls := sdkenabletrading.BuildEnableTradingApprovalCalls()
	nonce, err := rc.GetNonce(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("fetch nonce: %w", err)
	}
	deadline := relayer.BuildDeadline(240)
	batchTD, err := sdkenabletrading.BuildEnableTradingApprovalBatchTypedData(sdkenabletrading.ApprovalBatchParams{
		DepositWallet: wallet,
		ChainID:       sdkenabletrading.PolygonChainID,
		Nonce:         nonce,
		Deadline:      deadline,
		Calls:         calls,
	})
	if err != nil {
		return nil, err
	}
	sig, err := sdkenabletrading.SignDepositWalletApprovalBatch(privateKey, batchTD)
	if err != nil {
		return nil, err
	}
	tx, err := rc.SubmitWalletBatch(ctx, owner, wallet, nonce, sig, deadline, calls)
	if err != nil {
		return nil, fmt.Errorf("submit UI Enable Trading approval batch: %w", err)
	}
	return map[string]interface{}{
		"clobAuthSigned":          true,
		"apiKeysCreatedOrDerived": true,
		"tokenApprovalsSigned":    true,
		"tokenApprovalsSubmitted": true,
		"transactionID":           tx.TransactionID,
		"state":                   tx.State,
		"callCount":               len(calls),
		"includes":                []string{"clob-auth", "ui-token-approvals"},
	}, nil
}

type clobCredentialReadiness struct {
	Ready  bool
	Source string
}

func validateEnableTradingReadiness(ctx context.Context, privateKey, wallet, nonce string, clobCredentials clobCredentialReadiness, rpcURL string) (map[string]interface{}, error) {
	signer, err := auth.NewPrivateKeySigner(privateKey, 137)
	if err != nil {
		return nil, fmt.Errorf("init signer: %w", err)
	}
	if _, ok := new(big.Int).SetString(strings.TrimSpace(nonce), 10); !ok {
		return nil, fmt.Errorf("wallet nonce unavailable for enable-trading validation: %s", nonce)
	}
	clobTD, err := sdkenabletrading.BuildClobAuthTypedData(sdkenabletrading.ClobAuthParams{
		Address:   signer.Address(),
		ChainID:   sdkenabletrading.PolygonChainID,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
		Nonce:     0,
	})
	if err != nil {
		return nil, err
	}
	if _, err := sdkenabletrading.SignClobAuthTypedData(privateKey, clobTD); err != nil {
		return nil, fmt.Errorf("sign ClobAuth validation payload: %w", err)
	}
	calls := sdkenabletrading.BuildEnableTradingApprovalCalls()
	if err := sdkenabletrading.ValidateEnableTradingApprovalCalls(calls); err != nil {
		return nil, err
	}
	deadline := relayer.BuildDeadline(240)
	batchTD, err := sdkenabletrading.BuildEnableTradingApprovalBatchTypedData(sdkenabletrading.ApprovalBatchParams{
		DepositWallet: wallet,
		ChainID:       sdkenabletrading.PolygonChainID,
		Nonce:         nonce,
		Deadline:      deadline,
		Calls:         calls,
	})
	if err != nil {
		return nil, err
	}
	if _, err := sdkenabletrading.SignDepositWalletApprovalBatch(privateKey, batchTD); err != nil {
		return nil, fmt.Errorf("sign DepositWallet approval validation payload: %w", err)
	}

	approvalChecks, tokenApprovalsReady, err := enableTradingApprovalChecks(ctx, wallet, rpcURL)
	if err != nil {
		return nil, err
	}
	ready := clobCredentials.Ready && tokenApprovalsReady
	out := map[string]interface{}{
		"clobAuthSignable":          true,
		"clobCredentialsConfigured": clobCredentials.Source == "env",
		"clobCredentialsReady":      clobCredentials.Ready,
		"clobCredentialsSource":     clobCredentials.Source,
		"approvalCallsValid":        true,
		"approvalBatchSignable":     true,
		"tokenApprovalsReady":       tokenApprovalsReady,
		"approvalChecks":            approvalChecks,
		"ready":                     ready,
	}
	if !ready {
		out["nextAction"] = "polygolem deposit-wallet onboard"
	}
	return out, nil
}

func clobCredentialsReadyForCLI(ctx context.Context, privateKey string) clobCredentialReadiness {
	key, ok := clobL2CredentialsFromEnv()
	if ok && key.Validate() == nil {
		return clobCredentialReadiness{Ready: true, Source: "env"}
	}
	c := sdkclob.NewClient(sdkclob.Config{BaseURL: clobBaseURLFromEnv()})
	derived, err := c.DeriveAPIKey(ctx, privateKey)
	if err != nil {
		return clobCredentialReadiness{Ready: false, Source: "missing"}
	}
	if strings.TrimSpace(derived.Key) == "" || strings.TrimSpace(derived.Secret) == "" || strings.TrimSpace(derived.Passphrase) == "" {
		return clobCredentialReadiness{Ready: false, Source: "missing"}
	}
	return clobCredentialReadiness{Ready: true, Source: "derived"}
}

func enableTradingApprovalChecks(ctx context.Context, wallet, rpcURL string) ([]map[string]interface{}, bool, error) {
	required := []struct {
		label   string
		token   string
		spender string
	}{
		{label: "pusd_ctf", token: contracts.PUSD, spender: contracts.CTF},
		{label: "usdce_collateral_onramp", token: contracts.USDCE, spender: contracts.CollateralOnramp},
	}
	checks := make([]map[string]interface{}, 0, len(required))
	allReady := true
	for _, row := range required {
		allowance, err := rpc.ERC20Allowance(ctx, row.token, wallet, row.spender, rpcURL)
		if err != nil {
			return nil, false, fmt.Errorf("check ERC20 allowance %s: %w", row.label, err)
		}
		ready := allowance.Sign() > 0
		if !ready {
			allReady = false
		}
		checks = append(checks, map[string]interface{}{
			"name":      row.label,
			"token":     row.token,
			"spender":   row.spender,
			"allowance": allowance.String(),
			"ready":     ready,
		})
	}
	return checks, allReady, nil
}

func clobBaseURLFromEnv() string {
	if value := firstEnv("POLYMARKET_CLOB_URL", "CLOB_URL"); value != "" {
		return value
	}
	return clobBaseURL
}

func depositWalletDeployed(ctx context.Context, rc *relayer.Client, owner string, wallet string) (bool, error) {
	relayerDeployed, err := rc.IsDeployed(ctx, owner)
	if err != nil {
		return false, err
	}
	if relayerDeployed {
		return true, nil
	}
	status, err := contracts.DepositWalletDeployed(ctx, wallet, os.Getenv("POLYGON_RPC_URL"))
	if err != nil {
		return false, fmt.Errorf("relayer reported not deployed and on-chain code check failed: %w", err)
	}
	return status.Deployed, nil
}

func deploymentStatusSource(relayerDeployed bool, onchainDeployed bool) string {
	if relayerDeployed {
		return "relayer"
	}
	if onchainDeployed {
		return "polygon_code"
	}
	return "relayer_and_polygon_code"
}

func parsePUSDAmount(s string) (*big.Int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty amount")
	}
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		return nil, fmt.Errorf("amount must be unsigned decimal")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid amount %q", s)
	}
	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}
	if !decimalDigitsOnly(wholePart) {
		return nil, fmt.Errorf("invalid integer part: %s", parts[0])
	}
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
		if !decimalDigitsOnly(fracPart) {
			return nil, fmt.Errorf("invalid fractional part: %s", fracPart)
		}
		for len(fracPart) > 6 && strings.HasSuffix(fracPart, "0") {
			fracPart = strings.TrimSuffix(fracPart, "0")
		}
		if len(fracPart) > 6 {
			return nil, fmt.Errorf("pUSD supports at most 6 decimals")
		}
	}
	for len(fracPart) < 6 {
		fracPart += "0"
	}
	whole, ok := new(big.Int).SetString(wholePart, 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer part: %s", wholePart)
	}
	result := new(big.Int).Mul(whole, big.NewInt(1000000))
	if fracPart != "" {
		frac, ok := new(big.Int).SetString(fracPart, 10)
		if !ok {
			return nil, fmt.Errorf("invalid fractional part: %s", fracPart)
		}
		result.Add(result, frac)
	}
	return result, nil
}

func decimalDigitsOnly(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func builderConfigFromEnv() (auth.BuilderConfig, error) {
	bc := auth.BuilderConfig{
		Key:        firstEnv("POLYMARKET_BUILDER_API_KEY", "BUILDER_API_KEY"),
		Secret:     firstEnv("POLYMARKET_BUILDER_SECRET", "BUILDER_SECRET"),
		Passphrase: firstEnv("POLYMARKET_BUILDER_PASSPHRASE", "BUILDER_PASS_PHRASE"),
	}
	if !bc.Valid() {
		return auth.BuilderConfig{}, fmt.Errorf("builder credentials not configured: set POLYMARKET_BUILDER_API_KEY, POLYMARKET_BUILDER_SECRET, and POLYMARKET_BUILDER_PASSPHRASE (or BUILDER_API_KEY / BUILDER_SECRET / BUILDER_PASS_PHRASE)")
	}
	return bc, nil
}

type relayerClientAuthResult struct {
	Source      string `json:"source"`
	AutoMinted  bool   `json:"autoMinted"`
	PersistedTo string `json:"persistedTo,omitempty"`
}

// relayerClientFromEnv builds a relayer.Client from environment variables.
// Prefers the V2 plain-header scheme (RELAYER_API_KEY +
// RELAYER_API_KEY_ADDRESS, generated by `polygolem auth login`)
// when both are present in either process env or a known Polygolem env file;
// falls back to the legacy POLY_BUILDER_* HMAC scheme otherwise.
func relayerClientFromEnv() (*relayer.Client, error) {
	relayerURL := strings.TrimSpace(os.Getenv("POLYMARKET_RELAYER_URL"))
	if relayerURL == "" {
		relayerURL = defaultRelayerURL
	}

	if key, _, ok := relayerV2KeyFromProcessEnv(); ok {
		return relayer.NewV2(relayerURL, key, 137)
	}

	if bc, err := builderConfigFromEnv(); err == nil {
		return relayer.New(relayerURL, bc, 137)
	}
	if key, _, ok := relayerV2KeyFromFiles(); ok {
		return relayer.NewV2(relayerURL, key, 137)
	}
	return nil, fmt.Errorf("builder credentials not configured: set POLYMARKET_BUILDER_API_KEY, POLYMARKET_BUILDER_SECRET, and POLYMARKET_BUILDER_PASSPHRASE (or BUILDER_API_KEY / BUILDER_SECRET / BUILDER_PASS_PHRASE)")
}

func relayerClientForAutomation(ctx context.Context, stderr io.Writer, privateKey string) (*relayer.Client, relayerClientAuthResult, error) {
	relayerURL := strings.TrimSpace(os.Getenv("POLYMARKET_RELAYER_URL"))
	if relayerURL == "" {
		relayerURL = defaultRelayerURL
	}
	if key, source, ok := relayerV2KeyFromProcessEnv(); ok {
		client, err := relayer.NewV2(relayerURL, key, 137)
		return client, relayerClientAuthResult{Source: source}, err
	}
	if bc, err := builderConfigFromEnv(); err == nil {
		client, err := relayer.New(relayerURL, bc, 137)
		return client, relayerClientAuthResult{Source: "legacy-builder-env"}, err
	}
	if key, source, ok := relayerV2KeyFromFiles(); ok {
		client, err := relayer.NewV2(relayerURL, key, 137)
		return client, relayerClientAuthResult{Source: source}, err
	}

	key, target, err := mintRelayerV2KeyForAutomation(ctx, stderr, privateKey)
	if err != nil {
		return nil, relayerClientAuthResult{}, err
	}
	client, err := relayer.NewV2(relayerURL, key, 137)
	return client, relayerClientAuthResult{
		Source:      "auto-siwe-login",
		AutoMinted:  true,
		PersistedTo: target,
	}, err
}

func mintRelayerV2KeyForAutomation(ctx context.Context, stderr io.Writer, privateKey string) (relayer.V2APIKey, string, error) {
	signer, err := auth.NewPrivateKeySigner(privateKey, 137)
	if err != nil {
		return relayer.V2APIKey{}, "", fmt.Errorf("init signer: %w", err)
	}
	gammaURL := firstNonEmptyCLI(os.Getenv("POLYMARKET_GAMMA_URL"), defaultGammaBaseURL)
	relayerURL := firstNonEmptyCLI(os.Getenv("POLYMARKET_RELAYER_URL"), defaultRelayerV2BaseURL)
	if stderr != nil {
		fmt.Fprintf(stderr, "No relayer credentials loaded; running headless auth login automatically...\n")
	}
	loginCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	session, err := auth.NewSIWESession(signer, gammaURL)
	if err != nil {
		return relayer.V2APIKey{}, "", fmt.Errorf("new siwe session: %w", err)
	}
	if err := session.Login(loginCtx); err != nil {
		return relayer.V2APIKey{}, "", fmt.Errorf("siwe login: %w", err)
	}
	maker, err := auth.MakerAddressForSignatureType(signer.Address(), 137, 3)
	if err != nil {
		return relayer.V2APIKey{}, "", fmt.Errorf("derive deposit wallet maker: %w", err)
	}
	body := gamma.NewCreateProfileRequest(
		signer.Address(),
		maker,
		"metamask",
		time.Now().UnixMilli(),
	)
	if _, err := gamma.CreateProfile(loginCtx, session.HTTPClient(), gammaURL, body); err != nil && !strings.Contains(err.Error(), "HTTP 409") {
		return relayer.V2APIKey{}, "", fmt.Errorf("create profile: %w", err)
	}
	key, err := relayer.MintV2APIKey(loginCtx, session.HTTPClient(), relayerURL)
	if err != nil {
		return relayer.V2APIKey{}, "", fmt.Errorf("mint v2 relayer key: %w", err)
	}
	target := relayerEnvFileCandidates()[0]
	abs, err := filepath.Abs(target)
	if err != nil {
		return relayer.V2APIKey{}, "", fmt.Errorf("resolve relayer env file: %w", err)
	}
	if err := persistRelayerV2Key(abs, key, true); err != nil {
		return relayer.V2APIKey{}, "", fmt.Errorf("persist relayer key: %w", err)
	}
	_ = os.Setenv("RELAYER_API_KEY", key.Key)
	_ = os.Setenv("RELAYER_API_KEY_ADDRESS", key.Address)
	if stderr != nil {
		fmt.Fprintf(stderr, "Relayer credentials minted and saved to %s\n", abs)
	}
	return key, abs, nil
}

func relayerV2KeyFromProcessEnv() (relayer.V2APIKey, string, bool) {
	v2Key := strings.TrimSpace(os.Getenv("RELAYER_API_KEY"))
	v2Addr := strings.TrimSpace(os.Getenv("RELAYER_API_KEY_ADDRESS"))
	if v2Key != "" && v2Addr != "" {
		return relayer.V2APIKey{Key: v2Key, Address: v2Addr}, "env", true
	}
	return relayer.V2APIKey{}, "", false
}

func relayerV2KeyFromFiles() (relayer.V2APIKey, string, bool) {
	for _, path := range relayerEnvFileCandidates() {
		values, ok := readSimpleEnvFile(path)
		if !ok {
			continue
		}
		v2Key := strings.TrimSpace(values["RELAYER_API_KEY"])
		v2Addr := strings.TrimSpace(values["RELAYER_API_KEY_ADDRESS"])
		if v2Key != "" && v2Addr != "" {
			return relayer.V2APIKey{Key: v2Key, Address: v2Addr}, path, true
		}
	}
	return relayer.V2APIKey{}, "", false
}

func relayerEnvFileCandidates() []string {
	if override := strings.TrimSpace(os.Getenv("POLYGOLEM_RELAYER_ENV_FILE")); override != "" {
		return []string{override}
	}
	return []string{
		defaultRelayerEnvFile,
		"../.env.relayer-v2",
		".env.relayer-v2",
		"../go-bot/.env",
		"../.env",
		".env",
	}
}

func readSimpleEnvFile(path string) (map[string]string, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	out := make(map[string]string)
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key != "" {
			out[key] = value
		}
	}
	return out, true
}

func requirePrivateKey() (string, error) {
	key := strings.TrimSpace(os.Getenv("POLYMARKET_PRIVATE_KEY"))
	if key == "" {
		return "", fmt.Errorf("POLYMARKET_PRIVATE_KEY is required")
	}
	return key, nil
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	return ""
}

func printJSON(cmd *cobra.Command, v interface{}) error {
	return writeCommandJSON(cmd, v)
}

// dataAPIClient builds a pkg/data.Client honoring an optional override env.
// Defaults to the production Data API URL.
func dataAPIClient() *data.Client {
	base := firstEnv("POLYMARKET_DATA_API_URL")
	if base == "" {
		base = defaultDataAPIURL
	}
	return data.NewClient(data.Config{BaseURL: base})
}

// depositWalletSettlementStatusCmd is the read-only readiness gate live
// trading uses before it is allowed to place more orders. It checks the
// official V2 settlement path only: deployed deposit wallet, relayer
// credentials, Data API reachability, and CTF approvals for both V2
// collateral adapters.
func depositWalletSettlementStatusCmd(jsonOut bool) *cobra.Command {
	var rpcURL string
	cmd := &cobra.Command{
		Use:   "settlement-status",
		Short: "Check whether the deposit wallet is ready to redeem V2 winners",
		Long: `Read-only settlement readiness gate for V2 deposit-wallet trading.

Checks:
  - Deposit wallet has bytecode on Polygon
  - Polymarket relayer credentials are configured
  - Data API positions can be queried for the deposit wallet
  - CTF.setApprovalForAll(wallet, CtfCollateralAdapter) is true
  - CTF.setApprovalForAll(wallet, NegRiskCtfCollateralAdapter) is true

This command does not sign, submit, approve, redeem, or try a fallback. V2
deposit-wallet settlement is relayer + collateral adapter only: no direct EOA
submission path, no raw ConditionalTokens path, and no SAFE/PROXY shortcut.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()
			wallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			_, relayerErr := relayerClientFromEnv()
			status, err := settlement.CheckReadiness(cmd.Context(), dataAPIClient(), owner, wallet, settlement.ReadinessOptions{
				RPCURL:            firstNonEmptyCLI(rpcURL, os.Getenv("POLYGON_RPC_URL")),
				RelayerConfigured: relayerErr == nil,
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, status)
		},
	}
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Polygon RPC URL for code and adapter-approval checks (default: POLYGON_RPC_URL or public node)")
	return cmd
}

// depositWalletRedeemableCmd lists redeemable positions for the deposit
// wallet derived from POLYMARKET_PRIVATE_KEY. Read-only — no signing.
// Positions live in the deposit wallet, not the EOA, so the Data API
// `user` parameter must be the deposit wallet address.
func depositWalletRedeemableCmd(jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeemable",
		Short: "List redeemable positions held by the deposit wallet",
		Long: `Read-only list of positions where the Data API redeemable=true
flag is set. The 'user' parameter is the deposit wallet (not the EOA),
since POLY_1271 positions live in the wallet.

Use this before running 'redeem' to see what would be submitted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			wallet, err := auth.MakerAddressForSignatureType(signer.Address(), 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			rows, err := settlement.FindRedeemable(cmd.Context(), dataAPIClient(), wallet)
			if err != nil {
				return fmt.Errorf("find redeemable: %w", err)
			}
			return printJSON(cmd, map[string]interface{}{
				"depositWallet": wallet,
				"count":         len(rows),
				"positions":     rows,
			})
		},
	}
	return cmd
}

// depositWalletRedeemCmd builds the V2 redeem WALLET batch. Dry-run by
// default; --submit + --confirm REDEEM_WINNERS together authorize the
// live signing path. Pre-checks CTF.isApprovedForAll for each adapter
// the to-redeem set requires and refuses to sign if any approval is
// missing, pointing the operator at `approve-adapters`.
func depositWalletRedeemCmd(jsonOut bool) *cobra.Command {
	var submit bool
	var confirm string
	var limit int
	var rpcURL string
	cmd := &cobra.Command{
		Use:   "redeem",
		Short: "Redeem winning deposit-wallet positions via the V2 collateral adapter",
		Long: `Builds a WALLET batch that calls redeemPositions on the V2 collateral
adapter (CtfCollateralAdapter for binary markets, NegRiskCtfCollateralAdapter
for neg-risk). The adapter pulls the wallet's CTF tokens, redeems through
the legacy CT with USDC.e, wraps the proceeds back into pUSD, and sends pUSD
to the deposit wallet.

Without --submit, prints the calldata JSON for review.
With --submit, the operator must also pass --confirm REDEEM_WINNERS to
authorize the live-money WALLET batch.

Pre-check: requires CTF.setApprovalForAll(wallet, adapter) = true for
every adapter targeted by the redeem set. If any approval is missing,
fails closed with a pointer to 'deposit-wallet approve-adapters'.

NOTE: The V2 deposit-wallet redeem path is non-negotiable: the owner signs an
EIP-712 WALLET batch, the relayer submits it through the deposit-wallet
factory, and the wallet call targets CtfCollateralAdapter or
NegRiskCtfCollateralAdapter. If Polymarket's relayer rejects adapter approval
or redeem calls with "not in the allowed list", first verify the adapter
addresses against Polymarket's current contract reference; if they match, stop
and surface an upstream blocker. There is no direct EOA bypass, no raw
ConditionalTokens fallback, and no SAFE/PROXY shortcut for deposit-wallet
positions.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			signer, err := auth.NewPrivateKeySigner(key, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			wallet, err := auth.MakerAddressForSignatureType(signer.Address(), 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}
			rows, err := settlement.FindRedeemable(cmd.Context(), dataAPIClient(), wallet)
			if err != nil {
				return fmt.Errorf("find redeemable: %w", err)
			}
			if len(rows) == 0 {
				return printJSON(cmd, map[string]interface{}{
					"depositWallet": wallet,
					"state":         "nothing_to_redeem",
					"count":         0,
				})
			}

			// Build the calls so we can show the dry-run payload and
			// also know which adapters require the pre-check.
			calls := make([]relayer.DepositWalletCall, 0, len(rows))
			adaptersNeeded := make(map[string]struct{})
			for _, p := range rows {
				call, err := settlement.BuildRedeemCall(p)
				if err != nil {
					return fmt.Errorf("build call: %w", err)
				}
				calls = append(calls, call)
				adaptersNeeded[strings.ToLower(call.Target)] = struct{}{}
			}

			if !submit {
				return printJSON(cmd, map[string]interface{}{
					"depositWallet": wallet,
					"count":         len(rows),
					"positions":     rows,
					"calls":         calls,
					"path":          "relayer-adapter",
					"note":          "review calldata, then run with --submit --confirm REDEEM_WINNERS to sign and send",
				})
			}
			if confirm != "REDEEM_WINNERS" {
				return fmt.Errorf("--submit requires --confirm REDEEM_WINNERS (got %q)", confirm)
			}

			// Adapter approval pre-check (fail-closed).
			polygonRPC := firstEnv("POLYGON_RPC_URL")
			if rpcURL != "" {
				polygonRPC = rpcURL
			}
			missing := make([]string, 0, len(adaptersNeeded))
			for adapter := range adaptersNeeded {
				ok, err := rpc.IsApprovedForAll(cmd.Context(), contracts.CTF, wallet, adapter, polygonRPC)
				if err != nil {
					return fmt.Errorf("check isApprovedForAll(%s): %w", adapter, err)
				}
				if !ok {
					missing = append(missing, adapter)
				}
			}
			if len(missing) > 0 {
				return printJSON(cmd, map[string]interface{}{
					"ok":               false,
					"error":            "deposit wallet has not approved one or more V2 collateral adapters; run `polygolem deposit-wallet approve-adapters --submit --confirm APPROVE_ADAPTERS` first",
					"missingApprovals": missing,
					"depositWallet":    wallet,
				})
			}

			rc, _, err := relayerClientForAutomation(cmd.Context(), cmd.ErrOrStderr(), key)
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			result, err := settlement.SubmitRedeem(cmd.Context(), rc, key, rows, limit)
			if err != nil {
				if errors.Is(err, relayer.ErrRelayerAllowlistBlocked) {
					return printJSON(cmd, upstreamRelayerBlockJSON(wallet, "redeem", err))
				}
				return fmt.Errorf("submit redeem: %w", err)
			}
			return printJSON(cmd, map[string]interface{}{
				"transactionID": result.TransactionID,
				"state":         result.State,
				"wallet":        result.Wallet,
				"nonce":         result.Nonce,
				"deadline":      result.Deadline,
				"callCount":     result.CallCount,
				"redeemed":      result.Redeemed,
				"path":          "relayer-adapter",
				"proceedsToken": "pUSD",
			})
		},
	}
	cmd.Flags().BoolVar(&submit, "submit", false, "sign and submit the redeem batch (requires --confirm REDEEM_WINNERS)")
	cmd.Flags().StringVar(&confirm, "confirm", "", "live-money confirmation token; must be 'REDEEM_WINNERS' when --submit is set")
	cmd.Flags().IntVar(&limit, "limit", settlement.DefaultBatchLimit, "max positions per WALLET batch (deduplicated by conditionID)")
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Polygon RPC URL for the adapter-approval pre-check (default: POLYGON_RPC_URL or public node)")
	return cmd
}
