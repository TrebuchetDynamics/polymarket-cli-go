package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/gorilla/websocket"
)

type jsonEnvelopeForTest struct {
	OK      bool            `json:"ok"`
	Version string          `json:"version"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code     string `json:"code"`
		Category string `json:"category"`
		Message  string `json:"message"`
		Hint     string `json:"hint,omitempty"`
	} `json:"error,omitempty"`
	Meta struct {
		Command    string `json:"command"`
		TS         string `json:"ts"`
		DurationMS int64  `json:"duration_ms"`
	} `json:"meta"`
}

func executeRootForTest(args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &stderr})
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func parseJSONEnvelopeForTest(t *testing.T, body string) jsonEnvelopeForTest {
	t.Helper()
	var got jsonEnvelopeForTest
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("output is not a JSON envelope: %v\n%s", err, body)
	}
	if got.Version != "1" {
		t.Fatalf("version=%q, want 1\nenvelope=%s", got.Version, body)
	}
	if got.Meta.Command == "" {
		t.Fatalf("meta.command missing\nenvelope=%s", body)
	}
	if got.Meta.TS == "" {
		t.Fatalf("meta.ts missing\nenvelope=%s", body)
	}
	return got
}

func TestJSONVersionUsesSuccessEnvelope(t *testing.T) {
	stdout, stderr, err := executeRootForTest("--json", "version")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr=%q, want empty", stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false, want true\nenvelope=%s", stdout)
	}
	if got.Meta.Command != "version" {
		t.Fatalf("meta.command=%q, want version", got.Meta.Command)
	}
	var data struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data is not version payload: %v\n%s", err, got.Data)
	}
	if data.Version != "test-version" {
		t.Fatalf("data.version=%q, want test-version", data.Version)
	}
}

func TestJSONPreflightUsesSuccessEnvelope(t *testing.T) {
	stdout, stderr, err := executeRootForTest("--json", "preflight")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false, want true\nenvelope=%s", stdout)
	}
	if got.Meta.Command != "preflight" {
		t.Fatalf("meta.command=%q, want preflight", got.Meta.Command)
	}
	var data struct {
		OK     bool `json:"ok"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data is not preflight payload: %v\n%s", err, got.Data)
	}
	if len(data.Checks) == 0 {
		t.Fatalf("preflight checks empty\nenvelope=%s", stdout)
	}
}

func TestJSONGroupCommandUsesUsageErrorEnvelope(t *testing.T) {
	stdout, stderr, err := executeRootForTest("--json", "clob")
	if err == nil {
		t.Fatal("expected Execute to return usage error")
	}
	if stdout != "" {
		t.Fatalf("stdout=%q, want empty", stdout)
	}
	if got := ExitCode(err); got != 2 {
		t.Fatalf("ExitCode=%d, want 2", got)
	}
	if !ErrorAlreadyRendered(err) {
		t.Fatalf("error should be marked rendered: %v", err)
	}
	got := parseJSONEnvelopeForTest(t, stderr)
	if got.OK {
		t.Fatalf("ok=true, want false\nenvelope=%s", stderr)
	}
	if got.Meta.Command != "clob" {
		t.Fatalf("meta.command=%q, want clob", got.Meta.Command)
	}
	if got.Error == nil || got.Error.Code != "USAGE_SUBCOMMAND_UNKNOWN" || got.Error.Category != "usage" {
		t.Fatalf("unexpected error envelope: %+v\n%s", got.Error, stderr)
	}
}

