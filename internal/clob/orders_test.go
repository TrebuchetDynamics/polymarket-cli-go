package clob

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

const testOrderPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
const testBuilderCode = "0x1111111111111111111111111111111111111111111111111111111111111111"

func TestBuildSignedOrderPayloadV2UsesCurrentCLOBShape(t *testing.T) {
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
		orderType:   "FOK",
	}, time.UnixMilli(1778125000123), false)
	if err != nil {
		t.Fatal(err)
	}
	if order.Timestamp != "1778125000123" || order.Metadata != bytes32Zero || order.Builder != bytes32Zero || order.Expiration != "0" {
		t.Fatalf("v2 metadata fields not set: %+v", order)
	}
	if !strings.HasPrefix(order.Signature, "0x") || len(order.Signature) != 636 {
		t.Fatalf("signature shape=%q", order.Signature)
	}
	body, err := json.Marshal(order)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"\"taker\"", "\"nonce\"", "\"feeRateBps\""} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("v2 JSON contains %s: %s", forbidden, body)
		}
	}
}

func TestBuildSignedOrderPayloadUsesConfiguredBuilderCode(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	order, err := buildSignedOrderPayload(signer, orderDraft{
		tokenID:     big.NewInt(12345),
		side:        "BUY",
		makerAmount: "700000",
		takerAmount: "1400000",
		orderType:   "FOK",
		builderCode: testBuilderCode,
	}, time.UnixMilli(1778125000123), false)
	if err != nil {
		t.Fatal(err)
	}
	if order.Builder != testBuilderCode {
		t.Fatalf("builder=%q want %q", order.Builder, testBuilderCode)
	}
}

func TestBuildSignedOrderPayloadRejectsInvalidBuilderCode(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = buildSignedOrderPayload(signer, orderDraft{
		tokenID:     big.NewInt(12345),
		side:        "BUY",
		makerAmount: "700000",
		takerAmount: "1400000",
		orderType:   "FOK",
		builderCode: "0x1234",
	}, time.UnixMilli(1778125000123), false)
	if err == nil || !strings.Contains(err.Error(), "builder") {
		t.Fatalf("expected builder validation error, got %v", err)
	}
}

func TestBuildSignedOrderPayloadV2DepositWalletUsesEOASignerWithDepositMaker(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	order, err := buildSignedOrderPayload(signer, orderDraft{
		tokenID:     big.NewInt(12345),
		side:        "BUY",
		makerAmount: "700000",
		takerAmount: "1400000",
		orderType:   "FOK",
	}, time.UnixMilli(1778125000123), false)
	if err != nil {
		t.Fatal(err)
	}
	wantMaker := "0xfd5041047be8c192c725a66228f141196fa3cf9c"
	if !strings.EqualFold(order.Maker, wantMaker) || !strings.EqualFold(order.Signer, wantMaker) {
		t.Fatalf("maker/signer=%s/%s want deposit wallet %s", order.Maker, order.Signer, wantMaker)
	}
	if order.SignatureType != signatureTypePoly1271 {
		t.Fatalf("signature type=%d", order.SignatureType)
	}
	if len(order.Signature) != 636 {
		t.Fatalf("wrapped signature length=%d want 636", len(order.Signature))
	}
}

func TestCreateMarketOrderPostsV2PayloadWhenCLOBVersionIsTwo(t *testing.T) {
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tick-size":
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.001"}`))
		case "/fee-rate":
			_, _ = w.Write([]byte(`{"fee_rate_bps":0}`))
		case "/neg-risk":
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/order":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode order body: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true,"orderID":"0xabc","status":"matched"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	res, err := client.CreateMarketOrder(context.Background(), testOrderPrivateKey, MarketOrderParams{
		TokenID:   "12345",
		Side:      "buy",
		Amount:    "0.700000",
		Price:     "0.500000",
		OrderType: "FOK",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.OrderID != "0xabc" {
		t.Fatalf("response=%+v", res)
	}
	order, ok := posted["order"].(map[string]any)
	if !ok {
		t.Fatalf("posted order missing: %#v", posted)
	}
	for _, want := range []string{"timestamp", "metadata", "builder", "expiration", "signature"} {
		if _, ok := order[want]; !ok {
			t.Fatalf("posted v2 order missing %q: %#v", want, order)
		}
	}
	for _, forbidden := range []string{"taker", "nonce", "feeRateBps"} {
		if _, ok := order[forbidden]; ok {
			t.Fatalf("posted v2 order contains %q: %#v", forbidden, order)
		}
	}
	if posted["postOnly"] != false || posted["deferExec"] != false {
		t.Fatalf("post flags not explicit false: %#v", posted)
	}
}

