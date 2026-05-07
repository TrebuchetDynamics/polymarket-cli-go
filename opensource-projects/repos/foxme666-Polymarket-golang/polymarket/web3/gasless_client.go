package web3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/foxme666/Polymarket-golang/polymarket"
)

// PolymarketGaslessWeb3Client Polymarket Web3客户端（无gas交易）
// 仅支持：
// - Poly代理钱包 (signature_type=1)
// - Safe/Gnosis钱包 (signature_type=2)
type PolymarketGaslessWeb3Client struct {
	*BaseWeb3Client
	signer       *polymarket.Signer
	relayConfig  RelayConfig
	builderCreds *polymarket.ApiCreds
	httpClient   *http.Client
}

// NewPolymarketGaslessWeb3Client 创建新的PolymarketGaslessWeb3Client
func NewPolymarketGaslessWeb3Client(
	privateKey string,
	signatureType SignatureType,
	builderCreds *polymarket.ApiCreds,
	chainID int64,
	rpcURL string,
) (*PolymarketGaslessWeb3Client, error) {
	if signatureType != SignatureTypePolyProxy && signatureType != SignatureTypeSafe {
		return nil, fmt.Errorf("PolymarketGaslessWeb3Client only supports signature_type=1 (Poly proxy wallets) and signature_type=2 (Safe wallets)")
	}

	base, err := NewBaseWeb3Client(privateKey, signatureType, chainID, rpcURL)
	if err != nil {
		return nil, err
	}

	signer, err := polymarket.NewSigner(privateKey, int(chainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	return &PolymarketGaslessWeb3Client{
		BaseWeb3Client: base,
		signer:         signer,
		relayConfig:    DefaultRelayConfig,
		builderCreds:   builderCreds,
		httpClient:     &http.Client{},
	}, nil
}

// Execute 通过无gas中继执行交易
func (c *PolymarketGaslessWeb3Client) Execute(to common.Address, data []byte, operationName string, metadata string) (*TransactionReceipt, error) {
	var body *RelaySubmitRequest
	var err error

	switch c.signatureType {
	case SignatureTypePolyProxy:
		// 原始 to 地址会被包装在 ProxyCall 中
		// ProxyFactoryAddress 只在签名结构和请求中使用
		body, err = c.buildProxyRelayTransaction(to, data, metadata)
	case SignatureTypeSafe:
		body, err = c.buildSafeRelayTransaction(to, data, metadata)
	default:
		return nil, fmt.Errorf("invalid signature_type: %d", c.signatureType)
	}
	if err != nil {
		return nil, err
	}

	// 获取headers
	headers, err := c.getRelayHeaders(body)
	if err != nil {
		return nil, err
	}

	// 提交到中继
	resp, err := c.submitToRelay(body, headers)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Gasless txn submitted: %s\n", resp.TransactionHash)
	fmt.Printf("Transaction ID: %s\n", resp.TransactionID)
	fmt.Printf("State: %s\n", resp.State)

	// 等待确认
	if resp.TransactionHash != "" {
		receipt, err := c.waitForReceipt(common.HexToHash(resp.TransactionHash))
		if err != nil {
			return nil, err
		}

		if receipt.Status == 1 {
			fmt.Printf("%s succeeded\n", operationName)
		} else {
			fmt.Printf("%s failed\n", operationName)
		}

		return receipt, nil
	}

	return nil, fmt.Errorf("no transaction hash in response: %+v", resp)
}

// getRelayNonce 获取中继nonce
func (c *PolymarketGaslessWeb3Client) getRelayNonce(walletType string) (int, error) {
	url := fmt.Sprintf("%s/nonce?address=%s&type=%s", c.relayConfig.RelayURL, c.GetBaseAddress().Hex(), walletType)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to get nonce: %s", string(body))
	}

	var result struct {
		Nonce string `json:"nonce"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode nonce response: %w", err)
	}

	nonce, err := strconv.Atoi(result.Nonce)
	if err != nil {
		return 0, fmt.Errorf("invalid nonce value: %w", err)
	}

	return nonce, nil
}

// getRelayPayload 获取中继负载信息（包含动态 Relay 地址和 nonce）
// 这个方法调用 /relay-payload 端点，返回当前应该使用的 Relay 节点地址和 nonce
func (c *PolymarketGaslessWeb3Client) getRelayPayload(walletType string) (relayAddress string, nonce int, err error) {
	url := fmt.Sprintf("%s/relay-payload?address=%s&type=%s", c.relayConfig.RelayURL, c.GetBaseAddress().Hex(), walletType)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get relay payload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("failed to get relay payload: %s", string(body))
	}

	var result struct {
		Address string `json:"address"`
		Nonce   string `json:"nonce"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to decode relay payload response: %w", err)
	}

	nonceInt, err := strconv.Atoi(result.Nonce)
	if err != nil {
		return "", 0, fmt.Errorf("invalid nonce value: %w", err)
	}

	return result.Address, nonceInt, nil
}

// buildProxyRelayTransaction 构建Proxy中继交易
func (c *PolymarketGaslessWeb3Client) buildProxyRelayTransaction(to common.Address, data []byte, metadata string) (*RelaySubmitRequest, error) {
	// 使用 getRelayPayload 获取动态 Relay 地址和 nonce
	relayAddress, proxyNonce, err := c.getRelayPayload("PROXY")
	if err != nil {
		return nil, err
	}

	gasPrice := "0"
	relayerFee := "0"

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
	encodedTxnHex := "0x" + common.Bytes2Hex(proxyData)

	// 估算gas
	var gasLimit string
	gas, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: c.GetBaseAddress(),
		To:   &c.ProxyFactoryAddress,
		Data: proxyData,
	})
	if err != nil {
		gasLimit = "10000000"
	} else {
		gasLimit = strconv.FormatUint(uint64(float64(gas)*1.3)+100000, 10)
	}

	// 创建签名结构
	structBytes := CreateProxyStruct(
		c.GetBaseAddress().Hex(),
		c.ProxyFactoryAddress.Hex(),
		encodedTxnHex,
		relayerFee,
		gasPrice,
		gasLimit,
		strconv.Itoa(proxyNonce),
		c.relayConfig.RelayHub,
		relayAddress, // 使用动态获取的 Relay 地址
	)

	structHash := Keccak256Hash(structBytes)

	// 签名（eth_sign风格）
	prefixedHash := crypto.Keccak256(append([]byte("\x19Ethereum Signed Message:\n32"), common.FromHex(structHash)...))
	sig, err := crypto.Sign(prefixedHash, c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	// 调整v值
	if sig[64] < 27 {
		sig[64] += 27
	}
	signature := "0x" + common.Bytes2Hex(sig)

	proxyWalletAddr, err := c.GetPolyProxyAddress(c.account)
	if err != nil {
		return nil, err
	}

	return &RelaySubmitRequest{
		Data:        encodedTxnHex,
		From:        c.GetBaseAddress().Hex(),
		Metadata:    metadata,
		Nonce:       strconv.Itoa(proxyNonce),
		ProxyWallet: proxyWalletAddr.Hex(),
		Signature:   signature,
		SignatureParams: map[string]interface{}{
			"gasPrice":   gasPrice,
			"gasLimit":   gasLimit,
			"relayerFee": relayerFee,
			"relayHub":   c.relayConfig.RelayHub,
			"relay":      relayAddress, // 使用动态获取的 Relay 地址（与签名保持一致）
		},
		To:   c.ProxyFactoryAddress.Hex(),
		Type: "PROXY",
	}, nil
}

