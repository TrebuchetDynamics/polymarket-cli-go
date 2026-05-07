package clob

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestCreateOrDeriveAPIKeyUsesL1Auth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s, want POST", r.Method)
		}
		if r.URL.Path != "/auth/api-key" {
			t.Fatalf("path=%s, want /auth/api-key", r.URL.Path)
		}
		for _, header := range []string{"POLY_ADDRESS", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_NONCE"} {
			if strings.TrimSpace(r.Header.Get(header)) == "" {
				t.Fatalf("missing %s header", header)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiKey":"key-123","secret":"secret-123","passphrase":"pass-123"}`))
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)
	key, err := client.CreateOrDeriveAPIKey(context.Background(), "0x1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		t.Fatalf("CreateOrDeriveAPIKey returned error: %v", err)
	}
	if key.Key != "key-123" || key.Secret != "secret-123" || key.Passphrase != "pass-123" {
		t.Fatalf("unexpected key: %#v", key)
	}
}

func TestBalanceAllowanceUsesL2Auth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/derive-api-key":
			for _, header := range []string{"POLY_ADDRESS", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_NONCE"} {
				if strings.TrimSpace(r.Header.Get(header)) == "" {
					t.Fatalf("missing %s header", header)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"apiKey":"key-123","secret":"c2VjcmV0","passphrase":"pass-123"}`))
		case "/balance-allowance":
			if r.URL.Query().Get("asset_type") != "COLLATERAL" {
				t.Fatalf("asset_type=%q", r.URL.Query().Get("asset_type"))
			}
			if r.URL.Query().Get("signature_type") != "1" {
				t.Fatalf("signature_type=%q", r.URL.Query().Get("signature_type"))
			}
			for _, header := range []string{"POLY_ADDRESS", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_API_KEY", "POLY_PASSPHRASE"} {
				if strings.TrimSpace(r.Header.Get(header)) == "" {
					t.Fatalf("missing %s header", header)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"balance":"12.34","allowances":{"0xspender":"1000"}}`))
		default:
			t.Fatalf("unexpected path=%s", r.URL.Path)
		}
	}))
	defer server.Close()

	tc := transport.New(server.Client(), transport.DefaultConfig(server.URL+"/"))
	client := NewClient(server.URL+"/", tc)
	bal, err := client.BalanceAllowance(context.Background(), "0x1111111111111111111111111111111111111111111111111111111111111111", BalanceAllowanceParams{
		AssetType:     "COLLATERAL",
		SignatureType: 1,
	})
	if err != nil {
		t.Fatalf("BalanceAllowance returned error: %v", err)
	}
	if bal.Balance != "12.34" || bal.Allowances["0xspender"] != "1000" {
		t.Fatalf("unexpected balance: %#v", bal)
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
