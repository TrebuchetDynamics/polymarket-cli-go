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

	// V2 collateral adapters — split/merge/redeem from a deposit wallet
	// must route through these. The adapter pulls the wallet's CTF
	// position tokens, executes the underlying CT call with USDC.e, then
	// wraps the proceeds back into pUSD and sends pUSD to the wallet.
	// Source: https://docs.polymarket.com/developers/CTF/deployment-resources
	CtfCollateralAdapter        = "0xAdA100Db00Ca00073811820692005400218FcE1f"
	NegRiskCtfCollateralAdapter = "0xadA2005600Dec949baf300f4C6120000bDB6eAab"

	// V2 collateral ramps — convert between USDC/USDC.e and pUSD.
	CollateralOnramp  = "0x93070a847efEf7F70739046A929D47a521F5B8ee"
	CollateralOfframp = "0x2957922Eb93258b93368531d39fAcCA3B4dC5854"
	PermissionedRamp  = "0xebC2459Ec962869ca4c0bd1E06368272732BCb08"

	PUSD = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	CTF  = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"
	// USDC.e on Polygon. Polymarket's UI Enable Trading approval batch
	// approves this token to the V2 CollateralOnramp so the wallet can
	// route legacy collateral into pUSD when needed.
	USDCE = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
)

// Registry is the Polymarket Polygon contract registry used by polygolem.
type Registry struct {
	ChainID                     int    `json:"chainID"`
	DepositWalletFactory        string `json:"depositWalletFactory"`
	ProxyFactory                string `json:"proxyFactory"`
	GnosisSafeFactory           string `json:"gnosisSafeFactory"`
	CTFExchangeV2               string `json:"ctfExchangeV2"`
	NegRiskExchangeV2           string `json:"negRiskExchangeV2"`
	NegRiskAdapterV2            string `json:"negRiskAdapterV2"`
	CtfCollateralAdapter        string `json:"ctfCollateralAdapter"`
	NegRiskCtfCollateralAdapter string `json:"negRiskCtfCollateralAdapter"`
	CollateralOnramp            string `json:"collateralOnramp"`
	CollateralOfframp           string `json:"collateralOfframp"`
	PermissionedRamp            string `json:"permissionedRamp"`
	PUSD                        string `json:"pusd"`
	CTF                         string `json:"ctf"`
	USDCE                       string `json:"usdce"`
}

// PolygonMainnet returns the contract registry for Polymarket on Polygon.
func PolygonMainnet() Registry {
	return Registry{
		ChainID:                     PolygonChainID,
		DepositWalletFactory:        DepositWalletFactory,
		ProxyFactory:                ProxyFactory,
		GnosisSafeFactory:           GnosisSafeFactory,
		CTFExchangeV2:               CTFExchangeV2,
		NegRiskExchangeV2:           NegRiskExchangeV2,
		NegRiskAdapterV2:            NegRiskAdapterV2,
		CtfCollateralAdapter:        CtfCollateralAdapter,
		NegRiskCtfCollateralAdapter: NegRiskCtfCollateralAdapter,
		CollateralOnramp:            CollateralOnramp,
		CollateralOfframp:           CollateralOfframp,
		PermissionedRamp:            PermissionedRamp,
		PUSD:                        PUSD,
		CTF:                         CTF,
		USDCE:                       USDCE,
	}
}

// RedeemAdapterFor returns the V2 collateral adapter address that a
// deposit wallet must call redeemPositions on for a given market kind.
// The adapter pulls the wallet's CTF tokens, redeems through legacy
// ConditionalTokens with USDC.e, wraps proceeds into pUSD, and sends
// pUSD to the wallet. The adapter ignores the caller-supplied
// collateralToken arg, parentCollectionId arg, and indexSets array;
// only conditionId is used.
func RedeemAdapterFor(negRisk bool) string {
	if negRisk {
		return NegRiskCtfCollateralAdapter
	}
	return CtfCollateralAdapter
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
