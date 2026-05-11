package auth

import "testing"

func TestEIP712Domain(t *testing.T) {
	d := EIP712Domain{
		Name:              "Polymarket CTF Exchange",
		Version:           "2",
		ChainID:           137,
		VerifyingContract: "0xE111180000d2663C0091e4f400237545B87B996B",
	}
	sep, err := d.DomainSeparator()
	if err == nil {
		t.Fatal("expected error for unimplemented separator")
	}
	_ = sep
}

func TestIsValidSignatureType(t *testing.T) {
	if !IsValidSignatureType(SignatureTypeEOA) {
		t.Fatal("expected EOA to be valid")
	}
	if !IsValidSignatureType(SignatureTypePOLY1271) {
		t.Fatal("expected POLY1271 to be valid")
	}
	if IsValidSignatureType(99) {
		t.Fatal("expected 99 to be invalid")
	}
}

func TestHexToBigInt(t *testing.T) {
	n, err := HexToBigInt("0x123")
	if err != nil {
		t.Fatal(err)
	}
	if n.Int64() != 291 {
		t.Fatalf("expected 291, got %d", n.Int64())
	}
}
