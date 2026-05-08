# Polygolem SDK Requirements PRD

> **Audit status (2026-05-07):** This PRD predates the current codebase.
> Each requirement is annotated with one of:
>
> - ✅ **Fulfilled** — implemented and shipping.
> - ⚠️ **Partial / drifted** — implementation differs in scope or shape.
> - 🗒️ **Historical** — preserved for context; superseded by a later
>   decision documented in `docs/ARCHITECTURE.md` or
>   `docs/DEPOSIT-WALLET-MIGRATION.md`.
>
> The status reflects code reality at the audit date. The "why" prose is
> preserved unchanged.

Status: draft
Date: 2026-05-06
Scope: requirements for the future Polygolem SDK architecture and go-bot
consumer boundary

## Problem Statement

`polygolem` currently has a safe Go Phase 1 foundation: a Cobra CLI shell,
configuration, execution modes, preflight checks, read-only Gamma/CLOB clients,
structured output, and local paper state. The next architecture decision is how
to grow this into a reusable Polymarket SDK foundation without letting command
handlers, trading logic, authentication, and live-risk controls collapse into
one coupled CLI layer.

The SDK needs a clear requirement list for market discovery, authentication,
CLOB market data, orders, balances, allowances, streams, paper execution, and
future live execution. This document defines those needs before implementation.

For `go-bot`, Polygolem must become the only Polymarket protocol boundary. Any
Polymarket capability that the bot needs, including market discovery, CLOB
books, prices, history, auth readiness, balances, orders, trades, streams, and
future bridge or CTF flows, must come from Polygolem rather than direct bot
clients.

## Source Inputs

This PRD is based on:

- Current `polygolem` docs and packages: `internal/cli`, `internal/config`,
  `internal/modes`, `internal/preflight`, `internal/gamma`, `internal/clob`,
  `internal/paper`, `docs/ARCHITECTURE.md`, `docs/SAFETY.md`, and
  `docs/COMMANDS.md`.
- Local reference repositories in `polygolem/opensource-projects/repos`:
  `polymarket_cli`, `polymarket-go-gamma-client`, `polymarket-go`,
  `polymarket-go-sdk`, and `go-builder-signing-sdk`.
- Current Polymarket API and CLOB client documentation fetched through
  Context7 on 2026-05-06.
- Current `go-bot` integration scan on 2026-05-06, which found direct Gamma and
  CLOB clients under `internal/polymarket`, direct `POLYMARKET_GAMMA_URL` and
  `POLYMARKET_CLOB_URL` use, and paper mode CLOB book fetches that must move
  behind Polygolem.

Important reference lessons:

- `polymarket_cli`: keep the CLI thin, JSON-first, and read-only by default.
- `polymarket-go-gamma-client`: model Gamma API quirks explicitly, including
  inconsistent datetime and string-or-array fields.
- `polymarket-go`: split CLOB, Gamma, Data, WebSocket, signer, relayer, bridge,
  and Turnkey concerns into separate packages.
- `polymarket-go-sdk`: use domain modules, typed errors, order builders,
  rate-limit awareness, WebSocket lifecycle management, KMS/signer separation,
  and batch-size validation.
- `go-builder-signing-sdk`: treat builder attribution headers as a separate
  optional signing concern, not as user trading authentication.

## Goals

- Define the minimum SDK architecture required for safe market research,
  paper trading, and future gated live trading.
- Keep read-only workflows credential-free.
- Preserve Phase 1 safety: no live order placement, signing, token approval, or
  on-chain transaction path is enabled by this PRD.
- Create deep modules with small public surfaces: market discovery, auth,
  CLOB data, order building, execution, streams, account data, transport,
  safety, and paper state.
- Make the eventual SDK usable by the CLI, bots, daemons, tests, and operator
  tooling without duplicating protocol logic.
- Make `go-bot` consume Polymarket through Polygolem only, with stable typed
  interfaces and JSON contracts that can be mocked in TDD.

## Non-Goals

- This PRD does not authorize live trading implementation.
- This PRD does not choose a third-party SDK dependency for production use.
- This PRD does not add wallet automation, custody, bridge, CTF split/merge,
  redemption, or token approval flows to Phase 1.
- This PRD does not define strategy algorithms, market making rules, or
  high-frequency execution targets.
- This PRD does not allow `go-bot` to keep separate Polymarket protocol clients
  once Polygolem provides the needed capability.

## Actors

- Research operator: discovers markets, prices, liquidity, and metadata.
- Paper trader: simulates orders locally against read-only market data.
- Bot developer: builds strategy services on top of typed SDK modules.
- Live operator: future actor who can place or cancel live orders only after
  all safety gates and preflight checks pass.
- Compliance/safety reviewer: verifies that secrets, geography, wallet state,
  chain config, and execution gates are correct before any dangerous operation.

## Go-Bot Integration Rule

Polygolem is the sole source of Polymarket functionality for `go-bot`.

Rules:

- `go-bot` must not call Gamma, CLOB, Data API, WebSocket, Bridge, CTF,
  relayer, signer, auth, order, trade, balance, or allowance endpoints directly.
- `go-bot` must not construct Polymarket L1/L2 auth headers, CLOB signatures,
  builder headers, order payloads, or upstream request URLs outside Polygolem.
- `go-bot` may keep temporary local Polymarket-shaped domain structs only when
  they have no network, auth, signing, URL, or protocol behavior. These should
  be treated as migration debt and renamed once Polygolem contracts stabilize.
