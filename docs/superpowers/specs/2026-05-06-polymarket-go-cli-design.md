# Polymarket Go CLI Design

## Status

Approved by Juan on 2026-05-06.

## Mission

Build a clean, idiomatic Go replacement for the official Rust Polymarket CLI.
The Rust repository is the behavioral and protocol reference only. The Go
implementation must not translate Rust source line-by-line.

The project name is `polymarket-cli-go`. The repository will become a clean Go
repository rather than a mixed Rust/Go tree.

## Hard Engineering Rule

All implementation work is test-driven.

No production Go code may be written before a failing test demonstrates the
required behavior. Each implementation task must follow this loop:

1. Write a focused failing test.
2. Run the test and confirm the failure is expected.
3. Write the minimum production code needed to pass.
4. Run the focused test and broader relevant tests.
5. Refactor only while tests remain green.

This rule applies to features, bug fixes, refactors, safety gates, config
loading, command behavior, protocol clients, output formatting, and paper
trading behavior.

Generated files, dependency metadata, and documentation are not production Go
logic, but implementation plans must still include verification for them.

## Primary Goals

- Provide public market data access.
- Provide safe local-only paper trading simulation.
- Preserve a path to future live trading without enabling it by accident.
- Make every command automation-friendly with stable JSON output.
- Keep the CLI as a thin wrapper over typed, testable internal packages.
- Build maintainable infrastructure that can later support bots, daemons, and a
  reusable SDK.

## Non-Goals For Phase 1

- No real live order placement.
- No on-chain transaction submission.
- No automated market making.
- No high-frequency execution.
- No leverage.
- No custodial features.
- No browser automation.
- No strategy engine.
- No websocket streaming layer.
- No historical replay engine.
- No multi-account orchestration.

The design intentionally avoids turning Phase 1 into a full trading platform.

## Repository Direction

The cloned Rust source will be replaced by a Go module. Rust behavior will be
preserved in documentation at `docs/REFERENCE-RUST-CLI.md`.

The upstream Rust repository is MIT licensed, but the Go implementation will use
behavioral parity and protocol research, not copied source.

## Architecture

The long-term shape is:

```text
typed protocol clients
        |
reusable Go SDK boundary
        |
thin Cobra CLI
        |
future bots/services
```

The CLI must not become the core domain model. Command handlers parse flags,
call application services, and render output. Protocol rules, safety checks,
HTTP behavior, config validation, and paper trading state live behind packages
that can be tested without executing the CLI binary.

Use interfaces at integration boundaries, especially for HTTP clients, storage,
clock/time, and future signing/auth. Avoid interface pollution inside leaf
packages where concrete types are clearer.

## Package Boundaries

Phase 1 will use these package responsibilities:

- `cmd/polymarket-cli-go/main.go`: binary entry point only.
- `internal/cli`: Cobra command construction and dependency wiring.
- `internal/config`: Viper-backed config loading, defaults, environment
  binding, validation, and safe redaction.
- `internal/modes`: execution mode parsing and enforcement.
- `internal/preflight`: local and remote readiness checks.
- `internal/output`: stable table and JSON rendering plus structured errors.
- `internal/gamma`: typed read-only Gamma API client.
- `internal/clob`: typed read-only CLOB API client.
- `internal/markets`: market application services over Gamma/CLOB clients.
- `internal/orders`: order domain types and future live-trading service
  boundary. Phase 1 keeps live order execution blocked.
- `internal/paper`: local-only simulated orders, positions, fills, and PnL
  state.
- `internal/auth`: auth status and future API credential boundary.
- `internal/wallet`: wallet readiness checks and future signer boundary.
- `pkg/polymarket`: deferred until an actually reusable public SDK surface
  emerges.

## Execution Modes

The CLI has three execution modes.

### Read-Only Mode

Read-only is the default.

Allowed:

- market discovery
- market metadata
- prices
- order books
- public API health

Forbidden:

- signing
- order placement
- live mutations
- on-chain transactions

Read-only mode must not require wallet credentials.

### Paper Mode

Paper mode uses local persisted state only.

Allowed:

- simulated buys
- simulated sells
- local positions
- local PnL
- local reset

Forbidden:

- live order endpoints
- signing
- on-chain transactions
- authenticated trading mutations

Paper commands may call read-only market data endpoints for reference prices,
but must never call live trading endpoints.

### Live Mode

Live mode is hard-disabled by default.

Future live trading requires all gates:

- `POLYMARKET_LIVE_PROFILE=on`
- config value `live_trading_enabled: true`
- explicit CLI flag `--confirm-live`
- successful preflight validation

Preflight must validate:

- config validity
- wallet readiness
- auth readiness
- network health
- API health
- chain/network consistency

If any gate fails, the command aborts with a structured error. The CLI must not
silently downgrade to paper or read-only behavior.

Phase 1 implements live status and gate validation, but not live order
placement.

## Commands In Phase 1

The binary command name is `polymarket`.

Core:

- `polymarket version`
- `polymarket preflight`

Markets:

