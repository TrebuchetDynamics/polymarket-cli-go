package web3

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// 常量地址
var (
	AddressZero = common.HexToAddress("0x0000000000000000000000000000000000000000")
	HashZero    = common.Hash{}

	// Polygon 主网合约地址
	NegRiskAdapterAddress   = common.HexToAddress("0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296")
	ProxyFactoryAddress     = common.HexToAddress("0xaB45c5A4B0c941a2F231C04C3f49182e1A254052")
	SafeProxyFactoryAddress = common.HexToAddress("0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b")

	// 默认 RPC 端点
	DefaultPolygonRPC = "https://polygon-rpc.com"
)

// ChainConfig 链配置
type ChainConfig struct {
	ChainID           int64
	Exchange          common.Address
	Collateral        common.Address
	ConditionalTokens common.Address
	NegRiskExchange   common.Address
}

// 链配置映射
var chainConfigs = map[int64]*ChainConfig{
	137: { // Polygon 主网
		ChainID:           137,
		Exchange:          common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"),
		Collateral:        common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"),
		ConditionalTokens: common.HexToAddress("0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"),
		NegRiskExchange:   common.HexToAddress("0xC5d563A36AE78145C45a50134d48A1215220f80a"),
	},
	80002: { // Amoy 测试网
		ChainID:           80002,
		Exchange:          common.HexToAddress("0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40"),
		Collateral:        common.HexToAddress("0x9c4e1703476e875070ee25b56a58b008cfb8fa78"),
		ConditionalTokens: common.HexToAddress("0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB"),
		NegRiskExchange:   common.HexToAddress("0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"),
	},
}

// eth_call 调用策略
const (
	ethCallUntested = 0 // 未测试，需要自动检测
	ethCallPlain    = 1 // 无 gas 字段（适用于支持 CallDefaults 的节点）
	ethCallEIP1559  = 2 // maxFeePerGas + maxPriorityFeePerGas
	ethCallLegacy   = 3 // gasPrice (legacy mode)
)

// BaseWeb3Client Web3 基础客户端
type BaseWeb3Client struct {
	client    *ethclient.Client
	rpcClient *rpc.Client

	// eth_call 自动检测策略（Bor v2.6.0 兼容）
	ethCallMode   int         // 缓存的调用策略
	ethCallClient *rpc.Client // 使用的 RPC 客户端（可能是 fallback）
	fallbackRPC   *rpc.Client // 备用公共 RPC

	privateKey    *ecdsa.PrivateKey
	account       common.Address
	signatureType SignatureType
	chainID       int64
	config        *ChainConfig
	negRiskConfig *ChainConfig

	// 地址（根据签名类型不同）
	Address common.Address

	// 合约地址
	USDCAddress              common.Address
	ConditionalTokensAddress common.Address
	ExchangeAddress          common.Address
	NegRiskExchangeAddress   common.Address
	NegRiskAdapterAddress    common.Address
	ProxyFactoryAddress      common.Address
	SafeProxyFactoryAddress  common.Address
}

// NewBaseWeb3Client 创建基础 Web3 客户端
func NewBaseWeb3Client(
	privateKey string,
	signatureType SignatureType,
	chainID int64,
	rpcURL string,
) (*BaseWeb3Client, error) {
	if rpcURL == "" {
		rpcURL = DefaultPolygonRPC
	}

	// 连接到以太坊客户端（使用 rpc.Client 以支持 eth_call blockOverrides）
	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}
	client := ethclient.NewClient(rpcClient)

	// 解析私钥
	privKey, err := crypto.HexToECDSA(stripHexPrefix(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// 获取账户地址
	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}
	account := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取链配置
	config, ok := chainConfigs[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}

	// Neg risk 使用相同的配置但不同的交易所地址
	negRiskConfig := &ChainConfig{
		ChainID:           config.ChainID,
		Exchange:          config.NegRiskExchange,
		Collateral:        config.Collateral,
		ConditionalTokens: config.ConditionalTokens,
		NegRiskExchange:   config.NegRiskExchange,
	}

	c := &BaseWeb3Client{
		client:        client,
		rpcClient:     rpcClient,
		privateKey:    privKey,
		account:       account,
		signatureType: signatureType,
		chainID:       chainID,
		config:        config,
		negRiskConfig: negRiskConfig,

		USDCAddress:              config.Collateral,
		ConditionalTokensAddress: config.ConditionalTokens,
		ExchangeAddress:          config.Exchange,
		NegRiskExchangeAddress:   config.NegRiskExchange,
		NegRiskAdapterAddress:    NegRiskAdapterAddress,
		ProxyFactoryAddress:      ProxyFactoryAddress,
		SafeProxyFactoryAddress:  SafeProxyFactoryAddress,
	}

	// 设置地址（根据签名类型）
	if err := c.setupAddress(); err != nil {
		return nil, err
	}

	return c, nil
}