- `POLYMARKET_*` secrets and wallet inputs may exist as operator-provided
  environment variables, but `go-bot` should pass them through to Polygolem and
  must not interpret API secrets, passphrases, private keys, or signatures
  directly.
- Upstream base URLs such as `POLYMARKET_GAMMA_URL` and
  `POLYMARKET_CLOB_URL` are Polygolem configuration. `go-bot` should prefer
  `POLYGOLEM_*` config and should not own upstream Polymarket routing.
- Paper mode, replay mode, live-readiness checks, and future live execution all
  use the same Polygolem boundary. Paper mode may simulate execution locally,
  but real Polymarket market data still enters through Polygolem.

Acceptance criteria:

- Repository checks can prove that `go-bot` has no direct references to
  `clob.polymarket.com`, `gamma-api.polymarket.com`, `NewCLOBClient`,
  `NewGammaClient`, or authenticated Polymarket request construction outside
  Polygolem, docs, fixtures, and explicitly approved compatibility shims.
- A paper-mode run fetches live order books through Polygolem or a Polygolem
  test double, never by constructing a CLOB client in `go-bot`.
- Unit tests for bot services mock Polygolem interfaces, not Polymarket HTTP
  endpoints.

Observed paper-mode issues:

- Initial issue observed on 2026-05-06: running
  `timeout 60s go run ./cmd/mega-bot paper` printed
  `CLOB: https://clob.polymarket.com`.
- That run fetched books through the old direct `go-bot` CLOB path, received
  `clob_error=up=polymarket book status 404`, falls back to
  `orderbook source=synthetic`, and blocks decisions with
  `BLOCKED: live orderbook unavailable`.
- This was the first Phase 0 migration bug. Paper mode must call a
  Polygolem-backed `BookReader` or Polygolem test double, preserve the data
  source classification, and never construct a direct Polymarket CLOB client.
- Follow-up issue observed after routing through Polygolem: paper mode now
  reports `clob_error=up=polygolem book`, but Polygolem receives
  `HTTP 404 ... No orderbook exists for the requested token id` from the CLOB
  book endpoint.
- The remaining bug is market/token resolution. Paper mode is using a default
  token for market `btc-5m-5927042` that does not have an upstream orderbook.
  Polygolem must resolve the current active market and CLOB token IDs before
  book fetch, verify the market accepts orders, and return an explicit
  `unavailable` or `unresolved_token` classification instead of letting bot
  strategy code depend on stale/default token IDs.
- Follow-up issue observed on 2026-05-07: paper mode resolved the active
  market through Polygolem, but blocked with `spread_too_wide` because the
  Polygolem `BookReader` returned raw CLOB levels in upstream order. CLOB book
  snapshots may arrive with bids low-to-high and asks high-to-low, so Polygolem
  must normalize public book levels before exposing them to `go-bot`.
- Resolved behavior required: `BookReader` returns bids best-first
  high-to-low and asks best-first low-to-high, so bot risk checks see the true
  best spread instead of a synthetic `0.01`/`0.99` wide spread.

Observed live-mode issues:

- Live run blocker observed on 2026-05-07: `go-bot live` found the repo-local
  Polygolem binary, but failed with `unknown command "clob" for "polygolem"`.
  `go-bot` requires a compatibility command surface at
  `polygolem clob book`, `tick-size`, `create-api-key`, `balance`,
  `update-balance`, `orders`, `trades`, `create-order`, and `market-order`.
- Resolved requirement: Polygolem must own all CLOB L1/L2 authentication.
  `create-api-key` creates or derives CLOB API credentials at runtime with L1
  EIP-712 auth, redacts secret/passphrase in JSON, and private CLOB reads use
  L2 HMAC headers. The HMAC secret decoder must support URL-safe and standard
  base64 encodings.
- Live run blocker observed on 2026-05-07: CLOB
  `/balance-allowance` returned raw base units. `go-bot` interpreted
  `14000000` as `14,000,000` pUSD, which inflated the order budget to the
  profile cap. Polygolem must normalize collateral and conditional CLOB amounts
  from 6-decimal base units to human units at the CLI boundary and preserve the
  raw value as `balance_raw`.
- Resolved behavior required: `polygolem clob balance --asset-type collateral
  --output json` returns human `balance`, raw `balance_raw`, and
  `balance_decimals: 6`, while allowances remain raw strings for approval
  readiness checks.
- Live run blocker observed on 2026-05-07 after auth and funding readiness
  passed: `MODEL_VALIDATION_PATH=../logs/evidence/model_validation.json`
  expired at `2026-05-07T00:00:00Z`. The live loop correctly refused all
  trades with `validation_blocked:expired` and submitted zero orders. This is a
  model/operator gate, not a Polygolem protocol blocker, and must not be
  bypassed by SDK code.
- Implemented live protocol surface requirement: Polygolem order placement
  commands must build fixed 6-decimal maker/taker amounts, derive proxy/Safe
  maker addresses from the EOA for non-EOA signature types, sign the
  Polymarket CTF Exchange EIP-712 order, post to `/order` with L2 auth, and
  preserve the CLOB response fields that `go-bot` uses for submission audit and
  reconciliation.
- Safety requirement: even with the command surface implemented, `go-bot` owns
  strategy, validation, wallet, funding, daily cap, and safe-pause gates. A
  Polygolem command must fail loudly on invalid order inputs or insufficient
  FOK liquidity instead of silently downgrading to paper mode or changing order
  type.
