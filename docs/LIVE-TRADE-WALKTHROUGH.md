# Live Trade Walkthrough — EOA Private Key → Filled Sell

This document records, in full, the end-to-end operation of taking a private
key and producing a real, settled round-trip trade on Polymarket V2 mainnet.
Every transaction hash, gas figure, and pUSD amount in this doc is from the
**2026-05-08 live run** that is the canonical reference for the polygolem
deposit-wallet pipeline.

The numbers below are not approximations. They are reconciled against
on-chain receipts on Polygon mainnet.

---

## 0. Inputs

| Input | Value |
|---|---|
| EOA address | `0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C` |
| EOA private key | `POLYMARKET_PRIVATE_KEY` env var (hex, 64-char) |
| Deposit wallet (CREATE2-derived) | `0x21999a074344610057c9b2B362332388a44502D4` |
| Market | "Will Bitcoin hit $150k by June 30, 2026?" |
| Condition ID | `0xa0f4c4924ea1a8b410b4ce821c2a9955fad21a1b19bdcfde90816732278b3dd5` |
| YES token id | `13915689317269078219168496739008737517740566192006337297676041270492637394586` |

The deposit wallet address is a deterministic function of the EOA and the
Polymarket WalletFactory (CREATE2 salt + init code hash). It exists in the
math before it exists on-chain. Polygolem derives it via
`MakerAddressForSignatureType(eoa, 137, 3)` (`internal/auth`).

---

## 1. The Two Identities

Polymarket V2 separates two addresses that older guides conflate:

| Identity | Address | What it does |
|---|---|---|
| **EOA** | `0x33e4...` | Holds POL for gas. Signs every L1 (`/auth/api-key`, `/auth/derive-api-key`) and L2 (HMAC-tagged) request. Signs ClobAuth EIP-712 messages. **Never** holds pUSD or shares. |
| **Deposit wallet** | `0x21999a07...` | ERC-1967 proxy with `isValidSignature` (ERC-1271). Holds pUSD and conditional tokens. Is the `maker` and `signer` on every order. Signature type 3 (`POLY_1271`). |

The CLOB API key is **EOA-bound at the HTTP layer** but trades on behalf of
the deposit wallet via the order's `signatureType=3` field. This was empirically
verified by the 2026-05-08 Playwright capture and is documented in
[POLY_1271-SIGNING.md](./POLY_1271-SIGNING.md).

### How an order is authorised vs settled

When you place an order, the *signing* and the *settling* are two distinct
events with two distinct payers:

1. **Off-chain signing (you).** The EOA produces a 65-byte ECDSA signature
   over the order hash. The signature is wrapped in the ERC-7739 envelope so
   the on-chain Exchange can validate it via the deposit wallet's
   `isValidSignature(orderHash, sig)`. The deposit wallet contract unwraps
   the envelope, recovers the EOA, and accepts the signature because the EOA
   is its authorised owner (set when the wallet was deployed).
2. **On-chain settlement (Polymarket).** Once the matching engine pairs your
   order with a counterparty, the **matching-engine operator** submits a
   `fillOrders` tx that executes the pUSD ↔ CTF transfers. **You pay zero
   gas for this tx** even though it consumes hundreds of thousands of gas.

### Two Polymarket-run services — don't confuse them

Two different Polymarket-operated services pay gas on your behalf at
different points in the lifecycle. They are **not** the same infrastructure:

| Service | Endpoint / address | When it acts | What it pays for |
|---|---|---|---|
| **V2 Relayer** | `relayer-v2.polymarket.com/submit` | One-time onboarding | `WALLET-CREATE` (deposit wallet deployment) and the 6-call ERC-20 approval batch. |
| **Matching-engine operator** | Settler EOAs (e.g. `0x0484…de55`, `0x9b5b…e9ab`) on Polygon | Every matched order | `fillOrders` settlement txes that move pUSD and mint/burn CTF tokens. |

Throughout this doc, "**relayer**" means the V2 Relayer (onboarding sponsorship)
and "**operator**" means the matching-engine settler (trade settlement). Both
are Polymarket-run; both pay gas instead of you; they are separate systems.

---

