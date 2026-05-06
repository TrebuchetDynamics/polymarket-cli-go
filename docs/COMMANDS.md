# Commands

The Phase 1 binary command name is `polymarket`. Commands are designed around
safe read-only access, local paper state, and explicit status checks.

## Core

- `polymarket version`: prints the CLI version.
- `polymarket preflight`: reports local and remote readiness checks.

## Markets

- `polymarket markets search`: intended read-only market search.
- `polymarket markets get`: intended read-only lookup for one market.
- `polymarket markets active`: intended read-only active-market listing.

## Market Data

- `polymarket orderbook get`: intended read-only order book lookup.
- `polymarket prices get`: intended read-only price lookup.

## Paper

- `polymarket paper buy`: intended local simulated buy.
- `polymarket paper sell`: intended local simulated sell.
- `polymarket paper positions`: intended local position listing.
- `polymarket paper reset`: intended local paper-state reset.

Paper commands operate on local persisted state. They may use read-only market
data for reference pricing, but they must not send authenticated trading
mutations.

## Status

- `polymarket auth status`: reports authentication readiness without exposing
  credential material.
- `polymarket live status`: reports live gate state without enabling execution.

## Availability

The Phase 1 CLI shell currently exposes the command groups and foundational
packages. Some subcommands are planned surfaces and may still return skeleton
responses until their package wiring is completed by later tasks. Documentation
must track actual behavior and must not imply that real live execution is
available.
