# Track 2 — Godoc Layer: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Anyone reading the polygolem source can orient inside ~30 seconds per package. The public SDK (`pkg/*`) becomes genuinely usable on `pkg.go.dev` with package-level prose, every exported symbol documented, and at least one runnable `Example_*` per package.

**Architecture:** Two-tier strategy.

- **Tier A** — A `doc.go` file in every internal package containing only a package-level doc comment. Orientation only. Implementation files keep their existing comments where present; we either move those comments verbatim into `doc.go` or replace them with the canonical text from this plan when missing or thin.
- **Tier B** — Full godoc on every exported symbol in each `pkg/*` package, plus a `*_example_test.go` with at least one runnable `Example_*` function. Public-surface only — no `internal/...` references in **doc text** (signatures already leak `internal/*` types in some packages; that is a pre-existing API design issue logged below, not an in-scope fix).

**Tech Stack:** Go toolchain only (`go vet`, `go test`, `go doc`). No new dependencies. No `.revive.toml` (deferred per spec).

**Spec:** `docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § Track 2.

**Working tree caveat:** `main` carries substantial uncommitted WIP. Every task below ships an explicit per-task file allowlist. Use `git add <listed paths>` only — **never** `git add -A` or `git add .`. Documentation tasks must not modify code logic.

**Track 1 has shipped.** `docs/ARCHITECTURE.md` is canonical and lists 21 internal + 5 `pkg/` packages. The one-sentence purposes used below are the same ones in that file's Surface Map; keep them in sync if the architecture doc changes during Track 2.

---

## Task Inventory

| # | Task | Output | Commit? |
|---|---|---|---|
| 2A | `doc.go` for protocol clients (gamma, clob, dataapi, stream, relayer, rpc) — 6 packages | 6 new files | Yes (1 commit) |
| 2B | `doc.go` for cross-cutting primitives (auth, transport, polytypes, errors, output, config) — 6 packages | 6 new files | Yes (1 commit) |
| 2C | `doc.go` for execution layer (orders, execution, risk, paper, wallet, marketdiscovery) — 6 packages | 6 new files | Yes (1 commit) |
| 2D | `doc.go` for CLI / mode plumbing (cli, modes, preflight) — 3 packages | 3 new files | Yes (1 commit) |
| 2E | Full godoc + Example for `pkg/bookreader` | Modified + 1 new test file | Yes (1 commit) |
| 2F | Full godoc + Example for `pkg/bridge` | Modified + 1 new test file | Yes (1 commit) |
| 2G | Full godoc + Example for `pkg/gamma` | Modified + 1 new test file | Yes (1 commit) |
| 2H | Full godoc + Example for `pkg/marketresolver` | Modified | Yes (1 commit) |
| 2I | Full godoc + Example for `pkg/pagination` | Modified | Yes (1 commit) |
| 2J | Final Track 2 verification gate | Read-only verification | No |

Each task is one commit unless noted. Tasks 2E and 2F create a new `*_example_test.go` because no test file exists yet (bridge) or example tests are not present (bookreader has only a unit test). Tasks 2G/2H/2I add `Example_*` to existing or new files as listed per task.

---

## Audit-the-existing-comments rule (applies to Tasks 2A–2D)

Several internal packages already carry a one-line package comment on a non-`doc.go` file (verified at plan-write time). When you write `doc.go`, you must:

1. Use the canonical text given in the task (purpose + elaboration).
2. Then **delete** any pre-existing `// Package <name> ...` block on the implementation file (`*.go`, not `*_test.go`) so we don't have two competing package comments. Preserve the rest of the file unchanged.
3. If a package has no existing comment, only step 1 applies.

The pre-existing comments are concentrated in these files (re-verify with `grep` before editing — WIP may have moved them):

| Package | File with existing `// Package` comment |
|---|---|
| `internal/auth` | `internal/auth/auth.go` |
| `internal/dataapi` | `internal/dataapi/client.go` |
| `internal/errors` | `internal/errors/errors.go` |
| `internal/execution` | `internal/execution/executor.go` |
| `internal/marketdiscovery` | `internal/marketdiscovery/discovery.go` |
| `internal/polytypes` | `internal/polytypes/clob.go` |
| `internal/risk` | `internal/risk/breaker.go` |
| `internal/rpc` | `internal/rpc/transfer.go` |
| `internal/stream` | `internal/stream/client.go` |
| `internal/wallet` | `internal/wallet/derive.go` |

The rest (`cli`, `clob`, `config`, `gamma`, `modes`, `orders`, `output`, `paper`, `preflight`, `relayer`, `transport`) currently have a bare `package <name>` on the implementation file — only step 1 is required for those.

---

## Task 2A: `doc.go` for protocol clients (gamma, clob, dataapi, stream, relayer, rpc)

**Files:**
- Create: `internal/gamma/doc.go`
- Create: `internal/clob/doc.go`
- Create: `internal/dataapi/doc.go`
- Create: `internal/stream/doc.go`
- Create: `internal/relayer/doc.go`
- Create: `internal/rpc/doc.go`
- Modify: `internal/dataapi/client.go` (delete pre-existing 3-line `// Package dataapi …` block above `package dataapi`)
- Modify: `internal/stream/client.go` (delete pre-existing 3-line `// Package stream …` block above `package stream`)
- Modify: `internal/rpc/transfer.go` (delete pre-existing 1-line `// Package rpc …` block above `package rpc`)

**File allowlist (for `git add`):**
```
internal/gamma/doc.go
internal/clob/doc.go
internal/dataapi/doc.go
internal/dataapi/client.go
internal/stream/doc.go
internal/stream/client.go
internal/relayer/doc.go
internal/rpc/doc.go
internal/rpc/transfer.go
```

- [ ] **Step 1: Re-verify the existing-comment audit for this group**

```bash
grep -nE "^// Package " internal/gamma/*.go internal/clob/*.go internal/dataapi/*.go internal/stream/*.go internal/relayer/*.go internal/rpc/*.go 2>/dev/null
```

Expected: hits in `internal/dataapi/client.go`, `internal/stream/client.go`, `internal/rpc/transfer.go` and **only** those. If the WIP tree has moved them, adjust the implementation-file edits below to match where the `// Package` comment now lives.

- [ ] **Step 2: Write `internal/gamma/doc.go`**

Exact content:

```go
// Package gamma is the typed Gamma HTTP client used internally by polygolem
// — markets, events, search, tags, series, sports, comments, and profiles.
//
// Gamma is Polymarket's read-only metadata API. This client wraps it with
// retry, rate-limiting, and structured types from internal/polytypes. It
// performs no signing and never mutates state. Start with Client and
// Client.Search for orientation.
//
// This package is internal and not part of the polygolem public SDK.
// External consumers should use pkg/gamma instead.
package gamma
```

- [ ] **Step 3: Write `internal/clob/doc.go`**

Exact content:

```go
// Package clob is the CLOB API client — full read plus authenticated
// surface, EIP-712, POLY_1271, and ERC-7739 signing paths.
//
// Wraps Polymarket's central limit order book API. Read endpoints (books,
// midpoints, trades, markets) are usable without credentials. Mutating
// endpoints (create/cancel orders) require an L1 or L2 auth header from
// internal/auth and must be invoked only from live mode after gates pass.
// Start with Client for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package clob
```

- [ ] **Step 4: Write `internal/dataapi/doc.go`**

Exact content:

```go
// Package dataapi is the read-only Polymarket Data API client — positions,
// volume, trades history, and leaderboards.
//
// Base URL: https://data-api.polymarket.com. The client uses the shared
// internal/transport retry and rate-limiter and does not sign requests.
// Start with Client for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package dataapi
```

- [ ] **Step 5: Write `internal/stream/doc.go`**

Exact content:

```go
// Package stream provides typed WebSocket clients for Polymarket CLOB
// market streams with reconnect and event deduplication.
//
// The market client subscribes to public order book and trade updates for
// a set of asset IDs. Authenticated user streams are not implemented here.
// Start with MarketClient and SubscribeAssets for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package stream
```

- [ ] **Step 6: Write `internal/relayer/doc.go`**

Exact content:

```go
// Package relayer is the builder relayer client — WALLET-CREATE,
// WALLET batch, nonce reads, and operation polling.
//
// Used by the deposit-wallet flow to deploy a CREATE2 deposit wallet,
// submit batched calls under POLY_1271, and poll relayer operations until
// they reach a terminal state. Requires builder credentials supplied by
// internal/config. Start with Client for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package relayer
```

