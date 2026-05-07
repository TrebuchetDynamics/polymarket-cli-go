#!/bin/bash

# Polymarket 下单示例运行脚本

# 基本配置
export PRIVATE_KEY="${PRIVATE_KEY:-}"
export PRIVATE_KEY_ENC_FILE="${PRIVATE_KEY_ENC_FILE:-}"
export CHAIN_ID="${CHAIN_ID:-137}"
export CLOB_HOST="${CLOB_HOST:-https://clob.polymarket.com}"

# 账户类型和代理地址
export SIGNATURE_TYPE="${SIGNATURE_TYPE:-0}"  # 0=EOA, 1=Magic/Email, 2=Browser
export FUNDER="${FUNDER:-}"  # 代理钱包地址（可选）

# API凭证（可选，如果已有）
export CLOB_API_KEY="${CLOB_API_KEY:-}"
export CLOB_SECRET="${CLOB_SECRET:-}"
export CLOB_PASSPHRASE="${CLOB_PASSPHRASE:-}"
export CLOB_API_KEY_ENC_FILE="${CLOB_API_KEY_ENC_FILE:-}"
export CLOB_SECRET_ENC_FILE="${CLOB_SECRET_ENC_FILE:-}"
export CLOB_PASSPHRASE_ENC_FILE="${CLOB_PASSPHRASE_ENC_FILE:-}"

# 订单参数
export TOKEN_ID="${TOKEN_ID:-}"
export ORDER_SIDE="${ORDER_SIDE:-BUY}"
export PRICE="${PRICE:-0.5}"
export SIZE="${SIZE:-10.0}"
export ORDER_TYPE="${ORDER_TYPE:-LIMIT}"  # LIMIT 或 MARKET

# 运行程序
go run main.go
