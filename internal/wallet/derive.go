// Package wallet re-exports pkg/wallet as a compatibility shim.
// Deprecated: use pkg/wallet directly. This shim will be removed in v0.3.0.
package wallet

import "github.com/TrebuchetDynamics/polygolem/pkg/wallet"

var (
	DeriveProxyWallet = wallet.DeriveProxyWallet
	DeriveSafeWallet  = wallet.DeriveSafeWallet
	Readiness         = wallet.Readiness
)

type ReadyInfo = wallet.ReadyInfo
