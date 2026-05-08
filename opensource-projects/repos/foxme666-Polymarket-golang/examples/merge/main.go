package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/foxme666/Polymarket-golang/polymarket"
	"github.com/foxme666/Polymarket-golang/polymarket/web3"
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

	fmt.Println("\n=== Polymarket Merge Position 示例（支付 Gas） ===")
	fmt.Printf("Chain ID: %d\n", chainID)
	fmt.Printf("Signature Type: %d (%s)\n", sigType, getSignatureTypeName(sigType))
	fmt.Printf("Condition ID: %s\n", conditionID)
	fmt.Printf("Amount: %.6f 代币对\n", amount)
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

	// 执行合并
	fmt.Printf("\n=== 执行合并（Merge Position） ===\n")
	fmt.Printf("将 %.6f 个 Yes + No 代币对合并为 USDC...\n", amount)
	fmt.Println("注意: 此操作需要支付 POL Gas 费用")
	fmt.Println("      您需要同时持有相同数量的 Yes 和 No 代币")

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

	polBalanceAfter, err := client.GetPOLBalance()
	if err == nil {
		polFloatBefore, _ := polBalance.Float64()
		polFloatAfter, _ := polBalanceAfter.Float64()
		fmt.Printf("POL Balance: %.6f (Gas 费用: %.6f POL)\n", polFloatAfter, polFloatBefore-polFloatAfter)
	}

	fmt.Println("\n=== 完成 ===")
	fmt.Println("提示: Merge 操作将 Yes + No 代币对合并回 USDC")
	fmt.Println("      每对代币（1 Yes + 1 No）可兑换 1 USDC")
	fmt.Println("      这是 Split 操作的逆操作")
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
