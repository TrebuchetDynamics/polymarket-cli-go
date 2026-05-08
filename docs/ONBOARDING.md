# Onboarding — Single Source of Truth

**Last updated:** 2026-05-08
**Status:** Fully headless onboarding verified end-to-end against Polymarket prod with deposit wallet path
**Companion:** [BROWSER-SETUP.md](./BROWSER-SETUP.md) — only for users coming from Magic Link / social signup

This document supersedes any conflicting onboarding instructions in
README.md, ARCHITECTURE.md, BROWSER-SETUP.md, BUILDER-AUTO.md,
DEPOSIT-WALLET-DEPLOYMENT.md, DEPOSIT-WALLET-MIGRATION.md, or
LIVE-TRADING-BLOCKER-REPORT.md. If those disagree with this file, this
file is correct.

---

## TL;DR

| User type | Path | Browser required? |
|---|---|---|
| Fresh EOA, no Polymarket account | `auth headless-onboard --signature-type=3` | ❌ No |
| Existing EOA-based Polymarket user | `auth headless-onboard --skip-profile` | ❌ No |
| Existing Magic Link / social-signup user | Recover the EOA private key once via reveal.magic.link, then same as above | ✅ Once, only for key extraction |

The earlier "browser required for new users" claim (in older revisions
of this doc, BROWSER-SETUP.md, and LIVE-TRADING-BLOCKER-REPORT.md) was
based on a wrong root-cause attribution. The 401 we hit on
`/auth/api-key` was not the deposit-wallet ERC-1271 problem; it was a
profile-registration gate. `POST /gamma-api/profiles` is what
registers a maker address with Polymarket's backend, and after that,
sigtype-1 / sigtype-3 orders are accepted. polygolem now does this
automatically. See [BLOCKERS.md "CORRECTION 2026-05-08"](../BLOCKERS.md)
and [scripts/playwright-capture/](../scripts/playwright-capture/) for
the captured signup flow that proves it.

---

## The flow

```
┌──────────────────────────────┐
│  1. Hold an EOA private key  │   any 32-byte key; openssl rand -hex 32
└──────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────────────┐
│  2-4. polygolem auth headless-onboard                │
│       (a) SIWE personal_sign over standard message   │
│       (b) GET /gamma-api/login   → session cookie    │
│       (c) POST /gamma-api/profiles                   │
│           body.proxyWallet = MakerAddress(eoa,137,N) │
│       (d) POST /relayer-v2/relayer/api/auth  → key   │
└──────────────────────────────────────────────────────┘
              │
              ▼   only for sigtype-3 (deposit wallet)
┌──────────────────────────────────────────────────────┐
│  5. polygolem deposit-wallet onboard                 │
│       deploy (relayer, gasless) → approvals batch    │
│       → fund pUSD (EOA pays ~$0.01 POL gas)          │
└──────────────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────────────┐
│  6. polygolem clob create-order ...                  │
└──────────────────────────────────────────────────────┘
```

Steps 2–4 are one CLI call. Steps 5 is one CLI call. Total time on a
fresh EOA: ~30 seconds.

---

## Quickstart

### Fresh EOA, deposit wallet (recommended)

```bash
export POLYMARKET_PRIVATE_KEY=0x...
polygolem auth headless-onboard --signature-type=3 \
    --env-file=./.env.relayer-v2
polygolem deposit-wallet onboard --fund-amount 0.71
polygolem clob update-balance --asset-type collateral
polygolem clob create-order \
    --token <TOKEN_ID> --side buy --price 0.5 --size 10
```

### Existing Polymarket user

```bash
export POLYMARKET_PRIVATE_KEY=0x...
polygolem auth headless-onboard --skip-profile  # profile already registered
# everything else identical to fresh EOA path
```

### Sigtype-1 proxy instead of deposit wallet

```bash
polygolem auth headless-onboard --signature-type=1
# no deposit-wallet onboard step needed; proxy is implicit
```

---

## What works headlessly

| Operation | Command | Notes |
|---|---|---|
| Generate EOA | `openssl rand -hex 32` (or any wallet) | local |
| SIWE login | `auth headless-onboard` (step a-b) | gamma-api.polymarket.com |
| Profile registration | `auth headless-onboard` (step c) | `/profiles` |
| V2 relayer key mint | `auth headless-onboard` (step d) | `/relayer/api/auth` |
| Deposit wallet derive | `deposit-wallet derive` | local CREATE2 |
| Deposit wallet deploy | `deposit-wallet deploy --wait` | gasless via relayer, ~2 s |
| Approvals batch | `deposit-wallet approve --submit` | gasless |
| Fund pUSD | `deposit-wallet fund --amount N` | EOA pays gas ~$0.01 POL |
| Balance sync | `clob update-balance ...` | gasless |
| Order create/cancel | `clob create-order`, `clob cancel-order` | sigtypes 0, 1, 3 |
| Builder profile create | `clob create-builder-fee-key` | optional, fee revenue |

Once an EOA is registered, every subsequent operation is fully
headless. Re-registration is not needed across restarts.

---

## What requires a browser

These are intentional Polymarket UX choices that polygolem cannot and
should not replicate:

| Path | Why it's browser-only |
|---|---|
| Magic Link email signup | Magic SDK provisions a custodial EOA inside an iframe; the user has no key in hand. Recover at https://reveal.magic.link/polymarket. |
| Privy email/social signup | Same custodial pattern via Privy. |
| Google / Telegram / Steam OAuth | OAuth redirect handshake is browser-only by spec. |
| WebAuthn / passkey | Requires a platform authenticator. |

