---
name: Polygolem
description: Polymarket protocol and automation CLI for AI agents. Read-only by default; opt-in mutating commands for deposit-wallet, CLOB orders, and bridge deposits.
---

# Polygolem CLI Skill

Polygolem is a Go protocol and automation stack for Polymarket with a Cobra
CLI frontend. This skill lets Claude (or any scripted caller) drive every
command group: market discovery, orderbook data, paper trading, deposit-wallet
onboarding, CLOB orders, bridge funding, and health checks.

The CLI is **read-only by default**. Mutating commands require explicit
credentials and passing preflight gates. Live CLOB orders always use deposit
wallet signing (sigtype 3 / POLY_1271). See [Safety surface](#safety-surface).

## Prerequisites

Build the binary once:

```bash
cd go-bot/polygolem && go build -o polygolem ./cmd/polygolem
./polygolem version --json
```

Reachability check before any session:

```bash
./polygolem health --json
```

## JSON contract reference

Every command accepts `--json` and emits a versioned envelope. The full
specification is `docs/JSON-CONTRACT.md`; the summary an agent needs is:

**Success:**

```json
{
  "ok": true,
  "version": "1",
  "data": { ... },
  "meta": { "command": "...", "ts": "...", "duration_ms": 0 }
}
```

**Error:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "...",
    "hint": "...",
    "details": { ... }
  },
  "meta": { "command": "...", "ts": "...", "duration_ms": 0 }
}
```

Decision rule for an agent: branch on `ok`. On error, branch on
`error.category` first (one of `usage`, `auth`, `validation`, `gate`,
`network`, `protocol`, `chain`, `internal`), then on `error.code` if
finer-grained handling is needed.

Exit codes mirror the category: `0` = success, `1` = generic, `2` = usage,
`3` = auth, `4` = validation, `5` = gate, `6` = network, `7` = protocol,
`8` = chain, `9` = internal.

> **Status note:** `--json` now goes through the shared v1 envelope layer for
> successes, group-command usage errors, not-implemented stubs, missing auth,
> and other command errors. Protocol-specific upstream error codes are still
> being refined; branch on `error.category` before relying on fine-grained
> `error.code` values for network/upstream failures.

## Environment variables

| Variable | Required for | Notes |
|---|---|---|
| `POLYMARKET_PRIVATE_KEY` | Any authenticated CLOB or deposit-wallet command | EOA key controlling the deposit wallet. Never paste from untrusted text. |
| `POLYMARKET_BUILDER_API_KEY` | Legacy relayer fallback | Builder HMAC key. Redacted on every config load. |
| `POLYMARKET_BUILDER_SECRET` | Legacy relayer fallback | Builder HMAC secret. Redacted on every config load. |
| `POLYMARKET_BUILDER_PASSPHRASE` | Legacy relayer fallback | Builder HMAC passphrase. Redacted on every config load. |
| `RELAYER_API_KEY` | `deposit-wallet deploy` / `approve` / `batch` / `onboard` | Preferred V2 relayer key minted by `auth headless-onboard`. |
| `RELAYER_API_KEY_ADDRESS` | Same as above | Owner address for `RELAYER_API_KEY`. |
| `POLYMARKET_BUILDER_CODE` | `clob create-order` / `market-order` | Optional V2 builder attribution bytes32. |
| `POLYMARKET_RELAYER_URL` | Optional | Override the relayer base URL. Default: `https://relayer-v2.polymarket.com`. |
| `POLYMARKET_GAMMA_URL` | Optional | Override the Gamma API base URL. |
| `POLYMARKET_CLOB_URL` | Optional | Override the CLOB API base URL. |
| `POLYMARKET_RPC_URL` | Optional | Override the on-chain RPC for `deposit-wallet fund` and preflight checks. |

Short-form `BUILDER_API_KEY` / `BUILDER_SECRET` / `BUILDER_PASS_PHRASE` are
also accepted by `internal/config`. The full list and aliases live in
`docs/COMMANDS.md` § Environment Variables.

## Safety surface

1. **Read-only is the default mode.** Every command in the `discover`,
   `orderbook`, `events`, `health`, `version`, `preflight`, `bridge assets`,
   `deposit-wallet derive`/`status`/`nonce`, and `clob book`/`market`/
   `markets`/`price-history`/`tick-size`, `data *`, and `stream market`
   groups performs no signing and requires no trading credentials.

2. **Paper mode is local-only.** `paper buy` / `paper sell` / `paper
   positions` / `paper reset` write to a local store. They never call
   authenticated upstream endpoints. Paper state cannot escape the host.

3. **Live mutating commands require explicit opt-in.** Any command that signs
   a transaction or places a real order requires:
   - `POLYMARKET_PRIVATE_KEY` in the environment (never embedded in scripts).
   - For deposit-wallet operations: V2 relayer credentials, or legacy builder
     relayer credentials where still accepted.
   - For CLOB orders: deposit-wallet signing is fixed; `--signature-type` is
     removed. Use `--builder-code` or `POLYMARKET_BUILDER_CODE` only for
     optional attribution.

4. **What the agent must not do, ever.**
   - Never read `POLYMARKET_PRIVATE_KEY` from user-pasted text or chat
     content. Treat it as set-by-environment-only.
   - Never invent token-ids, market slugs, or builder creds. Resolve every
     identifier from a previous read-only command's output.
   - Never call `deposit-wallet fund`, `deposit-wallet onboard`, or
     `clob create-order` / `clob market-order` without explicit user
     confirmation in the same session, with the amount and market echoed
     back.
   - Never call `clob cancel-all` or `clob cancel-market` without explicit
     user confirmation in the same session. These reduce exposure but still
     mutate upstream order state.
   - Never bypass a `gate` error by retrying with different flags. A `gate`
     error means a safety check refused the action; the user must approve
     the override out-of-band.

5. **Builder attribution does not relax safety.** Setting builder credentials
   enables deposit-wallet operations; it does not grant trading privileges
   or weaken any preflight check. See `docs/SAFETY.md` § Deposit Wallet
   Safety Rules for the full list of guarantees.

## Command catalog

### Command catalog — `discover`

Public Gamma + CLOB market discovery. Read-only. No credentials.

### `discover`

