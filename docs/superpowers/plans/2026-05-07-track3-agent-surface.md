# Track 3 — Agent Surface: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce three documentation deliverables that turn polygolem into an agent-driveable surface: (1) the canonical `docs/JSON-CONTRACT.md` v1 envelope spec with full error-code taxonomy and exit-code matrix, (2) a populated `Track 3 — JSON envelope drift` section in `docs/AUDIT-FINDINGS.md`, and (3) a complete `SKILL.md` rewrite covering every command in the binary.

**Architecture:** Documentation-only. No Go source modifications. Three artifacts, four commits. The audit reads the binary's `--json` output (or its absence) per command and records drift; the spec is written from the design doc verbatim; the SKILL.md rewrite walks the same command tree Track 1 produced.

**Tech Stack:** Plain markdown edits. Bash to walk the command tree, run commands with `--json`, and grep verification. No code changes.

**Spec:** `docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § Track 3.

**Out of scope (carry forward):** Code changes to align command outputs to the v1 envelope. Findings are captured in `docs/AUDIT-FINDINGS.md` Track 3 section. Code-alignment is a separate follow-up plan after this one. Track 4 owns the docs-site `reference/json-contract.mdx` and `reference/error-codes.mdx`; this track produces only `docs/JSON-CONTRACT.md`.

---

## Working tree state

The working tree on `main` has uncommitted WIP across Go source files. Track 3 is documentation-only. Each task uses an explicit per-task allowlist; implementers `git add` only the listed paths. **Never** `git add -A`, `git add .`, or `git commit -a`.

---

## Track 3 dependencies

- Track 1 produced canonical `docs/COMMANDS.md` with all 50 command paths and the working-notes scaffold `docs/AUDIT-FINDINGS.md` containing an empty `Track 3 — JSON envelope drift` table.
- The ground-truth command snapshot from Track 1 lives at `/tmp/polygolem-cmds.txt` (50 command paths). Task 2 of this plan re-walks the binary if that file is missing or stale.

---

## Task Inventory

| # | Task | Output |
|---|---|---|
| 1 | Write `docs/JSON-CONTRACT.md` (envelope, codes, exits, versioning) | New file |
| 2 | Audit current `--json` output, populate AUDIT-FINDINGS Track 3 section | Modified file |
| 3 | Rewrite `SKILL.md` part A — frontmatter, overview, JSON contract reference, env-vars, safety surface | Refreshed file |
| 4 | Rewrite `SKILL.md` part B — full command catalog (12 groups, 50 paths) and common workflows | Completed file |
| 5 | Final Track 3 verification gate | Green |

Tasks 1–4 each produce exactly one commit. Task 5 produces no commit.

---

## Task 1: Write `docs/JSON-CONTRACT.md`

**Files:**
- Create: `docs/JSON-CONTRACT.md`

**File allowlist (commit):** `docs/JSON-CONTRACT.md`

**Why first:** SKILL.md (Tasks 3–4) and the audit table (Task 2) both reference error codes and the envelope shape. The contract must exist first so downstream artifacts can link to it without forward references.

- [ ] **Step 1: Confirm the file does not yet exist**

```bash
test ! -f docs/JSON-CONTRACT.md && echo "OK to create" || echo "ALREADY EXISTS — stop and report"
```

Expected: `OK to create`. If `ALREADY EXISTS`, stop and report — do not silently overwrite.

- [ ] **Step 2: Write `docs/JSON-CONTRACT.md` with the exact content below**

````markdown
# JSON Contract — v1

`polygolem` emits a single, versioned JSON envelope for every command invoked
with `--json`. This document is the canonical specification. Agent integrators,
shell scripts, and downstream consumers should treat this file as authoritative.

The envelope is designed so a caller can answer three questions without
parsing the inner payload:

1. Did it succeed? (`ok`)
2. What version of the contract is this? (`version`)
3. If it failed, what category and code? (`error.category`, `error.code`)

The exit code mirrors the error category for shell-script convenience.

## Versioning policy

- `version` is a string. Current value: `"1"`.
- **Non-breaking changes** that do not bump `version`: adding new top-level
  commands, adding new fields inside `data`, adding new error `code` strings,
  adding new fields inside `error.details`.
- **Breaking changes** that bump `version`: renaming or removing any envelope
  field, renaming or removing an existing error `code`, changing the type of
  an existing field, changing the meaning of an existing exit code.
- Future minor versions parse as `"1.1"`, `"1.2"`, etc. Major versions parse
  as `"2"`, `"3"`. Callers should accept any `version` whose major component
  matches the major they were written against.

## Success envelope

```json
{
  "ok": true,
  "version": "1",
  "data": { ... command-specific payload ... },
  "meta": {
    "command": "deposit-wallet onboard",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 2143
  }
}
```

Field semantics:

- `ok` — boolean, always `true` on success. The single field a caller checks.
- `version` — string, the contract version (currently `"1"`).
- `data` — object, command-specific payload. Field naming inside `data`
  preserves whatever the underlying upstream API returned. Polygolem does not
  re-case upstream payloads.
- `meta.command` — the full command path as the user invoked it
  (e.g., `"clob create-order"`, not just `"create-order"`).
- `meta.ts` — RFC 3339 / ISO 8601 timestamp in UTC, second precision or finer.
- `meta.duration_ms` — integer, wall-clock duration of the invocation in
  milliseconds.

## Error envelope

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "POLYMARKET_BUILDER_API_KEY is required for deposit-wallet onboard",
    "hint": "Create builder creds at polymarket.com/settings?tab=builder",
    "details": { ... optional structured context ... }
  },
  "meta": { "command": "deposit-wallet onboard", "ts": "2026-05-07T12:34:56Z", "duration_ms": 12 }
}
```

Field semantics:

- `ok` — boolean, always `false` on error.
- `version` — same as success envelope.
- `error.code` — `SCREAMING_SNAKE_CASE` stable string. See the taxonomy below.
- `error.category` — one of the eight categories defined in the next section.
- `error.message` — human-readable, single-sentence explanation suitable for
  surfacing to a user.
- `error.hint` — optional, single-sentence suggestion for resolution. Omit
  when no actionable hint exists.
- `error.details` — optional object with structured context (e.g., the
  offending flag name, the upstream HTTP status, the on-chain revert reason).
  Field shape varies by code.
- `meta` — always present, identical to success envelope.

