package rpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// HasCode checks whether an address has deployed bytecode on Polygon.
func HasCode(ctx context.Context, address string, rpcURL string) (bool, error) {
	address = strings.TrimSpace(address)
	if !common.IsHexAddress(address) {
		return false, fmt.Errorf("invalid address: %s", address)
	}
	if strings.TrimSpace(rpcURL) == "" {
		rpcURL = polygonRPC
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return false, fmt.Errorf("connect polygon rpc: %w", err)
	}
	defer client.Close()

	code, err := client.CodeAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return false, fmt.Errorf("eth_getCode: %w", err)
	}
	return len(code) > 0, nil
}
