package auth

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	gethmath "github.com/ethereum/go-ethereum/common/math"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
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
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
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

// SignEIP712 signs canonical EIP-712 typed data and returns the full 65-byte
// Ethereum signature with a 27/28 recovery byte.
func (s *PrivateKeySigner) SignEIP712(typed apitypes.TypedData) ([]byte, error) {
	hash, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidValue, "hash typed data", err)
	}
	sig, err := ethcrypto.Sign(hash, s.key)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidSignature, "signing failed", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}
	return sig, nil
}

const (
	proxyFactoryAddress = "0xaB45c5A4B0c941a2F231C04C3f49182e1A254052"
	safeFactoryAddress  = "0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b"
	proxyInitCodeHash   = "0xd21df8dc65880a8606f09fe0ce3df9b8869287ab0b058be05aa9e8af6330a00b"
	safeInitCodeHash    = "0x2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf"

	depositWalletFactoryAddress = "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"
	depositWalletImplAddress    = "0x58CA52ebe0DadfdF531Cde7062e76746de4Db1eB"

	// ERC-1967 proxy init code constants — verified against
	// Polymarket/py-builder-relayer-client builder/derive.py.
	erc1967Prefix = 0x61003D3D8160233D3973
	erc1967Const1 = "0xcc3735a920a3ca505d382bbc545af43d6000803e6038573d6000fd5b3d6000f3"
	erc1967Const2 = "0x5155f3363d3d373d3d363d7f360894a13ba1a3210667c828492db98dca3e2076"
)

// MakerAddressForSignatureType returns the CLOB maker/funder address for the
// configured signature type. EOA orders use the signer directly; proxy and Safe
// modes use Polymarket's deterministic CREATE2 wallet addresses on Polygon.
func MakerAddressForSignatureType(signerAddress string, chainID int64, signatureType int) (string, error) {
	signer := common.HexToAddress(signerAddress)
	switch signatureType {
	case 0:
		return signer.Hex(), nil
	case 1:
		if chainID != 137 {
			return "", fmt.Errorf("proxy wallet derivation only supports Polygon chain 137")
		}
		hash, err := decodeHexBytes(proxyInitCodeHash)
		if err != nil {
			return "", err
		}
		salt := ethcrypto.Keccak256(signer.Bytes())
		return ethcrypto.CreateAddress2(common.HexToAddress(proxyFactoryAddress), common.BytesToHash(salt), hash).Hex(), nil
	case 2:
		if chainID != 137 && chainID != 80002 {
			return "", fmt.Errorf("safe wallet derivation only supports Polygon chain 137 or Amoy chain 80002")
		}
		hash, err := decodeHexBytes(safeInitCodeHash)
		if err != nil {
			return "", err
		}
		salt := ethcrypto.Keccak256(common.LeftPadBytes(signer.Bytes(), 32))
		return ethcrypto.CreateAddress2(common.HexToAddress(safeFactoryAddress), common.BytesToHash(salt), hash).Hex(), nil
	case 3:
		if chainID != 137 {
			return "", fmt.Errorf("deposit wallet derivation only supports Polygon chain 137")
		}
		return deriveDepositWalletAddress(signer), nil
	default:
		return "", fmt.Errorf("unsupported signature type %d", signatureType)
	}
}

func EIP712ChainID(chainID int64) *gethmath.HexOrDecimal256 {
	return (*gethmath.HexOrDecimal256)(big.NewInt(chainID))
}

func decodeHexBytes(value string) ([]byte, error) {
	value = strings.TrimPrefix(value, "0x")
	out, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid init code hash: %w", err)
	}
	return out, nil
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

// deriveDepositWalletAddress computes the deterministic CREATE2 deposit wallet
// address for the given EOA owner.
func deriveDepositWalletAddress(owner common.Address) string {
	factory := common.HexToAddress(depositWalletFactoryAddress)
	impl := common.HexToAddress(depositWalletImplAddress)
	walletID := common.LeftPadBytes(owner.Bytes(), 32)

	addressType, _ := abi.NewType("address", "", nil)
	bytes32Type, _ := abi.NewType("bytes32", "", nil)
	argsType := abi.Arguments{
		{Type: addressType},
		{Type: bytes32Type},
	}
	args, _ := argsType.Pack(factory, common.BytesToHash(walletID))
	salt := ethcrypto.Keccak256Hash(args)

	initCode := depositWalletInitCode(impl, args)
	bytecodeHash := ethcrypto.Keccak256Hash(initCode)

	return ethcrypto.CreateAddress2(factory, salt, bytecodeHash.Bytes()).Hex()
}

// depositWalletInitCode builds the ERC-1967 proxy init bytecode.
// Verified against Polymarket/py-builder-relayer-client builder/derive.py.
func depositWalletInitCode(impl common.Address, args []byte) []byte {
	erc1967 := new(big.Int)
	erc1967.SetString("61003D3D8160233D3973", 16)
	argsLen := new(big.Int).Lsh(new(big.Int).SetInt64(int64(len(args))), 56)
	combined := new(big.Int).Add(erc1967, argsLen)

	c2 := hexDecode(erc1967Const2)
	c1 := hexDecode(erc1967Const1)
	six009 := hexDecode("0x6009")
	code := make([]byte, 10+20+2+len(c2)+len(c1)+len(args))
	combined.FillBytes(code[:10])
	pos := 10
	pos += copy(code[pos:], impl.Bytes())
	pos += copy(code[pos:], six009)
	pos += copy(code[pos:], c2)
	pos += copy(code[pos:], c1)
	copy(code[pos:], args)
	return code
}

// hexDecode decodes a 0x-prefixed hex string to bytes. Panics on invalid input
// — callers must only pass compile-time constants.
func hexDecode(value string) []byte {
	b, err := hex.DecodeString(strings.TrimPrefix(value, "0x"))
	if err != nil {
		panic(fmt.Sprintf("internal: invalid hex constant %q: %v", value, err))
	}
	return b
}
