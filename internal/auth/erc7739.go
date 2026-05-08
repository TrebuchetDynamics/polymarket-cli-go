package auth

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// WrapERC7739Signature produces a Solady-compatible ERC-7739 nested EIP-712
// wrapped signature suitable for a Polymarket deposit wallet's
// `isValidSignature` check. The deposit wallet's
// `_erc1271IsValidSignatureViaNestedEIP712` recomputes:
//
//	finalHash = keccak256(0x1901 || appDomainSep ||
//	            hashStruct(TypedDataSign{contents, DepositWallet inline domain}))
//
// recovers the EOA from innerSig, and returns ERC1271_MAGIC if the recovered
// address matches the wallet's owner.
//
// Layout: innerSig(65) || appDomainSep(32) || contents(32) || contentsType ||
// uint16BE(len(contentsType)).
//
// Caller responsibilities:
//   - appDomainSep is the OUTER domain separator the dapp/contract uses
//     (e.g. CTF Exchange V2 for orders, ClobAuthDomain for L1 mints).
//   - contents is the EIP-712 hashStruct of the original typed-data.
//   - contentsType is the canonical type string (must start with the primary
//     type name, e.g. "Order(...)" or "ClobAuth(...)").
//   - depositWalletAddress is the smart wallet that will validate the
//     signature; it must be deployed.
func WrapERC7739Signature(signer *PrivateKeySigner, depositWalletAddress string, chainID int64, appDomainSep [32]byte, contents [32]byte, contentsType string) (string, error) {
	if signer == nil {
		return "", fmt.Errorf("signer is required")
	}
	if depositWalletAddress == "" {
		return "", fmt.Errorf("depositWalletAddress is required")
	}
	if len(contentsType) == 0 || len(contentsType) > 0xffff {
		return "", fmt.Errorf("contentsType length %d out of range [1, 65535]", len(contentsType))
	}

	// TypedDataSign typehash inlines contentsType so the wallet's signature
	// validation reproduces the exact same struct.
	typeHashStr := "TypedDataSign(" +
		typedDataSignContentsField(contentsType) + " contents," +
		"string name,string version,uint256 chainId,address verifyingContract,bytes32 salt)" +
		contentsType
	typedDataSignTypehash := ethcrypto.Keccak256([]byte(typeHashStr))

	// hashStruct(TypedDataSign{contents, DepositWallet inline domain values}).
	// The INNER struct's domain identifies the WALLET (not the app).
	dwNameHash := ethcrypto.Keccak256([]byte("DepositWallet"))
	dwVerHash := ethcrypto.Keccak256([]byte("1"))
	dwChainIDBytes := common.LeftPadBytes(big.NewInt(chainID).Bytes(), 32)
	dwAddrBytes := common.LeftPadBytes(common.HexToAddress(depositWalletAddress).Bytes(), 32)
	dwSaltBytes := make([]byte, 32) // zeros

	tdsStruct := ethcrypto.Keccak256(
		typedDataSignTypehash,
		contents[:],
		dwNameHash,
		dwVerHash,
		dwChainIDBytes,
		dwAddrBytes,
		dwSaltBytes,
	)

	// finalHash = keccak256(0x1901 || appDomainSep || tdsStruct).
	finalHashInput := make([]byte, 0, 66)
	finalHashInput = append(finalHashInput, 0x19, 0x01)
	finalHashInput = append(finalHashInput, appDomainSep[:]...)
	finalHashInput = append(finalHashInput, tdsStruct...)
	finalHashSum := ethcrypto.Keccak256(finalHashInput)
	var finalHash [32]byte
	copy(finalHash[:], finalHashSum)

	innerSig, err := signer.SignRaw(finalHash)
	if err != nil {
		return "", fmt.Errorf("sign inner: %w", err)
	}

	var lenBuf [2]byte
	binary.BigEndian.PutUint16(lenBuf[:], uint16(len(contentsType)))
	sig := make([]byte, 0, 65+32+32+len(contentsType)+2)
	sig = append(sig, innerSig...)
	sig = append(sig, appDomainSep[:]...)
	sig = append(sig, contents[:]...)
	sig = append(sig, []byte(contentsType)...)
	sig = append(sig, lenBuf[:]...)
	return "0x" + hex.EncodeToString(sig), nil
}

// typedDataSignContentsField extracts the primary type name from a contentsType
// string (everything before the first '('). Solady's nested EIP-712 expects
// the TypedDataSign typehash to declare `<PrimaryType> contents,...`.
func typedDataSignContentsField(contentsType string) string {
	for i, c := range contentsType {
		if c == '(' {
			return contentsType[:i]
		}
	}
	return contentsType
}
