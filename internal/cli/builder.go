package cli

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
)

const (
	builderURLPath        = "https://polymarket.com/settings?tab=builder"
	defaultBuilderEnvFile = "../go-bot/.env.builder"
	defaultRelayerBaseURL = "https://relayer-v2.polymarket.com"
	defaultClobBaseURL    = "https://clob.polymarket.com"
)

// Polymarket issues UUID-shaped keys but does not stick to RFC 4122 version
// nibbles — observed values include v1-shape (existing accounts) and
// non-conforming version digits (fresh accounts). Match any 8-4-4-4-12 hex
// shape; the relayer is the authoritative validator.
var uuidShapePattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type builderCreds struct {
	Key        string
	Secret     string
	Passphrase string
}

type builderOnboardResult struct {
	WroteTo    string `json:"wrote_to"`
	Validated  bool   `json:"validated"`
	Permission string `json:"permission"`
}

func newBuilderCommand(jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "builder",
		Short: "Manage builder credentials",
		Long: `Builder helpers manage CLOB L2 credentials and legacy
builder-relayer HMAC credentials. Use 'builder auto' for CLOB L2 creds,
'auth headless-onboard' for V2 relayer keys, and 'clob
create-builder-fee-key' for order attribution.`,
	}
	cmd.AddCommand(newBuilderOnboardCommand(jsonOut))
	cmd.AddCommand(newBuilderAutoCommand(jsonOut))
	return cmd
}