- Live run blocker observed on 2026-05-07 after validation/funding passed:
  `go-bot` returned `no_market` for every live decision even though current
  Polymarket crypto up/down markets existed. Root cause: the go-bot fallback
  passed an empty Gamma base URL into Polygolem, and the resolver only knew
  BTC/ETH/SOL/XRP deterministic slugs. Polygolem must default empty Gamma URLs
  to `https://gamma-api.polymarket.com` and support all live bot assets
  (`BTC`, `ETH`, `SOL`, `XRP`, `DOGE`, `BNB`) in slug generation and search
  aliases.
- Live run blocker observed on 2026-05-07 after market discovery was fixed:
  CLOB `/tick-size` can return only `minimum_tick_size`, omitting
  `tick_size` and `minimum_order_size`; CLOB `/markets/{condition_id}` can
  return token prices as numbers. Polygolem must preserve omitted fields as
  empty strings, never stringify missing JSON fields as `"<nil>"`, and decode
  numeric/string CLOB prices into stable string fields before CLI output or
  order construction.
- Live run blocker observed on 2026-05-07 after metadata parsing was fixed:
  CLOB `/order` rejected signed orders with
  `{"error":"order_version_mismatch"}`. Root cause: Polygolem was signing the
  old V1 CTF Exchange EIP-712 order (`domain.version = "1"`, exchange
  `0x4bFb...`) while production CLOB `/version` returns `{"version":2}`.
  Polygolem live order placement must query `/version` before signing, build
  V2 order payloads for version 2, sign domain version `"2"` against
  `0xE111180000d2663C0091e4f400237545B87B996B`, include V2 post fields
  `timestamp`, `metadata`, `builder`, and `expiration`, omit V1-only signed
  fields `taker`, `nonce`, and `feeRateBps`, and send explicit
  `deferExec:false` and `postOnly:false`.
- Live run blocker observed on 2026-05-07 after V2 signing was fixed:
  CLOB `/order` rejected proxy-maker submissions with
  `{"error":"maker address not allowed, please use the deposit wallet flow"}`.
  Root cause: current production CLOB requires new API users to trade through
  the deposit-wallet flow. This still uses the configured EOA private key for
  L1/L2 auth and order signing, but `signatureType = 3` (`POLY_1271`) and the
  CLOB order `maker` and `signer` fields must both be the deterministic deposit
  wallet address.
  See [DEPOSIT-WALLET-MIGRATION.md](./DEPOSIT-WALLET-MIGRATION.md) for the
  full migration survival guide, common pitfalls, and implementation status.
- Live run blocker observed on 2026-05-07 after deposit-wallet order signing
  was added: `polygolem clob balance --asset-type collateral --signature-type
  deposit` returned CLOB pUSD balance `0.000000` and zero allowances, while the
  owner EOA still holds `0.709708` pUSD. Polymarket's migration guide is
  explicit that pUSD held by the EOA does not fund deposit-wallet orders.
  go-bot now blocks before submitting orders instead of generating more lost
  rows. Remaining live requirements for this account:
  - Polygolem must expose deposit-wallet onboarding primitives, not leave them
    in go-bot: derive expected wallet, deploy through relayer `WALLET-CREATE`,
    submit signed deposit-wallet `WALLET` batches, poll relayer transaction
    state, and refresh CLOB `balance-allowance` with `signature_type = 3`.
  - The relayer requires builder authorization in this environment. A raw
    unauthenticated `POST https://relayer-v2.polymarket.com/submit` with
    `WALLET-CREATE` returned `{"error":"invalid authorization"}`. Polygolem
    must support builder auth headers from configured builder credentials
    (`BUILDER_API_KEY`, `BUILDER_SECRET`, `BUILDER_PASS_PHRASE` or the chosen
    project env names) and keep these credentials separate from CLOB L2 creds.
  - Funding must move pUSD to the deposit wallet, not merely to the owner EOA.
    Approvals must be made from the deposit wallet through a relayer `WALLET`
    batch. Existing go-bot EOA auto-approval code is not sufficient for deposit
    wallet orders.
  - go-bot should continue treating the EOA as the signer identity, but live
    funding diagnostics and order audit payloads must distinguish
    `signer_eoa` from `funder/deposit_wallet`.

## Functional Requirements

### R1. Market Discovery ✅

> **Status:** Fulfilled. Implemented in `internal/marketdiscovery` and
> `internal/gamma`, exposed via `polygolem discover search|market|enrich`.
> Identifier normalization and Gamma+CLOB enrichment are covered by tests.

The SDK must provide a `MarketDiscovery` service over Gamma and CLOB metadata.

Requirements:

- Search markets, events, tags, series, sports metadata, and public profiles
  where supported by Gamma.
- List markets and events with filters for active/closed status, slug, ID,
  condition ID, CLOB token IDs, tag, related tags, volume, liquidity, start
  date, end date, sports fields, UMA status, and reward thresholds.
- Normalize and expose the identifiers separately: Gamma market ID, event ID,
  slug, condition ID, question ID, and CLOB token IDs.
- Preserve market fields needed by trading systems: question, outcomes,
  outcome prices, best bid, best ask, spread, last trade price, liquidity,
  volume windows, order book enabled flag, accepting orders flag, minimum tick
  size, minimum order size, fee rate, close time, resolution status, tags,
  series, and negative-risk flags.
- Support pagination helpers for Gamma offset pagination and CLOB cursor
  pagination.
- Provide an enrichment path that joins Gamma market metadata with CLOB market
  details, tick size, fee rate, negative-risk status, and optional order book
  snapshots.
