package relayer

import (
	"errors"
	"strings"
)

// ErrRelayerAllowlistBlocked is the sentinel error wrapped by relayer
// submission methods when Polymarket's relayer rejects a WALLET batch
// because one of the targeted contracts is not on its allowlist.
//
// Callers should detect this with errors.Is, verify that local contract
// constants match Polymarket's current contract reference, and stop if they
// do. The V2 deposit-wallet path is non-negotiable: signed EIP-712 WALLET
// batch -> CtfCollateralAdapter / NegRiskCtfCollateralAdapter -> relayer
// executeDepositWalletBatch. There is no safe direct EOA bypass and no raw
// ConditionalTokens fallback.
var ErrRelayerAllowlistBlocked = errors.New("relayer: allowlist block")

// allowlistRejectionMarkers are the case-insensitive substrings the
// Polymarket relayer returns in HTTP 400 bodies when it refuses a call
// in a WALLET batch:
//   - "calls to 0x… are not permitted"
//   - "setApprovalForAll operator 0x… is not in the allowed list"
//   - "call blocked: call[i] blocked: …"
var allowlistRejectionMarkers = []string{
	"not in the allowed list",
	"are not permitted",
	"call blocked",
}

// classifyAllowlistError wraps err with ErrRelayerAllowlistBlocked when
// the error string matches a known allowlist rejection. Returns err
// unchanged otherwise.
func classifyAllowlistError(err error) error {
	if err == nil {
		return nil
	}
	low := strings.ToLower(err.Error())
	for _, m := range allowlistRejectionMarkers {
		if strings.Contains(low, m) {
			return errors.Join(ErrRelayerAllowlistBlocked, err)
		}
	}
	return err
}
