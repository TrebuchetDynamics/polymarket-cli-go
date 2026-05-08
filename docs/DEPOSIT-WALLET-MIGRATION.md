# Deposit Wallet Migration — Bot Killer Survival Guide

> **Status**: ACTIVE MIGRATION — enforced for new API users as of May 2026
> **Impact**: Bots using EOA orders will be rejected with `"maker address not allowed, please use the deposit wallet flow"`
> **Last verified**: 2026-05-07 (live run)

---

## 1. What Happened

On **April 6, 2026**, Polymarket launched its largest infrastructure upgrade: CTF Exchange V2, a new native stablecoin (pUSD), and a rebuilt CLOB. On **April 28, 2026**, V1 was shut down and V2 became the only production path.

The migration introduced **deposit wallets** — a new wallet architecture using ERC-1271 (POLY_1271) smart contract signatures. For new API users, **EOA orders are no longer accepted** by the CLOB.

### Timeline

| Date | Event |
|------|-------|
| 2026-04-06 | CTF Exchange V2 and pUSD announced; preprod available |
| 2026-04-28 | V1 CLOB shut down; V2 production-only |
| 2026-05-01+ | Deposit wallet enforcement begins for new API users |
| 2026-05-07 | Our bot hit the wall: `"maker address not allowed, please use the deposit wallet flow"` |

### Sources

