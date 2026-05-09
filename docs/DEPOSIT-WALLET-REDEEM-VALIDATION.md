# Deposit Wallet Settlement Validation

> Date: 2026-05-09
> Status: resolved source of truth for V2 deposit-wallet settlement

This document records the source-of-truth ladder for Polymarket V2
deposit-wallet settlement. Use it before changing redeem code, adapter
constants, operator runbooks, or live-bot readiness gates.

## Executive Summary

The 2026-05-09 live settlement blocker was not a permanent relayer limitation.
The root cause was stale V2 collateral adapter constants in Polygolem. The
relayer correctly rejected the stale adapter addresses as not allowlisted.

After updating Polygolem to the current Polymarket contract registry, adapter
approvals and redeem both executed through the production deposit-wallet
relayer path.

Resolved live evidence:

| Check | Result |
|---|---|
| Official standard adapter | `0xAdA100Db00Ca00073811820692005400218FcE1f` |
| Official negative-risk adapter | `0xadA2005600Dec949baf300f4C6120000bDB6eAab` |
| Standard adapter CTF approval probe | executed, relayer tx `019e0ed9-2948-7764-ad69-d0e97cb70e6c` |
| Negative-risk adapter CTF approval probe | executed, relayer tx `019e0ed9-4985-7759-9aeb-b49784a514f5` |
| Full adapter approval batch | executed, relayer tx `019e0edb-8c8d-76e8-9d3c-8e0f0d251581` |
| Redeem batch | executed, relayer tx `019e0eda-a012-75d3-9aee-6c2c3dc61450` |
| Post-redeem readiness | `settlement-status.ready=true`, `redeemableCount=0` |
| Post-redeem CLOB collateral | `2.899475` pUSD |

## Current Production Invariants

The V2 deposit-wallet settlement path is adapter-based and non-negotiable:

1. Query Data API positions for the deposit wallet.
2. Treat `redeemable=true` as the redemption signal.
3. Build `redeemPositions(address,bytes32,bytes32,uint256[])` calldata.
4. Target `CtfCollateralAdapter` for standard markets or
   `NegRiskCtfCollateralAdapter` for negative-risk markets.
5. Sign the calls as an EIP-712 deposit-wallet `WALLET` batch.
6. Submit through Polymarket's relayer/operator path.
7. Confirm proceeds return as pUSD to the deposit wallet.

There is no supported fallback for deposit-wallet positions:

- No direct EOA factory `proxy(...)` or `deploy(...)` path.
- No raw `ConditionalTokens.redeemPositions(...)` fallback.
- No SAFE/PROXY relayer shortcut.
- No local "try another wallet type" switch.

Polymarket's current contract reference is the first source of truth for
addresses:

- Contracts: <https://docs.polymarket.com/resources/contracts>
- Deposit wallets: <https://docs.polymarket.com/trading/deposit-wallets>

The contracts page states that all listed Polymarket contracts are on Polygon
mainnet, chain ID 137, and lists these current V2 collateral adapters:

- `CtfCollateralAdapter`: `0xAdA100Db00Ca00073811820692005400218FcE1f`
- `NegRiskCtfCollateralAdapter`: `0xadA2005600Dec949baf300f4C6120000bDB6eAab`

## Live Incident: Stale Adapter Constants

The original symptom was `missing_adapter_approval` from:

```bash
go-bot settlement-status --json
```

The first hypothesis was "the relayer does not allow deposit-wallet adapter
approval or redeem calls." That was wrong. The relayer was rejecting the
specific stale adapter addresses we were sending.

The scientific method loop:

| Step | Experiment | Observation | Conclusion |
|---|---|---|---|
| 1 | Dry-run `approve-adapters` | Calls targeted stale adapter addresses | Local registry was suspect |
| 2 | Submit CTF approval to stale standard adapter | HTTP 400, operator not in allowed list | Relayer rejects stale target |
| 3 | Submit trading CTF approval to `CTFExchangeV2` | Executed | Relayer credentials and WALLET path work |
| 4 | Check official Polymarket contract docs | Adapters differed from local constants | Constants were stale |
| 5 | Submit CTF approval to official standard adapter | Executed | Official adapter is allowlisted |
| 6 | Submit CTF approval to official negative-risk adapter | Executed | Official neg-risk adapter is allowlisted |
| 7 | Run full 4-call `approve-adapters` | Executed | pUSD + CTF adapter readiness complete |
| 8 | Run `deposit-wallet redeem --submit` | Executed | Live winners redeemed through adapter path |
| 9 | Re-run settlement checks | `ready=true`, `redeemableCount=0` | Recovery complete |

