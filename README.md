# polygolem

Safe Polymarket SDK and CLI for Go. Zero external PolyMarket SDK dependencies — all types and patterns stolen from the ecosystem's best open-source projects.

```
Polymarket APIs → polygolem SDK → go-bot adapter → bot strategies
                                 → CLI commands → operator terminal
```

## Why

The Polymarket ecosystem has 8+ Go/Rust SDKs and CLI tools. None combine: read-only safety by default, paper trading with live market data, hard gated live execution, and a public SDK boundary for downstream bots.

Polygolem is the single source of truth for Polymarket protocol access in the megabot stack. `go-bot` must not construct its own CLOB clients, Gamma URLs, or auth headers — it goes through polygolem.

## Current Scope

**33 Go files across 20 packages.** All built, all tests passing.

```
internal/
├── transport/          HTTP client with retry, redaction, rate limiter, circuit breaker
├── errors/             Structured error taxonomy (NET-xxx, CLOB-xxx, AUTH-xxx)
├── polytypes/          Complete Gamma + CLOB types (100+ field Market, Decimal, NormalizedTime)
├── gamma/              Gamma API client — 18 methods (markets, events, search, tags, series)
├── clob/               CLOB API client — 17 methods (orderbook, price, midpoint, tick, fee, history)
├── dataapi/            Data API client — 11 methods (positions, volume, leaderboards)
├── marketdiscovery/    Gamma + CLOB enrichment service
├── auth/               L0/L1/L2 model, EIP-712 signing, HMAC, builder attribution, signer
├── wallet/             CREATE2 proxy/Safe address derivation, readiness checks
├── orders/             OrderIntent, fluent builder, validation, lifecycle states, amount math
├── execution/          Executor interface, PaperExecutor (local-only, no network)
├── risk/               Per-trade caps, daily loss limits, circuit breaker
├── stream/             WebSocket market client with reconnect + message deduplication
├── paper/              Local paper state persistence
├── preflight/          Safety gate checks
├── modes/              Execution mode enums
├── config/             Configuration loading
├── output/             JSON + table output
└── cli/                Cobra command wiring (thin handlers only)
```

### Public SDK (for go-bot)

```
pkg/
├── bookreader/         OrderBook reader backed by polygolem CLOB client
├── marketresolver/     Active market + token ID resolution from Gamma
├── bridge/             Bridge API — supported assets, deposit addresses, quotes
├── pagination/         Cursor and offset pagination helpers with batching
└── SKILL.md            AI agent integration (Claude Code)
```

## Quick Start

```bash
cd go-bot/polygolem
go build -o polygolem ./cmd/polygolem
```

### Read-only CLI (no credentials)

```bash
./polygolem discover search --query "btc 5m" --limit 10
./polygolem discover market --id "0x..."
./polygolem discover enrich --id "0x..."
./polygolem orderbook get --token-id "123..."
./polygolem orderbook price --token-id "123..." 
./polygolem orderbook midpoint --token-id "123..."
./polygolem orderbook spread --token-id "123..."
./polygolem orderbook tick-size --token-id "123..."
./polygolem orderbook fee-rate --token-id "123..."
./polygolem health
./polygolem preflight
./polygolem version
```

### As Go SDK dependency (for go-bot)

```go
import polybook "github.com/TrebuchetDynamics/polygolem/pkg/bookreader"
import polyresolver "github.com/TrebuchetDynamics/polygolem/pkg/marketresolver"
import polybridge "github.com/TrebuchetDynamics/polygolem/pkg/bridge"
import polypagination "github.com/TrebuchetDynamics/polygolem/pkg/pagination"
```

## Architecture

```
Transport Layer
  └── HTTP retry, rate limiter, circuit breaker, redaction

Protocol Layer
  ├── gamma/      (Gamma API, read-only, no auth)
  ├── clob/       (CLOB API, public + authenticated)
  ├── dataapi/    (Data API, read-only analytics)
  ├── stream/     (WebSocket, reconnect, dedup)
  └── bridge/     (Bridge API, deposits, quotes)

Domain Layer
  ├── polytypes/        (shared types)
  ├── marketdiscovery/  (Gamma + CLOB enrichment)
  ├── orders/           (intent, builder, lifecycle)
  ├── execution/        (paper executor, future live)
  ├── auth/             (L0/L1/L2, signer, HMAC)
  ├── wallet/           (derivation, readiness)
  └── risk/             (caps, limits, breaker)

Application Layer
  ├── cli/              (Cobra commands)
  └── pkg/              (public SDK boundary)
```

## Safety Model

| Mode | Credentials | Can Sign | Can Post | Can Mutate |
|------|-------------|----------|----------|------------|
| Read-only | None | No | No | No |
| Paper | None | No | No (local only) | No |
| Live (future) | Private key + API key | Yes | Yes | Gated |

Live execution requires: `POLYMARKET_LIVE_PROFILE=on`, `live_trading_enabled: true`, `--confirm-live`, successful `preflight`. No silent downgrade to paper/read-only.

## Dependencies

```
github.com/spf13/cobra          CLI routing
github.com/spf13/viper          Config loading
github.com/ethereum/go-ethereum ECDSA/secp256k1 (auth only)
github.com/gorilla/websocket    WebSocket (stream only)
golang.org/x/crypto             keccak256
```

No external Polymarket SDKs. All types stolen from reference repos, not vendored.

## Test

```bash
go test ./... -count=1
```

10 test packages, all passing. Mock HTTP tests for gamma, clob, output, preflight.

## Status

| Phase | Description | Status |
|-------|-------------|--------|
| 0 | Go-bot boundary cleanup | CLOB book fixed, market resolver live |
| A | Read-only SDK foundation | 18 gamma + 17 clob + 11 dataapi methods |
| B | Auth & readiness | EIP-712, HMAC, builder, wallet derivation |
| C | Orders & paper executor | Fluent builder, lifecycle, paper executor |
| D | Streams | Market WS + dedup |
| E | Gated live execution | Blocked — requires separate approved plan |

## Steal Log

Every pattern in this codebase was stolen from the best open-source Polymarket projects:

| Pattern | Source |
|---------|--------|
| Gamma types (100+ fields, NormalizedTime, StringOrArray) | polymarket-go-gamma-client |
| CLOB client API surface, error taxonomy, transport | polymarket-go-sdk |
| Auth model, EIP-712, HMAC, signer, builder | ybina/polymarket-go, go-builder-signing-sdk |
| Fluent order builder | rs-clob-client |
| WebSocket client, message dedup | polymarket-kit |
| CLI structure, JSON output, SKILL.md | vazic/polymarket_cli |
| Bridge API types | ybina/polymarket-go |
| Circuit breaker, rate limiter | polymarket-go-sdk |
| CREATE2 wallet derivation | polymarket-go, rs-clob-client |
| Lifecycle states, idempotency keys | polymarket-go-sdk |
| eth_call auto-detection | 0xNetuser/Polymarket-golang |

## Documentation

- [PRD](docs/PRD.md)
- [Implementation Plan](docs/IMPLEMENTATION-PLAN.md)
- [Phase 0 Go-Bot Migration](docs/PHASE0-GOBOT-MIGRATION.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Safety](docs/SAFETY.md)
