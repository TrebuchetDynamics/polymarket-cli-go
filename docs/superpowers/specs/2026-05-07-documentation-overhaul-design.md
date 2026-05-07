# Polygolem Documentation Overhaul — Design Spec

- **Date:** 2026-05-07
- **Status:** draft, awaiting user review
- **Author:** brainstormed with Claude
- **Scope:** all polygolem documentation surfaces

## 1. Context and Problem

Polygolem has matured rapidly: 21 internal packages, 5 public packages under
`pkg/`, the May 2026 deposit-wallet migration shipped end-to-end, and ~25
working tests. The documentation has not kept up.

Concrete drift:

- `docs/ARCHITECTURE.md` claims "no public Go SDK in Phase 1" and that the
  repo "intentionally avoids `pkg/` API". Both are now false: `pkg/bookreader`,
  `pkg/bridge`, `pkg/gamma`, `pkg/marketresolver`, and `pkg/pagination` exist.
- `docs/IMPLEMENTATION-PLAN.md` reads as a gap analysis where most "missing"
  items are now built. It is misleading on its face.
- `SKILL.md` documents only the read-only command subset. Deposit-wallet,
  CLOB trading, paper, and bridge command groups are absent. An agent that
  reads this file cannot drive the headline workflow.
- `docs-site/` is an Astro Starlight scaffold with a sidebar of 17 nav items
  but only 7 backing pages. The deposit-wallet flow — the headline feature
  the README leads with — is not in the sidebar at all.
- `docs/COMMANDS.md`, `docs/SAFETY.md`, and `docs/PRD.md` have not been
  audited against the current command tree.

The work below corrects the drift and brings every doc surface to a coherent,
launch-ready state. The driving use cases are: (1) a public GitHub launch,
(2) onboarding a future contributor or future-self, (3) AI-agent integration
via the Claude skill, and (4) eliminating doc rot before it compounds.

## 2. Goals and Non-Goals

### Goals

- Every doc reflects current code reality. No claims that contradict the source
  tree or the binary's behavior.
- Every doc surface has an explicit, single role. Readers know which surface
  to consult for which question.
- The public SDK (`pkg/*`) is genuinely usable via `pkg.go.dev`-style godoc.
- The Claude skill can drive every command group, not just the read-only
  subset.
- A versioned, stable JSON envelope is specified, error and exit codes are
  taxonomized, and current command output is audited against the envelope.
- The repo presents launch-ready: CONTRIBUTING, SECURITY, CHANGELOG, GitHub
  issue/PR templates, badges.

### Non-Goals

- **Implementing** the JSON envelope across the codebase. This spec captures
  the envelope design and the audit findings; the code-alignment work is a
  separate follow-up plan.
- Adding new exported APIs to `pkg/*` discovered while documenting. Such items
  are logged as findings; new exports get their own design.
- Deploying the docs-site (hosting, GitHub Pages config) — out of scope.
- Logo work, screencasts, marketing assets.
- A code-of-conduct file. The repo is solo today; adding CoC is performative
  without a contributor base. Trivial to add later.
- Release automation (goreleaser, homebrew tap, signed builds).

## 3. Doc Surface Roles (Source-of-Truth Map)

The repo keeps three user-visible surfaces plus inline godoc and the agent
manifest. Each has one job. Drift is prevented by making one location
canonical for each topic and the others link to it.

| Surface | Job | Audience | Canonical for |
|---|---|---|---|
| `README.md` | 60-second pitch + one-command onboarding + links | Drive-by GitHub visitor | Project pitch, install, headline command |
| `docs/*.md` | Long-form internal reference | Devs reading the repo | PRD, ARCHITECTURE, SAFETY, JSON-CONTRACT |
| `docs/history/` | Frozen-in-time archive | Archaeology | Past phase plans, behavioral references |
| `docs-site/` (Astro Starlight) | Polished tutorial + reference site | Users / integrators | Tutorial flow, per-command examples, concepts |
| `SKILL.md` | Agent-driveable command spec | Claude / scripted callers | The full command catalog for agents |
| Godoc on `pkg/*` | Per-symbol SDK reference | Go developers | Public Go API |
| Godoc on `internal/*` | Package-level orientation only | Source readers | Package purpose and entry points |

