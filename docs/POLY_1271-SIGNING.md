# POLY_1271 Signing Chain — sigtype 3 Full Flow

> **Status:** Live-verified against production CLOB V2
> **Last updated:** 2026-05-08
> **Companion:** [DEPOSIT-WALLET-DEPLOYMENT.md](./DEPOSIT-WALLET-DEPLOYMENT.md), [CONTRACTS.md](./CONTRACTS.md)

---

## The Full Chain

For sigtype 3 (POLY_1271 / deposit wallet) to work end-to-end, four conditions must be satisfied:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SIGTYPE 3 — FULL CHAIN                              │
│                                                                         │
│  Step 1: L2 Key Bound to Deposit Wallet                                │
│  ┌──────────────────────────────────────────────────────────────┐      │
│  │ POST /auth/api-key                                           │      │
│  │   POLY_ADDRESS = depositWallet (not EOA)                     │      │
│  │   POLY_SIGNATURE = raw 65-byte ECDSA from EOA                │      │
│  │   → L2 key is now "bound" to the deposit wallet address       │      │
│  └──────────────────────────────────────────────────────────────┘      │
│                              │                                          │
│                              ▼                                          │
│  Step 2: CLOB HTTP Gate Passes                                         │
│  ┌──────────────────────────────────────────────────────────────┐      │
│  │ POST /order (L2 HMAC headers)                                │      │
│  │   POLY_ADDRESS = depositWallet                                │      │
│  │   → CLOB checks: signer == address-of-API-KEY ✓               │      │
│  └──────────────────────────────────────────────────────────────┘      │
│                              │                                          │
│                              ▼                                          │
│  Step 3: Order Struct Correct                                          │
│  ┌──────────────────────────────────────────────────────────────┐      │
│  │ signedOrderPayload {                                          │      │
│  │   maker  = depositWallet                                      │      │
│  │   signer = depositWallet   ← must equal maker for sigtype 3   │      │
│  │   signatureType = 3                                          │      │
│  │   signature = ERC-7739 wrapped (636 hex chars)                │      │
│  │ }                                                             │      │
│  └──────────────────────────────────────────────────────────────┘      │
│                              │                                          │
│                              ▼                                          │
│  Step 4: On-Chain ERC-1271 Validates                                   │
│  ┌──────────────────────────────────────────────────────────────┐      │
│  │ CTF Exchange V2 calls:                                        │      │
│  │   depositWallet.isValidSignature(orderHash, wrappedSig)       │      │
│  │   → wallet validates EOA signature via ERC-1271 ✓             │      │
│  └──────────────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────────┘
```

## Step 1 — L2 Key Binding

The L2 API key must be bound to the **deposit wallet address**, not the EOA. This is done at key creation time via `POST /auth/api-key`.

```go
// internal/auth/l1.go — BuildL1HeadersForAddress
// ownerAddress overrides POLY_ADDRESS to the deposit wallet
headers, err := auth.BuildL1HeadersForAddress(
    privateKeyHex,  // EOA private key (signs the ClobAuth)
    chainID,
    timestamp,
    nonce,
    depositWallet,  // ← bind the L2 key to this address
)
// headers["POLY_ADDRESS"] = depositWallet (not EOA)
// headers["POLY_SIGNATURE"] = raw 65-byte ECDSA from EOA
```

**The ClobAuth signature is a raw 65-byte ECDSA from the EOA.** It is NOT ERC-7739 wrapped. The order signing path uses ERC-7739 wrapping, but the L1 auth path uses raw ECDSA.

From the [official Polymarket docs](https://docs.polymarket.com/trading/deposit-wallets):

> "The owner or session signer signs a nested TypedDataSign payload under the correct CTF Exchange V2 domain."

This applies to ORDER signing, not L1 auth. L1 auth uses the standard ClobAuth EIP-712 with raw ECDSA.

## Step 2 — CLOB HTTP Gate

When placing orders, CLOB checks that `POLY_ADDRESS` in the L2 HMAC headers matches the address the API key is bound to:

```
CLOB HTTP gate: signer (in headers) == address-of-API-KEY (from L2 key binding)
```

Since the L2 key is bound to the deposit wallet, `POLY_ADDRESS` in L2 headers must be the deposit wallet address.

## Step 3 — Order Struct

The signed order payload must have:

```go
order.Maker  = depositWallet  // holds the funds
order.Signer = depositWallet  // must equal maker for sigtype 3
order.SignatureType = 3       // POLY_1271
order.Signature = "0x..."     // ERC-7739 wrapped, 636 hex chars
```

**The order signature IS ERC-7739 wrapped.** It uses the nested TypedDataSign structure:

```
innerSig(65) || appDomainSep(32) || contents(32) || contentsType(186) || uint16BE(186)
= 317 bytes = 634 hex chars + "0x" = 636 chars total
```

Where:
- `innerSig` = EOA's ECDSA signature over the final hash
- `appDomainSep` = CTF Exchange V2 domain separator
- `contents` = hashStruct(Order)
- `contentsType` = V2 Order type string (186 bytes)

## Step 4 — On-Chain Validation

The CTF Exchange V2 calls `isValidSignature()` on the deposit wallet:

```solidity
// CTFExchangeV2._verifyPoly1271Signature()
bool valid = IDepositWallet(signer).isValidSignature(hash, signature);
```

The deposit wallet:
1. Unwraps the ERC-7739 envelope
2. Reconstructs the TypedDataSign hash
3. Verifies the EOA's ECDSA signature against it
4. Returns `0x1626ba7e` (ERC-1271 magic value) on success

## Key Distinction: L1 Auth vs Order Signing

| Aspect | L1 Auth (ClobAuth) | Order Signing (POLY_1271) |
|--------|-------------------|--------------------------|
| Signature type | Raw 65-byte ECDSA | ERC-7739 wrapped (636 chars) |
| EIP-712 domain | `ClobAuthDomain` v1 | `Polymarket CTF Exchange` v2 |
| Signer | EOA | EOA |
| `POLY_ADDRESS` | Deposit wallet (for key binding) | Deposit wallet (L2 header) |
| Purpose | Create/bind L2 API key | Authorize trade |

## Polygolem Implementation

| Component | File | Purpose |
|-----------|------|---------|
| L1 key binding | `internal/auth/l1.go::BuildL1HeadersForAddress` | Bind L2 key to deposit wallet |
| Order signing | `internal/clob/orders.go::buildSignedOrderPayload` | Build POLY_1271 order with correct maker/signer |
| ERC-7739 wrap | `internal/clob/orders.go::wrapPOLY1271Signature` | Wrap EOA sig in TypedDataSign envelope |
| ClobAuth | `internal/auth/eip712.go::SignClobAuth` | Raw ECDSA ClobAuth signing |

## Verification Checklist

- [ ] `POLY_ADDRESS` in `/auth/api-key` headers = deposit wallet (for key binding)
- [ ] `POLY_ADDRESS` in L2 order headers = deposit wallet
- [ ] Order `maker` = order `signer` = deposit wallet
- [ ] Order `signatureType` = 3
- [ ] Order `signature` is ERC-7739 wrapped (636 hex chars)
- [ ] L1 ClobAuth signature is raw 65-byte ECDSA (NOT wrapped)
- [ ] Deposit wallet is deployed (relayer `/deployed` returns true)
- [ ] Deposit wallet has approvals (6 contracts approved)

## Related

- [Deposit Wallets (Polymarket docs)](https://docs.polymarket.com/trading/deposit-wallets)
- [CLOB Authentication](https://docs.polymarket.com/developers/CLOB/authentication)
- [EIP-7739 — TypedDataSign](https://eips.ethereum.org/EIPS/eip-7739)
- [EIP-1271 — Standard Signature Validation](https://eips.ethereum.org/EIPS/eip-1271)
