# Builder API Key — Headless Issuance

> **Status:** Empirically verified 2026-05-07 against production.
> **Supersedes:** `docs/BUILDER-CREDENTIAL-ISSUANCE.md` (which claimed curl-only issuance was impossible — that conclusion was wrong).

## TL;DR

A single off-chain ECDSA signature plus one HTTP POST is the entire onboarding flow. No browser, no wallet UI, no on-chain transaction.

```
polygolem builder auto
```

reads `POLYMARKET_PRIVATE_KEY`, signs the canonical ClobAuth EIP-712 message, posts it to `https://clob.polymarket.com/auth/api-key`, validates the returned HMAC creds against the relayer, and persists them to `../go-bot/.env.builder` (mode `0600`).

## Empirical proof

Brand-new EOA `0xA02DBaa282D42d9A5496B6643373D8Db96eFEa64`, generated locally from OS entropy, never touched polymarket.com prior to the call:

```
$ POLYMARKET_PRIVATE_KEY=0x… polygolem builder auto
Signing ClobAuth and creating builder API key via https://clob.polymarket.com...
✓ Received creds (key=64f6a5fb-3ec0-569b-5329-3c8483853f19)
✓ HMAC-signed test request to relayer succeeded
✓ Wrote credentials to /tmp/throwaway-env.builder (mode 0600)
```

The same call ran twice returned the same key — `/auth/api-key` is idempotent per EOA.

## Full onboarding sequence

The minimum the user provides end-to-end is **(1) an EOA private key** and **(2) one USDC/pUSD transfer**. Everything else — account, profile, builder code, HMAC creds, deposit-wallet deploy, V2 approvals — happens via signatures the app generates locally and HTTP calls the app makes on the user's behalf. The deploy and approval transactions are paid by the Polymarket relayer; only the funding transfer comes from the user.

```mermaid
sequenceDiagram
    autonumber
    participant U as User
    participant A as App<br/>(polygolem / Arenaton)
    participant C as clob.polymarket.com
    participant R as relayer-v2.polymarket.com
    participant P as Polygon chain

    Note over U,A: Step 1 — auth (off-chain, no gas)
    U->>A: provide EOA private key<br/>(or app generates fresh)
    A->>A: sign ClobAuth EIP-712 locally
    A->>C: POST /auth/api-key<br/>(POLY_ADDRESS, POLY_SIGNATURE,<br/>POLY_TIMESTAMP, POLY_NONCE)
    C->>C: lazy-create account +<br/>builder profile + bytes32 code
    C-->>A: { apiKey, secret, passphrase }
    A->>R: GET /nonce (HMAC-signed)
    R-->>A: 200 → creds verified

    Note over A,P: Step 2 — deploy deposit wallet (relayer pays gas)
    A->>A: sign DepositWallet Batch EIP-712
    A->>R: POST /relay-payload<br/>(factory.deploy via proxy)
    R->>P: submit tx (relayer pays)
    P-->>R: receipt
    A->>R: poll GET /deployed
    R-->>A: { wallet: 0x… }

    Note over U,P: Step 3 — fund (only on-chain action by user)
    U->>P: transfer USDC.e or pUSD<br/>to deposit wallet

    Note over A,P: Step 4 — approvals (relayer pays gas)
    A->>A: sign 6× Batch (pUSD + CTF<br/>→ 3× V2 spenders)
    A->>R: POST /relay-payload (factory.proxy)
    R->>P: submit tx (relayer pays)
    P-->>R: receipt

    Note over A,C: Step 5 — trade
    A->>A: sign V2 Order EIP-712<br/>(POLY_1271 sigtype 3,<br/>builder code attached)
    A->>C: POST /order
    C-->>A: order accepted
```

**User-facing total cost:** one private key + one funding tx. Zero browser interaction.

## Wire format

### Request

`POST https://clob.polymarket.com/auth/api-key`

Headers (all required):

