package clob

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
)

func TestOrderBookGetUsesReadOnlyEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/book" {
			t.Fatalf("path = %q, want /book", r.URL.Path)
		}
		if r.URL.Query().Get("token_id") != "token-1" {
			t.Fatalf("token_id query = %q, want token-1", r.URL.Query().Get("token_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"market":"token-1","bids":[{"price":"0.40","size":"12"}],"asks":[{"price":"0.60","size":"8"}]}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)
	book, err := client.OrderBook(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("OrderBook returned error: %v", err)
	}
	if book.Market != "token-1" {
		t.Fatalf("Market = %q, want token-1", book.Market)
	}
}

func TestMarketByTokenCallsV2Endpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/markets-by-token/token-1" {
			t.Fatalf("path = %q, want /markets-by-token/token-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"condition_id":"condition-1",
			"primary_token_id":"token-yes",
			"secondary_token_id":"token-no"
		}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	got, err := client.MarketByToken(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("MarketByToken returned error: %v", err)
	}
	if got.ConditionID != "condition-1" || got.PrimaryTokenID != "token-yes" || got.SecondaryTokenID != "token-no" {
		t.Fatalf("unexpected market-by-token response: %+v", got)
	}
}

func TestPriceCallsCorrectEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token_id") != "abc" {
			t.Fatalf("token_id = %q, want abc", r.URL.Query().Get("token_id"))
		}
		if r.URL.Query().Get("side") != "BUY" {
			t.Fatalf("side = %q, want BUY", r.URL.Query().Get("side"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"price":"0.52"}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	price, err := client.Price(context.Background(), "abc", "BUY")
	if err != nil {
		t.Fatalf("Price returned error: %v", err)
	}
	if price != "0.52" {
		t.Fatalf("Price = %q, want 0.52", price)
	}
}

func TestPricesDoesNotFallbackToLegacyEndpoint(t *testing.T) {
	var sawLegacy bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/prices":
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
		case "/prices-post":
			sawLegacy = true
			http.Error(w, "legacy endpoint must not be called", http.StatusTeapot)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	_, err := client.Prices(context.Background(), []polytypes.BookParams{{TokenID: "token-1", Side: "BUY"}})
	if err == nil {
		t.Fatal("expected /prices error")
	}
	if sawLegacy {
		t.Fatal("Prices called legacy /prices-post fallback; V2-only clients must not")
	}
}

func TestMidpointCallsCorrectEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/midpoint" {
			t.Fatalf("path = %q, want /midpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"mid":"0.50"}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	mid, err := client.Midpoint(context.Background(), "tok")
	if err != nil {
		t.Fatalf("Midpoint returned error: %v", err)
	}
	if mid != "0.50" {
		t.Fatalf("Midpoint = %q, want 0.50", mid)
	}
}

func TestServerTime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"timestamp":"1234567890","iso":"2026-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	st, err := client.ServerTime(context.Background())
	if err != nil {
		t.Fatalf("ServerTime returned error: %v", err)
	}
	if st.Timestamp != "1234567890" {
		t.Fatalf("Timestamp = %q, want 1234567890", st.Timestamp)
	}
}

func TestTickSizeCallsCorrectEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tick-size" {
			t.Fatalf("path = %q, want /tick-size", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"minimum_tick_size":"0.01","minimum_order_size":"5","tick_size":"0.01"}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	ts, err := client.TickSize(context.Background(), "tok")
	if err != nil {
		t.Fatalf("TickSize returned error: %v", err)
	}
	if ts.MinimumTickSize != "0.01" {
		t.Fatalf("MinimumTickSize = %q, want 0.01", ts.MinimumTickSize)
	}
}

const builderFeeKeyTestPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

// EOA derived from builderFeeKeyTestPrivateKey. Per the 2026-05-08 web-UI
// capture, V2 CLOB authentication is EOA-bound at the HTTP layer (POLY_ADDRESS
// is the EOA, not the deposit wallet).
const builderFeeKeyTestEOA = "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"

// builderFeeKeyServer mounts the L2-derive endpoint that every authenticated
// builder-fee call needs alongside one extra handler for the test target.
func builderFeeKeyServer(t *testing.T, target string, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"l2-key","secret":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","passphrase":"l2-pass"}`))
		case target:
			handler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestCreateBuilderFeeKeyHitsCorrectEndpointWithL2Headers(t *testing.T) {
	var sawHeaders http.Header
	server := builderFeeKeyServer(t, "/auth/builder-api-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		sawHeaders = r.Header.Clone()
		_, _ = w.Write([]byte(`{"key":"fee-key-uuid","secret":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","passphrase":"fee-pass"}`))
	})
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	feeKey, err := client.CreateBuilderFeeKey(context.Background(), builderFeeKeyTestPrivateKey)
	if err != nil {
		t.Fatalf("CreateBuilderFeeKey returned error: %v", err)
	}
	if feeKey.Key != "fee-key-uuid" {
		t.Fatalf("Key = %q, want fee-key-uuid", feeKey.Key)
	}
	for _, want := range []string{"POLY_API_KEY", "POLY_PASSPHRASE", "POLY_TIMESTAMP", "POLY_SIGNATURE", "POLY_ADDRESS"} {
		if v := sawHeaders.Get(want); v == "" {
			t.Errorf("missing %s header", want)
		}
	}
}