- `polymarket markets search`
- `polymarket markets get`
- `polymarket markets active`

Market data:

- `polymarket orderbook get`
- `polymarket prices get`

Paper trading:

- `polymarket paper buy`
- `polymarket paper sell`
- `polymarket paper positions`
- `polymarket paper reset`

Status:

- `polymarket auth status`
- `polymarket live status`

Every command supports `--json`. JSON output must be stable and covered by
golden tests.

## Config Design

Use Viper through explicit instances, not package-level global state.

Config sources:

- defaults
- explicit config path flag
- standard user config path
- environment variables
- command flags

Initial config fields:

- `mode`
- `gamma_base_url`
- `clob_base_url`
- `request_timeout`
- `live_trading_enabled`
- `paper_state_path`

Credential-bearing values must not be printed unless redacted. Phase 1 should
avoid requiring credentials for read-only and paper workflows.

## Networking Design

All external API calls use:

- `context.Context`
- request timeouts
- typed request/response structs
- structured errors
- status-code-aware error handling

Retries are allowed only for safe idempotent read requests. Retry behavior must
be explicit and covered by tests.

The clients must expose enough response detail for diagnostics without hiding
HTTP failures.

## Output Design

`internal/output` owns human and JSON rendering.

JSON output must be:

- stable
- machine-readable
- versionable
- tested with golden fixtures

Errors in JSON mode use structured payloads with fields such as:

- `error.code`
- `error.message`
- `error.details`

Table output may be optimized for humans, but must not suppress failures or
hide critical safety information.

## Paper Trading Design

Paper trading is deliberately simple in Phase 1.

It records:

- simulated orders
- simulated fills
- current positions
- cash balance
- realized and unrealized PnL where enough data exists

The first implementation uses local JSON state. State storage lives behind a
small storage boundary so it can later move to SQLite or event sourcing.

Paper trading must be documented as simulation, not realistic execution. It
does not model queue position, slippage, partial fills, or latency in Phase 1.

## Rust CLI Audit Requirements

Create `docs/REFERENCE-RUST-CLI.md` before implementing protocol-dependent
behavior.

The audit must document:

- command structure
- auth model
- config format
- signing logic
- API endpoints
- market capabilities
- order and trading flows
- dependencies
- architectural problems
- maintenance concerns
- protocol assumptions
- license constraints

Initial observations from the cloned reference:

- Rust package name: `polymarket-cli`.
- Binary name: `polymarket`.
- License: MIT.
- Main dependency: `polymarket-client-sdk` with `gamma`, `data`, `bridge`,
  `clob`, and `ctf` features.
- CLI framework: Clap.
- Wallet config path: `~/.config/polymarket/config.json`.
- Private key priority: CLI flag, environment variable, config file.
- Signature types: `proxy`, `eoa`, `gnosis-safe`.
- Current Rust CLI includes authenticated CLOB order creation, cancellation,
  balance, rewards, API key, CTF, bridge, and wallet commands.
- The Go Phase 1 intentionally does not implement those live mutation paths.

## Testing Strategy

Required test categories:

- table-driven unit tests
- Cobra command execution tests
- config validation tests
- mode enforcement tests
- preflight tests
- mock HTTP server tests
- fixture parsing tests
- golden JSON output tests
- paper state tests

Required verification commands:

```bash
gofmt -w .
go vet ./...
go test ./...
```

Every implementation task in the plan must include:

- the failing test to write first
- the command that proves it fails
- the minimal implementation step
- the command that proves it passes
- the broader verification command when appropriate

## Documentation Deliverables

Phase 1 creates:

- `docs/REFERENCE-RUST-CLI.md`
- `docs/ARCHITECTURE.md`
- `docs/COMMANDS.md`
- `docs/SAFETY.md`

These docs must stay aligned with implemented behavior. They must not claim live
trading support before it exists and has been separately approved.

## Phase 1 Deliverables

- Rust CLI audit
- architecture documentation
- Go module skeleton
- Cobra CLI skeleton
- config system
- execution mode system
- preflight system
- read-only Gamma and CLOB clients
- read-only market commands
- stable JSON output support
- paper trading skeleton with local-only state
- tests and fixtures
- CI-ready verification commands

## Known Risks

- Polymarket APIs may drift or expose undocumented behavior.
- The Rust CLI supports live mutation paths that Phase 1 intentionally blocks.
- Paper trading can create false confidence if treated as realistic execution.
- Scope can expand into a full trading platform if future phases are not
  tightly controlled.
- AI-generated protocol code is risky around auth, signing, decimal handling,
  and order construction, so those areas require especially strict TDD,
  fixtures, and manual review.

## Acceptance Criteria

Phase 1 is acceptable only when:

- production Go code was written through TDD red-green-refactor loops
- read-only mode works without credentials
- paper mode never calls live trading endpoints
- live mode cannot mutate anything and reports blocked status clearly
- all implemented commands support `--json`
- structured errors are returned in JSON mode
- `gofmt -w .`, `go vet ./...`, and `go test ./...` have been run
- docs describe actual behavior without unsupported claims