`data` and `error` are mutually exclusive. A success envelope never carries
`error`; an error envelope never carries `data`.

## Field-naming convention

Envelope fields use `snake_case` (e.g., `duration_ms`, `version`). Inside
`data`, fields preserve upstream casing — Polymarket's Gamma API mixes
`camelCase` and `snake_case`, and polygolem does not re-shape upstream
payloads. Callers can rely on `snake_case` for the envelope itself but must
read each `data` payload's fields by the names the upstream API uses.

## Error code taxonomy

Two-level: `category` (broad) and `code` (specific). Categories are closed;
codes are open (new codes can be added without a `version` bump).

### `usage` — bad flags, missing args, conflicting options

| Code | Description |
|---|---|
| `USAGE_FLAG_MISSING` | A required flag was not provided on the command line. |
| `USAGE_FLAG_INVALID` | A flag was given a value the command does not accept (e.g., `--limit -1`). |
| `USAGE_FLAG_CONFLICT` | Two mutually exclusive flags were provided together (e.g., `--id` and `--slug`). |
| `USAGE_SUBCOMMAND_UNKNOWN` | A subcommand path does not resolve to a registered command. |

### `auth` — missing or invalid credentials, signature failure

| Code | Description |
|---|---|
| `AUTH_PRIVATE_KEY_MISSING` | `POLYMARKET_PRIVATE_KEY` is required by the command but not set or empty. |
| `AUTH_BUILDER_MISSING` | One or more of `POLYMARKET_BUILDER_API_KEY` / `_SECRET` / `_PASSPHRASE` is required but not set. |
| `AUTH_SIG_INVALID` | A signature (EIP-712, POLY_1271, or ERC-7739) failed local validation before submission. |
| `AUTH_API_KEY_REJECTED` | The CLOB API rejected an L2 API key (expired, revoked, or wrong account). |

### `validation` — input parse or range failures

| Code | Description |
|---|---|
| `VALIDATION_TOKEN_ID_INVALID` | `--token-id` is not a valid CLOB token id (wrong length, non-numeric, or unknown). |
| `VALIDATION_AMOUNT_OUT_OF_RANGE` | A numeric amount is below the minimum or above the maximum the command accepts. |
| `VALIDATION_PRICE_TICK_MISMATCH` | A submitted order price does not align with the market's tick size. |
| `VALIDATION_MARKET_IDENTIFIER_AMBIGUOUS` | A market identifier (id, slug, or token-id) matches zero or multiple markets. |

### `gate` — mode, safety, preflight, or risk refusal

| Code | Description |
|---|---|
| `GATE_LIVE_DISABLED` | Live mode is required for the command but the binary is running in read-only or paper mode. |
| `GATE_PREFLIGHT_FAILED` | A `preflight` check (RPC reachability, balance, allowance, or wallet status) failed. |
| `GATE_RISK_LIMIT` | A per-trade or daily risk cap was exceeded; the command refused to proceed. |
| `GATE_FUNDING_INSUFFICIENT` | The deposit wallet or EOA does not hold the funds the command would move. |

### `network` — HTTP failure, DNS, timeout, retry exhaustion

| Code | Description |
|---|---|
| `NETWORK_TIMEOUT` | A request exceeded its deadline; the underlying transport gave up. |
| `NETWORK_DNS` | DNS resolution for an upstream host failed. |
| `NETWORK_TLS` | TLS handshake or certificate verification failed against an upstream host. |
| `NETWORK_RETRY_EXHAUSTED` | The transport's retry budget was consumed without a successful response. |

### `protocol` — upstream API errored or returned an unexpected shape

| Code | Description |
|---|---|
| `PROTOCOL_GAMMA_4XX` | The Gamma API returned a 4xx response (bad request, auth, not found, or rate-limited). |
| `PROTOCOL_CLOB_5XX` | The CLOB API returned a 5xx response (server error, unavailable, or gateway timeout). |
| `PROTOCOL_RELAYER_REJECTED` | The builder relayer rejected a `WALLET-CREATE` or `WALLET` batch submission. |
| `PROTOCOL_UNEXPECTED_SHAPE` | An upstream response could not be decoded into the expected schema. |

### `chain` — on-chain RPC, nonce, or revert errors

| Code | Description |
|---|---|
| `CHAIN_NONCE_TOO_LOW` | The submitted transaction's nonce is below the account's current nonce on the destination chain. |
| `CHAIN_REVERTED` | A simulated or submitted transaction reverted; the revert reason is in `details.revert_reason` if known. |
| `CHAIN_INSUFFICIENT_FUNDS` | The sending account does not hold enough native gas token or ERC-20 balance to complete the call. |
| `CHAIN_RPC_UNAVAILABLE` | The configured RPC endpoint did not respond, or returned a transport-level error. |

### `internal` — bug; an invariant the code expects was violated

| Code | Description |
|---|---|
| `INTERNAL_PANIC` | A goroutine panicked and was recovered by the top-level handler; this is always a bug. |
| `INTERNAL_INVARIANT` | A `should-never-happen` precondition was violated (e.g., a non-nil expected value was nil). |
| `INTERNAL_UNIMPLEMENTED` | A code path was reached that has not yet been wired up; please file an issue. |
| `INTERNAL_STATE_CORRUPT` | Persistent state on disk (e.g., paper-mode positions) failed integrity checks. |

## Exit-code matrix

Wrapping shell scripts can branch on exit code without parsing JSON. The JSON
envelope remains authoritative; exit codes are a convenience.

| Exit | Meaning | Mapping |
|---|---|---|
| `0` | success | `ok: true` |
| `1` | generic | category did not map cleanly, or no envelope was emitted |
| `2` | usage | `error.category == "usage"` |
| `3` | auth | `error.category == "auth"` |
| `4` | validation | `error.category == "validation"` |
| `5` | gate | `error.category == "gate"` |
| `6` | network | `error.category == "network"` |
| `7` | protocol | `error.category == "protocol"` |
| `8` | chain | `error.category == "chain"` |
| `9` | internal | `error.category == "internal"` |

## Worked examples

### Success — `polygolem discover search --query "btc" --limit 1 --json`

```json
{
  "ok": true,
  "version": "1",
  "data": {
    "markets": [
      {
        "id": "0xabc...",
        "slug": "will-btc-be-above-100k",
        "question": "Will BTC be above $100k by end of 2026?",
        "active": true
      }
    ],
    "count": 1
  },
  "meta": {
    "command": "discover search",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 412
  }
}
```

