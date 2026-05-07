# Changelog

All notable changes to `polygolem` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — 2026-05-07

First tagged release. Includes everything shipped through Phase 0–E plus
the May 2026 deposit-wallet migration and the documentation overhaul.

### Added

- **Phase 0 — Go-bot boundary cleanup.** Polymarket protocol access moved
  out of `go-bot` and into a single Go module owned by `polygolem`.
- **Phase A — Read-only SDK foundation.** `internal/gamma`,
  `internal/clob` (read endpoints), `internal/dataapi`,
  `internal/marketdiscovery`, `internal/output`, `internal/transport`,
  and the public `pkg/bookreader`, `pkg/marketresolver`, `pkg/bridge`,
  `pkg/gamma`, `pkg/pagination` packages.
- **Phase B — Auth and readiness.** `internal/auth` (L0/L1/L2, EIP-712,
  builder attribution), `internal/config` (Viper-backed loading with
  redaction), `internal/preflight`, and `internal/modes`
  (read-only / paper / live).
- **Phase C — Orders and paper executor.** `internal/orders` (OrderIntent,
  fluent builder, validation, lifecycle states), `internal/execution`
  (paper executor), `internal/paper` (local-only persisted state),
  `internal/risk` (per-trade caps, daily loss limits, circuit breaker).
- **Phase D — Streams.** `internal/stream` WebSocket market client with
  reconnect and dedup.
- **Phase E — Gated live execution.** Live execution path gated by
  preflight + risk + funding checks; CLOB write endpoints accessible only
  with explicit signature type and gates passing.
- **Deposit-wallet migration (May 2026).** `internal/wallet`,
  `internal/relayer`, `internal/rpc`, and the `polygolem deposit-wallet *`
  command family — `derive`, `deploy`, `nonce`, `status`, `batch`,
  `approve`, `fund`, `onboard`. POLY_1271 order signing via
  `--signature-type deposit`.
- **CLI surface.** Cobra-based commands across `auth`, `bridge`, `clob`,
  `discover`, `events`, `health`, `live`, `orderbook`, `paper`,
  `preflight`, `version`, and `deposit-wallet` groups. Every command
  accepts `--json`.
- **Documentation overhaul (this release).**
  - Track 1 — audit & truth pass: `docs/ARCHITECTURE.md` rewritten,
    `docs/COMMANDS.md` regenerated from `--help`, `docs/PRD.md`
    annotated, `docs/SAFETY.md` extended with deposit-wallet rules,
    stale planning docs archived to `docs/history/`.
  - Track 2 — godoc on every exported identifier in `pkg/` and on
    package-level docs in `internal/`.
  - Track 3 — `SKILL.md` agent surface and the v1 JSON output contract.
  - Track 4 — Astro Starlight docs site under `docs-site/`.
  - Track 5 — `CONTRIBUTING.md`, `SECURITY.md`, this `CHANGELOG.md`,
    GitHub issue and PR templates, Dependabot config, README badges.

[Unreleased]: https://github.com/TrebuchetDynamics/polygolem/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v0.1.0
