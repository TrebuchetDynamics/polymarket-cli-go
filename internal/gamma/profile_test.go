package gamma

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewCreateProfileRequestMatchesCapturedShape(t *testing.T) {
	const (
		eoa   = "0x3075Af8096c3e5147af22Da45FE7c8496E70a306"
		proxy = "0x31bd9F0E315586352eb9B6141cEC154C9a71549D"
	)
	req := NewCreateProfileRequest(eoa, proxy, "metamask", 1778263464608)

	if req.Name != "0x31bd9F0E315586352eb9B6141cEC154C9a71549D-1778263464608" {
		t.Errorf("Name: got %q", req.Name)
	}
	if req.Pseudonym != proxy {
		t.Errorf("Pseudonym: got %q want %q", req.Pseudonym, proxy)
	}
	if req.ProxyWallet != proxy {
		t.Errorf("ProxyWallet: got %q want %q", req.ProxyWallet, proxy)
	}
	if !req.DisplayUsernamePublic || req.EmailOptIn || req.WalletActivated {
		t.Errorf("default flags wrong: %+v", req)
	}
	if len(req.Users) != 1 {
		t.Fatalf("Users: got %d want 1", len(req.Users))
	}
	u := req.Users[0]
	if u.Address != eoa {
		t.Errorf("Users[0].Address: got %q want %q", u.Address, eoa)
	}
	if u.ProxyWallet != proxy {
		t.Errorf("Users[0].ProxyWallet: got %q", u.ProxyWallet)
	}
	if !u.IsExternalAuth {
		t.Errorf("Users[0].IsExternalAuth: want true")
	}
	if u.Provider != "metamask" {
		t.Errorf("Users[0].Provider: got %q", u.Provider)
	}
	if len(u.Preferences) != 1 || len(u.WalletPreferences) != 1 {
		t.Errorf("Users[0] missing preferences blocks: %+v", u)
	}
}

func TestCreateProfileSendsSIWECookieAndExpectedBody(t *testing.T) {
	const (
		eoa   = "0x3075Af8096c3e5147af22Da45FE7c8496E70a306"
		proxy = "0x21999a074344610057c9b2B362332388a44502D4" // sigtype-3 deposit wallet
	)
	var (
		gotPath    string
		gotMethod  string
		gotCT      string
		gotCookie  string
		gotPayload CreateProfileRequest
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		if c, err := r.Cookie("polymarket-session"); err == nil {
			gotCookie = c.Value
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotPayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"id":"8040942",
			"name":"polygolem-test",
			"proxyWallet":"0x21999a074344610057c9b2b362332388a44502d4",
			"pseudonym":"Limping-Soul"
		}`))
	}))
	defer srv.Close()

	jar, _ := cookiejar.New(nil)
	parsedURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse srv URL: %v", err)
	}
	jar.SetCookies(parsedURL, []*http.Cookie{{Name: "polymarket-session", Value: "siwe-cookie"}})
	client := &http.Client{Jar: jar}

	body := NewCreateProfileRequest(eoa, proxy, "metamask", 1778263464608)
	resp, err := CreateProfile(context.Background(), client, srv.URL, body)
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if resp.ID != "8040942" {
		t.Errorf("response ID: got %q", resp.ID)
	}
	if resp.ProxyWallet != "0x21999a074344610057c9b2b362332388a44502d4" {
		t.Errorf("response proxyWallet: got %q", resp.ProxyWallet)
	}

	if gotPath != "/profiles" {
		t.Errorf("path: got %q want /profiles", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method: got %q want POST", gotMethod)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type: got %q", gotCT)
	}
	if gotCookie != "siwe-cookie" {
		t.Errorf("missing/wrong SIWE cookie: %q", gotCookie)
	}
	if gotPayload.ProxyWallet != proxy {
		t.Errorf("payload.ProxyWallet: got %q want %q", gotPayload.ProxyWallet, proxy)
	}
	if len(gotPayload.Users) != 1 || gotPayload.Users[0].Address != eoa {
		t.Errorf("payload.Users[0].Address: got %+v", gotPayload.Users)
	}
	if gotPayload.Users[0].Provider != "metamask" {
		t.Errorf("payload.Users[0].Provider: got %q", gotPayload.Users[0].Provider)
	}
}

func TestCreateProfileRejectsInvalidArgs(t *testing.T) {
	_, err := CreateProfile(context.Background(), nil, "x", CreateProfileRequest{})
	if err == nil || !strings.Contains(err.Error(), "client is required") {
		t.Errorf("nil client: got %v", err)
	}
	_, err = CreateProfile(context.Background(), &http.Client{}, "", CreateProfileRequest{})
	if err == nil || !strings.Contains(err.Error(), "gammaURL is required") {
		t.Errorf("empty url: got %v", err)
	}
	_, err = CreateProfile(context.Background(), &http.Client{}, "http://x", CreateProfileRequest{})
	if err == nil || !strings.Contains(err.Error(), "ProxyWallet is required") {
		t.Errorf("missing proxy: got %v", err)
	}
	_, err = CreateProfile(context.Background(), &http.Client{}, "http://x", CreateProfileRequest{ProxyWallet: "0x1"})
	if err == nil || !strings.Contains(err.Error(), "Users[0].Address is required") {
		t.Errorf("missing user: got %v", err)
	}
}

func TestCreateProfilePropagates409Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"profile already exists"}`))
	}))
	defer srv.Close()

	body := NewCreateProfileRequest("0xeoa", "0xproxy", "metamask", 1)
	_, err := CreateProfile(context.Background(), &http.Client{}, srv.URL, body)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP 409") {
		t.Errorf("expected HTTP 409 in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "profile already exists") {
		t.Errorf("expected response body in error, got: %v", err)
	}
}
