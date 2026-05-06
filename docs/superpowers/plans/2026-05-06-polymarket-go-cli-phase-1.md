# Polymarket Go CLI Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Phase 1 of `polymarket-cli-go`: a safe-by-default Go CLI with read-only market data, local-only paper trading, live-mode gates, structured JSON output, and docs that match actual behavior.

**Architecture:** Replace the Rust tree with a Go module shaped around typed internal packages and a thin Cobra CLI. Protocol clients, modes, config, output, preflight, and paper trading remain testable without executing the binary. Existing untracked `cmd/` and `pkg/` prototype files are treated as untrusted because they contain live order paths and do not follow the approved TDD safety model.

**Tech Stack:** Go, Cobra, Viper, standard `net/http`, `httptest`, `encoding/json`, table-driven tests, golden fixtures, GitHub Actions.

---

## Non-Negotiable TDD Rule

No production Go code may be written before a failing test proves the required behavior. For every production Go change:

1. Add the focused failing test.
2. Run the focused test and confirm the expected failure.
3. Add the minimum production code.
4. Run the focused test and confirm it passes.
5. Run the relevant package tests.
6. Commit only the files for that task.

Do not keep or adapt the current untracked prototype as production code unless each behavior is reintroduced through this red-green loop.

## Dirty Worktree Guard

Current observed dirty state includes untracked Go prototype files and tracked Rust deletions. Before implementation:

```bash
git status --short --branch
```

Rules:

- Do not commit unrelated untracked prototype files as-is.
- Do not commit tracked Rust deletions except in the dedicated repository-replacement task.
- Stage explicit file paths only.
- If a file contains live trading behavior, it must be deleted or replaced before any Phase 1 command is considered safe.

## File Structure

Create or replace these files during Phase 1:

- `go.mod`: module metadata for `github.com/TrebuchetDynamics/polymarket-cli-go`.
- `cmd/polymarket-cli-go/main.go`: binary entry point only.
- `internal/cli/root.go`: Cobra root command and dependency wiring.
- `internal/cli/root_test.go`: command construction tests.
- `internal/output/output.go`: JSON/table rendering and structured errors.
- `internal/output/output_test.go`: output contract tests.
- `internal/modes/modes.go`: read-only, paper, and live mode parsing and guards.
- `internal/modes/modes_test.go`: safety-gate tests.
- `internal/config/config.go`: Viper-backed explicit config loader.
- `internal/config/config_test.go`: config default, env, file, and redaction tests.
- `internal/preflight/preflight.go`: preflight checks and gate aggregation.
- `internal/preflight/preflight_test.go`: preflight pass/fail tests with fake probes.
- `internal/gamma/client.go`: typed read-only Gamma client.
- `internal/gamma/client_test.go`: mock HTTP tests for markets.
- `internal/clob/client.go`: typed read-only CLOB client.
- `internal/clob/client_test.go`: mock HTTP tests for order books and prices.
- `internal/markets/service.go`: market service over protocol clients.
- `internal/markets/service_test.go`: service behavior tests.
- `internal/paper/state.go`: local paper state model and JSON store.
- `internal/paper/state_test.go`: paper buy/sell/positions/reset tests.
- `internal/auth/status.go`: auth status model without credential exposure.
- `internal/auth/status_test.go`: auth status tests.
- `internal/wallet/status.go`: wallet readiness model without signing.
- `internal/wallet/status_test.go`: wallet readiness tests.
- `docs/REFERENCE-RUST-CLI.md`: Rust behavior audit.
- `docs/ARCHITECTURE.md`: package and dependency flow.
- `docs/COMMANDS.md`: implemented commands and JSON examples.
- `docs/SAFETY.md`: safety model and failure modes.
- `.github/workflows/ci.yml`: Go CI.

Do not create `pkg/polymarket` in Phase 1. The public SDK boundary remains deferred until internal packages stabilize.

## Task 1: Rust Reference Audit

**Files:**
- Create: `docs/REFERENCE-RUST-CLI.md`

