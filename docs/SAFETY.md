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

The EOA remains the cryptographic signing key for the ERC-7739 wrapper. The
deposit wallet is the CLOB account: it holds pUSD and appears as both order
`maker` and order `signer` in the V2 payload. These must never be confused:

- **EOA pUSD does NOT fund deposit-wallet orders.** CLOB reads the deposit wallet's balance.
- **Approvals must come from the deposit wallet** via relayer WALLET batch, not from the EOA.
- **Diagnostics must distinguish** `owner_eoa` from `funder/deposit_wallet` in logs and audit records.

### Builder Credential Isolation

V2 relayer keys (`RELAYER_API_KEY` / `RELAYER_API_KEY_ADDRESS`) and legacy
builder HMAC credentials (`BUILDER_API_KEY/SECRET/PASSPHRASE`) authenticate
with the relayer for WALLET-CREATE and WALLET batch operations. These are a
separate auth system and must never be:

- Reused as CLOB L2 credentials
- Stored alongside CLOB API keys in the same config section
- Added to order-signing headers (orders use `builderCode`, not builder HMAC)

### Relayer Auth vs Trading Auth

- **Relayer auth**: V2 relayer key or builder HMAC credentials → used for wallet lifecycle operations
- **Trading auth**: CLOB L1/L2 credentials → used for order placement and balance queries
- These systems are independent. A failed relayer call must not be retried as a CLOB call.

### Deposit-Wallet Balance Routing

When `signature_type = 3` (deposit), the CLOB balance endpoint returns the
**deposit wallet's pUSD balance**, not the EOA's. Live readiness must:

1. Check the CLOB balance with `signature_type = 3`
2. Verify the deposit wallet is deployed. Polygon `eth_getCode` is the source
   of truth; relayer `/deployed` is advisory and can return false while the
   contract already has bytecode.
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
   new accounts. Readiness must verify non-empty bytecode at the deposit
   wallet address before order submission.

7. **Builder attribution does not bypass safety.** Setting builder
   credentials enables deposit-wallet operations; it does not relax any
   gate or grant trading privileges.

8. **Decision-window safety.** Automated order placement must bind the
   strategy decision window to the selected market window. A signal for
   `2026-05-09T08:20:00Z` must not buy a market that starts at
   `2026-05-09T12:20:00Z`, even when the asset and timeframe match. The
   required SDK path is a strict window resolver that returns a
   `window_mismatch` status instead of silently falling back to a future
   market.

## Matched, Winning, And Redeemable

Matched, winning, and redeemable are separate states:

- `matched`: the CLOB filled the order and transferred or minted position
  shares.
- `winning`: the market resolved and the held token is the winning outcome.
- `redeemable`: Polymarket's Data API reports the held position can be
  redeemed.

A matched order is not proof that the market won. Redemption automation must
read the deposit wallet's Data API `/positions` rows and use
`redeemable=true` as the readiness signal. The current position schema exposes
`redeemable`, `mergeable`, `negativeRisk`, `outcome`, `outcomeIndex`,
`oppositeOutcome`, `oppositeAsset`, and `endDate`; it does not expose a
separate `resolved` boolean.

## V2 Redeem Readiness

Polymarket V2 uses collateral adapter contracts for pUSD-native CTF actions.
For deposit-wallet positions this path is non-negotiable: the owner signs an
EIP-712 WALLET batch, the relayer submits it through the deposit-wallet
factory, and the wallet call targets:

- `CtfCollateralAdapter` for standard binary markets:
  `0xAdA100Db00Ca00073811820692005400218FcE1f`
- `NegRiskCtfCollateralAdapter` for negative-risk markets:
  `0xadA2005600Dec949baf300f4C6120000bDB6eAab`

Do not call `ConditionalTokens.redeemPositions` directly from a V2
deposit-wallet flow. The adapter reads `conditionId`, detects the wallet's
current CTF balances, redeems through the underlying CTF path, wraps proceeds
back into pUSD, and returns pUSD to the deposit wallet.

SAFE and PROXY examples in upstream relayer clients are not deposit-wallet
precedent. Deposit wallets use `executeDepositWalletBatch(...)` / relayer
`WALLET` transactions, not the SAFE/PROXY `execute(...)` shortcut.

Adapter readiness is distinct from trading readiness. The existing trading
approval batch covers CLOB exchange spenders. V2 redeem requires the deposit
wallet to approve the collateral adapters with CTF `setApprovalForAll`; the
one-time adapter approval batch should also include pUSD `approve` for future
split support. Existing live wallets that only ran the trading approval batch
need a one-shot adapter-approval migration before their first V2 redeem.

The first-class `polygolem deposit-wallet approve-adapters`, `redeemable`,
and `redeem` commands build the V2 adapter path (commits `c77e735` and
`0593991`). Every signing path defaults to dry-run; submission requires both
`--submit` and a typed `--confirm` token (`APPROVE_ADAPTERS` for adapter
approvals, `REDEEM_WINNERS` for redeem). The redeem command runs an
`isApprovedForAll(wallet, adapter)` pre-check via `eth_call` and refuses to
sign if any approval is missing — the relayer never sees `/submit` when the
pre-check fails.

If the relayer rejects adapter approval or redeem calls as "not in the allowed
list", stop and treat it as an upstream relayer allowlist blocker. The
production `DepositWalletFactory.proxy()` entrypoint is `onlyOperator`, so the
owner EOA cannot bypass the relayer, and raw `ConditionalTokens.redeemPositions`
is not a V2 deposit-wallet fallback.
