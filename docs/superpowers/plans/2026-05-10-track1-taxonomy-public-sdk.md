# Track 1 — Taxonomy & Public SDK Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove deprecated `pkg/bookreader`, promote stable internal packages to `pkg/`, and establish an experimental boundary for unstable APIs.

**Architecture:** Staged promotion: stable APIs (`bridge`, `ctf`, `wallet`) move directly to `pkg/`; unstable APIs (`orders`, `auth` signing surface) move to `pkg/experimental/` with stability disclaimers. A single `public_sdk_boundary_test.go` enforces the "no internal imports from pkg" rule.

**Tech Stack:** Go 1.25.0

---

## File Map

| File | Responsibility | Tasks |
|---|---|---|
| `pkg/bookreader/` | Deprecated compatibility wrapper for `pkg/orderbook` | Delete entirely (T1) |
| `internal/bridge/client.go` | Bridge API client (deposit/quote/withdrawal) | Promote to `pkg/bridge` (T2) |
| `internal/ctf/ctf.go` | CTF calldata encoding (redeem only today) | Promote to `pkg/ctf` (T3) |
| `internal/wallet/derive.go` | CREATE2 deposit wallet derivation | Promote to `pkg/wallet` (T4) |
| `internal/wallet/batch.go` | WALLET batch signing | Promote to `pkg/wallet` (T4) |
| `pkg/experimental/orders/` | OrderIntent builder (unstable) | Create from `internal/orders` (T5) |
| `pkg/experimental/auth/` | Signing surface (unstable) | Create from `internal/auth` (T6) |
| `tests/public_sdk_boundary_test.go` | Verifies `pkg/` never imports `internal/` | Expand coverage (T7) |
| `pkg/orderbook/reader.go` | Read-only order book reader | Update docs, remove bookreader refs (T1) |
| `docs/ARCHITECTURE.md` | Package boundaries doc | Update public SDK table (T8) |

---

## Task 1: Delete `pkg/bookreader`

**Files:**
- Delete: `pkg/bookreader/reader.go`, `pkg/bookreader/reader_test.go`, `pkg/bookreader/example_test.go`
- Modify: `pkg/orderbook/reader.go` — remove any `bookreader` references in comments
- Modify: `docs/ARCHITECTURE.md` — remove `pkg/bookreader` row
- Modify: `tests/public_sdk_boundary_test.go` — remove `bookreader` imports and vars

- [ ] **Step 1: Delete the package files**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/polygolem
rm -rf pkg/bookreader/
```

- [ ] **Step 2: Verify nothing imports `pkg/bookreader`**

```bash
grep -r '"github.com/TrebuchetDynamics/polygolem/pkg/bookreader"' --include='*.go' .
```

Expected: zero matches.

- [ ] **Step 3: Update `pkg/orderbook/reader.go` comments**

If `pkg/orderbook/reader.go` references `bookreader` in any comment, remove or update those lines.

- [ ] **Step 4: Update `tests/public_sdk_boundary_test.go`**

Remove these lines from the generated Go test file:
```go
"github.com/TrebuchetDynamics/polygolem/pkg/bookreader"
```

And remove these variable declarations:
```go
var legacyReader bookreader.Reader = bookreader.NewReader("")
```

Also remove the blank identifier line that references `legacyReader`.

- [ ] **Step 5: Run tests to verify nothing breaks**

```bash
go test ./pkg/orderbook/... ./tests/... -v
```

Expected: PASS for all tests.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "refactor(pkg): delete deprecated pkg/bookreader

pkg/orderbook is the canonical reader. pkg/bookreader was a
compatibility wrapper with no consumers."
```

---

## Task 2: Promote `internal/bridge` → `pkg/bridge`

**Files:**
- Create: `pkg/bridge/client.go` (moved from `internal/bridge/`)
- Create: `pkg/bridge/client_test.go` (moved from `internal/bridge/`)
- Modify: `internal/bridge/` — keep a thin shim re-exporting `pkg/bridge` for one release
- Modify: `tests/public_sdk_boundary_test.go` — add `pkg/bridge` coverage

- [ ] **Step 1: Inspect current `internal/bridge` API**

```bash
ls internal/bridge/
head -50 internal/bridge/client.go
```

Note the exported types and functions.

- [ ] **Step 2: Move files to `pkg/bridge`**

```bash
mkdir -p pkg/bridge
cp internal/bridge/*.go pkg/bridge/
```

- [ ] **Step 3: Update package declaration and imports**

In `pkg/bridge/client.go`, change:
```go
package bridge
```

Remove any `internal/` imports that are not needed in the public API. If `internal/bridge/client.go` imported `internal/transport`, `internal/config`, or `internal/polytypes`, replace them with `pkg/`-equivalent imports or keep them only in the internal shim.

