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

const defaultPolygonRPC = "https://polygon-bor-rpc.publicnode.com"

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

	// Three-minute window covers SIWE login + V2 mint + WALLET-CREATE submit
	// + on-chain deploy confirm + wrapped L1 mint.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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

	// Wait for the deposit wallet to be deployed on-chain before minting
	// the L2 key — the wrapped L1 ClobAuth runs through the deposit
	// wallet's isValidSignature, which requires the contract to exist.
	fmt.Println("[probe] polling Polygon eth_getCode until the deposit wallet is on-chain ...")
	deployStart := time.Now()
	polygonRPC := firstNonEmpty(os.Getenv("POLYMARKET_RPC_URL"), defaultPolygonRPC)
	for attempt := 1; attempt <= 60; attempt++ {
		hasCode, derr := pollDepositWalletCode(ctx, http.DefaultClient, polygonRPC, depositWallet)
		if derr != nil {
			fmt.Printf("[probe] eth_getCode attempt %d error: %v\n", attempt, derr)
		} else if hasCode {
			fmt.Printf("[probe] deposit wallet deployed after %s\n", time.Since(deployStart).Round(time.Second))
			break
		}
		if attempt == 60 {
			return fmt.Errorf("deposit wallet not deployed after %d attempts", attempt)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	// Hypothesis A: re-mint the V2 relayer key AFTER the deposit wallet is deployed
	// — maybe the response shape changes once the indexer sees the wallet.
	fmt.Println("[probe] re-minting V2 relayer key after on-chain deploy ...")
	v2KeyAfter, remintErr := relayer.MintV2APIKey(ctx, session.HTTPClient(), "https://relayer-v2.polymarket.com")
	if remintErr != nil {
		fmt.Printf("[probe] re-mint failed: %v\n", remintErr)
	} else {
		fmt.Printf("[probe] V2 re-mint key=%s addr=%s (orig was key=%s addr=%s)\n", v2KeyAfter.Key, v2KeyAfter.Address, v2Key.Key, v2Key.Address)
		_ = v2KeyAfter // avoid unused if first-mint already had what we need
	}

	// Hypothesis B: forward the SIWE polymarketsession cookie to clob.polymarket.com
	// alongside the wrapped L1 headers. The browser's fetch sends *.polymarket.com
	// cookies cross-subdomain.
	fmt.Println("[probe] minting CLOB L2 API key — wrapped L1 + SIWE cookies forwarded ...")
	wrappedHeaders, err := auth.BuildL1HeadersForDepositWallet(keyHex, 137, time.Now().Unix(), 0, depositWallet)
	if err != nil {
		return fmt.Errorf("BuildL1HeadersForDepositWallet: %w", err)
	}
	gammaCookies := session.CookiesFor("https://gamma-api.polymarket.com/")
	cookieHeader := buildCookieHeader(gammaCookies)
	fmt.Printf("[probe] forwarding %d cookies to clob (manual Cookie header): %d bytes\n", len(gammaCookies), len(cookieHeader))

	mintReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://clob.polymarket.com/auth/api-key", bytes.NewReader([]byte("")))
	mintReq.Header.Set("Accept", "application/json")
	mintReq.Header.Set("Content-Type", "application/json")
	for k, v := range wrappedHeaders {
		mintReq.Header.Set(k, v)
	}
	if cookieHeader != "" {
		mintReq.Header.Set("Cookie", cookieHeader)
	}
	mintResp, err := http.DefaultClient.Do(mintReq)
	if err != nil {
		return fmt.Errorf("mint http: %w", err)
	}
	defer mintResp.Body.Close()
	mintBody, _ := io.ReadAll(mintResp.Body)
	fmt.Printf("[probe] mint with cookies → HTTP %d body=%s\n", mintResp.StatusCode, string(mintBody))

	if mintResp.StatusCode < 200 || mintResp.StatusCode > 299 {
		// Fallback: try without forwarding cookies (proves whether cookies were the missing piece)
		fmt.Println("[probe] FALLBACK — minting without cookies (existing wrapped path) ...")
		clobClient := clob.NewClient("https://clob.polymarket.com", nil)
		clobKey, err := clobClient.CreateAPIKeyForAddress(ctx, keyHex, depositWallet)
		if err != nil {
			fmt.Printf("[probe] no-cookie mint also failed: %v\n", err)
			return fmt.Errorf("CreateAPIKeyForAddress: %w", err)
		}
		fmt.Printf("[probe] no-cookie deposit-bound key: %s — but cookied path was supposed to work\n", clobKey.Key)
	}
	// If we reach here either the cookie path succeeded or fallback returned creds.
	// Best-effort: parse a key from the response we got.
	var clobKey struct {
		APIKey     string `json:"apiKey"`
		Secret     string `json:"secret"`
		Passphrase string `json:"passphrase"`
	}
	_ = json.Unmarshal(mintBody, &clobKey)
	if clobKey.APIKey != "" {
		fmt.Printf("[probe] cookied mint SUCCESS — apiKey=%s\n", clobKey.APIKey)
	}

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

// pollDepositWalletCode queries Polygon directly for contract code at addr.
// Returns true once eth_getCode is non-empty (deploy mined). Bypasses the
// relayer's /deployed endpoint, which silently returned false even after
// on-chain confirm in earlier probe runs.
func pollDepositWalletCode(ctx context.Context, client *http.Client, rpcURL, addr string) (bool, error) {
	if client == nil {
		client = http.DefaultClient
	}
	rpcURL = firstNonEmpty(rpcURL, defaultPolygonRPC)
	body, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getCode",
		"params":  []string{addr, "latest"},
		"id":      1,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var out struct {
		Result string `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return false, err
	}
	if out.Error != nil {
		return false, fmt.Errorf("rpc: %s", out.Error.Message)
	}
	// Empty contract code is "0x" or "0x0".
	return len(out.Result) > 4, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
