# Track 1 — Audit & Truth Pass: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring every existing polygolem doc surface (`README.md`, `docs/*.md`) into alignment with current code reality. Establish a known-good baseline so subsequent tracks (godoc, agent surface, docs-site, repo polish) build on truth.

**Architecture:** Audit-driven. Build the binary, derive ground truth from `--help` and the source tree, fix lies in place, archive stale docs to `docs/history/`, delete one obsolete planning doc, capture all drift in `docs/AUDIT-FINDINGS.md` for downstream tracks to consume.

**Tech Stack:** Plain markdown edits. Bash for file moves and grep verification. Go toolchain to rebuild the binary and to enumerate the command tree.

**Spec:** `docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § Track 1.

---

## Task Inventory

| # | Task | Output |
|---|---|---|
| 1 | Rebuild binary, enumerate command tree | `/tmp/polygolem-help.txt`, `/tmp/polygolem-cmds.txt` |
| 2 | Set up archive — create `docs/history/`, move 2 files, delete 1 | Archive in place |
| 3 | Create `docs/AUDIT-FINDINGS.md` working-notes scaffold | New file |
| 4 | Rewrite `docs/ARCHITECTURE.md` | Refreshed file |
| 5 | Regenerate `docs/COMMANDS.md` | Refreshed file |
| 6 | Audit & reconcile `docs/PRD.md` in place | Annotated file |
| 7 | Audit & extend `docs/SAFETY.md` | Refreshed file |
| 8 | Spot-fix `README.md` drift | Touched-up file |
| 9 | Final Track 1 verification gate | Green |

Each task is one commit unless noted.

---

## Task 1: Rebuild binary and snapshot the command tree

**Files:**
- Build: `cmd/polygolem/` (no edits, just build)
- Capture: `/tmp/polygolem-help.txt`, `/tmp/polygolem-cmds.txt` (working snapshots, not committed)

**Why first:** The audit needs ground truth. The shipped `./polygolem` predates `internal/cli/deposit_wallet.go` (the README headline) and likely other commands.

- [ ] **Step 1: Rebuild the binary**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/polygolem
go build -o polygolem ./cmd/polygolem
```

Expected: builds cleanly, no errors. If `go build` fails, **stop and report**. Track 1 cannot proceed against a non-building tree.

- [ ] **Step 2: Snapshot the top-level help**

```bash
./polygolem --help 2>&1 | tee /tmp/polygolem-help.txt
```

Expected: command list including (at minimum, per current source): `auth`, `bridge`, `clob`, `discover`, `events`, `health`, `live`, `orderbook`, `paper`, `preflight`, `version`. **Confirm `deposit-wallet` is now present.** If still absent, log this as a code-side bug in AUDIT-FINDINGS.md (Task 3) and continue — the audit reflects what's true today, not what we wish were true.

- [ ] **Step 3: Walk every command and snapshot full help**

```bash
> /tmp/polygolem-cmds.txt
for cmd in $(./polygolem --help 2>&1 | awk '/Available Commands:/,/^$/' | tail -n +2 | awk '{print $1}' | grep -v '^$\|completion\|help'); do
  echo "===== $cmd =====" >> /tmp/polygolem-cmds.txt
  ./polygolem $cmd --help >> /tmp/polygolem-cmds.txt 2>&1
  for sub in $(./polygolem $cmd --help 2>&1 | awk '/Available Commands:/,/^$/' | tail -n +2 | awk '{print $1}' | grep -v '^$'); do
    echo "===== $cmd $sub =====" >> /tmp/polygolem-cmds.txt
    ./polygolem $cmd $sub --help >> /tmp/polygolem-cmds.txt 2>&1
  done
done
wc -l /tmp/polygolem-cmds.txt
```

Expected: file is non-empty (should be hundreds of lines). This is the ground-truth command catalog Tasks 5 (COMMANDS.md) and Track 3 (SKILL.md) will consume.

- [ ] **Step 4: Verify ground truth files exist**

```bash
test -s /tmp/polygolem-help.txt && test -s /tmp/polygolem-cmds.txt && echo "OK"
```

Expected: `OK`.

- [ ] **Step 5: No commit**

This task produces working files in `/tmp` only. Nothing to stage. Move to Task 2.

---

## Task 2: Set up `docs/history/` archive

