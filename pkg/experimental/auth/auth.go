// Package auth provides an experimental public API for Polymarket
// signing primitives.
//
// WARNING: This package is experimental. APIs may change without notice.
package auth

import (
	"fmt"
	"math/big"
	"strings"
)

// EIP712Domain is the typed-data domain for Polymarket CTF Exchange V2.
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           int64
	VerifyingContract string
}

// DomainSeparator computes the EIP-712 domain separator.
func (d EIP712Domain) DomainSeparator() ([]byte, error) {
	if d.VerifyingContract == "" {
		return nil, fmt.Errorf("verifyingContract is required")
	}
	return nil, fmt.Errorf("not yet implemented")
}

// SignatureType constants.
const (
	SignatureTypeEOA        = 0
	SignatureTypePOLYProxy  = 1
	SignatureTypeGnosisSafe = 2
	SignatureTypePOLY1271   = 3
)

// IsValidSignatureType checks if the given type is a known signature type.
func IsValidSignatureType(t int) bool {
	switch t {
	case SignatureTypeEOA, SignatureTypePOLYProxy, SignatureTypeGnosisSafe, SignatureTypePOLY1271:
		return true
	}
	return false
}

// HexToBigInt converts a hex string (with or without 0x prefix) to *big.Int.
func HexToBigInt(s string) (*big.Int, error) {
	s = strings.TrimPrefix(s, "0x")
	n, ok := new(big.Int).SetString(s, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex: %s", s)
	}
	return n, nil
}
