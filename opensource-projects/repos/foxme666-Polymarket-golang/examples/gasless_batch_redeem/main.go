package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/0xNetuser/Polymarket-golang/polymarket"
	"github.com/0xNetuser/Polymarket-golang/polymarket/web3"
	"github.com/ethereum/go-ethereum/common"
)

// Position API 返回的头寸结构
type Position struct {
	ConditionID  string  `json:"conditionId"`
	Title        string  `json:"title"`
	Outcome      string  `json:"outcome"`
	OutcomeIndex int     `json:"outcomeIndex"`
	Size         float64 `json:"size"`
	CurrentValue float64 `json:"currentValue"`
	NegativeRisk bool    `json:"negativeRisk"`
	Redeemable   bool    `json:"redeemable"`
}

// GroupedPosition 按 conditionId 分组的头寸
type GroupedPosition struct {
	ConditionID  string
	Title        string
	NegativeRisk bool
	Outcomes     []OutcomeInfo
	TotalValue   float64
	TotalSize    float64
}

type OutcomeInfo struct {
	Outcome      string
	OutcomeIndex int
	Size         float64
	Value        float64
}

func main() {
	// 从环境变量读取配置
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("错误: 必须设置 PRIVATE_KEY 环境变量")
	}

	chainIDStr := os.Getenv("CHAIN_ID")
	if chainIDStr == "" {
		chainIDStr = "137"
	}
	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		log.Fatalf("无效的 CHAIN_ID: %v", err)
	}

	rpcURL := os.Getenv("RPC_URL")
	checkOnly := os.Getenv("CHECK_ONLY") == "true" || os.Getenv("CHECK_ONLY") == "1"

	// Builder 凭证
	var builderCreds *polymarket.ApiCreds
	apiKey := os.Getenv("BUILDER_API_KEY")
	apiSecret := os.Getenv("BUILDER_API_SECRET")
	apiPassphrase := os.Getenv("BUILDER_API_PASSPHRASE")
	if apiKey != "" && apiSecret != "" && apiPassphrase != "" {
		builderCreds = &polymarket.ApiCreds{
			APIKey:        apiKey,
			APISecret:     apiSecret,
			APIPassphrase: apiPassphrase,
		}
	} else if !checkOnly {
		log.Fatal("错误: Gasless 交易需要 Builder 凭证")
	}

	fmt.Println("\n=== Polymarket Auto Batch Redeem ===")
	fmt.Printf("Chain ID: %d\n", chainID)
	if checkOnly {
		fmt.Println("模式: 仅检查（不执行赎回）")
	} else {
		fmt.Println("模式: 执行赎回")
	}

	// 创建 Gasless Web3 客户端
	fmt.Println("\n正在创建客户端...")
	client, err := web3.NewPolymarketGaslessWeb3Client(
		privateKey,
		web3.SignatureTypePolyProxy,
		builderCreds,
		chainID,
		rpcURL,
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	proxyAddress := client.Address.Hex()
	fmt.Printf("Proxy Wallet: %s\n", proxyAddress)

	// 获取可赎回头寸
	fmt.Println("\n正在获取可赎回头寸...")
	positions, err := getRedeemablePositions(proxyAddress)
	if err != nil {
		log.Fatalf("获取头寸失败: %v", err)
	}

	if len(positions) == 0 {
		fmt.Println("✓ 没有可赎回的头寸")
		return
	}

	// 显示可赎回头寸
	fmt.Printf("\n找到 %d 个可赎回头寸:\n\n", len(positions))
	var totalValue float64
	for i, pos := range positions {
		fmt.Printf("%d. %s\n", i+1, truncate(pos.Title, 50))
		for _, out := range pos.Outcomes {
			status := "[LOSE]"
			if out.Value > 0 {
				status = "[WIN]"
			}
			fmt.Printf("   %s: Size %.4f, Value $%.4f %s\n", out.Outcome, out.Size, out.Value, status)
		}
		fmt.Printf("   ConditionID: %s\n", pos.ConditionID)
		fmt.Printf("   NegRisk: %t, Total Value: $%.4f\n", pos.NegativeRisk, pos.TotalValue)
		totalValue += pos.TotalValue
	}
	fmt.Printf("\n总可赎回价值: $%.4f\n", totalValue)

	if checkOnly {
		fmt.Println("\n(仅检查模式 - 不执行赎回)")
		fmt.Println("如需执行赎回，请移除 CHECK_ONLY 环境变量")
		return
	}

	// 查询当前余额
	fmt.Println("\n=== 当前余额 ===")
	usdcBalanceBefore, _ := client.GetUSDCBalance(common.Address{})
	if usdcBalanceBefore != nil {
		usdcFloat, _ := usdcBalanceBefore.Float64()
		fmt.Printf("USDC Balance: %.6f\n", usdcFloat)
	}

	// 构建批量赎回请求
	fmt.Printf("\n=== 执行批量赎回 ===\n")
	requests := make([]web3.RedeemRequest, len(positions))
	for i, pos := range positions {
		// 收集每个 outcome 的 size
		amounts := make([]float64, len(pos.Outcomes))
		for j, out := range pos.Outcomes {
			amounts[j] = out.Size
		}
		requests[i] = web3.RedeemRequest{
			ConditionID: common.HexToHash(pos.ConditionID),
			Amounts:     amounts,
			NegRisk:     pos.NegativeRisk,
		}
	}

	fmt.Printf("正在批量赎回 %d 个头寸...\n", len(requests))
	fmt.Println("(通过无 Gas 中继器提交批量交易)")

	receipt, err := client.RedeemPositions(requests)
	if err != nil {
		log.Fatalf("批量赎回失败: %v", err)
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

	// 查询赎回后余额
	fmt.Println("\n=== 赎回后余额 ===")
	usdcBalanceAfter, _ := client.GetUSDCBalance(common.Address{})
	if usdcBalanceAfter != nil && usdcBalanceBefore != nil {
		before, _ := usdcBalanceBefore.Float64()
		after, _ := usdcBalanceAfter.Float64()
		fmt.Printf("USDC Balance: %.6f (增加 %.6f)\n", after, after-before)
	}

	fmt.Println("\n=== 完成 ===")
	fmt.Printf("成功批量赎回 %d 个头寸\n", len(positions))
}

// getRedeemablePositions 从 Polymarket Data API 获取可赎回头寸
func getRedeemablePositions(walletAddress string) ([]GroupedPosition, error) {
	url := fmt.Sprintf("https://data-api.polymarket.com/positions?user=%s&sizeThreshold=0.01&redeemable=true&limit=100", walletAddress)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误 %d: %s", resp.StatusCode, string(body))
	}

	var positions []Position
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 按 conditionId 分组
	grouped := make(map[string]*GroupedPosition)
	for _, pos := range positions {
		if !pos.Redeemable {
			continue
		}

		if _, exists := grouped[pos.ConditionID]; !exists {
			grouped[pos.ConditionID] = &GroupedPosition{
				ConditionID:  pos.ConditionID,
				Title:        pos.Title,
				NegativeRisk: pos.NegativeRisk,
				Outcomes:     []OutcomeInfo{},
			}
		}

		g := grouped[pos.ConditionID]
		g.Outcomes = append(g.Outcomes, OutcomeInfo{
			Outcome:      pos.Outcome,
			OutcomeIndex: pos.OutcomeIndex,
			Size:         pos.Size,
			Value:        pos.CurrentValue,
		})
		g.TotalValue += pos.CurrentValue
		g.TotalSize += pos.Size
	}

	result := make([]GroupedPosition, 0, len(grouped))
	for _, g := range grouped {
		result = append(result, *g)
	}

	return result, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
