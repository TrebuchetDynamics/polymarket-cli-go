# Polymarket CLI & SDK Ecosystem — Complete Survey

Research date: 2026-05-08
Scope: Comprehensive survey of all known open-source Polymarket SDKs, clients, CLIs, trading bots, and tools.
Updated: Expanded to 35+ projects. Added feature matrix. Added polygolem gap analysis.

---

## Contents

1. [Official SDKs](#official-sdks)
2. [Official Relayer & Builder Clients](#official-relayer--builder-clients)
3. [Community Go SDKs](#community-go-sdks)
4. [Community Rust SDKs](#community-rust-sdks)
5. [Community Python & TypeScript SDKs](#community-python--typescript-sdks)
6. [Trading Bots](#trading-bots)
7. [Feature Matrix](#feature-matrix)
8. [Polygolem Comparison & Gap Analysis](#polygolem-comparison--gap-analysis)
9. [Cloned Repos](#cloned-repos)

---

## Official SDKs

Maintained by Polymarket. These are the canonical implementations.

| Repo | Language | Stars | Package | License | Status |
|------|----------|-------|---------|---------|--------|
| [Polymarket/clob-client-v2](https://github.com/Polymarket/clob-client-v2) | TypeScript | 15 | `@polymarket/clob-client-v2` | MIT | Active — V2 production |
| [Polymarket/py-clob-client-v2](https://github.com/Polymarket/py-clob-client-v2) | Python | 76 | `py-clob-client-v2` | MIT | Active — V2 production |
| [Polymarket/rs-clob-client-v2](https://github.com/Polymarket/rs-clob-client-v2) | Rust | 10 | `polymarket_client_sdk_v2` | MIT | Active — V2 production |
| [Polymarket/clob-client](https://github.com/Polymarket/clob-client) | TypeScript | 507 | `@polymarket/clob-client` | MIT | Legacy — V1 only |
| [Polymarket/py-clob-client](https://github.com/Polymarket/py-clob-client) | Python | 1,187 | `py-clob-client` | MIT | Legacy — V1 only |
| [Polymarket/rs-clob-client](https://github.com/Polymarket/rs-clob-client) | Rust | 646 | `polymarket_client_sdk` | MIT | Legacy — V1 only |
| [Polymarket/polymarket-sdk](https://github.com/Polymarket/polymarket-sdk) | TypeScript | 63 | `@polymarket/sdk` | MIT | Active — wallet primitives |
| [Polymarket/polymarket-us-python](https://github.com/Polymarket/polymarket-us-python) | Python | — | `polymarket-us` | — | Active — Polymarket US API |

### Official V2 SDK Features

All three official V2 SDKs share these capabilities:

- **L1 auth** — EIP-712 signature for API key derivation
- **L2 auth** — HMAC with API credentials for order operations
- **Order types** — GTC, GTD, FOK, FAK, Post Only
- **Market orders** — Amount-based market buys/sells
- **Batch operations** — Post/cancel multiple orders
- **Deposit wallet** — POLY_1271 / signature type 3 (via relayer client)
- **Builder attribution** — `builderCode` field on orders
- **V1/V2 protocol auto-detection** (Rust SDK)

### Official V1 SDK Features (Legacy)

The V1 SDKs support EOA (type 0), Proxy (type 1), and Gnosis Safe (type 2) wallets. They do NOT support deposit wallet (type 3). Production CLOB no longer accepts V1 orders as of April 28, 2026.

---

## Official Relayer & Builder Clients

| Repo | Language | Stars | Package | What |
|------|----------|-------|---------|------|
| [Polymarket/builder-relayer-client](https://github.com/Polymarket/builder-relayer-client) | TypeScript | — | `@polymarket/builder-relayer-client` | WALLET-CREATE, WALLET batch, Safe deploy |
| [Polymarket/py-builder-relayer-client](https://github.com/Polymarket/py-builder-relayer-client) | Python | — | `py-builder-relayer-client` | derive_deposit_wallet, WALLET batch |
| [GoPolymarket/go-builder-relayer-client](https://pkg.go.dev/github.com/GoPolymarket/go-builder-relayer-client) | Go | — | `go-builder-relayer-client` | Go relayer with deposit wallet |
| [Polymarket/go-builder-signing-sdk](https://github.com/Polymarket/go-builder-signing-sdk) | Go | 3 | `go-builder-signing-sdk` | HMAC builder auth (legacy, no builderCode field) |

Key contracts (verified on Polygonscan):
```
DepositWalletFactory:        0x00000000000Fb5C9ADea0298D729A0CB3823Cc07
DepositWalletImplementation: 0x58CA52ebe0DadfdF531Cde7062e76746de4Db1eB
```

---

## Community Go SDKs

### Active / Recently Updated

| Repo | Stars | License | Last Update | CLOB V2 | Deposit Wallet |
|------|-------|---------|-------------|---------|---------------|
| [nijaru/go-clob-client](https://github.com/nijaru/go-clob-client) | — | — | Apr 2026 | Yes | Unknown |
| [splicemood/polymarket-go-sdk](https://github.com/splicemood/polymarket-go-sdk) | — | — | Apr 2026 | Yes (fork) | Unknown |
| [GoPolymarket/polymarket-go-sdk](https://github.com/GoPolymarket/polymarket-go-sdk) | 47 | Apache-2.0 | Apr 2026 | v1.1.3 | No |

### V1-Only (Legacy — Architecture Reference Only)

| Repo | Stars | License | Key Strengths |
|------|-------|---------|---------------|
| [0xNetuser/Polymarket-golang](https://github.com/0xNetuser/Polymarket-golang) | 72 | MIT | Complete py-clob-client port. Web3 clients (gas + gasless). RFQ. CTF split/merge/redeem. Batch redeem. |
| [HuakunShen/polymarket-kit](https://github.com/HuakunShen/polymarket-kit) | 55 | MIT | Multi-language SDK (TS/Python/Go). Proxy server. WebSocket pool with dedup. MCP server. OpenAPI codegen. |
| [ybina/polymarket-go](https://github.com/ybina/polymarket-go) | 20 | MIT | Broadest API coverage. Turnkey wallet integration. Bridge. Safe deployment + approval. |
| [Polymarket/go-builder-signing-sdk](https://github.com/Polymarket/go-builder-signing-sdk) | 3 | — | Official builder HMAC signing. Reference implementation. |
| [ivanzzeth/polymarket-go-gamma-client](https://github.com/ivanzzeth/polymarket-go-gamma-client) | 30 | — | Focused Gamma API. Type-safe discovery. Still usable for reads. |
| [ivanzzeth/polymarket-go-real-time-data-client](https://github.com/ivanzzeth/polymarket-go-real-time-data-client) | 13 | — | WebSocket real-time data. Still usable. |
| [D8-X/polymarket-trader-go-sdk](https://github.com/D8-X/polymarket-trader-go-sdk) | 0 | — | V1-only |
| [aszxqaz/pmclient](https://github.com/aszxqaz/pmclient) | — | — | V1-only |
| [lajosdeme/polymarket-go-api](https://github.com/lajosdeme/polymarket-go-api) | — | — | V1-only |
| [CalderWhite/polymarket-gamma-go](https://github.com/CalderWhite/polymarket-gamma-go) | 11 | — | Read-only Gamma |
| [bububa/polymarket-client](https://github.com/bububa/polymarket-client) | 0 | — | V1-only |
| [monsterdev914/polymarket-trading-bot](https://github.com/monsterdev914/polymarket-trading-bot) | 27 | — | V1-only bot |
| [arjunprakash027/Mantis](https://github.com/arjunprakash027/Mantis) | 12 | — | V1-only |
| [vazic/polymarket_cli](https://github.com/vazic/polymarket_cli) | 0 | — | Go CLI for AI agents. Cobra + JSON output. |

---

## Community Rust SDKs

| Repo | Stars | License | What | Key Features |
|------|-------|---------|------|-------------|
| [tdergouzi/rs-clob-client](https://github.com/tdergouzi/rs-clob-client) | — | — | TS port to Rust | Full EIP-712 signing |
| [tdergouzi/rs-clob-client-v2](https://github.com/tdergouzi/rs-clob-client-v2) | — | — | Fork of official V2 | Deposit wallet support |

---

## Community Python & TypeScript SDKs

| Repo | Language | Stars | What |
|------|----------|-------|------|
| [qualiaenjoyer/polymarket-apis](https://github.com/qualiaenjoyer/polymarket-apis) | Python | — | derive_deposit_wallet, complete relayer flow |
| [cyl19970726/poly-sdk](https://github.com/cyl19970726/poly-sdk) | TypeScript | — | Unified SDK with TradingService, cache, rate limiter |
| [Polymarket/safe-wallet-integration](https://github.com/Polymarket/safe-wallet-integration) | TypeScript | — | Next.js reference implementation. Safe deploy, trading, fee collection. |
| [Polymarket/privy-safe-builder-example](https://github.com/Polymarket/privy-safe-builder-example) | TypeScript | — | Privy + Safe + builder relayer integration example |

---

## Trading Bots

### Rust Bots

| Repo | Stars | Strategy | Key Features |
|------|-------|----------|-------------|
| [taetaehoho/poly-kalshi-arb](https://github.com/taetaehoho/poly-kalshi-arb) | 427 | Cross-venue (Poly + Kalshi) | Lock-free orderbook cache, SIMD arb detection, circuit breaker, position tracking |
| [Sectionnaenumerate/Polymarket-Kalshi-btc-arbitrage-bot](https://github.com/Sectionnaenumerate/Polymarket-Kalshi-btc-arbitrage-bot) | 270 | Cross-venue (Poly + Kalshi) | Spread rule detection, late resolution arb, Rust core + Express layer |
| [rvenandowsley/Polymarket-crypto-5min-arbitrage-bot](https://github.com/rvenandowsley/Polymarket-crypto-5min-arbitrage-bot) | 96 | BTC 5m YES+NO arb | 20ms order placement, 50 checks/sec, merge/redeem, GTC/GTD/FOK/FAK |
| [PolybaseX/Polymarket-Trading-Bot-Rust](https://github.com/PolybaseX/Polymarket-Trading-Bot-Rust) | 95 | Dual limit + trailing | Limit buys at $0.45, hedge logic, backtest, simulation mode |
| [rvenandowsley/Polymarket-crypto-1hour-arbitrage-bot](https://github.com/rvenandowsley/Polymarket-crypto-1hour-arbitrage-bot) | 9 | BTC 1h YES+NO arb | Market discovery, order book monitoring, merge task |
| [Trum3it/polymarket-arbitrage-bot](https://github.com/Trum3it/polymarket-arbitrage-bot) | 20 | ETH+BTC arb | Market-neutral, simulation/production mode, auto market discovery |
| [gamma-trade-lab/polymarket-arbitrage-bot](https://github.com/gamma-trade-lab/polymarket-arbitrage-bot) | 8 | BTC 15m vs 5m overlap | Two-leg execution, one-leg fill protection, simulation, auto-redeem |
| [Poly-Tutor/Polymarket-15min-arbitrage-bot](https://github.com/Poly-Tutor/Polymarket-15min-arbitrage-bot) | 10 | 15m dump-and-hedge | Multi-asset (BTC/ETH/SOL/XRP), API or WebSocket source, trailing stop |
| [ApolloPolyX/Polymarket-Arbitrage-Bot-V2](https://github.com/ApolloPolyX/Polymarket-Arbitrage-Bot-V2) | 0 | Pre-order 5m/15m | Modular execution framework, backtest, simulation, operational binaries |
| [polymarket-traders/Polymarket-arb-bot](https://github.com/polymarket-traders/Polymarket-arb-bot) | 97 | Cross-venue (Poly + Kalshi) | Lock-free atomic cache, SIMD acceleration, concurrent execution, SQLite history |

### Python Bots

| Repo | Stars | Strategy | Key Features |
|------|-------|----------|-------------|
| [soldino777/polymarket-arb-bot](https://github.com/soldino777/polymarket-arb-bot) | 91 | Cross-market (Poly + Kalshi) | NLP name matching, parallel execution, partial fill handling, dashboard |
| [qntrade/polymarket-5min-15min-arbitrage-bot](https://github.com/qntrade/polymarket-5min-15min-arbitrage-bot) | 109 | Volatility arb | FastAPI backend, Next.js dashboard, Redis state, delta-neutral |
| [qntrade/polymarket-kalshi-arbitrage-bot](https://github.com/qntrade/polymarket-kalshi-arbitrage-bot) | 270 | Kalshi arb | Dual opportunity detection, fee calculation, auto-execution |
| [LvcidPsyche/polymarket-arbitrage-bot](https://github.com/LvcidPsyche/polymarket-arbitrage-bot) | 3 | Cross-platform | Node.js + Python hybrid, ML opportunity predictor, dashboard, silent execution |

### TypeScript Bots

| Repo | Stars | Strategy | Key Features |
|------|-------|----------|-------------|
| [figure-markets/polymarket-arbitrage-bot](https://github.com/figure-markets/polymarket-arbitrage-bot) | 218 | 15m dump-and-hedge | Multi-asset, auto-discovery, dump detection, official CLOB client |

---

## Feature Matrix

### Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Full support |
| 🟡 | Partial / limited support |
| ❌ | Not supported |
| ? | Unknown / not documented |
| — | Not applicable |

### SDKs & Clients

| Feature | polygolem | Official TS V2 | Official Py V2 | Official Rust V2 | GoPolymarket SDK | 0xNetuser Go | ybina Go | polymarket-kit | rs-clob-client |
|---------|-----------|----------------|----------------|------------------|------------------|--------------|----------|----------------|----------------|
| **Language** | Go | TypeScript | Python | Rust | Go | Go | Go | TS/Python/Go | Rust |
| **Stars** | — | 15 | 76 | 10 | 47 | 72 | 20 | 55 | 646 |
| **License** | MIT | MIT | MIT | MIT | Apache-2.0 | MIT | MIT | MIT | MIT |
| **Official** | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **CLOB V2** | ✅ | ✅ | ✅ | ✅ | 🟡 (v1.1.3) | ❌ | ❌ | ❌ | ❌ |
| **Deposit Wallet (Type 3)** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **EOA (Type 0)** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Proxy (Type 1)** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Safe (Type 2)** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CLOB Trading** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Market Orders** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Limit Orders (GTC)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Limit Orders (GTD)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ? | ✅ |
| **FOK Orders** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ? | ✅ |
| **FAK Orders** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ? | ✅ |
| **Post Only** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ? | ✅ |
| **Batch Order Posting** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ? | ✅ |
| **Order Cancellation** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Gamma API** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Data API** | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| **WebSocket** | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| **Bridge API** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ |
| **RFQ API** | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ |
| **Relayer Client** | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| **On-chain / Web3** | 🟡 | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **CTF Split/Merge/Redeem** | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Builder Attribution** | ✅ | ✅ | ✅ | ✅ | ✅ | 🟡 | ✅ | ❌ | ✅ |
| **Remote Builder Signing** | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| **Heartbeats** | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |
| **Turnkey Integration** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| **AWS KMS Signer** | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| **CLI Tool** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Paper / Simulation Trading** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Circuit Breaker / Risk** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Auto-pagination** | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| **Read-only Default** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **No External SDK Dep** | ✅ | ❌ (viem) | ❌ | ❌ (alloy) | ❌ | ❌ | ❌ | ❌ | ❌ |
| **ERC-7739 / POLY_1271** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **CREATE2 Derivation** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **V1/V2 Auto-detect** | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Test Coverage** | ✅ (27/28) | ? | ? | ✅ | ✅ (>=40%) | ? | ✅ | ? | ✅ |
| **CI/CD** | ✅ | ✅ | ✅ | ✅ | ✅ | ? | ? | ? | ✅ |

### Trading Bots

| Feature | polygolem | PolybaseX Rust | rvenandowsley 5m | figure-markets TS | soldino777 Py | qntrade Py | taetaehoho Rust |
|---------|-----------|----------------|------------------|-------------------|---------------|------------|-----------------|
| **Language** | Go | Rust | Rust | TypeScript | Python | Python | Rust |
| **Stars** | — | 95 | 96 | 218 | 91 | 109 | 427 |
| **CLOB V2** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Deposit Wallet** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Live Trading** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Simulation Mode** | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ |
| **Paper Trading** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Strategy Engine** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Backtesting** | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Cross-venue Arb** | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| **YES+NO Arb** | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| **Circuit Breaker** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Risk Limits** | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ |
| **WebSocket Data** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Auto-redeem** | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Position Tracking** | 🟡 | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ |
| **Dashboard / UI** | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| **P&L Tracking** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |

---

## Polygolem Comparison & Gap Analysis

### What polygolem does uniquely well

1. **Only Go implementation with full deposit wallet lifecycle** — No other Go SDK ships CREATE2 derivation, WALLET-CREATE, WALLET batch, ERC-7739 signing, and CLI onboarding in one repo.
2. **Read-only by default** — All `pkg/` public APIs are read-only. Authentication is explicit. This is a safety feature no other SDK emphasizes.
3. **Zero external SDK dependencies** — All signing, type definitions, and protocol logic implemented from spec. Only depends on `go-ethereum`, `gorilla/websocket`, `cobra`, `viper`.
4. **Headless V2 onboarding CLI** — `polygolem builder auto` + `auth headless-onboard` + `deposit-wallet onboard` is a complete zero-browser flow. Official SDKs require browser-based SIWE or Reown.
5. **Paper trading built-in** — `polygolem paper buy/positions/reset` for local simulation without API keys.
6. **Circuit breaker & risk management** — Per-trade caps, daily loss limits, and transport-level circuit breaker.
7. **V2-only deposit-wallet order signing** — Internal `orders` package signs POLY_1271 orders with deposit-wallet maker/signer, deposit-wallet-owned L2 headers, optional builder attribution, post-only support, batch posting, and heartbeats.
8. **Verified against official test vectors** — CREATE2 derivation verified against official Python SDK test vector.

### Where polygolem has gaps vs. the ecosystem

| Gap | Severity | Notes |
|-----|----------|-------|
| **RFQ API** | Low | Request-for-quote is a specialized feature. Most bots don't need it. |
| **CTF on-chain operations** | Medium | Split/merge/redeem positions. 0xNetuser's Go SDK has full Web3 clients for this. |
| **Remote builder signing** | Low | For client-side apps that can't expose builder secrets. GoPolymarket SDK has a signer-server pattern. |
| **AWS KMS / Turnkey signers** | Low | Institutional custody. GoPolymarket SDK and ybina SDK support this. |
| **Streaming pagination** | Low | `StreamData()` helper in GoPolymarket SDK and official Rust SDK. Polygolem has `pkg/pagination` but not streaming. |
| **Web3 / on-chain transfers** | Low | Direct USDC/conditional token transfers. Polygolem only has bridge deposit + relayer batch. |
| **V1/V2 auto-detection** | Low | Official Rust SDK auto-detects protocol from host. Polygolem requires explicit version awareness. |
| **Cross-venue arbitrage** | N/A | Out of scope for an SDK. Trading bots handle this. |
| **Dashboard / Web UI** | N/A | Out of scope. Bots like qntrade have Next.js dashboards. |
| **Machine learning** | N/A | Out of scope. LvcidPsyche bot has ML opportunity predictor. |

### Polygolem's strategic position

```
Uniqueness axis:
┌─────────────────────────────────────────────────────────────────┐
│  High    │  polygolem  │  taetaehoho bot  │  polymarket-kit   │
│          │  (Go + V2    │  (SIMD + lock-   │  (Multi-lang +    │
│          │   deposit    │   free + cross-  │   MCP + proxy)    │
│          │   wallet)    │   venue)         │                   │
│          │              │                  │                   │
│  Medium  │  GoPolymarket│  figure-markets  │  0xNetuser Go     │
│          │  SDK (AWS    │  (TS + dump-     │  (Web3 + gasless) │
│          │  KMS + WS)   │  hedge)          │                   │
│          │              │                  │                   │
│  Low     │  Official    │  Official Py     │  Official Rust    │
│          │  TS V2       │  V2              │  V2               │
│          │  (canonical) │  (canonical)     │  (canonical)      │
└─────────────────────────────────────────────────────────────────┘
         Niche / Specialized          →           Canonical / General
```

Polygolem occupies a **unique niche**: it's the only Go SDK for the deposit wallet flow. For teams building Go-based trading infrastructure that must use CLOB V2 with deposit wallets, polygolem is currently the only viable open-source option.

### Recommended ecosystem integrations for polygolem

1. **For CTF operations** — Study `0xNetuser/Polymarket-golang`'s `web3/` package. The `PolymarketWeb3Client` and `PolymarketGaslessWeb3Client` are well-architected references for split/merge/redeem.
2. **For WebSocket resilience** — Study `HuakunShen/polymarket-kit`'s `RedundantWSPool` and message deduplication. Also study `GoPolymarket/polymarket-go-sdk`'s heartbeat + reconnect policy.
3. **For institutional signers** — Study `GoPolymarket/polymarket-go-sdk`'s AWS KMS signer and `ybina/polymarket-go`'s Turnkey integration.
4. **For trading bot patterns** — Study `taetaehoho/poly-kalshi-arb` for circuit breaker design and `figure-markets/polymarket-arbitrage-bot` for dump-and-hedge strategy implementation.
5. **For builder attribution** — Study `GoPolymarket/polymarket-go-sdk`'s remote signer server pattern (`cmd/signer-server`).

---

## Cloned Repos (repos/)

All repos cloned with `--depth 1` for code study only. No `.env`, no API keys, no execution.

| Directory | Source | Relevance |
|---|---|---|
| `ctf-exchange-v2` | Polymarket/ctf-exchange-v2 | V2 exchange contracts — reference for POLY_1271 validation flow |
| `foxme666-Polymarket-golang` | 0xNetuser/Polymarket-golang | Complete Go CLOB port. Web3 clients, RFQ, gasless relay. |
| `go-builder-signing-sdk` | Polymarket/go-builder-signing-sdk | Official Go builder HMAC auth. Reference for header signing. |
| `polymarket_cli` | vazic/polymarket_cli | Go CLI for AI agents. Cobra + JSON output patterns. |
| `polymarket-go-gamma-client` | ivanzzeth/polymarket-go-gamma-client | Focused Gamma API. Type-safe market discovery. |
| `polymarket-go-sdk` | GoPolymarket/polymarket-go-sdk | Best Go CLOB SDK: REST, WS, pagination, Gamma, AWS KMS. |
| `polymarket-go` | ybina/polymarket-go | Broadest API coverage. Relayer, Bridge, Turnkey. |
| `Polymarket-golang` | 0xNetuser/Polymarket-golang | Complete py-clob-client port. L0/L1/L2 auth. |
| `polymarket-kit` | HuakunShen/polymarket-kit | WebSocket pool with dedup. OpenAPI schema. Multi-language. |
| `py-clob-client` | Polymarket/py-clob-client | Official Python V1 client. Reference for API contracts. |
| `rs-clob-client` | Polymarket/rs-clob-client | Official Rust V1 client. Reference semantics, types, auth. |

---

## Quick Reference: Go Package Names

```go
// Production CLOB SDK (V1-only, community)
go get github.com/GoPolymarket/polymarket-go-sdk

// Multi-language SDK with WebSocket pool (V1-only)
go get github.com/HuakunShen/polymarket-kit/go-client

// Trading + account management with Turnkey (V1-only)
go get github.com/ybina/polymarket-go

// Gamma API only
go get github.com/ivanzzeth/polymarket-go-gamma-client

// Complete CLOB client (V1-only, py-clob-client port)
go get github.com/0xNetuser/Polymarket-golang

// Real-time data WebSocket
go get github.com/ivanzzeth/polymarket-go-real-time-data-client

// Official builder signing SDK (legacy HMAC)
go get github.com/Polymarket/go-builder-signing-sdk

// Go relayer with deposit wallet
go get github.com/GoPolymarket/go-builder-relayer-client
```

---

## Research Methodology

This survey was compiled using:

1. **GitHub search** — `polymarket sdk`, `polymarket clob client`, `polymarket trading bot`, `polymarket arbitrage`
2. **Polymarket documentation** — [docs.polymarket.com](https://docs.polymarket.com)
3. **Direct repo analysis** — README, source code, go.mod, Cargo.toml, package.json
4. ** crates.io / npm / PyPI** — Package registry metadata
5. **Cloned repos** — `--depth 1` clones for structural analysis

Last updated: 2026-05-08