### Error — `polygolem deposit-wallet onboard --fund-amount 10 --json` without builder creds

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "AUTH_BUILDER_MISSING",
    "category": "auth",
    "message": "POLYMARKET_BUILDER_API_KEY is required for deposit-wallet onboard",
    "hint": "Create builder creds at polymarket.com/settings?tab=builder",
    "details": { "missing_vars": ["POLYMARKET_BUILDER_API_KEY"] }
  },
  "meta": {
    "command": "deposit-wallet onboard",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 12
  }
}
```

Exit code: `3` (`auth`).

### Error — `polygolem orderbook get --token-id "" --json`

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "VALIDATION_TOKEN_ID_INVALID",
    "category": "validation",
    "message": "--token-id is required and must be a non-empty CLOB token id",
    "hint": "Pass --token-id from `polygolem discover market` output"
  },
  "meta": {
    "command": "orderbook get",
    "ts": "2026-05-07T12:34:56Z",
    "duration_ms": 3
  }
}
```

Exit code: `4` (`validation`).

## Implementation status

The v1 envelope is the design target. Current command output partially
conforms; per-command drift from this contract is tracked in
`docs/AUDIT-FINDINGS.md` (Track 3 — JSON envelope drift section). Code
alignment is a separate follow-up project, not part of this documentation
overhaul.

## Source of truth

This file is canonical for the envelope shape, the eight error categories,
the listed example codes, and the exit-code matrix. The Starlight pages
`reference/json-contract.mdx` and `reference/error-codes.mdx` (built in
Track 4) link to this file. `SKILL.md` references this file for envelope
syntax instead of duplicating it.
````

- [ ] **Step 3: Verify internal consistency — every code referenced in examples is defined in the taxonomy**

```bash
codes_used=$(grep -oE '"code": "[A-Z_]+"' docs/JSON-CONTRACT.md | sort -u | sed 's/"code": "//;s/"//')
for c in $codes_used; do
  count=$(grep -cE "\\\`$c\\\`" docs/JSON-CONTRACT.md)
  if [ "$count" -lt 1 ]; then
    echo "UNDEFINED: $c referenced in example but not in taxonomy"
  fi
done
echo "DONE"
```

Expected: only `DONE`. No `UNDEFINED:` lines.

- [ ] **Step 4: Verify all eight categories are present with at least 4 codes each**

```bash
for cat in usage auth validation gate network protocol chain internal; do
  count=$(awk "/^### \\\`$cat\\\`/,/^### / { if (/^\| \\\`[A-Z_]+\\\` \|/) print }" docs/JSON-CONTRACT.md | wc -l)
  printf "%-12s %d codes\n" "$cat" "$count"
  if [ "$count" -lt 4 ]; then
    echo "FAIL: $cat has fewer than 4 codes"
  fi
done
```

Expected: every category prints `4 codes` (or more). No `FAIL:` lines.

- [ ] **Step 5: Verify exit-code matrix has 10 rows (0–9)**

```bash
awk '/^## Exit-code matrix/,/^## /' docs/JSON-CONTRACT.md | grep -cE '^\| `[0-9]` \|'
```

Expected: `10`.

- [ ] **Step 6: Commit**

```bash
git add docs/JSON-CONTRACT.md
git commit -m "$(cat <<'EOF'
docs: add JSON-CONTRACT.md — canonical v1 envelope spec

Defines the success and error envelopes, the eight error categories
(usage, auth, validation, gate, network, protocol, chain, internal) with
four example codes each (32 total), the exit-code matrix mapping
categories to shell exit codes 2-9, and the versioning policy.

Canonical reference for SKILL.md (Track 3) and the Starlight reference
pages (Track 4). Code alignment to the v1 envelope is a separate
follow-up plan; current drift is tracked in AUDIT-FINDINGS.md.

Part of Track 3 (Agent Surface) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Audit current `--json` output, populate AUDIT-FINDINGS Track 3 section

**Files:**
- Modify: `docs/AUDIT-FINDINGS.md` (only the `Track 3 — JSON envelope drift` section table)

**File allowlist (commit):** `docs/AUDIT-FINDINGS.md`

**Why now:** The contract from Task 1 is the target. SKILL.md (Tasks 3–4) needs to know which commands currently emit envelope-shaped output and which do not, so the per-command examples in SKILL.md can be flagged or annotated accurately.

### Audit methodology

For each command path in the binary (50 paths total — 12 top-level groups plus 38 subcommands):

1. Run the command with `--json` and minimal/typical input.
2. Capture stdout + stderr + exit code.
3. Decode stdout as JSON. If it does not decode, record `not-json`.
4. If it decodes, compare against the v1 envelope: does it have `ok`, `version`, `data`/`error`, `meta`?
5. For commands that require credentials we don't have (anything in `auth`, `clob` mutating, `deposit-wallet` mutating, `live`, `bridge deposit`), we expect an auth-side failure. Capture **the failure response shape** — the question is whether the failure path is envelope-conformant, not whether the command succeeds.
6. Record one row per command in the AUDIT-FINDINGS Track 3 table.

The audit is descriptive, not prescriptive — it captures current state. Code alignment is a separate plan.

### Drift category shorthand (for the table)

- **conformant** — output already matches v1 (`ok` + `version` + `data`/`error` + `meta`).
- **partial-success** — success path emits some JSON but missing one or more required envelope fields.
- **not-json** — command does not emit JSON at all (plain text, table, or empty).
- **error-not-json** — error path prints to stderr without JSON; exit code is non-zero but no envelope.
- **error-untyped** — error path emits JSON but lacks `error.code` / `error.category`.
- **flag-missing** — command does not accept `--json` (the flag is missing from `--help`).

A single command may be tagged with multiple shorthand values (e.g., `partial-success; error-untyped`).

- [ ] **Step 1: Confirm the binary is built and the command snapshot exists**

```bash
test -x ./polygolem && echo "binary OK" || (echo "rebuilding"; go build -o polygolem ./cmd/polygolem)
test -s /tmp/polygolem-cmds.txt && echo "snapshot OK" || echo "MISSING: rerun Track 1 Task 1 first"
```

Expected: `binary OK` and `snapshot OK`. If the snapshot is missing, stop and rerun Track 1 Task 1 (`/tmp/polygolem-cmds.txt`).

- [ ] **Step 2: Build the audit driver script**

Write the following to `/tmp/track3-audit.sh` (a working file, not committed):

```bash
cat > /tmp/track3-audit.sh <<'AUDIT'
#!/usr/bin/env bash
# Track 3 audit driver. Walks every command path in /tmp/polygolem-cmds.txt,
# runs it with --json plus typical/minimal input, and prints one TSV row per
# command: <path>\t<exit>\t<stdout-shape>\t<envelope-fields>\t<notes>
set -uo pipefail
BIN=./polygolem
OUT=/tmp/track3-audit.tsv
> "$OUT"