// newBuilderAutoCommand creates HMAC creds with the EOA's own ClobAuth
// signature — no browser, no paste. The CLOB /auth/api-key endpoint is
// idempotent under the same EOA: re-running this command derives the
// existing key when one was already issued.
func newBuilderAutoCommand(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var envFile, clobURL string
	var force, noValidate bool

	cmd := &cobra.Command{
		Use:   "auto",
		Short: "Mint CLOB L2 creds via ClobAuth signature",
		Long: `Signs the canonical ClobAuth EIP-712 message with the EOA loaded from
POLYMARKET_PRIVATE_KEY, posts it to /auth/api-key, and persists the
returned {apiKey, secret, passphrase} to a 0600 env file.

These are CLOB L2 trading creds — they authenticate book/balance reads,
relayer GETs (/nonce, /deployed), and orders signed by the same address.
They are NOT V2 Relayer API Keys: the relayer's POST /submit (used by
deposit-wallet deploy and approve flows) requires a separate key minted
by 'polygolem auth headless-onboard' or the settings-page Create button.
A profiled EOA without that relayer key will see relayer-write 401s even
with valid CLOB L2 creds. See docs/BUILDER-AUTO.md.

The endpoint is idempotent per EOA. Use 'builder onboard' for the
manual browser-capture flow.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKey := strings.TrimSpace(os.Getenv("POLYMARKET_PRIVATE_KEY"))
			if privateKey == "" {
				return errors.New("POLYMARKET_PRIVATE_KEY is required")
			}
			target := envFile
			if target == "" {
				target = defaultBuilderEnvFile
			}
			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve env file path: %w", err)
			}

			host := clobURL
			if host == "" {
				host = defaultClobBaseURL
			}

			stderr := cmd.ErrOrStderr()
			fmt.Fprintf(stderr, "Signing ClobAuth and creating builder API key via %s...\n", host)

			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
			defer cancel()

			c := clob.NewClient(host, nil)
			key, err := c.CreateOrDeriveAPIKey(ctx, privateKey)
			if err != nil {
				return fmt.Errorf("create or derive api key: %w", err)
			}
			creds := &builderCreds{
				Key:        key.Key,
				Secret:     key.Secret,
				Passphrase: key.Passphrase,
			}
			if err := validateBuilderCredentialFormat(creds.Key, creds.Secret, creds.Passphrase); err != nil {
				return fmt.Errorf("server returned malformed creds: %w", err)
			}
			fmt.Fprintf(stderr, "✓ Received creds (key=%s)\n", creds.Key)

			validated := false
			if !noValidate {
				if err := liveCheckBuilderCredentials(ctx, creds); err != nil {
					return fmt.Errorf("relayer HMAC check failed (use --no-validate to skip): %w", err)
				}
				validated = true
				fmt.Fprintln(stderr, "✓ HMAC-signed test request to relayer succeeded")
			} else {
				fmt.Fprintln(stderr, "skipped relayer validation (--no-validate)")
			}

			if err := persistBuilderCredentials(abs, *creds, force); err != nil {
				return fmt.Errorf("persist creds: %w", err)
			}
			fmt.Fprintf(stderr, "✓ Wrote credentials to %s (mode 0600)\n", abs)

			warnIfNoDepositKeySimple(stderr, privateKey)

			return w.printJSON(cmd, builderOnboardResult{
				WroteTo:    abs,
				Validated:  validated,
				Permission: "0600",
			})
		},
	}

	cmd.Flags().StringVar(&envFile, "env-file", "", "target env file (default: ../go-bot/.env.builder)")
	cmd.Flags().StringVar(&clobURL, "clob-url", "", "CLOB base URL (default: https://clob.polymarket.com)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing builder credentials")
	cmd.Flags().BoolVar(&noValidate, "no-validate", false, "skip the relayer HMAC liveness check")
	return cmd
}

func newBuilderOnboardCommand(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var envFile string
	var force, openBrowser, noValidate bool

	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Capture builder credentials and persist locally",
		Long: `Walks the operator through creating builder credentials at
polymarket.com/settings?tab=builder, validates the result, and writes it
to a 0600-permission env file (default: ../go-bot/.env.builder relative
to the polygolem cwd).

Validation does an HMAC-signed GET against the Polymarket relayer; if the
credentials are wrong, the file is not written. Use --no-validate to
skip the network step (offline use only).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := envFile
			if target == "" {
				target = defaultBuilderEnvFile
			}
			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve env file path: %w", err)
			}

			stderr := cmd.ErrOrStderr()
			fmt.Fprintf(stderr, "Builder credentials are required for the deposit-wallet flow.\n")
			fmt.Fprintf(stderr, "Create them once at:\n  %s\n\n", builderURLPath)
			fmt.Fprintf(stderr, "Steps:\n")
			fmt.Fprintf(stderr, "  1. Connect your Polymarket wallet (use the EOA tied to POLYMARKET_PRIVATE_KEY).\n")
			fmt.Fprintf(stderr, "  2. Click \"Create Builder API Key\".\n")
			fmt.Fprintf(stderr, "  3. Copy the three values shown — the secret won't be displayed again.\n\n")

			if openBrowser {
				if err := openInBrowser(builderURLPath); err != nil {
					fmt.Fprintf(stderr, "warning: could not open browser: %v\n", err)
				}
			}

			creds, err := readBuilderCredentialsFromStdin(cmd.InOrStdin(), stderr)
			if err != nil {
				return err
			}
			if err := validateBuilderCredentialFormat(creds.Key, creds.Secret, creds.Passphrase); err != nil {
				return fmt.Errorf("credential format invalid: %w", err)
			}

			validated := false
			if !noValidate {
				if err := liveCheckBuilderCredentials(cmd.Context(), creds); err != nil {
					return fmt.Errorf("live validation failed (use --no-validate to skip): %w", err)
				}
				validated = true
				fmt.Fprintln(stderr, "✓ HMAC-signed test request to relayer succeeded")
			} else {
				fmt.Fprintln(stderr, "skipped live validation (--no-validate)")
			}

			if err := persistBuilderCredentials(abs, *creds, force); err != nil {
				return fmt.Errorf("persist creds: %w", err)
			}
			fmt.Fprintf(stderr, "✓ Wrote credentials to %s (mode 0600)\n", abs)

			return w.printJSON(cmd, builderOnboardResult{
				WroteTo:    abs,
				Validated:  validated,
				Permission: "0600",
			})
		},
	}

	cmd.Flags().StringVar(&envFile, "env-file", "", "target env file (default: ../go-bot/.env.builder)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing builder credentials")
	cmd.Flags().BoolVar(&openBrowser, "open-browser", false, "attempt to open polymarket.com/settings?tab=builder")
	cmd.Flags().BoolVar(&noValidate, "no-validate", false, "skip the relayer HMAC liveness check")
	return cmd
}

