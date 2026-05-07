# CLOB V2 Hardening — Design

**Status:** approved (brainstormed 2026-05-07)
**Owner:** polygolem
**Track:** B-5 hardening (per `BLOCKERS.md`)
**Goal:** Close the gaps in polygolem's partially-landed CLOB V2 order
signing so the V2 wire format reaches production cleanly across all
signature types and market kinds.

---

## 1. Why this exists

A parallel session implemented the bulk of CLOB V2 conformance in the
working tree:

- V2 contract address `0xE111180000d2663C0091e4f400237545B87B996B`.
- V2 EIP-712 domain version `"2"`.
- V2 Order struct with `timestamp`, `metadata`, `builder` fields.
- V1 fields (`taker`, `nonce`, `feeRateBps`) removed from the V2 wire.
- `CLOBVersion()` `/version` lookup with version-gated dispatch
  (V1 path retained alongside V2).
- An empirical run confirmed V2 signing reaches the production CLOB —
  the response was a real backend error, not a signing failure.

What did **not** land:

1. **Three failing tests** in `internal/clob/orders_test.go`. The tests
   describe the intended contract; the implementation in
   `internal/clob/orders.go` falls short:
   - `TestBuildSignedOrderPayloadV2UsesCurrentCLOBShape` — fails because
     `buildSignedOrderPayload` returns an unsigned struct (signature is
     empty) when the test expects a fully signed V2 order.
   - `TestBuildSignedOrderPayloadV2DepositWalletUsesEOASignerWithDepositMaker`
     — fails because for signature type 3 (POLY_1271), the function
     sets `Signer = EOA` when it must set `Signer = deposit-wallet`
     (POLY_1271 convention puts the wallet contract address in both
     `Maker` and `Signer`; the EOA's ECDSA signature is wrapped into the
     `Signature` bytes for `isValidSignature` validation).
   - `TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo` — fails on
     a tick-size handling bug (`invalid minimum order size "<nil>"`).
2. **Per-market neg-risk verifyingContract selection.** V2's
   `signCLOBOrderV2` uses a single Exchange address for all markets,
   ignoring the `clob.Client.NegRisk(ctx, tokenID)` lookup. Neg-risk
   markets need `0xe2222d279d744050d28e00520010520000310F59`.
3. **V1 dead code.** The cutover was April 28; today is May 7. V1 is
   gone from production. The `CLOBVersion()` runtime check, V1 typed-data
   block, V1 dispatch path, and `clobExchangeAddress` constant are dead
   code that doubles the test surface and risks silent V1 fallback if
   the `/version` endpoint stops responding.
4. **No golden-vector tests.** The current tests assert wire shape and
   field presence but do not pin EIP-712 typed-data hashes against a
   reference implementation. Without that, a subtle struct/domain
   mistake could pass tests and fail in production.
5. **Light docs repositioning.** README, COMMANDS, PRD, SAFETY,
   DEPOSIT-WALLET-MIGRATION, BLOCKERS still carry framing from the
   pre-V2-conformance state.

This track closes each gap.

---

## 2. Scope

### 2.1 In-scope

1. **Fix the 3 failing tests** in `internal/clob/orders_test.go` by
   modifying the implementation in `internal/clob/orders.go` (and
   neighboring files as needed).
2. **Add per-market neg-risk verifyingContract selection.**
   `signAndPostOrder` (or a helper it calls) looks up
   `c.NegRisk(ctx, tokenID)` and threads the result into
   `signCLOBOrderV2`. The V2 typed-data picks
   `clobExchangeAddressV2` for regular markets and
   `negRiskExchangeAddressV2` for neg-risk markets.
3. **Remove V1 dead path.** Delete:
   - `clobExchangeAddress` constant (V1).
   - V1 `signedOrderPayload` struct.
   - V1 `signCLOBOrder` function (the V1 typed-data block at line 425
     range).
   - V1 branch of `signAndPostOrder` (the dispatch becomes
     unconditional V2).
   - V1 branch in `buildSignedOrderPayload` (becomes V2-only).
   - `CLOBVersion()` only if no other caller depends on it; otherwise
     leave the function but stop using it for dispatch.
4. **Add golden-vector tests.** New
   `internal/clob/orders_golden_test.go` (or extend
   `orders_test.go`). Pin EIP-712 typed-data hashes for fixtures
   covering the 4 (signature-type × neg-risk) combinations that real
   operators will hit:
   - sigtype 0 (EOA), regular market.
   - sigtype 1 (proxy), regular market.
   - sigtype 2 (Gnosis Safe), neg-risk market.
   - sigtype 3 (deposit-wallet POLY_1271), regular market.
   Hashes sourced from the `foxme666/Polymarket-golang` V2 fork after it
   is cloned into `opensource-projects/repos/`.