**Files:**
- Create: `docs/history/` (directory)
- Create: `docs/history/README.md`
- Move: `docs/REFERENCE-RUST-CLI.md` → `docs/history/REFERENCE-RUST-CLI.md`
- Move: `docs/PHASE0-GOBOT-MIGRATION.md` → `docs/history/PHASE0-GOBOT-MIGRATION.md`
- Delete: `docs/IMPLEMENTATION-PLAN.md`

- [ ] **Step 1: Verify the three target files currently exist**

```bash
test -f docs/REFERENCE-RUST-CLI.md && \
test -f docs/PHASE0-GOBOT-MIGRATION.md && \
test -f docs/IMPLEMENTATION-PLAN.md && echo "All present"
```

Expected: `All present`. If any file is missing, the spec is stale; stop and report.

- [ ] **Step 2: Create `docs/history/` directory and move two files**

```bash
mkdir -p docs/history
git mv docs/REFERENCE-RUST-CLI.md docs/history/REFERENCE-RUST-CLI.md
git mv docs/PHASE0-GOBOT-MIGRATION.md docs/history/PHASE0-GOBOT-MIGRATION.md
```

Expected: no errors. `git status` shows two `R ` (renamed) entries.

- [ ] **Step 3: Delete the obsolete plan**

```bash
git rm docs/IMPLEMENTATION-PLAN.md
```

Expected: `git status` shows `D ` (deleted) for the file.

- [ ] **Step 4: Write `docs/history/README.md`**

Exact content:

```markdown
# History

Archived documentation kept for archaeology. These files are frozen in time
and do not reflect the current codebase. For current docs, see the parent
`docs/` directory or the docs site.

| File | Captured | What it documents |
|---|---|---|
| `REFERENCE-RUST-CLI.md` | 2026-05-06 | Behavioral audit of the upstream Polymarket Rust CLI at commit `4b5a749`. Used as a reference target for polygolem's command shape. |
| `PHASE0-GOBOT-MIGRATION.md` | 2026-05-06 | The Phase 0 plan for moving direct Polymarket protocol access out of `go-bot` and into polygolem. The migration shipped; this file remains for anyone debugging the go-bot ↔ polygolem boundary. |

Older planning docs that no longer reflect reality have been deleted. See
`git log` for the historical trail.
```

- [ ] **Step 5: Verify the moves and delete are coherent**

```bash
ls docs/history/
test -f docs/history/REFERENCE-RUST-CLI.md && \
test -f docs/history/PHASE0-GOBOT-MIGRATION.md && \
test -f docs/history/README.md && \
test ! -f docs/REFERENCE-RUST-CLI.md && \
test ! -f docs/PHASE0-GOBOT-MIGRATION.md && \
test ! -f docs/IMPLEMENTATION-PLAN.md && echo "Archive OK"
```

Expected: `Archive OK`.

- [ ] **Step 6: Commit**

```bash
git add docs/history/ docs/REFERENCE-RUST-CLI.md docs/PHASE0-GOBOT-MIGRATION.md docs/IMPLEMENTATION-PLAN.md
git commit -m "$(cat <<'EOF'
docs: archive stale phase docs to docs/history/, delete IMPLEMENTATION-PLAN

Move REFERENCE-RUST-CLI.md and PHASE0-GOBOT-MIGRATION.md to docs/history/
with an index README. Delete IMPLEMENTATION-PLAN.md whose gap-analysis
content is fully stale (everything it lists as missing is now built; git
history is the archive).

Part of Track 1 (Audit & Truth Pass) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Create `docs/AUDIT-FINDINGS.md` working-notes scaffold

**Files:**
- Create: `docs/AUDIT-FINDINGS.md`

**Why now:** Tasks 4–8 will discover drift items as they audit. The scaffold gives them a single place to log everything. Track 3 will populate the `JSON drift` section. Track 5 (final task) will delete this file.

- [ ] **Step 1: Write `docs/AUDIT-FINDINGS.md` with this exact content**

```markdown
# Audit Findings — Documentation Overhaul

**Status:** working notes for the documentation overhaul project. Populated
during Track 1, consumed by Tracks 2–4, deleted at the end of Track 5.

This file is **not user-facing**. It captures drift items that surfaced
while auditing existing documentation. Items marked `→ code` describe a
discrepancy between docs and source where the **source** was wrong (i.e.,
the doc is correct, the code drifted). Items marked `→ docs` describe the
opposite. Items marked `→ defer` are out of scope for this project.

## Track 1 — Doc-side drift

(Populated by Tasks 4–8. One row per finding.)

