package auth

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/TrebuchetDynamics/polygolem/internal/errors"
	"golang.org/x/crypto/sha3"
)

// EIP-712 domain for Polymarket CLOB authentication.
// Mirrors the domain defined in py-clob-client and rs-clob-client.
const (
	clobAuthDomainName    = "ClobAuthDomain"
	clobAuthDomainVersion = "1"
)

// ClobAuthMessage is the EIP-712 typed data payload for ClobAuth.
type ClobAuthMessage struct {
	Address   string `json:"address"`
	Timestamp string `json:"timestamp"`
	Nonce     uint64 `json:"nonce"`
	Message   string `json:"message"`
}

const clobAuthDefaultMessage = "This message attests that I control the given wallet"

// EIP712Domain represents the EIP-712 domain separator.
type EIP712Domain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainID           uint64 `json:"chainId"`
	VerifyingContract string `json:"verifyingContract,omitempty"`
	Salt              string `json:"salt,omitempty"`
}

// eip712DomainType is the EIP-712 domain type definition.
var eip712DomainType = []struct{ Name, Type string }{
	{Name: "name", Type: "string"},
	{Name: "version", Type: "string"},
	{Name: "chainId", Type: "uint256"},
}

// clobAuthType is the EIP-712 type definition for ClobAuth.
var clobAuthType = []struct{ Name, Type string }{
	{Name: "address", Type: "address"},
	{Name: "timestamp", Type: "string"},
	{Name: "nonce", Type: "uint256"},
	{Name: "message", Type: "string"},
}

// HashTypedData computes the EIP-712 typed data hash.
// Implements the algorithm from EIP-712:
//
//	encode(domainSeparator, message) = 0x19 0x01 ‖ domainSeparator ‖ hashStruct(message)
//
// where:
//
//	domainSeparator = hashStruct(eip712Domain)
//	hashStruct(s) = keccak256(typeHash ‖ encodeData(s))
//	typeHash = keccak256(typeName ‖ "(" ‖ memberTypes ‖ ")")
func HashTypedData(domain EIP712Domain, message ClobAuthMessage) ([32]byte, error) {
	domainHash := hashStruct("EIP712Domain", eip712DomainType, domainValues(domain))
	messageHash := hashStruct("ClobAuth", clobAuthType, messageValues(message))

	// encode(domainSeparator, message) = \x19\x01 ‖ domainSeparator ‖ hashStruct(message)
	raw := make([]byte, 0, 66)
	raw = append(raw, 0x19, 0x01)
	raw = append(raw, domainHash[:]...)
	raw = append(raw, messageHash[:]...)

	return keccak256(raw), nil
}

func hashStruct(typeName string, typeFields []struct{ Name, Type string }, values []interface{}) [32]byte {
	typeHash := encodeType(typeName, typeFields)
	encodedData := encodeData(typeFields, values)

	h := sha3.NewLegacyKeccak256()
	h.Write(typeHash[:])
	h.Write(encodedData)
	var result [32]byte
	h.Sum(result[:0])
	return result
}

func encodeType(typeName string, fields []struct{ Name, Type string }) [32]byte {
	typeSig := typeName + "("
	for i, f := range fields {
		if i > 0 {
			typeSig += ","
		}
		typeSig += f.Type + " " + f.Name
	}
	typeSig += ")"
	return keccak256([]byte(typeSig))
}

func encodeData(fields []struct{ Name, Type string }, values []interface{}) []byte {
	var encoded []byte
	for i, f := range fields {
		v := values[i]
		switch f.Type {
		case "string":
			encoded = append(encoded, encodeString(v.(string))...)
		case "address":
			encoded = append(encoded, encodeAddress(v.(string))...)
		case "uint256":
			encoded = append(encoded, encodeUint256(v.(string))...)
		default:
			return nil
		}
	}
	return encoded
}

func domainValues(d EIP712Domain) []interface{} {
	return []interface{}{
		hashString(d.Name),           // name
		hashString(d.Version),        // version
		fmt.Sprintf("%d", d.ChainID), // chainId
	}
}

