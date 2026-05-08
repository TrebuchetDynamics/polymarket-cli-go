package auth

import (
	"testing"
)

func TestPrivateKeySignerDerivesAddress(t *testing.T) {
	// Well-known test key (do not use for real funds)
	signer, err := NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	if err != nil {
		t.Fatal(err)
	}
	if signer.Address() == "" {
		t.Fatal("expected non-empty address")
	}
	if signer.ChainID() != 137 {
		t.Fatalf("chainID = %d", signer.ChainID())
	}
}

func TestPrivateKeySignerRejectsEmptyKey(t *testing.T) {
	_, err := NewPrivateKeySigner("", 137)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestPrivateKeySignerRejectsInvalidKey(t *testing.T) {
	_, err := NewPrivateKeySigner("0xinvalid", 137)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestAPIKeyValidation(t *testing.T) {
	k := &APIKey{Key: "k", Secret: "s", Passphrase: "p"}
	if err := k.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAPIKeyMissingFields(t *testing.T) {
	k := &APIKey{Key: "", Secret: "s", Passphrase: "p"}
	if err := k.Validate(); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestRedact(t *testing.T) {
	if Redact("") != "" {
		t.Fatal("empty should stay empty")
	}
	if Redact("ab") != "[REDACTED]" {
		t.Fatalf("short: %s", Redact("ab"))
	}
	if Redact("abcdefghijkl") != "abcd...ijkl" {
		t.Fatalf("long: %s", Redact("abcdefghijkl"))
	}
}

func TestRedactedAPIKey(t *testing.T) {
	k := &APIKey{Key: "my-secret-key-123", Secret: "supersecret", Passphrase: "pass"}
	r := k.Redacted()
	if r.Key == "my-secret-key-123" {
		t.Fatal("key not redacted")
	}
	if r.Secret == "supersecret" {
		t.Fatal("secret not redacted")
	}
}

func TestSignHMAC(t *testing.T) {
	sig := SignHMAC("c2VjcmV0", 1700000000, "GET", "/book", nil)
	if sig == "" {
		t.Fatal("empty signature")
	}
}

func TestSignHMACWithBody(t *testing.T) {
	body := `{"token_id":"123"}`
	sig := SignHMAC("c2VjcmV0", 1700000000, "POST", "/order", &body)
	if sig == "" {
		t.Fatal("empty signature")
	}
}

func TestBuildL2Headers(t *testing.T) {
	k := &APIKey{Key: "key-1", Secret: "c2VjcmV0", Passphrase: "pass-1"}
	headers, err := BuildL2Headers(k, 1700000000, "GET", "/book", nil)
	if err != nil {
		t.Fatal(err)
	}
	if headers["POLY_API_KEY"] != "key-1" {
		t.Fatalf("API_KEY: %s", headers["POLY_API_KEY"])
	}
	if headers["POLY_PASSPHRASE"] != "pass-1" {
		t.Fatalf("PASSPHRASE: %s", headers["POLY_PASSPHRASE"])
	}
	if headers["POLY_SIGNATURE"] == "" {
		t.Fatal("empty signature")
	}
}

func TestBuildL2HeadersRejectsMissingKey(t *testing.T) {
	k := &APIKey{Key: "", Secret: "s", Passphrase: "p"}
	_, err := BuildL2Headers(k, 1700000000, "GET", "/book", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuilderHeaders(t *testing.T) {
	bc := &BuilderConfig{Key: "bk", Secret: "c2VjcmV0", Passphrase: "bp"}
	if !bc.Valid() {
		t.Fatal("builder should be valid")
	}
	headers, err := BuildBuilderHeaders(bc, 1700000000, "POST", "/order", nil)
	if err != nil {
		t.Fatal(err)
	}
	if headers["POLY_BUILDER_API_KEY"] != "bk" {
		t.Fatalf("BUILDER_API_KEY: %s", headers["POLY_BUILDER_API_KEY"])
	}
}

func TestAssertL1(t *testing.T) {
	if err := AssertL1(L0); err == nil {
		t.Fatal("L0 should fail L1 assertion")
	}
	if err := AssertL1(L1); err != nil {
		t.Fatal("L1 should pass L1 assertion")
	}
}

func TestAssertL2(t *testing.T) {
	if err := AssertL2(L1); err == nil {
		t.Fatal("L1 should fail L2 assertion")
	}
	if err := AssertL2(L2); err != nil {
		t.Fatal("L2 should pass L2 assertion")
	}
}

func TestEIP712HashTypedData(t *testing.T) {
	domain := EIP712Domain{Name: "ClobAuthDomain", Version: "1", ChainID: 137}
	msg := ClobAuthMessage{
		Address:   "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23",
		Timestamp: "1700000000",
		Nonce:     0,
		Message:   "This message attests that I control the given wallet",
	}
	hash, err := HashTypedData(domain, msg)
	if err != nil {
		t.Fatal(err)
	}
	if hash == [32]byte{} {
		t.Fatal("hash is zero")
	}
}

func TestCompactJSON(t *testing.T) {
	compact := CompactJSON(`{"key": "value", "nested": {"a": 1}}`)
	if compact != `{"key":"value","nested":{"a":1}}` {
		t.Fatalf("compact: %s", compact)
	}
}

func TestBuildL1HeadersForAddressOverridesPolyAddress(t *testing.T) {
	pk := "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	signer, err := NewPrivateKeySigner(pk, 137)
	if err != nil {
		t.Fatal(err)
	}
	override := "0x19bE70b1e4F59C0663a999C0dC6f5b3C68CFCaF3"

	headers, err := BuildL1HeadersForAddress(pk, 137, 1700000000, 0, override)
	if err != nil {
		t.Fatal(err)
	}
	if got := headers["POLY_ADDRESS"]; got != override {
		t.Errorf("POLY_ADDRESS = %s, want %s", got, override)
	}
	if got := headers["POLY_TIMESTAMP"]; got != "1700000000" {
		t.Errorf("POLY_TIMESTAMP = %s", got)
	}
	if headers["POLY_SIGNATURE"] == "" {
		t.Fatal("POLY_SIGNATURE missing")
	}

	defaults, err := BuildL1HeadersFromPrivateKey(pk, 137, 1700000000, 0)
	if err != nil {
		t.Fatal(err)
	}
	if defaults["POLY_ADDRESS"] != signer.Address() {
		t.Errorf("default POLY_ADDRESS = %s, want signer %s", defaults["POLY_ADDRESS"], signer.Address())
	}
	if headers["POLY_SIGNATURE"] == defaults["POLY_SIGNATURE"] {
		t.Fatal("override and default signatures must differ — typed-data value.address differs")
	}
}

func TestBuildL1HeadersForAddressEmptyFallsBackToSigner(t *testing.T) {
	pk := "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	signer, _ := NewPrivateKeySigner(pk, 137)

	headers, err := BuildL1HeadersForAddress(pk, 137, 1700000000, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if headers["POLY_ADDRESS"] != signer.Address() {
		t.Errorf("empty override → POLY_ADDRESS = %s, want signer %s", headers["POLY_ADDRESS"], signer.Address())
	}
}