- Treat Gamma response quirks as first-class types, not ad hoc parsing in CLI
  command handlers.

Acceptance criteria:

- A caller can find active CLOB-enabled markets for a query and receive token
  IDs, tick size, negative-risk status, best bid/ask, and liquidity in one
  typed result.
- A caller can distinguish "market exists" from "market cannot accept orders".
- No market discovery path requires credentials.

### R2. Public CLOB Market Data ✅

> **Status:** Fulfilled. Implemented in `internal/clob` and re-exposed via
> `pkg/bookreader`; surfaced through `polygolem clob book|market|tick-size|
> price-history` and the `polygolem orderbook` group. Bid/ask normalization
> (high-to-low / low-to-high) ships behind `BookReader`.

The SDK must provide typed public CLOB data clients.

Requirements:

- Health and server time endpoints.
- CLOB market list and single market lookup by condition ID.
- Order book lookup by token ID and batch order book lookup.
- Normalize order book levels so bids are sorted high-to-low and asks are
  sorted low-to-high before returning them through public SDK interfaces.
- Price, midpoint, spread, last trade price, and batch variants.
- Tick size, fee rate, and negative-risk lookup by token ID.
- Price history by condition ID or token ID.
- Explicit request limits for batch APIs, especially last trade price batch
  sizes.
- Decimal-safe models for prices, sizes, fees, and token amounts. Floating
  point values may be used only for display or non-authoritative summaries.

Acceptance criteria:

- Public CLOB calls work in read-only mode with no signer or API key.
- Request builders validate required token IDs, market IDs, side values, and
  batch-size limits before sending network requests.

### R3. Authentication Model ✅

> **Status:** Fulfilled. `internal/auth` covers L0/L1/L2, EIP-712 CLOB auth,
> POLY_1271/ERC-7739 deposit-wallet signing, builder attribution, and
> redaction; signer abstraction lives behind injectable interfaces. See
> `polygolem auth status` and `polygolem clob create-api-key`.

The SDK must model Polymarket authentication as explicit access levels.

Requirements:

- L0: public requests with no credentials.
- L1: wallet signer for API key creation/derivation using EIP-712 CLOB auth.
- L2: API key, secret, passphrase, timestamp, and HMAC signature for private
  CLOB endpoints.
- Builder attribution: V2 uses a per-order `builderCode` (bytes32) field for
  order attribution, replacing the legacy `POLY_BUILDER_*` HMAC headers used
  in V1. Builder credentials (`BUILDER_API_KEY/SECRET/PASSPHRASE`) authenticate
  with the relayer for WALLET-CREATE and WALLET batch flows; they must not be
  confused with user L2 credentials or with the per-order `builderCode`.
- Signature types: Deposit wallet / `POLY_1271` (type 3) is the **only**
  supported mode. EOA, proxy, and Gnosis Safe are blocked by CLOB V2.
  go-bot live execution keeps the EOA as owner/signer while using the
  deposit-wallet address as maker/funder.
- Funder, signer, and maker addresses must be represented separately.
- Builder relayer auth must be a first-class, separate credential set for
  deposit-wallet `WALLET-CREATE` and `WALLET` batch flows; never reuse CLOB L2
  credentials for relayer authorization.
- Signer abstraction must support local private key signing, and leave room for
  KMS, hardware wallet, Turnkey, or remote signing implementations.
- Auth status must report readiness without printing private keys, API secrets,
  passphrases, seed phrases, raw signatures, or bearer tokens.
- Server-time synchronization must be available for timestamp-sensitive
  signatures.
- Credentials must come from explicit config, environment, or injected
  providers. No package-level globals for secret material.

Acceptance criteria:

- L1-only clients cannot call L2 endpoints.
- L2 calls fail before network I/O if API credentials or signer are missing.
- Redaction tests prove credential values never appear in status output,
  structured errors, logs, or JSON output.

### R4. Wallet And Account Readiness ✅

> **Status:** Fulfilled. `internal/wallet`, `internal/preflight`, and
> `internal/auth` separate signer-EOA from funder/deposit-wallet, expose
> CREATE2 derivation, and surface readiness via `polygolem auth status`,
> `polygolem deposit-wallet onboard`, and `polygolem health`.

The SDK must separate wallet readiness from trading execution.

Requirements:

- Report chain ID, expected network, signer address, configured signature type,
  funder/profile address, and API key readiness.
- Support proxy, Safe, and deposit wallet address derivation checks where
  implemented.
- Readiness output must distinguish `signer_eoa` from `funder_address` /
  `deposit_wallet_address`; pUSD held by one must not be reported as buying
  power for the other.
- Support close-only/ban-status checks for authenticated accounts.
- Support geoblock/readiness checks as terminal preflight inputs.
- Do not deploy wallets, approve tokens, bridge funds, or mutate on-chain state
  as part of readiness checks.

Acceptance criteria:

- `auth status`, `live status`, and `preflight` can explain which dependency is
  missing without exposing secrets or attempting a mutation.

### R5. Order Builder ✅

> **Status:** Fulfilled. `internal/orders` exposes `OrderIntent`, a fluent
> builder, validation, and lifecycle states; V2 deposit-wallet
> (`signatureType=3`) signing is in `internal/auth` + `internal/clob`.
> Surfaced via `polygolem clob create-order` and `polygolem clob market-order`.

The SDK must provide an order builder that can build signable orders without
posting them.

Requirements:

