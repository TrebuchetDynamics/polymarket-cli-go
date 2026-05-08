package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// SIWE (Sign-In With Ethereum, EIP-4361) message construction and bearer
// token packaging for the Polymarket gamma-api login flow.
//
// Empirical reference: the Polymarket frontend bundle's `createSiweMessage`
// (viem) plus `useSignIn` mutation, which builds the same fields and ships
// them as `base64(JSON.stringify(fields) + ":::" + signature)` to
// `gamma-api.polymarket.com/login`. See
// `polydart/docs/HEADLESS-BUILDER-KEYS-INVESTIGATION.md`.

// PolymarketSIWE returns the canonical Polymarket SIWE statement and domain.
const (
	polymarketSIWEDomain    = "polymarket.com"
	polymarketSIWEStatement = "Welcome to Polymarket! Sign to connect."
	polymarketSIWEURI       = "https://polymarket.com"
	siweVersion             = "1"
)

// SIWEMessage holds the structured EIP-4361 fields. Field names match the
// viem JSON representation that ships to /login as the bearer payload.
type SIWEMessage struct {
	Domain         string `json:"domain"`
	Address        string `json:"address"`
	Statement      string `json:"statement"`
	URI            string `json:"uri"`
	Version        string `json:"version"`
	ChainID        int64  `json:"chainId"`
	Nonce          string `json:"nonce"`
	IssuedAt       string `json:"issuedAt"`
	ExpirationTime string `json:"expirationTime"`
}

// NewPolymarketSIWE builds a SIWEMessage with the Polymarket-specific
// domain, statement, and URI defaults. The address is checksummed; the
// nonce comes from `GET gamma-api.polymarket.com/nonce`. Issued-at and
// expiration are populated automatically with a 7-day window unless
// overridden.
func NewPolymarketSIWE(address, nonce string, chainID int64, now time.Time) SIWEMessage {
	checksummed := common.HexToAddress(address).Hex()
	return SIWEMessage{
		Domain:         polymarketSIWEDomain,
		Address:        checksummed,
		Statement:      polymarketSIWEStatement,
		URI:            polymarketSIWEURI,
		Version:        siweVersion,
		ChainID:        chainID,
		Nonce:          nonce,
		IssuedAt:       now.UTC().Format(time.RFC3339),
		ExpirationTime: now.Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339),
	}
}

// String renders the SIWEMessage in EIP-4361 plaintext form. This is the
// blob that gets personal_sign-hashed.
func (m SIWEMessage) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s wants you to sign in with your Ethereum account:\n", m.Domain)
	fmt.Fprintf(&b, "%s\n\n", m.Address)
	if m.Statement != "" {
		fmt.Fprintf(&b, "%s\n\n", m.Statement)
	}
	fmt.Fprintf(&b, "URI: %s\n", m.URI)
	fmt.Fprintf(&b, "Version: %s\n", m.Version)
	fmt.Fprintf(&b, "Chain ID: %d\n", m.ChainID)
	fmt.Fprintf(&b, "Nonce: %s\n", m.Nonce)
	fmt.Fprintf(&b, "Issued At: %s", m.IssuedAt)
	if m.ExpirationTime != "" {
		fmt.Fprintf(&b, "\nExpiration Time: %s", m.ExpirationTime)
	}
	return b.String()
}

// BuildSIWEBearerToken assembles the Polymarket /login bearer token:
// base64( JSON(message fields) + ":::" + 0x-prefixed signature hex ).
func BuildSIWEBearerToken(msg SIWEMessage, signature []byte) (string, error) {
	fields, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("marshal siwe fields: %w", err)
	}
	sigHex := "0x" + hexEncode(signature)
	combined := string(fields) + ":::" + sigHex
	return base64.StdEncoding.EncodeToString([]byte(combined)), nil
}

// hexEncode is a local helper to avoid pulling encoding/hex into the
// bigger import set used elsewhere.
func hexEncode(b []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hex[c>>4]
		out[i*2+1] = hex[c&0xf]
	}
	return string(out)
}