**If you signed up via any of these, you do have an EOA — it's just
custodial.** Run the recovery flow once (browser), extract the key,
then proceed exactly like a fresh EOA via this doc. After that, no
browser is ever needed.

---

## Step-by-step detail

### Step 1 — EOA private key

```bash
openssl rand -hex 32 | xargs -I{} echo "0x{}"
# or use any existing Ethereum wallet
export POLYMARKET_PRIVATE_KEY=0x...
```

### Step 2-4 — Auth + profile + relayer key

```bash
polygolem auth headless-onboard --signature-type=3
```

What it does (verified 2026-05-08 against prod with throwaway EOA
`0xe2A850703644BB8b341F2a07B34b2F9906120771` → profile id `8041373`,
relayer key `019e08d8-...`):

1. `GET /gamma-api/nonce` — fresh nonce
2. `personal_sign` over the EIP-4361 SIWE message
3. `GET /gamma-api/login` with `Authorization: Bearer base64(<SIWE-JSON>:::<sig>)` — session cookie issued
4. `POST /gamma-api/profiles` with body matching the captured web-UI shape:
   ```json
   {"displayUsernamePublic": true, "emailOptIn": false,
    "walletActivated": false, "name": "<proxy>-<unixms>",
    "pseudonym": "<proxy>", "proxyWallet": "<proxy>",
    "users": [{"address": "<EOA>", "isExternalAuth": true,
               "proxyWallet": "<proxy>", "username": "<proxy>-<unixms>",
               "provider": "metamask", "preferences": [...],
               "walletPreferences": [...]}]}
   ```
   `<proxy>` is `MakerAddressForSignatureType(eoa, 137, --signature-type)`.
5. `POST /relayer-v2/relayer/api/auth` with body `{}` — returns
   `{apiKey, address, createdAt}`
6. Writes `RELAYER_API_KEY` + `RELAYER_API_KEY_ADDRESS` to env file
   (mode 0600)

Flags:

| Flag | Default | Purpose |
|---|---|---|
| `--signature-type` | `3` | 0=EOA, 1=proxy, 3=deposit wallet |
| `--skip-profile` | `false` | skip step (4); use if profile already exists |
| `--env-file` | `../go-bot/.env.relayer-v2` | output path |
| `--gamma-url` | `https://gamma-api.polymarket.com` | override |
| `--relayer-url` | `https://relayer-v2.polymarket.com` | override |
| `--force` | `false` | overwrite existing env file |
| `--json` | (global) | structured output |

HTTP 409 from `/profiles` is auto-tolerated (re-runs are safe). Other
non-2xx fail loudly.

### Step 5 — Deposit wallet (sigtype-3 only)

```bash
polygolem deposit-wallet onboard --fund-amount 0.71
```

Internally:

1. Derive deposit wallet via CREATE2 (local)
2. WALLET-CREATE through the V2 relayer (gasless; ~2 s on-chain confirm)
3. Approval batch — pUSD + CTF for all V2 exchange spenders (gasless)
4. ERC-20 transfer from EOA to deposit wallet (`--fund-amount` pUSD; EOA pays gas ~$0.01 POL)

Skip if you're using sigtype 0/1 (no deposit wallet involved).

### Step 6 — Trade

```bash
polygolem clob update-balance --asset-type collateral
polygolem clob create-order \
    --token <TOKEN_ID> --side buy --price 0.5 --size 10
```

Order signing automatically uses the sigtype set at onboarding.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `HTTP 400 "maker address not allowed, please use the deposit wallet flow"` | Profile not registered (you ran the V1-style flow that skipped `/profiles`) | Run `auth headless-onboard` once. The /profiles call is the gate. |
| `HTTP 409 profile already exists` | Re-running onboarding | Auto-tolerated — proceed. Or pass `--skip-profile` next time. |
| `HTTP 401 Invalid L1 Request headers` on `/auth/api-key` | Calling the deprecated V1 L1 mint endpoint | Don't. V2 uses `/relayer/api/auth`, not `/auth/api-key`. The V1 endpoint requires ECDSA-recoverable POLY_SIGNATURE and rejects ERC-1271 wraps. |
| `Deposit wallet not deployed` | Skipped step 5 | `polygolem deposit-wallet deploy --wait` |
| `pUSD balance is zero` | Wallet not funded | `polygolem deposit-wallet fund --amount X` |
| `the order signer address has to be the address of the API KEY` | Mismatch between API-key owner and order signer | The V2 relayer key is bound to the EOA, not the deposit wallet. Sigtype-3 orders use `signer == maker == depositWallet` per V2 contract — the API key authenticates at the HTTP layer, the EIP-712 signature authenticates at the contract layer. They don't have to match. If you're seeing this, you may be on a stale V1 code path; rebuild from main. |

---

## Verification

```bash
polygolem auth status --check-deposit-key
```

Expected for sigtype-3 ready-to-trade:

```json
{
  "eoaAddress": "0x...",
  "depositWallet": "0x...",
  "depositWalletDeployed": true,
  "profileRegistered": true,
  "relayerKeyExists": true,
  "canTrade": true
}
```

---

## Reference

- [BLOCKERS.md](../BLOCKERS.md) — `CORRECTION 2026-05-08` section has the full captured request/response bodies for steps a–d
- [scripts/playwright-capture/](../scripts/playwright-capture/) — investigation tooling that produced the capture
- [BROWSER-SETUP.md](./BROWSER-SETUP.md) — only relevant for Magic Link / social-signup users who need to recover their EOA key
- [DEPOSIT-WALLET-DEPLOYMENT.md](./DEPOSIT-WALLET-DEPLOYMENT.md) — on-chain CREATE2 mechanics for step 5
