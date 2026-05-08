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

## Quick Start

**Existing Polymarket users:** Fully headless — see [docs/ONBOARDING.md](docs/ONBOARDING.md).

**New users:** One-time browser login required — see [docs/BROWSER-SETUP.md](docs/BROWSER-SETUP.md).

```bash
export POLYMARKET_PRIVATE_KEY="0x..."

# Full onboarding (existing users: fully headless; new users: browser step required)
polygolem auth headless-onboard                          # V2 relayer key
polygolem deposit-wallet onboard --fund-amount 0.71      # deploy + approve + fund
polygolem clob update-balance --asset-type collateral
polygolem clob create-order --token <ID> --side buy --price 0.5 --size 10
```

**Total cost: ~$0.01 POL for one funding transfer. Everything else is sponsored.**

> **New user limitation:** Pure headless onboarding is impossible for new deposit
> wallet users. Polymarket's L1 auth endpoint lacks ERC-1271 support. After one
> browser login, all trading is fully headless. See
> [docs/ONBOARDING.md](docs/ONBOARDING.md) for the full flow and
> [docs/BROWSER-SETUP.md](docs/BROWSER-SETUP.md) for the browser signup guide.

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
polygolem builder auto                                    # CLOB L2 credentials
polygolem auth headless-onboard                          # V2 relayer API key
polygolem clob create-builder-fee-key                    # V2 order attribution
polygolem deposit-wallet derive                          # predict CREATE2 address (local)
polygolem deposit-wallet deploy --wait                   # WALLET-CREATE via relayer
polygolem deposit-wallet status                          # deployed? approved? funded?
polygolem deposit-wallet approve --submit                # sign + submit 6-call batch
polygolem deposit-wallet fund --amount 0.71              # ERC-20 transfer EOA→wallet
polygolem deposit-wallet onboard --fund-amount 0.71      # deploy + approve + fund
polygolem clob create-api-key-for-address --owner 0xDepositWallet  # deposit-owned CLOB key
```

### CLOB Trading

```bash
polygolem clob balance --asset-type collateral
polygolem clob update-balance --asset-type collateral
polygolem clob create-api-key                             # EOA/bootstrap key
polygolem clob create-api-key-for-address --owner 0xDepositWallet
polygolem clob create-order --token ID --side buy --price 0.5 --size 10 --builder-code "$POLYMARKET_BUILDER_CODE"
polygolem clob create-order --token ID --side buy --price 0.5 --size 10 --post-only
polygolem clob create-order --token ID --side buy --price 0.5 --size 10 --order-type GTD --expiration 1778125000
polygolem clob batch-orders --orders-file orders.json --builder-code "$POLYMARKET_BUILDER_CODE"
polygolem clob market-order --token ID --side buy --amount 5 --builder-code "$POLYMARKET_BUILDER_CODE"
polygolem clob heartbeat --id keepalive-1
polygolem clob orders                                      # list open orders
polygolem clob trades                                      # trade history
polygolem clob cancel <order-id>                           # cancel single order
polygolem clob cancel-all                                  # cancel all orders
polygolem clob create-builder-fee-key
polygolem clob list-builder-fee-keys
polygolem clob revoke-builder-fee-key --key "$POLYMARKET_BUILDER_CODE"
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
| `pkg/clob` | CLOB market-data plus authenticated account/order DTOs |
| `pkg/stream` | Read-only public CLOB WebSocket market stream client |
| `pkg/orderbook` | Read-only CLOB order book reader |
| `pkg/bookreader` | Deprecated compatibility wrapper for `pkg/orderbook` |
| `pkg/marketresolver` | Market + token ID resolution |
| `pkg/builder` | Builder HMAC signer helpers for local or remote signing |
| `pkg/bridge` | Bridge API — supported assets, deposit addresses, quotes |
| `pkg/relayer` | Builder relayer primitives for wallet create and wallet batch flows |
| `pkg/ctf` | Conditional Tokens calldata and ID helpers |
| `pkg/pagination` | Cursor and offset pagination with concurrent batching |

## Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/clob` | CLOB v2 client — market data, auth, EIP-712, POLY_1271, ERC-7739 |
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
| `RELAYER_API_KEY` / `RELAYER_API_KEY_ADDRESS` | Deposit-wallet deploy and approval batches |
| `POLYMARKET_BUILDER_CODE` | Optional V2 order attribution |

CLOB L2 keys are created or derived on first use. V2 relayer keys are minted by
`polygolem auth headless-onboard`. The deposit wallet address is computed
locally from your key.

## Status

`v0.1.0` — Full deposit wallet lifecycle, CLOB v2 trading, builder auto,
universal SDK client. See [`CHANGELOG.md`](CHANGELOG.md).

```bash
go test ./...  # 27/28 packages pass (rpc intentionally untested)
```

## Docs

| Document | What It Covers | Status |
|----------|---------------|--------|
| [Onboarding](docs/ONBOARDING.md) | **Single source of truth** — complete deposit wallet flow, headless vs. browser, troubleshooting | **Canonical** |
| [Browser Setup](docs/BROWSER-SETUP.md) | One-time browser login for new users. Security guidance. Hardware wallet / WalletConnect support. | **Canonical** |
| [Commands](docs/COMMANDS.md) | Auto-generated CLI reference — every command and flag. | **Auto-generated** |
| [Safety](docs/SAFETY.md) | Read-only default, deposit wallet safety, risk breaker. | **Canonical** |
| [Contracts](docs/CONTRACTS.md) | Smart contract addresses, factory ABI, CREATE2 derivation. | **Canonical** |
| [Architecture](docs/ARCHITECTURE.md) | Package boundaries and dependency direction. | **Canonical** |
| [Deposit Wallet Migration](docs/DEPOSIT-WALLET-MIGRATION.md) | Bot killer survival guide, V1→V2 migration. | **Canonical** |
| [Polydart PRD](PRD_POLYDART.md) | Companion Dart SDK for Flutter / Arenaton. | **Canonical** |
| [Astro Docs](https://trebuchetdynamics.github.io/polygolem) | Full documentation site with guides, concepts, and API reference. | **Canonical** |