// setupAddress 根据签名类型设置地址
func (c *BaseWeb3Client) setupAddress() error {
	switch c.signatureType {
	case SignatureTypeEOA:
		c.Address = c.account
	case SignatureTypePolyProxy:
		addr, err := c.GetPolyProxyAddress(c.account)
		if err != nil {
			return err
		}
		c.Address = addr
	case SignatureTypeSafe:
		addr, err := c.GetSafeProxyAddress(c.account)
		if err != nil {
			return err
		}
		c.Address = addr
	default:
		return fmt.Errorf("invalid signature type: %d", c.signatureType)
	}
	return nil
}

// defaultCallMaxFeePerGas 用于 eth_call 的 gas 价格。
// Polygon Bor v2.6.0 升级后，某些节点会为 eth_call 填充极低的默认 maxFeePerGas（50 Mwei），
// 低于实际 baseFee（~100+ Gwei），导致 baseFee 校验失败。
// 此值仅用于只读 eth_call，不影响实际交易的 gas 费用。
var defaultCallMaxFeePerGas = big.NewInt(1e15) // 1000 Gwei

// polygonFallbackRPCs 用于 eth_call 降级的公共 RPC 端点。
// 部分 RPC 提供商（如 dRPC）的 Bor v2.6.0 节点存在 eth_call baseFee 校验 bug，
// 会忽略客户端设置的 gas 参数并填充极低的 maxFeePerGas，导致所有 eth_call 失败。
// 这些备用端点用于 eth_call 只读查询降级，不影响实际交易。
var polygonFallbackRPCs = map[int64][]string{
	137:   {"https://polygon-bor-rpc.publicnode.com", "https://1rpc.io/matic"},
	80002: {"https://rpc-amoy.polygon.technology"},
}

// ethCallStrategyTimeout 每种策略的超时时间（防止某些 RPC 端点 hang 住）
const ethCallStrategyTimeout = 5 * time.Second

// callContract 通过 eth_call 调用合约。
// 自动检测 RPC 提供商支持的调用策略（Bor v2.6.0 兼容性）：
//   - Plain: 无 gas 字段（适用于支持 CallDefaults 的 go-ethereum v1.16.7+ 节点）
//   - EIP-1559: 设置 maxFeePerGas + maxPriorityFeePerGas（适用于标准节点）
//   - Legacy: 设置 gasPrice（适用于 legacy 模式节点）
//
// 如果主 RPC 的所有策略都失败，自动降级到公共 Polygon RPC 端点。
// 检测结果会被缓存，后续调用直接使用已验证的策略。
func (c *BaseWeb3Client) callContract(ctx context.Context, to *common.Address, data []byte) ([]byte, error) {
	// 快速路径：使用已缓存的策略
	if c.ethCallMode > 0 && c.ethCallClient != nil {
		return c.rawEthCall(ctx, c.ethCallClient, to, data, c.ethCallMode)
	}

	// 自动检测：依次尝试各种策略（每种策略有独立超时，防止 hang）
	strategies := []int{ethCallPlain, ethCallEIP1559, ethCallLegacy}

	// 先尝试主 RPC
	if result, mode, err := c.tryStrategies(ctx, c.rpcClient, to, data, strategies); err == nil {
		c.ethCallMode = mode
		c.ethCallClient = c.rpcClient
		return result, nil
	}

	// 主 RPC 所有策略都失败，尝试公共 RPC 降级
	if fallbackURLs, ok := polygonFallbackRPCs[c.chainID]; ok {
		for _, fallbackURL := range fallbackURLs {
			fallbackClient, err := rpc.Dial(fallbackURL)
			if err != nil {
				continue
			}
			if result, mode, err := c.tryStrategies(ctx, fallbackClient, to, data, strategies); err == nil {
				c.ethCallMode = mode
				c.ethCallClient = fallbackClient
				c.fallbackRPC = fallbackClient
				fmt.Printf("eth_call: using fallback RPC (%s) for read-only calls\n", fallbackURL)
				return result, nil
			}
			fallbackClient.Close()
		}
	}

	return nil, fmt.Errorf("eth_call failed: all RPC endpoints and strategies exhausted")
}

