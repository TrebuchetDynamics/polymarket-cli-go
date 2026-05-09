# Blockers

Account: EOA `0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C`
Audit: 2026-05-07

> **⚠️ CORRECTION 2026-05-08 — This document contains outdated conclusions.**
> The proxy-wallet (sigtype-1) claims below were based on a Playwright capture of the
> polymarket.com web UI. However, **polygolem only supports deposit wallets (sigtype-3)**.
> Empirical live testing with real funds (see [LIVE-TRADING-BLOCKER-REPORT.md](docs/LIVE-TRADING-BLOCKER-REPORT.md))
> proved that deposit-wallet onboarding for new users **requires a one-time browser login**
> to create the deposit-wallet-owned CLOB API key. The `/profiles` registration described
> below does NOT create a deposit-wallet API key — it only registers a proxy-wallet profile,
> which polygolem does not use.
>
> **See instead:** [ONBOARDING.md](docs/ONBOARDING.md) — the single source of truth.

## OBSOLETE 2026-05-08 — Web-UI EOA signup uses sigtype-1 proxy, not deposit wallet

> **Note:** The following section documents a Playwright capture of the polymarket.com
> web UI signup flow. It observed proxy-wallet (sigtype-1) behavior. Polygolem does not
> support proxy wallets. The deposit-wallet (sigtype-3) path, which polygolem uses,
> was empirically tested and found to require browser login for new users.

The earlier framing (deposit wallet is the only supported mode) is
incorrect for new EOA accounts. A Playwright capture of the actual
polymarket.com signup flow (`scripts/playwright-capture/eoa-signup.mjs`,
EOA `0x3075Af8096c3e5147af22Da45FE7c8496E70a306`, run 2026-05-08T18:11Z)
showed Polymarket's frontend creates a **sigtype-1 proxy wallet** for
new users — `0x31bd9F0E315586352eb9B6141cEC154C9a71549D` — and binds the
account, builder profile, and relayer API key to that proxy.

**The 400 `"maker address not allowed, please use the deposit wallet
flow"` rejection seen in the 2026-05-07 verification was for an
unregistered EOA — Polymarket's backend had no profile for it. After
running the equivalent of the web-UI signup (SIWE login →
`POST /gamma-api/profiles`), the proxy is registered and sigtype-1
orders from that proxy are accepted.** The `deposit-wallet-migration`
doc still applies, but as a *parallel* path for users who explicitly
need a deposit wallet (custody/recovery features), not a *mandatory*
one.

### Captured V2 EOA signup flow (decoded from HAR)

```
1. GET  /gamma-api.polymarket.com/nonce                     → nonce
2. eth_requestAccounts via injected EIP-1193 provider       → [EOA]
3. personal_sign over standard SIWE message
   (statement: "Welcome to Polymarket! Sign to connect.")
4. GET  /gamma-api.polymarket.com/login
        Authorization: Bearer base64(SIWE-JSON ::: signature)
5. POST /gamma-api.polymarket.com/profiles
        body: {displayUsernamePublic, name, pseudonym,
               proxyWallet: <derived>,
               users: [{address: <EOA>, provider: "metamask",
                        proxyWallet: <derived>, ...}], ...}
   → 201 {profile, proxyWallet, users[]}
6. (optional, builder fees) POST /gamma-api.polymarket.com/builder-profiles
        body: {name, address: <proxyWallet>}
   → 201 {id, builderCode: {code: 0x<bytes32>, makerFeeRateBps,
                            takerFeeRateBps, enabled}, ...}
7. POST /relayer-v2.polymarket.com/relayer/api/auth
        body: {}
   → 200 {apiKey: <UUID>, address: <EOA>}
```

The bearer at step 4 is `base64(<SIWE-JSON>:::<sig>)` — a JSON SIWE
message + literal `:::` separator + raw 65-byte ECDSA signature, then
URL-safe base64. Same SIWE message as step 3 (pinned by `nonce`).

`POST /profiles` is what registers the proxy with the backend. The
proxy address is computed by Polymarket's server from the EOA — it's
the V1 CREATE2 proxy address polygolem already derives via
`MakerAddressForSignatureType(eoa, 137, 1)`.

### Implications for polygolem (OBSOLETE)