## 2. One-Time Setup (Per EOA)

These steps run once. After they complete, every subsequent trade is a single
HTTP POST.

### 2.1 V2 Relayer Key — `polygolem auth headless-onboard`

```
SIWE login → POST /profiles → POST /relayer-auth
→ persists POLYMARKET_RELAYER_API_KEY + POLYMARKET_RELAYER_API_KEY_ADDRESS to .env.relayer-v2
```

Gas: **0** (no on-chain tx).
Cost: **$0**.

### 2.2 CLOB API Key — one-time browser login (new users only)

For a brand-new EOA, the CLOB `/auth/api-key` endpoint requires a one-time
browser login to mint the deposit-wallet-bound API key. After that one login,
all subsequent calls are headless.

Existing users with an already-minted key skip this step entirely — polygolem
reads `POLYMARKET_CLOB_API_KEY` / `_SECRET` / `_PASSPHRASE` from the env and
uses them directly.

Gas: **0**.
Cost: **$0**.

See [BROWSER-SETUP.md](./BROWSER-SETUP.md) for the one-time procedure and
[BLOCKERS.md](../BLOCKERS.md) for the residual server-side ERC-1271 gap.

### 2.3 Deposit Wallet Deployment — `polygolem deposit-wallet deploy --wait`

`POST /relayer/submit` with `{"type":"WALLET-CREATE", ...}`. The Polymarket
relayer broadcasts the WalletFactory call **paying gas itself**. Polls until
the relayer reports `STATE_MINED`.

Gas (paid by relayer, not user): **~750k**.
Cost to user: **$0**.

### 2.4 ERC-20 Approvals — `polygolem deposit-wallet onboard --skip-deploy`

A 6-call relayer batch that approves the V2 Exchange, the neg-risk Exchange,
and the V1 Exchange to spend pUSD and conditional tokens from the deposit
wallet. All sponsored.

Gas (paid by relayer): **~250k–500k**.
Cost to user: **$0**.

After 2.4, `polygolem clob balance` shows the three exchange addresses with
`type(uint256).max` allowances for both pUSD and CTF tokens.

---

## 3. Per-Trade Pipeline (User-Paid Steps Only)

The remainder of this doc walks the four user-paid transactions and the two
operator-paid CLOB settlements that produced one filled buy and one filled
sell. Every receipt was confirmed on Polygon mainnet via
`https://polygon.drpc.org`.

### 3.1 POL → pUSD Swap — `polygolem deposit-wallet swap-pol-pusd`

The deposit wallet must hold pUSD. The EOA holds POL. Polygolem bridges via
Uniswap V3 multihop on Polygon, **without** any L2 bridge.

```bash
polygolem deposit-wallet swap-pol-pusd --out-pusd 0.5 --max-pol-in 10
```

Path encoding (output-first, V3 exactOutput convention):

```
pUSD ‖ 0x000bb8 (fee=3000, 0.30%)
‖ USDC.e ‖ 0x0001f4 (fee=500, 0.05%)
‖ WMATIC
```

The 0.30% fee tier on the pUSD/USDC.e leg is mandatory: the 0.05% pool is empty
(verified 2026-05-08: liquidity = 0). The 0.05% tier on USDC.e/WMATIC has the
deepest liquidity on Polygon. See `internal/rpc/swap.go:71-80`.

Excess input POL is auto-refunded by the router via `multicall(refundETH)`.
Native POL is auto-wrapped to WMATIC because the path begins with WMATIC.

#### Two swaps in this run

