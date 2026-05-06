# Commands

The Phase 1 binary command name is `polymarket`. Commands are designed around
safe read-only access, local paper state, and explicit status checks.

## Core

- `polymarket version`: prints the CLI version.
- `polymarket preflight`: reports local and remote readiness checks.

Examples:

```bash
polymarket version
polymarket --json version
polymarket preflight
```

## Markets

- `polymarket markets search`: intended read-only market search.
- `polymarket markets get`: intended read-only lookup for one market.
- `polymarket markets active`: intended read-only active-market listing.

Examples:

```bash
polymarket markets search "bitcoin"
polymarket --json markets active
polymarket --json markets get <market-id-or-slug>
```

## Market Data

- `polymarket orderbook get`: intended read-only order book lookup.
- `polymarket prices get`: intended read-only price lookup.

Examples:

```bash
polymarket --json orderbook get <token-id>
polymarket --json prices get <token-id>
```

## Paper

- `polymarket paper buy`: intended local simulated buy.
- `polymarket paper sell`: intended local simulated sell.
- `polymarket paper positions`: intended local position listing.
- `polymarket paper reset`: intended local paper-state reset.

Paper commands operate on local persisted state. They may use read-only market
data for reference pricing, but they must not send authenticated trading
mutations.

Examples:

```bash
polymarket --json paper buy --market <market-id> --outcome yes --size 10
polymarket --json paper sell --market <market-id> --outcome yes --size 5
polymarket --json paper positions
polymarket paper reset
```

## Status

- `polymarket auth status`: reports authentication readiness without exposing
  credential material.
- `polymarket live status`: reports live gate state without enabling execution.

Examples:

```bash
polymarket --json auth status
polymarket --json live status
```

## Automation Patterns

Use `--json` for scripts and agents so output remains stable and parseable.
Treat non-zero exits as failures and inspect structured error output.

```bash
set -euo pipefail

status="$(polymarket --json preflight)"
echo "$status" | jq .
```

For read-only market workflows:

```bash
set -euo pipefail

markets="$(polymarket --json markets search "election")"
echo "$markets" | jq .
```

## Availability

The Phase 1 CLI shell currently exposes the command groups and foundational
packages. Some subcommands are planned surfaces and may still return skeleton
responses until their package wiring is completed by later tasks. Documentation
must track actual behavior and must not imply that real live execution is
available.