// tryStrategies 依次尝试各种 eth_call 策略，每种策略有独立超时。
func (c *BaseWeb3Client) tryStrategies(ctx context.Context, client *rpc.Client, to *common.Address, data []byte, strategies []int) ([]byte, int, error) {
	var lastErr error
	for _, mode := range strategies {
		strategyCtx, cancel := context.WithTimeout(ctx, ethCallStrategyTimeout)
		result, err := c.rawEthCall(strategyCtx, client, to, data, mode)
		cancel()
		if err == nil {
			return result, mode, nil
		}
		lastErr = err
	}
	return nil, 0, lastErr
}

// rawEthCall 使用指定策略执行 eth_call。
// 使用 raw rpc.Client 而非 ethclient.CallContract，不发送 from 字段。
func (c *BaseWeb3Client) rawEthCall(ctx context.Context, client *rpc.Client, to *common.Address, data []byte, mode int) ([]byte, error) {
	type ethCallArgs struct {
		To       *common.Address `json:"to"`
		Data     hexutil.Bytes   `json:"data"`
		GasPrice *hexutil.Big    `json:"gasPrice,omitempty"`
		MaxFee   *hexutil.Big    `json:"maxFeePerGas,omitempty"`
		MaxTip   *hexutil.Big    `json:"maxPriorityFeePerGas,omitempty"`
	}

	args := ethCallArgs{To: to, Data: data}
	switch mode {
	case ethCallEIP1559:
		args.MaxFee = (*hexutil.Big)(defaultCallMaxFeePerGas)
		args.MaxTip = (*hexutil.Big)(new(big.Int))
	case ethCallLegacy:
		args.GasPrice = (*hexutil.Big)(defaultCallMaxFeePerGas)
	}

	var result hexutil.Bytes
	err := client.CallContext(ctx, &result, "eth_call", args, "latest")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetBaseAddress 获取基础 EOA 地址
func (c *BaseWeb3Client) GetBaseAddress() common.Address {
	return c.account
}

// GetPolyProxyAddress 获取 Polymarket 代理地址
func (c *BaseWeb3Client) GetPolyProxyAddress(address common.Address) (common.Address, error) {
	// 调用 CTFExchange 合约的 getPolyProxyWalletAddress 方法
	data, err := CTFExchangeABI.Pack("getPolyProxyWalletAddress", address)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := c.callContract(context.Background(), &c.ExchangeAddress, data)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to call contract: %w", err)
	}

	var proxyAddress common.Address
	err = CTFExchangeABI.UnpackIntoInterface(&proxyAddress, "getPolyProxyWalletAddress", result)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to unpack result: %w", err)
	}

	return proxyAddress, nil
}

// GetSafeProxyAddress 获取 Safe 代理地址
func (c *BaseWeb3Client) GetSafeProxyAddress(address common.Address) (common.Address, error) {
	// 调用 SafeProxyFactory 合约的 computeProxyAddress 方法
	data, err := SafeProxyFactoryABI.Pack("computeProxyAddress", address)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := c.callContract(context.Background(), &c.SafeProxyFactoryAddress, data)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to call contract: %w", err)
	}

	var proxyAddress common.Address
	err = SafeProxyFactoryABI.UnpackIntoInterface(&proxyAddress, "computeProxyAddress", result)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to unpack result: %w", err)
	}

	return proxyAddress, nil
}

