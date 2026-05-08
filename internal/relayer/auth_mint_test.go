package relayer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMintV2APIKeyHitsCorrectEndpointAndParsesResponse(t *testing.T) {
	var sawMethod, sawPath, sawBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		buf := make([]byte, 64)
		n, _ := r.Body.Read(buf)
		sawBody = string(buf[:n])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiKey":"019e0650-uuid","address":"0xabc","createdAt":"2026-05-08T00:00:00Z","updatedAt":"2026-05-08T00:00:00Z"}`))
	}))
	defer srv.Close()

	key, err := MintV2APIKey(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("MintV2APIKey: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method=%s, want POST", sawMethod)
	}
	if sawPath != "/relayer/api/auth" {
		t.Errorf("path=%q, want /relayer/api/auth", sawPath)
	}
	if sawBody != "{}" {
		t.Errorf("body=%q, want {}", sawBody)
	}
	if key.Key != "019e0650-uuid" || key.Address != "0xabc" {
		t.Errorf("key=%+v", key)
	}
}

func TestMintV2APIKeyReturnsErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	if _, err := MintV2APIKey(context.Background(), srv.Client(), srv.URL); err == nil {
		t.Fatal("expected error on 401")
	}
}

func TestMintV2APIKeyRejectsIncompleteResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiKey":"","address":""}`))
	}))
	defer srv.Close()

	if _, err := MintV2APIKey(context.Background(), srv.Client(), srv.URL); err == nil {
		t.Fatal("expected error for empty key/address")
	}
}

func TestMintV2APIKeyRejectsMissingArgs(t *testing.T) {
	if _, err := MintV2APIKey(context.Background(), nil, "https://x"); err == nil {
		t.Fatal("expected error for nil client")
	}
	if _, err := MintV2APIKey(context.Background(), &http.Client{}, ""); err == nil {
		t.Fatal("expected error for empty relayerURL")
	}
}

func TestV2APIKeyHeadersMatchSDKShape(t *testing.T) {
	key := V2APIKey{Key: "uuid-1", Address: "0xabc"}
	headers := key.V2Headers()
	if headers["RELAYER_API_KEY"] != "uuid-1" {
		t.Errorf("RELAYER_API_KEY=%q", headers["RELAYER_API_KEY"])
	}
	if headers["RELAYER_API_KEY_ADDRESS"] != "0xabc" {
		t.Errorf("RELAYER_API_KEY_ADDRESS=%q", headers["RELAYER_API_KEY_ADDRESS"])
	}
}
