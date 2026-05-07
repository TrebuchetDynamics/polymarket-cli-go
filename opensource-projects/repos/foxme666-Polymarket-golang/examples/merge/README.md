# Merge Position 示例（支付 Gas）

将 Yes + No 条件代币对合并回 USDC。

## 什么是 Merge？

Merge 是 Split 的逆操作，将两个互补的条件代币合并回 USDC：
- 1 个 Yes 代币 + 1 个 No 代币 = 1 USDC

例如：合并 10 对代币将获得 10 USDC。

## 使用场景

1. **退出头寸**：同时持有 Yes 和 No 代币时，可以合并退出
2. **套利**：当 Yes + No 价格总和 < 1 时，可以买入两方后合并获利
3. **减少持仓**：不想等待市场结算时，可以合并提前退出

## 环境变量

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `PRIVATE_KEY` | ✅ | - | 钱包私钥 |
| `CONDITION_ID` | ✅ | - | 市场的 Condition ID |
| `AMOUNT` | ❌ | 10.0 | 合并金额（代币对数量） |
| `CHAIN_ID` | ❌ | 137 | 链 ID（137=Polygon, 80002=Amoy） |
| `RPC_URL` | ❌ | polygon-rpc.com | RPC 节点地址 |
| `SIGNATURE_TYPE` | ❌ | 0 | 签名类型（0=EOA, 1=PolyProxy, 2=Safe） |
| `NEG_RISK` | ❌ | false | 是否为 NegRisk 市场 |

## 运行

```bash
# 基本用法
PRIVATE_KEY=0x... CONDITION_ID=0x... ./run.sh

# 指定金额
PRIVATE_KEY=0x... CONDITION_ID=0x... AMOUNT=50 ./run.sh

# NegRisk 市场
PRIVATE_KEY=0x... CONDITION_ID=0x... NEG_RISK=true ./run.sh
```

## 注意事项

1. **需要 POL**：此操作需要支付 Gas 费用
2. **需要代币对**：必须同时持有相等数量的 Yes 和 No 代币
3. **授权**：首次使用可能需要先授权合约

## Merge vs Redeem

| 操作 | 时机 | 条件 | 收益 |
|------|------|------|------|
| **Merge** | 市场结算前 | 需要 Yes + No | 1 对 = 1 USDC |
| **Redeem** | 市场结算后 | 只需获胜方代币 | 1 个获胜代币 = 1 USDC |

- Merge：不等结果，提前退出
- Redeem：等待结果，获胜方获得全部价值
