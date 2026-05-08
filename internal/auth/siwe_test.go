package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

const siweTestPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

func TestSIWEMessageMatchesEIP4361Format(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	msg := NewPolymarketSIWE(
		"0x9d8a62f656a8d1615c1294fd71e9cfb3e4855a4f",
		"abc123",
		137,
		now,
	)

	got := msg.String()
	want := "polymarket.com wants you to sign in with your Ethereum account:\n" +
		"0x9d8A62f656a8d1615C1294fd71e9CFb3E4855A4F\n\n" +
		"Welcome to Polymarket! Sign to connect.\n\n" +
		"URI: https://polymarket.com\n" +
		"Version: 1\n" +
		"Chain ID: 137\n" +
		"Nonce: abc123\n" +
		"Issued At: 2026-05-08T12:00:00Z\n" +
		"Expiration Time: 2026-05-15T12:00:00Z"

	if got != want {
		t.Fatalf("SIWE message mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSIWEAddressIsChecksummed(t *testing.T) {
	msg := NewPolymarketSIWE(
		"0X9D8A62F656A8D1615C1294FD71E9CFB3E4855A4F", // all uppercase
		"n",
		137,
		time.Unix(1, 0),
	)
	if msg.Address != "0x9d8A62f656a8d1615C1294fd71e9CFb3E4855A4F" {
		t.Fatalf("address not checksummed: %s", msg.Address)
	}
}

func TestSIWEBearerTokenRoundTrips(t *testing.T) {
	msg := NewPolymarketSIWE(
		"0x9d8a62f656a8d1615c1294fd71e9cfb3e4855a4f",
		"abc",
		137,
		time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
	)
	sig := make([]byte, 65)
	for i := range sig {
		sig[i] = byte(i)
	}

	token, err := BuildSIWEBearerToken(msg, sig)
	if err != nil {
		t.Fatalf("BuildSIWEBearerToken: %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token not valid base64: %v", err)
	}
	combined := string(decoded)
	if !strings.Contains(combined, ":::0x") {
		t.Fatalf("token missing :::0x signature separator: %s", combined)
	}
	if !strings.Contains(combined, `"chainId":137`) {
		t.Fatalf("token JSON missing chainId: %s", combined)
	}
	if !strings.Contains(combined, `"address":"0x9d8A62f656a8d1615C1294fd71e9CFb3E4855A4F"`) {
		t.Fatalf("token JSON missing checksummed address: %s", combined)
	}
}

func TestSIWESignPersonalMessageProducesValidSig(t *testing.T) {
	signer, err := NewPrivateKeySigner(siweTestPrivateKey, 137)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	msg := NewPolymarketSIWE(
		signer.Address(),
		"deadbeef",
		137,
		time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
	)
	sig, err := signer.SignPersonalMessage([]byte(msg.String()))
	if err != nil {
		t.Fatalf("SignPersonalMessage: %v", err)
	}
	if len(sig) != 65 {
		t.Fatalf("signature length=%d, want 65", len(sig))
	}
	if sig[64] != 27 && sig[64] != 28 {
		t.Fatalf("recovery byte=%d, want 27 or 28", sig[64])
	}
}
