package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
	"github.com/TrebuchetDynamics/polygolem/internal/rpc"
	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	"github.com/spf13/cobra"
)

const defaultRelayerURL = "https://relayer-v2.polymarket.com"

func depositWalletCmd(jsonOut bool) *cobra.Command {
	cmd := commandGroup("deposit-wallet", "Deposit wallet onboarding (WALLET-CREATE, nonce, batch, status)")

	cmd.AddCommand(depositWalletDeriveCmd(jsonOut))
	cmd.AddCommand(depositWalletDeployCmd(jsonOut))
	cmd.AddCommand(depositWalletDeployOnchainCmd(jsonOut))
	cmd.AddCommand(depositWalletNonceCmd(jsonOut))
	cmd.AddCommand(depositWalletStatusCmd(jsonOut))
	cmd.AddCommand(depositWalletBatchCmd(jsonOut))
	cmd.AddCommand(depositWalletApproveCmd(jsonOut))
	cmd.AddCommand(depositWalletFundCmd(jsonOut))
	cmd.AddCommand(depositWalletSwapCmd(jsonOut))
	cmd.AddCommand(depositWalletOnboardCmd(jsonOut))
	return cmd
}

// depositWalletDeployOnchainCmd lets the EOA call the deposit-wallet
// factory's deploy() function directly on Polygon — no relayer, no builder
// credentials. With --dry-run, only gas-estimation runs (no tx sent, no gas
// spent), which is enough to determine whether deploy() accepts EOA callers
// or is gated to admin/operator only.
func depositWalletDeployOnchainCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var dryRun bool
	var rpcURL string
	cmd := &cobra.Command{
		Use:   "deploy-onchain",
		Short: "Deploy the deposit wallet directly on-chain from the EOA (no relayer / no builder creds)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := requirePrivateKey()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			if dryRun {
				gas, err := rpc.DeployDepositWalletEstimate(ctx, key, rpcURL)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, map[string]interface{}{
					"dryRun":       true,
					"estimatedGas": gas,
					"deployGated":  false,
					"note":         "EstimateGas succeeded — deploy() accepts EOA callers; on-chain path is available without builder credentials",
				})
			}

			txHash, err := rpc.DeployDepositWalletOnchain(ctx, key, rpcURL)
			if err != nil {
				if txHash != "" {
					return fmt.Errorf("deploy-onchain failed (txHash=%s): %w", txHash, err)
				}
				return err
			}
			return w.printJSON(cmd, map[string]string{
				"txHash":         txHash,
				"factoryAddress": "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07",
				"status":         "deployed",
			})
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "estimate gas only; do not send a transaction (no gas spent)")
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Polygon RPC URL (default: public node)")
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
			rc, err := relayerClientFromEnv()
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
			rc, err := relayerClientFromEnv()
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
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check deposit wallet deployment status or transaction state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rc, err := relayerClientFromEnv()
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
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
			return printJSON(cmd, map[string]interface{}{
				"owner":                  owner,
				"depositWallet":          wallet,
				"deployed":               relayerDeployed || onchainDeployed,
				"relayerDeployed":        relayerDeployed,
				"onchainCodeDeployed":    onchainDeployed,
				"deploymentStatusSource": deploymentStatusSource(relayerDeployed, onchainDeployed),
				"walletNonce":            nonce,
			})
		},
	}
	cmd.Flags().StringVar(&txID, "tx-id", "", "transaction ID to poll")
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

			rc, err := relayerClientFromEnv()
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
	cmd.Flags().Int64Var(&deadline, "deadline", 240, "deadline seconds from now")
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
			rc, err := relayerClientFromEnv()
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
	var fundAmount string
	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Full deposit wallet onboarding: deploy + approve + fund",
		Long: `Run the complete deposit wallet setup sequence:

1. Derive the deterministic deposit wallet address
2. Deploy via WALLET-CREATE (skip with --skip-deploy if already deployed)
3. Submit the 6-call approval batch for pUSD and CTF (skip with --skip-approve)
4. Transfer pUSD from EOA to deposit wallet (requires --fund-amount)

After onboarding, sync CLOB:
  polygolem clob update-balance --asset-type collateral`,
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
			rc, err := relayerClientFromEnv()
			if err != nil {
				return fmt.Errorf("init relayer client: %w", err)
			}
			result := map[string]interface{}{
				"owner":         owner,
				"depositWallet": wallet,
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
				callsJSON, err := relayer.BuildApprovalCallsJSON()
				if err != nil {
					return fmt.Errorf("build approval calls: %w", err)
				}
				var calls []relayer.DepositWalletCall
				if err := json.Unmarshal([]byte(callsJSON), &calls); err != nil {
					return fmt.Errorf("parse approval calls: %w", err)
				}
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
				}
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
	cmd.Flags().StringVar(&fundAmount, "fund-amount", "", "pUSD amount to transfer from EOA to deposit wallet (e.g. 0.71)")
	return cmd
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
	bf, _, err := big.ParseFloat(s, 10, 6, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("parse amount: %w", err)
	}
	multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))
	result := new(big.Float).Mul(bf, multiplier)
	intResult, _ := result.Int(nil)
	return intResult, nil
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

// relayerClientFromEnv builds a relayer.Client from environment variables.
// Prefers the V2 plain-header scheme (RELAYER_API_KEY +
// RELAYER_API_KEY_ADDRESS, generated by `polygolem auth headless-onboard`)
// when both are present; falls back to the legacy POLY_BUILDER_* HMAC
// scheme otherwise.
func relayerClientFromEnv() (*relayer.Client, error) {
	relayerURL := strings.TrimSpace(os.Getenv("POLYMARKET_RELAYER_URL"))
	if relayerURL == "" {
		relayerURL = defaultRelayerURL
	}

	v2Key := strings.TrimSpace(os.Getenv("RELAYER_API_KEY"))
	v2Addr := strings.TrimSpace(os.Getenv("RELAYER_API_KEY_ADDRESS"))
	if v2Key != "" && v2Addr != "" {
		return relayer.NewV2(relayerURL, relayer.V2APIKey{Key: v2Key, Address: v2Addr}, 137)
	}

	bc, err := builderConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return relayer.New(relayerURL, bc, 137)
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
