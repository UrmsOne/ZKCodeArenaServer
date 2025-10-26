# 判题系统集成测试指南

## 📋 概述

本文档说明如何运行判题系统的集成测试,以验证系统的核心功能是否正常工作。

## 🎯 测试覆盖范围

集成测试脚本 (`scripts/integration_test.js`) 将自动测试以下功能:

### 1. **服务健康检查**
- 验证 API 服务是否正常运行
- 检查服务响应状态

### 2. **用户认证**
- 测试用户登录功能
- 验证 JWT token 生成

### 3. **题目管理**
- 获取题目列表
- 验证题目数据完整性

### 4. **代码提交与判题**
测试多种判题场景:
- ✅ **正确答案 (Accepted)** - 验证正确代码能通过所有测试用例
- ❌ **错误答案 (Wrong Answer)** - 验证错误逻辑被正确识别
- 🔧 **编译错误 (Compile Error)** - 验证编译错误检测
- 💥 **运行时错误 (Runtime Error)** - 验证运行时异常处理
- ⏱️ **超时 (Time Limit Exceeded)** - 验证超时检测(可选)

### 5. **数据库状态验证**
- 检查提交记录
- 验证没有遗留的 pending 任务
- 统计各状态的提交数量

## 🔧 前置条件

### 1. 环境依赖

确保以下服务已启动:

```bash
# MongoDB (默认端口: 27017)
# 地址: mongodb://42.194.245.236:27017

# go-judge 沙箱 (默认端口: 5050)
# 地址: http://localhost:5050

# API 服务器 (默认端口: 8080)
# 地址: http://localhost:8080
```

### 2. 测试数据

运行测试前,需要先初始化测试数据:

```bash
cd scripts
npm run init-test-data
```

这将创建:
- 2 个测试用户 (test_user_1, test_user_2)
- 3 个测试题目 (A+B Problem, 数组排序, 字符串反转)
- 15 个测试用例

### 3. Node.js 依赖

确保已安装所需的 npm 包:

```bash
cd scripts
npm install
```

## 🚀 运行测试

### 方式 1: 使用 npm 脚本 (推荐)

```bash
cd scripts
npm test
```

### 方式 2: 直接运行

```bash
cd scripts
node integration_test.js
```

### 方式 3: 自定义配置

可以通过环境变量自定义配置:

```bash
# 自定义 API 地址
API_URL=http://localhost:9000 npm test

# 自定义 MongoDB 地址
MONGO_URI=mongodb://localhost:27017 npm test

# 同时自定义多个配置
API_URL=http://localhost:9000 MONGO_URI=mongodb://localhost:27017 npm test
```

## 📊 测试输出

### 成功示例

```
========================================
判题系统集成测试
========================================
API 地址: http://localhost:8080
MongoDB: mongodb://42.194.245.236:27017
========================================

=== 测试 1: 服务健康检查 ===
✅ 服务健康检查

=== 测试 2: 用户登录 ===
✅ 用户登录成功

=== 测试 3: 获取题目列表 ===
✅ 获取题目列表 (共 3 个题目)

使用题目: A+B Problem (ID: 68e23c29fefc0ac8eec149ed)

=== 测试 4: 提交正确代码 ===
✅ 提交代码成功
   ⏳ 等待判题完成...
✅ 判题结果: accepted
   时间: 245ms, 内存: 15360KB
   通过用例: 5/5

=== 测试 5: 提交错误代码 (Wrong Answer) ===
✅ 提交代码成功
   ⏳ 等待判题完成...
✅ 判题结果: wrong_answer

=== 测试 6: 提交编译错误代码 ===
✅ 提交代码成功
   ⏳ 等待判题完成...
✅ 判题结果: compile_error
   编译错误信息: Main.java:8: error: ';' expected...

=== 测试 7: 提交运行时错误代码 ===
✅ 提交代码成功
   ⏳ 等待判题完成...
✅ 判题结果: runtime_error

=== 测试 8: 数据库状态验证 ===
✅ 数据库中有 4 条提交记录
✅ 没有遗留的 pending 任务
   提交状态分布:
     accepted: 1
     wrong_answer: 1
     compile_error: 1
     runtime_error: 1

========================================
测试总结
========================================
总测试数: 12
✅ 通过: 12
❌ 失败: 0
通过率: 100.00%

========================================  
```

