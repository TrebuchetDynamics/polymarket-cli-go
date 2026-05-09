# Polymarket Smart Contracts — Deposit Wallet Research

> **Date:** 2026-05-07
> **Status:** Complete — investigation confirms all-on-chain deployment is impossible
> **Last verified:** Live on-chain against Polygon mainnet (chainID 137) and Polygonscan verified source, 2026-05-09

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
| **CollateralOnramp** | `0x93070a847efEf7F70739046A929D47a521F5B8ee` | USDC/USDC.e to pUSD onramp |
| **CollateralOfframp** | `0x2957922Eb93258b93368531d39fAcCA3B4dC5854` | pUSD to USDC/USDC.e offramp |
| **PermissionedRamp** | `0xebC2459Ec962869ca4c0bd1E06368272732BCb08` | Permissioned collateral ramp |
| **CtfCollateralAdapter** | `0xAdA100Db00Ca00073811820692005400218FcE1f` | V2 split/merge/redeem adapter for standard CTF markets |
| **NegRiskCtfCollateralAdapter** | `0xadA2005600Dec949baf300f4C6120000bDB6eAab` | V2 split/merge/redeem adapter for negative-risk markets |

### 1.2 V2 Collateral Layer

pUSD is Polymarket's V2 collateral wrapper on Polygon. The proxy address is
`0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB`; local V2 source names it
`Polymarket USD`, symbol `pUSD`, with 6 decimals.

Deposit wallets must not call `ConditionalTokens.redeemPositions` directly for
V2 pUSD-native flows. The supported deposit-wallet path is an EIP-712 WALLET
batch submitted by the relayer through the deposit-wallet factory, with the
wallet call targeting a collateral adapter. Standard binary-market split,
merge, and redeem calls route through `CtfCollateralAdapter`; negative-risk
calls route through `NegRiskCtfCollateralAdapter`.

For redeem, the adapter exposes the same function shape as the legacy CTF
interface:

```solidity
redeemPositions(address collateralToken, bytes32 parentCollectionId, bytes32 conditionId, uint256[] indexSets)
```

The V2 adapter uses `conditionId`, reads the deposit wallet's current YES/NO
CTF balances, pulls those ERC-1155 positions into the adapter, performs the
underlying CTF redemption, wraps the USDC.e proceeds back into pUSD, and sends
pUSD to `msg.sender` (the deposit wallet). The caller-supplied collateral
address, parent collection, and `indexSets` are accepted for ABI compatibility
but ignored by the adapter.

This creates a separate approval requirement from trading:

| Approval Set | Spenders | Purpose |
|---|---|---|
| Trading approvals | `CTFExchangeV2`, `NegRiskExchangeV2`, `NegRiskAdapterV2` | CLOB order matching and settlement |
| Adapter approvals | `CtfCollateralAdapter`, `NegRiskCtfCollateralAdapter` | V2 split, merge, and redeem |

Today's trading approval batch is six calls: pUSD `approve` plus CTF
`setApprovalForAll` for each trading spender. V2 adapter readiness requires an
additional four-call WALLET batch: pUSD `approve` plus CTF
`setApprovalForAll` for each collateral adapter. Redeem itself needs the CTF
approval leg; the pUSD approval leg is included so the same one-time adapter
batch also covers split flows.

SAFE/PROXY relayer examples are separate wallet-type flows. They do not create
a deposit-wallet shortcut around the V2 adapter path, and raw
`ConditionalTokens.redeemPositions` must not be used as a deposit-wallet
fallback. If the relayer rejects adapter calls, first verify the local adapter
addresses against Polymarket's current contracts reference; stale constants are
a known failure mode.

### 1.3 Which Wallet Type Wins

**New API users (post-May 2026):** Deposit wallet (type 3 / POLY_1271) only. EOA is blocked by CLOB V2.
**Grandfathered users:** Proxy (type 1) or Safe (type 2) still work.

### 1.4 Deployment Status Source of Truth

There are two different deployment signals:

| Signal | Source | Meaning |
|---|---|---|
| Relayer deployed flag | `GET /deployed?address=<owner>` | Relayer/indexer view used by Polymarket's wallet lifecycle API |
| Contract bytecode | Polygon `eth_getCode(<depositWallet>)` | Chain truth: whether the deposit wallet has deployed code |

For trading safety, **on-chain bytecode wins**. The CTF Exchange V2
POLY_1271 path checks `maker.code.length > 0` before validating the ERC-1271
signature, so a deposit wallet with non-empty bytecode is deployed for
order-signing purposes even if the relayer `/deployed` endpoint returns a
false negative.

2026-05-09 live account evidence:

```text
owner:         0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C
depositWallet: 0x21999a074344610057c9b2B362332388a44502D4
relayer:       deployed=false
polygon code:  0x363d3d373d3d363d7f360894a13ba1a3210667c828492db...
status:        deployed for POLY_1271
```

Polygolem exposes this as the public SDK contract primitive:

```go
status, err := contracts.DepositWalletDeployed(ctx, depositWallet, "")
if err != nil {
    return err
}
if status.Deployed {
    // Safe to skip WALLET-CREATE; continue approvals/funding/readiness.
}
```