The durable lesson: `RELAYER_ALLOWLIST_BLOCKED` does not always mean
"Polymarket must change something." First verify that Polygolem is targeting
the current official contract addresses. If the addresses are current and the
relayer still rejects the batch, then stop and escalate as a real upstream
allowlist blocker.

## Operator Runbook

Use this sequence for live settlement triage.

### 1. Check Readiness

```bash
polygolem deposit-wallet settlement-status --json
```

Expected ready state:

- `walletDeployed=true`
- `relayerCredentials=true`
- `positionsReachable=true`
- `standardAdapterApproved=true`
- `negRiskAdapterApproved=true`
- `ready=true`

If adapter approval is missing, continue.

### 2. Verify Contract Constants

Compare local constants to the official Polymarket contracts page:

```bash
rg -n "CtfCollateralAdapter|NegRiskCtfCollateralAdapter" pkg internal docs
```

Current values must be:

```text
CtfCollateralAdapter        0xAdA100Db00Ca00073811820692005400218FcE1f
NegRiskCtfCollateralAdapter 0xadA2005600Dec949baf300f4C6120000bDB6eAab
```

If they differ, update Polygolem before sending any approval or redeem batch.

### 3. Approve Adapters

Dry-run first:

```bash
polygolem deposit-wallet approve-adapters --json
```

Submit only after reviewing the target addresses:

```bash
polygolem deposit-wallet approve-adapters \
  --submit \
  --confirm APPROVE_ADAPTERS \
  --json
```

This sends four idempotent calls:

1. pUSD `approve(CtfCollateralAdapter, MaxUint256)`
2. CTF `setApprovalForAll(CtfCollateralAdapter, true)`
3. pUSD `approve(NegRiskCtfCollateralAdapter, MaxUint256)`
4. CTF `setApprovalForAll(NegRiskCtfCollateralAdapter, true)`

### 4. Inspect Redeemable Positions

```bash
polygolem deposit-wallet redeemable --json
```

`redeemable=true` is the signal. The Data API does not expose a separate
`resolved` boolean for this flow.

### 5. Dry-Run Redeem

```bash
polygolem deposit-wallet redeem --json
```

Confirm every call target is one of the two official adapters.

### 6. Submit Redeem

```bash
polygolem deposit-wallet redeem \
  --submit \
  --confirm REDEEM_WINNERS \
  --json
```

The command fails closed before signing if
`CTF.isApprovedForAll(wallet, adapter)` is false.

### 7. Post-Checks

```bash
polygolem deposit-wallet settlement-status --json
polygolem deposit-wallet redeemable --json
polygolem clob update-balance --asset-type collateral --json
polygolem clob balance --asset-type collateral --json
```

Expected after a successful redeem:

- `redeemableCount=0` for the redeemed conditions.
- CLOB collateral reflects the returned pUSD after balance sync.
- No new live orders are placed unless `settlement-status.ready=true`.

## Validation Ladder

### Official API Docs

Source: Polymarket Data API current positions.

Fields Polygolem uses for settlement:

- `asset`
- `conditionId`
- `redeemable`
- `mergeable`
- `negativeRisk`
- `outcome`
- `outcomeIndex`
- `oppositeOutcome`
- `oppositeAsset`
- `endDate`

Live schema check:

```bash
curl -fsSL \
  'https://data-api.polymarket.com/positions?user=0x21999a074344610057c9b2B362332388a44502D4&limit=1&sizeThreshold=0' \
  | jq 'if length>0 then .[0] | keys else [] end'
```

### Official Contract Docs

Source: Polymarket contracts reference:
<https://docs.polymarket.com/resources/contracts>

The page lists:

- pUSD proxy: `0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB`
- Conditional Tokens: `0x4D97DCd97eC945f40cF65F87097ACe5EA0476045`
- CTF Exchange V2: `0xE111180000d2663C0091e4f400237545B87B996B`
- Negative Risk Exchange V2: `0xe2222d279d744050d28e00520010520000310F59`
- CTF collateral adapter: `0xAdA100Db00Ca00073811820692005400218FcE1f`
- Negative-risk CTF collateral adapter: `0xadA2005600Dec949baf300f4C6120000bDB6eAab`
- Deposit Wallet Factory: `0x00000000000Fb5C9ADea0298D729A0CB3823Cc07`

### Deployed ABI Checks

Adapter ABIs can be checked through Sourcify partial matches:

