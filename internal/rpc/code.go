package rpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
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

// isApprovedForAllSelector is the first 4 bytes of
// keccak256("isApprovedForAll(address,address)").
var isApprovedForAllSelector = []byte{0xe9, 0x85, 0xe9, 0xc5}

// IsApprovedForAll calls ERC-1155 isApprovedForAll(owner, operator) on the
// given token contract via eth_call. Returns false on RPC error to keep the
// caller fail-closed; the caller should treat any non-true result as
// "approval missing — refuse to sign".
func IsApprovedForAll(ctx context.Context, tokenAddress, owner, operator, rpcURL string) (bool, error) {
	tokenAddress = strings.TrimSpace(tokenAddress)
	owner = strings.TrimSpace(owner)
	operator = strings.TrimSpace(operator)
	if !common.IsHexAddress(tokenAddress) {
		return false, fmt.Errorf("invalid token address: %s", tokenAddress)
	}
	if !common.IsHexAddress(owner) {
		return false, fmt.Errorf("invalid owner address: %s", owner)
	}
	if !common.IsHexAddress(operator) {
		return false, fmt.Errorf("invalid operator address: %s", operator)
	}
	if strings.TrimSpace(rpcURL) == "" {
		rpcURL = polygonRPC
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return false, fmt.Errorf("connect polygon rpc: %w", err)
	}
	defer client.Close()

	data := make([]byte, 0, 4+32+32)
	data = append(data, isApprovedForAllSelector...)
	data = append(data, common.LeftPadBytes(common.HexToAddress(owner).Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(common.HexToAddress(operator).Bytes(), 32)...)

	addr := common.HexToAddress(tokenAddress)
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &addr, Data: data}, nil)
	if err != nil {
		return false, fmt.Errorf("eth_call isApprovedForAll: %w", err)
	}
	if len(out) < 32 {
		return false, fmt.Errorf("isApprovedForAll: short response (%d bytes)", len(out))
	}
	for i := 0; i < 31; i++ {
		if out[i] != 0 {
			return false, fmt.Errorf("isApprovedForAll: non-zero high bytes in bool: %x", out[:32])
		}
	}
	return out[31] == 1, nil
}
