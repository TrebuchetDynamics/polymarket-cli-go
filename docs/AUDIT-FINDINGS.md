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
| pkg/gamma re-export coverage | pkg/gamma/client.go vs internal/gamma/client.go | Resolved after Track 2: `pkg/gamma` exposes the 26 read-only Gamma methods and returns public `pkg/types` DTOs. |

## Track 3 — JSON envelope drift

(Populated by Track 3. Per-command list of where current `--json` output
diverges from the v1 envelope spec'd in
`docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § 3a.)

| Command | Current shape | Drift from v1 envelope |
|---|---|---|
| `auth` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json` entirely. v1 expects `error.code: USAGE_SUBCOMMAND_UNKNOWN` (or a help envelope) and a non-zero exit. |
| `auth status` | plain text `polygolem auth status: not implemented` on stdout; exit 0 | not-json; success-coded stub. v1 expects `error.code: INTERNAL_UNIMPLEMENTED`, exit 9, and an envelope. |
| `bridge` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. Same drift as `auth`. |
| `bridge assets` | bare object `{"supportedAssets":[...]}` | partial-success; payload is unwrapped. Missing `ok`, `version`, `meta`; should be wrapped in `data`. |
| `bridge deposit` | bare object `{"address":"...","note":"..."}` (positional `<address>`) | partial-success; payload is unwrapped. Auth/relayer failures not exercised here, but the success path already lacks the envelope. |
| `clob` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. |
| `clob balance` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, `category: auth`, exit 3. |
| `clob book` | stderr text `HTTP 404 .../book?token_id=...: {"error":"No orderbook exists..."}`; exit 1 (positional `<token-id>`) | error-not-json; expected `error.code: PROTOCOL_GAMMA_4XX` (or a `clob` analogue), exit 7, with upstream body in `error.details`. |
| `clob create-api-key` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `clob create-order` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `clob market` | stderr text `HTTP 404 .../markets/<id>: {"error":"market not found"}`; exit 1 (positional `<condition-id>`) | error-not-json; expected `error.code: PROTOCOL_UNEXPECTED_SHAPE` or 4xx-mapped code, exit 7. |
| `clob market-order` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `clob orders` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `clob price-history` | stderr text `HTTP 400 .../prices-history?...`; exit 1 (positional `<token-id>`) | error-not-json; expected `error.code: VALIDATION_TOKEN_ID_INVALID` or `PROTOCOL_*`, exit 4 or 7. |
| `clob tick-size` | stderr text `HTTP 404 .../tick-size?token_id=...: {"error":"market not found"}`; exit 1 (positional `<token-id>`) | error-not-json; expected `error.code` mapped from upstream 404, exit 7. |
| `clob trades` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `clob update-balance` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `deposit-wallet` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. |
| `deposit-wallet approve` | bare object `{"calls":[...],"note":"..."}` | partial-success; payload is unwrapped. Missing `ok`, `version`, `meta`; should be wrapped in `data`. |
| `deposit-wallet batch` | stderr text `builder credentials not configured: ...`; exit 1 | error-not-json; expected `error.code: AUTH_BUILDER_MISSING`, exit 3. |
| `deposit-wallet deploy` | stderr text `builder credentials not configured: ...`; exit 1 | error-not-json; expected `error.code: AUTH_BUILDER_MISSING`, exit 3. |
| `deposit-wallet derive` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `deposit-wallet fund` | stderr text `POLYMARKET_PRIVATE_KEY is required`; exit 1 | error-not-json; expected `error.code: AUTH_PRIVATE_KEY_MISSING`, exit 3. |
| `deposit-wallet nonce` | stderr text `builder credentials not configured: ...`; exit 1 | error-not-json; expected `error.code: AUTH_BUILDER_MISSING`, exit 3. |
| `deposit-wallet onboard` | stderr text `builder credentials not configured: ...`; exit 1 | error-not-json; expected `error.code: AUTH_BUILDER_MISSING`, exit 3. |
| `deposit-wallet status` | stderr text `builder credentials not configured: ...`; exit 1 | error-not-json; expected `error.code: AUTH_BUILDER_MISSING`, exit 3. |
| `discover` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. |
| `discover enrich` | stderr text `HTTP 404 .../markets/<id>: {"type":"not found error","error":"id not found"}`; exit 1 | error-not-json; expected `error.code: PROTOCOL_GAMMA_4XX`, exit 7, with upstream body in `error.details`. |
| `discover market` | stderr text `HTTP 422 .../markets/<slug>: {"type":"validation error",...}` (slug routed as id); exit 1 | error-not-json; also a code-side bug — `--slug` flag accepted but value is sent to `/markets/<id>` route. Expected `error.code: PROTOCOL_GAMMA_4XX` or `VALIDATION_MARKET_IDENTIFIER_AMBIGUOUS`, exit 7 or 4. |
| `discover search` | array of 1+ market objects at top level | partial-success; payload is a top-level JSON array. v1 requires the array under `data` with `ok`, `version`, `meta`. |
| `events` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. |
| `events list` | array of 20 event objects at top level | partial-success; payload is a top-level JSON array. Missing envelope; `--limit` flag also missing in code (silently ignored). |
| `health` | bare object `{"clob":"ok","gamma":"ok"}` | partial-success; missing `ok`, `version`, `meta`. Inner `clob`/`gamma` keys should sit under `data`. |
| `live` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. |
| `live status` | plain text `polygolem live status: not implemented` on stdout; exit 0 | not-json; success-coded stub. v1 expects `error.code: INTERNAL_UNIMPLEMENTED`, exit 9. |
| `orderbook` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. |
| `orderbook fee-rate` | stderr text `HTTP 404 .../fee-rate?token_id=...: {"error":"fee rate not found for market"}`; exit 1 | error-not-json; expected `error.code: PROTOCOL_GAMMA_4XX`, exit 7. |
| `orderbook get` | stderr text `HTTP 404 .../book?token_id=...: {"error":"No orderbook exists..."}`; exit 1 | error-not-json; expected `error.code: VALIDATION_TOKEN_ID_INVALID` (token resolves but book is closed) or 4xx-mapped, exit 4 or 7. |
| `orderbook last-trade` | bare object `{"price":"...","token_id":"..."}` | partial-success; payload is unwrapped. Missing `ok`, `version`, `meta`. |
| `orderbook midpoint` | stderr text `HTTP 404 .../midpoint?token_id=...: {"error":"No orderbook exists..."}`; exit 1 | error-not-json; expected `error.code: PROTOCOL_GAMMA_4XX`, exit 7. |
| `orderbook price` | stderr text `HTTP 404 .../price?token_id=...&side=BUY: {"error":"No orderbook exists..."}`; exit 1 | error-not-json; expected `error.code: PROTOCOL_GAMMA_4XX`, exit 7. |
| `orderbook spread` | stderr text `HTTP 404 .../spread?token_id=...: {"error":"No orderbook exists..."}`; exit 1 | error-not-json; expected `error.code: PROTOCOL_GAMMA_4XX`, exit 7. |
| `orderbook tick-size` | stderr text `HTTP 404 .../tick-size?token_id=...: {"error":"market not found"}`; exit 1 | error-not-json; expected `error.code: PROTOCOL_GAMMA_4XX`, exit 7. |
| `paper` | cobra group help text on stdout; exit 0 | not-json; group help bypasses `--json`. |
| `paper buy` | plain text `polygolem paper buy: not implemented` on stdout; exit 0 | not-json; success-coded stub. v1 expects `error.code: INTERNAL_UNIMPLEMENTED`, exit 9. |
| `paper positions` | plain text `polygolem paper positions: not implemented` on stdout; exit 0 | not-json; success-coded stub. v1 expects `error.code: INTERNAL_UNIMPLEMENTED`, exit 9. |
| `paper reset` | plain text `polygolem paper reset: not implemented` on stdout; exit 0 | not-json; success-coded stub. v1 expects `error.code: INTERNAL_UNIMPLEMENTED`, exit 9. |
| `paper sell` | plain text `polygolem paper sell: not implemented` on stdout; exit 0 | not-json; success-coded stub. v1 expects `error.code: INTERNAL_UNIMPLEMENTED`, exit 9. |
| `preflight` | bare object `{"checks":[...],"ok":true}` | partial-success; inner `ok` collides with envelope `ok`; missing `version`, `meta`; should be wrapped in `data`. |
| `version` | bare object `{"version":"dev"}` | partial-success; inner `version` collides with envelope `version` field name. Missing `ok`, `meta`; payload should sit under `data`. |

## Findings consumed downstream

(Filled in by each track as it consumes findings, so we know what's been
addressed.)

| Finding | Consumed by |
|---|---|
