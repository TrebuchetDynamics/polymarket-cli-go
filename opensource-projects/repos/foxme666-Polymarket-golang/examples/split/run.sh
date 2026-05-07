#!/bin/bash

# Polymarket Split Position 示例（支付 Gas）
# 将 USDC 拆分为 Yes + No 代币对

# 必需的环境变量
export PRIVATE_KEY="${PRIVATE_KEY:-}"           # 私钥（必需）
export PRIVATE_KEY_ENC_FILE="${PRIVATE_KEY_ENC_FILE:-}" # 加密私钥文件路径（可选）
export CONDITION_ID="${CONDITION_ID:-}"         # 市场 Condition ID（必需）

# 可选的环境变量
export CHAIN_ID="${CHAIN_ID:-137}"              # 链 ID（137=Polygon, 80002=Amoy 测试网）
export RPC_URL="${RPC_URL:-https://polygon-rpc.com}"  # RPC 节点地址
export SIGNATURE_TYPE="${SIGNATURE_TYPE:-0}"    # 0=EOA, 1=PolyProxy, 2=Safe
export AMOUNT="${AMOUNT:-10.0}"                 # 拆分金额（USDC）
export NEG_RISK="${NEG_RISK:-false}"            # 是否为 NegRisk 市场

# 检查必需的环境变量
if [ -z "$PRIVATE_KEY" ] && [ -z "$PRIVATE_KEY_ENC_FILE" ]; then
    echo "错误: 必须设置 PRIVATE_KEY 或 PRIVATE_KEY_ENC_FILE"
    echo "用法: PRIVATE_KEY=0x... CONDITION_ID=0x... ./run.sh"
    echo "或:  PRIVATE_KEY_ENC_FILE=/secrets/private_key.enc CONDITION_ID=0x... ./run.sh"
    exit 1
fi

if [ -z "$CONDITION_ID" ]; then
    echo "错误: 必须设置 CONDITION_ID 环境变量"
    echo "用法: PRIVATE_KEY=0x... CONDITION_ID=0x... ./run.sh"
    exit 1
fi

# 切换到脚本所在目录
cd "$(dirname "$0")"

# 运行示例
go run main.go