// buildSafeRelayTransaction 构建Safe中继交易
func (c *PolymarketGaslessWeb3Client) buildSafeRelayTransaction(to common.Address, data []byte, metadata string) (*RelaySubmitRequest, error) {
	safeNonce, err := c.getRelayNonce("SAFE")
	if err != nil {
		return nil, err
	}

	// 获取Safe交易哈希
	txHash, err := c.getSafeTransactionHash(to, data, big.NewInt(int64(safeNonce)))
	if err != nil {
		return nil, fmt.Errorf("failed to get safe transaction hash: %w", err)
	}

	// 签名
	prefixedHash := crypto.Keccak256(append([]byte("\x19Ethereum Signed Message:\n32"), txHash...))
	sig, err := crypto.Sign(prefixedHash, c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// 调整v值（Safe格式）
	switch sig[64] {
	case 0, 27:
		sig[64] = 0x1f
	case 1, 28:
		sig[64] = 0x20
	}
	signature := "0x" + common.Bytes2Hex(sig)

	safeProxyAddr, err := c.GetSafeProxyAddress(c.account)
	if err != nil {
		return nil, err
	}

	return &RelaySubmitRequest{
		Data:        "0x" + common.Bytes2Hex(data),
		From:        c.GetBaseAddress().Hex(),
		Metadata:    metadata,
		Nonce:       strconv.Itoa(safeNonce),
		ProxyWallet: safeProxyAddr.Hex(),
		Signature:   signature,
		SignatureParams: map[string]interface{}{
			"baseGas":        "0",
			"gasPrice":       "0",
			"gasToken":       AddressZero.Hex(),
			"operation":      "0",
			"refundReceiver": AddressZero.Hex(),
			"safeTxnGas":     "0",
		},
		To:   to.Hex(),
		Type: "SAFE",
	}, nil
}

// getSafeTransactionHash 获取Safe交易哈希
func (c *PolymarketGaslessWeb3Client) getSafeTransactionHash(to common.Address, data []byte, nonce *big.Int) ([]byte, error) {
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

// getRelayHeaders 获取中继请求headers
func (c *PolymarketGaslessWeb3Client) getRelayHeaders(body *RelaySubmitRequest) (map[string]string, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	if c.builderCreds == nil {
		// 使用签名服务器获取headers
		payload := map[string]interface{}{
			"method": "POST",
			"path":   "/submit",
			"body":   string(bodyJSON),
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Post(c.relayConfig.SignURL, "application/json", bytes.NewReader(payloadJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to get headers from sign server: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("sign server error: %s", string(respBody))
		}

		var headers map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
			return nil, fmt.Errorf("failed to decode headers: %w", err)
		}

		return headers, nil
	}

	// 使用本地 Builder 凭证创建headers（使用 POLY_BUILDER_* 头部）
	requestArgs := &polymarket.RequestArgs{
		Method:      "POST",
		RequestPath: "/submit",
		Body:        body,
	}

	headers, err := polymarket.CreateBuilderHeaders(c.builderCreds, requestArgs)
	if err != nil {
		return nil, err
	}

	return headers, nil
}

// submitToRelay 提交到中继
func (c *PolymarketGaslessWeb3Client) submitToRelay(body *RelaySubmitRequest, headers map[string]string) (*RelayResponse, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/submit", c.relayConfig.RelayURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit to relay: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("relay error: %s", string(respBody))
	}

	var result RelayResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode relay response: %w", err)
	}

	return &result, nil
}

// waitForReceipt 等待交易回执
// 添加轮询间隔和超时机制，避免 RPC 429 限流
func (c *PolymarketGaslessWeb3Client) waitForReceipt(txHash common.Hash) (*TransactionReceipt, error) {
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
func (c *PolymarketGaslessWeb3Client) SplitPosition(conditionID common.Hash, amount float64, negRisk bool) (*TransactionReceipt, error) {
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

	return c.Execute(to, data, "Split Position", "split")
}

// MergePosition 合并两个互补头寸为USDC
func (c *PolymarketGaslessWeb3Client) MergePosition(conditionID common.Hash, amount float64, negRisk bool) (*TransactionReceipt, error) {
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

	return c.Execute(to, data, "Merge Position", "merge")
}

// RedeemPosition 赎回头寸为USDC
func (c *PolymarketGaslessWeb3Client) RedeemPosition(conditionID common.Hash, amounts []float64, negRisk bool) (*TransactionReceipt, error) {
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

	return c.Execute(to, data, "Redeem Position", "redeem")
}

// ConvertPositions 转换NegRisk No头寸为Yes头寸和USDC
func (c *PolymarketGaslessWeb3Client) ConvertPositions(questionIDs []string, amount float64) (*TransactionReceipt, error) {
	amountInt := ToWei(amount, 6)
	negRiskMarketID := common.HexToHash(questionIDs[0][:len(questionIDs[0])-2] + "00")
	indexSet := big.NewInt(int64(GetIndexSet(questionIDs)))

	to := NegRiskAdapterAddress
	data, err := NegRiskAdapterABI.Pack("convertPositions", negRiskMarketID, indexSet, amountInt)
	if err != nil {
		return nil, err
	}

	return c.Execute(to, data, "Convert Positions", "convert")
}

// RedeemRequest 赎回请求
type RedeemRequest struct {
	ConditionID common.Hash
	Amounts     []float64
	NegRisk     bool
}

// RedeemPositions 批量赎回多个头寸（单次 gasless 交易）
// 将多个赎回操作打包到同一个 proxy(calls) 调用中
func (c *PolymarketGaslessWeb3Client) RedeemPositions(requests []RedeemRequest) (*TransactionReceipt, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no redeem requests provided")
	}

	// 构建多个 ProxyCall
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

	return c.ExecuteBatch(calls, "Batch Redeem Positions", "batch_redeem")
}

// ExecuteBatch 通过无gas中继执行批量交易
func (c *PolymarketGaslessWeb3Client) ExecuteBatch(calls []ProxyCall, operationName string, metadata string) (*TransactionReceipt, error) {
	var body *RelaySubmitRequest
	var err error

	switch c.signatureType {
	case SignatureTypePolyProxy:
		body, err = c.buildBatchProxyRelayTransaction(calls, metadata)
	default:
		return nil, fmt.Errorf("batch execution only supports PolyProxy signature type")
	}
	if err != nil {
		return nil, err
	}

	// 获取headers
	headers, err := c.getRelayHeaders(body)
	if err != nil {
		return nil, err
	}

	// 提交到中继
	resp, err := c.submitToRelay(body, headers)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Gasless batch txn submitted: %s\n", resp.TransactionHash)
	fmt.Printf("Transaction ID: %s\n", resp.TransactionID)
	fmt.Printf("State: %s\n", resp.State)

	// 等待确认
	if resp.TransactionHash != "" {
		receipt, err := c.waitForReceipt(common.HexToHash(resp.TransactionHash))
		if err != nil {
			return nil, err
		}

		if receipt.Status == 1 {
			fmt.Printf("%s succeeded\n", operationName)
		} else {
			fmt.Printf("%s failed\n", operationName)
		}

		return receipt, nil
	}

	return nil, fmt.Errorf("no transaction hash in response: %+v", resp)
}

// buildBatchProxyRelayTransaction 构建批量Proxy中继交易
func (c *PolymarketGaslessWeb3Client) buildBatchProxyRelayTransaction(calls []ProxyCall, metadata string) (*RelaySubmitRequest, error) {
	// 使用 getRelayPayload 获取动态 Relay 地址和 nonce
	relayAddress, proxyNonce, err := c.getRelayPayload("PROXY")
	if err != nil {
		return nil, err
	}

	gasPrice := "0"
	relayerFee := "0"

	// 编码代理交易 - 使用传入的 calls 数组
	proxyData, err := ProxyFactoryABI.Pack("proxy", calls)
	if err != nil {
		return nil, fmt.Errorf("failed to encode batch proxy transaction: %w", err)
	}
	encodedTxnHex := "0x" + common.Bytes2Hex(proxyData)

	// 估算gas（使用更大的默认值，因为是批量操作）
	var gasLimit string
	gas, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: c.GetBaseAddress(),
		To:   &c.ProxyFactoryAddress,
		Data: proxyData,
	})
	if err != nil {
		gasLimit = "15000000" // 批量操作使用更高的默认 gas limit
	} else {
		gasLimit = strconv.FormatUint(uint64(float64(gas)*1.5)+200000, 10)
	}

	// 创建签名结构
	structBytes := CreateProxyStruct(
		c.GetBaseAddress().Hex(),
		c.ProxyFactoryAddress.Hex(),
		encodedTxnHex,
		relayerFee,
		gasPrice,
		gasLimit,
		strconv.Itoa(proxyNonce),
		c.relayConfig.RelayHub,
		relayAddress,
	)

	structHash := Keccak256Hash(structBytes)

	// 签名（eth_sign风格）
	prefixedHash := crypto.Keccak256(append([]byte("\x19Ethereum Signed Message:\n32"), common.FromHex(structHash)...))
	sig, err := crypto.Sign(prefixedHash, c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	// 调整v值
	if sig[64] < 27 {
		sig[64] += 27
	}
	signature := "0x" + common.Bytes2Hex(sig)

	proxyWalletAddr, err := c.GetPolyProxyAddress(c.account)
	if err != nil {
		return nil, err
	}

	return &RelaySubmitRequest{
		Data:        encodedTxnHex,
		From:        c.GetBaseAddress().Hex(),
		Metadata:    metadata,
		Nonce:       strconv.Itoa(proxyNonce),
		ProxyWallet: proxyWalletAddr.Hex(),
		Signature:   signature,
		SignatureParams: map[string]interface{}{
			"gasPrice":   gasPrice,
			"gasLimit":   gasLimit,
			"relayerFee": relayerFee,
			"relayHub":   c.relayConfig.RelayHub,
			"relay":      relayAddress,
		},
		To:   c.ProxyFactoryAddress.Hex(),
		Type: "PROXY",
	}, nil
}