5. **Light docs repositioning** — `BLOCKERS.md`, `README.md`,
   `docs/COMMANDS.md`, `docs/PRD.md`, `docs/SAFETY.md`,
   `docs/DEPOSIT-WALLET-MIGRATION.md`. Surgical edits where V1
   contracts, V1 fields, or out-of-date framing appear. No global
   rebrand.
6. **Fetch `foxme666/Polymarket-golang`** into
   `opensource-projects/repos/` as the golden-vector source.
7. **Update `tests/docs_safety_test.go`** for any pinned line numbers
   or claims that change.

### 2.2 Out-of-scope (explicit non-goals)

1. **Builder attribution wiring.** `builder` field stays
   `bytes32(0)`. The UUID→bytes32 encoding is a separate research and
   wiring task.
2. **`go-bot/internal/app` live readiness gate.** The pUSD readiness
   check that fails `go-bot live` lives outside polygolem. Empirically
   the production CLOB is gating new API users to deposit-wallet flow
   regardless — that operational reality belongs to a follow-up.
3. **Fernet / encrypted secrets** (the `foxme666` fork's
   `PRIVATE_KEY_ENC_FILE` feature). Defer to its own track.
4. **API refactor** (`client.PlaceOrder(order)` builder). Out of
   scope.
5. **Pruning deposit-wallet code.** Stays.
6. **`signing/` subpackage extraction** (proposed earlier when this was
   greenfield work). Reduce scope: keep V2 code where it lives now in
   `internal/clob/orders.go`. Extraction is a follow-up if the file
   crosses 500 lines after V1 removal — currently it sits well under.

### 2.3 Already done elsewhere

- V2 struct (`signedOrderPayloadV2`).
- V2 typed-data block in `signCLOBOrderV2`.
- V2 contract address `clobExchangeAddressV2`.
- Removal of V1 fields from V2 wire JSON (`taker`, `nonce`,
  `feeRateBps`).
- `bytes32Zero` and `signatureTypePoly1271` constants.

---

## 3. The fixes, file by file

### 3.1 `internal/clob/orders.go` — `buildSignedOrderPayload`

The function builds the V2 struct but never signs it. The test calls it
expecting a signed payload back. The fix:

1. Move the V2 path to call `signCLOBOrderV2` before returning.
2. For signature type 3 (POLY_1271), set `Signer = maker` (the deposit
   wallet) instead of `signer.Address()` (the EOA).