- [ ] **Step 7: Write `internal/rpc/doc.go`**

Exact content:

```go
// Package rpc provides direct on-chain helpers for Polygon operations
// — primarily ERC-20 pUSD transfers from an EOA used by deposit-wallet
// funding.
//
// Calls go through a configured Polygon RPC endpoint and bypass the
// Polymarket relayer. Used only by live-mode commands behind the funding
// gate; no read-only or paper-mode code path depends on it.
//
// This package is internal and not part of the polygolem public SDK.
package rpc
```

- [ ] **Step 8: Strip the now-redundant comments on the three implementation files**

For `internal/dataapi/client.go` — delete only the 3 leading comment lines so the file begins with `package dataapi`:

```go
// Package dataapi provides read-only access to the Polymarket Data API.
// Base URL: https://data-api.polymarket.com
// Stolen patterns from ybina/polymarket-go and polymarket-kit.
package dataapi
```

becomes:

```go
package dataapi
```

For `internal/stream/client.go` — delete only the 2 leading comment lines:

```go
// Package stream provides typed WebSocket clients for Polymarket CLOB streams.
// Patterns stolen from polymarket-kit and polymarket-go-sdk.
package stream
```

becomes:

```go
package stream
```

For `internal/rpc/transfer.go` — delete only the 1 leading comment line:

```go
// Package rpc provides direct on-chain helpers for Polygon operations.
package rpc
```

becomes:

```go
package rpc
```

Use the Edit tool with the leading-comment-line-plus-`package <name>` block as `old_string` and `package <name>` as `new_string`. Do not touch anything else in those files.

- [ ] **Step 9: Verify the build and the doc**

```bash
go build ./internal/gamma ./internal/clob ./internal/dataapi ./internal/stream ./internal/relayer ./internal/rpc
go vet ./internal/gamma ./internal/clob ./internal/dataapi ./internal/stream ./internal/relayer ./internal/rpc
for p in gamma clob dataapi stream relayer rpc; do
  echo "===== internal/$p ====="
  go doc ./internal/$p | head -5
done
```

Expected: clean build, clean vet, and each `go doc` shows the new prose starting with `Package <name> ...`. If a package shows the old prose, the implementation-file strip was missed.

- [ ] **Step 10: Commit**

```bash
git add internal/gamma/doc.go internal/clob/doc.go internal/dataapi/doc.go internal/dataapi/client.go internal/stream/doc.go internal/stream/client.go internal/relayer/doc.go internal/rpc/doc.go internal/rpc/transfer.go
git commit -m "$(cat <<'EOF'
docs(godoc): add doc.go for protocol clients (gamma, clob, dataapi, stream, relayer, rpc)

Tier-A package comments — orientation only, one file per package. Strips the
redundant package comments from internal/dataapi/client.go,
internal/stream/client.go, and internal/rpc/transfer.go so each package has
exactly one canonical doc surface.

No code logic changed. No new exports.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2B: `doc.go` for cross-cutting primitives (auth, transport, polytypes, errors, output, config)

**Files:**
- Create: `internal/auth/doc.go`
- Create: `internal/transport/doc.go`
- Create: `internal/polytypes/doc.go`
- Create: `internal/errors/doc.go`
- Create: `internal/output/doc.go`
- Create: `internal/config/doc.go`
- Modify: `internal/auth/auth.go` (delete the 2-line `// Package auth …` block)
- Modify: `internal/polytypes/clob.go` (delete the 1-line `// Package polytypes …` block)
- Modify: `internal/errors/errors.go` (delete the 2-line `// Package errors …` block)

**File allowlist (for `git add`):**
```
internal/auth/doc.go
internal/auth/auth.go
internal/transport/doc.go
internal/polytypes/doc.go
internal/polytypes/clob.go
internal/errors/doc.go
internal/errors/errors.go
internal/output/doc.go
internal/config/doc.go
```

- [ ] **Step 1: Re-verify which implementation files carry pre-existing comments**

```bash
grep -nE "^// Package " internal/auth/*.go internal/transport/*.go internal/polytypes/*.go internal/errors/*.go internal/output/*.go internal/config/*.go 2>/dev/null
```

Expected: hits in `internal/auth/auth.go`, `internal/polytypes/clob.go`, `internal/errors/errors.go`. If hits elsewhere, adjust the strip-step to wherever the comment now lives.

- [ ] **Step 2: Write `internal/auth/doc.go`**

```go
// Package auth provides Polymarket authentication primitives — L0 / L1 / L2
// auth, EIP-712 signing, deposit-wallet CREATE2 derivation, and builder
// attribution.
//
// Used by every signed request to the CLOB and relayer. The default mode
// for polygolem is read-only and never enters this package; mutating
// commands acquire signers here behind explicit gates. Start with the
// Signer types and DeriveDepositWallet for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package auth
```

- [ ] **Step 3: Write `internal/transport/doc.go`**

```go
// Package transport is the shared HTTP client layer — retry, rate limiting,
// circuit breaking, and credential redaction.
//
// Every protocol client (gamma, clob, dataapi, relayer, bridge) sits on
// top of this. Configure with DefaultConfig and inject via Client. The
// redactor is wired in by internal/config so credentials never reach
// stdout or logs.
//
// This package is internal and not part of the polygolem public SDK.
package transport
```

- [ ] **Step 4: Write `internal/polytypes/doc.go`**

```go
// Package polytypes holds the Polymarket protocol-level types shared
// across CLOB, Gamma, Data API, and stream clients.
//
// These types mirror the on-the-wire JSON shapes Polymarket returns.
// Keeping them in one place avoids per-client drift and lets paper-mode
// and live-mode reuse the same structures. There is no logic here —
// only types and small helpers.
//
// This package is internal and not part of the polygolem public SDK.
package polytypes
```

- [ ] **Step 5: Write `internal/errors/doc.go`**

```go
// Package errors provides structured error types and code helpers used
// across polygolem clients and command handlers.
//
// Each error carries a stable code suitable for surfacing in the JSON
// envelope rendered by internal/output. Wrap protocol errors with the
// helpers here rather than reaching for fmt.Errorf so downstream
// consumers can switch on Code.
//
// This package is internal and not part of the polygolem public SDK.
package errors
```

- [ ] **Step 6: Write `internal/output/doc.go`**

```go
// Package output renders command results as either tables or stable JSON
// envelopes and emits structured error responses.
//
// Tables are designed for humans; the JSON envelope is the contract every
// command handler honors when --json is set. Handlers should call into
// this package rather than printing directly so the envelope shape stays
// stable.
//
// This package is internal and not part of the polygolem public SDK.
package output
```

- [ ] **Step 7: Write `internal/config/doc.go`**

```go
// Package config loads polygolem configuration via Viper — defaults,
// environment binding, file overrides, validation, and credential
// redaction.
//
// Every entry point reads config through Load. Builder credentials and
// private keys are redacted at load time so no downstream logger or JSON
// emitter ever sees the plaintext value.
//
// This package is internal and not part of the polygolem public SDK.
package config
```

- [ ] **Step 8: Strip pre-existing comments on the three implementation files**

For `internal/auth/auth.go` — replace:

```go
// Package auth provides Polymarket authentication primitives.
// Based on patterns from polymarket-go (ybina), polymarket-go-sdk, and go-builder-signing-sdk.
package auth
```

with:

```go
package auth
```

For `internal/polytypes/clob.go` — replace:

```go
// Package polytypes — CLOB types stolen from polymarket-go-sdk and rs-clob-client.
package polytypes
```

with:

```go
package polytypes
```

For `internal/errors/errors.go` — replace:

```go
// Package errors provides structured error codes for polygolem.
// Error taxonomy stolen from polymarket-go-sdk/pkg/errors.
package errors
```

with:

```go
package errors
```

- [ ] **Step 9: Verify**

```bash
go build ./internal/auth ./internal/transport ./internal/polytypes ./internal/errors ./internal/output ./internal/config
go vet ./internal/auth ./internal/transport ./internal/polytypes ./internal/errors ./internal/output ./internal/config
for p in auth transport polytypes errors output config; do
  echo "===== internal/$p ====="
  go doc ./internal/$p | head -5
done
```

Expected: clean build, clean vet, each `go doc` shows the new prose.

- [ ] **Step 10: Commit**

