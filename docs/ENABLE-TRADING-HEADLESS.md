# Headless Enable Trading

**Last updated:** 2026-05-10

Polymarket's web UI can show two remaining confirmations after the deposit
wallet is deployed:

1. **Enable Trading** - sign a ClobAuth EIP-712 message to create or derive
   CLOB L2 API keys.
2. **Approve Tokens** - sign a `DepositWallet.Batch` EIP-712 message that
   approves token spenders used by the UI enable-trading path.

Polygolem exposes these as SDK primitives in `pkg/enabletrading`.

The CLI integration is `polygolem deposit-wallet onboard`. After deploy and
the trading/adapter approval batch, onboarding signs ClobAuth, creates or
derives CLOB API keys, and submits the 2-call UI token approval batch. Skip it
only with `--skip-enable-trading`.

For an already-deployed wallet, use:

```bash
polygolem deposit-wallet enable-trading
```

If no V2 relayer key is configured, this command runs the SIWE login/profile
registration/key-mint sequence automatically, persists the relayer key to the
Polygolem env file, and then submits the approval batch. No browser or mobile
wallet confirmation is required for Polygolem's headless path.

Polymarket's website may still ask the browser to sign ClobAuth because the
site keeps browser-local API-key state. That is separate from Polygolem's
headless readiness. `deposit-wallet status --check-enable-trading` is the
CLI validation source of truth.

## Identity Model

Polymarket login signs with the EOA. The deposit wallet remains the trading
wallet.

| Identity | Role |
|---|---|
| EOA | Signs SIWE login, ClobAuth HTTP auth, relayer WALLET batches, and POLY_1271 order wrappers. |
| Deposit wallet | Holds pUSD, is the CLOB order maker/signer, receives CTF positions, and executes token approvals. |

The ClobAuth payload observed in the UI uses the EOA address. The
DepositWallet approval payload uses the deposit wallet as both
`domain.verifyingContract` and `message.wallet`.

## SDK Example

```go
package main

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/pkg/enabletrading"
)

func main() {
	ctx := context.Background()

	result, err := enabletrading.EnableTradingHeadless(ctx, enabletrading.EnableTradingParams{
		OwnerPrivateKey:       "0x...",
		DepositWalletAddress:  "0x...",
		CreateOrDeriveCLOBKey: true,
		ApproveTokens:         true,
		MaxApproval:           true,
		DryRun:                true,
	})
	if err != nil {
		panic(err)
	}

	_ = result.ClobAuthTypedData
	_ = result.ApprovalBatchTypedData
}
```

`DryRun` builds the exact typed data and planned actions without signing,
creating API keys, or submitting relayer batches. Use it for audits and tests.

## CLI Validation

```bash
polygolem deposit-wallet status --check-enable-trading
```

The validation path signs local throwaway ClobAuth and DepositWallet typed
data, validates the configured approval targets, verifies that CLOB
credentials are either configured or derivable via `/auth/derive-api-key`, and
checks the two on-chain ERC-20 allowances:

- pUSD allowance from deposit wallet to CTF.
- USDC.e allowance from deposit wallet to CollateralOnramp.

It does not print signatures or secrets.

## Lower-Level Primitives

```go
clobTD, err := enabletrading.BuildClobAuthTypedData(enabletrading.ClobAuthParams{
	Address:   eoa,
	ChainID:   137,
	Timestamp: "1778372101",
	Nonce:     0,
})

sig, err := enabletrading.SignClobAuthTypedData(privateKey, clobTD)

calls := enabletrading.BuildEnableTradingApprovalCalls()

batchTD, err := enabletrading.BuildEnableTradingApprovalBatchTypedData(enabletrading.ApprovalBatchParams{
	DepositWallet: depositWallet,
	ChainID:       137,
	Nonce:         walletNonce,
	Deadline:      deadline,
	Calls:         calls,
})

batchSig, err := enabletrading.SignDepositWalletApprovalBatch(privateKey, batchTD)
```

## Observed Approval Batch

The UI approval batch currently contains two ERC-20 `approve` calls:

| Token | Spender | Amount |
|---|---|---|
| pUSD `0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB` | CTF `0x4D97DCd97eC945f40cF65F87097ACe5EA0476045` | max uint256 |
| USDC.e `0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174` | CollateralOnramp `0x93070a847efEf7F70739046A929D47a521F5B8ee` | max uint256 |

This UI enable-trading batch is separate from:

- `deposit-wallet approve` - exchange trading approvals.
- `deposit-wallet approve-adapters` - split, merge, and redeem adapter approvals.

## Safety Rules

- `MaxApproval` must be explicit.
- Chain ID must be Polygon mainnet `137`.
- ClobAuth message text must exactly match Polymarket's control message.
- DepositWallet `verifyingContract` must equal `message.wallet`.
- Unknown approval targets or spenders fail closed.
- Private keys, API secrets, passphrases, and full signatures must not be
  printed or stored in logs.

## Current Scope

The SDK builds and signs the observed ClobAuth and approval-batch payloads,
creates or derives CLOB credentials, and submits the approval batch through
the V2 relayer when not in dry-run mode. The CLI wires those same signs into
`deposit-wallet onboard`, provides `deposit-wallet enable-trading` for already
deployed wallets, and exposes readiness validation with
`deposit-wallet status --check-enable-trading`.
