# Deposit Wallet Deployment — Full Pipeline

> **Status:** Production — polygolem ships the full deposit wallet lifecycle
> **Last verified:** 2026-05-07 — live on-chain
> **Companion:** [CONTRACTS.md](./CONTRACTS.md) — contract addresses, permission model, research

---

## Requirements — What You Need (and What You Don't)

### What You Need

| Requirement | Details |
|------------|---------|
| **EOA private key** | A Polygon EOA with a private key. Can be a fresh wallet with zero history and zero MATIC. |
| **pUSD** | ~$0.71 minimum to fund the deposit wallet. This is your trading collateral (not gas). |
| **Builder credentials** | Obtained programmatically via `polygolem builder auto` (ClobAuth EIP-712 signature). No browser needed. |

### What You DON'T Need (Every Absence Is a Feature)

| Don't Need | Why This Matters |
|------------|-----------------|
| ✅ **No KYC** | Builder program Unverified tier starts immediately. No ID, no SSN, no documents, no approval wait. |
| ✅ **No MATIC gas** | Polymarket relayer sponsors ALL on-chain operations. Deploy, approve, batch — fully gasless. |
| ✅ **No external account** | No exchange, no bridge, no third-party custody. Just your EOA in self-custody. |
| ✅ **No Polymarket email/phone login** | Wallet-only authentication. No personal data collected by Polymarket for builder access. |
| ✅ **No ongoing maintenance** | Builder credentials don't expire. Set once, use forever (rotate on compromise). |
| ✅ **No server dependency** | After initial deploy, post-deployment operations are permissionless. Submit batches directly from your EOA. |
| ✅ **No minimum wallet age** | Fresh EOA works. No history, no prior trading, no deposits needed beforehand. |
| ✅ **No premium/paid tier** | Unverified tier includes full relayer access with daily rate limits. Verified tier is optional. |

### Gas Sponsorship Breakdown

| Operation | Who pays gas? | Cost to you |
|-----------|--------------|-------------|
| `deploy` (WALLET-CREATE via relayer) | Polymarket relayer | FREE |
| `approve` (6-call WALLET batch) | Polymarket relayer | FREE |
| `trade` (order placement on CLOB) | Polymarket relayer | FREE |
| CTF split/merge/redeem | Polymarket relayer | FREE |
| `fund` (ERC-20 pUSD transfer EOA→wallet) | **You** | ~0.01 MATIC |

The only operation that costs you anything is the single `fund` transfer — send pUSD from your EOA to the deposit wallet. Everything else is sponsored.

### Replicate From Scratch — 3 Steps (Fully Automated)

```bash
# Step 1 — Generate a fresh EOA (no MATIC, no history, completely empty)
openssl rand -hex 32 > fresh_key.txt

# Step 2 — Builder profile + HMAC creds + wallet onboard (fully automated)
POLYMARKET_PRIVATE_KEY="0x$(cat fresh_key.txt)" \
  polygolem builder auto

POLYMARKET_PRIVATE_KEY="0x$(cat fresh_key.txt)" \
POLYMARKET_BUILDER_API_KEY="..." \
POLYMARKET_BUILDER_SECRET="..." \
POLYMARKET_BUILDER_PASSPHRASE="..." \
  polygolem deposit-wallet onboard --fund-amount 0.71 --json

# Step 3 — Sync and trade
POLYMARKET_PRIVATE_KEY="0x$(cat fresh_key.txt)" \
  polygolem clob update-balance --asset-type collateral --signature-type deposit

POLYMARKET_PRIVATE_KEY="0x$(cat fresh_key.txt)" \
  polygolem clob create-order --token ID --side buy --price 0.5 --size 10 --signature-type deposit
```

> **Note:** Builder credentials are obtained programmatically via `polygolem builder auto`. See [BUILDER-AUTO.md](./BUILDER-AUTO.md) for the full sequence diagram and empirical proof.

---

## 1. Pipeline Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                        DEPOSIT WALLET PIPELINE                        │
│                                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │  DERIVE  │───▶│  DEPLOY  │───▶│ APPROVE  │───▶│   FUND   │      │
│  │ (local)  │    │(relayer) │    │(relayer) │    │(RPC EOA) │      │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘      │
│       │               │               │               │            │
│       ▼               ▼               ▼               ▼            │
│  Predict        WALLET-CREATE    WALLET batch    ERC-20 transfer    │
│  CREATE2        via relayer      (6 approvals)   EOA→deposit wallet │
│  address        (builder creds)  (EOA signs)     (pUSD on Polygon)  │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                         ONBOARD                                │   │
│  │              One command: deploy + approve + fund              │   │
│  └──────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 2. Prerequisites