- [Polymarket V2 Migration Guide](https://docs.polymarket.com/v2-migration)
- [Deposit Wallet Migration Guide](https://docs.polymarket.com/trading/deposit-wallet-migration)
- [Builder Program Overview](https://docs.polymarket.com/builders/overview)
- [Polymarket April 2026 Upgrade Announcement (Bitcoin.com News)](https://news.bitcoin.com/polymarkets-april-2026-upgrade-new-stablecoin-faster-order-matching-smart-contract-wallet-support/)
- [CTF Exchange V2 Technical Analysis (MasterTP Blog)](https://blog.mastertp.com/posts/polymarket-ctf-exchange-v2-native-collateral-rebuild/)

---

## 2. Why This Is a Bot Killer

**Why this rejection occurs:** Per docs.polymarket.com/v2-migration,
the V2 backend rejects orders whose maker address has no
smart-account contract deployed and approved at the V2 Exchange. The
"use the deposit wallet flow" message is one suggested remediation;
for grandfathered accounts an existing proxy or Safe deployment also
satisfies the requirement.

### 2.1 The Rejection

Any `POST /order` with `signatureType = 0` (EOA) from an account classified as a "new API user" returns:

```json
{"error": "maker address not allowed, please use the deposit wallet flow"}
```

There is **no opt-out**. You cannot stay on EOA orders if you trigger this classification.

### 2.2 What Triggers "New API User" Classification

Polymarket has not publicly documented the exact criteria, but evidence suggests:
- Accounts that have never used proxy/Safe wallets
- Accounts that first interacted via API (not browser wallet)
- Accounts created after the V2 rollout date

**Existing proxy/Safe users are grandfathered** in this phase and should continue using their current setup.

### 2.3 The pUSD Trap

V2 changed the collateral token from **USDC.e** to **pUSD** (`0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB`). This is the #1 pitfall reported across GitHub issues:

> "I approved USDC.e for the V2 spenders thinking that was the right token. The CLOB exclusively tracks pUSD for V2 — USDC.e approvals are no-ops." — GitHub Issue #297

**Symptoms of the pUSD trap:**
- CLOB balance shows $0 despite on-chain USDC.e balance
- Allowances show 0 despite on-chain approvals
- Fund-pusd bridge deposits POL → pUSD to EOA (not deposit wallet)

### 2.4 The Split Cache Problem

The CLOB maintains **separate balance/allowance records per signature type** for the same wallet. A common scenario:

| Query | Result |
|-------|--------|
| `signature_type=1` (proxy) | Balance: $108.82, Allowances: 0 |
| `signature_type=2` (gnosis-safe) | Balance: $0.00, Allowances: MAX |
| `signature_type=3` (deposit) | Balance: $0.00, Allowances: 0 |

If your wallet was ever queried under multiple types, the CLOB cache can split. The fix is a full `update-balance` after funding and approving the deposit wallet.

---

## 3. The Deposit Wallet Architecture

### 3.1 Mental Model

```
┌─────────────────────────────────────────────────────┐
│                    YOUR EOA                          │
│  (holds private key, signs everything)               │
│  pUSD balance: 0.709708 ← NOT USED for orders        │
└──────────────┬──────────────────────────────────────┘
               │ owns
               ▼
┌─────────────────────────────────────────────────────┐
│              DEPOSIT WALLET                           │
│  (ERC-1967 proxy, deterministic CREATE2 address)     │
│  pUSD balance: 0.000000 ← THIS is what CLOB reads    │
│  Implements ERC-1271 isValidSignature()              │
└─────────────────────────────────────────────────────┘
```

**Critical rule from Polymarket docs:**
> *"pUSD held by the EOA does not count as CLOB buying power for deposit wallet orders."*

### 3.2 Signature Types

| Type | Name | Signer | Maker | Use Case |
|------|------|--------|-------|----------|
| 0 | EOA | EOA | EOA | Legacy (blocked for new API users) |
| 1 | POLY_PROXY | EOA | Proxy wallet | Magic Link / email accounts |
| 2 | POLY_GNOSIS_SAFE | EOA | Safe wallet | Browser wallet accounts |
| 3 | **POLY_1271** | EOA | **Deposit wallet** | New API users (required) |

### 3.3 ERC-7739 Wrapped Signatures

POLY_1271 orders use a nested ERC-7739 `TypedDataSign` wrapper:

```
signature = 0x || innerECDSA(65 bytes) || appDomainSeparator(32 bytes) 
            || contentsHash(32 bytes) || typeString(encoded) || typeStringLen(2 bytes)
```

The inner ECDSA is over: `keccak256(domainSeparator || contentsHash || typeString)`, where:
- `domainSeparator` = EIP-712 hash of `{ name: "DepositWallet", version: "1", chainId: 137, verifyingContract: depositWalletAddr }`
- `contentsHash` = EIP-712 hash of the Order struct under the CTF Exchange V2 domain
- `typeString` = `"Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"`

This is already implemented in `polygolem/internal/clob/orders.go` (636-byte output).

---

## 4. Onboarding Flow (Step by Step)

### Prerequisites

1. **EOA with private key** — you have this
2. **Builder Program credentials** — get from [polymarket.com/settings?tab=builder](https://polymarket.com/settings?tab=builder)
3. **POL for gas** on Polygon (minimum ~50 POL reserve recommended)
4. **pUSD** to fund the deposit wallet

### Step 1: Get Builder Credentials

Go to [polymarket.com/settings?tab=builder](https://polymarket.com/settings?tab=builder) and create a builder profile. You'll receive credentials that authenticate with the relayer.

**Environment variables:**
```bash
export POLYMARKET_BUILDER_API_KEY="your-builder-key"
export POLYMARKET_BUILDER_SECRET="your-builder-secret"
export POLYMARKET_BUILDER_PASSPHRASE="your-builder-passphrase"
# Optional: set custom relayer URL
export POLYMARKET_RELAYER_URL="https://relayer-v2.polymarket.com"
```

**Note on credential format**: The relayer may accept either:
- `POLY_BUILDER_API_KEY` style (HMAC headers) — documented in deposit wallet migration
- `RELAYER_API_KEY` style — documented in newer gasless transactions page

Try the `POLY_BUILDER_*` format first; if you get `"invalid authorization"`, switch to the `RELAYER_API_KEY` format.

### Step 2: Derive Deposit Wallet Address

```bash
cd go-bot/polygolem
POLYMARKET_PRIVATE_KEY="0x..." polygolem deposit-wallet derive --json
```

Output:
```json
{
  "owner": "0xYourEOA...",
  "depositWallet": "0xDeterministicDepositWallet..."
}
```

The deposit wallet address is **deterministic** — same EOA always produces the same wallet. Store both addresses.

### Step 3: Deploy the Deposit Wallet

```bash
POLYMARKET_PRIVATE_KEY="0x..." \
POLYMARKET_BUILDER_API_KEY="..." \
POLYMARKET_BUILDER_SECRET="..." \
POLYMARKET_BUILDER_PASSPHRASE="..." \
  polygolem deposit-wallet deploy --wait --json
```

This submits a `WALLET-CREATE` transaction to the relayer and polls until the wallet is deployed on Polygon. Expected output:

```json
{
  "transactionID": "tx-abc123",
  "state": "STATE_MINED",
  "owner": "0xYourEOA...",
  "depositWallet": "0xDeterministicDepositWallet..."
}
```

### Step 4: Fund the Deposit Wallet with pUSD

**This is the most critical step.** Your EOA has pUSD, but the deposit wallet needs its own pUSD.

**Option A: Direct ERC-20 Transfer (simplest)**
From your EOA, send pUSD to the deposit wallet address:
- Token: `0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB` (pUSD on Polygon)
- To: Your deposit wallet address (from Step 2)
- Amount: Your desired trading balance

This is a normal on-chain transaction requiring POL for gas.

**Option B: Polymarket Bridge**
Use `fund-pusd` which bridges POL → pUSD. Point it at the deposit wallet address instead of the EOA:
```bash
# This currently funds the EOA, not the deposit wallet
# Work needed: support --to-address flag
```

**Option C: WALLET Batch (not yet implemented in polygolem)**
Submit a relayer `WALLET` batch that calls `pUSD.transfer(depositWallet, amount)` from the EOA. This requires the EOA to have first approved the deposit wallet to spend its pUSD.

### Step 5: Approve Trading Contracts

The deposit wallet must approve pUSD and conditional token spenders. These approvals must be made **from the deposit wallet** (not the EOA) via a relayer `WALLET` batch.

**V2 Spender Addresses (Polygon mainnet):**
```
CTF Exchange:    0xE111180000d2663C0091e4f400237545B87B996B
Neg Risk:        0xe2222d279d744050d28e00520010520000310F59
Neg Risk Adapter: 0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296
```

**Manual approach**: Use the Polymarket UI at [polymarket.com](https://polymarket.com) to approve — the UI handles approvals for the deposit wallet automatically.

**Automated approach**: Implement WALLET batch with ERC-20 `approve()` calldata for each spender. Not yet implemented in polygolem (see Section 7).

### Step 6: Sync CLOB Balance

After funding and approvals, refresh the CLOB cache:

```bash
POLYMARKET_PRIVATE_KEY="0x..." \
  polygolem clob update-balance --asset-type collateral --json
```

Then verify the balance is visible:
```bash
POLYMARKET_PRIVATE_KEY="0x..." \
  polygolem clob balance --asset-type collateral --json
```

Expected output shows non-zero balance and non-zero allowances.

### Step 7: Place Orders

Orders must use:
- `signatureType = 3` (POLY_1271)
- `maker = depositWalletAddress`
- `signer = depositWalletAddress`
- ERC-7739 wrapped signature

Polygolem handles this automatically — sigtype 3 is the only type the SDK signs with.

---

## 5. Common Pitfalls

### 5.1 "pUSD shows on EOA but CLOB sees zero"

**Root cause**: pUSD must be in the deposit wallet, not the EOA. Transfer pUSD from EOA → deposit wallet address.

### 5.2 "invalid authorization" on WALLET-CREATE

**Root cause**: Builder/relayer credentials are missing or wrong format. Verify:
1. You have a builder profile at polymarket.com/settings?tab=builder
2. Environment variables are set correctly
3. Try alternate header format if POLY_BUILDER_* fails

### 5.3 "not enough balance / allowance"

**Common causes**:
- pUSD is still on EOA, not deposit wallet
- Allowances were set on EOA but need to be set on deposit wallet via WALLET batch
- CLOB cache is stale — run `update-balance`
- Approved USDC.e instead of pUSD (V2 uses pUSD, not USDC.e)

### 5.4 Split Cache (legacy)

Historical issue: pre-V2 wallets that were queried under multiple signature types could end up with a split CLOB cache (balance under one type, allowances under another). The V2 cutover removed `--signature-type` and forced sigtype 3 / deposit wallet, so this no longer applies. If you migrated a pre-V2 account, run `update-balance` once to repopulate the cache under sigtype 3.

### 5.5 V1 vs V2 Domain Mismatch

V2 orders use EIP-712 domain version `"2"` and new verifying contracts. V1 signatures are rejected. Polygolem uses V2 by default — do not override the domain version.

---

## 6. Builder Program: What You Need to Know

### 6.1 What Builders Get

| Benefit | Description |
|---------|-------------|
| **Relayer Access** | Gas-free wallet deployment, approvals, CTF operations |
| **Volume Tracking** | Orders attributed to your builder profile |
| **Leaderboard** | Public visibility on builders.polymarket.com |
| **Support** | Telegram + engineering support (Verified+ tier) |

### 6.2 Builder Code (V2 Order Attribution)

V2 moved builder attribution from HMAC headers into the order struct itself via the `builder` field (bytes32). This is separate from relayer authentication:
- **Order attribution**: `builderCode` field on each order → no more `POLY_BUILDER_*` HMAC headers on orders
- **Relayer authentication**: Still uses API keys (either `POLY_BUILDER_*` HMAC or `RELAYER_API_KEY` header)

### 6.3 No Approval Wait (Typically Instant)

Builder profile creation at polymarket.com/settings is generally instant. No application review period.

---

## 7. What Polygolem Implements (and What's Missing)

### ✅ Implemented

| Feature | Command/API |
|---------|-------------|
| Deposit wallet address derivation | `polygolem deposit-wallet derive` |
| WALLET-CREATE deployment | `polygolem deposit-wallet deploy [--wait]` |
| WALLET nonce fetch | `polygolem deposit-wallet nonce` |
| Transaction status polling | `polygolem deposit-wallet status [--tx-id]` |
| Deployment check | `polygolem deposit-wallet status` |
| V2 order signing (POLY_1271) | `polygolem clob create-order` |
| CLOB balance with sig type 3 | `polygolem clob balance` |
| CLOB balance sync | `polygolem clob update-balance` |
| ERC-7739 wrapped signatures | Automatic for deposit wallet orders |

### ❌ Not Yet Implemented

| Feature | Notes |
|---------|-------|
| WALLET batch signing (EIP-712) | DepositWallet.Batch domain + types |
| pUSD transfer to deposit wallet via relayer | ERC-20 transfer calldata in WALLET batch |
| Token approvals via WALLET batch | approve(pUSD) + setApprovalForAll(CTF) |
| Automated funding orchestration | Full deploy → fund → approve → sync flow |
| Live readiness for deposit wallet | go-bot live_wire.go deposit wallet checks |

---

## 8. Quick Reference: Env Vars

```bash
# Required for all operations
POLYMARKET_PRIVATE_KEY="0x..."

# Required for deposit wallet deploy/nonce/status
POLYMARKET_BUILDER_API_KEY="..."
POLYMARKET_BUILDER_SECRET="..."
POLYMARKET_BUILDER_PASSPHRASE="..."
# Or short-form:
BUILDER_API_KEY="..."
BUILDER_SECRET="..."
BUILDER_PASS_PHRASE="..."

# Optional
POLYMARKET_RELAYER_URL="https://relayer-v2.polymarket.com"  # default
POLYMARKET_SIGNATURE_TYPE="deposit"  # for live trading
```

---

## 9. External Resources

- [Polymarket Deposit Wallet Migration](https://docs.polymarket.com/trading/deposit-wallet-migration)
- [Polymarket V2 Migration Guide](https://docs.polymarket.com/v2-migration)
- [Builder Program](https://docs.polymarket.com/builders/overview)
- [Builder Relayer Client (TypeScript)](https://github.com/Polymarket/builder-relayer-client)
- [CLOB Client V2 (TypeScript)](https://github.com/Polymarket/clob-client-v2)
- [CLOB Client V2 (Python)](https://github.com/Polymarket/py-clob-client-v2)
- [CLOB Client V2 (Rust)](https://github.com/tdergouzi/rs-clob-client-v2)
- [Polymarket Discord (dev channel)](https://discord.gg/polymarket)
- [Builder Settings](https://polymarket.com/settings?tab=builder)

### Community Bug Reports (Relevant)

- [GitHub #339: POLY_PROXY accounts fail V2 auth (signer mismatch)](https://github.com/Polymarket/py-clob-client/issues/339)
- [GitHub #297: Proxy wallet balance correct but allowances stuck at 0](https://github.com/Polymarket/py-clob-client/issues/297)
- [GitHub #287: CLOB orders failing with "not enough balance / allowance"](https://github.com/Polymarket/py-clob-client/issues/287)
- [GitHub #248: POLY_ADDRESS header uses signer address instead of funderAddress](https://github.com/Polymarket/clob-client/issues/248)

---

## 10. Recovery Checklist

If your bot is down with deposit wallet issues:

- [ ] Confirm you're in the "new API user" category (EOA orders rejected)
- [ ] Create builder profile at polymarket.com/settings?tab=builder
- [ ] Set `POLYMARKET_BUILDER_API_KEY/SECRET/PASSPHRASE` env vars
- [ ] Run `polygolem deposit-wallet derive` — note the deposit wallet address
- [ ] Run `polygolem deposit-wallet deploy --wait` — deploy the wallet
- [ ] Transfer pUSD from EOA to deposit wallet address (ERC-20 transfer)
- [ ] Approve trading contracts via Polymarket UI or WALLET batch
- [ ] Run `polygolem clob update-balance --asset-type collateral`
- [ ] Run `polygolem clob balance --asset-type collateral` to verify
- [ ] Restart bot and verify no `"maker address not allowed"` errors
