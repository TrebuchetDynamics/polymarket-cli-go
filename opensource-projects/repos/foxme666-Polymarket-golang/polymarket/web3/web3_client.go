package web3

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// PolymarketWeb3Client Polymarket Web3客户端（支付gas）
// 支持：
// - EOA钱包 (signature_type=0)
// - Poly代理钱包 (signature_type=1)
// - Safe/Gnosis钱包 (signature_type=2)
type PolymarketWeb3Client struct {
	*BaseWeb3Client
}

// NewPolymarketWeb3Client 创建新的PolymarketWeb3Client
func NewPolymarketWeb3Client(privateKey string, signatureType SignatureType, chainID int64, rpcURL string) (*PolymarketWeb3Client, error) {
	base, err := NewBaseWeb3Client(privateKey, signatureType, chainID, rpcURL)
	if err != nil {
		return nil, err
	}

	return &PolymarketWeb3Client{
		BaseWeb3Client: base,
	}, nil
}

// Execute 执行链上交易
func (c *PolymarketWeb3Client) Execute(to common.Address, data []byte, operationName string) (*TransactionReceipt, error) {
	var tx *types.Transaction
	var err error

	switch c.signatureType {
	case SignatureTypeEOA:
		tx, err = c.buildEOATransaction(to, data)
	case SignatureTypePolyProxy:
		tx, err = c.buildProxyTransaction(to, data)
	case SignatureTypeSafe:
		tx, err = c.buildSafeTransaction(to, data)
	default:
		return nil, fmt.Errorf("invalid signature_type: %d", c.signatureType)
	}
	if err != nil {
		return nil, err
	}

	return c.executeTransaction(tx, operationName)
}

// buildEOATransaction 构建EOA钱包交易
func (c *PolymarketWeb3Client) buildEOATransaction(to common.Address, data []byte) (*types.Transaction, error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), c.account)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	adjustedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(105))
	adjustedGasPrice.Div(adjustedGasPrice, big.NewInt(100))

	// 估算gas
	gas, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: c.Address,
		To:   &to,
		Data: data,
	})
	if err != nil {
		gas = 500000 // 默认值
	}
	gas = uint64(float64(gas) * 1.05)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: adjustedGasPrice,
		Gas:      gas,
		To:       &to,
		Value:    big.NewInt(0),
		Data:     data,
	})

	return types.SignTx(tx, types.NewEIP155Signer(big.NewInt(c.chainID)), c.privateKey)
}

// buildProxyTransaction 构建Poly代理钱包交易
func (c *PolymarketWeb3Client) buildProxyTransaction(to common.Address, data []byte) (*types.Transaction, error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), c.account)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	adjustedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(105))
	adjustedGasPrice.Div(adjustedGasPrice, big.NewInt(100))

	// 编码代理交易 - 使用正确的切片类型
	calls := []ProxyCall{
		{
			TypeCode: 1,
			To:       to,
			Value:    big.NewInt(0),
			Data:     data,
		},
	}
	proxyData, err := ProxyFactoryABI.Pack("proxy", calls)
	if err != nil {
		return nil, fmt.Errorf("failed to encode proxy transaction: %w", err)
	}

	// 估算gas
	gas, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: c.Address,
		To:   &to,
		Data: data,
	})
	if err != nil {
		gas = 500000
	}
	gas = uint64(float64(gas)*1.05) + 100000

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: adjustedGasPrice,
		Gas:      gas,
		To:       &c.ProxyFactoryAddress,
		Value:    big.NewInt(0),
		Data:     proxyData,
	})

	return types.SignTx(tx, types.NewEIP155Signer(big.NewInt(c.chainID)), c.privateKey)
}

