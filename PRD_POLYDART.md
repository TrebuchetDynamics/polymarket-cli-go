# Polydart — Standalone Polymarket Dart SDK for Flutter

> **Status:** Draft — architecture confirmed, awaits implementation
> **Date:** 2026-05-07
> **Owner:** TrebuchetDynamics
> **License:** MIT (public)
> **Companion:** [polygolem docs/ONBOARDING.md](./docs/ONBOARDING.md) — reference pipeline
> **Companion:** [polygolem docs/CONTRACTS.md](./docs/CONTRACTS.md) — contract research

---

## 1. Vision

Polydart is the **official Dart-native Polymarket SDK** — a peer implementation to `polygolem` that brings the full Polymarket protocol stack to Flutter applications. It enables Flutter developers to build self-contained trading apps that query all Polymarket market data directly and place orders through a minimal server proxy.

**Polydart will always mirror the polygolem repository.** Every protocol module, API client, signing scheme, and safety boundary in polygolem has a corresponding Dart implementation in polydart. When polygolem evolves, polydart evolves in lockstep.

### 1.1 Target Use Case: Arenaton

Arenaton is a Flutter trading app that uses Reown/WalletConnect to connect with users' MetaMask wallets. Polydart provides:

- **Read-only market data** — direct from Flutter, no server needed
- **Order building + signing** — polydart builds the order data, Reown asks MetaMask to sign
- **Server proxy for deploy/batch** — a tiny server (~50 LOC) holds builder credentials and forwards relayer requests

---

## 2. Architecture — Confirmed Pipeline

### 2.1 The Three-Layer Model

```
┌──────────────────────────────────────────────────────────────────────┐
│                       ARENATON FLUTTER APP                           │
│                                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │  MARKET  │    │  ORDER   │    │  WALLET  │    │  REOWN   │      │
│  │  DATA    │    │ BUILDER  │    │ STATUS   │    │ SIGNING  │      │
│  │ (direct) │    │ (local)  │    │ (direct) │    │ (WC)     │      │
│  └────┬─────┘    └────┬─────┘    └────┬─────┘    └────┬─────┘      │
│       │               │               │               │            │
│       ▼               ▼               ▼               ▼            │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                       POLYDART (Dart SDK)                    │   │
│  │  src/auth  src/clob  src/gamma  src/wallet  src/orders       │   │
│  └──────────────────────────────────────────────────────────────┘   │
│       │                                       │                    │
│       │ (read-only + status)                  │ (deploy/batch)     │
│       ▼                                       ▼                    │
│  ┌──────────────┐                      ┌──────────────┐           │
│  │  Polymarket  │                      │  Relayer      │           │
│  │  APIs        │                      │  Proxy        │           │
│  │  (direct)    │                      │  (tiny server)│           │
│  └──────────────┘                      └──────┬───────┘           │
│                                               │                    │
│                                               ▼                    │
│                                        ┌──────────────┐           │
│                                        │  Builder      │           │
│                                        │  Relayer v2   │           │
│                                        └──────────────┘           │
└──────────────────────────────────────────────────────────────────────┘
```

### 2.2 What Goes Where

| Layer | What | Credentials |
|-------|------|-------------|
| **Flutter (Polydart)** | Market discovery, orderbook, prices, search, wallet status, order building | None needed (read-only) |
| **Reown/WalletConnect** | EOA signing: EIP-712 orders, batch signatures | MetaMask private key (user custody) |
| **Relayer Proxy (~50 LOC server)** | WALLET-CREATE, WALLET batch forwarding with builder headers | Builder API key/secret/passphrase |
| **Polymarket Relayer** | Actual on-chain deployment and submission | Operator role on factory |

---

## 3. Polydart Package Structure

