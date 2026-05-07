package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/foxme666/Polymarket-golang/polymarket"
)

func main() {
	// 从环境变量获取配置
	host := os.Getenv("CLOB_HOST")
	if host == "" {
		host = "https://clob.polymarket.com" // 默认值
	}

	chainIDStr := os.Getenv("CHAIN_ID")
	chainID := 137 // 默认Polygon主网
	if chainIDStr != "" {
		fmt.Sscanf(chainIDStr, "%d", &chainID)
	}

	privateKey, err := polymarket.ResolveSecret("PRIVATE_KEY", "PRIVATE_KEY_ENC_FILE")
	if err != nil {
		log.Fatalf("读取私钥失败: %v", err)
	}
	if privateKey == "" {
		log.Fatalf("错误: 必须设置 PRIVATE_KEY 或 PRIVATE_KEY_ENC_FILE")
	}

	funder := os.Getenv("FUNDER") // 可选，用于代理钱包

	// 获取签名类型（Magic钱包使用1，EOA使用0）
	signatureTypeStr := os.Getenv("SIGNATURE_TYPE")
	signatureType := 0 // 默认EOA
	if signatureTypeStr != "" {
		fmt.Sscanf(signatureTypeStr, "%d", &signatureType)
	}
	sigTypePtr := &signatureType

	// 创建客户端
	client, err := polymarket.NewClobClient(
		host,
		chainID,
		privateKey,
		nil,        // 初始时没有API凭证
		sigTypePtr, // 签名类型（0=EOA, 1=Magic/Email, 2=Browser proxy）
		funder,
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	fmt.Println("=== Polymarket 下单示例 ===")
	fmt.Printf("地址: %s\n", client.GetAddress())
	fmt.Printf("链ID: %d\n", chainID)
	fmt.Printf("签名类型: %d (0=EOA, 1=Magic/Email, 2=Browser)\n", signatureType)
	fmt.Println()

	// 检查是否需要创建或派生API凭证
	apiKey, err := polymarket.ResolveSecret("CLOB_API_KEY", "CLOB_API_KEY_ENC_FILE")
	if err != nil {
		log.Fatalf("读取 CLOB_API_KEY 失败: %v", err)
	}
	apiSecret, err := polymarket.ResolveSecret("CLOB_SECRET", "CLOB_SECRET_ENC_FILE")
	if err != nil {
		log.Fatalf("读取 CLOB_SECRET 失败: %v", err)
	}
	apiPassphrase, err := polymarket.ResolveSecret("CLOB_PASSPHRASE", "CLOB_PASSPHRASE_ENC_FILE")
	if err != nil {
		log.Fatalf("读取 CLOB_PASSPHRASE 失败: %v", err)
	}
	if (apiKey != "" || apiSecret != "" || apiPassphrase != "") && (apiKey == "" || apiSecret == "" || apiPassphrase == "") {
		log.Fatal("错误: API 凭证必须同时提供 (CLOB_API_KEY/SECRET/PASSPHRASE 或对应 *_ENC_FILE)")
	}

	var creds *polymarket.ApiCreds
	if apiKey != "" && apiSecret != "" && apiPassphrase != "" {
		fmt.Println("使用环境变量中的API凭证...")
		creds = &polymarket.ApiCreds{
			APIKey:        apiKey,
			APISecret:     apiSecret,
			APIPassphrase: apiPassphrase,
		}
		client.SetAPICreds(creds)
	} else {
		fmt.Println("未找到API凭证，正在创建或派生...")
		nonce := 0
		var err error
		creds, err = client.DeriveAPIKey(&nonce)
		if err != nil {
			fmt.Println("派生失败，尝试创建新的API密钥...")
			creds, err = client.CreateAPIKey(&nonce)
			if err != nil {
				log.Fatalf("创建API密钥失败: %v", err)
			}
			fmt.Println("⚠️  新API密钥已创建，请安全保存。")
			printCredsIfAllowed(creds)
		} else {
			fmt.Println("✓ 成功派生API密钥")
		}
	}

	// 获取订单参数
	tokenID := os.Getenv("TOKEN_ID")
	if tokenID == "" {
		log.Fatalf("错误: 必须设置 TOKEN_ID 环境变量（条件代币资产ID）")
	}

	orderSide := os.Getenv("ORDER_SIDE")
	if orderSide == "" {
		orderSide = "BUY" // 默认买入
	}
	if orderSide != "BUY" && orderSide != "SELL" {
		log.Fatalf("错误: ORDER_SIDE 必须是 BUY 或 SELL")
	}

	// 判断是限价订单还是市价订单
	orderType := os.Getenv("ORDER_TYPE") // "LIMIT" 或 "MARKET"
	useMarketOrder := orderType == "MARKET"

	if useMarketOrder {
		// 市价订单
		fmt.Println("\n=== 创建市价订单 ===")

		amountStr := os.Getenv("AMOUNT")
		if amountStr == "" {
			log.Fatalf("错误: 市价订单必须设置 AMOUNT 环境变量")
		}
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Fatalf("错误: AMOUNT 必须是有效的数字: %v", err)
		}

		// 获取订单类型（GTC, FOK, FAK）
		orderTypeStr := os.Getenv("MARKET_ORDER_TYPE")
		if orderTypeStr == "" {
			orderTypeStr = "GTC" // 默认Good Till Cancel
		}
		marketOrderType := polymarket.OrderType(orderTypeStr)

		// 获取nonce
		nonceStr := os.Getenv("NONCE")
		nonce := 0
		if nonceStr != "" {
			nonce, err = strconv.Atoi(nonceStr)
			if err != nil {
				log.Fatalf("错误: NONCE 必须是有效的整数: %v", err)
			}
		}

		// 创建市价订单参数
		marketOrderArgs := &polymarket.MarketOrderArgs{
			TokenID:    tokenID,
			Amount:     amount,
			Side:       orderSide,
			Price:      0, // 市价订单，价格会自动计算
			FeeRateBps: 0, // 会自动从市场获取
			Nonce:      nonce,
			Taker:      "", // 空字符串表示公开订单
			OrderType:  marketOrderType,
		}

		fmt.Printf("订单参数:\n")
		fmt.Printf("  TokenID: %s\n", marketOrderArgs.TokenID)
		fmt.Printf("  Side: %s\n", marketOrderArgs.Side)
		fmt.Printf("  Amount: %.6f\n", marketOrderArgs.Amount)
		fmt.Printf("  OrderType: %s\n", marketOrderType)
		fmt.Printf("  Nonce: %d\n", marketOrderArgs.Nonce)
		fmt.Println()

		// 创建并签名订单
		fmt.Println("正在创建并签名订单...")
		signedOrder, err := client.CreateMarketOrder(marketOrderArgs, nil)
		if err != nil {
			log.Fatalf("创建市价订单失败: %v", err)
		}

		fmt.Println("✓ 订单创建成功")
		printOrder(signedOrder)

		// 提交订单
		fmt.Println("\n正在提交订单到交易所...")
		result, err := client.PostOrder(signedOrder, marketOrderType)
		if err != nil {
			log.Fatalf("提交订单失败: %v", err)
		}

		fmt.Println("✓ 订单提交成功")
		printResult(result)

	} else {
		// 限价订单
		fmt.Println("\n=== 创建限价订单 ===")

		priceStr := os.Getenv("PRICE")
		if priceStr == "" {
			log.Fatalf("错误: 限价订单必须设置 PRICE 环境变量")
		}
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Fatalf("错误: PRICE 必须是有效的数字: %v", err)
		}

		sizeStr := os.Getenv("SIZE")
		if sizeStr == "" {
			log.Fatalf("错误: 限价订单必须设置 SIZE 环境变量（条件代币数量）")
		}
		size, err := strconv.ParseFloat(sizeStr, 64)
		if err != nil {
			log.Fatalf("错误: SIZE 必须是有效的数字: %v", err)
		}

		// 获取nonce
		nonceStr := os.Getenv("NONCE")
		nonce := 0
		if nonceStr != "" {
			nonce, err = strconv.Atoi(nonceStr)
			if err != nil {
				log.Fatalf("错误: NONCE 必须是有效的整数: %v", err)
			}
		}

		// 获取订单类型（GTD需要过期时间，GTC不需要）
		limitOrderType := os.Getenv("LIMIT_ORDER_TYPE")
		if limitOrderType == "" {
			limitOrderType = "GTC" // 默认GTC
		}

		// 获取过期时间
		// GTC订单：expiration = 0
		// GTD订单：需要设置过期时间
		expiration := 0
		if limitOrderType == "GTD" {
			expirationStr := os.Getenv("EXPIRATION")
			if expirationStr != "" {
				exp, err := strconv.Atoi(expirationStr)
				if err != nil {
					log.Fatalf("错误: EXPIRATION 必须是有效的Unix时间戳: %v", err)
				}
				expiration = exp
			} else {
				// GTD默认30天后过期
				expiration = int(time.Now().Add(30 * 24 * time.Hour).Unix())
			}
		}

		// 创建限价订单参数
		orderArgs := &polymarket.OrderArgs{
			TokenID:    tokenID,
			Price:      price,
			Size:       size,
			Side:       orderSide,
			FeeRateBps: 0, // 会自动从市场获取
			Nonce:      nonce,
			Expiration: expiration,
			Taker:      "", // 空字符串表示公开订单
		}

		fmt.Printf("订单参数:\n")
		fmt.Printf("  TokenID: %s\n", orderArgs.TokenID)
		fmt.Printf("  Side: %s\n", orderArgs.Side)
		fmt.Printf("  Price: %.6f\n", orderArgs.Price)
		fmt.Printf("  Size: %.6f\n", orderArgs.Size)
		fmt.Printf("  Nonce: %d\n", orderArgs.Nonce)
		fmt.Printf("  OrderType: %s\n", limitOrderType)
		if orderArgs.Expiration > 0 {
			fmt.Printf("  Expiration: %d (%s)\n", orderArgs.Expiration, time.Unix(int64(orderArgs.Expiration), 0).Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  Expiration: 0 (无过期时间)\n")
		}
		fmt.Println()

		// 方式1: 使用便捷方法（创建并提交）
		useConvenienceMethod := os.Getenv("USE_CONVENIENCE") != "false"
		if useConvenienceMethod {
			fmt.Println("使用便捷方法 (CreateAndPostOrder)...")
			result, err := client.CreateAndPostOrder(orderArgs, nil)
			if err != nil {
				log.Fatalf("创建并提交订单失败: %v", err)
			}
			fmt.Println("✓ 订单创建并提交成功")
			printResult(result)
		} else {
			// 方式2: 分步操作（先创建，再提交）
			fmt.Println("正在创建并签名订单...")
			signedOrder, err := client.CreateOrder(orderArgs, nil)
			if err != nil {
				log.Fatalf("创建订单失败: %v", err)
			}

			fmt.Println("✓ 订单创建成功")
			printOrder(signedOrder)

			// 提交订单
			fmt.Println("\n正在提交订单到交易所...")
			postOrderType := polymarket.OrderType(limitOrderType)
			result, err := client.PostOrder(signedOrder, postOrderType)
			if err != nil {
				log.Fatalf("提交订单失败: %v", err)
			}

			fmt.Println("✓ 订单提交成功")
			printResult(result)
		}
	}

	fmt.Println("\n=== 完成 ===")
}

func printCredsIfAllowed(creds *polymarket.ApiCreds) {
	if os.Getenv("PRINT_API_CREDS") != "1" {
		fmt.Println("提示: 设置 PRINT_API_CREDS=1 可输出 API 凭证")
		return
	}
	fmt.Printf("   API Key: %s\n", creds.APIKey)
	fmt.Printf("   Secret: %s\n", creds.APISecret)
	fmt.Printf("   Passphrase: %s\n", creds.APIPassphrase)
	fmt.Println()
}

// printOrder 打印订单信息
func printOrder(order *polymarket.SignedOrder) {
	fmt.Println("\n订单详情:")
	jsonData, err := json.MarshalIndent(order, "  ", "  ")
	if err != nil {
		fmt.Printf("  原始数据: %+v\n", order)
	} else {
		fmt.Println(string(jsonData))
	}
}

// printResult 打印API响应结果
func printResult(result interface{}) {
	fmt.Println("\nAPI响应:")
	jsonData, err := json.MarshalIndent(result, "  ", "  ")
	if err != nil {
		fmt.Printf("  原始数据: %+v\n", result)
	} else {
		fmt.Println(string(jsonData))
	}
}
