#!/bin/bash
# 钱包余额查询脚本

# 检查是否设置了私钥
if [ -z "$PRIVATE_KEY" ] && [ -z "$PRIVATE_KEY_ENC_FILE" ]; then
    echo "错误: 请设置环境变量 PRIVATE_KEY 或 PRIVATE_KEY_ENC_FILE"
    echo "使用方法:"
    echo "  export PRIVATE_KEY=\"your-private-key-hex\""
    echo "  # 或者:"
    echo "  export PRIVATE_KEY_ENC_FILE=\"/secrets/private_key.enc\""
    echo "  ./run.sh"
    exit 1
fi

# 运行程序
go run main.go