func TestJSONSkeletonUsesInternalErrorEnvelope(t *testing.T) {
	stdout, stderr, err := executeRootForTest("--json", "live", "status")
	if err == nil {
		t.Fatal("expected Execute to return internal error")
	}
	if stdout != "" {
		t.Fatalf("stdout=%q, want empty", stdout)
	}
	if got := ExitCode(err); got != 9 {
		t.Fatalf("ExitCode=%d, want 9", got)
	}
	got := parseJSONEnvelopeForTest(t, stderr)
	if got.OK {
		t.Fatalf("ok=true, want false\nenvelope=%s", stderr)
	}
	if got.Meta.Command != "live status" {
		t.Fatalf("meta.command=%q, want live status", got.Meta.Command)
	}
	if got.Error == nil || got.Error.Code != "INTERNAL_UNIMPLEMENTED" || got.Error.Category != "internal" {
		t.Fatalf("unexpected error envelope: %+v\n%s", got.Error, stderr)
	}
}

func TestJSONMissingPrivateKeyUsesAuthErrorEnvelope(t *testing.T) {
	t.Setenv("POLYMARKET_PRIVATE_KEY", "")

	stdout, stderr, err := executeRootForTest("--json", "clob", "create-api-key")
	if err == nil {
		t.Fatal("expected Execute to return auth error")
	}
	if stdout != "" {
		t.Fatalf("stdout=%q, want empty", stdout)
	}
	if got := ExitCode(err); got != 3 {
		t.Fatalf("ExitCode=%d, want 3", got)
	}
	if !ErrorAlreadyRendered(err) {
		t.Fatalf("error should be marked rendered: %v", err)
	}
	if !errors.Is(err, ErrExit) {
		t.Fatalf("error should wrap ErrExit: %v", err)
	}
	got := parseJSONEnvelopeForTest(t, stderr)
	if got.OK {
		t.Fatalf("ok=true, want false\nenvelope=%s", stderr)
	}
	if got.Meta.Command != "clob create-api-key" {
		t.Fatalf("meta.command=%q, want clob create-api-key", got.Meta.Command)
	}
	if got.Error == nil || got.Error.Code != "AUTH_PRIVATE_KEY_MISSING" || got.Error.Category != "auth" {
		t.Fatalf("unexpected error envelope: %+v\n%s", got.Error, stderr)
	}
}

func TestJSONAuthStatusMissingPrivateKeyUsesAuthErrorEnvelope(t *testing.T) {
	t.Setenv("POLYMARKET_PRIVATE_KEY", "")

	stdout, stderr, err := executeRootForTest("--json", "auth", "status")
	if err == nil {
		t.Fatal("expected Execute to return auth error")
	}
	if stdout != "" {
		t.Fatalf("stdout=%q, want empty", stdout)
	}
	if got := ExitCode(err); got != 3 {
		t.Fatalf("ExitCode=%d, want 3", got)
	}
	got := parseJSONEnvelopeForTest(t, stderr)
	if got.OK {
		t.Fatalf("ok=true, want false\nenvelope=%s", stderr)
	}
	if got.Meta.Command != "auth status" {
		t.Fatalf("meta.command=%q, want auth status", got.Meta.Command)
	}
	if got.Error == nil || got.Error.Code != "AUTH_PRIVATE_KEY_MISSING" || got.Error.Category != "auth" {
		t.Fatalf("unexpected error envelope: %+v\n%s", got.Error, stderr)
	}
}

func TestAuthExportKeyRequiresConfirm(t *testing.T) {
	t.Setenv("POLYMARKET_PRIVATE_KEY", "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")

	stdout, _, err := executeRootForTest("auth", "export-key")
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	if stdout != "" {
		t.Fatalf("stdout=%q, want empty", stdout)
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error=%q, want --confirm hint", err.Error())
	}
}

func TestAuthHeadlessOnboardHasProfileRegistrationFlags(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"auth", "headless-onboard"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	for _, name := range []string{"skip-profile", "signature-type"} {
		if flag := cmd.Flags().Lookup(name); flag == nil {
			t.Fatalf("%s flag missing", name)
		}
	}
	if got := cmd.Flags().Lookup("signature-type").DefValue; got != "3" {
		t.Fatalf("signature-type default=%q, want 3", got)
	}
}

