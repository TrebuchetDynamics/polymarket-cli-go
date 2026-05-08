# Polygolem Documentation

This directory contains the canonical documentation for polygolem. For a single source of truth, always prefer the docs listed here. Older or superseded docs reference these canonical versions.

## Quick Start

| I want to... | Read this |
|-------------|-----------|
| Install and run polygolem | [../README.md](../README.md) |
| Understand the deposit wallet onboarding flow | [ONBOARDING.md](./ONBOARDING.md) |
| Do the one-time browser signup (new users only) | [BROWSER-SETUP.md](./BROWSER-SETUP.md) |
| See all CLI commands | [COMMANDS.md](./COMMANDS.md) |
| Understand the architecture | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| Review safety and risk features | [SAFETY.md](./SAFETY.md) |

## Canonical Docs (Single Source of Truth)

### User Guides

| Doc | What It Covers | Status |
|-----|---------------|--------|
| [ONBOARDING.md](./ONBOARDING.md) | Complete deposit wallet onboarding flow: derive, deploy, approve, fund, trade. Includes headless steps AND browser login requirement for new users. | **Canonical** |
| [BROWSER-SETUP.md](./BROWSER-SETUP.md) | One-time browser login for new users. Security guidance for importing keys. Step-by-step with MetaMask/Rabby/WalletConnect. | **Canonical** |
| [COMMANDS.md](./COMMANDS.md) | Auto-generated CLI reference. Every command, flag, and example. | **Auto-generated** |
| [SAFETY.md](./SAFETY.md) | Read-only default, deposit wallet safety, risk breaker, circuit breaker. | **Canonical** |

### Technical Reference

| Doc | What It Covers | Status |
|-----|---------------|--------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Package boundaries, dependency direction, design decisions. | **Canonical** |
| [CONTRACTS.md](./CONTRACTS.md) | All smart contract addresses, factory ABI, CREATE2 derivation. | **Canonical** |
| [POLY_1271-SIGNING.md](./POLY_1271-SIGNING.md) | How POLY_1271 / deposit wallet signing works. | **Canonical** |

### Research & Findings

| Doc | What It Covers | Status |
|-----|---------------|--------|
| [INTEGRATION_PLAN.md](../opensource-projects/INTEGRATION_PLAN.md) | Ecosystem survey, 7 gap implementations, headless onboarding blocker analysis. | **Canonical** |
| [LIVE-TRADING-BLOCKER-REPORT.md](./LIVE-TRADING-BLOCKER-REPORT.md) | Empirical live trading test results with real funds. | **Canonical** |

## Deleted Docs

These docs contained outdated or false claims and have been removed:

| Doc | Why It Was Deleted |
|-----|-------------------|
| `BUILDER-AUTO.md` | Claimed "zero browser clicks" and deposit-wallet API key creation was headless. Proven false by live testing. |
| `BUILDER-CREDENTIAL-ISSUANCE.md` | Superseded document with conflated credential types. |
| `DEPOSIT-WALLET-DEPLOYMENT.md` | Claimed "no browser needed" for full onboarding. Missed the deposit-wallet API key blocker. |

## How We Keep Docs Consistent

1. **Empirical testing over claims** — If a doc says something works, it must be backed by a test or live verification.
2. **One topic, one canonical doc** — Onboarding is only in `ONBOARDING.md`. Browser setup is only in `BROWSER-SETUP.md`.
3. **Cross-references, not duplication** — Docs link to canonical sources rather than repeating information.
4. **Deprecation notices** — Outdated docs get a clear banner at the top pointing to the replacement.

---

*Last updated: 2026-05-08*