### 失败示例

如果测试失败,会显示详细的错误信息:

```
❌ 用户登录失败
   错误: 用户名或密码错误

失败的测试:
1. 用户登录失败
   用户名或密码错误
```

## 🐛 故障排查

### 问题 1: 服务健康检查失败

**症状**: `❌ 服务健康检查`

**可能原因**:
- API 服务器未启动
- 端口配置错误
- 防火墙阻止连接

**解决方法**:
```bash
# 检查服务是否运行
curl http://localhost:8080/health

# 检查端口占用
netstat -ano | findstr :8080

# 启动服务
go run cmd/main.go serve
```

### 问题 2: 用户登录失败

**症状**: `❌ 用户登录失败`

**可能原因**:
- 测试数据未初始化
- 用户凭证错误
- 数据库连接问题

**解决方法**:
```bash
# 重新初始化测试数据
cd scripts
npm run init-test-data

# 检查 MongoDB 连接
mongosh mongodb://42.194.245.236:27017/zk_code_arena
```

### 问题 3: 判题超时

**症状**: `❌ 等待判题结果超时`

**可能原因**:
- go-judge 沙箱未启动
- 判题服务未启动
- 消息队列阻塞

**解决方法**:
```bash
# 检查 go-judge 服务
curl http://localhost:5050/version

# 查看服务日志
# 检查判题服务是否正常消费任务

# 检查 MongoDB 中的提交状态
mongosh mongodb://42.194.245.236:27017/zk_code_arena
> db.submits.find({status: "pending"}).count()
> db.submits.find({status: "running"}).count()
```

### 问题 4: 数据库连接失败

**症状**: `❌ 数据库状态验证失败`

**可能原因**:
- MongoDB 未启动
- 连接字符串错误
- 网络问题

**解决方法**:
```bash
# 测试 MongoDB 连接
mongosh mongodb://42.194.245.236:27017

# 检查 MongoDB 服务状态
# Windows: services.msc 查找 MongoDB
# Linux: systemctl status mongod
```

## 📝 测试配置

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `API_URL` | `http://localhost:8080` | API 服务器地址 |
| `MONGO_URI` | `mongodb://42.194.245.236:27017` | MongoDB 连接字符串 |

### 测试用户

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `test_user_1` | `test123` | student |
| `test_user_2` | `test123` | student |

### 测试题目

| 题目名称 | 难度 | 时间限制 | 内存限制 |
|----------|------|----------|----------|
| A+B Problem | easy | 1000ms | 128MB |
| 数组排序 | medium | 2000ms | 256MB |
| 字符串反转 | easy | 1000ms | 128MB |

## 🔄 持续集成

可以将此测试脚本集成到 CI/CD 流程中:

```yaml
# .github/workflows/integration-test.yml
name: Integration Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mongodb:
        image: mongo:6
        ports:
          - 27017:27017
      
      go-judge:
        image: criyle/go-judge
        ports:
          - 5050:5050
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: |
          cd scripts
          npm install
      
      - name: Initialize test data
        run: |
          cd scripts
          npm run init-test-data
      
      - name: Start API server
        run: |
          go run cmd/main.go serve &
          sleep 5
      
      - name: Run integration tests
        run: |
          cd scripts
          npm test
```

## 📚 相关文档

- [判题系统重构设计文档](../判题系统重构/DESIGN_判题系统重构.md)
- [判题系统验证设计文档](./DESIGN_判题系统验证.md)
- [测试数据初始化脚本](../../scripts/init_test_data_node.js)
- [集成测试脚本](../../scripts/integration_test.js)

## ✅ 验收标准

集成测试通过的标准:

- ✅ 所有测试用例通过率 ≥ 90%
- ✅ 无遗留的 pending 或 running 状态任务
- ✅ 判题结果准确(AC/WA/CE/RE 判断正确)
- ✅ 响应时间合理(单次判题 < 30秒)
- ✅ 数据库状态一致性

## 🎉 总结

集成测试脚本提供了自动化的端到端测试,确保判题系统的核心功能正常工作。建议在每次代码变更后运行测试,以及时发现和修复问题。
