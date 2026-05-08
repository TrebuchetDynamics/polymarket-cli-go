package relayer

import (
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const (
	pusdAddress = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	ctfAddress  = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"

	ctfExchangeV2     = "0xE111180000d2663C0091e4f400237545B87B996B"
	negRiskExchangeV2 = "0xe2222d279d744050d28e00520010520000310F59"
	negRiskAdapterV2  = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"

	erc20ApproveSelector        = "095ea7b3"
	erc1155SetApprovalForAllSel = "a22cb465"
	maxUint256                  = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
)

func pad32Bytes(hexAddr string) string {
	hexAddr = strings.TrimPrefix(strings.TrimSpace(hexAddr), "0x")
	return strings.Repeat("0", 64-len(hexAddr)) + hexAddr
}

func buildApproveCall(tokenAddress, spenderAddress string) DepositWalletCall {
	token := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(tokenAddress), "0x"))
	spender := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(spenderAddress), "0x"))
	data := "0x" + erc20ApproveSelector + pad32Bytes(spender) + maxUint256
	return DepositWalletCall{
		Target: common.HexToAddress(token).Hex(),
		Value:  "0",
		Data:   data,
	}
}

func buildCTFApprovalCall(operatorAddress string) DepositWalletCall {
	operator := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(operatorAddress), "0x"))
	data := "0x" + erc1155SetApprovalForAllSel + pad32Bytes(operator) +
		"0000000000000000000000000000000000000000000000000000000000000001"
	return DepositWalletCall{
		Target: common.HexToAddress(ctfAddress).Hex(),
		Value:  "0",
		Data:   data,
	}
}

func buildTransferCall(tokenAddress, toAddress, amountHex string) DepositWalletCall {
	token := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(tokenAddress), "0x"))
	to := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(toAddress), "0x"))
	amount := strings.TrimPrefix(strings.TrimSpace(amountHex), "0x")
	data := "0xa9059cbb" + pad32Bytes(to) + pad32Bytes(amount)
	return DepositWalletCall{
		Target: common.HexToAddress(token).Hex(),
		Value:  "0",
		Data:   data,
	}
}

// BuildApprovalCalls returns the 6 calls needed to approve pUSD and CTF for
// all V2 exchange spenders. These must be submitted via WALLET batch from
// the deposit wallet.
func BuildApprovalCalls() []DepositWalletCall {
	calls := make([]DepositWalletCall, 0, 6)
	for _, spender := range []string{ctfExchangeV2, negRiskExchangeV2, negRiskAdapterV2} {
		calls = append(calls,
			buildApproveCall(pusdAddress, spender),
			buildCTFApprovalCall(spender),
		)
	}
	return calls
}

// BuildApprovalCallsJSON returns the approval calls as a JSON-marshalable
// slice for CLI --calls-json. It also returns the raw bytes for validation.
func BuildApprovalCallsJSON() (string, error) {
	calls := BuildApprovalCalls()
	raw, err := marshalCalls(calls)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func marshalCalls(calls []DepositWalletCall) ([]byte, error) {
	return json.Marshal(calls)
}
