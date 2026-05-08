// eoa_key_scout tests whether an EOA-owned CLOB API key can be used to
// place deposit-wallet (sigtype-3) orders. This is the critical hypothesis
// for pure-headless onboarding: if the server only checks L2 HMAC validity
// and on-chain ERC-1271 validity (not API-key-owner == order.signer), then
// an EOA key + deposit-wallet order should succeed.
//
// Run with POLYGOLEM_EOA_KEY_SCOUT_LIVE=1. The explicit environment gate is
// required because this command performs live SIWE, relayer, CLOB auth, wallet
// deploy, and order-post probes.
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
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if os.Getenv("POLYGOLEM_EOA_KEY_SCOUT_LIVE") != "1" {
		return fmt.Errorf("refusing to run live scout without POLYGOLEM_EOA_KEY_SCOUT_LIVE=1")
	}

	ctx := context.Background()

	// 1. Fresh EOA
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return err
	}
	keyHex := "0x" + hex.EncodeToString(keyBytes)

	// 2. SIWE login
	fmt.Println("[scout] SIWE login ...")
	signer, err := auth.NewPrivateKeySigner(keyHex, 137)
	if err != nil {
		return fmt.Errorf("signer: %w", err)
	}
	session, err := auth.NewSIWESession(signer, "https://gamma-api.polymarket.com")
	if err != nil {
		return fmt.Errorf("siwe session: %w", err)
	}
	if err := session.Login(ctx); err != nil {
		return fmt.Errorf("siwe login: %w", err)
	}
	fmt.Printf("[scout] SIWE OK — session cookie count=%d\n", len(session.CookiesFor("https://gamma-api.polymarket.com")))

	// 3. Mint V2 relayer key
	fmt.Println("[scout] minting V2 relayer key ...")
	v2Key, err := relayer.MintV2APIKey(ctx, session.HTTPClient(), "https://relayer-v2.polymarket.com")
	if err != nil {
		return fmt.Errorf("relayer mint: %w", err)
	}
	fmt.Printf("[scout] relayer key OK — addr=%s\n", v2Key.Address)

	// 4. Deploy deposit wallet
	fmt.Println("[scout] deploying deposit wallet ...")
	ownerAddr := signer.Address()
	depositWallet, _, err := deployDepositWallet(ctx, v2Key, ownerAddr)
	if err != nil {
		return fmt.Errorf("deploy: %w", err)
	}
	fmt.Printf("[scout] deposit wallet deployed: %s\n", depositWallet)

	// 5. Create EOA-owned API key (this path is PROVEN to work)
	fmt.Println("[scout] creating EOA-owned CLOB API key ...")
	eoaHeaders, err := auth.BuildL1HeadersFromPrivateKey(keyHex, 137, time.Now().Unix(), 0)
	if err != nil {
		return fmt.Errorf("l1 headers: %w", err)
	}

	mintReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://clob.polymarket.com/auth/api-key", bytes.NewReader([]byte("")))
	mintReq.Header.Set("Accept", "application/json")
	mintReq.Header.Set("Content-Type", "application/json")
	for k, v := range eoaHeaders {
		mintReq.Header.Set(k, v)
	}
	mintResp, err := http.DefaultClient.Do(mintReq)
	if err != nil {
		return fmt.Errorf("mint http: %w", err)
	}
	mintBody, _ := io.ReadAll(mintResp.Body)
	mintResp.Body.Close()
	fmt.Printf("[scout] EOA key mint → HTTP %d body=%s\n", mintResp.StatusCode, string(mintBody))
	if mintResp.StatusCode < 200 || mintResp.StatusCode > 299 {
		return fmt.Errorf("EOA key mint failed: HTTP %d", mintResp.StatusCode)
	}

	var eoaKey struct {
		APIKey     string `json:"apiKey"`
		Secret     string `json:"secret"`
		Passphrase string `json:"passphrase"`
	}
	json.Unmarshal(mintBody, &eoaKey)
	fmt.Printf("[scout] EOA apiKey=%s\n", eoaKey.APIKey)

	// TEST A: signer=depositWallet, maker=depositWallet, API key=EOA-owned
	fmt.Println("[scout] TEST A: signer=depositWallet, maker=depositWallet, API key=EOA-owned")
	orderPayloadA, err := buildDepositWalletOrder(keyHex, depositWallet, depositWallet, ownerAddr)
	if err != nil {
		return fmt.Errorf("build order A: %w", err)
	}
	if err := testOrderPost(ctx, eoaKey, depositWallet, depositWallet, ownerAddr, orderPayloadA); err != nil {
		return err
	}

	// TEST B: signer=EOA, maker=depositWallet, API key=EOA-owned
	fmt.Println("\n[scout] TEST B: signer=EOA, maker=depositWallet, API key=EOA-owned")
	orderPayloadB, err := buildDepositWalletOrder(keyHex, depositWallet, ownerAddr, ownerAddr)
	if err != nil {
		return fmt.Errorf("build order B: %w", err)
	}
	if err := testOrderPost(ctx, eoaKey, depositWallet, ownerAddr, ownerAddr, orderPayloadB); err != nil {
		return err
	}

	return nil
}

