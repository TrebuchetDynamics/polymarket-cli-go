package auth

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/TrebuchetDynamics/polygolem/internal/errors"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// PrivateKeySigner implements Signer using go-ethereum/secp256k1.
type PrivateKeySigner struct {
	key     *ecdsa.PrivateKey
	address string
	chainID int64
}

// NewPrivateKeySigner creates a signer from a 0x-prefixed hex private key.
func NewPrivateKeySigner(privateKeyHex string, chainID int64) (*PrivateKeySigner, error) {
	if privateKeyHex == "" {
		return nil, errors.New(errors.CodeMissingSigner, "private key is empty")
	}
	key, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidValue, "invalid private key", err)
	}
	addr := ethcrypto.PubkeyToAddress(key.PublicKey).Hex()
	return &PrivateKeySigner{
		key:     key,
		address: addr,
		chainID: chainID,
	}, nil
}

func (s *PrivateKeySigner) Address() string { return s.address }
func (s *PrivateKeySigner) ChainID() int64  { return s.chainID }

// SignHash signs a 32-byte hash using personal_sign prefix.
func (s *PrivateKeySigner) SignHash(hash [32]byte) ([]byte, error) {
	msg := ethcrypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		hash[:],
	)
	sig, err := ethcrypto.Sign(msg, s.key)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidSignature, "signing failed", err)
	}
	sig[64] += 27
	return sig, nil
}

func (s *PrivateKeySigner) SignTypedData(hash [32]byte, _ [32]byte) ([32]byte, error) {
	sig, err := s.SignHash(hash)
	if err != nil {
		return [32]byte{}, err
	}
	var result [32]byte
	copy(result[:], sig[:32])
	return result, nil
}

// GeneratePrivateKey creates a new random secp256k1 key.
func GeneratePrivateKey() (string, error) {
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(ethcrypto.FromECDSA(key)), nil
}

// PrivateKeyToAddress derives the Ethereum address from a hex private key.
func PrivateKeyToAddress(privateKeyHex string) (string, error) {
	signer, err := NewPrivateKeySigner(privateKeyHex, 0)
	if err != nil {
		return "", err
	}
	return signer.Address(), nil
}
