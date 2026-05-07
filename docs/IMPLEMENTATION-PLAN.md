# Polygolem Implementation Plan

Status: draft
Date: 2026-05-06
Based on: PRD v2026-05-06, 8 reference repos, Current polygolem state

---

## 1. Gap Analysis: Current State vs PRD Requirements

### What Exists (polygolem today)

| Module | Status | Coverage |
|--------|--------|----------|
| `cmd/polygolem` | cobra shell, root command | bare scaffold |
| `internal/cli` | command wiring | thin |
| `internal/config` | viper config loading | done |
| `internal/modes` | execution mode enum | done |
| `internal/preflight` | gate checks skeleton | done |
| `internal/output` | JSON/table format | done |
| `internal/gamma/client.go` | 1 method (ActiveMarkets), 6-field Market | 5% of PRD R1 |
| `internal/clob/client.go` | 1 method (OrderBook), 2-field Level | 3% of PRD R2 |
| `internal/paper` | persisted paper state | exists |

### What's Missing (by PRD requirement)

| PRD Req | Status | Gap |
|---------|--------|-----|
| R1 Market Discovery | 5% | No search, events, tags, series, sports, pagination, enrichment, normalization. Market struct has 6 of 50+ Gamma fields. |
| R2 Public CLOB Data | 3% | Only OrderBook. Missing: markets list, price, midpoint, spread, tick size, fee rate, neg risk, prices history, last trade price, batch variants. No decimal types. |
| R3 Auth Model | 0% | No L0/L1/L2 model. No EIP-712 or HMAC. No signer abstractions. No builder attribution. |
| R4 Wallet Readiness | 0% | Nothing. |
| R5 Order Builder | 0% | Nothing. |
| R6 Order Execution | 0% | Nothing. |
| R7 Balances/Positions | 0% | Nothing. |
| R8 WebSocket/Streaming | 0% | Nothing. |
| R9 Paper Trading | partial | Exists but paper fills don't use CLOB data for pricing. |
| R10 Safety Gates | partial | Exists as skeleton, needs auth/balance/geoblock checks. |
| R11 Transport/Errors | 0% | No retries, rate limits, structured errors, redaction. |
| R12 SDK Boundary | partial | Internal packages exist, but surface is too narrow. |

### Dependency Map

```
PRD phases map to implementation order:

Phase A (R1, R2, R10, R11) → [NOW] Read-only SDK
    ├── requires: internal/gamma expanded, internal/clob expanded
    ├── requires: internal/transport (retry, errors, redaction)
    └── requires: internal/marketdiscovery (enrichment)

Phase B (R3, R4) → Auth & Readiness
    ├── requires: Phase A
    ├── requires: internal/auth (L1/L2, signer, HMAC, builder)
    └── requires: internal/wallet (derivation, readiness)

Phase C (R5, R6, R7, R9) → Orders & Paper
    ├── requires: Phase B (for signing)
    ├── requires: internal/orders (builder, types, lifecycle)
    ├── requires: internal/execution (paper/live interface)
    ├── requires: internal/account (balances, positions, rewards)
    └── requires: internal/risk (caps, limits, breakers)

Phase D (R8) → Streams
    ├── requires: Phase A (public market WS)
    └── requires: Phase B (user WS requires L2 auth)

Phase E (R6 live path) → Gated Live Execution
    ├── requires: Phase B + Phase C + Phase D
    └── requires: ALL safety gates passing
```

---

## 2. What to Steal from Each Open-Source Project

### vazic/polymarket_cli — CLI Structure & Read-Only Pattern
```
Steal: Cobra command structure, JSON-first output, one-file-per-command,
       thin command handlers, CGO_ENABLED=0 build pattern
Apply: cmd/search.go, cmd/market.go, cmd/orderbook.go — your command files
       should be this thin
```

