package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/spf13/cobra"
)

func newAuthStatusCommand(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var checkDepositKey bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check authentication readiness and API key status",
		Long: `Inspects the current POLYMARKET_PRIVATE_KEY and reports:
  - EOA address and deposit wallet address
  - Whether the deposit wallet is deployed
  - Whether EOA-owned and deposit-wallet-owned API keys exist
  - Whether the setup is ready for trading

Use --check-deposit-key to test whether the deposit-wallet-owned API key
is functional (makes a live network call). Without this flag, the check
is faster but may report a stale key as existing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKey := strings.TrimSpace(os.Getenv("POLYMARKET_PRIVATE_KEY"))
			if privateKey == "" {
				return fmt.Errorf("POLYMARKET_PRIVATE_KEY is required")
			}

			signer, err := auth.NewPrivateKeySigner(privateKey, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()

			depositWallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}

			var deployed bool
			if rc, err := relayerClientFromEnv(); err == nil {
				deployed, _ = rc.IsDeployed(cmd.Context(), owner)
			}

			eoaKeyExists := false
			c := clob.NewClient(clobBaseURL, nil)
			if _, err := c.DeriveAPIKey(cmd.Context(), privateKey); err == nil {
				eoaKeyExists = true
			}

			depositKeyExists := false
			if deployed && checkDepositKey {
				_, err := c.DeriveAPIKeyForAddress(cmd.Context(), privateKey, owner)
				depositKeyExists = (err == nil)
			}

			canTrade := deployed && depositKeyExists

			result := map[string]interface{}{
				"eoaAddress":                owner,
				"depositWallet":             depositWallet,
				"depositWalletDeployed":     deployed,
				"eoaApiKeyExists":           eoaKeyExists,
				"depositWalletApiKeyExists": depositKeyExists,
				"canTrade":                  canTrade,
			}

			if !canTrade {
				if !deployed {
					result["nextStep"] = "Run: polygolem deposit-wallet deploy --wait"
				} else if !depositKeyExists {
					result["nextStep"] = "New users: complete one-time browser setup (see docs/BROWSER-SETUP.md)"
				}
				result["help"] = "https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/BROWSER-SETUP.md"
			}

			return w.printJSON(cmd, result)
		},
	}

	cmd.Flags().BoolVar(&checkDepositKey, "check-deposit-key", false, "make a live network call to verify the deposit-wallet API key exists")
	return cmd
}

func warnIfNoDepositKey(ctx context.Context, stderr io.Writer, privateKey string) {
	signer, err := auth.NewPrivateKeySigner(privateKey, 137)
	if err != nil {
		return
	}
	owner := signer.Address()

	depositWallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
	if err != nil {
		return
	}

	var deployed bool
	if rc, err := relayerClientFromEnv(); err == nil {
		deployed, _ = rc.IsDeployed(ctx, owner)
	}
	if !deployed {
		return
	}

	c := clob.NewClient(clobBaseURL, nil)
	_, err = c.DeriveAPIKeyForAddress(ctx, privateKey, owner)
	if err == nil {
		return
	}

	fmt.Fprintf(stderr, "\n⚠️  WARNING: Deposit wallet %s is deployed but no deposit-wallet-owned API key found.\n", depositWallet)
	fmt.Fprintf(stderr, "   Deposit-wallet orders require a deposit-wallet-owned API key.\n")
	fmt.Fprintf(stderr, "   If you're a new user, complete the one-time browser setup:\n")
	fmt.Fprintf(stderr, "   → docs/BROWSER-SETUP.md\n")
	fmt.Fprintf(stderr, "   → https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/BROWSER-SETUP.md\n\n")
}

func newAuthExportKeyCommand(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var confirm bool

	cmd := &cobra.Command{
		Use:   "export-key",
		Short: "Display private key for wallet import (use with care)",
		Long: `Displays the current POLYMARKET_PRIVATE_KEY and derived addresses
in formats suitable for wallet import. This is useful when a bot/agent
generated the key and the user needs to import it into MetaMask/Rabby/etc.
for the one-time Polymarket browser signup.

SECURITY WARNING: The private key will be printed to your terminal.
Anyone with access to your screen or shell history can steal your funds.
Use this only in a secure environment and clear your terminal history after.

Recommended flow for bot-generated keys:
  1. Run this command in a secure terminal
  2. Import the private key into a temporary wallet (MetaMask mobile, fresh browser profile)
  3. Connect to polymarket.com and complete signup
  4. Remove the imported account from the wallet
  5. Clear terminal history: history -c && clear`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKey := strings.TrimSpace(os.Getenv("POLYMARKET_PRIVATE_KEY"))
			if privateKey == "" {
				return fmt.Errorf("POLYMARKET_PRIVATE_KEY is required")
			}

			if !confirm {
				return fmt.Errorf("this command prints your private key to the terminal; pass --confirm to proceed")
			}

			signer, err := auth.NewPrivateKeySigner(privateKey, 137)
			if err != nil {
				return fmt.Errorf("init signer: %w", err)
			}
			owner := signer.Address()

			depositWallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
			if err != nil {
				return fmt.Errorf("derive deposit wallet: %w", err)
			}

			stderr := cmd.ErrOrStderr()
			fmt.Fprintf(stderr, "\n⚠️  SECURITY WARNING: Private key exposed below. Clear your terminal history after.\n\n")

			return w.printJSON(cmd, map[string]string{
				"eoaAddress":    owner,
				"depositWallet": depositWallet,
				"privateKey":    privateKey,
				"warning":       "Clear terminal history after import: history -c && clear",
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "acknowledge security risk and print the private key")
	return cmd
}

func warnIfNoDepositKeySimple(stderr io.Writer, privateKey string) {
	signer, err := auth.NewPrivateKeySigner(privateKey, 137)
	if err != nil {
		return
	}
	owner := signer.Address()

	depositWallet, err := auth.MakerAddressForSignatureType(owner, 137, 3)
	if err != nil {
		return
	}

	fmt.Fprintf(stderr, "\nℹ️  Note: If this is your first time using Polymarket with this key,\n")
	fmt.Fprintf(stderr, "   you may need to complete a one-time browser login to create\n")
	fmt.Fprintf(stderr, "   your deposit-wallet-owned API key.\n")
	fmt.Fprintf(stderr, "   Deposit wallet: %s\n", depositWallet)
	fmt.Fprintf(stderr, "   See: docs/BROWSER-SETUP.md\n\n")
}
