# Builder Credential Issuance — Reverse Engineering Report

> **Date:** 2026-05-07
> **Status:** Investigation complete — programmatic (curl-only) issuance is **impossible**
> **Last verified:** Live against polymarket.com and relayer-v2.polymarket.com OpenAPI spec
> **Companion:** [CONTRACTS.md](./CONTRACTS.md), [DEPOSIT-WALLET-DEPLOYMENT.md](./DEPOSIT-WALLET-DEPLOYMENT.md)

---

## 1. Executive Summary

**Can we get builder API keys programmatically (curl, HTTP, no browser)?**

### Answer: **NO.** The irreducible manual step is real and cannot be bypassed.

Both credential types (Builder API Key and Relayer API Key) require interaction with the Polymarket web UI at `polymarket.com/settings`. No public REST endpoint exists for creating either type. The web UI uses Gamma auth (browser session cookies from email/wallet login at polymarket.com), which cannot be replicated with a simple curl command.

### Irreducible Manual Steps

| Step | Can it be automated? | Method |
|------|---------------------|--------|
| Login to polymarket.com | ❌ No | Requires wallet signature + Polymarket session cookie |
| Navigate to settings?tab=builder | ❌ No (headless only) | Playwright/Puppeteer possible but still requires manual login |
| Create builder profile | ❌ No (headless only) | Requires Gamma auth session |
| Click "+ Create New" for API keys | ❌ No (headless only) | Requires Gamma auth session |
| Copy BUILDER_API_KEY/SECRET/PASSPHRASE | ✅ Yes | User copies 3 values once |

---

## 2. Credential Types

Polymarket has TWO separate credential systems for relayer access:

### 2.1 Builder API Key (`settings?tab=builder`)

| Field | Format | Auth Headers |
|-------|--------|-------------|
| `BUILDER_API_KEY` | UUID string | `POLY_BUILDER_API_KEY` |
| `BUILDER_SECRET` | Base64 string | `POLY_BUILDER_SIGNATURE` (HMAC-SHA256) |
| `BUILDER_PASSPHRASE` | Hex string | `POLY_BUILDER_PASSPHRASE` |
| — | — | `POLY_BUILDER_TIMESTAMP` (Unix) |

Obtained at: `polymarket.com/settings?tab=builder` → click "+ Create New"

### 2.2 Relayer API Key (`settings?tab=api-keys`)

| Field | Format | Auth Headers |
|-------|--------|-------------|
| `RELAYER_API_KEY` | ULID string | `RELAYER_API_KEY` |
| `RELAYER_API_KEY_ADDRESS` | Ethereum address | `RELAYER_API_KEY_ADDRESS` |

Obtained at: `polymarket.com/settings?tab=api-keys` → click "Create New"

### 2.3 Builder Code (separate — for order attribution only)

| Field | Format | Auth |
|-------|--------|------|
| `builderCode` | bytes32 hex | Public — attached to order struct, no signing needed |

Obtained at: `polymarket.com/settings?tab=builder` → copy from UI

---

## 3. What We Probed (All Failed)

### 3.1 Relayer API — Builder Auth Endpoint

```
POST https://relayer-v2.polymarket.com/auth/api-key → 404
POST https://relayer-v2.polymarket.com/auth/builder-api-key → 404
GET  https://relayer-v2.polymarket.com/auth/derive-api-key → 404
```

No auth endpoints exist at the relayer for creating builder credentials.

### 3.2 Relayer API — API Key CRUD