| # | tx | block | gasUsed | gasPrice (gwei) | gas cost (POL) |
|---|---|---:|---:|---:|---:|
| Swap 1 (out 0.725 pUSD) | [`0x689df05e…fca5622e`](https://polygonscan.com/tx/0x689df05e21f196c7869801eb52c1e2f01585e9f26887afa7cf45ea57fca5622e) | 86 588 695 | 193 134 | 277.501 | **0.053595** |
| Swap 2 (top-up, out 0.5 pUSD) | [`0xf26744ef…87ec4b3`](https://polygonscan.com/tx/0xf26744ef11d0aaace80665c0e8a37fcd345f379139706969d1991626c87ec4b3) | 86 592 289 | 192 190 | 275.748 | **0.052996** |

Both swaps land pUSD on the EOA, not the deposit wallet. They have to be
followed by a step-3.2 transfer.

### 3.2 EOA → Deposit Wallet pUSD Transfer — `polygolem deposit-wallet fund`

Standard ERC-20 `transfer(depositWallet, amount)` on the pUSD contract
(`0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB` on Polygon).

```bash
polygolem deposit-wallet fund --amount 0.5
```

| # | tx | block | gasUsed | gasPrice (gwei) | gas cost (POL) |
|---|---|---:|---:|---:|---:|
| Fund 1 (0.71 pUSD) | [`0xa2ac65c4…74ae191`](https://polygonscan.com/tx/0xa2ac65c4bb99117d0419f7ff851d8addaebad140856dba7df9a0bab0174ae191) | 86 588 885 | 55 868 | 277.803 | **0.015520** |
| Fund 2 (top-up, 0.5 pUSD) | [`0xa9f49d61…ee2fe00f6`](https://polygonscan.com/tx/0xa9f49d6195e5d728717d654aa72457ed7204207869ff878f6a1f115ee2fe00f6) | 86 592 295 | 38 768 | 275.488 | **0.010680** |

The first transfer costs more because it initialises the deposit wallet's
pUSD balance storage slot (~20k extra SSTORE).

### 3.3 CLOB Balance Refresh — `polygolem clob update-balance`

```bash
polygolem clob update-balance --asset-type collateral
```

`GET /balance-allowance/update?asset_type=COLLATERAL&signature_type=3`. The
endpoint returns HTTP 200 with an empty body — it just queues a refresh of
Polymarket's internal cache. No on-chain tx.

Gas: **0**.
Cost: **$0**.

(Polygolem tolerates the empty body since
[commit 8bf520b](https://github.com/TrebuchetDynamics/polygolem/commit/8bf520b);
prior versions surfaced an EOF decode error and the CLI was unusable.)

### 3.4 Market Buy — `polygolem clob market-order`

```bash
polygolem clob market-order \
  --token 13915689317269078219168496739008737517740566192006337297676041270492637394586 \
  --side buy --amount 1 --price 0.012 --order-type FOK
```

`amount=1` is pUSD spent (maker amount, capped at 2 decimals).
`price=0.012` is the worst per-share price the engine may pay (slippage cap).
`FOK` = fill-or-kill: the entire 1.00 pUSD must fill or the order is rejected.

#### CLOB minimums and decimal limits — what bit me here

The CLOB rejects orders that violate any of:

| Rule | Buy | Sell |
|---|---|---|
| Maker-amount decimals | ≤ 2 (pUSD) | ≤ 5 (shares) |
| Taker-amount decimals | ≤ 5 (shares) | ≤ 2 (pUSD) |
| Marketable order minimum | **$1.00 pUSD** | (same applies; sub-$1 fillable orders rejected) |
| Tick size | per market (this market: 0.001) |
| Min order size (post-only OK) | per market (this market: 5 shares) |

The first attempt of this trade used `--amount 0.055`, was rejected as
`"min size: $1"`. The wallet had to be topped up before a marketable trade
could fill.

#### Order response

```json
{
  "success": true,
  "orderID": "0x43083109e1d26284ddaf2618503df661fd60ff54f478ad9a8746948c423d793d",
  "status": "matched",
  "makingAmount": "1",
  "takingAmount": "86.606666",
  "transactionsHashes": [
    "0x74ad015d2de4d3c0e5559d72a49e3f85cd28c6c8ceafb7a942a1c89a6647823d"
  ]
}
```

The CLOB matched the order against the resting `0.011 × 39.28` and
`0.012 × 47.32` asks. Effective fill price: **1.00 / 86.606666 = 0.011546
pUSD per share** (a half-tick worse than the best ask, because the order
crossed two price levels).

#### On-chain settlement — paid by the operator, NOT by you

The CLOB operator submits a `fillOrders` tx that transfers pUSD out of the
deposit wallet, mints YES CTF tokens, and credits them to the deposit wallet.

| field | value |
|---|---|
| tx | [`0x74ad015d…4f7adc`](https://polygonscan.com/tx/0x74ad015d2de4d3c0e5559d72a49e3f85cd28c6c8ceafb7a942a1c89a6647823d) |
| block | 86 592 311 |
| gasUsed | 596 113 |
| effective gas price | 336.153 gwei |
| gas cost | **0.200385 POL** (paid by operator `0x0484…de55`) |
| `from` | Polymarket operator (`0x0484b1c7537083e8efb62865c30a885f277ade55`) |
| logs | 30 (Transfer events for pUSD + CTF mints + book updates) |

Even though this single settlement burned 0.20 POL of gas, the user pays
**zero** because the operator amortises that cost across the matching engine's
settlement batch (the same tx may settle multiple users' orders).

### 3.5 Limit Sell — `polygolem clob create-order --side sell`

```bash
polygolem clob create-order \
  --token 13915689317269078219168496739008737517740566192006337297676041270492637394586 \
  --side sell --price 0.010 --size 86.6 --order-type GTC
```

Sized at 86.6 shares (1-decimal precision) instead of the full 86.606666
because polygolem's order-amount math at 5-decimal sizes produced a sub-tick
price (`0.0099999930721263 < 0.010`) which the CLOB rejected. Working size
× price ratios that yield clean ticks is a quirk of the order encoding worth
internalising before automating size selection.

The order was a GTC limit; because the bid was 0.010 with 1364 shares of
size, the order crossed the book and matched immediately:

```json
{
  "success": true,
  "orderID": "0xa4c50b0145989d8332d85959be50e9cc84f816579d389a97eb3a984b3ceae47c",
  "status": "matched",
  "makingAmount": "86.6",
  "takingAmount": "0.866",
  "transactionsHashes": [
    "0x57a36b0bb2ad2c4d27ab9618eea57933f455845431914816d48e48f7f04f7adc"
  ]
}
```

Effective sell price: **0.866 / 86.6 = 0.010 pUSD per share** (exactly the
best bid, no slippage past the top of book).

| field | value |
|---|---|
| settlement tx | [`0x57a36b0b…4f7adc`](https://polygonscan.com/tx/0x57a36b0bb2ad2c4d27ab9618eea57933f455845431914816d48e48f7f04f7adc) |
| block | 86 592 697 |
| gasUsed | 368 829 |
| effective gas price | 329.449 gwei |
| gas cost | **0.121510 POL** (paid by operator `0x9b5b…e9ab`) |
| `from` | Polymarket operator (`0x9b5bd059a2adb5736eb0ee2da06a485e19a6e9ab`) |
| logs | 18 (Transfer events for CTF burns + pUSD credit) |

After this tx, deposit wallet holds:
- **0.939926 pUSD** (collateral)
- **0.006666 YES shares** (residue from the 86.606666 → 86.6 truncation)

---

## 4. Total Cost Breakdown — One Round-Trip

### 4.1 User-paid Polygon gas

| Step | Tx | gasUsed | POL cost |
|---|---|---:|---:|
| Swap 1 (POL → 0.725 pUSD) | `0x689df05e…` | 193 134 | 0.053595 |
| Fund 1 (0.71 pUSD → DW) | `0xa2ac65c4…` | 55 868 | 0.015520 |
| Swap 2 (top-up, POL → 0.5 pUSD) | `0xf26744ef…` | 192 190 | 0.052996 |
| Fund 2 (top-up, 0.5 pUSD → DW) | `0xa9f49d61…` | 38 768 | 0.010680 |
| **Total user gas** | | **480 (k gas)** | **0.132791 POL** |

At POL ≈ $0.07, total user-paid gas ≈ **$0.0093** (just under one cent).

A user who sized the initial swap correctly the first time would skip the
top-up pair, halving this to **0.069 POL ≈ $0.0048** (about half a cent).

### 4.2 Operator-paid settlement gas (informational; **user pays $0**)

| Step | Tx | gasUsed | POL cost (operator) |
|---|---|---:|---:|
| Buy settle | `0x74ad015d…` | 596 113 | 0.200385 |
| Sell settle | `0x57a36b0b…` | 368 829 | 0.121510 |

These are amortised across all orders in the operator's matching batch — your
share is effectively zero, even though a single tx burned more gas than every
user-paid step combined.

### 4.3 Polymarket protocol fees

V2 server-side maker and taker fees: **0 bps** as of 2026-05-08 (confirmed via
`maker_base_fee` / `taker_base_fee` returned by `GET /markets`). All fee logic
is server-driven; orders carry no on-order fee field. If Polymarket turns fees
on, polygolem's order code does not change — the matching engine simply takes
the fee from the fill.

### 4.4 Cross-spread cost (the actual P&L driver)

Average fill prices:

| Side | Effective price | Notes |
|---|---:|---|
| Buy | 0.011546 | Crossed 0.011 ask + 0.012 ask |
| Sell | 0.010000 | Sat exactly on the best bid |

Round-trip per share: `0.011546 - 0.010000 = 0.001546 pUSD`.
Across 86.6 shares: `0.001546 × 86.6 = 0.13388 pUSD` lost to spread.

This is by far the dominant cost. In a market with a 1-tick proportional
spread (`0.011 / 0.010 - 1 = 10 %`), a market round-trip of any size will
forfeit ~10 % of capital. **Spread cost dwarfs gas at this scale.**

### 4.5 Residue dust

86.606666 bought – 86.6 sold = **0.006666 YES shares**, worth ~0.000067 pUSD
at the current bid. Negligible. Sweepable later for ~$0.001 of additional
gas if you care.

### 4.6 Final P&L

| Source | Amount (pUSD) |
|---|---:|
| Spent (buy) | −1.000000 |
| Received (sell) | +0.866000 |
| Residue (held) | +0.000067 |
| **Net pUSD P&L** | **−0.133933** |
| Plus user-paid gas | −0.132791 POL ≈ −$0.0093 (independent currency, charged to EOA) |

The trade was an end-to-end pipeline test, not a profit attempt. The
~13.4 % spread loss is the cost of round-tripping a low-priced binary at a
1-tick spread — independent of polygolem, the broker, or the auth flow.

---

## 5. Replay Recipe

If you want to reproduce the run from a fresh EOA:

```bash
# (One-time) Mint relayer + CLOB API key
polygolem auth headless-onboard
# Browser login at https://polymarket.com → settings → API keys (existing users skip)
# Persist POLYMARKET_CLOB_API_KEY / _SECRET / _PASSPHRASE to .env

# (One-time, gasless via relayer) deposit wallet deploy + approvals
polygolem deposit-wallet onboard

# Per-trade
polygolem deposit-wallet swap-pol-pusd --out-pusd 1.5 --max-pol-in 10
polygolem deposit-wallet fund --amount 1.5
polygolem clob update-balance --asset-type collateral
polygolem clob market-order --token <ID> --side buy --amount 1 --price <slippage_cap> --order-type FOK
polygolem clob create-order --token <ID> --side sell --price <bid> --size <clean_size> --order-type GTC
```

Sizing notes: keep `amount` at 2 decimals (pUSD), keep `size` at a precision
that yields a clean `size × price` product to dodge sub-tick rejections, and
remember the **$1 marketable minimum**.

---

## 6. References

- Polymarket V2 deposit-wallet docs — https://docs.polymarket.com/trading/deposit-wallets
- Polygolem onboarding (single source of truth) — [ONBOARDING.md](./ONBOARDING.md)
- POLY_1271 signing details — [POLY_1271-SIGNING.md](./POLY_1271-SIGNING.md)
- Smart-contract addresses — [CONTRACTS.md](./CONTRACTS.md)
- Browser one-time setup — [BROWSER-SETUP.md](./BROWSER-SETUP.md)
- pUSD contract — `0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB`
- USDC.e (Polygon) — `0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174`
- WMATIC — `0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270`
- Uniswap V3 SwapRouter02 (Polygon) — `0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45`
- V2 Exchange — `0xE111180000d2663C0091e4f400237545B87B996B`
- V2 neg-risk Exchange — `0xe2222d279d744050d28e00520010520000310F59`
- V1 Exchange — `0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296`