- Supported sides: BUY and SELL.
- Supported order types: GTC, GTD, FOK, and FAK.
- Required order inputs: token ID, side, price or market-order guard price,
  size or USDC amount, order type, signature type, tick size, negative-risk
  flag, fee rate, nonce, expiration where required, signer, and funder.
- V2 deposit-wallet orders must use `signatureType = 3`, `maker =
  deposit_wallet`, `signer = deposit_wallet`, and an ERC-7739-wrapped
  `TypedDataSign` signature signed by the owner EOA or approved session signer.
- Validate price range, tick-size multiple, minimum order size, side, order
  type, expiration for GTD, decimal precision, fee-rate consistency, and
  negative-risk exchange selection.
- Generate salts through an injectable salt generator for deterministic tests.
- Use decimal-safe maker/taker amount calculations and documented rounding
  rules.
- Return a signable order payload before any network submission.
- Support offline signing tests using fixed fixtures.

Acceptance criteria:

- A signed order can be built deterministically in a test with fixed signer,
  salt, timestamp, token ID, tick size, fee rate, and amount.
- Invalid tick size, invalid price, missing negative-risk metadata, and missing
  signer fail before network I/O.

### R6. Order Execution And Lifecycle ✅

> **Status:** Fulfilled. `internal/execution` separates paper from live
> executors and enforces gates; `internal/clob` covers place/cancel/query
> with batch validation. Live commands (`polygolem live ...` and the
> `polygolem clob create-order|orders|trades`) ship behind preflight gates.

The SDK must expose execution as a separate service from order building.

Requirements:

- Execution interface must support place, place batch, cancel one, cancel
  batch, cancel all, cancel by market/asset, query order, list orders, list
  trades, and builder trades where credentials allow.
- Batch validation: maximum 15 orders per placement batch and maximum 3000
  order IDs per cancellation batch.
- Full order responses must preserve order ID, market, asset ID, side, price,
  original size, matched size, owner, maker address, order type, status,
  expiration, timestamps, transaction hashes, trade IDs, and error message.
- Trade responses must preserve market, asset ID, price, size, side, status,
  fee rate, maker/taker order IDs, transaction hash, and match time.
- Execution must model lifecycle states such as created, accepted, live,
  partial, matched, canceled, rejected, failed, mined, and confirmed.
- Non-idempotent requests must not be blindly retried. Any retry policy for
  order submission must require an idempotency strategy and operator approval.
- Live execution requires mode and gate validation outside the protocol client.

Acceptance criteria:

- Paper and live execution can share an interface while using different
  implementations.
- Live execution cannot be constructed or called unless all configured safety
  gates pass.
- Order and trade response types are not lossy wrappers around the API.

### R7. Balances, Allowances, Positions, And Rewards ✅

> **Status:** Fulfilled. `internal/clob` exposes `balance`/`update-balance`
> with `signature_type=3` deposit-wallet support and human/raw decimal
> normalization; `internal/dataapi` covers positions, trades, activity,
> top holders, leaderboards, and live volume. Surfaced via
> `polygolem clob balance|orders|order|trades` and `polygolem data *`.

The SDK must expose read-oriented account state before any live trading path is
enabled.

Requirements:

- Balance/allowance lookup for USDC collateral and conditional token assets.
- Explicit asset type handling: collateral vs conditional token.
- Signature type must be included where the API requires it.
- Deposit wallet balance and allowance checks must call CLOB with
  `signature_type = 3`; EOA pUSD balance is not valid buying power for
  deposit-wallet orders.
- Allowance refresh/update calls must be classified carefully. If a call can
  cause mutation or trigger on-chain behavior, it is gated as dangerous.
- Data API support for current positions, closed positions, trades, activity,
  holders, total value, markets traded, open interest, live volume, and
  leaderboards.
- Rewards and maker scoring endpoints should be typed, but not on the critical
  path for initial order safety.

Acceptance criteria:

- A future live preflight can prove sufficient balance and allowance for a
  proposed order without placing the order.
- Account state output redacts sensitive auth material and preserves enough
  fields for operator diagnostics.

### R8. WebSocket And Streaming ⚠️

> **Status:** Partial. `internal/stream` ships a public `MarketClient`
> with reconnect, ping/pong, dedup, and `SubscribeAssets`, exposed via
> `polygolem stream market`. The authenticated user stream and RTDS-channel
> isolation described here are not yet implemented.

The SDK must provide resilient typed streaming clients.

Requirements:

- Public market stream for order book, price changes, midpoint, last trade
  price, tick size changes, best bid/ask, new market, and market resolved
  events.
- Authenticated user stream for order and trade events.
- Subscription and unsubscription APIs for token IDs and markets.
- Initial dump support where the upstream protocol supports it.
- Ping/pong heartbeat, reconnect policy, connection state reporting, shutdown,
  and context cancellation.
- Message compatibility handling for known upstream naming variants, such as
  `asset_ids` and `assets_ids`, and event type aliases.
- Optional deduplication and sequence/hash validation hooks.
- RTDS streams are optional and should be isolated from CLOB streams.

Acceptance criteria:

- Streaming code can be tested with a local WebSocket server.
- Consumers can subscribe through typed channels or managed stream objects and
  close them without leaking goroutines.
- User streams require L2 credentials and fail clearly when credentials are
  missing.

### R9. Paper Trading ✅

> **Status:** Fulfilled. `internal/paper` plus `internal/execution`'s paper
> executor share the order-intent model with live; persisted JSON state
> sits behind a storage boundary. Surfaced via the `polygolem paper`
> command group with explicit "simulated" markers.

