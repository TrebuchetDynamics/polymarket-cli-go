package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/foxme666/Polymarket-golang/polymarket"
	"github.com/foxme666/Polymarket-golang/polymarket/web3"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	// 从环境变量或加密文件读取配置
	privateKey, err := polymarket.ResolveSecret("PRIVATE_KEY", "PRIVATE_KEY_ENC_FILE")
	if err != nil {
		log.Fatalf("读取私钥失败: %v", err)
	}
	if privateKey == "" {
		log.Fatal("错误: 必须设置 PRIVATE_KEY 或 PRIVATE_KEY_ENC_FILE")
	}

	chainIDStr := os.Getenv("CHAIN_ID")
	if chainIDStr == "" {
		chainIDStr = "137" // 默认 Polygon 主网
	}
	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		log.Fatalf("无效的 CHAIN_ID: %v", err)
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://polygon-rpc.com" // 默认公共 RPC
	}

	// 签名类型：0=EOA, 1=PolyProxy, 2=Safe
	sigTypeStr := os.Getenv("SIGNATURE_TYPE")
	if sigTypeStr == "" {
		sigTypeStr = "0" // 默认 EOA
	}
	sigType, err := strconv.Atoi(sigTypeStr)
	if err != nil || sigType < 0 || sigType > 2 {
		log.Fatal("错误: SIGNATURE_TYPE 必须是 0 (EOA), 1 (PolyProxy) 或 2 (Safe)")
	}

	// Condition ID（从市场获取）
	conditionID := os.Getenv("CONDITION_ID")
	if conditionID == "" {
		log.Fatal("错误: 必须设置 CONDITION_ID 环境变量")
	}

	// 赎回金额（逗号分隔，例如 "10.0,0" 表示赎回 10 个 Yes 代币和 0 个 No 代币）
	// 对于非 NegRisk 市场，使用默认值 [0, 0]，合约会自动处理
	amountsStr := os.Getenv("AMOUNTS")
	var amounts []float64
	if amountsStr != "" {
		parts := strings.Split(amountsStr, ",")
		for _, p := range parts {
			amt, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				log.Fatalf("无效的 AMOUNTS: %v", err)
			}
			amounts = append(amounts, amt)
		}
	} else {
		// 默认值：只用于 NegRisk 市场，非 NegRisk 市场会忽略
		amounts = []float64{0, 0}
	}

	// 是否为 NegRisk 市场
	negRiskStr := os.Getenv("NEG_RISK")
	negRisk := negRiskStr == "true" || negRiskStr == "1"

	fmt.Println("\n=== Polymarket Redeem Position 示例（支付 Gas） ===")
	fmt.Printf("Chain ID: %d\n", chainID)
	fmt.Printf("Signature Type: %d (%s)\n", sigType, getSignatureTypeName(sigType))
	fmt.Printf("Condition ID: %s\n", conditionID)
	fmt.Printf("Amounts: %v\n", amounts)
	fmt.Printf("NegRisk: %t\n", negRisk)
	fmt.Printf("RPC URL: %s\n", rpcURL)

	// 创建 Web3 客户端（支付 Gas）
	fmt.Println("\n正在创建 Web3 客户端...")
	client, err := web3.NewPolymarketWeb3Client(
		privateKey,
		web3.SignatureType(sigType),
		chainID,
		rpcURL,
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	fmt.Printf("✓ 客户端创建成功\n")
	fmt.Printf("  Address: %s\n", client.Address.Hex())

	// 查询当前余额
	fmt.Println("\n=== 当前余额 ===")
	polBalance, err := client.GetPOLBalance()
	if err != nil {
		log.Printf("获取 POL 余额失败: %v", err)
	} else {
		polFloat, _ := polBalance.Float64()
		fmt.Printf("POL Balance: %.4f (用于支付 Gas)\n", polFloat)
		if polFloat < 0.01 {
			log.Fatal("错误: POL 余额不足，需要至少 0.01 POL 支付 Gas")
		}
	}

	usdcBalanceBefore, err := client.GetUSDCBalance(common.Address{})
	if err != nil {
		log.Printf("获取 USDC 余额失败: %v", err)
	} else {
		usdcFloat, _ := usdcBalanceBefore.Float64()
		fmt.Printf("USDC Balance: %.6f\n", usdcFloat)
	}

	// 执行赎回
	fmt.Printf("\n=== 执行赎回（Redeem Position） ===\n")
	fmt.Println("将获胜代币赎回为 USDC...")
	fmt.Println("注意: 此操作需要支付 POL Gas 费用")
	fmt.Println("      只有市场结算后，获胜方代币才能赎回")

	receipt, err := client.RedeemPosition(common.HexToHash(conditionID), amounts, negRisk)
	if err != nil {
		log.Fatalf("赎回失败: %v", err)
	}

	fmt.Println("\n=== 交易结果 ===")
	fmt.Printf("Transaction Hash: %s\n", receipt.TxHash.Hex())
	fmt.Printf("Block Number: %d\n", receipt.BlockNumber)
	fmt.Printf("Gas Used: %d\n", receipt.GasUsed)
	if receipt.Status == 1 {
		fmt.Println("Status: ✓ 成功")
	} else {
		fmt.Println("Status: ✗ 失败")
	}

	// 查询赎回后的余额
	fmt.Println("\n=== 赎回后余额 ===")
	usdcBalanceAfter, err := client.GetUSDCBalance(common.Address{})
	if err != nil {
		log.Printf("获取 USDC 余额失败: %v", err)
	} else {
		usdcFloatBefore, _ := usdcBalanceBefore.Float64()
		usdcFloatAfter, _ := usdcBalanceAfter.Float64()
		fmt.Printf("USDC Balance: %.6f (增加 %.6f)\n", usdcFloatAfter, usdcFloatAfter-usdcFloatBefore)
	}

	polBalanceAfter, err := client.GetPOLBalance()
	if err == nil {
		polFloatBefore, _ := polBalance.Float64()
		polFloatAfter, _ := polBalanceAfter.Float64()
		fmt.Printf("POL Balance: %.6f (Gas 费用: %.6f POL)\n", polFloatAfter, polFloatBefore-polFloatAfter)
	}

	fmt.Println("\n=== 完成 ===")
	fmt.Println("提示: Redeem 操作在市场结算后，将获胜代币兑换为 USDC")
	fmt.Println("      - 如果结果为 Yes，每个 Yes 代币可兑换 1 USDC")
	fmt.Println("      - 如果结果为 No，每个 No 代币可兑换 1 USDC")
	fmt.Println("      - 失败方代币价值归零")
}

func getSignatureTypeName(sigType int) string {
	switch sigType {
	case 0:
		return "EOA"
	case 1:
		return "PolyProxy"
	case 2:
		return "Safe"
	default:
		return "Unknown"
	}
}