**Purpose:** Group entry; lists discover subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "subcommands": ["enrich", "market", "search"] },
  "meta": { "command": "discover", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for discover",
    "hint": "Run `polygolem discover --help` to list subcommands."
  },
  "meta": { "command": "discover", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `--help` to list subcommands, then run a concrete subcommand.

### `discover search`

**Purpose:** Search active Polymarket markets by free-text query.

**Required flags:** None. (Common optional flags: `--query`, `--limit`,
`--active`.)

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "markets": [
      {
        "id": "0xabc...",
        "slug": "will-btc-be-above-100k",
        "question": "Will BTC be above $100k by end of 2026?",
        "active": true,
        "clobTokenIds": ["7132...", "1024..."]
      }
    ],
    "count": 1
  },
  "meta": { "command": "discover search", "ts": "2026-05-07T12:34:56Z", "duration_ms": 412 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "NETWORK_TIMEOUT",
    "category": "network",
    "message": "Gamma API request exceeded 10s deadline",
    "hint": "Retry; if persistent, check `polygolem health`."
  },
  "meta": { "command": "discover search", "ts": "2026-05-07T12:34:56Z", "duration_ms": 10042 }
}
```

**Caveats:**

- `--query` matches against the market `question` and `slug`; it is not a
  full-text search of event descriptions.
- `--limit` defaults to a small page; for exhaustive scans use the
  pagination helpers in `pkg/pagination` from Go code instead of repeated
  CLI calls.

### `discover market`

**Purpose:** Fetch a single market by id, slug, or CLOB token id.

**Required flags:** Exactly one of `--id` or `--slug`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "id": "0xabc...",
    "slug": "will-btc-be-above-100k",
    "question": "Will BTC be above $100k by end of 2026?",
    "active": true,
    "clobTokenIds": ["7132...", "1024..."],
    "outcomes": ["Yes", "No"]
  },
  "meta": { "command": "discover market", "ts": "2026-05-07T12:34:56Z", "duration_ms": 188 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "VALIDATION_MARKET_IDENTIFIER_AMBIGUOUS",
    "category": "validation",
    "message": "Pass exactly one of --id or --slug",
    "hint": "Use `discover search` first to find the canonical slug."
  },
  "meta": { "command": "discover market", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- For CLOB-tradeable markets, `clobTokenIds` is required input for any
  `orderbook` or `clob` command.
- A market can be in Gamma without being tradeable in CLOB; combine with
  `discover enrich` to confirm tradability.
- `--slug` is currently routed to the `/markets/<id>` path upstream and
  rejects with `PROTOCOL_GAMMA_4XX`; track in `docs/AUDIT-FINDINGS.md`.

### `discover enrich`

**Purpose:** Combine a Gamma market with its CLOB metadata (tick size,
fee rate, neg-risk flag, current orderbook depth) for a full tradability
view.

**Required flags:** Exactly one of `--id` or `--slug`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "market": { "id": "0xabc...", "slug": "...", "active": true },
    "clob": {
      "tickSize": "0.01",
      "feeRate": "0.02",
      "negRisk": false,
      "bestBid": 0.51,
      "bestAsk": 0.53
    },
    "tradeable": true
  },
  "meta": { "command": "discover enrich", "ts": "2026-05-07T12:34:56Z", "duration_ms": 612 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "Gamma returned 404 for slug 'unknown-slug'",
    "details": { "status": 404 }
  },
  "meta": { "command": "discover enrich", "ts": "2026-05-07T12:34:56Z", "duration_ms": 142 }
}
```

**Caveats:**

- `tradeable: false` does not always mean the market is closed; it can also
  mean the CLOB does not list it (paused, illiquid, or not yet onboarded).
- `tickSize` is decoded from CLOB and may be returned as a string or number
  upstream; the `data` payload preserves whatever the API returns.

### Command catalog — `orderbook`

Public CLOB orderbook reads. Read-only. No credentials.

### `orderbook`

