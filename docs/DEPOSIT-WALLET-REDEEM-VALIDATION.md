# Deposit Wallet Redeem Validation

> Date: 2026-05-09
> Status: final source-of-truth for V2 deposit-wallet redeem blockers

This document records the source-of-truth ladder for the Polymarket V2
deposit-wallet redeem flow. Use it before changing redeem code or docs.

## Where We Landed

| Aspect | State |
|---|---|
| Via-EOA path / factory bypass | Removed. Verified impossible at the EVM level: `DepositWalletFactory.proxy(Batch[],bytes[])` is `onlyOperator`; `cast call --from 0x000000000000000000000000000000000000dEaD` reverts with `OnlyOperator()` (`0x27e1f1e5`). |
| Via-CTF path / raw `ConditionalTokens.redeemPositions` | Rejected for deposit-wallet positions. Raw CTF redeem is not the V2 pUSD-native deposit-wallet flow and must not be used as a fallback when relayer adapter calls are blocked. |
| Adapter-via-relayer path | Kept. Polygolem can build and sign adapter approval/redeem WALLET batches, but the production relayer currently rejects adapter calls with HTTP 400 `"not in the allowed list"`. |
| Live recovery | Blocked by relayer policy. The live SOL position had about 2.86 USDC.e of redeemable value; ETH resolved as a loss. There is no code-side workaround while `proxy()` remains operator-gated and adapter calls remain off the relayer allowlist. |
| New artifact | This file is the canonical validation ladder: official docs, deployed ABIs, contract source, RPC assertions, stale-doc inventory, and guardrails. |

## Conclusion

The V2 deposit-wallet redeem action is adapter-based and non-negotiable:

1. Find deposit-wallet positions where the Data API reports `redeemable=true`.
2. Build `redeemPositions(address,bytes32,bytes32,uint256[])` calldata.
3. Target `CtfCollateralAdapter` for standard markets or
   `NegRiskCtfCollateralAdapter` for negative-risk markets.
4. Sign the call set as an EIP-712 WALLET batch.
5. Submit the WALLET batch through Polymarket's relayer/operator path.

There is no safe fallback path for deposit-wallet positions:

- `DepositWalletFactory.deploy(...)` is `onlyOperator`.
- `DepositWalletFactory.proxy(...)` is also `onlyOperator`.
- The wallet implementation execution path is factory-mediated.
- Raw `ConditionalTokens.redeemPositions(...)` is not the pUSD-native V2
  deposit-wallet redeem path.
- SAFE/PROXY relayer examples are different wallet-type flows and do not apply
  to deposit-wallet positions.

If Polymarket's relayer rejects adapter approval or redeem calls as "not in
the allowed list", treat that as an upstream relayer allowlist blocker. Do not
route around it with raw CTF calls.

## Validation Ladder

### 1. Official API Docs

Source: Polymarket API Reference, `GET /positions`.

The official Data API current-position schema includes the fields polygolem
must use as redeem inputs:

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

Observed on 2026-05-09: the response contains `redeemable`, `mergeable`,
`negativeRisk`, `conditionId`, `outcome`, `outcomeIndex`, `oppositeOutcome`,
`oppositeAsset`, and `endDate`.

### 2. Official Polymarket Contract Source

Source: `opensource-projects/repos/ctf-exchange-v2`, remote
`https://github.com/Polymarket/ctf-exchange-v2.git`, main commit
`ccc0596074f4dfd62c944fbca4de252893b82b4b`.

The repo README lists the deployed Polygon V2 collateral adapters:

- `CtfCollateralAdapter`: `0xADa100874d00e3331D00F2007a9c336a65009718`
- `NegRiskCtfCollateralAdapter`: `0xAdA200001000ef00D07553cEE7006808F895c6F1`

The standard adapter source shows `redeemPositions(...)`:

```solidity
function redeemPositions(address, bytes32, bytes32 _conditionId, uint256[] calldata) external onlyUnpaused(USDCE)
```

It reads CTF balances from `msg.sender`, transfers those positions into the
adapter, redeems through legacy CT using USDC.e, wraps proceeds into pUSD, and
sends pUSD back to `msg.sender`. In this flow, `msg.sender` must be the deposit
wallet.

The negative-risk adapter overrides the internal redeem path and uses the
legacy neg-risk adapter with the adapter's current CTF balances.

### 3. ABI From Real Deployed Contracts

Adapter ABIs are available from Sourcify partial matches:

