package relayer

import (
	"context"

	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
)

// DepositWalletCodeDeployed checks Polygon directly for bytecode at a
// deterministic deposit-wallet address. Use this as the source of truth when
// the relayer /deployed endpoint returns a false negative.
//
// Deprecated: use contracts.DepositWalletDeployed for new code.
func DepositWalletCodeDeployed(ctx context.Context, depositWallet string, rpcURL string) (bool, error) {
	status, err := contracts.DepositWalletDeployed(ctx, depositWallet, rpcURL)
	if err != nil {
		return false, err
	}
	return status.Deployed, nil
}
