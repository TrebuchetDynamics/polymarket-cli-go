#!/bin/bash

# Polymarket Gasless Split 示例运行脚本

# 基本配置
export PRIVATE_KEY="${PRIVATE_KEY:-}"
export PRIVATE_KEY_ENC_FILE="${PRIVATE_KEY_ENC_FILE:-}"
export CHAIN_ID="${CHAIN_ID:-137}"
export RPC_URL="${RPC_URL:-}"  # 空=使用默认 Polygon RPC

# 签名类型: 1=PolyProxy, 2=Safe
# 注意: Gasless 模式不支持 EOA (0)
export SIGNATURE_TYPE="${SIGNATURE_TYPE:-1}"

# 拆分参数
export CONDITION_ID="${CONDITION_ID:-}"  # 必须设置
export AMOUNT="${AMOUNT:-10.0}"          # 拆分金额（USDC）
export NEG_RISK="${NEG_RISK:-true}"      # 是否为 NegRisk 市场

# Builder 凭证（必需，从 https://polymarket.com/builder 获取）
export BUILDER_API_KEY="${BUILDER_API_KEY:-}"
export BUILDER_API_SECRET="${BUILDER_API_SECRET:-}"
export BUILDER_API_PASSPHRASE="${BUILDER_API_PASSPHRASE:-}"
export BUILDER_API_KEY_ENC_FILE="${BUILDER_API_KEY_ENC_FILE:-}"
export BUILDER_API_SECRET_ENC_FILE="${BUILDER_API_SECRET_ENC_FILE:-}"
export BUILDER_API_PASSPHRASE_ENC_FILE="${BUILDER_API_PASSPHRASE_ENC_FILE:-}"

# 运行程序
go run main.go