run() {
  local cmd="$1"
  shift
  local out err rc
  out=$($BIN $cmd "$@" --json 2>/tmp/track3.err)
  rc=$?
  err=$(cat /tmp/track3.err)
  local shape="not-json"
  local fields=""
  if echo "$out" | jq -e . >/dev/null 2>&1; then
    shape="json"
    fields=$(echo "$out" | jq -r '[paths(scalars) | join(".")] | unique | join(",")' 2>/dev/null | head -c 300)
  fi
  printf "%s\t%s\t%s\t%s\t%s\n" "$cmd" "$rc" "$shape" "$fields" "${err:0:120}" >> "$OUT"
}

# Top-level / leaf commands with no required input
run "version"
run "health"
run "preflight"

# auth (read-only subcommand)
run "auth status"

# bridge — assets needs no input; deposit needs auth, will fail
run "bridge assets"
run "bridge deposit" --asset USDC --amount 1

# clob — read endpoints with token-id; mutating endpoints will fail without auth
TOKEN_ID="71321045679252212594626385532706912750332728571942532289631379312455583992563"
run "clob book" --token-id "$TOKEN_ID"
run "clob market" --token-id "$TOKEN_ID"
run "clob price-history" --token-id "$TOKEN_ID"
run "clob tick-size" --token-id "$TOKEN_ID"
run "clob trades" --token-id "$TOKEN_ID"
run "clob balance"
run "clob orders"
run "clob create-api-key"
run "clob create-order" --token-id "$TOKEN_ID" --side BUY --price 0.5 --size 10
run "clob market-order" --token-id "$TOKEN_ID" --side BUY --amount 10
run "clob update-balance"

# deposit-wallet — derive/status/nonce read-only; deploy/approve/batch/fund/onboard need builder creds
run "deposit-wallet derive"
run "deposit-wallet status"
run "deposit-wallet nonce"
run "deposit-wallet deploy"
run "deposit-wallet approve"
run "deposit-wallet batch" --calls-json '[]'
run "deposit-wallet fund" --amount 1
run "deposit-wallet onboard" --fund-amount 1

# discover — all read-only
run "discover search" --limit 1
run "discover market" --slug "will-btc-be-above-100k"
run "discover enrich" --slug "will-btc-be-above-100k"

# events
run "events list" --limit 1

# live — status only (no real live trading)
run "live status"

# orderbook — all need token-id
run "orderbook get" --token-id "$TOKEN_ID"
run "orderbook fee-rate" --token-id "$TOKEN_ID"
run "orderbook last-trade" --token-id "$TOKEN_ID"
run "orderbook midpoint" --token-id "$TOKEN_ID"
run "orderbook price" --token-id "$TOKEN_ID"
run "orderbook spread" --token-id "$TOKEN_ID"
run "orderbook tick-size" --token-id "$TOKEN_ID"

# paper — local-only simulation
run "paper positions"
run "paper buy" --token-id "$TOKEN_ID" --price 0.5 --size 10
run "paper sell" --token-id "$TOKEN_ID" --price 0.5 --size 10
run "paper reset"

echo "wrote $OUT"
wc -l "$OUT"
AUDIT
chmod +x /tmp/track3-audit.sh
```

Verify:

```bash
test -x /tmp/track3-audit.sh && echo "OK"
```

Expected: `OK`.

- [ ] **Step 3: Run the audit driver**

```bash
/tmp/track3-audit.sh
```

Expected: prints `wrote /tmp/track3-audit.tsv` and a line count of at least `35`. If the binary is missing or `jq` is unavailable, stop and report.

- [ ] **Step 4: Inspect the captured shapes**

```bash
column -t -s $'\t' /tmp/track3-audit.tsv | head -60
```

Expected: a TSV-aligned table with one row per command. The `shape` column is `json` or `not-json`; the `fields` column lists the dotted paths found in the output.

For each row, classify the drift using the shorthand (see methodology above). Note the rows where:

- `shape == "not-json"` and `rc == 0` → `not-json` for the success path.
- `shape == "not-json"` and `rc != 0` → `error-not-json`.
- `shape == "json"` and the `fields` column is missing `ok` or `version` or `meta.command` → `partial-success` or `error-untyped`.
- `--json` produces a usage error from cobra → `flag-missing`.

- [ ] **Step 5: Append one row per command to `docs/AUDIT-FINDINGS.md`**

The Track 3 section already has table headers from Track 1 Task 3. Append rows under the existing `| Command | Current shape | Drift from v1 envelope |` header. Sort rows by command path (alphabetical, top-level groups first, then subcommands).

Use this exact row format. Replace `<...>` with the audit findings:

```markdown
| `<full command path>` | <one-line description of stdout shape, e.g., "raw payload object", "table only", "stderr text + exit 1"> | <drift shorthand>; <one-sentence specifics> |
```

#### Sample rows (use as the format template; replace with actual findings)

```markdown
| `version` | bare object `{"version":"...","commit":"..."}` | partial-success; missing `ok`, `version` (envelope), `meta`. Inner `version` collides with envelope field name. |
| `discover search` | array of market objects under top-level | partial-success; success payload is an array, no envelope wrapper at all. |
| `deposit-wallet onboard` | stderr text "POLYMARKET_BUILDER_API_KEY required"; exit 1 | error-not-json; expected `error.code: AUTH_BUILDER_MISSING`, exit 3. |
| `orderbook get` | bare CLOB book payload | partial-success; missing envelope, exit code on bad token-id is 1 not 4. |
| `paper buy` | bare position object | partial-success; missing envelope. |
```

The implementer fills in real findings from `/tmp/track3-audit.tsv`. Every command in `/tmp/polygolem-cmds.txt` gets exactly one row.

- [ ] **Step 6: Verify every command path has a row**

```bash
expected=$(grep '^=====' /tmp/polygolem-cmds.txt | sed -E 's/^===== (.+) =====$/\1/' | sort -u)
while IFS= read -r cmd; do
  grep -qF "\`$cmd\`" docs/AUDIT-FINDINGS.md || echo "MISSING: $cmd"
