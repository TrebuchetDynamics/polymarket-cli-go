# Blockers

Account: EOA `0x33e4aD5A1367fbf7004c637F628A5b78c44Fa76C`
Audit: 2026-05-07

**Polygolem's only supported mode is deposit wallet (type 3 / POLY_1271).**
EOA, proxy, and Safe have been removed from the critical path. The CLI
defaults to `--signature-type deposit`, go-bot defaults to `"deposit"`,
and the live loop enforces it everywhere.

## Resolved

| # | Blocker | Fix |
|---|---------|-----|
| B-1 | `parseSignatureTypeFlag` missing "deposit" | +case 3 in `root.go` |
| B-2 | `deposit-wallet` subcommand not wired | +`root.AddCommand(depositWalletCmd)` |
| B-3 | `MakerAddressForSignatureType` no case 3 | +`deriveDepositWalletAddress` in `signer.go` |
| B-4 | Deposit wallet lifecycle (fund before deploy) | +guard in `live_wire.go` |
| B-5 | CLOB V1 order signing rejected by V2 backend | +V2 structs, signing, version dispatch in `orders.go` |

EOA (type 0), proxy (type 1), and Safe (type 2) were tested against the real
CLOB V2 — all three rejected:

```
HTTP 400: "maker address not allowed, please use the deposit wallet flow"
```

This account is classified as "new API user." Type 3 (deposit wallet / POLY_1271)
is the only accepted mode.

## What the deposit wallet is

Verified on [Polygonscan](https://polygonscan.com/address/0x58CA52ebe0DadfdF531Cde7062e76746de4Db1eB):
the implementation at `0x58CA...Db1eB` is a contract literally named
**`DepositWallet`** — Polymarket's own smart account (Solidity 0.8.34).

```
Contract Name:   DepositWallet
Factory:         0x00000000000Fb5C9ADea0298D729A0CB3823Cc07
Key methods:     isValidSignature (ERC-1271), execute (batch EIP-712),
                 owner(), initialize(address), transferOwnership
```

It is **not** a Gnosis Safe (type 2 uses different factory at `0xaacF...541b`).
It is **not** a proxy wallet (type 1 uses factory at `0xaB45...5052`).
All three are separate contracts at separate CREATE2 addresses.

For this EOA:
```
type 1 (proxy):   0xE6CEBea09b739246a53d5e40aCE665244e5acb13
type 2 (safe):    0x9A78d2fbEe46A6c3A59c1dAb786bae7df909AD02
type 3 (deposit): 0xd8F83c7021e1CA644Bc20177d1301d7FBa02f346
```

The CLOB requires a deployed `DepositWallet` contract at the type-3 address
for this account. It deploys via `WALLET-CREATE` on the Polymarket relayer,
signed by the EOA. Once deployed, pUSD must be transferred from EOA to the
deposit wallet, and 6 contract approvals submitted via a `WALLET` batch.

## Open

### B-6 — Builder credentials not configured

`go-bot/.env` is missing:
```
POLYMARKET_BUILDER_API_KEY=...
POLYMARKET_BUILDER_SECRET=...
POLYMARKET_BUILDER_PASSPHRASE=...
```

Get them at: https://polymarket.com/settings?tab=builder

Then:
```bash
polygolem deposit-wallet onboard --fund-amount 14
polygolem clob update-balance --asset-type collateral --signature-type deposit
go-bot live
```

## Files changed

| File | Change |
|------|--------|
| `internal/cli/root.go` | +deposit case in `parseSignatureTypeFlag`, +wire `depositWalletCmd` |
| `internal/auth/signer.go` | +deposit constants, +`deriveDepositWalletAddress`, +case 3 |
| `internal/clob/orders.go` | +V2 structs, +`signCLOBOrderV2`, +`CLOBVersion`, +version dispatch |
| `internal/clob/client.go` | +filter `<nil>` in `firstNonEmpty` |
| `go-bot/internal/app/live_wire.go` | +deposit wallet lifecycle, +`ensureDepositWalletFunded` |
| `go-bot/internal/polygolem/client.go` | +`DepositWalletStatus/Fund/Approve`, fix `--json` flags |
