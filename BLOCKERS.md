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

## Open

### B-9 — Builder credentials not configured

`go-bot/.env` is missing builder credentials. The live loop now fails early
with a clear actionable message instead of the confusing "pUSD balance is zero."

**Live output:**
```
deposit wallet derived: 0x21999a074344610057c9b2B362332388a44502D4
⚠ builder credentials required
  → https://polymarket.com/settings?tab=builder
  → Add POLYMARKET_BUILDER_API_KEY/SECRET/PASSPHRASE to go-bot/.env
  → Then restart: go-bot live
```

**Resolution:**
1. Go to https://polymarket.com/settings?tab=builder
2. Copy API key, secret, passphrase
3. Add to `go-bot/.env`:
   ```
   POLYMARKET_BUILDER_API_KEY=...
   POLYMARKET_BUILDER_SECRET=...
   POLYMARKET_BUILDER_PASSPHRASE=...
   ```
4. `go-bot live` — auto-deploys, funds, approves, starts trading

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