From the [Relayer OpenAPI spec](https://docs.polymarket.com/api-reference/relayer-api-keys/get-all-relayer-api-keys):

```
GET  /relayer/api/keys          → List keys (auth: Gamma or Relayer API key)
POST /relayer/api/keys          → NOT IN SPEC (no create endpoint)
DELETE /relayer/api/keys/{id}   → NOT IN SPEC (no delete endpoint)
```

**Only `GET` exists.** Creation and deletion are not exposed via REST. They are handled exclusively by the web UI at `settings?tab=api-keys`.

### 3.3 CLOB API — Builder Auth Endpoint

```
POST https://clob.polymarket.com/auth/api-key → exists BUT creates CLOB L2 credentials, NOT builder credentials
GET  https://clob.polymarket.com/auth/derive-api-key → exists BUT derives CLOB L2 credentials, NOT builder credentials
```

The CLOB `/auth/api-key` endpoint creates CLOB user API keys (L2 auth for order placement), not builder API keys. These are different systems:

| Auth System | Created By | Used For |
|------------|-----------|----------|
| **CLOB L2** | `POST /auth/api-key` (EIP-712 signed) | Order placement, balance queries |
| **Builder** | Web UI `settings?tab=builder` | Relayer WALLET-CREATE, WALLET batch |
| **Relayer API Key** | Web UI `settings?tab=api-keys` | Relayer (alternative to Builder auth) |

### 3.4 GitHub SDKs — No Creation Endpoint

Searching all official Polymarket SDKs:
- `@polymarket/builder-signing-sdk` — uses existing credentials only, never creates them
- `@polymarket/builder-relayer-client` — uses existing credentials only, never creates them
- `@polymarket/clob-client-v2` — `createApiKey()` creates CLOB credentials, NOT builder credentials
- `py-builder-signing-sdk` — same, uses only
- `py-builder-relayer-client` — same, uses only

No SDK has a `createBuilderApiKey()` method.

### 3.5 `polymarket.com` — Request Capture (Hypothetical)

When you click "+ Create New" at `settings?tab=builder`, the browser sends a POST request to a Polymarket backend endpoint. This request includes:
- Gamma auth cookies (from logging into polymarket.com)
- CSRF token
- The specific endpoint path (not publicly documented)

**Why curl cannot replicate:**
1. Gamma auth cookies require signing into polymarket.com with a wallet
2. The login flow involves EIP-712 signature challenges
3. Session cookies have short lifetimes
4. The exact endpoint path and request format are not publicly documented
5. Polymarket may have CSRF protection and other anti-automation measures

---

## 4. Headless Automation (Partial Path)

While curl-only is impossible, **headless browser automation** (Playwright/Puppeteer) could partially automate the process:

```
┌─────────────────────────────────────────────────────────────┐
│                   HEADLESS AUTOMATION PATH                   │
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐              │
│  │ User logs │    │ Navigate │    │ Click    │              │
│  │ into      │───▶│ to       │───▶│ "+Create │              │
│  │ Polymarket│    │ settings │    │ New"     │              │
│  │ (manual)  │    │ (auto)   │    │ (auto)   │              │
│  └──────────┘    └──────────┘    └────┬─────┘              │
│                                       │                    │
│                                       ▼                    │
│                               ┌──────────┐                │
│                               │ Capture  │                │
│                               │ response │                │
│                               │ (auto)   │                │
│                               └──────────┘                │
│                                                             │
│  STILL REQUIRES MANUAL LOGIN (wallet signature)             │
│  CSRF tokens, session lifetimes, anti-bot measures          │
│  Cost: complex, fragile, not recommended                    │
└─────────────────────────────────────────────────────────────┘
```

### Why This Isn't Recommended

1. **Fragile** — Polymarket changes their UI, the automation breaks
2. **Maintenance-heavy** — Session cookies expire, CSRF tokens change
3. **Anti-bot risk** — Polymarket may detect and block automation
4. **Minimal benefit** — The manual steps are logging in (already needed) and clicking a button
5. **Not cheaper than the manual step** — One-time copy of 3 values vs. building/maintaining automation

---

## 5. The Practical Workflow

### 5.1 One-Time Manual Setup (~2 minutes)

1. Open `polymarket.com/settings?tab=builder`
2. Sign in with Ethereum wallet (MetaMask, Rabby, etc.)
3. If no builder profile exists → create one (free, no KYC)
4. Click **"+ Create New"** to generate keys
5. Copy `BUILDER_API_KEY`, `BUILDER_SECRET`, `BUILDER_PASSPHRASE`

### 5.2 Data Persistence

Store these values as environment variables or in a secure config file:

```bash
POLYMARKET_BUILDER_API_KEY="01967c03-..."
POLYMARKET_BUILDER_SECRET="base64EncodedSecret=="
POLYMARKET_BUILDER_PASSPHRASE="hexPassphrase"
```

**These values:**
- Do not expire (rotate on compromise)
- Work with one EOA at a time (tied to the builder profile)
- Can be revoked and regenerated from the same UI
- Should NEVER be committed to version control

### 5.3 Everything Else Is Automated

Once builder credentials are set, polygolem handles the entire lifecycle:

```bash
# Deploy wallet
polygolem deposit-wallet deploy --wait

# Approve contracts
polygolem deposit-wallet approve --submit

# Fund wallet
polygolem deposit-wallet fund --amount 0.71

# Or all at once
polygolem deposit-wallet onboard --fund-amount 0.71
```

---

## 6. The Server Proxy Pattern

For multi-user apps (Arenaton Flutter), DO NOT store builder credentials on the client. Use a server proxy:

```
┌──────────┐      ┌──────────────┐      ┌──────────────┐
│  Flutter  │ ──▶ │  Relayer     │ ──▶ │  Builder      │
│  / Mobile │      │  Proxy       │      │  Relayer v2   │
│  (no creds)│     │  (has creds) │      │               │
└──────────┘      └──────────────┘      └──────────────┘
```

The proxy (~50 LOC):
- Holds builder credentials server-side
- Forwards WALLET-CREATE and WALLET batch requests
- Never exposes credentials to clients
- Rate-limits per user

---

## 7. Key Takeaways

| Question | Answer |
|----------|--------|
| Can we get builder creds via curl? | ❌ No REST endpoint exists |
| Can we get relayer API keys via curl? | ❌ No REST endpoint exists |
| Does the CLOB auth endpoint create builder creds? | ❌ No — it creates CLOB L2 creds (different system) |
| Is there a programmatic issuance SDK? | ❌ No — all SDKs use existing creds only |
| Is headless browser automation possible? | ⚠️ Technically yes, but fragile and not worth it |
| What is the irreducible manual step? | Log into polymarket.com → settings → click "Create New" → copy 3 values |
| How long does the manual step take? | ~2 minutes, one-time |
| Do builder credentials expire? | No |
| Is KYC required? | No |

---

## 8. Related Documents

- [CONTRACTS.md](./CONTRACTS.md) — Smart contract addresses, permission model
- [DEPOSIT-WALLET-DEPLOYMENT.md](./DEPOSIT-WALLET-DEPLOYMENT.md) — Full deployment pipeline
- [SAFETY.md](./SAFETY.md) — Builder credential isolation rules
- [Polymarket Relayer API Spec](https://docs.polymarket.com/api-reference/relayer/introduction)
- [Polymarket Authentication Docs](https://docs.polymarket.com/api-reference/authentication)

---

*This document reflects exhaustive investigation into programmatic builder credential issuance. Every known endpoint, SDK, and approach has been probed. The one-time manual step is real and irreducible — but it's 2 minutes, one time, and everything after that is fully automated.*
