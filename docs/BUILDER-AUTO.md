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

## Why the old doc was wrong

`docs/BUILDER-CREDENTIAL-ISSUANCE.md` asserted that builder creds are gated behind a Polymarket session cookie and that `/auth/api-key` only issues "CLOB L2" creds (a separate type from "Builder API Keys"). Both claims are false:

- There is one HMAC cred type. The same triple authenticates against `clob.polymarket.com` (L2 endpoints) and `relayer-v2.polymarket.com` (deposit-wallet/relay endpoints).
- The "+ Create New" button under Settings → Builder Codes calls the same endpoint; the browser's wallet popup is signing the same ClobAuth payload polygolem now signs locally.

The old doc should be deleted or replaced with a pointer to this one.