Cross-cutting reference content (JSON contract, error codes, env vars,
architecture) lives canonically in `docs/*.md`. The Starlight pages for these
topics are short prose plus a "Source of truth" link. CLI reference pages
defer to `polygolem <cmd> --help` as normative for flag semantics.

## 4. Tracks

The work is organized into five tracks executed in dependency order. Each
track has its own verification gate; downstream tracks should not start until
the upstream gate passes.

### Track 1 — Audit and Truth Pass

Foundation work. Establishes a known-good baseline so subsequent tracks build
on truth.

| File | Action | Notes |
|---|---|---|
| `docs/ARCHITECTURE.md` | **Rewrite** | New package map (21 internal, 5 public). New dependency-flow diagram. Document read-only / paper / live mode boundary as currently implemented, including deposit-wallet signature type. Drop the "no SDK" claims. |
| `docs/PRD.md` | **Audit + reconcile** | Read end-to-end against current code. Mark fulfilled requirements ✅, drift items ⚠️. Preserve historical "why". No rewrite from scratch. |
| `docs/COMMANDS.md` | **Regenerate** | Walk every command in the CLI tree. Each command present with correct flags and one example. Cross-check against `polygolem --help`. |
| `docs/SAFETY.md` | **Audit + extend** | Verify read-only-by-default claims hold. Add rules for builder credential handling, batch signing, and live-money commands. |
| `README.md` | **Spot-fix only** | Mostly fresh. Fix only contradictions surfaced by the audit. No rewrite. |
| `docs/REFERENCE-RUST-CLI.md` | **Move** to `docs/history/` | Keeps the behavioral audit available without misleading new readers. |
| `docs/PHASE0-GOBOT-MIGRATION.md` | **Move** to `docs/history/` | Useful for anyone debugging the go-bot ↔ polygolem boundary. |
| `docs/IMPLEMENTATION-PLAN.md` | **Delete** | Gap-analysis content is now stale across the board. Git history is the archive. |
| `docs/history/README.md` | **Create** | One-paragraph index of what each archived doc captured and when. |

**Track output:** `docs/AUDIT-FINDINGS.md` — a working-notes file listing
every drift item discovered while auditing. Tracks 2, 3, and 4 consume this
file. Deleted at the end of Track 5.

**Verification:**
- `git grep -nE "no public Go SDK|intentionally avoids \`pkg/\`"` returns no
  hits in `docs/`.
- Every command in `polygolem --help` appears in `docs/COMMANDS.md`.
- `docs/history/README.md` exists; both archived files moved.

### Track 2 — Internal Documentation (godoc layer)

Two-tier strategy: orientation for everything, full reference for the public
SDK.

#### Tier A — `doc.go` for every package

One file per package containing only the package-level comment. Format:

```go
// Package <name> <one-sentence purpose>.
//
// <2-4 sentence elaboration: what problem it solves, what it does NOT do,
// the most important type or function to look at first.>
//
// This package is internal and not part of the polygolem public SDK.
package <name>
```

Targets (21 internal packages, plus auditing any existing comments to merge,
not duplicate): `auth`, `cli`, `clob`, `config`, `dataapi`, `errors`,
`execution`, `gamma`, `marketdiscovery`, `modes`, `orders`, `output`,
`paper`, `polytypes`, `preflight`, `relayer`, `risk`, `rpc`, `stream`,
`transport`, `wallet`.

#### Tier B — full godoc on `pkg/*`

For each of `pkg/bookreader`, `pkg/bridge`, `pkg/gamma`, `pkg/marketresolver`,
`pkg/pagination`:

1. Package comment: purpose, when to use, when not to use, stability promise.
2. Godoc on every exported symbol (types, methods, functions, constants).
   One-line minimum; longer for non-obvious behavior.