### 2.1 One-Time Manual Step (Builder Credentials)

1. Open `polymarket.com/settings?tab=builder` in a browser
2. Sign in with your Ethereum wallet (MetaMask, Rabby, etc.)
3. Create a builder profile (free, no KYC)
4. Click **"+ Create New"** to generate Relayer API keys
5. Copy these three values:

| Variable | Source |
|----------|--------|
| `POLYMARKET_BUILDER_API_KEY` | UUID shown in the UI |
| `POLYMARKET_BUILDER_SECRET` | Base64 secret shown in the UI |
| `POLYMARKET_BUILDER_PASSPHRASE` | Hex passphrase shown in the UI |

**Store these securely.** They are shown once in the UI. If lost, revoke and create new ones.

### 2.2 Required for All Operations

| Variable | Required for |
|----------|-------------|
| `POLYMARKET_PRIVATE_KEY` | All authenticated commands (EOA signing) |
| `POLYMARKET_BUILDER_API_KEY` | Wallet deploy, batch, onboard |
| `POLYMARKET_BUILDER_SECRET` | Wallet deploy, batch, onboard |
| `POLYMARKET_BUILDER_PASSPHRASE` | Wallet deploy, batch, onboard |

---

## 3. Step-by-Step Pipeline

### 3.1 Derive — Predict Wallet Address (Local)

**No credentials needed.** Pure CREATE2 computation from the EOA address.

```bash
polygolem deposit-wallet derive
```

**What it does:**
1. Reads `POLYMARKET_PRIVATE_KEY` to derive the EOA address
2. Computes `salt = keccak256(abi.encodePacked(eoaAddress))`
3. Executes CREATE2 formula locally: `keccak256(0xff + factory + salt + initCodeHash)[12:]`
4. Returns the predicted deposit wallet address

**Under the hood:**
```
EOA → keccak256(abi.encodePacked(eoa)) → salt
salt → CREATE2(factory=0x000...b5C9, salt, initCodeHash) → deposit_wallet_address
```

**Output:**
```json
{
  "eoa": "0x21999a074344610057c9b2B362332388a44502D4",
  "deposit_wallet": "0x...",
  "factory": "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07",
  "derived_with": "CREATE2 + keccak256(eoa)"
}
```

### 3.2 Deploy — Create the Wallet (Relayer)

**Requires builder credentials.** Sends a `WALLET-CREATE` request to Polymarket's relayer.

```bash
polygolem deposit-wallet deploy --wait
```

**What it does:**
1. Computes predicted wallet address (same as `derive`)
2. Signs builder HMAC headers with `BUILDER_API_KEY/SECRET/PASSPHRASE`
3. POSTs `{"type": "WALLET-CREATE", "from": eoa, "to": factory}` to `relayer-v2.polymarket.com/submit`
4. Relayer calls `factory.deploy()` (relayer holds operator role)
5. Factory deploys ERC-1967 proxy clone at the predicted CREATE2 address
6. Returns transaction ID and polls until confirmed (`--wait`)

**Why the relayer is required:**
- The factory's `deploy()` function is role-gated to the relayer's EOA
- Calling directly from any other EOA reverts
- The CREATE2 address embeds the factory address — no bypass possible

**Deployment is gas-sponsored** — the relayer pays MATIC gas.

```json
{
  "transaction_id": "tx_abc123...",
  "type": "WALLET-CREATE",
  "deposit_wallet": "0x...",
  "state": "confirmed",
  "transaction_hash": "0x..."
}
```

### 3.3 Approve — Authorize Trading Contracts (Relayer)

**Requires builder credentials + EOA key.** The deposit wallet must approve 6 contracts to spend pUSD and outcome tokens.

```bash
# Review the 6 calls before submitting
polygolem deposit-wallet approve

# Sign and submit
polygolem deposit-wallet approve --submit
```