```bash
git add internal/auth/doc.go internal/auth/auth.go internal/transport/doc.go internal/polytypes/doc.go internal/polytypes/clob.go internal/errors/doc.go internal/errors/errors.go internal/output/doc.go internal/config/doc.go
git commit -m "$(cat <<'EOF'
docs(godoc): add doc.go for cross-cutting primitives (auth, transport, polytypes, errors, output, config)

Tier-A package comments. Strips the redundant package comments from
internal/auth/auth.go, internal/polytypes/clob.go, and
internal/errors/errors.go.

No code logic changed. No new exports.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2C: `doc.go` for execution layer (orders, execution, risk, paper, wallet, marketdiscovery)

**Files:**
- Create: `internal/orders/doc.go`
- Create: `internal/execution/doc.go`
- Create: `internal/risk/doc.go`
- Create: `internal/paper/doc.go`
- Create: `internal/wallet/doc.go`
- Create: `internal/marketdiscovery/doc.go`
- Modify: `internal/execution/executor.go` (delete the 2-line `// Package execution …` block)
- Modify: `internal/risk/breaker.go` (delete the 1-line `// Package risk …` block)
- Modify: `internal/wallet/derive.go` (delete the 2-line `// Package wallet …` block)
- Modify: `internal/marketdiscovery/discovery.go` (delete the 1-line `// Package marketdiscovery …` block)

**File allowlist (for `git add`):**
```
internal/orders/doc.go
internal/execution/doc.go
internal/execution/executor.go
internal/risk/doc.go
internal/risk/breaker.go
internal/paper/doc.go
internal/wallet/doc.go
internal/wallet/derive.go
internal/marketdiscovery/doc.go
internal/marketdiscovery/discovery.go
```

- [ ] **Step 1: Re-verify which implementation files carry pre-existing comments**

```bash
grep -nE "^// Package " internal/orders/*.go internal/execution/*.go internal/risk/*.go internal/paper/*.go internal/wallet/*.go internal/marketdiscovery/*.go 2>/dev/null
```

Expected: hits in `internal/execution/executor.go`, `internal/risk/breaker.go`, `internal/wallet/derive.go`, `internal/marketdiscovery/discovery.go`.

- [ ] **Step 2: Write `internal/orders/doc.go`**

```go
// Package orders defines OrderIntent, the fluent builder, validation
// rules, and order lifecycle states used by both paper and live executors.
//
// An OrderIntent is the protocol-agnostic shape an executor accepts.
// Construction goes through Builder so invariants (size, side, market,
// signature type) are checked before any network call. Lifecycle states
// are explicit so paper and live can share the same state machine.
//
// This package is internal and not part of the polygolem public SDK.
package orders
```

- [ ] **Step 3: Write `internal/execution/doc.go`**

```go
// Package execution defines the executor interface and ships the
// paper-mode implementation. A live executor satisfies the same contract.
//
// Executors take a validated OrderIntent and return a typed result.
// Paper executors update internal/paper state; live executors call
// internal/clob and internal/relayer behind the live-mode gate. Handlers
// depend on the interface, never on a concrete executor.
//
// This package is internal and not part of the polygolem public SDK.
package execution
```

- [ ] **Step 4: Write `internal/risk/doc.go`**

```go
// Package risk provides per-trade caps, daily loss limits, and the
// circuit breaker that gates live order submission.
//
// Live commands consult risk before any signing or submission step.
// Read-only and paper modes do not call into this package. Limits are
// configured in internal/config and never derived from market data.
//
// This package is internal and not part of the polygolem public SDK.
package risk
```

- [ ] **Step 5: Write `internal/paper/doc.go`**

```go
// Package paper holds local-only paper-trading state — positions, fills,
// and persisted snapshots.
//
// Paper state lives entirely on disk and never reaches an authenticated
// Polymarket endpoint. The paper executor in internal/execution writes
// here; live mode does not touch this package. Useful for replay, edge
// validation, and offline development.
//
// This package is internal and not part of the polygolem public SDK.
package paper
```

- [ ] **Step 6: Write `internal/wallet/doc.go`**

```go
// Package wallet provides deposit-wallet primitives — CREATE2 derivation,
// status checks, deploy and batch-signing helpers.
//
// Address derivation is non-mutating and used by read-only deposit-wallet
// commands. Deploy and batch operations sit behind builder credentials
// and the live gate. See docs/DEPOSIT-WALLET-MIGRATION.md for the May 2026
// signature-type migration this package implements.
//
// This package is internal and not part of the polygolem public SDK.
package wallet
```

- [ ] **Step 7: Write `internal/marketdiscovery/doc.go`**

```go
// Package marketdiscovery provides high-level market discovery by
// joining Gamma metadata with CLOB tick-size and orderbook details.
//
// Wraps internal/gamma and internal/clob so command handlers (`discover`,
// `discover enrich`, `discover market`) can return one denormalized view
// instead of stitching Gamma plus CLOB calls per command. Read-only;
// safe in every mode.
//
// This package is internal and not part of the polygolem public SDK.
package marketdiscovery
```

- [ ] **Step 8: Strip pre-existing comments on the four implementation files**

`internal/execution/executor.go` — replace:

```go
// Package execution provides the order execution interface.
// Paper and live implementations share the same contract.
package execution
```

with `package execution`.

`internal/risk/breaker.go` — replace:

```go
// Package risk provides per-trade caps, limits, and circuit breakers.
package risk
```

with `package risk`.

`internal/wallet/derive.go` — replace:

```go
// Package wallet provides non-mutating wallet readiness checks.
// Address derivation stolen from polymarket-go and rs-clob-client.
package wallet
```

with `package wallet`.

`internal/marketdiscovery/discovery.go` — replace:

```go
// Package marketdiscovery provides market enrichment by joining Gamma metadata with CLOB details.
package marketdiscovery
```

with `package marketdiscovery`.

- [ ] **Step 9: Verify**

```bash
go build ./internal/orders ./internal/execution ./internal/risk ./internal/paper ./internal/wallet ./internal/marketdiscovery
go vet ./internal/orders ./internal/execution ./internal/risk ./internal/paper ./internal/wallet ./internal/marketdiscovery
for p in orders execution risk paper wallet marketdiscovery; do
  echo "===== internal/$p ====="
  go doc ./internal/$p | head -5
done
```

Expected: clean build, clean vet, each `go doc` shows the new prose.

- [ ] **Step 10: Commit**

```bash
git add internal/orders/doc.go internal/execution/doc.go internal/execution/executor.go internal/risk/doc.go internal/risk/breaker.go internal/paper/doc.go internal/wallet/doc.go internal/wallet/derive.go internal/marketdiscovery/doc.go internal/marketdiscovery/discovery.go
git commit -m "$(cat <<'EOF'
docs(godoc): add doc.go for execution layer (orders, execution, risk, paper, wallet, marketdiscovery)

Tier-A package comments. Strips the redundant package comments from
internal/execution/executor.go, internal/risk/breaker.go,
internal/wallet/derive.go, and internal/marketdiscovery/discovery.go.

No code logic changed. No new exports.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2D: `doc.go` for CLI / mode plumbing (cli, modes, preflight)

**Files:**
- Create: `internal/cli/doc.go`
- Create: `internal/modes/doc.go`
- Create: `internal/preflight/doc.go`

**File allowlist (for `git add`):**
```
internal/cli/doc.go
internal/modes/doc.go
internal/preflight/doc.go
```

None of these three packages currently carry a pre-existing `// Package …` comment, so this task is pure additions.

- [ ] **Step 1: Re-verify no pre-existing comments**

```bash
grep -nE "^// Package " internal/cli/*.go internal/modes/*.go internal/preflight/*.go 2>/dev/null
```

Expected: **no output**. If any file has a `// Package …` block, add a strip step modeled on Tasks 2A–2C and add that file to the allowlist.

- [ ] **Step 2: Write `internal/cli/doc.go`**

```go
// Package cli builds the polygolem Cobra command tree and wires command
// handlers to typed protocol, execution, and safety packages.
//
// Every command lives here and delegates to a typed package — handlers do
// not contain protocol logic. The default invocation enters read-only
// mode; live commands require an explicit signature type and gate pass.
// Start with NewRootCmd for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package cli
```

- [ ] **Step 3: Write `internal/modes/doc.go`**

