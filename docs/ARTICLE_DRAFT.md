# How Polymarket V2 Actually Works (And How to Automate It)

**Author:** Trebuchet Dynamics
**Date:** 2026-05-08
**Status:** Production-verified with live funds on Polygon mainnet

---

## 1. Polymarket V2 in 30 Seconds

Polymarket is the largest prediction market. In April 2026 they migrated to V2: new CLOB, new stablecoin (pUSD), and **deposit wallets** — smart contract wallets that prevent ghost fills.

**Ghost fills:** Orders that appear executed in the book but never settle because the signer didn't actually hold the funds. V2 fixes this by requiring the wallet that *places* orders to be the wallet that *holds* funds.

**Deposit wallet architecture:**
- ERC-1967 proxy deployed via relayer `WALLET-CREATE`
- CREATE2 deterministic address (predictable before deployment)
- ERC-1271 signature validation (`isValidSignature`)
- Orders use `signatureType=3` (`POLY_1271`) with ERC-7739 wrapped signatures

> **Official docs:** [Polymarket Deposit Wallet Documentation](https://docs.polymarket.com/trading/deposit-wallets)

---

## 2. How Polymarket Auth Actually Works

Every guide, doc, and community post explains what the endpoints are. Almost none explain the **actual flow** a fresh account goes through. We captured it with Playwright — injected stub wallet, full request logging, automated signup. Here's what happens:

### Step 1: Proxy Profile Registration

Polymarket registers a proxy profile for every new account. Not a deposit wallet — a proxy profile.

```
>>> GET https://gamma-api.polymarket.com/profiles
response: [{"proxyWallet":"0x4c72e9fd06a3478e29f69879db9a058ebe35ef84",...}]
```

The `proxyWallet` equals the EOA address. This is the account identity in Polymarket's system.

### Step 2: V2 Relayer Key

The relayer issues an API key for gasless transactions (deploy, approve, batch):

```
>>> POST https://relayer-v2.polymarket.com/relayer/api/auth
<<< 200
response: {"apiKey":"019e0928-79bd-782e-9c7b-d02a4282d92c",...}
```

### Step 3: Deposit Wallet Deployment

The relayer deploys the deposit wallet on-chain:

```
>>> POST https://relayer-v2.polymarket.com/submit
body: {"type":"WALLET-CREATE","from":"0x8968...","to":"0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"}
<<< 200
response: {"transactionHash":"0x2d8c0469...","state":"STATE_MINED","type":"WALLET-CREATE"}
```

### Step 4: CLOB API Key Creation

This is the part every doc gets vague about. Here's exactly what the browser sends:

```
>>> POST https://clob.polymarket.com/auth/api-key
poly_address: 0x8968CE148788015103F291F679AA4D17e0b1f088
poly_signature: 0x6b140a67106e1d85fc820954d480221404d2548341dc6909c562fb01bf136e2346b3aedb305f95ad137f6ec4637a867e31fdc6b59d2d2e5621dd5e0834dafe041c
poly_timestamp: 1778270141
poly_nonce: 0
<<< 200
response: {"apiKey":"b28f3795-db29-6a67-e0f8-892db2d96030",...}
```

**Key facts:**
- `POLY_ADDRESS` is the **deposit wallet**, not the EOA
- `POLY_SIGNATURE` is **65 bytes of standard EOA ECDSA**, not ERC-7739 wrapped
- The EOA signs the ClobAuth message, but the server binds the API key to the deposit wallet
- No ERC-1271 validation happens at the HTTP layer

**Why this matters:** The deposit wallet's identity is established at the API key binding step. After that, every authenticated request uses `POLY_ADDRESS=depositWallet` and L2 HMAC headers. The order's `signatureType=3` tells the CTF Exchange to call `isValidSignature` on-chain.

> **Evidence:** Full Playwright capture with fresh EOA `0x4c72...f84` → proxy profile → relayer key → WALLET-CREATE → API key bound to deposit wallet `0x8968...`. All request/response logs in [BLOCKERS.md](https://github.com/TrebuchetDynamics/polygolem/blob/main/BLOCKERS.md) § "CORRECTION 2026-05-08".

---

## 3. The Full Headless Flow

With the auth flow understood, here's how to do everything headlessly:

### Prerequisites

```bash
# Install polygolem
go install github.com/TrebuchetDynamics/polygolem/cmd/polygolem@latest

# Set your EOA private key
export POLYMARKET_PRIVATE_KEY="0x..."
```

### Step 1: Register Profile + Mint Relayer Key

```bash
polygolem auth login
```

This does: SIWE login → `POST /profiles` → mint V2 relayer key → persist to `.env.relayer-v2`.

### Step 2: Deploy Deposit Wallet

```bash
polygolem deposit-wallet deploy --wait
```

Submits `WALLET-CREATE` to relayer, polls until `STATE_MINED`.

### Step 3: Create CLOB API Key

```bash
polygolem clob create-api-key-for-address \
  --owner $(polygolem deposit-wallet derive | jq -r '.depositWallet')
```

EOA signs ClobAuth, `POLY_ADDRESS` = deposit wallet, server binds API key to deposit wallet.

### Step 4: Fund and Approve

```bash
polygolem deposit-wallet onboard --fund-amount 0.71
```

Derives address → deploy (skip if done) → submit 6-call approval batch → transfer pUSD from EOA to deposit wallet.

### Step 5: Trade

```bash
polygolem clob create-order \
  --token <TOKEN_ID> \
  --side buy \
  --price 0.5 \
  --size 10
```

Order uses `signatureType=3`, `maker=depositWallet`, `signer=depositWallet`, ERC-7739 wrapped signature. CTF Exchange validates via `isValidSignature`.

---

## 4. What polygolem Is

Polygolem is a **Swiss Army knife for Polymarket**. One tool for everything:

### Query Everything

```bash
polygolem discover search --query "bitcoin 150k" --limit 5
polygolem discover market --id "0xbd31dc8..."
polygolem clob book <token-id>
polygolem orderbook spread <token-id>
polygolem clob markets --cursor ""
polygolem events list
```

### Account + Wallet

```bash
polygolem auth status --check-deposit-key
polygolem deposit-wallet derive
polygolem deposit-wallet deploy --wait
polygolem deposit-wallet onboard --fund-amount 0.71
polygolem deposit-wallet status
```

### Trading

```bash
polygolem clob create-order --token <ID> --side buy --price 0.5 --size 10
polygolem clob batch-orders --orders-file orders.json
polygolem clob cancel <order-id>
polygolem clob cancel-all
polygolem clob orders
polygolem clob trades
```

### Builder + Attribution

```bash
polygolem builder auto
polygolem clob create-builder-fee-key
polygolem clob list-builder-fee-keys
```

### Read-Only (No Credentials)

```bash
polygolem discover search --query "btc 5m"
polygolem orderbook get --token-id "123..."
polygolem health
polygolem version
```

### Key Features

- **No external SDKs** — All types, signing, protocol logic from spec
- **Go-native** — Single binary, `go install`, MIT license
- **Deposit wallet exclusive** — The only mode Polymarket accepts for new users
- **Full CLI + SDK** — Use commands or import `pkg/` packages
- **Live-tested** — Verified with real funds on Polygon mainnet

---

## 5. Why This Matters

**For traders:** Deposit wallets prevent ghost fills. Your orders are backed by real pUSD in a smart contract wallet. No more "book says filled, balance says zero."

**For developers:** The auth flow is simpler than it looks. Standard EOA ECDSA for L1, ERC-7739 only for order signing. Don't over-engineer what the browser doesn't.

**For automation:** Polymarket login signs with the EOA; the deposit wallet
remains the trading wallet. `polygolem auth login`, `builder auto`,
deposit-wallet deploy, approvals, funding, orders, cancels, and settlement can
run headlessly.

---

## 6. Try It Yourself

```bash
# 1. Generate fresh EOA
export POLYMARKET_PRIVATE_KEY=$(polygolem auth export-key --generate | jq -r '.privateKey')

# 2. Full headless signup
polygolem auth login
polygolem deposit-wallet deploy --wait
polygolem builder auto
polygolem deposit-wallet onboard --fund-amount 0.71

# 3. Check status
polygolem auth status --check-deposit-key

# 4. Query a market
polygolem discover search --query "bitcoin" --limit 1

# 5. Place an order (requires pUSD in deposit wallet)
polygolem clob create-order --token <ID> --side buy --price 0.5 --size 10
```

> **Safety:** Use testnet or <$1 pUSD on mainnet. Don't fund a test wallet with more than you're willing to lose.

---

## Citations

1. Polymarket Deposit Wallet Documentation — https://docs.polymarket.com/trading/deposit-wallets
2. Polymarket Authentication Documentation — https://docs.polymarket.com/api-reference/authentication
3. Polymarket Quickstart Guide — https://docs.polymarket.com/trading/quickstart
4. Polygolem Repository — https://github.com/TrebuchetDynamics/polygolem
5. Polygolem ONBOARDING.md — https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ONBOARDING.md
6. Polygolem BLOCKERS.md (Playwright capture evidence) — https://github.com/TrebuchetDynamics/polygolem/blob/main/BLOCKERS.md

---

*Last updated: 2026-05-08 — Verified with live funds on Polygon mainnet*
