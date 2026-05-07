package web3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GetMarketIndex 从question_id中提取市场索引（最后2个十六进制字符）
func GetMarketIndex(questionID string) int {
	if len(questionID) < 2 {
		return 0
	}
	hexStr := questionID[len(questionID)-2:]
	val, _ := hex.DecodeString(hexStr)
	if len(val) > 0 {
		return int(val[0])
	}
	return 0
}

// GetIndexSet 从question_ids计算位索引集
func GetIndexSet(questionIDs []string) int {
	indexSet := 0
	seen := make(map[int]bool)
	for _, qid := range questionIDs {
		idx := GetMarketIndex(qid)
		if !seen[idx] {
			seen[idx] = true
			indexSet |= (1 << idx)
		}
	}
	return indexSet
}

// SplitSignature 将签名分割为r, s, v组件
func SplitSignature(signatureHex string) (r, s [32]byte, v uint8, err error) {
	signatureHex = strings.TrimPrefix(signatureHex, "0x")
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return r, s, v, fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(sigBytes) != 65 {
		return r, s, v, fmt.Errorf("invalid signature length: %d", len(sigBytes))
	}

	copy(r[:], sigBytes[:32])
	copy(s[:], sigBytes[32:64])
	v = sigBytes[64]

	// 标准化 v 值
	if v < 27 {
		if v == 0 || v == 1 {
			v += 27
		} else {
			return r, s, v, fmt.Errorf("invalid signature v value: %d", v)
		}
	}

	return r, s, v, nil
}

// GetPackedSignature 获取打包的签名用于Safe交易
func GetPackedSignature(r, s [32]byte, v uint8) []byte {
	// 调整 v 值用于 Safe
	switch v {
	case 0, 1:
		v += 31
	case 27, 28:
		v += 4
	}

	packed := make([]byte, 65)
	copy(packed[:32], r[:])
	copy(packed[32:64], s[:])
	packed[64] = v
	return packed
}

// CreateProxyStruct 创建代理钱包签名的结构哈希
func CreateProxyStruct(
	fromAddress string,
	to string,
	data string,
	txFee string,
	gasPrice string,
	gasLimit string,
	nonce string,
	relayHubAddress string,
	relayAddress string,
) []byte {
	prefix := []byte("rlx:")

	fromAddr := common.HexToAddress(fromAddress)
	toAddr := common.HexToAddress(to)
	dataBytes := common.FromHex(data)

	txFeeInt, _ := new(big.Int).SetString(txFee, 10)
	gasPriceInt, _ := new(big.Int).SetString(gasPrice, 10)
	gasLimitInt, _ := new(big.Int).SetString(gasLimit, 10)
	nonceInt, _ := new(big.Int).SetString(nonce, 10)

	relayHubAddr := common.HexToAddress(relayHubAddress)
	relayAddr := common.HexToAddress(relayAddress)

	// 构建结构
	var result []byte
	result = append(result, prefix...)
	result = append(result, fromAddr.Bytes()...)
	result = append(result, toAddr.Bytes()...)
	result = append(result, dataBytes...)
	result = append(result, common.LeftPadBytes(txFeeInt.Bytes(), 32)...)
	result = append(result, common.LeftPadBytes(gasPriceInt.Bytes(), 32)...)
	result = append(result, common.LeftPadBytes(gasLimitInt.Bytes(), 32)...)
	result = append(result, common.LeftPadBytes(nonceInt.Bytes(), 32)...)
	result = append(result, relayHubAddr.Bytes()...)
	result = append(result, relayAddr.Bytes()...)

	return result
}

// Keccak256 计算keccak256哈希
func Keccak256(data []byte) []byte {
	return crypto.Keccak256(data)
}

// Keccak256Hash 计算keccak256哈希并返回十六进制字符串
func Keccak256Hash(data []byte) string {
	return "0x" + hex.EncodeToString(crypto.Keccak256(data))
}

// ToWei 将金额转换为wei（指定小数位数）
func ToWei(amount float64, decimals int) *big.Int {
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountBig := new(big.Float).SetFloat64(amount)
	amountBig.Mul(amountBig, new(big.Float).SetInt(multiplier))

	result := new(big.Int)
	amountBig.Int(result)
	return result
}

// FromWei 将wei转换为金额
func FromWei(amount *big.Int, decimals int) float64 {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountFloat := new(big.Float).SetInt(amount)
	divisorFloat := new(big.Float).SetInt(divisor)
	amountFloat.Quo(amountFloat, divisorFloat)

	result, _ := amountFloat.Float64()
	return result
}
