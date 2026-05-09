// Package funding exposes explicit on-chain funding operations needed after
// deposit-wallet onboarding.
//
// These methods submit Polygon mainnet transactions. Callers must enforce
// their own live-mode gates before invoking them.
package funding

import (
	"context"
	"math/big"

	"github.com/TrebuchetDynamics/polygolem/internal/rpc"
)

// TransferPUSD sends pUSD from the EOA derived from privateKeyHex to
// toAddress. amount is in base units with 6 pUSD decimals.
func TransferPUSD(ctx context.Context, privateKeyHex, toAddress string, amount *big.Int, rpcURL string) (string, error) {
	return rpc.TransferPUSD(ctx, privateKeyHex, toAddress, amount, rpcURL)
}
