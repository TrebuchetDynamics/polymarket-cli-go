# Polymarket Go SDK

[English](README.md) | [中文](README_zh.md)

Polymarket CLOB（中央限价订单簿）API 的 Go 语言 SDK，完整实现了 [py-clob-client](https://github.com/Polymarket/py-clob-client) 的所有核心功能。


Follow at X:  @netu5er


## 特性

- ✅ **完整的 API 覆盖**: 实现 py-clob-client 的所有端点
- ✅ **三种认证级别**: L0（只读）、L1（私钥）、L2（完整访问）
- ✅ **订单管理**: 创建、提交、取消和查询订单
- ✅ **市场数据**: 订单簿、价格、价差和市场信息
- ✅ **RFQ 支持**: 报价请求功能
- ✅ **类型安全**: 使用 Go 的类型系统提供强类型支持
- ✅ **EIP-712 签名**: 完整支持以太坊消息签名

## 安装

```bash
go get github.com/foxme666/Polymarket-golang@v0.0.1
```

## 快速开始

### Level 0（只读模式）

```go
package main

import (
    "fmt"
    "github.com/foxme666/Polymarket-golang/polymarket"
)

func main() {
    // 创建只读客户端
    client, err := polymarket.NewClobClient(
        "https://clob.polymarket.com",
        137, // Polygon 链 ID
        "",  // 无私钥
        nil, // 无 API 凭证
        nil, // 无签名类型
        "",  // 无 funder 地址
    )
    if err != nil {
        panic(err)
    }

    // 健康检查
    ok, err := client.GetOK()
    fmt.Println("服务器状态:", ok)

    // 获取订单簿
    orderBook, err := client.GetOrderBook("token-id")
    if err != nil {
        panic(err)
    }
    fmt.Printf("订单簿: %+v\n", orderBook)
}
```

### Level 1（私钥认证）

```go
client, err := polymarket.NewClobClient(
    "https://clob.polymarket.com",
    137,
    "your-private-key-hex",
    nil,
    nil,
    "",
)

// 创建或派生 API 凭证
creds, err := client.CreateOrDeriveAPIKey(nil)
if err != nil {
    panic(err)
}
fmt.Printf("API Key: %s\n", creds.APIKey)
```

### Level 2（完整功能）

```go
creds := &polymarket.ApiCreds{
    APIKey:      "your-api-key",
    APISecret:   "your-api-secret",
    APIPassphrase: "your-passphrase",
}

client, err := polymarket.NewClobClient(
    "https://clob.polymarket.com",
    137,
    "your-private-key-hex",
    creds,
    nil,
    "",
)

// 查询余额
balance, err := client.GetBalanceAllowance(&polymarket.BalanceAllowanceParams{
    AssetType: polymarket.AssetTypeCollateral,
})
if err != nil {
    panic(err)
}
fmt.Printf("余额: %+v\n", balance)
```

## 示例

### 查询余额

```bash
cd examples/check_balance
export PRIVATE_KEY="your-private-key"
export CHAIN_ID="137"  # 可选，默认为 137 (Polygon)
export SIGNATURE_TYPE="0"  # 可选，0=EOA, 1=Magic/Email, 2=Browser
export FUNDER=""  # 可选，用于代理钱包
go run main.go
```

### 创建并提交订单

```go
// 创建订单
orderArgs := &polymarket.OrderArgs{
    TokenID:    "token-id",
    Price:      0.5,
    Size:       100.0,
    Side:       "BUY",
    FeeRateBps: 0,
    Nonce:      1,
    Expiration: 1234567890,
}

order, err := client.CreateOrder(orderArgs, nil)
if err != nil {
    panic(err)
}

// 提交订单
result, err := client.PostOrder(order, polymarket.OrderTypeGTC)
if err != nil {
    panic(err)
}
fmt.Printf("订单已提交: %+v\n", result)
```

### 原始订单模式（跳过服务器请求）

默认情况下，`CreateOrder` 会从服务器获取市场的 `tick_size`、`neg_risk` 和 `fee_rate`。如果你想跳过这些 API 调用并自行提供这些值，可以使用 `RawOrder` 选项：

```go
orderArgs := &polymarket.OrderArgs{
    TokenID:    "token-id",
    Price:      0.56,
    Size:       100.0,
    Side:       "BUY",
    Expiration: 0,
}

// 使用 RawOrder 模式跳过服务器请求
// 必须提供 TickSize 和 NegRisk
tickSize := polymarket.TickSize("0.01")
negRisk := false
options := &polymarket.PartialCreateOrderOptions{
    RawOrder: true,       // 跳过 tick_size/neg_risk/fee_rate 的服务器请求
    TickSize: &tickSize,  // RawOrder 模式下必须提供
    NegRisk:  &negRisk,   // RawOrder 模式下必须提供
}

order, err := client.CreateOrder(orderArgs, options)
// 或者
result, err := client.CreateAndPostOrder(orderArgs, options)
```

**注意**：在 `RawOrder` 模式下，`TickSize` 和 `NegRisk` 是**必需的**。库仍然会使用 `TickSize` 将价格/数量转换为正确的金额。

### 订单类型

`CreateAndPostOrder` 支持通过 `OrderType` 选项指定不同的订单类型：

```go
// FAK 订单（Fill And Kill）- 部分成交后取消剩余
orderType := polymarket.OrderTypeFAK
options := &polymarket.PartialCreateOrderOptions{
    OrderType: &orderType,
}
result, err := client.CreateAndPostOrder(orderArgs, options)

// FOK 订单（Fill Or Kill）- 全部成交或取消
orderType := polymarket.OrderTypeFOK
options := &polymarket.PartialCreateOrderOptions{
    OrderType: &orderType,
}
result, err := client.CreateAndPostOrder(orderArgs, options)
```

| 订单类型 | 说明 |
|----------|------|
| `GTC` | Good Till Cancel - 直到取消（默认） |
| `FOK` | Fill Or Kill - 全部成交或取消 |
| `FAK` | Fill And Kill - 部分成交后取消剩余 |
| `GTD` | Good Till Date - 直到指定时间（需要设置 `Expiration`） |

### Post Only 订单

Post Only 订单只有在为订单簿添加流动性（成为 maker 订单）时才会被接受。如果 Post Only 订单会立即匹配并成为 taker，则会被拒绝。

```go
// 使用 PostOrderWithOptions 提交 Post Only 订单
result, err := client.PostOrderWithOptions(signedOrder, polymarket.OrderTypeGTC, true) // postOnly=true

// 或使用 PostOrders 批量提交，设置 PostOnly 字段
args := []polymarket.PostOrdersArgs{
    {Order: signedOrder1, OrderType: polymarket.OrderTypeGTC, PostOnly: true},
    {Order: signedOrder2, OrderType: polymarket.OrderTypeGTD, PostOnly: true},
}
result, err := client.PostOrders(args)
```

**注意**：Post Only 仅对 `GTC` 和 `GTD` 订单类型有效。

### 心跳 API

心跳 API 允许您在连接丢失时自动取消所有订单。一旦启动，您必须每 10 秒发送一次心跳，否则所有订单将被取消。

```go
// 使用可选 ID 发送心跳
heartbeatID := "my-session-123"
result, err := client.PostHeartbeat(&heartbeatID)

// 或不使用 ID 发送
result, err := client.PostHeartbeat(nil)
```

**使用场景**：防止交易系统意外离线时留下过期订单。

## Web3 客户端

SDK 包含两个 Web3 客户端用于链上操作：

### PolymarketWeb3Client（支付 Gas）

```go
import "github.com/foxme666/Polymarket-golang/polymarket/web3"

// 创建 Web3 客户端（需要支付 gas）
client, err := web3.NewPolymarketWeb3Client(
    "your-private-key",
    web3.SignatureTypePolyProxy, // 0=EOA, 1=PolyProxy, 2=Safe
    137,                         // Chain ID
    "",                          // RPC URL（空=默认）
)

// 获取余额
polBalance, _ := client.GetPOLBalance()
usdcBalance, _ := client.GetUSDCBalance(common.Address{})
tokenBalance, _ := client.GetTokenBalance("token-id", common.Address{})

// 设置所有必要的授权
receipts, _ := client.SetAllApprovals()

// 分割 USDC 为头寸
receipt, _ := client.SplitPosition(conditionID, 100.0, true) // negRisk=true

// 合并头寸为 USDC
receipt, _ := client.MergePosition(conditionID, 100.0, true)

// 转账 USDC
receipt, _ := client.TransferUSDC(recipient, 50.0)

// 转账条件代币
receipt, _ := client.TransferToken("token-id", recipient, 50.0)
```

### PolymarketGaslessWeb3Client（无 Gas）

```go
// 创建无 Gas Web3 客户端（通过中继器交易，无需 gas）
// 仅支持 signature_type=1 (PolyProxy) 或 signature_type=2 (Safe)
client, err := web3.NewPolymarketGaslessWeb3Client(
    "your-private-key",
    web3.SignatureTypePolyProxy,
    nil,  // 可选：builder 凭证
    137,
    "",
)

// 与 PolymarketWeb3Client 相同的操作
receipt, _ := client.SplitPosition(conditionID, 100.0, true)
receipt, _ := client.MergePosition(conditionID, 100.0, true)
```

## 项目结构

```
polymarket/
├── client.go                  # 主客户端结构
├── client_api.go              # API 方法（健康检查、API 密钥、市场数据等）
├── client_orders.go           # 订单管理方法（提交、取消、查询）
├── client_order_creation.go   # 订单创建方法（CreateOrder, CreateMarketOrder）
├── client_misc.go             # 其他功能（只读 API 密钥、订单评分、市场查询等）
├── rfq_client.go              # RFQ 客户端便捷方法
├── config.go                  # 合约配置
├── constants.go               # 常量定义
├── endpoints.go               # API 端点常量
├── http_client.go             # HTTP 客户端
├── http_helpers.go            # HTTP 辅助函数（查询参数构建）
├── signer.go                  # 签名器
├── signing_internal.go        # 签名实现（EIP-712, HMAC）
├── types.go                   # 类型定义
├── utilities.go               # 工具函数
├── order_summary_wrapper.go   # OrderSummary 包装器
├── headers/                   # 认证头（包装函数）
│   └── headers.go
├── order_builder/             # 订单构建器
│   ├── order_builder.go       # 订单构建器实现
│   └── helpers.go             # 订单构建辅助函数
├── rfq/                       # RFQ 客户端
│   ├── rfq_client.go          # RFQ 客户端实现
│   └── types.go               # RFQ 类型定义
└── web3/                      # Web3 客户端（链上操作）
    ├── base_client.go         # 基础 Web3 客户端（共享逻辑）
    ├── web3_client.go         # PolymarketWeb3Client（支付 gas）
    ├── gasless_client.go      # PolymarketGaslessWeb3Client（无 gas）
    ├── types.go               # Web3 类型定义
    ├── helpers.go             # Web3 辅助函数
    ├── abi_loader.go          # ABI 加载工具
    └── abis/                   # 合约 ABI 文件
```

## 已实现功能

### ✅ 核心功能
- [x] 基础类型定义（所有 py-clob-client 的类型）
- [x] 签名器 - EIP-712 和 HMAC 签名
- [x] L0/L1/L2 认证头生成
- [x] HTTP 客户端封装
- [x] 客户端基础结构（支持三种模式）

### ✅ API 方法
- [x] **健康检查**: `GetOK()`, `GetServerTime()`
- [x] **API 密钥管理**: `CreateAPIKey()`, `DeriveAPIKey()`, `CreateOrDeriveAPIKey()`, `GetAPIKeys()`, `DeleteAPIKey()`
- [x] **市场数据**: `GetMidpoint()`, `GetMidpoints()`, `GetPrice()`, `GetPrices()`, `GetSpread()`, `GetSpreads()`
- [x] **订单簿**: `GetOrderBook()`, `GetOrderBooks()`, `GetOrderBookHash()`
- [x] **市场信息**: `GetTickSize()`, `GetNegRisk()`, `GetFeeRateBps()` (带缓存)
- [x] **最后成交价**: `GetLastTradePrice()`, `GetLastTradesPrices()`

### ✅ 订单管理
- [x] **订单提交**: `PostOrder()`, `PostOrders()`, `PostOrderWithOptions()`（支持 PostOnly）
- [x] **订单取消**: `Cancel()`, `CancelOrders()`, `CancelAll()`, `CancelMarketOrders()`
- [x] **订单查询**: `GetOrders()`, `GetOrder()`
- [x] **交易查询**: `GetTrades()`
- [x] **余额查询**: `GetBalanceAllowance()`
- [x] **通知管理**: `GetNotifications()`, `DropNotifications()`
- [x] **心跳**: `PostHeartbeat()` - 保持订单活跃（10秒无心跳自动取消订单）

### ✅ 订单构建和创建
- [x] 订单构建器完整实现（使用 go-order-utils）
- [x] 订单创建方法：`CreateOrder()`, `CreateMarketOrder()`, `CreateAndPostOrder()`
- [x] 市价计算：`CalculateMarketPrice()`
- [x] 舍入配置和金额计算

### ✅ 只读 API 密钥管理
- [x] `CreateReadonlyAPIKey()` - 创建只读 API 密钥
- [x] `GetReadonlyAPIKeys()` - 获取只读 API 密钥列表
- [x] `DeleteReadonlyAPIKey()` - 删除只读 API 密钥
- [x] `ValidateReadonlyAPIKey()` - 验证只读 API 密钥

### ✅ RFQ 客户端功能
- [x] `CreateRfqRequest()` - 创建 RFQ 请求
- [x] `CancelRfqRequest()` - 取消 RFQ 请求
- [x] `GetRfqRequests()` - 获取 RFQ 请求列表
- [x] `CreateRfqQuote()` - 创建 RFQ 报价
- [x] `CancelRfqQuote()` - 取消 RFQ 报价
- [x] `GetRfqRequesterQuotes()` - 获取针对自己请求的报价（请求方视角）
- [x] `GetRfqQuoterQuotes()` - 获取自己创建的报价（报价方视角）
- [x] `GetRfqBestQuote()` - 获取最佳 RFQ 报价
- [x] `AcceptQuote()` - 接受报价（请求方）
- [x] `ApproveOrder()` - 批准订单（报价方）
- [x] `GetRfqConfig()` - 获取 RFQ 配置

### ✅ Web3 客户端功能
- [x] `PolymarketWeb3Client` - 链上交易（支付 gas）
  - [x] 支持 EOA、PolyProxy 和 Safe 钱包
  - [x] 余额查询（POL、USDC、条件代币）
  - [x] 授权管理 (`SetAllApprovals()`)
  - [x] 头寸操作 (`SplitPosition()`, `MergePosition()`, `RedeemPosition()`, `RedeemPositions()`, `ConvertPositions()`)
  - [x] 代币转账 (`TransferUSDC()`, `TransferToken()`)
- [x] `PolymarketGaslessWeb3Client` - 无 gas 交易（通过中继器）
  - [x] 支持 PolyProxy 和 Safe 钱包
  - [x] 与 Web3Client 相同的操作，无需支付 gas
  - [x] **需要 Builder 凭证**（从 Polymarket 获取）
  - [x] **动态 Relay 地址**：自动从 `/relay-payload` 端点获取当前中继节点地址
  - [x] **批量赎回**: `RedeemPositions()` - 单次交易赎回多个 conditionId

### ✅ 其他功能
- [x] 订单评分：`IsOrderScoring()`, `AreOrdersScoring()`
- [x] 市场查询：`GetMarkets()`, `GetSimplifiedMarkets()`, `GetSamplingMarkets()`, `GetSamplingSimplifiedMarkets()`
- [x] 市场详情：`GetMarket()`, `GetMarketTradesEvents()`
- [x] 余额更新：`UpdateBalanceAllowance()`
- [x] 订单簿哈希：`GetOrderBookHash()`
- [x] Builder 交易：`GetBuilderTrades()`

## 功能对比

| 功能 | py-clob-client | Go SDK | 状态 |
|------|---------------|--------|------|
| 基础客户端 | ✅ | ✅ | 完成 |
| L0/L1/L2 认证 | ✅ | ✅ | 完成 |
| API 密钥管理 | ✅ | ✅ | 完成 |
| 市场数据查询 | ✅ | ✅ | 完成 |
| 订单簿查询 | ✅ | ✅ | 完成 |
| 订单提交/取消 | ✅ | ✅ | 完成 |
| 订单查询 | ✅ | ✅ | 完成 |
| 交易查询 | ✅ | ✅ | 完成 |
| 余额查询 | ✅ | ✅ | 完成 |
| 订单构建器 | ✅ | ✅ | 完整实现 |
| RFQ 功能 | ✅ | ✅ | 完整实现 |
| 只读 API 密钥 | ✅ | ✅ | 完整实现 |

## 环境变量

示例程序使用以下环境变量：

- `PRIVATE_KEY` (必需，除非设置了 `PRIVATE_KEY_ENC_FILE`): 您的以太坊私钥（十六进制格式）
- `CHAIN_ID` (可选): 链 ID（默认: 137，Polygon）
- `SIGNATURE_TYPE` (可选): 签名类型（0=EOA, 1=Magic/Email, 2=Browser，默认: 0）
- `FUNDER` (可选): 代理钱包的资金持有者地址
- `CLOB_HOST` (可选): CLOB API 主机（默认: https://clob.polymarket.com）
- `CLOB_API_KEY` (可选): L2 认证的 API 密钥
- `CLOB_SECRET` (可选): L2 认证的 API 密钥
- `CLOB_PASSPHRASE` (可选): L2 认证的 API 密钥
- `TOKEN_ID` (可选): 条件代币余额查询的 token ID
- `PRIVATE_KEY_ENC_FILE` (可选): 加密私钥文件路径（Fernet token），可替代 `PRIVATE_KEY`
- `CLOB_API_KEY_ENC_FILE` (可选): 加密的 `CLOB_API_KEY` 文件路径（Fernet token）
- `CLOB_SECRET_ENC_FILE` (可选): 加密的 `CLOB_SECRET` 文件路径（Fernet token）
- `CLOB_PASSPHRASE_ENC_FILE` (可选): 加密的 `CLOB_PASSPHRASE` 文件路径（Fernet token）
- `CLOB_*_ENC_FILE` (可选): 仅在需要复用固定的 L2 凭证时设置；否则可从 L1 自动派生。
- `BUILDER_API_KEY_ENC_FILE` / `BUILDER_API_SECRET_ENC_FILE` / `BUILDER_API_PASSPHRASE_ENC_FILE` (可选): 加密的 Builder 凭证文件（gasless）
- `POLY_ENC_KEY` (使用 *_ENC_FILE 时必需): Fernet key 或口令
- `POLY_ENC_SALT` (POLY_ENC_KEY 为口令时必需): PBKDF2 盐值
- `PRINT_API_CREDS` (可选): 设为 `1` 时输出新生成的 API 凭证
- `DEBUG_HEADERS` (可选): 设为 `1` 时输出认证头（已脱敏）
- `DEBUG_HEADERS_FULL` (可选): 设为 `1` 时输出完整认证头

## 更新日志

### v0.2.7 (2026-03-14)

#### Bug 修复

- **修复 Polygon Bor v2.6.0 `eth_call` 兼容性问题** - 使用原始 `eth_call` RPC 配合 `blockOverrides`（`baseFeePerGas: 0`）绕过 baseFee 校验
  - Bor v2.6.0（[公告](https://forum.polygon.technology/t/bor-v2-6-0-and-erigon-v3-4-0-for-mainnet-and-amoy/21757)）同步了上游 go-ethereum 的 `eth_call` 校验逻辑，节点现在会拒绝 `maxFeePerGas` 低于 `baseFee` 的调用
  - Bor 节点的 `setDefaults` 会注入冲突的 gas 字段，无法通过 `CallMsg` 的 gas 参数解决（设置 `GasFeeCap`/`GasTipCap` 导致 "both gasPrice and maxFeePerGas specified"；设置 `GasPrice` 会被忽略）
  - 修复方案：将所有 `ethclient.CallContract` 替换为原始 `rpc.Client.CallContext` 的 `eth_call`，通过第 4 个参数 `blockOverrides: {"baseFeePerGas": "0x0"}` 将区块上下文的 baseFee 设为 0，从而在 gas 校验前绕过限制
  - `estimateGas` 调用移除了 gas 字段，所有调用点均有 fallback gas limit
  - 仅影响只读合约调用，真实交易的 gas price 不受影响
  - 此问题影响所有 Polygon RPC 提供商，非特定提供商问题

### v0.2.6 (2026-01-28)

#### Bug 修复

- **修复 RPC 429 限流错误** - `waitForReceipt` 现在包含正确的轮询间隔和超时机制
  - 添加 4 秒轮询间隔
  - 添加 5 分钟超时限制
  - 同时适用于 `PolymarketWeb3Client` 和 `PolymarketGaslessWeb3Client`

### v0.2.5 (2026-01-24)

#### 新功能

- **统一批量赎回 (Unified Batch Redeem)** - 在有 Gas 和无 Gas 客户端中均支持单笔交易赎回多个 conditionId
  - 更新了 `RedeemPositions(requests []RedeemRequest)` 以支持 `PolyProxy` 链上单笔打包（此前为串行）
  - 统一了 `PolymarketWeb3Client` 和 `PolymarketGaslessWeb3Client` 之间的 API
  - 新增 `ExecuteBatch` 方法用于链上通用的多调用执行
  - 在 `examples/gasless_batch_redeem` 中新增自动发现并赎回的示例
  - 大幅降低了大批量头寸管理时的 Gas 成本和交易等待时间

### v0.2.3 (2026-01-24)

#### Bug 修复

- **Gasless Web3 客户端 - 修复签名验证失败问题**
  - 修复 `SignatureParams.relay` 使用 `/relay-payload` 端点返回的动态中继地址
  - 之前签名使用动态中继地址，而 `SignatureParams.relay` 使用静态配置地址，导致签名验证失败
  - 撤销了错误的 `to = ProxyFactoryAddress` 改动，该改动导致 `ProxyCall.To` 目标地址错误
  - 现在签名生成和请求参数都使用一致的动态中继地址

### v0.2.1 (2026-01-24)

#### 改进

- **Gasless Web3 客户端 - 动态 Relay 地址**
  - 新增 `getRelayPayload()` 方法，从 `/relay-payload` 端点获取动态中继节点地址
  - Polymarket 的中继服务可能会动态分配不同的中继器节点，现在代码会实时获取当前应该使用的 Relay 地址

## 参考

- [py-clob-client](https://github.com/Polymarket/py-clob-client) - Python 官方实现
- [Polymarket CLOB 文档](https://docs.polymarket.com/developers/CLOB)
- [go-order-utils](https://github.com/polymarket/go-order-utils) - 订单构建工具库

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎贡献！请随时提交 Pull Request。
