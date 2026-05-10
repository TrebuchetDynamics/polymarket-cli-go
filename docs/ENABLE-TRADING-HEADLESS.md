# Headless Polymarket Enable Trading

Polygolem exposes a headless compatibility layer for Polymarket's web UI
"Enable Trading" flow. It is intended for server/CLI automation where the
operator controls an imported/generated EOA private key and does not want a
browser wallet popup.

## UI flow covered

Polymarket shows three onboarding rows:

1. Deploy Wallet — deploy a smart contract deposit wallet to enable trading.
2. Enable Trading — sign a `ClobAuth` EIP-712 message to create or recover CLOB API keys.
3. Approve Tokens — sign and submit a `DepositWallet.Batch` approval payload.

The SDK helpers make those payloads explicit and testable.

## Public package

```go
import "github.com/TrebuchetDynamics/polygolem/pkg/enabletrading"
```

### Dry-run example

Dry-run builds the CLOB auth typed data and the approval batch typed data, but
does not sign, submit, or create credentials.

```go
result, err := enabletrading.EnableTradingHeadless(ctx, enabletrading.EnableTradingParams{
    OwnerPrivateKey:       os.Getenv("POLYMARKET_PRIVATE_KEY"),
    DepositWalletAddress:  "0xYourDepositWallet",
    CreateOrDeriveCLOBKey: true,
    ApproveTokens:         true,
    MaxApproval:           true,
    DryRun:                true,
    ClobAuthTimestamp:     "1778372101",
    WalletNonce:           "6",
    ApprovalDeadline:      "1778373936",
})
if err != nil {
    return err
}

// Safe to inspect typed-data shape. Do not log private keys or live credentials.
fmt.Println(result.PlannedActions)
fmt.Printf("CLOB primary type: %s\n", result.ClobAuthTypedData.PrimaryType)
fmt.Printf("Approval primary type: %s\n", result.ApprovalBatchTypedData.PrimaryType)
```

### Low-level helpers

```go
clobTD, err := enabletrading.BuildClobAuthTypedData(enabletrading.ClobAuthParams{
    Address:   ownerEOA,
    ChainID:   137,
    Timestamp: "1778372101",
    Nonce:     0,
})

sig, err := enabletrading.SignClobAuthTypedData(privateKey, clobTD)

calls := enabletrading.BuildEnableTradingApprovalCalls()
approvalTD, err := enabletrading.BuildEnableTradingApprovalBatchTypedData(enabletrading.ApprovalBatchParams{
    DepositWallet: depositWallet,
    ChainID:       137,
    Nonce:         "6",
    Deadline:      "1778373936",
    Calls:         calls,
})
```

## Safety behavior

The package fails closed on:

- non-Polygon chain IDs; only chain ID `137` is accepted,
- empty CLOB auth address or timestamp,
- CLOB auth messages that do not match Polymarket's expected control text,
- empty deposit wallet address, nonce, deadline, or approval call list,
- approval calls whose targets are not the observed Polymarket Enable Trading token targets,
- approval calls that are not ERC-20 `approve(spender, maxUint256)` payloads.

`MaxApproval` must be explicit when requesting token approvals through the
high-level dry-run flow. This is intentional because the observed UI payload
uses max-uint allowances.

## Secret handling

Never log or commit:

- private keys,
- API keys,
- API secrets,
- API passphrases,
- bearer/session tokens,
- seed phrases.

The typed-data helpers do not need secrets. Signing helpers accept the private
key only long enough to produce a signature.

## Current implementation notes

- Existing Polygolem internals already include CLOB L1 key creation/derivation,
  deterministic deposit-wallet derivation, relayer WALLET-CREATE submission,
  WALLET nonce fetching, and DepositWallet.Batch signing/submission.
- `pkg/enabletrading` currently exposes the stable public typed-data and dry-run
  surface. Live submission remains composed through `pkg/relayer`, `pkg/clob`,
  and the internal orchestrator until the exact production relayer allowance
  readiness endpoint is pinned in a public API.
- The observed web UI approval calls are intentionally versioned in code and
  tests. If Polymarket changes targets or spenders, update the named constants
  and tests together rather than accepting arbitrary targets silently.

## Troubleshooting

- Wrong chain: ensure signer and payload chain ID are Polygon mainnet `137`.
- Undeployed wallet: use `pkg/relayer.OnboardDepositWallet` or relayer
  `SubmitWalletCreate` before submitting wallet batches.
- Rejected CLOB auth: verify the `ClobAuth` message text and timestamp/nonce.
- Expired deadline: rebuild the approval batch with a fresh deadline.
- Approval target rejected: verify Polymarket's current token/spender registry
  before changing allowlisted targets.
- API key creation failure: try derive existing API key first; if creation still
  fails, inspect the CLOB `/auth/api-key` response without logging secrets.