```
polydart/
├── lib/
│   ├── src/
│   │   ├── auth/               # EIP-712, POLY_1271, ERC-7739, CREATE2
│   │   ├── clob/               # CLOB V2 client (read + write)
│   │   ├── gamma/              # Gamma API client
│   │   ├── dataapi/            # Data API client
│   │   ├── relayer/            # Builder relayer proxy client
│   │   ├── wallet/             # Deposit wallet lifecycle
│   │   ├── orders/             # OrderIntent, builders, validation
│   │   ├── stream/             # WebSocket market streams
│   │   ├── transport/          # HTTP retry, rate limit, circuit breaker
│   │   ├── types/              # Protocol types (mirrors polygolem internal/polytypes)
│   │   ├── config/             # Configuration, validation, redaction
│   │   ├── modes/              # Read-only / paper / live gates
│   │   ├── risk/               # Per-trade caps, daily loss limits
│   │   ├── paper/              # Local simulation state
│   │   └── execution/          # Order execution surface
│   └── polydart.dart           # Public SDK surface
├── test/
│   ├── unit/
│   ├── integration/
│   └── fixtures/               # Shared test vectors with polygolem
├── example/
│   └── arenaton_demo/          # Demo Flutter app
└── pubspec.yaml
```

---

## 4. Core Modules

### 4.1 `auth` — Cryptographic Primitives

**Mirrors:** `polygolem/internal/auth`

| Capability | Dart Implementation |
|-----------|---------------------|
| EIP-712 typed data signing | Delegated to Reown/WalletConnect |
| POLY_1271 order signatures | Custom — appends `0x03` signature type byte |
| ERC-7739 context | Custom — wraps order hash with chain/verifying contract |
| Deposit wallet CREATE2 | Keccak-256 local computation (`pointycastle`) |
| Address derivation | `ecdsa` pubKey → keccak → last 20 bytes |

**Key constraint:** No private key storage. All signing flows through Reown/WalletConnect. The Flutter app has access to the EOA address (from MetaMask) but never the private key.

### 4.2 `clob` — CLOB V2 Client

**Mirrors:** `polygolem/internal/clob`

- **Read endpoints** (no auth): book, trades, prices, spread, midpoint, tick size, neg risk, fee rate, last trade
- **Write endpoints** (signing required): create-order, cancel, update-balance
- **Signature type:** `POLY_1271` (deposit wallet) only — EOA/proxy/Safe blocked by CLOB V2

### 4.3 `gamma` — Market Discovery

**Mirrors:** `polygolem/internal/gamma`

- Search, markets, events, tags, series, sports, comments, profiles
- Read-only, no credentials
- Handles Gamma API quirks (inconsistent datetime, string-or-array fields)

### 4.4 `relayer` — Builder Relayer Proxy Client

**Mirrors:** `polygolem/internal/relayer`

- Client talks to **your server proxy**, not Polymarket's relayer directly
- Methods: `requestDeploy(eoaAddress)`, `submitBatch(eoaAddress, walletAddress, signedBatch)`
- Server proxy adds builder HMAC headers and forwards to `relayer-v2.polymarket.com`

### 4.5 `wallet` — Deposit Wallet Lifecycle

**Mirrors:** `polygolem/internal/wallet`

- `derive(eoaAddress)` → predict CREATE2 address (local computation)
- `deploy(eoaAddress)` → calls server proxy → relayer WALLET-CREATE
- `status(eoaAddress)` → checks deployment, balance, allowances
- `approve(eoaAddress, walletAddress)` → builds batch, user signs via Reown, server proxy submits

### 4.6 `orders` — Order Building

**Mirrors:** `polygolem/internal/orders`

```dart
final order = await client.orders
  .buy(tokenId: '123...')
  .atPrice(0.5)
  .forSize(10)
  .withBuilder('0x...')   // builder code from polymarket.com/settings?tab=builder
  .build();

// Sign via Reown/MetaMask
final signature = await client.reown.signTypedData(order.eip712Data);

// Submit to CLOB
await client.clob.submit(order, signature);
```

### 4.7 `stream` — Real-Time Data

**Mirrors:** `polygolem/internal/stream`

- WebSocket orderbook updates from `wss://ws-subscriptions-clob.polymarket.com/ws/`
- Auto-reconnect with exponential backoff
- Deduplication by message hash
- Callbacks: `onBook`, `onPriceChange`, `onLastTrade`

---

## 5. Reown/WalletConnect Integration

### 5.1 Security Model