```bash
curl -fsSL \
  'https://repo.sourcify.dev/contracts/partial_match/137/0xADa100874d00e3331D00F2007a9c336a65009718/metadata.json' \
  | jq -r '.output.abi[] | select(.type=="function") | [.name, (.inputs|map(.type)|join(",")), .stateMutability] | @tsv'

curl -fsSL \
  'https://repo.sourcify.dev/contracts/partial_match/137/0xAdA200001000ef00D07553cEE7006808F895c6F1/metadata.json' \
  | jq -r '.output.abi[] | select(.type=="function") | [.name, (.inputs|map(.type)|join(",")), .stateMutability] | @tsv'
```

Both deployed adapter ABIs include:

```text
redeemPositions    address,bytes32,bytes32,uint256[]    nonpayable
mergePositions     address,bytes32,bytes32,uint256[],uint256
splitPosition      address,bytes32,bytes32,uint256[],uint256
```

Factory ABI/source is verified on Polygonscan for
`0x00000000000Fb5C9ADea0298D729A0CB3823Cc07`. The verified source shows:

```solidity
function deploy(address[] calldata _owners, bytes32[] calldata _ids) external onlyOperator
function proxy(Batch[] calldata _batches, bytes[] calldata _signatures) external onlyOperator
```

### 4. Polygon RPC Assertions

Selectors:

```bash
cast sig 'redeemPositions(address,bytes32,bytes32,uint256[])'
# 0x01b7037c

cast sig 'setApprovalForAll(address,bool)'
# 0xa22cb465

cast sig 'proxy((address,uint256,uint256,(address,uint256,bytes)[])[],bytes[])'
# 0x0a3c4405

cast sig 'deploy(address[],bytes32[])'
# 0x77835641
```

Factory implementation:

```bash
cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  0x00000000000Fb5C9ADea0298D729A0CB3823Cc07 \
  'implementation()(address)'
# 0x58CA52ebe0DadfdF531Cde7062e76746de4Db1eB
```

Wallet implementation source at `0x58CA52ebe0DadfdF531Cde7062e76746de4Db1eB`
is also verified on Polygonscan. Its source shows:

```solidity
modifier onlyFactory() {
    require(msg.sender == factory(), OnlyFactory());
    _;
}

function execute(Batch calldata _batch, bytes calldata _signature) external onlyFactory
```

Adapter immutables:

```bash
cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  0xADa100874d00e3331D00F2007a9c336a65009718 \
  'CONDITIONAL_TOKENS()(address)'
# 0x4D97DCd97eC945f40cF65F87097ACe5EA0476045

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  0xADa100874d00e3331D00F2007a9c336a65009718 \
  'COLLATERAL_TOKEN()(address)'
# 0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  0xADa100874d00e3331D00F2007a9c336a65009718 \
  'USDCE()(address)'
# 0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  0xAdA200001000ef00D07553cEE7006808F895c6F1 \
  'NEG_RISK_ADAPTER()(address)'
# 0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  0xAdA200001000ef00D07553cEE7006808F895c6F1 \
  'WRAPPED_COLLATERAL()(address)'
# 0x3A3BD7bb9528E159577F7C2e685CC81A765002E2
```

Factory role-gate proof:

```bash
cast sig 'OnlyOperator()'
# 0x27e1f1e5

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  --from 0x000000000000000000000000000000000000dEaD \
  0x00000000000Fb5C9ADea0298D729A0CB3823Cc07 \
  'proxy((address,uint256,uint256,(address,uint256,bytes)[])[],bytes[])' \
  '[]' '[]'
# execution reverted, data: "0x27e1f1e5"

cast call --rpc-url https://polygon-bor-rpc.publicnode.com \
  --from 0x000000000000000000000000000000000000dEaD \
  0x00000000000Fb5C9ADea0298D729A0CB3823Cc07 \
  'deploy(address[],bytes32[])' \
  '[0x000000000000000000000000000000000000dEaD]' \
  '[0x0000000000000000000000000000000000000000000000000000000000000000]'
# execution reverted, data: "0x27e1f1e5"
```

## Stale Docs Found

| Stale claim | Where found | Correct statement |
|---|---|---|
| Factory `proxy(Batch[],bytes[])` is ungated or permissionless. | `docs/CONTRACTS.md`, docs-site contract page, older planning notes. | `proxy(...)` is `onlyOperator`; direct EOA submission reverts with `OnlyOperator()`. |
| A direct EOA factory `proxy(...)` path can recover adapter approvals or redeem. | Earlier local design notes. | No direct factory fallback exists; the relayer/operator path is required. |
| Raw CTF redeem is the fallback when relayer adapter calls are rejected. | CLI/docs wording before this validation pass. | Raw `ConditionalTokens.redeemPositions` is not the V2 pUSD-native deposit-wallet flow; relayer rejection is an upstream blocker. |
| `builder-relayer-client/examples/redeem.ts` proves raw CTF is a deposit-wallet route. | Official example source, if read without `src/client.ts`. | The example calls `RelayClient.execute(...)`, which switches only between SAFE and PROXY. Deposit wallets use `executeDepositWalletBatch(...)` and the WALLET path. |
| "All redeem code has landed, only runbook remains." | `BLOCKERS.md` B-12. | Code can build adapter batches, but live recovery still depends on relayer allowlist acceptance or an official Polymarket route. |

