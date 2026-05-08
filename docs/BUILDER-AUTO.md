# Builder API Key — Headless Issuance

> **Status:** Empirically verified 2026-05-07 against production.
> **Supersedes:** `docs/BUILDER-CREDENTIAL-ISSUANCE.md` (which claimed curl-only issuance was impossible — that conclusion was wrong).

## Three Credential Types — No Longer Conflated

Polygolem interacts with three distinct Polymarket credential systems. They are not the same thing.

| Credential | Endpoint | Auth Required | Headless? | Used For |
|-----------|----------|--------------|-----------|----------|
| **CLOB L2 Trading Key** | `POST /auth/api-key` | L1 EIP-712 (EOA signs) | ✅ Yes — `builder auto` | Order placement, balance queries, trade history |
| **Builder Fee Key** | `POST /auth/builder-api-key` | L2 HMAC (existing L2 creds) | ✅ Yes — `CreateBuilderFeeKey` | V2 order `builder` field attribution |
| **Relayer API Key** | `POST /relayer/api/auth` | SIWE-backed Gamma session | ✅ Yes — `auth headless-onboard` | WALLET-CREATE, WALLET batch, relayer operations |

### Builder Fee Key (headless, new)

`POST /auth/builder-api-key` at `clob.polymarket.com` takes L2 HMAC headers (from existing CLOB L2 creds) and returns a builder fee key. This key goes in the V2 order `builder` field for attribution. Fully headless — no cookie, no browser.

```go
// After obtaining L2 creds via builder auto or CreateOrDeriveAPIKey:
feeKey, err := client.CreateBuilderFeeKey(ctx, privateKey)
// feeKey.Key → used as builderCode in V2 order struct
```

This is distinct from the Relayer API Key, which is minted through the
SIWE-backed `auth headless-onboard` flow and used by relayer `WALLET-CREATE`
and `WALLET` batch calls.

### Relayer API Key (SIWE-backed)

The relayer's `/relayer/api/auth` endpoint requires a Gamma session cookie.
Polygolem obtains that cookie headlessly with EIP-4361 SIWE against
`gamma-api.polymarket.com/login`, then mints and persists
`RELAYER_API_KEY` + `RELAYER_API_KEY_ADDRESS` via
`polygolem auth headless-onboard`. The settings-page button remains a manual
fallback, not the primary path.

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

## Production V2 reality check (2026-05-07)

**Sigtype 3 (deposit wallet) is the only signature type accepted by `/order` in production.** The 2026-04-28 V2 cutover finished off sigtypes 0/1/2 (EOA / proxy / safe). The CLOB rejects them at the maker-address check before any balance or signing validation runs.

Verified live against `clob.polymarket.com` from a profiled, builder-registered EOA on an open, accepting-orders market (`0x0b4cc3b7…d134bee`, 5-share buy at $0.01):

```
--signature-type eoa|proxy|safe → HTTP 400 maker address not allowed, please use the deposit wallet flow
--signature-type deposit        → HTTP 400 the order signer address has to be the address of the API KEY
```

So the only working path is sigtype 3, and sigtype 3 has its own coupling: **the L2 API key has to be owned by the deposit-wallet address, not the EOA.** Calling `clob create-api-key` with `POLYMARKET_PRIVATE_KEY` mints a key owned by the EOA — that key signs valid HMAC headers, but `/order` rejects it because the order signer (the deposit wallet via ERC-1271) does not match the API-key owner (the EOA).

The well-formed end-to-end path is therefore:

1. Mint CLOB L2 trading creds via `polygolem builder auto` (`POST /auth/api-key`, headless).
2. Mint a CLOB Builder Fee Key via `polygolem clob create-builder-fee-key` (`POST /auth/builder-api-key`, headless — needs L2 creds from step 1).
3. Mint a Relayer API Key via `polygolem auth headless-onboard` (SIWE login + `POST /relayer/api/auth`).
4. Deploy the deposit wallet via `polygolem deposit-wallet deploy` (`POST relayer-v2/submit`, gated on the Relayer API Key from step 3).
5. Mint a CLOB API key **owned by the deposit wallet**, not the EOA, when an owner-scoped key is required: `polygolem clob create-api-key-for-address --owner <deposit-wallet>`.
6. Submit orders signed by the deposit wallet (sigtype 3).

