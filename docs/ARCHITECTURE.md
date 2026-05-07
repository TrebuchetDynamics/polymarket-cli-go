# Architecture

`polygolem` is a Go protocol and automation stack for Polymarket with a
Cobra-based CLI frontend. The CLI is a thin shell over typed, testable
internal packages and a small public SDK in `pkg/`.

## Surface map

### Public SDK (`pkg/`)

Stable interfaces for downstream Go consumers (e.g., `go-bot`).

| Package | Purpose |
|---|---|
| `pkg/bookreader` | Read-only CLOB order-book reader. |
| `pkg/bridge` | Bridge API client — supported assets, deposit addresses, quotes. |
| `pkg/gamma` | Read-only Gamma API surface for embedded use. |
| `pkg/marketresolver` | Resolve market identifiers (ID, slug, token-id) to a canonical view. |
| `pkg/pagination` | Cursor and offset pagination with concurrent batching. |

### Internal packages (`internal/`)

Implementation. Not part of the public SDK contract.

| Package | Purpose |
|---|---|
| `internal/auth` | L0/L1/L2 auth, EIP-712, deposit-wallet CREATE2 derivation, builder attribution, signers. |
| `internal/cli` | Cobra command construction and dependency wiring. |
| `internal/clob` | CLOB API client — full read + authenticated surface, EIP-712, POLY_1271, ERC-7739. |
| `internal/config` | Viper-backed config loading, defaults, environment binding, validation, redaction. |
| `internal/dataapi` | Data API client — positions, volume, leaderboards. |
| `internal/errors` | Structured error types and code helpers. |
| `internal/execution` | Paper executor today; live executor surface for future use. |
| `internal/gamma` | Typed Gamma HTTP client — markets, events, search, tags, series, sports, comments, profiles. |
| `internal/marketdiscovery` | High-level market discovery service that combines Gamma and CLOB. |
| `internal/modes` | Read-only / paper / live mode parsing and gate checks. |
| `internal/orders` | OrderIntent, fluent builder, validation, lifecycle states. |
| `internal/output` | Stable table and JSON rendering plus structured errors. |
| `internal/paper` | Local-only paper positions, fills, and persisted state. |
| `internal/polytypes` | Polymarket protocol-level types shared across clients. |
| `internal/preflight` | Local and remote readiness checks. |
| `internal/relayer` | Builder relayer client — WALLET-CREATE, WALLET batch, nonce, polling. |
| `internal/risk` | Per-trade caps, daily loss limits, circuit breaker. |
| `internal/rpc` | Direct on-chain transfers (e.g., ERC-20 pUSD from EOA). |
| `internal/stream` | WebSocket market client with reconnect and dedup. |
| `internal/transport` | HTTP retry, rate limiter, circuit breaker, redaction. |
| `internal/wallet` | Deposit-wallet primitives — derive, deploy, status, batch signing. |

## Dependency direction

```text
cmd/polygolem
        |
internal/cli
        |
internal/{config, modes, preflight, output, errors}
        |
internal/{gamma, clob, dataapi, stream, relayer, rpc}   ← protocol clients
        |
internal/{auth, transport, polytypes}                   ← cross-cutting primitives
        |
internal/{wallet, orders, execution, risk, paper, marketdiscovery}
        |
pkg/{bookreader, bridge, gamma, marketresolver, pagination}   ← public re-exposed surface
```

Command handlers parse flags, call package APIs, and render output via
`internal/output`. Protocol clients do not know about Cobra. Safety packages
do not depend on command text. Paper state stays local and never reaches
authenticated mutation endpoints.

Cobra command handlers must not contain protocol or trading business logic.
That logic belongs in typed clients, application services, safety gates, and
paper-state packages where it is testable without executing the binary.

## Mode system

Mode selection starts in configuration and CLI flags, then flows through
`internal/modes` before command handlers call protocol clients or paper
state.

- **Read-only** (default): public market data only. May use
  `internal/gamma`, `internal/clob` (read endpoints), `internal/dataapi`,
  `internal/marketdiscovery`, and `internal/output`. Forbids signing or
  any mutation.
- **Paper**: local simulation. Combines read-only reference data with
  `internal/paper` state. Simulated actions stay local. Authenticated
  mutation APIs remain off-limits.
- **Live**: gated. Requires preflight + risk + funding gates to pass.
  Live execution operates through `internal/execution`, `internal/orders`,
  `internal/clob` (write endpoints), `internal/relayer`, `internal/rpc`,
  and `internal/wallet`. The default `polygolem` invocation does not enter
  live mode.

## Signature types

Polygolem supports **deposit wallet (POLY_1271 / type 3)** exclusively.
EOA, proxy, and Gnosis Safe are blocked by CLOB V2 and are not supported.

| Value | Status |
|-------|--------|
| `deposit` | ✅ Deposit wallet (POLY_1271). Required for all trading. |
| `eoa` | ❌ Blocked by CLOB V2 |
| `proxy` | ❌ Blocked by CLOB V2 |
| `safe` / `gnosis-safe` | ❌ Blocked by CLOB V2 |

Builder credentials are required for deposit wallet deployment via the
relayer. Order attribution uses the on-order `builder` bytes32 field (V2).

## Public SDK boundary

`pkg/` exists. It is small by design and grows when an internal capability
proves stable enough to expose. Do not move code into `pkg/` without an
SDK-level commitment to keep its API stable across minor versions.

## Safety boundaries

- Read-only is the default mode and is exercised by every public command.
- Paper mode never calls authenticated endpoints.
- Live commands require explicit signature-type, gates passing, and
  builder credentials where applicable.
- Builder credentials and private keys are redacted by `internal/config`
  on every load.
