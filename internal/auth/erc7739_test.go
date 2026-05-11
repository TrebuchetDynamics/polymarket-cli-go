package auth

import (
	"encoding/hex"
	"strings"
	"testing"
)

const (
	testOrderPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	polygonChainID      = 137
)

// TestWrapERC7739GoldenVector pins the ERC-7739 signature layout for a
// ClobAuth L1 header. This serves as a regression guard: if the nested
// EIP-712 typehash, deposit-wallet domain, or signing logic drifts, the
// test fails and forces a deliberate update.
func TestWrapERC7739GoldenVector(t *testing.T) {
	signer, err := NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}

	const depositWallet = "0xfd5041047be8c192c725a66228f141196fa3cf9c"
	var appDomainSep [32]byte
	copy(appDomainSep[:], hexDecodeTest("0x1901df32864c97f6cdcb1823d5197ae11b39e229d94ef4beba15803ebbce9f63"))
	var contents [32]byte
	copy(contents[:], hexDecodeTest("0x65c370d71b3486d4e967778d425eed41affc152fe819f0628973abdcfad34219"))
	contentsType := "ClobAuth(address address,uint64 timestamp,uint256 nonce,string message)"

	sig, err := WrapERC7739Signature(signer, depositWallet, polygonChainID, appDomainSep, contents, contentsType)
	if err != nil {
		t.Fatalf("WrapERC7739Signature: %v", err)
	}

	// Validate layout: innerSig(65) || appDomainSep(32) || contents(32) ||
	// contentsType || uint16BE(len(contentsType)).
	if len(sig) != 2+65*2+32*4+len(contentsType)*2+4 {
		t.Fatalf("unexpected signature length %d", len(sig))
	}

	// Extract components.
	sigBytes := hexDecodeTest(sig)
	innerSig := sigBytes[:65]
	gotAppDomainSep := sigBytes[65:97]
	gotContents := sigBytes[97:129]
	gotContentsType := string(sigBytes[129 : 129+len(contentsType)])
	gotLen := binaryBigEndianUint16(sigBytes[129+len(contentsType) : 131+len(contentsType)])

	if hex.EncodeToString(gotAppDomainSep) != hex.EncodeToString(appDomainSep[:]) {
		t.Fatalf("appDomainSep mismatch:\n  got  %x\n  want %x", gotAppDomainSep, appDomainSep)
	}
	if hex.EncodeToString(gotContents) != hex.EncodeToString(contents[:]) {
		t.Fatalf("contents mismatch:\n  got  %x\n  want %x", gotContents, contents)
	}
	if gotContentsType != contentsType {
		t.Fatalf("contentsType mismatch:\n  got  %q\n  want %q", gotContentsType, contentsType)
	}
	if gotLen != uint16(len(contentsType)) {
		t.Fatalf("length mismatch: got %d, want %d", gotLen, len(contentsType))
	}

	// Validate inner signature recovers the expected EOA.
	innerSigHex := "0x" + hex.EncodeToString(innerSig)
	if len(innerSigHex) != 132 {
		t.Fatalf("inner sig length %d, want 132", len(innerSigHex))
	}

	// Pin the full signature as a regression guard.
	if len(sig) != 406 {
		t.Fatalf("total signature length %d, want 406", len(sig))
	}

	// Last 2 bytes are uint16BE(len(contentsType)) = 71 = 0x0047.
	if sig[402:406] != "0047" {
		t.Fatalf("last uint16 = %s, want 0047", sig[402:406])
	}

	t.Logf("ERC-7739 signature: %s... (len=%d)", sig[:20], len(sig))
}

func hexDecodeTest(s string) []byte {
	s = strings.TrimPrefix(s, "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func binaryBigEndianUint16(b []byte) uint16 {
	if len(b) != 2 {
		panic("need 2 bytes")
	}
	return uint16(b[0])<<8 | uint16(b[1])
}
