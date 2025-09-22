#!/bin/bash

# ZK Code Arena Server 开发环境启动脚本

set -e

echo "🔧 启动 ZK Code Arena Server 开发环境..."

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.23+"
    exit 1
fi

# 检查 Air 是否安装
if ! command -v air &> /dev/null; then
    echo "📦 安装 Air 热更新工具..."
    go install github.com/cosmtrek/air@latest
fi

# 检查 MongoDB 是否运行
if ! docker ps | grep -q zk-arena-mongo-dev; then
    echo "🐳 启动 MongoDB..."
    docker run -d --name zk-arena-mongo-dev -p 27017:27017 mongo:7.0
fi

# 检查 Redis 是否运行
if ! docker ps | grep -q zk-arena-redis-dev; then
    echo "🐳 启动 Redis..."
    docker run -d --name zk-arena-redis-dev -p 6379:6379 redis:7.2-alpine
fi

# 检查 go-judge 是否运行
if ! docker ps | grep -q zk-arena-judge-dev; then
    echo "🐳 启动 go-judge..."
    docker run -d --name zk-arena-judge-dev -p 5050:5050 --privileged criyle/go-judge:latest
fi

# 下载依赖
echo "📦 下载 Go 依赖..."
go mod download

# 启动 Air 热更新
echo "🔥 启动 Air 热更新..."
air
