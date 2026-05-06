# Safety

Phase 1 is safe by default. The CLI supports read-only workflows and local paper
state while keeping real execution hard-disabled.

## Read-Only Default

Read-only mode is the default. It may call public market-data APIs for markets,
order books, prices, and health checks. It must not require wallet credentials,
sign payloads, or submit mutations.

## Paper Mode

Paper mode is local-only. Simulated buys, simulated sells, positions, and reset
operations are stored in local state. Paper behavior must not call authenticated
trading endpoints or on-chain transaction paths.

## Live Gates

Future live-capable commands require all gates:

- `POLYMARKET_LIVE_PROFILE=on`
- `live_trading_enabled: true`
- `--confirm-live`
- successful `preflight`

All four gates must pass before any future live-capable command may proceed.
Phase 1 implements status and validation boundaries only; it does not implement
real execution.

## Failure Behavior

If any gate fails, the command must abort with a structured error and a non-zero
exit code. The CLI must not silently downgrade to paper mode or read-only mode,
because that would hide operator intent and make automation unsafe.

## Credential Handling

Read-only and paper workflows should not require private keys. Any future
credential-aware status output must redact sensitive values and report readiness
without printing secrets.