func TestCreateLimitOrderPostsConfiguredBuilderCode(t *testing.T) {
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tick-size":
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.001"}`))
		case "/fee-rate":
			_, _ = w.Write([]byte(`{"fee_rate_bps":0}`))
		case "/neg-risk":
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/order":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode order body: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true,"orderID":"0xabc","status":"matched"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)
	client.SetBuilderCode(testBuilderCode)

	_, err := client.CreateLimitOrder(context.Background(), testOrderPrivateKey, CreateOrderParams{
		TokenID:   "12345",
		Side:      "buy",
		Price:     "0.500000",
		Size:      "1.400000",
		OrderType: "GTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	order, ok := posted["order"].(map[string]any)
	if !ok {
		t.Fatalf("posted order missing: %#v", posted)
	}
	if order["builder"] != testBuilderCode {
		t.Fatalf("posted builder=%#v want %s", order["builder"], testBuilderCode)
	}
}

func TestOrderQueriesSingleOrderByID(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/order/0xabc":
			gotPath = r.URL.Path
			if r.Method != http.MethodGet {
				t.Fatalf("method=%s want GET", r.Method)
			}
			_, _ = w.Write([]byte(`{"id":"0xabc","status":"LIVE","market":"0xmarket"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	order, err := client.Order(context.Background(), testOrderPrivateKey, "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/order/0xabc" || order.ID != "0xabc" || order.Status != "LIVE" {
		t.Fatalf("path=%q order=%+v", gotPath, order)
	}
}

func TestCancelOrdersDeletesBatchEndpointWithOrderIDs(t *testing.T) {
	var posted map[string][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/orders":
			if r.Method != http.MethodDelete {
				t.Fatalf("method=%s want DELETE", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode cancel body: %v", err)
			}
			_, _ = w.Write([]byte(`{"canceled":["0x1","0x2"],"not_canceled":{}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	res, err := client.CancelOrders(context.Background(), testOrderPrivateKey, []string{"0x1", "0x2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(posted["orderIDs"]) != 2 || len(res.Canceled) != 2 {
		t.Fatalf("posted=%#v response=%+v", posted, res)
	}
}

func TestCancelMarketDeletesMarketEndpointWithFilters(t *testing.T) {
	var posted map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/cancel-market-orders":
			if r.Method != http.MethodDelete {
				t.Fatalf("method=%s want DELETE", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode cancel market body: %v", err)
			}
			_, _ = w.Write([]byte(`{"canceled":["0x1"],"not_canceled":{}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	res, err := client.CancelMarket(context.Background(), testOrderPrivateKey, CancelMarketParams{
		Market: "0xmarket",
		Asset:  "123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if posted["market"] != "0xmarket" || posted["asset_id"] != "123" || len(res.Canceled) != 1 {
		t.Fatalf("posted=%#v response=%+v", posted, res)
	}
}

func TestSignCLOBOrderUsesNegRiskExchangeAddressWhenFlagged(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testOrderPrivateKey, polygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	payload := signedOrderPayload{
		Salt:          1,
		Maker:         signer.Address(),
		Signer:        signer.Address(),
		TokenID:       "12345",
		MakerAmount:   "700000",
		TakerAmount:   "1400000",
		Side:          "BUY",
		SignatureType: signatureTypePoly1271,
		Timestamp:     "1778125000123",
		Metadata:      bytes32Zero,
		Builder:       bytes32Zero,
	}
	sigRegular, err := signCLOBOrder(signer, payload, false)
	if err != nil {
		t.Fatal(err)
	}
	sigNegRisk, err := signCLOBOrder(signer, payload, true)
	if err != nil {
		t.Fatal(err)
	}
	if sigRegular == sigNegRisk {
		t.Fatalf("regular and neg-risk signatures must differ; both = %q", sigRegular)
	}
}
