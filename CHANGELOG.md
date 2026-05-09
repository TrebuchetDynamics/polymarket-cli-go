# Changelog

All notable changes to `polygolem` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/TrebuchetDynamics/polygolem/compare/v2026.5.9...HEAD
[v2026.5.9]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v2026.5.9
[0.1.0]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v0.1.0
