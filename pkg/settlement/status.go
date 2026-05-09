package settlement

import (
	"context"
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/rpc"
	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	"github.com/TrebuchetDynamics/polygolem/pkg/data"
)

const (
	StatusReady                    = "ready"
	StatusDepositWalletNotDeployed = "deposit_wallet_not_deployed"
	StatusMissingRelayerCreds      = "missing_relayer_credentials"
	StatusMissingAdapterApproval   = "missing_adapter_approval"
	StatusDataAPIUnavailable       = "data_api_unavailable"
	StatusRPCError                 = "rpc_error"
)

// ReadinessOptions configures the read-only settlement readiness check.
type ReadinessOptions struct {
	// RPCURL is the Polygon JSON-RPC endpoint used for eth_getCode and
	// CTF.isApprovedForAll checks. Empty uses the package default.
	RPCURL string
	// RelayerConfigured tells the checker whether the caller has usable
	// Polymarket relayer credentials loaded. The checker does not read env.
	RelayerConfigured bool
}

// AdapterApproval reports one required V2 collateral adapter approval.
type AdapterApproval struct {
	Adapter  string `json:"adapter"`
	Approved bool   `json:"approved"`
}

// Readiness is the read-only settlement gate result. Ready means the deposit
// wallet is deployed, relayer credentials are configured, the Data API can be
// queried, and both V2 collateral adapters are approved for CTF movement.
type Readiness struct {
	Ready                 bool                 `json:"ready"`
	Status                string               `json:"status"`
	Owner                 string               `json:"owner,omitempty"`
	DepositWallet         string               `json:"depositWallet"`
	DepositWalletDeployed bool                 `json:"depositWalletDeployed"`
	RelayerConfigured     bool                 `json:"relayerConfigured"`
	RequiredAdapters      []string             `json:"requiredAdapters"`
	AdapterApprovals      []AdapterApproval    `json:"adapterApprovals"`
	MissingApprovals      []string             `json:"missingApprovals,omitempty"`
	RedeemableCount       int                  `json:"redeemableCount"`
	RedeemablePositions   []RedeemablePosition `json:"redeemablePositions,omitempty"`
	Reason                string               `json:"reason,omitempty"`
	NextAction            string               `json:"nextAction,omitempty"`
}

// RequiredAdapters returns the V2 collateral adapters a deposit wallet must
// approve before split/merge/redeem can work in both binary and neg-risk
// markets.
func RequiredAdapters() []string {
	return []string{contracts.CtfCollateralAdapter, contracts.NegRiskCtfCollateralAdapter}
}

// CheckReadiness performs the settlement readiness gate without signing or
// submitting anything. dataClient may be nil to skip the Data API probe; live
// trading callers should pass one so redeemable detection is proven too.
func CheckReadiness(ctx context.Context, dataClient *data.Client, owner string, depositWallet string, opts ReadinessOptions) (*Readiness, error) {
	depositWallet = strings.TrimSpace(depositWallet)
	if depositWallet == "" {
		return nil, fmt.Errorf("settlement: deposit wallet is required")
	}
	out := &Readiness{
		Status:            StatusReady,
		Owner:             strings.TrimSpace(owner),
		DepositWallet:     depositWallet,
		RelayerConfigured: opts.RelayerConfigured,
		RequiredAdapters:  RequiredAdapters(),
	}

	codeStatus, err := contracts.DepositWalletDeployed(ctx, depositWallet, opts.RPCURL)
	if err != nil {
		out.Status = StatusRPCError
		out.Reason = err.Error()
		out.NextAction = "fix POLYGON_RPC_URL and retry `polygolem deposit-wallet settlement-status`"
		return out, nil
	}
	out.DepositWalletDeployed = codeStatus.Deployed
	if !codeStatus.Deployed {
		out.Status = StatusDepositWalletNotDeployed
		out.Reason = "deposit wallet has no bytecode on Polygon"
		out.NextAction = "polygolem deposit-wallet deploy --wait"
		return out, nil
	}

	if dataClient != nil {
		rows, err := FindRedeemable(ctx, dataClient, depositWallet)
		if err != nil {
			out.Status = StatusDataAPIUnavailable
			out.Reason = err.Error()
			out.NextAction = "restore Data API access before live trading or winner redemption"
			return out, nil
		}
		out.RedeemablePositions = rows
		out.RedeemableCount = len(rows)
	}

	for _, adapter := range out.RequiredAdapters {
		approved, err := rpc.IsApprovedForAll(ctx, contracts.CTF, depositWallet, adapter, opts.RPCURL)
		if err != nil {
			out.Status = StatusRPCError
			out.Reason = fmt.Sprintf("check CTF.isApprovedForAll(%s): %v", adapter, err)
			out.NextAction = "fix POLYGON_RPC_URL and retry `polygolem deposit-wallet settlement-status`"
			return out, nil
		}
		out.AdapterApprovals = append(out.AdapterApprovals, AdapterApproval{
			Adapter:  adapter,
			Approved: approved,
		})
		if !approved {
			out.MissingApprovals = append(out.MissingApprovals, adapter)
		}
	}

	if !opts.RelayerConfigured {
		out.Status = StatusMissingRelayerCreds
		out.Reason = "Polymarket relayer credentials are not configured"
		out.NextAction = "add RELAYER_API_KEY and RELAYER_API_KEY_ADDRESS to go-bot/.env.relayer-v2"
		return out, nil
	}
	if len(out.MissingApprovals) > 0 {
		out.Status = StatusMissingAdapterApproval
		out.Reason = "deposit wallet has not approved every V2 collateral adapter required for settlement"
		out.NextAction = "polygolem deposit-wallet approve-adapters --submit --confirm APPROVE_ADAPTERS"
		return out, nil
	}

	out.Ready = true
	out.Status = StatusReady
	return out, nil
}
