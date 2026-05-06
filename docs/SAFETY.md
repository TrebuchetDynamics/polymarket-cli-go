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

## Preflight

Preflight checks config validity, wallet readiness, auth readiness, network consistency, API health, and chain consistency.
It aggregates local configuration, credential readiness, remote API reachability,
and expected network identity into one pass/fail result.

Automation must treat any preflight failure as terminal. A failed preflight
means the requested operation is not safe to continue, and scripts should stop
instead of retrying a different mode or assuming a partial result is usable.

## Failure Behavior

If any gate fails, the command must abort with a structured error and a non-zero
exit code. The CLI must not silently downgrade to paper mode or read-only mode,
because that would hide operator intent and make automation unsafe.

## Dangerous Operations

Dangerous operations include real order submission, payload signing, on-chain transactions, token approvals, private-key handling, and authenticated trading mutations.
Phase 1 intentionally contains no code path for those operations.

Future work that introduces any dangerous operation must add explicit tests,
structured errors, credential redaction, preflight coverage, and live-gate
enforcement before it is exposed through the CLI frontend.

## Credential Handling

Read-only and paper workflows should not require private keys. Any future
credential-aware status output must redact sensitive values and report readiness
without printing secrets.
