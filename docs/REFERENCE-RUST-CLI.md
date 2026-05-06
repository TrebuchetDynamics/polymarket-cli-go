# Rust Reference CLI Audit

## Source

- Reference repository: https://github.com/Polymarket/polymarket-cli
- Audited commit: `4b5a749`
- Audit date: 2026-05-06
- License: MIT, from `Cargo.toml` at the audited commit.
- Extracted evidence files:
  - `/tmp/polymarket-cli-Cargo.toml`
  - `/tmp/polymarket-cli-main.rs`
  - `/tmp/polymarket-cli-config.rs`
  - `/tmp/polymarket-cli-auth.rs`
  - `/tmp/polymarket-cli-clob.rs`

This Rust CLI is a behavioral and protocol reference only. The Go CLI must not
copy Rust source, and it must not blindly translate Rust implementation details.
Use the reference to understand command shape, authentication behavior, and
Polymarket protocol assumptions, then implement Go behavior independently.

## Behavioral Use

The Rust binary is named `polymarket` and describes itself as a CLI to browse
markets, trade, and manage positions. It uses global `--output`, `--private-key`,
and `--signature-type` flags.

Useful behavioral references for Go:

- Global output mode supports `table` and `json`, defaulting to `table`.
- Global private key flag overrides other key sources.
- Global signature type accepts `eoa`, `proxy`, or `gnosis-safe`.
- Top-level public/read workflows include markets, events, tags, series,
  comments, profiles, sports metadata, data, bridge, and status.
- Wallet, approvals, CLOB, CTF, and some bridge flows include wallet or
  mutation semantics and must be treated carefully in Go Phase 1.

## Command Structure

The Rust top-level command enum includes:

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

The CLOB command group includes unauthenticated read commands and authenticated
commands. Read-oriented CLOB commands include health, price, batch prices,
midpoint, midpoints, spread, spreads, book, books, last trade, last trades,
market, markets, sampling markets, simplified markets, tick size, fee rate,
negative-risk status, price history, server time, and geoblock status.

Authenticated CLOB commands include order reads, trade reads, balance reads,
order creation, cancellation, balance allowance refresh, notifications, rewards,
API key operations, and account status.

## Auth Model

Private key resolution priority in Rust is:

1. `--private-key` CLI flag
2. `POLYMARKET_PRIVATE_KEY` environment variable
3. `~/.config/polymarket/config.json`

Signature type resolution priority in Rust is:

1. `--signature-type` CLI flag
2. `POLYMARKET_SIGNATURE_TYPE` environment variable
3. `~/.config/polymarket/config.json`
4. default `proxy`

The supported signature type strings are:

- `proxy`, mapped to the SDK proxy signature type
- `gnosis-safe`, mapped to the SDK Gnosis Safe signature type
- `eoa`, mapped to the SDK EOA signature type

Rust's parser falls back to EOA for unknown signature strings after config
resolution. Go should not inherit that fallback automatically; it should define
and test its own validation behavior explicitly.

Rust sets the signer chain ID to Polymarket's Polygon constant from the SDK. It
also uses `POLYMARKET_RPC_URL` for provider creation and otherwise defaults to
`https://polygon.drpc.org`.

## Config Format

Rust stores wallet config at:

```text
~/.config/polymarket/config.json
```

The serialized JSON fields are:

```json
{
  "private_key": "...",
  "chain_id": 137,
  "signature_type": "proxy"
}
```

The audited Rust config type requires `private_key` and `chain_id`. The
`signature_type` field has a default of `proxy` for deserialization if absent.
On Unix, Rust creates the config directory with mode `0700` and writes the
config file with mode `0600`.

## Live Mutation Paths

The Rust reference includes live mutation paths. Go Phase 1 blocks live
mutations and must not implement or expose these paths as executable live
actions.

CLOB live mutation paths observed in `src/commands/clob.rs`:

- `create-order`: builds, signs, and posts a limit order through `post_order`.
- `post-orders`: builds, signs, and posts multiple limit orders through
  `post_orders`.
- market buy/sell command: builds, signs, and posts an immediate-style order
  through `post_order`.
- `cancel`: cancels a single order through `cancel_order`.
- `cancel-orders`: cancels multiple orders through `cancel_orders`.
- `cancel-all`: cancels all open orders through `cancel_all_orders`.
- `cancel-market`: cancels market/asset-filtered orders through
  `cancel_market_orders`.
- `update-balance`: refreshes balance allowance through
  `update_balance_allowance`.
- `delete-notifications`: deletes notifications through `delete_notifications`.
- `create-api-key`: creates or derives an API key through
  `create_or_derive_api_key`.
- `delete-api-key`: deletes the current API key through `delete_api_key`.

Other top-level groups named `approve`, `ctf`, `bridge`, `wallet`, and `setup`
also signal wallet or protocol mutation risk from their command descriptions.
They are out of scope for Go Phase 1 unless separately audited and guarded.

## Protocol Assumptions

The Rust reference assumes:

- Polymarket CLOB operations run against Polygon.
- The SDK owns core CLOB, Gamma, Data, Bridge, and CTF protocol mechanics.
- Token IDs are numeric strings parsed as unsigned 256-bit integers.
- Condition IDs are accepted as strings or 256-bit hex values depending on the
  command.
- Order sides are `buy` and `sell`.
- Order types include `GTC`, `FOK`, `GTD`, and `FAK`.
- Market buy amounts are USDC amounts; market sell amounts are share amounts.
- Authenticated CLOB calls require a resolved private key and signature type.
- Unauthenticated CLOB market-data calls can use a default CLOB client.

These assumptions should become explicit Go tests before production behavior is
implemented. Do not rely on the Rust SDK as a substitute for Go-side protocol
contracts.

## Maintenance Concerns

- The Rust CLI depends on `polymarket-client-sdk = "0.4"` with features for
  Gamma, Data, Bridge, CLOB, and CTF. SDK behavior may move faster than this Go
  port.
- Rust's command surface is broad and mixes read-only queries with live trading,
  wallet, approval, bridge, and CTF operations. Go must keep Phase 1 scope narrow
  and mutation-blocked.
- Rust accepts unknown signature strings as EOA after parsing. That behavior is
  risky to clone without an explicit product decision.
- Config storage contains a private key on disk. Go should keep wallet handling
  outside Phase 1 unless a separate tested threat model exists.
- Some Rust commands are authenticated but read-only; others mutate remote state.
  The Go CLI should classify commands by side effect, not merely by whether they
  require authentication.
- Temporary audit files came from Git object reads, not working-tree Rust files,
  so the audit is independent of any local Rust file edits.

## License Constraints

The audited Rust project declares MIT license metadata in `Cargo.toml`.
MIT permits use, copying, modification, and distribution subject to preserving
the copyright and permission notice.

For this Go port:

- Treat Rust as a reference, not source material to copy.
- Do not paste Rust implementation bodies into Go.
- Preserve attribution in documentation where Rust behavior influenced design.
- Re-check repository license files before vendoring or copying any upstream
  artifact beyond small behavioral facts.
