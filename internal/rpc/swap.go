package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Uniswap V3 SwapRouter02 on Polygon. Native POL is auto-wrapped to WMATIC
// when the multicall path begins with WMATIC. Excess input is returned via
// refundETH() called inside the same multicall.
const (
	uniswapV3SwapRouter02 = "0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"
	wmaticAddr            = "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270"
	usdceAddr             = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
)

const swapRouter02ABI = `[
{"name":"exactOutput","type":"function","stateMutability":"payable","inputs":[{"name":"params","type":"tuple","components":[{"name":"path","type":"bytes"},{"name":"recipient","type":"address"},{"name":"amountOut","type":"uint256"},{"name":"amountInMaximum","type":"uint256"}]}],"outputs":[{"name":"amountIn","type":"uint256"}]},
{"name":"multicall","type":"function","stateMutability":"payable","inputs":[{"name":"data","type":"bytes[]"}],"outputs":[{"name":"","type":"bytes[]"}]},
{"name":"refundETH","type":"function","stateMutability":"payable","inputs":[],"outputs":[]}
]`

// SwapPOLForExactPUSD swaps native POL for `amountPUSDOut` (6-decimal pUSD
// base units) via Uniswap V3 (WMATIC → USDC.e → pUSD multihop, both legs
// at 0.05% fee tier). msg.value = maxPOLIn; the router refunds unspent POL
// via multicall(refundETH). The pUSD lands on the EOA controlling
// privateKeyHex; transfer to deposit wallet via [TransferPUSD] separately.
//
// Returns the swap transaction hash. The caller is responsible for waiting
// for confirmation (use [PollTxStatus] or eth_getTransactionReceipt).
func SwapPOLForExactPUSD(ctx context.Context, privateKeyHex string, amountPUSDOut, maxPOLIn *big.Int, rpcURL string) (string, error) {
	privateKeyHex = strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x")
	if amountPUSDOut == nil || amountPUSDOut.Sign() <= 0 {
		return "", fmt.Errorf("amountPUSDOut must be positive")
	}
	if maxPOLIn == nil || maxPOLIn.Sign() <= 0 {
		return "", fmt.Errorf("maxPOLIn must be positive")
	}
	if rpcURL == "" {
		rpcURL = polygonRPC
	}
	priv, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return "", fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	fromAddr := crypto.PubkeyToAddress(priv.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("gas price: %w", err)
	}

	// Multihop path for V3 exactOutput is encoded output-first:
	//   pUSD || fee3000 || USDC.e || fee500 || WMATIC
	// pUSD/USDC.e: 0.30% tier — only tier with liquidity (0.05% pool is empty).
	// USDC.e/WMATIC: 0.05% tier — deepest WMATIC pool on Polygon.
	fee500 := []byte{0x00, 0x01, 0xf4}
	fee3000 := []byte{0x00, 0x0b, 0xb8}
	path := make([]byte, 0, 66)
	path = append(path, common.HexToAddress(pusdAddr).Bytes()...)
	path = append(path, fee3000...)
	path = append(path, common.HexToAddress(usdceAddr).Bytes()...)
	path = append(path, fee500...)
	path = append(path, common.HexToAddress(wmaticAddr).Bytes()...)

	parsed, err := abi.JSON(strings.NewReader(swapRouter02ABI))
	if err != nil {
		return "", fmt.Errorf("parse abi: %w", err)
	}

	type exactOutputParams struct {
		Path            []byte
		Recipient       common.Address
		AmountOut       *big.Int
		AmountInMaximum *big.Int
	}
	params := exactOutputParams{
		Path:            path,
		Recipient:       fromAddr,
		AmountOut:       amountPUSDOut,
		AmountInMaximum: maxPOLIn,
	}

	exactOutputData, err := parsed.Pack("exactOutput", params)
	if err != nil {
		return "", fmt.Errorf("pack exactOutput: %w", err)
	}
	refundETHData, err := parsed.Pack("refundETH")
	if err != nil {
		return "", fmt.Errorf("pack refundETH: %w", err)
	}
	multicallData, err := parsed.Pack("multicall", [][]byte{exactOutputData, refundETHData})
	if err != nil {
		return "", fmt.Errorf("pack multicall: %w", err)
	}

	routerAddr := common.HexToAddress(uniswapV3SwapRouter02)
	msg := ethereum.CallMsg{From: fromAddr, To: &routerAddr, Value: maxPOLIn, Data: multicallData}
	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("estimate gas (swap may revert with these params): %w", err)
	}
	gasLimit = gasLimit * 12 / 10 // 20% buffer

	tx := types.NewTransaction(nonce, routerAddr, maxPOLIn, gasLimit, gasPrice, multicallData)
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(137)), priv)
	if err != nil {
		return "", fmt.Errorf("sign tx: %w", err)
	}
	if err := client.SendTransaction(ctx, signed); err != nil {
		return "", fmt.Errorf("send tx: %w", err)
	}
	return signed.Hash().Hex(), nil
}