func messageValues(m ClobAuthMessage) []interface{} {
	return []interface{}{
		m.Address,                  // address
		m.Timestamp,                // timestamp
		fmt.Sprintf("%d", m.Nonce), // nonce
		hashString(m.Message),      // message (hashed because it's over 32 bytes)
	}
}

func hashString(s string) string {
	h := keccak256([]byte(s))
	return "0x" + fmt.Sprintf("%x", h)
}

func encodeString(s string) []byte {
	return encodeBytes([]byte(s), "string")
}

func encodeAddress(addr string) []byte {
	// Strip 0x prefix, pad to 32 bytes
	clean := addr
	if len(clean) >= 2 && clean[:2] == "0x" {
		clean = clean[2:]
	}
	b := make([]byte, 32)
	copy(b[12:], hexToBytes(clean))
	return b
}

func encodeUint256(n string) []byte {
	bi, ok := new(big.Int).SetString(n, 10)
	if !ok {
		return make([]byte, 32)
	}
	b := bi.Bytes()
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

func encodeBytes(b []byte, _ string) []byte {
	h := keccak256(b)
	return h[:]
}

func keccak256(data []byte) [32]byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	var result [32]byte
	h.Sum(result[:0])
	return result
}

func hexToBytes(hex string) []byte {
	if len(hex) == 0 {
		return nil
	}
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}
	b := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		fmt.Sscanf(hex[i:i+2], "%x", &b[i/2])
	}
	return b
}

// BuildClobAuthTypedData builds the EIP-712 typed data for Polymarket ClobAuth.
// Used for L1 API key creation/derivation.
func BuildClobAuthTypedData(address string, chainID uint64, timestamp string, nonce uint64) (EIP712Domain, ClobAuthMessage) {
	domain := EIP712Domain{
		Name:    clobAuthDomainName,
		Version: clobAuthDomainVersion,
		ChainID: chainID,
	}
	message := ClobAuthMessage{
		Address:   address,
		Timestamp: timestamp,
		Nonce:     nonce,
		Message:   clobAuthDefaultMessage,
	}
	return domain, message
}

// L1HeaderMap builds the L1 authentication headers for API key operations.
func L1HeaderMap(address string, signature [32]byte, timestamp string, nonce uint64) map[string]string {
	return map[string]string{
		"POLY_ADDRESS":   address,
		"POLY_SIGNATURE": fmt.Sprintf("0x%x", signature),
		"POLY_TIMESTAMP": timestamp,
		"POLY_NONCE":     fmt.Sprintf("%d", nonce),
	}
}

// --- EIP-712 TypedData wrapper for signing ---

// TypedData is the top-level EIP-712 structure.
type TypedData struct {
	Types       TypedDataTypes `json:"types"`
	PrimaryType string         `json:"primaryType"`
	Domain      EIP712Domain   `json:"domain"`
	Message     map[string]interface{} `json:"message"`
}

type TypedDataTypes map[string][]TypedDataField

type TypedDataField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TypedDataFromClobAuth builds the full EIP-712 TypedData from a ClobAuth.
func TypedDataFromClobAuth(domain EIP712Domain, msg ClobAuthMessage) TypedData {
	return TypedData{
		Types: TypedDataTypes{
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
		Domain:      domain,
		Message: map[string]interface{}{
			"address":   msg.Address,
			"timestamp": msg.Timestamp,
			"nonce":     fmt.Sprintf("%d", msg.Nonce),
			"message":   msg.Message,
		},
	}
}

// MarshalJSON serializes the full typed data for signing via personal_sign.
func (td TypedData) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"types":       td.Types,
		"primaryType": td.PrimaryType,
		"domain":      td.Domain,
		"message":     td.Message,
	})
}

// SignClobAuth performs the full L1 signing flow:
// 1. Builds EIP-712 domain and message
// 2. Computes typed data hash
// 3. Signs with provided signer
func SignClobAuth(signer Signer, chainID uint64, timestamp string, nonce uint64) ([32]byte, error) {
	domain, msg := BuildClobAuthTypedData(signer.Address(), chainID, timestamp, nonce)
	hash, err := HashTypedData(domain, msg)
	if err != nil {
		return [32]byte{}, errors.Wrap(errors.CodeInvalidSignature, "EIP-712 hash failed", err)
	}
	return signer.SignTypedData(hash, [32]byte{})
}
