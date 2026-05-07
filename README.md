# polygolem

Safe Polymarket SDK and CLI for Go. Read-only by default, no credentials needed.
All types stolen from the ecosystem's best open-source projects — no external Polymarket SDKs.

## Install

```bash
go install github.com/TrebuchetDynamics/polygolem/cmd/polygolem@latest
```

Or build from source:
```bash
git clone https://github.com/TrebuchetDynamics/polygolem
cd polygolem && go build -o polygolem ./cmd/polygolem
```

## What can you do with it?

### 1. Search markets and check odds — no wallet, no API key

```bash
# Find active BTC markets
polygolem discover search --query "btc 5m" --limit 5

# Get details for a specific market
polygolem discover market --id "0xbd31dc8..."

# Get everything at once — Gamma metadata + CLOB tick size, fee, orderbook
polygolem discover enrich --id "0xbd31dc8..."
```

Output is always JSON:
```json
{
  "market": {"question": "Bitcoin up in 5 minutes?", "lastTradePrice": 0.52, ...},
  "tick_size": {"minimum_tick_size": "0.01", "minimum_order_size": "5"},
  "neg_risk": false,
  "fee_rate_bps": 0,
  "order_book": {"bids": [...], "asks": [...]}
}
```

### 2. Read order books and prices in real time

```bash
# Get L2 order book depth
polygolem orderbook get --token-id "123456789..."

# Check best bid/ask, midpoint, spread, tick size
polygolem orderbook price --token-id "123..."
polygolem orderbook midpoint --token-id "123..."
polygolem orderbook spread --token-id "123..."
polygolem orderbook tick-size --token-id "123..."
polygolem orderbook fee-rate --token-id "123..."
```

### 3. Check API health

```bash
polygolem health
# {"gamma": "ok", "clob": "ok"}
```

### 4. Bridge: check supported assets and get deposit quotes

```bash
# No CLI yet — use as Go library:
```

```go
import "github.com/TrebuchetDynamics/polygolem/pkg/bridge"

bridge := bridge.NewClient("", nil)
assets, _ := bridge.GetSupportedAssets(ctx)
// [{ChainID: "137", ChainName: "Polygon", Token: {Symbol: "POL", Decimals: 18}, ...}]

quote, _ := bridge.GetQuote(ctx, bridge.QuoteRequest{
    FromAmountBaseUnit: "1000000000000000000", // 1 POL
    FromChainID: "137", FromTokenAddress: "0x...",
    RecipientAddress: "0xYourWallet", ToChainID: "137",
    ToTokenAddress: "0xUSDC...",
})
// {EstOutputUsd: 0.23, EstFeeBreakdown: {GasUsd: 0.02, ...}}
```

### 5. Use as a Go SDK in your own bot

```go
import (
    "github.com/TrebuchetDynamics/polygolem/pkg/bookreader"
    "github.com/TrebuchetDynamics/polygolem/pkg/marketresolver"
    "github.com/TrebuchetDynamics/polygolem/pkg/pagination"
)

// Resolve active token IDs for BTC 5m markets
resolver := marketresolver.NewResolver("")
result := resolver.ResolveTokenIDs(ctx, "BTC", "5m")
// {Status: "available", UpTokenID: "...", DownTokenID: "...", ConditionID: "0x..."}

// Fetch order books through polygolem (never construct your own CLOB client)
reader := bookreader.NewReader("")
book, _ := reader.OrderBook(ctx, result.UpTokenID)
// {Bids: [{Price: 0.48, Size: 100}, ...], Asks: [{Price: 0.52, Size: 50}, ...]}

// Build orders with the fluent builder
intent, _ := orders.NewBuilder(result.UpTokenID, polytypes.SideBuy).
    Price("0.49").Size("10").
    TickSize("0.01").FeeRateBps(0).
    Build()

// Auto-paginate through all markets
all, _ := pagination.CollectAll(ctx, func(ctx context.Context, cursor string) ([]CLOBMarket, string, error) {
    resp, _ := clob.Markets(ctx, cursor)
    return resp.Data, resp.NextCursor, nil
})
```

### 6. Paper trade locally against real market data

```go
executor := execution.NewPaperExecutor("1000") // $1000 starting cash
resp, _ := executor.Place(ctx, &intent)
// {Success: true, OrderID: "paper-1", Status: "matched"}
```

## Safety

| Mode | Credentials | Can Sign | Can Post | Can Mutate |
|------|-------------|----------|----------|------------|
| Read-only | None | No | No | No |
| Paper | None | No | No (local only) | No |
| Live (future) | Private key + API key | Yes | Yes | Gated |

Read-only is the default. No API keys, no wallet, no risk. Live execution is hard-disabled until all gates pass: `POLYMARKET_LIVE_PROFILE=on`, `--confirm-live`, successful preflight.

## Packages

| Package | What it does |
|---------|-------------|
| `internal/gamma` | Gamma API client — 18 methods (markets, events, search, tags, series) |
| `internal/clob` | CLOB API client — 17 methods (orderbook, price, midpoint, tick, fee, history) |
| `internal/dataapi` | Data API client — 11 methods (positions, volume, leaderboards) |
| `internal/auth` | L0/L1/L2 auth model, EIP-712 signing, HMAC, builder attribution |
| `internal/wallet` | CREATE2 proxy/Safe address derivation |
| `internal/orders` | OrderIntent, fluent builder, validation, lifecycle states |
| `internal/execution` | PaperExecutor (local-only), future live executor |
| `internal/stream` | WebSocket market client with reconnect + dedup |
| `internal/risk` | Per-trade caps, daily loss limits, circuit breaker |
| `internal/transport` | HTTP retry, rate limiter, circuit breaker, redaction |
| `pkg/bookreader` | Public OrderBook reader for go-bot |
| `pkg/marketresolver` | Public market + token ID resolution |
| `pkg/bridge` | Public Bridge API — supported assets, deposit addresses, quotes |
| `pkg/pagination` | Cursor and offset pagination with concurrent batching |

## Dependencies

```
github.com/spf13/cobra          CLI
github.com/spf13/viper          Config
github.com/ethereum/go-ethereum Secp256k1 (auth only)
github.com/gorilla/websocket    WebSocket (stream only)
golang.org/x/crypto             Keccak256
```

## Status

| Phase | Status |
|-------|--------|
| 0 — Go-bot boundary cleanup | Done |
| A — Read-only SDK foundation | Done |
| B — Auth & readiness | Done |
| C — Orders & paper executor | Done |
| D — Streams | Done |
| E — Gated live execution | Blocked (requires separate plan) |

```bash
go test ./... -count=1   # 10 packages, all passing
```

## Docs

- [PRD](docs/PRD.md) — full requirements
- [Implementation Plan](docs/IMPLEMENTATION-PLAN.md) — architecture decisions
- [Phase 0 Migration](docs/PHASE0-GOBOT-MIGRATION.md) — go-bot integration
