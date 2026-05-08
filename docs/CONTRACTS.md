# Polymarket Smart Contracts — Deposit Wallet Research

> **Date:** 2026-05-07
> **Status:** Complete — investigation confirms all-on-chain deployment is impossible
> **Last verified:** Live on-chain against Polygon mainnet (chainID 137)

---

## 1. Contract Registry

### 1.1 Core Contracts

| Contract | Address | Role |
|----------|---------|------|
| **DepositWalletFactory** | `0x00000000000Fb5C9ADea0298D729A0CB3823Cc07` | CREATE2 deploys per-user deposit wallets (POLY_1271) |
| **Proxy Factory** | `0xaB45c5A4B0c941a2F231C04C3f49182e1A254052` | Older POLY_PROXY wallets (type 1, grandfathered) |
| **Gnosis Safe Factory** | `0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b` | Gnosis Safe wallets (type 2) |
| **CTF Exchange V2** | `0xE111180000d2663C0091e4f400237545B87B996B` | Order matching and settlement |
| **Neg Risk CTF Exchange** | `0xe2222d279d744050d28e00520010520000310F59` | Neg-risk markets |
| **Neg Risk Adapter** | `0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296` | Neg-risk adapter |
| **pUSD (proxy)** | `0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB` | Collateral token |
| **pUSD (impl)** | `0x6bBCef9f7ef3B6C592c99e0f206a0DE94Ad0925f` | Collateral implementation |
| **CTF** | `0x4D97DCd97eC945f40cF65F87097ACe5EA0476045` | Conditional Tokens Framework (ERC-1155) |

### 1.2 Which Wallet Type Wins

**New API users (post-May 2026):** Deposit wallet (type 3 / POLY_1271) only. EOA is blocked by CLOB V2.
**Grandfathered users:** Proxy (type 1) or Safe (type 2) still work.

---

## 2. DepositWalletFactory (`0x00000000000Fb5C9ADea0298D729A0CB3823Cc07`)

### 2.1 Factory Properties

| Property | Value |
|----------|-------|
| Compiler | v0.8.34+commit.80d5c536, optimization 1000000 runs |
| License | NA (unverified but decompilable) |
| CREATE2 salt | `keccak256(abi.encodePacked(ownerAddress))` |
| Implementation | Set via `initialize(address _admin, address _operator, address _implementation)` |
| Upgradeable | ERC-1967 proxy pattern |

### 2.2 ABI — Public/External Functions

```
deploy(address[] _owners, bytes32[] _ids)
proxy(Batch[] _batches, bytes[] _signatures)           ← UNGATED
predictWalletAddress(address _implementation, bytes32 _id)
implementation()
authorizedImplementation(address)
addAdmin/removeAdmin/addOperator/removeOperator/...
grantRoles/revokeRoles/hasAllRoles/hasAnyRole/rolesOf
```

### 2.3 ABI Errors (Permission Model)

```
OnlyAdmin, OnlyOperator, OnlyRoles
Unauthorized, UnauthorizedCallContext
```

Presence of `OnlyAdmin` and `OnlyOperator` errors confirms role-gated functions.

### 2.4 deploy() is Role-Gated

**Empirical proof:** Calling `deploy()` from a non-privileged EOA reverts with "execution reverted." Confirmed on 2026-05-07 against production.

**Who has the operator role:** Polymarket's relayer EOA. No third party holds this role.

**Conclusion:** Only the Polymarket relayer can deploy deposit wallets. Direct EOA deployment is impossible.

### 2.5 proxy(Batch[], bytes[]) is Ungated

The `proxy()` function uses internal signature validation — it does not check caller roles:

> *proxy(Batch[], bytes[]) appears UNGATED (signature validation internal). This means after a wallet is deployed, anyone can submit signed batches on its behalf — including the wallet owner directly from their EOA.*

**This is the key permissionless path.** Post-deployment, the wallet owner controls everything via EOA-signed batches.

### 2.6 predictWalletAddress() — View Function

Ungated. Predicts the CREATE2 address without deploying. This is the path polygolem uses in `deposit-wallet derive`.

---

## 3. CREATE2 Address Derivation

### 3.1 Why You Can't Bypass the Factory

