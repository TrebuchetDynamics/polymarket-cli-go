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

## Track 1 — Code-side drift surfaced incidentally

(Items where audit revealed a code bug. Not fixed in Track 1; logged for
later.)

| Finding | File:line | Notes |
|---|---|---|
| Prior COMMANDS.md documented `markets search` / `markets get` / `markets active` but no `markets` group exists in the binary | docs/COMMANDS.md (pre-Task 5) | Binary exposes market discovery via `discover search`, `discover market`, and `discover enrich` instead. Doc-side stale; no code change needed. |
| Prior COMMANDS.md documented `prices get` but no `prices` command exists in the binary | docs/COMMANDS.md (pre-Task 5) | Closest equivalent is `clob price-history` and `orderbook price`. Doc-side stale; no code change needed. |

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
