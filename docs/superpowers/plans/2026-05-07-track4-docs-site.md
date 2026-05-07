# Track 4 — Docs Site Overhaul: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the polygolem Astro Starlight site into the front door for the project. A drive-by visitor should be able to run `polygolem deposit-wallet onboard` in under 10 minutes. The headline (deposit-wallet) currently has no presence in the sidebar; this plan adds a dedicated section, refreshes 6 existing pages, and creates 26 new pages, then verifies the build is clean.

**Architecture:** Sidebar-first. Update `astro.config.mjs` to the spec'd structure once, accept that intermediate builds will warn until backing pages exist, then create pages section by section in the same order they appear in the sidebar. Cross-cutting reference docs (`docs/JSON-CONTRACT.md`, `docs/ARCHITECTURE.md`, `docs/SAFETY.md`, `docs/COMMANDS.md`, `SKILL.md`) remain canonical in `docs/`; Starlight pages are short prose plus a "Source of truth" footer pointing back to them. CLI reference pages cite `polygolem <cmd> --help` as normative.

**Tech Stack:** Astro 5 + `@astrojs/starlight` ^0.32. MDX content under `docs-site/src/content/docs/`. Build via `cd docs-site && npm run build`.

**Spec:** `docs/superpowers/specs/2026-05-07-documentation-overhaul-design.md` § Track 4.

**Dependencies on other tracks:**
- Track 1 must have produced canonical `docs/ARCHITECTURE.md`, `docs/COMMANDS.md`, `docs/SAFETY.md`. Tasks 5 and 7 link to those.
- Track 2 must have produced full godoc on `pkg/*`. Task 2 (refresh of `reference/sdk`) links to pkg.go.dev with a note that the SDK is published from main.
- Track 3 must have produced `docs/JSON-CONTRACT.md` and `SKILL.md`. Tasks 7 and 8 link to those.

If a Track 4 task discovers that one of those upstream artifacts does not yet exist, **stop and report** — do not invent canonical content here.

**Working tree caveat:** `main` has uncommitted WIP unrelated to docs-site (e.g., `internal/cli/deposit_wallet.go`, `internal/relayer/`, `pkg/gamma/`, `docs/DEPOSIT-WALLET-MIGRATION.md`). Each task ships a per-task file allowlist; implementers use `git add <specific paths>` and never `git add -A` / `git add .`.

---

## Task Inventory

| # | Task | Output | Pages |
|---|---|---|---|
| 1 | Update `astro.config.mjs` sidebar to new 6-section structure | Refreshed config | 0 (config only) |
| 2 | Refresh 6 existing pages to align with new sidebar and headlines | Edited mdx files | 6 refresh |
| 3 | Create Deposit Wallet section (4 pages) | New section | 4 new |
| 4a | Create new Guides — orderbook-data, paper-trading, placing-orders | New mdx files | 3 new |
| 4b | Create new Guides — bridge-funding, go-bot-integration | New mdx files | 2 new |
| 5a | Create Concepts — markets-events-tokens, modes, signature-types, builder-attribution | New mdx files | 4 new |
| 5b | Create Concepts — poly-1271, safety, architecture | New mdx files | 3 new |
| 6 | Create reference/cli per-group pages (6 pages) | New mdx files | 6 new |
| 7 | Create reference link pages — json-contract, error-codes, env-vars | New mdx files | 3 new |
| 8 | Create agents/claude-skill page | New mdx file | 1 new |
| 9 | Final Track 4 verification gate | Green build, manual sidebar walk | 0 |

Total: 33 sidebar entries (7 existing × refresh + 26 new). Each task is one commit unless explicitly marked.

---

## Conventions for every new page

Every new mdx file under this plan MUST satisfy:

1. **Frontmatter:** at minimum `title:` and `description:`. Optionally `sidebar:` for ordering hints.
2. **Source-of-truth footer.** Last section is a small block:
   ```markdown
   ## Source of truth

   - [`docs/<FILE>.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/<FILE>.md)
   - `polygolem <cmd> --help` (run locally)
   ```
   Pages that don't have a single canonical doc may omit the footer; pages that link to multiple sources include all relevant ones.
3. **No placeholders in code blocks.** A reader should be able to copy-paste any block as-is. Token IDs / market IDs may be obvious examples (`"123..."`), but any `<your-id>` syntax is forbidden unless explicitly marked as "replace this." Prefer an annotated comment.
4. **Expected output.** Every command example shows truncated JSON output where it materially helps the reader. JSON in fenced ` ```json ` blocks. Truncate noisy fields with `...`.
5. **At least one diagram or table per concept page.** Concept pages without a table or diagram are incomplete.
6. **Deposit-wallet pages assume zero prior knowledge.** Define every term inline the first time it appears in the page.
7. **Internal links use Starlight slug form.** E.g., `/deposit-wallet/onboard` (no `.mdx`, no leading domain).

---

## Task 1: Update `docs-site/astro.config.mjs` sidebar

**Files:**
- Modify: `docs-site/astro.config.mjs`

**Allowlist for `git add`:** `docs-site/astro.config.mjs`

**Why first:** Astro Starlight resolves sidebar slugs at build time. Updating the sidebar shape first lets every subsequent task verify "the page I added shows up where the spec says it should." Intermediate `npm run build` runs between Tasks 1 and 9 will emit warnings for slugs that point to not-yet-created pages; this is **expected** and acceptable. The Task 9 gate is what must be clean.

