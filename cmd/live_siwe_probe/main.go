// live_siwe_probe is a one-shot diagnostic: generates a throwaway EOA,
// runs the SIWE login flow against gamma-api.polymarket.com production,
// attempts the relayer-v2 auth mint, then attempts a real WALLET-CREATE
// against the relayer with the V2 API key. Captures the wire trace.
// Disposable — run once, delete.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return fmt.Errorf("rand: %w", err)
	}
	keyHex := "0x" + hex.EncodeToString(keyBytes)

	priv, err := ethcrypto.HexToECDSA(keyHex[2:])
	if err != nil {
		return fmt.Errorf("hex to ecdsa: %w", err)
	}
	addr := ethcrypto.PubkeyToAddress(priv.PublicKey).Hex()

	fmt.Printf("[probe] throwaway EOA address = %s\n", addr)
	fmt.Printf("[probe] (private key not printed; key bytes len=%d)\n", len(keyBytes))

	signer, err := auth.NewPrivateKeySigner(keyHex, 137)
	if err != nil {
		return fmt.Errorf("signer: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := auth.NewSIWESession(signer, "https://gamma-api.polymarket.com")
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}

	fmt.Printf("[probe] running SIWE login against gamma-api.polymarket.com...\n")
	if err := session.Login(ctx); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	cookies := session.CookiesFor("https://gamma-api.polymarket.com/")
	fmt.Printf("[probe] login OK — captured %d cookies for gamma-api:\n", len(cookies))
	for _, c := range cookies {
		val := c.Value
		if len(val) > 24 {
			val = val[:8] + "…(redacted, len=" + fmt.Sprintf("%d", len(c.Value)) + ")…"
		}
		fmt.Printf("  - %s = %s (Domain=%s Path=%s Secure=%v HttpOnly=%v)\n",
			c.Name, val, c.Domain, c.Path, c.Secure, c.HttpOnly)
	}

	relayerCookies := session.CookiesFor("https://relayer-v2.polymarket.com/")
	fmt.Printf("[probe] cookies the jar would send to relayer-v2.polymarket.com: %d\n", len(relayerCookies))
	for _, c := range relayerCookies {
		fmt.Printf("  - %s (Domain=%s)\n", c.Name, c.Domain)
	}

	if len(relayerCookies) == 0 {
		fmt.Println("[probe] no cookies for relayer-v2 — Domain attribute is gamma-api-only.")
		fmt.Println("[probe] will forward cookies manually to relayer-v2.")
	}

	fmt.Println("[probe] attempting POST relayer-v2.polymarket.com/relayer/api/auth ...")
	v2Key, err := relayer.MintV2APIKey(ctx, session.HTTPClient(), "https://relayer-v2.polymarket.com")
	if err != nil {
		return fmt.Errorf("v2 mint: %w", err)
	}
	fmt.Printf("[probe] V2 API key minted: apiKey=%s address=%s\n", v2Key.Key, v2Key.Address)

	fmt.Println("[probe] attempting WALLET-CREATE via relayer with V2 auth ...")
	if err := probeWalletCreate(ctx, v2Key, addr); err != nil {
		fmt.Printf("[probe] WALLET-CREATE: %v\n", err)
	}

	fmt.Println("[probe] DONE")
	return nil
}

func probeWalletCreate(ctx context.Context, key relayer.V2APIKey, ownerAddr string) error {
	url := "https://relayer-v2.polymarket.com/submit"
	payload := map[string]string{
		"type": "WALLET-CREATE",
		"from": ownerAddr,
		"to":   "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	for k, v := range key.V2Headers() {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	preview := string(respBody)
	if len(preview) > 600 {
		preview = preview[:600] + "…(truncated)"
	}
	fmt.Printf("[probe] WALLET-CREATE response: HTTP %d body=%s\n", resp.StatusCode, preview)
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		fmt.Println("[probe] WALLET-CREATE accepted by relayer")
		return nil
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode)
}

func probeRelayerMint(ctx context.Context, session *auth.SIWESession, gammaCookies []*http.Cookie) error {
	// Try via the SIWE session's client first (jar-based attach). If that
	// returns 0 cookies for the relayer host, manually forward.
	client := session.HTTPClient()

	// Attempt 1: jar-based attach.
	if err := postRelayerAuth(ctx, client, "[probe attempt 1: jar-based]", nil); err != nil {
		fmt.Printf("[probe] attempt 1 failed: %v\n", err)
	}

	// Attempt 2: manually forward gamma-api cookies as Cookie header.
	manualHeader := buildCookieHeader(gammaCookies)
	if err := postRelayerAuth(ctx, client, "[probe attempt 2: manual Cookie header]", &manualHeader); err != nil {
		return fmt.Errorf("attempt 2 failed: %w", err)
	}
	return nil
}

func postRelayerAuth(ctx context.Context, client *http.Client, label string, manualCookieHeader *string) error {
	url := "https://relayer-v2.polymarket.com/relayer/api/auth"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if manualCookieHeader != nil {
		req.Header.Set("Cookie", *manualCookieHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	preview := string(body)
	if len(preview) > 400 {
		preview = preview[:400] + "…(truncated)"
	}
	fmt.Printf("%s HTTP %d, body=%s\n", label, resp.StatusCode, preview)
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		fmt.Printf("%s SUCCESS — relayer minted creds\n", label)
		return nil
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode)
}

func buildCookieHeader(cookies []*http.Cookie) string {
	var parts []byte
	for i, c := range cookies {
		if i > 0 {
			parts = append(parts, ';', ' ')
		}
		parts = append(parts, []byte(c.Name+"="+c.Value)...)
	}
	return string(parts)
}
