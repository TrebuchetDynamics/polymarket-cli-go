# Architecture and Taxonomy Improvement Plan

Status: proposed

Scope: package names, public SDK contracts, docs language, and command taxonomy.
This plan does not change live trading behavior or wallet safety gates.

## Verdict

Polygolem's top-level taxonomy is directionally right:

- API families: Gamma, CLOB, Data API, Bridge, Relayer, WebSocket.
- Wallet model: deposit wallet only for live trading, with POLY_1271 / type 3 signing.
- Safety model: read-only by default, paper local-only, live gated.

The main gaps are not missing nouns. The gaps are inconsistent boundaries:

- Public SDK packages expose some `internal/*` protocol types.
- `pkg/universal` is doing too much without a crisp contract.
- `pkg/bookreader` should become `pkg/orderbook`, or be deprecated in favor of it.
- The word "builder" is overloaded across CLOB API credentials, builder relayer credentials, and order attribution.
- Docs describe the current surface well, but there is no canonical glossary or ADR that protects names from drifting.

## Canonical Vocabulary

Use these names consistently in code, CLI help, docs, and tests.

| Canonical term | Meaning | Avoid |
|---|---|---|
| Gamma | Public discovery/content API: markets, events, tags, series, comments | Categories as a top-level API family |
| CLOB | Central limit order book API: books, markets, balances, orders, trades, cancellations | Exchange API |
| Data API | Analytics API: positions, holders, leaderboards, live volume | Analytics when naming packages |
| Bridge | Polymarket bridge/deposit API | Funding API |
| Relayer | Builder relayer API for deposit-wallet deployment and batches | Builder API when referring to relayer calls |
| WebSocket | Realtime market or user streams | Events for stream package names |
| Deposit Wallet | Polymarket smart wallet used for current live trading | Proxy wallet, Safe wallet |
| POLY_1271 | CLOB signature type 3 deposit-wallet order signing | Magic value 3 without a name |
| CLOB API Credentials | L2 CLOB key, secret, and passphrase derived from wallet auth | Builder credentials |
| Builder Relayer Credentials | Builder API key, secret, and passphrase used by the relayer | CLOB credentials |
| Builder Code | Order attribution bytes32 field | Builder credentials |
| Condition ID | Market identifier used by CLOB for market-level actions | Market ID when the endpoint expects a condition |
| Token ID | Outcome token identifier used for books and token-level actions | Asset ID unless matching WebSocket docs |

## Target Package Map

Public SDK packages should describe stable consumer contracts:

| Target package | Role |
|---|---|
| `pkg/gamma` | Stable Gamma discovery client |
| `pkg/clob` | Stable CLOB read/account/trading client once write gates are ready for public SDK use |
| `pkg/data` | Stable Data API analytics client |
| `pkg/stream` | Stable WebSocket clients |
| `pkg/bridge` | Stable Bridge client |
| `pkg/orderbook` | Stable order-book reader; replace or deprecate `pkg/bookreader` |
| `pkg/types` | Exported shared protocol DTOs used by public packages |
| `pkg/client` | Optional facade across stable SDK packages |

Internal packages should keep implementation detail names:

| Internal package family | Role |
|---|---|
| `internal/{gamma,clob,dataapi,stream,bridge,relayer}` | Protocol clients |
| `internal/auth` | L0/L1/L2 auth, EIP-712, POLY_1271, credential derivation |
| `internal/wallet` | Deposit-wallet derivation, status, batch signing |
| `internal/workflows/*` | CLI-level orchestration such as onboarding and market discovery |
| `internal/polytypes` | Internal CLOB/enrichment protocol types plus aliases for public Gamma DTOs |

## Work Plan

### Phase 1 - Freeze the language

Add a glossary and ADRs before renaming code:

- `CONTEXT.md`: canonical terms, unsupported terms, and safety boundaries.
- `docs/adr/0001-polymarket-api-taxonomy.md`: why Gamma, CLOB, Data API, Bridge, Relayer, and WebSocket stay separate.
- `docs/adr/0002-deposit-wallet-only-live-trading.md`: why EOA, proxy, and Safe modes remain unsupported for current production trading.
- `docs/adr/0003-public-sdk-type-boundary.md`: rule that public packages should not require callers to name `internal/*` types.

Verification:

- Docs search has one meaning for "builder credentials".
- PR template asks whether public SDK signatures expose internal packages.

### Phase 2 - Clean auth and builder names

Separate the three concepts currently hidden behind "builder":

- `CLOBAPIKey`: CLOB L2 key, secret, passphrase.
- `BuilderRelayerCredentials`: relayer key, secret, passphrase.
- `BuilderCode`: order attribution bytes32 value.

Verification:

- CLI help and docs never say "builder creds" without saying relayer or attribution.
- Tests cover env var loading for both CLOB and relayer credential names.

### Phase 3 - Promote public DTOs

Create `pkg/types` or another clearly named public DTO package, then re-export
or migrate stable structs out of `internal/polytypes`, `internal/dataapi`,
`internal/clob`, and `internal/stream` where public packages expose them.

Progress:

- 2026-05-08: Data API DTOs were promoted first. `pkg/data` is the canonical
  read-only public Data API client, and `pkg/universal` now returns
  `pkg/types` for Data API positions, trades, activity, holders, portfolio
  value, markets traded, open interest, leaderboard, and live volume.
- 2026-05-08: Gamma DTOs were promoted next. `pkg/gamma` and the Gamma methods
  on `pkg/universal` now return `pkg/types` for markets, events, tags, series,
  comments, profiles, search, sports metadata, and keyset pagination. The
  internal `polytypes` package aliases those DTOs for existing internal callers.

Verification:

- `go doc ./pkg/...` does not show public methods whose callers must name `internal/*` types.
- A small external-module compile test imports `pkg/gamma`, `pkg/universal` or `pkg/client`, and `pkg/bridge`.

### Phase 4 - Split the facade

Decide whether `pkg/universal` stays as a compatibility facade or becomes
`pkg/client`. Keep protocol-specific packages as the primary SDK surface.

Rules:

- Keep `pkg/universal` only if it is a convenience facade with no unique business logic.
- Put protocol details in protocol packages.
- Put multi-step operator actions in workflow packages, not in the facade.

Verification:

- Each public method has a clear owner package.
- `pkg/universal` tests prove delegation, not protocol behavior.

### Phase 5 - Rename order-book package safely

Introduce `pkg/orderbook` and keep `pkg/bookreader` as a deprecated wrapper for
one minor release if compatibility matters.

Verification:

- Existing consumers still compile.
- New docs and examples use `pkg/orderbook`.

### Phase 6 - Reduce CLI orchestration depth

Move multi-step command behavior out of `internal/cli/root.go` into workflow
packages where it can be tested without Cobra.

Initial workflow candidates:

- `internal/workflows/walletlifecycle`
- `internal/workflows/marketdiscovery`
- `internal/workflows/tradingaccount`
- `internal/workflows/streaming`

Verification:

- Cobra tests cover flags and command wiring.
- Workflow tests cover behavior with fake clients.
- Wallet and live-order gates remain tested outside Cobra.

## Definition of Done

- One canonical glossary exists and is linked from README, docs, and Starlight.
- Public SDK signatures do not leak `internal/*` packages.
- Builder, relayer, CLOB credential, and builder-code terms are unambiguous.
- `pkg/orderbook` is the documented name for order-book reads.
- `pkg/universal` is either a thin compatibility facade or replaced by `pkg/client`.
- CI includes a repository hygiene check for generated command docs and SDK boundary drift.
