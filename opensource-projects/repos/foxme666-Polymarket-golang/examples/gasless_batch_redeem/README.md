# Gasless Batch Redeem 示例

自动从 Polymarket API 获取可赎回头寸并批量赎回。**单次 gasless 交易完成所有赎回操作。**

## 功能

- 🔍 自动查询 Polymarket Data API 获取所有可赎回头寸
- 📊 按 conditionId 分组并显示详情（标题、outcome、size、value）
- ⚡ 单次 gasless 交易批量赎回所有头寸
- 🔒 支持仅检查模式（不执行赎回）

## 用法

```bash
# 设置环境变量
export PRIVATE_KEY="0x..."
export BUILDER_API_KEY="..."
export BUILDER_API_SECRET="..."
export BUILDER_API_PASSPHRASE="..."

# 仅检查可赎回头寸（不执行赎回）
CHECK_ONLY=true ./run.sh

# 执行批量赎回
./run.sh
```

## 环境变量

| 变量 | 说明 | 必需 |
|------|------|------|
| PRIVATE_KEY | 私钥 | ✓ |
| BUILDER_API_KEY | Builder API Key | 赎回时必需 |
| BUILDER_API_SECRET | Builder API Secret | 赎回时必需 |
| BUILDER_API_PASSPHRASE | Builder API Passphrase | 赎回时必需 |
| CHECK_ONLY | 设为 `true` 仅检查 | ✗ |
| CHAIN_ID | 链 ID (默认 137) | ✗ |
| RPC_URL | 自定义 RPC URL | ✗ |

## 获取 Builder 凭证

1. 访问 https://polymarket.com/builder
2. 创建或登录 Builder 账户
3. 生成 API Key

## 示例输出

```
=== Polymarket Auto Batch Redeem ===
Chain ID: 137
模式: 执行赎回

正在创建客户端...
Proxy Wallet: 0xA94f7ad2De9cf1FA03dAD8e88876331519cCBa4A

正在获取可赎回头寸...

找到 84 个可赎回头寸:

1. Bitcoin Up or Down - January 24, 4:00AM-4:15AM ET
   Up: Size 10.0000, Value $10.0000 [WIN]
   ConditionID: 0x4bb0cfd0b903a...
   NegRisk: false, Total Value: $10.0000
...

总可赎回价值: $19.98

=== 执行批量赎回 ===
正在批量赎回 84 个头寸...
Gasless batch txn submitted: 0xf567f79c8cc3...
Transaction ID: 019beea4-8217-7f1f-96c9-...
Status: ✓ 成功

=== 赎回后余额 ===
USDC Balance: 184.76 (增加 19.98)
```