```go
// Package modes parses and gates the polygolem operating modes —
// read-only, paper, and live.
//
// Mode selection comes from configuration and CLI flags. Read-only is the
// default and never reaches authenticated mutation endpoints. Paper stays
// local. Live requires preflight, risk, and funding gates to pass before
// any signed call goes out.
//
// This package is internal and not part of the polygolem public SDK.
package modes
```

- [ ] **Step 4: Write `internal/preflight/doc.go`**

```go
// Package preflight runs local and remote readiness probes before
// polygolem performs any state-changing operation.
//
// A Probe is a context-aware function returning an error. Probes cover
// builder credentials, deposit-wallet status, RPC reachability, and
// relayer health. The aggregate result gates entry into live mode.
//
// This package is internal and not part of the polygolem public SDK.
package preflight
```

- [ ] **Step 5: Verify**

```bash
go build ./internal/cli ./internal/modes ./internal/preflight
go vet ./internal/cli ./internal/modes ./internal/preflight
for p in cli modes preflight; do
  echo "===== internal/$p ====="
  go doc ./internal/$p | head -5
done
```

Expected: clean build, clean vet, each `go doc` shows the new prose.

- [ ] **Step 6: Verify all 21 internal packages now have a `doc.go`**

```bash
missing=0
for p in auth cli clob config dataapi errors execution gamma marketdiscovery modes orders output paper polytypes preflight relayer risk rpc stream transport wallet; do
  if [[ ! -f "internal/$p/doc.go" ]]; then
    echo "MISSING: internal/$p/doc.go"
    missing=1
  fi
done
[[ $missing -eq 0 ]] && echo "All 21 doc.go files present"
```

Expected: `All 21 doc.go files present`.

- [ ] **Step 7: Commit**

```bash
git add internal/cli/doc.go internal/modes/doc.go internal/preflight/doc.go
git commit -m "$(cat <<'EOF'
docs(godoc): add doc.go for CLI / mode plumbing (cli, modes, preflight)

Tier-A package comments — completes coverage of all 21 internal/* packages.

No code logic changed. No new exports.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2E: Full godoc + Example for `pkg/bookreader`

**Files:**
- Modify: `pkg/bookreader/reader.go` (replace package comment, add godoc on every exported symbol)
- Create: `pkg/bookreader/example_test.go` (new file with `Example_*`)

**File allowlist (for `git add`):**
```
pkg/bookreader/reader.go
pkg/bookreader/example_test.go
```

Exported surface to document (verified with `go doc ./pkg/bookreader`):

| Symbol | Kind | Notes |
|---|---|---|
| `Reader` | interface | One method `OrderBook(ctx, tokenID) (OrderBook, error)`. |
| `OrderBook` | struct | Fields: `MarketID`, `TokenID`, `Bids`, `Asks`, `LastTradePrice`. |
| `Level` | struct | Fields: `Price`, `Size`. |
| `NewReader(clobBaseURL string) Reader` | func | Constructor returning the interface. |
| `Reader.OrderBook(ctx, tokenID) (OrderBook, error)` | method | Method on the interface. |

- [ ] **Step 1: Read the file once before editing**

```bash
wc -l pkg/bookreader/reader.go
```

You'll edit only the package doc comment and add godoc above each exported declaration. Do not change any function bodies, types, or signatures.

- [ ] **Step 2: Replace the package comment**

The current file opens with:

```go
// Package bookreader provides a public BookReader interface and Polygolem implementation.
// This is the Phase 0 boundary between go-bot and polygolem — replaces direct CLOB clients.
package bookreader
```

Replace those three lines with:

```go
// Package bookreader is a read-only Polymarket CLOB order-book reader.
//
// Use bookreader when you want top-of-book price discovery for one or
// more token IDs without pulling in the full polygolem CLI. The Reader
// interface is the only public entry point; NewReader returns a
// production implementation backed by the CLOB HTTP API.
//
// When not to use this package:
//   - For authenticated CLOB operations (create or cancel orders) — those
//     are not part of the public SDK.
//   - For low-latency streaming — use a WebSocket client instead.
//
// Stability: the Reader interface, OrderBook, Level, and NewReader are
// part of the polygolem public SDK and follow semver. Internal helpers
// remain unexported and may change.
package bookreader
```

- [ ] **Step 3: Add godoc on every exported symbol**

Above `type OrderBook struct` (the existing one-line comment is fine; expand to a full sentence):

```go
// OrderBook is a snapshot of one Polymarket CLOB market.
// Bids are sorted highest-price first and Asks lowest-price first.
// LastTradePrice may be zero if the snapshot does not include a trade
// reference.
type OrderBook struct {
```

Above `type Level struct`:

```go
// Level is one price level in the order book — a single Price with the
// total Size resting at that price.
type Level struct {
```

Above `type Reader interface`:

```go
// Reader fetches CLOB order books by ERC-1155 token ID.
// Implementations must be safe for concurrent use by multiple goroutines.
type Reader interface {
	// OrderBook returns the current order-book snapshot for tokenID.
	// The returned OrderBook is sorted best-first on each side.
	OrderBook(ctx context.Context, tokenID string) (OrderBook, error)
}
```

(Keep the body of `Reader` as the existing single-method interface; only add the surrounding doc comments.)

Above `func NewReader`:

```go
// NewReader returns a Reader backed by the polygolem CLOB client at
// clobBaseURL. Pass an empty string to use the Polymarket production CLOB
// URL. The returned Reader uses the package's default HTTP transport with
// retry and rate limiting.
func NewReader(clobBaseURL string) Reader {
```

Unexported symbols (`reader`, `convertBook`, `convertLevels`, `parseFloat`) do not require godoc.

- [ ] **Step 4: Create `pkg/bookreader/example_test.go`**

This package has only `reader_test.go` today; the example goes in a new file so it renders separately on `pkg.go.dev`.

Exact content:

```go
package bookreader_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/TrebuchetDynamics/polygolem/pkg/bookreader"
)

// Example_orderBook demonstrates fetching a CLOB order-book snapshot for a
// token ID. A test HTTP server stands in for the production CLOB so the
// example is hermetic and runnable with `go test ./pkg/bookreader`.
func Example_orderBook() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"market": "condition-1",
			"asset_id": "token-1",
			"bids": [{"price": "0.42", "size": "10"}],
			"asks": [{"price": "0.58", "size": "10"}]
		}`))
	}))
	defer server.Close()

	reader := bookreader.NewReader(server.URL)
	book, err := reader.OrderBook(context.Background(), "token-1")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("best bid=%.2f best ask=%.2f\n", book.Bids[0].Price, book.Asks[0].Price)
	// Output: best bid=0.42 best ask=0.58
}
```

- [ ] **Step 5: Verify build, vet, tests, and rendered doc**

```bash
go build ./pkg/bookreader
go vet ./pkg/bookreader
go test ./pkg/bookreader
go doc ./pkg/bookreader
go doc ./pkg/bookreader.Reader
go doc ./pkg/bookreader.OrderBook
go doc ./pkg/bookreader.NewReader
```

Expected: build and vet clean, tests green (the new `Example_orderBook` runs as a test), and each `go doc` invocation shows the new prose. The package overview should mention "read-only Polymarket CLOB order-book reader".

- [ ] **Step 6: Commit**

```bash
git add pkg/bookreader/reader.go pkg/bookreader/example_test.go
git commit -m "$(cat <<'EOF'
docs(godoc): full godoc and runnable Example for pkg/bookreader

Adds Tier-B package doc with use-cases and stability promise. Documents
every exported symbol (Reader, OrderBook, Level, NewReader). Adds
Example_orderBook in a new pkg/bookreader/example_test.go using httptest
so the example is hermetic and runs under `go test`.

No new exports. No code logic changed.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2F: Full godoc + Example for `pkg/bridge`

**Files:**
- Modify: `pkg/bridge/bridge.go` (replace package comment, add godoc on every exported symbol)
- Create: `pkg/bridge/example_test.go` (new file with `Example_*`)

**File allowlist (for `git add`):**
```
pkg/bridge/bridge.go
pkg/bridge/example_test.go
```

Exported surface to document (verified with `go doc ./pkg/bridge`):

- Constructor: `NewClient(baseURL string, tc *transport.Client) *Client`
- Client type with methods: `CreateDepositAddress`, `GetSupportedAssets`, `GetDepositStatus`, `GetQuote`
- Types: `DepositAddress`, `CreateDepositAddressResponse`, `TokenInfo`, `SupportedAsset`, `SupportedAssetsResponse`, `DepositTransaction`, `DepositStatusResponse`, `QuoteRequest`, `FeeBreakdown`, `QuoteResponse`

> **Public-surface caveat:** `NewClient` takes `*transport.Client` from `internal/transport`. Per spec, **no `internal/...` references in doc TEXT**. The signature itself is a pre-existing leak; we document it by saying "pass nil to use the default transport." The example must use `bridge.NewClient(server.URL, nil)`. If a reviewer wants the leak fixed, log it on `docs/AUDIT-FINDINGS.md` under `Track 1 — Code-side drift surfaced incidentally` — do not change the signature in this task.

- [ ] **Step 1: Replace the package comment**

The current file opens with:

```go
// Package bridge provides read-only Bridge API readiness checks.
// Stolen from ybina/polymarket-go/client/bridge/bridge.go.
// Base URL: https://bridge.polymarket.com
package bridge
```

Replace those four lines with:

```go
// Package bridge is a client for the Polymarket Bridge API — supported
// assets, deposit addresses, deposit-status polling, and quotes.
//
// Use bridge to discover which assets can be bridged into Polymarket and
// to surface a deposit address for an EOA. The client is HTTP-only and
// performs no signing; it is safe to use in read-only mode.
//
// When not to use this package:
//   - For on-chain transfers — use a Polygon RPC client directly.
//   - For order placement — see the polygolem CLOB surface.
//
// Stability: the Client constructor, methods, and request/response types
// are part of the polygolem public SDK and follow semver. Pass a nil
// transport to NewClient to use the package default; advanced callers may
// supply their own.
package bridge
```

- [ ] **Step 2: Add godoc on every exported symbol**

Above `const defaultBridgeBaseURL`: leave as-is (unexported).

Above `type Client struct` — the existing one-line comment is acceptable; expand to:

```go
// Client provides read-only access to the Polymarket Bridge API.
// Construct via NewClient. Methods are safe for concurrent use.
type Client struct {
```

Above `func NewClient`:

```go
// NewClient returns a Bridge API client.
// If baseURL is empty, the production Bridge URL is used.
// If tc is nil, a default transport with retry and rate limiting is
// constructed.
func NewClient(baseURL string, tc *transport.Client) *Client {
```

For each exported type in the `// --- Types ---` block, add a one-line godoc comment immediately above the `type` line:

```go
// DepositAddress carries the per-chain deposit addresses returned by the
// Bridge for a given Polymarket account.
type DepositAddress struct { ... }

// CreateDepositAddressResponse is the response shape for POST /deposit.
type CreateDepositAddressResponse struct { ... }

// TokenInfo describes one token (name, symbol, address, decimals) as
// reported by the Bridge.
type TokenInfo struct { ... }

// SupportedAsset is one entry in the Bridge's supported-assets list,
// pairing a chain with the token usable as deposit collateral.
type SupportedAsset struct { ... }

// SupportedAssetsResponse is the response shape for GET /supported-assets.
type SupportedAssetsResponse struct { ... }

// DepositTransaction describes one deposit attempt observed by the Bridge.
// Status is a Bridge-defined string; clients should treat unknown values
// as opaque.
type DepositTransaction struct { ... }

// DepositStatusResponse is the response shape for GET /status/{address}.
type DepositStatusResponse struct { ... }

// QuoteRequest is the input to GetQuote — the source token and amount,
// recipient, and target token on the Polymarket side.
type QuoteRequest struct { ... }

// FeeBreakdown enumerates the fee components a Bridge quote includes.
// All percent fields are expressed as a fraction (0.01 = 1%).
type FeeBreakdown struct { ... }

// QuoteResponse is the response shape for POST /quote — estimated input
// and output USD, an estimated time, the fee breakdown, and a quote ID
// that the caller must echo when accepting the quote.
type QuoteResponse struct { ... }
```

For each exported method in the `// --- Methods ---` block:

```go
// CreateDepositAddress requests the Bridge mint a deposit address for the
// given Polymarket-side address. The Bridge returns a per-chain address
// set; only one of EVM/SVM/BTC is typically populated per request.
func (c *Client) CreateDepositAddress(ctx context.Context, address string) (*CreateDepositAddressResponse, error) {

// GetSupportedAssets returns the assets the Bridge currently accepts as
// deposit collateral.
func (c *Client) GetSupportedAssets(ctx context.Context) (*SupportedAssetsResponse, error) {

// GetDepositStatus polls the Bridge for outstanding and recent deposit
// transactions targeting depositAddress.
func (c *Client) GetDepositStatus(ctx context.Context, depositAddress string) (*DepositStatusResponse, error) {

// GetQuote asks the Bridge to price a deposit move described by req.
// The returned QuoteID is the handle the caller will echo in a follow-up
// accept call.
func (c *Client) GetQuote(ctx context.Context, req QuoteRequest) (*QuoteResponse, error) {
```

- [ ] **Step 3: Create `pkg/bridge/example_test.go`**

Bridge has no test file yet. Create one solely for the example.

Exact content:

```go
package bridge_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/TrebuchetDynamics/polygolem/pkg/bridge"
)

// Example_supportedAssets demonstrates listing the Bridge's supported
// deposit assets. A test HTTP server stands in for the production Bridge
// so the example is hermetic and runs under `go test`.
func Example_supportedAssets() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"supportedAssets": [{
				"chainId": "137",
				"chainName": "Polygon",
				"token": {"name": "USDC", "symbol": "USDC", "address": "0x", "decimals": 6},
				"minCheckoutUsd": 5
			}]
		}`))
	}))
	defer server.Close()

	client := bridge.NewClient(server.URL, nil)
	resp, err := client.GetSupportedAssets(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	first := resp.SupportedAssets[0]
	fmt.Printf("%s on %s, min=%.0f USD\n", first.Token.Symbol, first.ChainName, first.MinCheckoutUsd)
	// Output: USDC on Polygon, min=5 USD
}
```

- [ ] **Step 4: Verify build, vet, tests, and rendered doc**

```bash
go build ./pkg/bridge
go vet ./pkg/bridge
go test ./pkg/bridge
go doc ./pkg/bridge
go doc ./pkg/bridge.Client
go doc ./pkg/bridge.NewClient
go doc ./pkg/bridge.QuoteRequest
```

Expected: clean build, clean vet, tests green, package doc shows the new "client for the Polymarket Bridge API" summary.

- [ ] **Step 5: Commit**

```bash
git add pkg/bridge/bridge.go pkg/bridge/example_test.go
git commit -m "$(cat <<'EOF'
docs(godoc): full godoc and runnable Example for pkg/bridge

