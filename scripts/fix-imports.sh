#!/bin/bash

# 修复 Go 导入问题的脚本

set -e

echo "🔧 修复 Go 导入问题..."

# 检查是否安装了 goimports
if ! command -v goimports &> /dev/null; then
    echo "📦 安装 goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
fi

# 修复所有 Go 文件的导入
echo "🔍 检查并修复导入..."
find . -name "*.go" -not -path "./vendor/*" -not -path "./tmp/*" | while read -r file; do
    echo "处理文件: $file"
    goimports -w "$file"
done

echo "✅ 导入修复完成！"

# 检查是否还有编译错误
echo "🔍 检查编译错误..."
if go build ./... 2>&1 | grep -q "imported and not used"; then
    echo "⚠️  仍有未使用的导入，请手动检查："
    go build ./... 2>&1 | grep "imported and not used"
else
    echo "✅ 没有发现未使用的导入！"
fi