`polygolem deposit-wallet status` reports both fields:

```json
{
  "deployed": true,
  "deploymentStatusSource": "polygon_code",
  "relayerDeployed": false,
  "onchainCodeDeployed": true
}
```

`polygolem deposit-wallet deploy --wait` also checks bytecode before
submitting `WALLET-CREATE`; if code already exists, it returns
`state=already_deployed` and skips the relayer mutation.

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
proxy(Batch[] _batches, bytes[] _signatures)           ← OPERATOR-GATED + signature-validated
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

### 2.5 proxy(Batch[], bytes[]) is Operator-Gated

Polygonscan verified source for `DepositWalletFactory` shows both deployment
and batch proxying are restricted to the factory operator role:

```solidity
function deploy(address[] calldata _owners, bytes32[] calldata _ids) external onlyOperator
function proxy(Batch[] calldata _batches, bytes[] calldata _signatures) external onlyOperator
```

The `proxy()` path still validates each wallet owner's EOA signature, but the
transaction submitter must also be a factory operator. A direct owner EOA call
to `proxy()` is therefore not a valid fallback path.

2026-05-09 Polygon RPC proof:

```bash
cast sig 'OnlyOperator()'
# 0x27e1f1e5

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  --from 0x000000000000000000000000000000000000dEaD \
  0x00000000000Fb5C9ADea0298D729A0CB3823Cc07 \
  'proxy((address,uint256,uint256,(address,uint256,bytes)[])[],bytes[])' \
  '[]' '[]'
# execution reverted, data: "0x27e1f1e5"
```

**Operational conclusion:** post-deployment wallet mutations still go through
Polymarket's relayer/operator surface. The wallet owner controls authorization
with EOA-signed batches, but the factory does not expose a permissionless
submission route.

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
| `relayer-v2.polymarket.com` | `GET /deployed` | Builder HMAC | Relayer/indexer deployment view; use Polygon bytecode as fallback/source of truth |

### 5.2 Transaction Types

| Type | Gated By | Purpose |
|------|----------|---------|
| `WALLET-CREATE` | Factory `deploy()` role-gate (relayer only) | Deploy deposit wallet |
| `WALLET` batch | Factory `proxy()` operator gate plus EOA signature validation | Execute calls from wallet |
| `SAFE` deploy | Relayer | Deploy Gnosis Safe |
| `PROXY` deploy | Relayer | Deploy POLY_PROXY |

### 5.3 Gas Sponsorship

The relayer pays gas for all on-chain operations. Users need pUSD for trading amounts, not MATIC for gas:

- Wallet deployment: gas-sponsored
- Trading approval batch: gas-sponsored
- Adapter approval batch: gas-sponsored when accepted by the relayer allowlist
- CTF split/merge/redeem through V2 collateral adapters: gas-sponsored when accepted by the relayer allowlist
- Deposit-wallet WALLET batches: gas-sponsored
- EOA-to-wallet funding transfer: user pays Polygon gas

---

## 6. Headless Automation Boundary

### 6.1 What Is Automated

1. **CLOB L2 credential issuance** — `polygolem builder auto`
2. **V2 relayer key issuance** — `polygolem auth headless-onboard`
3. **Builder fee key issuance** — `polygolem clob create-builder-fee-key`

### 6.2 Wallet Lifecycle Automation

1. Wallet address prediction (`derive`) — local CREATE2 computation
2. Wallet deployment status (`status`) — relayer view plus Polygon `eth_getCode`
3. Wallet deployment (`deploy`) — skip when bytecode exists; otherwise POST to relayer with builder creds
4. Token approvals (`approve`) — build the trading approval batch, EOA signs,
   relayer submits
5. Adapter approvals — submit the separate collateral-adapter batch before
   split/merge/redeem
6. Wallet funding (`fund`) — ERC-20 transfer EOA → deposit wallet
7. Order creation → signing → submission — full CLOB flow
8. Balance checks, orderbook reads — all read-only CLOB/Gamma endpoints
9. Entire onboarding: `deploy → approve → fund` as one command

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
| Direct EOA call to `proxy()` | Polygonscan verified source + live `eth_call` | ❌ Reverts with `OnlyOperator()` unless caller has operator role |
| Direct EOA call to wallet `execute()` | Verified call context | ❌ Wallet execution is factory-mediated |
| Raw `ConditionalTokens.redeemPositions` fallback | V2 adapter source review + official relayer client wallet-type split | ❌ Not a deposit-wallet fallback; SAFE/PROXY examples do not apply |
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
- Polygonscan verified source: `0x00000000000Fb5C9ADea0298D729A0CB3823Cc07` (DepositWalletFactory; `deploy()` and `proxy()` are `onlyOperator`)
- Polygonscan: `0xb6F9C7E68A38c21BeDfD873bC5a378236f7ba987` (DepositWalletFactory — alternate address)
- Polygonscan: `0x21999a074344610057c9b2B362332388a44502D4` (example deposit wallet)

---

*This document is the definitive reference for Polymarket's deposit wallet architecture and the deployment pipeline. It reflects live on-chain reality as of 2026-05-09.*
