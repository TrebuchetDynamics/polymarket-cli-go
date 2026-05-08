# JSON Contract — v1

`polygolem` emits a single, versioned JSON envelope for every command invoked
with `--json`. This document is the canonical specification. Agent integrators,
shell scripts, and downstream consumers should treat this file as authoritative.

The envelope is designed so a caller can answer three questions without
parsing the inner payload:

1. Did it succeed? (`ok`)
2. What version of the contract is this? (`version`)
3. If it failed, what category and code? (`error.category`, `error.code`)

The exit code mirrors the error category for shell-script convenience.

## Versioning policy

- `version` is a string. Current value: `"1"`.
- **Non-breaking changes** that do not bump `version`: adding new top-level
  commands, adding new fields inside `data`, adding new error `code` strings,
  adding new fields inside `error.details`.
- **Breaking changes** that bump `version`: renaming or removing any envelope
  field, renaming or removing an existing error `code`, changing the type of
  an existing field, changing the meaning of an existing exit code.
- Future minor versions parse as `"1.1"`, `"1.2"`, etc. Major versions parse
  as `"2"`, `"3"`. Callers should accept any `version` whose major component
  matches the major they were written against.

## Success envelope

```json
{
  "ok": true,
  "version": "1",
  "data": { ... command-specific payload ... },
  "meta": {
    "command": "deposit-wallet onboard",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 2143
  }
}
```

Field semantics:

- `ok` — boolean, always `true` on success. The single field a caller checks.
- `version` — string, the contract version (currently `"1"`).
- `data` — object, command-specific payload. Field naming inside `data`
  preserves whatever the underlying upstream API returned. Polygolem does not
  re-case upstream payloads.
- `meta.command` — the full command path as the user invoked it
  (e.g., `"clob create-order"`, not just `"create-order"`).
- `meta.ts` — RFC 3339 / ISO 8601 timestamp in UTC, second precision or finer.
- `meta.duration_ms` — integer, wall-clock duration of the invocation in
  milliseconds.