func TestAuthLoginCommandExplainsEOASignerAndDepositWallet(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"auth", "login"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if cmd == nil {
		t.Fatal("auth login command missing")
	}
	if cmd.CommandPath() != "polygolem auth login" {
		t.Fatalf("found command path %q, want polygolem auth login", cmd.CommandPath())
	}
	signatureTypeFlag := cmd.Flags().Lookup("signature-type")
	if signatureTypeFlag == nil {
		t.Fatal("signature-type flag missing")
	}
	if got := signatureTypeFlag.DefValue; got != "3" {
		t.Fatalf("signature-type default=%q, want 3", got)
	}
	help := cmd.Long
	for _, want := range []string{
		"Polymarket login signs with the EOA",
		"deposit wallet remains the trading wallet",
		"mints V2 relayer credentials",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("auth login help missing %q:\n%s", want, help)
		}
	}
}

func TestAuthCLOBProbeUsesConfiguredCredentialsForReadOnlyEndpoints(t *testing.T) {
	var requests []string
	var sawAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, "mutation attempted", http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/auth/derive-api-key":
			http.Error(w, "derive should not be called", http.StatusTeapot)
		case "/data/orders":
			sawAPIKey = r.Header.Get("POLY_API_KEY")
			_, _ = w.Write([]byte(`[{"id":"0xorder","status":"ORDER_STATUS_LIVE"}]`))
		case "/data/trades":
			_, _ = w.Write([]byte(`[{"id":"trade-1","status":"MATCHED"}]`))
		case "/balance-allowance":
			if got := r.URL.Query().Get("asset_type"); got != "COLLATERAL" {
				t.Errorf("asset_type=%q, want COLLATERAL", got)
			}
			if got := r.URL.Query().Get("signature_type"); got != "3" {
				t.Errorf("signature_type=%q, want 3", got)
			}
			_, _ = w.Write([]byte(`{"balance":"1000000","allowance":"999"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := clob.NewClient(server.URL, nil)
	result, err := runCLOBCredentialProbe(context.Background(), client, "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", auth.APIKey{
		Key:        "configured-key",
		Secret:     "c2VjcmV0",
		Passphrase: "pass",
	})
	if err != nil {
		t.Fatalf("runCLOBCredentialProbe returned error: %v", err)
	}
	if sawAPIKey != "configured-key" {
		t.Fatalf("POLY_API_KEY=%q, want configured-key", sawAPIKey)
	}
	if result.CredentialSource != "configured_clob_l2" || !result.ReadOnly || result.DeriveAPIKeyCalled {
		t.Fatalf("unexpected result metadata: %+v", result)
	}
	if result.Orders.Count != 1 || result.Trades.Count != 1 || result.BalanceAllowance.Balance != "1000000" {
		t.Fatalf("unexpected probe result: %+v", result)
	}
	for _, request := range requests {
		if strings.Contains(request, "/auth/derive-api-key") {
			t.Fatalf("probe called derive endpoint; requests=%v", requests)
		}
		if !strings.HasPrefix(request, "GET ") {
			t.Fatalf("probe made non-read request; requests=%v", requests)
		}
	}
}

func TestAuthCLOBProbeCommandIsRegistered(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"auth", "clob-probe"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if cmd == nil {
		t.Fatal("auth clob-probe command missing")
	}
}

func TestJSONAuthExportKeyConfirmedOutputsWalletImportData(t *testing.T) {
	const privateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	t.Setenv("POLYMARKET_PRIVATE_KEY", privateKey)

	stdout, stderr, err := executeRootForTest("--json", "auth", "export-key", "--confirm")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stderr, "SECURITY WARNING") {
		t.Fatalf("stderr=%q, want security warning", stderr)
	}
	got := parseJSONEnvelopeForTest(t, stdout)
	if !got.OK {
		t.Fatalf("ok=false, want true\nenvelope=%s", stdout)
	}
	if got.Meta.Command != "auth export-key" {
		t.Fatalf("meta.command=%q, want auth export-key", got.Meta.Command)
	}
	var data struct {
		EOAAddress    string `json:"eoaAddress"`
		DepositWallet string `json:"depositWallet"`
		PrivateKey    string `json:"privateKey"`
	}
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data is not export-key payload: %v\n%s", err, got.Data)
	}
	if data.PrivateKey != privateKey {
		t.Fatalf("privateKey=%q, want configured key", data.PrivateKey)
	}
	if data.EOAAddress == "" || data.DepositWallet == "" {
		t.Fatalf("derived addresses missing: %+v", data)
	}
}

func TestJSONMissingPositionalArgUsesUsageErrorEnvelope(t *testing.T) {
	stdout, stderr, err := executeRootForTest("--json", "clob", "book")
	if err == nil {
		t.Fatal("expected Execute to return usage error")
	}
	if stdout != "" {
		t.Fatalf("stdout=%q, want empty", stdout)
	}
	if got := ExitCode(err); got != 2 {
		t.Fatalf("ExitCode=%d, want 2", got)
	}
	got := parseJSONEnvelopeForTest(t, stderr)
	if got.OK {
		t.Fatalf("ok=true, want false\nenvelope=%s", stderr)
	}
	if got.Meta.Command != "clob book" {
		t.Fatalf("meta.command=%q, want clob book", got.Meta.Command)
	}
	if got.Error == nil || got.Error.Code != "USAGE_ARG_INVALID" || got.Error.Category != "usage" {
		t.Fatalf("unexpected error envelope: %+v\n%s", got.Error, stderr)
	}
}

func TestVersionCommandPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
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
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"--json", "preflight"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var got jsonEnvelopeForTest
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("preflight stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	var data struct {
		OK     bool `json:"ok"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("preflight data is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(data.Checks) == 0 {
		t.Fatalf("preflight JSON checks is empty: %s", stdout.String())
	}
}

func TestDocumentedSubcommandsAreRegistered(t *testing.T) {
	for _, args := range [][]string{
		{"discover", "search"},
		{"discover", "markets"},
		{"discover", "market"},
		{"discover", "enrich"},
		{"discover", "tags"},
		{"discover", "series"},
		{"discover", "comments"},
		{"orderbook", "get"},
		{"orderbook", "price"},
		{"orderbook", "midpoint"},
		{"orderbook", "spread"},
		{"orderbook", "tick-size"},
		{"orderbook", "fee-rate"},
		{"orderbook", "last-trade"},
		{"clob", "book"},
		{"clob", "tick-size"},
		{"clob", "create-api-key"},
		{"clob", "create-api-key-for-address"},
		{"clob", "create-builder-fee-key"},
		{"clob", "list-builder-fee-keys"},
		{"clob", "revoke-builder-fee-key"},
		{"clob", "balance"},
		{"clob", "update-balance"},
		{"clob", "orders"},
		{"clob", "order"},
		{"clob", "trades"},
		{"clob", "cancel"},
		{"clob", "cancel-orders"},
		{"clob", "cancel-all"},
		{"clob", "cancel-market"},
		{"clob", "create-order"},
		{"clob", "batch-orders"},
		{"clob", "market-order"},
		{"clob", "heartbeat"},
		{"clob", "price-history"},
		{"clob", "market"},
		{"clob", "market-by-token"},
		{"clob", "markets"},
		{"data", "positions"},
		{"data", "closed-positions"},
		{"data", "trades"},
		{"data", "activity"},
		{"data", "holders"},
		{"data", "value"},
		{"data", "markets-traded"},
		{"data", "open-interest"},
		{"data", "leaderboard"},
		{"data", "live-volume"},
		{"stream", "market"},
		{"events", "list"},
		{"bridge", "assets"},
		{"bridge", "deposit"},
		{"health"},
		{"paper", "buy"},
		{"paper", "sell"},
		{"paper", "positions"},
		{"paper", "reset"},
		{"auth", "status"},
		{"auth", "export-key"},
		{"auth", "login"},
		{"auth", "headless-onboard"},
		{"live", "status"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
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

func TestCLOBCreateAPIKeyForAddressHasOwnerFlag(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"clob", "create-api-key-for-address"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	flag := cmd.Flags().Lookup("owner")
	if flag == nil {
		t.Fatal("owner flag missing")
	}
	if flag.DefValue != "" {
		t.Fatalf("default owner=%q, want empty", flag.DefValue)
	}
}

func TestCLOBOrderCommandsHaveBuilderCodeFlag(t *testing.T) {
	for _, args := range [][]string{
		{"clob", "create-order"},
		{"clob", "batch-orders"},
		{"clob", "market-order"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
			cmd, _, err := root.Find(args)
			if err != nil {
				t.Fatalf("Find returned error: %v", err)
			}
			flag := cmd.Flags().Lookup("builder-code")
			if flag == nil {
				t.Fatal("builder-code flag missing")
			}
			if flag.DefValue != "" {
				t.Fatalf("default builder-code=%q, want empty", flag.DefValue)
			}
		})
	}
}

func TestCLOBBatchOrdersHasOrdersFileFlag(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"clob", "batch-orders"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	flag := cmd.Flags().Lookup("orders-file")
	if flag == nil {
		t.Fatal("orders-file flag missing")
	}
	if flag.DefValue != "" {
		t.Fatalf("default orders-file=%q, want empty", flag.DefValue)
	}
}

func TestCLOBHeartbeatHasIDFlag(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"clob", "heartbeat"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	flag := cmd.Flags().Lookup("id")
	if flag == nil {
		t.Fatal("id flag missing")
	}
	if flag.DefValue != "" {
		t.Fatalf("default id=%q, want empty", flag.DefValue)
	}
}

func TestParseBatchOrderParamsAcceptsTokenAliasAndPostOnly(t *testing.T) {
	body := strings.NewReader(`[
		{"token":"12345","side":"buy","price":"0.5","size":"10","orderType":"GTC","postOnly":true},
		{"tokenID":"12346","side":"sell","price":"0.6","size":"5","orderType":"GTD","expiration":"1778125000"}
	]`)
	got, err := parseBatchOrderParams(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if got[0].TokenID != "12345" || !got[0].PostOnly {
		t.Fatalf("first order=%+v", got[0])
	}
	if got[1].TokenID != "12346" || got[1].Expiration != "1778125000" {
		t.Fatalf("second order=%+v", got[1])
	}
}

func TestBuilderCodeFromFlagOrEnv(t *testing.T) {
	envBuilderCode := "0x1111111111111111111111111111111111111111111111111111111111111111"
	flagBuilderCode := "0x2222222222222222222222222222222222222222222222222222222222222222"
	t.Setenv("POLYMARKET_BUILDER_CODE", envBuilderCode)

	if got := builderCodeFromFlagOrEnv(""); got != envBuilderCode {
		t.Fatalf("env builder code=%q, want %q", got, envBuilderCode)
	}
	if got := builderCodeFromFlagOrEnv(flagBuilderCode); got != flagBuilderCode {
		t.Fatalf("flag builder code=%q, want %q", got, flagBuilderCode)
	}
}

func TestCLOBL2CredentialsFromEnvUsesCanonicalNames(t *testing.T) {
	t.Setenv("POLYMARKET_CLOB_API_KEY", "clob-key")
	t.Setenv("POLYMARKET_CLOB_SECRET", "clob-secret")
	t.Setenv("POLYMARKET_CLOB_PASSPHRASE", "clob-pass")

	key, ok := clobL2CredentialsFromEnv()
	if !ok {
		t.Fatal("expected configured CLOB L2 credentials")
	}
	if key.Key != "clob-key" || key.Secret != "clob-secret" || key.Passphrase != "clob-pass" {
		t.Fatalf("credentials=%+v", key)
	}
}

func TestCLOBL2CredentialsFromEnvUsesShortAliases(t *testing.T) {
	t.Setenv("CLOB_API_KEY", "clob-key")
	t.Setenv("CLOB_SECRET", "clob-secret")
	t.Setenv("CLOB_PASSPHRASE", "clob-pass")

	key, ok := clobL2CredentialsFromEnv()
	if !ok {
		t.Fatal("expected configured CLOB L2 credentials")
	}
	if key.Key != "clob-key" || key.Secret != "clob-secret" || key.Passphrase != "clob-pass" {
		t.Fatalf("credentials=%+v", key)
	}
}

func TestCLOBL2CredentialsFromEnvTreatsPartialConfigAsConfigured(t *testing.T) {
	t.Setenv("POLYMARKET_CLOB_API_KEY", "clob-key")

	key, ok := clobL2CredentialsFromEnv()
	if !ok {
		t.Fatal("expected partial CLOB L2 config to be surfaced")
	}
	if key.Key != "clob-key" || key.Secret != "" || key.Passphrase != "" {
		t.Fatalf("credentials=%+v", key)
	}
}

func TestPreflightRejectsInvalidBuilderCodeEnv(t *testing.T) {
	t.Setenv("POLYMARKET_BUILDER_CODE", "0x1234")

	result := runLocalPreflight(context.Background(), "test-version")
	if result.OK {
		t.Fatal("preflight should fail when POLYMARKET_BUILDER_CODE is malformed")
	}
	for _, check := range result.Checks {
		if check.Name == "clob_builder_code" {
			if check.Status != "fail" || !strings.Contains(check.Message, "builder") {
				t.Fatalf("unexpected builder-code check: %+v", check)
			}
			return
		}
	}
	t.Fatalf("clob_builder_code check missing: %+v", result.Checks)
}

// TestCLOBSignatureTypeFlagRemoved verifies that the --signature-type flag
// has been removed from every CLOB command. Polymarket V2 (post 2026-04-28
// cutover) accepts only sigtype 3 (POLY_1271 / deposit wallet); exposing a
// flag would only let callers pick a value that production rejects.
func TestCLOBSignatureTypeFlagRemoved(t *testing.T) {
	for _, args := range [][]string{
		{"clob", "balance"},
		{"clob", "update-balance"},
		{"clob", "create-order"},
		{"clob", "market-order"},
		{"deposit-wallet", "derive"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
			cmd, _, err := root.Find(args)
			if err != nil {
				t.Fatalf("Find returned error: %v", err)
			}
			if flag := cmd.Flags().Lookup("signature-type"); flag != nil {
				t.Fatalf("signature-type flag still present (default=%q); should be removed", flag.DefValue)
			}
		})
	}
}

func TestCLOBCreateOrderExpirationDefaultsToZero(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"clob", "create-order"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	flag := cmd.Flags().Lookup("expiration")
	if flag == nil {
		t.Fatal("expiration flag missing")
	}
	if flag.DefValue != "0" {
		t.Fatalf("default expiration=%q, want 0", flag.DefValue)
	}
}

func TestCLOBCreateOrderHasPostOnlyFlag(t *testing.T) {
	root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd, _, err := root.Find([]string{"clob", "create-order"})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	flag := cmd.Flags().Lookup("post-only")
	if flag == nil {
		t.Fatal("post-only flag missing")
	}
	if flag.DefValue != "false" {
		t.Fatalf("default post-only=%q, want false", flag.DefValue)
	}
}

func TestStreamMarketReadsFromLocalWebSocket(t *testing.T) {
	upgrader := websocket.Upgrader{}
	subscriptions := make(chan []string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		var sub struct {
			Type     string   `json:"type"`
			AssetIDs []string `json:"assets_ids"`
		}
		if err := conn.ReadJSON(&sub); err != nil {
			t.Errorf("read subscription: %v", err)
			return
		}
		subscriptions <- sub.AssetIDs
		if err := conn.WriteJSON(map[string]any{
			"event_type": "book",
			"asset_id":   "token-1",
			"market":     "market-1",
			"timestamp":  "1",
			"bids":       []map[string]string{{"price": "0.50", "size": "10"}},
			"asks":       []map[string]string{{"price": "0.51", "size": "12"}},
		}); err != nil {
			t.Errorf("write stream message: %v", err)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{
		"--json",
		"stream", "market",
		"--url", "ws" + strings.TrimPrefix(server.URL, "http"),
		"--asset-ids", "token-1",
		"--max-messages", "1",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	gotSubscription := <-subscriptions
	if len(gotSubscription) != 1 || gotSubscription[0] != "token-1" {
		t.Fatalf("subscription=%v, want [token-1]", gotSubscription)
	}
	var envelope jsonEnvelopeForTest
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stream stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	var got struct {
		EventType string `json:"event_type"`
		AssetID   string `json:"asset_id"`
		Market    string `json:"market"`
	}
	if err := json.Unmarshal(envelope.Data, &got); err != nil {
		t.Fatalf("stream data is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if got.EventType != "book" || got.AssetID != "token-1" || got.Market != "market-1" {
		t.Fatalf("unexpected stream output: %+v", got)
	}
}

func TestMarketDataLiveEmitsEnrichedSnapshots(t *testing.T) {
	upgrader := websocket.Upgrader{}
	subscriptions := make(chan []string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		var sub struct {
			Type     string   `json:"type"`
			AssetIDs []string `json:"assets_ids"`
		}
		if err := conn.ReadJSON(&sub); err != nil {
			t.Errorf("read subscription: %v", err)
			return
		}
		subscriptions <- sub.AssetIDs
		if err := conn.WriteJSON(map[string]any{
			"event_type": "book",
			"asset_id":   "token-1",
			"market":     "market-1",
			"timestamp":  "1",
			"bids": []map[string]string{
				{"price": "0.49", "size": "10"},
				{"price": "0.51", "size": "3"},
			},
			"asks": []map[string]string{
				{"price": "0.55", "size": "2"},
				{"price": "0.53", "size": "4"},
			},
		}); err != nil {
			t.Errorf("write stream message: %v", err)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{
		"--json",
		"marketdata", "live",
		"--url", "ws" + strings.TrimPrefix(server.URL, "http"),
		"--asset-ids", "token-1",
		"--max-messages", "1",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	gotSubscription := <-subscriptions
	if len(gotSubscription) != 1 || gotSubscription[0] != "token-1" {
		t.Fatalf("subscription=%v, want [token-1]", gotSubscription)
	}
	var envelope jsonEnvelopeForTest
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("marketdata stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	var got struct {
		EventType string `json:"event_type"`
		AssetID   string `json:"asset_id"`
		BestBid   string `json:"best_bid"`
		BestAsk   string `json:"best_ask"`
		Midpoint  string `json:"midpoint"`
	}
	if err := json.Unmarshal(envelope.Data, &got); err != nil {
		t.Fatalf("marketdata data is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if got.EventType != "book" || got.AssetID != "token-1" || got.BestBid != "0.51" || got.BestAsk != "0.53" || got.Midpoint != "0.52" {
		t.Fatalf("unexpected marketdata output: %+v", got)
	}
}

func TestDocumentedSubcommandArgsAreNotHandledByParentOnly(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"discover", "search", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "polygolem discover search") {
		t.Fatalf("discover search was not handled by its own command:\n%s", stdout.String())
	}
}

func TestNormalizeCollateralBalanceResponseScalesBaseUnits(t *testing.T) {
	raw := map[string]interface{}{
		"balance": "14000000",
		"allowances": map[string]string{
			"0xspender": "1000000",
		},
	}

	got := normalizeCollateralBalanceResponse(raw)

	if got["balance"] != "14.000000" {
		t.Fatalf("balance=%v", got["balance"])
	}
	if !reflect.DeepEqual(got["allowances"], raw["allowances"]) {
		t.Fatalf("allowances changed: %#v", got["allowances"])
	}
}