Tier-B package doc with use-cases and stability promise. Godoc on Client,
NewClient, every method, and every exported request/response type.
Example_supportedAssets added in a new pkg/bridge/example_test.go (the
package had no test file).

No new exports. No code logic changed. NewClient still accepts a
*transport.Client; the leak is pre-existing and documented as "pass nil
to use the default."

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2G: Full godoc + Example for `pkg/gamma`

**Files:**
- Modify: `pkg/gamma/client.go` (add package comment, godoc on every exported symbol)
- Create: `pkg/gamma/example_test.go`

**File allowlist (for `git add`):**
```
pkg/gamma/client.go
pkg/gamma/example_test.go
```

Exported surface to document (verified with `go doc ./pkg/gamma` and `go doc ./pkg/gamma.Client`):

- `func NewClient(baseURL string) *Client`
- `func DefaultConfig(baseURL string) transport.Config`
- `Client` with methods: `HealthCheck`, `ActiveMarkets`, `Markets`, `MarketByID`, `Events`, `EventByID`, `Series`, `Search`, `Tags`, `SportsMetadata`, `Comments`

> **Public-surface caveat:** Every method returns or takes `internal/polytypes` types. This is a pre-existing leak. Doc text must not reference `internal/polytypes`; instead say "the Polymarket protocol type for X". `DefaultConfig` returns `transport.Config` (an internal type) — same treatment. **Do not change signatures.**
>
> **Missing-but-needed export to log (not fix here):** `internal/gamma.Client` exposes `EventBySlug`, `EventsByTag`, `Profile`, and others not re-exported by `pkg/gamma`. Per spec ("Missing-but-needed exports get logged, not added"), append a row to `docs/AUDIT-FINDINGS.md` under `Track 1 — Code-side drift surfaced incidentally`:
>
> ```markdown
> | pkg/gamma re-exports only 11 of internal/gamma.Client's methods | pkg/gamma/client.go vs internal/gamma/client.go | Surfaced during Track 2; SDK consumers cannot reach EventBySlug, profile lookups, etc. through the public package. |
> ```

