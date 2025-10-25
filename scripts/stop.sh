#!/bin/bash

# ZK Code Arena Server 停止脚本

set -e

echo "🛑 停止 ZK Code Arena Server..."

# 停止服务
if [ "$1" = "dev" ]; then
    echo "🔧 停止开发环境..."
    docker-compose -f docker-compose.dev.yml down
else
    echo "🏭 停止生产环境..."
    docker-compose down
fi

echo "✅ ZK Code Arena Server 已停止"