done <<< "$expected"
```

Expected: no `MISSING:` lines. The Track 3 table covers all 50 command paths.

- [ ] **Step 7: Verify table integrity**

```bash
awk '/^## Track 3 — JSON envelope drift/,/^## Findings consumed downstream/' docs/AUDIT-FINDINGS.md \
  | grep -cE '^\| `[a-z]'
```

Expected: at least `50` (one row per command path; allow more if a command needs split rows for success vs error paths).

- [ ] **Step 8: Commit**

```bash
git add docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs: populate AUDIT-FINDINGS Track 3 section with per-command drift

Walks all 50 command paths in the polygolem binary, runs each with --json
and typical/minimal input, and records the current stdout shape plus its
drift from the v1 envelope spec'd in docs/JSON-CONTRACT.md. Auth-required
commands without credentials capture the failure shape, which is the
relevant question for envelope conformance.

Findings feed the separate code-alignment follow-up plan; this commit is
documentation only.

Part of Track 3 (Agent Surface) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Rewrite `SKILL.md` part A — frame, contract reference, env-vars, safety

**Files:**
- Modify: `SKILL.md` (replace existing content with the structure below; the command catalog body is filled in by Task 4)

**File allowlist (commit):** `SKILL.md`

**Why split:** SKILL.md grows large because it covers 50 command paths. Splitting the rewrite into part A (frame + cross-cutting reference) and part B (catalog + workflows) keeps each commit reviewable and lets the verification gate at Task 5 check the structural pieces independently.

- [ ] **Step 1: Read the current SKILL.md so you know what content (if any) is worth preserving**

```bash
wc -l SKILL.md
cat SKILL.md
```

Note: the current SKILL.md is 103 lines and covers only the read-only subset (`discover`, `orderbook`, `health`, `version`, `preflight`). Task 3 + Task 4 replace it entirely. Preserve nothing — we are starting from the v1 envelope and the full command tree.

- [ ] **Step 2: Replace `SKILL.md` with the part-A scaffold below**

The scaffold has placeholder markers (`<!-- TASK 4: ... -->`) where Task 4 inserts the command catalog and workflows. Task 4 must replace those markers with content; the verification gate (Task 5) refuses commits that leave them in place.

Exact content:

````markdown
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
````

- [ ] **Step 3: Verify the scaffold has the expected structure**

```bash
grep -cE '^## ' SKILL.md
```

Expected: at least `5` (Prerequisites, JSON contract reference, Environment
variables, Safety surface, See also). Task 4 will add 2 more (Command
catalog, Common workflows).

```bash
grep -cE '<!-- TASK 4:' SKILL.md
```

Expected: `2`. Both placeholder markers must be present so Task 4 has a
defined insertion point.

```bash
grep -q "v1 envelope" SKILL.md && grep -q "snake_case" SKILL.md || true
grep -q "AUTH_BUILDER_MISSING" SKILL.md && echo "envelope refs OK"
```

Expected: `envelope refs OK`.

- [ ] **Step 4: Commit**

```bash
git add SKILL.md
git commit -m "$(cat <<'EOF'
docs: rewrite SKILL.md frame — JSON contract, env vars, safety surface

Replaces the read-only-only SKILL.md (103 lines, 5 commands) with the
agent-spec frame: project overview, v1 JSON envelope summary linking to
docs/JSON-CONTRACT.md, full env-var reference, and a safety-surface
section enumerating read-only defaults, paper-mode locality, live opt-in
requirements, and explicit must-not rules for the agent.

Command catalog and common workflows are inserted by Task 4 at the
TASK 4 placeholder markers.

Part of Track 3 (Agent Surface) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Rewrite `SKILL.md` part B — full command catalog and common workflows

**Files:**
- Modify: `SKILL.md` (replace the two TASK 4 placeholder markers with full content)

**File allowlist (commit):** `SKILL.md`

### Catalog format (mirror exactly for every command-group section)

For each of the 12 command groups, add one `## Command catalog — <group>`
section. Inside it, every command path in that group gets one `### <full
path>` block with the following exact subheadings:

- **Purpose** — one sentence.
- **Required flags** — bullet list. Empty list ("None.") if the command
  takes no required flags.
- **Env vars consumed** — bullet list of variables actually read by the
  command. "None." for read-only public-data commands.
- **Sample success JSON** — fenced JSON block with the v1 envelope shape.
  Use `...` inside `data` for noisy fields. If the command currently emits
  a partial envelope, mark the block ` ```json title="target v1 envelope" `
  and add a one-line note pointing to the AUDIT-FINDINGS row.
- **Sample error JSON** — fenced JSON block with the most likely failure
  for that command (auth-missing for credentialed commands;
  validation-invalid for input-bound commands).
- **Caveats** — bullet list. One bullet per real-world gotcha. Use "None."
  if there are no caveats.

The order inside each group: top-level group entry first (`### <group>` with
no path; describes the group as a whole), then subcommands alphabetically.

### Worked example — the `discover` group

This is the EXACT format Task 4 must mirror for the other 11 groups. Insert
this content at the first `<!-- TASK 4: COMMAND CATALOG GOES HERE -->`
marker (followed by sections for the other 11 groups in this order:
`orderbook`, `clob`, `deposit-wallet`, `paper`, `bridge`, `events`,
`health`, `version`, `preflight`, `auth`, `live`).

````markdown
## Command catalog

### Command catalog — `discover`

Public Gamma + CLOB market discovery. Read-only. No credentials.

#### `discover search`

**Purpose:** Search active Polymarket markets by free-text query.

**Required flags:** None. (Common optional flags: `--query`, `--limit`,
`--active`.)

**Env vars consumed:** None.

**Sample success JSON:**

