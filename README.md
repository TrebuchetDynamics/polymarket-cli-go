# polygolem

[![CI](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml/badge.svg)](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TrebuchetDynamics/polygolem)](go.mod)
[![Latest Release](https://img.shields.io/github/v/tag/TrebuchetDynamics/polygolem?label=release&sort=semver)](https://github.com/TrebuchetDynamics/polygolem/releases)

Safe Polymarket SDK and CLI for Go — **deposit wallet (type 3 / POLY_1271) only.**
Read-only by default. No external SDKs. All types, signing, and protocol logic
implemented from spec. CLOB V2 with version-gated order signing, ERC-1967 CREATE2
wallet derivation (verified against official Python SDK), relayer client, and
full CLI for the deposit wallet lifecycle.

**EOA, proxy, and Gnosis Safe are not supported.** Polymarket CLOB V2 requires
deposit wallet for new API users. Polygolem is built exclusively for type 3
(POLY_1271) — the only mode that works on current production.

## One Env Var. Everything Else Auto-Generated.

**`POLYMARKET_PRIVATE_KEY` is the only environment variable.**
Builder credentials, CLOB L2 keys — polygolem generates and persists them all.
No browser, no polymarket.com, no copy-paste.

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

# 1. Builder credentials — auto-generated, stored independently
polygolem builder auto

# 2. Full deposit wallet onboarding (builder creds auto-loaded)
polygolem deposit-wallet onboard --fund-amount 0.71 --json

# 3. Sync and trade
polygolem clob update-balance --asset-type collateral
polygolem clob create-order --token ID --side buy --price 0.5 --size 10
```

**Total cost: ~$0.01 POL for one funding transfer. Everything else is sponsored.**

under the hood, `builder auto` signs a local EIP-712 ClobAuth message and posts it to the CLOB API — this lazy-creates a builder profile, issues `BUILDER_API_KEY` + `BUILDER_SECRET` + `BUILDER_PASSPHRASE`, and assigns a `bytes32` builder code. the creds are persisted to an env file that downstream commands (`deposit-wallet`, `clob`) pick up automatically.

## Install

```bash
git clone https://github.com/TrebuchetDynamics/polygolem
cd polygolem && go build -o polygolem ./cmd/polygolem
```

## Command Inventory

### Public — No Credentials Needed

```bash
polygolem discover search --query "btc 5m" --limit 5
polygolem discover market --id "0xbd31dc8..."
polygolem discover enrich --id "0xbd31dc8..."
polygolem orderbook get --token-id "123..."
polygolem orderbook price --token-id "123..."
polygolem orderbook spread --token-id "123..."
polygolem clob book <token-id>
polygolem clob markets --cursor ""
polygolem clob market <condition-id>
polygolem health
polygolem version
```

### Builder + Deposit Wallet

```bash
polygolem builder auto                                    # programmatic, no browser
polygolem deposit-wallet derive                          # predict CREATE2 address (local)
polygolem deposit-wallet deploy --wait                   # WALLET-CREATE via relayer
polygolem deposit-wallet status                          # deployed? approved? funded?
polygolem deposit-wallet approve --submit                # sign + submit 6-call batch
polygolem deposit-wallet fund --amount 0.71              # ERC-20 transfer EOA→wallet
polygolem deposit-wallet onboard --fund-amount 0.71      # deploy + approve + fund
```

### CLOB Trading

```bash
polygolem clob balance --asset-type collateral
polygolem clob update-balance --asset-type collateral
polygolem clob create-order --token ID --side buy --price 0.5 --size 10
polygolem clob create-order --token ID --side buy --price 0.5 --size 10 --order-type GTD --expiration 1778125000
polygolem clob market-order --token ID --side buy --amount 5
polygolem clob orders                                      # list open orders
polygolem clob trades                                      # trade history
polygolem clob cancel <order-id>                           # cancel single order
polygolem clob cancel-all                                  # cancel all orders
polygolem clob create-api-key
```

### Bridge

```bash
polygolem bridge assets
polygolem bridge deposit <wallet-address>
```

### Paper — Local Simulation

```bash
polygolem paper buy
polygolem paper positions
polygolem paper reset
```

## Go SDK — `pkg/`

| Package | What It Does |
|---------|-------------|
| `pkg/universal` | Single client for Gamma + CLOB + Data API + Discovery + Stream (70+ methods) |
| `pkg/gamma` | Read-only Gamma API — 26 methods returning public `pkg/types` DTOs |
| `pkg/data` | Read-only Data API analytics client using public DTOs |
| `pkg/types` | Shared public DTOs for SDK packages |
| `pkg/clob` | Read-only CLOB market-data client returning public `pkg/types` DTOs |
| `pkg/stream` | Read-only public CLOB WebSocket market stream client |
| `pkg/orderbook` | Read-only CLOB order book reader |
| `pkg/bookreader` | Deprecated compatibility wrapper for `pkg/orderbook` |
| `pkg/marketresolver` | Market + token ID resolution |
| `pkg/bridge` | Bridge API — supported assets, deposit addresses, quotes |
| `pkg/relayer` | Builder relayer primitives for wallet create and wallet batch flows |
| `pkg/pagination` | Cursor and offset pagination with concurrent batching |

## Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/clob` | CLOB v2 client — 37 methods, EIP-712, POLY_1271, ERC-7739 |
| `internal/gamma` | Gamma API client — 26 methods |
| `internal/dataapi` | Data API — positions, volume, leaderboards |
| `internal/relayer` | Builder relayer — WALLET-CREATE, WALLET batch, nonce |
| `internal/auth` | L0/L1/L2 auth, CREATE2 derivation, builder attribution |
| `internal/stream` | WebSocket market stream implementation behind `pkg/stream` |
| `internal/transport` | HTTP retry, rate limiter, circuit breaker, redaction |
| `internal/orders` | OrderIntent, fluent builder, validation |
| `internal/wallet` | Deposit wallet primitives — derive, deploy, status |
| `internal/risk` | Per-trade caps, daily loss limits, circuit breaker |

## Env Vars

| Variable | Required |
|----------|----------|
| `POLYMARKET_PRIVATE_KEY` | All authenticated operations |

That's it. Builder credentials are auto-generated by `polygolem builder auto` and stored independently — you never touch them. CLOB L2 keys are derived on first use. The deposit wallet address is computed locally from your key.

## Status

`v0.1.0` — Full deposit wallet lifecycle, CLOB v2 trading, builder auto,
universal SDK client. See [`CHANGELOG.md`](CHANGELOG.md).

```bash
go test ./...  # 27/28 packages pass (rpc intentionally untested)
```

## Docs

| Document | What It Covers |
|----------|---------------|
| [Builder Auto](docs/BUILDER-AUTO.md) | Zero-browser onboarding — full sequence diagram, costs, Reown flow |
| [Deposit Wallet Deployment](docs/DEPOSIT-WALLET-DEPLOYMENT.md) | Full pipeline — derive, deploy, approve, fund, requirements checklist |
| [Contracts](docs/CONTRACTS.md) | All smart contract addresses, factory ABI, CREATE2, permission model |
| [Deposit Wallet Migration](docs/DEPOSIT-WALLET-MIGRATION.md) | Bot killer survival guide, V1→V2 migration |
| [Builder Credential Issuance](docs/BUILDER-CREDENTIAL-ISSUANCE.md) | Superseded by BUILDER-AUTO.md — builder auto is programmatic |
| [Polydart PRD](PRD_POLYDART.md) | Companion Dart SDK for Flutter / Arenaton |
| [Safety](docs/SAFETY.md) | Read-only default, deposit wallet safety rules |
| [Commands](docs/COMMANDS.md) | Full command reference |
| [Architecture](docs/ARCHITECTURE.md) | Package boundaries and dependency direction |
| [Astro Docs](https://trebuchetdynamics.github.io/polygolem) | Full documentation site with guides, concepts, and API reference |
