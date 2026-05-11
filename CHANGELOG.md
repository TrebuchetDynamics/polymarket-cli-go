# Changelog

All notable changes to `polygolem` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.1.1] — 2026-05-11

### Added

- **Crypto market discovery commands.**
  - `polygolem discover crypto` — search active crypto markets by asset and
    interval (5m, 15m, 1h, 4h) with optional CLOB price/spread enrichment.
  - `polygolem discover crypto-window` — deterministic slug resolution for
    the current time-windowed market (`btc-updown-5m-<unix>`). Bypasses search
    index lag; hits the exact window directly.
  - `polygolem discover crypto-5m` — resolves all 7 active 5-minute crypto
    markets (BTC, ETH, SOL, XRP, BNB, DOGE, HYPE) in a single call with
    consolidated token IDs, condition IDs, and optional live prices.
- **Paper trading commands.**
  - `polygolem paper buy` / `polygolem paper sell` — simulate orders against
    live CLOB best ask/bid with local $10,000 starting cash.
  - `polygolem paper positions` / `polygolem paper reset` — inspect and wipe
    paper state.
  - `polygolem paper trade` — one-command workflow: resolve current window,
    fetch live price, and execute paper trade in a single step.
  - `polygolem paper crypto` — discover crypto markets and return token IDs
    ready for paper trading.
- **`CryptoWindowSlug` exported from `pkg/marketresolver`.** Deterministic
  slug generator for downstream consumers (bots, dashboards) that need to
  construct Polymarket crypto event slugs without hitting search.
- **V2 settlement gate (`pkg/settlement`, `polygolem deposit-wallet settlement-status`).**
  Read-only readiness check: deposit-wallet bytecode, relayer credentials,
  Data API positions reachability, and CTF approvals for both V2 collateral
  adapters.
- **Live E2E demo tests.** `tests/e2e_live_market_extraction_test.go`,
  `e2e_multi_market_stress_test.go`, `e2e_polygolem_demo_test.go` —
  production-validated read-only flows against live Polymarket APIs.
- **Coverage gate at 60%** with baseline tracking in CI.
- **Property-based tests** (`testing/quick`) for critical path validation.

### Changed

- **CLI root.go split** from 1,677 lines into 10 domain-specific files
  (`cmd_discover.go`, `cmd_clob.go`, `cmd_paper.go`, etc.).
- **Deposit-wallet redeem docs and CLI help** now hard-disable fallback
  thinking: V2 settlement is relayer + collateral adapter only, with no
  direct EOA, raw CTF, SAFE, or PROXY route.
- **`RELAYER_ALLOWLIST_BLOCKED`** now tells operators to verify the local
  contract registry against Polymarket's current contract reference before
  escalating. The stale upstream issue tracker is no longer surfaced as a
  current source of truth.
- **`OnboardDepositWallet` ships a 10-call post-deploy batch** (6 trading +
  4 adapter approvals) so new wallets are redeem-ready out of the box.
- **`ResolveTokenIDsAt` fails closed on slug-hit window mismatch** instead of
  silently substituting a different window.

### Fixed

- **7 Go-specific bugs** identified and fixed:
  - `hexToBytes`/`hexDecodeInto` now use `encoding/hex` with panic on invalid
    hex instead of manual parsing.
  - `buildOrderTypedData` now returns `(apitypes.TypedData, error)` with
    proper error handling.
  - WebSocket race condition fixed with `mc.mu` protecting `mc.conn`.
  - Deposit-wallet deploy false-negative: `eth_getCode` is now the source of
    truth when the relayer reports `STATE_FAILED`.
  - Position schema decode: JSON tags corrected from snake_case to camelCase
    to match live Polymarket Data API.
  - `CtfCollateralAdapter` and `NegRiskCtfCollateralAdapter` updated to
    current official Polygon addresses.
  - CLOB market buy rounding aligned with V2 expectations.

### Removed

- Removed the deprecated `deposit-wallet deploy-onchain` command and internal
  direct EOA factory deploy helper. The production deposit-wallet factory gates
  `deploy(...)` and `proxy(...)` behind `onlyOperator`, so the relayer
  `WALLET-CREATE` path is the only supported Polygolem deploy surface.
- Removed deprecated `pkg/bookreader` in favor of `pkg/orderbook`.

## [v2026.5.9] — 2026-05-09

Release version: `v0.1.0`.

### Added

- **Market-window guard (`pkg/marketresolver`).** New strict
  `ResolveTokenIDsForWindow(asset, timeframe, windowStart)` returns
  `StatusAvailable` only when the matched market's `startDate` exactly
  equals `windowStart`. New `StatusWindowMismatch` distinguishes
  wrong-window from no-market — intended as the only resolver entry
  point on the live order-placement path. `CryptoMarket` and
  `ResolveResult` now carry `StartDate`/`EndDate`.
