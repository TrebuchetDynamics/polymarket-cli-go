# polygolem

Safe Polymarket SDK and CLI for Go. Read-only by default. No external SDKs —
all types, signing, and protocol logic implemented from spec.

**Solves the May 2026 Polymarket deposit wallet migration** — the breaking change
that rejected EOA orders for new API users. Full tooling: derive, deploy via
relayer, WALLET batch signing, approvals, on-chain funding, and POLY_1271 order
placement.

## One Command to Onboard a New Polymarket Account

```bash
# Creates builder profile (free at polymarket.com/settings?tab=builder), then:
POLYMARKET_PRIVATE_KEY="0x..." \
POLYMARKET_BUILDER_API_KEY="..." \
POLYMARKET_BUILDER_SECRET="..." \
POLYMARKET_BUILDER_PASSPHRASE="..." \
  polygolem deposit-wallet onboard --fund-amount 0.71 --json
```

That's it. Deploys the wallet, approves 6 trading contracts via WALLET batch, and
transfers 0.71 pUSD from EOA. Then sync:

```bash
POLYMARKET_PRIVATE_KEY="0x..." \
  polygolem clob update-balance --asset-type collateral --signature-type deposit
```

Ready to trade. See [Deposit Wallet Migration Guide](docs/DEPOSIT-WALLET-MIGRATION.md).

## How It Works (Two Tiers)

**Tier 1 — Your own bot** (you hold the key):
Every command is a standalone CLI tool. Deploy wallets, sign batches, place
POLY_1271 orders directly from your infrastructure.

**Tier 2 — Your users' app** (they hold their keys, you hold one builder account):
Arenaton's server calls polygolem as a subprocess. Users never see a CLI.
One builder account deploys wallets for all users. Builder creds are in the
server environment, never exposed to Flutter.

| User action | Server calls |
|-------------|-------------|
| Enable trading | `deposit-wallet onboard --fund-amount X` |
| Check status | `deposit-wallet status` |
| Place order | `clob create-order --signature-type deposit` |
| Check balance | `clob balance --signature-type deposit` |

## Install

```bash
git clone https://github.com/TrebuchetDynamics/polygolem
cd polygolem && go build -o polygolem ./cmd/polygolem
```

## Command Inventory

### Public (no credentials needed)
```bash
polygolem discover search --query "btc 5m" --limit 5
polygolem discover market --id "0xbd31dc8..."
polygolem discover enrich --id "0xbd31dc8..."
polygolem orderbook get --token-id "123..."
polygolem orderbook price --token-id "123..."
polygolem orderbook spread --token-id "123..."
polygolem health
polygolem version
```

### Deposit Wallet (builder creds required for deploy/batch/onboard)
```bash
polygolem deposit-wallet derive                    # no creds needed
polygolem deposit-wallet deploy --wait             # needs builder
polygolem deposit-wallet nonce                     # needs builder
polygolem deposit-wallet status                    # needs builder
polygolem deposit-wallet batch --calls-json '...'  # EIP-712 sign + submit
polygolem deposit-wallet approve                   # review calldata
polygolem deposit-wallet approve --submit          # sign + send 6-call batch
polygolem deposit-wallet fund --amount 0.71        # ERC-20 transfer EOA→wallet
polygolem deposit-wallet onboard --fund-amount 0.71 # deploy + approve + fund
```

### CLOB Trading (private key required)
```bash
polygolem clob book <token-id>
polygolem clob balance --asset-type collateral --signature-type deposit
polygolem clob update-balance --asset-type collateral --signature-type deposit
polygolem clob create-order --token ID --side buy --price 0.5 --size 10 --signature-type deposit
polygolem clob market-order --token ID --side buy --amount 5 --signature-type deposit
polygolem clob orders
polygolem clob trades
polygolem clob create-api-key
```

### Bridge
```bash
polygolem bridge assets
polygolem bridge deposit <wallet-address>
```

### Paper (local simulation)
```bash
polygolem paper buy
polygolem paper positions
polygolem paper reset
```

## Packages

| Package | What it does |
|---------|-------------|
| `internal/relayer` | Builder relayer client — WALLET-CREATE, WALLET batch, nonce, polling |
| `internal/rpc` | Direct on-chain transfers (ERC-20 pUSD from EOA) |
| `internal/clob` | CLOB API client — 17 methods + EIP-712 + POLY_1271 + ERC-7739 signing |
| `internal/auth` | L0/L1/L2 auth, EIP-712, deposit wallet CREATE2 derivation, builder attribution |
| `internal/gamma` | Gamma API client — 18 methods (markets, events, search, tags, series) |
| `internal/dataapi` | Data API client — 11 methods (positions, volume, leaderboards) |
| `internal/orders` | OrderIntent, fluent builder, validation, lifecycle states |
| `internal/execution` | PaperExecutor (local-only), future live executor |
| `internal/stream` | WebSocket market client with reconnect + dedup |
| `internal/risk` | Per-trade caps, daily loss limits, circuit breaker |
| `internal/transport` | HTTP retry, rate limiter, circuit breaker, redaction |
| `pkg/bookreader` | Public OrderBook reader for go-bot |
| `pkg/marketresolver` | Public market + token ID resolution |
| `pkg/bridge` | Public Bridge API — supported assets, deposit addresses, quotes |
| `pkg/pagination` | Cursor and offset pagination with concurrent batching |

## Status

| Phase | Status |
|-------|--------|
| Phase 0 — Go-bot boundary cleanup | ✅ Done |
| Phase A — Read-only SDK foundation | ✅ Done |
| Phase B — Auth & readiness | ✅ Done |
| Phase C — Orders & paper executor | ✅ Done |
| Phase D — Streams | ✅ Done |
| Phase E — Gated live execution | ✅ Done |
| Deposit wallet migration (May 2026) | ✅ Done |

```bash
go test ./...   # 25+ packages, all passing
```

## Env Vars

| Variable | Required for |
|----------|-------------|
| `POLYMARKET_PRIVATE_KEY` | All authenticated commands |
| `POLYMARKET_BUILDER_API_KEY` | Deposit wallet deploy/batch/onboard |
| `POLYMARKET_BUILDER_SECRET` | Deposit wallet deploy/batch/onboard |
| `POLYMARKET_BUILDER_PASSPHRASE` | Deposit wallet deploy/batch/onboard |
| `POLYMARKET_RELAYER_URL` | Override relayer URL (default: relayer-v2.polymarket.com) |

Short-form `BUILDER_API_KEY` / `BUILDER_SECRET` / `BUILDER_PASS_PHRASE` also accepted.

## Docs

- [Deposit Wallet Migration Guide](docs/DEPOSIT-WALLET-MIGRATION.md) — bot killer survival guide
- [PRD](docs/PRD.md) — full requirements
- [Safety](docs/SAFETY.md) — read-only default, deposit wallet safety rules
- [Commands](docs/COMMANDS.md) — full command reference
- [Architecture](docs/ARCHITECTURE.md) — package boundaries