### ivanzzeth/polymarket-go-gamma-client — Gamma API Types
```
Steal: Complete Market type (50+ fields), NormalizedTime, StringOrArray,
       GetMarketsParams query builder, Events/Series/Search/Tags API shapes,
       pointer-based optional params pattern
Apply: internal/gamma/types.go — copy the Market fields you need
       internal/gamma/params.go — copy the query param structs
       internal/gamma/normalize.go — copy NormalizedTime, StringOrArray logic
```

### ybina/polymarket-go — Package Separation
```
Steal: Clean package boundaries (clob/ vs ws/ vs relayer/ vs data/ vs gamma/),
       signer abstraction (PrivateKey vs Turnkey), L1/L2 header builders,
       order builder pipeline (build → sign → post), WebSocket callbacks
Apply: internal/auth/header.go — L1/L2 header struct shapes
       internal/orders/builder.go — the 3-step pipeline pattern
       Package boundary discipline (never mix streaming with REST)
```

### GoPolymarket/polymarket-go-sdk — Production Architecture
```
Steal: 5-layer model (App→Execution→Protocol→Security→Transport),
       OrderBuilder fluent API, error code taxonomy (AUTH-001, CLOB-001, etc.),
       WS dual-channel (market + user), reconnect policy config,
       subscription ref-counting, Stream[T] generic wrapper,
       IdempotencyKey format, 6-state lifecycle model,
       Rate limiter + circuit breaker transport layer
Apply: internal/errors/codes.go — error code taxonomy
       internal/transport/ — retry, circuit breaker, rate limiter design
       internal/orders/lifecycle.go — 6-state model
       internal/stream/refcount.go — subscription ref-counting
```

### Polymarket/rs-clob-client — Rust Type Design
```
Steal: OrderType enum (GTC/FOK/GTD/FAK + Unknown), Side enum (Buy=0/Sell=1),
       SignatureType (EOA=0/Proxy=1/Safe=2), Contract addresses per chain,
       amount types with USDC decimal awareness, feature flags as compile-time
Apply: internal/orders/types.go — OrderType, Side, SignatureType constants
       internal/wallet/addresses.go — CREATE2 derivation for proxy/safe
       internal/config/contracts.go — contract address table per chain
```

### Polymarket/go-builder-signing-sdk — Builder Auth
```
Steal: POLY_BUILDER_* header constants, HMAC generation flow,
       LocalSigner + RemoteSigner pattern, URL-safe base64 encoding
Apply: internal/auth/builder.go — complete builder header generation
```

### 0xNetuser/Polymarket-golang — Web3 & eth_call
```
Steal: eth_call auto-detection strategy (plain/EIP-1559/legacy) for Bor v2.6.0
       Gasless client relay payload fetching, batch redeem pattern
       (Phase E only — not needed now)
Apply: internal/web3/ — Phase E only
```

### HuakunShen/polymarket-kit — WebSocket Pool
```
Steal: RedundantWSPool design (N parallel connections → dedup),
       TTL-based message deduplication, CAS reconnect mutex,
       reference-counted dynamic subscriptions
Apply: internal/stream/pool.go — Phase D only
```

---

## 3. Phase A Implementation Plan: Read-Only SDK Foundation

### Target: Build the complete read-only API surface

This phase touches R1 (Market Discovery), R2 (Public CLOB Data), R10 (Safety Gates),
R11 (Transport/Errors), and R12 (SDK Boundary).

### 3.1 New Packages to Create

