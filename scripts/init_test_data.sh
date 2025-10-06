#!/bin/bash

# 测试数据初始化脚本执行器
# 
# 功能: 连接到 MongoDB 并执行测试数据初始化
# 使用方法: ./scripts/init_test_data.sh

set -e  # 遇到错误立即退出

# 配置
MONGO_URI="mongodb://42.194.245.236:27017"
DB_NAME="zk_code_arena"
SCRIPT_PATH="scripts/init_test_data.js"

echo "========================================="
echo "测试数据初始化工具"
echo "========================================="
echo ""
echo "MongoDB URI: $MONGO_URI"
echo "数据库名称: $DB_NAME"
echo "脚本路径: $SCRIPT_PATH"
echo ""

# 检查 mongosh 是否安装
if ! command -v mongosh &> /dev/null; then
    echo "❌ 错误: 未找到 mongosh 命令"
    echo "请先安装 MongoDB Shell:"
    echo "  https://www.mongodb.com/try/download/shell"
    exit 1
fi

# 检查脚本文件是否存在
if [ ! -f "$SCRIPT_PATH" ]; then
    echo "❌ 错误: 未找到脚本文件 $SCRIPT_PATH"
    exit 1
fi

# 测试 MongoDB 连接
echo "📡 测试 MongoDB 连接..."
if ! mongosh "$MONGO_URI/$DB_NAME" --quiet --eval "db.runCommand({ ping: 1 })" > /dev/null 2>&1; then
    echo "❌ 错误: 无法连接到 MongoDB"
    echo "请检查:"
    echo "  1. MongoDB 服务是否运行"
    echo "  2. 网络连接是否正常"
    echo "  3. MongoDB URI 是否正确"
    exit 1
fi
echo "✅ MongoDB 连接成功"
echo ""

# 执行初始化脚本
echo "🚀 开始执行初始化脚本..."
echo "========================================="
mongosh "$MONGO_URI/$DB_NAME" < "$SCRIPT_PATH"
echo "========================================="
echo ""

# 完成
echo "✅ 测试数据初始化完成!"
echo ""
echo "📖 下一步:"
echo "  1. 启动应用服务: go run cmd/*.go run"
echo "  2. 运行集成测试: ./scripts/test_judge_flow.sh"
echo "  3. 查看手动验证指南: docs/判题系统验证/MANUAL_TEST_GUIDE.md"
echo ""
