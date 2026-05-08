package relayer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// V2 Relayer API auth.
//
// The live mint returns `{apiKey, address, createdAt, updatedAt}` — a
// single UUID and the EOA address. Downstream relayer calls authenticate
// with two plain headers:
//
//   RELAYER_API_KEY:         <apiKey UUID>
//   RELAYER_API_KEY_ADDRESS: <0x address>
//
// No HMAC, no secret, no passphrase, no timestamp signature. Confirmed
// against `Polymarket/relayer-client` bundle (`06yo__3skxzq9.js` line
// ~61849: `getRelayerApiKeys(e) → headers={RELAYER_API_KEY: e.apiKey,
// RELAYER_API_KEY_ADDRESS: e.address}`).
//
// This file ships the V2 path. The legacy POLY_BUILDER_* HMAC headers
// (in `internal/auth/l2.go::BuildBuilderHeaders`) coexist for now and
// remain wired through `BuilderConfig`.

// V2APIKey is the V2 relayer API key triple. The "triple" is misleading —
// V2 uses two values (Key + Address) where V1 used three (Key, Secret,
// Passphrase). The `CreatedAt` field is captured for diagnostics.
type V2APIKey struct {
	Key       string `json:"apiKey"`
	Address   string `json:"address"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// V2Headers returns the plain headers the V2 relayer expects on every
// authenticated request: `RELAYER_API_KEY` + `RELAYER_API_KEY_ADDRESS`.
func (k V2APIKey) V2Headers() map[string]string {
	return map[string]string{
		"RELAYER_API_KEY":         k.Key,
		"RELAYER_API_KEY_ADDRESS": k.Address,
	}
}

// MintV2APIKey calls `POST {relayerURL}/relayer/api/auth` with body `{}`,
// authenticated by the cookies in the supplied http.Client (which the
// caller has populated via [auth.SIWESession].Login).
//
// Each call mints a new key — the caller should persist the result and
// not re-mint per request.
func MintV2APIKey(ctx context.Context, client *http.Client, relayerURL string) (V2APIKey, error) {
	if client == nil {
		return V2APIKey{}, fmt.Errorf("client is required")
	}
	if strings.TrimSpace(relayerURL) == "" {
		return V2APIKey{}, fmt.Errorf("relayerURL is required")
	}
	url := strings.TrimRight(relayerURL, "/") + "/relayer/api/auth"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return V2APIKey{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return V2APIKey{}, fmt.Errorf("mint v2 api key: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return V2APIKey{}, fmt.Errorf("HTTP %d %s: %s", resp.StatusCode, url, string(body))
	}

	var key V2APIKey
	if err := json.Unmarshal(body, &key); err != nil {
		return V2APIKey{}, fmt.Errorf("decode response: %w", err)
	}
	if strings.TrimSpace(key.Key) == "" || strings.TrimSpace(key.Address) == "" {
		return V2APIKey{}, fmt.Errorf("relayer returned incomplete key (apiKey=%q address=%q)", key.Key, key.Address)
	}
	return key, nil
}