// buildSafeTransaction 构建Safe钱包交易
func (c *PolymarketWeb3Client) buildSafeTransaction(to common.Address, data []byte) (*types.Transaction, error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), c.account)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	adjustedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(105))
	adjustedGasPrice.Div(adjustedGasPrice, big.NewInt(100))

	// 获取Safe nonce
	safeNonceData, err := SafeABI.Pack("nonce")
	if err != nil {
		return nil, fmt.Errorf("failed to pack nonce call: %w", err)
	}

	result, err := c.callContract(context.Background(), &c.Address, safeNonceData)
	if err != nil {
		return nil, fmt.Errorf("failed to get safe nonce: %w", err)
	}

	var safeNonce *big.Int
	if err := SafeABI.UnpackIntoInterface(&safeNonce, "nonce", result); err != nil {
		return nil, fmt.Errorf("failed to unpack safe nonce: %w", err)
	}

	// 获取交易哈希
	txHash, err := c.getSafeTransactionHash(to, data, safeNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to get safe transaction hash: %w", err)
	}

	// 签名
	prefixedHash := crypto.Keccak256(append([]byte("\x19Ethereum Signed Message:\n32"), txHash...))
	sig, err := crypto.Sign(prefixedHash, c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign safe transaction: %w", err)
	}

	// 调整v值
	if sig[64] == 0 || sig[64] == 1 {
		sig[64] += 31
	} else if sig[64] == 27 || sig[64] == 28 {
		sig[64] += 4
	}

	// 估算gas
	gas, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: c.Address,
		To:   &to,
		Data: data,
	})
	if err != nil {
		gas = 500000
	}
	gas = uint64(float64(gas)*1.05) + 100000

	// 编码execTransaction调用
	execData, err := SafeABI.Pack("execTransaction",
		to,
		big.NewInt(0),
		data,
		uint8(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		AddressZero,
		AddressZero,
		sig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encode execTransaction: %w", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: adjustedGasPrice,
		Gas:      gas,
		To:       &c.Address,
		Value:    big.NewInt(0),
		Data:     execData,
	})

	return types.SignTx(tx, types.NewEIP155Signer(big.NewInt(c.chainID)), c.privateKey)
}

// getSafeTransactionHash 获取Safe交易哈希
func (c *PolymarketWeb3Client) getSafeTransactionHash(to common.Address, data []byte, nonce *big.Int) ([]byte, error) {
	txHashData, err := SafeABI.Pack("getTransactionHash",
		to,
		big.NewInt(0),
		data,
		uint8(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		AddressZero,
		AddressZero,
		nonce,
	)
	if err != nil {
		return nil, err
	}

	result, err := c.callContract(context.Background(), &c.Address, txHashData)
	if err != nil {
		return nil, err
	}

	var hash [32]byte
	if err := SafeABI.UnpackIntoInterface(&hash, "getTransactionHash", result); err != nil {
		return nil, err
	}

	return hash[:], nil
}

// executeTransaction 执行交易并等待回执
func (c *PolymarketWeb3Client) executeTransaction(tx *types.Transaction, operationName string) (*TransactionReceipt, error) {
	err := c.client.SendTransaction(context.Background(), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	fmt.Printf("Txn hash: %s\n", tx.Hash().Hex())

	receipt, err := c.waitForReceipt(tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to wait for receipt: %w", err)
	}

	if receipt.Status == 1 {
		fmt.Printf("%s succeeded\n", operationName)
	} else {
		fmt.Printf("%s failed\n", operationName)
	}

	gasUsed := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), receipt.EffectiveGasPrice)
	fmt.Printf("Paid %f POL for gas\n", FromWei(gasUsed, 18))

	return receipt, nil
}

// waitForReceipt 等待交易回执
// 添加轮询间隔和超时机制，避免 RPC 429 限流
func (c *PolymarketWeb3Client) waitForReceipt(txHash common.Hash) (*TransactionReceipt, error) {
	const (
		pollInterval = 4 * time.Second // 轮询间隔
		timeout      = 5 * time.Minute // 总超时时间
	)

	deadline := time.Now().Add(timeout)

	for {
		receipt, err := c.client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return FromEthReceipt(receipt, c.account), nil
		}

		// 检查是否超时
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for transaction receipt: %s", txHash.Hex())
		}

		// 等待后重试
		time.Sleep(pollInterval)
	}
}

// SplitPosition 分割USDC为两个互补头寸
func (c *PolymarketWeb3Client) SplitPosition(conditionID common.Hash, amount float64, negRisk bool) (*TransactionReceipt, error) {
	amountInt := ToWei(amount, 6)

	var to common.Address
	var data []byte
	var err error

	if negRisk {
		to = NegRiskAdapterAddress
		data, err = NegRiskAdapterABI.Pack("splitPosition", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)}, amountInt)
	} else {
		to = c.ConditionalTokensAddress
		data, err = ConditionalTokensABI.Pack("splitPosition", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)}, amountInt)
	}
	if err != nil {
		return nil, err
	}

	return c.Execute(to, data, "Split Position")
}

// MergePosition 合并两个互补头寸为USDC
func (c *PolymarketWeb3Client) MergePosition(conditionID common.Hash, amount float64, negRisk bool) (*TransactionReceipt, error) {
	amountInt := ToWei(amount, 6)

	var to common.Address
	var data []byte
	var err error

	if negRisk {
		to = NegRiskAdapterAddress
		data, err = NegRiskAdapterABI.Pack("mergePositions", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)}, amountInt)
	} else {
		to = c.ConditionalTokensAddress
		data, err = ConditionalTokensABI.Pack("mergePositions", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)}, amountInt)
	}
	if err != nil {
		return nil, err
	}

	return c.Execute(to, data, "Merge Position")
}

