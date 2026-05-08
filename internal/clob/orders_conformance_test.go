package clob

import (
	"math/big"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
)

func TestGTC_OrderTypeAccepted(t *testing.T) {
	if got := normalizeOrderType("GTC", ""); got != "GTC" {
		t.Fatalf("GTC not accepted: %s", got)
	}
}

func TestGTD_OrderTypeAccepted(t *testing.T) {
	if got := normalizeOrderType("GTD", ""); got != "GTD" {
		t.Fatalf("GTD not accepted: %s", got)
	}
}

func TestFOK_OrderTypeAccepted(t *testing.T) {
	if got := normalizeOrderType("FOK", ""); got != "FOK" {
		t.Fatalf("FOK not accepted: %s", got)
	}
}

func TestFAK_OrderTypeAccepted(t *testing.T) {
	if got := normalizeOrderType("FAK", ""); got != "FAK" {
		t.Fatalf("FAK not accepted: %s", got)
	}
}

func TestGTD_ExpirationPassesThrough(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	expirationUnix := "1778125000123"
	tokenID := big.NewInt(12345)
	order, err := buildSignedOrderPayload(signer, orderDraft{
		tokenID:     tokenID,
		side:        "BUY",
		makerAmount: "700000",
		takerAmount: "1400000",
		orderType:   "GTD",
		expiration:  expirationUnix,
	}, time.UnixMilli(1778125000123), false)
	if err != nil {
		t.Fatal(err)
	}
	if order.Expiration != expirationUnix {
		t.Fatalf("expiration=%q want %q", order.Expiration, expirationUnix)
	}
}

func TestGTC_DefaultExpirationZero(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	tokenID := big.NewInt(12345)
	order, err := buildSignedOrderPayload(signer, orderDraft{
		tokenID:     tokenID,
		side:        "BUY",
		makerAmount: "700000",
		takerAmount: "1400000",
		orderType:   "GTC",
		expiration:  "0",
	}, time.UnixMilli(1778125000123), false)
	if err != nil {
		t.Fatal(err)
	}
	if order.Expiration != "0" {
		t.Fatalf("GTC expiration=%q want 0", order.Expiration)
	}
}

func TestInvalidOrderType_FallsBack(t *testing.T) {
	if got := normalizeOrderType("INVALID", "GTC"); got != "GTC" {
		t.Fatalf("invalid type not falling back: %s", got)
	}
}

func TestEmptyOrderType_UsesFallback(t *testing.T) {
	if got := normalizeOrderType("", "FOK"); got != "FOK" {
		t.Fatalf("empty type fallback failed: %s", got)
	}
}

func TestCreateOrderParams_HasExpirationField(t *testing.T) {
	p := CreateOrderParams{
		TokenID:    "123",
		Side:       "BUY",
		Price:      "0.5",
		Size:       "10",
		OrderType:  "GTD",
		Expiration: "1778125000123",
	}
	if p.Expiration != "1778125000123" {
		t.Fatalf("Expiration field not set: %s", p.Expiration)
	}
}