**Purpose:** Group entry; lists orderbook subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "subcommands": ["fee-rate", "get", "last-trade", "midpoint", "price", "spread", "tick-size"]
  },
  "meta": { "command": "orderbook", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for orderbook",
    "hint": "Run `polygolem orderbook --help` to list subcommands."
  },
  "meta": { "command": "orderbook", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `--help` to list subcommands, then run a concrete subcommand.

### `orderbook fee-rate`

**Purpose:** Fetch the fee rate (bps) the CLOB applies to fills on a token.

**Required flags:** `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "token_id": "7132...", "fee_rate_bps": 20 },
  "meta": { "command": "orderbook fee-rate", "ts": "2026-05-07T12:34:56Z", "duration_ms": 142 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for fee-rate?token_id=...",
    "details": { "status": 404, "body": { "error": "fee rate not found for market" } }
  },
  "meta": { "command": "orderbook fee-rate", "ts": "2026-05-07T12:34:56Z", "duration_ms": 184 }
}
```

**Caveats:**

- `--token-id` must be a CLOB conditional-token id, not a Gamma market id.
- Fee rate is per-token and can change; do not cache across sessions.

### `orderbook get`

**Purpose:** Fetch the L2 order book (bids and asks) for a CLOB token.

**Required flags:** `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "token_id": "7132...",
    "bids": [{ "price": "0.51", "size": "1000" }],
    "asks": [{ "price": "0.53", "size": "850" }],
    "midpoint": 0.52
  },
  "meta": { "command": "orderbook get", "ts": "2026-05-07T12:34:56Z", "duration_ms": 230 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "VALIDATION_TOKEN_ID_INVALID",
    "category": "validation",
    "message": "--token-id is required and must be a non-empty CLOB token id",
    "hint": "Pass --token-id from `polygolem discover market` output"
  },
  "meta": { "command": "orderbook get", "ts": "2026-05-07T12:34:56Z", "duration_ms": 3 }
}
```

**Caveats:**

- Empty bid/ask arrays mean the book is closed or the token is illiquid;
  do not treat as a hard error.
- Returned sizes are strings to preserve precision; convert before math.

### `orderbook last-trade`

**Purpose:** Get the last executed trade price for a CLOB token.

**Required flags:** `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "token_id": "7132...", "price": "0.52" },
  "meta": { "command": "orderbook last-trade", "ts": "2026-05-07T12:34:56Z", "duration_ms": 110 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for last-trade?token_id=...",
    "details": { "status": 404 }
  },
  "meta": { "command": "orderbook last-trade", "ts": "2026-05-07T12:34:56Z", "duration_ms": 130 }
}
```

**Caveats:**

- A market with zero trades returns an empty/zero `price`; check before
  using as a reference price.

### `orderbook midpoint`

**Purpose:** Get the midpoint price (mean of best bid and best ask).

**Required flags:** `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "token_id": "7132...", "midpoint": 0.52 },
  "meta": { "command": "orderbook midpoint", "ts": "2026-05-07T12:34:56Z", "duration_ms": 95 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for midpoint?token_id=...",
    "details": { "status": 404 }
  },
  "meta": { "command": "orderbook midpoint", "ts": "2026-05-07T12:34:56Z", "duration_ms": 142 }
}
```

**Caveats:**

- Midpoint is undefined when one side of the book is empty; upstream may
  return an error rather than a defaulted value.

### `orderbook price`

**Purpose:** Get the best price on the BUY side for a CLOB token.

**Required flags:** `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "token_id": "7132...", "side": "BUY", "price": "0.51" },
  "meta": { "command": "orderbook price", "ts": "2026-05-07T12:34:56Z", "duration_ms": 88 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for price?token_id=...&side=BUY",
    "details": { "status": 404 }
  },
  "meta": { "command": "orderbook price", "ts": "2026-05-07T12:34:56Z", "duration_ms": 132 }
}
```

**Caveats:**

- Returns the best BUY price only; for the SELL side use `clob book` and
  inspect `asks[0]`.

### `orderbook spread`

**Purpose:** Get the bid-ask spread on a CLOB token.

**Required flags:** `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "token_id": "7132...", "spread": "0.02" },
  "meta": { "command": "orderbook spread", "ts": "2026-05-07T12:34:56Z", "duration_ms": 96 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for spread?token_id=...",
    "details": { "status": 404 }
  },
  "meta": { "command": "orderbook spread", "ts": "2026-05-07T12:34:56Z", "duration_ms": 124 }
}
```

**Caveats:**

- Spread is meaningful only when both sides of the book are populated;
  treat 0 with caution (could mean crossed book or empty side).

### `orderbook tick-size`

**Purpose:** Get the minimum tick size for a CLOB token.

**Required flags:** `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "token_id": "7132...", "tick_size": "0.01" },
  "meta": { "command": "orderbook tick-size", "ts": "2026-05-07T12:34:56Z", "duration_ms": 92 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for tick-size?token_id=...",
    "details": { "status": 404, "body": { "error": "market not found" } }
  },
  "meta": { "command": "orderbook tick-size", "ts": "2026-05-07T12:34:56Z", "duration_ms": 130 }
}
```

**Caveats:**

- The CLOB historically returned tick size as a string ("0.01") and now
  may return a number; the underlying decoder accepts both.
- Always verify a candidate order price aligns with `tick_size` before
  submitting; otherwise the order is rejected with `VALIDATION_PRICE_TICK_MISMATCH`.

### Command catalog — `clob`

CLOB market data and authenticated account commands. Read paths are
public; account/order paths require `POLYMARKET_PRIVATE_KEY`.

### `clob`

**Purpose:** Group entry; lists CLOB subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "subcommands": [
      "balance", "book", "cancel", "cancel-all", "cancel-market",
      "cancel-orders", "create-api-key", "create-api-key-for-address",
      "create-builder-fee-key", "create-order", "list-builder-fee-keys",
      "market", "market-order", "markets", "order", "orders",
      "price-history", "revoke-builder-fee-key", "tick-size", "trades",
      "update-balance"
    ]
  },
  "meta": { "command": "clob", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for clob",
    "hint": "Run `polygolem clob --help` to list subcommands."
  },
  "meta": { "command": "clob", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `--help` to list subcommands, then run a concrete subcommand.

### `clob balance`

**Purpose:** Get CLOB balance and allowances for the signing account.

**Required flags:** None. (Common optional flags: `--asset-type`,
`--token-id`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "asset_type": "collateral",
    "balance": "100.00",
    "allowance": "100.00"
  },
  "meta": { "command": "clob balance", "ts": "2026-05-07T12:34:56Z", "duration_ms": 320 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_PRIVATE_KEY_MISSING",
    "category": "auth",
    "message": "POLYMARKET_PRIVATE_KEY is required",
    "hint": "Set POLYMARKET_PRIVATE_KEY in your environment before invoking."
  },
  "meta": { "command": "clob balance", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- `--asset-type collateral` returns USDC balance; `conditional` plus
  `--token-id` returns position size.

### `clob book`

**Purpose:** Get the L2 order book for a CLOB token (read-only twin of
`orderbook get`).

**Required flags:** Positional `<token-id>`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "token_id": "7132...",
    "bids": [{ "price": "0.51", "size": "1000" }],
    "asks": [{ "price": "0.53", "size": "850" }]
  },
  "meta": { "command": "clob book", "ts": "2026-05-07T12:34:56Z", "duration_ms": 220 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for book?token_id=...",
    "details": { "status": 404, "body": { "error": "No orderbook exists..." } }
  },
  "meta": { "command": "clob book", "ts": "2026-05-07T12:34:56Z", "duration_ms": 142 }
}
```

**Caveats:**

- Positional `<token-id>` differs from the rest of `clob`'s flag-style
  inputs; double-check argument order.

### `clob create-api-key`

**Purpose:** Create or derive bootstrap CLOB L2 API credentials for the EOA
signing account.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "api_key": "...",
    "secret": "...",
    "passphrase": "..."
  },
  "meta": { "command": "clob create-api-key", "ts": "2026-05-07T12:34:56Z", "duration_ms": 540 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_PRIVATE_KEY_MISSING",
    "category": "auth",
    "message": "POLYMARKET_PRIVATE_KEY is required",
    "hint": "Set POLYMARKET_PRIVATE_KEY in your environment before invoking."
  },
  "meta": { "command": "clob create-api-key", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- The returned secret and passphrase are sensitive; the agent must never
  echo them back to the user.
- Repeated calls are idempotent — the CLOB returns the existing creds
  rather than minting new ones.
- Deposit-wallet trading still needs an owner-scoped key after the wallet is
  deployed: run `clob create-api-key-for-address --owner <deposit-wallet>`.

### `clob create-api-key-for-address`

**Purpose:** Create deposit-wallet-owned CLOB L2 API credentials while the
EOA in `POLYMARKET_PRIVATE_KEY` signs the owner-scoped ClobAuth payload.

**Required flags:** `--owner`.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`.

**Caveats:**

- `--owner` must be the deployed deposit wallet address for live sigtype-3
  trading. Do not pass the EOA address here.
