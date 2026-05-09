# Polygolem Live Safety + V2 Settlement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close two live-money blockers found while debugging the 2026-05-09 trading session against deposit wallet `0x21999a074344610057c9b2B362332388a44502D4`:

1. The market resolver can return a **wrong-window** market when its deterministic slug lookup misses or its event payload is timeshifted. Live evidence: SOL signal bar 2026-05-09 08:20 UTC bought a market that starts 12:20 UTC; ETH signal bars 07:40/08:00 UTC bought a market starting 08:40 UTC. Orders matched, but on the wrong market window. Must fail-closed before adding more funds.

2. There is no V2-correct redeem path. `pkg/ctf` encodes the legacy `redeemPositions` selector but the **V2 deposit wallet must call the V2 collateral adapter**, not raw ConditionalTokens. The current 6-call WALLET-CREATE approval batch covers trading spenders but **does not approve the V2 collateral adapters**, so today's wallet cannot redeem even after market resolution.

This plan is polygolem-scope only. go-bot consumes via SDK.

**Architecture:**
- `pkg/marketresolver` gains a `ResolveTokenIDsForWindow` strict method and a `StatusWindowMismatch` status. Existing `ResolveTokenIDsAt` becomes fail-closed when the slug-hit market's `StartDate` does not match the requested window.
- `pkg/types.Position` and `internal/dataapi.Position` gain `Redeemable`, `Mergeable`, `NegativeRisk`, plus the resolution-relevant fields the official Data API schema documents.
- `pkg/contracts.Registry` gains `CtfCollateralAdapter`, `NegRiskCtfCollateralAdapter`, `CollateralOnramp`, `CollateralOfframp`, `PermissionedRamp`, plus a `RedeemAdapterFor(negRisk bool) string` helper.
- `internal/relayer/approvals.go` gains `BuildAdapterApprovalCalls()` (4 calls: pUSD `approve` + CTF `setApprovalForAll` for both V2 adapters) and `OnboardDepositWallet` includes them in the post-deploy batch so new wallets are redeem-ready out of the box. Existing live wallets need a one-shot `polygolem deposit-wallet approve-adapters` migration.
- New `pkg/settlement` package: `FindRedeemable`, `BuildRedeemCall`, `SubmitRedeem`. Calldata reuses `pkg/ctf.RedeemPositionsData`; only the call target switches between `CtfCollateralAdapter` and `NegRiskCtfCollateralAdapter` based on `Position.NegativeRisk`.
- New CLI commands `polygolem deposit-wallet approve-adapters --submit --confirm APPROVE_ADAPTERS`, `... redeemable`, `... redeem [--dry-run] [--limit N] [--submit --confirm REDEEM_WINNERS]`.

**Tech Stack:** Go 1.22+, `github.com/ethereum/go-ethereum/{common,accounts/abi,crypto}`, `httptest` server mocks, no live RPC required for tests.

**Spec:** This file. Cross-references `docs/superpowers/specs/2026-05-07-clob-v2-conformance-design.md` for V2 typed-data context.