- [ ] **Step 1: Confirm the file matches the audited shape**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
wc -l docs-site/astro.config.mjs
grep -n 'label:\|slug:' docs-site/astro.config.mjs
```

Expected: ~58 lines, with the four current sections (Getting Started, Guides, Concepts, Reference). If the structure has drifted from this baseline, **stop and report**.

- [ ] **Step 2: Replace the sidebar definition with the new 6-section structure**

Open `docs-site/astro.config.mjs` and replace the entire `sidebar: [ ... ]` array (lines 16–54 in the audited file) with the block below. Leave everything outside the array (`site:`, `integrations:`, `title:`, `description:`, `logo:`, `social:`, `customCss:`) untouched.

```js
      sidebar: [
        {
          label: "Getting Started",
          items: [
            { label: "Introduction", slug: "" },
            { label: "Installation", slug: "getting-started/installation" },
            { label: "Quick Start", slug: "getting-started/quickstart" },
          ],
        },
        {
          label: "Deposit Wallet (May 2026)",
          items: [
            { label: "Why this matters", slug: "deposit-wallet/why" },
            { label: "One-command onboarding", slug: "deposit-wallet/onboard" },
            { label: "Step-by-step flow", slug: "deposit-wallet/flow" },
            { label: "Troubleshooting", slug: "deposit-wallet/troubleshooting" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Market Discovery", slug: "guides/market-discovery" },
            { label: "Orderbook Data", slug: "guides/orderbook-data" },
            { label: "Paper Trading", slug: "guides/paper-trading" },
            { label: "Placing Real Orders", slug: "guides/placing-orders" },
            { label: "Bridge & Funding", slug: "guides/bridge-funding" },
            { label: "Go-Bot Integration", slug: "guides/go-bot-integration" },
          ],
        },
        {
          label: "Concepts",
          items: [
            { label: "Polymarket API Overview", slug: "concepts/polymarket-api" },
            { label: "Markets, Events & Tokens", slug: "concepts/markets-events-tokens" },
            { label: "Modes (read-only / paper / live)", slug: "concepts/modes" },
            { label: "Signature Types", slug: "concepts/signature-types" },
            { label: "Builder Attribution", slug: "concepts/builder-attribution" },
            { label: "POLY_1271 Order Signing", slug: "concepts/poly-1271" },
            { label: "Safety Model", slug: "concepts/safety" },
            { label: "Architecture", slug: "concepts/architecture" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "CLI Commands", slug: "reference/cli" },
            { label: "discover", slug: "reference/cli/discover" },
            { label: "orderbook", slug: "reference/cli/orderbook" },
            { label: "deposit-wallet", slug: "reference/cli/deposit-wallet" },
            { label: "clob", slug: "reference/cli/clob" },
            { label: "paper", slug: "reference/cli/paper" },
            { label: "bridge", slug: "reference/cli/bridge" },
            { label: "Go SDK", slug: "reference/sdk" },
            { label: "JSON Contract", slug: "reference/json-contract" },
            { label: "Error & Exit Codes", slug: "reference/error-codes" },
            { label: "Environment Variables", slug: "reference/env-vars" },
          ],
        },
        {
          label: "For Agents",
          items: [
            { label: "Using polygolem from Claude", slug: "agents/claude-skill" },
          ],
        },
      ],
```

- [ ] **Step 3: Confirm structural shape**

```bash
grep -c '{ label:' docs-site/astro.config.mjs
```

Expected: `33` (3 + 4 + 6 + 8 + 11 + 1).

- [ ] **Step 4: Run the build and observe warnings**

```bash
cd docs-site && npm run build 2>&1 | tail -40
```

Expected: build completes, but emits `[WARN]` lines for every slug whose backing mdx file does not yet exist. This is acceptable for Task 1. **Do not** try to silence warnings here — Tasks 2–8 add the missing pages, and Task 9 verifies a fully clean build. If `npm run build` exits non-zero (rather than warning), capture the error and resolve it before commit.

- [ ] **Step 5: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/astro.config.mjs
git commit -m "$(cat <<'EOF'
docs(site): restructure Starlight sidebar to 6-section overhaul layout

Adds the Deposit Wallet (May 2026) headline section, expands Guides /
Concepts / Reference per the Track 4 spec, and adds a For Agents section
for the SKILL.md surface. Backing pages added in subsequent commits;
build will warn on missing slugs until Tasks 2-8 land.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Refresh the 6 existing pages

**Files (allowlist for `git add`):**
- `docs-site/src/content/docs/index.mdx`
- `docs-site/src/content/docs/getting-started/installation.mdx`
- `docs-site/src/content/docs/getting-started/quickstart.mdx`
- `docs-site/src/content/docs/guides/market-discovery.mdx`
- `docs-site/src/content/docs/concepts/polymarket-api.mdx`
- `docs-site/src/content/docs/reference/cli.mdx`
- `docs-site/src/content/docs/reference/sdk.mdx`

**Why one task:** These are surgical refreshes — adding the deposit-wallet headline, fixing one or two stale lines, and adding "Source of truth" footers. They do not justify seven separate commits.

- [ ] **Step 1: Refresh `index.mdx`**

The current homepage does not mention deposit-wallet. The spec calls it the headline. Add a deposit-wallet hero block immediately after the opening tagline. Keep existing content; do not delete the "What can you do?" or "No credentials needed" sections. Add a card linking to `/deposit-wallet/why` in the bottom CardGrid.

Specifically:
- Keep the title and the "Polygolem is the single source of truth..." paragraph.
- Insert this block **before** the existing `## What can you do?` heading:
  ```markdown
  ## New: One-command deposit wallet (May 2026)

  Polymarket required all new accounts to migrate to **deposit wallets**
  (POLY_1271 signing) on May 14, 2026. Polygolem ships a single command
  that derives, deploys, approves, and funds the wallet:

  ```bash
  polygolem deposit-wallet onboard --fund-amount 50
  ```

  See [Why this matters](/deposit-wallet/why) and the
  [step-by-step flow](/deposit-wallet/flow).
  ```
- In the closing `<CardGrid>` block, add `<Card title="Deposit Wallet" href="/deposit-wallet/why" />` as the first card so it leads.

- [ ] **Step 2: Refresh `getting-started/installation.mdx`**

The current page is mostly correct. Two small fixes:
- After the `Verify` section, add a short subsection:
  ```markdown
  ## Verify deposit-wallet support

  ```bash
  ./polygolem deposit-wallet --help
  ```

  If the output lists `derive`, `status`, `deploy`, `batch`, `approve`,
  `fund`, `nonce`, and `onboard`, your build includes the May 2026
  signing path.
  ```
- Add a `## Source of truth` footer:
  ```markdown
  ## Source of truth

  - [`README.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/README.md)
  - `polygolem version`, `polygolem health`
  ```

- [ ] **Step 3: Refresh `getting-started/quickstart.mdx`**

The existing page has three workflows. Make the deposit-wallet workflow the **first** example. Renumber.

- Rename the current `## 1. Find and inspect a market` to `## 2. Find and inspect a market`.
- Renumber `## 2. Check if a market is tradable` → `## 3.` and `## 3. Use as a Go SDK` → `## 4.`.
- Insert this new first section above the renumbered list:
  ```markdown
  ## 1. Onboard a deposit wallet (May 2026 path)

  Polymarket requires deposit wallets for every new account from May 14, 2026.
  One command derives the wallet, deploys it on-chain, approves spend, and
  funds it from your EOA:

  ```bash
  export POLYMARKET_PRIVATE_KEY="0x..."             # signer EOA
  export POLYMARKET_BUILDER_API_KEY="..."           # from your builder dashboard
  export POLYMARKET_BUILDER_SECRET="..."
  export POLYMARKET_BUILDER_PASSPHRASE="..."

  polygolem deposit-wallet onboard --fund-amount 50 --json
  ```

  Expected JSON (truncated):

  ```json
  {
    "deposit_wallet": "0xabc...def",
    "deployed": true,
    "approved": true,
    "funded": { "amount_pusd": "50", "tx_hash": "0x..." },
    "ready": true
  }
  ```

  See the [step-by-step flow](/deposit-wallet/flow) for what each phase does.
  ```
- Add a `## Source of truth` footer linking to `docs/COMMANDS.md` and `docs/DEPOSIT-WALLET-MIGRATION.md`.

- [ ] **Step 4: Refresh `guides/market-discovery.mdx`**

This page is largely correct. Fixes:
- Confirm every Gamma method named in the table actually exists in `internal/gamma`. Verify with:
  ```bash
  for m in Markets Events Search Tags Series Teams Comments Profiles SportsMarketTypes MarketByToken EventsKeyset MarketsKeyset; do
    grep -rqn "func.*$m" internal/gamma pkg/gamma 2>/dev/null || echo "MISSING: $m"
  done
  ```
  If anything prints `MISSING:`, remove that row from the table rather than fabricating.
- Add a `## Source of truth` footer:
  ```markdown
  ## Source of truth

  - [`docs/COMMANDS.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/COMMANDS.md) — `discover *` reference
  - `polygolem discover --help`
  ```

- [ ] **Step 5: Refresh `concepts/polymarket-api.mdx`**

This page is structurally correct. One required addition: a row noting where deposit wallets live in the Polymarket stack. Append at the end (before any footer) a new section:

```markdown
## Builder Relayer (`relayer-v2.polymarket.com`)

**Builder credentials required (L2-equivalent for batch/proxy operations).**

Used by `polygolem deposit-wallet deploy/batch/approve/onboard`. Submits
calldata batches to the relayer; the relayer pays gas and brokers the
on-chain action. See
[Architecture](/concepts/architecture) and
[Builder Attribution](/concepts/builder-attribution).
```

Add a `## Source of truth` footer linking to `docs/ARCHITECTURE.md`.

- [ ] **Step 6: Refresh `reference/cli.mdx`**

The current command tree is incomplete (missing `auth`, `bridge`, `clob`, `events`, `live`, `paper`, `deposit-wallet`). Replace the whole file with this **index** version that:
- Lists every top-level command group with one-line descriptions and a link to the per-group reference page.
- Notes that the canonical reference is `polygolem <cmd> --help` and `docs/COMMANDS.md`.

