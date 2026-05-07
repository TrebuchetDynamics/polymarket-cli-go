#!/bin/bash

# Polymarket Auto Batch Redeem 示例运行脚本
# 自动获取可赎回头寸并批量赎回

# 基本配置 - 请设置环境变量
export PRIVATE_KEY="${PRIVATE_KEY:-}"
export CHAIN_ID="${CHAIN_ID:-137}"
export RPC_URL="${RPC_URL:-}"

# 设置为 true 只检查不赎回
export CHECK_ONLY="${CHECK_ONLY:-false}"

# Builder 凭证（必需，从 https://polymarket.com/builder 获取）
export BUILDER_API_KEY="${BUILDER_API_KEY:-}"
export BUILDER_API_SECRET="${BUILDER_API_SECRET:-}"
export BUILDER_API_PASSPHRASE="${BUILDER_API_PASSPHRASE:-}"

# 检查必要的环境变量
if [ -z "$PRIVATE_KEY" ]; then
    echo "错误: 必须设置 PRIVATE_KEY 环境变量"
    exit 1
fi

if [ "$CHECK_ONLY" != "true" ]; then
    if [ -z "$BUILDER_API_KEY" ] || [ -z "$BUILDER_API_SECRET" ] || [ -z "$BUILDER_API_PASSPHRASE" ]; then
        echo "错误: 执行赎回需要 Builder 凭证"
        echo "设置 CHECK_ONLY=true 可以仅检查可赎回头寸"
        exit 1
    fi
fi

# 运行程序
go run main.go
