# Gasless Redeem Position 示例

这个示例演示如何使用 `PolymarketGaslessWeb3Client` 进行无 Gas 的头寸赎回操作。

## 什么是赎回（Redeem）？

赎回是在市场**结算后**将获胜的条件代币兑换为 USDC 的操作：
- 如果您持有获胜结果的代币，每个代币可兑换 $1
- 如果您持有失败结果的代币，价值为 $0

**重要**：只有在市场结算后才能赎回。未结算的市场需要使用合并（Merge）操作。

## 环境变量

| 变量 | 必需 | 说明 | 默认值 |
|------|------|------|--------|
| `PRIVATE_KEY` | ✅ | 私钥（带或不带 0x 前缀） | - |
| `CONDITION_ID` | ✅ | 已结算市场的 Condition ID | - |
| `BUILDER_API_KEY` | ✅ | Builder API Key（从 Polymarket 获取） | - |
| `BUILDER_API_SECRET` | ✅ | Builder API Secret | - |
| `BUILDER_API_PASSPHRASE` | ✅ | Builder API Passphrase | - |
| `CHAIN_ID` | ❌ | 链 ID | `137` (Polygon) |
| `RPC_URL` | ❌ | 自定义 RPC URL | 默认 Polygon RPC |
| `SIGNATURE_TYPE` | ❌ | 签名类型 (1=PolyProxy, 2=Safe) | `1` |
| `AMOUNTS` | ❌ | 赎回数量（Yes,No） | `10.0,10.0` |
| `NEG_RISK` | ❌ | 是否为 NegRisk 市场 | `true` |

### Amounts 说明

`AMOUNTS` 参数是逗号分隔的两个数字，分别表示要赎回的 Yes 和 No 代币数量：
- 格式：`<Yes数量>,<No数量>`
- 示例：`10.0,10.0` = 赎回 10 个 Yes 代币和 10 个 No 代币
- 示例：`50.0,0` = 只赎回 50 个 Yes 代币
- 示例：`0,100.0` = 只赎回 100 个 No 代币

## 运行示例

### 方法 1: 使用环境变量

```bash
export PRIVATE_KEY="your-private-key"
export CONDITION_ID="0x..."  # 已结算市场的 condition ID

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
PRIVATE_KEY="0x..." CONDITION_ID="0x..." \
  BUILDER_API_KEY="..." BUILDER_API_SECRET="..." BUILDER_API_PASSPHRASE="..." \
  go run main.go
```

## 输出示例

```
=== Polymarket Gasless Redeem 示例 ===
Chain ID: 137
Signature Type: 1 (PolyProxy)
Condition ID: 0x...
Amounts: [10 10] (Yes, No 代币数量)
NegRisk: true

正在创建 Gasless Web3 客户端...
✓ 客户端创建成功
  Base Address: 0x...
  Proxy Address: 0x...

=== 当前余额 ===
POL Balance: 0.0000
USDC Balance: 0.000000

=== 执行赎回 ===
正在赎回 10.000000 Yes 和 10.000000 No 代币...
(通过无 Gas 中继器提交交易)
Gasless txn submitted: 0x...
Transaction ID: ...
State: submitted

Redeem Position succeeded

=== 交易结果 ===
Transaction Hash: 0x...
Block Number: 12345678
Gas Used: 120000
Status: ✓ 成功

=== 赎回后余额 ===
USDC Balance: 50.000000 (增加 50.000000)

=== 完成 ===
```

## 赎回逻辑

假设您持有一个已结算市场的代币：

| 场景 | Yes 代币 | No 代币 | 结果 | 获得 USDC |
|------|---------|---------|------|----------|
| Yes 获胜 | 100 | 0 | Yes | $100 |
| Yes 获胜 | 0 | 100 | Yes | $0 |
| No 获胜 | 100 | 0 | No | $0 |
| No 获胜 | 0 | 100 | No | $100 |
| Yes 获胜 | 50 | 50 | Yes | $50 |

## 如何检查市场是否已结算

```go
// 使用 CLOB API 获取市场信息
client, _ := polymarket.NewClobClient(...)
market, _ := client.GetMarket(conditionID)

// 检查市场状态
if market["closed"].(bool) && market["resolved"].(bool) {
    fmt.Println("市场已结算，可以赎回")
}
```

## 注意事项

1. **市场必须已结算**：只有结算后的市场才能赎回
2. **Builder 凭证必需**：Gasless 交易需要 Builder API 凭证
3. **失败代币无价值**：只有获胜结果的代币可以兑换 USDC
4. **钱包类型**：Gasless 只支持 PolyProxy 和 Safe 钱包
5. **配额限制**：中继器有每日配额限制

## 相关操作

| 操作 | 方法 | 说明 | 适用场景 |
|------|------|------|---------|
| 拆分 | `SplitPosition()` | USDC → Yes + No | 进入市场 |
| 合并 | `MergePosition()` | Yes + No → USDC | 退出未结算市场 |
| 赎回 | `RedeemPosition()` | 获胜代币 → USDC | 结算后领取收益 |
| 转换 | `ConvertPositions()` | NegRisk No → Yes + USDC | 特殊转换 |

## 常见问题

### Q: 市场未结算可以赎回吗？
A: 不可以。未结算市场请使用 `MergePosition()` 将代币对合并回 USDC。

### Q: 我持有失败结果的代币怎么办？
A: 失败结果的代币价值为 $0，赎回不会获得任何 USDC，但会清除您的代币余额。

### Q: 如何知道哪个结果获胜？
A: 通过 Polymarket API 或网站查看市场的结算结果。