```markdown
---
title: CLI Commands
description: Index of every polygolem command group with links to per-group reference pages.
---

`polygolem` exposes one CLI per command group. Each group has a dedicated
reference page below; canonical flag semantics live in
`polygolem <cmd> --help` and the generated `docs/COMMANDS.md`.

## Command groups

| Group | Auth | Reference |
|---|---|---|
| `discover` | none | [discover](/reference/cli/discover) |
| `orderbook` | none | [orderbook](/reference/cli/orderbook) |
| `deposit-wallet` | builder + EOA | [deposit-wallet](/reference/cli/deposit-wallet) |
| `clob` | EOA + L2 | [clob](/reference/cli/clob) |
| `paper` | none (local) | [paper](/reference/cli/paper) |
| `bridge` | none | [bridge](/reference/cli/bridge) |

## Top-level utilities

| Command | Purpose |
|---|---|
| `polygolem health` | Probe Gamma + CLOB reachability. |
| `polygolem preflight` | Run safety / readiness gates. |
| `polygolem version` | Print version. |
| `polygolem auth` | Inspect/derive auth-related identifiers. |
| `polygolem events` | Stream WebSocket events (market or user). |
| `polygolem live` | Live-trading commands (gated). |

## Global flags

| Flag | Description |
|---|---|
| `--json` | Emit JSON instead of a human table. |
| `--help` | Show help for the current command. |

## Source of truth

- [`docs/COMMANDS.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/COMMANDS.md)
- `polygolem <cmd> --help`
```

- [ ] **Step 7: Refresh `reference/sdk.mdx`**

The page documents `bookreader`, `marketresolver`, `bridge`, `pagination` — but the current `pkg/` directory also includes `pkg/gamma`. Add a `pkg/gamma` section. Also add a note about pkg.go.dev publication:

- After the title block, add:
  ```markdown
  Full godoc: <https://pkg.go.dev/github.com/TrebuchetDynamics/polygolem>
  (published from `main`; may lag local `pkg/` by minutes after a push).
  ```
- Add a new section between `pkg/bridge` and `pkg/pagination`:
  ```markdown
  ## `pkg/gamma`

  Read-only Gamma API surface for embedded use.

  ```go
  import "github.com/TrebuchetDynamics/polygolem/pkg/gamma"

  client := gamma.NewClient("")
  markets, err := client.Markets(ctx, gamma.MarketsParams{Limit: 10})
  ```

  | Method | Returns |
  |---|---|
  | `Markets(ctx, params)` | `[]Market` |
  | `Events(ctx, params)` | `[]Event` |
  | `Search(ctx, query)` | `SearchResponse` |
  ```
  Verify signatures match `pkg/gamma/` before committing; remove rows that don't exist.
- Add a `## Source of truth` footer linking to pkg.go.dev and `docs/ARCHITECTURE.md`.

- [ ] **Step 8: Build incremental — should still warn but not error**

```bash
cd docs-site && npm run build 2>&1 | tail -20
```

Expected: build exits 0. Warnings for not-yet-created pages remain.

- [ ] **Step 9: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/index.mdx \
        docs-site/src/content/docs/getting-started/installation.mdx \
        docs-site/src/content/docs/getting-started/quickstart.mdx \
        docs-site/src/content/docs/guides/market-discovery.mdx \
        docs-site/src/content/docs/concepts/polymarket-api.mdx \
        docs-site/src/content/docs/reference/cli.mdx \
        docs-site/src/content/docs/reference/sdk.mdx
git commit -m "$(cat <<'EOF'
docs(site): refresh 6 existing pages for the overhaul layout

