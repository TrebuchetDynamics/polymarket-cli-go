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
	depositWalletFactoryAddr = "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"
	deployGasLimitFloor      = 600_000
)

// deploySelector = first 4 bytes of keccak256("deploy(address[],bytes32[])")
var deploySelector = crypto.Keccak256([]byte("deploy(address[],bytes32[])"))[:4]

// DeployDepositWalletEstimate does a gas-estimation-only call against the
// Polymarket DepositWalletFactory's deploy(...) function from the EOA. No
// transaction is signed or sent; no gas is spent. Returns the estimated
// gas on success, or an error containing the revert reason if deploy() is
// gated to admin/operator (or otherwise rejects the EOA caller).
//
// Use this to test whether the on-chain deploy path is available without
// risking gas.
func DeployDepositWalletEstimate(ctx context.Context, privateKeyHex, rpcURL string) (uint64, error) {
	privateKeyHex = strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x")
	if rpcURL == "" {
		rpcURL = polygonRPC
	}
	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return 0, fmt.Errorf("invalid private key: %w", err)
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return 0, fmt.Errorf("connect polygon rpc: %w", err)
	}
	defer client.Close()

	from := crypto.PubkeyToAddress(key.PublicKey)
	factory := common.HexToAddress(depositWalletFactoryAddr)
	calldata := encodeDeployCalldata(from)

	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		To:   &factory,
		Data: calldata,
	})
	if err != nil {
		return 0, fmt.Errorf("estimate gas: %w (deploy() likely gated to admin/operator; on-chain path not available from EOA)", err)
	}
	return gas, nil
}

// encodeDeployCalldata builds the ABI-encoded calldata for
// factory.deploy([owner], [bytes32(owner-padded)]) — single-element arrays.
func encodeDeployCalldata(owner common.Address) []byte {
	walletID := common.LeftPadBytes(owner.Bytes(), 32)
	out := make([]byte, 0, 4+32*6)
	out = append(out, deploySelector...)
	out = append(out, common.LeftPadBytes(big.NewInt(0x40).Bytes(), 32)...) // offset to _owners
	out = append(out, common.LeftPadBytes(big.NewInt(0xa0).Bytes(), 32)...) // offset to _ids
	out = append(out, common.LeftPadBytes(big.NewInt(1).Bytes(), 32)...)    // _owners length
	out = append(out, common.LeftPadBytes(owner.Bytes(), 32)...)            // _owners[0]
	out = append(out, common.LeftPadBytes(big.NewInt(1).Bytes(), 32)...)    // _ids length
	out = append(out, walletID...)                                          // _ids[0]
	return out
}

// DeployDepositWalletOnchain calls the Polymarket DepositWalletFactory's
// deploy(address[] _owners, bytes32[] _ids) function directly from the EOA
// (no relayer / no builder credentials). The EOA pays gas. Returns the tx
// hash on success or an error if the tx reverts (e.g. if deploy() is gated
// to admin/operator only).
//
// walletId is the deterministic per-owner identifier:
//
//	walletId = bytes32(owner) // owner address left-padded to 32 bytes
//
// per docs.polymarket.com/trading/deposit-wallets.
func DeployDepositWalletOnchain(ctx context.Context, privateKeyHex, rpcURL string) (string, error) {
	privateKeyHex = strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x")
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

	from := crypto.PubkeyToAddress(key.PublicKey)
	factory := common.HexToAddress(depositWalletFactoryAddr)
	calldata := encodeDeployCalldata(from)

	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("suggest gas: %w", err)
	}
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("get chain id: %w", err)
	}

	// Estimate gas — surfaces a revert with a useful error message before
	// sending. If estimate fails, the tx would fail too; report the reason.
	if _, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		To:   &factory,
		Data: calldata,
	}); err != nil {
		return "", fmt.Errorf("estimate gas: %w (likely deploy() is gated to admin/operator and rejects direct EOA calls)", err)
	}

	tx := types.NewTransaction(nonce, factory, big.NewInt(0), deployGasLimitFloor, gasPrice, calldata)
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
				return txHash, fmt.Errorf("deploy reverted: %s (deploy() is gated; on-chain path not available)", txHash)
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