| Header | Value |
| --- | --- |
| `POLY_ADDRESS` | EOA address, hex with `0x` prefix |
| `POLY_TIMESTAMP` | Unix seconds, decimal string |
| `POLY_NONCE` | Decimal string (use `0`) |
| `POLY_SIGNATURE` | Hex `0x…` 65-byte ECDSA over the EIP-712 hash below |

EIP-712 typed data:

```
domain:  { name: "ClobAuthDomain", version: "1", chainId: 137 }
type:    ClobAuth(address address, string timestamp, uint256 nonce, string message)
value:   { address:   <EOA>,
           timestamp: <unix_seconds_as_string>,
           nonce:     0,
           message:   "This message attests that I control the given wallet" }
```

Signed digest = `keccak256(0x1901 || keccak256(domainSep) || keccak256(structHash))` — standard EIP-712.

### Response

```json
{ "apiKey": "<uuid-shape>", "secret": "<base64>", "passphrase": "<random>" }
```

Some legacy responses use `api_key`, `passPhrase`, or `pass_phrase` — `internal/clob/client.go` accepts all variants.

The `apiKey` value is **UUID-shaped (8-4-4-4-12 hex)** but does **not** conform to RFC 4122 version constraints. Observed values include version-nibbles `1` (existing accounts) and `e` or `5` (fresh accounts). Validators that require a `4` are wrong.

The endpoint also has a `GET /auth/derive-api-key` companion that returns the same triple deterministically when one already exists; `CreateOrDeriveAPIKey` in `internal/clob/client.go` falls back to it on conflict.

## What the backend does on first contact

A new EOA's first signed `POST /auth/api-key` lazy-creates:

1. The user account
2. The builder profile row (visible at `polymarket.com/settings?tab=builder` if you log in with that key)
3. The bytes32 `builderCode` (used in V2 `Order.builder` for fee attribution)
4. The HMAC triple returned in the response

The `polymarket.com/settings?tab=builder` UI exposes state; it does not create it. There is no separate "register as builder" endpoint or transaction.

## Validation

`polygolem builder auto` validates by HMAC-signing a `GET /nonce` against `https://relayer-v2.polymarket.com` using the freshly-issued creds. Server-side HMAC verification doubles as a profile-existence check — only a registered builder address gets a non-error response. Verified for both pre-existing and brand-new EOAs.

## Implementation pointers

| Concern | Location |
| --- | --- |
| ClobAuth EIP-712 typed data | `internal/auth/eip712.go` |
| L1 header builder | `internal/auth/l1.go` |
| `CreateOrDeriveAPIKey` HTTP client | `internal/clob/client.go:65` |
| `polygolem builder auto` CLI | `internal/cli/builder.go` (`newBuilderAutoCommand`) |
| Persisted env file shape | `internal/cli/builder.go:persistBuilderCredentials` |

## Operational notes

- Idempotent: re-running `builder auto` for the same EOA returns the same creds. Use `--force` to overwrite the local env file with re-fetched values.
- `--no-validate` skips the relayer round-trip — useful for offline runs but loses the existence check.
- The throwaway-key proof above generated and discarded the EOA in seconds; the orphan account on Polymarket is harmless.

## End-to-End Cost to the User

| Item | Cost | Who Pays |
|------|------|---------|
| Generate EOA key | Free | N/A |
| Builder profile + HMAC creds | Free | N/A |
| Wallet deploy | Free (gas sponsored) | Polymarket relayer |
| 6 contract approvals | Free (gas sponsored) | Polymarket relayer |
| Place orders | Free (gas sponsored) | Polymarket relayer |
| **Fund wallet (one tx)** | **~$0.01 POL** | **User** |
| pUSD to trade with | Whatever you deposit | User (your money) |

**Total hard cost to user:** ~$0.01 in POL gas for one transfer. Everything else is free and automated.

---

## Flow A — Wallet Already Deployed (Returning User)