```
internal/
├── transport/           [NEW] HTTP transport, retries, errors, redaction
│   ├── client.go        — Transport struct: baseURL, http.Client, retry policy
│   ├── errors.go        — structured error types with codes
│   ├── retry.go         — exponential backoff for idempotent GETs
│   └── redact.go        — redaction helpers for secrets in logs/errors
│
├── polytypes/           [NEW] Shared Polymarket types (steal from gamma-client)
│   ├── market.go        — Market (50+ fields from Gamma), Event, Series, Tag
│   ├── clob.go          — OrderBook, Level, Price, TickSize, FeeRate
│   ├── decimal.go       — Decimal-safe type wrapper (string-backed, not float64)
│   └── normalize.go     — NormalizedTime, StringOrArray, outcome parser
│
├── gamma/               [EXPAND] Gamma REST client
│   ├── client.go        — KEEP existing, ADD 15+ methods
│   ├── params.go        — query param structs (GetMarketsParams, etc.)
│   ├── markets.go       — GetMarkets, GetMarketByID, GetMarketBySlug
│   ├── events.go        — GetEvents, GetEventByID, GetEventBySlug
│   ├── search.go        — Search (cross-entity)
│   ├── tags.go          — GetTags, GetTagByID, etc.
│   ├── series.go        — GetSeries, GetSeriesByID
│   └── sports.go        — GetTeams, GetSportsMetadata
│
├── clob/                [EXPAND] CLOB REST client
│   ├── client.go        — KEEP existing, ADD 15+ methods
│   ├── markets.go       — GetMarkets, GetMarket (by condition ID)
│   ├── orderbook.go     — GetOrderBook, GetOrderBooks (batch)
│   ├── prices.go        — GetPrice, GetPrices, GetMidpoint, GetMidpoints
│   ├── spreads.go       — GetSpread, GetSpreads (batch)
│   ├── ticks.go         — GetTickSize, GetTickSizes, GetFeeRateBps, GetNegRisk
│   └── history.go       — GetPricesHistory, GetLastTradePrice
│
├── marketdiscovery/     [NEW] Market enrichment service
│   ├── discovery.go     — joins Gamma metadata + CLOB details
│   └── enrichment.go    — EnrichedMarket type with merged fields
│
└── preflight/           [EXPAND] Safety gates
    └── checks.go        — ADD network health, gamma health, clob health checks
```

### 3.2 What Changes in Existing Packages

#### `internal/gamma/client.go` — Expand from 1 method to 20+

```go
// Keep existing patterns, add:
func (c *Client) Search(ctx context.Context, params *SearchParams) (*SearchResponse, error)
func (c *Client) Markets(ctx context.Context, params *GetMarketsParams) ([]polytypes.Market, error)
func (c *Client) MarketByID(ctx context.Context, id string) (*polytypes.Market, error)
func (c *Client) MarketBySlug(ctx context.Context, slug string) (*polytypes.Market, error)
func (c *Client) Events(ctx context.Context, params *GetEventsParams) ([]polytypes.Event, error)
func (c *Client) Tags(ctx context.Context, params *GetTagsParams) ([]polytypes.Tag, error)
// ... etc
```

**Steal from:** `polymarket-go-gamma-client` — copy all method signatures, all query param structs, all types. The methods are simple HTTP GETs with query params; the value is in the complete type system.

#### `internal/clob/client.go` — Expand from 1 method to 15+

```go
func (c *Client) GetMarkets(ctx context.Context, nextCursor string) (*polytypes.PaginatedMarkets, error)
func (c *Client) GetMarket(ctx context.Context, conditionID string) (*polytypes.CLOBMarket, error)
func (c *Client) GetOrderBooks(ctx context.Context, tokenIDs []string) ([]polytypes.OrderBook, error)
func (c *Client) GetPrice(ctx context.Context, tokenID, side string) (polytypes.Decimal, error)
func (c *Client) GetPrices(ctx context.Context, params []BookParams) (map[string]Price, error)
func (c *Client) GetMidpoint(ctx context.Context, tokenID string) (polytypes.Decimal, error)
func (c *Client) GetMidpoints(ctx context.Context, params []BookParams) (map[string]polytypes.Decimal, error)
func (c *Client) GetSpread(ctx context.Context, tokenID string) (polytypes.Decimal, error)
func (c *Client) GetTickSize(ctx context.Context, tokenID string) (polytypes.TickSize, error)
func (c *Client) GetFeeRateBps(ctx context.Context, tokenID string) (int, error)
func (c *Client) GetNegRisk(ctx context.Context, tokenID string) (bool, error)
func (c *Client) GetLastTradePrice(ctx context.Context, tokenID string) (polytypes.Decimal, error)
func (c *Client) GetPricesHistory(ctx context.Context, params *PriceHistoryParams) ([]polytypes.MarketPrice, error)
```