**What it does:**
1. Builds a batch of 6 `approve()` calls targeting the deposit wallet
2. Fetches the current WALLET nonce from the relayer
3. Constructs the EIP-712 signed envelope (POLY_1271)
4. EOA signs the batch (private key)
5. POSTs the signed batch as `{"type": "WALLET", "from": eoa, "to": factory, ...}` to the relayer
6. Relayer calls `factory.proxy(batch[], signatures[])` which executes the 6 calls through the wallet

**The 6 approvals:**

| # | Target Contract | Token |
|---|----------------|-------|
| 1 | CTF Exchange V2 | pUSD (collateral) |
| 2 | Neg Risk CTF Exchange | pUSD (collateral) |
| 3 | Neg Risk Adapter | pUSD (collateral) |
| 4 | CTF Exchange V2 | Outcome tokens (ERC-1155) |
| 5 | Neg Risk CTF Exchange | Outcome tokens (ERC-1155) |
| 6 | Neg Risk Adapter | Outcome tokens (ERC-1155) |

**Security:** The batch is EOA-signed. `factory.proxy()` validates the signature internally — it does not check caller roles. Anyone can submit a validly-signed batch.

### 3.4 Fund — Transfer pUSD from EOA (Direct RPC)

**Requires EOA key only.** Transfers pUSD from the EOA to the deposit wallet via direct Polygon RPC.

```bash
polygolem deposit-wallet fund --amount 0.71
```

**What it does:**
1. Reads the EOA's pUSD balance
2. Validates the amount is ≤ balance
3. Constructs an ERC-20 `transfer(to: deposit_wallet, amount: X)` call
4. Signs and submits the transaction directly to Polygon (no relayer)
5. Waits for confirmation

**Why direct RPC and not relayer:** The relayer handles factory-interfaced operations (deploy, proxy). Direct ERC-20 transfers from EOA are standard on-chain transactions.

**Minimum amount:** At least 0.50 pUSD recommended to cover trading + fees.

### 3.5 Onboard — All-in-One Automation

```bash
POLYMARKET_PRIVATE_KEY="0x..." \
POLYMARKET_BUILDER_API_KEY="..." \
POLYMARKET_BUILDER_SECRET="..." \
POLYMARKET_BUILDER_PASSPHRASE="..." \
  polygolem deposit-wallet onboard --fund-amount 0.71 --json
```

**What it does:**
1. `derive` — predicts wallet address
2. `deploy` — deploys wallet via relayer
3. `approve --submit` — approves 6 contracts via relayer batch
4. `fund` — transfers pUSD from EOA to wallet

Each step is gated. Failure of any step aborts the composite. The wallet is left in a recoverable, inspectable state (`deposit-wallet status`).

---

## 4. Post-Deployment Operations

### 4.1 Status Check

```bash
polygolem deposit-wallet status
```

Checks: deployment status, pUSD balance, allowance status, nonce.

### 4.2 Balance Sync

```bash
polygolem clob update-balance --asset-type collateral --signature-type deposit
```

Syncs the CLOB's cached balance with the on-chain state. Required after funding.

### 4.3 Trading

```bash
# Create limit order
polygolem clob create-order \
  --token ID --side buy --price 0.5 --size 10 \
  --signature-type deposit

# Market order
polygolem clob market-order \
  --token ID --side buy --amount 5 \
  --signature-type deposit
```

Orders are POLY_1271 signed — the deposit wallet's contract signature format.

---

## 5. Relayer Proxy Pattern (Server-Based)

### 5.1 Why a Server Proxy

Builder credentials must never be exposed to clients (mobile apps, browsers). The **relayer proxy pattern** keeps builder credentials server-side:

```
┌──────────┐      ┌──────────────┐      ┌──────────────┐
│  Flutter  │ ──▶ │  Relayer     │ ──▶ │  Builder      │
│  / Mobile │      │  Proxy       │      │  Relayer v2   │
└──────────┘      │  (server)    │      └──────────────┘
                  └──────────────┘
                       │
                  holds builder
                  credentials
```

### 5.2 Minimal Proxy (~50 LOC)

```go
// POST /relay/deploy
// POST /relay/batch
// Forwards with builder HMAC headers.
// Builder creds are environment-only, never in response.
```

The proxy:
- Never stores or logs private keys
- Never stores or logs builder credentials
- Only forwards WALLET-CREATE and WALLET batch requests
- Rate-limits per user
- Circuit-breaks on repeated failures

