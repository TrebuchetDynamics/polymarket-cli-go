// indexer_probe is a long-running diagnostic that tests whether a Polymarket
// backend indexer eventually registers an EOA→depositWallet relationship some
// time after the on-chain WALLET-CREATE.
//
// Two modes:
//
//  1. Bootstrap (no STATE_FILE): generate a fresh EOA, deploy its deposit
//     wallet via the V2 relayer, persist the EOA private key + deposit wallet
//     address to STATE_FILE, then exit.
//
//  2. Probe (STATE_FILE exists): load state, run SIWE login + V2 mint +
//     wrapped L1 mint against /auth/api-key, attempt one sigtype-3 limit
//     order at price 0.001 (rests far below market). Log the outcome with a
//     timestamp. Exit.
//
// Run with `go run ./cmd/indexer_probe` once to bootstrap, then schedule
// the binary every 15 minutes (cron / systemd-timer / `while true; do …; sleep 900; done`).
//
// Env vars:
//
//	STATE_FILE        path to persistent state (default ./indexer_probe.state.json)
//	PROBE_TOKEN_ID    optional CLOB token id for the order test (default = active 'Jesus' market YES)
//
// The state file contains the throwaway EOA private key — DO NOT commit it.
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
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	defaultStateFile = "./indexer_probe.state.json"
	defaultTokenID   = "69324317355037271422943965141382095011871956039434394956830818206664869608517" // Jesus market YES
	gammaURL         = "https://gamma-api.polymarket.com"
	relayerURL       = "https://relayer-v2.polymarket.com"
	clobURL          = "https://clob.polymarket.com"
	polygonRPC       = "https://polygon-bor-rpc.publicnode.com"
)

type state struct {
	EOAKey          string `json:"eoa_key_hex"`
	EOAAddress      string `json:"eoa_address"`
	DepositWallet   string `json:"deposit_wallet"`
	DeployTxHash    string `json:"deploy_tx_hash"`
	DeployedAtUnix  int64  `json:"deployed_at_unix"`
	BootstrappedISO string `json:"bootstrapped_iso"`
}