**Steal from:** `polymarket-go-sdk` — method signatures, batch param structs, pagination cursor pattern. Don't import the SDK, copy the API contract.

#### `internal/polytypes/` — New shared type package

Types to steal:

| Type | Source repo | Fields to copy |
|------|-------------|----------------|
| `Market` | polymarket-go-gamma-client | All 50+ fields |
| `Event` | polymarket-go-gamma-client | All fields including Markets[] |
| `OrderBook` | Existing polygolem + expand | Add Hash, Timestamp fields |
| `OrderBookLevel` | Existing polygolem | Already done |
| `TickSize` | polymarket-go-sdk | MinimumTickSize, MinimumOrderSize, TickSize |
| `Decimal` | New (string-backed) | Marshal/Unmarshal as string, avoid float64 |
| `NormalizedTime` | polymarket-go-gamma-client | Multi-format time parsing |
| `StringOrArray` | polymarket-go-gamma-client | JSON-encoded string arrays |
| `Side` | rs-clob-client | Buy=0, Sell=1 |
| `OrderType` | rs-clob-client | GTC, FOK, GTD, FAK + Unknown |
| `SignatureType` | rs-clob-client | EOA=0, Proxy=1, GnosisSafe=2 |

#### `internal/transport/` — New HTTP transport layer

```go
// Steal from polymarket-go-sdk transport design:
type Config struct {
    BaseURL         string
    Timeout         time.Duration
    UserAgent       string
    RetryMax        int           // default 3
    RetryBaseDelay  time.Duration // default 100ms
    RetryMaxDelay   time.Duration // default 2s
    RateLimitPerSec float64       // 0 = disabled
}

type Client struct {
    http   *http.Client
    config Config
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error
func (c *Client) Delete(ctx context.Context, path string) error
```

**Steal from:** `polymarket-go-sdk/pkg/transport` — retry with exponential backoff, rate limiter, error normalization.

#### `internal/errors/` — Structured error taxonomy

**Steal from:** `polymarket-go-sdk/pkg/errors` — complete error code taxonomy:

```
MARKET-001: Token ID required
MARKET-002: Invalid side value
CLOB-001: Order not found
CLOB-002: Insufficient balance
CLOB-003: Rate limit exceeded
NET-001: Connection refused
NET-002: Request timeout
AUTH-001: Missing signer
AUTH-002: Missing API credentials
AUTH-003: Invalid signature
SAFETY-001: Live trading disabled
SAFETY-002: Preflight failed
...
```

### 3.3 CLI Commands to Add (Phase A)

All commands remain read-only. No credentials needed.

```
polygolem discover
    search --query "election" --limit 20
    search --query "btc" --active --limit 10
    market --id "0x..."
    market --slug "will-btc-be-above"
    market enrich --id "0x..."     NEW: Gamma + CLOB merged view
    tags --active
    events --tag "crypto"

polygolem orderbook
    show --token-id "123..."       (already exists)
    batch --token-ids "123...,456..."
    price --token-id "123..." --side buy
    prices --token-ids "123...,456..."
    midpoint --token-id "123..."
    spread --token-id "123..."
    tick-size --token-id "123..."
    neg-risk --token-id "123..."
    fee-rate --token-id "123..."
    history --token-id "123..." --interval 1h

polygolem preflight
    health                          NEW: Gamma + CLOB health check
    status                          (already exists)
```

### 3.4 Implementation Order within Phase A