When the user returns, skip deploy and approve:

```
┌─────────────────────────────────────────────────────────────────┐
│            RETURNING USER — WALLET ALREADY DEPLOYED              │
│                                                                 │
│  1. App has EOA private key (or MetaMask connected)            │
│  2. Derive wallet address (local CREATE2 — instant)             │
│  3. GET /deployed?address=EOA → { deployed: true }             │
│                                                                 │
│     ✅ Wallet exists — skip deploy                              │
│     ✅ Approvals exist — skip approve                           │
│                                                                 │
│  4. Check balance (CLOB /balance-allowance)                     │
│                                                                 │
│     ┌──────────────────┬──────────────────────┐                 │
│     │  Has pUSD        │  Needs funding        │                 │
│     │  → Trade         │  → Fund first         │                 │
│     └──────────────────┴──────────────────────┘                 │
│                                                                 │
│  5. Trade — build order → sign EIP-712 → POST to CLOB          │
└─────────────────────────────────────────────────────────────────┘
```

```bash
polygolem deposit-wallet status
# {
#   "eoa": "0x...",
#   "deposit_wallet": "0x...",
#   "deployed": true,
#   "approvals": 6,
#   "pUSD_balance": "5.00",
#   "ready_to_trade": true
# }
```

---

## Funding Flow — POL → pUSD → Deposit Wallet

### The POL Requirement

The user needs ~0.01 POL for exactly ONE transaction — the ERC-20 pUSD transfer from EOA to deposit wallet. Every other operation is gas-sponsored.

### Path 1: Exchange → POL → pUSD (Recommended)

```
┌─────────────────────────────────────────────────────────────────┐
│                    FUNDING FLOW — FRESH EOA                      │
│                                                                 │
│  Step 1: Get POL (for one transfer only, ~$0.01)                │
│    Buy POL on any exchange → withdraw to EOA on Polygon          │
│                                                                 │
│  Step 2: Get USDC on Polygon                                    │
│    Option A: Buy USDC on exchange → withdraw to EOA on Polygon  │
│    Option B: Bridge from Ethereum/Arbitrum/Base via Across.to   │
│    Option C: Use Polymarket Bridge API (auto-converts to pUSD)  │
│                                                                 │
│  Step 3: Convert USDC → pUSD                                    │
│    Polymarket bridge does this automatically                    │
│    Or call CollateralOnramp.deposit() manually                  │
│                                                                 │
│  Step 4: Transfer pUSD EOA → Deposit Wallet                     │
│    polygolem deposit-wallet fund --amount X                     │
│    (This is the one tx you pay gas for — ~$0.01 POL)            │
└─────────────────────────────────────────────────────────────────┘
```

### Path 2: Bridge API (Programmatic)

```bash
# Get deposit addresses for your EOA
curl -X POST https://bridge.polymarket.com/deposit \
  -H "Content-Type: application/json" \
  -d '{"address": "0xYOUR_EOA_ADDRESS"}'
# Response: { "evm": "0x...", "svm": "...", "btc": "..." }

# Bridge auto-converts to pUSD and credits your EOA
polygolem bridge deposit <wallet-address>

# Then transfer to deposit wallet
polygolem deposit-wallet fund --amount X
```

---

## The Deposit Wallet — Beyond Polymarket

The deposit wallet is an ERC-1967 proxy smart contract on Polygon. It's NOT Polymarket-specific — it's a general-purpose smart contract wallet implementing ERC-1271.

### What the Wallet CAN Do

| Capability | Standard | Polymarket-Specific? |
|-----------|----------|---------------------|
| Hold pUSD | ERC-20 | No — any ERC-20 works |
| Hold USDC, USDC.e | ERC-20 | No |
| Hold POL (native) | Native balance | No |
| Hold outcome tokens | ERC-1155 (CTF) | Yes (CTF-specific) |
| Sign typed data | ERC-1271 (EIP-1271) | No — any protocol can validate |
| Execute batched calls | factory.proxy(batch[], sig[]) | No — general-purpose proxy |
| Approve token spenders | ERC-20 approve via batch | No — any spender |
| Interact with ANY contract | Via proxy batch calls | No — fully programmable |