- [ ] **Step 1: Add the package doc comment**

The current file starts with `package gamma` and no comment. Insert above `package gamma`:

```go
// Package gamma is a read-only client for the Polymarket Gamma API
// surfaced for embedded use by downstream Go consumers.
//
// Use gamma when you need typed access to Polymarket markets, events,
// search, tags, series, sports metadata, or comments without pulling in
// the full polygolem CLI. The client performs no signing and is safe in
// read-only contexts.
//
// When not to use this package:
//   - For order book reads — use pkg/bookreader.
//   - For order placement or cancellation — Gamma does not host the
//     mutating CLOB surface.
//
// Stability: Client, NewClient, DefaultConfig, and every method on Client
// are part of the polygolem public SDK and follow semver. Method
// signatures currently expose protocol types defined inside polygolem;
// those types may be re-homed in a future release without changing the
// public method set.
package gamma
```

- [ ] **Step 2: Add godoc on the type and constructor**

Above `type Client struct`:

```go
// Client is the public read-only Gamma API client.
// Construct via NewClient. Methods are safe for concurrent use.
type Client struct {
```

Above `func NewClient`:

```go
// NewClient returns a Gamma client targeting baseURL.
// If baseURL is empty, the production Gamma URL is used. The client uses
// the package default HTTP transport with retry and rate limiting.
func NewClient(baseURL string) *Client {
```

Above `func DefaultConfig`:

```go
// DefaultConfig returns the transport config the Gamma client uses by
// default for baseURL — exposed for callers that want to inspect or
// extend the retry, timeout, and rate-limit defaults.
func DefaultConfig(baseURL string) transport.Config {
```

- [ ] **Step 3: Add godoc on every method on `*Client`**

Above each method, in this order:

```go
// HealthCheck pings the Gamma /health endpoint and returns the parsed
// response. Use this for readiness probes; it does not validate auth.
func (c *Client) HealthCheck(ctx context.Context) (*polytypes.HealthResponse, error) {

// ActiveMarkets returns markets currently flagged active by Gamma.
// Equivalent to Markets with the active filter set.
func (c *Client) ActiveMarkets(ctx context.Context) ([]polytypes.Market, error) {

// Markets lists markets matching the given filter parameters.
// Pass nil for default behavior (server-defined defaults).
func (c *Client) Markets(ctx context.Context, params *polytypes.GetMarketsParams) ([]polytypes.Market, error) {

// MarketByID fetches a single market by its Gamma ID.
func (c *Client) MarketByID(ctx context.Context, id string) (*polytypes.Market, error) {

// Events lists events matching the given filter parameters.
// Pass nil for default behavior.
func (c *Client) Events(ctx context.Context, params *polytypes.GetEventsParams) ([]polytypes.Event, error) {

// EventByID fetches a single event by its Gamma ID.
func (c *Client) EventByID(ctx context.Context, id string) (*polytypes.Event, error) {

// Series lists market series matching the given filter parameters.
// Pass nil for default behavior.
func (c *Client) Series(ctx context.Context, params *polytypes.GetSeriesParams) ([]polytypes.Series, error) {

// Search performs Gamma's public search across events, markets, tags, and
// profiles. Pass non-nil params; an empty Q returns server defaults.
func (c *Client) Search(ctx context.Context, params *polytypes.SearchParams) (*polytypes.SearchResponse, error) {

// Tags lists tags matching the given filter parameters.
// Pass nil for default behavior.
func (c *Client) Tags(ctx context.Context, params *polytypes.GetTagsParams) ([]polytypes.Tag, error) {

// SportsMetadata returns the current sports metadata catalog used by
// sports-event markets.
func (c *Client) SportsMetadata(ctx context.Context) ([]polytypes.SportMetadata, error) {

// Comments returns comments matching the given query — by parent entity
// or author, with optional pagination via params.
func (c *Client) Comments(ctx context.Context, params *polytypes.CommentQuery) ([]polytypes.Comment, error) {
```

- [ ] **Step 4: Create `pkg/gamma/example_test.go`**

Exact content:

```go
package gamma_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/TrebuchetDynamics/polygolem/pkg/gamma"
)

// Example_healthCheck demonstrates a Gamma readiness probe. A test HTTP
// server stands in for the production Gamma API so the example is
// hermetic and runs under `go test`.
func Example_healthCheck() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := gamma.NewClient(server.URL)
	resp, err := client.HealthCheck(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("status:", resp.Status)
	// Output: status: ok
}
```

> Note: this assumes `polytypes.HealthResponse` has a field named `Status` of string type — verify before writing the example. If the field is named differently, adjust the assertion line and the `// Output:` comment so they match. Run `go doc github.com/TrebuchetDynamics/polygolem/internal/polytypes.HealthResponse` to confirm. If the type does not unmarshal a single string field, choose `Example_activeMarkets` instead and have the test server return `[]` plus assert `len(markets) == 0`.

- [ ] **Step 5: Verify build, vet, tests, and rendered doc**

```bash
go build ./pkg/gamma
go vet ./pkg/gamma
go test ./pkg/gamma
go doc ./pkg/gamma
go doc ./pkg/gamma.Client
go doc ./pkg/gamma.NewClient
```

Expected: clean build, clean vet, tests green, package doc shows the new "read-only client for the Polymarket Gamma API" summary.

- [ ] **Step 6: Append the missing-export finding to `docs/AUDIT-FINDINGS.md`**

Append one row under `Track 1 — Code-side drift surfaced incidentally` (the row is the one in the caveat above the steps). Use a single Edit on `docs/AUDIT-FINDINGS.md`. Do not stage the file in this commit if Track 1 is the owner of that document — instead stage it here only if the surrounding workflow agrees. **Default action:** stage and commit it together with the godoc changes for transparency.

- [ ] **Step 7: Commit**

```bash
git add pkg/gamma/client.go pkg/gamma/example_test.go docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs(godoc): full godoc and runnable Example for pkg/gamma

Tier-B package doc with use-cases and stability promise. Godoc on Client,
NewClient, DefaultConfig, and every method. Example_healthCheck added in
a new pkg/gamma/example_test.go using httptest.

Logs the partial re-export of internal/gamma.Client in AUDIT-FINDINGS.md
under Track 1 code-side drift; per spec, missing-but-needed exports are
not added in this track.

No new exports. No code logic changed.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2H: Full godoc + Example for `pkg/marketresolver`

**Files:**
- Modify: `pkg/marketresolver/resolver.go` (replace package comment, add godoc on every exported symbol, add `Example_*`)

**File allowlist (for `git add`):**
```
pkg/marketresolver/resolver.go
```

> The package already has `resolver_test.go` with one `httptest`-backed test that hits both the `/public-search` and `/events` endpoints. Adding `Example_resolveTokenIDs` to a new `example_test.go` is fine, but a single example colocated in `resolver.go` is acceptable Go practice when the package is small. To minimize touched files, this task adds the example **inside the same file** the resolver lives in, immediately below the public type/method definitions but inside `_test.go` would still be cleaner — pick one consistent approach:
>
> **Decision (this plan):** put the `Example_resolveTokenIDs` in a new `example_test.go` to match Tasks 2E/2F/2G. Update the file allowlist below accordingly.

Revised file allowlist:
```
pkg/marketresolver/resolver.go
pkg/marketresolver/example_test.go
```

Exported surface to document (verified with `go doc ./pkg/marketresolver` and `go doc ./pkg/marketresolver.Resolver`):

- `Resolver` (struct) + `NewResolver(gammaBaseURL string) *Resolver`
- Methods: `ResolveCryptoMarkets`, `ResolveTokenIDs`, `ResolveTokenIDsAt`, `ValidateToken`
- Types: `CryptoMarket`, `MarketStatus` (string type), `ResolveResult`
- Constants: `StatusAvailable`, `StatusUnavailable`, `StatusStaleToken`, `StatusUnresolved`

The package has these existing one-line comments already on most symbols (`Resolver`, `CryptoMarket`, `ResolveResult`, `NewResolver`, `ResolveCryptoMarkets`, `ResolveTokenIDsAt`, `ResolveTokenIDs`, `ValidateToken`, `MarketStatus`). Edit them in place to be slightly fuller; do not delete existing prose.

- [ ] **Step 1: Replace the package comment**

Current opener:

```go
// Package marketresolver resolves active Polymarket markets and token IDs.
// Replaces go-bot's direct Gamma client and default token IDs per PRD Phase 0.
package marketresolver
```

Replace with:

```go
// Package marketresolver resolves Polymarket market identifiers — slug,
// asset, timeframe, or window-start time — into canonical token IDs.
//
// Use marketresolver when a downstream consumer (for example a trading
// bot) needs to convert a human-friendly identifier into the up/down
// token IDs and condition ID needed to place an order. The resolver
// performs only Gamma reads; it does not sign or mutate anything.
//
// When not to use this package:
//   - For full Gamma metadata access — use pkg/gamma directly.
//   - For order book pricing — use pkg/bookreader.
//
// Stability: Resolver, NewResolver, the four Resolve methods,
// ValidateToken, MarketStatus and its constants, ResolveResult, and
// CryptoMarket are part of the polygolem public SDK and follow semver.
package marketresolver
```

- [ ] **Step 2: Expand or add godoc on each exported symbol**

For each existing one-line comment, leave the first line and append a second sentence describing edge cases. Add comments where missing.

```go
// CryptoMarket represents a resolved crypto up/down market with token IDs.
// Slug and Question come from the Gamma event payload. UpTokenID and
// DownTokenID may be empty if the market's outcomes do not include both
// "up"/"yes" and "down"/"no".
type CryptoMarket struct { ... }