> **These implications were written based on the proxy-wallet capture below.
> They do not apply to polygolem's deposit-wallet (sigtype-3) path, which
> requires browser login for new users. See ONBOARDING.md for the current state.**

- The `BuildL1HeadersForDepositWallet` / ERC-7739 wrap path is the
  *deposit wallet variant*, not the path new accounts take. Keep it
  for accounts that want deposit wallets, but stop treating it as the
  only path.
- Headless onboarding for a fresh EOA is feasible in pure HTTP — no
  browser needed at runtime. The capture script (Playwright +
  injected EIP-1193 provider) was the investigation tool; production
  polygolem can replicate steps 1-7 above with `net/http` + our
  existing SIWE signer.
- The "Relayer API Key" returned at step 7 is the only API key in V2;
  there is no separate CLOB API key. Trading-tab settings are pure
  UX (FAK/FOK, max button, auto-redeem) — no key creation there.
- Builder profile creation (step 6) is purely a `POST` with the proxy
  address — no signing required beyond the SIWE session cookie.
  Polymarket assigns the bytes32 `builderCode` server-side.

### Reference EOA + proxy used in the capture

```
EOA:           0x3075Af8096c3e5147af22Da45FE7c8496E70a306
proxyWallet:   0x31bd9F0E315586352eb9B6141cEC154C9a71549D
profile id:    8040942
pseudonym:     Limping-Soul
builderCode:   0x47743457d6825e8b6bd996845190ab63bcad2cd74ae096c627bbc2b259173f8b
relayer key:   019e08ca-b8fb-710f-9f5a-94f7227838f2
```

(The EOA private key is in `scripts/playwright-capture/.eoa-key.json` —
gitignored, throwaway test EOA only.)

---

## Original framing (kept for context — see correction above)

**Polygolem's only supported mode is deposit wallet (type 3 / POLY_1271).**
EOA, proxy, and Safe were tested against CLOB V2 — all rejected with
`"maker address not allowed, please use the deposit wallet flow"`.

The HTTP 400 "maker address not allowed, please use the deposit
wallet flow" rejection is the documented V2 enforcement, not a
heuristic. Per docs.polymarket.com/v2-migration and
docs.polymarket.com/trading/deposit-wallet-migration, deposit
wallets are mandatory for new API users; existing proxy/Safe users
are grandfathered. Sigtype 0/1/2 remain useful for grandfathered
accounts but are documented as legacy paths for new users.

## Resolved

| # | Blocker | Fix |
|---|---------|-----|
| B-1 | `parseSignatureTypeFlag` missing "deposit" | +case 3 in `root.go` |
| B-2 | `deposit-wallet` subcommand not wired | +`root.AddCommand(depositWalletCmd)` |
| B-3 | `MakerAddressForSignatureType` no case 3 | +CREATE2 derivation, `case 3` in `signer.go` |
| B-4 | Deposit wallet lifecycle (fund before deploy) | +guard in `live_wire.go` |
| B-5 | CLOB V1 order signing rejected by V2 backend | +V2 structs, `signCLOBOrderV2`, `CLOBVersion`, dispatch |
| B-6 | CREATE2 derivation wrong (simple concat) | Corrected to ERC-1967 formula, verified against official test vector |
| B-7 | `signer` field wrong for type 3 (was EOA) | Fixed: `signer == maker == depositWallet` for type 3 |
| B-8 | Docs referenced EOA/proxy/Safe as viable | All docs updated: deposit wallet only |

### CREATE2 derivation

Original formula (wrong): `c2 || impl || c1` → produced `0xd8F83c...f346`.

Correct ERC-1967 formula (verified against `Polymarket/py-builder-relayer-client`):
```
PREFIX(10) || impl(20) || 0x6009(2) || CONST2(33) || CONST1(33) || args
```
Produces `0x21999a074344610057c9b2B362332388a44502D4`.

Verified against official test vector:
```
owner:    0xA60601A4d903af91855C52BFB3814f6bA342f201
expected: 0x8b60BF0f650Bf7a0d93F10D72375b37De18F8c40  ✅ matches
```

