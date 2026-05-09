# Deposit-Wallet Live Trading Blocker Report — 2026-05-08

**Test Account:** 0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C
**Deposit Wallet:** 0x21999a074344610057c9b2B362332388a44502D4
**Funds:** Low (exact balance unknown due to balance endpoint failure)

> **Scope correction:** This report documents a live sigtype-3 deposit-wallet
> probe. It does not describe the corrected default V2 signup model. A later
> Playwright capture, recorded in `BLOCKERS.md` under "CORRECTION 2026-05-08",
> showed that fresh web-UI EOA signup registers a sigtype-1 proxy profile via
> `POST /profiles`, then mints a V2 relayer key. That proxy-profile path can be
> replicated headlessly. The blocker below applies to the deposit-wallet-owned
> CLOB API-key path, which remains unsupported by Polymarket's L1 auth endpoint.

---

## Executive Summary

The sigtype-3 deposit-wallet-only path is blocked for fresh accounts at the
deposit-wallet-owned CLOB API key. Polymarket's L1 auth endpoint rejects the
ERC-1271-style deposit-wallet key derivation used by that path, so balance,
order, cancellation, and heartbeat calls fail before reaching CLOB order
validation.

Separate finding: `GET /transaction?id=...` returns an array in production.
Polygolem's relayer polling used to expect a single object, which caused
`deposit-wallet deploy --wait` to fail after the transaction was already
confirmed on-chain. That decoder bug is covered by regression tests and fixed
in the relayer client.

---

## Test Results

### Step 1: Auth Status Check ✅ / ❌

```bash
polygolem auth status --check-deposit-key
```

**Result:**
```json
{
  "canTrade": false,
  "depositWallet": "0x21999a074344610057c9b2B362332388a44502D4",
  "depositWalletApiKeyExists": false,
  "depositWalletDeployed": false,
  "eoaAddress": "0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C",
  "eoaApiKeyExists": true,
  "nextStep": "Run: polygolem deposit-wallet deploy --wait"
}
```

**Finding:** EOA API key exists (created headlessly), but deposit-wallet API key does not exist.

---

### Step 2: Relayer Credential Minting ✅

```bash
polygolem auth headless-onboard
```

**Result:** ✅ Success
- SIWE login: 4 cookies captured
- V2 Relayer API Key minted: `019e08cd-de88-7a96-b9d4-ab1938a20cfb`
- Credentials persisted to `.env.relayer-v2`

**Finding:** Relayer credentials work perfectly headlessly.

---

### Step 3: Deposit Wallet Deployment ⚠️ PARTIAL

```bash
polygolem deposit-wallet deploy --wait
```

**Result:** Deployment transaction submitted and **confirmed on-chain** (tx: `0xf5272733...`), but CLI polling failed with:

```
WALLET-CREATE poll: relayer: get transaction: decode response:
  json: cannot unmarshal array into Go value of type relayer.RelayerTransaction
```

**Root cause:** The relayer's `GET /transaction?id=...` endpoint returns an **array** of transactions, but polygolem expects a **single object**.

**On-chain status (via curl):**
```json
[{
  "transactionID": "019e08ce-32a8-7dc1-8026-fcd03e396fd7",
  "transactionHash": "0xf527273371afdc9881e01ee77a06ceac7982779a74710aed71a8bf426db43194",
  "state": "STATE_CONFIRMED",
  "type": "WALLET-CREATE"
}]
```

**Finding:** The wallet IS deployed on-chain, but:
1. `PollTransaction` had a JSON decoding bug. Fixed: `GetTransaction` now
   accepts both legacy object responses and production array responses.
2. `deposit-wallet status` still reported `deployed: false` in this live run.
   That remains a separate live follow-up; no local evidence yet proves whether
   `/deployed` expects a different address, lags behind transaction state, or
   behaved differently because the profile/proxy model was not registered.

---

### Step 4: Deposit-Wallet-Owned API Key Creation ❌ BLOCKED

```bash
polygolem clob create-api-key-for-address --owner 0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C
```

**Result:**
```
HTTP 401 https://clob.polymarket.com/auth/api-key: {"error":"Invalid L1 Request headers"}
```

**Finding:** This is the **primary blocker**. Polymarket's L1 auth endpoint does not support ERC-1271 validation for deposit wallets. Confirmed empirically.

---

### Step 5: Balance Check ❌ BLOCKED

```bash
polygolem clob balance --asset-type collateral
```

