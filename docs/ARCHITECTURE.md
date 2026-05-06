# Architecture

`polymarket-cli-go` is a Go Phase 1 CLI with a thin command layer over typed,
testable internal packages. The Rust CLI is retained only as a behavioral
reference in `docs/REFERENCE-RUST-CLI.md`.

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

## Phase 1 SDK Boundary

There is no public Go SDK in Phase 1. Reusable behavior remains under
`internal/` until the stable package surface is proven by CLI use and tests.
The repository intentionally avoids a `pkg/` API until a future phase defines a
supported SDK contract.

## Safety Boundaries

Read-only market data and local paper state are the only operational surfaces in
Phase 1. Live-capable execution is represented by status and gate validation,
not by order submission or on-chain mutation code.
