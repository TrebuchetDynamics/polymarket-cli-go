# Commands

The Phase 1 binary command name is `polygolem`. Commands are designed around
safe read-only access, local paper state, and explicit status checks.

## Core

- `polygolem version`: prints the CLI version.
- `polygolem preflight`: reports local and remote readiness checks.

Examples:

```bash
polygolem version
polygolem --json version
polygolem preflight
```

## Markets

- `polygolem markets search`: intended read-only market search.
- `polygolem markets get`: intended read-only lookup for one market.
- `polygolem markets active`: intended read-only active-market listing.

Examples:

```bash
polygolem markets search "bitcoin"
polygolem --json markets active
polygolem --json markets get <market-id-or-slug>
```

## Market Data

- `polygolem orderbook get`: intended read-only order book lookup.
- `polygolem prices get`: intended read-only price lookup.

Examples:

```bash
polygolem --json orderbook get <token-id>
polygolem --json prices get <token-id>
```

## Paper

- `polygolem paper buy`: intended local simulated buy.
- `polygolem paper sell`: intended local simulated sell.
- `polygolem paper positions`: intended local position listing.
- `polygolem paper reset`: intended local paper-state reset.

Paper commands operate on local persisted state. They may use read-only market
data for reference pricing, but they must not send authenticated trading
mutations.

Examples:

```bash
polygolem --json paper buy --market <market-id> --outcome yes --size 10
polygolem --json paper sell --market <market-id> --outcome yes --size 5
polygolem --json paper positions
polygolem paper reset
```

## Status

- `polygolem auth status`: reports authentication readiness without exposing
  credential material.
- `polygolem live status`: reports live gate state without enabling execution.

Examples:

```bash
polygolem --json auth status
polygolem --json live status
```

## Automation Patterns

Use `--json` for scripts and agents so output remains stable and parseable.
Treat non-zero exits as failures and inspect structured error output.

```bash
set -euo pipefail

status="$(polygolem --json preflight)"
echo "$status" | jq .
```

For read-only market workflows:

```bash
set -euo pipefail

markets="$(polygolem --json markets search "election")"
echo "$markets" | jq .
```

## Availability

The Phase 1 CLI shell currently exposes the command groups and foundational
packages. Some subcommands are planned surfaces and may still return skeleton
responses until their package wiring is completed by later tasks. Documentation
must track actual behavior and must not imply that real live execution is
available.