// Resolver finds active markets from the Gamma API.
// Methods are safe for concurrent use; each call is independent.
type Resolver struct { ... }

// NewResolver creates a market resolver targeting the given Gamma base URL.
// If gammaBaseURL is empty, the production Gamma URL is used.
func NewResolver(gammaBaseURL string) *Resolver { ... }

// ResolveCryptoMarkets finds active CLOB-enabled up/down markets for an asset.
// Returns only accepting, non-closed markets with valid token IDs.
// asset is matched case-insensitively; concurrent Gamma searches are
// fanned out per timeframe.
func (r *Resolver) ResolveCryptoMarkets(ctx context.Context, asset string) ([]CryptoMarket, error) { ... }

// ResolveTokenIDsAt resolves token IDs for a specific crypto window.
// Crypto up/down markets use deterministic slugs such as
// btc-updown-5m-1778114700, where the suffix is the UTC window start
// epoch. Falls back to ResolveTokenIDs when the slug lookup misses.
func (r *Resolver) ResolveTokenIDsAt(ctx context.Context, asset, timeframe string, windowStart time.Time) ResolveResult { ... }

// ResolveTokenIDs resolves token IDs for a given asset+timeframe.
// Returns StatusUnresolved if no active accepting market is found.
// Source records which Gamma path produced the result, useful for
// debugging stale-token issues.
func (r *Resolver) ResolveTokenIDs(ctx context.Context, asset, timeframe string) ResolveResult { ... }

// ValidateToken checks if a token ID is still valid by basic format checks.
// Returns StatusStaleToken if the CLOB returns an error for the token (a
// fuller validation requires CLOB access in the bookreader layer); for
// now it returns StatusUnresolved on empty/non-numeric token IDs and
// StatusAvailable otherwise.
func (r *Resolver) ValidateToken(ctx context.Context, tokenID string) MarketStatus { ... }

// MarketStatus classifies market availability returned by ResolveResult.
type MarketStatus string

// Market status values reported by ResolveResult.
const (
	// StatusAvailable means the resolver found an accepting non-closed
	// market with valid up/down token IDs.
	StatusAvailable   MarketStatus = "available"
	// StatusUnavailable means the resolver found a market but it is not
	// accepting orders (paused or closed).
	StatusUnavailable MarketStatus = "unavailable"
	// StatusStaleToken means a previously valid token ID can no longer
	// be priced; use ResolveTokenIDs again to discover the current one.
	StatusStaleToken  MarketStatus = "stale_token"
	// StatusUnresolved means no active matching market could be found.
	StatusUnresolved  MarketStatus = "unresolved"
)

// ResolveResult is the structured result of a market/token resolution.
// Source identifies which Gamma path produced the answer (deterministic
// slug, public search, or an error string).
type ResolveResult struct { ... }
```

The `Level`, `parseFloat` etc. helpers remain unexported; do not touch them.

- [ ] **Step 3: Create `pkg/marketresolver/example_test.go`**

The example must use httptest so it is hermetic. Modeled on the existing `TestResolveTokenIDsAtUsesDeterministicCryptoSlug`.

Exact content:

```go
package marketresolver_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/TrebuchetDynamics/polygolem/pkg/marketresolver"
)

// Example_resolveTokenIDsAt demonstrates resolving the up/down token IDs
// for a deterministic Polymarket crypto window slug. A test HTTP server
// stands in for the production Gamma API so the example is hermetic and
// runs under `go test`.
func Example_resolveTokenIDsAt() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events" && r.URL.Query().Get("slug") == "btc-updown-5m-1778114700" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{
				"id": "event-1",
				"slug": "btc-updown-5m-1778114700",
				"active": true,
				"closed": false,
				"markets": [{
					"id": "market-1",
					"conditionId": "condition-1",
					"slug": "btc-updown-5m-1778114700",
					"question": "Bitcoin Up or Down",
					"outcomes": ["Up", "Down"],
					"active": true,
					"closed": false,
					"enableOrderBook": true,
					"acceptingOrders": true,
					"clobTokenIds": "[\"up-token\", \"down-token\"]"
				}]
			}]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	resolver := marketresolver.NewResolver(server.URL)
	got := resolver.ResolveTokenIDsAt(context.Background(), "BTC", "5m", time.Unix(1778114700, 0).UTC())
	fmt.Printf("status=%s up=%s down=%s\n", got.Status, got.UpTokenID, got.DownTokenID)
	// Output: status=available up=up-token down=down-token
}
```

- [ ] **Step 4: Verify build, vet, tests, and rendered doc**

```bash
go build ./pkg/marketresolver
go vet ./pkg/marketresolver
go test ./pkg/marketresolver
go doc ./pkg/marketresolver
go doc ./pkg/marketresolver.Resolver
go doc ./pkg/marketresolver.MarketStatus
```

Expected: clean build, clean vet, tests green (existing four tests plus the new example), and package doc shows the new "resolves Polymarket market identifiers" summary.

- [ ] **Step 5: Commit**

```bash
git add pkg/marketresolver/resolver.go pkg/marketresolver/example_test.go
git commit -m "$(cat <<'EOF'
docs(godoc): full godoc and runnable Example for pkg/marketresolver

Tier-B package doc with use-cases and stability promise. Expands the
existing one-liners on every exported symbol — Resolver, NewResolver, the
four Resolve methods, ValidateToken, MarketStatus and its four constants,
ResolveResult, and CryptoMarket. Adds Example_resolveTokenIDsAt in a new
pkg/marketresolver/example_test.go using httptest.

No new exports. No code logic changed.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2I: Full godoc + Example for `pkg/pagination`

**Files:**
- Modify: `pkg/pagination/pagination.go` (replace package comment, expand godoc on every exported symbol)
- Modify: `pkg/pagination/pagination_test.go` (append `Example_*` to the existing test file — the package is generic and test file conventions match)

**File allowlist (for `git add`):**
```
pkg/pagination/pagination.go
pkg/pagination/pagination_test.go
```

Exported surface to document (verified with `go doc ./pkg/pagination`):

- `Page[T any]` — function type
- `OffsetPage[T any]` — function type
- `StreamResult[T any]` — generic struct (`Items`, `Err`)
- `StreamPages[T any]` — generic func returning `<-chan StreamResult[T]`
- `CollectAll[T any]` — generic func
- `CollectOffset[T any]` — generic func with `limit int`
- `Batch[T, R any]` — generic func parallelizing over batches

- [ ] **Step 1: Replace the package comment**

Current opener:

```go
// Package pagination provides auto-pagination helpers for cursor and offset-based APIs.
// Stolen from polymarket-go-sdk's StreamData and MarketsAll patterns.
package pagination
```

Replace with:

