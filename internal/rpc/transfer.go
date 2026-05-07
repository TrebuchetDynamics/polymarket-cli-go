package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	polygonRPC     = "https://polygon-bor-rpc.publicnode.com"
	pusdAddr       = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	defaultGasLimit = 100000
)

var erc20TransferSelector = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

// TransferPUSD sends pUSD from the EOA (derived from privateKeyHex) to the
// deposit wallet address. amount is in base units (e.g. "709708" for 0.709708 pUSD).
// Returns the transaction hash on success.
func TransferPUSD(ctx context.Context, privateKeyHex, toAddress string, amount *big.Int, rpcURL string) (string, error) {
	privateKeyHex = strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x")
	toAddress = strings.TrimSpace(toAddress)
	if !common.IsHexAddress(toAddress) {
		return "", fmt.Errorf("invalid to address: %s", toAddress)
	}
	if amount == nil || amount.Sign() <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}
	if rpcURL == "" {
		rpcURL = polygonRPC
	}

	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return "", fmt.Errorf("connect polygon rpc: %w", err)
	}
	defer client.Close()

	fromAddr := crypto.PubkeyToAddress(key.PublicKey)
	toAddr := common.HexToAddress(toAddress)
	tokenAddr := common.HexToAddress(pusdAddr)

	callData := make([]byte, 0, 68)
	callData = append(callData, erc20TransferSelector...)
	callData = append(callData, common.LeftPadBytes(toAddr.Bytes(), 32)...)
	callData = append(callData, common.LeftPadBytes(amount.Bytes(), 32)...)

	nonce, err := client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("suggest gas: %w", err)
	}

	msg := ethereum.CallMsg{
		From: fromAddr,
		To:   &tokenAddr,
		Data: callData,
	}
	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		gasLimit = defaultGasLimit
	}
	if gasLimit < defaultGasLimit {
		gasLimit = defaultGasLimit
	}

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("get chain id: %w", err)
	}

	tx := types.NewTransaction(nonce, tokenAddr, big.NewInt(0), gasLimit, gasPrice, callData)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		return "", fmt.Errorf("sign tx: %w", err)
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("send tx: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	for i := 0; i < 30; i++ {
		receipt, err := client.TransactionReceipt(ctx, signedTx.Hash())
		if err == nil {
			if receipt.Status != types.ReceiptStatusSuccessful {
				return txHash, fmt.Errorf("transaction reverted: %s", txHash)
			}
			return txHash, nil
		}
		select {
		case <-ctx.Done():
			return txHash, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return txHash, fmt.Errorf("tx sent but confirmation timed out: %s", txHash)
}