## Implementation Guardrails

- Keep `pkg/settlement.BuildRedeemCall` targeting V2 collateral adapters.
- Keep pre-submit checks on `CTF.isApprovedForAll(wallet, adapter)`.
- Do not add `--via-eoa` for factory `proxy(...)` unless verified deployed
  source changes and RPC role-gate proof changes.
- Do not submit raw `ConditionalTokens.redeemPositions(...)` as a V2 pUSD
  deposit-wallet redeem fallback.
- Do not add `adapter-pusd`, `ctf-usdce`, or `auto` route switches. The only
  supported deposit-wallet route is the adapter-targeted WALLET batch.
- Do not adapt SAFE/PROXY examples to deposit-wallet positions.
- If relayer allowlist rejects adapter calls, surface a structured upstream
  blocker and stop.

## Upstream Tracker State (2026-05-09)

The exact deposit-wallet redeem blocker is already filed upstream and was
closed without response. Filing duplicate issues is unlikely to help; treat
this as a quiet upstream block, not a code task.

- `Polymarket/builder-relayer-client#29` — *"Deposit Wallet redemption path:
  relayer blocks calls to CtfCollateralAdapter, no documented alternative"*.
  Filed 2026-05-05 by `fylorn`. Closed 2026-05-06 with zero comments.
  Reproduction matches ours exactly: HTTP 400 `"calls to
  0xADa100874d00e3331D00F2007a9c336a65009718 are not permitted"` and
  `"setApprovalForAll operator … is not in the allowed list"` on a `WALLET`
  batch from a deposit wallet.
- `Polymarket/builder-relayer-client#30` — open, server-side precheck bug
  (`PRECHECK_SKIPPED: zero position balance`) for Safe-wallet NegRisk redeem.
  Different wallet type and different rejection path; not adjacent to the
  deposit-wallet allowlist issue.

The official `Polymarket/builder-relayer-client/examples/redeem.ts` covers
SAFE/PROXY wallets only — `RelayClient.execute()` switches on `relayTxType`
between `SAFE` and `PROXY`; deposit wallets use the separate
`executeDepositWalletBatch(...)` path. No deposit-wallet redeem example is
shipped, which is consistent with the `WALLET`-batch-only adapter route
being the intended deposit-wallet redeem path.

The relayer allowlist itself is not in any public Polymarket repo
(`builder-relayer-client`, `py-builder-relayer-client`, `relayer-deposits`,
and `ctf-exchange-v2` were all checked); there is no upstream PR target.

## Regression Tests

- `pkg/settlement`: unit coverage for redeemable filtering, adapter target
  selection, calldata encoding, condition de-duplication, batch limits, and
  successful relayer submission.
- `internal/cli.TestDepositWalletRedeemHelpRejectsDirectCTFFallback`: CLI
  long-help must not advertise direct CTF as a fallback.
- `tests.TestPolygolemSettlementE2EStopsAtRelayerAllowlistBlocker`: public SDK
  e2e coverage for Data API `redeemable=true`, adapter-only WALLET calls,
  no raw CTF target, and structured stop on relayer `"not in the allowed list"`.

## Sources

- Polymarket API Reference: https://docs.polymarket.com/api-reference/core/get-current-positions-for-a-user
- Polymarket CTF Exchange V2: https://github.com/Polymarket/ctf-exchange-v2
- CtfCollateralAdapter source: `opensource-projects/repos/ctf-exchange-v2/src/adapters/CtfCollateralAdapter.sol`
- NegRiskCtfCollateralAdapter source: `opensource-projects/repos/ctf-exchange-v2/src/adapters/NegRiskCtfCollateralAdapter.sol`
- Sourcify adapter metadata:
  `https://repo.sourcify.dev/contracts/partial_match/137/0xADa100874d00e3331D00F2007a9c336a65009718/metadata.json`
- Sourcify neg-risk adapter metadata:
  `https://repo.sourcify.dev/contracts/partial_match/137/0xAdA200001000ef00D07553cEE7006808F895c6F1/metadata.json`
- Polygonscan factory source:
  `https://polygonscan.com/address/0x00000000000Fb5C9ADea0298D729A0CB3823Cc07#code`
- Polygonscan deposit wallet implementation source:
  `https://polygonscan.com/address/0x58CA52ebe0DadfdF531Cde7062e76746de4Db1eB#code`