Paper execution must remain local-only.

Requirements:

- Paper buys, sells, positions, fills, cash balance, realized PnL, and
  unrealized PnL where enough market data exists.
- Paper fills may use public market data for reference pricing, but must not
  call authenticated endpoints.
- Paper execution should use the same order intent and validation types as live
  execution where practical.
- Persist state behind a storage boundary so JSON can later move to SQLite or
  event sourcing.
- Every paper output must identify itself as simulated.

Acceptance criteria:

- Tests prove paper operations do not call authenticated mutation endpoints.
- Paper positions can be replayed from persisted state.

### R10. Safety Gates And Preflight ✅

> **Status:** Fulfilled. `internal/modes`, `internal/preflight`, and
> `internal/risk` enforce read-only/paper/live gates and structured
> blocked-error paths. Surfaced via `polygolem preflight`, `polygolem
> health`, and the `--confirm-live` + `POLYMARKET_LIVE_PROFILE` gates on
> live commands.

Safety is an SDK requirement, not only a CLI requirement.

Requirements:

- Read-only is the default and requires no credentials.
- Paper mode cannot sign, approve, place orders, cancel live orders, or call
  authenticated mutation endpoints.
- Live mode requires all gates:
  `POLYMARKET_LIVE_PROFILE=on`, `live_trading_enabled: true`,
  `--confirm-live`, and successful preflight.
- Preflight must include config validity, wallet readiness, auth readiness,
  network consistency, API health, chain consistency, geoblock/compliance
  status, balance/allowance sufficiency, and close-only status where relevant.
- Dangerous operations include real order submission, payload signing,
  on-chain transactions, token approvals, private-key handling, authenticated
  trading mutations, bridge operations, CTF split/merge/redeem, and wallet
  deployment.
- Every dangerous operation must have a structured error path when blocked.
- No code may silently downgrade a requested live action to paper or read-only.

Acceptance criteria:

- A blocked dangerous operation returns a stable machine-readable error.
- Live execution commands are available only as gated protocol primitives; the
  bot must still block before command invocation unless validation, funding,
  wallet, risk, safe-pause, and operator gates pass.

### R11. Transport, Errors, And Observability ✅

> **Status:** Fulfilled. `internal/transport` centralizes HTTP retry,
> rate limiting, circuit breaking, and redaction; `internal/errors`
> provides structured categories. All clients accept `context.Context`
> and load base URLs from `internal/config`.

The SDK must centralize network behavior.

Requirements:

- All API calls accept `context.Context`.
- Configurable base URLs for Gamma, CLOB, Data, WebSocket, and staging.
- Request timeout, user agent, rate-limit policy, retry policy, and circuit
  breaker configuration.
- Retries allowed only for safe idempotent reads by default.
- Structured errors with categories for auth, wallet, CLOB, Gamma, Data,
  transport, WebSocket, rate limit, geoblock, validation, and safety gate
  failures.
- HTTP status, endpoint family, request ID where present, and upstream error
  message should be preserved in diagnostics.
- Logs must redact secrets and avoid dumping signed payloads unless an explicit
  secure debug mode is introduced.

Acceptance criteria:

- Mock HTTP tests can assert headers, paths, query parameters, body shapes,
  status-code handling, retries, and redaction.

### R12. Public SDK Boundary ⚠️

> **Status:** Partial / drifted. The "keep everything in `internal/`"
> stance shifted: `pkg/{bookreader,bridge,gamma,marketresolver,pagination}`
> are now exposed as a small stable surface. The remaining requirements
> (thin Cobra handlers, application services above protocol clients) hold;
> see `docs/ARCHITECTURE.md` for the current package map.

The SDK surface should remain internal until stable.

Requirements:

- Keep reusable behavior under `internal/` while the CLI and tests prove the
  contracts.
- Use application services above protocol clients:
  `marketdiscovery`, `orders`, `account`, `execution`, `risk`, and `paper`.
- Introduce a public `pkg/polygolem` or `pkg/polymarket` facade only after
  internal types are stable, documented, and covered by compatibility tests.
- Avoid copying third-party SDK source. Use reference repos for protocol
  understanding and tests.
- Prefer small interfaces at integration boundaries: HTTP doer, signer,
  clock, storage, WebSocket dialer, and executor.

Acceptance criteria:

- Cobra command handlers remain thin and do not contain protocol rules,
  signing logic, order math, retry policy, or safety policy.
- SDK modules can be tested without executing the CLI binary.

### R13. Go-Bot Consumer Boundary ⚠️

> **Status:** Partial. Polygolem-side primitives exist (`pkg/bookreader`,
> `pkg/marketresolver`, `pkg/bridge`, `internal/clob`, `internal/dataapi`,
> `internal/stream`, `internal/execution`) and CLI JSON output is
> regenerated in `docs/COMMANDS.md`. Go-bot-side adoption (full removal of
> direct `internal/polymarket` clients, repository guard) is tracked
> outside this repo and is not verifiable here.

Polygolem must expose the capabilities `go-bot` needs without making the bot
know Polymarket protocol details.

Required Polygolem-backed interfaces for `go-bot`:

- `MarketDiscovery`: search and list active, closing, closed, resolved,
  CLOB-enabled, sports, tagged, and slug-based markets.
- `MarketResolver`: resolve by Gamma market ID, event ID, slug, condition ID,
  question ID, or CLOB token ID, and return all canonical identifiers.
- `BookReader`: fetch single and batch order books, best bid/ask, midpoint,
  spread, last trade price, tick size, fee rate, and negative-risk metadata.