// GetPOLBalance 获取 POL 余额
func (c *BaseWeb3Client) GetPOLBalance() (*big.Float, error) {
	balance, err := c.client.BalanceAt(context.Background(), c.account, nil)
	if err != nil {
		return nil, err
	}

	// 转换为 POL（18 位小数）
	balanceFloat := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetInt(big.NewInt(1e18))
	result := new(big.Float).Quo(balanceFloat, divisor)

	return result, nil
}

// GetUSDCBalance 获取 USDC 余额
func (c *BaseWeb3Client) GetUSDCBalance(address common.Address) (*big.Float, error) {
	if address == (common.Address{}) {
		address = c.Address
	}

	// 调用 USDC 合约的 balanceOf 方法
	data, err := USDCABI.Pack("balanceOf", address)
	if err != nil {
		return nil, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := c.callContract(context.Background(), &c.USDCAddress, data)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var balance *big.Int
	err = USDCABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	// 转换为 USDC（6 位小数）
	balanceFloat := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetInt(big.NewInt(1e6))
	resultFloat := new(big.Float).Quo(balanceFloat, divisor)

	return resultFloat, nil
}

// GetTokenBalance 获取条件代币余额
func (c *BaseWeb3Client) GetTokenBalance(tokenID string, address common.Address) (*big.Float, error) {
	if address == (common.Address{}) {
		address = c.Address
	}

	tokenIDBig := new(big.Int)
	tokenIDBig.SetString(tokenID, 10)

	// 调用 ConditionalTokens 合约的 balanceOf 方法
	data, err := ConditionalTokensABI.Pack("balanceOf", address, tokenIDBig)
	if err != nil {
		return nil, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := c.callContract(context.Background(), &c.ConditionalTokensAddress, data)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var balance *big.Int
	err = ConditionalTokensABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	// 转换（6 位小数）
	balanceFloat := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetInt(big.NewInt(1e6))
	resultFloat := new(big.Float).Quo(balanceFloat, divisor)

	return resultFloat, nil
}

// GetTokenComplement 获取互补代币 ID
func (c *BaseWeb3Client) GetTokenComplement(tokenID string) (string, error) {
	tokenIDBig := new(big.Int)
	tokenIDBig.SetString(tokenID, 10)

	// 首先尝试 NegRiskExchange
	data, err := NegRiskExchangeABI.Pack("getComplement", tokenIDBig)
	if err != nil {
		return "", fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := c.callContract(context.Background(), &c.NegRiskExchangeAddress, data)
	if err == nil && len(result) > 0 {
		var complement *big.Int
		err = NegRiskExchangeABI.UnpackIntoInterface(&complement, "getComplement", result)
		if err == nil && complement.Sign() > 0 {
			return complement.String(), nil
		}
	}

	// 回退到 CTFExchange
	data, err = CTFExchangeABI.Pack("getComplement", tokenIDBig)
	if err != nil {
		return "", fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err = c.callContract(context.Background(), &c.ExchangeAddress, data)
	if err != nil {
		return "", fmt.Errorf("failed to call contract: %w", err)
	}

	var complement *big.Int
	err = CTFExchangeABI.UnpackIntoInterface(&complement, "getComplement", result)
	if err != nil {
		return "", fmt.Errorf("failed to unpack result: %w", err)
	}

	return complement.String(), nil
}

// GetConditionIDNegRisk 获取 neg risk 市场的 condition ID
func (c *BaseWeb3Client) GetConditionIDNegRisk(questionID common.Hash) (common.Hash, error) {
	// 调用 NegRiskAdapter 合约的 getConditionId 方法
	data, err := NegRiskAdapterABI.Pack("getConditionId", questionID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := c.callContract(context.Background(), &c.NegRiskAdapterAddress, data)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to call contract: %w", err)
	}

	var conditionID [32]byte
	err = NegRiskAdapterABI.UnpackIntoInterface(&conditionID, "getConditionId", result)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to unpack result: %w", err)
	}

	return common.BytesToHash(conditionID[:]), nil
}

// encodeUSDCApprove 编码 USDC 授权交易
func (c *BaseWeb3Client) encodeUSDCApprove(spender common.Address) ([]byte, error) {
	return USDCABI.Pack("approve", spender, MaxUint256())
}

// encodeConditionalTokensApprove 编码条件代币授权交易
func (c *BaseWeb3Client) encodeConditionalTokensApprove(spender common.Address) ([]byte, error) {
	return ConditionalTokensABI.Pack("setApprovalForAll", spender, true)
}

// encodeTransferUSDC 编码 USDC 转账交易
func (c *BaseWeb3Client) encodeTransferUSDC(to common.Address, amount *big.Int) ([]byte, error) {
	return USDCABI.Pack("transfer", to, amount)
}

// encodeTransferToken 编码条件代币转账交易
func (c *BaseWeb3Client) encodeTransferToken(tokenID string, to common.Address, amount *big.Int) ([]byte, error) {
	tokenIDBig := new(big.Int)
	tokenIDBig.SetString(tokenID, 10)
	return ConditionalTokensABI.Pack("safeTransferFrom", c.Address, to, tokenIDBig, amount, []byte{})
}

// encodeSplit 编码拆分仓位交易
func (c *BaseWeb3Client) encodeSplit(conditionID common.Hash, amount *big.Int) ([]byte, error) {
	return ConditionalTokensABI.Pack("splitPosition", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)}, amount)
}

// encodeMerge 编码合并仓位交易
func (c *BaseWeb3Client) encodeMerge(conditionID common.Hash, amount *big.Int) ([]byte, error) {
	return ConditionalTokensABI.Pack("mergePositions", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)}, amount)
}