| Component | Holds | Purpose |
|-----------|-------|---------|
| **MetaMask** (user device) | EOA private key | All signing |
| **Reown SDK** (Flutter) | WalletConnect session | Bridge between app and MetaMask |
| **Polydart** (Flutter) | EOA address, session | Build typed data, request signatures |
| **Server Proxy** (tiny server) | Builder API key/secret/passphrase | Forward relayer requests |
| **Polymarket Relayer** | Operator role on factory | Execute on-chain transactions |

### 5.2 Signing Flow

```
User creates order in Arenaton
         │
         ▼
Polydart builds EIP-712 typed data
         │
         ▼
Reown sends signTypedData request to MetaMask
         │
         ▼
User reviews and approves in MetaMask
         │
         ▼
MetaMask returns EOA signature
         │
         ▼
Polydart wraps as POLY_1271 (adds 0x03 signature type byte)
         │
         ▼
Polydart POSTs to CLOB with POLY_1271 signature
```

### 5.3 What Reown Handles

- `personal_sign` — for authentication messages
- `eth_signTypedData` — for EIP-712 order signing
- `eth_signTypedData_v4` — for deposit wallet batch signing
- `eth_sendTransaction` — for direct token transfers (EOA → wallet)

### 5.4 What Reown Does NOT Handle

- Builder credential management (server-side)
- Relayer authentication (server proxy)
- CREATE2 address computation (local Dart)
- Order validation (Polydart validation layer)

---

## 6. Unified Market Data Client

Polydart ships a **universal client** that wraps all Polymarket public data APIs:

```dart
final client = Polydart.universal();

// Gamma — market discovery
final markets = await client.activeMarkets();
final results = await client.search(query: 'btc 5m');

// CLOB — order books & pricing
final book = await client.orderBook(tokenId: '123...');
final price = await client.price(tokenId: '123...', side: 'BUY');
final history = await client.pricesHistory(market: '0x...', interval: '1h');

// Data API — volume & leaderboard
final volume = await client.liveVolume(limit: 10);
final board = await client.traderLeaderboard(limit: 100);

// Enriched — Gamma + CLOB combined
final enriched = await client.enrichedMarkets(limit: 50);

// Health
final health = await client.healthCheck(); // pings Gamma, CLOB, Data API
```

**This client mirrors `polygolem/pkg/universal`.** Every method has a Dart equivalent.

---

## 7. Mode System

**Mirrors:** `polygolem/docs/SAFETY.md`

| Mode | Reown | Server Proxy | CLOB Writes |
|------|-------|-------------|-------------|
| **Read-only** | Not needed | Not needed | Blocked |
| **Paper** | EOA address only | Not needed | Blocked (local sim) |
| **Live** | Reown provider connected | Required for deploy/batch | Full access with POLY_1271 signing |

**Gates:**
- Live mode requires Reown session + server proxy URL + preflight checks
- `risk` module enforces per-trade caps and daily loss limits
- Circuit breaker halts trading on repeated errors

---

## 8. The Server Proxy (Minimal)

**~50 LOC of Dart or Go.** The server proxy is the ONLY server component needed. Everything else is client-side.

```dart
// Dart version (can also be Go reusing polygolem)
// Endpoints:
//   POST /relay/deploy    ← { "eoaAddress": "0x..." }
//                          → forwards WALLET-CREATE with builder HMAC
//   POST /relay/batch     ← { signed batch data }
//                          → forwards WALLET batch with builder HMAC
```

**What it does:**
1. Accepts requests from authenticated Flutter users
2. Adds builder HMAC-SHA256 headers (using server-side builder creds)
3. Forwards to `https://relayer-v2.polymarket.com/submit`
4. Returns relayer response to Flutter client

**What it does NOT do:**
- Store or touch private keys
- Store or log builder credentials in responses
- Handle signing (signing is client-side via Reown)
- Handle order placement (direct CLOB from Flutter)
- Rate-limit per user
- Circuit-break on repeated failures

---

## 9. Implementation Phases

### Phase 1 — Foundation (v0.1.0)
- [ ] `types` — all protocol types (mirrors polygolem `internal/polytypes`)
- [ ] `gamma` — read-only market discovery
- [ ] `clob` — read-only endpoints
- [ ] `transport` — HTTP client with retry/rate limit
- [ ] `config` — env binding, validation
- [ ] `universal` — unified market data client
- [ ] Unit tests for all above

