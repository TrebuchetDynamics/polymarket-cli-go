# Polymarket CLI & SDK Ecosystem — Go & Rust

Research date: 2026-05-07
Scope: Curated Go and Rust projects useful for building a PolyMarket mega-bot API.
Updated: Added deposit wallet (type 3 / POLY_1271) repos. Marked V1-only repos as deprecated.

---

## Deposit Wallet (Type 3 / POLY_1271) Support

Polymarket's CLOB V2 requires deposit wallet (signature type 3) for new API users.
These repos implement the full deposit wallet flow: CREATE2 derivation, WALLET-CREATE
relayer deployment, ERC-1271 signature verification, and WALLET batch operations.

### Official Polymarket implementations

| Repo | Language | Deposit Wallet |
|------|----------|---------------|
| [Polymarket/clob-client-v2](https://github.com/Polymarket/clob-client-v2) | TypeScript | ✅ Full — POLY_1271 signing, V2 order struct |
| [Polymarket/py-clob-client-v2](https://github.com/Polymarket/py-clob-client-v2) | Python | ✅ Full — POLY_1271, V2 domain |
| [Polymarket/rs-clob-client-v2](https://github.com/Polymarket/rs-clob-client-v2) | Rust | ✅ Full — SignatureType::Poly1271 |
| [Polymarket/builder-relayer-client](https://github.com/Polymarket/builder-relayer-client) | TypeScript | ✅ Full — WALLET-CREATE, deposit wallet factory |
| [Polymarket/py-builder-relayer-client](https://github.com/Polymarket/py-builder-relayer-client) | Python | ✅ Full — derive_deposit_wallet, WALLET batch |
| [GoPolymarket/go-builder-relayer-client](https://pkg.go.dev/github.com/GoPolymarket/go-builder-relayer-client) | Go | ✅ Full — Go relayer with deposit wallet |

Key contracts (verified on Polygonscan):
```
DepositWalletFactory:        0x00000000000Fb5C9ADea0298D729A0CB3823Cc07
DepositWalletImplementation: 0x58CA52ebe0DadfdF531Cde7062e76746de4Db1eB
```
The implementation is a Polymarket-proprietary `DepositWallet` contract (Solidity 0.8.34),
NOT a Gnosis Safe. It supports `isValidSignature` (ERC-1271), batch `execute`, owner
management, and session signers.

### Community repos with deposit wallet

| Repo | Language | Deposit Wallet |
|------|----------|---------------|
| [qualiaenjoyer/polymarket-apis](https://github.com/qualiaenjoyer/polymarket-apis) | Python | ✅ derive_deposit_wallet, complete relayer flow |
| [tdergouzi/rs-clob-client-v2](https://github.com/tdergouzi/rs-clob-client-v2) | Rust | ✅ Fork of official V2 with deposit wallet |

### polygolem's deposit wallet implementation

Polygolem is the **only known Go implementation** that ships the full deposit wallet
lifecycle without depending on an external SDK:

- `internal/auth/signer.go` — CREATE2 derivation (verified against official Python test vector)
- `internal/relayer/` — WALLET-CREATE, WALLET batch, nonce management
- `internal/clob/orders.go` — V2 order signing with version-gated dispatch
- `internal/cli/deposit_wallet.go` — Full CLI: derive, deploy, fund, approve, onboard

### POLY_1271 protocol details (from official Polymarket contracts)

Signature type 3 (`POLY_1271`) is defined in [Polymarket/ctf-exchange-v2](https://github.com/Polymarket/ctf-exchange-v2/blob/main/src/exchange/libraries/Structs.sol):

```solidity
enum SignatureType { EOA=0, POLY_PROXY=1, POLY_GNOSIS_SAFE=2, POLY_1271=3 }
```

**Validation flow** (from `Signatures.sol`):
```solidity
function verifyPoly1271Signature(address signer, address maker, bytes32 hash, bytes memory signature)
    internal view returns (bool) {
    return (signer == maker) && maker.code.length > 0
        && SignatureCheckerLib.isValidSignatureNow(maker, hash, signature);
}
```

**V2 EIP-712 Order type** (from official Python SDK):
```python
ORDER_TYPE_STRING = (
    "Order(uint256 salt,address maker,address signer,uint256 tokenId,"
    "uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,"
    "uint256 timestamp,bytes32 metadata,bytes32 builder)"
)
ORDER_TYPEHASH = 0xbb86318a2138f5fa8ae32fbe8e659f8fcf13cc6ae4014a707893055433818589
```

**ERC-7739 TypedDataSign wrapping** (required for type 3):
The SDK wraps the EOA's ECDSA signature in an ERC-7739 `TypedDataSign` envelope:
```
signature || appDomainSep || contentsHash || typeString
```
The inner domain for TypedDataSign uses the deposit wallet address as `verifyingContract`.
This produces a ~636-byte signature that the CLOB passes to `depositWallet.isValidSignature()`.

The Soladity type:
```
TypedDataSign(Order contents,string name,string version,uint256 chainId,
              address verifyingContract,bytes32 salt)
```

**Key requirements:**
- `signer` == `maker` == deposit wallet address ✅ (fixed in our code)
- Deposit wallet must be deployed on-chain (`maker.code.length > 0`)
- Signature must be ERC-7739 wrapped (not plain ECDSA)
- Outer domain: `{name: "Polymarket CTF Exchange", version: "2"}`
- Inner TypedDataSign domain: `{name: "DepositWallet", version: "1", verifyingContract: depositWallet}`

### Go repos with CLOB V2 support (April-May 2026)

| Repo | Status |
|------|--------|
| [nijaru/go-clob-client](https://github.com/nijaru/go-clob-client) | Updated April 2026 |
| [splicemood/polymarket-go-sdk](https://github.com/splicemood/polymarket-go-sdk) | Fork with V2 support |
| [GoPolymarket/polymarket-go-sdk](https://github.com/GoPolymarket/polymarket-go-sdk) | v1.1.3 (April 28 2026) |

None of these have POLY_1271 / ERC-7739 support confirmed — polygolem remains the leading Go
implementation for the deposit wallet flow.

---

## Deprecated — V1-only repos (no CLOB V2 or deposit wallet support)

---

## Official Polymarket CLI

| Repo | Language | Stars | License | Why it matters |
|---|---|---|---|---|
| [Polymarket/polymarket-cli](https://github.com/Polymarket/polymarket-cli) | Rust | 2540 | — | **Primary CLI gateway.** Browse markets, place orders, manage positions, onchain contracts. Outputs JSON for scripts/agents. Has interactive REPL. `brew install polymarket`. |

Key features:
- `polymarket search --query "..."` — Gamma API search
- `polymarket market --id "0x..."` — Market pricing/details
- `polymarket orderbook --token-id "..."` — L2 depth
- `polymarket order buy/sell` — CLOB authenticated orders
- Config at `~/.config/polymarket/config.json`
- Wallet resolution, RPC provider, CLOB auth

---

## Go SDKs & Clients

## Deprecated — V1-only repos (no CLOB V2 or deposit wallet support)

These repos predate the April 28 2026 CLOB V2 cutover. They sign V1 orders
that are no longer accepted by production CLOB. Useful for architecture
reference only — do not use for live trading.

### Tier A — Production-grade, now deprecated

| Repo | Stars | Status |
|---|---|---|
| [GoPolymarket/polymarket-go-sdk](https://github.com/GoPolymarket/polymarket-go-sdk) | 47 | ⚠️ V1-only |
| [0xNetuser/Polymarket-golang](https://github.com/0xNetuser/Polymarket-golang) | 72 | ⚠️ V1-only |
| [Polymarket/go-builder-signing-sdk](https://github.com/Polymarket/go-builder-signing-sdk) | 3 | ⚠️ HMAC-only, no builderCode field |
| [HuakunShen/polymarket-kit](https://github.com/HuakunShen/polymarket-kit) | 55 | ⚠️ V1-only |
| [ybina/polymarket-go](https://github.com/ybina/polymarket-go) | 20 | ⚠️ V1-only |

### Tier B — Specialized, now deprecated

| Repo | Stars | Status |
|---|---|---|
| [ivanzzeth/polymarket-go-gamma-client](https://github.com/ivanzzeth/polymarket-go-gamma-client) | 30 | ⚠️ Read-only Gamma, still usable |
| [ivanzzeth/polymarket-go-real-time-data-client](https://github.com/ivanzzeth/polymarket-go-real-time-data-client) | 13 | ⚠️ Read-only WS, still usable |
| [D8-X/polymarket-trader-go-sdk](https://github.com/D8-X/polymarket-trader-go-sdk) | 0 | ⚠️ V1-only |
| [aszxqaz/pmclient](https://github.com/aszxqaz/pmclient) | — | ⚠️ V1-only |
| [lajosdeme/polymarket-go-api](https://github.com/lajosdeme/polymarket-go-api) | — | ⚠️ V1-only |

### Tier C — Reference, now deprecated

| Repo | Stars | Status |
|---|---|---|
| [CalderWhite/polymarket-gamma-go](https://github.com/CalderWhite/polymarket-gamma-go) | 11 | ⚠️ Read-only Gamma, still usable |
| [bububa/polymarket-client](https://github.com/bububa/polymarket-client) | 0 | ⚠️ V1-only |
| [monsterdev914/polymarket-trading-bot](https://github.com/monsterdev914/polymarket-trading-bot) | 27 | ⚠️ V1-only |
| [arjunprakash027/Mantis](https://github.com/arjunprakash027/Mantis) | 12 | ⚠️ V1-only |
| [vazic/polymarket_cli](https://github.com/vazic/polymarket_cli) | 0 | ⚠️ V1-only |

---

## Rust SDKs & Clients

### Official

| Repo | Stars | License | What | Key Features |
|---|---|---|---|---|
| [Polymarket/rs-clob-client](https://github.com/Polymarket/rs-clob-client) | 646 | MIT | **Official** Rust CLOB client | Typed CLOB requests, dual auth flows, alloy support, order builders, serde, async-first with reqwest |
| [Polymarket/rs-clob-client-v2](https://github.com/Polymarket/rs-clob-client-v2) | 10 | MIT | **Official** Rust CLOB V2 client | Uses v2 endpoints (`clob-v2.polymarket.com`), modular features, MSRV 1.88 |

Feature flags (common to both official clients):

| Feature | Description |
|---|---|
| `clob` | Core CLOB client for order placement, market data, auth |
| `ws` | WebSocket client for real-time orderbook, price, user events |
| `rtds` | Real-time data streams (Binance, Chainlink crypto prices, comments) |
| `data` | Data API client (positions, trades, leaderboards, analytics) |
| `gamma` | Gamma API client (market/event discovery, search, metadata) |
| `bridge` | Bridge API client (cross-chain deposits: EVM, Solana, Bitcoin) |
| `rfq` | RFQ API (submit/query quotes) |
| `heartbeats` | Auto heartbeat; disconnect cancels all open orders |
| `ctf` | CTF API client (split/merge/redeem on binary & neg risk markets) |
| `tracing` | Structured logging via `tracing` |

### Community / Third-party

| Repo | Stars | What |
|---|---|---|
| [tdergouzi/rs-clob-client](https://github.com/tdergouzi/rs-clob-client) | — | Rust port of TS `@polymarket/clob-client`. Full EIP-712 signing. |

---

## Rust Trading Bots & Strategy Repos

These are the most directly applicable for strategy design and execution patterns in a mega-bot.

| Repo | Stars | Strategy | Key Features |
|---|---|---|---|
| [PolybaseX/Polymarket-Trading-Bot-Rust](https://github.com/PolybaseX/Polymarket-Trading-Bot-Rust) | 95 | Dual Limit Same-Size + 5m BTC + Trailing | Limit buys at $0.45, hedge if only one fills, trailing stop, backtest, simulation mode |
| [PolyScripts/polymarket-5min-15min-1hr-btc-arbitrage-trading-bot-rust](https://github.com/PolyScripts/polymarket-5min-15min-1hr-btc-arbitrage-trading-bot-rust) | 68 | BTC 5m/15m arbitrage | 20ms order placement, 50 checks/sec, market-neutral dual-leg, DRY_RUN mode |
| [Sectionnaenumerate/Polymarket-Kalshi-btc-arbitrage-bot](https://github.com/Sectionnaenumerate/Polymarket-Kalshi-btc-arbitrage-bot) | 270 | Cross-venue (Poly + Kalshi) | Spread rule detection, late resolution arbitrage, Rust core + Express layer, HTTP API (Axum) |
| [taetaehoho/poly-kalshi-arb](https://github.com/taetaehoho/poly-kalshi-arb) | 427 | Cross-venue (Poly + Kalshi) | Lock-free orderbook cache, SIMD arb detection, concurrent execution, circuit breaker, position tracking |
| [Trum3it/polymarket-arbitrage-bot](https://github.com/Trum3it/polymarket-arbitrage-bot) | 20 | ETH+BTC arbitrage | Market-neutral strategy, simulation/production mode, auto market discovery |
| [gamma-trade-lab/polymarket-arbitrage-bot](https://github.com/gamma-trade-lab/polymarket-arbitrage-bot) | 8 | BTC 15m vs 5m overlap | Two-leg execution, one-leg fill protection, simulation mode, auto-redeem |
| [Poly-Tutor/Polymarket-15min-arbitrage-bot](https://github.com/Poly-Tutor/Polymarket-15min-arbitrage-bot) | 10 | 15m dump-and-hedge | Multi-asset (BTC/ETH/SOL/XRP), API or WebSocket data source, trailing stop hedge |

---

## Architecture Assessment for polygolem API

### Best candidates for direct integration

1. **`GoPolymarket/polymarket-go-sdk`** — Most production-ready Go CLOB SDK.
   - Layered architecture: Application → Execution → Protocol → Security → Transport
   - AWS KMS signer support (institutional-grade)
   - WebSocket with auto-reconnect + heartbeat
   - Order builder with tick size, fee rate, neg risk awareness
   - Gamma API, pagination helpers, execution contracts
   - **Risk**: Community-maintained, not official. Audit before depending.

2. **`0xNetuser/Polymarket-golang`** — Complete py-clob-client port.
   - Gasless Web3 client (relay-based)
   - CTF exchange operations (split/merge/redeem)
   - Heartbeat system for open orders
   - **Risk**: No explicit license. Audit before use.

3. **`HuakunShen/polymarket-kit`** — Multi-language SDK with proxy.
   - Redundant WebSocket pool with message deduplication
   - OpenAPI schema generation → can generate any language client
   - MCP server for AI-driven research
   - **Risk**: Work in progress, not all APIs covered. Proxy dependency adds hop.

4. **`ybina/polymarket-go`** — Most comprehensive API coverage.
   - Includes Turnkey wallet management (institutional custody)
   - Bridge API for cross-chain deposits
   - Unified signer (PrivateKey / Turnkey)
   - Relayer client with Safe deployment + approval flow
   - **Best fit** if you need wallet/account management beyond trading.

5. **`Polymarket/polymarket-cli`** (Rust) — Official CLI as external tool.
   - Already designed as JSON-output API for agents
   - Can be wrapped as Go subprocess or Docker sidecar
   - **Less surface area** for auth bugs — CLI handles signing
   - **Trade-off**: Process overhead per call, higher latency

### Recommended Approach for polygolem

```
Phase 1 (Read-only):  polymarket-kit go-client OR pmclient (lightweight)
                      → Gamma search + CLOB orderbook queries
                      → WebSocket market data via polymarket-kit ws_pool

Phase 2 (Paper):      GoPolymarket/polymarket-go-sdk (with fake executor)
                      OR polymorph-cli wrapped via internal/polycli

Phase 3 (Live):       ybina/polymarket-go (full CLOB + relayer + bridge)
                      + Polymarket/go-builder-signing-sdk (auth validation)
```

### Key design patterns from these repos

- **Separation of concerns**: All mature Go SDKs split into `clob/`, `ws/`, `gamma/`, `data/`, `auth/`, `signer/` packages
- **Dual auth**: L1 (EIP-712) for key derivation, L2 (HMAC) for order operations
- **Builder attribution**: Separate builder credentials flow for rewards
- **Safe/Proxy wallets**: Gnosis Safe deployment + approval flows for institutional custody
- **WebSocket resilience**: Auto-reconnect, heartbeat, ping/pong, message deduplication
- **Order building**: Fluent builders with tick size validation, fee rate, neg risk awareness
- **Simulation modes**: Most rust bots have `DRY_RUN` or `simulation_mode` flags
- **Circuit breakers**: Position limits, daily loss limits, consecutive error thresholds, cooldown

---

## Rust Bot Strategy Patterns Worth Adopting

1. **Dual-limit same-size** (PolybaseX): Place symmetric limit buys at period start, hedge unfilled leg
2. **Dump-and-hedge** (Poly-Tutor): Detect price drops, buy opposite leg with trailing stop
3. **Cross-venue spread** (Sectionnaenumerate, taetaehoho): Kalshi YES price > Polymarket YES by threshold → buy Poly
4. **Time-window only** (multiple): Only evaluate signals in first N minutes of each period
5. **Circuit breaker stack** (taetaehoho): Max position per market, max total, max daily loss, consecutive error halt, cooldown

---

## Cloned Repos (repos/)

All repos cloned with `--depth 1` for code study only. No `.env`, no API keys, no execution.

| Directory | Source | Relevance to polygolem Phase 1 |
|---|---|---|
| `polymarket-go-sdk` | GoPolymarket/polymarket-go-sdk | Best Go CLOB SDK: REST, WS, pagination, Gamma. Clean layered architecture. |
| `polymarket-kit` | HuakunShen/polymarket-kit | WebSocket pool with dedup. OpenAPI schema. Go Gamma/CLOB/Data clients. |
| `polymarket-go` | ybina/polymarket-go | Broadest API coverage. Relayer, Bridge, Turnkey. Good for auth patterns. |
| `polymarket-go-gamma-client` | ivanzzeth/polymarket-go-gamma-client | Focused Gamma API. Type-safe market/event discovery. |
| `Polymarket-golang` | 0xNetuser/Polymarket-golang | Complete CLOB Go port from py-clob-client. L0/L1/L2 auth. |
| `go-builder-signing-sdk` | Polymarket/go-builder-signing-sdk | Official Go builder auth. Reference for correct signing. |
| `polymarket_cli` | vazic/polymarket_cli | Go CLI for AI agent integration. Similar philosophy (cobra + JSON output). |
| `rs-clob-client` | Polymarket/rs-clob-client | Official Rust CLOB client. Reference semantics for API contracts, types, auth. |

## Quick Reference: Go Package Names

```go
// Production CLOB SDK
go get github.com/GoPolymarket/polymarket-go-sdk

// Multi-language SDK with WebSocket pool
go get github.com/HuakunShen/polymarket-kit/go-client

// Trading + account management with Turnkey
go get github.com/ybina/polymarket-go

// Gamma API only
go get github.com/ivanzzeth/polymarket-go-gamma-client

// Complete CLOB client (py-clob-client port)
go get github.com/0xNetuser/Polymarket-golang

// Real-time data WebSocket
go get github.com/ivanzzeth/polymarket-go-real-time-data-client

// Official builder signing SDK
go get github.com/Polymarket/go-builder-signing-sdk
```