// RedeemPosition 赎回头寸为USDC
func (c *PolymarketWeb3Client) RedeemPosition(conditionID common.Hash, amounts []float64, negRisk bool) (*TransactionReceipt, error) {
	var to common.Address
	var data []byte
	var err error

	if negRisk {
		to = NegRiskAdapterAddress
		intAmounts := make([]*big.Int, len(amounts))
		for i, amt := range amounts {
			intAmounts[i] = ToWei(amt, 6)
		}
		data, err = NegRiskAdapterABI.Pack("redeemPositions", conditionID, intAmounts)
	} else {
		to = c.ConditionalTokensAddress
		data, err = ConditionalTokensABI.Pack("redeemPositions", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)})
	}
	if err != nil {
		return nil, err
	}

	return c.Execute(to, data, "Redeem Position")
}

// ConvertPositions 转换NegRisk No头寸为Yes头寸和USDC
func (c *PolymarketWeb3Client) ConvertPositions(questionIDs []string, amount float64) (*TransactionReceipt, error) {
	amountInt := ToWei(amount, 6)
	negRiskMarketID := common.HexToHash(questionIDs[0][:len(questionIDs[0])-2] + "00")
	indexSet := big.NewInt(int64(GetIndexSet(questionIDs)))

	to := NegRiskAdapterAddress
	data, err := NegRiskAdapterABI.Pack("convertPositions", negRiskMarketID, indexSet, amountInt)
	if err != nil {
		return nil, err
	}

	return c.Execute(to, data, "Convert Positions")
}

// SetCollateralApproval 设置USDC授权
func (c *PolymarketWeb3Client) SetCollateralApproval(spender common.Address) (*TransactionReceipt, error) {
	to := c.USDCAddress
	data, err := USDCABI.Pack("approve", spender, MaxUint256())
	if err != nil {
		return nil, err
	}
	return c.Execute(to, data, "Collateral Approval")
}

// SetConditionalTokensApproval 设置条件代币授权
func (c *PolymarketWeb3Client) SetConditionalTokensApproval(spender common.Address) (*TransactionReceipt, error) {
	to := c.ConditionalTokensAddress
	data, err := ConditionalTokensABI.Pack("setApprovalForAll", spender, true)
	if err != nil {
		return nil, err
	}
	return c.Execute(to, data, "Conditional Tokens Approval")
}

// SetAllApprovals 设置所有必要的授权
func (c *PolymarketWeb3Client) SetAllApprovals() ([]*TransactionReceipt, error) {
	var receipts []*TransactionReceipt

	fmt.Println("Approving ConditionalTokens as spender on USDC")
	r, err := c.SetCollateralApproval(c.ConditionalTokensAddress)
	if err != nil {
		return receipts, err
	}
	receipts = append(receipts, r)

	fmt.Println("Approving CTFExchange as spender on USDC")
	r, err = c.SetCollateralApproval(c.ExchangeAddress)
	if err != nil {
		return receipts, err
	}
	receipts = append(receipts, r)

	fmt.Println("Approving NegRiskCtfExchange as spender on USDC")
	r, err = c.SetCollateralApproval(c.NegRiskExchangeAddress)
	if err != nil {
		return receipts, err
	}
	receipts = append(receipts, r)

	fmt.Println("Approving NegRiskAdapter as spender on USDC")
	r, err = c.SetCollateralApproval(NegRiskAdapterAddress)
	if err != nil {
		return receipts, err
	}
	receipts = append(receipts, r)

	fmt.Println("Approving CTFExchange as spender on ConditionalTokens")
	r, err = c.SetConditionalTokensApproval(c.ExchangeAddress)
	if err != nil {
		return receipts, err
	}
	receipts = append(receipts, r)

	fmt.Println("Approving NegRiskCtfExchange as spender on ConditionalTokens")
	r, err = c.SetConditionalTokensApproval(c.NegRiskExchangeAddress)
	if err != nil {
		return receipts, err
	}
	receipts = append(receipts, r)

	fmt.Println("Approving NegRiskAdapter as spender on ConditionalTokens")
	r, err = c.SetConditionalTokensApproval(NegRiskAdapterAddress)
	if err != nil {
		return receipts, err
	}
	receipts = append(receipts, r)

	fmt.Println("All approvals set!")
	return receipts, nil
}

// TransferUSDC 转账USDC
func (c *PolymarketWeb3Client) TransferUSDC(recipient common.Address, amount float64) (*TransactionReceipt, error) {
	balance, err := c.GetUSDCBalance(common.Address{})
	if err != nil {
		return nil, err
	}
	balanceFloat, _ := balance.Float64()
	if balanceFloat < amount {
		return nil, fmt.Errorf("insufficient USDC balance: %f < %f", balanceFloat, amount)
	}

	amountInt := ToWei(amount, 6)
	to := c.USDCAddress
	data, err := USDCABI.Pack("transfer", recipient, amountInt)
	if err != nil {
		return nil, err
	}
	return c.Execute(to, data, "USDC Transfer")
}