func main() {
	stateFile := envOr("STATE_FILE", defaultStateFile)
	tokenID := envOr("PROBE_TOKEN_ID", defaultTokenID)

	st, err := loadState(stateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load state: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	if st == nil {
		log("BOOTSTRAP", "no state file; generating fresh EOA + deploying deposit wallet")
		st, err = bootstrap(ctx)
		if err != nil {
			log("BOOTSTRAP_FAIL", err.Error())
			os.Exit(1)
		}
		if err := saveState(stateFile, st); err != nil {
			fmt.Fprintf(os.Stderr, "save state: %v\n", err)
			os.Exit(1)
		}
		log("BOOTSTRAP_OK", fmt.Sprintf("eoa=%s deposit=%s tx=%s", st.EOAAddress, st.DepositWallet, st.DeployTxHash))
		log("HINT", fmt.Sprintf("schedule periodic re-tries: while true; do %s; sleep 900; done >> probe.log 2>&1", os.Args[0]))
		return
	}

	age := time.Since(time.Unix(st.DeployedAtUnix, 0))
	log("PROBE", fmt.Sprintf("eoa=%s deposit=%s wallet_age=%s", st.EOAAddress, st.DepositWallet, age.Round(time.Second)))

	if err := probeOnce(ctx, st, tokenID); err != nil {
		log("PROBE_FAIL", err.Error())
		os.Exit(1)
	}
}

func bootstrap(ctx context.Context) (*state, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("rand: %w", err)
	}
	keyHex := "0x" + hex.EncodeToString(keyBytes)
	priv, err := ethcrypto.HexToECDSA(keyHex[2:])
	if err != nil {
		return nil, err
	}
	eoa := ethcrypto.PubkeyToAddress(priv.PublicKey).Hex()

	signer, err := auth.NewPrivateKeySigner(keyHex, 137)
	if err != nil {
		return nil, err
	}
	session, err := auth.NewSIWESession(signer, gammaURL)
	if err != nil {
		return nil, err
	}
	if err := session.Login(ctx); err != nil {
		return nil, fmt.Errorf("siwe login: %w", err)
	}
	v2Key, err := relayer.MintV2APIKey(ctx, session.HTTPClient(), relayerURL)
	if err != nil {
		return nil, fmt.Errorf("v2 mint: %w", err)
	}
	rc, err := relayer.NewV2(relayerURL, v2Key, 137)
	if err != nil {
		return nil, fmt.Errorf("relayer v2: %w", err)
	}
	tx, err := rc.SubmitWalletCreate(ctx, eoa)
	if err != nil {
		return nil, fmt.Errorf("WALLET-CREATE: %w", err)
	}
	depositWallet, err := auth.MakerAddressForSignatureType(eoa, 137, 3)
	if err != nil {
		return nil, fmt.Errorf("derive deposit wallet: %w", err)
	}

	// Wait for on-chain confirmation so subsequent probes can attempt the
	// wrapped L1 mint without a "wallet not deployed" false negative.
	deployed := false
	for i := 0; i < 60; i++ {
		hasCode, _ := pollCode(ctx, depositWallet)
		if hasCode {
			deployed = true
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	if !deployed {
		return nil, fmt.Errorf("deposit wallet not deployed within bootstrap window")
	}

	return &state{
		EOAKey:          keyHex,
		EOAAddress:      eoa,
		DepositWallet:   depositWallet,
		DeployTxHash:    tx.TransactionHash,
		DeployedAtUnix:  time.Now().Unix(),
		BootstrappedISO: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func probeOnce(ctx context.Context, st *state, tokenID string) error {
	// Attempt 1: wrapped L1 mint with POLY_ADDRESS=depositWallet.
	clobC := clob.NewClient(clobURL, nil)
	wrappedKey, wrappedErr := clobC.CreateAPIKeyForAddress(ctx, st.EOAKey, st.DepositWallet)
	if wrappedErr == nil && wrappedKey.Key != "" {
		log("L1_WRAPPED_OK", fmt.Sprintf("apiKey=%s — indexer registered the wallet", wrappedKey.Key))
		// Try the sigtype-3 order — this is the real proof.
		return tryOrder(ctx, st.EOAKey, tokenID)
	}
	log("L1_WRAPPED_FAIL", wrappedErr.Error())

	// Attempt 2: EOA-bound L1 mint, then sigtype-3 order. If the gate now passes,
	// the indexer-aware CREATE2 path is live.
	clobKey, eoaErr := clobC.CreateOrDeriveAPIKey(ctx, st.EOAKey)
	if eoaErr != nil {
		log("L1_EOA_FAIL", eoaErr.Error())
		return eoaErr
	}
	log("L1_EOA_OK", fmt.Sprintf("apiKey=%s", clobKey.Key))
	return tryOrder(ctx, st.EOAKey, tokenID)
}

func tryOrder(ctx context.Context, keyHex, tokenID string) error {
	c := clob.NewClient(clobURL, nil)
	resp, err := c.CreateLimitOrder(ctx, keyHex, clob.CreateOrderParams{
		TokenID:   tokenID,
		Side:      "BUY",
		Price:     "0.001",
		Size:      "5",
		OrderType: "GTC",
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "the order signer address has to be the address of the API KEY") {
			log("ORDER_GATE", "still failing the API-KEY gate")
		} else if strings.Contains(msg, "insufficient") || strings.Contains(msg, "balance") || strings.Contains(msg, "allowance") || strings.Contains(msg, "1271") {
			log("ORDER_GATE_PASSED", fmt.Sprintf("gate cleared, downstream error: %s", msg))
		} else {
			log("ORDER_OTHER", msg)
		}
		return err
	}
	log("ORDER_ACCEPTED", fmt.Sprintf("orderID=%s status=%s — cancel manually", resp.OrderID, resp.Status))
	return nil
}

func pollCode(ctx context.Context, addr string) (bool, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0", "method": "eth_getCode", "params": []string{addr, "latest"}, "id": 1,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, polygonRPC, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var out struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return false, err
	}
	return len(out.Result) > 4, nil
}

func loadState(path string) (*state, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s state
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.EOAKey == "" || s.DepositWallet == "" {
		return nil, fmt.Errorf("incomplete state file (missing eoa_key_hex or deposit_wallet)")
	}
	return &s, nil
}

func saveState(path string, s *state) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func log(tag, msg string) {
	fmt.Printf("%s [%s] %s\n", time.Now().UTC().Format(time.RFC3339), tag, msg)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