```
Step 1: internal/polytypes/          [1-2 days]
    Copy Market, Event, OrderBook types from gamma-client + rs-clob-client.
    Implement Decimal (string-backed). Implement NormalizedTime.

Step 2: internal/transport/ + internal/errors/  [1 day]
    Build retry transport, error codes.

Step 3: internal/gamma/ expansion   [2-3 days]
    Add all 20+ methods using transport client.
    Wire polytypes throughout.

Step 4: internal/clob/ expansion    [2-3 days]
    Add all 15+ methods.
    Wire polytypes.

Step 5: internal/marketdiscovery/   [1 day]
    Enrichment: Gamma market → fetch CLOB tick/fee/negrisk → merged type.

Step 6: cmd/ commands               [2 days]
    Wire new commands. Keep handlers thin (one-liner delegates to SDK).

Step 7: Tests                       [2-3 days]
    Mock HTTP tests for gamma + clob.
    Golden JSON tests for CLI output.
```

---

## 4. Phase B Outline: Auth & Readiness

### When Phase A is complete

```
internal/auth/
├── auth.go           — L0/L1/L2 mode enum, SignatureType enum
├── signer.go         — Signer interface, PrivateKeySigner
├── eip712.go         — BuildClobEip712Signature (steal from ybina)
├── l1.go             — CreateL1Headers (steal from polymarket-go)
├── l2.go             — CreateL2Headers, HMAC signing (steal from polymarket-go)
├── builder.go        — BuilderConfig, POLY_BUILDER_* headers (steal from go-builder-signing-sdk)
├── keys.go           — API key management (create, derive, list, delete wrappers)
└── redact.go         — Redaction: never log private keys or passphrases

internal/wallet/
├── wallet.go         — Wallet readiness: chain ID, signer address, funder
├── derive.go         — Proxy wallet derivation (CREATE2), Safe derivation
├── checks.go         — Close-only status, geoblock, API health, chain consistency
└── status.go         — ReadinessReport type

CLI additions:
polygolem auth
    status              — Show readiness without secrets
    derive-key          — L1: derive API key (requires private key)
    create-key          — L1: create new API key
    list-keys           — L2: list existing keys
    delete-key          — L2: delete key

polygolem wallet
    status              — Address, chain, signature type, funder
    derive-proxy        — Compute proxy wallet address
    derive-safe         — Compute Safe wallet address
```

**Steal from:** `ybina/polymarket-go/tools/headers`, `go-builder-signing-sdk`, `Polymarket/rs-clob-client` (contract addresses)

---

## 5. Phase C Outline: Order Domain & Paper Executor

### When Phase B provides signer infrastructure

```
internal/orders/
├── types.go          — OrderIntent, SignedOrder, OrderResponse, Trade
├── builder.go        — OrderBuilder (steal fluent pattern from rs-clob-client)
├── sign.go           — SignOrder (EIP-712 struct hash → signer.Sign)
├── lifecycle.go      — 6-state model (created→accepted→partial→filled/canceled/rejected)
├── validate.go       — Tick size, price range, fee rate, neg risk, batch size
└── fixtures_test.go  — Deterministic signing tests with fixed salts

internal/execution/
├── executor.go       — Executor interface (Place, Cancel, Query)
├── paper.go          — PaperExecutor: local-only simulation
├── live.go           — LiveExecutor: gated, unavailable in Phase C
└── idempotency.go    — Canonical key: tenant;strategy;client_order_id

internal/account/
├── balances.go       — BalanceAllowance (USDC + conditional tokens)
├── positions.go      — Current/closed positions from Data API
├── rewards.go        — User earnings, reward percentages, rewards markets
└── dataapi.go        — Data API client (positions, trades, analytics)

internal/risk/
├── caps.go           — Per-trade amount caps, open-order limits
├── limits.go         — Daily loss limits, consecutive error thresholds
├── breaker.go        — Circuit breaker: halt on violations
└── policy.go         — RiskPolicy configuration

CLI additions:
polygolem order
    build --token-id "..." --side buy --price 0.55 --size 10
    build --market-order --token-id "..." --side sell --amount 100

polygolem paper
    buy --token-id "..." --price 0.55 --size 10
    sell --token-id "..." --price 0.45 --size 5
    positions
    pnl
    history

polygolem account
    balance (Phase E: requires L2 auth)
    positions
    rewards
```