## Error envelope

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "POLYMARKET_BUILDER_API_KEY is required for deposit-wallet onboard",
    "hint": "Create builder creds at polymarket.com/settings?tab=builder",
    "details": { ... optional structured context ... }
  },
  "meta": { "command": "deposit-wallet onboard", "ts": "2026-05-07T12:34:56Z", "duration_ms": 12 }
}
```

Field semantics:

- `ok` — boolean, always `false` on error.
- `version` — same as success envelope.
- `error.code` — `SCREAMING_SNAKE_CASE` stable string. See the taxonomy below.
- `error.category` — one of the eight categories defined in the next section.
- `error.message` — human-readable, single-sentence explanation suitable for
  surfacing to a user.
- `error.hint` — optional, single-sentence suggestion for resolution. Omit
  when no actionable hint exists.
- `error.details` — optional object with structured context (e.g., the
  offending flag name, the upstream HTTP status, the on-chain revert reason).
  Field shape varies by code.
- `meta` — always present, identical to success envelope.

`data` and `error` are mutually exclusive. A success envelope never carries
`error`; an error envelope never carries `data`.

## Field-naming convention

Envelope fields use `snake_case` (e.g., `duration_ms`, `version`). Inside
`data`, fields preserve upstream casing — Polymarket's Gamma API mixes
`camelCase` and `snake_case`, and polygolem does not re-shape upstream
payloads. Callers can rely on `snake_case` for the envelope itself but must
read each `data` payload's fields by the names the upstream API uses.

## Error code taxonomy

Two-level: `category` (broad) and `code` (specific). Categories are closed;
codes are open (new codes can be added without a `version` bump).

### `usage` — bad flags, missing args, conflicting options

| Code | Description |
|---|---|
| `USAGE_FLAG_MISSING` | A required flag was not provided on the command line. |
| `USAGE_FLAG_INVALID` | A flag was given a value the command does not accept (e.g., `--limit -1`). |
| `USAGE_FLAG_CONFLICT` | Two mutually exclusive flags were provided together (e.g., `--id` and `--slug`). |
| `USAGE_SUBCOMMAND_UNKNOWN` | A subcommand path does not resolve to a registered command. |

### `auth` — missing or invalid credentials, signature failure

| Code | Description |
|---|---|
| `AUTH_PRIVATE_KEY_MISSING` | `POLYMARKET_PRIVATE_KEY` is required by the command but not set or empty. |
| `AUTH_BUILDER_MISSING` | One or more of `POLYMARKET_BUILDER_API_KEY` / `_SECRET` / `_PASSPHRASE` is required but not set. |
| `AUTH_SIG_INVALID` | A signature (EIP-712, POLY_1271, or ERC-7739) failed local validation before submission. |
| `AUTH_API_KEY_REJECTED` | The CLOB API rejected an L2 API key (expired, revoked, or wrong account). |

### `validation` — input parse or range failures

| Code | Description |
|---|---|
| `VALIDATION_TOKEN_ID_INVALID` | `--token-id` is not a valid CLOB token id (wrong length, non-numeric, or unknown). |
| `VALIDATION_AMOUNT_OUT_OF_RANGE` | A numeric amount is below the minimum or above the maximum the command accepts. |
| `VALIDATION_PRICE_TICK_MISMATCH` | A submitted order price does not align with the market's tick size. |
| `VALIDATION_MARKET_IDENTIFIER_AMBIGUOUS` | A market identifier (id, slug, or token-id) matches zero or multiple markets. |

### `gate` — mode, safety, preflight, or risk refusal

| Code | Description |
|---|---|
| `GATE_LIVE_DISABLED` | Live mode is required for the command but the binary is running in read-only or paper mode. |
| `GATE_PREFLIGHT_FAILED` | A `preflight` check (RPC reachability, balance, allowance, or wallet status) failed. |
| `GATE_RISK_LIMIT` | A per-trade or daily risk cap was exceeded; the command refused to proceed. |
| `GATE_FUNDING_INSUFFICIENT` | The deposit wallet or EOA does not hold the funds the command would move. |

### `network` — HTTP failure, DNS, timeout, retry exhaustion

| Code | Description |
|---|---|
| `NETWORK_TIMEOUT` | A request exceeded its deadline; the underlying transport gave up. |
| `NETWORK_DNS` | DNS resolution for an upstream host failed. |
| `NETWORK_TLS` | TLS handshake or certificate verification failed against an upstream host. |
| `NETWORK_RETRY_EXHAUSTED` | The transport's retry budget was consumed without a successful response. |

### `protocol` — upstream API errored or returned an unexpected shape

| Code | Description |
|---|---|
| `PROTOCOL_GAMMA_4XX` | The Gamma API returned a 4xx response (bad request, auth, not found, or rate-limited). |
| `PROTOCOL_CLOB_5XX` | The CLOB API returned a 5xx response (server error, unavailable, or gateway timeout). |
| `PROTOCOL_RELAYER_REJECTED` | The builder relayer rejected a `WALLET-CREATE` or `WALLET` batch submission. |
| `PROTOCOL_UNEXPECTED_SHAPE` | An upstream response could not be decoded into the expected schema. |

### `chain` — on-chain RPC, nonce, or revert errors

| Code | Description |
|---|---|
| `CHAIN_NONCE_TOO_LOW` | The submitted transaction's nonce is below the account's current nonce on the destination chain. |
| `CHAIN_REVERTED` | A simulated or submitted transaction reverted; the revert reason is in `details.revert_reason` if known. |
| `CHAIN_INSUFFICIENT_FUNDS` | The sending account does not hold enough native gas token or ERC-20 balance to complete the call. |
| `CHAIN_RPC_UNAVAILABLE` | The configured RPC endpoint did not respond, or returned a transport-level error. |

### `internal` — bug; an invariant the code expects was violated

| Code | Description |
|---|---|
| `INTERNAL_PANIC` | A goroutine panicked and was recovered by the top-level handler; this is always a bug. |
| `INTERNAL_INVARIANT` | A `should-never-happen` precondition was violated (e.g., a non-nil expected value was nil). |
| `INTERNAL_UNIMPLEMENTED` | A code path was reached that has not yet been wired up; please file an issue. |
| `INTERNAL_STATE_CORRUPT` | Persistent state on disk (e.g., paper-mode positions) failed integrity checks. |

## Exit-code matrix

Wrapping shell scripts can branch on exit code without parsing JSON. The JSON
envelope remains authoritative; exit codes are a convenience.

| Exit | Meaning | Mapping |
|---|---|---|
| `0` | success | `ok: true` |
| `1` | generic | category did not map cleanly, or no envelope was emitted |
| `2` | usage | `error.category == "usage"` |
| `3` | auth | `error.category == "auth"` |
| `4` | validation | `error.category == "validation"` |
| `5` | gate | `error.category == "gate"` |
| `6` | network | `error.category == "network"` |
| `7` | protocol | `error.category == "protocol"` |
| `8` | chain | `error.category == "chain"` |
| `9` | internal | `error.category == "internal"` |

## Worked examples

### Success — `polygolem discover search --query "btc" --limit 1 --json`

```json
{
  "ok": true,
  "version": "1",
  "data": {
    "markets": [
      {
        "id": "0xabc...",
        "slug": "will-btc-be-above-100k",
        "question": "Will BTC be above $100k by end of 2026?",
        "active": true
      }
    ],
    "count": 1
  },
  "meta": {
    "command": "discover search",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 412
  }
}
```

### Error — `polygolem deposit-wallet onboard --fund-amount 10 --json` without builder creds

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
  "meta": {
    "command": "deposit-wallet onboard",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 12
  }
}
```

Exit code: `3` (`auth`).

### Error — `polygolem orderbook get --token-id "" --json`

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
  "meta": {
    "command": "orderbook get",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 3
  }
}
```

Exit code: `4` (`validation`).

## Implementation status

The shared CLI layer emits the v1 envelope for `--json` success output, group
commands invoked without a subcommand, not-implemented command stubs, missing
private-key or builder credentials, and other command errors. The remaining
alignment work is protocol-specific classification: upstream 4xx/5xx,
transport, and chain failures still need more precise `error.code` values and
structured `error.details`.

## Source of truth

This file is canonical for the envelope shape, the eight error categories,
the listed example codes, and the exit-code matrix. The Starlight pages
`reference/json-contract.mdx` and `reference/error-codes.mdx` (built in
Track 4) link to this file. `SKILL.md` references this file for envelope
syntax instead of duplicating it.