// TransferToken 转账条件代币
func (c *PolymarketWeb3Client) TransferToken(tokenID string, recipient common.Address, amount float64) (*TransactionReceipt, error) {
	balance, err := c.GetTokenBalance(tokenID, common.Address{})
	if err != nil {
		return nil, err
	}
	balanceFloat, _ := balance.Float64()
	if balanceFloat < amount {
		return nil, fmt.Errorf("insufficient token balance: %f < %f", balanceFloat, amount)
	}

	amountInt := ToWei(amount, 6)
	tokenIDBig := new(big.Int)
	tokenIDBig.SetString(tokenID, 10)

	to := c.ConditionalTokensAddress
	data, err := ConditionalTokensABI.Pack("safeTransferFrom", c.Address, recipient, tokenIDBig, amountInt, []byte{})
	if err != nil {
		return nil, err
	}
	return c.Execute(to, data, "Token Transfer")
}

// RedeemPositions 批量赎回多个头寸
// 对于 PolyProxy 钱包，这将在单次链上交易中执行所有赎回操作
func (c *PolymarketWeb3Client) RedeemPositions(requests []RedeemRequest) (*TransactionReceipt, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no redeem requests provided")
	}

	// 仅 PolyProxy 支持高效打包执行
	if c.signatureType == SignatureTypePolyProxy {
		calls := make([]ProxyCall, 0, len(requests))
		for _, req := range requests {
			var to common.Address
			var data []byte
			var err error

			if req.NegRisk {
				to = NegRiskAdapterAddress
				intAmounts := make([]*big.Int, len(req.Amounts))
				for i, amt := range req.Amounts {
					intAmounts[i] = ToWei(amt, 6)
				}
				data, err = NegRiskAdapterABI.Pack("redeemPositions", req.ConditionID, intAmounts)
			} else {
				to = c.ConditionalTokensAddress
				data, err = ConditionalTokensABI.Pack("redeemPositions", c.USDCAddress, HashZero, req.ConditionID, []*big.Int{big.NewInt(1), big.NewInt(2)})
			}
			if err != nil {
				return nil, fmt.Errorf("failed to encode redeem for condition %s: %w", req.ConditionID.Hex(), err)
			}

			calls = append(calls, ProxyCall{
				TypeCode: 1,
				To:       to,
				Value:    big.NewInt(0),
				Data:     data,
			})
		}
		return c.ExecuteBatch(calls, "Batch Redeem Positions")
	}

	// 对于非打包支持的钱包类型（如 EOA），退回到串行执行
	var lastReceipt *TransactionReceipt
	for _, req := range requests {
		r, err := c.RedeemPosition(req.ConditionID, req.Amounts, req.NegRisk)
		if err != nil {
			return lastReceipt, err
		}
		lastReceipt = r
	}
	return lastReceipt, nil
}

// ExecuteBatch 执行批量链上交易
func (c *PolymarketWeb3Client) ExecuteBatch(calls []ProxyCall, operationName string) (*TransactionReceipt, error) {
	var tx *types.Transaction
	var err error

	switch c.signatureType {
	case SignatureTypePolyProxy:
		tx, err = c.buildBatchProxyTransaction(calls)
	default:
		return nil, fmt.Errorf("batch execution only supported for PolyProxy signature type on-chain")
	}

	if err != nil {
		return nil, err
	}

	return c.executeTransaction(tx, operationName)
}

// buildBatchProxyTransaction 构建批量 Poly 代理钱包交易
func (c *PolymarketWeb3Client) buildBatchProxyTransaction(calls []ProxyCall) (*types.Transaction, error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), c.account)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	adjustedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(110)) // 批量交易适当提高气价稳定性
	adjustedGasPrice.Div(adjustedGasPrice, big.NewInt(100))

	proxyData, err := ProxyFactoryABI.Pack("proxy", calls)
	if err != nil {
		return nil, fmt.Errorf("failed to encode batch proxy transaction: %w", err)
	}

	// 估算gas
	gas, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: c.account,
		To:   &c.ProxyFactoryAddress,
		Data: proxyData,
	})
	if err != nil {
		gas = 15000000 // 默认大的 gas limit
	} else {
		gas = uint64(float64(gas)*1.2) + 100000
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: adjustedGasPrice,
		Gas:      gas,
		To:       &c.ProxyFactoryAddress,
		Value:    big.NewInt(0),
		Data:     proxyData,
	})

	return types.SignTx(tx, types.NewEIP155Signer(big.NewInt(c.chainID)), c.privateKey)
}