func TestListBuilderFeeKeysHitsPluralPath(t *testing.T) {
	var sawMethod string
	server := builderFeeKeyServer(t, "/auth/builder-api-keys", func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		_, _ = w.Write([]byte(`[{"key":"fee-1","created_at":"2026-05-08T00:00:00Z"},{"key":"fee-2"}]`))
	})
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	rows, err := client.ListBuilderFeeKeys(context.Background(), builderFeeKeyTestPrivateKey)
	if err != nil {
		t.Fatalf("ListBuilderFeeKeys returned error: %v", err)
	}
	if sawMethod != http.MethodGet {
		t.Fatalf("method = %s, want GET", sawMethod)
	}
	if len(rows) != 2 || rows[0].Key != "fee-1" || rows[1].Key != "fee-2" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestRevokeBuilderFeeKeyHitsScopedPath(t *testing.T) {
	var sawMethod, sawPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/auth/derive-api-key":
			_, _ = w.Write([]byte(`{"apiKey":"l2-key","secret":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","passphrase":"l2-pass"}`))
		default:
			sawMethod, sawPath = r.Method, r.URL.Path
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	if err := client.RevokeBuilderFeeKey(context.Background(), builderFeeKeyTestPrivateKey, "fee-1"); err != nil {
		t.Fatalf("RevokeBuilderFeeKey returned error: %v", err)
	}
	if sawMethod != http.MethodDelete {
		t.Fatalf("method = %s, want DELETE", sawMethod)
	}
	if sawPath != "/auth/builder-api-key/fee-1" {
		t.Fatalf("path = %q, want /auth/builder-api-key/fee-1", sawPath)
	}
}

func TestRevokeBuilderFeeKeyRejectsEmpty(t *testing.T) {
	tc := transport.New(nil, transport.DefaultConfig("http://invalid.local/"))
	client := NewClient("http://invalid.local/", tc)
	if err := client.RevokeBuilderFeeKey(context.Background(), builderFeeKeyTestPrivateKey, "  "); err == nil {
		t.Fatal("expected error for empty builderKey")
	}
}

func TestCreateAPIKeyForAddressIgnoresOwnerAndSignsWithEOA(t *testing.T) {
	// Per the 2026-05-08 web-UI capture, the CLOB API key is EOA-bound at
	// the HTTP layer regardless of which maker (proxy or deposit wallet)
	// the user signs orders with. CreateAPIKeyForAddress retains its
	// ownerAddress parameter for source-compat, but POLY_ADDRESS is the
	// EOA derived from the private key.
	var sawAddress string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/api-key" || r.Method != http.MethodPost {
			t.Fatalf("path=%s method=%s", r.URL.Path, r.Method)
		}
		sawAddress = r.Header.Get("POLY_ADDRESS")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiKey":"k","secret":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","passphrase":"p"}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)
	deposit := "0x19bE70b1e4F59C0663a999C0dC6f5b3C68CFCaF3"

	key, err := client.CreateAPIKeyForAddress(context.Background(), builderFeeKeyTestPrivateKey, deposit)
	if err != nil {
		t.Fatalf("CreateAPIKeyForAddress: %v", err)
	}
	if !strings.EqualFold(sawAddress, builderFeeKeyTestEOA) {
		t.Fatalf("POLY_ADDRESS = %s, want EOA %s (deposit %s should be ignored)", sawAddress, builderFeeKeyTestEOA, deposit)
	}
	if key.Key != "k" {
		t.Fatalf("Key = %s", key.Key)
	}
}

func TestBalanceAllowanceUsesEOABoundL2Auth(t *testing.T) {
	// Per the 2026-05-08 web-UI capture: HTTP-layer auth (POLY_ADDRESS,
	// HMAC) is EOA-bound. Deposit-wallet identity rides on the
	// signature_type=3 query param, not POLY_ADDRESS.
	var deriveAddress string
	var balanceAddress string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/derive-api-key":
			deriveAddress = r.Header.Get("POLY_ADDRESS")
			_, _ = w.Write([]byte(`{"apiKey":"deposit-key","secret":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","passphrase":"pass"}`))
		case "/balance-allowance":
			balanceAddress = r.Header.Get("POLY_ADDRESS")
			if got := r.URL.Query().Get("signature_type"); got != "3" {
				t.Fatalf("signature_type=%q want 3", got)
			}
			_, _ = w.Write([]byte(`{"balance":"1000000","allowance":"999"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	resp, err := client.BalanceAllowance(context.Background(), builderFeeKeyTestPrivateKey, BalanceAllowanceParams{
		AssetType: "COLLATERAL",
	})
	if err != nil {
		t.Fatalf("BalanceAllowance returned error: %v", err)
	}
	if resp.Balance != "1000000" {
		t.Fatalf("balance=%q", resp.Balance)
	}
	if !strings.EqualFold(deriveAddress, builderFeeKeyTestEOA) {
		t.Fatalf("derive POLY_ADDRESS=%s want EOA %s", deriveAddress, builderFeeKeyTestEOA)
	}
	if !strings.EqualFold(balanceAddress, builderFeeKeyTestEOA) {
		t.Fatalf("balance POLY_ADDRESS=%s want EOA %s", balanceAddress, builderFeeKeyTestEOA)
	}
}

func TestUpdateBalanceAllowanceTreatsEmptyBodyAsSuccess(t *testing.T) {
	// Live behavior (verified 2026-05-08): /balance-allowance/update returns
	// HTTP 200 with an empty body — the endpoint just queues the refresh.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"apiKey":"deposit-key","secret":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","passphrase":"pass"}`))
		case "/balance-allowance/update":
			if got := r.URL.Query().Get("signature_type"); got != "3" {
				t.Fatalf("signature_type=%q want 3", got)
			}
			w.WriteHeader(http.StatusOK) // empty body
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)

	resp, err := client.UpdateBalanceAllowance(context.Background(), builderFeeKeyTestPrivateKey, BalanceAllowanceParams{
		AssetType: "COLLATERAL",
	})
	if err != nil {
		t.Fatalf("UpdateBalanceAllowance returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("UpdateBalanceAllowance returned nil response")
	}
}