```go
// Package pagination provides generic helpers for paginating cursor-based
// and offset-based HTTP APIs and for parallelizing batch work.
//
// Use pagination when a Polymarket (or any) endpoint returns paged
// results and you want a tight loop around either the streamed pages or
// the fully collected slice. Helpers are generic over the page item type;
// callers supply the per-page fetch function.
//
// When not to use this package:
//   - For single-page fetches — call the underlying API directly.
//   - When concurrency is not desired and a simple for-loop is clearer.
//
// Stability: Page, OffsetPage, StreamResult, StreamPages, CollectAll,
// CollectOffset, and Batch are part of the polygolem public SDK and
// follow semver.
package pagination
```

- [ ] **Step 2: Expand the godoc on each exported symbol**

```go
// Page fetches one page of cursor-based data given a cursor.
// Returns the page's items, the next cursor (empty string ends iteration),
// and any error. The first call receives an empty cursor.
type Page[T any] func(ctx context.Context, cursor string) ([]T, string, error)

// StreamResult is a single page result emitted on the channel returned by
// StreamPages. Exactly one of Items or Err is meaningful per result.
type StreamResult[T any] struct {
	Items []T
	Err   error
}

// StreamPages iterates through all pages of a cursor-based endpoint.
// Calls pageFn until the next cursor is empty or an error occurs.
// Returns a channel that closes when iteration completes or ctx is
// cancelled. Errors are delivered as a final StreamResult before close.
func StreamPages[T any](ctx context.Context, pageFn Page[T]) <-chan StreamResult[T] { ... }

// CollectAll consumes a cursor-paged stream and returns all items.
// Returns the first error encountered; partial results are discarded.
func CollectAll[T any](ctx context.Context, pageFn Page[T]) ([]T, error) { ... }

// OffsetPage fetches one page of an offset-based API.
// Returns the page's items, the count returned (used to detect the last
// page), and any error.
type OffsetPage[T any] func(ctx context.Context, offset, limit int) ([]T, int, error)

// CollectOffset iterates through all pages of an offset-based endpoint.
// Stops when pageFn returns fewer than limit items. Returns the first
// error encountered; partial results are discarded.
func CollectOffset[T any](ctx context.Context, pageFn OffsetPage[T], limit int) ([]T, error) { ... }

// Batch splits items into chunks of at most maxBatchSize and runs fn on
// each batch concurrently. Returns the per-batch results in input order.
// Returns the first error encountered. fn may be invoked concurrently;
// callers are responsible for synchronizing any shared state.
func Batch[T, R any](ctx context.Context, items []T, maxBatchSize int, fn func(context.Context, []T) (R, error)) ([]R, error) { ... }
```

- [ ] **Step 3: Append `Example_collectAll` to `pkg/pagination/pagination_test.go`**

The existing file is `package pagination`. Appending an example in the same package keeps it close to the existing tests; `go test` will execute it. Use a deterministic in-memory page function so no httptest is needed.

Append to the bottom of `pkg/pagination/pagination_test.go`:

```go

// Example_collectAll demonstrates iterating through every page of a
// cursor-based source and collecting all items. The page function is
// in-memory so the example is hermetic.
func Example_collectAll() {
	pageFn := func(ctx context.Context, cursor string) ([]int, string, error) {
		switch cursor {
		case "":
			return []int{1, 2}, "p2", nil
		case "p2":
			return []int{3}, "", nil
		}
		return nil, "", nil
	}

	items, err := CollectAll(context.Background(), pageFn)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(items)
	// Output: [1 2 3]
}
```

> Note: `Example_collectAll` uses `fmt`, which is not currently imported by `pagination_test.go`. Add `"fmt"` to the existing import block in that file. Do not reorder or remove other imports.

- [ ] **Step 4: Verify build, vet, tests, and rendered doc**

```bash
go build ./pkg/pagination
go vet ./pkg/pagination
go test ./pkg/pagination
go doc ./pkg/pagination
go doc ./pkg/pagination.StreamPages
go doc ./pkg/pagination.Batch
```

Expected: clean build, clean vet, tests green (four existing tests plus `Example_collectAll`), and package doc shows the new "generic helpers for paginating cursor-based and offset-based HTTP APIs" summary.

- [ ] **Step 5: Commit**

```bash
git add pkg/pagination/pagination.go pkg/pagination/pagination_test.go
git commit -m "$(cat <<'EOF'
docs(godoc): full godoc and runnable Example for pkg/pagination

Tier-B package doc with use-cases and stability promise. Godoc on Page,
OffsetPage, StreamResult, StreamPages, CollectAll, CollectOffset, and
Batch. Appends Example_collectAll to pkg/pagination/pagination_test.go.

No new exports. No code logic changed.

Part of Track 2 (Godoc Layer) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2J: Final Track 2 verification gate

**Files:** none modified — read-only verification.

This task does not produce a commit. It either passes (Track 2 done) or
identifies regressions to loop back on.

- [ ] **Step 1: Every internal package has a `doc.go`**

```bash
missing=0
for p in auth cli clob config dataapi errors execution gamma marketdiscovery modes orders output paper polytypes preflight relayer risk rpc stream transport wallet; do
  if [[ ! -f "internal/$p/doc.go" ]]; then
    echo "MISSING: internal/$p/doc.go"
    missing=1
  fi
done
[[ $missing -eq 0 ]] && echo "All 21 doc.go files present"
```

Expected: `All 21 doc.go files present`.

- [ ] **Step 2: No package has two competing `// Package …` comments**

```bash
for p in auth cli clob config dataapi errors execution gamma marketdiscovery modes orders output paper polytypes preflight relayer risk rpc stream transport wallet; do
  count=$(grep -lE "^// Package $p\b" internal/$p/*.go 2>/dev/null | wc -l)
  if [[ "$count" -ne 1 ]]; then
    echo "internal/$p has $count files with a package comment (expected 1)"
  fi
done
```

Expected: no output.

- [ ] **Step 3: Every `pkg/*` has a package comment, full godoc, and at least one Example**

```bash
for p in bookreader bridge gamma marketresolver pagination; do
  echo "===== pkg/$p ====="
  go doc ./pkg/$p | head -3
done
echo "--- Example functions ---"
grep -r "^func Example" pkg/ --include='*.go'
```

Expected: each `go doc` shows a package summary in its first non-blank lines (not just `package $p` followed by symbols). The `grep` should show at least five `Example_*` lines, one per public package.

- [ ] **Step 4: Run the full spec-mandated verification**

```bash
go vet ./...
go test ./...
go doc ./pkg/bookreader
go doc ./pkg/bridge
go doc ./pkg/gamma
go doc ./pkg/marketresolver
go doc ./pkg/pagination
```

Expected: `go vet` clean, `go test` green, every `go doc` invocation prints package-level prose followed by a non-empty type/func listing.

- [ ] **Step 5: No `internal/...` references in the doc TEXT of the public packages**

Doc text only — signatures may still reference `internal/...` as a pre-existing API issue. The check below scans for `internal/` inside `//` comment lines in `pkg/`:

```bash
grep -rnE "^//.*internal/" pkg/
```

Expected: **no output**. If any line appears, fix the comment in place (the violation will be a `// ... internal/foo ...` doc string, not a Go import).

- [ ] **Step 6: Build still produces a working binary**

```bash
go build -o /tmp/polygolem-track2-verify ./cmd/polygolem
/tmp/polygolem-track2-verify --help >/dev/null && echo "binary OK"
rm -f /tmp/polygolem-track2-verify
```

Expected: `binary OK`. Track 2 is documentation-only and must not regress the binary.

- [ ] **Step 7: If all checks pass, mark Track 2 complete**

No file change required. Inform the user that Track 2 verification has passed. If any check fails, return to the relevant earlier task and fix in place rather than papering over.

---

## Out of scope (re-stated)

- Code refactoring of any kind. Touching function bodies, signatures, or dependencies is outside this track.
- Adding new exported APIs to any `pkg/*` package. Missing-but-needed exports surfaced (for example the partial `pkg/gamma` re-export of `internal/gamma`) are logged in `docs/AUDIT-FINDINGS.md` and not addressed here.
- Removing the pre-existing `internal/...` type leaks in public signatures (`pkg/bridge.NewClient` taking `*transport.Client`, `pkg/gamma.*` returning `polytypes.*` types, `pkg/gamma.DefaultConfig` returning `transport.Config`). Doc text avoids them; signatures stay as-is.
- `.revive.toml`, CI wiring, or `pkg.go.dev` publication scripts — deferred per spec.
- Track 1 / 3 / 4 / 5 work — separate plans.