- `PriceHistoryReader`: fetch condition-ID or token-ID price history with
  stable timestamps and decimal-safe prices.
- `AccountReader`: report auth readiness, close-only state, balances,
  allowances, positions, open orders, trades, fills, and rewards where
  credentials allow.
- `OrderExecutor`: validate, build, dry-run, place, cancel, query, and list
  orders through paper or future gated live implementations.
- `StreamSubscriber`: subscribe to public market streams and future
  authenticated user streams through typed events.
- `BridgeAndCTFReadiness`: report future bridge, collateral, split, merge,
  redeem, and funding readiness without mutating state unless a separate live
  safety plan approves it.

Contract requirements:

- Every Polygolem response consumed by `go-bot` must include a stable schema
  version, source, timestamp, request context where safe, and structured error
  category.
- Polygolem must provide JSON CLI output and in-process Go interfaces with the
  same semantics so `go-bot` can migrate from command execution to SDK calls
  without changing strategy logic.
- Polygolem must classify data as real, stale, synthetic, simulated, or
  unavailable. `go-bot` decides strategy behavior from that classification, but
  Polygolem owns the upstream fetch and validation.
- Polygolem must surface rate-limit and retry metadata so `go-bot` can slow
  down without learning endpoint-specific limits.
- Polygolem must preserve enough upstream identifiers for reconciliation:
  token ID, condition ID, market ID, order ID, trade ID, transaction hash,
  maker, owner, funder, side, price, size, status, and timestamps.

Acceptance criteria:

- `go-bot` paper mode can run with a fake Polygolem backend in tests and with
  the real Polygolem read-only backend in operator paper mode.
- Direct `internal/polymarket` protocol clients are either removed from
  `go-bot` or converted into non-network compatibility/domain types with a
  documented removal plan.
- New Polymarket API capabilities needed by `go-bot` are added to Polygolem
  first, then consumed through the Polygolem adapter.

## Proposed Architecture

Target dependency direction:

```text
Polymarket APIs
  -> Polygolem protocol SDK and CLI
  -> go-bot Polygolem adapter interfaces
  -> go-bot app services, strategies, paper mode, and future live mode
```

Recommended Polygolem internal modules:

- `internal/transport`: HTTP/WebSocket transport, rate limits, retries, errors,
  and redaction.
- `internal/gamma`: Gamma REST client and Gamma-specific response normalization.
- `internal/clob`: public CLOB data client and future private CLOB endpoint
  client.
- `internal/dataapi`: positions, activity, holders, volume, and leaderboard
  data.
- `internal/stream`: CLOB WebSocket and optional RTDS streaming lifecycle.
- `internal/auth`: access-level model, API keys, signer interfaces, L1/L2
  header builders, builder attribution, and redaction.
- `internal/wallet`: signer/funder readiness, chain consistency, proxy/Safe
  derivation checks, and non-mutating wallet diagnostics.
- `internal/marketdiscovery`: Gamma + CLOB market enrichment service.
- `internal/orders`: order intent, order builder, signed order payloads,
  order/trade models, and lifecycle states.
- `internal/execution`: paper/live executor interface, live gate enforcement,
  idempotency policy, and cancellation/query flows.
- `internal/account`: balances, allowances, rewards, and position summaries.
- `internal/risk`: per-trade caps, open-order caps, slippage limits,
  daily-loss gates, close-only handling, and circuit breakers.
- `internal/paper`: local paper state and simulated execution.
- `internal/cli`: command parsing, dependency wiring, and output only.

Recommended `go-bot` integration modules:

- `internal/polygolem`: adapter over Polygolem CLI/SDK contracts, schema
  validation, command execution, and test doubles.
- Bot-facing interfaces in app packages: market discovery, book reading, price
  history, account readiness, order execution, and stream subscription.
- No bot-facing module should own a Polymarket upstream URL, auth header,
  signer, CLOB order schema, or endpoint-specific retry policy.

## User Stories

1. As a research operator, I want to search active markets by query, so that I
   can find candidates without using credentials.
2. As a research operator, I want enriched market results with token IDs, tick
   size, fee rate, spread, liquidity, and negative-risk status, so that I can
   decide if a market is tradable.
3. As a bot developer, I want market discovery to normalize Gamma and CLOB
   identifiers, so that I do not mix up event IDs, condition IDs, and token IDs.
4. As a paper trader, I want to submit simulated orders through the same order
   intent model used by future live execution, so that paper workflows exercise
   realistic validation without sending live orders.
5. As a live operator, I want preflight to explain missing wallet, auth,
   balance, allowance, network, geoblock, or close-only readiness, so that I can
   fix the exact blocker before trading.
6. As a bot developer, I want to build and sign an order separately from
   posting it, so that I can test order math and review payloads before
   execution.
7. As a safety reviewer, I want dangerous operations blocked by explicit gates,
   so that a config or CLI mistake cannot place a live order.
8. As an operator, I want typed order and trade responses with full status and
   transaction fields, so that I can reconcile fills and diagnose failures.
9. As a streaming consumer, I want reconnecting typed market and user streams,
   so that long-running bots can recover from normal WebSocket disconnects.
10. As a maintainer, I want protocol clients isolated from Cobra, so that API
    behavior can be tested with local HTTP fixtures.
11. As a `go-bot` maintainer, I want all Polymarket reads and writes routed
    through Polygolem, so that strategy code is not coupled to upstream API
    shape, auth mechanics, or endpoint-specific failure modes.