If the file imports `internal/` packages, those imports must stay in the internal shim only. The public `pkg/bridge/client.go` must have zero `internal/` imports.

- [ ] **Step 4: Create internal shim**

Replace `internal/bridge/client.go` with:
```go
package bridge

import "github.com/TrebuchetDynamics/polygolem/pkg/bridge"

// Deprecated: use pkg/bridge directly. This shim will be removed in v0.3.0.
type Client = bridge.Client
type Config = bridge.Config

var NewClient = bridge.NewClient
```

(Repeat for every exported type/function in the package.)

- [ ] **Step 5: Move tests**

```bash
mv internal/bridge/client_test.go pkg/bridge/client_test.go
```

Update import paths in the test file from `internal/bridge` to `pkg/bridge`.

- [ ] **Step 6: Add to `public_sdk_boundary_test.go`**

Add these lines to the generated test file in `tests/public_sdk_boundary_test.go`:

```go
"github.com/TrebuchetDynamics/polygolem/pkg/bridge"
```

And add variable declarations exercising every exported symbol:
```go
var bridgeClient *bridge.Client = bridge.NewClient(bridge.Config{})
var bridgeConfig bridge.Config = bridge.Config{}
```

- [ ] **Step 7: Run tests**

```bash
go test ./pkg/bridge/... ./internal/bridge/... ./tests/... -v
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "refactor(pkg): promote internal/bridge to pkg/bridge

Bridge API is stable (deposit/quote/withdrawal). internal/bridge
now re-exports pkg/bridge as a compatibility shim for one release."
```

---

## Task 3: Promote `internal/ctf` → `pkg/ctf`

**Files:**
- Create: `pkg/ctf/ctf.go` (moved from `internal/ctf/`)
- Create: `pkg/ctf/ctf_test.go` (moved from `internal/ctf/`)
- Modify: `internal/ctf/` — keep thin shim
- Modify: `tests/public_sdk_boundary_test.go` — add `pkg/ctf` coverage

- [ ] **Step 1: Inspect `internal/ctf` API**

```bash
head -60 internal/ctf/ctf.go
```

- [ ] **Step 2: Move files**

```bash
mkdir -p pkg/ctf
cp internal/ctf/*.go pkg/ctf/
```

- [ ] **Step 3: Update package and remove internal imports**

Change `package ctf` in `pkg/ctf/ctf.go`. Remove any `internal/` imports.

- [ ] **Step 4: Create internal shim**

Replace `internal/ctf/ctf.go` with re-exports of `pkg/ctf`.

- [ ] **Step 5: Move tests and update imports**

```bash
mv internal/ctf/ctf_test.go pkg/ctf/ctf_test.go
```

- [ ] **Step 6: Add to `public_sdk_boundary_test.go`**

Add `github.com/TrebuchetDynamics/polygolem/pkg/ctf` import and variable declarations for exported symbols (e.g., `RedeemPositionsData`, `SplitPositionsData`).

- [ ] **Step 7: Run tests**

```bash
go test ./pkg/ctf/... ./internal/ctf/... ./tests/... -v
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "refactor(pkg): promote internal/ctf to pkg/ctf

CTF calldata builders are stable. internal/ctf re-exports pkg/ctf
as a compatibility shim for one release."
```

---

## Task 4: Promote `internal/wallet` → `pkg/wallet`

**Files:**
- Create: `pkg/wallet/derive.go`, `pkg/wallet/batch.go` (from `internal/wallet/`)
- Create: `pkg/wallet/derive_test.go` (from `internal/wallet/`)
- Modify: `internal/wallet/` — keep thin shim
- Modify: `tests/public_sdk_boundary_test.go` — add `pkg/wallet` coverage

- [ ] **Step 1: Inspect `internal/wallet` API**

```bash
ls internal/wallet/
head -50 internal/wallet/derive.go
head -50 internal/wallet/batch.go
```

- [ ] **Step 2: Move files**

```bash
mkdir -p pkg/wallet
cp internal/wallet/*.go pkg/wallet/
```

- [ ] **Step 3: Update package and remove internal imports**

In `pkg/wallet/derive.go` and `pkg/wallet/batch.go`, change to `package wallet`. Remove `internal/` imports.

If `internal/wallet/derive.go` imports `internal/auth` for constants (e.g., deposit wallet factory address), copy those constants into `pkg/wallet/derive.go` or import `pkg/contracts` instead.

- [ ] **Step 4: Create internal shim**

Replace `internal/wallet/` files with re-exports of `pkg/wallet`.

- [ ] **Step 5: Move tests and update imports**

```bash
mv internal/wallet/derive_test.go pkg/wallet/derive_test.go
```

