# polygolem

[![CI](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml/badge.svg)](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TrebuchetDynamics/polygolem)](go.mod)
[![Latest Release](https://img.shields.io/github/v/tag/TrebuchetDynamics/polygolem?label=release&sort=semver)](https://github.com/TrebuchetDynamics/polygolem/releases)

**Production-safe Polymarket infrastructure for the deposit-wallet era.**

A single Go binary and SDK for trading on Polymarket V2 — built specifically
for the deposit-wallet signing path (Polymarket's current production model
for new API users; sometimes called *type 3* or *POLY_1271* in the docs).
No external SDKs. No Python runtime. No opaque signing wrappers. Wallet
derivation, signing, relayer flows, and protocol logic are implemented
directly from spec in Go.

*For operators who want verifiable trading infrastructure instead of opaque
wrappers.*

---

## Why polygolem Exists

In April 2026 Polymarket migrated to V2: a new exchange, a new stablecoin
(pUSD), and a new requirement that orders be placed by **deposit wallets**
(ERC-1967 proxies that validate signatures via ERC-1271) instead of EOAs.

Most existing Polymarket bots, wrappers, and unofficial SDKs still assume
EOA signing. Many of them silently broke; the rest produce ghost fills that
appear in the book and never settle.

polygolem only knows the production-safe path:

- **Deposit-wallet only** — `signatureType=3` (`POLY_1271`) for every order
- **Local signing** — your private key never leaves the process
- **Spec-implemented protocol** — no shimmed Python or JS SDKs in the trust path
- **CREATE2 wallet derivation verified** against the official Polymarket Python SDK
- **Read-only by default** — every authenticated command requires explicit credentials

EOA, proxy, and Gnosis Safe paths are intentionally not supported.

---

## What Works Today

Verified end-to-end against Polygon mainnet on the 2026-05-08 reference run:

- Headless V2 relayer onboarding + deposit-wallet deploy and funding
- CLOB V2 limit and market orders (post-only / GTC / GTD / FOK) with cancels
- Builder-code attribution (V2 bytes32 model)
- Public market discovery + streaming (Gamma, CLOB, Data API, WebSocket)
- Local risk controls — per-trade caps, daily loss limits, circuit breaker

→ [Full walkthrough — every tx hash, gas figure, and pUSD movement](docs/LIVE-TRADE-WALKTHROUGH.md)

---

## Safety Model

polygolem is trading infrastructure. Trust matters more than features.

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

## Try It in 60 Seconds (No Credentials Needed)

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

## Demo Pipeline

The planned `polygolem demo` command is the new-user showroom for the SDK and
CLI. It should walk through the Polymarket API graph in one run, choosing live
read-only data and explaining how Gamma, CLOB, Data, and CTF identifiers connect.

```bash
polygolem demo              # default: L1 public, unauthenticated, read-only
polygolem demo --layer l1   # explicit public API demo
polygolem demo --layer l2   # authenticated CLOB L2 read-only account context
polygolem demo --query btc --limit 5 --seed 123
polygolem demo --json
```

`--layer l1` uses only public APIs. It lists a small set of active campaigns
(Polymarket events/series/markets), shows their tags/categories, picks one
market, then reads Gamma metadata, CLOB market/order-book/price/tick/history
data, and Data API analytics. No signup, private key, funding, approvals,
orders, or cancels are involved.

`--layer l2` runs the same public pipeline and adds authenticated read-only
CLOB context from configured L2 credentials: open orders, trade history, and
balance/allowance. It must remain read-only: no `POST /order`, no cancels, no
wallet deployment, no approvals, and no funding.

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

---

## Go SDK

If you'd rather embed polygolem in a larger Go service, every CLI subcommand
is a thin wrapper around an importable `pkg/` package:

| Package | What it does |
|---|---|
| [`pkg/universal`](pkg/universal) | One typed client over Gamma + CLOB + Data API + Stream + Discovery (70+ methods) |
| [`pkg/clob`](pkg/clob) | CLOB V2 — market data, orders, balances, builder fees |
| [`pkg/gamma`](pkg/gamma) | Read-only Gamma market discovery (26 methods) |
| [`pkg/data`](pkg/data) | Data API — positions, volume, leaderboards |
| [`pkg/stream`](pkg/stream) | Public CLOB WebSocket market stream |
| [`pkg/orderbook`](pkg/orderbook) | Read-only book + price + spread |
| [`pkg/relayer`](pkg/relayer) | V2 Relayer client — WALLET-CREATE, batch, nonce |
| [`pkg/builder`](pkg/builder) | Builder HMAC signing (local or remote) |
| [`pkg/marketresolver`](pkg/marketresolver) | Market + token ID resolution |
| [`pkg/types`](pkg/types) | Shared public DTOs |

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

Internal implementation details live under `internal/` and are documented in
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

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
| [Contracts](docs/CONTRACTS.md) | All contract addresses, factory ABI, CREATE2 derivation. |
| [Architecture](docs/ARCHITECTURE.md) | Package boundaries and dependency direction. |
| [Commands](docs/COMMANDS.md) | Auto-generated CLI reference. |
| [Deposit Wallet Migration](docs/DEPOSIT-WALLET-MIGRATION.md) | V1→V2 survival guide for older bots. |
| [Astro Docs site](https://trebuchetdynamics.github.io/polygolem) | Searchable HTML version of the full doc tree. |

---

## Status

`v0.1.0` — production-validated against Polygon mainnet on **2026-05-08**
([reference run](docs/LIVE-TRADE-WALKTHROUGH.md)).

**Current scope**

- Deposit-wallet lifecycle (derive, deploy, approve, fund)
- CLOB V2 trading (limit, market, post-only, GTC / GTD / FOK, cancels)
- Builder-code attribution
- Universal SDK client over Gamma + CLOB + Data + Stream

**Still hardening**

- Broader exchange abstractions
- Release signing
- Extended automation surfaces

See [`CHANGELOG.md`](CHANGELOG.md) for per-version detail.