12. As a paper-mode operator, I want real order books to come through
    Polygolem, so that paper trading exercises the same market-data boundary
    future live trading will use.

## Implementation Decisions

- Use internal modules first; defer any public SDK facade until contracts are
  stable.
- Use direct typed clients and reference repos for behavior. Do not vendor or
  copy third-party SDK code without a separate license and maintenance review.
- Preserve read-only and paper safety while defining future live-capable
  interfaces.
- Treat order building, signing, and posting as three separate steps.
- Keep signer and API key credentials separate from builder attribution
  credentials.
- Model money, prices, sizes, and fees with decimal-safe types.
- Validate batch-size, tick-size, fee-rate, order-type, and live-gate rules
  before network calls where possible.
- Keep all network behavior behind injectable transports and contexts.
- Treat Polygolem as the only Polymarket protocol owner for `go-bot`; direct
  protocol clients in `go-bot` are migration debt.
- Keep `go-bot` strategy code dependent on small Polygolem-backed interfaces,
  not on Polygolem command strings or raw JSON envelopes.
- Deprecate direct `POLYMARKET_GAMMA_URL` and `POLYMARKET_CLOB_URL` handling in
  `go-bot`; Polygolem owns upstream routing and staging selection.

## Testing Decisions

- Unit tests for order math, rounding, tick-size validation, order type
  validation, negative-risk exchange selection, and signature fixture behavior.
- Mock HTTP tests for Gamma, CLOB, Data API, auth headers, error handling,
  pagination, rate limits, and redaction.
- WebSocket tests with local servers for subscribe, unsubscribe, initial dump,
  event parsing, ping/pong, reconnect, and shutdown.
- Mode and preflight tests proving read-only, paper, and future live gates
  behave correctly.
- Paper tests proving no authenticated mutation endpoints are called.
- Golden JSON tests for CLI outputs and structured errors.
- Compatibility tests for known upstream field aliases and response shape
  differences.
- Boundary tests or repository checks proving `go-bot` does not introduce new
  direct Polymarket protocol clients, base URLs, auth header builders, or order
  payload builders outside Polygolem.
- Contract tests proving `go-bot` paper mode and market-data workflows can run
  against a fake Polygolem backend before using the real read-only backend.

## Phasing

### Phase 0 - Go-Bot Boundary Cleanup

- Fix paper mode first: replace direct CLOB book fetching with a
  Polygolem-backed book reader and add a failing test for the current bypass.
- Fix paper market/token resolution next: Polygolem must map the bot's target
  interval market to an active CLOB-enabled condition and token pair before
  fetching books, and tests must cover stale/default token IDs returning 404.
- Inventory every direct Polymarket protocol use in `go-bot`, including
  `internal/polymarket`, `POLYMARKET_GAMMA_URL`, `POLYMARKET_CLOB_URL`, bridge
  URLs, paper-mode CLOB fetches, market extraction, price history, and replay
  data paths.
- Define the minimal Polygolem adapter interfaces needed by existing bot
  workflows.
- Add TDD contract tests that fail while paper mode, extraction, and price
  history bypass Polygolem.
- Move each workflow to Polygolem-backed adapters, leaving only pure domain
  structs or fixtures outside Polygolem.
- Add a repository guard so new direct Polymarket protocol access in `go-bot`
  fails review or tests.

### Phase A - Read-Only SDK Foundation

- Complete market search, market lookup, active market listing, order book, and
  price command wiring over typed clients.
- Add market discovery enrichment over Gamma and public CLOB data.
- Add Data API read-only account/market analytics where useful for research.

### Phase B - Auth And Readiness

- Add signer interfaces, L1/L2 auth header builders, API key readiness checks,
  builder attribution model, and redaction tests.
- Add non-mutating wallet/account readiness and close-only/geoblock checks.
- Add deposit-wallet address derivation and status reporting that separates
  signer EOA, funder wallet, CLOB balance, and on-chain balance.

### Phase C - Order Domain And Paper Executor

- Add order intent, order builder, signed-order fixtures, paper execution, and
  order lifecycle models.
- Keep live posting unavailable.

### Phase D - Streams

- Add public market WebSocket streams first.
- Add authenticated user streams only after L2 auth readiness is tested.

### Phase E - Gated Live Execution

- Add live execution only through an approved plan and TDD fixtures for CLOB
  signing, L2 order posting, reconciliation reads, and blocked-gate behavior.
- Require all live gates, preflight, risk caps, balance/allowance checks,
  structured errors, and non-idempotent retry rules before enabling any real
  order placement or cancellation.
- For deposit-wallet accounts, require relayer deployment, pUSD funding to the
  deposit wallet, relayer-batch approvals from the deposit wallet, and CLOB
  `signature_type = 3` balance refresh before live order placement can start.

## Open Questions

- Should the eventual public facade be named `pkg/polygolem` or
  `pkg/polymarket`?
- Which signer backends are required first after local private-key fixtures:
  KMS, Turnkey, hardware wallet, or remote signer?
- Should account/position data come from the public Data API first, or should
  authenticated CLOB order/trade queries be prioritized?
- What minimum risk model is required before future live execution:
  per-order cap only, or per-market, per-strategy, daily-loss, and
  consecutive-error gates?
- Should WebSocket deduplication be mandatory in the first stream release or
  introduced after basic typed streams are stable?
- Should temporary `go-bot/internal/polymarket` domain structs be renamed to a
  neutral package such as `marketdata` after protocol behavior moves into
  Polygolem?
