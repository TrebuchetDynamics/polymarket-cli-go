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

## Deposit Wallet Safety

Polymarket requires deposit wallet (POLY_1271 / signature type 3) for all
trading. EOA, proxy, and Gnosis Safe are blocked by CLOB V2. This
introduces safety rules beyond the direct EOA model.

### Signer vs Funder Separation

The EOA remains the signing key. The deposit wallet is the funder (the address that
holds pUSD and is the CLOB `maker`/`signer`). These must never be confused:

- **EOA pUSD does NOT fund deposit-wallet orders.** CLOB reads the deposit wallet's balance.
- **Approvals must come from the deposit wallet** via relayer WALLET batch, not from the EOA.
- **Diagnostics must distinguish** `signer_eoa` from `funder/deposit_wallet` in logs and audit records.

### Builder Credential Isolation

Builder credentials (`BUILDER_API_KEY/SECRET/PASSPHRASE`) authenticate with the
relayer for WALLET-CREATE and WALLET batch operations. These are a separate auth
system and must never be:

- Reused as CLOB L2 credentials
- Stored alongside CLOB API keys in the same config section
- Added to order-signing headers (orders use `builderCode`, not builder HMAC)

### Relayer Auth vs Trading Auth

- **Relayer auth**: Builder HMAC credentials → used for wallet lifecycle operations
- **Trading auth**: CLOB L1/L2 credentials → used for order placement and balance queries
- These systems are independent. A failed relayer call must not be retried as a CLOB call.

### Deposit-Wallet Balance Routing

When `signature_type = 3` (deposit), the CLOB balance endpoint returns the
**deposit wallet's pUSD balance**, not the EOA's. Live readiness must:

1. Check the CLOB balance with `signature_type = 3`
2. Verify the deposit wallet is deployed (relayer `/deployed` endpoint)
3. Verify collateral allowances are non-zero
4. Block before order submission if any check fails

See [DEPOSIT-WALLET-MIGRATION.md](./DEPOSIT-WALLET-MIGRATION.md) for the full onboarding flow,
common pitfalls, and recovery steps.

## Deposit Wallet Safety Rules

The May 2026 deposit-wallet migration introduces a new family of commands
(`polygolem deposit-wallet *`) that perform on-chain or relayer-bound
operations. These rules apply.

1. **Builder credentials are required and redacted.** `--deploy`,
   `--batch`, `--approve --submit`, and `--onboard` require
   `POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`, and
   `POLYMARKET_BUILDER_PASSPHRASE`. Configuration loading redacts all three
   on every load; no command emits them in JSON output.

2. **Read-only deposit-wallet commands stay read-only.**
   `deposit-wallet derive`, `deposit-wallet status`, and
   `deposit-wallet nonce` perform no on-chain or relayer mutations.

3. **Batch signing requires explicit calldata input.**
   `deposit-wallet batch --calls-json` requires structured input. The CLI
   does not synthesize calls. The `approve` shortcut shows calldata before
   submission unless `--submit` is passed.

4. **Funding moves real money.** `deposit-wallet fund --amount X`
   transfers ERC-20 pUSD from the EOA to the deposit wallet via direct RPC.
   The amount must be specified explicitly. There is no default.

5. **Onboarding is the only multi-step composite.**
   `deposit-wallet onboard --fund-amount X` performs deploy → approve →
   fund. Each step is gated; failure of any step aborts the composite and
   leaves the wallet in a recoverable state visible to
   `deposit-wallet status`.

6. **POLY_1271 orders use the deployed wallet's signature path.**
   `clob create-order` and
   `clob market-order` sign with the deposit
   wallet's POLY_1271 path. Orders signed without the deposit signature
   type after the May 2026 cutoff will be rejected by Polymarket for
   new accounts.

7. **Builder attribution does not bypass safety.** Setting builder
   credentials enables deposit-wallet operations; it does not relax any
   gate or grant trading privileges.
