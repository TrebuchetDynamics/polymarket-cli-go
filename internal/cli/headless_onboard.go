package cli

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/gamma"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
	"github.com/spf13/cobra"
)

// Headless onboarding wires SIWE login + V2 relayer key mint and persists
// the resulting RELAYER_API_KEY + RELAYER_API_KEY_ADDRESS to a 0600 env
// file. Companion to `polygolem builder auto`, which mints CLOB L2 creds.
//
// See docs/HEADLESS-BUILDER-KEYS-INVESTIGATION.md for the V2 split-cred
// model.

const (
	defaultGammaBaseURL     = "https://gamma-api.polymarket.com"
	defaultRelayerV2BaseURL = "https://relayer-v2.polymarket.com"
	defaultRelayerEnvFile   = "../go-bot/.env.relayer-v2"
)

func newAuthHeadlessOnboardCommand(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var envFile, gammaURL, relayerURL string
	var force, skipProfile bool
	var signatureType int

	cmd := &cobra.Command{
		Use:   "headless-onboard",
		Short: "Run SIWE login + mint V2 Relayer API Key (headless; does NOT create CLOB API key)",
		Long: `Headless replacement for the polymarket.com signup flow. Steps:

  1. Sign a Polymarket SIWE message with the EOA from POLYMARKET_PRIVATE_KEY.
  2. Trade the signature for a polymarket session cookie at
     gamma-api.polymarket.com/login.
  3. Register the EOA + maker (proxy or deposit wallet, per --signature-type)
     with gamma-api.polymarket.com/profiles. Skip with --skip-profile if the
     profile already exists.
  4. Mint a V2 Relayer API Key at relayer-v2.polymarket.com/relayer/api/auth.
  5. Persist {RELAYER_API_KEY, RELAYER_API_KEY_ADDRESS} to a 0600 env file.

The /profiles step is what registers the maker address with Polymarket's
backend so subsequent CLOB orders are accepted (without it, fresh EOAs
get HTTP 400 "maker address not allowed"). See BLOCKERS.md "CORRECTION
2026-05-08" for the captured signup flow this command replicates.`,
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

			target := envFile
			if target == "" {
				target = defaultRelayerEnvFile
			}
			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve env file path: %w", err)
			}

			gURL := strings.TrimSpace(gammaURL)
			if gURL == "" {
				gURL = defaultGammaBaseURL
			}
			rURL := strings.TrimSpace(relayerURL)
			if rURL == "" {
				rURL = defaultRelayerV2BaseURL
			}

			stderr := cmd.ErrOrStderr()
			fmt.Fprintf(stderr, "Running SIWE login at %s ...\n", gURL)

			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			session, err := auth.NewSIWESession(signer, gURL)
			if err != nil {
				return fmt.Errorf("new siwe session: %w", err)
			}
			if err := session.Login(ctx); err != nil {
				return fmt.Errorf("siwe login: %w", err)
			}
			cookies := session.CookiesFor(gURL + "/")
			fmt.Fprintf(stderr, "✓ SIWE login OK (%d cookies captured)\n", len(cookies))

			maker, err := auth.MakerAddressForSignatureType(signer.Address(), 137, signatureType)
			if err != nil {
				return fmt.Errorf("derive maker (sigtype %d): %w", signatureType, err)
			}
			fmt.Fprintf(stderr, "  maker (sigtype %d): %s\n", signatureType, maker)

			profileID := ""
			if !skipProfile {
				fmt.Fprintf(stderr, "Registering profile at %s/profiles ...\n", gURL)
				body := gamma.NewCreateProfileRequest(
					signer.Address(),
					maker,
					"metamask",
					time.Now().UnixMilli(),
				)
				profile, perr := gamma.CreateProfile(ctx, session.HTTPClient(), gURL, body)
				if perr != nil {
					if strings.Contains(perr.Error(), "HTTP 409") {
						fmt.Fprintf(stderr, "  profile already exists (409) — continuing\n")
					} else {
						return fmt.Errorf("create profile: %w", perr)
					}
				} else {
					profileID = profile.ID
					fmt.Fprintf(stderr, "✓ Profile registered (id=%s, proxyWallet=%s)\n", profile.ID, profile.ProxyWallet)
				}
			} else {
				fmt.Fprintf(stderr, "  --skip-profile set — not calling /profiles\n")
			}

			fmt.Fprintf(stderr, "Minting V2 Relayer API Key at %s ...\n", rURL)
			v2Key, err := relayer.MintV2APIKey(ctx, session.HTTPClient(), rURL)
			if err != nil {
				return fmt.Errorf("mint v2 api key: %w", err)
			}
			fmt.Fprintf(stderr, "✓ V2 API key minted (apiKey=%s, address=%s)\n", v2Key.Key, v2Key.Address)

			if err := persistRelayerV2Key(abs, v2Key, force); err != nil {
				return fmt.Errorf("persist relayer key: %w", err)
			}
			fmt.Fprintf(stderr, "✓ Wrote credentials to %s (mode 0600)\n", abs)

			return w.printJSON(cmd, map[string]string{
				"wroteTo":        abs,
				"relayerApiKey":  v2Key.Key,
				"relayerAddress": v2Key.Address,
				"createdAt":      v2Key.CreatedAt,
				"permission":     "0600",
				"sessionCookies": fmt.Sprintf("%d", len(cookies)),
				"gammaURL":       gURL,
				"relayerURL":     rURL,
				"signerAddress":  signer.Address(),
				"makerAddress":   maker,
				"signatureType":  fmt.Sprintf("%d", signatureType),
				"profileID":      profileID,
			})
		},
	}

	cmd.Flags().StringVar(&envFile, "env-file", "", "target env file (default: ../go-bot/.env.relayer-v2)")
	cmd.Flags().StringVar(&gammaURL, "gamma-url", "", "Gamma API base URL (default: https://gamma-api.polymarket.com)")
	cmd.Flags().StringVar(&relayerURL, "relayer-url", "", "Relayer base URL (default: https://relayer-v2.polymarket.com)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing env file")
	cmd.Flags().BoolVar(&skipProfile, "skip-profile", false, "skip the /profiles registration step (use if profile already exists)")
	cmd.Flags().IntVar(&signatureType, "signature-type", 3, "maker derivation: 0=EOA, 1=proxy, 3=deposit wallet (default 3)")
	return cmd
}

func persistRelayerV2Key(target string, key relayer.V2APIKey, force bool) error {
	if !force {
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("env file %s exists; pass --force to overwrite", target)
		} else if !stderrors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat target: %w", err)
		}
	}
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir env file dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".env.relayer-v2.tmp.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	body := fmt.Sprintf(
		"# Polymarket V2 relayer credentials\n"+
			"# Generated by polygolem auth headless-onboard.\n"+
			"# Created: %s\n"+
			"RELAYER_API_KEY=%s\n"+
			"RELAYER_API_KEY_ADDRESS=%s\n",
		key.CreatedAt,
		key.Key,
		key.Address,
	)
	if _, err := tmp.WriteString(body); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return fmt.Errorf("chmod 0600: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("rename temp to target: %w", err)
	}
	return nil
}
