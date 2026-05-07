package wallet

import (
	"math/big"
	"golang.org/x/crypto/sha3"
)

// Contract addresses for Polygon mainnet (chainID 137).
const (
	ProxyFactoryAddr  = "0xaB45c5A4B0c941a2F231C04C3f49182e1A254052"
	SafeFactoryAddr   = "0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b"
	PolygonChainID    = 137

	// CREATE2 init code hashes
	proxyInitCodeHash = "0xd21df8dc65880a8606f09fe0ce3df9b8869287ab0b058be05aa9e8af6330a00b"
	safeInitCodeHash  = "0x2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf"
)

// DeriveProxyWallet computes the deterministic proxy wallet address via CREATE2.
// Proxy wallets are used by Polymarket Magic/email accounts.
func DeriveProxyWallet(eoa string) string {
	return deriveCreate2(ProxyFactoryAddr, proxySalt(eoa), proxyInitCodeHash)
}

// DeriveSafeWallet computes the deterministic Gnosis Safe address via CREATE2.
// Safe wallets are used by Polymarket browser accounts.
func DeriveSafeWallet(eoa string) string {
	return deriveCreate2(SafeFactoryAddr, safeSalt(eoa), safeInitCodeHash)
}

func proxySalt(eoa string) []byte {
	clean := strip0x(eoa)
	if len(clean) != 40 {
		return make([]byte, 32)
	}
	salt := make([]byte, 32)
	hexDecodeInto(salt[12:], clean)
	return salt
}

func safeSalt(eoa string) []byte {
	clean := strip0x(eoa)
	salt := make([]byte, 32)
	hexDecodeInto(salt, clean)
	return salt
}

// deriveCreate2 implements EIP-1014 CREATE2 address derivation.
// address = keccak256(0xff + factory + salt + keccak256(initCode))[12:]
func deriveCreate2(factory string, salt []byte, initCodeHash string) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte{0xff})
	hash.Write(hexToBytes(strip0x(factory)))
	hash.Write(salt)
	hash.Write(hexToBytes(strip0x(initCodeHash)))
	result := hash.Sum(nil)
	return "0x" + toHex(result[12:])
}

// ReadyInfo holds wallet readiness information.
type ReadyInfo struct {
	ChainID       int64  `json:"chain_id"`
	EOA           string `json:"eoa,omitempty"`
	ProxyWallet   string `json:"proxy_wallet,omitempty"`
	SafeWallet    string `json:"safe_wallet,omitempty"`
	HasSigner     bool   `json:"has_signer"`
}

// Readiness returns wallet readiness info for the given EOA.
func Readiness(chainID int64, eoa string) ReadyInfo {
	info := ReadyInfo{
		ChainID: chainID,
		EOA:     eoa,
	}
	if eoa != "" {
		info.HasSigner = true
		info.ProxyWallet = DeriveProxyWallet(eoa)
		info.SafeWallet = DeriveSafeWallet(eoa)
	}
	return info
}

// helpers

func strip0x(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	}
	return s
}

func hexToBytes(s string) []byte {
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		n := new(big.Int)
		n.SetString(s[i:i+2], 16)
		b[i/2] = byte(n.Uint64())
	}
	return b
}

func hexDecodeInto(dst []byte, src string) {
	for i := 0; i < len(src) && i/2 < len(dst); i += 2 {
		n := new(big.Int)
		n.SetString(src[i:i+2], 16)
		dst[i/2] = byte(n.Uint64())
	}
}

func toHex(b []byte) string {
	h := ""
	for _, v := range b {
		h += string("0123456789abcdef"[v>>4])
		h += string("0123456789abcdef"[v&0xf])
	}
	return h
}