**Result:**
```
derive deposit-wallet api key: HTTP 401 https://clob.polymarket.com/auth/derive-api-key:
  {"error":"Invalid L1 Request headers"}
```

**Finding:** Balance check fails because it attempts to derive the deposit-wallet API key first. Cannot even read balance without the deposit key.

---

### Step 6: Order Placement ❌ BLOCKED

```bash
polygolem clob create-order --token 8501497159083948713316135768103773293754490207922884688769443031624417212426 --side buy --price 0.01 --size 0.01
```

**Result:**
```
derive deposit-wallet api key: HTTP 401 https://clob.polymarket.com/auth/derive-api-key:
  {"error":"Invalid L1 Request headers"}
```

**Finding:** Order placement fails at the API-key derivation step — never reaches order validation.

---

## Blocker Matrix

| Operation | Status | Blocker | Severity |
|-----------|--------|---------|----------|
| SIWE login | ✅ Works | None | — |
| Relayer key mint | ✅ Works | None | — |
| Deposit wallet deploy | ✅ Fixed in code | `PollTransaction` now decodes array or object responses | Medium |
| Deposit wallet status | ⚠️ Needs follow-up | `IsDeployed` returned false despite on-chain deployment in this live run | Low |
| Deposit-wallet API key | ❌ **Blocked** | Server returns 401, no ERC-1271 support | **Critical** |
| Balance check | ❌ **Blocked** | Depends on deposit-wallet API key | **Critical** |
| Order creation | ❌ **Blocked** | Depends on deposit-wallet API key | **Critical** |
| Batch orders | ❌ **Blocked** | Depends on deposit-wallet API key | **Critical** |
| Market order | ❌ **Blocked** | Depends on deposit-wallet API key | **Critical** |
| Order cancellation | ❌ **Blocked** | Depends on deposit-wallet API key | **Critical** |
| Heartbeat | ❌ **Blocked** | Depends on deposit-wallet API key | **Critical** |
| Public market data | ✅ Works | None | — |

---

## Critical Finding: Deposit-Wallet API Key is the Single Point of Failure

Every authenticated CLOB operation in polygolem routes through `depositWalletAPIKey()` which:
1. Derives the deposit wallet address from the EOA
2. Calls `DeriveAPIKeyForAddress()` with the deposit wallet as owner
3. This sends L1 auth headers with `POLY_ADDRESS=<deposit_wallet>`
4. Server returns 401 because it cannot validate ERC-1271 signatures

**This means:** Without browser signup, **zero** authenticated operations are possible. Not balance checks, not orders, not cancellations.

---

## Secondary Bug: Relayer Polling

`PollTransaction` in `internal/relayer/` expects:
```go
var tx RelayerTransaction
json.Unmarshal(body, &tx)
```

But the relayer returns:
```json
[{"transactionID": "...", "state": "STATE_CONFIRMED", ...}]
```

**Fix:** Change to:
```go
var txs []RelayerTransaction
json.Unmarshal(body, &txs)
tx := txs[0]
```

---

## Workaround Status

| Workaround | Viable? | Notes |
|------------|---------|-------|
| Corrected V2 proxy-profile signup | ❌ Not for Polygolem | Proxy signup is a different wallet family; Polygolem's production path is deposit wallet / POLY_1271. |
| Browser login to create deposit-wallet-owned API key | ✅ Yes for new users | One-time browser flow mints the deposit-wallet CLOB key; after that, Polygolem runs headlessly. |
| EOA-owned API key + deposit orders | ❌ No | Server enforces owner gate |
| Wait for Polymarket fix | ❌ Unknown | No timeline, no public commitment |
| Deposit-wallet-only headless trading | ❌ No | Blocked by deposit-wallet API-key derivation on `/auth/derive-api-key` |

---

## Recommendations

1. **Done in this slice:** Fix `PollTransaction` / `GetTransaction` to decode
   production array responses.
2. **Next live follow-up:** Re-test `IsDeployed` after profile registration and
   record whether `/deployed` expects owner, proxy, or deposit-wallet address.
3. **Default onboarding:** Prefer the corrected proxy-profile signup flow for
   fresh EOAs: SIWE login, `POST /profiles`, then V2 relayer auth.
4. **Deposit-wallet path:** Keep documenting the ERC-1271 L1-auth limitation
   until Polymarket supports deposit-wallet API-key derivation.

---

*Report generated from live test with real funds on Polygon mainnet.*