func testOrderPost(ctx context.Context, eoaKey struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}, depositWallet, signerAddr, ownerAddr string, orderPayload []byte) error {
	bodyStr := string(orderPayload)
	l2Headers, err := auth.BuildL2Headers(&auth.APIKey{
		Key:        eoaKey.APIKey,
		Secret:     eoaKey.Secret,
		Passphrase: eoaKey.Passphrase,
	}, time.Now().Unix(), http.MethodPost, "/order", &bodyStr)
	if err != nil {
		return fmt.Errorf("l2 headers: %w", err)
	}
	l2Headers["POLY_ADDRESS"] = ownerAddr

	orderReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://clob.polymarket.com/order", bytes.NewReader(orderPayload))
	orderReq.Header.Set("Accept", "application/json")
	orderReq.Header.Set("Content-Type", "application/json")
	for k, v := range l2Headers {
		orderReq.Header.Set(k, v)
	}

	orderResp, err := http.DefaultClient.Do(orderReq)
	if err != nil {
		return fmt.Errorf("order http: %w", err)
	}
	orderBody, _ := io.ReadAll(orderResp.Body)
	orderResp.Body.Close()
	fmt.Printf("[scout] order post → HTTP %d body=%s\n", orderResp.StatusCode, string(orderBody))

	msg := string(orderBody)
	if strings.Contains(msg, "the order owner has to be the owner of the API KEY") {
		fmt.Println("[scout] FINDING: Owner gate ENFORCED — order owner must match API key owner.")
	} else if strings.Contains(msg, "the order signer address has to be the address of the API KEY") {
		fmt.Println("[scout] FINDING: Signer gate ENFORCED — order signer must match API key owner.")
	} else if strings.Contains(msg, "insufficient") || strings.Contains(msg, "balance") || strings.Contains(msg, "allowance") || strings.Contains(msg, "1271") || strings.Contains(msg, "wallet") {
		fmt.Println("[scout] FINDING: Gate PASSED — order reached validation stage!")
	} else if orderResp.StatusCode >= 200 && orderResp.StatusCode <= 299 {
		fmt.Println("[scout] FINDING: ORDER ACCEPTED (HTTP 2xx)!")
	} else {
		fmt.Println("[scout] FINDING: unexpected response — needs analysis")
	}
	return nil
}

func deployDepositWallet(ctx context.Context, v2Key relayer.V2APIKey, ownerAddr string) (string, string, error) {
	rc, err := relayer.NewV2("https://relayer-v2.polymarket.com", v2Key, 137)
	if err != nil {
		return "", "", err
	}
	tx, err := rc.SubmitWalletCreate(ctx, ownerAddr)
	if err != nil {
		return "", "", err
	}
	depositWallet, err := auth.MakerAddressForSignatureType(ownerAddr, 137, 3)
	if err != nil {
		return "", "", err
	}
	return depositWallet, tx.TransactionHash, nil
}

func buildDepositWalletOrder(privateKeyHex, depositWallet, signerAddr, ownerAddr string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"order": map[string]interface{}{
			"salt":          123456789,
			"maker":         depositWallet,
			"signer":        signerAddr,
			"tokenId":       "69324317355037271422943965141382095011871956039434394956830818206664869608517",
			"makerAmount":   "5000",
			"takerAmount":   "5000000",
			"side":          "BUY",
			"expiration":    "0",
			"signatureType": 3,
			"timestamp":     fmt.Sprintf("%d", time.Now().UnixMilli()),
			"metadata":      "0x0000000000000000000000000000000000000000000000000000000000000000",
			"builder":       "0x0000000000000000000000000000000000000000000000000000000000000000",
			"signature":     "0x00",
		},
		"owner":     ownerAddr,
		"orderType": "GTC",
		"postOnly":  false,
		"deferExec": false,
	})
}
