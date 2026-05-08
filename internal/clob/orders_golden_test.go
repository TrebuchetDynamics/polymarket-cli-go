package clob

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// goldenFixture pins one neg-risk variant of the sigtype-3 (POLY_1271,
// deposit wallet) signing path. Since the 2026-04-28 V2 cutover, sigtypes
// 0/1/2 are dead — `clob/order` rejects them — so we only pin sigtype 3.
//
// expectedHash and expectedSig are populated by running the test once with
// placeholder values, capturing the actual outputs, and pasting them in.
// They serve as regression pins: if the V2 wire format drifts (struct
// fields, EIP-712 domain, signing logic), these tests fail and force the
// change to be deliberate.
type goldenFixture struct {
	name         string
	negRisk      bool
	expectedHash string // 0x-prefixed 64-hex EIP-712 struct hash
	expectedSig  string // 0x-prefixed POLY_1271 wrapped signature (636 chars)
}

var goldenFixtures = []goldenFixture{
	{
		name:         "poly1271_regular",
		negRisk:      false,
		expectedHash: "0x5f003d3e8bb3b6e0571bf6a33ca824ba97a222293b53355223ec0a98f7ec17cf",
		expectedSig:  "0x8d13d68146f12b705d3e4dd73f6ac0bb9b4f8c8895d6ff6b82b2c981d039408a302077259f4710c02fd5417ec051471779794fd078971df44fca2df2ea2ce81c1b3264e159346253e26a64e00b69032db0e7d32f94628de3e6eecb50304d7af3d25f003d3e8bb3b6e0571bf6a33ca824ba97a222293b53355223ec0a98f7ec17cf4f726465722875696e743235362073616c742c61646472657373206d616b65722c61646472657373207369676e65722c75696e7432353620746f6b656e49642c75696e74323536206d616b6572416d6f756e742c75696e743235362074616b6572416d6f756e742c75696e743820736964652c75696e7438207369676e6174757265547970652c75696e743235362074696d657374616d702c62797465733332206d657461646174612c62797465733332206275696c6465722900ba",
	},
}

func TestGoldenVectorsV2OrderSigning(t *testing.T) {
	origSalt := orderSalt
	origNow := orderNow
	t.Cleanup(func() {
		orderSalt = origSalt
		orderNow = origNow
	})
	orderSalt = func() (uint64, error) { return 1, nil }
	orderNow = func() time.Time { return time.UnixMilli(1778125000123) }

	signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}

	for _, fx := range goldenFixtures {
		t.Run(fx.name, func(t *testing.T) {
			payload, err := buildSignedOrderPayload(signer, orderDraft{
				tokenID:     big.NewInt(12345),
				side:        "BUY",
				makerAmount: "700000",
				takerAmount: "1400000",
				orderType:   "GTC",
			}, time.UnixMilli(1778125000123), fx.negRisk)
			if err != nil {
				t.Fatal(err)
			}

			// Compute the EIP-712 typed-data hash for the order.
			typedData := buildOrderTypedData(payload, fx.negRisk)
			_, rawDataStr, err := apitypes.TypedDataAndHash(typedData)
			if err != nil {
				t.Fatalf("typed-data hash: %v", err)
			}
			rawData := []byte(rawDataStr)
			gotHash := "0x" + hex.EncodeToString(rawData[34:66]) // structHash slice

			if gotHash != fx.expectedHash {
				t.Fatalf("hash mismatch for %s:\n  got  %s\n  want %s", fx.name, gotHash, fx.expectedHash)
			}
			if payload.Signature != fx.expectedSig {
				t.Fatalf("signature mismatch for %s:\n  got  %s\n  want %s", fx.name, payload.Signature, fx.expectedSig)
			}
		})
	}
}