### 5.3 What the Proxy Does NOT Handle

- Signing (EOA signs locally)
- Key management (keys stay with user)
- Order placement (direct CLOB, not through proxy)
- Market data (direct Gamma/CLOB, no credentials needed)
- Balance transfers (direct RPC or bridge)

---

## 6. Security Model

### 6.1 Credential Boundaries

| Component | Holds | Never Sees |
|-----------|-------|-----------|
| **User (Flutter/MetaMask)** | EOA private key | Builder credentials |
| **Relayer Proxy (server)** | Builder creds | EOA private key |
| **Polymarket Relayer** | Operator role on factory | Nothing user-specific |
| **Deposit Wallet (on-chain)** | pUSD, tokens, approvals | No credentials |

### 6.2 Attack Surface

| Threat | Mitigation |
|--------|-----------|
| Builder credential leak | Server proxy only; rotate on compromise; rate-limit |
| EOA private key leak | MetaMask/Reown custody; user controls key |
| Malicious batch signing | User sees full batch in signing prompt |
| Proxy compromise | Limited blast radius (deploy/batch only, no funds access) |
| Relayer compromise | Polymarket's infra — out of scope for polygolem |

### 6.3 Redaction Policy

`internal/config` redacts all three builder credentials on every load. No command emits builder credentials in JSON output, logs, or error messages.

---

## 7. Environment Variables

| Variable | Required for |
|----------|-------------|
| `POLYMARKET_PRIVATE_KEY` | All authenticated commands |
| `POLYMARKET_BUILDER_API_KEY` | Relayer: deploy, batch, onboard |
| `POLYMARKET_BUILDER_SECRET` | Relayer: deploy, batch, onboard |
| `POLYMARKET_BUILDER_PASSPHRASE` | Relayer: deploy, batch, onboard |
| `POLYMARKET_GAMMA_URL` | Override Gamma URL (default: `gamma-api.polymarket.com`) |
| `POLYMARKET_CLOB_URL` | Override CLOB URL (default: `clob.polymarket.com`) |
| `POLYMARKET_RELAYER_URL` | Override relayer URL (default: `relayer-v2.polymarket.com`) |

Short-form alternatives (`BUILDER_API_KEY`, `BUILDER_SECRET`, `BUILDER_PASS_PHRASE`) also accepted.

---

## 8. Error Recovery

### 8.1 Deploy Failed

**Symptom:** `WALLET-CREATE` returned error or timed out.

**Recovery:**
1. Check `deposit-wallet status` — the wallet may have been deployed despite timeout
2. If not deployed, re-run `deposit-wallet deploy --wait`
3. Validated builder credentials are still valid

### 8.2 Approve Failed

**Symptom:** Batch submission failed.

**Recovery:**
1. Check `deposit-wallet status` for current nonce
2. Check individual allowance statuses
3. Re-run `deposit-wallet approve --submit` (uses current nonce)

### 8.3 Fund Failed

**Symptom:** ERC-20 transfer reverted or insufficient balance.

**Recovery:**
1. Verify EOA has sufficient pUSD
2. Check `deposit-wallet status` for current wallet balance
3. Re-run with corrected amount

### 8.4 Builder Credentials Lost

**Symptom:** Relayer returns 401 Unauthorized.

**Recovery:**
1. Go to `polymarket.com/settings?tab=builder`
2. Revoke old keys and create new ones
3. Update environment variables
4. Existing wallet is unaffected (credentials are auth, not custody)

---

## 9. Related Documents

- [CONTRACTS.md](./CONTRACTS.md) — Smart contract addresses, permission model, on-chain research
- [DEPOSIT-WALLET-MIGRATION.md](./DEPOSIT-WALLET-MIGRATION.md) — Bot killer survival guide, V1→V2 migration
- [SAFETY.md](./SAFETY.md) — Security model, credential handling, deposit wallet safety rules
- [COMMANDS.md](./COMMANDS.md) — Full CLI command reference
- [ARCHITECTURE.md](./ARCHITECTURE.md) — Package boundaries, dependency direction

---

*This document defines the complete deposit wallet deployment pipeline as implemented in polygolem. Every step is live-tested against Polymarket production. The only manual step is copying builder credentials from the Polymarket web UI — everything else is automated.*