// validateBuilderCredentialFormat checks the loose shape of the three
// credential parts. It does NOT contact Polymarket. It catches typos and
// obviously malformed input before we attempt a network call.
func validateBuilderCredentialFormat(key, secret, passphrase string) error {
	key = strings.TrimSpace(key)
	secret = strings.TrimSpace(secret)
	passphrase = strings.TrimSpace(passphrase)

	if key == "" {
		return errors.New("api key is empty")
	}
	if !uuidShapePattern.MatchString(key) {
		return fmt.Errorf("api key %q is not in UUID shape", key)
	}
	if secret == "" {
		return errors.New("secret is empty")
	}
	if _, err := base64.StdEncoding.DecodeString(secret); err != nil {
		if _, err2 := base64.URLEncoding.DecodeString(secret); err2 != nil {
			return fmt.Errorf("secret is not valid base64: %w", err)
		}
	}
	if passphrase == "" {
		return errors.New("passphrase is empty")
	}
	return nil
}

// liveCheckBuilderCredentials does an HMAC-signed GET against the
// Polymarket relayer's nonce endpoint. The relayer rejects bad builder
// credentials with a 4xx; a successful round trip means the creds are
// recognised. The owner address used here is a well-known throwaway —
// the call is read-only and does not change relayer state.
func liveCheckBuilderCredentials(ctx context.Context, creds *builderCreds) error {
	bc := auth.BuilderConfig{
		Key:        creds.Key,
		Secret:     creds.Secret,
		Passphrase: creds.Passphrase,
	}
	if !bc.Valid() {
		return errors.New("builder config invalid after format validation; this is a bug")
	}

	rc, err := relayer.New(defaultRelayerBaseURL, bc, 137)
	if err != nil {
		return fmt.Errorf("init relayer client: %w", err)
	}

	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 0x000…001 is a never-deployed throwaway. The relayer should still
	// authenticate the request (or reject builder creds) before deciding
	// what to return for that owner.
	if _, err := rc.GetNonce(checkCtx, "0x0000000000000000000000000000000000000001"); err != nil {
		return err
	}
	return nil
}

// persistBuilderCredentials writes the creds atomically to target with
// 0600 mode. If target exists and force is false, returns an error.
func persistBuilderCredentials(target string, creds builderCreds, force bool) error {
	if !force {
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("env file %s exists; pass --force to overwrite", target)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat target: %w", err)
		}
	}
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir env file dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".env.builder.tmp.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	body := fmt.Sprintf(
		"# Polymarket builder credentials\n"+
			"# Generated by polygolem builder onboard.\n"+
			"# Source: %s\n"+
			"POLYMARKET_BUILDER_API_KEY=%s\n"+
			"POLYMARKET_BUILDER_SECRET=%s\n"+
			"POLYMARKET_BUILDER_PASSPHRASE=%s\n",
		builderURLPath,
		creds.Key,
		creds.Secret,
		creds.Passphrase,
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

// readBuilderCredentialsFromStdin prompts for the three values one at a
// time. Each line is read up to a newline; trailing whitespace is
// stripped. No masking — the caller chose this command knowing the
// secret will appear briefly on the terminal.
func readBuilderCredentialsFromStdin(in io.Reader, stderr io.Writer) (*builderCreds, error) {
	r := bufio.NewReader(in)
	read := func(label string) (string, error) {
		fmt.Fprintf(stderr, "%s: ", label)
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("read %s: %w", label, err)
		}
		return strings.TrimSpace(line), nil
	}
	key, err := read("POLY_BUILDER_API_KEY")
	if err != nil {
		return nil, err
	}
	secret, err := read("POLY_BUILDER_SECRET")
	if err != nil {
		return nil, err
	}
	pass, err := read("POLY_BUILDER_PASSPHRASE")
	if err != nil {
		return nil, err
	}
	return &builderCreds{Key: key, Secret: secret, Passphrase: pass}, nil
}

// openInBrowser tries to open a URL with the platform's default opener.
// Best-effort — caller treats failure as a warning.
func openInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if cmd == nil {
		return errors.New("no browser opener for this platform")
	}
	return cmd.Start()
}

// silence unused-import warnings for json in case marshaling is added
// later through a helper.
var _ = json.Marshal
