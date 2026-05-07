# Track 5 — Repo Polish: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the polygolem repository up to a presentable open-source baseline — contribution norms, vulnerability reporting, a changelog, GitHub issue/PR templates, dependency automation, and refreshed README badges. Close the documentation overhaul project by deleting `docs/AUDIT-FINDINGS.md` once Tracks 1–4 have consumed every finding, and run the full Track 5 verification gate.

**Architecture:** Pure repo-meta work. Add seven small new files (`CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md`, two `.github/ISSUE_TEMPLATE/*.md`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/dependabot.yml`). Refresh `README.md` (badges + trim Status table). Optionally add a `docs-site-build` job to `.github/workflows/ci.yml`. Delete `docs/AUDIT-FINDINGS.md`. Each of these is one commit with a tight per-task allowlist.

**Tech Stack:** Plain markdown and YAML. Bash for verification. Go toolchain (`go build`, `go test`, `go vet`, `gofmt`) plus the docs-site `npm run build` for the verification gate.

**Spec:** `docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § Track 5.

**Dependencies:**

- Tracks 1–4 must be merged. Track 5's badges and CHANGELOG seed assume Track 2 (godoc), Track 3 (SKILL.md, JSON contract), and Track 4 (docs-site) have shipped.
- `docs/AUDIT-FINDINGS.md` exists from Track 1 and was populated by Tracks 2–4 as findings were consumed. Track 5's final task **deletes it** — confirm there is nothing left unconsumed before deleting.
- Project repo URL: `https://github.com/TrebuchetDynamics/polygolem`.

**Working tree note:** `main` carries uncommitted WIP unrelated to Track 5. Every task uses an explicit per-file `git add` allowlist. Never use `git add -A` or `git add .`.

---

## Task Inventory

| # | Task | Output | Commit |
|---|---|---|---|
| 1 | Add `CONTRIBUTING.md` | New file | Yes |
| 2 | Add `SECURITY.md` | New file | Yes |
| 3 | Add `CHANGELOG.md` | New file | Yes |
| 4 | Add GitHub issue + PR templates | 3 new files | Yes |
| 5 | Add `.github/dependabot.yml` | New file | Yes |
| 6 | README badges + trim Status table | Refreshed `README.md` | Yes |
| 7 | (Optional) Add `docs-site-build` CI job | Modified `.github/workflows/ci.yml` | Yes (skip if Track 4 docs-site is not green) |
| 8 | Delete `docs/AUDIT-FINDINGS.md` and run the Track 5 verification gate | Deletion commit + green gate | Yes (deletion only) |

Each task is one commit unless noted. Task 8's gate run is verification-only; the file deletion is its own tiny commit.

---

## Task 1: Add `CONTRIBUTING.md`

**Files:**
- Create: `CONTRIBUTING.md`

**Allowlist:** `CONTRIBUTING.md`

- [ ] **Step 1: Verify no `CONTRIBUTING.md` already exists**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
test ! -f CONTRIBUTING.md && echo "absent OK"
```

Expected: `absent OK`. If a file is already present, **stop and report** — Track 5 is additive; do not overwrite an existing contribution policy without an explicit follow-up.

- [ ] **Step 2: Write `CONTRIBUTING.md` with this exact content**

```markdown
# Contributing to polygolem

Thanks for considering a contribution. This file describes how to build,
test, file an issue, and what we expect from a pull request.

## Build, test, lint

`polygolem` is a single Go module. Standard toolchain only.

```bash
# Build the binary
go build -o polygolem ./cmd/polygolem

# Run all tests
go test ./...

# Static analysis
go vet ./...

# Formatting (writes in place; CI fails if anything is reformatted)
gofmt -w .
```

The CI workflow at `.github/workflows/ci.yml` runs the same four steps on
every push and pull request.

## TDD-first discipline

`polygolem` is a TDD-first project. Behavior changes land with tests, and
new tests fail before the implementation lands. The test layout follows
the standard Go convention: `*_test.go` siblings inside each package, plus
end-to-end checks under `tests/`.

If a change cannot be expressed as a failing test first (rare — usually
docs-only or repo-meta), say so explicitly in the pull request.

## Documentation surfaces

`polygolem` keeps documentation in five places. When you change behavior,
update the surfaces that lose accuracy:

| Surface | Audience | Source |
|---|---|---|
| `README.md` | Drive-by readers, install + headline pitch. | Repo root. |
| `docs/*.md` | Operators and integrators (architecture, commands, safety, PRD). | `docs/`. |
| Astro docs site | Long-form web docs, search-indexed. | `docs-site/`. |
| `SKILL.md` | Agentic consumers (Claude Code skill manifest). | Repo root. |
| Godoc comments | Go SDK consumers (`pkg/`) and contributors (`internal/`). | Inline in `.go` files. |

A change to a CLI flag typically touches `README.md`, `docs/COMMANDS.md`,
the docs-site equivalent, and `SKILL.md`. A change to a `pkg/` API touches
godoc and the docs-site reference.

## Filing an issue

Open issues at https://github.com/TrebuchetDynamics/polygolem/issues.
Use the **Bug report** template for behavior bugs and the **Feature
request** template for proposals. Include the exact command you ran and
the JSON output (with `--json`) when applicable.

For security-sensitive reports — anything touching private keys, the
deposit-wallet flow, signing paths, or builder credentials — follow
`SECURITY.md` instead. Do not file public issues for those.

## Filing a pull request

Use the pull-request template. The checklist is short:

- Tests pass locally (`go test ./...`).
- Godoc updated if the exported surface changed (`pkg/` or exported
  identifiers in `internal/`).
- `CHANGELOG.md` `## [Unreleased]` section updated with a one-line
  description of the change.

Where the planning artifacts live, in case a PR references them:

- Specs: `docs/superpowers/specs/`
- Plans: `docs/superpowers/plans/`
- Working audit notes (when present): `docs/AUDIT-FINDINGS.md` (created
  per project, deleted when consumed).

Thanks for reading.
```

- [ ] **Step 3: Verify the file exists and is non-empty**

```bash
test -s CONTRIBUTING.md && wc -l CONTRIBUTING.md
```

Expected: line count between 60 and 100.

- [ ] **Step 4: Commit**

```bash
git add CONTRIBUTING.md
git commit -m "$(cat <<'EOF'
docs: add CONTRIBUTING.md

Documents the build / test / lint loop, the TDD-first discipline, the
five doc surfaces (README, docs/, docs-site, SKILL.md, godoc), how to
file issues, and the PR expectations.

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Add `SECURITY.md`

**Files:**
- Create: `SECURITY.md`

**Allowlist:** `SECURITY.md`

- [ ] **Step 1: Verify no `SECURITY.md` already exists**

```bash
test ! -f SECURITY.md && echo "absent OK"
```

Expected: `absent OK`.

- [ ] **Step 2: Write `SECURITY.md` with this exact content**

```markdown
# Security Policy

`polygolem` handles private keys, signs Polymarket protocol messages,
deploys deposit wallets, and submits on-chain transfers. Vulnerability
reports are taken seriously.

## Reporting a vulnerability

Do **not** file a public GitHub issue.

Use one of:

1. **GitHub Security Advisories** — preferred. Open a private advisory at
   https://github.com/TrebuchetDynamics/polygolem/security/advisories/new.
2. **Email** — `security@trebuchetdynamics.com`. PGP optional; encrypt if
   the report includes secrets or transcripts.

Include: affected version (commit SHA or tag), reproduction steps, a
proof-of-concept where possible, and impact you observed.

## Expected response time

- Acknowledgement within **3 business days**.
- Initial assessment (in scope / not / needs more info) within **7 business days**.
- Coordinated disclosure timeline agreed before any public write-up.

## In scope

- Private-key handling — anywhere a key is read, held in memory, or
  passed to a signer (`internal/auth`, `internal/wallet`, `internal/rpc`).
- Deposit-wallet flow — derive, deploy, batch, approve, fund, onboard
  (`internal/wallet`, `internal/relayer`, `polygolem deposit-wallet *`).
- Signing paths — EIP-712 order signing, POLY_1271 on-chain signature
  verification, ERC-7739 typed-data flows (`internal/clob`,
  `internal/auth`).
- Builder credential handling — API key / secret / passphrase loading,
  redaction in logs and JSON output (`internal/config`,
  `internal/output`).
- JSON-output redaction — any path where `--json` could leak a secret,
  private key, or unredacted credential.
- Order-execution gates — bypasses of read-only / paper / live mode
  separation.

## Out of scope

- Polymarket protocol issues themselves (CLOB pricing, settlement,
  resolution). Report those to Polymarket directly via their channels.
- Polygon network or Ethereum tooling (e.g., go-ethereum) bugs. Report
  upstream.
- Issues that require an attacker who already has the operator's private
  key or shell access on the machine running polygolem.
- Third-party dependencies — file separately upstream; we will track via
  Dependabot updates.

Thank you for keeping `polygolem` users safe.
```

- [ ] **Step 3: Verify the file exists and is non-empty**

```bash
test -s SECURITY.md && wc -l SECURITY.md
```

Expected: line count between 30 and 60.

- [ ] **Step 4: Commit**

```bash
git add SECURITY.md
git commit -m "$(cat <<'EOF'
docs: add SECURITY.md

Documents private vulnerability reporting via GitHub Security Advisories
or email, expected response times, and the in-scope / out-of-scope
boundary (private keys, deposit-wallet, signing, builder credentials,
JSON redaction in scope; Polymarket protocol issues out of scope).

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Add `CHANGELOG.md`

**Files:**
- Create: `CHANGELOG.md`

**Allowlist:** `CHANGELOG.md`

- [ ] **Step 1: Verify no `CHANGELOG.md` already exists**

```bash
test ! -f CHANGELOG.md && echo "absent OK"
```

Expected: `absent OK`.

- [ ] **Step 2: Write `CHANGELOG.md` with this exact content**

```markdown
# Changelog

All notable changes to `polygolem` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — 2026-05-07

First tagged release. Includes everything shipped through Phase 0–E plus
the May 2026 deposit-wallet migration and the documentation overhaul.

### Added

- **Phase 0 — Go-bot boundary cleanup.** Polymarket protocol access moved
  out of `go-bot` and into a single Go module owned by `polygolem`.
- **Phase A — Read-only SDK foundation.** `internal/gamma`,
  `internal/clob` (read endpoints), `internal/dataapi`,
  `internal/marketdiscovery`, `internal/output`, `internal/transport`,
  and the public `pkg/bookreader`, `pkg/marketresolver`, `pkg/bridge`,
  `pkg/gamma`, `pkg/pagination` packages.
- **Phase B — Auth and readiness.** `internal/auth` (L0/L1/L2, EIP-712,
  builder attribution), `internal/config` (Viper-backed loading with
  redaction), `internal/preflight`, and `internal/modes`
  (read-only / paper / live).
- **Phase C — Orders and paper executor.** `internal/orders` (OrderIntent,
  fluent builder, validation, lifecycle states), `internal/execution`
  (paper executor), `internal/paper` (local-only persisted state),
  `internal/risk` (per-trade caps, daily loss limits, circuit breaker).
- **Phase D — Streams.** `internal/stream` WebSocket market client with
  reconnect and dedup.
- **Phase E — Gated live execution.** Live execution path gated by
  preflight + risk + funding checks; CLOB write endpoints accessible only
  with explicit signature type and gates passing.
- **Deposit-wallet migration (May 2026).** `internal/wallet`,
  `internal/relayer`, `internal/rpc`, and the `polygolem deposit-wallet *`
  command family — `derive`, `deploy`, `nonce`, `status`, `batch`,
  `approve`, `fund`, `onboard`. POLY_1271 order signing via
  `--signature-type deposit`.
- **CLI surface.** Cobra-based commands across `auth`, `bridge`, `clob`,
  `discover`, `events`, `health`, `live`, `orderbook`, `paper`,
  `preflight`, `version`, and `deposit-wallet` groups. Every command
  accepts `--json`.
- **Documentation overhaul (this release).**
  - Track 1 — audit & truth pass: `docs/ARCHITECTURE.md` rewritten,
    `docs/COMMANDS.md` regenerated from `--help`, `docs/PRD.md`
    annotated, `docs/SAFETY.md` extended with deposit-wallet rules,
    stale planning docs archived to `docs/history/`.
  - Track 2 — godoc on every exported identifier in `pkg/` and on
    package-level docs in `internal/`.
  - Track 3 — `SKILL.md` agent surface and the v1 JSON output contract.
  - Track 4 — Astro Starlight docs site under `docs-site/`.
  - Track 5 — `CONTRIBUTING.md`, `SECURITY.md`, this `CHANGELOG.md`,
    GitHub issue and PR templates, Dependabot config, README badges.

[Unreleased]: https://github.com/TrebuchetDynamics/polygolem/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/TrebuchetDynamics/polygolem/releases/tag/v0.1.0
```

- [ ] **Step 3: Verify the file exists and is non-empty**

```bash
test -s CHANGELOG.md && grep -q "## \[Unreleased\]" CHANGELOG.md && grep -q "## \[0.1.0\] — 2026-05-07" CHANGELOG.md && echo "OK"
```

Expected: `OK`.

- [ ] **Step 4: Commit**

```bash
git add CHANGELOG.md
git commit -m "$(cat <<'EOF'
docs: add CHANGELOG.md (Keep a Changelog format)

Seeds the file with an empty Unreleased section and a 0.1.0 entry dated
2026-05-07 summarizing Phase 0-E, the deposit-wallet migration, and the
five-track documentation overhaul. Going forward every PR updates the
Unreleased section.

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Add GitHub issue and pull-request templates

**Files:**
- Create: `.github/ISSUE_TEMPLATE/bug_report.md`
- Create: `.github/ISSUE_TEMPLATE/feature_request.md`
- Create: `.github/PULL_REQUEST_TEMPLATE.md`

**Allowlist:** `.github/ISSUE_TEMPLATE/`, `.github/PULL_REQUEST_TEMPLATE.md`

- [ ] **Step 1: Verify the templates do not already exist**

```bash
test ! -f .github/ISSUE_TEMPLATE/bug_report.md && \
test ! -f .github/ISSUE_TEMPLATE/feature_request.md && \
test ! -f .github/PULL_REQUEST_TEMPLATE.md && echo "absent OK"
```

Expected: `absent OK`.

- [ ] **Step 2: Create `.github/ISSUE_TEMPLATE/` directory**

```bash
mkdir -p .github/ISSUE_TEMPLATE
```

Expected: directory exists. `.github/` already contains `workflows/`, so the parent directory is in place.

- [ ] **Step 3: Write `.github/ISSUE_TEMPLATE/bug_report.md` with this exact content**

```markdown
---
name: Bug report
about: Report a defect in polygolem behavior
title: "[bug] "
labels: bug
assignees: ''
---

## Reproduction

Exact command(s) you ran. Use `--json` where applicable.

```bash
polygolem ...
```

## Expected behavior

What you expected to happen.

## Actual behavior

What actually happened. Paste the JSON output (with `--json`) when
relevant. Redact any private keys, builder credentials, or wallet
addresses you do not want public.

## Environment

- `polygolem version` output:
- OS and arch (e.g., `uname -a`):
- Go version (`go version`):
- Mode (read-only / paper / live):

## Additional context

Anything else useful — relevant config, network conditions, links to a
specific market or token id, etc.

## Security note

If this issue involves private-key handling, deposit-wallet flow,
signing paths, builder credentials, or JSON-output redaction, please
**stop** and follow `SECURITY.md` instead of filing a public issue.
```

- [ ] **Step 4: Write `.github/ISSUE_TEMPLATE/feature_request.md` with this exact content**

```markdown
---
name: Feature request
about: Propose a new capability or enhancement
title: "[feature] "
labels: enhancement
assignees: ''
---

## Problem

What problem does this solve? Who is affected? Reference a concrete
workflow if you can.

## Proposed solution

Describe what you would like to see. CLI shape, package API, JSON
field — be concrete enough that a reviewer can push back.

## Alternatives considered

Other shapes you thought about and why this one is better.

## Scope

- Mode: read-only / paper / live (or N/A).
- Surfaces touched: README / docs / docs-site / SKILL.md / godoc / CLI
  / package API.
- Backwards-compat impact: none / additive / breaking.

## Additional context

Links, references, related issues, or prior art.
```

- [ ] **Step 5: Write `.github/PULL_REQUEST_TEMPLATE.md` with this exact content**

```markdown
## Summary

One or two sentences describing what this PR does and why.

## Changes

- Bullet list of the substantive changes.

## Checklist

- [ ] Tests pass locally (`go test ./...`).
- [ ] `go vet ./...` is clean.
- [ ] `gofmt -l .` produces no output.
- [ ] Godoc updated if the exported surface changed (`pkg/` or exported
      identifiers in `internal/`).
- [ ] `CHANGELOG.md` `## [Unreleased]` section updated with a one-line
      entry.
- [ ] Doc surfaces (`README.md`, `docs/*.md`, `docs-site/`, `SKILL.md`)
      updated where this change affects them.
- [ ] No secrets, private keys, builder credentials, or `.env` files
      committed.

## Test plan

How a reviewer can verify this change locally. Commands, expected
output, sample data.

## Related

Issue numbers, prior PRs, or specs/plans under `docs/superpowers/`.
```

- [ ] **Step 6: Verify all three files exist and are non-empty**

```bash
test -s .github/ISSUE_TEMPLATE/bug_report.md && \
test -s .github/ISSUE_TEMPLATE/feature_request.md && \
test -s .github/PULL_REQUEST_TEMPLATE.md && echo "OK"
```

Expected: `OK`.

- [ ] **Step 7: Commit**

```bash
git add .github/ISSUE_TEMPLATE/bug_report.md .github/ISSUE_TEMPLATE/feature_request.md .github/PULL_REQUEST_TEMPLATE.md
git commit -m "$(cat <<'EOF'
docs: add GitHub issue and pull-request templates

Adds .github/ISSUE_TEMPLATE/bug_report.md, feature_request.md, and
.github/PULL_REQUEST_TEMPLATE.md. Bug template asks for the exact command
and JSON output. PR template mirrors the contributor checklist
(tests, godoc, CHANGELOG, no secrets) and points security-sensitive
reports to SECURITY.md.

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Add `.github/dependabot.yml`

**Files:**
- Create: `.github/dependabot.yml`

**Allowlist:** `.github/dependabot.yml`

- [ ] **Step 1: Verify no `dependabot.yml` already exists**

```bash
test ! -f .github/dependabot.yml && echo "absent OK"
```

Expected: `absent OK`.

- [ ] **Step 2: Write `.github/dependabot.yml` with this exact content**

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "08:00"
      timezone: "UTC"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "go"
    commit-message:
      prefix: "deps"
      include: "scope"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "08:00"
      timezone: "UTC"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "github-actions"
    commit-message:
      prefix: "deps"
      include: "scope"
```

- [ ] **Step 3: Verify the file exists and parses as YAML**

```bash
test -s .github/dependabot.yml && \
python3 -c "import yaml,sys; yaml.safe_load(open('.github/dependabot.yml'))" && echo "YAML OK"
```

Expected: `YAML OK`. If `python3` or PyYAML are unavailable, fall back to:

```bash
test -s .github/dependabot.yml && grep -q "package-ecosystem: \"gomod\"" .github/dependabot.yml && echo "OK"
```

Expected: `OK`.

- [ ] **Step 4: Commit**

```bash
git add .github/dependabot.yml
git commit -m "$(cat <<'EOF'
chore: add dependabot config for weekly Go module + GitHub Actions updates

Configures Dependabot to open weekly PRs on Monday 08:00 UTC for Go
module updates and GitHub Actions version bumps. Caps open PRs at 5 per
ecosystem, labels them, and uses a "deps" commit prefix.

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: README badges + trim Status table

**Files:**
- Modify: `README.md`

**Allowlist:** `README.md`

**Approach:** Surgical edits only. Add a badge row immediately under the
`# polygolem` title. Trim the `## Status` table down to a single
"current state" line and point readers at `CHANGELOG.md` for detail. Do
**not** touch the deposit-wallet pitch, the Tier 1 / Tier 2 explainer,
the command inventory, the packages table, the env-var table, or the
docs link list.

- [ ] **Step 1: Confirm the README is in its current shape**

```bash
head -1 README.md
grep -n "^## Status" README.md
grep -n "^| Phase " README.md
```

Expected:
- Line 1: `# polygolem`
- A `## Status` heading exists.
- A `| Phase | Status |` row exists below it.

If any of these do not match, the README has drifted from the audit
state — **stop and report** before editing.

- [ ] **Step 2: Insert the badge row immediately after the H1 title**

Edit `README.md`. Find the line:

```
# polygolem
```

Replace it with this exact block (preserving the blank line that already
follows):

```markdown
# polygolem

[![CI](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml/badge.svg)](https://github.com/TrebuchetDynamics/polygolem/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TrebuchetDynamics/polygolem)](go.mod)
[![Latest Release](https://img.shields.io/github/v/tag/TrebuchetDynamics/polygolem?label=release&sort=semver)](https://github.com/TrebuchetDynamics/polygolem/releases)
```

Note on the release badge: until the first `v0.1.0` tag is pushed,
shields.io will render this as "no releases" or "no tags". That is
acceptable; the badge updates automatically once a tag exists.

- [ ] **Step 3: Replace the Status section**

Find the current Status section, which starts at `## Status` and includes
the phase table plus the trailing fenced `go test ./...` block. Replace
the entire section (from `## Status` through the closing ` ``` ` of the
`go test` block) with this exact block:

```markdown
## Status

`v0.1.0` — Phase 0 through Phase E plus the May 2026 deposit-wallet
migration are shipped. See [`CHANGELOG.md`](CHANGELOG.md) for the full
release log.

```bash
go test ./...
```
```

- [ ] **Step 4: Verify the badge row and trimmed Status section are both present, and the deposit-wallet pitch is untouched**

```bash
grep -q '!\[CI\]' README.md && \
grep -q '!\[License: MIT\]' README.md && \
grep -q '!\[Go Version\]' README.md && \
grep -q '!\[Latest Release\]' README.md && \
grep -q '## Status' README.md && \
grep -q 'See \[`CHANGELOG.md`\]' README.md && \
grep -q 'deposit-wallet onboard --fund-amount 0.71' README.md && \
echo "README OK"
```

Expected: `README OK`.

- [ ] **Step 5: Verify the old Status table is gone**

```bash
grep -nE '^\| Phase 0 — Go-bot boundary cleanup' README.md && echo "STALE STATUS ROW STILL PRESENT" || echo "removed OK"
```

Expected: `removed OK`.

- [ ] **Step 6: Commit**

```bash
git add README.md
git commit -m "$(cat <<'EOF'
docs: add README badges, trim Status table to a CHANGELOG pointer

Adds CI / license / Go version / latest release badges under the title.
Replaces the seven-row phase table with a single "v0.1.0 shipped" line
that points at CHANGELOG.md for the release log. Deposit-wallet pitch,
command inventory, package table, env vars, and docs links are
unchanged.

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7 (OPTIONAL): Add `docs-site-build` job to CI

**Skip this task if Track 4 has not shipped a docs-site that builds
clean from a fresh clone.** Run `cd docs-site && npm ci && npm run build`
once locally before starting; if it fails, skip Task 7 and address the
docs-site breakage in a Track 4 follow-up plan instead. A red CI is
worse than a missing job.

**Files:**
- Modify: `.github/workflows/ci.yml`

**Allowlist:** `.github/workflows/ci.yml`

- [ ] **Step 1: Sanity-check the docs-site builds locally**

```bash
test -d docs-site && cd docs-site && npm ci && npm run build && cd -
```

Expected: build completes with no error. If anything fails, **abort Task 7** and proceed to Task 8.

- [ ] **Step 2: Confirm the current CI workflow shape**

```bash
cat .github/workflows/ci.yml
```

Expected: a single `test` job that runs `gofmt`, `git diff --exit-code`, `go vet`, and `go test`. If the file has been refactored since Track 1, **stop and report** so the additions can be re-targeted.

- [ ] **Step 3: Replace `.github/workflows/ci.yml` with this exact content**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: gofmt -w .
      - run: git diff --exit-code
      - run: go vet ./...
      - run: go test ./...

  docs-site-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"
          cache-dependency-path: docs-site/package-lock.json
      - run: npm ci
        working-directory: docs-site
      - run: npm run build
        working-directory: docs-site
```

- [ ] **Step 4: Verify the `test` job is unchanged and the new job is present**

```bash
grep -q "^  test:" .github/workflows/ci.yml && \
grep -q "^  docs-site-build:" .github/workflows/ci.yml && \
grep -q "go-version-file: go.mod" .github/workflows/ci.yml && \
grep -q "working-directory: docs-site" .github/workflows/ci.yml && \
echo "CI OK"
```

Expected: `CI OK`.

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "$(cat <<'EOF'
ci: add docs-site-build job

Runs npm ci and npm run build under docs-site/ on every push and PR so
that breakage in the Astro Starlight site fails CI rather than going
unnoticed. Existing Go test job is unchanged.

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: Delete `docs/AUDIT-FINDINGS.md` and run the Track 5 verification gate

**Files:**
- Delete: `docs/AUDIT-FINDINGS.md` (commit)
- No other modifications. Verification commands only.

**Allowlist:** `docs/AUDIT-FINDINGS.md`

- [ ] **Step 1: Confirm Tracks 2–4 consumed every audit finding**

```bash
test -f docs/AUDIT-FINDINGS.md && cat docs/AUDIT-FINDINGS.md
```

Open the file. Every row in the `Track 1 — Doc-side drift` and
`Track 3 — JSON envelope drift` tables must have a non-empty
**Resolution** entry, and the `Findings consumed downstream` table
should reference each consumer. If any row is unresolved, **stop**.
Either file a follow-up plan to capture the remaining items elsewhere
(e.g., open issues), or finish consuming them in the appropriate track,
**before** deleting this file.

`Track 1 — Code-side drift surfaced incidentally` rows describe code
bugs that were logged for later. They do not block deletion as long as
they have been moved into GitHub issues or another tracking surface;
note in the deletion commit message where they went.

- [ ] **Step 2: Delete the file**

```bash
git rm docs/AUDIT-FINDINGS.md
```

Expected: `git status` shows `D ` for `docs/AUDIT-FINDINGS.md`.

- [ ] **Step 3: Verify deletion**

```bash
test ! -f docs/AUDIT-FINDINGS.md && echo "deleted OK"
```

Expected: `deleted OK`.

- [ ] **Step 4: Commit the deletion**

```bash
git add docs/AUDIT-FINDINGS.md
git commit -m "$(cat <<'EOF'
docs: delete docs/AUDIT-FINDINGS.md (overhaul complete)

Tracks 1-4 consumed every doc-side and JSON-envelope finding logged in
this file. Any remaining code-side drift items have been moved to
GitHub issues. Per the documentation overhaul spec, AUDIT-FINDINGS.md
is a working-notes file deleted at the end of Track 5.

Part of Track 5 (Repo Polish) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 5: Run the Track 5 verification gate (no further commits)**

This is the gate from the spec § Track 5. All five commands must succeed.

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem

# 1. Build
go build ./cmd/polygolem
echo "build: $?"

# 2. Tests
go test ./...
echo "test: $?"

# 3. Formatting (must produce no output)
gofmt -l .
echo "gofmt exit: $?"

# 4. Vet
go vet ./...
echo "vet: $?"

# 5. Docs-site build
cd docs-site && npm run build && cd -
echo "docs-site: $?"
```

Expected for each:
- `go build ./cmd/polygolem` — exits 0, produces a `polygolem` binary in the repo root.
- `go test ./...` — exits 0, all packages PASS.
- `gofmt -l .` — exits 0 with **no output**.
- `go vet ./...` — exits 0, no warnings.
- `cd docs-site && npm run build` — exits 0, Astro completes the build.

If any of the five fails, the gate is red. Fix the failure and re-run
the gate; do not paper over.

- [ ] **Step 6: Confirm the working tree is clean of Track 5 artifacts**

```bash
git status --short -- CONTRIBUTING.md SECURITY.md CHANGELOG.md .github/ISSUE_TEMPLATE/ .github/PULL_REQUEST_TEMPLATE.md .github/dependabot.yml .github/workflows/ci.yml README.md docs/AUDIT-FINDINGS.md
```

Expected: **no output** for any path that Tasks 1–8 produced (all
committed). Other unrelated WIP entries on `main` are fine — they are
out of scope for Track 5.

- [ ] **Step 7: Mark Track 5 complete**

No further file change required. Inform the user that Track 5
verification has passed and that the documentation overhaul project
is closed pending tag and release. Tagging `v0.1.0` is out of scope
(see "Out of scope" below) — leave the release-management decision to
the maintainer.

---

## Out of scope (re-stated)

- Release automation — `goreleaser`, signed builds, Homebrew tap,
  publishing the `v0.1.0` tag itself.
- Code-of-conduct file.
- Logo work, marketing site, README screenshots, asciinema casts.
- Doc-link checkers beyond the Astro build pass already exercised by
  Task 7's optional CI job.
- Any Go code change. Track 5 is repo-meta only; if the verification
  gate surfaces a code regression introduced by another track, fix it
  in that track's follow-up plan, not here.
