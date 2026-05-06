# Architecture

`polymarket-cli-go` is a Go protocol and automation stack with a CLI frontend.
The Cobra command tree is a thin frontend over typed, testable internal
packages. The Rust CLI is retained only as a behavioral reference in
`docs/REFERENCE-RUST-CLI.md`.

## Package Boundaries

- `cmd/polymarket-cli-go`: binary entry point and process exit handling.
- `internal/cli`: Cobra command construction and dependency wiring.
- `internal/config`: explicit Viper-backed config loading, defaults,
  environment binding, validation, and redaction.
- `internal/modes`: read-only, paper, and live-mode parsing and gate checks.
- `internal/preflight`: local and remote readiness checks.
- `internal/output`: stable table and JSON rendering plus structured errors.
- `internal/gamma`: typed read-only Gamma HTTP client.
- `internal/clob`: typed read-only CLOB HTTP client.
- `internal/paper`: local-only paper positions, fills, and persisted state.

## Dependency Flow

The intended flow is:

```text
protocol clients -> application services -> thin Cobra CLI
```

The package-level dependency direction is:

```text
cmd/polymarket-cli-go
        |
internal/cli
        |
config, modes, preflight, output
        |
gamma, clob, paper
```

Command handlers parse flags, call package APIs, and render output. Protocol
clients do not know about Cobra. Safety packages do not depend on command text.
Paper state stays local and does not call live mutation endpoints.

Cobra command handlers must not contain protocol or trading business logic.
That logic belongs in typed clients, application services, safety gates, and
paper-state packages where it can be tested without executing the binary.

## Mode System

Mode selection starts in configuration and CLI flags, then flows through
`internal/modes` before command handlers call protocol clients or paper state.
Command handlers should pass the selected mode into application services rather
than deciding safety policy inline.

Read-only mode permits public market data and forbids signing or mutations. It
is the default mode and may use `internal/gamma`, `internal/clob`, and
`internal/output` for public data retrieval and rendering.

Paper mode permits local simulation and forbids live endpoints. It may combine
read-only reference data with `internal/paper` state, but simulated actions must
remain local and must not reach authenticated mutation APIs.

Live mode is disabled unless every gate passes. In Phase 1, live mode is a
status and validation surface only: `internal/config`, `internal/modes`, and
`internal/preflight` can explain gate state, but no package should execute real
trading or on-chain operations.

## Phase 1 SDK Boundary

There is no public Go SDK in Phase 1. Reusable behavior remains under
`internal/` until the stable package surface is proven by CLI use and tests.
The repository intentionally avoids a `pkg/` API until a future phase defines a
supported SDK contract.

## Safety Boundaries

Read-only market data and local paper state are the only operational surfaces in
Phase 1. Live-capable execution is represented by status and gate validation,
not by order submission or on-chain mutation code.
