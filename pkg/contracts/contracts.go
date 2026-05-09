// Package contracts exposes Polymarket Polygon contract addresses and
// contract-level readiness checks.
package contracts

import (
	"context"
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/rpc"
)

const (
	PolygonChainID = 137
	PolygonRPC     = "https://polygon-bor-rpc.publicnode.com"

	DepositWalletFactory = "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"
	ProxyFactory         = "0xaB45c5A4B0c941a2F231C04C3f49182e1A254052"
	GnosisSafeFactory    = "0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b"

	CTFExchangeV2     = "0xE111180000d2663C0091e4f400237545B87B996B"
	NegRiskExchangeV2 = "0xe2222d279d744050d28e00520010520000310F59"
	NegRiskAdapterV2  = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"

	PUSD = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	CTF  = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"
)

// Registry is the Polymarket Polygon contract registry used by polygolem.
type Registry struct {
	ChainID              int    `json:"chainID"`
	DepositWalletFactory string `json:"depositWalletFactory"`
	ProxyFactory         string `json:"proxyFactory"`
	GnosisSafeFactory    string `json:"gnosisSafeFactory"`
	CTFExchangeV2        string `json:"ctfExchangeV2"`
	NegRiskExchangeV2    string `json:"negRiskExchangeV2"`
	NegRiskAdapterV2     string `json:"negRiskAdapterV2"`
	PUSD                 string `json:"pusd"`
	CTF                  string `json:"ctf"`
}

// PolygonMainnet returns the contract registry for Polymarket on Polygon.
func PolygonMainnet() Registry {
	return Registry{
		ChainID:              PolygonChainID,
		DepositWalletFactory: DepositWalletFactory,
		ProxyFactory:         ProxyFactory,
		GnosisSafeFactory:    GnosisSafeFactory,
		CTFExchangeV2:        CTFExchangeV2,
		NegRiskExchangeV2:    NegRiskExchangeV2,
		NegRiskAdapterV2:     NegRiskAdapterV2,
		PUSD:                 PUSD,
		CTF:                  CTF,
	}
}

// DeploymentStatus reports whether a contract address has bytecode on-chain.
type DeploymentStatus struct {
	Address  string `json:"address"`
	Deployed bool   `json:"deployed"`
	Source   string `json:"source"`
}

// ContractDeployed checks Polygon eth_getCode for non-empty bytecode.
func ContractDeployed(ctx context.Context, address string, rpcURL string) (DeploymentStatus, error) {
	address = strings.TrimSpace(address)
	deployed, err := rpc.HasCode(ctx, address, rpcURL)
	if err != nil {
		return DeploymentStatus{}, err
	}
	return DeploymentStatus{
		Address:  address,
		Deployed: deployed,
		Source:   "polygon_eth_getCode",
	}, nil
}

// DepositWalletDeployed checks the deterministic deposit-wallet address on
// Polygon. It is the contract-level source of truth for POLY_1271 readiness.
func DepositWalletDeployed(ctx context.Context, depositWallet string, rpcURL string) (DeploymentStatus, error) {
	status, err := ContractDeployed(ctx, depositWallet, rpcURL)
	if err != nil {
		return DeploymentStatus{}, fmt.Errorf("deposit wallet code check: %w", err)
	}
	return status, nil
}