3. At least one runnable `Example_*` function in `*_example_test.go`. These
   compile, run as tests, and render on `pkg.go.dev`.
4. No `internal/...` references in doc text. Public-surface only.

**Verification:**
```bash
go vet ./...                        # malformed doc strings caught
go test ./...                       # examples must compile and pass
go doc ./pkg/bookreader             # spot-check rendering
go doc ./pkg/bridge
go doc ./pkg/gamma
go doc ./pkg/marketresolver
go doc ./pkg/pagination
```

Optional add: `.revive.toml` enabling `package-comments` and `exported` rules
so future drift gets caught by linting. CI wiring is out of scope here.

### Track 3 — Agent Surface

Three deliverables: the v1 JSON envelope spec, the JSON-contract reference
docs, and the SKILL.md rewrite.

#### 3a. v1 JSON Envelope

Every command's `--json` output adopts a single envelope:

**Success:**
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

**Error:**
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
  "meta": { "command": "...", "ts": "...", "duration_ms": 12 }
}
```

Decisions:
- `ok` is the single boolean an agent checks.
- `version` is a string (not int) so future minor versions like `"1.1"` parse.
- `data` and `error` are mutually exclusive — never both.
- `meta` is always present, even on error, for uniform logs and analytics.
- Field naming inside the envelope is `snake_case` (matches Polymarket's
  upstream APIs). Inside `data`, fields preserve whatever the underlying
  upstream API returned. We do not re-case upstream payloads.

#### 3b. Error code taxonomy

Two-level: `category` (broad) plus `code` (specific). Categories:

| Category | When | Example codes |
|---|---|---|
| `usage` | Bad flags, missing args, conflicting options | `USAGE_FLAG_MISSING`, `USAGE_FLAG_INVALID` |
| `auth` | Missing/invalid creds, bad signature | `AUTH_PRIVATE_KEY_MISSING`, `AUTH_BUILDER_MISSING`, `AUTH_SIG_INVALID` |
| `validation` | Input parse / range failures | `VALIDATION_TOKEN_ID_INVALID`, `VALIDATION_AMOUNT_OUT_OF_RANGE` |
| `gate` | Mode / safety / preflight refusal | `GATE_LIVE_DISABLED`, `GATE_PREFLIGHT_FAILED`, `GATE_RISK_LIMIT` |
| `network` | HTTP failure, DNS, timeout, retry exhausted | `NETWORK_TIMEOUT`, `NETWORK_DNS`, `NETWORK_TLS` |
| `protocol` | Upstream API errored or returned unexpected shape | `PROTOCOL_GAMMA_4XX`, `PROTOCOL_CLOB_5XX`, `PROTOCOL_RELAYER_REJECTED` |
| `chain` | On-chain RPC, nonce, revert | `CHAIN_NONCE_TOO_LOW`, `CHAIN_REVERTED`, `CHAIN_INSUFFICIENT_FUNDS` |
| `internal` | Bug. Code asserts something that should never happen | `INTERNAL_PANIC`, `INTERNAL_INVARIANT` |

Codes are stable strings in `SCREAMING_SNAKE`. Adding new codes is
non-breaking; renaming or removing is breaking and bumps `version`.

#### 3c. Exit-code matrix

| Exit | Meaning |
|---|---|
| `0` | success (`ok: true`) |
| `1` | generic failure when category does not map cleanly |
| `2` | `usage` |
| `3` | `auth` |
| `4` | `validation` |
| `5` | `gate` |
| `6` | `network` |
| `7` | `protocol` |
| `8` | `chain` |
| `9` | `internal` |

Wrapping shell scripts can branch on exit code without parsing JSON. The
JSON envelope remains authoritative; exit codes are a convenience.

#### 3d. Documentation deliverables

- `docs/JSON-CONTRACT.md` — canonical envelope spec, all codes with
  descriptions, exit-code matrix, versioning policy (when `version` bumps,
  what counts as breaking).
- `docs-site/.../reference/json-contract.mdx` — short prose plus a "Source of
  truth" link to `docs/JSON-CONTRACT.md`.
- `docs-site/.../reference/error-codes.mdx` — same pattern for codes.
- `docs/AUDIT-FINDINGS.md` (deliverable from Track 1) gets a **JSON drift
  section**: per-command list of where current output diverges from the v1
  envelope. This is the handoff to the code-alignment plan that lives outside
  this work.

#### 3e. SKILL.md rewrite

Replace the current read-only-only file with a complete agent spec.

Sections:

- **Overview** — what polygolem does, what it does not.
- **JSON contract reference** — short, with link to full doc.
- **Command catalog** by group: `discover`, `orderbook`, `clob`,
  `deposit-wallet`, `paper`, `bridge`, `health`, `version`, `preflight`. For
  each command: purpose, required flags, env vars consumed, sample success
  JSON snippet, sample error JSON snippet, notable caveats.
- **Common workflows** — recipes the agent is expected to execute end-to-end.
  At minimum: "onboard a new account", "place a trade",
  "check market depth", "find a tradeable market".
- **Safety surface** — read-only is default; what flags are required to opt
  into mutating commands; what the agent must not do (e.g., never use
  `POLYMARKET_PRIVATE_KEY` from user-pasted text without explicit
  confirmation).
- **Env-var reference** — every variable, when each is required, redaction
  guarantees.

**Out of scope (explicit):** code changes to align command outputs to the v1
envelope. Captured as findings here, implemented in a separate follow-up plan
after this brainstorm/spec/plan cycle finishes.

**Verification:**
- `docs/JSON-CONTRACT.md` exists and is internally consistent (every code
  used in examples is defined in the taxonomy).
- `SKILL.md` covers every command group present in `polygolem --help`.
- A spot-check: pick three commands (one read-only, one mutating, one
  paper), confirm SKILL.md examples are runnable as documented.

### Track 4 — Public docs-site (Astro Starlight)

**Goal:** A polished site that turns a drive-by visitor into someone who has
run `polygolem deposit-wallet onboard` in under 10 minutes.

#### Sidebar redesign

The current sidebar buries the headline. New structure:

```
Getting Started
├── Introduction                     index.mdx                              [refresh]
├── Installation                     getting-started/installation
└── Quick Start                      getting-started/quickstart             [refresh]

