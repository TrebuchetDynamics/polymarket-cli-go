package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testSIWEFixedNonce = "test-nonce-abc123"

func newSIWETestServer(t *testing.T, captureBearer *string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/nonce", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("/nonce method=%s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"nonce": testSIWEFixedNonce})
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("/login method=%s, want GET", r.Method)
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("/login missing Bearer token: %q", auth)
		}
		if captureBearer != nil {
			*captureBearer = strings.TrimPrefix(auth, "Bearer ")
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "polymarketSession",
			Value: "session-cookie-value",
			Path:  "/",
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"type":    "EOA",
			"address": "0x9d8A62f656a8d1615C1294fd71e9CFb3E4855A4F",
		})
	})
	return httptest.NewServer(mux)
}

func TestSIWELoginGetsNonceSignsAndPersistsCookie(t *testing.T) {
	var bearer string
	srv := newSIWETestServer(t, &bearer)
	defer srv.Close()

	signer, err := NewPrivateKeySigner(siweTestPrivateKey, 137)
	if err != nil {
		t.Fatal(err)
	}
	fixedTime := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)

	session, err := NewSIWESession(signer, srv.URL, WithSIWEClock(func() time.Time { return fixedTime }))
	if err != nil {
		t.Fatal(err)
	}

	if err := session.Login(context.Background()); err != nil {
		t.Fatalf("Login: %v", err)
	}

	// Bearer token should decode to {fields}:::sig.
	rawToken, err := base64.StdEncoding.DecodeString(bearer)
	if err != nil {
		t.Fatalf("bearer not base64: %v", err)
	}
	combined := string(rawToken)
	if !strings.Contains(combined, `"nonce":"`+testSIWEFixedNonce+`"`) {
		t.Errorf("bearer JSON missing nonce: %s", combined)
	}
	if !strings.Contains(combined, ":::0x") {
		t.Errorf("bearer missing :::0x sig separator: %s", combined)
	}
	if !strings.Contains(combined, `"chainId":137`) {
		t.Errorf("bearer missing chainId: %s", combined)
	}

	// Cookie jar should now hold the session cookie.
	cookies := session.CookiesFor(srv.URL + "/")
	var found bool
	for _, c := range cookies {
		if c.Name == "polymarketSession" && c.Value == "session-cookie-value" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("polymarketSession cookie missing; got %d cookies", len(cookies))
	}
}

func TestSIWELoginFailsWhenNonceServerErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"nope"}`, http.StatusInternalServerError)
	}))
	defer srv.Close()

	signer, err := NewPrivateKeySigner(siweTestPrivateKey, 137)
	if err != nil {
		t.Fatal(err)
	}
	session, err := NewSIWESession(signer, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Login(context.Background()); err == nil {
		t.Fatal("expected error when /nonce returns 500")
	}
}

func TestSIWELoginRejectsEmptyNonce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"nonce": ""})
	}))
	defer srv.Close()

	signer, err := NewPrivateKeySigner(siweTestPrivateKey, 137)
	if err != nil {
		t.Fatal(err)
	}
	session, err := NewSIWESession(signer, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Login(context.Background()); err == nil {
		t.Fatal("expected error when nonce is empty")
	}
}

func TestNewSIWESessionRejectsMissingArgs(t *testing.T) {
	if _, err := NewSIWESession(nil, "https://x"); err == nil {
		t.Fatal("expected error for nil signer")
	}
	signer, _ := NewPrivateKeySigner(siweTestPrivateKey, 137)
	if _, err := NewSIWESession(signer, ""); err == nil {
		t.Fatal("expected error for empty gammaURL")
	}
}
