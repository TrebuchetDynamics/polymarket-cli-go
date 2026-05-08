// parity_walletbatch emits the intermediate hashes of a canonical
// DepositWallet.Batch typed-data sample. polydart locks these values into
// `test/wallet/deposit_wallet_signing_test.dart`.
//
//	go run ./cmd/parity_walletbatch
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	gethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func main() {
	walletAddress := "0xfd5041047be8c192c725a66228f141196fa3cf9c"
	nonce := "7"
	deadline := "1750000000"
	calls := []map[string]interface{}{
		{
			"target": "0x6c030c5cc283f791b26816f325b9c632d964f8a1",
			"value":  "0",
			"data":   "0xa9059cbb000000000000000000000000000000000000000000000000000000000000bee0000000000000000000000000000000000000000000000000000000000000000a",
		},
		{
			"target": "0x4f9b03a3c34e9ff7c20de0f5d5b4a9b3a82a8ac4",
			"value":  "1000000",
			"data":   "0x",
		},
	}

	td := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Call": {
				{Name: "target", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "data", Type: "bytes"},
			},
			"Batch": {
				{Name: "wallet", Type: "address"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
				{Name: "calls", Type: "Call[]"},
			},
		},
		PrimaryType: "Batch",
		Domain: apitypes.TypedDataDomain{
			Name:              "DepositWallet",
			Version:           "1",
			ChainId:           (*gethmath.HexOrDecimal256)(big.NewInt(137)),
			VerifyingContract: walletAddress,
		},
		Message: map[string]interface{}{
			"wallet":   walletAddress,
			"nonce":    nonce,
			"deadline": deadline,
			"calls":    toIfaceSlice(calls),
		},
	}

	hash, rawData, err := apitypes.TypedDataAndHash(td)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	raw := []byte(rawData)
	domainSep := raw[2:34]
	structHash := raw[34:66]

	// Per-call hashStruct values (so polydart can lock them too).
	callType := td.Types["Call"]
	callTypeHash, err := td.HashStruct("Call", apitypes.TypedDataMessage(calls[0]))
	_ = callType // unused; kept for clarity
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	callHash0 := callTypeHash
	callHash1, err := td.HashStruct("Call", apitypes.TypedDataMessage(calls[1]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := map[string]string{
		"domainSeparator": hex.EncodeToString(domainSep),
		"batchStructHash": hex.EncodeToString(structHash),
		"finalDigest":     hex.EncodeToString(hash),
		"callHash0":       hex.EncodeToString(callHash0),
		"callHash1":       hex.EncodeToString(callHash1),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func toIfaceSlice(in []map[string]interface{}) []interface{} {
	out := make([]interface{}, len(in))
	for i, m := range in {
		out[i] = m
	}
	return out
}
