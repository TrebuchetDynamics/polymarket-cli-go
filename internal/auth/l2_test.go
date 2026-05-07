package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"testing"
)

func TestSignHMACDecodesURLSafeBase64Secrets(t *testing.T) {
	secretBytes := []byte{251, 255, 255, 1, 2, 3}
	secret := base64.URLEncoding.EncodeToString(secretBytes)
	timestamp := int64(1000000)
	method := "GET"
	path := "/balance-allowance"

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(strconv.FormatInt(timestamp, 10) + method + path))
	want := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	if got := SignHMAC(secret, timestamp, method, path, nil); got != want {
		t.Fatalf("signature=%q, want %q", got, want)
	}
}