Lead the homepage and quickstart with deposit-wallet. Convert
reference/cli into an index that delegates to per-group pages. Add
pkg/gamma to the SDK reference. Add Source-of-truth footers pointing to
docs/*.md so prose drift is bounded.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Create the Deposit Wallet section (4 pages)

**Files (allowlist):**
- Create: `docs-site/src/content/docs/deposit-wallet/why.mdx`
- Create: `docs-site/src/content/docs/deposit-wallet/onboard.mdx`
- Create: `docs-site/src/content/docs/deposit-wallet/flow.mdx`
- Create: `docs-site/src/content/docs/deposit-wallet/troubleshooting.mdx`

**Why one task:** Cohesive section. Reads as a single narrative; one commit captures the headline ship.

- [ ] **Step 1: Verify the directory does not yet exist**

```bash
test ! -d docs-site/src/content/docs/deposit-wallet && echo "OK"
mkdir -p docs-site/src/content/docs/deposit-wallet
```

- [ ] **Step 2: Create `deposit-wallet/why.mdx` — full prose, zero prior knowledge**

Required structure (write fully fleshed prose, not a stub):

```markdown
---
title: Why deposit wallets matter
description: The May 14 2026 Polymarket migration explained — why every new account needs a deposit wallet and how polygolem handles it.
---

If you opened a Polymarket account **before** May 14 2026, you can keep
trading with your EOA-signed orders. If you opened — or are opening — an
account after that date, **you must use a deposit wallet** to sign orders.
This page explains why, and what changes for you.

## TL;DR

- Polymarket requires **deposit wallets** for new accounts after May 14 2026.
- A deposit wallet is a smart contract owned by your EOA that signs orders
  via **POLY_1271** (Polymarket's variant of EIP-1271 / ERC-7739).
- `polygolem deposit-wallet onboard` derives, deploys, approves, and funds
  the wallet in one command.
- Your existing EOA is still the **owner** — you keep custody.

## What is a deposit wallet?

A deposit wallet is a smart contract that:

1. Is **deterministically derived** from your EOA address (CREATE2).
2. Signs Polymarket orders using **POLY_1271** instead of an ECDSA EOA.
3. Holds your USDC.e balance for trading.
4. Is **owned by your EOA** — only your private key can authorize batched
   actions through the Polymarket builder relayer.

The address is derived before deployment, so you can fund it before the
contract exists on-chain. `polygolem deposit-wallet derive` shows you the
address from your EOA alone — no API calls, no on-chain transactions.

## Why did Polymarket make this change?

Two practical reasons:

1. **Gas-paid order flow.** The deposit wallet pattern lets the Polymarket
   builder relayer batch and pay gas for trader actions. Traders sign
   off-chain with their EOA; the relayer brokers the on-chain effect.
2. **Account abstraction-style UX.** A deposit wallet behaves like a
   stable trading account, decoupled from the user's hot key. The EOA can
   rotate without affecting the trading address.

## What changes for me as a user?

| Before (legacy EOA accounts) | After (deposit-wallet accounts) |
|---|---|
| Sign each order with EOA private key | Sign each order via POLY_1271 |
| Deposit USDC.e to your EOA | Deposit USDC.e to the deposit wallet |
| `--signature-type eoa` on `clob create-order` | `--signature-type deposit` |
| No on-chain deploy | One-time deploy + approve at onboarding |

## What does polygolem do for me?

`polygolem deposit-wallet onboard --fund-amount 50` runs the entire
onboarding sequence in one command. See
[One-command onboarding](/deposit-wallet/onboard) for the headline path
and [Step-by-step flow](/deposit-wallet/flow) for what each phase does.

## What if it fails partway through?

The composite is idempotent. Each phase (derive → deploy → approve → fund)
checks current state and skips if already done. Re-running is safe. See
[Troubleshooting](/deposit-wallet/troubleshooting).

## Source of truth

- [`docs/DEPOSIT-WALLET-MIGRATION.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/DEPOSIT-WALLET-MIGRATION.md)
- [`docs/SAFETY.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/SAFETY.md) § Deposit Wallet Safety Rules
- `polygolem deposit-wallet --help`
```

- [ ] **Step 3: Create `deposit-wallet/onboard.mdx` — the headline page**

Required structure (full prose):

```markdown
---
title: One-command onboarding
description: A single polygolem command derives, deploys, approves, and funds your Polymarket deposit wallet.
---

`polygolem deposit-wallet onboard` is the headline command. It runs the
entire deposit-wallet onboarding sequence end-to-end: derive → deploy →
approve → fund. If any step has already happened (you re-ran it), that
step is skipped and the composite continues.

## Prerequisites

You need:

- A funded EOA (Polygon) with enough native MATIC for the relayer call
  (or builder-paid gas if the relayer covers it for your account).
- Bridged USDC.e on Polygon at the EOA. Use a bridge like Bungee or
  Polymarket's bridge to move USDC from another chain. See
  [Bridge & Funding](/guides/bridge-funding).
- Builder credentials issued by Polymarket:
  - `POLYMARKET_BUILDER_API_KEY`
  - `POLYMARKET_BUILDER_SECRET`
  - `POLYMARKET_BUILDER_PASSPHRASE`
- Your EOA private key in `POLYMARKET_PRIVATE_KEY` (hex, 0x-prefixed).

## The command

```bash
export POLYMARKET_PRIVATE_KEY="0x..."
export POLYMARKET_BUILDER_API_KEY="..."
export POLYMARKET_BUILDER_SECRET="..."
export POLYMARKET_BUILDER_PASSPHRASE="..."

polygolem deposit-wallet onboard --fund-amount 50 --json
```

`--fund-amount` is in **USDC.e units (decimal)**. `50` means 50 USDC.e.

## Expected output

```json
{
  "deposit_wallet": "0xabc...def",
  "owner": "0x123...456",
  "deployed": true,
  "approved": {
    "spender": "0x...",
    "amount": "115792089237316195423570985008687907853269984665640564039457584007913129639935"
  },
  "funded": {
    "amount_pusd": "50",
    "tx_hash": "0x...",
    "block_number": 60000000
  },
  "ready": true
}
```

`ready: true` means the wallet can sign POLY_1271 orders and trade.

## Optional flags

| Flag | Default | Effect |
|---|---|---|
| `--fund-amount <N>` | (required) | USDC.e to transfer from EOA into the deposit wallet. |
| `--no-fund` | off | Skip the funding phase (deploy + approve only). |
| `--dry-run` | off | Show what would happen without submitting any transaction. |
| `--json` | off | Emit structured JSON. |

## What if I just want to derive the address?

```bash
polygolem deposit-wallet derive --json
```

This is read-only and emits no transactions. Useful for pre-funding by
sending USDC.e directly to the derived address before deploying.

## After onboarding

Use `--signature-type deposit` on every `clob create-order` /
`clob market-order` call:

```bash
polygolem clob create-order \
  --token-id "123..." \
  --side BUY --price 0.51 --size 10 \
  --signature-type deposit
```

## Source of truth

- [`docs/DEPOSIT-WALLET-MIGRATION.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/DEPOSIT-WALLET-MIGRATION.md)
- `polygolem deposit-wallet onboard --help`
```

- [ ] **Step 4: Create `deposit-wallet/flow.mdx` — step-by-step diagram**

Required structure (full prose, with a diagram):

```markdown
---
title: Step-by-step deposit-wallet flow
description: What each phase of polygolem deposit-wallet onboard does, and why.
---

`polygolem deposit-wallet onboard` is a composite of four phases. This
page shows what each does so you understand exactly what the one-command
path is doing on your behalf.

## The flow at a glance

```text
        EOA (POLYMARKET_PRIVATE_KEY)
              |
              | derives address (CREATE2, off-chain)
              v
     ┌──────────────────────────┐
     │  Deposit Wallet (0xabc..) │
     └──────────────────────────┘
       ^         ^         ^
       |         |         |
   deploy()   approve()   fund()
   (relayer)  (relayer)   (RPC ERC-20 transfer)
```

Each phase is independently runnable as its own subcommand:

| Phase | Subcommand | What it does | On-chain? |
|---|---|---|---|
| 1. Derive | `deposit-wallet derive` | Compute the deposit-wallet address (CREATE2). | No |
| 2. Deploy | `deposit-wallet deploy` | Submit the deployment calldata batch via the builder relayer. | Yes |
| 3. Approve | `deposit-wallet approve --submit` | Approve Polymarket spender for max USDC.e. | Yes |
| 4. Fund | `deposit-wallet fund --amount N` | Transfer USDC.e from EOA → deposit wallet via direct RPC. | Yes |

## Phase 1 — Derive

```bash
polygolem deposit-wallet derive --json
```

Read-only. Computes the deterministic address from your EOA.

```json
{ "deposit_wallet": "0xabc...def", "owner": "0x123...456" }
```

## Phase 2 — Deploy

```bash
polygolem deposit-wallet deploy --json
```

Submits a builder-relayer batch to deploy the wallet contract. Requires
builder credentials. Relayer pays gas. Polls until the batch lands.

## Phase 3 — Approve

```bash
polygolem deposit-wallet approve --submit --json
```

Approves the Polymarket spender for unlimited USDC.e from the deposit
wallet. Without `--submit`, prints the calldata and exits. With `--submit`,
broadcasts via the relayer.

## Phase 4 — Fund

```bash
polygolem deposit-wallet fund --amount 50 --json
```

Transfers 50 USDC.e from the EOA to the deposit-wallet address using a
direct ERC-20 `transfer()` over RPC. The EOA pays gas in MATIC.

## Status check

At any point:

```bash
polygolem deposit-wallet status --json
```

Returns the current phase per address, including what is and isn't done.

## Source of truth

- [`docs/DEPOSIT-WALLET-MIGRATION.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/DEPOSIT-WALLET-MIGRATION.md)
- `polygolem deposit-wallet <subcmd> --help`
```

- [ ] **Step 5: Create `deposit-wallet/troubleshooting.mdx`**

Required structure (table-driven; full prose for each row):

```markdown
---
title: Deposit-wallet troubleshooting
description: Failure modes for polygolem deposit-wallet, what they mean, and how to recover.
---

The onboarding composite is idempotent — re-running is safe. This page
indexes the failure modes you may hit.

## Common errors

### "missing builder credentials"

```text
error: missing POLYMARKET_BUILDER_API_KEY (or BUILDER_SECRET / BUILDER_PASSPHRASE)
```

You set `POLYMARKET_PRIVATE_KEY` but not the builder triple. `deploy`,
`approve --submit`, `batch`, and `onboard` all require the builder triple.
`derive` and `status` do not.

**Fix:** export all three:
```bash
export POLYMARKET_BUILDER_API_KEY="..."
export POLYMARKET_BUILDER_SECRET="..."
export POLYMARKET_BUILDER_PASSPHRASE="..."
```

### "wallet already deployed"

Not actually an error — `onboard` will skip the deploy phase and continue
to approve and fund. Confirm with:
```bash
polygolem deposit-wallet status --json
```

### "fund: insufficient USDC.e balance on EOA"

The EOA does not have enough USDC.e for the requested `--fund-amount`.

**Fix:** bridge more USDC.e to the EOA. See
[Bridge & Funding](/guides/bridge-funding).

### "fund: insufficient native MATIC for gas"

The EOA does not have enough MATIC to pay for the ERC-20 transfer.

**Fix:** transfer a small amount of MATIC (0.01 is plenty) to the EOA.

### "relayer batch timed out"

The builder relayer accepted the batch but the wallet did not deploy
within the polling window. Most often a transient RPC issue.

**Fix:** re-run `polygolem deposit-wallet deploy`. Idempotent — it
detects existing deployment and skips.

### "POLY_1271 signature rejected"

After onboarding, an order signed with `--signature-type deposit` was
rejected. Check `polygolem deposit-wallet status` for `deployed: true`
and `approved: true`.

## Recovery commands

| Situation | Command |
|---|---|
| What state am I in? | `polygolem deposit-wallet status --json` |
| Re-run end-to-end | `polygolem deposit-wallet onboard --fund-amount N` |
| Just deploy (already derived) | `polygolem deposit-wallet deploy` |
| Just approve | `polygolem deposit-wallet approve --submit` |
| Just fund | `polygolem deposit-wallet fund --amount N` |

## Source of truth

- [`docs/DEPOSIT-WALLET-MIGRATION.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/DEPOSIT-WALLET-MIGRATION.md)
- [`docs/SAFETY.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/SAFETY.md) § Deposit Wallet Safety Rules
- `polygolem deposit-wallet --help`
```

- [ ] **Step 6: Build and verify the four pages render**

```bash
cd docs-site && npm run build 2>&1 | grep -E "deposit-wallet|warn|error" | head -20
```

Expected: no errors. The four deposit-wallet warnings from Task 1 are now resolved.

- [ ] **Step 7: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/deposit-wallet/why.mdx \
        docs-site/src/content/docs/deposit-wallet/onboard.mdx \
        docs-site/src/content/docs/deposit-wallet/flow.mdx \
        docs-site/src/content/docs/deposit-wallet/troubleshooting.mdx
git commit -m "$(cat <<'EOF'
docs(site): add Deposit Wallet (May 2026) section — 4 pages

Why this matters, one-command onboarding, step-by-step flow, and
troubleshooting. Headline section of the overhaul: turns the May 14 2026
migration from a buried CLI subcommand into a four-page narrative a
new user can act on in under 10 minutes.

Source of truth links point to docs/DEPOSIT-WALLET-MIGRATION.md and
docs/SAFETY.md.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4a: Create new Guides — orderbook-data, paper-trading, placing-orders

**Files (allowlist):**
- Create: `docs-site/src/content/docs/guides/orderbook-data.mdx`
- Create: `docs-site/src/content/docs/guides/paper-trading.mdx`
- Create: `docs-site/src/content/docs/guides/placing-orders.mdx`

For each page write a full guide following the outline below. Each guide must include at least one runnable command, expected JSON output (truncated), at least one table, and a Source-of-truth footer.

- [ ] **Step 1: `guides/orderbook-data.mdx`**

Outline:
- Frontmatter: title "Orderbook Data", description "Read CLOB book depth, midpoints, spreads, and tick sizes — no auth required."
- Intro paragraph: orderbook is the L0 read surface; no credentials needed.
- Section "Get a full book": `polygolem orderbook get --token-id "123..."` with truncated JSON.
- Section "Best prices": `polygolem orderbook price --token-id "123..."` with output.
- Section "Midpoint and spread": commands for `midpoint` and `spread`.
- Section "Tradability fields": `tick-size`, `fee-rate`, with output. Note that `minimum_order_size` is enforced server-side.
- Section "From Go": `pkg/bookreader` example mirroring the SDK reference.
- Source-of-truth footer: `docs/COMMANDS.md` § orderbook, `polygolem orderbook --help`.

- [ ] **Step 2: `guides/paper-trading.mdx`**

Outline:
- Frontmatter: title "Paper Trading", description "Simulate orders against live market data without touching authenticated endpoints."
- Intro: paper mode is local-only; cite `internal/paper`.
- Section "Open a paper position": example `polygolem paper open --token-id "..." --side BUY --size 10 --price 0.51` with output.
- Section "List positions and PnL": `polygolem paper positions`, `polygolem paper pnl`.
- Section "Close": `polygolem paper close --id "..."`.
- Section "State location": where paper state is persisted on disk (cite `internal/paper`).
- Section "Why paper mode never calls live endpoints" — short paragraph explaining the safety boundary in `internal/modes`.
- Source-of-truth footer: `docs/SAFETY.md`, `docs/COMMANDS.md` § paper, `polygolem paper --help`.

- [ ] **Step 3: `guides/placing-orders.mdx`**

Outline:
- Frontmatter: title "Placing Real Orders", description "Limit and market orders against Polymarket CLOB with deposit-wallet POLY_1271 signing."
- Prereqs: deposit wallet onboarded (link to `/deposit-wallet/onboard`), USDC.e in deposit wallet, `--signature-type deposit`.
- Section "Limit order": `polygolem clob create-order --token-id "..." --side BUY --price 0.51 --size 10 --signature-type deposit` with output.
- Section "Market order": `polygolem clob market-order ...` example.
- Section "Cancel and query": `polygolem clob cancel --order-id "..."`, `polygolem clob orders`.
- Section "Builder attribution": one paragraph linking to `/concepts/builder-attribution`.
- Section "Live mode gates": one paragraph linking to `/concepts/safety` — preflight, risk, funding gates must pass.
- Source-of-truth footer: `docs/COMMANDS.md` § clob, `docs/SAFETY.md`, `polygolem clob --help`.

- [ ] **Step 4: Build incremental**

```bash
cd docs-site && npm run build 2>&1 | grep -E "guides/(orderbook-data|paper-trading|placing-orders)|error" | head
```

Expected: zero hits (warnings resolved, no new errors).

- [ ] **Step 5: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/guides/orderbook-data.mdx \
        docs-site/src/content/docs/guides/paper-trading.mdx \
        docs-site/src/content/docs/guides/placing-orders.mdx
git commit -m "$(cat <<'EOF'
docs(site): add Guides — orderbook-data, paper-trading, placing-orders

Three new guides covering the L0 read surface, the paper-mode local
simulation surface, and the deposit-wallet-signed live order surface.
Each links back to docs/COMMANDS.md and docs/SAFETY.md as canonical.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4b: Create new Guides — bridge-funding, go-bot-integration

**Files (allowlist):**
- Create: `docs-site/src/content/docs/guides/bridge-funding.mdx`
- Create: `docs-site/src/content/docs/guides/go-bot-integration.mdx`

- [ ] **Step 1: `guides/bridge-funding.mdx`**

Outline:
- Frontmatter: title "Bridge & Funding", description "Move USDC.e onto Polygon and into your Polymarket deposit wallet."
- Intro: clarify Polygon vs other chains; explain "bridging" in one sentence.
- Section "List supported assets": `polygolem bridge assets --json` with output.
- Section "Get a deposit address": `polygolem bridge deposit-address --address "0x..."` with output.
- Section "Get a quote": `polygolem bridge quote --from-chain ethereum --asset USDC --amount 100 --json` with output.
- Section "Track a deposit": `polygolem bridge status --deposit-address "0x..."`.
- Section "From Go": `pkg/bridge` snippet.
- Section "After bridging": link to `/deposit-wallet/onboard` for fund step.
- Source-of-truth footer: `docs/COMMANDS.md` § bridge, `polygolem bridge --help`.

- [ ] **Step 2: `guides/go-bot-integration.mdx`**

Outline:
- Frontmatter: title "Go-Bot Integration", description "Embed polygolem's public SDK in a downstream Go trading bot."
- Intro: polygolem ships a stable `pkg/` surface; downstream consumers (e.g., go-bot) import directly.
- Section "What's in `pkg/`": list with one-line descriptions, link to `/reference/sdk`.
- Section "Minimal embedding example": full Go `main.go` that resolves a market, reads a book, returns prices.
- Section "Authenticated flows": brief — for live trading, embed and configure auth via env vars; refer to `/concepts/signature-types`.
- Section "Versioning": one paragraph — polygolem follows semver on `pkg/`; pin in your `go.mod`.
- Source-of-truth footer: `docs/ARCHITECTURE.md`, pkg.go.dev URL.

- [ ] **Step 3: Build**

```bash
cd docs-site && npm run build 2>&1 | grep -E "guides/(bridge-funding|go-bot-integration)|error" | head
```

Expected: zero hits.

- [ ] **Step 4: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/guides/bridge-funding.mdx \
        docs-site/src/content/docs/guides/go-bot-integration.mdx
git commit -m "$(cat <<'EOF'
docs(site): add Guides — bridge-funding, go-bot-integration

Bridge guide walks list-assets → quote → deposit-address → status using
polygolem bridge subcommands and pkg/bridge. Go-bot guide shows the
embedding pattern downstream consumers (incl. go-bot) use against pkg/*.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5a: Create Concepts — markets-events-tokens, modes, signature-types, builder-attribution

**Files (allowlist):**
- Create: `docs-site/src/content/docs/concepts/markets-events-tokens.mdx`
- Create: `docs-site/src/content/docs/concepts/modes.mdx`
- Create: `docs-site/src/content/docs/concepts/signature-types.mdx`
- Create: `docs-site/src/content/docs/concepts/builder-attribution.mdx`

Each page must include a table or diagram per the conventions. `concepts/modes.mdx` is the **exemplar full-prose page** — write it first and use it as the style anchor for the other concepts.

- [ ] **Step 1: `concepts/modes.mdx` — EXEMPLAR — write fully**

Use exactly this content:

```markdown
---
title: Modes (read-only / paper / live)
description: Polygolem's three execution modes, what each can and cannot do, and how the mode is chosen.
---

Polygolem operates in one of three **modes** at any time. The mode is
chosen at startup and gates every subsequent operation. This page is the
canonical user-facing description; the implementation lives in
`internal/modes`.

## The three modes

| Mode | Default? | Can read public data? | Can use paper state? | Can sign / submit live orders? |
|---|---|---|---|---|
| **Read-only** | yes | yes | no | no |
| **Paper** | no | yes | yes | no |
| **Live** | no | yes | no | yes (gated) |

## Read-only

The default. Selected when no mode flag is passed and no live-only command
is invoked. Touches only public Polymarket endpoints (Gamma, CLOB read,
Data API, Bridge read, market WebSocket channel). Forbids any signing
operation. This is what runs when you do:

```bash
polygolem discover search --query "btc"
polygolem orderbook get --token-id "..."
```

Read-only mode requires no credentials. Setting
`POLYMARKET_PRIVATE_KEY` does not, by itself, change the mode.

## Paper

A local simulation mode. Combines read-only reference data with
`internal/paper` state so simulated trades behave like live trades except
no authenticated endpoint is ever called. State is persisted locally
between invocations. Selected by paper-prefixed commands:

```bash
polygolem paper open --token-id "..." --side BUY --size 10 --price 0.51
polygolem paper positions
polygolem paper close --id "..."
```

Paper mode is **local-only** by construction. The paper executor in
`internal/execution` cannot reach `internal/clob`'s authenticated write
methods — they sit behind a mode gate.

## Live

The only mode that can submit real orders or move real funds. Selected
either by an explicitly live command (`polygolem clob create-order`,
`polygolem deposit-wallet deploy`) or by `polygolem live *`. Each live
operation must pass:

1. **Preflight** — connectivity and config validity (`polygolem preflight`).
2. **Risk** — per-trade cap, daily loss limit, circuit breaker (`internal/risk`).
3. **Funding** — adequate USDC.e and gas (where applicable).
4. **Signature** — correct `--signature-type` for your account vintage.

Live mode is never the default. Bare `polygolem` invocations cannot end
up in live mode without an explicit live-mutating command.

## Mode selection rules

```text
       startup
          |
   parse flags + env
          |
   v ----------------- v
   any live-only cmd?  -- yes --> live (subject to gates)
          |
          | no
          v
   any paper-prefixed cmd?  -- yes --> paper
          |
          | no
          v
        read-only  (default)
```

## Why this is a hard boundary

Mode lives in `internal/modes` and is consumed by every protocol client
before it accepts a write. There is no "soft" mode flag. A handler that
attempts a write while paper-mode is active fails closed at the package
boundary, not at the CLI layer.

## Source of truth

- [`docs/SAFETY.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/SAFETY.md)
- [`docs/ARCHITECTURE.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ARCHITECTURE.md) § Mode system
- `internal/modes/`
```

- [ ] **Step 2: `concepts/markets-events-tokens.mdx`**

Outline (mirror modes.mdx structure):
- Frontmatter.
- Intro: define the three primitives.
- Table comparing Market, Event, Token: what it is, IDs used, where to fetch from.
- Worked example: an event resolves to N markets, each market resolves to 2 tokens (yes/no or up/down).
- Section "Identifiers you'll see": `condition_id`, `market_id`, `token_id`, `slug`, `event_id` — when each is used.
- Section "From the SDK": one paragraph linking to `pkg/marketresolver` and `pkg/gamma`.
- Source-of-truth footer: `docs/ARCHITECTURE.md`, `internal/polytypes`, `pkg/marketresolver` godoc.

- [ ] **Step 3: `concepts/signature-types.mdx`**

Outline:
- Frontmatter.
- Intro: orders need a signature; the `--signature-type` flag picks which path.
- Table from Track 1's ARCHITECTURE.md (eoa / proxy / gnosis-safe / deposit), one column "Use when".
- Note "After May 14 2026, new accounts must use `deposit`."
- Cross-link to `/concepts/poly-1271` (the technical signing detail) and `/deposit-wallet/why`.
- Source-of-truth footer: `docs/ARCHITECTURE.md` § Signature types, `internal/auth`.

- [ ] **Step 4: `concepts/builder-attribution.mdx`**

Outline:
- Frontmatter.
- Intro: builder attribution is how Polymarket credits the entity that submitted an order; orthogonal to signature type.
- Section "What is a builder?" — one paragraph: a registered Polymarket builder account with API credentials.
- Section "How polygolem attaches builder attribution": cite `internal/auth` builder fields; mention env vars.
- Section "What it does NOT do": does not relax safety gates, does not grant trading privileges (cite `docs/SAFETY.md` § Deposit Wallet Safety Rules item 7).
- Source-of-truth footer: `docs/SAFETY.md`, `docs/ARCHITECTURE.md`, `internal/auth`.

- [ ] **Step 5: Build**

```bash
cd docs-site && npm run build 2>&1 | grep -E "concepts/(markets-events-tokens|modes|signature-types|builder-attribution)|error" | head
```

Expected: zero hits.

- [ ] **Step 6: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/concepts/markets-events-tokens.mdx \
        docs-site/src/content/docs/concepts/modes.mdx \
        docs-site/src/content/docs/concepts/signature-types.mdx \
        docs-site/src/content/docs/concepts/builder-attribution.mdx
git commit -m "$(cat <<'EOF'
docs(site): add Concepts — markets-events-tokens, modes, signatures, builder

Four conceptual primer pages. concepts/modes is the exemplar full-prose
page; the other three follow its structure. Cross-links into
concepts/poly-1271 and deposit-wallet/why for downstream depth.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5b: Create Concepts — poly-1271, safety, architecture

**Files (allowlist):**
- Create: `docs-site/src/content/docs/concepts/poly-1271.mdx`
- Create: `docs-site/src/content/docs/concepts/safety.mdx`
- Create: `docs-site/src/content/docs/concepts/architecture.mdx`

`safety.mdx` and `architecture.mdx` are explicitly **link-out** pages: short prose plus a strong pointer to `docs/SAFETY.md` / `docs/ARCHITECTURE.md` (Track 1 outputs). Do not duplicate content; summarize and link.

- [ ] **Step 1: `concepts/poly-1271.mdx`**

Outline:
- Frontmatter.
- Intro: POLY_1271 is Polymarket's variant of EIP-1271 + ERC-7739 used by deposit wallets to sign orders.
- Section "What is EIP-1271?" — one paragraph, the smart-contract `isValidSignature` check.
- Section "What does POLY_1271 add?" — Polymarket's domain-separated typed data + ERC-7739 wrapping for replay safety.
- Section "When is POLY_1271 used?" — `--signature-type deposit` on orders, batch operations through the builder relayer.
- Section "Where it lives in polygolem" — `internal/auth` (signing primitives), `internal/clob` (typed-data construction).
- Diagram (text): order → typed-data → POLY_1271 hash → wallet `isValidSignature` → CLOB accepts.
- Source-of-truth footer: `internal/auth`, `internal/clob`, `docs/ARCHITECTURE.md`.

- [ ] **Step 2: `concepts/safety.mdx` — link-out page**

Use exactly this content shape:

```markdown
---
title: Safety Model
description: Polygolem's safety boundaries — read-only by default, paper-mode local-only, live-mode gated.
---

The full safety model is documented in
[`docs/SAFETY.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/SAFETY.md).
This page is a quick orientation; treat the linked doc as canonical.

## The four guarantees

1. **Read-only is the default mode** and is exercised by every public command.
2. **Paper mode never calls authenticated endpoints.**
3. **Live commands require explicit signature-type, gates passing, and
   builder credentials where applicable.**
4. **Builder credentials and private keys are redacted** by `internal/config`
   on every load.

## Deposit-wallet specific rules

The May 14 2026 deposit-wallet migration adds a family of rules covered
in `docs/SAFETY.md` § Deposit Wallet Safety Rules. Highlights:

- Read-only deposit-wallet commands (`derive`, `status`, `nonce`) stay
  read-only.
- Funding moves real money — no default `--amount`.
- `onboard` is the only multi-step composite; failure leaves a
  recoverable state visible to `deposit-wallet status`.

## Source of truth

- [`docs/SAFETY.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/SAFETY.md)
- [`docs/DEPOSIT-WALLET-MIGRATION.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/DEPOSIT-WALLET-MIGRATION.md)
```

- [ ] **Step 3: `concepts/architecture.mdx` — link-out page**

Use exactly this content shape:

```markdown
---
title: Architecture
description: Polygolem's package layout, dependency direction, and mode/signing boundaries.
---

The full architecture is documented in
[`docs/ARCHITECTURE.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ARCHITECTURE.md).
This page is a quick orientation; treat the linked doc as canonical.

## Layered surface

```text
cmd/polygolem
       |
internal/cli
       |
internal/{config, modes, preflight, output, errors}
       |
internal/{gamma, clob, dataapi, stream, relayer, rpc}   ← protocol clients
       |
internal/{auth, transport, polytypes}                   ← cross-cutting
       |
internal/{wallet, orders, execution, risk, paper, marketdiscovery}
       |
pkg/{bookreader, bridge, gamma, marketresolver, pagination}   ← public SDK
```

## Public SDK

Five public packages in `pkg/`: `bookreader`, `bridge`, `gamma`,
`marketresolver`, `pagination`. See [Go SDK](/reference/sdk) for the
in-site reference and pkg.go.dev for full godoc.

## Cobra handlers stay thin

Command handlers parse flags, call package APIs, and render output via
`internal/output`. They contain no protocol or trading business logic.

## Source of truth

- [`docs/ARCHITECTURE.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ARCHITECTURE.md)
- pkg.go.dev (published from `main`)
```

- [ ] **Step 4: Build**

```bash
cd docs-site && npm run build 2>&1 | grep -E "concepts/(poly-1271|safety|architecture)|error" | head
```

Expected: zero hits.

- [ ] **Step 5: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/concepts/poly-1271.mdx \
        docs-site/src/content/docs/concepts/safety.mdx \
        docs-site/src/content/docs/concepts/architecture.mdx
git commit -m "$(cat <<'EOF'
docs(site): add Concepts — poly-1271, safety, architecture

POLY_1271 explains the deposit-wallet signing path. safety and
architecture are deliberate link-out pages: short orientation + canonical
pointer to docs/SAFETY.md and docs/ARCHITECTURE.md (Track 1 output) so
prose drift is bounded to one place.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Create reference/cli per-group pages (6 pages)

**Files (allowlist):**
- Create: `docs-site/src/content/docs/reference/cli/discover.mdx`
- Create: `docs-site/src/content/docs/reference/cli/orderbook.mdx`
- Create: `docs-site/src/content/docs/reference/cli/deposit-wallet.mdx`
- Create: `docs-site/src/content/docs/reference/cli/clob.mdx`
- Create: `docs-site/src/content/docs/reference/cli/paper.mdx`
- Create: `docs-site/src/content/docs/reference/cli/bridge.mdx`

**Approach:** Each per-group page mirrors the structure used by `docs/COMMANDS.md` (Track 1). Pull `Usage:`, `Flags:`, and one example per subcommand from `polygolem <group> --help` and `polygolem <group> <sub> --help`. **Do not invent flags.** If you need to refer to a runtime help capture file, use `/tmp/polygolem-cmds.txt` from Track 1 if still present, else regenerate it locally.

Common page structure (apply to all six):

```markdown
---
title: <group> command
description: <one-line description of the group>
---

<one-paragraph intro: what this group does, what mode it requires>

## Subcommands

For each subcommand:

### <group> <subcommand>

<one-line description>

**Usage:**

```
polygolem <group> <subcommand> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|

**Example:**

```bash
polygolem <group> <subcommand> ...
```

```json
{ "...": "..." }
```

## Source of truth

- [`docs/COMMANDS.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/COMMANDS.md) § <group>
- `polygolem <group> --help`
```

- [ ] **Step 1: Capture help output for the six groups**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
mkdir -p /tmp/polygolem-help-task6
for g in discover orderbook deposit-wallet clob paper bridge; do
  ./polygolem "$g" --help > "/tmp/polygolem-help-task6/${g}.txt" 2>&1
  for sub in $(./polygolem "$g" --help 2>&1 | awk '/Available Commands:/,/^$/' | tail -n +2 | awk '{print $1}' | grep -v '^$'); do
    echo "===== $g $sub =====" >> "/tmp/polygolem-help-task6/${g}.txt"
    ./polygolem "$g" "$sub" --help >> "/tmp/polygolem-help-task6/${g}.txt" 2>&1
  done
done
ls -la /tmp/polygolem-help-task6/
```

Expected: six non-empty files. If any is empty, the group is missing from the binary — log and adjust the page to note that.

- [ ] **Step 2: Create each per-group page**

Walk the six groups in order: `discover`, `orderbook`, `deposit-wallet`, `clob`, `paper`, `bridge`. For each:

1. Open `/tmp/polygolem-help-task6/<group>.txt`.
2. Use the top of the file (the group's own `--help`) for the page intro, listing what subcommands exist.
3. For each `===== <group> <sub> =====` block, emit a `### <group> <sub>` section using the structure above. Pull every flag listed in the `Flags:` block into the table; do not omit and do not invent.
4. The `Example:` block must be a working command using known example values from existing docs (e.g., `--token-id "123..."`). The `json` block under it must show realistic truncated output. If the implementer cannot derive a realistic output without running the command, they may print `<output omitted — run locally to inspect>` for read-only commands, but **only** as a last resort.

Specific notes per group:

- **`discover`**: subcommands `search`, `market`, `enrich`. The existing `reference/cli.mdx` (pre-refresh) had reasonable example outputs — reuse them.
- **`orderbook`**: subcommands `get`, `price`, `midpoint`, `spread`, `tick-size`, `fee-rate`, possibly `last-trade`. Use the same `--token-id "123..."` pattern.
- **`deposit-wallet`**: subcommands `derive`, `status`, `deploy`, `batch`, `approve`, `fund`, `nonce`, `onboard`. Cross-link to `/deposit-wallet/onboard` and `/deposit-wallet/flow` in the page intro. Note: `batch` requires `--calls-json`.
- **`clob`**: subcommands likely include `create-order`, `market-order`, `cancel`, `orders`, `trades`, `tick-size`, `fee-rate` (read mirrors). For mutating ones note `--signature-type deposit` requirement post-May 2026.
- **`paper`**: subcommands likely include `open`, `close`, `positions`, `pnl`. Mark every example clearly as local-only.
- **`bridge`**: subcommands `assets`, `deposit-address`, `quote`, `status`.

- [ ] **Step 3: Build**

```bash
cd docs-site && npm run build 2>&1 | grep -E "reference/cli/(discover|orderbook|deposit-wallet|clob|paper|bridge)|error" | head
```

Expected: zero hits.

- [ ] **Step 4: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/reference/cli/discover.mdx \
        docs-site/src/content/docs/reference/cli/orderbook.mdx \
        docs-site/src/content/docs/reference/cli/deposit-wallet.mdx \
        docs-site/src/content/docs/reference/cli/clob.mdx \
        docs-site/src/content/docs/reference/cli/paper.mdx \
        docs-site/src/content/docs/reference/cli/bridge.mdx
git commit -m "$(cat <<'EOF'
docs(site): add per-group CLI reference pages — 6 pages

discover, orderbook, deposit-wallet, clob, paper, bridge. Each lists
every subcommand pulled from polygolem <group> --help with one runnable
example per subcommand. Source of truth links back to docs/COMMANDS.md.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: Create reference link pages — json-contract, error-codes, env-vars

**Files (allowlist):**
- Create: `docs-site/src/content/docs/reference/json-contract.mdx`
- Create: `docs-site/src/content/docs/reference/error-codes.mdx`
- Create: `docs-site/src/content/docs/reference/env-vars.mdx`

These are **short link pages**. The canonical content lives in `docs/JSON-CONTRACT.md` (Track 3 output) and `docs/COMMANDS.md` § Environment Variables (Track 1 output). Error codes are documented in Track 3 alongside the JSON contract.

If `docs/JSON-CONTRACT.md` does not exist when this task runs, **stop and report** — Track 3 has not landed.

- [ ] **Step 1: `reference/json-contract.mdx`**

```markdown
---
title: JSON Contract
description: Polygolem's stable JSON output envelope for --json mode.
---

Every `polygolem` command run with `--json` emits a stable envelope. The
canonical contract lives in
[`docs/JSON-CONTRACT.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/JSON-CONTRACT.md).
This page is a quick orientation.

## Envelope shape

```json
{
  "ok": true,
  "data": { "...": "..." },
  "error": null,
  "meta": { "command": "...", "version": "..." }
}
```

- `ok` is `true` for success, `false` for any error.
- `data` is command-specific and omitted on error.
- `error` is structured (see [Error & Exit Codes](/reference/error-codes))
  on failure and `null` on success.
- `meta` always carries the command path and polygolem version.

## Why a stable contract?

Downstream agents (e.g., the Claude skill — see
[Using polygolem from Claude](/agents/claude-skill)) rely on the envelope
shape to parse output. The contract is versioned; breaking changes
require a major version bump.

## Source of truth

- [`docs/JSON-CONTRACT.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/JSON-CONTRACT.md)
```

- [ ] **Step 2: `reference/error-codes.mdx`**

```markdown
---
title: Error & Exit Codes
description: Polygolem's structured error format and process exit codes.
---

Polygolem emits structured errors in JSON mode and uses non-zero exit
codes for every error class. The canonical reference lives in
[`docs/JSON-CONTRACT.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/JSON-CONTRACT.md)
§ Errors.

## Error envelope

```json
{
  "ok": false,
  "data": null,
  "error": {
    "code": "E_AUTH_MISSING",
    "message": "missing POLYMARKET_BUILDER_API_KEY",
    "hint": "export POLYMARKET_BUILDER_API_KEY=..."
  },
  "meta": { "command": "...", "version": "..." }
}
```

## Exit-code classes

| Range | Meaning |
|---|---|
| `0` | Success. |
| `1` | Generic error. |
| `2` | Usage error (bad flags). |
| `3` | Auth / credentials missing or invalid. |
| `4` | Network / upstream error. |
| `5` | Safety-gate failure. |

## Source of truth

- [`docs/JSON-CONTRACT.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/JSON-CONTRACT.md) § Errors
- `internal/errors/`
```

The exact error codes and exit-code values must match `internal/errors/`. If the Track 3 spec or `internal/errors/` lists a different set, replace the table with whatever Track 3 documents — Track 3 is canonical.

- [ ] **Step 3: `reference/env-vars.mdx`**

```markdown
---
title: Environment Variables
description: Every environment variable polygolem reads at startup.
---

Polygolem loads configuration via Viper. Environment variables override
config files and flags follow normal precedence rules. Canonical list:
[`docs/COMMANDS.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/COMMANDS.md)
§ Environment Variables.

## Authenticated commands

| Variable | Required for |
|---|---|
| `POLYMARKET_PRIVATE_KEY` | Any signing operation. |
| `POLYMARKET_BUILDER_API_KEY` | `deposit-wallet deploy/batch/approve --submit/onboard`. |
| `POLYMARKET_BUILDER_SECRET` | Same as above. |
| `POLYMARKET_BUILDER_PASSPHRASE` | Same as above. |

Short-form aliases: `BUILDER_API_KEY`, `BUILDER_SECRET`, `BUILDER_PASS_PHRASE`.

## Endpoint overrides

| Variable | Default |
|---|---|
| `POLYMARKET_RELAYER_URL` | `https://relayer-v2.polymarket.com` |
| `POLYMARKET_GAMMA_URL` | `https://gamma-api.polymarket.com` |
| `POLYMARKET_CLOB_URL` | `https://clob.polymarket.com` |

## Redaction

`internal/config` redacts every credential field on every load. No
command emits `POLYMARKET_BUILDER_*` or `POLYMARKET_PRIVATE_KEY` in JSON
output.

## Source of truth

- [`docs/COMMANDS.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/COMMANDS.md) § Environment Variables
- `internal/config/`
```

Verify the endpoint defaults against `internal/config/` before committing; replace the table values if they differ.

- [ ] **Step 4: Build**

```bash
cd docs-site && npm run build 2>&1 | grep -E "reference/(json-contract|error-codes|env-vars)|error" | head
```

Expected: zero hits.

- [ ] **Step 5: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/reference/json-contract.mdx \
        docs-site/src/content/docs/reference/error-codes.mdx \
        docs-site/src/content/docs/reference/env-vars.mdx
git commit -m "$(cat <<'EOF'
docs(site): add reference link pages — json-contract, error-codes, env-vars

Three short orientation pages that point at the canonical sources:
docs/JSON-CONTRACT.md (Track 3) for envelope and error-code shape, and
docs/COMMANDS.md (Track 1) for environment variables. Prose duplication
kept minimal so drift only happens in one place.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: Create agents/claude-skill page

**Files (allowlist):**
- Create: `docs-site/src/content/docs/agents/claude-skill.mdx`

If `SKILL.md` does not exist at the repo root when this task runs, **stop and report** — Track 3 has not landed.

- [ ] **Step 1: Create the directory**

```bash
mkdir -p docs-site/src/content/docs/agents
```

- [ ] **Step 2: Write `agents/claude-skill.mdx`**

Use this content shape:

```markdown
---
title: Using polygolem from Claude
description: How an agent (Claude Code, custom Claude harnesses) consumes polygolem via SKILL.md.
---

Polygolem ships a Claude-compatible skill description at
[`SKILL.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/SKILL.md).
This page summarizes how an agent uses it.

## What the skill provides

- A natural-language summary of every command group.
- The stable JSON envelope contract — agents parse it, not human prose.
- Example invocations for the most common workflows (discover, orderbook,
  paper, deposit-wallet onboard).

## Loading the skill

Drop `SKILL.md` next to your other skills (e.g. under
`~/.claude/skills/` or a project-level skills directory). Claude will
surface it like any other skill — by name and trigger description.

## Programmatic consumption

For agents that don't use Claude's skill loader, the canonical JSON
envelope is documented at
[`docs/JSON-CONTRACT.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/JSON-CONTRACT.md).
Always invoke polygolem with `--json` from an agent.

## Worked example

```bash
polygolem deposit-wallet status --json
```

Returns the envelope shape documented in
[JSON Contract](/reference/json-contract). The agent parses `data` for
deployment / approval / funding state and acts based on `ok`.

## Source of truth

- [`SKILL.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/SKILL.md)
- [`docs/JSON-CONTRACT.md`](https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/JSON-CONTRACT.md)
```

- [ ] **Step 3: Build**

```bash
cd docs-site && npm run build 2>&1 | grep -E "agents/claude-skill|error" | head
```

Expected: zero hits.

- [ ] **Step 4: Commit**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
git add docs-site/src/content/docs/agents/claude-skill.mdx
git commit -m "$(cat <<'EOF'
docs(site): add For Agents — Using polygolem from Claude

Single-page agent surface that points at SKILL.md (Track 3) and the JSON
envelope contract. Establishes the agent-facing entry point of the docs
site without duplicating SKILL.md content.

Part of Track 4 (Docs Site Overhaul) of the documentation overhaul.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: Final Track 4 verification gate

**Files:** none modified — read-only verification. No commit.

This task either passes (Track 4 done) or identifies regressions to loop back on. Spec-mandated check: `cd docs-site && npm run build` produces **zero warnings**.

- [ ] **Step 1: Clean build with zero warnings**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site
rm -rf dist .astro
npm run build 2>&1 | tee /tmp/polygolem-docs-build.txt | tail -40
grep -E '\bWARN\b|\bwarning\b|\berror\b' /tmp/polygolem-docs-build.txt | grep -v 'no warning'
```

Expected: the final `grep` prints **nothing**. If any `WARN`/`warning`/`error` lines remain, identify the slug and fix the corresponding page in its source task. Do not paper over.

- [ ] **Step 2: Every sidebar slug has a backing file**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
slugs=$(grep -oE 'slug: "[^"]*"' docs-site/astro.config.mjs | sed -E 's/slug: "([^"]*)"/\1/')
missing=0
for s in $slugs; do
  if [ -z "$s" ]; then
    f="docs-site/src/content/docs/index.mdx"
  else
    f="docs-site/src/content/docs/${s}.mdx"
  fi
  if [ ! -f "$f" ]; then
    echo "MISSING: $s -> $f"
    missing=$((missing + 1))
  fi
done
echo "Missing: $missing"
```

Expected: `Missing: 0`. If any `MISSING:` line prints, fix in the corresponding task.

- [ ] **Step 3: Total count of mdx files matches sidebar count**

```bash
find docs-site/src/content/docs -name '*.mdx' | wc -l
```

Expected: `33`. If lower, a page is missing; if higher, an orphaned page exists outside the sidebar (acceptable but flag in the final report).

- [ ] **Step 4: Every Source-of-truth link is a valid path**

For each linked `docs/*.md` and `SKILL.md` reference inside the docs site, confirm the target exists:

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem
grep -hoE 'docs/[A-Z][A-Z0-9_-]*\.md' docs-site/src/content/docs/**/*.mdx docs-site/src/content/docs/*.mdx 2>/dev/null | sort -u | while read -r f; do
  test -f "$f" || echo "MISSING SOURCE-OF-TRUTH: $f"
done
grep -l 'SKILL.md' docs-site/src/content/docs/**/*.mdx 2>/dev/null && (test -f SKILL.md || echo "MISSING SOURCE-OF-TRUTH: SKILL.md")
```

Expected: no `MISSING SOURCE-OF-TRUTH:` lines.

- [ ] **Step 5: Manual sidebar walk**

Open `docs-site/dist/index.html` (or run `npm run preview` and open `http://localhost:4321/`). Walk every sidebar entry top to bottom. Confirm:

