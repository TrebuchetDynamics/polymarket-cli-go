# Redeem Position 示例（支付 Gas）

市场结算后，将获胜代币赎回为 USDC。

## 什么是 Redeem？

Redeem 操作在市场结算后，将获胜方的代币兑换为 USDC：
- 如果结果为 **Yes**：每个 Yes 代币可兑换 1 USDC，No 代币价值归零
- 如果结果为 **No**：每个 No 代币可兑换 1 USDC，Yes 代币价值归零

## 使用场景

市场结算后，持有获胜方代币的用户需要调用 Redeem 来领取 USDC。

## 环境变量

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `PRIVATE_KEY` | ✅ | - | 钱包私钥 |
| `CONDITION_ID` | ✅ | - | 市场的 Condition ID |
| `AMOUNTS` | ❌ | - | 赎回金额（逗号分隔，如 "10.0,0"） |
| `CHAIN_ID` | ❌ | 137 | 链 ID（137=Polygon, 80002=Amoy） |
| `RPC_URL` | ❌ | polygon-rpc.com | RPC 节点地址 |
| `SIGNATURE_TYPE` | ❌ | 0 | 签名类型（0=EOA, 1=PolyProxy, 2=Safe） |
| `NEG_RISK` | ❌ | false | 是否为 NegRisk 市场 |

## 运行

```bash
# 基本用法（自动赎回所有可赎回代币）
PRIVATE_KEY=0x... CONDITION_ID=0x... ./run.sh

# NegRisk 市场，指定赎回金额
PRIVATE_KEY=0x... CONDITION_ID=0x... NEG_RISK=true AMOUNTS="100.0,0" ./run.sh
```

## AMOUNTS 参数说明

对于 NegRisk 市场，`AMOUNTS` 是一个逗号分隔的数组：
- 格式：`"YesAmount,NoAmount"`
- 例如：`"10.0,0"` 表示赎回 10 个 Yes 代币和 0 个 No 代币

对于非 NegRisk 市场，此参数可以省略，合约会自动处理。

## 注意事项

1. **市场必须已结算**：只有市场结算后才能 Redeem
2. **只有获胜方有价值**：失败方代币价值归零，无法赎回
3. **需要 POL**：此操作需要支付 Gas 费用

## 如何检查市场是否已结算

```go
// 通过 API 查询市场状态
market, _ := client.GetMarket(conditionID)
if market.Resolved {
    fmt.Println("市场已结算，获胜方:", market.Outcome)
}
```

## 完整流程示例

```
1. Split: 10 USDC → 10 Yes + 10 No
2. 交易: 卖出 10 No → 获得 USDC
3. 持有: 10 Yes
4. 等待市场结算...
5. 结果为 Yes
6. Redeem: 10 Yes → 10 USDC

净利润 = 10 USDC (Redeem) + 卖出 No 的收入 - 10 USDC (Split) - Gas
```