```json
{
  "ok": true,
  "version": "1",
  "data": {
    "markets": [
      {
        "id": "0xabc...",
        "slug": "will-btc-be-above-100k",
        "question": "Will BTC be above $100k by end of 2026?",
        "active": true,
        "clobTokenIds": ["7132...", "1024..."]
      }
    ],
    "count": 1
  },
  "meta": { "command": "discover search", "ts": "2026-05-07T12:34:56Z", "duration_ms": 412 }
}
```

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "NETWORK_TIMEOUT",
    "category": "network",
    "message": "Gamma API request exceeded 10s deadline",
    "hint": "Retry; if persistent, check `polygolem health`."
  },
  "meta": { "command": "discover search", "ts": "2026-05-07T12:34:56Z", "duration_ms": 10042 }
}
```

**Caveats:**

- `--query` matches against the market `question` and `slug`; it is not a
  full-text search of event descriptions.
- `--limit` defaults to a small page; for exhaustive scans use the
  pagination helpers in `pkg/pagination` from Go code instead of repeated
  CLI calls.

#### `discover market`

**Purpose:** Fetch a single market by id, slug, or CLOB token id.

**Required flags:** Exactly one of `--id`, `--slug`, or `--token-id`.

**Env vars consumed:** None.

**Sample success JSON:**

```json
{
  "ok": true,
  "version": "1",
  "data": {
    "id": "0xabc...",
    "slug": "will-btc-be-above-100k",
    "question": "Will BTC be above $100k by end of 2026?",
    "active": true,
    "clobTokenIds": ["7132...", "1024..."],
    "outcomes": ["Yes", "No"]
  },
  "meta": { "command": "discover market", "ts": "2026-05-07T12:34:56Z", "duration_ms": 188 }
}
```

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "VALIDATION_MARKET_IDENTIFIER_AMBIGUOUS",
    "category": "validation",
    "message": "Pass exactly one of --id, --slug, or --token-id",
    "hint": "Use `discover search` first to find the canonical slug."
  },
  "meta": { "command": "discover market", "ts": "2026-05-07T12:34:56Z", "duration_ms": 4 }
}
```

**Caveats:**

- For CLOB-tradeable markets, `clobTokenIds` is required input for any
  `orderbook` or `clob` command.
- A market can be in Gamma without being tradeable in CLOB; combine with
  `discover enrich` to confirm tradability.

#### `discover enrich`

**Purpose:** Combine a Gamma market with its CLOB metadata (tick size,
fee rate, neg-risk flag, current orderbook depth) for a full tradability
view.

**Required flags:** Exactly one of `--id` or `--slug`.

**Env vars consumed:** None.

**Sample success JSON:**

```json
{
  "ok": true,
  "version": "1",
  "data": {
    "market": { "id": "0xabc...", "slug": "...", "active": true },
    "clob": {
      "tickSize": "0.01",
      "feeRate": "0.02",
      "negRisk": false,
      "bestBid": 0.51,
      "bestAsk": 0.53
    },
    "tradeable": true
  },
  "meta": { "command": "discover enrich", "ts": "2026-05-07T12:34:56Z", "duration_ms": 612 }
}
```

**Sample error JSON:**

```json
{
  "ok": false,
  "version": "1",
  "error": {
    "code": "PROTOCOL_GAMMA_4XX",
    "category": "protocol",
    "message": "Gamma returned 404 for slug 'unknown-slug'",
    "details": { "status": 404 }
  },
  "meta": { "command": "discover enrich", "ts": "2026-05-07T12:34:56Z", "duration_ms": 142 }
}
```

**Caveats:**

- `tradeable: false` does not always mean the market is closed; it can also
  mean the CLOB does not list it (paused, illiquid, or not yet onboarded).
- `tickSize` is decoded from CLOB and may be returned as a string or number
  upstream; the `data` payload preserves whatever the API returns.
````

### Sections to write (apply the same format)

Task 4 must add a `### Command catalog — <group>` section for **every**
group below, with one `### <full path>` subsection per command path. The
counts come from `/tmp/polygolem-cmds.txt`.

| Group | Subcommands to document |
|---|---|
| `discover` | `discover search`, `discover market`, `discover enrich` (already shown above; copy verbatim). |
| `orderbook` | `orderbook get`, `orderbook fee-rate`, `orderbook last-trade`, `orderbook midpoint`, `orderbook price`, `orderbook spread`, `orderbook tick-size`. |
| `clob` | `clob book`, `clob market`, `clob price-history`, `clob tick-size`, `clob trades`, `clob balance`, `clob orders`, `clob create-api-key`, `clob create-order`, `clob market-order`, `clob update-balance`. |
| `deposit-wallet` | `deposit-wallet derive`, `deposit-wallet status`, `deposit-wallet nonce`, `deposit-wallet deploy`, `deposit-wallet approve`, `deposit-wallet batch`, `deposit-wallet fund`, `deposit-wallet onboard`. |
| `paper` | `paper positions`, `paper buy`, `paper sell`, `paper reset`. |
| `bridge` | `bridge assets`, `bridge deposit`. |
| `events` | `events list`. |
| `health` | (single-command group; document `health` itself.) |
| `version` | (single-command group; document `version` itself.) |
| `preflight` | (single-command group; document `preflight` itself.) |
| `auth` | `auth status`. |
| `live` | `live status`. |