**Source-of-truth references:**
- `opensource-projects/repos/ctf-exchange-v2/src/adapters/CtfCollateralAdapter.sol` — V2 redeem logic.
- `opensource-projects/repos/ctf-exchange-v2/src/adapters/NegRiskCtfCollateralAdapter.sol` — neg-risk variant.
- `opensource-projects/repos/ctf-exchange-v2/src/collateral/CollateralToken.sol` — pUSD definition (`name="Polymarket USD"`, `symbol="pUSD"`, 6 decimals).
- `opensource-projects/repos/ctf-exchange-v2/README.md` — Polygon mainnet addresses for all V2 contracts.
- Polymarket Data API GET /positions schema: `redeemable: boolean`, `mergeable: boolean`, `negativeRisk: boolean` (https://docs.polymarket.com/api-reference/core/get-current-positions-for-a-user.md).
- Polymarket V2 redeem semantics (official): *"Polymarket uses thin collateral adapter contracts for pUSD-native CTF actions. Approve the adapter once, then route split, merge, and redeem actions through it."*

**Tradeoffs / known-unknowns flagged for review during implementation:**

- **Approval batch backwards compatibility.** Adding 4 calls to `OnboardDepositWallet` is fine for new wallets but cannot retroactively flow into existing wallets — they must run `approve-adapters` once. Reviewers must confirm this is acceptable rather than requiring a forced migration path.
- **WALLET batch size limit.** The hard cap is whatever the relayer + Polygon block gas accept. We default to 10 calls per redeem batch. The empirical ceiling needs measuring (Task 7 step) — flag if a real run exceeds it and fall back to chunking.
- **`indexSets` for binary markets.** `CtfCollateralAdapter._redeemPositions` ignores the caller's `indexSets` array and uses `CTFHelpers.partition()` (= `[1, 2]`) internally. We pass `[]uint256{}` (empty) to keep calldata minimal; the adapter's signature matches the legacy CT signature for source-compatibility.
- **NegRisk question count.** `NegRiskCtfCollateralAdapter._redeemPositions` reads CTF balances internally and redeems whatever the wallet holds. No off-chain question-count discovery needed for redeem — just the `conditionId`. Confirmed in `NegRiskCtfCollateralAdapter.sol:155-163`.

---

## File map

| File | Responsibility | Tasks |
|---|---|---|
| `pkg/marketresolver/resolver.go` | window-guard fields, strict resolver method, status enum addition | T2 |
| `pkg/marketresolver/resolver_test.go` | unit tests for window guard + slug-hit verification | T2 |
| `internal/dataapi/client.go` | extend `Position` decoding with V2 fields | T3 |
| `pkg/types/data.go` | mirror new fields on the public DTO | T3 |
| `pkg/data/client.go`, `pkg/universal/client.go` | pass-through (no logic change) | T3 |
| `pkg/contracts/contracts.go` | adapter + ramp constants, `RedeemAdapterFor` helper | T4 |
| `pkg/contracts/contracts_test.go` | registry assertions | T4 |
| `internal/relayer/approvals.go` | `BuildAdapterApprovalCalls` + adapter spender constants | T5 |
| `pkg/relayer/onboard.go` | bake adapter approvals into `OnboardDepositWallet` for new wallets | T5 |
| `internal/cli/deposit_wallet.go` | `approve-adapters`, `redeemable`, `redeem` subcommands | T5, T7 |
| `internal/cli/deposit_wallet_test.go` | CLI tests | T5, T7 |
| `pkg/settlement/settlement.go` | `FindRedeemable`, `BuildRedeemCall`, `SubmitRedeem` | T6 |
| `pkg/settlement/settlement_test.go` | unit tests with httptest Data API + relayer mocks | T6 |
| `tests/public_sdk_boundary_test.go` | re-pin signatures for new SDK surface | T2, T3, T4, T5, T6 |
| `tests/repository_hygiene_test.go` | require `pkg/settlement` exists | T6 |
| `docs/CONTRACTS.md`, `docs/SAFETY.md`, `docs/COMMANDS.md`, `BLOCKERS.md`, `CHANGELOG.md` | text + regen | T8 |
| `docs-site/src/content/docs/docs/concepts/{contracts,deposit-wallets,safety}.mdx` | Starlight mirror | T8 |
| `docs-site/src/content/docs/docs/guides/{deposit-wallet-lifecycle,redeem-winners}.mdx` | new redeem guide + lifecycle update | T8 |
| `docs-site/src/content/docs/docs/reference/sdk.mdx` | `pkg/settlement` SDK reference | T8 |

---

## Task 1: Verify research preconditions are still true

The plan was built on snapshots of upstream contracts and docs. Confirm nothing has rotated before starting.

**Files:**
- Read: `opensource-projects/repos/ctf-exchange-v2/README.md`, `opensource-projects/repos/ctf-exchange-v2/src/adapters/CtfCollateralAdapter.sol`

- [ ] **Step 1: Confirm V2 adapter addresses match the plan**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
grep -E "CtfCollateralAdapter\b.*0x|NegRiskCtfCollateralAdapter\b.*0x" \
  opensource-projects/repos/ctf-exchange-v2/README.md
```
Expected: lines containing
```
[CtfCollateralAdapter]...0xAdA100Db00Ca00073811820692005400218FcE1f
[NegRiskCtfCollateralAdapter]...0xadA2005600Dec949baf300f4C6120000bDB6eAab
```

- [ ] **Step 2: Confirm `redeemPositions` signature on the adapter**

```bash
sed -n '114,140p' opensource-projects/repos/ctf-exchange-v2/src/adapters/CtfCollateralAdapter.sol
```
Expected: function declaration `redeemPositions(address, bytes32, bytes32 _conditionId, uint256[]) external onlyUnpaused(USDCE)`. The collateral address, parent collection, and `uint256[]` args are accepted but ignored; the implementation reads `_conditionId`.

- [ ] **Step 3: Confirm Data API position schema**

```bash
# Quick re-fetch (no auth required)
curl -fsSL "https://data-api.polymarket.com/positions?user=0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C&limit=1" \
  | python3 -c "import json,sys;rows=json.load(sys.stdin);print(json.dumps(list(rows[0].keys()) if rows else [], indent=2))"
```
Expected: a key list that contains `redeemable`, `mergeable`, `negativeRisk`, `outcome`, `outcomeIndex`, `oppositeOutcome`, `oppositeAsset`, `endDate`. If the live account currently has no positions, substitute any whale address from a recent winning market.

- [ ] **Step 4: Confirm current approval batch does NOT include the V2 adapters**

```bash
grep -n "0xAdA100Db00Ca00073811820692005400218FcE1f\|0xadA2005600Dec949baf300f4C6120000bDB6eAab" \
  internal/relayer/approvals.go
```
Expected: **no matches**. (If a match appears, someone landed Task 5 already; skim the diff and restart from the next missing task.)

- [ ] **Step 5: No commit for this task** — it is a verification gate.

---

## Task 2: Market resolver decision-window guard

**The bug.** `pkg/marketresolver.Resolver.ResolveTokenIDsAt(asset, timeframe, windowStart)` builds a deterministic slug `<asset>-updown-<tf>-<unix>` and on slug hit returns the first accepting market with matching timeframe. It does **not** verify that the matched market's `StartDate` equals `windowStart`. On slug miss it falls through to `ResolveTokenIDs` which returns *any* accepting market — silently selecting a future or stale window.

**The fix.** Carry `StartDate` and `EndDate` through `CryptoMarket` and `ResolveResult`. Add `StatusWindowMismatch` to distinguish "no market" from "wrong-window market" (the dangerous case). Add a strict `ResolveTokenIDsForWindow` method that never falls through. `ResolveTokenIDsAt` now fails-closed on a slug-hit window mismatch instead of returning the wrong market.

**Files:**
- Modify: `pkg/marketresolver/resolver.go`, `pkg/marketresolver/resolver_test.go`
- Modify: `tests/public_sdk_boundary_test.go`

- [ ] **Step 1: Add fields to `CryptoMarket` and `ResolveResult`**

In `pkg/marketresolver/resolver.go`:

```go
// CryptoMarket struct:
//   ...
//   Closed      bool
//   Question    string
//   Slug        string
//   StartDate   time.Time   // NEW — Gamma market.startDate
//   EndDate     time.Time   // NEW — Gamma market.endDate

// ResolveResult struct:
//   ...
//   Source      string       `json:"source"`
//   StartDate   time.Time    `json:"start_date,omitempty"` // NEW
//   EndDate     time.Time    `json:"end_date,omitempty"`   // NEW
```

Populate them in `marketsFromGamma` from `m.StartDate.Time()` and `m.EndDate.Time()` (the existing `polytypes.Market` has `StartDate`/`EndDate` as `NormalizedTime`; expose `.Time()` if not already present — verify with `grep -n "func.*NormalizedTime.*Time" pkg/types/normalized*.go internal/polytypes/*.go`).

`firstAcceptingMarket` carries them through into the returned `ResolveResult`.

- [ ] **Step 2: Add the new status value**

```go
const (
    StatusAvailable      MarketStatus = "available"
    StatusUnavailable    MarketStatus = "unavailable"
    StatusStaleToken     MarketStatus = "stale_token"
    StatusUnresolved     MarketStatus = "unresolved"
    StatusWindowMismatch MarketStatus = "window_mismatch" // NEW
)
```

- [ ] **Step 3: Make `ResolveTokenIDsAt` fail-closed on slug-hit window mismatch**

```go
func (r *Resolver) ResolveTokenIDsAt(ctx context.Context, asset, timeframe string, windowStart time.Time) ResolveResult {
    if slug := cryptoWindowSlug(asset, timeframe, windowStart); slug != "" {
        if evt, err := r.gamma.EventBySlug(ctx, slug); err == nil {
            if result, ok := firstAcceptingMarket(asset, timeframe, marketsFromGamma(asset, evt.Markets)); ok {
                if !windowStart.IsZero() && !result.StartDate.Equal(windowStart.UTC().Truncate(time.Second)) {
                    return ResolveResult{
                        Status:    StatusWindowMismatch,
                        Asset:     asset,
                        Timeframe: timeframe,
                        Source: fmt.Sprintf("gamma:slug_hit_window_mismatch:%s:got=%s want=%s",
                            slug, result.StartDate.UTC().Format(time.RFC3339),
                            windowStart.UTC().Format(time.RFC3339)),
                        StartDate: result.StartDate,
                        EndDate:   result.EndDate,
                    }
                }
                result.Source = "gamma:event_slug:" + slug
                return result
            }
        }
    }
    return r.ResolveTokenIDs(ctx, asset, timeframe)
}
```

- [ ] **Step 4: Add the strict method**

```go
// ResolveTokenIDsForWindow returns StatusAvailable only when the matched
// market's startDate exactly equals windowStart. Returns StatusUnresolved
// on slug miss or fallback-search miss. Returns StatusWindowMismatch when
// a slug hit is found but its startDate disagrees with windowStart. Never
// falls through to an unanchored search.
func (r *Resolver) ResolveTokenIDsForWindow(ctx context.Context, asset, timeframe string, windowStart time.Time) ResolveResult {
    if windowStart.IsZero() {
        return ResolveResult{Status: StatusUnresolved, Asset: asset, Timeframe: timeframe, Source: "windowStart_zero"}
    }
    slug := cryptoWindowSlug(asset, timeframe, windowStart)
    if slug == "" {
        return ResolveResult{Status: StatusUnresolved, Asset: asset, Timeframe: timeframe, Source: "no_slug_for_asset_timeframe"}
    }
    evt, err := r.gamma.EventBySlug(ctx, slug)
    if err != nil {
        return ResolveResult{Status: StatusUnresolved, Asset: asset, Timeframe: timeframe, Source: fmt.Sprintf("gamma:slug_miss:%s:%v", slug, err)}
    }
    result, ok := firstAcceptingMarket(asset, timeframe, marketsFromGamma(asset, evt.Markets))
    if !ok {
        return ResolveResult{Status: StatusUnresolved, Asset: asset, Timeframe: timeframe, Source: "gamma:slug_event_no_accepting_market:" + slug}
    }
    if !result.StartDate.Equal(windowStart.UTC().Truncate(time.Second)) {
        return ResolveResult{
            Status:    StatusWindowMismatch,
            Asset:     asset,
            Timeframe: timeframe,
            StartDate: result.StartDate,
            EndDate:   result.EndDate,
            Source: fmt.Sprintf("gamma:slug_hit_window_mismatch:%s:got=%s want=%s",
                slug, result.StartDate.UTC().Format(time.RFC3339),
                windowStart.UTC().Format(time.RFC3339)),
        }
    }
    result.Source = "gamma:event_slug_strict:" + slug
    return result
}
```

- [ ] **Step 5: Update package doc**

In the package doc block at the top of `resolver.go`, add a paragraph:

> Decision-window safety: prefer `ResolveTokenIDsForWindow` when the caller has
> a binding window start (the typical live-trading case). It returns
> `StatusWindowMismatch` rather than silently substituting a different window.
> `ResolveTokenIDsAt` and `ResolveTokenIDs` are best-effort and may return any
> currently-accepting market; do not use them on the order-placement path.

- [ ] **Step 6: Add tests**

In `pkg/marketresolver/resolver_test.go`, add four cases backed by an `httptest.Server` that mocks `gamma.polymarket.com`:

1. `TestResolveTokenIDsForWindow_HappyPath` — slug hits, market `startDate == windowStart` → returns `StatusAvailable` with the matched up/down token IDs.
2. `TestResolveTokenIDsForWindow_RejectsWrongWindow` — slug hits, market `startDate = windowStart + 5*time.Minute` → returns `StatusWindowMismatch`. Source string contains `slug_hit_window_mismatch`.
3. `TestResolveTokenIDsForWindow_NeverFallsThrough` — slug 404 → returns `StatusUnresolved`. Asserts the resolver did **not** call the Gamma `/events?...` search endpoint (mock counts requests).
4. `TestResolveTokenIDsAt_FailsClosedOnSlugHitMismatch` — exact reproducer of the SOL 08:20-vs-12:20 trap; returns `StatusWindowMismatch`.

- [ ] **Step 7: Re-pin SDK boundary**

```bash
grep -n "marketresolver\." tests/public_sdk_boundary_test.go
```
If the test pins specific symbol names, add `ResolveTokenIDsForWindow` and `StatusWindowMismatch`. Re-run `go test ./tests/...` and confirm green.

- [ ] **Step 8: Verify**

```bash
go test ./pkg/marketresolver/... ./tests/... -count=1
```
Expected: all green.

- [ ] **Step 9: Commit**

```bash
git add pkg/marketresolver/resolver.go pkg/marketresolver/resolver_test.go tests/public_sdk_boundary_test.go
git commit -m "fix(marketresolver): fail-closed decision-window guard

ResolveTokenIDsAt silently returned wrong-window markets when the slug
hit a market whose startDate did not match windowStart. Live evidence:
SOL 08:20 UTC signal bought a market starting 12:20 UTC; ETH 07:40/08:00
signals bought a market starting 08:40 UTC. Orders matched on the
wrong 5m windows.

- ResolveTokenIDsAt now returns StatusWindowMismatch on slug-hit
  startDate disagreement instead of substituting the wrong market.
- New strict ResolveTokenIDsForWindow never falls through to the
  unanchored search; intended as the only resolver entry point on
  the live order-placement path.
- CryptoMarket and ResolveResult carry StartDate/EndDate from Gamma.
- New StatusWindowMismatch status distinguishes wrong-window from
  no-market.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Position schema additions for V2 redeem detection

**The gap.** `pkg/types.Position` has no `redeemable` field, so go-bot cannot detect when a winning position is ready to redeem. The official Data API returns `redeemable`, `mergeable`, `negativeRisk`, `outcome`, `outcomeIndex`, `oppositeOutcome`, `oppositeAsset`, `endDate` — none of which polygolem currently surfaces.

**Files:**
- Modify: `internal/dataapi/client.go`, `pkg/types/data.go`, `pkg/data/client.go`, `pkg/universal/client.go`
- Modify: `tests/public_sdk_boundary_test.go`

- [ ] **Step 1: Extend the internal Position type**

In `internal/dataapi/client.go`, append to the `Position` struct (preserve existing field ordering and tags; new fields go at the end):

```go
type Position struct {
    TokenID         string  `json:"asset"`           // existing
    ConditionID     string  `json:"conditionId"`     // existing
    // ... existing fields ...
    Redeemable      bool    `json:"redeemable"`      // NEW
    Mergeable       bool    `json:"mergeable"`       // NEW
    NegativeRisk    bool    `json:"negativeRisk"`    // NEW
    Outcome         string  `json:"outcome"`         // NEW
    OutcomeIndex    int     `json:"outcomeIndex"`    // NEW
    OppositeOutcome string  `json:"oppositeOutcome"` // NEW
    OppositeAsset   string  `json:"oppositeAsset"`   // NEW
    EndDate         string  `json:"endDate"`         // NEW
    Title           string  `json:"title"`           // NEW (optional, useful for CLI output)
    Slug            string  `json:"slug"`            // NEW
    EventSlug       string  `json:"eventSlug"`       // NEW
}
```

Verify the existing JSON tags. The current code may use `json:"token_id"` (snake_case); the upstream API uses camelCase (`asset`). Run:

```bash
grep -nB1 -A12 "type Position struct" internal/dataapi/client.go
```
If the existing tags are wrong (snake_case), do **not** change them in this task — flag and stop. The wrong-tag case is a pre-existing bug requiring its own commit.

- [ ] **Step 2: Mirror on the public DTO**

In `pkg/types/data.go`, mirror the same fields with the same JSON tags. Keep the existing field order; append new fields at the end.

- [ ] **Step 3: Pass-through layers compile clean**

```bash
go build ./pkg/data/... ./pkg/universal/...
```
Expected: no compile errors. The pass-through clients shouldn't require changes since they `return c.inner.Method(...)`.

- [ ] **Step 4: Add httptest unit test**

In `internal/dataapi/client_test.go`, add a test that serves a fixture JSON containing `redeemable: true`, `mergeable: false`, `negativeRisk: true`, etc., and asserts each field decodes correctly.

- [ ] **Step 5: Re-pin SDK boundary**

If `tests/public_sdk_boundary_test.go` references `Position` field names, add the new fields to the boundary check.

- [ ] **Step 6: Verify**

```bash
go test ./internal/dataapi/... ./pkg/data/... ./pkg/universal/... ./pkg/types/... ./tests/... -count=1
```
Expected: all green.

- [ ] **Step 7: Commit**

```bash
git add internal/dataapi/client.go internal/dataapi/client_test.go pkg/types/data.go tests/public_sdk_boundary_test.go
git commit -m "feat(dataapi): surface redeemable/mergeable/negativeRisk on Position

Polymarket's GET /positions response carries redeemable, mergeable,
negativeRisk, outcome, outcomeIndex, oppositeOutcome, oppositeAsset,
endDate, title, slug, and eventSlug. Without these fields a settlement
worker has no way to detect when a winning position is ready to redeem
or which collateral adapter (Ctf vs NegRiskCtf) to route the redeem
through.

Additive only. Existing PnL/size/avgPrice fields unchanged.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: pkg/contracts — V2 collateral adapters and ramps

**The gap.** `pkg/contracts.Registry` exposes `CTFExchangeV2`, `NegRiskExchangeV2`, `NegRiskAdapterV2`, `PUSD`, `CTF` — sufficient for trading but missing the V2 collateral layer needed for redeem (and for any future split/merge or pUSD off-ramp work).

**Files:**
- Modify: `pkg/contracts/contracts.go`, `pkg/contracts/contracts_test.go`
- Modify: `tests/public_sdk_boundary_test.go`, `docs-site/src/content/docs/docs/reference/sdk.mdx`

- [ ] **Step 1: Add constants and registry fields**

In `pkg/contracts/contracts.go`, append to the constant block:

```go
const (
    // ... existing ...

    // V2 collateral adapters — route split/merge/redeem through these
    // from the deposit wallet. They wrap legacy CT calls and return pUSD.
    CtfCollateralAdapter        = "0xAdA100Db00Ca00073811820692005400218FcE1f"
    NegRiskCtfCollateralAdapter = "0xadA2005600Dec949baf300f4C6120000bDB6eAab"

    // V2 collateral ramps — convert between USDC/USDC.e and pUSD.
    CollateralOnramp  = "0x93070a847efEf7F70739046A929D47a521F5B8ee"
    CollateralOfframp = "0x2957922Eb93258b93368531d39fAcCA3B4dC5854"
    PermissionedRamp  = "0xebC2459Ec962869ca4c0bd1E06368272732BCb08"
)
```

Add the same fields to `Registry` and `PolygonMainnet()`. Keep JSON tags lowerCamelCase to match the existing pattern.

- [ ] **Step 2: Add the redeem-adapter helper**

```go
// RedeemAdapterFor returns the V2 collateral adapter address that a
// deposit wallet must call redeemPositions on for a given market kind.
// The adapter wraps the legacy CT call and returns pUSD to the caller.
func RedeemAdapterFor(negRisk bool) string {
    if negRisk {
        return NegRiskCtfCollateralAdapter
    }
    return CtfCollateralAdapter
}
```

- [ ] **Step 3: Add tests**

In `pkg/contracts/contracts_test.go`:

1. `TestPolygonMainnetIncludesV2Adapters` — registry exposes both adapter addresses and all three ramp addresses with the documented values.
2. `TestRedeemAdapterFor` — `false → CtfCollateralAdapter`, `true → NegRiskCtfCollateralAdapter`.

- [ ] **Step 4: Re-pin SDK boundary**

```bash
grep -n "contracts\." tests/public_sdk_boundary_test.go
```
Add the new constants and `RedeemAdapterFor` to the boundary check.

- [ ] **Step 5: Update Starlight SDK reference**

In `docs-site/src/content/docs/docs/reference/sdk.mdx` under `## pkg/contracts`, list the new constants and `RedeemAdapterFor`.

- [ ] **Step 6: Verify**

```bash
go test ./pkg/contracts/... ./tests/... -count=1
```

- [ ] **Step 7: Commit**

```bash
git add pkg/contracts/ tests/public_sdk_boundary_test.go docs-site/src/content/docs/docs/reference/sdk.mdx
git commit -m "feat(contracts): expose V2 collateral adapters and ramps

V2 deposit-wallet split/merge/redeem must route through the V2
collateral adapter (CtfCollateralAdapter, or NegRiskCtfCollateralAdapter
for neg-risk markets). The adapter pulls the wallet's CTF tokens, calls
ConditionalTokens.redeemPositions internally with USDC.e, then wraps
the proceeds back into pUSD and sends pUSD to the wallet.

Adds the two V2 adapters, the three V2 collateral ramps (Onramp,
Offramp, PermissionedRamp), and a RedeemAdapterFor(negRisk bool)
helper. Backed by source-of-truth references in
opensource-projects/repos/ctf-exchange-v2.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Adapter approval batch + onboard integration + CLI migration command

**The gap.** Today's WALLET-CREATE post-deploy approval batch (`internal/relayer/approvals.go:BuildApprovalCalls`) ships 6 calls: pUSD `approve` + CTF `setApprovalForAll` for `{CTFExchangeV2, NegRiskExchangeV2, NegRiskAdapterV2}`. None of those is a V2 collateral adapter. `CtfCollateralAdapter.redeemPositions` does `safeBatchTransferFrom(msg.sender, address(this), positionIds, amounts, "")` and therefore requires `CTF.setApprovalForAll(adapter, true)` from the deposit wallet. Without this approval, redeem reverts.

**The fix.** New helper `BuildAdapterApprovalCalls()` returns 4 calls: pUSD `approve` + CTF `setApprovalForAll` for both V2 adapters. New CLI command `polygolem deposit-wallet approve-adapters` submits the batch as a standalone WALLET op (idempotent — re-issuing on a wallet that already approved is a no-op). `OnboardDepositWallet` is extended to include the 4 calls in the post-deploy batch so new wallets are redeem-ready without operator intervention.

**Files:**
- Modify: `internal/relayer/approvals.go`, `internal/relayer/approvals_test.go` (create if absent)
- Modify: `pkg/relayer/onboard.go`, `pkg/relayer/onboard_test.go` (or the existing relayer_test.go)
- Modify: `internal/cli/deposit_wallet.go`, `internal/cli/deposit_wallet_test.go`

- [ ] **Step 1: Add adapter spender constants and helper**

In `internal/relayer/approvals.go`, alongside the existing constants:

```go
const (
    // ... existing ...
    ctfCollateralAdapter        = "0xAdA100Db00Ca00073811820692005400218FcE1f"
    negRiskCtfCollateralAdapter = "0xadA2005600Dec949baf300f4C6120000bDB6eAab"
)

// BuildAdapterApprovalCalls returns the 4 calls a deposit wallet must
// submit before V2 split/merge/redeem can succeed. Idempotent: re-issuing
// on a wallet that already approved is a no-op (max-approval is sticky;
// setApprovalForAll(true) is sticky).
//
// Required because CtfCollateralAdapter.redeemPositions calls
// safeBatchTransferFrom(msg.sender, address(this), ...) on the CTF.
func BuildAdapterApprovalCalls() []DepositWalletCall {
    calls := make([]DepositWalletCall, 0, 4)
    for _, spender := range []string{ctfCollateralAdapter, negRiskCtfCollateralAdapter} {
        calls = append(calls,
            buildApproveCall(pusdAddress, spender),
            buildCTFApprovalCall(spender),
        )
    }
    return calls
}
```

- [ ] **Step 2: Add a unit test for calldata correctness**

In `internal/relayer/approvals_test.go` (create the file if it doesn't exist):

```go
func TestBuildAdapterApprovalCallsCalldata(t *testing.T) {
    calls := BuildAdapterApprovalCalls()
    if len(calls) != 4 {
        t.Fatalf("len=%d want 4", len(calls))
    }
    // Call 0: pUSD.approve(CtfCollateralAdapter, MaxUint256)
    if !strings.EqualFold(calls[0].Target, "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB") {
        t.Fatalf("call0 target=%s", calls[0].Target)
    }
    if !strings.HasPrefix(strings.ToLower(calls[0].Data), "0x095ea7b3") {
        t.Fatalf("call0 selector=%s", calls[0].Data[:10])
    }
    if !strings.Contains(strings.ToLower(calls[0].Data), "ada100874d00e3331d00f2007a9c336a65009718") {
        t.Fatalf("call0 spender not encoded: %s", calls[0].Data)
    }
    // Call 1: CTF.setApprovalForAll(CtfCollateralAdapter, true)
    if !strings.EqualFold(calls[1].Target, "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045") {
        t.Fatalf("call1 target=%s", calls[1].Target)
    }
    if !strings.HasPrefix(strings.ToLower(calls[1].Data), "0xa22cb465") {
        t.Fatalf("call1 selector=%s", calls[1].Data[:10])
    }
    // Calls 2 & 3: same shape but for NegRiskCtfCollateralAdapter
    if !strings.Contains(strings.ToLower(calls[2].Data), "ada200001000ef00d07553cee7006808f895c6f1") {
        t.Fatalf("call2 spender not encoded: %s", calls[2].Data)
    }
}
```

- [ ] **Step 3: Bake adapter approvals into `OnboardDepositWallet`**

In `pkg/relayer/onboard.go`, find the line that builds the approval batch:

```go
calls := BuildApprovalCalls()
```

Replace with:

```go
calls := append(BuildApprovalCalls(), BuildAdapterApprovalCalls()...)
```

(Verify the import path; `BuildAdapterApprovalCalls` lives in `internal/relayer` and is re-exported via `pkg/relayer` if needed — match the existing re-export pattern for `BuildApprovalCalls`.)

Update the onboard test to assert the batch length is now 10.

- [ ] **Step 4: Add the `approve-adapters` CLI subcommand**

In `internal/cli/deposit_wallet.go`, register a new subcommand alongside `deposit-wallet onboard`:

```go
func depositWalletApproveAdaptersCmd(jsonOut bool) *cobra.Command {
    var submit bool
    var confirm string
    var wait bool
    var timeout time.Duration
    cmd := &cobra.Command{
        Use:   "approve-adapters",
        Short: "Approve V2 collateral adapters for redeem (one-shot per wallet)",
        Long: "Submits the 4-call approval batch (pUSD.approve + CTF.setApprovalForAll for " +
            "CtfCollateralAdapter and NegRiskCtfCollateralAdapter). Required once per " +
            "deposit wallet before V2 redeem will succeed. Idempotent.",
        Args: cobra.NoArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Build the 4 calls and print them by default.
            // 2. Submit only with --submit --confirm APPROVE_ADAPTERS.
            // 3. Init signer + relayer client (mirror existing onboard path).
            // 4. Verify deposit wallet is deployed (reuse the dual-source check).
            // 5. Fetch nonce, build deadline, sign WALLET batch over BuildAdapterApprovalCalls().
            // 6. Submit + (optionally) poll until terminal.
            // 7. Emit JSON envelope { state, transactionID, calls: [...] }.
        },
    }
    cmd.Flags().BoolVar(&submit, "submit", false, "sign and submit the adapter approval batch")
    cmd.Flags().StringVar(&confirm, "confirm", "", "must be APPROVE_ADAPTERS when --submit is set")
    cmd.Flags().BoolVar(&wait, "wait", true, "poll until transaction reaches terminal state")
    cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "max wait time for --wait")
    return cmd
}
```

Wire it into the `deposit-wallet` parent command.

- [ ] **Step 5: CLI test**

In `internal/cli/deposit_wallet_test.go`, add `TestDepositWalletApproveAdaptersHappyPath` using an httptest relayer mock that:
- Returns `deployed=true` from `/deployed`
- Returns nonce `42` from `/nonce`
- Captures the WALLET batch posted to `/submit`, asserts the calls array has length 4 and matches `BuildAdapterApprovalCalls()` byte-for-byte.
- Returns `STATE_CONFIRMED` from `/transaction`.

- [ ] **Step 6: Regenerate generated docs**

```bash
go run ./cmd/polygolem_docs
```
Expected: `docs/COMMANDS.md` and `docs-site/src/content/docs/docs/reference/cli.mdx` pick up the new subcommand.

- [ ] **Step 7: Verify**

```bash
go test ./internal/relayer/... ./pkg/relayer/... ./internal/cli/... -count=1
go run ./cmd/polygolem_docs -check
```
Expected: all green.

- [ ] **Step 8: Commit**

```bash
git add internal/relayer/approvals.go internal/relayer/approvals_test.go pkg/relayer/onboard.go pkg/relayer/onboard_test.go internal/cli/deposit_wallet.go internal/cli/deposit_wallet_test.go docs/COMMANDS.md docs-site/src/content/docs/docs/reference/cli.mdx
git commit -m "feat(relayer,cli): approve V2 collateral adapters for redeem

V2 deposit-wallet redemption requires the wallet to have approved the
V2 collateral adapter (CtfCollateralAdapter for binary markets,
NegRiskCtfCollateralAdapter for neg-risk) on the CTF contract. The
existing 6-call WALLET-CREATE approval batch covers trading spenders
(CTFExchangeV2, NegRiskExchangeV2, legacy NegRiskAdapter) but not the
V2 collateral adapters, so today's wallets cannot redeem.

- BuildAdapterApprovalCalls() returns the 4 calls (pUSD.approve + CTF
  .setApprovalForAll for both V2 adapters). Idempotent.
- OnboardDepositWallet now ships a 10-call post-deploy batch so new
  wallets are redeem-ready out of the box.
- New CLI: polygolem deposit-wallet approve-adapters — one-shot
  migration for existing live wallets (e.g. 0x21999a07...).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: pkg/settlement — V2 redeem SDK

**The gap.** No public Go API turns a redeemable position into a deposit-wallet WALLET batch. go-bot's eventual settlement worker needs `FindRedeemable`, `BuildRedeemCall`, and `SubmitRedeem` so it can stay schedule-free in polygolem's repo.

**Files:**
- Create: `pkg/settlement/settlement.go`, `pkg/settlement/settlement_test.go`
- Modify: `tests/public_sdk_boundary_test.go`, `tests/repository_hygiene_test.go`

- [ ] **Step 1: Package skeleton**

```go
// Package settlement turns redeemable Polymarket V2 positions into
// deposit-wallet WALLET batches that route through the V2 collateral
// adapters and return pUSD to the wallet.
//
// Stability: FindRedeemable, BuildRedeemCall, SubmitRedeem, and
// RedeemablePosition are part of the polygolem public SDK and follow
// semver.
package settlement

import (
    "context"
    "encoding/hex"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/common"

    "github.com/TrebuchetDynamics/polygolem/pkg/contracts"
    "github.com/TrebuchetDynamics/polygolem/pkg/ctf"
    "github.com/TrebuchetDynamics/polygolem/pkg/data"
    "github.com/TrebuchetDynamics/polygolem/pkg/relayer"
)

type RedeemablePosition struct {
    TokenID      string  `json:"tokenID"`
    ConditionID  string  `json:"conditionID"`
    Size         float64 `json:"size"`
    NegativeRisk bool    `json:"negativeRisk"`
    EndDate      string  `json:"endDate"`
    Title        string  `json:"title"`
}

// FindRedeemable returns positions with redeemable=true for owner via
// the Data API. Does not call Gamma; NegativeRisk is taken from the
// position payload.
func FindRedeemable(ctx context.Context, dataClient *data.Client, owner string) ([]RedeemablePosition, error)

// BuildRedeemCall encodes redeemPositions(address(0), bytes32(0), conditionId, [])
// targeting the V2 collateral adapter for the position's market kind.
// Calldata reuses pkg/ctf.RedeemPositionsData; only the call target
// switches.
func BuildRedeemCall(p RedeemablePosition) (relayer.DepositWalletCall, error)

// SubmitRedeem groups positions by conditionID, builds one DepositWalletCall
// per condition, signs a single WALLET batch (capped at limit calls), and
// submits via the relayer. Returns the relayer transaction. Idempotent on
// already-redeemed positions because the adapter zero-pays them.
func SubmitRedeem(
    ctx context.Context,
    rc *relayer.Client,
    privateKey string,
    positions []RedeemablePosition,
    limit int,
) (*relayer.RelayerTransaction, error)
```

- [ ] **Step 2: Implement `FindRedeemable`**

```go
func FindRedeemable(ctx context.Context, dataClient *data.Client, owner string) ([]RedeemablePosition, error) {
    rows, err := dataClient.Positions(ctx, owner)
    if err != nil {
        return nil, fmt.Errorf("settlement: positions: %w", err)
    }
    out := make([]RedeemablePosition, 0, len(rows))
    for _, p := range rows {
        if !p.Redeemable {
            continue
        }
        out = append(out, RedeemablePosition{
            TokenID:      p.TokenID,
            ConditionID:  p.ConditionID,
            Size:         p.Size,
            NegativeRisk: p.NegativeRisk,
            EndDate:      p.EndDate,
            Title:        p.Title,
        })
    }
    return out, nil
}
```

- [ ] **Step 3: Implement `BuildRedeemCall`**

```go
func BuildRedeemCall(p RedeemablePosition) (relayer.DepositWalletCall, error) {
    if p.ConditionID == "" {
        return relayer.DepositWalletCall{}, fmt.Errorf("settlement: empty conditionID")
    }
    cid := common.HexToHash(p.ConditionID)
    // Adapter ignores collateralToken arg + indexSets (uses partition() = [1,2]
    // internally). Pass zero values for minimal calldata.
    data, err := ctf.RedeemPositionsData(common.Address{}, common.Hash{}, cid, []*big.Int{})
    if err != nil {
        return relayer.DepositWalletCall{}, fmt.Errorf("settlement: encode redeem: %w", err)
    }
    return relayer.DepositWalletCall{
        Target: contracts.RedeemAdapterFor(p.NegativeRisk),
        Value:  "0",
        Data:   "0x" + hex.EncodeToString(data),
    }, nil
}
```

- [ ] **Step 4: Implement `SubmitRedeem`**

Mirror the `OnboardDepositWallet` flow:

1. Init signer from privateKey.
2. Derive owner + deposit wallet via `relayer.DepositWalletAddress`.
3. Verify deployed (relayer + on-chain fallback — reuse `relayer.DepositWalletCodeDeployed` if relayer false).
4. Group `positions` by `conditionID` (collapse duplicates from neg-risk markets where YES and NO show as separate rows but redeem with one call).
5. Build one `DepositWalletCall` per unique condition. Cap at `limit` (default 10).
6. Fetch nonce, build deadline, sign WALLET batch, submit, poll, return transaction.

If `len(positions) > limit`, return only the first `limit` and the caller decides whether to call again.

- [ ] **Step 5: Tests**

In `pkg/settlement/settlement_test.go`:

1. `TestFindRedeemableFiltersNonRedeemable` — httptest Data API returns three positions: two `redeemable=false`, one `redeemable=true`. `FindRedeemable` returns exactly one row.
2. `TestBuildRedeemCallBinary` — encodes calldata identical to `ctf.RedeemPositionsData(common.Address{}, common.Hash{}, cid, []*big.Int{})`, target = `CtfCollateralAdapter`.
3. `TestBuildRedeemCallNegRisk` — `NegativeRisk: true` → target = `NegRiskCtfCollateralAdapter`.
4. `TestSubmitRedeemHappyPath` — httptest relayer; intercepts `/submit`, asserts the batch contains the expected calls in the expected order, returns `STATE_CONFIRMED`.
5. `TestSubmitRedeemIdempotenceAfterAlreadyRedeemed` — two positions where the wallet's CTF balance is now zero; the adapter would zero-pay. The test asserts `SubmitRedeem` does not refuse to build the call (idempotence is a contract-level property, not a settlement-package gate).
6. `TestSubmitRedeemRespectsLimit` — 15 positions, `limit=10` → batch length 10, return value indicates 5 more pending.

- [ ] **Step 6: Update repo hygiene**

In `tests/repository_hygiene_test.go`:

```go
if _, err := os.Stat(filepath.Join(root, "pkg/settlement")); err != nil {
    t.Fatalf("pkg/settlement public boundary is missing: %v", err)
}
```

- [ ] **Step 7: Re-pin SDK boundary**

In `tests/public_sdk_boundary_test.go`, add `pkg/settlement` symbols to the boundary test alongside `pkg/contracts`.

- [ ] **Step 8: Verify**

```bash
go test ./pkg/settlement/... ./pkg/ctf/... ./pkg/contracts/... ./tests/... -count=1
```

- [ ] **Step 9: Commit**

```bash
git add pkg/settlement/ tests/public_sdk_boundary_test.go tests/repository_hygiene_test.go
git commit -m "feat(settlement): V2 redeem SDK for deposit wallets

New pkg/settlement turns redeemable=true Data API positions into
WALLET batches routed through the V2 collateral adapter. Calldata
reuses pkg/ctf.RedeemPositionsData; the adapter reads conditionId and
ignores the collateralToken, parentCollection, and indexSets args
(uses partition() = [1,2] internally), so we pass zero values for
ignored fields.

Public SDK surface: FindRedeemable, BuildRedeemCall, SubmitRedeem,
RedeemablePosition. NegativeRisk routes to NegRiskCtfCollateralAdapter;
binary markets to CtfCollateralAdapter. Idempotent: re-running on a
wallet whose CTF balance is zero is a zero-pay no-op contract-side.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: CLI commands — `redeemable` and `redeem`

**The goal.** Operator-callable commands for ad-hoc redemption and dry-run inspection. go-bot's settlement worker must call `pkg/settlement` directly; the CLI is for humans and scripted operator runbooks only.

**Files:**
- Modify: `internal/cli/deposit_wallet.go`, `internal/cli/deposit_wallet_test.go`
- Modify: `docs/COMMANDS.md`, `docs-site/src/content/docs/docs/reference/cli.mdx` (regenerated)

- [ ] **Step 1: `deposit-wallet redeemable` subcommand**

Read-only; lists redeemable positions for the EOA's deposit wallet:

```go
func depositWalletRedeemableCmd(jsonOut bool) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "redeemable",
        Short: "List redeemable positions for the deposit wallet",
        Args:  cobra.NoArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Resolve EOA + deposit wallet from POLYMARKET_PRIVATE_KEY.
            // 2. Call settlement.FindRedeemable(ctx, dataClient, depositWalletAddress).
            //    NOTE: Polymarket Data API takes the deposit wallet as the
            //    "user" param, not the EOA, because positions live in the
            //    deposit wallet.
            // 3. Print JSON envelope with { count, positions: [...] }.
        },
    }
    return cmd
}
```

- [ ] **Step 2: `deposit-wallet redeem` subcommand**

```go
func depositWalletRedeemCmd(jsonOut bool) *cobra.Command {
    var dryRun bool
    var limit int
    var submit bool
    var confirm string
    cmd := &cobra.Command{
        Use:   "redeem",
        Short: "Redeem winning positions via deposit-wallet WALLET batch",
        Args:  cobra.NoArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Find redeemable positions.
            // 2. Build calls, print them, and return unless --submit is set.
            // 3. --submit requires --confirm REDEEM_WINNERS.
            // 4. Before submit: verify adapter approvals exist via on-chain
            //    isApprovedForAll(wallet, adapter); if false, return a
            //    structured error pointing to `approve-adapters`.
            // 5. Call settlement.SubmitRedeem(...).
            // 6. Print { transactionID, state, redeemed: [...positions...] }.
        },
    }
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "build calls and exit without signing")
    cmd.Flags().IntVar(&limit, "limit", 10, "max positions per WALLET batch")
    cmd.Flags().BoolVar(&submit, "submit", false, "sign and submit the redeem batch")
    cmd.Flags().StringVar(&confirm, "confirm", "", "must be REDEEM_WINNERS when --submit is set")
    return cmd
}
```

- [ ] **Step 3: Adapter-approval pre-check helper**

Add to `internal/rpc/code.go` (or a new `internal/rpc/erc1155.go`):

```go
// IsApprovedForAll calls CTF.isApprovedForAll(owner, operator) via eth_call.
// Returns false on RPC error to fail-closed.
func IsApprovedForAll(ctx context.Context, ctfAddress, owner, operator, rpcURL string) (bool, error)
```

The `redeem` subcommand uses this against `CtfCollateralAdapter` and `NegRiskCtfCollateralAdapter` (filtered by which positions are present in the to-redeem set) before submitting. If false, it returns:

```json
{
  "ok": false,
  "error": "deposit wallet has not approved CtfCollateralAdapter; run `polygolem deposit-wallet approve-adapters --submit --confirm APPROVE_ADAPTERS --wait` first",
  "missingApprovals": ["0xAdA100Db00Ca00073811820692005400218FcE1f"]
}
```

- [ ] **Step 4: CLI tests**

1. `TestDepositWalletRedeemableJSON` — httptest Data API returns one redeemable position; CLI prints `count=1` and the position.
2. `TestDepositWalletRedeemDryRun` — `--dry-run --json` builds calls, prints them, never invokes relayer.
3. `TestDepositWalletRedeemRefusesWithoutAdapterApproval` — `IsApprovedForAll` mock returns false; CLI fails with the structured error and suggests `approve-adapters`. Relayer mock asserts `/submit` was never called.
4. `TestDepositWalletRedeemHappyPath` — adapter approval present, relayer returns `STATE_CONFIRMED`; CLI prints `transactionID` and the redeemed list.

- [ ] **Step 5: Regenerate generated docs**

```bash
go run ./cmd/polygolem_docs
```

- [ ] **Step 6: Verify**

```bash
go test ./internal/cli/... ./internal/rpc/... ./pkg/settlement/... -count=1
go run ./cmd/polygolem_docs -check
```

- [ ] **Step 7: Commit**

```bash
git add internal/cli/deposit_wallet.go internal/cli/deposit_wallet_test.go internal/rpc/ docs/COMMANDS.md docs-site/src/content/docs/docs/reference/cli.mdx
git commit -m "feat(cli): deposit-wallet redeemable + redeem with adapter pre-check

- redeemable: read-only list of redeemable positions via Data API
- redeem [--dry-run] [--limit N]: signs WALLET batch through V2
  collateral adapters and submits to relayer. Pre-check verifies
  CTF.isApprovedForAll(wallet, adapter) before signing; if missing,
  fails closed with a clear pointer to \`approve-adapters\`.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Documentation + CHANGELOG + BLOCKERS update

**Files:**
- Modify: `CHANGELOG.md`, `BLOCKERS.md`, `docs/CONTRACTS.md`, `docs/SAFETY.md`
- Modify: `docs-site/src/content/docs/docs/concepts/{contracts,deposit-wallets,safety}.mdx`
- Modify: `docs-site/src/content/docs/docs/guides/deposit-wallet-lifecycle.mdx`
- Create: `docs-site/src/content/docs/docs/guides/redeem-winners.mdx`

- [ ] **Step 1: BLOCKERS.md — close B-10 sub-points and open B-11 if needed**

Update B-10 to mark the wrong-window trap and the missing-redeem-path as fixed by this plan. If POL reserve is still the only unresolved live blocker, leave that as the sole remaining open item.

- [ ] **Step 2: CHANGELOG.md — Unreleased entry**

```markdown
### Added

- **Market-window guard.** New `marketresolver.ResolveTokenIDsForWindow`
  fails closed when the resolver cannot bind to the exact decision
  window. New `StatusWindowMismatch` distinguishes wrong-window from
  no-market. `ResolveTokenIDsAt` now fails closed instead of silently
  substituting a different window.
- **V2 redeem primitives.** New `pkg/settlement` package
  (`FindRedeemable`, `BuildRedeemCall`, `SubmitRedeem`). New CLI
  commands `polygolem deposit-wallet redeemable`, `redeem [--dry-run]`,
  and `approve-adapters` (one-shot migration for existing wallets).
- **V2 collateral adapter registry.** `pkg/contracts` exposes
  `CtfCollateralAdapter`, `NegRiskCtfCollateralAdapter`,
  `CollateralOnramp`, `CollateralOfframp`, `PermissionedRamp`, plus
  `RedeemAdapterFor(negRisk bool) string`.
- **Position schema.** `pkg/types.Position` and `internal/dataapi.Position`
  now surface `redeemable`, `mergeable`, `negativeRisk`, `outcome`,
  `outcomeIndex`, `oppositeOutcome`, `oppositeAsset`, `endDate`,
  `title`, `slug`, `eventSlug`.

### Changed

- **Onboard batch is now 10 calls.** `OnboardDepositWallet` includes
  the 4 adapter-approval calls so new wallets are redeem-ready out of
  the box. Existing wallets must run `deposit-wallet approve-adapters`
  once.
```

- [ ] **Step 3: docs/CONTRACTS.md — V2 collateral layer section**

Add a section after the existing `1.3 Deployment Status Source of Truth`:

> ### 1.4 V2 Collateral Layer
>
> V2 redeem does **not** call ConditionalTokens directly. Polymarket
> publishes thin collateral adapters that bridge between the legacy
> CTF and the V2 pUSD wrapper:
>
> | Contract | Address | Use |
> |---|---|---|
> | CtfCollateralAdapter | `0xAdA100Db00Ca00073811820692005400218FcE1f` | Binary up/down markets |
> | NegRiskCtfCollateralAdapter | `0xadA2005600Dec949baf300f4C6120000bDB6eAab` | Neg-risk multi-outcome markets |
> | CollateralOnramp | `0x93070a847efEf7F70739046A929D47a521F5B8ee` | USDC/USDC.e → pUSD |
> | CollateralOfframp | `0x2957922Eb93258b93368531d39fAcCA3B4dC5854` | pUSD → USDC/USDC.e |
> | PermissionedRamp | `0xebC2459Ec962869ca4c0bd1E06368272732BCb08` | EIP-712 witness-signed wrap/unwrap |
>
> The deposit wallet must approve the adapters before redeem will
> succeed: `pUSD.approve(adapter, MaxUint256)` and
> `CTF.setApprovalForAll(adapter, true)`. New wallets pick this up
> automatically via the 10-call onboard batch; existing wallets must
> run `polygolem deposit-wallet approve-adapters` once.

- [ ] **Step 4: docs/SAFETY.md — market-window guard rule**

Add a numbered rule to the `Live Readiness` section:

> 8. **Decision-window safety.** Order placement must use
>    `marketresolver.ResolveTokenIDsForWindow` rather than the
>    unanchored `ResolveTokenIDs` or the slug-then-fallback
>    `ResolveTokenIDsAt`. The strict resolver returns
>    `StatusWindowMismatch` when the matched market's `startDate`
>    disagrees with the requested window — this is a fail-closed
>    signal that must abort the order. Live evidence: 2026-05-09
>    SOL/ETH 5m signals filled buys against future market windows
>    because the previous resolver silently substituted a different
>    market.

- [ ] **Step 5: Starlight mirrors**

Mirror the contracts.md V2 collateral layer into
`docs-site/src/content/docs/docs/concepts/contracts.mdx`.

Mirror the safety.md rule into
`docs-site/src/content/docs/docs/concepts/safety.mdx`.

Update `docs-site/src/content/docs/docs/guides/deposit-wallet-lifecycle.mdx` Phase 4 (Approve) to mention the 10-call batch and what each call does.

- [ ] **Step 6: New guide — redeem winners**

Create `docs-site/src/content/docs/docs/guides/redeem-winners.mdx`:

```mdx
---
title: Redeeming Winning Positions
description: How polygolem detects, batches, and submits V2 redemption transactions for resolved deposit-wallet positions.
---

# Redeeming Winning Positions

Polymarket V2 deposit wallets do not redeem through the legacy
ConditionalTokens contract directly. They route split, merge, and
redeem actions through thin collateral adapters that wrap pUSD ↔ CTF
operations and return pUSD to the caller.

## Detection

`GET /positions?user=<depositWallet>` exposes `redeemable: boolean`.
A position is ready to redeem when both the market is resolved and
the wallet still holds the winning CTF balance.

```bash
polygolem deposit-wallet redeemable --json
```

## Submission

```bash
polygolem deposit-wallet redeem --dry-run --json   # inspect calldata
polygolem deposit-wallet redeem --submit --confirm REDEEM_WINNERS --json
```

The `redeem` command:

1. Lists `redeemable=true` positions for the wallet.
2. Checks `CTF.isApprovedForAll(wallet, adapter)` for every adapter
   the to-redeem set requires; refuses to sign if any approval is
   missing and points to `approve-adapters`.
3. Builds one `DepositWalletCall` per unique conditionID, targeting
   `CtfCollateralAdapter` (binary) or `NegRiskCtfCollateralAdapter`
   (neg-risk).
4. Signs a single WALLET batch via the relayer and polls until
   terminal.

## SDK

```go
import "github.com/TrebuchetDynamics/polygolem/pkg/settlement"

positions, _ := settlement.FindRedeemable(ctx, dataClient, depositWallet)
tx, _ := settlement.SubmitRedeem(ctx, relayerClient, privateKey, positions, 10)
```

See [Smart Contracts §1.4](/docs/concepts/contracts/#v2-collateral-layer) for adapter addresses and [Deposit Wallet Lifecycle](/docs/guides/deposit-wallet-lifecycle/) for the full lifecycle context.
```

- [ ] **Step 7: Verify**

```bash
go run ./cmd/polygolem_docs -check
go test ./tests/... -count=1
cd docs-site && npm run build
```
Expected: docs-check green, all tests green, Astro build clean.

- [ ] **Step 8: Commit**

```bash
git add CHANGELOG.md BLOCKERS.md docs/CONTRACTS.md docs/SAFETY.md docs-site/src/content/docs/docs/concepts/ docs-site/src/content/docs/docs/guides/
git commit -m "docs: V2 collateral layer + redeem guide + window-guard rule

Documents the new pkg/settlement and approve-adapters surface, the V2
collateral adapter addresses, and the decision-window safety rule.
New Starlight guide at /docs/guides/redeem-winners/.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: Redeploy Starlight + final verification

- [ ] **Step 1: Push the work**

```bash
git push
```

- [ ] **Step 2: Redeploy**

```bash
cd docs-site
set -a && . ../.env && set +a
npm run deploy 2>&1 | tail -10
```

Expected: `✨ Deployment complete!` line with a `https://<hash>.polygolem.pages.dev` URL.

- [ ] **Step 3: Verify content live on production**

```bash
curl -s https://polygolem.pages.dev/docs/guides/redeem-winners/ \
  | grep -oE "approve-adapters|CtfCollateralAdapter|FindRedeemable" | sort -u
```

Expected:
```
approve-adapters
CtfCollateralAdapter
FindRedeemable
```

- [ ] **Step 4: Sanity probes against the live wallet**

After this point the live operator runs (manually, gated by an explicit
operator decision because each step is live-money):

```bash
# 1) Approve adapters once.
POLYMARKET_PRIVATE_KEY=… RELAYER_API_KEY=… RELAYER_API_KEY_ADDRESS=… \
  polygolem deposit-wallet approve-adapters --submit --confirm APPROVE_ADAPTERS --wait --json

# 2) List anything redeemable.
POLYMARKET_PRIVATE_KEY=… polygolem deposit-wallet redeemable --json

# 3) Dry-run redeem (no signing).
POLYMARKET_PRIVATE_KEY=… polygolem deposit-wallet redeem --dry-run --json

# 4) Submit only after operator approval.
POLYMARKET_PRIVATE_KEY=… RELAYER_API_KEY=… RELAYER_API_KEY_ADDRESS=… \
  polygolem deposit-wallet redeem --submit --confirm REDEEM_WINNERS --json
```

These steps are out-of-scope for the agentic worker — they require a live
operator and live funds. Document the runbook in `docs/SAFETY.md` if not
already covered.

---

## What this leaves for go-bot

- **Settlement worker loop.** Periodic poll (`60s` for active markets,
  `5m` for the long tail) that calls `settlement.FindRedeemable` and
  `settlement.SubmitRedeem` with a sane batch limit.
- **Order-placement gate.** Replace any use of
  `marketresolver.ResolveTokenIDs[At]` on the live decision path with
  `ResolveTokenIDsForWindow`. On `StatusWindowMismatch`, abort the
  order, log the source string, and emit a metric.
- **Telemetry.** Counters for `StatusWindowMismatch`, redeem attempts,
  redeem failures with structured reasons (especially missing-adapter-
  approval, which should be impossible after migration but is still
  worth alerting on).

---

## Verification matrix

| What | Where | Pass criterion |
|---|---|---|
| Window guard rejects wrong-window slug hit | `pkg/marketresolver/resolver_test.go` | `StatusWindowMismatch` returned |
| Strict resolver never falls through | same | mock Gamma `/events` search call count == 0 |
| Position decodes redeemable | `internal/dataapi/client_test.go` | fixture w/ `redeemable:true` decodes correctly |
| pkg/contracts exposes V2 adapters | `pkg/contracts/contracts_test.go` | constants and `RedeemAdapterFor` correct |
| Adapter approval calldata is correct | `internal/relayer/approvals_test.go` | selector + spender + max-uint encoded |
| Onboard batch is 10 calls | `pkg/relayer/onboard_test.go` | `len(calls) == 10` |
| approve-adapters CLI submits batch | `internal/cli/deposit_wallet_test.go` | relayer mock sees the 4-call batch |
| Settlement filters non-redeemable | `pkg/settlement/settlement_test.go` | one-of-three positions returned |
| Settlement targets correct adapter | same | `NegativeRisk` toggles target |
| Redeem CLI refuses without approval | `internal/cli/deposit_wallet_test.go` | structured error, no `/submit` call |
| Redeem CLI happy path | same | `STATE_CONFIRMED` returned |
| Docs round-trip | `cmd/polygolem_docs -check` | exit 0 |
| Starlight builds | `npm run build` in `docs-site/` | exit 0 |
| Starlight deploy live | `curl polygolem.pages.dev/...` | new content visible |
