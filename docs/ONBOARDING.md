# Deposit Wallet Onboarding - Single Source of Truth

**Last updated:** 2026-05-10
**Status:** Production - verified with live funds on Polygon mainnet
**Companion:** [BROWSER-SETUP.md](./BROWSER-SETUP.md) - manual signing fallback only

## TL;DR

Polymarket login signs with the EOA. The deposit wallet is not the website
login signer; it is the V2 trading wallet.

| Identity | Role |
|---|---|
| EOA | Signs SIWE login, ClobAuth HTTP auth, relayer WALLET batches, and POLY_1271 order wrappers. |
| Deposit wallet | Holds pUSD, is the CLOB order maker/signer, receives CTF positions, approves adapters, and redeems winners. |

For the account in the current bot env, polymarket.com showing the EOA
`0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C` in the sign-in prompt is
expected. Polygolem derives the corresponding deposit wallet separately.

## Standard Headless Flow

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

# 1. Deposit wallet deploy + relayer auth + approvals + UI Enable Trading signs + funding.
polygolem deposit-wallet onboard --fund-amount 0.71

# 2. Sync CLOB cache and verify readiness.
polygolem clob update-balance --asset-type collateral
polygolem deposit-wallet status --check-enable-trading
polygolem auth status --check-deposit-key