```bash
curl -fsSL \
  'https://repo.sourcify.dev/contracts/partial_match/137/0xAdA100Db00Ca00073811820692005400218FcE1f/metadata.json' \
  | jq -r '.output.abi[] | select(.type=="function") | [.name, (.inputs|map(.type)|join(",")), .stateMutability] | @tsv'

curl -fsSL \
  'https://repo.sourcify.dev/contracts/partial_match/137/0xadA2005600Dec949baf300f4C6120000bDB6eAab/metadata.json' \
  | jq -r '.output.abi[] | select(.type=="function") | [.name, (.inputs|map(.type)|join(",")), .stateMutability] | @tsv'
```

Both adapter ABIs include:

```text
redeemPositions    address,bytes32,bytes32,uint256[]    nonpayable
mergePositions     address,bytes32,bytes32,uint256[],uint256
splitPosition      address,bytes32,bytes32,uint256[],uint256
```

### Factory Role-Gate Proof

Factory ABI/source is verified on Polygonscan for:

```text
0x00000000000Fb5C9ADea0298D729A0CB3823Cc07
```

The verified source shows both deployment and proxy execution are operator
gated:

```solidity
function deploy(address[] calldata _owners, bytes32[] calldata _ids) external onlyOperator
function proxy(Batch[] calldata _batches, bytes[] calldata _signatures) external onlyOperator
```

Direct EOA attempts revert with `OnlyOperator()`:

```bash
cast sig 'OnlyOperator()'
# 0x27e1f1e5

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  --from 0x000000000000000000000000000000000000dEaD \
  0x00000000000Fb5C9ADea0298D729A0CB3823Cc07 \
  'proxy((address,uint256,uint256,(address,uint256,bytes)[])[],bytes[])' \
  '[]' '[]'
# execution reverted, data: "0x27e1f1e5"
```

This is why Polygolem removed the direct EOA deploy/proxy recovery path from
the operator surface.

## Removed Claims

| Deprecated claim | Correct statement |
|---|---|
| Stale adapter addresses are V2 production targets. | Use the current Polymarket contracts reference before every registry update. |
| `RELAYER_ALLOWLIST_BLOCKED` always means an upstream relayer bug. | First verify target addresses. Stale local constants can cause the same error. |
| Direct EOA factory `deploy(...)` or `proxy(...)` can recover a wallet. | The production factory is `onlyOperator`; EOA calls revert. |
| Raw `ConditionalTokens.redeemPositions(...)` can recover V2 deposit-wallet positions. | V2 deposit-wallet settlement routes through collateral adapters and returns pUSD. |
| SAFE/PROXY relayer examples apply to deposit wallets. | Deposit wallets use `executeDepositWalletBatch(...)` / WALLET batches. |

## Implementation Guardrails

- Keep `pkg/contracts` pinned to the official Polymarket contracts reference.
- Keep tests that fail if adapter constants drift back to stale addresses.
- Keep `pkg/settlement.BuildRedeemCall` targeting V2 collateral adapters.
- Keep `CTF.isApprovedForAll(wallet, adapter)` as a pre-submit check.
- Keep `settlement-status` as a live-bot readiness gate.
- Do not reintroduce direct EOA deploy/proxy commands.
- Do not add raw CTF, SAFE, PROXY, or "auto route" fallback switches for
  deposit-wallet settlement.
- If the relayer rejects a batch after the target addresses have been verified
  against the official docs, surface `RELAYER_ALLOWLIST_BLOCKED` and stop.

## Regression Coverage

Required test coverage:

- `pkg/contracts`: official adapter constants and `RedeemAdapterFor`.
- `internal/relayer`: allowlist rejection classification.
- `internal/relayer`: 4-call adapter approval batch shape.
- `pkg/settlement`: binary and negative-risk redeem call targets.
- `internal/cli`: redeem help rejects direct EOA/raw CTF fallback thinking.
- `internal/cli`: structured allowlist error has no deprecated issue tracker.
- `pkg/settlement`: readiness statuses for deployed wallet, reachable Data API,
  relayer credentials, and adapter approvals.

## Sources

- Polymarket Contracts: <https://docs.polymarket.com/resources/contracts>
- Polymarket Deposit Wallets: <https://docs.polymarket.com/trading/deposit-wallets>
- Polymarket Data API positions: <https://data-api.polymarket.com/positions>
- Sourcify adapter metadata:
  `https://repo.sourcify.dev/contracts/partial_match/137/<adapter>/metadata.json`
- Polygon mainnet RPC assertions against `0x00000000000Fb5C9ADea0298D729A0CB3823Cc07`