- [ ] **Step 1: Gather reference evidence without relying on working-tree Rust files**

Run:

```bash
git show 4b5a749:Cargo.toml > /tmp/polymarket-cli-Cargo.toml
git show 4b5a749:src/main.rs > /tmp/polymarket-cli-main.rs
git show 4b5a749:src/config.rs > /tmp/polymarket-cli-config.rs
git show 4b5a749:src/auth.rs > /tmp/polymarket-cli-auth.rs
git show 4b5a749:src/commands/clob.rs > /tmp/polymarket-cli-clob.rs
```

Expected: all commands exit 0.

- [ ] **Step 2: Write the audit doc**

Create `docs/REFERENCE-RUST-CLI.md` with these sections:

```markdown
# Rust CLI Reference Audit

## Source

- Repository: https://github.com/Polymarket/polymarket-cli
- Audited commit: 4b5a749
- License: MIT
- Audit date: 2026-05-06

## Behavioral Use

The Rust CLI is a behavioral and protocol reference. The Go CLI must not copy
Rust source or blindly translate implementation details.

## Command Structure

- `setup`
- `shell`
- `markets`
- `events`
- `tags`
- `series`
- `comments`
- `profiles`
- `sports`
- `approve`
- `clob`
- `ctf`
- `data`
- `bridge`
- `wallet`
- `status`
- `upgrade`

## Auth Model

The Rust CLI resolves private keys from CLI flag, environment variable, then
config file. It supports `proxy`, `eoa`, and `gnosis-safe` signature types.

## Config Format

Rust stores config at `~/.config/polymarket/config.json` with private key,
chain ID, and signature type. The Go Phase 1 CLI avoids credential requirements
for read-only and paper commands.

## Live Mutation Paths

The Rust CLI includes CLOB order creation, market orders, order cancellation,
balance, rewards, API-key management, CTF operations, approvals, and bridge
flows. The Go Phase 1 CLI intentionally blocks live mutations.

## Protocol Assumptions

- Polygon chain ID 137 is the primary chain target.
- CLOB read APIs can be used without wallet credentials.
- Gamma market APIs can be used without wallet credentials.

## Maintenance Concerns

- Rust command handlers include many protocol surfaces in a single binary.
- Authenticated trading paths are convenient but dangerous for a Phase 1 Go
  rewrite.
- JSON output parity needs golden tests to prevent drift.

## License Constraints

The upstream Rust project is MIT licensed. Behavioral parity is allowed. Source
copying is avoided so the Go implementation remains independently designed.
```

- [ ] **Step 3: Verify audit doc exists**

Run:

```bash
test -s docs/REFERENCE-RUST-CLI.md
```

Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
git add docs/REFERENCE-RUST-CLI.md
git commit -m "docs: audit rust polymarket cli reference"
```

## Task 2: Replace Prototype With Go Module Baseline

**Files:**
- Replace: `go.mod`
- Create: `cmd/polymarket-cli-go/main.go`
- Create: `internal/cli/root_test.go`
- Create: `internal/cli/root.go`

- [ ] **Step 1: Remove unsafe prototype files from the working tree**

The current prototype contains live order paths. Remove only untracked prototype
paths during this task:

```bash
command -v trash
trash cmd pkg go.mod
```

Expected: `command -v trash` prints a path, and `trash` exits 0.

If `trash` is unavailable, run:

```bash
git status --short
```

Expected if unavailable: report `trash` missing and stop for a manual decision
before deleting untracked paths.

- [ ] **Step 2: Create dependency metadata**

Create `go.mod`:

```go
module github.com/TrebuchetDynamics/polymarket-cli-go

go 1.23