The CREATE2 formula is:

```
address = keccak256(0xff + factory_address + salt + keccak256(initCode))[12:]
```

Where:
- `factory_address` = `0x00000000000Fb5C9ADea0298D729A0CB3823Cc07` (hard-coded)
- `salt` = `keccak256(abi.encodePacked(EOA_address))`
- `initCode` = `keccak256(clone_bytecode)`

**The factory address is embedded in the derivation.** Even if you knew the exact init code and salt, deploying from YOUR address would produce a DIFFERENT address. The CLOB only recognizes wallets at factory-derived addresses.

### 3.2 polygolem Derivation (Go)

`internal/wallet/derive.go` implements CREATE2 local computation:

```go
func deriveCreate2(factory string, salt []byte, initCodeHash string) string {
    hash := sha3.NewLegacyKeccak256()
    hash.Write([]byte{0xff})
    hash.Write(hexToBytes(strip0x(factory)))
    hash.Write(salt)
    hash.Write(hexToBytes(strip0x(initCodeHash)))
    result := hash.Sum(nil)
    return "0x" + toHex(result[12:])
}
```

This is pure local computation — no on-chain call needed for derivation.

---

## 4. Builder Credentials

### 4.1 What They Are

| Credential | Format | Used For |
|-----------|--------|----------|
| `BUILDER_API_KEY` | UUID | Relayer authentication |
| `BUILDER_SECRET` | Base64 | HMAC-SHA256 signing |
| `BUILDER_PASSPHRASE` | Hex string | Authentication identifier |
| `builderCode` | bytes32 | Order attribution (public, on-chain) |

### 4.2 How to Obtain

Preferred V2 relayer keys are minted by `polygolem auth headless-onboard`
through SIWE login plus `POST /relayer/api/auth`. The settings page remains a
manual fallback for legacy builder-relayer HMAC credentials.

- **No KYC required** for the Unverified tier
- **Geoblock applies** — 33 restricted countries are blocked
- **No programmatic endpoint** exists at `relayer-v2.polymarket.com/auth/api-key` (returns 404)
- The V2 relayer key endpoint is `relayer-v2.polymarket.com/relayer/api/auth`
  and requires a SIWE-backed session.

### 4.3 Builder Credentials vs CLOB Auth

| Auth System | Credentials | Scope |
|------------|-------------|-------|
| **Relayer auth** | `RELAYER_API_KEY` + `RELAYER_API_KEY_ADDRESS`, or legacy `BUILDER_API_KEY` / `SECRET` / `PASSPHRASE` + HMAC-SHA256 | Wallet lifecycle (deploy, batch, approve) |
| **CLOB L1** | EOA private key + EIP-712 typed data signature | Create/derive CLOB API keys |
| **CLOB L2** | CLOB `apiKey` / `secret` / `passphrase` + HMAC-SHA256 | Order placement, balance queries, trades |

These systems are **independent**. Builder credentials cannot be substituted for CLOB auth and vice versa.

### 4.4 Security Rules

- Builder credentials must NEVER be exposed client-side (mobile, browser)
- Use environment variables or a secrets manager on the server
- Credentials are redacted by `internal/config` on every load
- Rotate builder keys if compromised

---

## 5. Relayer v2

### 5.1 Endpoints

| Host | Path | Auth | Purpose |
|------|------|------|---------|
| `relayer-v2.polymarket.com` | `POST /submit` | Builder HMAC | WALLET-CREATE, WALLET batch |
| `relayer-v2.polymarket.com` | `GET /nonce` | Builder HMAC | Current WALLET nonce |
| `relayer-v2.polymarket.com` | `GET /transaction` | Builder HMAC | Poll transaction status |
| `relayer-v2.polymarket.com` | `GET /deployed` | Builder HMAC | Check if wallet deployed |

### 5.2 Transaction Types

| Type | Gated By | Purpose |
|------|----------|---------|
| `WALLET-CREATE` | Factory `deploy()` role-gate (relayer only) | Deploy deposit wallet |
| `WALLET` batch | EOA signature validation (internal to `proxy()`) | Execute calls from wallet |
| `SAFE` deploy | Relayer | Deploy Gnosis Safe |
| `PROXY` deploy | Relayer | Deploy POLY_PROXY |