**Steal from:**
- OrderBuilder: `polymarket-go-sdk/pkg/clob/order_builder.go` (fluent pattern)
- Lifecycle: `polymarket-go-sdk/pkg/execution/lifecycle.go` (6 states)
- Balance/positions: `ybina/polymarket-go/client/data/` (Data API shapes)
- Risk: `taetaehoho/poly-kalshi-arb` (circuit breaker design from Rust)

---

## 6. Phase D Outline: Streams

### When read-only + auth foundations are stable

```
internal/stream/
├── config.go         — WebSocket config: URL, reconnect, heartbeat
├── market.go         — Public market WS client (orderbook, prices, midpoint, etc.)
├── user.go           — Authenticated user WS client (orders, trades)
├── refcount.go       — Subscription reference counting
├── lifecycle.go      — Connect, reconnect, ping/pong, shutdown, context cancel
├── events.go         — Event type dispatch: book, price_change, last_trade_price, etc.
├── dedup.go          — TTL-based message deduplication (steal from polymarket-kit)
└── pool.go           — RedundantWSPool (steal from polymarket-kit)

CLI additions:
polygolem stream
    market --token-ids "123...,456..."
    user --markets "0x...,0x..."        (requires L2 auth)
```

**Steal from:**
- Dual-channel design: `polymarket-go-sdk/pkg/clob/ws/`
- RedundantWSPool + dedup: `HuakunShen/polymarket-kit/go-client/client/`
- Event type parsing: `polymarket-kit/types/websocket.go` (7 event types)
- Reconnect policy: `polymarket-go-sdk/pkg/clob/ws/config.go`

---

## 7. Phase E Outline: Gated Live Execution

### Requires separate approved plan. NOT implemented until all gates pass.

---

## 8. Package Dependency Tree

```
                    cmd/polygolem (cobra, thin handlers)
                           │
              ┌────────────┼────────────┐
              │            │            │
    marketdiscovery     orders       stream
         │              │   │           │
    ┌────┴────┐    ┌────┴───┴───┐   ┌──┴──┐
  gamma     clob  auth  execution paper ws
    │        │     │      │        │    │
    └────────┴─────┴──────┴────────┴────┘
                    │
              ┌─────┴─────┐
          transport     polytypes
                    │
              errors (no deps, used by all)
```

### Dependency rules:
- `polytypes` depends on nothing (leaf package)
- `errors` depends on nothing
- `transport` depends on `errors`
- `gamma`, `clob` depend on `transport` + `polytypes` + `errors`
- `marketdiscovery` depends on `gamma` + `clob` + `polytypes`
- `auth` depends on `transport` + `errors` (optional: go-ethereum for EIP-712 in Phase B)
- `orders` depends on `auth` + `polytypes` + `errors`
- `execution` depends on `orders` + `auth` + `transport`
- `paper` depends on `execution` + `polytypes`
- `stream` depends on `transport` + `polytypes` + `auth`

### No circular dependencies. All internal packages.

---

## 9. Go Module Dependencies Plan

### Current (Phase A start)
```
github.com/spf13/cobra v1.9.1
github.com/spf13/viper v1.21.0
```

### Phase A additions (minimal)
```
No new deps. All HTTP via stdlib net/http.
```

### Phase B (auth) — careful dependency choice
```
Option A: Add go-ethereum only for EIP-712       [heavy: 100+ transitive deps]
Option B: Implement EIP-712 signing with stdlib   [light: ~200 lines of crypto code]
Option C: Use external signer binary (polymarket-cli) [zero Go deps, subprocess overhead]
```

Recommendation: **Option B** — EIP-712 domain separator, struct hash, and secp256k1 signing
can be done with Go stdlib `crypto/ecdsa`, `crypto/sha3`, and basic RLP encoding.
The `polymarket-go-sdk` already does this with `go-ethereum`'s `signer.SignTypedData`,
but the algorithm itself is well-defined and implementable in ~200 lines.

