# polygolem

[![CI](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml/badge.svg)](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TrebuchetDynamics/polygolem)](go.mod)
[![Latest Release](https://img.shields.io/github/v/tag/TrebuchetDynamics/polygolem?label=release&sort=semver)](https://github.com/TrebuchetDynamics/polygolem/releases)

**Production-safe Polymarket infrastructure for Go developers.**

A single binary + Go SDK for trading on Polymarket V2 through deposit wallets.
No Python. No npm. No opaque wrappers.

```bash
go install github.com/TrebuchetDynamics/polygolem/cmd/polygolem@latest
polygolem health
# {"clob":"ok","gamma":"ok"}
```

---

## Who This Is For

- **Bot developers** building automated trading strategies in Go
- **Quant developers** who want deterministic, compiled infrastructure with type safety
- **Operators** running headless trading systems that need auditability and local signing
- **Engineers** embedding Polymarket data and execution into larger Go services
- **Developers** who want one dependency, not a Python virtualenv + npm + Docker compose

If you are writing a Polymarket bot in Python or TypeScript, the [official CLOB clients](https://github.com/Polymarket/py-clob-client) are the right choice. If you are building in Go, or you want a single static binary with no runtime dependencies, polygolem is the only production-ready option.

---

## The Problem We Solve

Polymarket migrated to V2 in April 2026. The new model requires **deposit wallets** (ERC-1967 proxies with ERC-1271 validation) instead of EOAs as order makers. This broke most existing tooling.

| | Official Python/TS SDKs | polygolem |
|---|---|---|
| **Language** | Python / TypeScript | Go |
| **Dependencies** | pip/npm + 10+ transitive packages | Go stdlib + `cobra` |
| **Distribution** | Package manager install | Single static binary |
| **V2 deposit wallet** | Supported (with known bugs) | Supported, production-validated |
| **EOA signing** | Supported (produces ghost fills on V2) | **Blocked** — deposit-wallet only |
| **Version negotiation** | Hardcoded `CLOB_VERSION = "1"` → breaks on upgrades | Dynamic `/version` query before signing |
| **Credential security** | Auth headers leaked in error logs ([#327](https://github.com/Polymarket/clob-client/issues/327)) | Redacted in all output and logs |
| **Tick size caching** | In-memory per-instance, stale on update | Fresh fetch per order placement |
| **API key propagation** | 2-minute delay, no status polling | Derived on-demand with immediate use |
| **Local signing** | Optional (can use remote signers) | **Required** — key never leaves process |
| **External SDK in trust path** | Yes (Polymarket Python/TS SDKs) | No — all protocol code in this repo |
| **Go embedding** | Not possible | Native `pkg/` packages |
| **Read-only default** | No | Yes — credentials required explicitly |

The official clients are excellent for Python/TS ecosystems. They support EOA, proxy, and Safe signing modes for backward compatibility. Polygolem does not — it only implements the V2 deposit-wallet production path, because EOA orders on V2 produce ghost fills that appear in the book but never settle.

**Concrete issues we avoid:**

- **Hardcoded `CLOB_VERSION = "1"`** in `py-clob-client` caused mass `order_version_mismatch` failures when Polymarket upgraded their EIP-712 domain in April 2026. Polygolem queries `/version` dynamically before every signing session.
- **Auth headers leaked in error logs** (TypeScript client [#327](https://github.com/Polymarket/clob-client/issues/327)). Polygolem redacts all secrets in errors, logs, and JSON output — tested and enforced.
- **Tick size caching bugs** ([#265](https://github.com/Polymarket/clob-client/issues/265)) cause valid orders to be rejected because stale tick sizes are cached per client instance. Polygolem fetches tick sizes fresh per order placement.
- **No official Go client exists** — only scattered community efforts with varying completeness. Polygolem is a unified, production-validated Go-native SDK.

---

## Try It in 60 Seconds (No Credentials Needed)

All outputs below come from live Polymarket APIs.

```bash
git clone https://github.com/TrebuchetDynamics/polygolem
cd polygolem && go build -o polygolem ./cmd/polygolem

# Health check
./polygolem health
# {"clob":"ok","gamma":"ok"}

# Current BTC 5-minute up/down market
./polygolem discover crypto-5m --asset BTC
# Returns token IDs, condition ID, window boundaries

# Live order book
./polygolem orderbook price --token-id 1391568931...637394586
# {"price":"0.012","token_id":"1391568931...637394586"}

# All 7 active 5m crypto markets
./polygolem discover crypto-5m --enrich
# BTC, ETH, SOL, XRP, BNB, DOGE, HYPE with live prices
```

Read-only is the default for everything until you set `POLYMARKET_PRIVATE_KEY`.

---

## Production Validation

> **Production-validated:** Polygon mainnet · 2026-05-11 reference run
>
> [Every tx hash, gas figure, and pUSD movement](docs/LIVE-TRADE-WALKTHROUGH.md)
> is documented from EOA private key to filled buy + sell.

Core trading flows validated today:

- Headless V2 relayer onboarding (SIWE + profile + relayer key mint)
- Deposit-wallet deploy + funding
- CLOB V2 order signing, placement, and cancellation
- Advanced order types (FOK, GTD, post-only)
- Market discovery, streaming, and paper trading

---

## The V2 Identity Model

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

Your EOA signs; the deposit wallet holds funds and is the on-order maker;
Polymarket-run services pay every gas fee except your single ERC-20 funding
transfer. See [the walkthrough](docs/LIVE-TRADE-WALKTHROUGH.md) for the full
lifecycle with real txes.

---

## Trade in Four Commands

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

# One-command onboarding: auth + deploy + approve + fund
polygolem deposit-wallet onboard --fund-amount 0.71

# Sync CLOB balance
polygolem clob update-balance --asset-type collateral

# Place a market FOK buy
polygolem clob market-order \
  --token <ID> --side buy --amount 1 --price 0.012 --order-type FOK
# {
#   "success": true,
#   "orderID": "0x43083109...c423d793d",
#   "status": "matched",
#   "makingAmount": "1",
#   "takingAmount": "86.606666"
# }
```

After onboarding, every trade is fully headless. Total user-paid cost on the
reference run was **~$0.01 in POL gas** for the single ERC-20 transfer that
funds the deposit wallet.

---

## Safety Model

| Guard | What it does |
|---|---|
| **Read-only by default** | No credentials = no authenticated operations |
| **Deposit-wallet only** | Cannot accidentally sign as EOA, proxy, or Safe |
| **Local signing** | Private key never leaves the process |
| **No external SDKs** | All wallet derivation, EIP-712, ERC-7739, and relayer code is in this repo |
| **Pre-trade caps + daily limits + circuit breaker** | Configurable in `internal/risk` |
| **Secret redaction** | API keys and signatures are redacted in logs |

See [docs/SAFETY.md](docs/SAFETY.md) for the full model.

---

## Go SDK

Every CLI subcommand is a thin wrapper around importable `pkg/` packages:

| Package | What it does |
|---|---|
| [`pkg/universal`](pkg/universal) | One typed client over Gamma + CLOB + Data API + Stream + Discovery (70+ methods) |
| [`pkg/clob`](pkg/clob) | CLOB V2 — market data, orders, balances, builder fees |
| [`pkg/gamma`](pkg/gamma) | Read-only Gamma market discovery (26 methods) |
| [`pkg/stream`](pkg/stream) | Public CLOB WebSocket market stream |
| [`pkg/marketdata`](pkg/marketdata) | Live share-price snapshots from stream events |
| [`pkg/relayer`](pkg/relayer) | V2 Relayer client — WALLET-CREATE, batch, nonce |
| [`pkg/settlement`](pkg/settlement) | V2 winner redemption planning, adapter calls, readiness gates |
| [`pkg/marketresolver`](pkg/marketresolver) | Deterministic crypto window resolution (BTC/ETH/SOL/XRP/BNB/DOGE/HYPE) |

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/TrebuchetDynamics/polygolem/pkg/marketresolver"
    "github.com/TrebuchetDynamics/polygolem/pkg/universal"
)

ctx := context.Background()
client := universal.NewClient(universal.Config{})

// Resolve current BTC 5m window
resolver := marketresolver.NewResolver("")
result := resolver.ResolveTokenIDsForWindow(ctx, "BTC", "5m", time.Now().UTC())
// result.Status = "available"
// result.UpTokenID = "208311606920..."
// result.DownTokenID = "988679547673..."

price, _ := client.Price(ctx, result.UpTokenID, "buy")
spread, _ := client.Spread(ctx, result.UpTokenID)
fmt.Printf("BTC 5m YES — price %s, spread %s\n", price, spread)
```

Full package boundaries in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## Crypto Market Discovery

Polymarket runs 5-minute up/down markets for major crypto assets. Polygolem
discovers them deterministically — no search index lag:

```bash
# All 7 active 5m markets in one call
polygolem discover crypto-5m --enrich

# Specific window
polygolem discover crypto-window --asset BTC --interval 5m

# Paper trade the current window in one step
polygolem paper trade --asset BTC --interval 5m --side up --size 1
```

Assets supported: BTC, ETH, SOL, XRP, BNB, DOGE, HYPE.

---

## Common Workflows

| I want to... | Run |
|---|---|
| Find an active market | `polygolem discover search --query "..."` |
| List all 5m crypto markets | `polygolem discover crypto-5m` |
| Inspect the book | `polygolem clob book <token-id>` |
| Check deposit wallet status | `polygolem deposit-wallet status` |
| Place a limit buy | `polygolem clob create-order --token <ID> --side buy --price 0.5 --size 10` |
| Place a market FOK buy | `polygolem clob market-order --token <ID> --side buy --amount 1 --price <cap>` |
| Cancel everything | `polygolem clob cancel-all` |
| Read collateral balance | `polygolem clob balance --asset-type collateral` |
| Paper trade | `polygolem paper trade --asset BTC --interval 5m --side up` |

Full CLI reference: [docs/COMMANDS.md](docs/COMMANDS.md).

---

## Environment

- **Required for authenticated commands:** `POLYMARKET_PRIVATE_KEY`.
- **Automatic:** V2 relayer key is minted and persisted on first wallet use;
  CLOB L2 keys are created or derived on demand.
- **Optional:** `POLYMARKET_BUILDER_CODE` for V2 order attribution.

The deposit wallet address is derived locally from the private key; no API
call required. Full env reference in [docs/ONBOARDING.md](docs/ONBOARDING.md).

---

## Docs

| Document | What it covers |
|---|---|
| [Live Trade Walkthrough](docs/LIVE-TRADE-WALKTHROUGH.md) | End-to-end reference run: every tx, gas figure, and pUSD movement |
| [Onboarding](docs/ONBOARDING.md) | Complete deposit wallet flow, troubleshooting |
| [Headless Enable Trading](docs/ENABLE-TRADING-HEADLESS.md) | SDK for UI ClobAuth and token-approval signing |
| [Browser Fallback](docs/BROWSER-SETUP.md) | Manual signing when headless login is blocked |
| [Safety](docs/SAFETY.md) | Risk controls, deposit-wallet-only enforcement |
| [Contracts](docs/CONTRACTS.md) | Contract addresses, factory ABI, CREATE2 derivation |
| [Architecture](docs/ARCHITECTURE.md) | Package boundaries and dependency direction |
| [Commands](docs/COMMANDS.md) | Auto-generated CLI reference |
| [Deposit Wallet Migration](docs/DEPOSIT-WALLET-MIGRATION.md) | V1→V2 survival guide |
| [polygolem.trebuchetdynamics.com](https://polygolem.trebuchetdynamics.com) | Searchable docs site |

---

## Status

`v0.1.1` — production-validated against Polygon mainnet on **2026-05-11**
([reference run](docs/LIVE-TRADE-WALKTHROUGH.md)).

Core trading flows are production-validated today.

See [`CHANGELOG.md`](CHANGELOG.md) for per-version detail.