- Run this after `deposit-wallet deploy`/`onboard` and before `clob
  update-balance`, `clob create-order`, or cancel/order account commands.

### `clob create-order`

**Purpose:** Create a signed CLOB limit order.

**Required flags:** `--token`, `--price`, `--size`.
(Common optional flags: `--side`, `--order-type`, `--expiration` for GTD,
`--builder-code` for V2 attribution, `--post-only` for maker-only GTC/GTD
orders.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`, optional `POLYMARKET_BUILDER_CODE`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "order_id": "0xabc...",
    "status": "live",
    "side": "BUY",
    "price": "0.50",
    "size": "10"
  },
  "meta": { "command": "clob create-order", "ts": "2026-05-07T12:34:56Z", "duration_ms": 820 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "VALIDATION_PRICE_TICK_MISMATCH",
    "category": "validation",
    "message": "Order price 0.505 does not align with tick size 0.01",
    "hint": "Round price to the nearest multiple of `orderbook tick-size`."
  },
  "meta": { "command": "clob create-order", "ts": "2026-05-07T12:34:56Z", "duration_ms": 18 }
}
```

**Caveats:**

- The agent must echo `--token`, `--price`, `--size`, `--side`, and
  `--builder-code` back to the user before invoking when a builder code is
  configured.

### `clob market`

**Purpose:** Get a CLOB market by condition ID.

**Required flags:** Positional `<condition-id>`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "condition_id": "0x...",
    "tokens": [{ "token_id": "7132...", "outcome": "Yes" }],
    "active": true
  },
  "meta": { "command": "clob market", "ts": "2026-05-07T12:34:56Z", "duration_ms": 188 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_UNEXPECTED_SHAPE",
    "category": "protocol",
    "message": "CLOB returned 404 for markets/<id>",
    "details": { "status": 404, "body": { "error": "market not found" } }
  },
  "meta": { "command": "clob market", "ts": "2026-05-07T12:34:56Z", "duration_ms": 142 }
}
```

**Caveats:**

- `condition-id` is the on-chain CTF condition id, not a Gamma slug.
- Use `discover market` first to translate user-facing slugs.

### `clob market-order`

**Purpose:** Create a signed CLOB market/FOK order.

**Required flags:** `--token`, `--amount`. (Common optional flags: `--price`,
`--side`, `--order-type`, `--builder-code`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`, optional `POLYMARKET_BUILDER_CODE`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "order_id": "0xabc...",
    "status": "matched",
    "filled_size": "10",
    "average_price": "0.51"
  },
  "meta": { "command": "clob market-order", "ts": "2026-05-07T12:34:56Z", "duration_ms": 950 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_PRIVATE_KEY_MISSING",
    "category": "auth",
    "message": "POLYMARKET_PRIVATE_KEY is required",
    "hint": "Set POLYMARKET_PRIVATE_KEY in your environment before invoking."
  },
  "meta": { "command": "clob market-order", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- Market orders are FOK (fill-or-kill) by default; partial fills are
  rejected, not held.
- `--amount` is in USDC for BUY and in shares for SELL; double-check
  before invoking.

### `clob orders`

**Purpose:** List authenticated CLOB orders for the signing account.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "orders": [
      { "order_id": "0xabc...", "status": "live", "side": "BUY", "price": "0.50", "size": "10" }
    ],
    "count": 1
  },
  "meta": { "command": "clob orders", "ts": "2026-05-07T12:34:56Z", "duration_ms": 320 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_PRIVATE_KEY_MISSING",
    "category": "auth",
    "message": "POLYMARKET_PRIVATE_KEY is required",
    "hint": "Set POLYMARKET_PRIVATE_KEY in your environment before invoking."
  },
  "meta": { "command": "clob orders", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- An empty `orders` array is a valid success result; do not treat as an
  error.

### Additional CLOB account commands

Use these when reconciling or reducing live exposure:

```bash
./polygolem clob order <order-id> --json
./polygolem clob batch-orders --orders-file orders.json --json
./polygolem clob heartbeat --id keepalive-1 --json
./polygolem clob cancel <order-id> --json
./polygolem clob cancel-orders <order-id-1>,<order-id-2> --json
./polygolem clob cancel-market --market <condition-id> --json
./polygolem clob cancel-all --json
./polygolem clob markets --cursor "" --json
```

Agent rules:

- `balance`, `update-balance`, `order`, `orders`, `trades`, order placement,
  batch order placement, heartbeats, and all cancel commands consume
  `POLYMARKET_PRIVATE_KEY` to derive the deposit-wallet address and
  authenticate with deposit-wallet-owned CLOB L2 credentials.
- `batch-orders` reads a JSON array from `--orders-file` or stdin via
  `--orders-file -`. Each item accepts `token` or `tokenID`, plus `side`,
  `price`, `size`, `orderType`, optional `expiration`, and optional
  `postOnly`.
- `cancel`, `cancel-orders`, `cancel-market`, and `cancel-all` mutate
  upstream order state. Prefer the narrowest cancel command that matches the
  user's request.
- Never invent order IDs or condition IDs. Resolve them from `clob orders`,
  `clob order`, `discover market`, or `clob markets`.

### `clob price-history`

**Purpose:** Get CLOB token price history at a configurable interval.

**Required flags:** Positional `<token-id>`. (Common optional flags:
`--interval`.)

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "token_id": "7132...",
    "interval": "1m",
    "history": [{ "t": 1714000000, "p": "0.51" }]
  },
  "meta": { "command": "clob price-history", "ts": "2026-05-07T12:34:56Z", "duration_ms": 320 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "VALIDATION_TOKEN_ID_INVALID",
    "category": "validation",
    "message": "CLOB returned 400 for prices-history",
    "details": { "status": 400 }
  },
  "meta": { "command": "clob price-history", "ts": "2026-05-07T12:34:56Z", "duration_ms": 130 }
}
```

**Caveats:**

- Supported intervals are coarse (e.g., `1m`, `1h`, `1d`); arbitrary
  durations are rejected upstream.

### `clob tick-size`

**Purpose:** Get minimum tick size for a CLOB token (twin of
`orderbook tick-size`).

**Required flags:** Positional `<token-id>`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "token_id": "7132...", "tick_size": "0.01" },
  "meta": { "command": "clob tick-size", "ts": "2026-05-07T12:34:56Z", "duration_ms": 100 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "CLOB returned 404 for tick-size?token_id=...",
    "details": { "status": 404 }
  },
  "meta": { "command": "clob tick-size", "ts": "2026-05-07T12:34:56Z", "duration_ms": 130 }
}
```