Deposit Wallet (May 2026)         ← NEW SECTION, headline placement
├── Why this matters                 deposit-wallet/why                     [NEW]
├── One-command onboarding           deposit-wallet/onboard                 [NEW]
├── Step-by-step flow                deposit-wallet/flow                    [NEW]
└── Troubleshooting                  deposit-wallet/troubleshooting         [NEW]

Guides
├── Market Discovery                 guides/market-discovery                [refresh]
├── Orderbook Data                   guides/orderbook-data                  [NEW]
├── Paper Trading                    guides/paper-trading                   [NEW]
├── Placing Real Orders              guides/placing-orders                  [NEW]
├── Bridge & Funding                 guides/bridge-funding                  [NEW]
└── Go-Bot Integration               guides/go-bot-integration              [NEW]

Concepts
├── Polymarket API Overview          concepts/polymarket-api                [refresh]
├── Markets, Events & Tokens         concepts/markets-events-tokens         [NEW]
├── Modes (read-only / paper / live) concepts/modes                         [NEW]
├── Signature Types                  concepts/signature-types               [NEW]
├── Builder Attribution              concepts/builder-attribution           [NEW]
├── POLY_1271 Order Signing          concepts/poly-1271                     [NEW]
├── Safety Model                     concepts/safety                        [NEW]
└── Architecture                     concepts/architecture                  [NEW]