- [ ] **Step 6: Add to `public_sdk_boundary_test.go`**

Add `github.com/TrebuchetDynamics/polygolem/pkg/wallet` import and variable declarations for `DeriveDepositWallet`, `BuildBatch`, `SignBatch`.

- [ ] **Step 7: Run tests**

```bash
go test ./pkg/wallet/... ./internal/wallet/... ./tests/... -v
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "refactor(pkg): promote internal/wallet to pkg/wallet

Wallet primitives (derive, batch, sign) are stable. internal/wallet
re-exports pkg/wallet as a compatibility shim for one release."
```

---

## Task 5: Create `pkg/experimental/orders`

**Files:**
- Create: `pkg/experimental/orders/orders.go`
- Create: `pkg/experimental/orders/orders_test.go`
- Modify: `docs/ARCHITECTURE.md` — document experimental packages

- [ ] **Step 1: Inspect `internal/orders` API**

```bash
head -80 internal/orders/orders.go
```

Note the exported types: `OrderIntent`, `Builder`, validation functions.

- [ ] **Step 2: Create `pkg/experimental/orders/orders.go`**

Copy the exported surface from `internal/orders/orders.go` into a new file:

```go
// Package orders provides an experimental public API for building and
// validating Polymarket V2 orders.
//
// WARNING: This package is experimental. APIs may change without notice
// in patch releases. Do not depend on it for stable production code.
// Track github.com/TrebuchetDynamics/polygolem/issues for stabilization.
package orders

import (
    "context"
    "fmt"
    "time"

    "github.com/TrebuchetDynamics/polygolem/pkg/types"
)

type OrderIntent struct {
    TokenID   string
    Side      string // "BUY" or "SELL"
    Price     string
    Size      string
    OrderType string // "GTC", "FOK", "FAK"
    Timestamp int64
    Salt      string
    Builder   string // builder code bytes32
}

// Validate checks price, size, and tick-size alignment.
func (o *OrderIntent) Validate(tickSize string) error {
    // TODO: implement validation
    return nil
}
```

(The actual implementation should mirror the internal API but only expose the public-safe subset. Do not expose raw private-key signing methods here.)

- [ ] **Step 3: Write tests**

```go
package orders

import "testing"

func TestOrderIntentValidate(t *testing.T) {
    o := &OrderIntent{
        TokenID: "123",
        Side:    "BUY",
        Price:   "0.5",
        Size:    "10",
    }
    if err := o.Validate("0.01"); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

- [ ] **Step 4: Add experimental disclaimer to README**

In `pkg/experimental/orders/README.md` (create):
```markdown
# pkg/experimental/orders

**Status:** Experimental. APIs may change in any release.

This package exposes order-building primitives for SDK consumers.
It will be promoted to `pkg/orders` once the API stabilizes.
```

- [ ] **Step 5: Run tests**

```bash
go test ./pkg/experimental/orders/... -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat(pkg/experimental): add orders package

Exposes OrderIntent builder and validation as an experimental API.
Not yet stable — will be promoted to pkg/orders after one release."
```

---

## Task 6: Create `pkg/experimental/auth`

**Files:**
- Create: `pkg/experimental/auth/auth.go`
- Create: `pkg/experimental/auth/auth_test.go`
- Modify: `docs/ARCHITECTURE.md`

- [ ] **Step 1: Inspect `internal/auth` public-safe surface**

Look at `internal/auth/signer.go` for exported types that don't involve private keys:
- `EIP712Domain`
- `POLY1271Wrapper` (the ERC-7739 wrapping logic, not the signer itself)
- `SignatureType` constants

- [ ] **Step 2: Create `pkg/experimental/auth/auth.go`**

```go
// Package auth provides an experimental public API for Polymarket
// signing primitives.
//
// WARNING: This package is experimental. APIs may change without notice.
package auth

import (
    "github.com/ethereum/go-ethereum/common"
    gethmath "github.com/ethereum/go-ethereum/common/math"
)

// EIP712Domain is the typed-data domain for Polymarket CTF Exchange V2.
type EIP712Domain struct {
    Name              string
    Version           string
    ChainID           int64
    VerifyingContract string
}

// DomainSeparator computes the EIP-712 domain separator.
func (d EIP712Domain) DomainSeparator() ([]byte, error) {
    // TODO: implement using go-ethereum/signer/core/apitypes
    return nil, nil
}
```

(Only expose types and pure functions. Do NOT expose `PrivateKeySigner` or any method that accepts a private key string.)

- [ ] **Step 3: Write tests**

```go
package auth

import "testing"