**Hardening track closed (2026-05-07):** The CLOB V2 conformance work
described in docs/superpowers/specs/2026-05-07-clob-v2-conformance-design.md
was implemented across commits 297b9c5 (foxme666 reference fetch),
b0b04c1 (sign in build), 4a1e2b8 (tick-size + postOnly), bbc183e
(neg-risk routing), 0b0c278 (POLY_1271 ERC-7739 envelope), 6b8ddc0
(V1 dead code removal), a481200 (V2 suffix drop), fd365f0 (golden
vectors). All clob tests pass.

### CLOB V2 signing

Orders now use:
- EIP-712 domain: `{name: "Polymarket CTF Exchange", version: "2"}`
- Exchange: `0xE111180000d2663C0091e4f400237545B87B996B`
- Order struct: `salt, maker, signer, tokenId, makerAmount, takerAmount, side, signatureType, timestamp, metadata, builder`
- Version-gated dispatch: checks `/version`, uses V2 when available, falls back to V1

### POLY_1271 order requirements (from official contracts)

From [Polymarket/ctf-exchange-v2 Signatures.sol](https://github.com/Polymarket/ctf-exchange-v2):
```solidity
function verifyPoly1271Signature(address signer, address maker, bytes32 hash, bytes memory signature)
    internal view returns (bool) {
    return (signer == maker) && maker.code.length > 0
        && SignatureCheckerLib.isValidSignatureNow(maker, hash, signature);
}
```

- `signer == maker` == deposit wallet address ✅
- `maker.code.length > 0` — wallet must be deployed ✅ (via WALLET-CREATE)
- `isValidSignature` — EOA signs, DepositWallet contract verifies

## Current status

### B-10 — 2026-05-09 live readiness remediation

Live account:

```
EOA:            0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C
depositWallet: 0x21999a074344610057c9b2B362332388a44502D4
```

Current live state after the 2026-05-09T07:20Z remediation:

- Deposit wallet is deployed on-chain. `eth_getCode` at
  `0x21999a074344610057c9b2B362332388a44502D4` returns non-empty bytecode
  (`0x363d3d37...`), even though the relayer `/deployed` endpoint returns
  `false`.
- POL gas balance is `36.952069` POL.
- EOA on-chain pUSD is `0.000000` pUSD.
- CLOB collateral balance is `1.064918` pUSD.
- CLOB pUSD/CTF allowances are present for three exchange spenders.
- Open CLOB orders are empty.
- Live order budget at `LIVE_ORDER_ALLOCATION_PCT=95` is about `1.0117` pUSD.
- Local live policy pins `MINIMUM_POL=30` for this funded account.

Live transactions broadcast in this run:

```
POL->pUSD swap, amount=1 POL, min reserve=30 POL
tx_hashes:
  0xe95af436e337500533a3b4c02e613693d20d0ca70bad48e05ddb3b0d86fc7f95
  0xbc9f30dfbf3ebd6845d1b112a07cc4cc9658ec99ddb27b50fae5799f47cd74eb

manual EOA->deposit-wallet pUSD transfer, amount=0.10
tx_hash:
  0x14e982e8cccafdf4d8c0bd4a38e902686b7921df30b35f04cba38497280d1474

live startup residual EOA->deposit-wallet pUSD transfer, amount=0.025383
tx_hash:
  0x8f19b1768763b362bf4bdb4c3a63b814434dfa73fce942b400d7f47cfbdfeee1
```

Operational fixes made during this run:

- `go-bot/internal/polygolem` now defaults the SDK relayer URL to
  `https://relayer-v2.polymarket.com` when `POLYMARKET_RELAYER_URL` is unset.
- `go-bot live` now defaults evidence reads to repo `logs/evidence`, matching
  `diagnose-gates` and `run-evidence`.
- Deposit-wallet startup funding is now a no-op when the EOA has zero pUSD;
  CLOB collateral readiness remains the actual funding gate.
- Six-asset MTF inference freshness was restored by backfilling the raw 1m
  candle gap for BTC, ETH, SOL, XRP, DOGE, and BNB.
- `go-bot live` now requires only live-entry applicable gates:
  W01, W02, W03, W04, W05, W07, W09, W10, and W11. W06 is a post-fill
  paper/live divergence monitor; W08 and W12 are not applicable to the
  current crypto MTF path because it does not assume maker rewards or LLM
  calibration.
- `go-bot run-evidence` now populates W08-W12. W09 checks the documented CLOB
  `/simplified-markets` schema, W10 checks `https://polymarket.com/api/geoblock`
  without storing the detected IP, and W11 checks CLOB collateral, allowances,
  and order-read operational readiness.
- Local `go-bot/.env` now sets `MINIMUM_POL=30`, matching the explicit operator
  reserve used for the successful guarded live startup.
- Local `go-bot/.env` now sets `LIVE_ORDER_ALLOCATION_PCT=95` for this
  underfunded account so market-FOK fallback can spend enough USD amount to
  exercise the live path when a limit order would be below market size.
- `go-bot` now requires `LIVE_MARKET_FALLBACK=true` before converting a
  below-min-size limit intent into a buy `market-order` FOK, and it passes the
  executable price as the market order's slippage cap.
- Polygolem market-order signing now matches the official py-clob-client shape:
  buy market orders are built from USD amount and price cap without a local
  minimum-share precheck; the production CLOB remains the final acceptance gate.
- `go-bot wallet-status` no longer reports empty EOA pUSD as `BLOCKED` when
  CLOB collateral is already funded; the live-trading funding checklist now
  follows the CLOB collateral source of truth.

Evidence status after the 2026-05-09T07:20Z remediation:

```
diagnose-gates:
  populated=12 empty=0 stale=0
  W01-W12 OK

guarded live verification:
  LIVE_MAX_BUY_PRICE=0.0001
  MINIMUM_POL=30
  LIVE_ORDER_ALLOCATION_PCT=95
  gate_blocked=0
  funding_blocked=0
  validation_fail=0
  submitted=0
  accepted=0
  no_trade=6
  processed=12
  decided=6
```

No active guarded live-startup blockers remain after refreshing six-asset
candle data and pinning the account reserve policy. The `timeout` wrapper used
for verification stops the long-running live loop after it prints the summary;
the zero `submitted` count above confirms no order was sent.

Known non-blocking trap:

- Deposit-wallet relayer deploy status remains a false-negative:

   ```
   tx_id=019e0ab5-166e-7aca-81a9-2f7bbd7463a7
   type=WALLET-CREATE
   state=STATE_FAILED
   from=0x33e4ad5a1367fbf7004c637f628a5b78c44fa76c
   to=0x00000000000fb5c9adea0298d729a0cb3823cc07
   nonce=0
   ```

  The current source of truth must be Polygon `eth_getCode`, not relayer
  `/deployed`. Tooling should skip another `WALLET-CREATE` when the derived
  deposit wallet already has code.

### B-11 — 2026-05-09 live alpha run

Live alpha was reached on 2026-05-09T08:01Z with the same account:

```
EOA:            0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C
depositWallet: 0x21999a074344610057c9b2B362332388a44502D4
accepted order: 0xf06cbfacf8101227c240f5d532977000c2d202b490019674ff62bf8c4a057e86
settlement tx:  0xc78239f4fee60721d71e9228eb1de6206dbd5a70f225b9a970c5cc8b6b577f22
status:         matched
makingAmount:   1.069999
takingAmount:   2.098038
```

Additional blockers found and fixed in the live loop:

- Submissions now get a client-side UUID before insert, so accepted/lost
  transitions do not depend on Postgres-generated IDs that the in-memory
  caller never sees.
- Market-FOK buy amounts now use CLOB-compatible USD/share precision.
- CLOB balances below the executable $1 market-buy minimum are treated as
  `LOW`, not live-ready.
- Confirmed auto-swap now handles both zero pUSD and dust pUSD, then funds
  the deposit wallet and refreshes CLOB collateral.
- A stranded router-leg USDC.e balance was wrapped into pUSD with
  `wrap-usdce-to-pusd --amount all --min-pol-reserve 10 --confirm EXECUTE_SWAP`.
- `LIVE_MAX_SIGNAL_AGE` is configurable; the live alpha run used `15m` to
  tolerate normal 5m candle/API lag.
- Production `go-bot live` now sets `MaxSubmissionsPerTick=1`, preventing
  multiple same-tick submissions from racing a stale upstream collateral
  balance after the first fill.

Post-run state:

- `audit.live_submissions`: `accepted=2`, `lost=32`, `pending=0`.
- Active safe pauses: `0`.
- Active live locks: `0`.
- CLOB collateral is dust: `0.021734` pUSD, correctly below the $1 executable
  minimum.
- Final no-auto-swap startup check stops at
  `funding_blocked: order budget below minimum`; no new order is sent.

### B-12 — 2026-05-09 V2 settlement and wrong-window remediation plan

Live order review found two issues that must be fixed before adding funds or
raising live order size:

1. Some accepted 5m buys matched against markets whose `starts_at` did not
   equal the strategy decision window.
2. The current Polygolem trading approval batch covers exchange spenders but
   not the V2 collateral adapters required for redeeming winning positions.

The required remediation path is:

- Polygolem market resolution must expose a strict decision-window resolver
  and return a `window_mismatch` status instead of falling back to another
  available market.
- Data API position DTOs must expose `redeemable`, `mergeable`,
  `negativeRisk`, `outcome`, `outcomeIndex`, `oppositeOutcome`,
  `oppositeAsset`, and `endDate`; there is no separate `resolved` field.
- `pkg/contracts` must expose the V2 collateral adapters and ramp addresses:
  `CtfCollateralAdapter=0xADa100874d00e3331D00F2007a9c336a65009718`,
  `NegRiskCtfCollateralAdapter=0xAdA200001000ef00D07553cEE7006808F895c6F1`.
- Existing deposit wallets need a one-shot adapter-approval WALLET batch
  before V2 split/merge/redeem. New wallet onboarding should include both the
  six trading approvals and the four adapter-readiness approvals.
- `pkg/settlement` should become the SDK surface go-bot calls directly:
  find `redeemable=true` positions, build adapter-targeted redeem calls, and
  submit a capped WALLET batch only after adapter approvals are present.

The process is now documented in `docs/SAFETY.md`, `docs/CONTRACTS.md`,
`docs/DEPOSIT-WALLET-REDEEM-VALIDATION.md`, the README, and the Starlight
guide set.

**Implementation status (2026-05-09 PM):** the SDK/CLI can build the V2
adapter approval and redeem WALLET batches, but live recovery still depends on
Polymarket's relayer accepting those adapter calls. Verified factory source and
Polygon RPC show `DepositWalletFactory.proxy(...)` is `onlyOperator`; direct EOA
submission is not a fallback. If the relayer returns "not in the allowed list",
the remaining blocker is upstream relayer allowlist support or an official
Polymarket redeem route.

Commits:

- `5ece04a` `fix(marketresolver)` — fail-closed decision-window guard
  (`StatusWindowMismatch`, strict `ResolveTokenIDsForWindow`).
- `0800fe4` `feat(contracts)` — V2 collateral adapter and ramp registry,
  `RedeemAdapterFor` helper.
- `c77e735` `feat(relayer,cli)` — `BuildAdapterApprovalCalls`,
  `OnboardDepositWallet` 10-call batch, `deposit-wallet approve-adapters`
  CLI with `--submit --confirm APPROVE_ADAPTERS` gating.
- `753a576` `fix(dataapi)` — Position camelCase tags + V2 redemption
  fields (`redeemable`, `mergeable`, `negativeRisk`, …). Closes a
  pre-existing decode bug where snake_case tags silently zeroed every
  field against the live API.
- `9d3226d` `feat(settlement)` — `pkg/settlement` SDK
  (`FindRedeemable`, `BuildRedeemCall`, `SubmitRedeem`).
- `0593991` `feat(cli)` — `deposit-wallet redeemable` / `redeem` with
  `CTF.isApprovedForAll` adapter pre-check and
  `--submit --confirm REDEEM_WINNERS` gating.

Operator runbook to recover the existing redeemable position on
`0x21999a07…02D4`:

```
# 1. One-shot adapter approval (idempotent).
polygolem deposit-wallet approve-adapters --submit --confirm APPROVE_ADAPTERS --json

# 2. Inspect what's redeemable.
polygolem deposit-wallet redeemable --json

# 3. Dry-run.
polygolem deposit-wallet redeem --json

# 4. Submit. Pre-check refuses to sign if any
#    isApprovedForAll(wallet, adapter) is false.
polygolem deposit-wallet redeem --submit --confirm REDEEM_WINNERS --json
```

If step 1 or step 4 fails with a relayer allowlist rejection, stop. Do not use
raw `ConditionalTokens.redeemPositions`, direct EOA calls, or SAFE/PROXY
relayer examples as a V2 deposit-wallet redeem workaround.

### B-9 — Builder credentials configured

The earlier missing-builder-credentials blocker is no longer active in this
workspace. Guarded live startup reaches `clob api key ready` and passes the
funding gate. Keep the credentials only in local ignored env files; do not
commit or print them.

### Future work

- **ERC-7739 TypedDataSign wrapping**: The SDK wraps the EOA signature in a
  `TypedDataSign` envelope for type 3 orders. Polygolem's `signCLOBOrderV2`
  does direct EIP-712 signing which works for types 0/1/2 but type 3 may need
  the ERC-7739 wrapper. The Python SDK wraps as:
  `sig || appDomainSep || contentsHash || typeString`. Needs verification once
  the deposit wallet is deployed and a test order is attempted.

- **Neg-risk exchange address**: Currently hardcoded to regular exchange V2
  address. Per-market selection needed for neg-risk markets.

## Files changed

| File | Change |
|------|--------|
| `internal/cli/root.go` | +deposit case, +wire `depositWalletCmd` |
| `internal/auth/signer.go` | +deposit constants, +correct CREATE2 derivation, +case 3 |
| `internal/clob/orders.go` | +V2 structs, +`signCLOBOrderV2`, +`CLOBVersion`, +dispatch, signer fix |
| `internal/clob/client.go` | +filter `<nil>` in `firstNonEmpty` |
| `go-bot/internal/app/live_wire.go` | +deposit wallet lifecycle, +`ensureDepositWalletFunded` |
| `go-bot/internal/polygolem/client.go` | +`DepositWalletStatus/Fund/Approve`, fix `--json` flags |
| `README.md` | Deposit wallet only |
| `docs/ARCHITECTURE.md` | Signature types: deposit only |
| `docs/COMMANDS.md` | Default: `deposit`, remove EOA/proxy/safe |
| `docs/SAFETY.md` | Deposit wallet required for all trading |
| `docs/PRD.md` | Deposit wallet is the only supported mode |
| `opensource-projects/README.md` | Deposit wallet repos, mark V1 as deprecated |
| `docs-site/src/content/docs/concepts/deposit-wallets.mdx` | Full signature type reference |
| `docs-site/src/content/docs/concepts/secrets-management.mdx` | Secrets tiers |
| `docs-site/src/content/docs/guides/deposit-wallet-lifecycle.mdx` | Lifecycle + recovery |

## Post-hardening operator verification (procedure)

The CLOB V2 hardening track ships a code-correct V2 implementation
verified by golden-vector and httptest-mock unit tests. Manual
end-to-end verification against the production V2 backend is the
final acceptance gate. The operator runs the steps below once when
they are ready to confirm the wire format reaches the backend
correctly.

This is **not** a CI step.

### Step 1 — Build the latest binary

```bash
cd go-bot/polygolem
go build -o polygolem ./cmd/polygolem
```

### Step 2 — Probe sigtype 0 (EOA) — expected to be rejected by docs design

```bash
POLYMARKET_PRIVATE_KEY=$(grep ^POLYMARKET_PRIVATE_KEY ../.env | cut -d= -f2) \
./polygolem clob create-order \
    --token-id <known-active-token-id> \
    --side buy --price 0.01 --size 1 \
    --signature-type eoa --json
```

Expected: HTTP 400 with `"maker address not allowed, please use the
deposit wallet flow"`. Per
docs.polymarket.com/trading/deposit-wallet-migration this rejection
is the documented enforcement for new API users — it proves V2
signing reaches the backend correctly.

### Step 3 — Probe sigtype 3 (deposit wallet) — expected to succeed once funded

Prerequisites:

1. Builder credentials in `go-bot/.env.builder` (run `./polygolem
   builder onboard` if absent).
2. Deposit wallet deployed and funded:
   ```bash
   ./polygolem deposit-wallet onboard --fund-amount <pUSD-amount> --json
   ```
3. Allowances set:
   ```bash
   ./polygolem clob update-balance \
       --asset-type collateral \
       --signature-type deposit
   ```

Then post a tiny test order at a price far from market so it rests
on the book without filling:

```bash
./polygolem clob create-order \
    --token-id <known-active-token-id> \
    --side buy --price 0.01 --size 1 \
    --signature-type deposit --json
```

Expected: `{"success": true, "orderID": "0x…", "status": "live"}`.
Cancel afterward to avoid an unintentional fill.

### Step 4 — Record outputs

After running steps 2 and 3, append the actual outputs (with
sensitive values redacted as needed) under a new sub-section here in
`BLOCKERS.md`:

```markdown
### Verification run on YYYY-MM-DD

**sigtype 0 (EOA)** — expected docs-defined rejection:
\`\`\`
$ ./polygolem clob create-order --signature-type eoa ...
{"error": "maker address not allowed, please use the deposit wallet flow"}
\`\`\`

**sigtype 3 (deposit wallet)** — expected success:
\`\`\`
$ ./polygolem clob create-order --signature-type deposit ...
{"success": true, "orderID": "0x…"}
\`\`\`

V2 hardening track verified end-to-end.
```

The track's success criterion is: sigtype 0 returns the documented
rejection (proving V2 signing reaches the backend) AND sigtype 3
returns success once builder credentials and the deposit wallet are
onboarded.

### Verification run on 2026-05-07

**sigtype 0 (EOA)** — documented rejection received as expected:

```
$ ./polygolem clob create-order \
    --token 926377706175971731068420551849041218012736398250875962020506643091812084572 \
    --side buy --price 0.01 --size 1 \
    --signature-type eoa --json
HTTP 400 https://clob.polymarket.com/order: {"error":"maker address not allowed, please use the deposit wallet flow"}
```

Token used: open YES-outcome of the regular (non-neg-risk) market
"Will Bitcoin hit $X by …" cohort
("Trump announces US blockade of Hormuz lifted by …", event 36173,
market 1972137).

**Reading of this result:** polygolem's sigtype-0 V2-signed order
reached the production CLOB backend and was recognized as a
well-formed V2 order at the EOA-signing layer. The backend rejected
with the documented V2 enforcement message because the EOA is
classified as a new API user that must use the deposit wallet path.

**Scope of what was actually verified:** sigtypes 0 / 1 / 2
(EOA / proxy / Safe) all share the same signing path — a single
EIP-712 signature over the V2 Order typed-data, against the CTF
Exchange V2 domain (regular or neg-risk per market). This probe
exercises that path. **It says nothing about sigtype 3 (POLY_1271).**
Sigtype 3 uses a different, more involved ERC-7739 wrapped signature
which is structurally distinct.

**sigtype 3 (deposit wallet) — not run.**

Prerequisites missing:
- Builder credentials not configured (B-6 still open). `./polygolem
  deposit-wallet status` reports the missing env vars.
- Deposit wallet at `0x21999a074344610057c9b2B362332388a44502D4` is
  derived but not deployed/approved/funded.

Running this step requires real on-chain transactions (deploy
deposit wallet, fund with pUSD) and a real signed order against
production. Operator should complete the onboarding flow
(`./polygolem builder onboard` then `./polygolem deposit-wallet
onboard --fund-amount <pUSD>` then `./polygolem clob update-balance
--asset-type collateral --signature-type deposit`) before re-running
the probe.

**Status of the POLY_1271 wrapper implementation (corrected at
`3c00f2d`):**

The ERC-7739 envelope is now implemented per the canonical
specification at `docs.polymarket.com/trading/deposit-wallets`:

- OUTER `domSep` in `keccak256(0x1901 || domSep || hashStruct(...))` is the
  CTF Exchange V2 domain (regular `0xE111…996B` or neg-risk
  `0xe2222d2…0F59` per market).
- INNER `TypedDataSign` struct's inline domain fields describe the
  WALLET (`name = "DepositWallet"`, `version = "1"`, `chainId = 137`,
  `verifyingContract = <deposit-wallet>`, `salt = 0`).

Two earlier implementations had this backwards (T4 commit `b5bc00d`:
both DepositWallet; T12 commit `7b3f2a2`: outer DepositWallet, inner
Exchange — both wrong) before the correction at `3c00f2d`.

**The wrapper has NOT yet been verified against a real
`isValidSignature` check on a Polymarket deposit wallet contract.**
The 4 golden vectors pin polygolem's own output as a regression guard
but do not validate against a real V2 deposit-wallet verifier. Real
sigtype-3 verification is a future step that requires the operator to
complete the deposit-wallet onboarding flow above and post a small
test order.
