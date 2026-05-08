// parity_erc7739 emits the intermediate ERC-7739 / POLY_1271 wrap-signature
// hashes for the canonical polydart OrderV2Draft. The values it prints are
// pasted into polydart `test/auth/erc7739_test.dart` as parity vectors.
//
// Run from the polygolem repo root:
//
//	go run ./cmd/parity_erc7739
package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	gethmath "github.com/ethereum/go-ethereum/common/math"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	chainID                  = int64(137)
	clobExchangeAddress      = "0xE111180000d2663C0091e4f400237545B87B996B"
	bytes32Zero              = "0x0000000000000000000000000000000000000000000000000000000000000000"
	depositWalletDomainName  = "DepositWallet"
	depositWalletDomainVer   = "1"
	canonicalContentsType    = "Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"
	testEoaAddress           = "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"
	testDepositWalletAddress = "0xfd5041047be8c192c725a66228f141196fa3cf9c"
)

func main() {
	// Match polydart canonical draft. For sigtype=3, polygolem's
	// buildSignedOrderPayload sets maker=signer=depositWallet.
	salt := "1"
	maker := testDepositWalletAddress
	signer := testDepositWalletAddress
	tokenID := "12345"
	makerAmount := "5500000"
	takerAmount := "10000000"
	side := "BUY"
	signatureType := int64(3) // poly1271
	timestamp := "1700000000000"

	td := buildOrderTypedData(salt, maker, signer, tokenID, makerAmount, takerAmount, side, signatureType, timestamp, false)

	_, rawDataStr, err := apitypes.TypedDataAndHash(td)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rawData := []byte(rawDataStr)
	appDomainSep := rawData[2:34]
	contents := rawData[34:66]

	tdsTypeHash := ethcrypto.Keccak256([]byte(
		"TypedDataSign(Order contents,string name,string version,uint256 chainId,address verifyingContract,bytes32 salt)" + canonicalContentsType,
	))

	dwNameHash := ethcrypto.Keccak256([]byte(depositWalletDomainName))
	dwVerHash := ethcrypto.Keccak256([]byte(depositWalletDomainVer))
	dwChainIDBytes := common.LeftPadBytes(big.NewInt(chainID).Bytes(), 32)
	dwAddrBytes := common.LeftPadBytes(common.HexToAddress(testDepositWalletAddress).Bytes(), 32)
	dwSaltBytes := make([]byte, 32)

	tdsStruct := ethcrypto.Keccak256(
		tdsTypeHash, contents, dwNameHash, dwVerHash, dwChainIDBytes, dwAddrBytes, dwSaltBytes,
	)

	finalInput := make([]byte, 0, 66)
	finalInput = append(finalInput, 0x19, 0x01)
	finalInput = append(finalInput, appDomainSep...)
	finalInput = append(finalInput, tdsStruct...)
	finalHash := ethcrypto.Keccak256(finalInput)

	// Synthesize a deterministic 65-byte placeholder innerSig so we can
	// also lock the assembled wrap layout. The polydart side mirrors this
	// fake innerSig in its assemble test.
	innerSig := make([]byte, 65)
	for i := range innerSig {
		innerSig[i] = byte(0xa0 + (i % 16))
	}

	var lenBuf [2]byte
	binary.BigEndian.PutUint16(lenBuf[:], uint16(len(canonicalContentsType)))
	wrap := make([]byte, 0, 317)
	wrap = append(wrap, innerSig...)
	wrap = append(wrap, appDomainSep...)
	wrap = append(wrap, contents...)
	wrap = append(wrap, []byte(canonicalContentsType)...)
	wrap = append(wrap, lenBuf[:]...)

	out := map[string]string{
		"appDomainSep":        hex.EncodeToString(appDomainSep),
		"contents":            hex.EncodeToString(contents),
		"tdsTypeHash":         hex.EncodeToString(tdsTypeHash),
		"dwNameHash":          hex.EncodeToString(dwNameHash),
		"dwVerHash":           hex.EncodeToString(dwVerHash),
		"tdsStruct":           hex.EncodeToString(tdsStruct),
		"finalHash":           hex.EncodeToString(finalHash),
		"contentsType":        canonicalContentsType,
		"contentsTypeHex":     hex.EncodeToString([]byte(canonicalContentsType)),
		"placeholderInnerSig": hex.EncodeToString(innerSig),
		"assembledWrap":       hex.EncodeToString(wrap),
		"wrapLength":          fmt.Sprintf("%d", len(wrap)),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildOrderTypedData(salt, maker, signer, tokenID, makerAmount, takerAmount, side string, signatureType int64, timestamp string, negRisk bool) apitypes.TypedData {
	verifyingContract := clobExchangeAddress
	sideInt := int64(0)
	if side == "SELL" {
		sideInt = 1
	}
	tokenIDBig, _ := new(big.Int).SetString(tokenID, 10)
	makerAmountBig, _ := new(big.Int).SetString(makerAmount, 10)
	takerAmountBig, _ := new(big.Int).SetString(takerAmount, 10)
	timestampBig, _ := new(big.Int).SetString(timestamp, 10)
	saltBig, _ := new(big.Int).SetString(salt, 10)
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
				{Name: "timestamp", Type: "uint256"},
				{Name: "metadata", Type: "bytes32"},
				{Name: "builder", Type: "bytes32"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "Polymarket CTF Exchange",
			Version:           "2",
			ChainId:           (*gethmath.HexOrDecimal256)(big.NewInt(chainID)),
			VerifyingContract: verifyingContract,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          (*gethmath.HexOrDecimal256)(saltBig),
			"maker":         common.HexToAddress(maker).Hex(),
			"signer":        common.HexToAddress(signer).Hex(),
			"tokenId":       (*gethmath.HexOrDecimal256)(tokenIDBig),
			"makerAmount":   (*gethmath.HexOrDecimal256)(makerAmountBig),
			"takerAmount":   (*gethmath.HexOrDecimal256)(takerAmountBig),
			"side":          (*gethmath.HexOrDecimal256)(big.NewInt(sideInt)),
			"signatureType": (*gethmath.HexOrDecimal256)(big.NewInt(signatureType)),
			"timestamp":     (*gethmath.HexOrDecimal256)(timestampBig),
			"metadata":      common.HexToHash(bytes32Zero).Hex(),
			"builder":       common.HexToHash(bytes32Zero).Hex(),
		},
	}
}