- **V2 collateral adapter registry (`pkg/contracts`).**
  `CtfCollateralAdapter`, `NegRiskCtfCollateralAdapter`,
  `CollateralOnramp`, `CollateralOfframp`, and `PermissionedRamp`
  constants and registry fields. New helper
  `RedeemAdapterFor(negRisk bool) string` selects the right adapter
  for a market kind.
- **V2 redeem SDK (`pkg/settlement`).** `FindRedeemable`,
  `BuildRedeemCall`, `SubmitRedeem`, and the `RedeemablePosition` /
  `RedeemResult` DTOs. Calldata reuses `pkg/ctf.RedeemPositionsData`;
  the adapter ignores `collateralToken`, `parentCollectionId`, and
  `indexSets` (uses `CTFHelpers.partition()=[1,2]` internally), so we
  pass zero values. `SubmitRedeem` dedupes by `conditionID` (collapses
  YES/NO splits) and caps batches at `DefaultBatchLimit=10`.
- **Adapter approval primitives (`pkg/relayer`, `internal/relayer`).**
  `BuildAdapterApprovalCalls()` returns the 4 calls a deposit wallet
  must submit before V2 split/merge/redeem (pUSD `approve` + CTF
  `setApprovalForAll` for both V2 collateral adapters). Idempotent.
- **Operator CLI surface.**
  - `polygolem deposit-wallet approve-adapters` — one-shot migration
    for existing wallets. Dry-run by default; `--submit` requires
    `--confirm APPROVE_ADAPTERS`.
  - `polygolem deposit-wallet redeemable` — read-only list of
    redeemable positions for the deposit wallet.
  - `polygolem deposit-wallet redeem [--limit N] [--rpc-url URL]
    [--submit --confirm REDEEM_WINNERS]` — pre-checks
    `CTF.isApprovedForAll(wallet, adapter)` and refuses to sign with a
    structured pointer to `approve-adapters` if any approval is
    missing.
- **Adapter-approval pre-check (`internal/rpc.IsApprovedForAll`).**
  ERC-1155 approval check via `eth_call` (selector `0xe985e9c5`),
  used by the redeem CLI to fail-closed before signing.
- **Position V2 fields.** `pkg/types.Position` and
  `internal/dataapi.Position` now surface `Redeemable`, `Mergeable`,
  `NegativeRisk`, `Outcome`, `OutcomeIndex`, `OppositeOutcome`,
  `OppositeAsset`, `EndDate`, `Title`, `Slug`, `EventSlug`, `Icon`,
  `EventID`, `ProxyWallet`, `InitialValue`, `CurrentValue`,
  `TotalBought`, `RealizedPnl`, `PercentRealized`, `CashPnl`,
  `PercentPnl`.

### Changed

- **`OnboardDepositWallet` ships a 10-call post-deploy batch** (6
  trading + 4 adapter approvals) so new wallets are redeem-ready out
  of the box. Existing live wallets must run
  `deposit-wallet approve-adapters` once.
- **`ResolveTokenIDsAt` fails closed on slug-hit window mismatch**
  instead of silently substituting a different window.

### Fixed