### ERC-1271 Interoperability

The wallet implements `isValidSignature(bytes32 hash, bytes signature)` per EIP-1271. Any protocol that accepts ERC-1271 signatures can validate EOA-signed messages through this wallet. The wallet itself has no private key — the EOA signs, the wallet validates.

### Potential Use Cases Beyond Polymarket

| Use Case | How |
|----------|-----|
| Other prediction markets | Settle in ERC-20 tokens on Polygon |
| DeFi (Aave, Uniswap) | Approve + interact via proxy batch |
| NFT marketplaces | Hold/trade ERC-721/ERC-1155 |
| Multi-market aggregators | One wallet, multiple platforms, unified balance |
| DAO voting | ERC-1271 signatures for snapshot |
| Account abstraction | Extendable via factory upgrade |

### Wallet Interface

```solidity
// Core capabilities of the deposit wallet
interface IDepositWallet {
    // ERC-1271 — any protocol calls this to validate EOA signatures
    function isValidSignature(bytes32 hash, bytes calldata signature) 
        external view returns (bytes4 magicValue);
    // Token receiver (ERC-1155 compatible)
    function onERC1155Received(address, address, uint256, uint256, bytes calldata) 
        external returns (bytes4);
    // The wallet owner
    function owner() external view returns (address);
}

// Factory — deploy + execute
interface IDepositWalletFactory {
    function deploy(address[] owners, bytes32[] ids) external;
    function proxy(Batch[] batches, bytes[] signatures) external;  // UNGATED
    function predictWalletAddress(address impl, bytes32 id) external view returns (address);
}
```

**Key insight:** The deposit wallet is a standard ERC-1271 smart contract wallet. Polymarket uses it as the funder address. Any protocol that accepts smart contract wallet signatures can use it.

---

## Polygolem API Surface

| Command | What it does | Gas Cost |
|---------|-------------|----------|
| `polygolem builder auto` | Create builder profile + HMAC creds (ClobAuth EIP-712) | Free |
| `polygolem deposit-wallet derive` | Predict CREATE2 wallet address (local) | Free |
| `polygolem deposit-wallet status` | Check deployed?, approvals?, balance? | Free |
| `polygolem deposit-wallet deploy --wait` | WALLET-CREATE via relayer | Sponsored |
| `polygolem deposit-wallet approve --submit` | 6-call WALLET batch via relayer | Sponsored |
| `polygolem deposit-wallet fund --amount X` | ERC-20 transfer EOA → deposit wallet | ~$0.01 POL |
| `polygolem deposit-wallet onboard --fund-amount X` | deploy + approve + fund (all-in-one) | ~$0.01 POL |
| `polygolem bridge deposit <addr>` | Get EVM/Solana/BTC deposit addresses | Free |
| `polygolem clob create-order ...` | Place order (POLY_1271 signed) | Sponsored |

---

## Why the old doc was wrong

`docs/BUILDER-CREDENTIAL-ISSUANCE.md` asserted that builder creds are gated behind a Polymarket session cookie and that `/auth/api-key` only issues "CLOB L2" creds (a separate type from "Builder API Keys"). Both claims are false:

- There is one HMAC cred type. The same triple authenticates against `clob.polymarket.com` (L2 endpoints) and `relayer-v2.polymarket.com` (deposit-wallet/relay endpoints).
- The "+ Create New" button under Settings → Builder Codes calls the same endpoint; the browser's wallet popup is signing the same ClobAuth payload polygolem now signs locally.

**BUILDER-AUTO.md is the authoritative document. BUILD-CREDENTIAL-ISSUANCE.md is superseded and should be treated as historical investigation.**
