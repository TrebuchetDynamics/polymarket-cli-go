# Gasless Merge Position 示例

这个示例演示如何使用 `PolymarketGaslessWeb3Client` 进行无 Gas 的头寸合并操作。

## 什么是合并（Merge）？

合并是拆分的逆操作：
- **拆分 (Split)**：USDC → Yes 代币 + No 代币
- **合并 (Merge)**：Yes 代币 + No 代币 → USDC

当您同时持有一对市场的 Yes 和 No 代币时，可以将它们合并回 USDC。每对代币可兑换 1 USDC。

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
| `AMOUNT` | ❌ | 合并数量（代币对数量） | `10.0` |
| `NEG_RISK` | ❌ | 是否为 NegRisk 市场 | `true` |

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
=== Polymarket Gasless Merge 示例 ===
Chain ID: 137
Signature Type: 1 (PolyProxy)
Condition ID: 0x...
Amount: 10.000000 代币
NegRisk: true

正在创建 Gasless Web3 客户端...
✓ 客户端创建成功
  Base Address: 0x...
  Proxy Address: 0x...

=== 当前余额 ===
POL Balance: 0.0000
USDC Balance: 90.000000

=== 执行合并 ===
正在合并 10.000000 代币对为 USDC...
(通过无 Gas 中继器提交交易)
Gasless txn submitted: 0x...
Transaction ID: ...
State: submitted

Merge Position succeeded

=== 交易结果 ===
Transaction Hash: 0x...
Block Number: 12345678
Gas Used: 150000
Status: ✓ 成功

=== 合并后余额 ===
USDC Balance: 100.000000 (增加 10.000000)

=== 完成 ===
```

## 使用场景

1. **退出市场**：当您不再想持有某个市场的头寸时
2. **套利**：如果 Yes + No 价格 < $1，可以买入两者然后合并获利
3. **资金回收**：将闲置的代币对转换回 USDC

## 注意事项

1. **Builder 凭证必需**：Gasless 交易需要 Builder API 凭证
2. **代币配对**：您必须同时持有等量的 Yes 和 No 代币才能合并
3. **钱包类型**：Gasless 只支持 PolyProxy 和 Safe 钱包
4. **配额限制**：中继器有每日配额限制

## 相关操作

| 操作 | 方法 | 说明 |
|------|------|------|
| 拆分 | `SplitPosition()` | USDC → Yes + No |
| 合并 | `MergePosition()` | Yes + No → USDC |
| 赎回 | `RedeemPosition()` | 市场结算后赎回 |
| 转换 | `ConvertPositions()` | NegRisk No → Yes + USDC |