> Empirically verified 2026-05-07 against profiled EOA `0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C`: registering a builder code without minting a Relayer API Key still returns `HTTP 401 invalid authorization` from `relayer-v2/submit`. The "Builder Keys: No builder API keys yet" UI label refers specifically to the Relayer API Key row.

## Full onboarding sequence

The minimum the user provides end-to-end is **(1) an EOA private key** and
**(2) one USDC/pUSD transfer**. Everything else — CLOB L2 creds, V2 relayer
keys, deposit-wallet deploy, V2 approvals, and builder-code attribution —
happens via signatures the app generates locally and HTTP calls the app makes
on the user's behalf. The deploy and approval transactions are paid by the
Polymarket relayer; only the funding transfer comes from the user.

> Earlier revisions treated the settings-page click as mandatory. The current
> CLI uses `auth headless-onboard` to mint the V2 relayer key without a
> browser. If relayer credentials are missing, `deposit-wallet deploy` returns
> `HTTP 401 invalid authorization`.

```mermaid
sequenceDiagram
    autonumber
    participant U as User
    participant A as App<br/>(polygolem / Arenaton)
    participant C as clob.polymarket.com
    participant R as relayer-v2.polymarket.com
    participant P as Polygon chain

    Note over U,A: Step 1 — L2 trading creds (off-chain, no gas)
    U->>A: provide EOA private key<br/>(or app generates fresh)
    A->>A: sign ClobAuth EIP-712 locally
    A->>C: POST /auth/api-key<br/>(POLY_ADDRESS, POLY_SIGNATURE,<br/>POLY_TIMESTAMP, POLY_NONCE)
    C-->>A: { apiKey, secret, passphrase }

    Note over A,C: Step 2 — Builder Fee Key (off-chain, no gas)
    A->>C: POST /auth/builder-api-key<br/>(L2 HMAC headers)
    C-->>A: { key, secret, passphrase }<br/>→ goes in V2 order builder field

    Note over U,A,R: Step 3 — Relayer API Key (headless SIWE)
    A->>C: SIWE login at gamma-api.polymarket.com/login
    A->>R: POST /relayer/api/auth
    R-->>A: RELAYER_API_KEY + RELAYER_API_KEY_ADDRESS

    Note over A,P: Step 4 — deploy deposit wallet (relayer pays gas)
    A->>A: sign DepositWallet Batch EIP-712
    A->>R: POST /relay-payload<br/>(factory.deploy via proxy)
    R->>P: submit tx (relayer pays)
    P-->>R: receipt
    A->>R: poll GET /deployed
    R-->>A: { wallet: 0x… }

    Note over U,P: Step 5 — fund (only on-chain action by user)
    U->>P: transfer USDC.e or pUSD<br/>to deposit wallet

    Note over A,P: Step 6 — approvals (relayer pays gas)
    A->>A: sign 6× Batch (pUSD + CTF<br/>→ 3× V2 spenders)
    A->>R: POST /relay-payload (factory.proxy)
    R->>P: submit tx (relayer pays)
    P-->>R: receipt

    Note over A,C: Step 7 — trade
    A->>A: sign V2 Order EIP-712<br/>(POLY_1271 sigtype 3,<br/>builder fee key attached)
    A->>C: POST /order
    C-->>A: order accepted
```

**User-facing total cost:** one private key + one funding tx. **Zero browser clicks** (resolved 2026-05-08; see § SIWE resolution below).

## SIWE resolution (2026-05-08)

The "browser-mediated wallet challenge" that earlier revisions of this doc treated as a real gate **was never a gate**. A focused frontend-bundle pass identified `/login/internal` as an SSR-only bootstrap function inside the bundled `@polymarket/relayer-client` SDK — not the user login flow. The actual user flow that mints the Polymarket session cookie is canonical EIP-4361 SIWE against `gamma-api.polymarket.com/login`, replicable from any HTTP client + EOA signer.

Live-validated 2026-05-08 against production with a throwaway EOA: SIWE login → cookie → V2 Relayer API Key mint → real `WALLET-CREATE` accepted by `relayer-v2/submit` (transactionHash `0xeb820d76…62e45`). Cloudflare's `__cf_bm` cookie was set but did not challenge.