3. Wrap the raw ECDSA signature into a POLY_1271 signature blob (318
   bytes; the wire format `Signature` field is 636 hex chars without
   `0x` prefix or 638 with — verify against the failing test
   expectation `len(order.Signature) != 636` and the foxme666
   reference's wrapping helper).
4. Remove the V1 branch (`clobVersion < 2` path) entirely.
5. Drop the `clobVersion int64` parameter from
   `buildSignedOrderPayload`'s signature once V1 is gone — but keep the
   `time.Time` parameter for testability (the test pins
   `time.UnixMilli(1778125000123)`).

Pseudocode for the corrected function:

```go
func buildSignedOrderPayload(signer *auth.PrivateKeySigner, draft orderDraft, ts time.Time, negRisk bool) (signedOrderPayloadV2, error) {
    salt, err := generateOrderSalt()
    if err != nil { return signedOrderPayloadV2{}, err }

    maker, err := auth.MakerAddressForSignatureType(signer.Address(), polygonChainID, draft.signatureType)
    if err != nil { return signedOrderPayloadV2{}, err }

    orderSigner := signer.Address()
    if draft.signatureType == signatureTypePoly1271 {
        orderSigner = maker
    }

    payload := signedOrderPayloadV2{
        Salt:          salt,
        Maker:         maker,
        Signer:        orderSigner,
        TokenID:       draft.tokenID.String(),
        MakerAmount:   draft.makerAmount,
        TakerAmount:   draft.takerAmount,
        Side:          draft.side,
        Expiration:    "0",
        SignatureType: draft.signatureType,
        Timestamp:     fmt.Sprintf("%d", ts.UnixMilli()),
        Metadata:      bytes32Zero,
        Builder:       bytes32Zero,
    }

    sig, err := signCLOBOrderV2(signer, payload, negRisk)
    if err != nil { return signedOrderPayloadV2{}, err }
    if draft.signatureType == signatureTypePoly1271 {
        sig, err = wrapPOLY1271Signature(sig, signer.Address())
        if err != nil { return signedOrderPayloadV2{}, err }
    }
    payload.Signature = sig
    return payload, nil
}
```

The test asserts a non-prefixed-or-prefixed 636-char signature for
sigtype 3 and a 132-char (`0x` + 130 hex) signature for sigtype 0 —
the wrapping helper applies only when sigtype == 3.

### 3.2 `internal/clob/orders.go` — `signCLOBOrderV2`

Add a `negRisk bool` parameter. Inside the function, choose
`VerifyingContract` based on the flag:

```go
verifyingContract := clobExchangeAddressV2
if negRisk {
    verifyingContract = negRiskExchangeAddressV2
}
```

Add the `negRiskExchangeAddressV2` constant (or import from
`internal/relayer/approvals.go` if the visibility allows).

### 3.3 `internal/clob/orders.go` — `signAndPostOrder`

After V1 removal, this becomes:

```go
func (c *Client) signAndPostOrder(ctx context.Context, privateKey string, draft orderDraft) (*OrderPlacementResponse, error) {
    signer, err := auth.NewPrivateKeySigner(privateKey, polygonChainID)
    if err != nil { return nil, err }

    key, err := c.DeriveAPIKey(ctx, privateKey)
    if err != nil { return nil, fmt.Errorf("derive api key: %w", err) }

    negRisk, err := c.NegRisk(ctx, draft.tokenID.String())
    if err != nil { return nil, fmt.Errorf("neg-risk lookup: %w", err) }

    payload, err := buildSignedOrderPayload(signer, draft, time.Now(), negRisk.NegRisk)
    if err != nil { return nil, err }

    body := sendOrderPayloadV2{
        Order:     payload,
        Owner:     key.Key,
        OrderType: draft.orderType,
        DeferExec: false,
    }
    return postOrder(ctx, c, privateKey, &key, body)
}
```

The V1 branch is removed. The `CLOBVersion()` call is removed. The
existing `postOrder` helper stays.

### 3.4 `internal/clob/orders.go` — Tick-size `<nil>` fix

`TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo` fails with
`invalid minimum order size "<nil>"`. The mock returns
`{"minimum_tick_size":"0.001"}` only — no `minimum_order_size` field.
Somewhere a Go `<nil>` value is being string-formatted into the parse
input.

Locate the bug in either `internal/clob/client.go::TickSize` (response
parsing) or `internal/clob/orders.go::validateMinimumOrderSize` (the
caller). The fix is one of:

- Filter the `<nil>` literal alongside empty in the same path
  `firstNonEmpty` already filters.
- Treat a missing `minimum_order_size` as "no minimum constraint"
  rather than parsing it.

Either way, the diff is a few lines. The test pins the desired
behaviour.

### 3.5 V1 dead code removal

Delete:

```
internal/clob/orders.go:
  - line 22: clobExchangeAddress constant (V1)
  - lines 67-79: signedOrderPayload struct (V1)
  - lines 100-104 (or thereabouts): sendOrderPayload struct (V1)
  - lines 365-422 (or thereabouts): signCLOBOrder function (V1 typed-data)
  - signAndPostOrderV1 helper if it exists (collapsed into signAndPostOrder)
  - V1 branch of buildSignedOrderPayload
  - V1 branch of signAndPostOrder
```

If `CLOBVersion()` is only used for V1/V2 dispatch, delete it too.
If it's used elsewhere (operator-side `clob version` command, telemetry,
etc.), leave it but unused-by-orders.

After deletion, rename `signedOrderPayloadV2` → `signedOrderPayload`,
`sendOrderPayloadV2` → `sendOrderPayload`, `signCLOBOrderV2` →
`signCLOBOrder`, `signAndPostOrderV2` → `signAndPostOrder` (if those
suffixed names exist). The `V2` qualifier becomes redundant.

### 3.6 Golden-vector tests

New file `internal/clob/orders_golden_test.go` (or extend
`orders_test.go`). Each fixture:

- Fixed private key (use the existing `testOrderPrivateKey`).
- Fixed salt — bypass random generation by exporting a test seam, e.g.
  `var saltSource = generateOrderSalt` and patching it in the test.
- Fixed timestamp.
- Fixed token ID, side, amounts.
- Fixed neg-risk flag.

Asserts:

1. EIP-712 typed-data hash matches the value derived from the
   foxme666 reference for the same inputs.
2. Final signature hex matches the value derived from the foxme666
   reference (or, if foxme666 doesn't ship hash fixtures, derived twice
   via two independent computations and asserted equal).

The four fixtures match the (sigtype × neg-risk) combinations operators
will actually hit:

| Fixture | sigtype | neg-risk | Maker | Signer |
|---|---|---|---|---|
| 1 | 0 (EOA) | false | EOA | EOA |
| 2 | 1 (proxy) | false | Proxy CREATE2 | EOA |
| 3 | 2 (Safe) | true | Safe CREATE2 | EOA |
| 4 | 3 (POLY_1271) | false | Deposit-wallet CREATE2 | Deposit-wallet CREATE2 |

