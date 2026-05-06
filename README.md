# polymarket-cli-go

`polymarket-cli-go` is a Go Phase 1 command-line foundation for safe
Polymarket research and automation.

Phase 1 focuses on:

- read-only market, order book, and price access
- local-only paper trading state
- explicit live-mode status and gate checks
- stable table and JSON output foundations
- testable internal Go packages

It does not submit real orders, sign transactions, or perform on-chain
mutations in Phase 1.

## Current Scope

Implemented foundation packages include the Cobra CLI shell, config loading,
execution modes, preflight checks, structured output, read-only Gamma and CLOB
clients, and local paper state.

The binary command name is `polymarket`. The repository module name is
`github.com/TrebuchetDynamics/polymarket-cli-go`.

## Safety Model

Read-only behavior is the default. Paper trading uses local persisted state
only. Live execution remains hard-disabled until a future phase implements and
tests every required gate.

Future live-capable commands must require:

- `POLYMARKET_LIVE_PROFILE=on`
- `live_trading_enabled: true`
- `--confirm-live`
- successful `preflight`

If any gate fails, the command must abort with a structured error. It must not
quietly switch to paper or read-only behavior.

## Verification

Run the focused documentation safety check:

```bash
go test ./tests -run TestDocumentationSafety -count=1
```

Run the full test suite:

```bash
go test ./...
```

The documentation safety test also checks for unsafe upstream or live-execution
claims that do not apply to this Go Phase 1 repository.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Commands](docs/COMMANDS.md)
- [Safety](docs/SAFETY.md)
- [Rust reference audit](docs/REFERENCE-RUST-CLI.md)