Polygolem ships this flow as `polygolem auth headless-onboard`. Polydart consumers get `SIWESession` + `mintV2APIKey` + `V2APIKey.v2Headers()`. See `polydart/docs/HEADLESS-BUILDER-KEYS-INVESTIGATION.md` for the full investigation including the addendum that pinned down the SIWE flow.

## V2 auth schemes (2026-05-08)

Two relayer auth schemes coexist on V2 `relayer-v2/submit`:

1. **POLY_BUILDER_* HMAC** (legacy, V1-era). Per-request HMAC-SHA256 over timestamp+method+path+body. Implemented in `internal/auth/l2.go::BuildBuilderHeaders` and surfaced via `relayer.New(...)`.
2. **RELAYER_API_KEY plain headers** (V2). Two flat headers (`RELAYER_API_KEY`, `RELAYER_API_KEY_ADDRESS`); no HMAC, no clock skew, no body coupling. Implemented in `internal/relayer/auth_mint.go::V2APIKey` and surfaced via `relayer.NewV2(...)`.

Static analysis of `@polymarket/builder-relayer-client@0.0.9` (the canonical V2 SDK) confirms it still emits POLY_BUILDER_* headers on `/submit`, so legacy creds remain wire-valid. Polygolem prefers V2 when both `RELAYER_API_KEY` and `RELAYER_API_KEY_ADDRESS` are in env (via `relayerClientFromEnv()` in `internal/cli/deposit_wallet.go`), falling back to POLY_BUILDER_* HMAC otherwise. See report at `/tmp/poly-builder-vs-relayer-key.md` for the SDK-level evidence.

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

A new EOA's first signed `POST /auth/api-key` issues an HMAC triple
(`apiKey` / `secret` / `passphrase`). The endpoint is idempotent per EOA.

**Read access** — those creds authenticate against the relayer for read
endpoints (`GET /nonce`, `GET /deployed`). HMAC verifies on the server
side; the EOA can poll its own nonce and deployment status.

**Relayer writes — gated on Relayer API Keys.** The relayer's
`POST /submit` endpoint (used for `WALLET-CREATE`, `WALLET` batches)
returns `HTTP 401 invalid authorization` for any EOA whose triple was
issued by `/auth/api-key` rather than by the V2 relayer key issuer.
Verified 2026-05-07 before `auth headless-onboard` was implemented:

- Throwaway EOA `0xf76Ca737f9c768fc3562fbFbF8A748A4718f2a61` (no
  browser interaction): builder-auto succeeded, `/nonce` + `/deployed`
  200, `/submit` 401.
- Profiled EOA `0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C` (registered
  builder code, builder enabled, **no Builder Keys minted**):
  same result — `/submit` 401.

The settings page's "Create" button and `polygolem auth headless-onboard`
both mint relayer-write creds; `/auth/api-key` does not.

**CLOB writes — sigtype 3 only.** `clob.polymarket.com/order` accepts
sigtype 3 (deposit wallet) and nothing else; sigtypes 0/1/2 return
`maker address not allowed, please use the deposit wallet flow`.
Sigtype 3 additionally requires the order signer to match the API-key
owner: since `builder auto` mints a key owned by the EOA, sigtype-3
orders signed by the deposit wallet fail with `the order signer
address has to be the address of the API KEY` until a fresh API key is
minted from the deposit wallet itself.

So the canonical onboarding **today** is:

| Step | Mechanism | Coverage from polygolem |
| --- | --- | --- |
| Mint CLOB L2 creds (read access + signing identity) | `polygolem builder auto` | ✅ headless |
| Authenticate relayer reads (`/nonce`, `/deployed`) | same creds | ✅ headless |
| Mint CLOB Builder Fee Key (`/auth/builder-api-key`) | `polygolem clob create-builder-fee-key` | ✅ headless |
| Mint **Relayer API Key** (`/relayer/api/auth`) | `polygolem auth headless-onboard` | ✅ headless |
| Deploy deposit wallet (`WALLET-CREATE` via `/submit`) | `polygolem deposit-wallet deploy` | ✅ headless after Relayer API Key exists |
| Approve V2 spenders (6× ERC-20/ERC-1155 approvals) | `polygolem deposit-wallet approve` | ✅ headless |
| Mint a deposit-wallet-owned CLOB API key | `polygolem clob create-api-key-for-address --owner <deposit-wallet>` | ✅ headless |
| Place orders (sigtype 3) | `polygolem clob create-order` | ✅ headless after the key swap |