# 3. Trade with the deposit wallet path.
polygolem clob create-order --token <TOKEN_ID> --side buy --price 0.5 --size 10
```

`deposit-wallet onboard` and `deposit-wallet enable-trading` first load V2
relayer credentials from env or Polygolem env files. If none exist, they run
the SIWE login/profile/key-mint flow automatically, persist the relayer key to
the 0600 env file, and continue with the WALLET batch. `polygolem auth login`
is still useful when you want to inspect or refresh authentication explicitly,
but it is no longer a required manual pre-step for the deposit-wallet flow.

`auth headless-onboard` remains as a compatibility command for older
automation. New scripts should use `polygolem auth login`.

## UI Enable Trading Prompts

If polymarket.com still shows "Enable Trading" after wallet deployment, it is
asking for two typed-data operations:

- EOA-signed ClobAuth to create or derive CLOB API keys.
- DepositWallet Batch signing for UI token approvals.

`polygolem deposit-wallet onboard` now performs both operations by default:
it signs ClobAuth locally, creates or derives CLOB API keys, then submits the
2-call UI token approval WALLET batch. Use `--skip-enable-trading` only when
you intentionally want deploy/approve/fund without these UI readiness signs.

If the wallet is already deployed, run the focused command:

```bash
polygolem deposit-wallet enable-trading
```

This command is also automatic: when relayer credentials are missing, it signs
the SIWE login locally, registers the profile if needed, mints the V2 relayer
key, persists it, creates or derives CLOB API credentials, and submits the UI
approval batch. No mobile-wallet or browser popup is required for Polygolem's
headless path.

Validate the result with:

```bash
polygolem deposit-wallet status --check-enable-trading
```

Important: polymarket.com can still ask the browser to sign ClobAuth because
the website stores browser-local API-key state. That prompt does not mean
Polygolem cannot trade headlessly. Use the status command above as the
headless source of truth.

The SDK surface is `pkg/enabletrading`. For audits, start with dry-run:

```go
result, err := enabletrading.EnableTradingHeadless(ctx, enabletrading.EnableTradingParams{
    OwnerPrivateKey:       privateKey,
    DepositWalletAddress:  depositWallet,
    CreateOrDeriveCLOBKey: true,
    ApproveTokens:         true,
    MaxApproval:           true,
    DryRun:                true,
})
```

See [ENABLE-TRADING-HEADLESS.md](./ENABLE-TRADING-HEADLESS.md) for the typed
data, approval calls, and live-submission safety rules.

## What `auth login` Does

`polygolem auth login` handles the website sign-in step from the CLI:

1. Fetches a SIWE nonce from Gamma.
2. Builds the same message polymarket.com displays.
3. Signs it locally with the EOA from `POLYMARKET_PRIVATE_KEY`.
4. Exchanges the signature for a Polymarket session cookie.
5. Registers the EOA + maker profile for signature type 3 by default.
6. Mints V2 relayer credentials and writes them to a 0600 env file.

It does not export the private key and does not make the deposit wallet sign
the SIWE message. That distinction is intentional.

## Headless Coverage

| Operation | Headless? | Command |
|---|---:|---|
| Derive deposit wallet | Yes | `deposit-wallet derive` |
| Polymarket SIWE login | Yes | `auth login` |
| Profile registration | Yes | `auth login` |
| V2 relayer key mint | Yes | `auth login` or automatic in wallet commands |
| CLOB L2 key create/derive | Yes | `builder auto` or `clob create-api-key` |
| Deposit wallet deploy | Yes | `deposit-wallet deploy` / `deposit-wallet onboard` |
| Trading approvals | Yes | `deposit-wallet approve` / `deposit-wallet onboard` |
| UI Enable Trading signs | Yes | `deposit-wallet onboard` / `deposit-wallet enable-trading` |
| Adapter approvals for redeem | Yes | `deposit-wallet approve-adapters` |
| Funding deposit wallet | Yes | `deposit-wallet fund` |
| Orders and cancels | Yes | `clob create-order`, `market-order`, `cancel*` |
| Redeem winners | Yes, gated by allowlist/readiness | `deposit-wallet settlement-status`, `redeemable`, `redeem` |

## Prerequisites

- `POLYMARKET_PRIVATE_KEY` - Polygon EOA private key, 0x-prefixed.
- pUSD in the EOA if you plan to fund the deposit wallet.
- A little POL for the one ERC-20 funding transfer.
- Relayer credentials are optional for live wallet commands. If
  `RELAYER_API_KEY` + `RELAYER_API_KEY_ADDRESS` are not configured or present
  in a Polygolem env file, `deposit-wallet onboard`, `deploy`, `approve`,
  `approve-adapters`, `nonce`, `status`, `enable-trading`, and live `redeem`
  mint and persist V2 relayer credentials automatically from
  `POLYMARKET_PRIVATE_KEY`.
- CLOB L2 credentials are optional for the normal trading path. Polygolem can
  derive or create the EOA-bound CLOB key from `POLYMARKET_PRIVATE_KEY` when
  needed. Use `builder auto`, `clob create-api-key`, or env when you want
  pre-provisioned credentials for probes and automation.

## Verification

```bash
polygolem auth status --check-deposit-key
```

Expected ready shape:

```json
{
  "eoaAddress": "0x...",
  "depositWallet": "0x...",
  "depositWalletDeployed": true,
  "eoaApiKeyExists": true,
  "depositWalletApiKeyExists": true,
  "canTrade": true
}
```

The `depositWalletApiKeyExists` field name is historical. Current V2 HTTP
auth is EOA-signed; deposit-wallet identity is carried by the POLY_1271 order
payload and the `signature_type=3` CLOB query path.

For the UI Enable Trading gate, prefer:

```bash
polygolem deposit-wallet status --check-enable-trading
```

That command reports whether the ClobAuth key is configured or derivable,
whether the two UI token allowances are live on-chain, and whether the account
is ready for Polygolem's headless trading path.

## Manual Fallback

Use [BROWSER-SETUP.md](./BROWSER-SETUP.md) only when `polygolem auth login`
or `builder auto` is blocked by an upstream API change, account policy, or
local network restriction. The browser signs the same SIWE message with the
EOA; it does not change the deposit wallet address.

## Troubleshooting

### polymarket.com asks my EOA to sign

That is correct. Polymarket login signs with the EOA. The deposit wallet is
the trading wallet after login.

### "Invalid L1 Request headers"

Refresh CLOB credentials:

```bash
polygolem builder auto --force
```

Then retry `polygolem auth status --check-deposit-key`.

### "the order owner has to be the owner of the API KEY"

The CLOB L2 key and the POLY_1271 order identity are out of sync. Re-run:

```bash
polygolem builder auto --force
polygolem clob update-balance --asset-type collateral
```

### "Deposit wallet not deployed"

```bash
polygolem deposit-wallet status --json
polygolem deposit-wallet deploy --wait
```

On-chain bytecode is the source of truth. If `onchainCodeDeployed=true`, the
wallet exists even when the relayer index says otherwise.

### "Insufficient balance"

Fund the deposit wallet, not just the EOA:

```bash
polygolem deposit-wallet fund --amount 0.71
```

## Technical Background

- [LIVE-TRADING-BLOCKER-REPORT.md](./LIVE-TRADING-BLOCKER-REPORT.md)
- [DEPOSIT-WALLET-MIGRATION.md](./DEPOSIT-WALLET-MIGRATION.md)
- [POLY_1271-SIGNING.md](./POLY_1271-SIGNING.md)

This document is the canonical onboarding flow. If another doc contradicts it,
this doc is correct.
