// Package auth provides Polymarket authentication primitives.
// Based on patterns from polymarket-go (ybina), polymarket-go-sdk, and go-builder-signing-sdk.
package auth

import (
	"fmt"

	"github.com/TrebuchetDynamics/polygolem/internal/errors"
)

// AccessLevel defines the authentication tier.
type AccessLevel int

const (
	L0 AccessLevel = iota // public, no credentials
	L1                    // wallet signer for key creation/derivation
	L2                    // API key + HMAC for authenticated operations
)

func (l AccessLevel) String() string {
	switch l {
	case L0:
		return "L0"
	case L1:
		return "L1"
	case L2:
		return "L2"
	default:
		return fmt.Sprintf("L%d", l)
	}
}

// APIKey holds L2 credentials for HMAC-authenticated requests.
type APIKey struct {
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

func (k *APIKey) Validate() error {
	if k == nil {
		return errors.New(errors.CodeMissingCreds, "API credentials not set")
	}
	if k.Key == "" {
		return errors.New(errors.CodeMissingCreds, "API key is empty")
	}
	if k.Secret == "" {
		return errors.New(errors.CodeMissingCreds, "API secret is empty")
	}
	if k.Passphrase == "" {
		return errors.New(errors.CodeMissingCreds, "API passphrase is empty")
	}
	return nil
}

func (k *APIKey) Redacted() APIKey {
	return APIKey{
		Key:        Redact(k.Key),
		Secret:     Redact(k.Secret),
		Passphrase: Redact(k.Passphrase),
	}
}

// BuilderConfig holds credentials for builder attribution headers.
// Builder credentials must be kept separate from user L2 credentials (per PRD R3).
type BuilderConfig struct {
	Key        string
	Secret     string
	Passphrase string
}

func (bc *BuilderConfig) Valid() bool {
	return bc != nil && bc.Key != "" && bc.Secret != "" && bc.Passphrase != ""
}

// Redact replaces a secret value with a safe representation.
func Redact(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 8 {
		return "[REDACTED]"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

// Signer is the abstraction for signing operations.
// Supports local private key and leaves room for KMS/Turnkey/HSM implementations.
type Signer interface {
	// Address returns the checksummed Ethereum address.
	Address() string

	// ChainID returns the configured chain identifier.
	ChainID() int64

	// SignHash signs a 32-byte hash (personal_sign style).
	SignHash(hash [32]byte) ([]byte, error)

	// SignTypedData signs an EIP-712 typed data hash.
	SignTypedData(domainHash, structHash [32]byte) ([32]byte, error)
}

// Status reports the current auth readiness without exposing secrets.
type Status struct {
	AccessLevel    AccessLevel `json:"access_level"`
	HasSigner      bool        `json:"has_signer"`
	HasAPIKey      bool        `json:"has_api_key"`
	HasBuilder     bool        `json:"has_builder"`
	SignerAddress  string      `json:"signer_address,omitempty"`
	ChainID        int64       `json:"chain_id"`
	SignatureType  string      `json:"signature_type"`
}

// AssertL1 returns an error if the access level is below L1.
func AssertL1(level AccessLevel) error {
	if level < L1 {
		return errors.New(errors.CodeMissingSigner, "L1 authentication required: no signer configured")
	}
	return nil
}

// AssertL2 returns an error if the access level is below L2.
func AssertL2(level AccessLevel) error {
	if level < L2 {
		return errors.New(errors.CodeMissingCreds, "L2 authentication required: no API credentials configured")
	}
	return nil
}
