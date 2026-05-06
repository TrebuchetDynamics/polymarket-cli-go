package auth

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewSigner(privateKeyHex string) (*Signer, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	
	address := crypto.PubkeyToAddress(key.PublicKey)
	return &Signer{
		privateKey: key,
		address:    address,
	}, nil
}

func (s *Signer) Address() string {
	return s.address.Hex()
}

func (s *Signer) PrivateKey() *ecdsa.PrivateKey {
	return s.privateKey
}

func (s *Signer) DeriveProxyAddress(chainID int) string {
	// Polymarket proxy wallet is deterministically derived from EOA
	// This is a simplified version - actual derivation uses CREATE2
	// For now, return EOA address as fallback
	return s.address.Hex()
}
