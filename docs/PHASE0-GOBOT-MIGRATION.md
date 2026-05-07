# Phase 0 — Go-Bot Boundary Cleanup Plan

Date: 2026-05-06
Status: inventory complete, migration pending

## Inventory

7 files in `go-bot/internal/polymarket/` contain direct Polymarket protocol access:

| File | Lines | Violation |
|------|-------|-----------|
| `gamma_client.go` | 351 | Constructs Gamma URLs, HTTP clients, parses upstream responses |
| `clob_client.go` | 233 | Constructs CLOB URLs, HTTP clients, parses order books and trades |
| `types.go` | 89 | Domain types (OK to keep, but should reference polygolem types) |
| `gamma_resolved.go` | ~200 | Resolved market parsing |
| `gamma_client_test.go` | ~100 | Tests against real Gamma API shapes |
| `clob_client_test.go` | ~100 | Tests against real CLOB API shapes |
| `gamma_resolved_test.go` | ~80 | Tests for resolved markets |

## Migration Map: go-bot → polygolem

| go-bot usage | Replace with polygolem |
|---|---|
| `NewGammaClient(baseURL, httpClient, now)` | `internal/gamma.NewClient(gammaBaseURL, transportClient)` |
| `GammaClient.ActiveBTCMarkets()` | `internal/marketdiscovery.Service.EnrichedMarkets()` |
| `GammaClient.ActiveCryptoMarkets()` | `internal/gamma.Client.Markets()` + tag filtering |
| `NewCLOBClient(baseURL, httpClient, now)` | `internal/clob.NewClient(clobBaseURL, transportClient)` |
| `CLOBClient.OrderBook(tokenID)` | `internal/clob.Client.OrderBook(ctx, tokenID)` |
| `CLOBClient.RecentTrades(marketID)` | `internal/clob.Client.LastTradesPrices()` or Data API |
| `Market` domain struct | `internal/polytypes.Market` (100+ fields) |
| `OrderBook` domain struct | `internal/polytypes.OrderBook` |
| `Trade` domain struct | `internal/polytypes.Trade` |
| `MarketSource` interface | `internal/marketdiscovery.Service` |
| `OrderBookSource` interface | `internal/clob.Client` |
| `TradeSource` interface | `internal/dataapi.Client` |

## R13 Go-Bot Consumer Interfaces

Polygolem exposes these interfaces for go-bot consumption:

```go
// MarketDiscovery — market search and listing
type MarketDiscovery interface {
    Search(ctx context.Context, query string, limit int) ([]polytypes.EnrichedMarket, error)
    ActiveMarkets(ctx context.Context, tag string, limit int) ([]polytypes.Market, error)
    MarketByID(ctx context.Context, id string) (*polytypes.Market, error)
    MarketBySlug(ctx context.Context, slug string) (*polytypes.Market, error)
}

// BookReader — CLOB order book and price data
type BookReader interface {
    OrderBook(ctx context.Context, tokenID string) (*polytypes.OrderBook, error)
    Price(ctx context.Context, tokenID, side string) (string, error)
    Midpoint(ctx context.Context, tokenID string) (string, error)
    Spread(ctx context.Context, tokenID string) (string, error)
    TickSize(ctx context.Context, tokenID string) (*polytypes.TickSize, error)
    FeeRateBps(ctx context.Context, tokenID string) (int, error)
    NegRisk(ctx context.Context, tokenID string) (bool, error)
}

// AccountReader — account state (requires L2 auth in Phase E)
type AccountReader interface {
    Positions(ctx context.Context, user string) ([]Position, error)
    TotalValue(ctx context.Context, user string) (*TotalValue, error)
}

// OrderExecutor — order lifecycle (paper now, live Phase E)
type OrderExecutor interface {
    Place(ctx context.Context, intent *orders.OrderIntent) (*orders.OrderResponse, error)
    Cancel(ctx context.Context, orderID string) error
    List(ctx context.Context) ([]orders.OrderResponse, error)
}

// StreamSubscriber — WebSocket streams (Phase D)
type StreamSubscriber interface {
    SubscribeOrderBook(ctx context.Context, assetIDs []string, callback func(BookMessage)) error
    Close() error
}
```

## Migration Steps

### Step 1: Add polygolem adapter in go-bot
```
go-bot/internal/polygolem/
├── adapter.go       — wraps polygolem internal packages
├── interfaces.go    — go-bot-specific interfaces (above)
└── fake.go          — test doubles returning fixtures
```

### Step 2: Wire paper mode through adapter
Replace direct CLOB calls in paper mode with polygolem's `BookReader`.

### Step 3: Wire market discovery through adapter
Replace `ActiveBTCMarkets()` with polygolem's `MarketDiscovery`.

### Step 4: Mark old types as deprecated
Add deprecation comments to `go-bot/internal/polymarket/types.go`.

### Step 5: Remove direct protocol clients
After all consumers use the adapter, delete `gamma_client.go` and `clob_client.go`.

### Step 6: Add repo guard
CI check: no new `clob.polymarket.com`, `gamma-api.polymarket.com`, or `data-api.polymarket.com` in go-bot outside polygolem adapter.

## Acceptance Criteria (from PRD R13)

- [ ] go-bot has no direct references to `clob.polymarket.com`, `gamma-api.polymarket.com`
- [ ] go-bot has no `NewCLOBClient`, `NewGammaClient` outside polygolem adapter
- [ ] Paper-mode run fetches order books through polygolem
- [ ] Unit tests mock polygolem interfaces, not Polymarket HTTP endpoints
- [ ] CI guard prevents new direct Polymarket protocol access