### Phase 2 — Authentication + Wallet (v0.2.0)
- [ ] `auth` — EIP-712 typed data building (signing via Reown)
- [ ] `wallet` — CREATE2 derivation, status checks
- [ ] `relayer` — server proxy client
- [ ] Reown/WalletConnect integration

### Phase 3 — Full Trading (v0.3.0)
- [ ] `orders` — OrderIntent builder, validation
- [ ] `clob` — write endpoints (create-order, cancel)
- [ ] `paper` — local simulation state
- [ ] `stream` — WebSocket market data
- [ ] `risk` — per-trade caps, daily limits

### Phase 4 — Polish (v0.4.0)
- [ ] `dataapi` — positions, volume, leaderboards
- [ ] Example Flutter app (Arenaton demo)
- [ ] Documentation site
- [ ] CI/CD with Flutter integration tests
- [ ] pub.dev package published

---

## 10. Dependencies

```yaml
dependencies:
  flutter:
    sdk: flutter
  http: ^1.2.0
  web_socket_channel: ^2.4.0
  web3dart: ^2.7.0
  pointycastle: ^3.7.0
  reown_appkit: ^1.0.0      # WalletConnect / Reown
  shared_preferences: ^2.2.0
  hive: ^2.2.0
  freezed_annotation: ^2.4.0
  json_annotation: ^4.8.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  build_runner: ^2.4.0
  freezed: ^2.4.0
  json_serializable: ^6.7.0
  mockito: ^5.4.0
```

---

## 11. Testing Strategy

**Mirrors polygolem test structure.**

| Test Type | Coverage |
|-----------|----------|
| Unit | Every module independently |
| Integration | Against Polymarket testnet/staging |
| Fixtures | Shared EIP-712 hashes, CREATE2 addresses with polygolem |
| Property-based | Order validation, price/size bounds |
| E2E | Full flow: connect → deploy → fund → trade |
| Widget | Flutter widget tests for UI components |

---

## 12. Security Rules

### 12.1 Private Key Handling

- **Private keys NEVER enter the Flutter app.**
- All signing is done by MetaMask via Reown/WalletConnect.
- The app only stores the EOA address (public).

### 12.2 Builder Credential Handling

- **Builder credentials NEVER enter the Flutter app.**
- Stored server-side in the relayer proxy.
- Never returned in API responses.
- Rotated on compromise.

### 12.3 Order Validation

- All orders are validated client-side before signing.
- Price/size bounds enforced by Polydart.
- Circuit breaker halts on repeated errors.
- User reviews full order in MetaMask before signing.

### 12.4 Network Security

- All API calls use HTTPS with certificate pinning.
- WebSocket connections use WSS.
- Rate limiting on all authenticated endpoints.
- Circuit breaker on the server proxy.

---

## 13. Success Criteria

- [ ] All polygolem `pkg/` APIs have polydart equivalents
- [ ] Read-only market data works with zero server dependency
- [ ] Order signing works via Reown/MetaMask
- [ ] Wallet deployment works via server proxy
- [ ] Server proxy < 100 LOC
- [ ] Shared test vectors pass in both repos (polygolem + polydart)
- [ ] Example Arenaton demo app works end-to-end
- [ ] CI passes: Dart analysis, tests, integration tests
- [ ] pub.dev package published

---

## 14. Open Questions

1. **Reown vs WalletConnect v3:** Which library is more stable for production?
2. **Server proxy language:** Go (polygolem reuse, one binary) or Dart (single stack)?
3. **Paper state storage:** `shared_preferences` vs `hive` vs `drift`?
4. **Flutter minimum version:** 3.16+ or 3.19+?
5. **Builder fee passthrough:** Does Arenaton charge builder fees on user orders?

---

*This PRD defines the Polydart SDK and the deployment pipeline for the Arenaton Flutter app. Every design decision is based on confirmed on-chain research (see polygolem docs/CONTRACTS.md). The only manual step in the entire pipeline is the one-time copy of builder credentials from polymarket.com/settings?tab=builder — everything else is automated.*
