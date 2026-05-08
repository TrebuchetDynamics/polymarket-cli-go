package auth

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	gethmath "github.com/ethereum/go-ethereum/common/math"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// BuildL1HeadersFromPrivateKey builds Polymarket CLOB L1 auth headers for
// API-key creation and derivation. It signs the canonical ClobAuth EIP-712
// message with a local EOA private key. POLY_ADDRESS is the EOA address.
func BuildL1HeadersFromPrivateKey(privateKeyHex string, chainID int64, timestamp int64, nonce int64) (map[string]string, error) {
	return BuildL1HeadersForAddress(privateKeyHex, chainID, timestamp, nonce, "")
}

// BuildL1HeadersForDepositWallet builds CLOB L1 auth headers wrapped with the
// ERC-7739 nested EIP-712 envelope for use against a Polymarket deposit
// wallet's `isValidSignature`. The EOA holding privateKey produces the inner
// ECDSA, but POLY_ADDRESS is the smart wallet and POLY_SIGNATURE is the
// wrapped form Solady's ERC1271 mixin expects.
//
// Required when minting an L2 API key bound to a deposit wallet (so the L2
// key's address matches `Order.signer == Order.maker == depositWallet` for
// sigtype 3 — both the CLOB HTTP gate and the on-chain
// `_verifyPoly1271Signature` use this equality).
//
// The deposit wallet must be deployed at depositWalletAddress before this
// call (the contract's `isValidSignature` runs on-chain).
func BuildL1HeadersForDepositWallet(privateKeyHex string, chainID int64, timestamp int64, nonce int64, depositWalletAddress string) (map[string]string, error) {
	if depositWalletAddress == "" {
		return nil, fmt.Errorf("depositWalletAddress is required")
	}
	signer, err := NewPrivateKeySigner(privateKeyHex, chainID)
	if err != nil {
		return nil, err
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	// Build the ClobAuth typed data with depositWallet as the address field.
	typed := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    clobAuthDomainName,
			Version: clobAuthDomainVersion,
			ChainId: (*gethmath.HexOrDecimal256)(big.NewInt(chainID)),
		},
		Message: apitypes.TypedDataMessage{
			"address":   depositWalletAddress,
			"timestamp": strconv.FormatInt(timestamp, 10),
			"nonce":     (*gethmath.HexOrDecimal256)(big.NewInt(nonce)),
			"message":   clobAuthDefaultMessage,
		},
	}
	// TypedDataAndHash returns rawData = "\x19\x01" || domainSep(32) || structHash(32).
	_, rawDataStr, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return nil, fmt.Errorf("hash ClobAuth typed data: %w", err)
	}
	rawData := []byte(rawDataStr)
	if len(rawData) != 66 {
		return nil, fmt.Errorf("unexpected rawData length %d", len(rawData))
	}
	var appDomainSep, contents [32]byte
	copy(appDomainSep[:], rawData[2:34])
	copy(contents[:], rawData[34:66])

	const clobAuthContentsType = "ClobAuth(address address,string timestamp,uint256 nonce,string message)"
	wrappedSig, err := WrapERC7739Signature(signer, depositWalletAddress, chainID, appDomainSep, contents, clobAuthContentsType)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"POLY_ADDRESS":   depositWalletAddress,
		"POLY_SIGNATURE": wrappedSig,
		"POLY_TIMESTAMP": strconv.FormatInt(timestamp, 10),
		"POLY_NONCE":     strconv.FormatInt(nonce, 10),
	}, nil
}

// BuildL1HeadersForAddress is the smart-wallet variant. When ownerAddress is
// non-empty it overrides both POLY_ADDRESS and the typed-data value.address;
// the EOA still produces the signature, but the CLOB validates it via
// ERC-1271 against the smart-wallet contract (proxy / Safe / deposit-wallet).
//
// Required for sigtype-3 deposit-wallet API-key minting: without the override,
// the L2 key is bound to the EOA and the CLOB's "the order signer address has
// to be the address of the API KEY" gate rejects sigtype-3 orders whose
// signer is the deposit wallet.
//
// The smart-wallet must already be deployed at ownerAddress for ERC-1271
// validation to succeed.
func BuildL1HeadersForAddress(privateKeyHex string, chainID int64, timestamp int64, nonce int64, ownerAddress string) (map[string]string, error) {
	signer, err := NewPrivateKeySigner(privateKeyHex, chainID)
	if err != nil {
		return nil, err
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	polyAddress := ownerAddress
	if polyAddress == "" {
		polyAddress = signer.Address()
	}
	typed := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    clobAuthDomainName,
			Version: clobAuthDomainVersion,
			ChainId: (*gethmath.HexOrDecimal256)(big.NewInt(chainID)),
		},
		Message: apitypes.TypedDataMessage{
			"address":   polyAddress,
			"timestamp": strconv.FormatInt(timestamp, 10),
			"nonce":     (*gethmath.HexOrDecimal256)(big.NewInt(nonce)),
			"message":   clobAuthDefaultMessage,
		},
	}
	hash, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return nil, err
	}
	sig, err := ethcrypto.Sign(hash, signer.key)
	if err != nil {
		return nil, err
	}
	if sig[64] < 27 {
		sig[64] += 27
	}
	return map[string]string{
		"POLY_ADDRESS":   polyAddress,
		"POLY_SIGNATURE": fmt.Sprintf("0x%x", sig),
		"POLY_TIMESTAMP": strconv.FormatInt(timestamp, 10),
		"POLY_NONCE":     strconv.FormatInt(nonce, 10),
	}, nil
}
