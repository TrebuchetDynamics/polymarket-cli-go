package web3

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// SignatureType 签名类型
type SignatureType int

const (
	// SignatureTypeEOA EOA钱包 (0)
	SignatureTypeEOA SignatureType = 0
	// SignatureTypePolyProxy Poly代理钱包 (1)
	SignatureTypePolyProxy SignatureType = 1
	// SignatureTypeSafe Safe/Gnosis钱包 (2)
	SignatureTypeSafe SignatureType = 2
)

// TransactionReceipt 交易回执
type TransactionReceipt struct {
	TxHash            common.Hash    `json:"transactionHash"`
	TxIndex           uint           `json:"transactionIndex"`
	BlockHash         common.Hash    `json:"blockHash"`
	BlockNumber       uint64         `json:"blockNumber"`
	Status            uint64         `json:"status"` // 1 = success, 0 = failure
	Type              uint8          `json:"type"`
	GasUsed           uint64         `json:"gasUsed"`
	CumulativeGasUsed uint64         `json:"cumulativeGasUsed"`
	EffectiveGasPrice *big.Int       `json:"effectiveGasPrice"`
	From              common.Address `json:"from"`
	To                common.Address `json:"to"`
	ContractAddress   common.Address `json:"contractAddress,omitempty"`
	Logs              []*types.Log   `json:"logs"`
}

// FromEthReceipt 从 go-ethereum 的 Receipt 转换
func FromEthReceipt(receipt *types.Receipt, from common.Address) *TransactionReceipt {
	return &TransactionReceipt{
		TxHash:            receipt.TxHash,
		TxIndex:           receipt.TransactionIndex,
		BlockHash:         receipt.BlockHash,
		BlockNumber:       receipt.BlockNumber.Uint64(),
		Status:            receipt.Status,
		Type:              receipt.Type,
		GasUsed:           receipt.GasUsed,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		EffectiveGasPrice: receipt.EffectiveGasPrice,
		From:              from,
		ContractAddress:   receipt.ContractAddress,
		Logs:              receipt.Logs,
	}
}

// RelayConfig 中继器配置
type RelayConfig struct {
	RelayURL     string
	SignURL      string
	RelayHub     string
	RelayAddress string
}

// DefaultRelayConfig 默认中继器配置
var DefaultRelayConfig = RelayConfig{
	RelayURL:     "https://relayer-v2.polymarket.com",
	SignURL:      "https://builder-signing-server.vercel.app/sign",
	RelayHub:     "0xD216153c06E857cD7f72665E0aF1d7D82172F494",
	RelayAddress: "0x7db63fe6d62eb73fb01f8009416f4c2bb4fbda6a",
}

// ProxyTransaction 代理交易结构（用于 JSON）
type ProxyTransaction struct {
	TypeCode int    `json:"typeCode"`
	To       string `json:"to"`
	Value    int    `json:"value"`
	Data     string `json:"data"`
}

// ProxyCall ABI 编码用的代理调用结构
// 必须匹配 ProxyWalletFactory.proxy(tuple[]) 的参数类型
type ProxyCall struct {
	TypeCode uint8
	To       common.Address
	Value    *big.Int
	Data     []byte
}

// SafeTransaction Safe交易结构
type SafeTransaction struct {
	To        string `json:"to"`
	Data      string `json:"data"`
	Operation int    `json:"operation"`
	Value     int    `json:"value"`
}

// RelaySubmitRequest 中继提交请求
type RelaySubmitRequest struct {
	Data            string                 `json:"data"`
	From            string                 `json:"from"`
	Metadata        string                 `json:"metadata"`
	Nonce           string                 `json:"nonce"`
	ProxyWallet     string                 `json:"proxyWallet"`
	Signature       string                 `json:"signature"`
	SignatureParams map[string]interface{} `json:"signatureParams"`
	To              string                 `json:"to"`
	Type            string                 `json:"type"` // "PROXY" or "SAFE"
}

// RelayResponse 中继响应
type RelayResponse struct {
	TransactionHash string `json:"transactionHash"`
	TransactionID   string `json:"transactionID"`
	State           string `json:"state"`
}

// MaxUint256 返回最大uint256值
func MaxUint256() *big.Int {
	maxUint256 := new(big.Int)
	maxUint256.Exp(big.NewInt(2), big.NewInt(256), nil)
	maxUint256.Sub(maxUint256, big.NewInt(1))
	return maxUint256
}
