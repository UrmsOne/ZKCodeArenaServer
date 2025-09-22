#!/bin/bash

# ZK Code Arena Server 启动脚本

set -e

echo "🚀 启动 ZK Code Arena Server..."

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 检查环境变量
if [ -f .env ]; then
    echo "📄 加载环境变量..."
    export $(cat .env | xargs)
fi

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p logs
mkdir -p deploy/mongo/init
mkdir -p deploy/nginx/conf.d

# 构建并启动服务
echo "🔨 构建并启动服务..."
if [ "$1" = "dev" ]; then
    echo "🔧 开发模式启动..."
    docker-compose -f docker-compose.dev.yml up --build
else
    echo "🏭 生产模式启动..."
    docker-compose up --build -d
fi

echo "✅ ZK Code Arena Server 启动完成！"
echo "🌐 访问地址: http://localhost"
echo "📊 API 文档: http://localhost/api/v1"
echo "💊 健康检查: http://localhost/health"
