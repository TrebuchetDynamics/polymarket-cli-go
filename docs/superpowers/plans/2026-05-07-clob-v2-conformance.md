# CLOB V2 Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the gaps in polygolem's partially-landed CLOB V2 order signing so the V2 wire format reaches production correctly across all signature types and market kinds.

**Architecture:** Surgical edits inside `internal/clob/orders.go` and `internal/clob/client.go`. No new packages. The V2 typed-data block already exists; the work is signing in the right place, wrapping POLY_1271 sigs per ERC-7739 (OUTER domain = CTF Exchange V2, INNER struct inline domain = `DepositWallet`), threading neg-risk through verifying-contract selection, and removing now-dead V1 paths.

**Note:** the original Task 4 description below had the wrapper outer/inner domains backwards; this was caught during operator probe research (post-T11) and fixed at commit `3c00f2d`. The corrected structure is documented in the spec at `docs/superpowers/specs/2026-05-07-clob-v2-conformance-design.md` §3.1. The skeleton inside Task 4 below reflects the OLD (incorrect) design — read it as the implementation history, not as guidance for new work.

**Tech Stack:** Go 1.22+, `github.com/ethereum/go-ethereum/signer/core/apitypes`, `github.com/ethereum/go-ethereum/common`, golden-vectored unit tests against the `foxme666/Polymarket-golang` V2 fork.

**Spec:** `docs/superpowers/specs/2026-05-07-clob-v2-conformance-design.md` (commits `23c8790`, `d7b0165`).

---

## File map

| File | Responsibility | Tasks that touch it |
|---|---|---|
| `opensource-projects/repos/foxme666-Polymarket-golang/` | external V2 reference, golden-vector source | T1 |
| `internal/clob/orders.go` | order construction, V2 EIP-712 typed-data, signing, posting | T2, T3, T4, T6, T7, T8 |
| `internal/clob/client.go` | CLOB read API helpers; `toString` nil-handling | T5 |
| `internal/clob/orders_test.go` | unit tests for the above | T3, T4, T5, T6, T7, T9 |
| `BLOCKERS.md` | track-level status | T10 |
| `README.md` | top-level repo framing | T10 |
| `docs/COMMANDS.md`, `docs/PRD.md`, `docs/SAFETY.md`, `docs/DEPOSIT-WALLET-MIGRATION.md` | user-facing prose | T10 |
| `tests/docs_safety_test.go` | docs/code drift pins | T10 |

---

## Task 1: Fetch the foxme666/Polymarket-golang V2 fork

The fork is the source of truth for golden-vector hashes and the POLY_1271 ERC-7739 wrapper layout. Clone it into the workspace's reference repo directory.

**Files:**
- Create: `opensource-projects/repos/foxme666-Polymarket-golang/` (cloned)

- [ ] **Step 1: Confirm the repo is reachable**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/polygolem
ls opensource-projects/repos/ | head -20
```
Expected: a list of existing reference repos (no `foxme666-Polymarket-golang` yet).

- [ ] **Step 2: Clone the fork**

```bash
git clone --depth 1 https://github.com/foxme666/Polymarket-golang.git \
  opensource-projects/repos/foxme666-Polymarket-golang
```
Expected: `Cloning into 'opensource-projects/repos/foxme666-Polymarket-golang'... done.`

If the URL 404s, fall back to:
```bash
git clone https://github.com/Polymarket/polymarket-go.git \
  opensource-projects/repos/foxme666-Polymarket-golang
```
and note the actual upstream in the commit message.

- [ ] **Step 3: Verify V2 content present**

```bash
ls opensource-projects/repos/foxme666-Polymarket-golang/
grep -rn "0xE111180000d2663C0091e4f400237545B87B996B\|signatureType.*3\|POLY_1271\|TypedDataSign\|DepositWallet" \
  opensource-projects/repos/foxme666-Polymarket-golang/ | head -20
```
Expected: matches for V2 contract address, sigtype 3 references, and either `TypedDataSign` or `DepositWallet` references.

- [ ] **Step 4: Commit**

```bash
git add opensource-projects/repos/foxme666-Polymarket-golang
git commit -m "chore: add foxme666/Polymarket-golang as V2 golden-vector reference

Cloned for the CLOB V2 hardening track per
docs/superpowers/specs/2026-05-07-clob-v2-conformance-design.md §3.6.
Source of truth for POLY_1271 ERC-7739 wrapping and golden-vector hashes.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Verify V2 Order field ordering matches the docs

EIP-712 typed-data hashes are field-order-sensitive. The spec requires that polygolem's `signCLOBOrderV2` field array matches the canonical ordering published at `docs.polymarket.com/v2-migration`:

```
Order(uint256 salt, address maker, address signer, uint256 tokenId,
      uint256 makerAmount, uint256 takerAmount,
      uint8 side, uint8 signatureType,
      uint256 timestamp, bytes32 metadata, bytes32 builder)
```

Polygolem currently has the right ordering (verified in `internal/clob/orders.go:446-458` during planning). This task is the explicit verification gate.

**Files:**
- Read: `internal/clob/orders.go:438-486` (`signCLOBOrderV2` typed-data block)

- [ ] **Step 1: Read the current V2 typed-data field array**

```bash
sed -n '446,458p' internal/clob/orders.go
```
Expected output:
```
"Order": {
    {Name: "salt", Type: "uint256"},
    {Name: "maker", Type: "address"},
    {Name: "signer", Type: "address"},
    {Name: "tokenId", Type: "uint256"},
    {Name: "makerAmount", Type: "uint256"},
    {Name: "takerAmount", Type: "uint256"},
    {Name: "side", Type: "uint8"},
    {Name: "signatureType", Type: "uint8"},
    {Name: "timestamp", Type: "uint256"},
    {Name: "metadata", Type: "bytes32"},
    {Name: "builder", Type: "bytes32"},
},
```

- [ ] **Step 2: Diff against the canonical ordering**

