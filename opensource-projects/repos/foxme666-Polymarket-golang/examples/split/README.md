# Split Position 示例（支付 Gas）

将 USDC 拆分为 Yes + No 条件代币对。

## 什么是 Split？

Split 操作将 USDC 抵押品拆分为两个互补的条件代币：
- **Yes 代币**：如果市场结果为 Yes，可兑换 1 USDC
- **No 代币**：如果市场结果为 No，可兑换 1 USDC

例如：拆分 10 USDC 将获得 10 个 Yes 代币和 10 个 No 代币。

## 使用场景

1. **做市**：拆分后可以在订单簿上同时挂买卖单
2. **套利**：当 Yes + No 价格总和 > 1 时，可以拆分后卖出获利
3. **对冲**：持有一方头寸时，拆分可以获得另一方代币进行对冲

## 环境变量

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `PRIVATE_KEY` | ✅ | - | 钱包私钥 |
| `CONDITION_ID` | ✅ | - | 市场的 Condition ID |
| `AMOUNT` | ❌ | 10.0 | 拆分金额（USDC） |
| `CHAIN_ID` | ❌ | 137 | 链 ID（137=Polygon, 80002=Amoy） |
| `RPC_URL` | ❌ | polygon-rpc.com | RPC 节点地址 |
| `SIGNATURE_TYPE` | ❌ | 0 | 签名类型（0=EOA, 1=PolyProxy, 2=Safe） |
| `NEG_RISK` | ❌ | false | 是否为 NegRisk 市场 |

## 运行

```bash
# 基本用法
PRIVATE_KEY=0x... CONDITION_ID=0x... ./run.sh

# 指定金额
PRIVATE_KEY=0x... CONDITION_ID=0x... AMOUNT=100 ./run.sh

# 使用自定义 RPC
PRIVATE_KEY=0x... CONDITION_ID=0x... RPC_URL=https://your-rpc.com ./run.sh
```

## 获取 Condition ID

可以通过 Polymarket API 或网页获取市场的 Condition ID：

```bash
# 从市场 API 获取
curl "https://clob.polymarket.com/markets" | jq '.[] | {question, condition_id}'
```

## 注意事项

1. **需要 POL**：此操作需要支付 Gas 费用，确保钱包有足够的 POL
2. **需要 USDC**：确保钱包有足够的 USDC 进行拆分
3. **授权**：首次使用可能需要先授权 ConditionalTokens 合约使用 USDC

## 与 Gasless 版本的区别

| 特性 | 支付 Gas 版本 | Gasless 版本 |
|------|--------------|--------------|
| Gas 费用 | 用户支付 POL | 中继器支付 |
| 速度 | 更快（直接上链） | 稍慢（经过中继） |
| 依赖 | 只需 RPC | 需要 Builder 凭证 |
| 适用场景 | 有 POL 的用户 | 无 POL 的用户 |
