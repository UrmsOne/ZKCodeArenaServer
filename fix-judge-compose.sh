#!/bin/bash
# 修复 docker-compose.judge.yml 的路径配置

sed -i 's|context: ./deploy/judge|context: .|g' docker-compose.judge.yml

echo "✅ 修改完成！查看修改结果："
grep -A 1 "build:" docker-compose.judge.yml