The canonical order is `salt, maker, signer, tokenId, makerAmount, takerAmount, side, signatureType, timestamp, metadata, builder`. The current code matches. **No change needed.**

If the file shows a different order, edit `internal/clob/orders.go` to match the canonical and re-run all clob tests.

- [ ] **Step 3: No commit (no change). Note in subsequent commit**

Move on to Task 3. If you did edit, commit with:
```bash
git commit -m "fix(clob): reorder V2 Order EIP-712 fields to match docs

Per docs.polymarket.com/v2-migration the canonical Order field order is
salt, maker, signer, tokenId, makerAmount, takerAmount, side,
signatureType, timestamp, metadata, builder. EIP-712 type-encoding is
order-sensitive.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Fix Test 1 — `buildSignedOrderPayload` must sign the payload

`TestBuildSignedOrderPayloadV2UsesCurrentCLOBShape` calls `buildSignedOrderPayload` and asserts the returned payload has a populated `Signature` field. The current implementation returns an unsigned struct. The signing happens later, only in `signAndPostOrderV2`. Move signing into `buildSignedOrderPayload` so direct callers (and tests) get a signed payload.

**Files:**
- Modify: `internal/clob/orders.go:488-524` (`buildSignedOrderPayload`)
- Test: `internal/clob/orders_test.go:19-56` (`TestBuildSignedOrderPayloadV2UsesCurrentCLOBShape`)

- [ ] **Step 1: Run the test to confirm the failure**

```bash
go test ./internal/clob/ -run TestBuildSignedOrderPayloadV2UsesCurrentCLOBShape -v
```
Expected: `FAIL` with `signature shape=""`.

- [ ] **Step 2: Edit `buildSignedOrderPayload` to sign the V2 payload**

In `internal/clob/orders.go`, replace the V2 branch (lines 498-512) so the function signs the payload before returning. The full new V2 branch:

```go
if clobVersion >= 2 {
    payload := signedOrderPayloadV2{
        Salt:          salt,
        Maker:         maker,
        Signer:        signer.Address(),
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
    sig, err := signCLOBOrderV2(signer, payload)
    if err != nil {
        return nil, err
    }
    payload.Signature = sig
    return payload, nil
}
```

(POLY_1271 sigtype 3 specifics — `Signer = Maker` and ERC-7739 wrapping — are handled in Task 4. For now this fix makes Test 1 pass for sigtype 0; Test 2 will still fail.)

- [ ] **Step 3: Run Test 1 to confirm it passes**

```bash
go test ./internal/clob/ -run TestBuildSignedOrderPayloadV2UsesCurrentCLOBShape -v
```
Expected: `PASS`.

- [ ] **Step 4: Run Test 2 to confirm the remaining failure mode**

```bash
go test ./internal/clob/ -run TestBuildSignedOrderPayloadV2DepositWalletUsesEOASignerWithDepositMaker -v
```
Expected: `FAIL`. The maker/signer mismatch failure should still surface (Task 4 fixes it).

- [ ] **Step 5: Commit**

```bash
git add internal/clob/orders.go
git commit -m "fix(clob): buildSignedOrderPayload signs the V2 payload before returning

Test 1 (TestBuildSignedOrderPayloadV2UsesCurrentCLOBShape) called
buildSignedOrderPayload expecting a signed V2 order; the function
returned an unsigned struct. Signing happened only in
signAndPostOrderV2. Move signing into the build function so direct
callers and tests receive a signed payload.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Fix Test 2 — POLY_1271 sigtype 3 with ERC-7739 envelope

`TestBuildSignedOrderPayloadV2DepositWalletUsesEOASignerWithDepositMaker` asserts:

1. For `signatureType == signatureTypePoly1271` (3), `Maker` and `Signer` must both equal the deposit-wallet address (CREATE2-derived from the EOA). The EOA address must NOT appear on the order.
2. The wrapped signature length must be exactly 636 hex characters.

The 636-char output is produced by an ERC-7739 `TypedDataSign` envelope per `docs.polymarket.com/trading/deposit-wallet-migration`. The wrapper signs a separate EIP-712 typed-data with a `DepositWallet` domain. The exact byte layout matches the foxme666 fork's helper.

**Files:**
- Modify: `internal/clob/orders.go` (`buildSignedOrderPayload`, add `wrapPOLY1271Signature` helper)
- Test: `internal/clob/orders_test.go:58-89`

- [ ] **Step 1: Locate the foxme666 wrapper**

```bash
grep -rn "TypedDataSign\|DepositWallet\|poly_1271\|poly1271\|wrapSignature" \
  opensource-projects/repos/foxme666-Polymarket-golang/ | head -20
```
Read the matched file(s). Identify:
- The EIP-712 type definition for `TypedDataSign`.
- The exact bytes appended after the inner ECDSA signature (the
  ERC-7739 "content type" header).
- How the deposit wallet address is fed to the inner domain.

If the fork doesn't have the wrapper directly, search ancillary references:
```bash
grep -rn "TypedDataSign\|0x190101" opensource-projects/repos/ | head -10
```

- [ ] **Step 2: Run Test 2 to confirm the failure baseline**

```bash
go test ./internal/clob/ -run TestBuildSignedOrderPayloadV2DepositWalletUsesEOASignerWithDepositMaker -v
```
Expected: `FAIL` with `maker/signer=…/… want deposit wallet 0xfd5041047be8c192c725a66228f141196fa3cf9c`.

- [ ] **Step 3: Update `buildSignedOrderPayload` to set Signer = Maker for sigtype 3**

In `internal/clob/orders.go`, before constructing the `payload` literal in the V2 branch:

```go
orderSigner := signer.Address()
if draft.signatureType == signatureTypePoly1271 {
    orderSigner = maker
}
payload := signedOrderPayloadV2{
    Salt:          salt,
    Maker:         maker,
    Signer:        orderSigner,
    TokenID:       draft.tokenID.String(),
    // … rest unchanged
}
```

- [ ] **Step 4: Add the `wrapPOLY1271Signature` helper**

Append to `internal/clob/orders.go` (near the other signing helpers):

```go
// wrapPOLY1271Signature wraps a raw ECDSA signature into the ERC-7739
// TypedDataSign envelope expected by Polymarket deposit wallets.
//
// Per docs.polymarket.com/trading/deposit-wallet-migration the deposit
// wallet's isValidSignature validates an inner EIP-712 signature whose
// domain identifies the wallet itself:
//
//   name:              "DepositWallet"
//   version:           "1"
//   chainId:           137
//   verifyingContract: <deposit wallet address>
//   salt:              0x000…
//
// The exact byte layout matches the foxme666/Polymarket-golang fork
// helper; the output is 636 hex chars.
func wrapPOLY1271Signature(signer *auth.PrivateKeySigner, depositWallet string, innerSig string, orderTypedHash [32]byte) (string, error) {
    // Implementation tracks the foxme666 reference. Copy the exact byte
    // layout: inner ECDSA signature, then the appended ERC-7739 content
    // bytes (typed-data hash | content type header | length prefix).
    // See opensource-projects/repos/foxme666-Polymarket-golang/<file> for
    // the canonical implementation.
    // …
}
```

The exact byte assembly comes from the foxme666 fork located in Step 1. Reproduce it line-for-line, then verify the output length is 636 hex chars.

- [ ] **Step 5: Wire the wrapper into `buildSignedOrderPayload`**

After the existing `sig, err := signCLOBOrderV2(signer, payload)` call, branch on sigtype:

```go
if draft.signatureType == signatureTypePoly1271 {
    typedHash, err := signCLOBOrderV2TypedHash(signer, payload)
    if err != nil {
        return nil, err
    }
    sig, err = wrapPOLY1271Signature(signer, maker, sig, typedHash)
    if err != nil {
        return nil, err
    }
}
payload.Signature = sig
return payload, nil
```

If `signCLOBOrderV2TypedHash` does not yet exist, extract a small helper from `signCLOBOrderV2` that returns the EIP-712 hash bytes (separate from signing). This helper is needed both inside `signCLOBOrderV2` and inside `wrapPOLY1271Signature`.

- [ ] **Step 6: Run Test 2 to confirm pass**

```bash
go test ./internal/clob/ -run TestBuildSignedOrderPayloadV2DepositWalletUsesEOASignerWithDepositMaker -v
```
Expected: `PASS`.

- [ ] **Step 7: Run all clob tests; Test 3 should still fail**

```bash
go test ./internal/clob/ -v 2>&1 | tail -20
```
Expected: Tests 1 and 2 pass; Test 3 still fails on `invalid minimum order size "<nil>"`.

- [ ] **Step 8: Commit**

```bash
git add internal/clob/orders.go
git commit -m "fix(clob): POLY_1271 sigtype 3 signs with ERC-7739 TypedDataSign envelope

For signatureType=3 the deposit wallet's isValidSignature validates a
wrapped signature, not a raw ECDSA. Per
docs.polymarket.com/trading/deposit-wallet-migration the wrapper signs
a TypedDataSign payload with a DepositWallet domain whose
verifyingContract is the deposit wallet itself. Set Order.Signer to
the deposit wallet (not the EOA) and emit the 636-char wrapped
signature. Reference: foxme666/Polymarket-golang.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Fix Test 3 — tick-size `<nil>` handling and `postOnly` field

`TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo` fails on `invalid minimum order size "<nil>"`. The `/tick-size` mock returns only `{"minimum_tick_size":"0.001"}` (no `minimum_order_size`). The `toString` lambda inside `client.TickSize` formats a nil interface{} as the literal string `"<nil>"`. Treat nil as empty.

The same test asserts the POST body has `postOnly: false` and `deferExec: false`. The current `sendOrderPayloadV2` struct lacks a `PostOnly` field; add it.

**Files:**
- Modify: `internal/clob/client.go:326-337` (the `toString` closure)
- Modify: `internal/clob/orders.go:100-105` (`sendOrderPayloadV2`)
- Modify: `internal/clob/orders.go:340-345` (where `sendOrderPayloadV2` is constructed)
- Test: `internal/clob/orders_test.go:91-154`

- [ ] **Step 1: Run Test 3 to confirm the baseline failure**

```bash
go test ./internal/clob/ -run TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo -v
```
Expected: `FAIL` with `invalid minimum order size "<nil>"`.

- [ ] **Step 2: Patch `toString` to handle nil**

In `internal/clob/client.go`, replace lines 326-337 with:

```go
toString := func(v interface{}) string {
    if v == nil {
        return ""
    }
    switch val := v.(type) {
    case string:
        return val
    case float64:
        return strconv.FormatFloat(val, 'f', -1, 64)
    case json.Number:
        return val.String()
    default:
        return fmt.Sprintf("%v", val)
    }
}
```

- [ ] **Step 3: Add `PostOnly` to `sendOrderPayloadV2`**

In `internal/clob/orders.go`, replace the `sendOrderPayloadV2` struct definition (lines 100-105) with:

```go
type sendOrderPayloadV2 struct {
    Order     signedOrderPayloadV2 `json:"order"`
    Owner     string               `json:"owner"`
    OrderType string               `json:"orderType"`
    PostOnly  bool                 `json:"postOnly"`
    DeferExec bool                 `json:"deferExec"`
}
```

- [ ] **Step 4: Set `PostOnly: false` in the construction site**

In `internal/clob/orders.go`, find the `sendOrderPayloadV2{` literal in `signAndPostOrderV2` (around line 340) and add the field:

```go
payload := sendOrderPayloadV2{
    Order:     unsigned,
    Owner:     key.Key,
    OrderType: draft.orderType,
    PostOnly:  false,
    DeferExec: false,
}
```

- [ ] **Step 5: Run Test 3 to confirm pass**

```bash
go test ./internal/clob/ -run TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo -v
```
Expected: `PASS`.

- [ ] **Step 6: Run all clob tests**

```bash
go test ./internal/clob/ -v 2>&1 | tail -20
```
Expected: all three target tests pass; no regressions.

- [ ] **Step 7: Commit**

```bash
git add internal/clob/client.go internal/clob/orders.go
git commit -m "fix(clob): tick-size nil handling and explicit postOnly field

(1) client.TickSize's toString lambda was formatting a nil interface{}
as the literal string \"<nil>\", causing parseRat to fail downstream
when /tick-size omitted minimum_order_size. Treat nil as empty.

(2) sendOrderPayloadV2 now includes PostOnly bool, set explicitly to
false at construction. Test 3 asserts posted[\"postOnly\"] == false.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Per-market neg-risk verifyingContract selection

The V2 typed-data verifyingContract differs between regular markets (`0xE111…996B`) and neg-risk markets (`0xe2222d279d744050d28e00520010520000310F59`). The current `signCLOBOrderV2` always uses the regular address. Look up `c.NegRisk(ctx, tokenID)` per-market and thread the flag through.

**Files:**
- Modify: `internal/clob/orders.go` (add constant; modify `signCLOBOrderV2`, `buildSignedOrderPayload`, `signAndPostOrderV2`)
- Test: `internal/clob/orders_test.go` (add a new test asserting neg-risk routing)

- [ ] **Step 1: Write the failing test**

Append to `internal/clob/orders_test.go`:

```go
func TestSignCLOBOrderV2UsesNegRiskExchangeAddressWhenFlagged(t *testing.T) {
    signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
    if err != nil { t.Fatal(err) }
    payload := signedOrderPayloadV2{
        Salt:          1,
        Maker:         signer.Address(),
        Signer:        signer.Address(),
        TokenID:       "12345",
        MakerAmount:   "700000",
        TakerAmount:   "1400000",
        Side:          "BUY",
        SignatureType: 0,
        Timestamp:     "1778125000123",
        Metadata:      bytes32Zero,
        Builder:       bytes32Zero,
    }
    sigRegular, err := signCLOBOrderV2(signer, payload, false)
    if err != nil { t.Fatal(err) }
    sigNegRisk, err := signCLOBOrderV2(signer, payload, true)
    if err != nil { t.Fatal(err) }
    if sigRegular == sigNegRisk {
        t.Fatalf("regular and neg-risk signatures must differ; both = %q", sigRegular)
    }
}
```

- [ ] **Step 2: Run the test; expect compile failure**

```bash
go test ./internal/clob/ -run TestSignCLOBOrderV2UsesNegRiskExchangeAddressWhenFlagged -v
```
Expected: build failure — `signCLOBOrderV2` does not take a `bool` argument yet.

- [ ] **Step 3: Add the neg-risk constant and reshape `signCLOBOrderV2`**

In `internal/clob/orders.go` near the other constants:

```go
const (
    clobExchangeAddress         = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E" // V1 (removed in T7)
    clobExchangeAddressV2       = "0xE111180000d2663C0091e4f400237545B87B996B"
    negRiskExchangeAddressV2    = "0xe2222d279d744050d28e00520010520000310F59"
    zeroAddress                 = "0x0000000000000000000000000000000000000000"
    bytes32Zero                 = "0x0000000000000000000000000000000000000000000000000000000000000000"
    signatureTypePoly1271       = 3
)
```

Modify `signCLOBOrderV2`'s signature and `VerifyingContract` selection:

```go
func signCLOBOrderV2(signer *auth.PrivateKeySigner, order signedOrderPayloadV2, negRisk bool) (string, error) {
    // … unchanged sideInt / typed-data type definition …
    verifyingContract := clobExchangeAddressV2
    if negRisk {
        verifyingContract = negRiskExchangeAddressV2
    }
    typed := apitypes.TypedData{
        // … unchanged ...
        Domain: apitypes.TypedDataDomain{
            Name:              "Polymarket CTF Exchange",
            Version:           "2",
            ChainId:           auth.EIP712ChainID(polygonChainID),
            VerifyingContract: verifyingContract,
        },
        // … unchanged …
    }
    // … unchanged signing …
}
```

- [ ] **Step 4: Update callers**

Update `buildSignedOrderPayload` to accept and pass the flag:

```go
func buildSignedOrderPayload(signer *auth.PrivateKeySigner, draft orderDraft, clobVersion int64, ts time.Time, negRisk bool) (interface{}, error) {
    // … unchanged salt, maker derivation …
    if clobVersion >= 2 {
        // … construct payload … (Tasks 3 + 4 logic stays)
        sig, err := signCLOBOrderV2(signer, payload, negRisk)
        // … rest unchanged …
    }
    // … V1 branch unchanged for now …
}
```

Update `signAndPostOrderV2` to do the neg-risk lookup and pass through:

```go
func (c *Client) signAndPostOrderV2(ctx context.Context, privateKey string, signer *auth.PrivateKeySigner, key *auth.APIKey, salt uint64, maker string, draft orderDraft) (*OrderPlacementResponse, error) {
    nr, err := c.NegRisk(ctx, draft.tokenID.String())
    if err != nil {
        return nil, fmt.Errorf("neg-risk lookup: %w", err)
    }
    negRisk := nr.NegRisk
    // … now construct + sign payload using signCLOBOrderV2(signer, unsigned, negRisk) …
}
```

`signAndPostOrderV1` stays untouched in this task (deleted in Task 7).

`buildSignedOrderPayload`'s test callers in `orders_test.go` already pass V2 path; update those call sites to add the new bool argument:

```go
// Test 1
payload, err := buildSignedOrderPayload(signer, orderDraft{…}, 2, time.UnixMilli(1778125000123), false)
// Test 2
payload, err := buildSignedOrderPayload(signer, orderDraft{…}, 2, time.UnixMilli(1778125000123), false)
```

- [ ] **Step 5: Patch the existing test 3 mock to return a `/neg-risk` response**

Test 3's httptest handler does not currently mock `/neg-risk`. Add:

```go
case "/neg-risk":
    _, _ = w.Write([]byte(`{"neg_risk":false}`))
```

- [ ] **Step 6: Run the new test and the suite**

```bash
go test ./internal/clob/ -v 2>&1 | tail -25
```
Expected: all clob tests pass, including the new neg-risk routing test.

- [ ] **Step 7: Commit**

```bash
git add internal/clob/orders.go internal/clob/orders_test.go
git commit -m "feat(clob): per-market neg-risk verifyingContract selection in V2 signing

V2 has two Exchange contracts: regular
(0xE111180000d2663C0091e4f400237545B87B996B) and neg-risk
(0xe2222d279d744050d28e00520010520000310F59). The current V2 path
always used the regular address, producing invalid signatures for
neg-risk markets. Look up clob.NegRisk(tokenId) and pass the flag
through buildSignedOrderPayload to signCLOBOrderV2.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Remove V1 dead code

The CLOB V1 cutover was April 28, 2026. V1 is dead. The runtime version-dispatch and V1 typed-data block are dead code. Removal also drops the `versionCalled` assertion from Test 3 (no `/version` lookup remains).

**Files:**
- Modify: `internal/clob/orders.go` (delete V1 constants, structs, functions, branches)
- Modify: `internal/clob/orders_test.go` (drop `/version` mock and `versionCalled` assertion)

- [ ] **Step 1: Delete `clobExchangeAddress` (V1) constant**

In `internal/clob/orders.go`, remove the line:
```go
clobExchangeAddress   = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E" // V1
```

- [ ] **Step 2: Delete the V1 `signedOrderPayload` struct and the V1 `sendOrderPayload` struct**

Remove lines defining `signedOrderPayload` (with `Taker, Nonce, FeeRateBps`) and the V1 `sendOrderPayload`. Keep `signedOrderPayloadV2` and `sendOrderPayloadV2` for now (renamed in T8).

- [ ] **Step 3: Delete `signCLOBOrder` (V1) function**

Remove the entire V1 typed-data block (the function with `Version: "1"` and the V1 field set including `taker`, `expiration`, `nonce`, `feeRateBps`).

- [ ] **Step 4: Delete `signAndPostOrderV1` method**

Remove the entire V1 method.

- [ ] **Step 5: Delete `CLOBVersion` and the V1 dispatch branch**

In `signAndPostOrder`, replace the version-switch with a direct V2 call:

```go
func (c *Client) signAndPostOrder(ctx context.Context, privateKey string, draft orderDraft) (*OrderPlacementResponse, error) {
    signer, err := auth.NewPrivateKeySigner(privateKey, polygonChainID)
    if err != nil {
        return nil, err
    }
    key, err := c.DeriveAPIKey(ctx, privateKey)
    if err != nil {
        return nil, fmt.Errorf("derive api key: %w", err)
    }
    salt, err := generateOrderSalt()
    if err != nil {
        return nil, err
    }
    maker, err := auth.MakerAddressForSignatureType(signer.Address(), polygonChainID, draft.signatureType)
    if err != nil {
        return nil, err
    }
    return c.signAndPostOrderV2(ctx, privateKey, signer, &key, salt, maker, draft)
}
```

Then remove the `CLOBVersion` method entirely.

- [ ] **Step 6: Drop the V1 branch in `buildSignedOrderPayload`**

`buildSignedOrderPayload` becomes V2-only. Remove the `clobVersion int64` parameter, return the concrete `signedOrderPayloadV2` instead of `interface{}`:

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
        typedHash, err := signCLOBOrderV2TypedHash(signer, payload, negRisk)
        if err != nil { return signedOrderPayloadV2{}, err }
        sig, err = wrapPOLY1271Signature(signer, maker, sig, typedHash)
        if err != nil { return signedOrderPayloadV2{}, err }
    }
    payload.Signature = sig
    return payload, nil
}
```

- [ ] **Step 7: Update `signAndPostOrderV2` to call the now-V2-only build function**

```go
func (c *Client) signAndPostOrderV2(ctx context.Context, privateKey string, signer *auth.PrivateKeySigner, key *auth.APIKey, salt uint64, maker string, draft orderDraft) (*OrderPlacementResponse, error) {
    nr, err := c.NegRisk(ctx, draft.tokenID.String())
    if err != nil {
        return nil, fmt.Errorf("neg-risk lookup: %w", err)
    }
    unsigned, err := buildSignedOrderPayload(signer, draft, time.Now(), nr.NegRisk)
    if err != nil {
        return nil, err
    }
    payload := sendOrderPayloadV2{
        Order:     unsigned,
        Owner:     key.Key,
        OrderType: draft.orderType,
        PostOnly:  false,
        DeferExec: false,
    }
    return c.postOrder(ctx, privateKey, key, payload, draft.orderType)
}
```

The unused `salt`, `maker` parameters fall away — drop them from the signature, and update the caller in `signAndPostOrder` accordingly. (Salt and maker are now derived inside `buildSignedOrderPayload`.)

- [ ] **Step 8: Drop `/version` from Test 3's mock and the `versionCalled` assertion**

In `internal/clob/orders_test.go::TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo`, remove:

```go
case "/version":
    versionCalled = true
    _, _ = w.Write([]byte(`{"version":2}`))
```

and:

```go
if !versionCalled {
    t.Fatal("expected CLOB /version lookup before signing")
}
```

(the `versionCalled` variable becomes unused; remove its declaration too).

Also, since Tests 1 and 2 call `buildSignedOrderPayload` without the `clobVersion` int parameter now (T6 added negRisk), update them to drop the `2,` argument. The expected signatures now look like:

```go
payload, err := buildSignedOrderPayload(signer, orderDraft{…}, time.UnixMilli(1778125000123), false)
```

- [ ] **Step 9: Run the entire test suite**

```bash
go test ./... 2>&1 | tail -30
```
Expected: all packages pass.

- [ ] **Step 10: Verify V1 leftovers are gone**

```bash
grep -n "clobExchangeAddress\b" internal/clob/orders.go
grep -n "signedOrderPayload\b\|sendOrderPayload\b" internal/clob/orders.go
grep -n "CLOBVersion\b\|signAndPostOrderV1\|signCLOBOrder\b" internal/clob/orders.go
```
Expected: each grep returns either zero matches or only the V2-suffixed identifiers (those are renamed in T8).

- [ ] **Step 11: Commit**

```bash
git add internal/clob/orders.go internal/clob/orders_test.go
git commit -m "chore(clob): remove CLOB V1 dead code

V1 cutover was April 28, 2026 (per docs.polymarket.com/v2-migration).
V1 is dead in production. Remove the runtime version-dispatch, V1
Order struct, V1 typed-data block, V1 send payload, V1 dispatch
branch, and the CLOBVersion() helper. Update Test 3 to drop the
/version mock and versionCalled assertion. buildSignedOrderPayload
becomes V2-only; loses the clobVersion int parameter and returns
the concrete signedOrderPayloadV2 type.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Rename V2-suffixed identifiers

After V1 removal the V2 qualifier is redundant. Rename:
- `signedOrderPayloadV2` → `signedOrderPayload`
- `sendOrderPayloadV2` → `sendOrderPayload`
- `signCLOBOrderV2` → `signCLOBOrder`
- `signCLOBOrderV2TypedHash` (if introduced in T4) → `signCLOBOrderTypedHash`
- `clobExchangeAddressV2` → `clobExchangeAddress`
- `negRiskExchangeAddressV2` → `negRiskExchangeAddress`
- `signAndPostOrderV2` → `signAndPostOrder` (after deleting the wrapper that was just calling through)

**Files:**
- Modify: `internal/clob/orders.go`, `internal/clob/orders_test.go`

- [ ] **Step 1: Rename in implementation**

```bash
# Order matters: replace the longer V2 names first so the prefix
# `signCLOBOrderV2` is consumed before `signCLOBOrderV2TypedHash`
# is also matched by `signCLOBOrderV2`.
sed -i \
  -e 's/signCLOBOrderV2TypedHash/signCLOBOrderTypedHash/g' \
  -e 's/signedOrderPayloadV2/signedOrderPayload/g' \
  -e 's/sendOrderPayloadV2/sendOrderPayload/g' \
  -e 's/signCLOBOrderV2/signCLOBOrder/g' \
  -e 's/clobExchangeAddressV2/clobExchangeAddress/g' \
  -e 's/negRiskExchangeAddressV2/negRiskExchangeAddress/g' \
  internal/clob/orders.go internal/clob/orders_test.go
```

- [ ] **Step 2: Merge `signAndPostOrderV2` into the dispatcher**

After the sed pass, `signAndPostOrder` calls a `signAndPostOrder` method on itself (same name as the wrapper that previously dispatched). Resolve by inlining: delete the trampoline `signAndPostOrder` (the one that just calls through to V2) and rename the V2 method (`signAndPostOrderV2` post-rename → `signAndPostOrder`).

Concretely: after the sed, both methods are named `signAndPostOrder`. Delete the wrapper (the small one calling `c.signAndPostOrder(...) `), keep the substantive method.

- [ ] **Step 3: Build to ensure the rename compiles**

```bash
go build ./...
```
Expected: clean build.

- [ ] **Step 4: Run the full test suite**

```bash
go test ./... 2>&1 | tail -20
```
Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/clob/orders.go internal/clob/orders_test.go
git commit -m "refactor(clob): drop V2 suffix from CLOB type and helper names

After V1 removal the V2 qualifier is redundant. signedOrderPayloadV2 →
signedOrderPayload, sendOrderPayloadV2 → sendOrderPayload,
signCLOBOrderV2 → signCLOBOrder, clobExchangeAddressV2 →
clobExchangeAddress, negRiskExchangeAddressV2 →
negRiskExchangeAddress. Merge signAndPostOrderV2 wrapper.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: Golden-vector fixtures

Pin the EIP-712 typed-data hashes for four (sigtype × neg-risk) combinations. Hashes sourced from the foxme666 fork. If foxme666 ships hash fixtures directly, copy them. If not, compute the V2 typed-data hash by running the foxme666 reference's signing function with the same inputs and capture the intermediate hash.

**Files:**
- Modify: `internal/clob/orders.go` (introduce a test seam so salt and timestamp can be deterministic)
- Test: `internal/clob/orders_golden_test.go` (new)

- [ ] **Step 1: Add a test seam for salt and timestamp**

In `internal/clob/orders.go`, replace direct `generateOrderSalt()` / `time.Now()` calls inside `buildSignedOrderPayload` with package-level variables that tests can override:

```go
var (
    orderSalt = generateOrderSalt
    orderNow  = time.Now
)

func buildSignedOrderPayload(signer *auth.PrivateKeySigner, draft orderDraft, ts time.Time, negRisk bool) (signedOrderPayload, error) {
    if ts.IsZero() {
        ts = orderNow()
    }
    salt, err := orderSalt()
    if err != nil {
        return signedOrderPayload{}, err
    }
    // … rest unchanged …
}
```

(`signAndPostOrder` continues to call `buildSignedOrderPayload(signer, draft, time.Time{}, negRisk)` so the override applies; explicit timestamps in tests bypass `orderNow`.)

- [ ] **Step 2: Capture reference hashes from the foxme666 fork**

For each fixture, identify the foxme666 helper that produces the V2 typed-data hash and run it once with the exact inputs below to capture the expected hash and signature. Record the resulting hex strings.

Fixture 1 — sigtype 0 (EOA), regular market:
```
privateKey:    testOrderPrivateKey  // existing const
salt:          uint64(1)
tokenId:       big.NewInt(12345)
side:          "BUY"
makerAmount:   "700000"
takerAmount:   "1400000"
sigType:       0
timestampMs:   1778125000123
negRisk:       false
```

Fixture 2 — sigtype 1 (proxy), regular market: same as 1 but `sigType: 1`.

Fixture 3 — sigtype 2 (Safe), neg-risk market: `sigType: 2`, `negRisk: true`.

Fixture 4 — sigtype 3 (POLY_1271), regular market: `sigType: 3`, `negRisk: false`.

If the foxme666 fork does not have a directly callable hashing helper, write a tiny throwaway main.go in a tmp dir that imports the fork's signer and prints the hash. Once the hash is captured, discard the throwaway.

- [ ] **Step 3: Write the golden-vector test file**

Create `internal/clob/orders_golden_test.go`:

```go
package clob

import (
    "math/big"
    "testing"
    "time"

    "github.com/TrebuchetDynamics/polygolem/internal/auth"
)

type goldenFixture struct {
    name        string
    sigType     int
    negRisk     bool
    expectedHash string // 0x-prefixed
    expectedSig  string // 0x-prefixed
}

var goldenFixtures = []goldenFixture{
    {name: "eoa_regular", sigType: 0, negRisk: false, expectedHash: "<paste from step 2>", expectedSig: "<paste>"},
    {name: "proxy_regular", sigType: 1, negRisk: false, expectedHash: "<paste>", expectedSig: "<paste>"},
    {name: "safe_negrisk", sigType: 2, negRisk: true, expectedHash: "<paste>", expectedSig: "<paste>"},
    {name: "poly1271_regular", sigType: 3, negRisk: false, expectedHash: "<paste>", expectedSig: "<paste>"},
}

func TestGoldenVectorsAgainstFoxme666Reference(t *testing.T) {
    origSalt := orderSalt
    origNow := orderNow
    t.Cleanup(func() { orderSalt = origSalt; orderNow = origNow })

    orderSalt = func() (uint64, error) { return 1, nil }
    orderNow = func() time.Time { return time.UnixMilli(1778125000123) }

    signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
    if err != nil { t.Fatal(err) }

    for _, fx := range goldenFixtures {
        t.Run(fx.name, func(t *testing.T) {
            payload, err := buildSignedOrderPayload(signer, orderDraft{
                tokenID:       big.NewInt(12345),
                side:          "BUY",
                makerAmount:   "700000",
                takerAmount:   "1400000",
                signatureType: fx.sigType,
                orderType:     "GTC",
            }, time.UnixMilli(1778125000123), fx.negRisk)
            if err != nil { t.Fatal(err) }

            hash, err := signCLOBOrderTypedHash(signer, payload, fx.negRisk)
            if err != nil { t.Fatal(err) }
            gotHash := hashHex(hash)
            if gotHash != fx.expectedHash {
                t.Fatalf("hash mismatch for %s: got %s want %s", fx.name, gotHash, fx.expectedHash)
            }
            if payload.Signature != fx.expectedSig {
                t.Fatalf("signature mismatch for %s: got %s want %s", fx.name, payload.Signature, fx.expectedSig)
            }
        })
    }
}

func hashHex(h [32]byte) string {
    const hexChars = "0123456789abcdef"
    out := make([]byte, 2+64)
    out[0] = '0'; out[1] = 'x'
    for i, b := range h {
        out[2+i*2] = hexChars[b>>4]
        out[3+i*2] = hexChars[b&0x0f]
    }
    return string(out)
}
```

If `signCLOBOrderTypedHash` is not yet exported from the implementation file (it was introduced as a helper in T4 if needed), expose it now (lower-case keeps it package-private; that's fine since the test is in the same package).

- [ ] **Step 4: Run the golden-vector test**

```bash
go test ./internal/clob/ -run TestGoldenVectorsAgainstFoxme666Reference -v
```
Expected: all four sub-tests pass.

If a hash mismatch surfaces, recompute the reference value from the foxme666 fork (Step 2) and re-pin. Do not change the implementation to match a wrong reference.

- [ ] **Step 5: Run the entire suite**

```bash
go test ./... 2>&1 | tail -30
```
Expected: all packages pass.

- [ ] **Step 6: Commit**

```bash
git add internal/clob/orders.go internal/clob/orders_golden_test.go
git commit -m "test(clob): golden-vector fixtures for V2 order signing

Pin EIP-712 typed-data hashes and final signatures for four
(sigtype × neg-risk) combinations against the foxme666/Polymarket-golang
V2 fork. Adds an orderSalt/orderNow test seam so timestamps and salts
are deterministic in tests. Catches silent struct/domain drift before
production.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 10: Light docs repositioning

Surgical edits where text claims V1 contracts, client-computed `feeRateBps`, the legacy HMAC builder model, or the wrong empirical-rejection explanation. Update `BLOCKERS.md` to point readers at the V2 migration docs as the authoritative source.

**Files:**
- Modify: `BLOCKERS.md`
- Modify: `README.md`
- Modify: `docs/COMMANDS.md`, `docs/PRD.md`, `docs/SAFETY.md`, `docs/DEPOSIT-WALLET-MIGRATION.md`
- Modify: `tests/docs_safety_test.go` (update any pinned line numbers / claims)

- [ ] **Step 1: Update `BLOCKERS.md`**

In the B-5 section, replace the empirical-finding paragraph that claims "Polymarket gates new API users to deposit-wallet" with the docs-grounded language:

```markdown
The HTTP 400 "maker address not allowed, please use the deposit
wallet flow" rejection is the documented V2 enforcement, not a
heuristic. Per docs.polymarket.com/v2-migration and
docs.polymarket.com/trading/deposit-wallet-migration, deposit
wallets are mandatory for new API users; existing proxy/Safe users
are grandfathered. Sigtype 0/1/2 remain useful for grandfathered
accounts but are documented as legacy.
```

Keep the resolution status table; mark B-5 ✅ when this plan completes.

- [ ] **Step 2: Update `README.md`**

Find any text that mentions V1 Exchange addresses, `feeRateBps` as a client-computed field, or implies the deposit-wallet migration is the only V2 change. Replace with V2-correct language:

```bash
grep -n "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E\|feeRateBps\|signature type" README.md
```

For each match, edit so the surrounding paragraph reflects:
- V2 Exchange addresses (regular and neg-risk).
- Server-side fees (no client-computed `feeRateBps`).
- Four signature types (EOA, proxy, Safe, deposit-wallet); deposit wallet mandatory for new API users.
- Builder attribution is a per-order `builderCode` bytes32, not HMAC.

Avoid global rebrand. Stay surgical.

- [ ] **Step 3: Update `docs/COMMANDS.md`, `docs/PRD.md`, `docs/SAFETY.md`**

Same approach: grep for V1 addresses, `feeRateBps`, and obsolete framing; edit in place.

```bash
for f in docs/COMMANDS.md docs/PRD.md docs/SAFETY.md; do
    echo "=== $f ==="
    grep -n "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E\|feeRateBps\|builder.*HMAC" "$f"
done
```

- [ ] **Step 4: Update `docs/DEPOSIT-WALLET-MIGRATION.md`**

Add a paragraph at the top of section 2 (or wherever the rejection error first appears):

```markdown
**Why this rejection occurs:** Per docs.polymarket.com/v2-migration,
the V2 backend rejects orders whose maker address has no smart-account
contract deployed and approved at the V2 Exchange. The "use the
deposit wallet flow" message is one suggested remediation; for
grandfathered accounts an existing proxy or Safe deployment also
satisfies the requirement.
```

- [ ] **Step 5: Update `tests/docs_safety_test.go` pins**

```bash
go test ./tests/ -run TestDocsSafety -v 2>&1 | tail -20
```
If the test fails on changed line numbers or content pins, edit the assertions in `tests/docs_safety_test.go` to match the new doc content.

- [ ] **Step 6: Run the entire test suite**

```bash
go test ./... 2>&1 | tail -30
```
Expected: all packages pass.

- [ ] **Step 7: Commit**

```bash
git add BLOCKERS.md README.md docs/COMMANDS.md docs/PRD.md docs/SAFETY.md docs/DEPOSIT-WALLET-MIGRATION.md tests/docs_safety_test.go
git commit -m "docs: align V2 prose with official Polymarket docs

Replace V1 Exchange addresses with V2 (regular + neg-risk). Drop
client-computed feeRateBps language. Reframe builder model from
legacy HMAC to per-order builderCode bytes32. Re-explain the
empirical 'use the deposit wallet flow' rejection as the documented
V2 enforcement, citing docs.polymarket.com/v2-migration and
docs.polymarket.com/trading/deposit-wallet-migration.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 11: Operator end-to-end verification (manual)

Not a CI step — a recorded acceptance gate that the operator runs once after the plan lands. Document the steps and the expected outputs in `BLOCKERS.md` so the next operator (or a future audit) can re-run them.

**Files:**
- Modify: `BLOCKERS.md` (append a "post-hardening verification" section)

- [ ] **Step 1: Build the latest binary**

```bash
cd /home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/polygolem
go build -o polygolem ./cmd/polygolem
```

- [ ] **Step 2: Probe sigtype 0 (EOA) — expected to be rejected by docs design**

```bash
POLYMARKET_PRIVATE_KEY=$(grep ^POLYMARKET_PRIVATE_KEY ../.env | cut -d= -f2) \
./polygolem clob create-order \
    --token-id <known-active-token-id> \
    --side buy --price 0.01 --size 1 \
    --signature-type eoa --json
```
Expected: HTTP 400 with `"maker address not allowed, please use the deposit wallet flow"`. This is the docs-defined enforcement and proves V2 signing reaches the backend correctly.

- [ ] **Step 3: Probe sigtype 3 (deposit wallet) — expected to succeed once funded**

This requires:
1. Builder credentials in `.env.builder` (run `./polygolem builder onboard` if absent).
2. Deposit wallet deployed and funded: `./polygolem deposit-wallet onboard --fund-amount <pUSD-amount> --json`.
3. Allowances set: `./polygolem clob update-balance --asset-type collateral --signature-type deposit`.

Then:

```bash
./polygolem clob create-order \
    --token-id <known-active-token-id> \
    --side buy --price 0.01 --size 1 \
    --signature-type deposit --json
```
Expected: `{"success": true, "orderID": "0x…", "status": "live"}`. Cancel afterward to avoid an unintentional fill.

- [ ] **Step 4: Append a verification section to `BLOCKERS.md`**

Append the actual outputs (redacted of token IDs / order IDs as needed) under a new section:

```markdown
## Post-hardening verification (2026-05-XX)

### sigtype 0 (EOA) — expected docs-defined rejection

```
$ ./polygolem clob create-order --signature-type eoa ...
{"error": "maker address not allowed, please use the deposit wallet flow"}
```

This proves V2 signing reaches the backend. Per
docs.polymarket.com/v2-migration this rejection is by design for
new API users.

### sigtype 3 (deposit wallet) — expected success

```
$ ./polygolem clob create-order --signature-type deposit ...
{"success": true, "orderID": "0x…"}
```

V2 hardening track complete.
```

- [ ] **Step 5: Commit**

```bash
git add BLOCKERS.md
git commit -m "docs(blockers): post-hardening V2 verification recorded

V2 signing reaches the production CLOB. sigtype 0 is rejected by
documented design (new API user → deposit wallet). sigtype 3
returns success once builder credentials and the deposit wallet
are onboarded and funded.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Acceptance gate

After all 11 tasks are committed:

```bash
go build ./...                    # clean
go test ./...                     # clean
grep -n "clobExchangeAddress\b\|signedOrderPayload\b" internal/clob/orders.go  # only the renamed (non-V2-suffixed) constants and structs
grep -rn "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E" internal/  # zero matches in non-test source
grep -n "feeRateBps" internal/clob/orders.go  # zero matches (server-side fees in V2)
```

If all checks pass and `BLOCKERS.md` § B-5 is marked ✅ with the post-hardening verification appended, the track is done.
