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

func TestCreateLimitOrderUsesDepositWalletBoundL2Auth(t *testing.T) {
	wantDepositWallet := "0xfd5041047be8c192c725a66228f141196fa3cf9c"
	var deriveAddress string
	var orderAddress string
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tick-size":
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.001","minimum_order_size":"1"}`))
		case "/neg-risk":
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case "/auth/derive-api-key":
			deriveAddress = r.Header.Get("POLY_ADDRESS")
			_, _ = w.Write([]byte(`{"apiKey":"deposit-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/order":
			orderAddress = r.Header.Get("POLY_ADDRESS")
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode order body: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true,"orderID":"0xabc","status":"live"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	_, err := client.CreateLimitOrder(context.Background(), testOrderPrivateKey, CreateOrderParams{
		TokenID:   "12345",
		Side:      "buy",
		Price:     "0.500000",
		Size:      "2.000000",
		OrderType: "GTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(deriveAddress, wantDepositWallet) {
		t.Fatalf("derive POLY_ADDRESS=%s want deposit wallet %s", deriveAddress, wantDepositWallet)
	}
	if !strings.EqualFold(orderAddress, wantDepositWallet) {
		t.Fatalf("order POLY_ADDRESS=%s want deposit wallet %s", orderAddress, wantDepositWallet)
	}
	order, ok := posted["order"].(map[string]any)
	if !ok {
		t.Fatalf("posted order missing: %#v", posted)
	}
	if !strings.EqualFold(order["maker"].(string), wantDepositWallet) || !strings.EqualFold(order["signer"].(string), wantDepositWallet) {
		t.Fatalf("maker/signer=%v/%v want deposit wallet %s", order["maker"], order["signer"], wantDepositWallet)
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

func TestCreateLimitOrderWithPostOnlyGTCIncludesPostOnlyInPayload(t *testing.T) {
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tick-size":
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.001"}`))
		case "/neg-risk":
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/order":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode order body: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true,"orderID":"0xabc","status":"live"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	_, err := client.CreateLimitOrder(context.Background(), testOrderPrivateKey, CreateOrderParams{
		TokenID:   "12345",
		Side:      "buy",
		Price:     "0.500000",
		Size:      "2.000000",
		OrderType: "GTC",
		PostOnly:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if posted["postOnly"] != true {
		t.Fatalf("postOnly=%v want true", posted["postOnly"])
	}
}

func TestCreateLimitOrderWithPostOnlyFOKRejectsValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tick-size":
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.001"}`))
		case "/neg-risk":
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	_, err := client.CreateLimitOrder(context.Background(), testOrderPrivateKey, CreateOrderParams{
		TokenID:   "12345",
		Side:      "buy",
		Price:     "0.500000",
		Size:      "2.000000",
		OrderType: "FOK",
		PostOnly:  true,
	})
	if err == nil {
		t.Fatal("expected error for PostOnly with FOK order type")
	}
	if !strings.Contains(err.Error(), "postOnly") {
		t.Fatalf("expected postOnly validation error, got %v", err)
	}
}