Earlier revisions of this doc claimed the first `/auth/api-key` POST
lazy-created the full builder profile end-to-end. That was over-stated
based on a single observation against an EOA that already had a
manually-created profile and Relayer API Keys. They also conflated the
CLOB Builder Fee Key (headless) with the Relayer API Key (now headless via SIWE)
under one "Builder API Keys" label. The corrected behaviour and split
above is what production enforces today.

## Validation

`polygolem builder auto` validates by HMAC-signing a `GET /nonce` against `https://relayer-v2.polymarket.com` using the freshly-issued creds. Server-side HMAC verification doubles as a profile-existence check — only a registered builder address gets a non-error response. Verified for both pre-existing and brand-new EOAs.

## Implementation pointers

### Public Go SDK (semver-stable)

| Concern | Public path |
| --- | --- |
| `CreateOrDeriveAPIKey` / `DeriveAPIKey` (Step 1) | `pkg/universal.Client` |
| `CreateBuilderFeeKey` / `ListBuilderFeeKeys` / `RevokeBuilderFeeKey` (Step 2) | `pkg/clob.Client`; `CreateBuilderFeeKey` is also on `pkg/universal.Client` |
| `BalanceAllowance` / `UpdateBalanceAllowance` | `pkg/universal.Client` |
| `CreateLimitOrder` / `CreateMarketOrder` (Step 7) | `pkg/universal.Client` |
| Relayer client (Steps 4 & 6) | `pkg/relayer.New` |
| `BuildApprovalCalls` (Step 6 calldata) | `pkg/relayer.BuildApprovalCalls` |
| `SignWalletBatch` (Step 6 signing) | `pkg/relayer.SignWalletBatch` |

### Internal sources

| Concern | Location |
| --- | --- |
| ClobAuth EIP-712 typed data | `internal/auth/eip712.go` |
| L1 header builder | `internal/auth/l1.go` |
| `CreateOrDeriveAPIKey` HTTP client | `internal/clob/client.go:65` |
| Relayer HTTP client | `internal/relayer/client.go` |
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
| CLOB L2 creds (`builder auto`) | Free, headless | N/A |
| CLOB Builder Fee Key (`clob create-builder-fee-key`) | Free, headless | N/A |
| **Mint Relayer API Key** (`auth headless-onboard`) | Free, headless | N/A |
| Wallet deploy | Free (gas sponsored) | Polymarket relayer |
| 6 contract approvals | Free (gas sponsored) | Polymarket relayer |
| Place orders | Free (gas sponsored) | Polymarket relayer |
| **Fund wallet (one tx)** | **~$0.01 POL** | **User** |
| pUSD to trade with | Whatever you deposit | User (your money) |

**Total hard cost to user:** ~$0.01 in POL gas for one transfer. Everything else is automated.

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

## Funding Flow — POL + pUSD → Deposit Wallet

### The Only Two Things the User Needs

| Need | Amount | Purpose |
|------|--------|---------|
| POL (MATIC) | ~0.01 | Gas for ONE ERC-20 transfer (EOA → deposit wallet) |
| pUSD | Whatever you want to trade | Trading collateral |

Everything else is gas-sponsored by the Polymarket relayer. One transaction, paid in POL, gets you funded. After that, zero gas forever.

### The pUSD Pipeline

```
┌─────────────────────────────────────────────────────────────────┐
│                    pUSD FUNDING FLOW                             │
│                                                                 │
│  Step 1: Get POL (~$0.01)                                       │
│    Any exchange → withdraw POL to EOA on Polygon                 │
│    This is the ONLY gas you'll ever pay.                        │
│                                                                 │
│  Step 2: Get pUSD on Polygon                                    │
│    Option A: Polymarket Bridge API (auto-converts USDC → pUSD)  │
│    Option B: Polymarket.com deposit (converts USDC → pUSD)      │
│    Option C: Call CollateralOnramp.deposit(USDC) on-chain       │
│                                                                 │
│  Step 3: Transfer pUSD EOA → Deposit Wallet                     │
│    polygolem deposit-wallet fund --amount X                     │
│    One ERC-20 transfer. That's it. Gas: ~$0.01 POL.             │
│                                                                 │
│  After this: zero POL needed. All trading is gas-sponsored.     │
└─────────────────────────────────────────────────────────────────┘
```

