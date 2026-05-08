# Changelog

All notable changes to `polygolem` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/TrebuchetDynamics/polygolem/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v0.1.0