- Each link resolves (no 404).
- Each page renders (no blank page or template error).
- Every page has a Source-of-truth footer where required by conventions, except the explicit exceptions (e.g., simple landing tiles).
- Every code block on the deposit-wallet pages and the homepage looks copy-paste runnable.

Log any visual issues; fix in the source task and re-run the build.

- [ ] **Step 6: No accidental edits to repo files outside the allowlist**

```bash
git status --short | grep -vE '^\?\? |docs-site/|docs/superpowers/plans/'
```

Expected: empty. The Track 4 plan permits edits only under `docs-site/` (and this plan file itself, already committed externally). If anything else shows up, **stop and report** — a previous task touched out-of-scope files.

- [ ] **Step 7: If all checks pass, mark Track 4 complete**

No file change required. Inform the user that Track 4 verification has passed and propose moving to Track 5 planning. If any check fails, return to the relevant earlier task and fix in place rather than papering over.

---

## Out of scope (re-stated)

- Hosting, deploy automation, GitHub Pages publishing config.
- Search backend changes (Pagefind default — leave it).
- Internationalization or theme/dark-mode tweaks.
- Authoring `docs/JSON-CONTRACT.md`, `SKILL.md`, or any canonical
  `docs/*.md`. Those belong to Tracks 1–3. Track 4 only links to them.
- Code changes anywhere outside `docs-site/`.
