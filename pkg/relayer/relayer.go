// Package relayer is the public Go SDK surface for the Polymarket Builder
// Relayer V2 (https://relayer-v2.polymarket.com).
//
// Use this package with builder relayer credentials accepted by the
// Polymarket relayer. CLOB L2 credentials and builder relayer credentials
// are separate auth concepts; callers should not assume they are
// interchangeable unless they have validated that flow for the account.
// The relayer pays gas for deposit-wallet deploy and proxy/batch
// submission, so the only on-chain cost the end user pays in the canonical
// onboarding flow is a single pUSD transfer into the deployed deposit wallet.
//
// The package re-exports types and constructors from internal/relayer and
// internal/auth via type aliases — there is no behavioral wrapper layer,
// just stable import paths for SDK consumers.
//
// Stability: every exported identifier here follows polygolem's public
// SDK semver guarantee.
package relayer

import (
	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	internalrelayer "github.com/TrebuchetDynamics/polygolem/internal/relayer"
)

// Client is the relayer HTTP client. Every request authenticates with
// the POLY_BUILDER_* HMAC-SHA256 header set built from a [BuilderConfig].
type Client = internalrelayer.Client

// BuilderConfig holds the builder relayer {API key, secret, passphrase}
// triple used to authenticate relayer requests.
type BuilderConfig = auth.BuilderConfig

// PrivateKeySigner signs DepositWallet.Batch EIP-712 payloads from a raw
// EOA private key. Use [NewSigner] to construct one.
type PrivateKeySigner = auth.PrivateKeySigner

// RelayerTransaction is the relayer's tracked-transaction record.
type RelayerTransaction = internalrelayer.RelayerTransaction

// RelayerTransactionState is the lifecycle state of a tracked transaction.
type RelayerTransactionState = internalrelayer.RelayerTransactionState

// DepositWalletCall is a single (target, value, data) tuple inside a
// signed wallet batch.
type DepositWalletCall = internalrelayer.DepositWalletCall

// NonceResponse is the wire shape of GET /nonce.
type NonceResponse = internalrelayer.NonceResponse

// DeployedResponse is the wire shape of GET /deployed.
type DeployedResponse = internalrelayer.DeployedResponse

// State constants — re-exported from internal/relayer.
const (
	StateNew       = internalrelayer.StateNew
	StateExecuted  = internalrelayer.StateExecuted
	StateMined     = internalrelayer.StateMined
	StateInvalid   = internalrelayer.StateInvalid
	StateConfirmed = internalrelayer.StateConfirmed
	StateFailed    = internalrelayer.StateFailed
)

// New constructs a relayer Client. baseURL defaults to the production
// relayer when empty; chainID defaults to 137 (Polygon mainnet) when 0.
// Returns an error when [BuilderConfig] is incomplete.
func New(baseURL string, bc BuilderConfig, chainID int64) (*Client, error) {
	return internalrelayer.New(baseURL, bc, chainID)
}

// NewSigner builds a [PrivateKeySigner] from a 0x-prefixed (or unprefixed)
// 64-char hex private key. chainID defaults to 137 when 0.
func NewSigner(privateKeyHex string, chainID int64) (*PrivateKeySigner, error) {
	if chainID == 0 {
		chainID = 137
	}
	return auth.NewPrivateKeySigner(privateKeyHex, chainID)
}

// BuildApprovalCalls returns the six wallet calls a deposit wallet must
// execute to permit trading on V2: pUSD `approve` + CTF
// `setApprovalForAll` for each of the three V2 spenders (CTF Exchange,
// NegRisk Exchange, NegRisk Adapter).
func BuildApprovalCalls() []DepositWalletCall {
	return internalrelayer.BuildApprovalCalls()
}

// BuildDeadline returns a unix-seconds deadline string suitable for use
// in DepositWallet.Batch payloads. Defaults to now+240s when
// secondsFromNow is non-positive (matches the TypeScript relayer client).
func BuildDeadline(secondsFromNow int64) string {
	return internalrelayer.BuildDeadline(secondsFromNow)
}

// SignWalletBatch signs the EIP-712 DepositWallet.Batch payload and
// returns a 0x-prefixed 65-byte ECDSA signature ready for SubmitWalletBatch.
func SignWalletBatch(signer *PrivateKeySigner, walletAddress, nonce, deadline string, calls []DepositWalletCall) (string, error) {
	return internalrelayer.SignWalletBatch(signer, walletAddress, nonce, deadline, calls)
}