For commands that mutate or sign, the `Sample error JSON` block must show
the most-likely failure mode per [the JSON contract](#json-contract-reference):

- Auth-required commands without creds → `AUTH_PRIVATE_KEY_MISSING` or
  `AUTH_BUILDER_MISSING`.
- CLOB order commands without a signature type → `USAGE_FLAG_MISSING`.
- `deposit-wallet fund` with `--amount 0` → `VALIDATION_AMOUNT_OUT_OF_RANGE`.
- `deposit-wallet onboard` without a deployed wallet but with builder creds
  also reachable → success path with multi-step `data.steps` array.
- `clob create-order` with a price off the tick grid →
  `VALIDATION_PRICE_TICK_MISMATCH`.
- `live status` when live mode is disabled → `GATE_LIVE_DISABLED`.

For commands flagged in Task 2's audit as currently non-conformant, add
this exact note immediately after the `Sample success JSON` block:

```markdown
> **Drift:** current output is not yet envelope-conformant; see
> `docs/AUDIT-FINDINGS.md` (Track 3 row for `<full command path>`). The
> JSON above is the v1 target shape.
```

### Common workflows

Replace the second `<!-- TASK 4: COMMON WORKFLOWS GO HERE -->` marker with
this section. Use exactly this content:

````markdown
## Common workflows

### 1. Find a tradeable market for a topic

```bash
./polygolem discover search --query "btc" --limit 5 --json | jq '.data.markets[] | {slug, question}'
SLUG=$(./polygolem discover search --query "btc" --limit 1 --json | jq -r '.data.markets[0].slug')
./polygolem discover enrich --slug "$SLUG" --json | jq '{tradeable: .data.tradeable, tickSize: .data.clob.tickSize}'
```

Decision rule for the agent:

- `data.tradeable == true` and `data.clob.bestBid > 0` and
  `data.clob.bestAsk > 0` → market is tradeable.
- `data.tradeable == false` → stop; report to user and pick another
  market.

### 2. Check market depth before quoting a size

```bash
TOKEN_ID=$(./polygolem discover market --slug "$SLUG" --json | jq -r '.data.clobTokenIds[0]')
./polygolem orderbook get --token-id "$TOKEN_ID" --json | jq '{
  bids: .data.bids[0:3],
  asks: .data.asks[0:3],
  midpoint: .data.midpoint
}'
./polygolem orderbook spread --token-id "$TOKEN_ID" --json | jq '.data.spread'
```

The agent uses the top-of-book + midpoint to compute slippage tolerance
before submitting any order.

### 3. Onboard a new account end-to-end (deposit-wallet)

Pre-flight: the user must have set `POLYMARKET_PRIVATE_KEY`,
`POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`, and
`POLYMARKET_BUILDER_PASSPHRASE` in their environment **before** the agent
starts.

```bash
./polygolem preflight --json
./polygolem deposit-wallet derive --json
./polygolem deposit-wallet status --json
./polygolem deposit-wallet onboard --fund-amount 25 --json
```

Decision rule:

- `preflight` returns `ok: false` with `error.category == "gate"` → stop;
  surface the failed checks to the user.
- `deposit-wallet status` returns `data.deployed: true` → skip `onboard`
  deploy step; only `fund` is needed.
- `deposit-wallet onboard` requires explicit `--fund-amount`; the agent
  must echo the amount back to the user before running this command.

### 4. Place a paper trade (always safe to try first)

```bash
./polygolem paper buy --token-id "$TOKEN_ID" --price 0.50 --size 10 --json
./polygolem paper positions --json | jq '.data.positions'
```

Paper mode never touches authenticated endpoints. Use it as a dry-run
before any live order.

### 5. Place a live CLOB order (requires explicit confirmation)

```bash
# 1. confirm tradability
./polygolem orderbook tick-size --token-id "$TOKEN_ID" --json
# 2. quote round-trip first
./polygolem clob book --token-id "$TOKEN_ID" --json | jq '.data | {bestBid: .bids[0], bestAsk: .asks[0]}'
# 3. live order — REQUIRES user confirmation in the same turn
./polygolem clob create-order \
  --token-id "$TOKEN_ID" \
  --side BUY --price 0.50 --size 10 \
  --signature-type deposit \
  --json
```

Decision rule:

- Echo `--token-id`, `--side`, `--price`, `--size`, and `--signature-type`
  back to the user verbatim before running step 3.
- After the May 2026 cutoff, `--signature-type` other than `deposit` is
  rejected for new accounts; the agent must default to `deposit`.
- On `error.category == "gate"` or `"validation"`, the agent stops and
  reports — it does not retry with adjusted flags.
````

- [ ] **Step 1: Replace the first TASK 4 marker with the full command catalog**

Insert the discover-group section above verbatim, then add the other 11
groups using the same format. Each `### <full path>` subsection has
exactly the six subheadings (Purpose, Required flags, Env vars consumed,
Sample success JSON, Sample error JSON, Caveats).

For each command, draw flag information from `polygolem <cmd> --help`
(captured in `/tmp/polygolem-cmds.txt`). Do not invent flags; do not
omit required flags surfaced by `--help`.

- [ ] **Step 2: Replace the second TASK 4 marker with the workflows section**

Insert the workflows section above verbatim. Five workflows are required
at minimum: find-tradeable-market, check-depth, onboard, paper-trade,
live-order.

- [ ] **Step 3: Verify both placeholder markers are gone**

```bash
grep -c '<!-- TASK 4:' SKILL.md
```

Expected: `0`. If `1` or `2`, a marker was missed.

- [ ] **Step 4: Verify every command path from the binary appears in SKILL.md**

```bash
while IFS= read -r cmd; do
  grep -qF "### \`$cmd\`" SKILL.md || grep -qE "^### \`?$cmd\`?\$" SKILL.md || echo "MISSING: $cmd"
done < <(grep '^=====' /tmp/polygolem-cmds.txt | sed -E 's/^===== (.+) =====$/\1/' | sort -u)
```

Expected: no `MISSING:` lines. SKILL.md covers every command path the
binary exposes.

- [ ] **Step 5: Verify each documented command has the six required subheadings**

```bash
awk '/^### `/{cmd=$0; p=0; r=0; e=0; s=0; x=0; c=0; next}
     /^\*\*Purpose:\*\*/{p=1}
     /^\*\*Required flags:\*\*/{r=1}
     /^\*\*Env vars consumed:\*\*/{e=1}
     /^\*\*Sample success JSON:\*\*/{s=1}
     /^\*\*Sample error JSON:\*\*/{x=1}
     /^\*\*Caveats:\*\*/{c=1}
     /^### `/ && cmd!=""{
       if(!(p&&r&&e&&s&&x&&c)) print "INCOMPLETE:", cmd, p,r,e,s,x,c;
     }
     END{
       if(cmd!="" && !(p&&r&&e&&s&&x&&c)) print "INCOMPLETE:", cmd, p,r,e,s,x,c;
     }' SKILL.md
```

Expected: no `INCOMPLETE:` lines.

- [ ] **Step 6: Verify the workflows section is present and has at least five workflows**

```bash
grep -cE '^### [0-9]\.' SKILL.md
```

Expected: `5` or more (numbered workflow headings).

- [ ] **Step 7: Verify every error-code referenced in SKILL.md is defined in JSON-CONTRACT.md**

```bash
grep -oE '"code": "[A-Z_]+"' SKILL.md | sort -u | sed 's/"code": "//;s/"//' | while read c; do
  grep -qE "\\\`$c\\\`" docs/JSON-CONTRACT.md || echo "UNDEFINED: $c"
