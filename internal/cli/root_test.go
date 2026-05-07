package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/clob"
)

func TestVersionCommandPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer

	root := NewRootCommand(Options{
		Version: "test-version",
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "test-version") {
		t.Fatalf("version output %q does not include test-version", got)
	}
}

func TestJSONFlagIsAcceptedAndPreflightEmitsJSON(t *testing.T) {
	var stdout bytes.Buffer

	root := NewRootCommand(Options{
		Version: "test-version",
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	root.SetArgs([]string{"--json", "preflight"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var got struct {
		OK     bool `json:"ok"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("preflight stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got.Checks) == 0 {
		t.Fatalf("preflight JSON checks is empty: %s", stdout.String())
	}
	for _, check := range got.Checks {
		if check.Name == "" {
			t.Fatalf("preflight check has empty name: %+v", got.Checks)
		}
		if check.Status == "" {
			t.Fatalf("preflight check %q has empty status", check.Name)
		}
	}
}

func TestAuthStatusReportsReadinessWithoutSecrets(t *testing.T) {
	t.Setenv("POLYMARKET_PRIVATE_KEY", "0x1111111111111111111111111111111111111111111111111111111111111111")
	t.Setenv("POLYMARKET_SIGNATURE_TYPE", "proxy")
	t.Setenv("POLYMARKET_CLOB_API_KEY", "api-secret-value")
	t.Setenv("POLYMARKET_CLOB_SECRET", "clob-secret-value")
	t.Setenv("POLYMARKET_CLOB_PASS_PHRASE", "passphrase-secret-value")

	var stdout bytes.Buffer
	root := NewRootCommand(Options{
		Version: "test-version",
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	root.SetArgs([]string{"--json", "auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var got struct {
		AccessLevel   string `json:"access_level"`
		HasSigner     bool   `json:"has_signer"`
		HasAPIKey     bool   `json:"has_api_key"`
		SignatureType string `json:"signature_type"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("auth status stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if got.AccessLevel != "L2" || !got.HasSigner || !got.HasAPIKey || got.SignatureType != "proxy" {
		t.Fatalf("unexpected auth status: %+v", got)
	}
	for _, secret := range []string{"111111111111", "api-secret-value", "clob-secret-value", "passphrase-secret-value"} {
		if strings.Contains(stdout.String(), secret) {
			t.Fatalf("auth status leaked secret %q in %s", secret, stdout.String())
		}
	}
}

func TestLiveStatusReportsGateFailures(t *testing.T) {
	t.Setenv("POLYMARKET_LIVE_PROFILE", "")
	t.Setenv("POLYMARKET_LIVE_TRADING_ENABLED", "false")

	var stdout bytes.Buffer
	root := NewRootCommand(Options{
		Version: "test-version",
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	root.SetArgs([]string{"--json", "live", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var got struct {
		Allowed  bool `json:"allowed"`
		Failures []struct {
			Code string `json:"code"`
		} `json:"failures"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("live status stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if strings.Contains(stdout.String(), `"Code"`) || !strings.Contains(stdout.String(), `"code"`) {
		t.Fatalf("live status failure fields should use lowercase JSON keys: %s", stdout.String())
	}
	if got.Allowed {
		t.Fatalf("live status allowed trading without gates: %s", stdout.String())
	}
	var codes []string
	for _, failure := range got.Failures {
		codes = append(codes, failure.Code)
	}
	for _, want := range []string{"env_gate_required", "config_gate_required", "cli_confirmation_required"} {
		if !containsString(codes, want) {
			t.Fatalf("live status missing %q in %v", want, codes)
		}
	}
}

func TestDocumentedSubcommandsAreRegistered(t *testing.T) {
	for _, args := range [][]string{
		{"discover", "search"},
		{"discover", "market"},
		{"discover", "enrich"},
		{"orderbook", "get"},
		{"orderbook", "price"},
		{"orderbook", "midpoint"},
		{"orderbook", "spread"},
		{"orderbook", "tick-size"},
		{"orderbook", "fee-rate"},
		{"clob", "book"},
		{"clob", "tick-size"},
		{"clob", "orders"},
		{"clob", "trades"},
		{"clob", "create-api-key"},
		{"clob", "balance"},
		{"clob", "update-balance"},
		{"clob", "create-order"},
		{"clob", "market-order"},
		{"paper", "buy"},
		{"paper", "sell"},
		{"paper", "positions"},
		{"paper", "reset"},
		{"auth", "status"},
		{"live", "status"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			root := NewRootCommand(Options{
				Version: "test-version",
				Stdout:  &stdout,
				Stderr:  &bytes.Buffer{},
			})
			root.SetArgs(append(args, "--help"))

			if err := root.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			wantPath := "polygolem " + strings.Join(args, " ")
			if !strings.Contains(stdout.String(), wantPath) {
				t.Fatalf("help output does not identify exact command path %q:\n%s", wantPath, stdout.String())
			}
		})
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestDocumentedSubcommandArgsAreNotHandledByParentOnly(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{
		Version: "test-version",
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	root.SetArgs([]string{"discover", "search", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "polygolem discover search") {
		t.Fatalf("discover search was not handled by its own command:\n%s", stdout.String())
	}
}

func TestBalanceAllowanceOutputNormalizesCollateralBaseUnits(t *testing.T) {
	out, err := balanceAllowanceOutput(&clob.BalanceAllowanceResponse{
		Balance:    "14000000",
		Allowances: map[string]string{"0xspender": "1000000"},
	}, "COLLATERAL")
	if err != nil {
		t.Fatalf("balanceAllowanceOutput returned error: %v", err)
	}

	if out.Balance != "14" {
		t.Fatalf("balance=%q, want human pUSD units", out.Balance)
	}
	if out.BalanceRaw != "14000000" {
		t.Fatalf("balance_raw=%q", out.BalanceRaw)
	}
	if out.BalanceDecimals != 6 {
		t.Fatalf("balance_decimals=%d", out.BalanceDecimals)
	}
	if out.Allowances["0xspender"] != "1000000" {
		t.Fatalf("allowances were not preserved raw: %#v", out.Allowances)
	}
}
