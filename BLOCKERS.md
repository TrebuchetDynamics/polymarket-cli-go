# Blockers

Account: EOA `0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C`
Audit: 2026-05-07

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
