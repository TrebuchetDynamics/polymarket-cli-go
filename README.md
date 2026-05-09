# polygolem

[![CI](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml/badge.svg)](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TrebuchetDynamics/polygolem)](go.mod)
[![Latest Release](https://img.shields.io/github/v/tag/TrebuchetDynamics/polygolem?label=release&sort=semver)](https://github.com/TrebuchetDynamics/polygolem/releases)

**Production-safe Polymarket infrastructure for the deposit-wallet era.**

A single Go binary and SDK for trading on Polymarket V2.

Built for Polymarket's current deposit-wallet production model — with local
signing, no external SDKs, and no opaque runtime layers.

*For operators who want verifiable trading infrastructure instead of opaque
wrappers — validated against live Polygon mainnet flows, not mocks or paper
environments.*

---

## Try It in 60 Seconds (No Credentials Needed)

All outputs below come from live Polygon mainnet responses.

```bash
git clone https://github.com/TrebuchetDynamics/polygolem
cd polygolem && go build -o polygolem ./cmd/polygolem

./polygolem health
# {"clob":"ok","gamma":"ok"}

./polygolem orderbook price --token-id 1391568931...637394586
# {"price":"0.012","token_id":"1391568931...637394586"}

./polygolem orderbook spread --token-id 1391568931...637394586
# {"spread":"0.002","token_id":"1391568931...637394586"}

./polygolem discover search --query "btc 150k" --limit 3
```

That's polygolem talking to live Polymarket — no key, no credentials, no
sign-up. Read-only is the default for everything until you set
`POLYMARKET_PRIVATE_KEY`.

---

## Production Validation

> **Production-validated:** Polygon mainnet · 2026-05-08 reference run
>
> [Every tx hash, gas figure, and pUSD movement](docs/LIVE-TRADE-WALKTHROUGH.md)
> is documented from EOA private key to filled buy + sell.

Core trading flows are production-validated today: headless relayer onboarding,
deposit-wallet deploy + funding, CLOB V2 trading + cancels, advanced order
types, market discovery, streaming, and local risk controls.

---

## Why polygolem Exists

In April 2026 Polymarket migrated to V2: a new exchange, a new stablecoin
(pUSD), and a new requirement that orders be placed by **deposit wallets**
(ERC-1967 proxies that validate signatures via ERC-1271) instead of EOAs.

That broke the old assumption that an EOA is the order maker. Many existing
Polymarket bots, wrappers, and unofficial SDKs still sign as EOAs; those paths
can produce ghost fills that appear in the book and never settle.

polygolem only knows the production-safe path:

- **Deposit-wallet only** — the current production order model
  (`signatureType=3` / `POLY_1271`, validated on-chain via ERC-1271)
- **Local signing** — your private key never leaves the process
- **Spec-implemented protocol** — no shimmed Python or JS SDK in the trust path
- **CREATE2 wallet derivation verified** against the official Polymarket Python SDK
- **Read-only by default** — every authenticated command requires explicit credentials

EOA, proxy, and Gnosis Safe paths are intentionally not supported.

> **Why Go?**
>
> One language. One binary. No hidden runtime layers between your private key
> and the exchange. No transitive npm or pip dependencies that get to see your
> signing key.

The on-chain identity model the rest of this README is built around:

```
  EOA  ──signs──▶  Order
   │              (signatureType=3, maker=DepositWallet, signer=DepositWallet)
   │
   ▼ derives (CREATE2)              ▼ submitted by
 Deposit Wallet  ◀──holds pUSD──    Polymarket matching engine
 (ERC-1967 proxy,                   (gas-sponsored fillOrders settlement)
  validates signatures              ──┐
  via ERC-1271)                       │
                                      ▼
 V2 Relayer  ──sponsors──▶  WALLET-CREATE + approval batch
 (relayer-v2.polymarket.com)
```

The EOA signs; the deposit wallet holds funds and is the on-order maker;
Polymarket-run services pay every gas fee except your single ERC-20 funding
transfer. See [docs/LIVE-TRADE-WALKTHROUGH.md](docs/LIVE-TRADE-WALKTHROUGH.md)
for the full lifecycle with real txes.

---

## What Works Today

Covered by the 2026-05-08 Polygon mainnet reference run:

**Trading**

- Headless V2 relayer onboarding
- Deposit-wallet deploy + funding
- CLOB V2 trading + cancels
- Advanced order types

**Market Data**

- Gamma + CLOB market discovery
- Public CLOB WebSocket market stream
- Order book price + spread helpers

**Safety**

- Read-only default CLI/SDK surface
- Local risk controls
- Secret redaction

---

## Safety Model

polygolem is trading infrastructure. Trust matters more than features.
The SDK defaults toward preventing irreversible mistakes.

| | |
|---|---|
| **Read-only by default** | Authenticated operations require an explicit private key in env. |
| **Deposit-wallet only** | Cannot accidentally sign as an EOA, proxy, or Safe. |
| **Local signing** | The process holds the key; no signing service in the trust path. |
| **No external SDKs** | All wallet derivation, EIP-712, ERC-7739, and relayer calls are in this repo. |
| **Pre-trade caps + daily limits + circuit breaker** | Configurable risk controls in `internal/risk`. |
| **Secret redaction** | API keys and signatures are redacted in logs (`internal/transport`). |

See [docs/SAFETY.md](docs/SAFETY.md) for the full safety model.

---

## Trade in Four Commands

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

polygolem auth headless-onboard                     # mint V2 relayer key (gasless)
polygolem deposit-wallet onboard --fund-amount 0.71 # deploy + approve + fund
polygolem clob update-balance --asset-type collateral
polygolem clob market-order --token <ID> --side buy --amount 1 --price 0.012 --order-type FOK
# {
#   "success": true,
#   "orderID": "0x43083109...c423d793d",
#   "status": "matched",
#   "makingAmount": "1",
#   "takingAmount": "86.606666",
#   "transactionsHashes": ["0x74ad015d...4f7adc"]
# }
```

**After onboarding, every trade is fully headless.** Total user-paid cost on
the reference run was **~$0.01 in POL gas** for the single ERC-20 transfer
that funds the deposit wallet — `WALLET-CREATE`, the 6-call approval batch,
and every CLOB settlement are sponsored by Polymarket-run services. See
[the walkthrough](docs/LIVE-TRADE-WALKTHROUGH.md) for the per-tx breakdown.

> ⚠️ **New users need one browser login.** Polymarket's L1 auth endpoint
> (`/auth/api-key`) does not currently support ERC-1271 validation, so a
> brand-new EOA needs one browser login at polymarket.com to mint the
> deposit-wallet-bound CLOB API key. **After that, all trading is fully
> headless.** Existing Polymarket users with an already-minted CLOB key skip
> this entirely. See [docs/BROWSER-SETUP.md](docs/BROWSER-SETUP.md).

### After a Fill

`status=matched` means the CLOB filled the order. It does not mean the market
won. Polygolem treats post-trade state as three separate checks:

- `matched`: the order filled and shares moved into the deposit wallet.
- `winning`: the market resolved in favor of the held outcome.
- `redeemable`: the Data API reports the held winning position can be redeemed.

For V2 deposit wallets, redeem has one supported production path: the owner
signs an EIP-712 WALLET batch, Polymarket's relayer submits it through the
deposit-wallet factory, and the wallet call targets a pUSD collateral adapter.
Standard markets use `CtfCollateralAdapter`; negative-risk markets use
`NegRiskCtfCollateralAdapter`. Existing deposit wallets that only ran the
six-call trading approval batch need a separate adapter-approval migration
before their first V2 redeem.

The first-class SDK/CLI settlement surface is `pkg/settlement` plus
`deposit-wallet settlement-status`, `deposit-wallet redeemable`, and
`deposit-wallet redeem`. `settlement-status` is the read-only readiness gate:
it checks wallet bytecode, relayer credentials, Data API reachability, and
adapter approvals before a live bot should place more orders. The redeem
commands build the V2 adapter path and fail closed on missing adapter
approvals. If the relayer rejects adapter calls as not allowlisted, stop; the
production factory does not expose a direct EOA fallback and raw
`ConditionalTokens` redeem is not a deposit-wallet fallback. SAFE/PROXY
relayer examples do not apply to deposit-wallet positions. See
[docs/SAFETY.md](docs/SAFETY.md),
[docs/CONTRACTS.md](docs/CONTRACTS.md), and
[docs/DEPOSIT-WALLET-REDEEM-VALIDATION.md](docs/DEPOSIT-WALLET-REDEEM-VALIDATION.md).

---

## Go SDK

If you'd rather embed polygolem in a larger Go service, every CLI subcommand
is a thin wrapper around importable `pkg/` packages:

| Package | What it does |
|---|---|
| [`pkg/universal`](pkg/universal) | One typed client over Gamma + CLOB + Data API + Stream + Discovery (70+ methods) |
| [`pkg/clob`](pkg/clob) | CLOB V2 — market data, orders, balances, builder fees |
| [`pkg/gamma`](pkg/gamma) | Read-only Gamma market discovery (26 methods) |
| [`pkg/stream`](pkg/stream) | Public CLOB WebSocket market stream |
| [`pkg/relayer`](pkg/relayer) | V2 Relayer client — WALLET-CREATE, batch, nonce |
| [`pkg/settlement`](pkg/settlement) | V2 winner redemption planning, adapter calls, and readiness gates |

```go
import (
    "context"
    "fmt"

    "github.com/TrebuchetDynamics/polygolem/pkg/universal"
)

c := universal.NewClient(universal.Config{})
ctx := context.Background()

const btcYesToken = "13915689317269078219168496739008737517740566192006337297676041270492637394586"

price, _ := c.Price(ctx, btcYesToken, "buy")
spread, _ := c.Spread(ctx, btcYesToken)
fmt.Printf("BTC $150k YES — price %s, spread %s\n", price, spread)
// BTC $150k YES — price 0.012, spread 0.002
```

Full package boundaries, dependency direction, and internal implementation
details are documented in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## Common Workflows

| I want to... | Run |
|---|---|
| Find an active market | `polygolem discover search --query "..."` |
| Inspect the book | `polygolem clob book <token-id>` |
| Check my deposit wallet status | `polygolem deposit-wallet status` |
| Place a limit buy | `polygolem clob create-order --token <ID> --side buy --price 0.5 --size 10` |
| Place a market FOK buy | `polygolem clob market-order --token <ID> --side buy --amount 1 --price <slippage_cap>` |
| Cancel everything | `polygolem clob cancel-all` |
| Read my collateral balance | `polygolem clob balance --asset-type collateral` |

Full CLI reference (auto-generated, every flag and example):
[docs/COMMANDS.md](docs/COMMANDS.md).

---

## Environment

- **Required for any authenticated command:** `POLYMARKET_PRIVATE_KEY`.
- **Auto-minted by polygolem on first use:** V2 relayer key, CLOB L2 key
  (existing users) — persisted to local env files.
- **Optional:** `POLYMARKET_BUILDER_CODE` for V2 order attribution.

The deposit wallet address is derived locally from the private key; no API
call required. Full env reference in [docs/ONBOARDING.md](docs/ONBOARDING.md).

---

## Docs

| Document | What it covers |
|---|---|
| [Live Trade Walkthrough](docs/LIVE-TRADE-WALKTHROUGH.md) | End-to-end 2026-05-08 reference run: every tx, gas figure, and pUSD movement from EOA private key to a filled buy + sell. |
| [Onboarding](docs/ONBOARDING.md) | Single source of truth — complete deposit wallet flow, troubleshooting. |
| [Browser Setup](docs/BROWSER-SETUP.md) | One-time browser login for new users; security guidance. |
| [Safety](docs/SAFETY.md) | Risk controls, deposit-wallet-only enforcement, circuit breakers. |
| [Contracts](docs/CONTRACTS.md) | Contract addresses, factory ABI, CREATE2 derivation, and deployment status source-of-truth rules. |
| [Architecture](docs/ARCHITECTURE.md) | Package boundaries and dependency direction. |
| [Commands](docs/COMMANDS.md) | Auto-generated CLI reference. |
| [Deposit Wallet Migration](docs/DEPOSIT-WALLET-MIGRATION.md) | V1→V2 survival guide for older bots. |
| [polygolem.trebuchetdynamics.com](https://polygolem.trebuchetdynamics.com) | Documentation site — searchable HTML version of the full doc tree, with landing page. |

---

## Status

`v0.1.0` — production-validated against Polygon mainnet on **2026-05-08**
([reference run](docs/LIVE-TRADE-WALKTHROUGH.md)).

Core trading flows are production-validated today.

Release signing, broader exchange abstractions, and extended automation
surfaces are still hardening.

Planned: a guided read-only demo pipeline for onboarding and API graph
exploration.

See [`CHANGELOG.md`](CHANGELOG.md) for per-version detail.