func TestCreateLimitOrderWithPostOnlyGTDSucceeds(t *testing.T) {
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tick-size":
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.001"}`))
		case "/neg-risk":
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/order":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode order body: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true,"orderID":"0xabc","status":"live"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	_, err := client.CreateLimitOrder(context.Background(), testOrderPrivateKey, CreateOrderParams{
		TokenID:    "12345",
		Side:       "buy",
		Price:      "0.500000",
		Size:       "2.000000",
		OrderType:  "GTD",
		Expiration: "1778125000",
		PostOnly:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if posted["postOnly"] != true {
		t.Fatalf("postOnly=%v want true", posted["postOnly"])
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

func TestCancelOrderUsesDepositWalletBoundL2Auth(t *testing.T) {
	wantDepositWallet := "0xfd5041047be8c192c725a66228f141196fa3cf9c"
	var deriveAddress string
	var cancelAddress string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			deriveAddress = r.Header.Get("POLY_ADDRESS")
			_, _ = w.Write([]byte(`{"apiKey":"deposit-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/order":
			cancelAddress = r.Header.Get("POLY_ADDRESS")
			if r.Method != http.MethodDelete {
				t.Fatalf("method=%s want DELETE", r.Method)
			}
			_, _ = w.Write([]byte(`{"canceled":["0x1"],"not_canceled":{}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	res, err := client.CancelOrder(context.Background(), testOrderPrivateKey, "0x1")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Canceled) != 1 {
		t.Fatalf("response=%+v", res)
	}
	if !strings.EqualFold(deriveAddress, wantDepositWallet) {
		t.Fatalf("derive POLY_ADDRESS=%s want deposit wallet %s", deriveAddress, wantDepositWallet)
	}
	if !strings.EqualFold(cancelAddress, wantDepositWallet) {
		t.Fatalf("cancel POLY_ADDRESS=%s want deposit wallet %s", cancelAddress, wantDepositWallet)
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

func TestCreateBatchOrdersPostsArrayToOrdersEndpoint(t *testing.T) {
	var posted []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tick-size":
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.001"}`))
		case "/neg-risk":
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/orders":
			if r.Method != http.MethodPost {
				t.Fatalf("method=%s want POST", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode batch body: %v", err)
			}
			_, _ = w.Write([]byte(`[{"success":true,"orderID":"0xabc","status":"live"},{"success":true,"orderID":"0xdef","status":"live"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	res, err := client.CreateBatchOrders(context.Background(), testOrderPrivateKey, []CreateOrderParams{
		{TokenID: "12345", Side: "buy", Price: "0.500000", Size: "2.000000", OrderType: "GTC"},
		{TokenID: "12346", Side: "sell", Price: "0.600000", Size: "3.000000", OrderType: "GTC"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(posted) != 2 {
		t.Fatalf("posted %d orders, want 2", len(posted))
	}
	if len(res.Orders) != 2 {
		t.Fatalf("response orders=%d want 2", len(res.Orders))
	}
	if res.Orders[0].OrderID != "0xabc" || res.Orders[1].OrderID != "0xdef" {
		t.Fatalf("order IDs=%v", []string{res.Orders[0].OrderID, res.Orders[1].OrderID})
	}
}

func TestCreateBatchOrdersRejectsEmptyBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	_, err := client.CreateBatchOrders(context.Background(), testOrderPrivateKey, nil)
	if err == nil {
		t.Fatal("expected error for empty batch")
	}
}

func TestCreateBatchOrdersRejectsOversizedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	params := make([]CreateOrderParams, MaxBatchPostSize+1)
	for i := range params {
		params[i] = CreateOrderParams{TokenID: "12345", Side: "buy", Price: "0.5", Size: "1"}
	}
	_, err := client.CreateBatchOrders(context.Background(), testOrderPrivateKey, params)
	if err == nil {
		t.Fatal("expected error for oversized batch")
	}
	if !strings.Contains(err.Error(), "15") {
		t.Fatalf("expected max size error, got %v", err)
	}
}

func TestHeartbeatPostsToHeartbeatsEndpoint(t *testing.T) {
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/v1/heartbeats":
			if r.Method != http.MethodPost {
				t.Fatalf("method=%s want POST", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode heartbeat body: %v", err)
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	if err := client.Heartbeat(context.Background(), testOrderPrivateKey, ""); err != nil {
		t.Fatal(err)
	}
	if posted["heartbeat_id"] != nil {
		t.Fatalf("expected nil heartbeat_id, got %v", posted["heartbeat_id"])
	}
}

func TestHeartbeatWithIDIncludesIDInBody(t *testing.T) {
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/v1/heartbeats":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode heartbeat body: %v", err)
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	if err := client.Heartbeat(context.Background(), testOrderPrivateKey, "hb-123"); err != nil {
		t.Fatal(err)
	}
	if posted["heartbeat_id"] != "hb-123" {
		t.Fatalf("heartbeat_id=%v want hb-123", posted["heartbeat_id"])
	}
}

func TestAutoHeartbeatSendsMultiplePings(t *testing.T) {
	var count int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/v1/heartbeats":
			count++
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	cancel := client.AutoHeartbeat(context.Background(), testOrderPrivateKey, 50*time.Millisecond)
	defer cancel()

	time.Sleep(180 * time.Millisecond)
	if count < 2 {
		t.Fatalf("expected at least 2 heartbeats, got %d", count)
	}
}

func TestAutoHeartbeatCancelStopsPings(t *testing.T) {
	var count int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"owner-key","secret":"c2VjcmV0","passphrase":"pass"}`))
		case "/v1/heartbeats":
			count++
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	cancel := client.AutoHeartbeat(context.Background(), testOrderPrivateKey, 50*time.Millisecond)
	time.Sleep(120 * time.Millisecond)
	cancel()

	before := count
	time.Sleep(120 * time.Millisecond)
	if count != before {
		t.Fatalf("heartbeat count changed after cancel: before=%d after=%d", before, count)
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