// encodeRedeem 编码赎回仓位交易
func (c *BaseWeb3Client) encodeRedeem(conditionID common.Hash) ([]byte, error) {
	return ConditionalTokensABI.Pack("redeemPositions", c.USDCAddress, HashZero, conditionID, []*big.Int{big.NewInt(1), big.NewInt(2)})
}

// encodeRedeemNegRisk 编码 neg risk 赎回仓位交易
func (c *BaseWeb3Client) encodeRedeemNegRisk(conditionID common.Hash, amounts []*big.Int) ([]byte, error) {
	return NegRiskAdapterABI.Pack("redeemPositions", conditionID, amounts)
}

// encodeConvert 编码转换仓位交易
func (c *BaseWeb3Client) encodeConvert(negRiskMarketID common.Hash, indexSet *big.Int, amount *big.Int) ([]byte, error) {
	return NegRiskAdapterABI.Pack("convertPositions", negRiskMarketID, indexSet, amount)
}

// encodeProxy 编码代理交易
func (c *BaseWeb3Client) encodeProxy(proxyTxn interface{}) ([]byte, error) {
	return ProxyFactoryABI.Pack("proxy", []interface{}{proxyTxn})
}

// GetTransactionOpts 获取交易选项
func (c *BaseWeb3Client) GetTransactionOpts() (*bind.TransactOpts, error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), c.account)
	if err != nil {
		return nil, err
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	// 增加 5% 的 gas 价格
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(105))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))

	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, big.NewInt(c.chainID))
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(1000000)
	auth.GasPrice = gasPrice

	return auth, nil
}

// WaitForReceipt 等待交易收据
func (c *BaseWeb3Client) WaitForReceipt(txHash common.Hash) (*types.Receipt, error) {
	return bind.WaitMined(context.Background(), c.client, &types.Transaction{})
}

// Client 返回底层的 ethclient
func (c *BaseWeb3Client) Client() *ethclient.Client {
	return c.client
}

// PrivateKey 返回私钥
func (c *BaseWeb3Client) PrivateKey() *ecdsa.PrivateKey {
	return c.privateKey
}

// ChainID 返回链 ID
func (c *BaseWeb3Client) ChainID() int64 {
	return c.chainID
}

// SignatureType 返回签名类型
func (c *BaseWeb3Client) SignatureTypeValue() SignatureType {
	return c.signatureType
}

// stripHexPrefix 移除 0x 前缀
func stripHexPrefix(s string) string {
	if len(s) >= 2 && s[0:2] == "0x" {
		return s[2:]
	}
	return s
}
