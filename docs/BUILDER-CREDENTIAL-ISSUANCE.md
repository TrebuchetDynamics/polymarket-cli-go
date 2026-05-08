# Builder Credential Issuance — How to Get Your Keys

> **Date:** 2026-05-07
> **Status:** Complete — one manual step, then fully automated
> **Companion:** [CONTRACTS.md](./CONTRACTS.md), [DEPOSIT-WALLET-DEPLOYMENT.md](./DEPOSIT-WALLET-DEPLOYMENT.md)

---

## 1. The Bottom Line

Getting builder credentials takes **~2 minutes, one time, forever.** After that, polygolem handles everything automatically.

### What's Great About This

| Don't Need | Why It's a Win |
|------------|---------------|
| ✅ No KYC | Unverified tier starts immediately. No ID, no documents, no approval wait. |
| ✅ No paid tier | Full relayer access is free. Verified tier is optional for higher rate limits. |
| ✅ No expiration | Builder credentials never expire. Rotate only if compromised. |
| ✅ No programmatic complexity | No API integration needed for credential management. Copy 3 values once. |
| ✅ No multi-step web flow | One page: `polymarket.com/settings?tab=builder`, click "Create New", copy keys. |

---

## 2. The Two-Minute Setup

1. Open `polymarket.com/settings?tab=builder` in a browser
2. Sign in with your Ethereum wallet (MetaMask, Rabby, etc.)
3. If no builder profile exists → create one (free, instant, no KYC)
4. Click **"+ Create New"** to generate API keys
5. Copy these three values:

| Variable | What It Looks Like |
|----------|-------------------|
| `BUILDER_API_KEY` | UUID string (e.g., `01967c03-b8c8-7000-8f68-8b8eaec6fd3d`) |
| `BUILDER_SECRET` | Base64-encoded string |
| `BUILDER_PASSPHRASE` | Hex string |

Store them as environment variables:

```bash
POLYMARKET_BUILDER_API_KEY="01967c03-..."
POLYMARKET_BUILDER_SECRET="base64EncodedSecret=="
POLYMARKET_BUILDER_PASSPHRASE="hexPassphrase"
```

That's it. You're done with the manual part. Everything after this is automated via polygolem.

---

## 3. Two Credential Types (Both from the Web UI)

Polymarket has two separate credential systems for relayer access:

### 3.1 Builder API Key (`settings?tab=builder`)

| Field | Format | Auth Headers |
|-------|--------|-------------|
| `BUILDER_API_KEY` | UUID string | `POLY_BUILDER_API_KEY` |
| `BUILDER_SECRET` | Base64 string | `POLY_BUILDER_SIGNATURE` (HMAC-SHA256) |
| `BUILDER_PASSPHRASE` | Hex string | `POLY_BUILDER_PASSPHRASE` |
| — | — | `POLY_BUILDER_TIMESTAMP` (Unix) |

Used for: WALLET-CREATE, WALLET batch, order attribution via relayer

### 3.2 Relayer API Key (`settings?tab=api-keys`)

| Field | Format | Auth Headers |
|-------|--------|-------------|
| `RELAYER_API_KEY` | ULID string | `RELAYER_API_KEY` |
| `RELAYER_API_KEY_ADDRESS` | Ethereum address | `RELAYER_API_KEY_ADDRESS` |

Used for: Alternative to Builder API Key for relayer access

**Both are created in the Polymarket web UI.** Both support the same relayer endpoints.

### 3.3 Builder Code (for Order Attribution Only)

| Field | Format | Auth |
|-------|--------|------|
| `builderCode` | bytes32 hex | Public — attached to order struct, no signing needed |

The builder code is a public identifier. You include it in every order to get credit on the Builder Leaderboard. No HMAC, no headers — just a bytes32 field in the order.

---

## 4. What We Investigated (For Reference)

### 4.1 Why Curl-Only Doesn't Work

Relayer API endpoints probed:

```
POST https://relayer-v2.polymarket.com/auth/api-key         → 404
POST https://relayer-v2.polymarket.com/auth/builder-api-key → 404
GET  https://relayer-v2.polymarket.com/auth/derive-api-key  → 404
```

CLOB endpoints create CLOB L2 credentials (for trading), NOT builder credentials:

```
POST https://clob.polymarket.com/auth/api-key        → creates CLOB L2 (different system)
GET  https://clob.polymarket.com/auth/derive-api-key → derives CLOB L2 (different system)
```

The Relayer OpenAPI spec only lists:

```
GET /relayer/api/keys → List keys (auth: Gamma auth or Relayer API key)
```

No `POST /relayer/api/keys` exists. Creation is exclusively via the web UI.

### 4.2 Why This Isn't a Problem

The manual step is:
- **2 minutes** total
- **One-time** (credentials never expire)
- **No KYC** (no documents, no approval wait)
- **Free** (Unverified tier)
- **Wallet-only** (no email/phone needed)

For comparison, setting up an exchange API key takes just as long. This is standard for any API credential issuance.

---

## 5. Headless Automation (Not Recommended)

While curl-only is unavailable, headless browser automation (Playwright/Puppeteer) could theoretically automate the web UI interaction. However:

- Requires maintaining a login session to polymarket.com
- Session cookies expire
- CSRF tokens must be handled
- Polymarket UI changes would break automation
- Much more fragile than 2 minutes of manual work

**Recommendation:** Just copy the keys once. It's faster, more reliable, and doesn't need maintenance.

---

## 6. Server Proxy for Multi-User Apps

For Flutter/mobile apps like Arenaton, builder credentials must stay server-side:

```
Flutter App (no creds) → Server Proxy (holds creds) → Builder Relayer v2
```

The proxy (~50 LOC) forwards WALLET-CREATE and WALLET batch requests, adding builder HMAC headers from server-side environment variables. Builder credentials are never exposed to client code.

---

## 7. Summary

| Question | Answer |
|----------|--------|
| Can we get builder creds via curl? | No — but the 2-minute manual step is standard for API key issuance |
| Do builder credentials expire? | Never — set once, use forever |
| Is KYC required? | Not at all — Unverified tier starts immediately |
| Is the process free? | Yes — Unverified tier includes full relayer access |
| What's the only manual step? | polymarket.com → settings → "Create New" → copy 3 values |
| How long does it take? | ~2 minutes, one-time |
| What about multi-user apps? | Use a server proxy — builder creds stay server-side |

---

## 8. Related Documents

- [CONTRACTS.md](./CONTRACTS.md) — Smart contract addresses, permission model
- [DEPOSIT-WALLET-DEPLOYMENT.md](./DEPOSIT-WALLET-DEPLOYMENT.md) — Full deployment pipeline
- [SAFETY.md](./SAFETY.md) — Builder credential isolation rules
- [Polymarket Relayer API Spec](https://docs.polymarket.com/api-reference/relayer/introduction)
- [Polymarket Authentication Docs](https://docs.polymarket.com/api-reference/authentication)
