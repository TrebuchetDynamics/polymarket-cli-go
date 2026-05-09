# Deposit Wallet Onboarding — Single Source of Truth

**Last updated:** 2026-05-09
**Status:** Production — verified with live funds on Polygon mainnet
**Companion:** [BROWSER-SETUP.md](./BROWSER-SETUP.md) — one-time browser login guide for new users

---

## TL;DR

| User Type | What You Need | Browser Required? |
|-----------|--------------|-------------------|
| **Existing Polymarket user** | `POLYMARKET_PRIVATE_KEY` | ❌ No — fully headless |
| **New user (fresh EOA)** | `POLYMARKET_PRIVATE_KEY` + pUSD | ⚠️ One-time browser login |

**For new users:** the supported path is browser login once to mint the
deposit-wallet-owned CLOB API key. Polygolem keeps the rest of the lifecycle
headless: relayer credentials, wallet deploy, trading approvals, adapter
approvals, funding, balance sync, orders, cancels, settlement readiness, and
winner redemption.

**Cost:** ~$0.01 POL for one pUSD funding transfer. Everything else (deploy, approve, orders) is gasless via relayer.

---

## Prerequisites

- `POLYMARKET_PRIVATE_KEY` — Polygon EOA private key (0x-prefixed hex)
- ~$0.71 pUSD — trading collateral (not gas)
- ~0.01 POL — gas for the single pUSD transfer

**Optional for new users:**
- A browser (Chrome, Firefox, etc.)
- MetaMask, Rabby, or WalletConnect-compatible wallet

---

## The Two Paths

### Path A: Existing Polymarket Users (Fully Headless)

If you already have a Polymarket account and a deposit-wallet-owned API key:

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

# 1. V2 relayer credentials (headless)
polygolem auth headless-onboard

# 2. Deploy + approve + fund (headless)
polygolem deposit-wallet onboard --fund-amount 0.71

# 3. Sync balance
polygolem clob update-balance --asset-type collateral

# 4. Trade
polygolem clob create-order --token <TOKEN_ID> --side buy --price 0.5 --size 10
```

**Verify:**
```bash
polygolem auth status --check-deposit-key
# Expected: canTrade: true
```

---

### Path B: New Users (One-Time Browser Login Required)

If you've never used Polymarket with this key:

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

# Step 1: Derive deposit wallet address (local)
polygolem deposit-wallet derive
# Save the depositWallet address

# Step 2: V2 relayer credentials (headless)
polygolem auth headless-onboard

# Step 3: Deploy deposit wallet (headless)
polygolem deposit-wallet deploy --wait

# Step 4: Browser login (REQUIRED — see BROWSER-SETUP.md)
# Go to https://polymarket.com, connect wallet, let it detect your deposit wallet
# This creates the deposit-wallet-owned CLOB API key in the background
# Full guide: docs/BROWSER-SETUP.md

# Step 5: Approve + fund (headless)
polygolem deposit-wallet approve --submit
polygolem deposit-wallet fund --amount 0.71

# Step 6: Sync and trade (headless)
polygolem clob update-balance --asset-type collateral
polygolem clob create-order --token <TOKEN_ID> --side buy --price 0.5 --size 10
```

**Why browser login is required:**
- Polymarket's `/auth/api-key` endpoint requires an EIP-712 ClobAuth signature
- Deposit wallets are ERC-1271 smart contracts — they can't produce raw ECDSA signatures
- Polymarket's L1 auth endpoint does not support ERC-1271 `isValidSignature` validation
- Result: `HTTP 401 {"error":"Invalid L1 Request headers"}` when trying headless

**After browser login:**
- The API key is permanently associated with your deposit wallet
- Polygolem derives it automatically on demand
- All future operations are fully headless

---

## What Works Headlessly vs. What Doesn't

| Operation | Headless? | Notes |
|-----------|-----------|-------|
| SIWE login | ✅ Yes | `auth headless-onboard` |
| Relayer key mint | ✅ Yes | `auth headless-onboard` |
| Deposit wallet deploy | ✅ Yes | `deposit-wallet deploy`; skips `WALLET-CREATE` when Polygon bytecode already exists |
| Deposit wallet status | ✅ Yes | `deposit-wallet status`; reports both relayer and Polygon bytecode status |
| **Deposit-wallet CLOB API key** | ❌ **No** | **Requires browser login for new users** |
| Balance check | ✅ Yes | After API key exists |
| Order creation | ✅ Yes | After API key exists |
| Batch orders | ✅ Yes | After API key exists |
| Market order | ✅ Yes | After API key exists |
| Order cancellation | ✅ Yes | After API key exists |
| Heartbeat | ✅ Yes | After API key exists |
| Builder fee key | ✅ Yes | `clob create-builder-fee-key` |

---

## Bot-Generated Keys

If your bot/agent generated `POLYMARKET_PRIVATE_KEY`:

```bash
# Display the key for wallet import (use with care)
polygolem auth export-key --confirm

# Then follow BROWSER-SETUP.md Step 4
```

**Security:** Clear terminal history after import:
```bash
history -c && clear
```

---

## Troubleshooting

### "Invalid L1 Request headers" (HTTP 401)

You're trying to create or derive a deposit-wallet-owned API key headlessly. This is **expected to fail** for new users. Complete the browser login in [BROWSER-SETUP.md](./BROWSER-SETUP.md).

### "the order owner has to be the owner of the API KEY" (HTTP 400)

You're trying to place a deposit-wallet order with an EOA-owned API key. Deposit-wallet orders **must** use a deposit-wallet-owned API key.

### "Deposit wallet not deployed"

Run `polygolem deposit-wallet status` first. If `onchainCodeDeployed` is
`true`, the wallet exists even when `relayerDeployed` is `false`.

If both are false, run:

```bash
polygolem deposit-wallet deploy --wait
```

The wallet must exist on-chain before browser signup and POLY_1271 order
signing.

### "Insufficient balance"

Fund your deposit wallet:
```bash
polygolem deposit-wallet fund --amount 0.71
```

---

## Verification

```bash
polygolem auth status --check-deposit-key
```

Expected for ready-to-trade:
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

---

## Technical Background

For the full technical analysis:
- [LIVE-TRADING-BLOCKER-REPORT.md](./LIVE-TRADING-BLOCKER-REPORT.md)
- [INTEGRATION_PLAN.md](../opensource-projects/INTEGRATION_PLAN.md) — Appendix A

---

*This document is the single source of truth for polygolem onboarding. If another doc contradicts it, this doc is correct.*