require (
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
)
```

Run:

```bash
go mod tidy
```

Expected: `go.sum` is created or updated, and command exits 0.

- [ ] **Step 3: Write the failing root command test**

Create `internal/cli/root_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
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
```

- [ ] **Step 4: Run test to verify it fails**

Run:

```bash
go test ./internal/cli -run TestVersionCommandPrintsVersion -count=1
```

Expected: FAIL because `NewRootCommand` and `Options` are undefined.

- [ ] **Step 5: Write minimal CLI implementation**

Create `internal/cli/root.go`:

```go
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
	Version string
	Stdout  io.Writer
	Stderr  io.Writer
}

func NewRootCommand(opts Options) *cobra.Command {
	if opts.Version == "" {
		opts.Version = "dev"
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	root := &cobra.Command{
		Use:           "polymarket",
		Short:         "Safe Polymarket CLI for research and automation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "polymarket %s\n", opts.Version)
			return err
		},
	})

	return root
}
```

Create `cmd/polymarket-cli-go/main.go`:

```go
package main

import (
	"os"

	"github.com/TrebuchetDynamics/polymarket-cli-go/internal/cli"
)

var version = "dev"

func main() {
	root := cli.NewRootCommand(cli.Options{Version: version})
	if err := root.Execute(); err != nil {
		_, _ = root.ErrOrStderr().Write([]byte(err.Error() + "\n"))
		os.Exit(1)
	}
}
```

- [ ] **Step 6: Run focused and package tests**

Run:

```bash
go test ./internal/cli -run TestVersionCommandPrintsVersion -count=1
go test ./...
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum cmd/polymarket-cli-go/main.go internal/cli/root.go internal/cli/root_test.go
git commit -m "feat: add go cli baseline"
```

## Task 3: Structured Output Contract

**Files:**
- Create: `internal/output/output_test.go`
- Create: `internal/output/output.go`

- [ ] **Step 1: Write failing JSON error test**

Create `internal/output/output_test.go`:

```go
package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteErrorJSONUsesStableEnvelope(t *testing.T) {
	var buf bytes.Buffer

	err := WriteError(&buf, FormatJSON, Error{
		Code:    "live_gate_failed",
		Message: "live trading requires --confirm-live",
		Details: map[string]string{"gate": "cli_confirmation"},
	})
	if err != nil {
		t.Fatalf("WriteError returned error: %v", err)
	}

	var got map[string]map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("error output is not JSON: %v\n%s", err, buf.String())
	}
	if got["error"]["code"] != "live_gate_failed" {
		t.Fatalf("unexpected code: %#v", got)
	}
	if got["error"]["message"] != "live trading requires --confirm-live" {
		t.Fatalf("unexpected message: %#v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/output -run TestWriteErrorJSONUsesStableEnvelope -count=1
```

Expected: FAIL because output package functions are undefined.

- [ ] **Step 3: Write minimal output implementation**

Create `internal/output/output.go`:

```go
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error Error `json:"error"`
}

func WriteJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func WriteError(w io.Writer, format Format, err Error) error {
	if format == FormatJSON {
		return WriteJSON(w, errorEnvelope{Error: err})
	}
	_, writeErr := fmt.Fprintf(w, "Error: %s\n", err.Message)
	return writeErr
}
```

- [ ] **Step 4: Run output tests**

Run:

```bash
go test ./internal/output -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/output/output.go internal/output/output_test.go
git commit -m "feat: add structured output contract"
```

## Task 4: Execution Modes And Live Gates

**Files:**
- Create: `internal/modes/modes_test.go`
- Create: `internal/modes/modes.go`

- [ ] **Step 1: Write failing live-gate tests**

Create `internal/modes/modes_test.go`:

```go
package modes

import "testing"

func TestDefaultModeIsReadOnly(t *testing.T) {
	mode, err := Parse("")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if mode != ReadOnly {
		t.Fatalf("mode = %q, want %q", mode, ReadOnly)
	}
}

func TestLiveModeRequiresAllGates(t *testing.T) {
	result := ValidateLiveGates(LiveGateInput{
		EnvEnabled:    true,
		ConfigEnabled: true,
		ConfirmLive:   false,
		PreflightOK:   true,
	})
	if result.Allowed {
		t.Fatal("live mode allowed without CLI confirmation")
	}
	if result.Failures[0].Code != "cli_confirmation_required" {
		t.Fatalf("unexpected failures: %#v", result.Failures)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/modes -count=1
```

Expected: FAIL because `Parse`, `ReadOnly`, `ValidateLiveGates`, and related types are undefined.

- [ ] **Step 3: Write minimal modes implementation**

Create `internal/modes/modes.go`:

```go
package modes

import "fmt"

type Mode string

const (
	ReadOnly Mode = "read-only"
	Paper    Mode = "paper"
	Live     Mode = "live"
)

type Failure struct {
	Code    string
	Message string
}

type LiveGateInput struct {
	EnvEnabled    bool
	ConfigEnabled bool
	ConfirmLive   bool
	PreflightOK   bool
}

type LiveGateResult struct {
	Allowed  bool
	Failures []Failure
}

func Parse(value string) (Mode, error) {
	switch value {
	case "", string(ReadOnly):
		return ReadOnly, nil
	case string(Paper):
		return Paper, nil
	case string(Live):
		return Live, nil
	default:
		return "", fmt.Errorf("unknown mode %q", value)
	}
}

func ValidateLiveGates(input LiveGateInput) LiveGateResult {
	var failures []Failure
	if !input.EnvEnabled {
		failures = append(failures, Failure{Code: "env_gate_required", Message: "POLYMARKET_LIVE_PROFILE must be on"})
	}
	if !input.ConfigEnabled {
		failures = append(failures, Failure{Code: "config_gate_required", Message: "live_trading_enabled must be true"})
	}
	if !input.ConfirmLive {
		failures = append(failures, Failure{Code: "cli_confirmation_required", Message: "--confirm-live is required"})
	}
	if !input.PreflightOK {
		failures = append(failures, Failure{Code: "preflight_required", Message: "preflight must pass"})
	}
	return LiveGateResult{Allowed: len(failures) == 0, Failures: failures}
}
```

- [ ] **Step 4: Run modes tests**

Run:

```bash
go test ./internal/modes -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modes/modes.go internal/modes/modes_test.go
git commit -m "feat: add execution mode gates"
```

## Task 5: Config Loading With Viper Instances

**Files:**
- Create: `internal/config/config_test.go`
- Create: `internal/config/config.go`

- [ ] **Step 1: Write failing config tests**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaultsToReadOnlyAndSafeURLs(t *testing.T) {
	cfg, err := Load(Options{})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Mode != "read-only" {
		t.Fatalf("Mode = %q, want read-only", cfg.Mode)
	}
	if cfg.LiveTradingEnabled {
		t.Fatal("live trading must default to disabled")
	}
	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("RequestTimeout = %s, want 10s", cfg.RequestTimeout)
	}
}

func TestLoadReadsExplicitConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("mode: paper\npaper_state_path: /tmp/paper.json\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(Options{ConfigPath: path})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Mode != "paper" {
		t.Fatalf("Mode = %q, want paper", cfg.Mode)
	}
	if cfg.PaperStatePath != "/tmp/paper.json" {
		t.Fatalf("PaperStatePath = %q", cfg.PaperStatePath)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/config -count=1
```

Expected: FAIL because config package is undefined.

- [ ] **Step 3: Write minimal config implementation**

Create `internal/config/config.go`:

```go
package config

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Options struct {
	ConfigPath string
	EnvPrefix  string
}

type Config struct {
	Mode               string
	GammaBaseURL       string
	CLOBBaseURL        string
	RequestTimeout     time.Duration
	LiveTradingEnabled bool
	PaperStatePath     string
}

func Load(opts Options) (Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if opts.EnvPrefix == "" {
		opts.EnvPrefix = "POLYMARKET"
	}
	v.SetEnvPrefix(opts.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("mode", "read-only")
	v.SetDefault("gamma_base_url", "https://gamma-api.polymarket.com")
	v.SetDefault("clob_base_url", "https://clob.polymarket.com")
	v.SetDefault("request_timeout", "10s")
	v.SetDefault("live_trading_enabled", false)
	v.SetDefault("paper_state_path", "")
	if opts.ConfigPath != "" {
		v.SetConfigFile(opts.ConfigPath)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, err
		}
	}
	timeout, err := time.ParseDuration(v.GetString("request_timeout"))
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Mode:               v.GetString("mode"),
		GammaBaseURL:       v.GetString("gamma_base_url"),
		CLOBBaseURL:        v.GetString("clob_base_url"),
		RequestTimeout:     timeout,
		LiveTradingEnabled: v.GetBool("live_trading_enabled"),
		PaperStatePath:     v.GetString("paper_state_path"),
	}
	if cfg.Mode == "" {
		return Config{}, errors.New("mode is required")
	}
	return cfg, nil
}
```

- [ ] **Step 4: Run config tests**

Run:

```bash
go test ./internal/config -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add safe config loading"
```

## Task 6: Preflight Aggregation

**Files:**
- Create: `internal/preflight/preflight_test.go`
- Create: `internal/preflight/preflight.go`

- [ ] **Step 1: Write failing preflight tests**

Create `internal/preflight/preflight_test.go`:

```go
package preflight

import (
	"context"
	"errors"
	"testing"
)

func TestRunReportsProbeFailures(t *testing.T) {
	checks := []Check{
		{Name: "gamma", Probe: func(context.Context) error { return nil }},
		{Name: "clob", Probe: func(context.Context) error { return errors.New("503") }},
	}
	result := Run(context.Background(), checks)
	if result.OK {
		t.Fatal("preflight should fail when a probe fails")
	}
	if result.Checks[1].Status != "fail" {
		t.Fatalf("second status = %q", result.Checks[1].Status)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/preflight -count=1
```

Expected: FAIL because preflight types are undefined.

- [ ] **Step 3: Write minimal preflight implementation**

Create `internal/preflight/preflight.go`:

```go
package preflight

import "context"

type Probe func(context.Context) error

type Check struct {
	Name  string
	Probe Probe
}

type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Result struct {
	OK     bool          `json:"ok"`
	Checks []CheckResult `json:"checks"`
}

func Run(ctx context.Context, checks []Check) Result {
	result := Result{OK: true}
	for _, check := range checks {
		err := check.Probe(ctx)
		checkResult := CheckResult{Name: check.Name, Status: "pass"}
		if err != nil {
			result.OK = false
			checkResult.Status = "fail"
			checkResult.Message = err.Error()
		}
		result.Checks = append(result.Checks, checkResult)
	}
	return result
}
```

- [ ] **Step 4: Run preflight tests**

Run:

```bash
go test ./internal/preflight -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/preflight/preflight.go internal/preflight/preflight_test.go
git commit -m "feat: add preflight checks"
```

## Task 7: Paper State Is Local-Only

**Files:**
- Create: `internal/paper/state_test.go`
- Create: `internal/paper/state.go`

- [ ] **Step 1: Write failing paper state tests**

Create `internal/paper/state_test.go`:

```go
package paper

import "testing"

func TestBuyUpdatesLocalPositionWithoutExternalExecution(t *testing.T) {
	state := NewState("USD", 100)
	fill, err := state.Buy(Order{
		MarketID: "market-1",
		TokenID:  "yes-token",
		Price:    0.25,
		Size:     10,
	})
	if err != nil {
		t.Fatalf("Buy returned error: %v", err)
	}
	if fill.Live {
		t.Fatal("paper fill must not be live")
	}
	if state.Cash != 97.5 {
		t.Fatalf("Cash = %v, want 97.5", state.Cash)
	}
	if state.Positions["yes-token"].Size != 10 {
		t.Fatalf("position size = %v", state.Positions["yes-token"].Size)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/paper -count=1
```

Expected: FAIL because paper package types are undefined.

- [ ] **Step 3: Write minimal paper state implementation**

Create `internal/paper/state.go`:

```go
package paper

import "fmt"

type Order struct {
	MarketID string  `json:"market_id"`
	TokenID  string  `json:"token_id"`
	Price    float64 `json:"price"`
	Size     float64 `json:"size"`
}

type Fill struct {
	MarketID string  `json:"market_id"`
	TokenID  string  `json:"token_id"`
	Price    float64 `json:"price"`
	Size     float64 `json:"size"`
	Live     bool    `json:"live"`
}

type Position struct {
	TokenID string  `json:"token_id"`
	Size    float64 `json:"size"`
	Cost    float64 `json:"cost"`
}

type State struct {
	Currency  string              `json:"currency"`
	Cash      float64             `json:"cash"`
	Positions map[string]Position `json:"positions"`
	Fills     []Fill              `json:"fills"`
}

func NewState(currency string, cash float64) *State {
	return &State{Currency: currency, Cash: cash, Positions: map[string]Position{}}
}

func (s *State) Buy(order Order) (Fill, error) {
	cost := order.Price * order.Size
	if cost > s.Cash {
		return Fill{}, fmt.Errorf("insufficient paper cash")
	}
	s.Cash -= cost
	pos := s.Positions[order.TokenID]
	pos.TokenID = order.TokenID
	pos.Size += order.Size
	pos.Cost += cost
	s.Positions[order.TokenID] = pos
	fill := Fill{MarketID: order.MarketID, TokenID: order.TokenID, Price: order.Price, Size: order.Size, Live: false}
	s.Fills = append(s.Fills, fill)
	return fill, nil
}
```

- [ ] **Step 4: Run paper tests**

Run:

```bash
go test ./internal/paper -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/paper/state.go internal/paper/state_test.go
git commit -m "feat: add local paper state"
```

## Task 8: Read-Only Protocol Clients

**Files:**
- Create: `internal/gamma/client_test.go`
- Create: `internal/gamma/client.go`
- Create: `internal/clob/client_test.go`
- Create: `internal/clob/client.go`

- [ ] **Step 1: Write failing Gamma mock-server test**

Create `internal/gamma/client_test.go`:

```go
package gamma

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActiveMarketsUsesContextAndParsesMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Fatalf("path = %q, want /markets", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1","slug":"m-1","question":"Will it rain?","active":true,"closed":false}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	markets, err := client.ActiveMarkets(context.Background())
	if err != nil {
		t.Fatalf("ActiveMarkets returned error: %v", err)
	}
	if len(markets) != 1 || markets[0].Slug != "m-1" {
		t.Fatalf("unexpected markets: %#v", markets)
	}
}
```

- [ ] **Step 2: Write failing CLOB mock-server test**

Create `internal/clob/client_test.go`:

```go
package clob

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderBookGetUsesReadOnlyEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/book" {
			t.Fatalf("path = %q, want /book", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"market":"token-1","bids":[{"price":"0.40","size":"12"}],"asks":[{"price":"0.60","size":"8"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	book, err := client.OrderBook(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("OrderBook returned error: %v", err)
	}
	if book.Market != "token-1" {
		t.Fatalf("Market = %q, want token-1", book.Market)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run:

```bash
go test ./internal/gamma ./internal/clob -count=1
```

Expected: FAIL because both clients are undefined.

- [ ] **Step 4: Write minimal Gamma client**

Create `internal/gamma/client.go`:

```go
package gamma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Market struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Question string `json:"question"`
	Active   bool   `json:"active"`
	Closed   bool   `json:"closed"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, http: httpClient}
}

func (c *Client) ActiveMarkets(ctx context.Context) ([]Market, error) {
	u, err := url.Parse(c.baseURL + "/markets")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("active", "true")
	q.Set("closed", "false")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("gamma status %d", resp.StatusCode)
	}
	var markets []Market
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, err
	}
	return markets, nil
}
```

- [ ] **Step 5: Write minimal CLOB client**

Create `internal/clob/client.go`:

```go
package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type OrderBook struct {
	Market string  `json:"market"`
	Bids   []Level `json:"bids"`
	Asks   []Level `json:"asks"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, http: httpClient}
}

func (c *Client) OrderBook(ctx context.Context, tokenID string) (OrderBook, error) {
	u, err := url.Parse(c.baseURL + "/book")
	if err != nil {
		return OrderBook{}, err
	}
	q := u.Query()
	q.Set("token_id", tokenID)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return OrderBook{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return OrderBook{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return OrderBook{}, fmt.Errorf("clob status %d", resp.StatusCode)
	}
	var book OrderBook
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return OrderBook{}, err
	}
	return book, nil
}
```

- [ ] **Step 6: Run client tests**

Run:

```bash
go test ./internal/gamma ./internal/clob -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/gamma/client.go internal/gamma/client_test.go internal/clob/client.go internal/clob/client_test.go
git commit -m "feat: add read-only protocol clients"
```

## Task 9: Cobra Commands For Phase 1

**Files:**
- Modify: `internal/cli/root.go`
- Modify: `internal/cli/root_test.go`
- Create: `internal/cli/commands_test.go`

- [ ] **Step 1: Write failing command visibility test**

Create `internal/cli/commands_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpListsPhaseOneCommands(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	help := stdout.String()
	for _, want := range []string{"preflight", "markets", "orderbook", "prices", "paper", "auth", "live"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
}
```

- [ ] **Step 2: Run command test to verify it fails**

Run:

```bash
go test ./internal/cli -run TestHelpListsPhaseOneCommands -count=1
```

Expected: FAIL because Phase 1 commands are not yet attached.

- [ ] **Step 3: Add minimal command skeletons that do not mutate state**

Modify `internal/cli/root.go` by adding these subcommands inside `NewRootCommand`:

```go
	root.AddCommand(&cobra.Command{
		Use:   "preflight",
		Short: "Run safety and connectivity preflight checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "preflight: not configured")
			return err
		},
	})
	root.AddCommand(&cobra.Command{Use: "markets", Short: "Read Polymarket market data"})
	root.AddCommand(&cobra.Command{Use: "orderbook", Short: "Read CLOB order books"})
	root.AddCommand(&cobra.Command{Use: "prices", Short: "Read CLOB prices"})
	root.AddCommand(&cobra.Command{Use: "paper", Short: "Manage local paper trading state"})
	root.AddCommand(&cobra.Command{Use: "auth", Short: "Inspect auth readiness"})
	root.AddCommand(&cobra.Command{Use: "live", Short: "Inspect live trading gate status"})
```

- [ ] **Step 4: Run command tests**

Run:

```bash
go test ./internal/cli -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/cli/root.go internal/cli/root_test.go internal/cli/commands_test.go
git commit -m "feat: add phase one command skeletons"
```

## Task 10: Repository Replacement Commit

**Files:**
- Delete: Rust source and release files from upstream clone.
- Keep: `.gitattributes`, `.gitignore`, `README.md` until Go docs replace them.
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Confirm Go tests pass before deleting Rust files**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Delete Rust-specific files**

Use `trash` where available:

```bash
trash Cargo.toml Cargo.lock Formula install.sh scripts src tests .github/workflows/release.yml
```

If `trash` is unavailable, stop and ask for confirmation before deleting tracked files.

- [ ] **Step 3: Write Go CI workflow**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: gofmt -w .
      - run: git diff --exit-code
      - run: go vet ./...
      - run: go test ./...
```

- [ ] **Step 4: Verify replacement diff**

Run:

```bash
git status --short
git diff --stat
```

Expected: Rust files are deleted, Go files and docs are present, and no unsafe prototype `pkg/` live trading paths remain.

- [ ] **Step 5: Run full verification**

Run:

```bash
gofmt -w .
go vet ./...
go test ./...
```

Expected: all commands exit 0.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "chore: replace rust clone with go phase one foundation"
```

## Task 11: Phase 1 Documentation

**Files:**
- Create: `docs/ARCHITECTURE.md`
- Create: `docs/COMMANDS.md`
- Create: `docs/SAFETY.md`

- [ ] **Step 1: Write architecture doc**

Create `docs/ARCHITECTURE.md`:

```markdown
# Architecture

`polymarket-cli-go` is a Go CLI built around internal typed packages and a thin
Cobra command layer.

## Dependency Flow

CLI commands call application services. Services call typed protocol clients or
local paper state. Protocol clients use `context.Context`, timeouts, typed
responses, and structured errors.

## Package Boundaries

- `internal/cli`: command wiring only
- `internal/config`: explicit Viper config loading
- `internal/modes`: read-only, paper, and live mode gates
- `internal/preflight`: readiness checks
- `internal/output`: JSON/table output
- `internal/gamma`: read-only Gamma client
- `internal/clob`: read-only CLOB client
- `internal/paper`: local simulation state

## Public SDK

No public `pkg/polymarket` SDK exists in Phase 1. A public SDK boundary should
only be introduced after internal APIs prove stable.
```

- [ ] **Step 2: Write commands doc**

Create `docs/COMMANDS.md`:

```markdown
# Commands

Every command supports automation-friendly output. Use `--json` when scripting.

## Core

- `polymarket version`
- `polymarket preflight`

## Markets

- `polymarket markets search`
- `polymarket markets get`
- `polymarket markets active`

## Market Data

- `polymarket orderbook get`
- `polymarket prices get`

## Paper Trading

- `polymarket paper buy`
- `polymarket paper sell`
- `polymarket paper positions`
- `polymarket paper reset`

## Status

- `polymarket auth status`
- `polymarket live status`
```

- [ ] **Step 3: Write safety doc**

Create `docs/SAFETY.md`:

```markdown
# Safety

Read-only mode is the default and requires no wallet credentials.

Paper mode uses local persisted state only. It must never call live order
endpoints, signing endpoints, or on-chain transaction paths.

Live mode is hard-disabled by default. Future live trading requires all gates:

- `POLYMARKET_LIVE_PROFILE=on`
- `live_trading_enabled: true`
- `--confirm-live`
- successful preflight validation

If any gate fails, the CLI aborts with a structured error. It never silently
downgrades to paper mode or read-only mode.
```

- [ ] **Step 4: Verify docs mention no unsupported live trading**

Run:

```bash
rg -n "<unsafe live claim terms>" docs README.md
```

Expected: no claim that Phase 1 supports real execution.

- [ ] **Step 5: Commit**

```bash
git add docs/ARCHITECTURE.md docs/COMMANDS.md docs/SAFETY.md
git commit -m "docs: document go cli architecture commands and safety"
```

## Task 12: Final Verification And Push

**Files:**
- No new files unless verification reveals a failing test that requires a TDD fix.

- [ ] **Step 1: Run full verification**

Run:

```bash
gofmt -w .
go vet ./...
go test ./...
```

Expected: all commands exit 0.

- [ ] **Step 2: Inspect git status**

Run:

```bash
git status --short --branch
```

Expected: clean worktree on `main` before push.

- [ ] **Step 3: Push to TrebuchetDynamics remote**

Run:

```bash
git push trebuchet main
```

Expected: push exits 0 and updates `TrebuchetDynamics/polymarket-cli-go`.

- [ ] **Step 4: Capture final evidence**

Run:

```bash
git rev-parse --short HEAD
git ls-remote trebuchet refs/heads/main
```

Expected: local HEAD and remote `refs/heads/main` match.

## Self-Review Checklist

- Every production Go task starts with a failing test.
- Existing unsafe prototype code is not trusted or committed as-is.
- Live order placement remains a non-goal for Phase 1.
- Paper trading is local-only.
- JSON output has a stable envelope.
- Protocol clients are read-only and tested with mock servers.
- Docs do not claim unsupported live trading.
- Final verification uses `gofmt -w .`, `go vet ./...`, and `go test ./...`.
