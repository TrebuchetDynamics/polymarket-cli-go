package polymarket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// ParseRawOrderBookSummary 解析原始订单簿摘要
func ParseRawOrderBookSummary(rawObs map[string]interface{}) (*OrderBookSummary, error) {
	bids := []OrderSummary{}
	if bidsRaw, ok := rawObs["bids"].([]interface{}); ok {
		for _, bidRaw := range bidsRaw {
			if bid, ok := bidRaw.(map[string]interface{}); ok {
				bids = append(bids, OrderSummary{
					Price: fmt.Sprintf("%v", bid["price"]),
					Size:  fmt.Sprintf("%v", bid["size"]),
				})
			}
		}
	}

	asks := []OrderSummary{}
	if asksRaw, ok := rawObs["asks"].([]interface{}); ok {
		for _, askRaw := range asksRaw {
			if ask, ok := askRaw.(map[string]interface{}); ok {
				asks = append(asks, OrderSummary{
					Price: fmt.Sprintf("%v", ask["price"]),
					Size:  fmt.Sprintf("%v", ask["size"]),
				})
			}
		}
	}

	obs := &OrderBookSummary{
		Market:       getString(rawObs, "market"),
		AssetID:      getString(rawObs, "asset_id"),
		Timestamp:    getString(rawObs, "timestamp"),
		MinOrderSize: getString(rawObs, "min_order_size"),
		NegRisk:      getBool(rawObs, "neg_risk"),
		TickSize:     getString(rawObs, "tick_size"),
		Bids:         bids,
		Asks:         asks,
		Hash:         getString(rawObs, "hash"),
	}

	return obs, nil
}

// GenerateOrderBookSummaryHash 生成订单簿摘要哈希
func GenerateOrderBookSummaryHash(orderbook *OrderBookSummary) string {
	// 临时清空hash
	originalHash := orderbook.Hash
	orderbook.Hash = ""

	// 序列化为JSON
	jsonData, err := json.Marshal(orderbook)
	if err != nil {
		orderbook.Hash = originalHash
		return ""
	}

	// SHA1哈希
	hash := sha1.Sum(jsonData)
	hashStr := fmt.Sprintf("%x", hash)

	// 恢复hash
	orderbook.Hash = hashStr
	return hashStr
}

// OrderToJSON 将订单转换为JSON格式
// 格式与 Python py_order_utils.SignedOrder.dict() 完全一致
func OrderToJSON(order *SignedOrder, owner string, orderType OrderType) map[string]interface{} {
	return OrderToJSONWithPostOnly(order, owner, orderType, false)
}

// OrderToJSONWithPostOnly 将订单转换为JSON格式（支持 PostOnly）
func OrderToJSONWithPostOnly(order *SignedOrder, owner string, orderType OrderType, postOnly bool) map[string]interface{} {
	// 将签名从 []byte 转换为 hex 字符串（带 0x 前缀）
	var signatureHex string
	if order.Signature != nil {
		// 检查签名是否已经是 hex 格式的字符串（base64 编码）
		sigStr := string(order.Signature)
		if strings.HasPrefix(sigStr, "0x") {
			signatureHex = sigStr
		} else {
			// 尝试解码 base64
			decoded, err := base64.StdEncoding.DecodeString(sigStr)
			if err == nil {
				signatureHex = "0x" + hex.EncodeToString(decoded)
			} else {
				// 直接转换为 hex
				signatureHex = "0x" + hex.EncodeToString(order.Signature)
			}
		}
	}

	// 将地址转换为 checksummed 格式
	makerAddr := common.HexToAddress(order.Maker.Hex())
	takerAddr := common.HexToAddress(order.Taker.Hex())
	signerAddr := common.HexToAddress(order.Signer.Hex())

	// 将 side 从数字转换为字符串 "BUY" 或 "SELL"
	// Python: BUY = 0, SELL = 1
	sideStr := "BUY"
	if order.Side.Int64() == 1 {
		sideStr = "SELL"
	}

	// 将SignedOrder转换为字典
	// 格式与 Python py_order_utils.SignedOrder.dict() 完全一致
	orderDict := map[string]interface{}{
		"salt":          order.Salt.Int64(),      // 整数，不是字符串
		"maker":         makerAddr.Hex(),
		"signer":        signerAddr.Hex(),
		"taker":         takerAddr.Hex(),
		"tokenId":       order.TokenId.String(),
		"makerAmount":   order.MakerAmount.String(),
		"takerAmount":   order.TakerAmount.String(),
		"expiration":    order.Expiration.String(),
		"nonce":         order.Nonce.String(),
		"feeRateBps":    order.FeeRateBps.String(),
		"side":          sideStr,                          // 字符串 "BUY" 或 "SELL"
		"signatureType": int(order.SignatureType.Int64()), // 整数
		"signature":     signatureHex,
	}
	return map[string]interface{}{
		"order":     orderDict,
		"owner":     owner,
		"orderType": string(orderType),
		"postOnly":  postOnly,
	}
}

// IsTickSizeSmaller 检查tick size是否更小
func IsTickSizeSmaller(a, b TickSize) bool {
	aFloat, _ := strconv.ParseFloat(string(a), 64)
	bFloat, _ := strconv.ParseFloat(string(b), 64)
	return aFloat < bFloat
}

// PriceValid 检查价格是否有效
func PriceValid(price float64, tickSize TickSize) bool {
	tickSizeFloat, _ := strconv.ParseFloat(string(tickSize), 64)
	return price >= tickSizeFloat && price <= 1.0-tickSizeFloat
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

