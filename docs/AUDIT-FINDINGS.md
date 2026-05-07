# Audit Findings — Documentation Overhaul

**Status:** working notes for the documentation overhaul project. Populated
during Track 1, consumed by Tracks 2–4, deleted at the end of Track 5.

This file is **not user-facing**. It captures drift items that surfaced
while auditing existing documentation. Items marked `→ code` describe a
discrepancy between docs and source where the **source** was wrong (i.e.,
the doc is correct, the code drifted). Items marked `→ docs` describe the
opposite. Items marked `→ defer` are out of scope for this project.

## Track 1 — Doc-side drift

(Populated by Tasks 4–8. One row per finding.)

| Finding | Doc | Code reference | Direction | Resolution |
|---|---|---|---|---|
| ARCHITECTURE.md claimed no public SDK and "intentionally avoids `pkg/`" | docs/ARCHITECTURE.md | `pkg/{bookreader,bridge,gamma,marketresolver,pagination}` exist | → docs | Rewritten in Task 4 |
| ARCHITECTURE.md missed 13+ internal packages | docs/ARCHITECTURE.md | `internal/{auth,clob,dataapi,errors,execution,marketdiscovery,orders,polytypes,relayer,risk,rpc,stream,transport,wallet}` | → docs | Rewritten in Task 4 |
| COMMANDS.md missing 28 of 50 command paths from binary (auth, bridge, bridge assets, bridge deposit, clob, clob market, clob price-history, clob tick-size, deposit-wallet, deposit-wallet approve, deposit-wallet batch, deposit-wallet fund, deposit-wallet onboard, discover, discover enrich, discover market, discover search, events, events list, health, live, orderbook, orderbook fee-rate, orderbook last-trade, orderbook midpoint, orderbook price, orderbook spread, orderbook tick-size) | docs/COMMANDS.md | `polygolem --help` walked recursively shows 50 paths | → docs | Regenerated in Task 5 |
| PRD R8 (WebSocket And Streaming) requires authenticated user streams and RTDS isolation; only public `MarketClient` exists | docs/PRD.md | `internal/stream/client.go` exposes `MarketClient`/`SubscribeAssets` only; no user stream | → docs | Annotated ⚠️ in Task 6; current behavior covered by `docs/ARCHITECTURE.md` and `docs/COMMANDS.md` (`events list`) |
| PRD R12 (Public SDK Boundary) said "keep everything in `internal/`" and defer `pkg/`; five public packages now ship | docs/PRD.md | `pkg/{bookreader,bridge,gamma,marketresolver,pagination}` | → docs | Annotated ⚠️ in Task 6; current surface documented in `docs/ARCHITECTURE.md` |
| PRD R13 (Go-Bot Consumer Boundary) acceptance criteria reach into go-bot repo; not verifiable from polygolem | docs/PRD.md | Polygolem side ships required interfaces; go-bot adoption tracked elsewhere | → docs | Annotated ⚠️ in Task 6; cross-repo follow-up out of scope |
| README Packages table method counts are stale: `internal/clob` says "17 methods" (actual ~34 on `Client` in `internal/clob/client.go` plus 4 in `orders.go`); `internal/gamma` says "18 methods" (actual 27 on `Client` in `internal/gamma/client.go`) | README.md (Packages table) | `internal/clob/client.go`, `internal/clob/orders.go`, `internal/gamma/client.go` | → docs | Deferred in Task 8: choosing a canonical "method count" (Client receivers only vs all exported funcs) is ambiguous; not a smallest-edit fix. Restructure or drop the numbers in a later track. |
| README "Docs" list linked to `docs/IMPLEMENTATION-PLAN.md` which does not exist on disk | README.md | `docs/` directory listing | → docs | Fixed in Task 8 by removing the broken bullet (smallest edit). |

## Track 1 — Code-side drift surfaced incidentally

(Items where audit revealed a code bug. Not fixed in Track 1; logged for
later.)

| Finding | File:line | Notes |
|---|---|---|
| Prior COMMANDS.md documented `markets search` / `markets get` / `markets active` but no `markets` group exists in the binary | docs/COMMANDS.md (pre-Task 5) | Binary exposes market discovery via `discover search`, `discover market`, and `discover enrich` instead. Doc-side stale; no code change needed. |
| Prior COMMANDS.md documented `prices get` but no `prices` command exists in the binary | docs/COMMANDS.md (pre-Task 5) | Closest equivalent is `clob price-history` and `orderbook price`. Doc-side stale; no code change needed. |
| tests/docs_safety_test.go pinned old ARCHITECTURE/COMMANDS phrases and old REFERENCE-RUST-CLI path | tests/docs_safety_test.go:17,60,77-83 | Pins updated to track current canonical wording; COMMANDS.md got back its Automation section (set -euo pipefail, jq examples). Closed in Track 1 fixup. |
| pkg/gamma re-exports only 11 of internal/gamma.Client's methods | pkg/gamma/client.go vs internal/gamma/client.go | Surfaced during Track 2; SDK consumers cannot reach EventBySlug, profile lookups, etc. through the public package. |

## Track 3 — JSON envelope drift

(Populated by Track 3. Per-command list of where current `--json` output
diverges from the v1 envelope spec'd in
`docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § 3a.)

| Command | Current shape | Drift from v1 envelope |
|---|---|---|

## Findings consumed downstream

(Filled in by each track as it consumes findings, so we know what's been
addressed.)

| Finding | Consumed by |
|---|---|