- **Position schema decode bug.** `Position` JSON tags were
  snake_case (`token_id`, `condition_id`, `avg_price`, `unrealized_pnl`,
  `side`, `market_id`) but the live Polymarket Data API returns
  camelCase (`asset`, `conditionId`, `avgPrice`, `cashPnl`, no
  `side`/`market_id` at all). Existing tests round-tripped through
  their own snake_case fixtures and passed; against the real API
  every field would decode as zero. Tags now match the documented
  upstream schema; `Side`, `MarketID`, and `UnrealizedPnl` are
  removed (the API doesn't return them); `CashPnl`/`PercentPnl` take
  over from `UnrealizedPnl`.
- **Wrong-window market trap.** Live evidence on 2026-05-09: SOL
  08:20 UTC and ETH 07:40/08:00 signals filled buys against future
  market windows because `ResolveTokenIDsAt` silently substituted a
  different market when the slug-hit returned the wrong `startDate`.
  Fixed with the `StatusWindowMismatch` fail-closed signal and the
  strict `ResolveTokenIDsForWindow` entry point.
- **Deposit-wallet deploy false-negative trap.** The relayer `/deployed`
  endpoint can return `false` after a stale `WALLET-CREATE` row is marked
  `STATE_FAILED` even when the deposit wallet is fully deployed on Polygon.
  Polygolem and the go-bot SDK now treat `eth_getCode` at the derived
  deposit-wallet address as the source of truth.
  - `polygolem deposit-wallet status` falls back to `eth_getCode` when the
    relayer reports not deployed; the JSON envelope adds
    `relayerDeployed`, `onchainCodeDeployed`, and `deploymentStatusSource`,
    and renames the long-standing `wallerNonce` typo to `walletNonce`.
  - `polygolem deposit-wallet deploy --wait` checks `eth_getCode` before
    submitting `WALLET-CREATE` and exits with `state=already_deployed` when
    the wallet already has code. New `--rpc-url` flag overrides the
    Polygon RPC endpoint (default: `POLYGON_RPC_URL` env, then public node).
  - `pkg/relayer.DepositWalletAddress` and
    `pkg/relayer.DepositWalletCodeDeployed` (wraps `internal/rpc.HasCode`)
    expose the dual-source check to SDK consumers.
  - `go-bot/internal/polygolem.Client.DepositWalletStatus` treats on-chain
    code as the source of truth when the relayer reports false.

## [0.1.0] — 2026-05-07

First tagged release. Includes everything shipped through Phase 0–E plus
the May 2026 deposit-wallet migration and the documentation overhaul.

### Added

- **Builder auto — programmatic CLOB L2 credentials.** `polygolem builder auto`
  mints CLOB L2 HMAC credentials via local ClobAuth EIP-712 signing. Single
  env var (`POLYMARKET_PRIVATE_KEY`) required. See `docs/ONBOARDING.md`.
- **Universal market data client (`pkg/universal`).** Single client wrapping
  Gamma + CLOB + Data API + Discovery + Stream (70+ methods). Query all
  Polymarket public data through one typed surface.
- **Full Gamma API surface (`pkg/gamma`, 26 methods).** MarketBySlug,
  EventBySlug, SeriesByID, TagByID/TagBySlug, RelatedTagsByID/BySlug,
  Teams, CommentByID, CommentsByUser, PublicProfile, SportsMarketTypes,
  MarketByToken, EventsKeyset, MarketsKeyset.
- **CLOB V2 order management.** Cancel order (`clob cancel`), cancel all
  (`clob cancel-all`), typed `OrderRecord` and `TradeRecord` responses
  (replacing `json.RawMessage`), GTD expiration support
  (`--expiration` flag).
- **CreateBuilderFeeKey.** `POST /auth/builder-api-key` via L2 HMAC auth.
  Mints builder fee key for V2 order `builder` field attribution. Fully
  headless — no cookie, no browser.
- **SDK contracts documented.** All public types and method signatures in
  `pkg/` documented as Go interface contracts in Astro docs.
- **Polytypes reference.** V2 data types (`Market`, `Event`, `OrderBook`,
  `signedOrderPayload`, `EnrichedMarket`, `PriceHistory`, `OrderRecord`,
  `TradeRecord`, `CancelOrdersResponse`) documented with JSON field tags.
- **Deposit wallet pipeline documentation.** `docs/ONBOARDING.md`
  with full pipeline (derive → deploy → approve → fund → onboard),
  requirements checklist, gas sponsorship breakdown, replication steps.
  `docs/CONTRACTS.md` with all smart contract addresses, factory ABI,
  CREATE2 derivation, permission model, alternate deployment paths.
- **Astro docs site (25+ pages).** Guides (Builder Auto, Universal Client,
  Market Discovery, Deposit Wallet Lifecycle, Orderbook Data, Paper Trading,
  Bridge & Funding, Go-Bot Integration), Concepts (API Overview, Smart
  Contracts, POLY_1271 Deposit Wallets, Secrets, Markets/Events/Tokens,
  Safety, Architecture), Reference (CLI, Go SDK Contracts, Protocol Types,
  Internal Packages, Gamma/CLOB/Data/Stream APIs, Coverage Matrix).
- **Polydart PRD.** `PRD_POLYDART.md` — companion Dart SDK design for
  Arenaton Flutter with Reown/WalletConnect, server proxy, confirmed
  pipeline.
- **Test coverage.** Added tests for `internal/errors`,
  `internal/marketdiscovery`, `internal/stream`. 29/29 packages pass
  CI (gofmt + vet + test).
- **Orderbook taxonomy.** `pkg/orderbook` re-exports with typed reader
  interface from `pkg/bookreader`.

### Changed

- **CLOB API reference updated for V2.** Accurate commands, POLY_1271
  signing flow, ERC-7739 TypedDataSign wrapper documentation, V2 order
  envelope fields.
- **Safety model extended for deposit wallet V2.** Signer vs funder
  separation, builder credential isolation, deposit-wallet balance routing,
  relayer auth vs trading auth rules.
- **Architecture updated.** 6 `pkg/` + 21 `internal/` packages documented
  with dependency direction diagram.
- **README rewritten.** One env var focus, accurate command inventory,
  builder auto front-and-center, SDK tables, docs links.
- **Credential documentation.** Split three credential types: CLOB L2 Trading Key
  (headless for existing users), Builder Fee Key (headless via L2 HMAC), Relayer API Key
  (headless via SIWE). See `docs/ONBOARDING.md`.

[Unreleased]: https://github.com/TrebuchetDynamics/polygolem/compare/v0.1.1...HEAD
[v0.1.1]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v0.1.1
[v2026.5.9]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v2026.5.9
[0.1.0]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v0.1.0