### Phase D (streams)
```
github.com/gorilla/websocket   [1 dep, standard choice]
```
Or use stdlib `nhooyr.io/websocket` for context-native API.

---

## 10. Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Decimal type | Custom `polytypes.Decimal` (string-backed) | Avoid float64 precision issues. MVP: store as string, parse with `math/big.Rat`. |
| EIP-712 signing | Implement with stdlib crypto | Avoid go-ethereum dependency. ~200 lines. Already solved in `polymarket-go`. |
| WebSocket library | gorilla/websocket | Industry standard, 1 transitive dep. |
| Transport retry | Build custom, 30 lines | Exponential backoff on GET requests only. POST never retried. |
| Error codes | String taxonomy (CLOB-001, etc.) | Machine-readable, stable, copied from polymarket-go-sdk. |
| Market enrichment | New `marketdiscovery` package | Separation of concerns: Gamma and CLOB clients don't know about each other. |
| Paper state storage | JSON file per session | Simple now, swappable to SQLite later via `storage` interface. |
| CLI command handlers | 1-3 lines max | Delegates immediately to SDK service. No logic in cmd/. |

---

## 11. Test Strategy (per PRD Testing Decisions)

### Phase A Tests

```
internal/gamma/client_test.go
    - Mock HTTP: search returns markets, market by ID, events, tags
    - Error: 404, 500, malformed JSON, Gamma quirk fields
    - Pagination: offset queries, empty results

internal/clob/client_test.go
    - Mock HTTP: orderbook, price, midpoint, spread, tick size, neg risk, fee rate
    - Batch: 5 token IDs produce 5 results
    - Error: invalid token ID format, 429 rate limit

internal/polytypes/normalize_test.go
    - NormalizedTime: "2020-11-02T16:31:01Z", "2020-11-02 16:31:01+00", "January 2, 2006"
    - StringOrArray: `["Yes","No"]`, `"[\"Yes\",\"No\"]"`, `[["A","B"]]`

internal/transport/client_test.go
    - Retry: transient 500s, backoff timing, max retries
    - Redaction: private key, passphrase, bearer token never in output

internal/marketdiscovery/discovery_test.go
    - Enrichment: Gamma market → CLOB tick/fee/negrisk → merged EnrichedMarket

tests/cli_output_test.go
    - Golden JSON: `polygolem discover search --query "btc"` output
    - Golden JSON: `polygolem orderbook show --token-id "123"` output
```

### Phase B Tests
```
- Auth: L1 header fixture, L2 HMAC header fixture, redaction
- Wallet: proxy/safe derivation against known CREATE2 addresses
- Preflight: missing key → structured error, geoblock → blocked
```

### Phase C Tests
```
- Order builder: deterministic signing with fixed salt/timestamp
- Order validation: invalid tick size, price out of range, missing neg risk
- Paper executor: buy/sell produce correct paper state, no real HTTP calls
```

---

## 12. File Count Estimate

| Phase | New Files | Modified Files | Total Lines (est.) |
|-------|-----------|----------------|---------------------|
| Phase A | ~25 | ~5 | ~3000 |
| Phase B | ~15 | ~3 | ~2000 |
| Phase C | ~20 | ~5 | ~3500 |
| Phase D | ~10 | ~2 | ~2000 |
| TOTAL | ~70 | ~15 | ~10500 |

---

## 13. Next Immediate Action

**Start Phase A, Step 1: `internal/polytypes/`**

1. Create `internal/polytypes/market.go` — copy the complete Market type from `polymarket-go-gamma-client/types.go`
2. Create `internal/polytypes/clob.go` — OrderBook, Level, TickSize, FeeRate types
3. Create `internal/polytypes/decimal.go` — string-backed Decimal with MarshalJSON/UnmarshalJSON
4. Create `internal/polytypes/normalize.go` — NormalizedTime, StringOrArray
5. Run `go test ./internal/polytypes/` — should compile, zero external deps
