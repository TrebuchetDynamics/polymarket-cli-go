package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xelhaku/polymarket-cli-go/pkg/auth"
	"github.com/xelhaku/polymarket-cli-go/pkg/clob"
	"github.com/xelhaku/polymarket-cli-go/pkg/config"
	"github.com/xelhaku/polymarket-cli-go/pkg/gamma"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	ctx := context.Background()

	switch cmd {
	case "setup":
		runSetup(ctx)
	case "balance":
		runBalance(ctx)
	case "create-order":
		runCreateOrder(ctx)
	case "market-order":
		runMarketOrder(ctx)
	case "book":
		runBook(ctx)
	case "markets":
		runMarkets(ctx)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Polymarket CLI Go - CLOB V2 Client

Usage:
  polymarket <command> [options]

Commands:
  setup          First-time wallet setup and deposit wallet flow
  balance        Check CLOB balance and allowances
  create-order   Place a limit order
  market-order   Place a market order (FOK)
  book <token>   View order book for a token
  markets        List active markets
  help           Show this help

Environment:
  POLYMARKET_PRIVATE_KEY    Required. Your EOA private key
  POLYMARKET_HOST           CLOB host (default: https://clob.polymarket.com)
  POLYMARKET_GAMMA_URL      Gamma API (default: https://gamma-api.polymarket.com)
`)
}

func runSetup(ctx context.Context) {
	cfg := config.Load()
	
	signer, err := auth.NewSigner(cfg.PrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid private key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wallet Address: %s\n", signer.Address())
	fmt.Printf("Chain ID: %d\n", cfg.ChainID)
	
	// Derive proxy wallet if needed for deposit flow
	proxyAddr := signer.DeriveProxyAddress(cfg.ChainID)
	fmt.Printf("Proxy Wallet: %s\n", proxyAddr)
	
	client := clob.NewClient(cfg.CLOBHost, cfg.ChainID, signer)
	
	// Create or derive API key
	creds, err := client.CreateOrDeriveAPIKey(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "API key creation failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("API Key: %s...\n", creds.Key[:8])
	fmt.Println("\nSetup complete!")
	fmt.Printf("Next: deposit pUSD to %s and run 'polymarket balance'\n", proxyAddr)
}

func runBalance(ctx context.Context) {
	cfg := config.Load()
	signer, err := auth.NewSigner(cfg.PrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid private key: %v\n", err)
		os.Exit(1)
	}

	client := clob.NewClient(cfg.CLOBHost, cfg.ChainID, signer)
	creds, err := client.CreateOrDeriveAPIKey(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "API key error: %v\n", err)
		os.Exit(1)
	}
	client.SetCredentials(creds)

	bal, err := client.GetBalance(ctx, "collateral")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Balance error: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(bal, "", "  ")
	fmt.Println(string(output))
}

func runCreateOrder(ctx context.Context) {
	cfg := config.Load()
	signer, err := auth.NewSigner(cfg.PrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid private key: %v\n", err)
		os.Exit(1)
	}

	client := clob.NewClient(cfg.CLOBHost, cfg.ChainID, signer)
	creds, err := client.CreateOrDeriveAPIKey(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "API key error: %v\n", err)
		os.Exit(1)
	}
	client.SetCredentials(creds)

	// Parse args
	order := parseOrderArgs()
	
	resp, err := client.CreateOrder(ctx, order)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Order failed: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(output))
}

func runMarketOrder(ctx context.Context) {
	cfg := config.Load()
	signer, err := auth.NewSigner(cfg.PrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid private key: %v\n", err)
		os.Exit(1)
	}

	client := clob.NewClient(cfg.CLOBHost, cfg.ChainID, signer)
	creds, err := client.CreateOrDeriveAPIKey(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "API key error: %v\n", err)
		os.Exit(1)
	}
	client.SetCredentials(creds)

	order := parseMarketOrderArgs()
	
	resp, err := client.CreateMarketOrder(ctx, order)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Market order failed: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(output))
}

func runBook(ctx context.Context) {
	cfg := config.Load()
	client := clob.NewClient(cfg.CLOBHost, cfg.ChainID, nil)

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: polymarket book <token-id>\n")
		os.Exit(1)
	}

	book, err := client.GetOrderBook(ctx, os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Book error: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(book, "", "  ")
	fmt.Println(string(output))
}

func runMarkets(ctx context.Context) {
	cfg := config.Load()
	client := gamma.NewClient(cfg.GammaURL)

	markets, err := client.GetActiveMarkets(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Markets error: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(markets, "", "  ")
	fmt.Println(string(output))
}

func parseOrderArgs() clob.Order {
	order := clob.Order{}
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--token":
			i++; order.TokenID = os.Args[i]
		case "--side":
			i++; order.Side = os.Args[i]
		case "--price":
			i++; fmt.Sscanf(os.Args[i], "%f", &order.Price)
		case "--size":
			i++; fmt.Sscanf(os.Args[i], "%f", &order.Size)
		case "--order-type":
			i++; order.OrderType = os.Args[i]
		}
	}
	if order.OrderType == "" {
		order.OrderType = "GTC"
	}
	return order
}

func parseMarketOrderArgs() clob.MarketOrder {
	order := clob.MarketOrder{}
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--token":
			i++; order.TokenID = os.Args[i]
		case "--side":
			i++; order.Side = os.Args[i]
		case "--amount":
			i++; fmt.Sscanf(os.Args[i], "%f", &order.Amount)
		case "--order-type":
			i++; order.OrderType = os.Args[i]
		}
	}
	if order.OrderType == "" {
		order.OrderType = "FOK"
	}
	return order
}
