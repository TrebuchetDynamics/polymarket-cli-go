---
name: Polygolem
description: Polymarket protocol and automation CLI for AI agents. Read-only by default; opt-in mutating commands for deposit-wallet, CLOB orders, and bridge deposits.
---

# Polygolem CLI Skill

Polygolem is a Go protocol and automation stack for Polymarket with a Cobra
CLI frontend. This skill lets Claude (or any scripted caller) drive every
command group: market discovery, orderbook data, paper trading, deposit-wallet
onboarding, CLOB orders, bridge funding, and health checks.

The CLI is **read-only by default**. Mutating commands require explicit
credentials and, for live signing, a `--signature-type` flag plus passing
preflight gates. See [Safety surface](#safety-surface).

## Prerequisites

Build the binary once:

```bash
cd go-bot/polygolem && go build -o polygolem ./cmd/polygolem
./polygolem version --json
```

Reachability check before any session:

```bash
./polygolem health --json
```

## JSON contract reference

Every command accepts `--json` and emits a versioned envelope. The full
specification is `docs/JSON-CONTRACT.md`; the summary an agent needs is:

**Success:**

```json
{
  "ok": true,
  "version": "1",
  "data": { ... },
  "meta": { "command": "...", "ts": "...", "duration_ms": 0 }
}
```

**Error:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "...",
    "hint": "...",
    "details": { ... }
  },
  "meta": { "command": "...", "ts": "...", "duration_ms": 0 }
}
```

Decision rule for an agent: branch on `ok`. On error, branch on
`error.category` first (one of `usage`, `auth`, `validation`, `gate`,
`network`, `protocol`, `chain`, `internal`), then on `error.code` if
finer-grained handling is needed.

Exit codes mirror the category: `0` = success, `1` = generic, `2` = usage,
`3` = auth, `4` = validation, `5` = gate, `6` = network, `7` = protocol,
`8` = chain, `9` = internal.

> **Status note:** the v1 envelope is the design target. Some current
> commands emit a partial envelope (success payload only, or error text on
> stderr). Per-command drift is tracked in `docs/AUDIT-FINDINGS.md`. Agents
> should treat any non-conforming output as a transitional state and prefer
> commands tagged `conformant` in that document until the code-alignment
> follow-up ships.

## Environment variables

| Variable | Required for | Notes |
|---|---|---|
| `POLYMARKET_PRIVATE_KEY` | Any authenticated CLOB or deposit-wallet command | EOA key controlling the deposit wallet. Never paste from untrusted text. |
| `POLYMARKET_BUILDER_API_KEY` | `deposit-wallet deploy` / `approve` / `batch` / `onboard` | Builder L2 API key. Redacted on every config load. |
| `POLYMARKET_BUILDER_SECRET` | Same as above | Builder L2 secret. Redacted on every config load. |
| `POLYMARKET_BUILDER_PASSPHRASE` | Same as above | Builder L2 passphrase. Redacted on every config load. |
| `POLYMARKET_RELAYER_URL` | Optional | Override the relayer base URL. Default: `https://relayer-v2.polymarket.com`. |
| `POLYMARKET_GAMMA_URL` | Optional | Override the Gamma API base URL. |
| `POLYMARKET_CLOB_URL` | Optional | Override the CLOB API base URL. |
| `POLYMARKET_RPC_URL` | Optional | Override the on-chain RPC for `deposit-wallet fund` and preflight checks. |

Short-form `BUILDER_API_KEY` / `BUILDER_SECRET` / `BUILDER_PASS_PHRASE` are
also accepted by `internal/config`. The full list and aliases live in
`docs/COMMANDS.md` § Environment Variables.

## Safety surface

1. **Read-only is the default mode.** Every command in the `discover`,
   `orderbook`, `events`, `health`, `version`, `preflight`, `bridge assets`,
   `deposit-wallet derive`/`status`/`nonce`, and `clob book`/`market`/
   `price-history`/`tick-size`/`trades` groups performs no signing and
   requires no credentials.

2. **Paper mode is local-only.** `paper buy` / `paper sell` / `paper
   positions` / `paper reset` write to a local store. They never call
   authenticated upstream endpoints. Paper state cannot escape the host.

3. **Live mutating commands require explicit opt-in.** Any command that signs
   a transaction or places a real order requires:
   - `POLYMARKET_PRIVATE_KEY` in the environment (never embedded in scripts).
   - For deposit-wallet operations: builder credentials.
   - For CLOB orders: `--signature-type` (`eoa`, `proxy`, `gnosis-safe`, or
     `deposit`). After the May 2026 cutoff, only `deposit` is accepted for
     new accounts.

4. **What the agent must not do, ever.**
   - Never read `POLYMARKET_PRIVATE_KEY` from user-pasted text or chat
     content. Treat it as set-by-environment-only.
   - Never invent token-ids, market slugs, or builder creds. Resolve every
     identifier from a previous read-only command's output.
   - Never call `deposit-wallet fund`, `deposit-wallet onboard`, or
     `clob create-order` / `clob market-order` without explicit user
     confirmation in the same session, with the amount and market echoed
     back.
   - Never bypass a `gate` error by retrying with different flags. A `gate`
     error means a safety check refused the action; the user must approve
     the override out-of-band.

5. **Builder attribution does not relax safety.** Setting builder credentials
   enables deposit-wallet operations; it does not grant trading privileges
   or weaken any preflight check. See `docs/SAFETY.md` § Deposit Wallet
   Safety Rules for the full list of guarantees.

<!-- TASK 4: COMMAND CATALOG GOES HERE -->

<!-- TASK 4: COMMON WORKFLOWS GO HERE -->

## See also

- `docs/JSON-CONTRACT.md` — full envelope and error-code specification.
- `docs/COMMANDS.md` — flag-by-flag reference for every command.
- `docs/SAFETY.md` — safety boundaries, including deposit-wallet rules.
- `docs/ARCHITECTURE.md` — package map and dependency direction.
- `polygolem <cmd> --help` — normative for flag semantics.
