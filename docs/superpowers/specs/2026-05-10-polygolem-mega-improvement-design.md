# Polygolem Mega Improvement Plan

> **Status:** Approved after grilling — ready for implementation plan
> **Date:** 2026-05-10
> **Scope:** Taxonomy, Architecture, TDD, E2E, Features, and V2 Accuracy

---

## 1. Executive Summary

Polygolem is production-validated (`v0.1.0`) against Polymarket V2 deposit-wallet flows. This plan maps the **six improvement vectors** required to evolve it from a hardened single-operator CLI into a **tier-1 open-source Polymarket SDK**.

The work is organized into **five parallel tracks** (not sequential phases), each producing working, testable software independently.

---

## 2. Current State Assessment

### Strengths
- Production-validated V2 deposit-wallet flow (2026-05-08 reference run)
- Excellent documentation culture (walkthroughs, canonical docs, Astro site)
- Safety-first design (read-only default, risk controls, circuit breakers)
- Clean `pkg/` / `internal/` boundary with 64 test files / 438 test functions
- Settlement support (V2 collateral adapters, redeem paths)

### Critical Gaps

| Vector | Gap | Risk |
|--------|-----|------|
| **Taxonomy** | Deprecated `pkg/bookreader` still shipped; no public `pkg/orders`, `pkg/wallet`, `pkg/auth` | SDK consumers cannot build orders or manage wallets without importing `internal/` |
| **Architecture** | No telemetry/metrics; no plugin/extension surface | Operational blindness; no third-party extension path |
| **TDD** | No property-based tests; no benchmarks; coverage unknown | Regressions in edge cases (price/size bounds, neg-risk routing) |
| **E2E** | Only read-only live E2E exists; no authenticated trading E2E; no contract simulation | Changes to signing or relayer paths are only validated manually with real funds |
| **Features** | Missing Heartbeats, User WebSocket, Withdrawals, GraphQL, CTF split/merge, builder trades | Automated bots cannot keep orders alive; SDK lacks full V2 surface |
| **Accuracy** | Neg-risk exchange address hardcoded per-order instead of per-market selection; no cursor pagination | Orders on neg-risk markets may hit wrong exchange; large result sets are unstable |

---

## 3. Track Breakdown (Parallel, Independent)

### Track 1 — Taxonomy & Public SDK Cleanup
**Goal:** Remove deprecated surfaces and promote stable internals to `pkg/` so Polydart and other consumers have a complete public API.

**Decisions from grilling:**
- **Staged promotion** (not big-bang): Promote `pkg/bridge`, `pkg/ctf`, and `pkg/wallet` first (stable APIs).
- **Experimental track** for unstable APIs: `pkg/experimental/orders` and `pkg/experimental/auth` with clear "API may change" disclaimers. Promote to stable `pkg/` after one release cycle.
- **Immediate deletion**: `pkg/bookreader` (deprecated compatibility wrapper).

**Tasks:**
1. Delete `pkg/bookreader` and update all references.
2. Promote `internal/bridge` → `pkg/bridge` (deposit/quote/withdrawal methods).
3. Promote `internal/ctf` → `pkg/ctf` (split/merge/redeem calldata builders).
4. Promote `internal/wallet` → `pkg/wallet` (DeriveDepositWallet, BuildBatch, SignBatch).
5. Create `pkg/experimental/orders` — expose OrderIntent builder and validation. Mark unstable.
6. Create `pkg/experimental/auth` — expose Signer, EIP712Domain, POLY1271Wrapper. Keep private-key handling internal.
7. Expand `tests/public_sdk_boundary_test.go` to verify zero `internal/` leakage.

### Track 2 — Architecture & Observability
**Goal:** Add telemetry, rate-limit enforcement, and a plugin boundary.

**Decisions from grilling:**
- **No go-bot extraction** — `go-bot` is already in its own repo (`polymarket-mega-bot`). Probes (`live_siwe_probe`, `indexer_probe`) stay in `polygolem/cmd/` as polygolem-native diagnostics.
- **Structured logs as primary surface** — `log/slog` with JSON output. Every HTTP request is a log line with `method`, `duration_ms`, `status_code`. No tracing backend required.
- **Rate-limit enforcement** — token-bucket enforcement matching Polymarket's published limits.

**Tasks:**
1. Add `internal/telemetry` — structured logging (`log/slog`) on all protocol clients.
2. Add `internal/ratelimit` — token-bucket enforcement per endpoint family (CLOB read/write, Gamma, Data API).
3. Replace hardcoded `negRiskExchangeAddress` with per-market lookup via `pkg/contracts` based on `ClobMarketInfo.negRisk` flag.
4. Define `pkg/plugins` interface boundary — `MarketDataPlugin` and `RiskPlugin` for third-party extensions.

### Track 3 — TDD Hardening
**Goal:** Move from "tests exist" to "tests prevent regressions."

**Decisions from grilling:**
- **Coverage gate**: Measure current coverage first, then set gate at `max(current, 60%)`.
- **Behavioral rule**: Every exported function in `pkg/` must have at least one test asserting its contract (not just "doesn't panic"). CLI commands exempt from numeric gate but must have E2E/integration tests.
- **Property-based + golden vectors + benchmarks + contract simulation**.

**Tasks:**
1. Measure coverage baseline (`go test -coverprofile`).
2. Expand golden vectors: neg-risk order signing, market-order FOK/FAK, batch order hash consistency.
3. Add property-based tests (price/size bounds, tick-size alignment, CREATE2 derivation).
4. Add benchmarks (`BenchmarkSignCLOBOrderV2`, `BenchmarkCreate2Derive`, `BenchmarkStreamDedup`).
5. Refactor complex tests to strict table-driven format with named subtests.
6. Add contract simulation tests using `go-ethereum/simulated.Backend` (CTF approvals, deposit wallet `isValidSignature` mock).

