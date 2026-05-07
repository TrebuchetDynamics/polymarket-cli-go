---
name: Polygolem
description: Safe read-only Polymarket CLI for market discovery, orderbook data, and paper trading research
---

# Polygolem CLI Skill

This skill allows Claude to interact with Polymarket prediction markets through Polygolem, a safe Go CLI designed for research and AI agent integration. All commands output structured JSON by default.

## Prerequisites

Ensure the `polygolem` binary is built and available:

```bash
cd go-bot/polygolem && go build -o polygolem ./cmd/polygolem
```

Verify with:
```bash
./polygolem health
```

## Commands

Always use `--json` flag for structured data. Commands are grouped by domain.

### Discover — Market Discovery (Gamma API, read-only, no credentials)

Search for active markets:
```bash
./polygolem discover search --limit 10
./polygolem discover search --query "btc 5m" --limit 5
```

Get market details by ID or slug:
```bash
./polygolem discover market --id "0x..."
./polygolem discover market --slug "will-btc-be-above-100k"
```

Enrich a Gamma market with CLOB data (tick size, fee rate, neg risk, orderbook):
```bash
./polygolem discover enrich --id "0x..."
```

### Orderbook — CLOB Market Data (read-only, no credentials)

Get L2 orderbook depth:
```bash
./polygolem orderbook get --token-id "123456789..."
```

Get best price, midpoint, spread, tick size, or fee rate:
```bash
./polygolem orderbook price --token-id "123456789..."
./polygolem orderbook midpoint --token-id "123456789..."
./polygolem orderbook spread --token-id "123456789..."
./polygolem orderbook tick-size --token-id "123456789..."
./polygolem orderbook fee-rate --token-id "123456789..."
```

### Health — API Reachability

Check Gamma and CLOB connectivity:
```bash
./polygolem health
```

### Version & Preflight

```bash
./polygolem version
./polygolem preflight
```

## How to Use This Skill

When a user asks about Polymarket markets:

1. **For odds on an event**: Search → get market ID → get market details → report prices.
   ```bash
   ./polygolem discover search --query "btc 5m" --limit 5
   ./polygolem discover market --id "<conditionId from search>"
   ```

2. **For orderbook depth**: Get the orderbook for a specific token.
   ```bash
   ./polygolem orderbook get --token-id "<clobTokenId>"
   ```

3. **For market tradability assessment**: Enrich a market to check tick size, fee rate, neg risk, and whether it accepts orders.
   ```bash
   ./polygolem discover enrich --id "<marketId>"
   ```

4. **For API health**: Quick connectivity check.
   ```bash
   ./polygolem health
   ```

## Safety

Polygolem is **read-only by default**. It does not place orders, sign transactions, or require credentials. Paper mode is simulated locally. Live execution is hard-disabled until explicit operator approval and gate checks pass.
