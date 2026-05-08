// live_siwe_probe is a one-shot diagnostic that exercises the full
// headless onboarding pipeline against production with a throwaway EOA:
//
//  1. Generate a fresh EOA from crypto/rand
//  2. SIWE login at gamma-api.polymarket.com
//  3. Mint a V2 Relayer API Key
//  4. WALLET-CREATE via relayer-v2/submit (deposit wallet deploy)
//  5. Poll until the WALLET-CREATE tx confirms
//  6. Mint a CLOB L2 API key via /auth/api-key
//  7. Build a sigtype-3 limit order (Order.signer == Order.maker ==
//     depositWallet, ERC-7739 wrapped sig) and POST /order
//  8. Capture the server's response — distinguishes between
//     "order signer..." (gate still in place) vs "insufficient balance"
//     (gate passed, deposit wallet just empty)
//
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
	"strings"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
	"github.com/TrebuchetDynamics/polygolem/internal/transport"
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

	fmt.Println("[probe] WALLET-CREATE via internal/relayer (V2 auth) ...")
	depositWallet, txHash, err := probeWalletCreateViaSDK(ctx, v2Key, signer.Address())
	if err != nil {
		fmt.Printf("[probe] WALLET-CREATE: %v\n", err)
		return err
	}
	fmt.Printf("[probe] WALLET-CREATE accepted, depositWallet=%s tx=%s\n", depositWallet, txHash)

	// Optional: poll for confirmation. Skip the wait — the order test below
	// will tell us either way. If wallet isn't deployed yet, /order will
	// return an ERC-1271 sig validation error (different from the API key
	// error we want to investigate).
	fmt.Println("[probe] not polling for tx confirmation; if deploy isn't done, sigtype-3 will fail at sig validation")

	fmt.Println("[probe] minting CLOB L2 API key with EOA ...")
	clobClient := clob.NewClient("https://clob.polymarket.com", nil)
	clobKey, err := clobClient.CreateOrDeriveAPIKey(ctx, keyHex)
	if err != nil {
		return fmt.Errorf("CreateOrDeriveAPIKey: %w", err)
	}
	fmt.Printf("[probe] CLOB L2 API key minted: %s\n", clobKey.Key)

	fmt.Println("[probe] attempting sigtype-3 limit order on token 6932…517 (active 'Jesus' market) ...")
	if err := probeCreateOrder(ctx, keyHex); err != nil {
		fmt.Printf("[probe] /order: %v\n", err)
	}

	fmt.Println("[probe] DONE")
	return nil
}

func probeWalletCreateViaSDK(ctx context.Context, v2Key relayer.V2APIKey, ownerAddr string) (string, string, error) {
	rc, err := relayer.NewV2("https://relayer-v2.polymarket.com", v2Key, 137)
	if err != nil {
		return "", "", fmt.Errorf("relayer.NewV2: %w", err)
	}
	tx, err := rc.SubmitWalletCreate(ctx, ownerAddr)
	if err != nil {
		return "", "", fmt.Errorf("SubmitWalletCreate: %w", err)
	}
	depositWallet, err := auth.MakerAddressForSignatureType(ownerAddr, 137, 3)
	if err != nil {
		return "", "", fmt.Errorf("derive deposit wallet: %w", err)
	}
	return depositWallet, tx.TransactionHash, nil
}

func probeCreateOrder(ctx context.Context, privateKeyHex string) error {
	tc := transport.New(nil, transport.DefaultConfig("https://clob.polymarket.com"))
	c := clob.NewClient("https://clob.polymarket.com", tc)
	resp, err := c.CreateLimitOrder(ctx, privateKeyHex, clob.CreateOrderParams{
		// Yes token of "Will Jesus Christ return before 2027?" — confirmed
		// active, accepting_orders=true, tick=0.001 (verified 2026-05-07).
		TokenID:   "69324317355037271422943965141382095011871956039434394956830818206664869608517",
		Side:      "BUY",
		Price:     "0.001",
		Size:      "5",
		OrderType: "GTC",
	})
	if err != nil {
		// The interesting cases:
		//   "the order signer address has to be the address of the API KEY"
		//     → the gate is still in place; scout's hypothesis is wrong.
		//   "insufficient balance" / "no allowance" / "ERC-1271 ..."
		//     → the API-key gate passed; the wallet is just empty / not deployed yet.
		msg := err.Error()
		if strings.Contains(msg, "the order signer address has to be the address of the API KEY") {
			fmt.Println("[probe] FINDING: API-KEY gate is still in place — the scout's hypothesis is WRONG. Need a different approach.")
		} else if strings.Contains(msg, "insufficient") || strings.Contains(msg, "balance") || strings.Contains(msg, "allowance") || strings.Contains(msg, "1271") || strings.Contains(msg, "wallet") {
			fmt.Println("[probe] FINDING: API-KEY gate PASSED — failure is downstream (balance / allowance / wallet not deployed). The scout's hypothesis is CORRECT.")
		} else {
			fmt.Printf("[probe] FINDING: unexpected error shape — %s\n", msg)
		}
		return err
	}
	fmt.Printf("[probe] /order accepted: %+v\n", resp)
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