### Track 4 — E2E & Integration Validation
**Goal:** Close the "only read-only E2E" gap.

**Decisions from grilling:**
- **Mock conformance server for CI** — expand existing mock to validate V2 signatures, builder codes, order lifecycle. Zero financial risk.
- **Tiny real orders for manual verification only** — post $0.01 limit at extreme price, verify `live`, cancel. Document in operator runbook.
- **No live authenticated tests in PR CI** — too risky with real funds.

**Tasks:**
1. Expand `tests/e2e_public_sdk_test.go` mock server to cover batch orders, heartbeats, cancel-all.
2. Add `tests/e2e_contract_sim_test.go` using `simulated.Backend` (CREATE2 match, ERC-1271 verification).
3. Add WebSocket E2E: user channel subscription, heartbeat ping/pong, auto-reconnect under partition.
4. Expand `tests/docs_safety_test.go` to verify every CLI command in `docs/COMMANDS.md` has a corresponding test.

### Track 5 — V2 Feature Parity
**Goal:** Close every gap between polygolem and the official Polymarket V2 API surface.

**Decisions from grilling:**
- **Priority order**: Heartbeats (critical) → cursor pagination (high) → builder trades/scoring (medium) → CTF split/merge, withdrawals (medium) → defer GraphQL, user WebSocket, streaming pagination (low effort/value).
- **Add `docs/V2-PARITY-MAP.md`** documenting supported / partially supported / not supported endpoints with rationale.

**Tasks:**
1. **Heartbeats** — `pkg/clob.SendHeartbeat` + `polygolem clob heartbeat`. Manual first, standalone manager later.
2. **Cursor pagination** — `pkg/gamma.MarketsKeyset` and `EventsKeyset` using `next_cursor`/`after_cursor`.
3. **Builder trades / order scoring** — `pkg/clob.BuilderTrades` and `GetOrderScoringStatus`.
4. **CTF split/merge** — `pkg/ctf.SplitPositions` and `MergePositions` + CLI commands.
5. **Withdrawals** — `pkg/bridge.Withdraw` + CLI for bridging pUSD.
6. **Neg-risk per-market exchange** — dynamic lookup based on `ClobMarketInfo.negRisk` flag.
7. **Rate-limit status exposure** — `internal/transport` exposes `RateLimitStatus` headers.
8. **Defer to future**: User WebSocket, GraphQL/subgraph, streaming pagination, geoblock SDK method.

---

## 4. Success Criteria

| Track | Criteria |
|-------|----------|
| 1 | `pkg/` imports zero `internal/` packages; `pkg/bookreader` deleted; experimental packages created with stability disclaimers |
| 2 | Every protocol client emits structured logs; rate limits enforced and testable; neg-risk exchange selected per-market |
| 3 | CI fails if coverage < `max(current, 60%)`; golden vectors cover neg-risk + batch + FAK; benchmarks run on every PR |
| 4 | Mock conformance server covers batch + heartbeats + cancel-all; contract sim tests run in CI; docs drift test passes |
| 5 | Heartbeats + cursor pagination + builder trades shipped; `docs/V2-PARITY-MAP.md` exists and is accurate |

---

## 5. Risk & Mitigation

| Risk | Mitigation |
|------|------------|
| Promoting `internal/` → `pkg/` locks in APIs too early | Experimental packages with stability disclaimers; one-release shim layer |
| Authenticated E2E requires real funds | Mock server for CI; tiny real orders for manual verification only; gated by env var |
| Heartbeats interval unknown | Discover empirically; default to 30s with configurable override |
| GraphQL/subgraph schemas drift | Deferred; if pursued later, generate Go types from schema via `gqlgen` and pin versions |
| Neg-risk exchange address changes upstream | Dynamic lookup via `ClobMarketInfo` instead of hardcoded constant |

---

## 6. Priority Matrix (Within Each Track)

### Track 1 — Do first
- Delete `pkg/bookreader`
- Promote `pkg/bridge`, `pkg/ctf`, `pkg/wallet`
- Expand `public_sdk_boundary_test.go`

### Track 2 — Do first
- Structured logging on all protocol clients
- Per-market neg-risk exchange lookup
- Rate-limit enforcement

### Track 3 — Do first
- Coverage baseline measurement
- Golden vectors for neg-risk + batch
- Property-based tests for order validation

### Track 4 — Do first
- Mock conformance server expansion
- Contract simulation tests

### Track 5 — Do first
- Heartbeats (manual CLI/SDK)
- Cursor pagination
- Builder trades / scoring

---

## 7. Appendix: Decisions Log

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Parallel tracks, not sequential phases | Each vector can ship independently; reduces time-to-value |
| 2 | Staged SDK promotion (stable first, experimental for unstable) | Avoids locking in APIs that may change (orders, auth) |
| 3 | go-bot stays separate; no extraction needed | Already in its own repo (`polymarket-mega-bot`) |
| 4 | Structured logs (not full OTel) | Adds value to CLI + go-bot without tracing backend |
| 5 | Coverage gate at `max(current, 60%)` | Avoids gaming; sets realistic baseline |
| 6 | Mock conformance server + manual real orders | Zero-risk CI + operator acceptance gate |
| 7 | Heartbeats → cursor pagination → builder trades first | Highest effort/value ratio |
| 8 | Manual heartbeat first, standalone manager later | Simple, testable, matches existing patterns |

---

*This spec is approved. Proceeding to implementation plan.*