Reference
├── CLI Commands                     reference/cli                          [refresh]
│   ├── discover                     reference/cli/discover                 [NEW]
│   ├── orderbook                    reference/cli/orderbook                [NEW]
│   ├── deposit-wallet               reference/cli/deposit-wallet           [NEW]
│   ├── clob                         reference/cli/clob                     [NEW]
│   ├── paper                        reference/cli/paper                    [NEW]
│   └── bridge                       reference/cli/bridge                   [NEW]
├── Go SDK                           reference/sdk                          [refresh]
├── JSON Contract                    reference/json-contract                [NEW]
├── Error & Exit Codes               reference/error-codes                  [NEW]
└── Environment Variables            reference/env-vars                     [NEW]

For Agents
└── Using polygolem from Claude       agents/claude-skill                    [NEW]
```

Approximately 33 pages total: 7 existing (6 refresh + `installation` left
as-is), ~26 to create.

#### Source-of-truth strategy (drift prevention)

1. Cross-cutting reference docs (JSON contract, error codes, env vars,
   architecture) are canonical in `docs/*.md`. Starlight pages are short
   prose plus a "Source of truth" link. MDX includes were considered and
   rejected — the build-complexity cost is not worth it for this many pages.
2. CLI reference pages link to `polygolem <cmd> --help` as normative for
   flag semantics. Pages document workflows and examples.
3. Each page has a footer snippet: "Source of truth: `<link>`".

#### Per-page quality bar

- Every code block runnable as written. No `<your-id>` placeholders unless
  explicitly marked.
- Every command example shows expected JSON output, truncated where noisy.
- Every concept page has at least one diagram or table — no walls of prose.
- Deposit-wallet pages assume zero prior knowledge — this is the entry point
  for new users.

#### Verification

```bash
cd docs-site && npm run build         # Astro build passes with zero warnings
```

Manually walk every sidebar link. Broken links fail the track.

#### Out of scope (explicit)

- Hosting and deploy. Site builds locally; deploying to GitHub Pages is a
  separate concern.
- Search backend. Pagefind is the Starlight default — leave it.
- Internationalization, dark-mode tweaks beyond Starlight defaults.

### Track 5 — Repo polish

#### Files to add

| File | Content | Notes |
|---|---|---|
| `CONTRIBUTING.md` | Build, run tests, file an issue, the TDD expectation, how the doc surfaces split. | Short, ~80 lines. |
| `SECURITY.md` | How to report a vulnerability privately. Email or GitHub Security Advisories. | ~30 lines. **Critical** for a tool that holds private keys. |
| `CHANGELOG.md` | Keep-a-Changelog format. Seed with `## [Unreleased]` plus `## [0.1.0] — 2026-05-07` summarizing Phase 0–E plus deposit-wallet migration. | Going forward, every PR updates `[Unreleased]`. |
| `.github/ISSUE_TEMPLATE/bug_report.md` | Standard bug template — reproduction, expected, actual, command + JSON output. | |
| `.github/ISSUE_TEMPLATE/feature_request.md` | Standard. | |
| `.github/PULL_REQUEST_TEMPLATE.md` | Checklist: tests pass, godoc updated if exported surface changed, CHANGELOG updated. | |
| `.github/dependabot.yml` | Weekly Go module updates. | One-time setup, infinite payoff. |

#### Files to refresh

| File | Change |
|---|---|
| `README.md` | Add badges (CI status, license, Go version, latest tag). Shorten the "Status" table — move detail to CHANGELOG. Keep the deposit-wallet pitch front-and-center. |
| `.github/workflows/ci.yml` | Optional: add a `docs-site-build` job that runs `npm ci && npm run build` so doc-site breakage fails CI. Decide at plan time. |

#### Out of scope (explicit)

- Release automation (goreleaser, signed builds, homebrew tap).
- Code-of-conduct file. Repo is solo today; trivial to add later.
- Logo work, marketing site, README screenshots / asciinema casts.
- Doc-link checkers beyond the Astro build pass from Track 4.

#### Verification

```bash
go build ./cmd/polygolem               # binary still builds
go test ./...                          # all packages still green
gofmt -l .                             # zero output
go vet ./...                           # zero output
cd docs-site && npm run build          # Astro build green
```

## 5. Sequencing and Dependencies

```
Track 1 (Audit & Truth Pass)
    │
    ├── produces: AUDIT-FINDINGS.md, refreshed ARCHITECTURE/PRD/COMMANDS/SAFETY
    │
    ├──> Track 2 (godoc): needs accurate package list from refreshed ARCHITECTURE.
    │
    ├──> Track 3 (Agent Surface): needs accurate command catalog from refreshed COMMANDS.
    │
    ├──> Track 4 (docs-site): needs all reference docs canonical and accurate.
    │              ↑
    │              │ Track 4 also depends on Track 3 for reference/json-contract.mdx
    │              │ and reference/error-codes.mdx, and on Track 2 for the SDK page.
    │
    └──> Track 5 (Repo polish): can start anytime; final verification gate
                                 depends on Track 4 (docs-site build).
```

Tracks 2 and 3 can run in parallel after Track 1 completes. Track 4 starts
once both 2 and 3 are done. Track 5 can be interleaved but its final
verification waits on Track 4.

## 6. Success Criteria

The overhaul is done when all of the following hold:

- [ ] `git grep` finds zero contradictions between docs and code (the audit
      asserts each known-drift item has been resolved).
- [ ] `go test ./...` passes; `go vet ./...` clean; `gofmt -l .` empty.
- [ ] `go doc ./pkg/<each>` renders complete package, type, and example
      documentation for all five public packages.
- [ ] `docs-site` builds with `npm run build` with zero warnings; every
      sidebar link resolves.
- [ ] `SKILL.md` covers every command in `polygolem --help`, with sample
      success and error JSON for each command group.
- [ ] `docs/JSON-CONTRACT.md` exists, is internally consistent, and is
      referenced from at least Starlight, SKILL.md, and ARCHITECTURE.
- [ ] `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md`, and the GitHub
      issue/PR templates exist.
- [ ] `README.md` has CI / license / Go-version badges and no claims that
      contradict the audited docs.
- [ ] `docs/history/` contains the two archived files plus a one-paragraph
      `README.md`.
- [ ] `docs/IMPLEMENTATION-PLAN.md` is removed.
- [ ] `docs/AUDIT-FINDINGS.md` ends the project with its **JSON drift
      section** populated, then is itself deleted as the very last step
      (the findings live on as the input to the follow-up code-alignment
      plan, which is captured in a separate spec).

## 7. Follow-up Work (out of this spec)

Captured here so the through-line is not lost when implementation begins.

1. **JSON envelope code alignment.** Implement the v1 envelope across every
   command. Findings come from `docs/AUDIT-FINDINGS.md` (Track 3 JSON drift
   section). Gets its own brainstorm → spec → plan → implement cycle.
2. **CI wiring for godoc-lint.** If `.revive.toml` is added in Track 2,
   wire `revive` into `.github/workflows/ci.yml` as a non-blocking job, then
   later promote to blocking.
3. **Docs-site deploy.** GitHub Pages or other hosting. Currently builds
   locally only.
4. **Code-of-conduct, release automation, logo work.** Deferred per Track 5.

## 8. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Audit surfaces more drift than expected; Track 1 balloons. | Time-box Track 1. Anything beyond a fixed budget gets logged in `AUDIT-FINDINGS.md` and deferred, not chased. |
| `pkg/*` godoc reveals API design weaknesses. | Track 2 logs but does not redesign. New design is its own follow-up spec. |
| The v1 JSON envelope, once specified, makes a large fraction of current command output non-conformant. | Expected. The audit findings are the deliverable; alignment is a separate plan. The spec's value is in pinning the target. |
| Starlight build complexity (MDX, sidebar config) consumes time. | Track 4 builds incrementally; verification is `npm run build` clean, not pixel-perfect styling. |
| Doc rot returns six months from now. | Per-page "Source of truth" footers and the `revive` lint option (Track 2) are the mitigations. Long-term: doc-tests-as-CI is on the follow-up list. |

## 9. Open Questions

None remain from brainstorm. All scoping decisions resolved by user during
the brainstorming session 2026-05-07.