| Finding | Doc | Code reference | Direction | Resolution |
|---|---|---|---|---|

## Track 1 — Code-side drift surfaced incidentally

(Items where audit revealed a code bug. Not fixed in Track 1; logged for
later.)

| Finding | File:line | Notes |
|---|---|---|

## Track 3 — JSON envelope drift

(Populated by Track 3. Per-command list of where current `--json` output
diverges from the v1 envelope spec'd in
`docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § 3a.)

| Command | Current shape | Drift from v1 envelope |
|---|---|---|

## Findings consumed downstream

(Filled in by each track as it consumes findings, so we know what's been
addressed.)

| Finding | Consumed by |
|---|---|
```

- [ ] **Step 2: Verify the file exists and is non-empty**

```bash
test -s docs/AUDIT-FINDINGS.md && echo "OK"
```

Expected: `OK`.

- [ ] **Step 3: Commit**

```bash
git add docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs: add AUDIT-FINDINGS.md working-notes scaffold

Working file for the documentation overhaul project. Captures drift items
discovered during Track 1 audits and JSON envelope drift from Track 3.
Deleted at the end of Track 5 once findings have been consumed.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Rewrite `docs/ARCHITECTURE.md`

**Files:**
- Modify: `docs/ARCHITECTURE.md` (full rewrite)
- Modify: `docs/AUDIT-FINDINGS.md` (log the lies that were fixed)

- [ ] **Step 1: Confirm the lies in the current file**

```bash
git grep -nE "no public Go SDK|intentionally avoids \`pkg/\`" docs/ARCHITECTURE.md
```

Expected: at least 2 hits inside `docs/ARCHITECTURE.md`. These are the strings the rewrite must eliminate.

- [ ] **Step 2: List the actual package surface**

```bash
ls internal/ | sort
ls pkg/ | sort
```

Expected output:
```
auth, cli, clob, config, dataapi, errors, execution, gamma, marketdiscovery,
modes, orders, output, paper, polytypes, preflight, relayer, risk, rpc,
stream, transport, wallet
```
(21 internal packages)
```
bookreader, bridge, gamma, marketresolver, pagination
```
(5 public packages)

If the listing differs from this, **stop and report** — the spec needs an update.

- [ ] **Step 3: Replace the entire contents of `docs/ARCHITECTURE.md` with the following**

```markdown
# Architecture

`polygolem` is a Go protocol and automation stack for Polymarket with a
Cobra-based CLI frontend. The CLI is a thin shell over typed, testable
internal packages and a small public SDK in `pkg/`.

## Surface map

### Public SDK (`pkg/`)

Stable interfaces for downstream Go consumers (e.g., `go-bot`).

| Package | Purpose |
|---|---|
| `pkg/bookreader` | Read-only CLOB order-book reader. |
| `pkg/bridge` | Bridge API client — supported assets, deposit addresses, quotes. |
| `pkg/gamma` | Read-only Gamma API surface for embedded use. |
| `pkg/marketresolver` | Resolve market identifiers (ID, slug, token-id) to a canonical view. |
| `pkg/pagination` | Cursor and offset pagination with concurrent batching. |

### Internal packages (`internal/`)

Implementation. Not part of the public SDK contract.

| Package | Purpose |
|---|---|
| `internal/auth` | L0/L1/L2 auth, EIP-712, deposit-wallet CREATE2 derivation, builder attribution, signers. |
| `internal/cli` | Cobra command construction and dependency wiring. |
| `internal/clob` | CLOB API client — full read + authenticated surface, EIP-712, POLY_1271, ERC-7739. |
| `internal/config` | Viper-backed config loading, defaults, environment binding, validation, redaction. |
| `internal/dataapi` | Data API client — positions, volume, leaderboards. |
| `internal/errors` | Structured error types and code helpers. |
| `internal/execution` | Paper executor today; live executor surface for future use. |
| `internal/gamma` | Typed Gamma HTTP client — markets, events, search, tags, series, sports, comments, profiles. |
| `internal/marketdiscovery` | High-level market discovery service that combines Gamma and CLOB. |
| `internal/modes` | Read-only / paper / live mode parsing and gate checks. |
| `internal/orders` | OrderIntent, fluent builder, validation, lifecycle states. |
| `internal/output` | Stable table and JSON rendering plus structured errors. |
| `internal/paper` | Local-only paper positions, fills, and persisted state. |
| `internal/polytypes` | Polymarket protocol-level types shared across clients. |
| `internal/preflight` | Local and remote readiness checks. |
| `internal/relayer` | Builder relayer client — WALLET-CREATE, WALLET batch, nonce, polling. |
| `internal/risk` | Per-trade caps, daily loss limits, circuit breaker. |
| `internal/rpc` | Direct on-chain transfers (e.g., ERC-20 pUSD from EOA). |
| `internal/stream` | WebSocket market client with reconnect and dedup. |
| `internal/transport` | HTTP retry, rate limiter, circuit breaker, redaction. |
| `internal/wallet` | Deposit-wallet primitives — derive, deploy, status, batch signing. |

## Dependency direction

```text
cmd/polygolem
        |
internal/cli
        |
internal/{config, modes, preflight, output, errors}
        |
internal/{gamma, clob, dataapi, stream, relayer, rpc}   ← protocol clients
        |
internal/{auth, transport, polytypes}                   ← cross-cutting primitives
        |
internal/{wallet, orders, execution, risk, paper, marketdiscovery}
        |
pkg/{bookreader, bridge, gamma, marketresolver, pagination}   ← public re-exposed surface
```

Command handlers parse flags, call package APIs, and render output via
`internal/output`. Protocol clients do not know about Cobra. Safety packages
do not depend on command text. Paper state stays local and never reaches
authenticated mutation endpoints.

Cobra command handlers must not contain protocol or trading business logic.
That logic belongs in typed clients, application services, safety gates, and
paper-state packages where it is testable without executing the binary.

## Mode system

Mode selection starts in configuration and CLI flags, then flows through
`internal/modes` before command handlers call protocol clients or paper
state.

- **Read-only** (default): public market data only. May use
  `internal/gamma`, `internal/clob` (read endpoints), `internal/dataapi`,
  `internal/marketdiscovery`, and `internal/output`. Forbids signing or
  any mutation.
- **Paper**: local simulation. Combines read-only reference data with
  `internal/paper` state. Simulated actions stay local. Authenticated
  mutation APIs remain off-limits.
- **Live**: gated. Requires preflight + risk + funding gates to pass.
  Live execution operates through `internal/execution`, `internal/orders`,
  `internal/clob` (write endpoints), `internal/relayer`, `internal/rpc`,
  and `internal/wallet`. The default `polygolem` invocation does not enter
  live mode.

## Signature types

Live commands accept a `--signature-type` flag. Supported values:

| Value | Description |
|---|---|
| `eoa` | Plain externally-owned account; rejected by Polymarket for new accounts after May 2026. Retained for legacy keys. |
| `proxy` | Proxy-wallet signing. |
| `gnosis-safe` | Gnosis Safe signing. |
| `deposit` | Deposit wallet (POLY_1271). The supported path for new accounts after May 2026. See `docs/DEPOSIT-WALLET-MIGRATION.md`. |

Builder attribution for orders is handled in `internal/auth` and is
orthogonal to the signature type.

## Public SDK boundary

`pkg/` exists. It is small by design and grows when an internal capability
proves stable enough to expose. Do not move code into `pkg/` without an
SDK-level commitment to keep its API stable across minor versions.

## Safety boundaries

- Read-only is the default mode and is exercised by every public command.
- Paper mode never calls authenticated endpoints.
- Live commands require explicit signature-type, gates passing, and
  builder credentials where applicable.
- Builder credentials and private keys are redacted by `internal/config`
  on every load.
```

- [ ] **Step 4: Verify the lies are gone**

```bash
git grep -nE "no public Go SDK|intentionally avoids \`pkg/\`" docs/
```

Expected: **no output** (exit code 1). If hits remain, the rewrite missed a spot.

- [ ] **Step 5: Verify every package is referenced**

```bash
for p in $(ls internal/); do
  grep -q "internal/$p" docs/ARCHITECTURE.md || echo "MISSING: internal/$p"
done
for p in $(ls pkg/); do
  grep -q "pkg/$p" docs/ARCHITECTURE.md || echo "MISSING: pkg/$p"
done
```

Expected: no `MISSING:` lines.

- [ ] **Step 6: Log the finding in `docs/AUDIT-FINDINGS.md`**

Append a row under the `Track 1 — Doc-side drift` table:

```markdown
| ARCHITECTURE.md claimed no public SDK and "intentionally avoids `pkg/`" | docs/ARCHITECTURE.md | `pkg/{bookreader,bridge,gamma,marketresolver,pagination}` exist | → docs | Rewritten in Task 4 |
| ARCHITECTURE.md missed 13+ internal packages | docs/ARCHITECTURE.md | `internal/{auth,clob,dataapi,errors,execution,marketdiscovery,orders,polytypes,relayer,risk,rpc,stream,transport,wallet}` | → docs | Rewritten in Task 4 |
```

- [ ] **Step 7: Commit**

```bash
git add docs/ARCHITECTURE.md docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs: rewrite ARCHITECTURE.md to reflect 21 internal + 5 pkg/ packages

Drops the stale "no public Go SDK in Phase 1" / "intentionally avoids pkg/"
claims. Documents the actual surface map, dependency direction, mode
system, and signature types (including the May 2026 deposit signature).

Audit findings logged in docs/AUDIT-FINDINGS.md.

Part of Track 1 (Audit & Truth Pass) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Regenerate `docs/COMMANDS.md`

**Files:**
- Modify: `docs/COMMANDS.md` (full regenerate)
- Modify: `docs/AUDIT-FINDINGS.md` (log any drift)

- [ ] **Step 1: Read the current file in full**

```bash
wc -l docs/COMMANDS.md
cat docs/COMMANDS.md
```

Note: which command groups are documented vs missing relative to `/tmp/polygolem-help.txt` from Task 1.

- [ ] **Step 2: Build the canonical command list from `/tmp/polygolem-cmds.txt`**

```bash
grep '^=====' /tmp/polygolem-cmds.txt | sed -E 's/^===== (.+) =====$/\1/' | sort -u
```

Expected: a list of every command and subcommand actually present in the binary, one per line, with full paths preserved (e.g., `bridge`, `bridge assets`, `clob create-order`). This is the authoritative command catalog for the rest of this task.

- [ ] **Step 3: Replace `docs/COMMANDS.md` with a structured catalog**

Use this template. For every command and subcommand listed in Step 2, fill in a `### <command-path>` section. Pull `Usage:`, `Flags:`, and one runnable example per command from `/tmp/polygolem-cmds.txt`. **Do not invent flags or examples not visible in `--help` output.**

```markdown
# Commands

Complete command reference for `polygolem`. Generated by walking
`polygolem --help` recursively. Re-run Task 5 of the documentation overhaul
plan to regenerate when commands are added or changed.

Source of truth for flag semantics: `polygolem <cmd> --help`.

## Conventions

- All commands accept `--json` to emit structured JSON instead of tables.
- Read-only commands do not require credentials.
- Authenticated commands consume environment variables; see
  [Environment Variables](#environment-variables).
- Live-mutating commands require `--signature-type` and pass a gate check
  before submitting.

## Commands

For every command path produced by Step 2 (top-level and subcommand alike),
emit one `### <full command path>` section using the shape below. Pull
`Usage:` and the `Flags:` block verbatim from the matching block in
`/tmp/polygolem-cmds.txt`. Every flag listed by `--help` must appear in the
table.

### Worked example — `### bridge assets`

````markdown
### bridge assets

List supported assets that can be bridged into Polymarket.

**Usage:**

```
polygolem bridge assets [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--json` | bool | `false` | Emit JSON output instead of a table. |
| `-h, --help` | bool | `false` | Help for `assets`. |

**Example:**

```bash
polygolem bridge assets --json
```
````

Repeat that exact shape for every command in the catalog. Order: top-level
commands first (alphabetically), then subcommand paths (alphabetically).
Do not collapse subcommands under their parent — each command path gets
its own H3.

## Environment Variables

| Variable | Required for |
|---|---|
| `POLYMARKET_PRIVATE_KEY` | All authenticated commands. |
| `POLYMARKET_BUILDER_API_KEY` | Deposit-wallet deploy/batch/onboard. |
| `POLYMARKET_BUILDER_SECRET` | Deposit-wallet deploy/batch/onboard. |
| `POLYMARKET_BUILDER_PASSPHRASE` | Deposit-wallet deploy/batch/onboard. |
| `POLYMARKET_RELAYER_URL` | Override relayer URL (default: relayer-v2.polymarket.com). |

Short-form `BUILDER_API_KEY` / `BUILDER_SECRET` / `BUILDER_PASS_PHRASE` are
also accepted.
```

If a command listed in Step 2 is missing from the rewritten file, **the task is not done.** This is the single most common drift trap.

- [ ] **Step 4: Verify every command in the binary is documented**

```bash
while IFS= read -r cmd; do
  grep -qF "### $cmd" docs/COMMANDS.md || echo "MISSING: $cmd"
done < <(grep '^=====' /tmp/polygolem-cmds.txt | sed -E 's/^===== (.+) =====$/\1/' | sort -u)
```

Expected: no `MISSING:` lines. (`grep -F` so multi-word paths like
`clob create-order` match literally; `<(…)` keeps spaces intact.)

- [ ] **Step 5: Log findings in `docs/AUDIT-FINDINGS.md`**

For each command that was in the binary but missing from the prior `docs/COMMANDS.md`, append a row under `Track 1 — Doc-side drift`. For each command that was in the prior `COMMANDS.md` but **not** in the binary, append a row under `Track 1 — Code-side drift surfaced incidentally`. Be terse — one row per item.

- [ ] **Step 6: Commit**

```bash
git add docs/COMMANDS.md docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs: regenerate COMMANDS.md from polygolem --help

Walks every command and subcommand currently in the binary; produces a
structured catalog with one section per command. Cross-references against
/tmp/polygolem-cmds.txt; missing commands logged as drift in
AUDIT-FINDINGS.md.

Part of Track 1 (Audit & Truth Pass) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Audit & reconcile `docs/PRD.md` in place

**Files:**
- Modify: `docs/PRD.md` (annotate, do not rewrite)
- Modify: `docs/AUDIT-FINDINGS.md` (log items)

**Approach:** PRD has historical "why" content worth preserving. Walk it section by section. For each requirement, mark its current state with one of:

- ✅ **Fulfilled** — code matches the requirement.
- ⚠️ **Partial / drifted** — code addresses some but not all of the requirement, or interpretation has shifted.
- 🗒️ **Historical** — the requirement was deliberately changed or dropped; preserve the text as historical context but mark it.

Do **not** delete sections. Do **not** rewrite prose. Add status markers and an inline note where useful.

- [ ] **Step 1: Read the entire PRD**

```bash
wc -l docs/PRD.md
```

Open the file. Read every section. Maintain a running tally on paper / in a temporary file of which requirements are fulfilled vs partial vs historical.

- [ ] **Step 2: Add a status legend at the top of the PRD**

Insert this block immediately after the top-level title:

```markdown
> **Audit status (2026-05-07):** This PRD predates the current codebase.
> Each requirement is annotated with one of:
>
> - ✅ **Fulfilled** — implemented and shipping.
> - ⚠️ **Partial / drifted** — implementation differs in scope or shape.
> - 🗒️ **Historical** — preserved for context; superseded by a later
>   decision documented in `docs/ARCHITECTURE.md` or
>   `docs/DEPOSIT-WALLET-MIGRATION.md`.
>
> The status reflects code reality at the audit date. The "why" prose is
> preserved unchanged.
```

- [ ] **Step 3: Walk every requirement and add an inline status marker**

For each requirement heading (e.g., `### R1 Market Discovery`), append the
status emoji to the heading line and add one inline note immediately
below explaining current state:

```markdown
### R1 Market Discovery ✅

> **Status:** Fulfilled. Implemented in `internal/gamma`,
> `internal/marketdiscovery`, and exposed via `polygolem discover *`.
> See `docs/ARCHITECTURE.md` for the package map.
```

Use compact notes — two sentences max per requirement. Reference packages
or commands. Do not edit the original prose body of the requirement.

- [ ] **Step 4: Verify every requirement heading has a status marker**

```bash
grep -nE '^###\s+R\d+' docs/PRD.md | grep -vE '✅|⚠️|🗒️' && echo "ABOVE LINES MISSING STATUS"
```

Expected: no `ABOVE LINES MISSING STATUS` output (the grep should print nothing). If headings appear without an emoji, fix them.

- [ ] **Step 5: Log findings in `docs/AUDIT-FINDINGS.md`**

For every ⚠️ or 🗒️ requirement, append a row under
`Track 1 — Doc-side drift` describing what changed and which doc now
documents the current behavior. For every code bug or gap surfaced
incidentally, append to `Track 1 — Code-side drift surfaced incidentally`.

- [ ] **Step 6: Commit**

```bash
git add docs/PRD.md docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs: annotate PRD with current status per requirement

Walks every R-numbered requirement. Marks each as Fulfilled / Partial /
Historical without rewriting historical "why" prose. Drift items logged in
AUDIT-FINDINGS.md with pointers to current authoritative docs.

Part of Track 1 (Audit & Truth Pass) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: Audit & extend `docs/SAFETY.md`

**Files:**
- Modify: `docs/SAFETY.md`
- Modify: `docs/AUDIT-FINDINGS.md` if drift found

- [ ] **Step 1: Read current SAFETY.md and confirm read-only-by-default still holds in code**

```bash
cat docs/SAFETY.md
grep -rn "ModeReadOnly\|DefaultMode\|defaultMode" internal/modes/ internal/cli/
```

Verify: `internal/modes` (or wherever the default is set) still defaults to read-only. If not, log as `→ code` drift in AUDIT-FINDINGS.md and continue.

- [ ] **Step 2: Confirm coverage of the existing surface**

Verify SAFETY.md covers, at minimum:
- Read-only as default mode.
- Paper-mode local-only constraint.
- Live-mode gate requirement.
- Credential redaction guarantees.

If any are missing, add them by appending paragraphs in the appropriate existing section.

- [ ] **Step 3: Add a new section "Deposit Wallet Safety Rules"**

Append the following section at the end of SAFETY.md (before any final footer if present):

```markdown
## Deposit Wallet Safety Rules

The May 2026 deposit-wallet migration introduces a new family of commands
(`polygolem deposit-wallet *`) that perform on-chain or relayer-bound
operations. These rules apply.

1. **Builder credentials are required and redacted.** `--deploy`,
   `--batch`, `--approve --submit`, and `--onboard` require
   `POLYMARKET_BUILDER_API_KEY`, `POLYMARKET_BUILDER_SECRET`, and
   `POLYMARKET_BUILDER_PASSPHRASE`. Configuration loading redacts all three
   on every load; no command emits them in JSON output.

2. **Read-only deposit-wallet commands stay read-only.**
   `deposit-wallet derive`, `deposit-wallet status`, and
   `deposit-wallet nonce` perform no on-chain or relayer mutations.

3. **Batch signing requires explicit calldata input.**
   `deposit-wallet batch --calls-json` requires structured input. The CLI
   does not synthesize calls. The `approve` shortcut shows calldata before
   submission unless `--submit` is passed.

4. **Funding moves real money.** `deposit-wallet fund --amount X`
   transfers ERC-20 pUSD from the EOA to the deposit wallet via direct RPC.
   The amount must be specified explicitly. There is no default.

5. **Onboarding is the only multi-step composite.**
   `deposit-wallet onboard --fund-amount X` performs deploy → approve →
   fund. Each step is gated; failure of any step aborts the composite and
   leaves the wallet in a recoverable state visible to
   `deposit-wallet status`.

6. **POLY_1271 orders use the deployed wallet's signature path.**
   `clob create-order --signature-type deposit` and
   `clob market-order --signature-type deposit` sign with the deposit
   wallet's POLY_1271 path. Orders signed without the deposit signature
   type after the May 2026 cutoff will be rejected by Polymarket for
   new accounts.

7. **Builder attribution does not bypass safety.** Setting builder
   credentials enables deposit-wallet operations; it does not relax any
   gate or grant trading privileges.
```

- [ ] **Step 4: Verify the new section renders and existing content is intact**

```bash
grep -n "^## " docs/SAFETY.md
git diff docs/SAFETY.md
```

Expected: the diff is **purely additive** in the deposit-wallet section, plus any clarifying paragraphs added in Step 2. No deletions of pre-existing safety claims.

- [ ] **Step 5: Commit**

```bash
git add docs/SAFETY.md docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs: extend SAFETY.md with deposit-wallet safety rules

Adds a Deposit Wallet Safety Rules section covering builder credential
redaction, read-only deposit-wallet commands, explicit calldata input
for batch signing, real-money funding, the onboarding composite, POLY_1271
order signing, and the boundary between builder attribution and safety
gates.

Part of Track 1 (Audit & Truth Pass) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: Spot-fix `README.md` drift

**Files:**
- Modify: `README.md` (no rewrite — only fix contradictions)
- Modify: `docs/AUDIT-FINDINGS.md`

- [ ] **Step 1: Cross-reference README.md against the audit so far**

For each of the following claims in README.md, verify against current code:

- The deposit-wallet pitch refers to commands that exist in the binary
  (`./polygolem deposit-wallet onboard --help` succeeds).
- Every package listed in the "Packages" table exists at the path given.
- The "Status" table phases match the spec / current code state.
- Env-var list matches what `internal/config` actually reads.

```bash
./polygolem deposit-wallet --help >/dev/null 2>&1 && echo "deposit-wallet OK" || echo "deposit-wallet MISSING"
for p in $(grep -oE 'internal/\w+|pkg/\w+' README.md | sort -u); do
  test -d "$p" || echo "MISSING: $p"
done
```

Expected: `deposit-wallet OK` and no `MISSING:` lines.

- [ ] **Step 2: Fix any contradiction inline**

Make the smallest edit that removes the contradiction. Do not rewrite
sections. Do not add new sections. If a fact is genuinely ambiguous, log
it in AUDIT-FINDINGS.md instead of guessing.

- [ ] **Step 3: Verify the README still reads coherently**

```bash
head -50 README.md
tail -30 README.md
```

Expected: no orphan headings, no half-edited paragraphs.

- [ ] **Step 4: Commit (only if changes were needed)**

```bash
git add README.md docs/AUDIT-FINDINGS.md
git diff --cached --quiet && echo "No changes — skip commit" || git commit -m "$(cat <<'EOF'
docs: spot-fix README.md drift surfaced by Track 1 audit

Smallest-edit fixes for contradictions between README.md and the actual
binary / package layout. No rewrite. Items logged in AUDIT-FINDINGS.md.

Part of Track 1 (Audit & Truth Pass) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: Final Track 1 verification gate

**Files:** none modified — read-only verification.

This task does not produce a commit. It either passes (Track 1 done) or
identifies regressions to loop back on.

- [ ] **Step 1: No lies remain**

```bash
git grep -nE "no public Go SDK|intentionally avoids \`pkg/\`|Phase 0\b.*pending|Phase 1\b.*disabled" docs/ README.md
```

Expected: **no output**, OR all hits are inside `docs/history/` (archived files are allowed to lie).

```bash
git grep -nE "no public Go SDK|intentionally avoids \`pkg/\`" docs/ README.md | grep -v '^docs/history/'
```

Expected: **no output**.

- [ ] **Step 2: Every command in the binary is in COMMANDS.md**

```bash
while IFS= read -r cmd; do
  grep -qF "### $cmd" docs/COMMANDS.md || echo "MISSING: $cmd"
done < <(grep '^=====' /tmp/polygolem-cmds.txt | sed -E 's/^===== (.+) =====$/\1/' | sort -u)
```

Expected: no `MISSING:` lines.

- [ ] **Step 3: Every package on disk is in ARCHITECTURE.md**

```bash
for p in $(ls internal/); do
  grep -q "internal/$p" docs/ARCHITECTURE.md || echo "MISSING: internal/$p"
done
for p in $(ls pkg/); do
  grep -q "pkg/$p" docs/ARCHITECTURE.md || echo "MISSING: pkg/$p"
done
```

Expected: no `MISSING:` lines.

- [ ] **Step 4: Archive is in place**

```bash
test -f docs/history/REFERENCE-RUST-CLI.md && \
test -f docs/history/PHASE0-GOBOT-MIGRATION.md && \
test -f docs/history/README.md && \
test ! -f docs/REFERENCE-RUST-CLI.md && \
test ! -f docs/PHASE0-GOBOT-MIGRATION.md && \
test ! -f docs/IMPLEMENTATION-PLAN.md && \
echo "Archive OK"
```

Expected: `Archive OK`.

- [ ] **Step 5: AUDIT-FINDINGS.md is populated**

```bash
test -s docs/AUDIT-FINDINGS.md && \
grep -q "Track 1 — Doc-side drift" docs/AUDIT-FINDINGS.md && \
echo "Findings file OK"
```

Expected: `Findings file OK`. The file should have at least 3 rows in the doc-side drift table (the absolute minimum: ARCHITECTURE.md lies × 2, plus at least one COMMANDS finding).

- [ ] **Step 6: Build and tests are still green**

```bash
go build ./cmd/polygolem
go vet ./...
go test ./...
```

Expected: all three pass. Track 1 is documentation-only and must not have regressed code state.

- [ ] **Step 7: If all checks pass, mark Track 1 complete**

No file change required. Inform the user that Track 1 verification has
passed and propose moving to Track 2 planning.

If any check fails, return to the relevant earlier task and fix in place
rather than papering over.

---

## Out of scope (re-stated)

- Code changes to fix anything found in `Track 1 — Code-side drift surfaced
  incidentally`. Those are logged for future work.
- Track 2/3/4/5 work — separate plans.
- Changes to `docs/superpowers/specs/` content; the spec is locked for
  this implementation cycle.