> **pUSD homogeneity strategy:** We commit to pUSD as the settlement token for ALL markets — Polymarket's existing markets AND any future markets built on the Arenaton/Polydart stack. One token, one wallet, one funding pipeline. No bridging, no wrapping, no multi-chain fragmentation.

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

### pUSD-Native Markets — Your Own Prediction Markets

The deposit wallet + pUSD gives you a turnkey settlement layer:

| Capability | Why It Matters |
|-----------|---------------|
| **Single collateral token** | All markets settle in pUSD. No per-market token fragmentation. |
| **One wallet, all markets** | Users fund once, trade on Polymarket AND your custom markets. |
| **ERC-1271 signatures** | EOA signs, deposit wallet validates. No per-market key management. |
| **Batch execution** | Approve + trade in one `factory.proxy()` call. No per-market approval flow. |
| **Same gas model** | Polymarket relayer sponsors gas. Your markets can use the same or similar relayer. |
| **1:1 USDC backing** | pUSD is fully backed on-chain. No algorithmic risk. No peg to maintain. |

### Wallet Interface

```solidity
// Core capabilities of the deposit wallet
interface IDepositWallet {
    // ERC-1271 — any protocol calls this to validate EOA signatures
    function isValidSignature(bytes32 hash, bytes calldata signature) 
        external view returns (bytes4 magicValue);
    function owner() external view returns (address);
}

// Factory — deploy + execute
interface IDepositWalletFactory {
    function proxy(Batch[] batches, bytes[] signatures) external;  // UNGATED
    function predictWalletAddress(address impl, bytes32 id) external view returns (address);
}
```

---

## Polygolem API Surface

| Command | What it does | Gas Cost |
|---------|-------------|----------|
| `polygolem builder auto` | Create or derive CLOB L2 creds (ClobAuth EIP-712) | Free |
| `polygolem auth headless-onboard` | Mint V2 relayer API key via SIWE | Free |
| `polygolem clob create-builder-fee-key` | Mint CLOB builder attribution key | Free |
| `polygolem clob create-api-key-for-address --owner <wallet>` | Mint owner-scoped CLOB L2 creds | Free |
| `polygolem deposit-wallet derive` | Predict CREATE2 wallet address (local) | Free |
| `polygolem deposit-wallet status` | Check deployed?, approvals?, balance? | Free |
| `polygolem deposit-wallet deploy --wait` | WALLET-CREATE via relayer | Sponsored |
| `polygolem deposit-wallet approve --submit` | 6-call WALLET batch via relayer | Sponsored |
| `polygolem deposit-wallet fund --amount X` | ERC-20 transfer EOA → deposit wallet | ~$0.01 POL |
| `polygolem deposit-wallet onboard --fund-amount X` | deploy + approve + fund (all-in-one) | ~$0.01 POL |
| `polygolem bridge deposit <addr>` | Get EVM/Solana/BTC deposit addresses | Free |
| `polygolem clob create-order ...` | Place order (POLY_1271 signed) | Sponsored |

---

## What the old doc got right and wrong

`docs/BUILDER-CREDENTIAL-ISSUANCE.md` asserted that builder creds are gated behind a Polymarket session cookie and that `/auth/api-key` only issues "CLOB L2" creds (a separate type from "Builder API Keys").

**The two-cred-types observation was correct.** As of the V2 cutover, `/auth/api-key` issues CLOB L2 creds, and Relayer API Keys are a distinct key minted by `auth headless-onboard` or the settings page. See § Three Credential Types above for the empirical 401 evidence.

**The session-cookie claim was wrong.** The browser flow mints a standard session through SIWE; polygolem now reproduces that flow headlessly and then calls the relayer key endpoint.

**Bottom line:** `polygolem builder auto` mints CLOB L2 creds, `polygolem auth headless-onboard` mints relayer creds, and `polygolem clob create-builder-fee-key` mints order-attribution keys.
