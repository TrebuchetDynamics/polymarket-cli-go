# polygolem

[![CI](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml/badge.svg)](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TrebuchetDynamics/polygolem)](go.mod)
[![Latest Release](https://img.shields.io/github/v/tag/TrebuchetDynamics/polygolem?label=release&sort=semver)](https://github.com/TrebuchetDynamics/polygolem/releases)

**Production-safe Polymarket infrastructure for the deposit-wallet era.**

A single Go binary and SDK for trading on Polymarket V2. Built specifically
for the deposit-wallet (POLY_1271 / type 3) signing path — the only path that
works for new production API users since the April 2026 migration. No external
SDKs, no Python runtime, no opaque signing wrappers. Every byte of wallet
derivation, EIP-712 payload, ERC-7739 envelope, and relayer call is
implemented directly from spec in Go.

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

Verified against Polygon mainnet on the 2026-05-08 reference run
([full walkthrough with every tx hash, gas figure, and pUSD movement](docs/LIVE-TRADE-WALKTHROUGH.md)):

- Headless V2 relayer onboarding (SIWE → `/profiles` → `/relayer-auth`)
- Deposit-wallet derivation, deployment, and approvals (relayer-sponsored, $0 user gas)
- POL → pUSD swap via Uniswap V3 multihop (no L2 bridge required)
- pUSD funding (EOA → deposit wallet ERC-20 transfer)
- CLOB V2 limit and market orders with post-only / GTC / GTD / FOK
- Builder-code attribution (V2 bytes32 model)
- Cancels (single, batch, all, per-market)
- Public market discovery via Gamma + CLOB
- Public WebSocket market stream
- Local risk controls (per-trade caps, daily loss limits, circuit breaker)

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

## Trade in Four Commands

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

polygolem auth headless-onboard                     # mint V2 relayer key (gasless)
polygolem deposit-wallet onboard --fund-amount 0.71 # deploy + approve + fund
polygolem clob update-balance --asset-type collateral
polygolem clob create-order --token <ID> --side buy --price 0.5 --size 10
```

**Total user-paid cost: ~$0.01 in POL gas** for the single ERC-20 transfer
that funds the deposit wallet. WALLET-CREATE, the 6-call approval batch, and
every CLOB settlement are sponsored by Polymarket-run services. See
[the walkthrough](docs/LIVE-TRADE-WALKTHROUGH.md) for the per-tx gas
breakdown.

> **One-time browser step for new users.** Polymarket's L1 auth endpoint
> (`/auth/api-key`) does not currently support ERC-1271 validation, so a
> brand-new EOA needs one browser login at polymarket.com to mint the
> deposit-wallet-bound CLOB API key. After that, everything is headless. See
> [docs/BROWSER-SETUP.md](docs/BROWSER-SETUP.md). Existing Polymarket users
> with an already-minted CLOB key skip this entirely.

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
import "github.com/TrebuchetDynamics/polygolem/pkg/universal"

c := universal.NewClient(universal.Config{})
price, _ := c.OrderbookPrice(ctx, tokenID)
fmt.Println(price)
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

| Variable | When required |
|---|---|
| `POLYMARKET_PRIVATE_KEY` | Any authenticated command. |
| `POLYMARKET_RELAYER_API_KEY` / `_ADDRESS` | Auto-minted by `auth headless-onboard`. |
| `POLYMARKET_CLOB_API_KEY` / `_SECRET` / `_PASSPHRASE` | One-time browser-minted for new users; persisted afterwards. |
| `POLYMARKET_BUILDER_CODE` | Optional V2 order attribution. |

The deposit wallet address is derived locally from the private key — no API
call required.

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

`v0.1.0` — Full deposit-wallet lifecycle, CLOB V2 trading, builder-code
attribution, universal SDK client. See [`CHANGELOG.md`](CHANGELOG.md).
