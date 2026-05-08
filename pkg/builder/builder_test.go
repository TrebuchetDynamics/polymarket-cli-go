package builder

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestLocalSignerCreatesAllRequiredHeaders(t *testing.T) {
	signer, err := NewLocalSigner(LocalSignerConfig{
		Key:        "test-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("test-secret")),
		Passphrase: "test-pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	headers, err := signer.CreateHeaders("POST", "/order", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{POLY_BUILDER_API_KEY, POLY_BUILDER_PASSPHRASE, POLY_BUILDER_TIMESTAMP, POLY_BUILDER_SIGNATURE} {
		if headers[key] == "" {
			t.Fatalf("missing header %s", key)
		}
	}
}

func TestLocalSignerSignatureIsValidHMAC(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("my-secret"))
	signer, err := NewLocalSigner(LocalSignerConfig{
		Key:        "key",
		Secret:     secret,
		Passphrase: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	ts := int64(1234567890)
	body := `{"test":true}`
	headers, err := signer.CreateHeaders("POST", "/order", &body, &ts)
	if err != nil {
		t.Fatal(err)
	}

	message := strconv.FormatInt(ts, 10) + "POST" + "/order" + body
	decodedSecret, _ := base64.StdEncoding.DecodeString(secret)
	h := hmac.New(sha256.New, decodedSecret)
	h.Write([]byte(message))
	expectedSig := base64.StdEncoding.EncodeToString(h.Sum(nil))
	expectedSig = strings.ReplaceAll(expectedSig, "+", "-")
	expectedSig = strings.ReplaceAll(expectedSig, "/", "_")

	if headers[POLY_BUILDER_SIGNATURE] != expectedSig {
		t.Fatalf("signature mismatch: got %s want %s", headers[POLY_BUILDER_SIGNATURE], expectedSig)
	}
}

func TestLocalSignerRejectsEmptyConfig(t *testing.T) {
	_, err := NewLocalSigner(LocalSignerConfig{})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestRemoteSignerCallsServerAndReturnsHeaders(t *testing.T) {
	var gotPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("auth header=%q want Bearer test-token", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		resp := map[string]string{
			"POLY_BUILDER_API_KEY":     "remote-key",
			"POLY_BUILDER_TIMESTAMP":   "1234567890",
			"POLY_BUILDER_PASSPHRASE":  "remote-pass",
			"POLY_BUILDER_SIGNATURE":   "remote-sig",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	signer, err := NewRemoteSigner(RemoteSignerConfig{URL: server.URL, Token: "test-token"})
	if err != nil {
		t.Fatal(err)
	}

	body := `{"order":true}`
	headers, err := signer.CreateHeaders("POST", "/order", &body, nil)
	if err != nil {
		t.Fatal(err)
	}
	if headers[POLY_BUILDER_API_KEY] != "remote-key" {
		t.Fatalf("api key=%q want remote-key", headers[POLY_BUILDER_API_KEY])
	}
	if gotPayload["method"] != "POST" || gotPayload["path"] != "/order" {
		t.Fatalf("payload=%v", gotPayload)
	}
}

func TestRemoteSignerRejectsEmptyURL(t *testing.T) {
	_, err := NewRemoteSigner(RemoteSignerConfig{URL: "", Token: "token"})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestRemoteSignerRejectsEmptyToken(t *testing.T) {
	_, err := NewRemoteSigner(RemoteSignerConfig{URL: "http://example.com", Token: ""})
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestGenSignatureURLSafeBase64(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte{0xFF, 0xFE, 0xFD, 0xFC})
	sig := GenSignature(secret, 1, "GET", "/test", nil)
	if strings.Contains(sig, "+") || strings.Contains(sig, "/") {
		t.Fatalf("signature not URL-safe: %s", sig)
	}
}
