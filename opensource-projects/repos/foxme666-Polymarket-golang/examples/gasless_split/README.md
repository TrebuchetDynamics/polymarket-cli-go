# Gasless Split Position 示例

这个示例演示如何使用 `PolymarketGaslessWeb3Client` 进行无 Gas 的头寸拆分操作。

## 什么是 Gasless 交易？

Gasless 交易通过 Polymarket 的中继器提交，用户不需要支付 POL (gas) 费用。这对于：
- 新用户（没有 POL）非常友好
- 减少交易成本
- 简化用户体验

**注意**：Gasless 模式仅支持 PolyProxy (1) 和 Safe (2) 钱包，不支持 EOA (0)。

## 环境变量

| 变量 | 必需 | 说明 | 默认值 |
|------|------|------|--------|
| `PRIVATE_KEY` | ✅ | 私钥（带或不带 0x 前缀） | - |
| `CONDITION_ID` | ✅ | 市场的 Condition ID | - |
| `BUILDER_API_KEY` | ✅ | Builder API Key（从 Polymarket 获取） | - |
| `BUILDER_API_SECRET` | ✅ | Builder API Secret | - |
| `BUILDER_API_PASSPHRASE` | ✅ | Builder API Passphrase | - |
| `CHAIN_ID` | ❌ | 链 ID | `137` (Polygon) |
| `RPC_URL` | ❌ | 自定义 RPC URL | 默认 Polygon RPC |
| `SIGNATURE_TYPE` | ❌ | 签名类型 (1=PolyProxy, 2=Safe) | `1` |
| `AMOUNT` | ❌ | 拆分金额 (USDC) | `10.0` |
| `NEG_RISK` | ❌ | 是否为 NegRisk 市场 | `true` |

## 如何获取 Condition ID

1. 访问 Polymarket 网站找到你感兴趣的市场
2. 使用 CLOB API 获取市场信息：

```go
client, _ := polymarket.NewClobClient(...)
markets, _ := client.GetMarkets(nil, "")
// 从市场信息中获取 condition_id
```

或者使用 Gamma API：
```
GET https://gamma-api.polymarket.com/markets?_limit=10
```

## 运行示例

### 方法 1: 使用环境变量

```bash
export PRIVATE_KEY="your-private-key"
export CONDITION_ID="0x..."
export AMOUNT="10.0"

# Builder 凭证（必需）
export BUILDER_API_KEY="your-builder-api-key"
export BUILDER_API_SECRET="your-builder-api-secret"
export BUILDER_API_PASSPHRASE="your-builder-passphrase"

go run main.go
```

### 方法 2: 使用运行脚本

```bash
chmod +x run.sh

# 编辑 run.sh 设置参数
./run.sh
```

### 方法 3: 一行命令

```bash
PRIVATE_KEY="0x..." CONDITION_ID="0x..." AMOUNT="10.0" \
  BUILDER_API_KEY="..." BUILDER_API_SECRET="..." BUILDER_API_PASSPHRASE="..." \
  go run main.go
```

## 输出示例

```
=== Polymarket Gasless Split 示例 ===
Chain ID: 137
Signature Type: 1 (PolyProxy)
Condition ID: 0x...
Amount: 10.000000 USDC
NegRisk: true

正在创建 Gasless Web3 客户端...
✓ 客户端创建成功
  Base Address: 0x...
  Proxy Address: 0x...

=== 当前余额 ===
POL Balance: 0.0000
USDC Balance: 100.000000

=== 执行拆分 ===
正在将 10.000000 USDC 拆分为头寸...
(通过无 Gas 中继器提交交易)
Gasless txn submitted: 0x...
Transaction ID: ...
State: submitted

Split Position succeeded

=== 交易结果 ===
Transaction Hash: 0x...
Block Number: 12345678
Gas Used: 150000
Status: ✓ 成功

=== 拆分后余额 ===
USDC Balance: 90.000000 (减少 10.000000)

=== 完成 ===
```

## 拆分后的头寸

拆分操作将 USDC 转换为两个互补的条件代币：
- **Yes 代币**：如果结果为 Yes，价值 $1
- **No 代币**：如果结果为 No，价值 $1

例如，拆分 10 USDC 将获得：
- 10 个 Yes 代币
- 10 个 No 代币

## 相关操作

| 操作 | 方法 | 说明 |
|------|------|------|
| 拆分 | `SplitPosition()` | USDC → Yes + No |
| 合并 | `MergePosition()` | Yes + No → USDC |
| 赎回 | `RedeemPosition()` | 市场结算后赎回 |
| 转换 | `ConvertPositions()` | NegRisk No → Yes + USDC |

## 注意事项

1. **Builder 凭证必需**：Gasless 交易需要 Builder API 凭证，从 [Polymarket Builder](https://polymarket.com/builder) 获取
2. **钱包类型**：Gasless 只支持 PolyProxy 和 Safe 钱包，不支持 EOA
3. **授权**：首次操作前需要设置代币授权（可能需要支付 gas）
4. **中继延迟**：Gasless 交易可能比直接交易稍慢
5. **配额限制**：中继器有每日配额限制，超出配额后需等待重置