done
```

Expected: no `UNDEFINED:` lines. Every code an example uses must be
listed in the JSON-CONTRACT taxonomy.

- [ ] **Step 8: Commit**

```bash
git add SKILL.md
git commit -m "$(cat <<'EOF'
docs: complete SKILL.md catalog and workflows for all 50 commands

Replaces the two TASK 4 placeholder markers with the full command
catalog (12 groups, 50 command paths) and five canonical workflows
(find-tradeable-market, check-depth, onboard, paper-trade, live-order).

Each documented command has six subheadings: Purpose, Required flags,
Env vars consumed, Sample success JSON, Sample error JSON, Caveats.
Sample envelopes mirror docs/JSON-CONTRACT.md; commands flagged as
non-conformant in AUDIT-FINDINGS get a "Drift" note pointing to the
target v1 shape.

Part of Track 3 (Agent Surface) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Final Track 3 verification gate

**Files:** none modified — read-only verification.

This task does not produce a commit. It either passes (Track 3 done) or
identifies regressions to loop back on.

- [ ] **Step 1: `docs/JSON-CONTRACT.md` exists and is internally consistent**

```bash
test -s docs/JSON-CONTRACT.md && echo "exists"
codes_used=$(grep -oE '"code": "[A-Z_]+"' docs/JSON-CONTRACT.md | sort -u | sed 's/"code": "//;s/"//')
for c in $codes_used; do
  grep -qE "\\\`$c\\\`" docs/JSON-CONTRACT.md || echo "UNDEFINED: $c"
done
```

Expected: `exists` and no `UNDEFINED:` lines.

- [ ] **Step 2: All eight error categories defined with at least four codes each**

```bash
for cat in usage auth validation gate network protocol chain internal; do
  count=$(awk "/^### \\\`$cat\\\`/,/^### / { if (/^\| \\\`[A-Z_]+\\\` \|/) print }" docs/JSON-CONTRACT.md | wc -l)
  printf "%-12s %d\n" "$cat" "$count"
  [ "$count" -ge 4 ] || echo "FAIL: $cat"
done
```

Expected: every category prints `>= 4`. No `FAIL:` lines.

- [ ] **Step 3: Exit-code matrix has 10 rows (0–9)**

```bash
awk '/^## Exit-code matrix/,/^## /' docs/JSON-CONTRACT.md | grep -cE '^\| `[0-9]` \|'
```

Expected: `10`.

- [ ] **Step 4: AUDIT-FINDINGS Track 3 section has one row per command path**

```bash
expected=$(grep '^=====' /tmp/polygolem-cmds.txt | sed -E 's/^===== (.+) =====$/\1/' | sort -u)
while IFS= read -r cmd; do
  grep -qF "\`$cmd\`" docs/AUDIT-FINDINGS.md || echo "MISSING: $cmd"
done <<< "$expected"
```

Expected: no `MISSING:` lines.

- [ ] **Step 5: SKILL.md covers every command in `polygolem --help`**

```bash
while IFS= read -r cmd; do
  grep -qF "### \`$cmd\`" SKILL.md || grep -qE "^### \`?$cmd\`?\$" SKILL.md || echo "MISSING: $cmd"
done < <(grep '^=====' /tmp/polygolem-cmds.txt | sed -E 's/^===== (.+) =====$/\1/' | sort -u)
```

Expected: no `MISSING:` lines.

- [ ] **Step 6: SKILL.md placeholder markers are all replaced**

```bash
grep -c '<!-- TASK 4:' SKILL.md
```

Expected: `0`.

- [ ] **Step 7: Every error code referenced in SKILL.md is defined in JSON-CONTRACT.md**

```bash
grep -oE '"code": "[A-Z_]+"' SKILL.md | sort -u | sed 's/"code": "//;s/"//' | while read c; do
  grep -qE "\\\`$c\\\`" docs/JSON-CONTRACT.md || echo "UNDEFINED: $c"
done
```

Expected: no `UNDEFINED:` lines.

- [ ] **Step 8: Spot-check three command examples are runnable as documented**

Pick one read-only, one mutating (auth-required), one paper command. Run
each with the flags shown in its SKILL.md example block.

```bash
# read-only
./polygolem health --json >/tmp/spot-health.json && jq -e '.' /tmp/spot-health.json >/dev/null && echo "health: runs and emits JSON"

# auth-required (we expect it to fail, but the failure mode must match what SKILL.md documents)
./polygolem deposit-wallet status --json 2>&1 | head -c 400; echo
echo "deposit-wallet status: failure shape captured above; cross-check against SKILL.md sample error JSON"

# paper
./polygolem paper positions --json >/tmp/spot-paper.json 2>&1; head -c 400 /tmp/spot-paper.json; echo
echo "paper positions: output captured above; cross-check against SKILL.md sample success JSON"
```

Expected: `health: runs and emits JSON`. For the other two, manually
confirm that the captured output matches the structure SKILL.md documents
(or, where the audit flagged drift, that the SKILL.md block carries the
`> **Drift:**` note pointing to AUDIT-FINDINGS).

If any spot-check reveals a SKILL.md example that doesn't reflect the
documented behavior — and isn't already flagged with a Drift note — return
to Task 4 and fix the example or add the note.

- [ ] **Step 9: Build and tests are still green**

```bash
go build ./cmd/polygolem
go vet ./...
go test ./...
```

Expected: all three pass. Track 3 is documentation-only and must not have
regressed code state.

- [ ] **Step 10: If all checks pass, mark Track 3 complete**

No file change required. Inform the user that Track 3 verification has
passed and propose moving to Track 4 (docs-site) planning.

If any check fails, return to the relevant earlier task and fix in place
rather than papering over.

---

## Out of scope (re-stated)

- **Code changes to align command outputs to the v1 envelope.** Findings
  are captured in `docs/AUDIT-FINDINGS.md`. Code-alignment is its own
  brainstorm → spec → plan → implement cycle, not part of this plan.
- **Starlight pages** `reference/json-contract.mdx` and
  `reference/error-codes.mdx`. Owned by Track 4. They link to
  `docs/JSON-CONTRACT.md` produced here.
- **New error codes beyond the 32 listed.** Adding codes is non-breaking
  and can happen at any time; this plan defines the taxonomy floor, not
  the ceiling.
- **Changes to `docs/superpowers/specs/` content.** The spec is locked for
  this implementation cycle.
