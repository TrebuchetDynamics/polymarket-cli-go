package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	// 如果未设置，使用默认的 Polygon RPC

	// 签名类型：1=PolyProxy, 2=Safe
	sigTypeStr := os.Getenv("SIGNATURE_TYPE")
	if sigTypeStr == "" {
		sigTypeStr = "1" // 默认 PolyProxy
	}
	sigType, err := strconv.Atoi(sigTypeStr)
	if err != nil || (sigType != 1 && sigType != 2) {
		log.Fatal("错误: SIGNATURE_TYPE 必须是 1 (PolyProxy) 或 2 (Safe)")
	}

	// Condition ID（从市场获取）
	conditionID := os.Getenv("CONDITION_ID")
	if conditionID == "" {
		log.Fatal("错误: 必须设置 CONDITION_ID 环境变量")
	}

	// 合并金额（代币数量）
	amountStr := os.Getenv("AMOUNT")
	if amountStr == "" {
		amountStr = "10.0" // 默认 10 个代币
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Fatalf("无效的 AMOUNT: %v", err)
	}

	// 是否为 NegRisk 市场
	negRiskStr := os.Getenv("NEG_RISK")
	negRisk := negRiskStr == "true" || negRiskStr == "1"

	// Builder 凭证（必需）
	var builderCreds *polymarket.ApiCreds
	apiKey, err := polymarket.ResolveSecret("BUILDER_API_KEY", "BUILDER_API_KEY_ENC_FILE")
	if err != nil {
		log.Fatalf("读取 BUILDER_API_KEY 失败: %v", err)
	}
	apiSecret, err := polymarket.ResolveSecret("BUILDER_API_SECRET", "BUILDER_API_SECRET_ENC_FILE")
	if err != nil {
		log.Fatalf("读取 BUILDER_API_SECRET 失败: %v", err)
	}
	apiPassphrase, err := polymarket.ResolveSecret("BUILDER_API_PASSPHRASE", "BUILDER_API_PASSPHRASE_ENC_FILE")
	if err != nil {
		log.Fatalf("读取 BUILDER_API_PASSPHRASE 失败: %v", err)
	}
	if (apiKey != "" || apiSecret != "" || apiPassphrase != "") && (apiKey == "" || apiSecret == "" || apiPassphrase == "") {
		log.Fatal("错误: Builder 凭证必须同时提供 (BUILDER_API_KEY/SECRET/PASSPHRASE 或对应 *_ENC_FILE)")
	}
	if apiKey != "" && apiSecret != "" && apiPassphrase != "" {
		builderCreds = &polymarket.ApiCreds{
			APIKey:        apiKey,
			APISecret:     apiSecret,
			APIPassphrase: apiPassphrase,
		}
		fmt.Println("使用本地 Builder 凭证进行签名")
	} else {
		log.Fatal("错误: Gasless 交易需要 Builder 凭证 (BUILDER_API_KEY, BUILDER_API_SECRET, BUILDER_API_PASSPHRASE)")
	}

	fmt.Println("\n=== Polymarket Gasless Merge 示例 ===")
	fmt.Printf("Chain ID: %d\n", chainID)
	fmt.Printf("Signature Type: %d (%s)\n", sigType, getSignatureTypeName(sigType))
	fmt.Printf("Condition ID: %s\n", conditionID)
	fmt.Printf("Amount: %.6f 代币\n", amount)
	fmt.Printf("NegRisk: %t\n", negRisk)

	// 创建 Gasless Web3 客户端
	fmt.Println("\n正在创建 Gasless Web3 客户端...")
	client, err := web3.NewPolymarketGaslessWeb3Client(
		privateKey,
		web3.SignatureType(sigType),
		builderCreds,
		chainID,
		rpcURL,
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	fmt.Printf("✓ 客户端创建成功\n")
	fmt.Printf("  Base Address: %s\n", client.GetBaseAddress().Hex())
	fmt.Printf("  Proxy Address: %s\n", client.Address.Hex())

	// 查询当前余额
	fmt.Println("\n=== 当前余额 ===")
	polBalance, err := client.GetPOLBalance()
	if err != nil {
		log.Printf("获取 POL 余额失败: %v", err)
	} else {
		polFloat, _ := polBalance.Float64()
		fmt.Printf("POL Balance: %.4f\n", polFloat)
	}

	usdcBalanceBefore, err := client.GetUSDCBalance(common.Address{})
	if err != nil {
		log.Printf("获取 USDC 余额失败: %v", err)
	} else {
		usdcFloat, _ := usdcBalanceBefore.Float64()
		fmt.Printf("USDC Balance: %.6f\n", usdcFloat)
	}

	// 执行合并
	fmt.Printf("\n=== 执行合并 ===\n")
	fmt.Printf("正在合并 %.6f 代币对为 USDC...\n", amount)
	fmt.Println("(通过无 Gas 中继器提交交易)")

	receipt, err := client.MergePosition(common.HexToHash(conditionID), amount, negRisk)
	if err != nil {
		log.Fatalf("合并失败: %v", err)
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

	// 查询合并后的余额
	fmt.Println("\n=== 合并后余额 ===")
	usdcBalanceAfter, err := client.GetUSDCBalance(common.Address{})
	if err != nil {
		log.Printf("获取 USDC 余额失败: %v", err)
	} else {
		usdcFloatBefore, _ := usdcBalanceBefore.Float64()
		usdcFloatAfter, _ := usdcBalanceAfter.Float64()
		fmt.Printf("USDC Balance: %.6f (增加 %.6f)\n", usdcFloatAfter, usdcFloatAfter-usdcFloatBefore)
	}

	fmt.Println("\n=== 完成 ===")
	fmt.Println("提示: 合并操作将 Yes + No 代币对转换回 USDC")
	fmt.Println("      每对代币可兑换 1 USDC")
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