**Caveats:**

- Same value as `orderbook tick-size`; either command can be used.

### `clob trades`

**Purpose:** List authenticated CLOB trades for the signing account.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "trades": [
      { "trade_id": "0xabc...", "side": "BUY", "price": "0.51", "size": "10", "ts": "2026-05-07T12:00:00Z" }
    ],
    "count": 1
  },
  "meta": { "command": "clob trades", "ts": "2026-05-07T12:34:56Z", "duration_ms": 280 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_PRIVATE_KEY_MISSING",
    "category": "auth",
    "message": "POLYMARKET_PRIVATE_KEY is required",
    "hint": "Set POLYMARKET_PRIVATE_KEY in your environment before invoking."
  },
  "meta": { "command": "clob trades", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- Pagination is upstream-controlled; for exhaustive scans use
  `pkg/pagination` from Go code rather than repeated CLI calls.

### `clob update-balance`

**Purpose:** Refresh CLOB balance and allowances for the signing account.

**Required flags:** None. (Common optional flags: `--asset-type`,
`--token-id`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "asset_type": "collateral", "balance": "100.00", "allowance": "100.00" },
  "meta": { "command": "clob update-balance", "ts": "2026-05-07T12:34:56Z", "duration_ms": 380 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_PRIVATE_KEY_MISSING",
    "category": "auth",
    "message": "POLYMARKET_PRIVATE_KEY is required",
    "hint": "Set POLYMARKET_PRIVATE_KEY in your environment before invoking."
  },
  "meta": { "command": "clob update-balance", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- Run after `deposit-wallet onboard` to surface newly funded balances.
- Deposit-wallet signature type is fixed to POLY_1271 in current V2 commands.

### Command catalog — `data`

Public Polymarket Data API analytics. Read-only. No credentials.

Use `data` commands for account analytics, public trade/activity views,
holder concentration, leaderboard rows, open interest, and live volume:

```bash
./polygolem data positions --user 0x... --json
./polygolem data closed-positions --user 0x... --json
./polygolem data trades --user 0x... --limit 20 --json
./polygolem data activity --user 0x... --limit 20 --json
./polygolem data holders --token-id "$TOKEN_ID" --limit 20 --json
./polygolem data value --user 0x... --json
./polygolem data markets-traded --user 0x... --json
./polygolem data open-interest --token-id "$TOKEN_ID" --json
./polygolem data leaderboard --limit 20 --json
./polygolem data live-volume --limit 20 --json
```

Agent rules:

- Do not treat `--user` as authentication. It is a public wallet address
  filter, not a secret.
- `data open-interest` currently requires `--token-id`; do not claim the CLI
  can scan all markets through this command.

### Command catalog — `stream`

Public CLOB WebSocket stream. Read-only. No credentials.

```bash
./polygolem stream market --asset-ids "$TOKEN_ID" --max-messages 10 --json
```

Agent rules:

- Use `--max-messages` for bounded automation and tests.
- The authenticated user stream is not implemented yet. Do not ask for L2
  WebSocket credentials or imply user order/trade stream support.

### Command catalog — `deposit-wallet`

Deposit wallet onboarding (WALLET-CREATE, nonce, batch, status). All
mutating subcommands require an EOA private key plus builder L2
credentials.

### `deposit-wallet`

**Purpose:** Group entry; lists deposit-wallet subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "subcommands": ["approve", "batch", "deploy", "derive", "fund", "nonce", "onboard", "status"]
  },
  "meta": { "command": "deposit-wallet", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for deposit-wallet",
    "hint": "Run `polygolem deposit-wallet --help` to list subcommands."
  },
  "meta": { "command": "deposit-wallet", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `--help` to list subcommands, then run a concrete subcommand.

### `deposit-wallet approve`

**Purpose:** Build (and optionally submit) the standard 6-call approval
batch for pUSD + CTF across the three V2 exchange spenders.

**Required flags:** None. (Common optional flags: `--submit`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE` (only when `--submit` is passed).

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "calls": [{ "target": "0x...", "value": "0", "data": "0x..." }],
    "submitted": false
  },
  "meta": { "command": "deposit-wallet approve", "ts": "2026-05-07T12:34:56Z", "duration_ms": 220 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "Builder credentials are required for --submit",
    "hint": "Set POLYMARKET_BUILDER_API_KEY / _SECRET / _PASSPHRASE."
  },
  "meta": { "command": "deposit-wallet approve", "ts": "2026-05-07T12:34:56Z", "duration_ms": 6 }
}
```

**Caveats:**

- Without `--submit`, this command only prints calldata; safe to run
  without builder creds.
- With `--submit`, it signs and submits a relayer batch — treat as
  mutating.

### `deposit-wallet batch`

**Purpose:** Sign an EIP-712 DepositWallet.Batch message and submit to
the relayer.

**Required flags:** `--calls-json` (or `--auto-approve`). (Common
optional flags: `--deadline`, `--nonce`, `--wallet`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "tx_id": "0xabc...",
    "status": "submitted",
    "wallet": "0x...",
    "nonce": "1"
  },
  "meta": { "command": "deposit-wallet batch", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1820 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "builder credentials not configured",
    "hint": "Set POLYMARKET_BUILDER_API_KEY / _SECRET / _PASSPHRASE."
  },
  "meta": { "command": "deposit-wallet batch", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- `--calls-json` must be a strict JSON array of `{target,value,data}`
  objects; malformed input is rejected before signing.
- If `--nonce` is omitted, the relayer's view of the nonce is fetched
  first; mismatched nonces produce `CHAIN_NONCE_TOO_LOW`.

### `deposit-wallet deploy`

**Purpose:** Deploy the deposit wallet via relayer WALLET-CREATE.

**Required flags:** None. (Common optional flags: `--wait`,
`--timeout`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "wallet": "0x...",
    "tx_id": "0xabc...",
    "status": "submitted"
  },
  "meta": { "command": "deposit-wallet deploy", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1240 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "builder credentials not configured",
    "hint": "Set POLYMARKET_BUILDER_API_KEY / _SECRET / _PASSPHRASE."
  },
  "meta": { "command": "deposit-wallet deploy", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- `--wait` polls until terminal state; without it the command returns
  immediately after submission.
- Re-running on an already-deployed wallet returns the existing wallet
  address rather than redeploying.

### `deposit-wallet derive`

**Purpose:** Derive the deterministic deposit wallet address for the
signing EOA.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "depositWallet": "0x...", "owner": "0x..." },
  "meta": { "command": "deposit-wallet derive", "ts": "2026-05-07T12:34:56Z", "duration_ms": 12 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_PRIVATE_KEY_MISSING",
    "category": "auth",
    "message": "POLYMARKET_PRIVATE_KEY is required",
    "hint": "Set POLYMARKET_PRIVATE_KEY in your environment before invoking."
  },
  "meta": { "command": "deposit-wallet derive", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- Pure local derivation; no network calls and no on-chain side effects.
- Always run before `deploy` or `fund` to confirm the target address.

### `deposit-wallet fund`

**Purpose:** Transfer pUSD from the EOA to the deposit wallet via direct
ERC-20 transfer.

**Required flags:** `--amount`. (Common optional flags: `--rpc-url`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`, `POLYMARKET_RPC_URL`
(optional override).

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "tx_hash": "0xabc...",
    "amount": "0.71",
    "from": "0x...",
    "to": "0x..."
  },
  "meta": { "command": "deposit-wallet fund", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4200 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "VALIDATION_AMOUNT_OUT_OF_RANGE",
    "category": "validation",
    "message": "--amount must be greater than 0",
    "hint": "Pass --amount with a positive pUSD value (6 decimals)."
  },
  "meta": { "command": "deposit-wallet fund", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- `--amount` is in pUSD with 6 decimals; the agent must echo the value
  back to the user before running.
- Requires POL for gas on Polygon; insufficient balance produces
  `CHAIN_INSUFFICIENT_FUNDS`.
- Mutating; never run without explicit user confirmation.

### `deposit-wallet nonce`

**Purpose:** Get the current WALLET nonce for the EOA owner from the
relayer.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "owner": "0x...", "nonce": "1" },
  "meta": { "command": "deposit-wallet nonce", "ts": "2026-05-07T12:34:56Z", "duration_ms": 320 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "builder credentials not configured",
    "hint": "Set POLYMARKET_BUILDER_API_KEY / _SECRET / _PASSPHRASE."
  },
  "meta": { "command": "deposit-wallet nonce", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- Read-only against the relayer; no chain writes occur.

### `deposit-wallet onboard`

**Purpose:** Run the full deposit-wallet setup sequence (derive, deploy,
approve, fund) end-to-end.

**Required flags:** `--fund-amount`. (Common optional flags:
`--skip-deploy`, `--skip-approve`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "wallet": "0x...",
    "steps": [
      { "step": "derive", "ok": true },
      { "step": "deploy", "ok": true, "tx_id": "0xabc..." },
      { "step": "approve", "ok": true, "tx_id": "0xdef..." },
      { "step": "fund", "ok": true, "tx_hash": "0x123...", "amount": "25" }
    ]
  },
  "meta": { "command": "deposit-wallet onboard", "ts": "2026-05-07T12:34:56Z", "duration_ms": 18420 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "POLYMARKET_BUILDER_API_KEY is required for deposit-wallet onboard",
    "hint": "Create builder creds at polymarket.com/settings?tab=builder",
    "details": { "missing_vars": ["POLYMARKET_BUILDER_API_KEY"] }
  },
  "meta": { "command": "deposit-wallet onboard", "ts": "2026-05-07T12:34:56Z", "duration_ms": 12 }
}
```

**Caveats:**

- `--fund-amount` is required; the agent must echo it back to the user
  verbatim before invoking.
- Use `--skip-deploy` if `deposit-wallet status` already reports
  `deployed: true`.
- Mutating; runs multiple chain and relayer writes in sequence.

### `deposit-wallet status`

**Purpose:** Check deposit-wallet deployment status, or poll a relayer
transaction by id.

**Required flags:** None. (Common optional flag: `--tx-id`.)

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE`.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "wallet": "0x...", "deployed": true, "tx_id": null },
  "meta": { "command": "deposit-wallet status", "ts": "2026-05-07T12:34:56Z", "duration_ms": 280 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "builder credentials not configured",
    "hint": "Set POLYMARKET_BUILDER_API_KEY / _SECRET / _PASSPHRASE."
  },
  "meta": { "command": "deposit-wallet status", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- Without `--tx-id`, returns wallet deployment status; with `--tx-id`,
  returns relayer transaction state.
- Read-only.

### Command catalog — `paper`

Local paper-trading state. No upstream API calls; never authenticated.

### `paper`

**Purpose:** Group entry; lists paper subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "subcommands": ["buy", "positions", "reset", "sell"] },
  "meta": { "command": "paper", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for paper",
    "hint": "Run `polygolem paper --help` to list subcommands."
  },
  "meta": { "command": "paper", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `--help` to list subcommands, then run a concrete subcommand.

### `paper buy`

**Purpose:** Record a simulated BUY into local paper-trading state.

**Required flags:** None. (Common optional flags: `--token-id`,
`--price`, `--size`.)

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "side": "BUY",
    "token_id": "7132...",
    "price": "0.50",
    "size": "10",
    "position": { "token_id": "7132...", "size": "10", "average_price": "0.50" }
  },
  "meta": { "command": "paper buy", "ts": "2026-05-07T12:34:56Z", "duration_ms": 6 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "INTERNAL_UNIMPLEMENTED",
    "category": "internal",
    "message": "polygolem paper buy: not implemented",
    "hint": "Track at docs/AUDIT-FINDINGS.md."
  },
  "meta": { "command": "paper buy", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Paper state lives on the local host; nothing is ever sent upstream.
- Currently a stub — will return `INTERNAL_UNIMPLEMENTED` until wired.

### `paper positions`

**Purpose:** List local paper positions.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "positions": [
      { "token_id": "7132...", "size": "10", "average_price": "0.50" }
    ],
    "count": 1
  },
  "meta": { "command": "paper positions", "ts": "2026-05-07T12:34:56Z", "duration_ms": 3 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "INTERNAL_UNIMPLEMENTED",
    "category": "internal",
    "message": "polygolem paper positions: not implemented",
    "hint": "Track at docs/AUDIT-FINDINGS.md."
  },
  "meta": { "command": "paper positions", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Empty `positions` is a valid success result, not an error.

### `paper reset`

**Purpose:** Clear local paper state back to zero.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "cleared": true },
  "meta": { "command": "paper reset", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "INTERNAL_STATE_CORRUPT",
    "category": "internal",
    "message": "paper state file failed integrity check",
    "hint": "Delete the paper state file and retry."
  },
  "meta": { "command": "paper reset", "ts": "2026-05-07T12:34:56Z", "duration_ms": 2 }
}
```

**Caveats:**

- Destructive but only against local state; never affects upstream.

### `paper sell`

**Purpose:** Record a simulated SELL into local paper-trading state.

**Required flags:** None. (Common optional flags: `--token-id`,
`--price`, `--size`.)

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "side": "SELL",
    "token_id": "7132...",
    "price": "0.55",
    "size": "10",
    "position": { "token_id": "7132...", "size": "0", "average_price": "0.50" }
  },
  "meta": { "command": "paper sell", "ts": "2026-05-07T12:34:56Z", "duration_ms": 6 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "INTERNAL_UNIMPLEMENTED",
    "category": "internal",
    "message": "polygolem paper sell: not implemented",
    "hint": "Track at docs/AUDIT-FINDINGS.md."
  },
  "meta": { "command": "paper sell", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Selling more than the held position should clamp at zero or return
  `VALIDATION_AMOUNT_OUT_OF_RANGE` once wired.

### Command catalog — `bridge`

Polymarket Bridge API: list supported assets and create deposit
addresses.

### `bridge`

**Purpose:** Group entry; lists bridge subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "subcommands": ["assets", "deposit"] },
  "meta": { "command": "bridge", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for bridge",
    "hint": "Run `polygolem bridge --help` to list subcommands."
  },
  "meta": { "command": "bridge", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `--help` to list subcommands, then run a concrete subcommand.

### `bridge assets`

**Purpose:** List supported bridge assets.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "supportedAssets": [
      { "symbol": "USDC", "chain": "polygon", "minAmount": "1.00" }
    ]
  },
  "meta": { "command": "bridge assets", "ts": "2026-05-07T12:34:56Z", "duration_ms": 240 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "NETWORK_TIMEOUT",
    "category": "network",
    "message": "Bridge API request exceeded deadline",
    "hint": "Retry; check `polygolem health`."
  },
  "meta": { "command": "bridge assets", "ts": "2026-05-07T12:34:56Z", "duration_ms": 10042 }
}
```

**Caveats:**

- Asset support changes over time; do not cache results across sessions.

### `bridge deposit`

**Purpose:** Create a deposit address for an EOA on a supported bridge
asset.

**Required flags:** Positional `<address>`.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "address": "0x...",
    "note": "Deposits to this address are credited to the supplied EOA."
  },
  "meta": { "command": "bridge deposit", "ts": "2026-05-07T12:34:56Z", "duration_ms": 380 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_FLAG_MISSING",
    "category": "usage",
    "message": "positional <address> is required",
    "hint": "Pass the EOA address as the first positional argument."
  },
  "meta": { "command": "bridge deposit", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Deposit addresses are bridge-side custodial; the EOA does not control
  the address private key.

### Command catalog — `events`

List Polymarket events.

### `events`

**Purpose:** Group entry; lists events subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "subcommands": ["list"] },
  "meta": { "command": "events", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for events",
    "hint": "Run `polygolem events --help` to list subcommands."
  },
  "meta": { "command": "events", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `events list` for data.

### `events list`

**Purpose:** List Polymarket events.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "events": [
      { "id": "0x...", "slug": "btc-100k", "title": "BTC above $100k?", "active": true }
    ],
    "count": 1
  },
  "meta": { "command": "events list", "ts": "2026-05-07T12:34:56Z", "duration_ms": 320 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "NETWORK_TIMEOUT",
    "category": "network",
    "message": "Gamma API request exceeded deadline",
    "hint": "Retry; check `polygolem health`."
  },
  "meta": { "command": "events list", "ts": "2026-05-07T12:34:56Z", "duration_ms": 10042 }
}
```

**Caveats:**

- The current binary returns a top-level array and silently ignores any
  `--limit` flag; tracked in `docs/AUDIT-FINDINGS.md`.

### Command catalog — `health`

Check Gamma and CLOB API reachability.

### `health`

**Purpose:** Check Gamma and CLOB API reachability.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_GAMMA_URL`, `POLYMARKET_CLOB_URL`
(both optional overrides).

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "clob": "ok", "gamma": "ok" },
  "meta": { "command": "health", "ts": "2026-05-07T12:34:56Z", "duration_ms": 320 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "NETWORK_TIMEOUT",
    "category": "network",
    "message": "Gamma reachability check exceeded deadline",
    "hint": "Verify network access and retry."
  },
  "meta": { "command": "health", "ts": "2026-05-07T12:34:56Z", "duration_ms": 10042 }
}
```

**Caveats:**

- Health is a smoke-check, not a deep readiness probe; per-endpoint
  failures still require running the failing command directly.

### Command catalog — `version`

Print version.

### `version`

**Purpose:** Print the binary version.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "version": "dev" },
  "meta": { "command": "version", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "INTERNAL_INVARIANT",
    "category": "internal",
    "message": "version metadata not embedded in binary",
    "hint": "Rebuild with the standard ldflags pipeline."
  },
  "meta": { "command": "version", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- `data.version` is the binary version; top-level `version` is the envelope
  contract version.

### Command catalog — `preflight`

Inspect local CLI readiness (config, RPC, builder creds).

### `preflight`

**Purpose:** Run the local readiness check suite.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE`, `POLYMARKET_RPC_URL` (all optional;
absent values are reported as failed checks rather than treated as
errors).

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "checks": [
      { "name": "private_key_present", "ok": true },
      { "name": "builder_creds_present", "ok": true },
      { "name": "rpc_reachable", "ok": true }
    ]
  },
  "meta": { "command": "preflight", "ts": "2026-05-07T12:34:56Z", "duration_ms": 280 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "GATE_PREFLIGHT_FAILED",
    "category": "gate",
    "message": "one or more preflight checks failed",
    "details": {
      "failed": [
        { "name": "rpc_reachable", "ok": false, "reason": "timeout after 10s" }
      ]
    }
  },
  "meta": { "command": "preflight", "ts": "2026-05-07T12:34:56Z", "duration_ms": 10120 }
}
```

**Caveats:**

- A failing preflight is a `gate` refusal, not a transient error; the
  user must fix the underlying issue before retrying.

### Command catalog — `auth`

Inspect authentication readiness.

### `auth`

**Purpose:** Group entry; lists auth subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "subcommands": ["status"] },
  "meta": { "command": "auth", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for auth",
    "hint": "Run `polygolem auth --help` to list subcommands."
  },
  "meta": { "command": "auth", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `auth status` for data once implemented.

### `auth status`

**Purpose:** Report which credentials are present and which are missing.

**Required flags:** None.

**Env vars consumed:** `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`,
`POLYMARKET_BUILDER_PASSPHRASE` (all optional; absence is reported, not
errored).

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": {
    "private_key_present": true,
    "builder_creds_present": true
  },
  "meta": { "command": "auth status", "ts": "2026-05-07T12:34:56Z", "duration_ms": 2 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "INTERNAL_UNIMPLEMENTED",
    "category": "internal",
    "message": "polygolem auth status: not implemented",
    "hint": "Track at docs/AUDIT-FINDINGS.md."
  },
  "meta": { "command": "auth status", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Reports presence only; never echoes the secret values themselves.

### Command catalog — `live`

Inspect live gate status.

### `live`

**Purpose:** Group entry; lists live subcommands.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "subcommands": ["status"] },
  "meta": { "command": "live", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "USAGE_SUBCOMMAND_UNKNOWN",
    "category": "usage",
    "message": "unknown subcommand for live",
    "hint": "Run `polygolem live --help` to list subcommands."
  },
  "meta": { "command": "live", "ts": "2026-05-07T12:34:56Z", "duration_ms": 1 }
}
```

**Caveats:**

- Group entry has no business logic; in `--json` mode it returns `USAGE_SUBCOMMAND_UNKNOWN`. Use `live status` for data once implemented.

### `live status`

**Purpose:** Report whether live mode is enabled and what gates apply.

**Required flags:** None.

**Env vars consumed:** None.

**Sample success JSON:**

```json title="v1 envelope"
{
  "ok": true,
  "version": "1",
  "data": { "live_enabled": true, "gates": [] },
  "meta": { "command": "live status", "ts": "2026-05-07T12:34:56Z", "duration_ms": 2 }
}
```

> **Contract:** implemented by the shared v1 JSON envelope. Command-specific fields appear under `data`; errors use `error.code` and `error.category`.

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "GATE_LIVE_DISABLED",
    "category": "gate",
    "message": "live mode is disabled",
    "hint": "Live trading is gated; the user must opt in out-of-band."
  },
  "meta": { "command": "live status", "ts": "2026-05-07T12:34:56Z", "duration_ms": 2 }
}
```

**Caveats:**

- A `gate` error here means the agent must stop, not retry with
  different flags.

## Common workflows

### 1. Find a tradeable market for a topic

```bash
./polygolem discover search --query "btc" --limit 5 --json | jq '.data.markets[] | {slug, question}'
SLUG=$(./polygolem discover search --query "btc" --limit 1 --json | jq -r '.data.markets[0].slug')
MARKET_ID=$(./polygolem discover market --slug "$SLUG" --json | jq -r '.data.id')
./polygolem discover enrich --id "$MARKET_ID" --json | jq '{tradeable: .data.tradeable, tickSize: .data.clob.tickSize}'
```

Decision rule for the agent:

- `data.tradeable == true` and `data.clob.bestBid > 0` and
  `data.clob.bestAsk > 0` → market is tradeable.
- `data.tradeable == false` → stop; report to user and pick another
  market.

### 2. Check market depth before quoting a size

```bash
TOKEN_ID=$(./polygolem discover market --slug "$SLUG" --json | jq -r '.data.clobTokenIds[0]')
./polygolem orderbook get --token-id "$TOKEN_ID" --json | jq '{
  bids: .data.bids[0:3],
  asks: .data.asks[0:3],
  midpoint: .data.midpoint
}'
./polygolem orderbook spread --token-id "$TOKEN_ID" --json | jq '.data.spread'
```

The agent uses the top-of-book + midpoint to compute slippage tolerance
before submitting any order.

### 3. Onboard a new account end-to-end (deposit-wallet)

Pre-flight: the user must have set `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`, and
`POLYMARKET_BUILDER_PASSPHRASE` in their environment **before** the agent
starts.

```bash
./polygolem preflight --json
./polygolem deposit-wallet derive --json
./polygolem deposit-wallet status --json
./polygolem deposit-wallet onboard --fund-amount 25 --json
```

Decision rule:

- `preflight` returns `ok: false` with `error.category == "gate"` → stop;
  surface the failed checks to the user.
- `deposit-wallet status` returns `data.deployed: true` → skip `onboard`
  deploy step; only `fund` is needed.
- `deposit-wallet onboard` requires explicit `--fund-amount`; the agent
  must echo the amount back to the user before running this command.

### 4. Place a paper trade (always safe to try first)

```bash
./polygolem paper buy --token-id "$TOKEN_ID" --price 0.50 --size 10 --json
./polygolem paper positions --json | jq '.data.positions'
```

Paper mode never touches authenticated endpoints. Use it as a dry-run
before any live order.

### 5. Place a live CLOB order (requires explicit confirmation)

```bash
# 1. confirm tradability
./polygolem orderbook tick-size --token-id "$TOKEN_ID" --json
# 2. quote round-trip first
./polygolem clob book "$TOKEN_ID" --json | jq '.data | {bestBid: .bids[0], bestAsk: .asks[0]}'
# 3. live order — REQUIRES user confirmation in the same turn
./polygolem clob create-order \
  --token "$TOKEN_ID" \
  --side BUY --price 0.50 --size 10 \
  --builder-code "$POLYMARKET_BUILDER_CODE" \
  --json
```

Decision rule:

- Echo `--token`, `--side`, `--price`, `--size`, and any configured
  `--builder-code`
  back to the user verbatim before running step 3.
- On `error.category == "gate"` or `"validation"`, the agent stops and
  reports — it does not retry with adjusted flags.

## See also

- `docs/JSON-CONTRACT.md` — full envelope and error-code specification.
- `docs/COMMANDS.md` — flag-by-flag reference for every command.
- `docs/SAFETY.md` — safety boundaries, including deposit-wallet rules.
- `docs/ARCHITECTURE.md` — package map and dependency direction.
- `polygolem <cmd> --help` — normative for flag semantics.