### 5.3 Gas Sponsorship

The relayer pays gas for all on-chain operations. Users need pUSD for trading amounts, not MATIC for gas:

- Wallet deployment: gas-sponsored
- Token approvals (6-call batch): gas-sponsored
- CTF splits/merges/redemptions: gas-sponsored
- Transfers: gas-sponsored

---

## 6. Headless Automation Boundary

### 6.1 What Is Automated

1. **CLOB L2 credential issuance** — `polygolem builder auto`
2. **V2 relayer key issuance** — `polygolem auth headless-onboard`
3. **Builder fee key issuance** — `polygolem clob create-builder-fee-key`

### 6.2 Wallet Lifecycle Automation

1. Wallet address prediction (`derive`) — local CREATE2 computation
2. Wallet deployment (`deploy`) — POST to relayer with builder creds
3. Token approvals (`approve`) — build 6-call batch, EOA signs, relayer submits
4. Wallet funding (`fund`) — ERC-20 transfer EOA → deposit wallet
5. Order creation → signing → submission — full CLOB flow
6. Balance checks, orderbook reads — all read-only CLOB/Gamma endpoints
7. Entire onboarding: `deploy → approve → fund` as one command

### 6.3 The polygolem Flow

```bash
# Step 1: One-time manual — copy builder credentials from polymarket.com/settings?tab=builder
# Step 2: Everything else automated
POLYMARKET_PRIVATE_KEY="0x..." \
POLYMARKET_BUILDER_API_KEY="..." \
POLYMARKET_BUILDER_SECRET="..." \
POLYMARKET_BUILDER_PASSPHRASE="..." \
  polygolem deposit-wallet onboard --fund-amount 0.71 --json
```

---

## 7. Alternate Paths Explored (All Dead Ends)

| Path | Investigated | Verdict |
|------|-------------|---------|
| Direct EOA call to `deploy()` | Live on-chain test | ❌ Reverts (role-gated) |
| Old ProxyFactory `proxy()` | Source code review | ❌ Creates type 1 wallets, not type 3 |
| Self-deploy on implementation | Polygonscan source check | ❌ No such function; implementation unverified |
| Separate permissionless factory | Full contract search | ❌ Only one DepositWalletFactory exists |
| Meta-transaction relay (OpenGSN) | Architecture analysis | ❌ No third-party relay holds operator role |
| Programmatic builder creds | Endpoint probing | ❌ `POST /auth/api-key` at relayer-v2 → 404 |
| Bridge API | Docs review | ❌ Only handles asset transfers |

---

## 8. Sources

- [Polymarket Docs — Deposit Wallets](https://docs.polymarket.com/trading/deposit-wallets)
- [Polymarket Docs — Builder Program](https://docs.polymarket.com/builders/overview)
- [Polymarket Docs — Contracts](https://docs.polymarket.com/resources/contracts)
- [Polymarket Docs — Authentication](https://docs.polymarket.com/api-reference/authentication)
- [EIP-7739 — TypedDataSign](https://eips.ethereum.org/EIPS/eip-7739)
- [EIP-1271 — Standard Signature Validation](https://eips.ethereum.org/EIPS/eip-1271)
- [Polymarket GitHub — proxy-factories](https://github.com/Polymarket/proxy-factories)
- [Polymarket GitHub — ctf-exchange](https://github.com/Polymarket/ctf-exchange)
- [Polymarket GitHub — builder-relayer-client](https://github.com/Polymarket/builder-relayer-client)
- [Polymarket GitHub — builder-signing-sdk](https://github.com/Polymarket/builder-signing-sdk)
- Polygonscan: `0x00000000000Fb5C9ADea0298D729A0CB3823Cc07` (DepositWalletFactory)
- Polygonscan: `0xb6F9C7E68A38c21BeDfD873bC5a378236f7ba987` (DepositWalletFactory — alternate address)
- Polygonscan: `0x21999a074344610057c9b2B362332388a44502D4` (example deposit wallet)

---

*This document is the definitive reference for Polymarket's deposit wallet architecture and the deployment pipeline. It reflects live on-chain reality as of 2026-05-07.*