### 3.7 Docs repositioning (light)

| File | Change |
|---|---|
| `BLOCKERS.md` | Mark B-5 truly closed when this lands; demote B-6 to operational. Note the empirical "new API user → deposit-wallet only" finding so future readers don't repeat the discovery work. |
| `README.md` | Where text claims V1 contract addresses or client-computed fees, update to V2. Keep deposit-wallet section. Do not promote any single sigtype as "the V2 default." |
| `docs/COMMANDS.md` | Same surgical edits where V1 details surface. |
| `docs/PRD.md` | Same. |
| `docs/SAFETY.md` | Same. |
| `docs/DEPOSIT-WALLET-MIGRATION.md` | Add a paragraph noting that for new API accounts, Polymarket gates orders to the deposit-wallet path even when V2 signing is correct on EOA / proxy. This is empirical, not a contract requirement. |
| `tests/docs_safety_test.go` | Update any pins that change. |
| `docs-site/src/content/docs/*` | Mirror README changes where the docs-site duplicates them. |

---

## 4. Test strategy

1. **Existing failing tests pass** — the three tests are the
   acceptance criteria. They are already written; the implementation
   needs to match.
2. **Golden-vector tests added** — four fixtures covering the
   (sigtype × neg-risk) combinations.
3. **HTTP fixture extended** —
   `TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo` already
   covers the wire body shape. After V1 removal it can drop the
   `versionCalled` assertion and the `/version` mock endpoint.
4. **`tests/docs_safety_test.go`** — updated pins.
5. **Operator-side end-to-end verification** — post a tiny test order
   with `--signature-type proxy` (sigtype 1) against the production V2
   backend. Empirically this will return HTTP 400
   "use the deposit wallet flow" for new API users; we accept that
   response as proof the V2 signature reached the backend correctly.
   For accounts already onboarded to deposit-wallet, run
   `--signature-type deposit` (sigtype 3) to confirm a real
   `Success: true` response.

CI does not need to run the operator-side step.

---

## 5. Success criteria

1. `go build ./...` clean.
2. `go test ./...` clean — including the four golden-vector fixtures.
3. Search for `clobExchangeAddress` (the V1 constant name) returns zero
   matches in non-test source.
4. Search for `feeRateBps` returns zero matches in `internal/clob/`
   source (it can still appear in `internal/clob/client.go::FeeRateBps`
   read-only API helper if any caller needs the server-side fee for
   display).
5. `BLOCKERS.md` § B-5 actually closed (the parallel session's claim
   becomes true).
6. README and docs no longer reference V1 contracts or client-computed
   `feeRateBps` as if they were current.

---

## 6. Decisions log

| Decision | Choice | Rationale |
|---|---|---|
| Scope boundary | polygolem-only | go-bot live gate is operational, not a code blocker — separate follow-up. |
| V1 path | drop entirely | Cutover was April 28. V1 is dead. Dispatch is dead code. |
| Builder attribution | zero bytes32 | Encoding research blocks scope. Follow-up. |
| Test strategy | extend existing tests + add golden vectors | Three existing failing tests are the de-facto contract; add hash pins on top. |
| Structure | keep code in `internal/clob/orders.go` | Subpackage extraction was overkill; greenfield framing was wrong. File stays under 500 lines after V1 removal. |
| Golden-vector source | foxme666/Polymarket-golang | April 2026 V2 fork. Cleanest reference. |
| Fernet encryption | defer | Orthogonal. |
| Reframe docs as gasless-Safe-as-default | no | All four sigtypes first-class. |
| Deposit-wallet code | keep | Empirically the only working path for new API accounts. |
| `signing/` subpackage extraction | drop | Was greenfield-thinking; not warranted for surgical fixes. |

---

## 7. Plan handoff

After this spec is approved, the next step is `superpowers:writing-plans`
to produce the bite-sized implementation plan at
`docs/superpowers/plans/2026-05-07-clob-v2-conformance.md`.

The plan will sequence:

1. Fetch `foxme666/Polymarket-golang` into
   `opensource-projects/repos/`.
2. Test 1 fix (sign the V2 payload in `buildSignedOrderPayload`).
3. Test 2 fix (POLY_1271: Signer = wallet; wrap signature).
4. Test 3 fix (tick-size `<nil>` handling).
5. Add per-market neg-risk verifyingContract selection.
6. Wire `c.NegRisk(ctx, tokenID)` into `signAndPostOrder`.
7. Remove V1 dead code (constant, struct, function, dispatch).
8. Rename V2-suffixed identifiers back to unsuffixed.
9. Add golden-vector fixtures (4).
10. Light docs repositioning + `tests/docs_safety_test.go` pins.
11. Operator end-to-end verification.