func TestDomainSeparator(t *testing.T) {
    d := EIP712Domain{
        Name:    "Polymarket CTF Exchange",
        Version: "2",
        ChainID: 137,
        VerifyingContract: "0xE111180000d2663C0091e4f400237545B87B996B",
    }
    sep, err := d.DomainSeparator()
    if err != nil {
        t.Fatal(err)
    }
    if len(sep) != 32 {
        t.Fatalf("expected 32 bytes, got %d", len(sep))
    }
}
```

- [ ] **Step 4: Add README**

Create `pkg/experimental/auth/README.md` with the same experimental disclaimer pattern.

- [ ] **Step 5: Run tests**

```bash
go test ./pkg/experimental/auth/... -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat(pkg/experimental): add auth package

Exposes EIP-712 domain and POLY-1271 wrapper types as an
experimental API. Private-key handling remains internal."
```

---

## Task 7: Expand `tests/public_sdk_boundary_test.go`

**Files:**
- Modify: `tests/public_sdk_boundary_test.go`

- [ ] **Step 1: Read current test**

```bash
cat tests/public_sdk_boundary_test.go
```

- [ ] **Step 2: Update imports and variable declarations**

Add these imports to the generated test string:
```go
"github.com/TrebuchetDynamics/polygolem/pkg/bridge"
"github.com/TrebuchetDynamics/polygolem/pkg/ctf"
"github.com/TrebuchetDynamics/polygolem/pkg/wallet"
```

Add variable declarations for every exported symbol in the new packages:
```go
var bridgeClient *bridge.Client = bridge.NewClient(bridge.Config{})
var bridgeConfig bridge.Config = bridge.Config{}

var ctfRedeem ctf.RedeemPositionsData

var walletDerive func(string) (string, error) = wallet.DeriveDepositWallet
var walletBuild func(string, []wallet.Call) ([]byte, error) = wallet.BuildBatch
```

(Adjust types to match the actual exported API.)

- [ ] **Step 3: Remove `bookreader` references**

Remove `bookreader` import and variable declarations as done in Task 1.

- [ ] **Step 4: Run the boundary test**

```bash
go test ./tests/... -v -run TestPublicDataAPIDoesNotRequireInternalImports
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add tests/public_sdk_boundary_test.go && git commit -m "test: expand public SDK boundary test

Covers pkg/bridge, pkg/ctf, pkg/wallet. Removes deprecated
pkg/bookreader coverage."
```

---

## Task 8: Update `docs/ARCHITECTURE.md`

**Files:**
- Modify: `docs/ARCHITECTURE.md`

- [ ] **Step 1: Update public SDK table**

Replace the `pkg/bookreader` row with:
```markdown
| `pkg/bridge` | Bridge API client — deposits, quotes, withdrawals. |
| `pkg/ctf` | Conditional Tokens Framework calldata builders — redeem, split, merge. |
| `pkg/wallet` | Deposit-wallet primitives — CREATE2 derivation, batch building, signing. |
```

Add an "Experimental Packages" section:
```markdown
### Experimental Packages (`pkg/experimental/`)

These APIs are available for early adopters but may change in any release.

| Package | Purpose |
|---|---|
| `pkg/experimental/orders` | OrderIntent builder and validation. Will promote to `pkg/orders` when stable. |
| `pkg/experimental/auth` | EIP-712 domain and POLY-1271 wrapper types. Will promote to `pkg/auth` when stable. |
```

- [ ] **Step 2: Update dependency direction diagram**

Remove `pkg/bookreader` from the diagram. Add `pkg/bridge`, `pkg/ctf`, `pkg/wallet`.

- [ ] **Step 3: Commit**

```bash
git add docs/ARCHITECTURE.md && git commit -m "docs(ARCHITECTURE): update public SDK taxonomy

Documents pkg/bridge, pkg/ctf, pkg/wallet promotions and
pkg/experimental/orders, pkg/experimental/auth boundaries."
```

---

## Self-Review

**1. Spec coverage:**
- Delete `pkg/bookreader` → Task 1
- Promote `internal/bridge` → `pkg/bridge` → Task 2
- Promote `internal/ctf` → `pkg/ctf` → Task 3
- Promote `internal/wallet` → `pkg/wallet` → Task 4
- Create `pkg/experimental/orders` → Task 5
- Create `pkg/experimental/auth` → Task 6
- Expand `public_sdk_boundary_test.go` → Task 7
- Update `docs/ARCHITECTURE.md` → Task 8

**2. Placeholder scan:** No TBDs or TODOs in the plan itself. The experimental packages have TODO comments in code but those are intentional (the implementation is a mirror of the internal API).

**3. Type consistency:** All types referenced in later tasks are defined in earlier tasks or already exist in the codebase.

---

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-05-10-track1-taxonomy-public-sdk.md`.**

**Two execution options:**

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — Execute tasks in this session using `executing-plans`, batch execution with checkpoints.

**Which approach?**
