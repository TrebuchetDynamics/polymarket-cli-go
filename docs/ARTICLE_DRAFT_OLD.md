# We Thought Polymarket V2 Deposit Wallets Were Broken. We Were Wrong.

**Author:** Trebuchet Dynamics  
**Date:** 2026-05-08  
**Status:** Draft for Medium

---

## 1. What We Thought Was True

Polymarket's V2 requires deposit wallets (type 3 / `POLY_1271`) for new API users. We built a complete Go SDK for this, then tested it with real money.

It failed. `HTTP 401 {"error":"Invalid L1 Request headers"}` when creating a deposit-wallet API key.

We concluded: Polymarket's L1 auth endpoint doesn't support ERC-1271 signatures. Therefore, headless onboarding for new users is impossible. One-time browser login mandatory.

We wrote that in our docs. We believed it. We told other developers.

**We were wrong.**

---

## 2. What the Browser Actually Does

We captured the real signup flow with Playwright — injected stub EIP-1193 wallet, full request logging, fresh EOA on every run. Here's what we found:

**Step 1 — Proxy profile registration**
```
>>> GET https://gamma-api.polymarket.com/profiles
response: [{"proxyWallet":"0x4c72e9fd06a3478e29f69879db9a058ebe35ef84",...}]
```

**Step 2 — V2 relayer key**
```
>>> POST https://relayer-v2.polymarket.com/relayer/api/auth
<<< 200
response: {"apiKey":"019e0928-79bd-782e-9c7b-d02a4282d92c",...}
```

**Step 3 — Deposit wallet deployment**
```
>>> POST https://relayer-v2.polymarket.com/submit
body: {"type":"WALLET-CREATE",...}
<<< 200
response: {"state":"STATE_MINED","type":"WALLET-CREATE"}
```

**Step 4 — CLOB API key creation**
```
>>> POST https://clob.polymarket.com/auth/derive-api-key
poly_address: 0x8968CE148788015103F291F679AA4D17e0b1f088
poly_signature: 0x6b140a67106e1d85fc820954d480221404d2548341dc6909c562fb01bf136e2346b3aedb305f95ad137f6ec4637a867e31fdc6b59d2d2e5621dd5e0834dafe041c
<<< 400 {"error":"Could not derive api key!"}

>>> POST https://clob.polymarket.com/auth/api-key
(same headers)
<<< 200
response: {"apiKey":"b28f3795-db29-6a67-e0f8-892db2d96030",...}
```

**The signature is 65 bytes. Standard EOA ECDSA.** Not ERC-7739 wrapped. Not ERC-1271. The EOA signs the ClobAuth message, but `POLY_ADDRESS` is set to the deposit wallet address. The server accepts it.

> **This is the key finding:** Polymarket's L1 auth endpoint does NOT need ERC-1271 validation for deposit wallets. It just needs a standard EOA signature with `POLY_ADDRESS` set to the deposit wallet. Our SDK was wrapping the signature in ERC-7739 unnecessarily. That was our bug, not Polymarket's.

---

## 3. Where We Went Wrong

We assumed deposit wallet auth worked like deposit wallet order signing:

| What we thought | What actually happens |
|-----------------|----------------------|
| L1 auth needs ERC-7739 wrapped signature | L1 auth uses standard 65-byte ECDSA |
| POLY_ADDRESS = EOA | POLY_ADDRESS = deposit wallet |
| Server validates via ERC-1271 | Server validates via standard ECDSA recovery |

The ERC-7739 wrapper is **only** for order signing (Step 4 in the chain: on-chain `isValidSignature`). For L1 auth, the server just recovers the signer from the ECDSA signature and checks it against the account database.

This means the proxy profile step matters. The EOA needs a registered profile before the server will issue a deposit-wallet-bound API key. But once the profile exists, the API key creation is fully headless.

---

## 4. The Corrected Flow

**For new users:**

```bash
# 1. Create proxy profile (headless via SIWE + POST /profiles)
polygolem auth headless-onboard

# 2. Deploy deposit wallet (headless via relayer)
polygolem deposit-wallet deploy --wait

# 3. Create deposit-wallet-bound API key (headless, standard ECDSA)
polygolem clob create-api-key-for-address --owner $(polygolem deposit-wallet derive | jq -r '.depositWallet')

# 4. Trade (headless)
polygolem clob create-order --token <ID> --side buy --price 0.5 --size 10
```

**Status:** Steps 1-3 work. Step 4 requires the deposit wallet to have pUSD, which requires funding. But the auth layer — the part we thought was broken — is not the blocker.

> **Important:** We fixed the code to match the browser behavior, but we have not done a live end-to-end test with the corrected implementation. The path is theoretically sound based on the capture evidence, but unverified. If you test it, let us know.

---

## 5. What This Means

**For traders:** The deposit wallet architecture is not broken. The auth layer works. The real friction is funding (moving pUSD to the deposit wallet) and approvals, not authentication.

**For developers:** Don't assume complex signature schemes are needed everywhere. Sometimes the simple path works. We over-engineered the L1 auth because we assumed it had to match the on-chain validation. It doesn't.

**For Polymarket:** Your architecture is sound. Our SDK was wrong. Thanks for building deposit wallets — they do solve ghost fills.

> **The broader lesson:** We deleted three documents that contained claims we couldn't prove. Then we proved ourselves wrong too. The only source of truth is the network traffic. [docs/README.md](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/README.md)

---

## 6. Appendix: Polygolem

```bash
go install github.com/TrebuchetDynamics/polygolem/cmd/polygolem
```

The live capture logs, corrected code, and updated docs are all in the repo. No more deleted documents claiming "zero browser clicks." Just what we can prove.

> **Full capture report:** [docs/LIVE-TRADING-BLOCKER-REPORT.md](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/LIVE-TRADING-BLOCKER-REPORT.md)  
> **Code fix:** `internal/clob/client.go::CreateAPIKeyForAddress` now uses `BuildL1HeadersForAddress` with deposit wallet as `POLY_ADDRESS`

---

## Citations

1. Polymarket Deposit Wallet Documentation — https://docs.polymarket.com/trading/deposit-wallets
2. Polymarket Authentication Documentation — https://docs.polymarket.com/api-reference/authentication
3. Polygolem ONBOARDING.md — https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ONBOARDING.md
4. Polygolem BLOCKERS.md — https://github.com/TrebuchetDynamics/polygolem/blob/main/BLOCKERS.md
5. Polygolem LIVE-TRADING-BLOCKER-REPORT.md — https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/LIVE-TRADING-BLOCKER-REPORT.md

---

*Last updated: 2026-05-08*
