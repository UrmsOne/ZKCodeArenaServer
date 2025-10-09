# 接口完备性巡检 v1 - 待办事项

## ✅ 已完成任务（最新更新）

### ✅ 请求参数统一整理 - 2025/10/09
**完成时间**: 2025年10月9日  
**完成内容**:
- ✅ 创建 `pkg/models/requests.go` 统一管理所有请求参数
- ✅ 消除所有匿名结构体，替换为命名类型
- ✅ 更新 Swagger 注释使用正确的模型引用
- ✅ 重新生成 Swagger 文档，所有请求模型正确识别
- ✅ 清理 `course.go` 中的重复定义
- ✅ 编写详细的改进文档 `REQUEST_PARAMS_UNIFICATION.md`

**改进效果**:
- 19 个请求类型统一定义
- 5 个 server 文件更新
- Swagger 文档质量大幅提升
- 代码可维护性显著增强

**相关文档**: `docs/接口完备性巡检_v1/REQUEST_PARAMS_UNIFICATION.md`

---

## 高优先级任务

### 1. Swagger 接口注释补充 ✅ 已完成基础框架
**预估工作量**: 已完成 60%（剩余 4-6 小时）  
**紧急程度**: 中  
**建议时间**: 可选优化

**详细任务**:
- [ ] **第一批**：核心认证模块（2-3 小时）
  - [ ] POST /user/register - 用户注册
  - [ ] POST /user/login - 用户登录
  - [ ] GET /user/profile - 获取用户资料
  - [ ] PUT /user/profile - 更新用户资料

- [ ] **第二批**：题目和提交模块（3-4 小时）
  - [ ] GET /problems - 获取题目列表
  - [ ] GET /problems/:id - 获取题目详情
  - [ ] POST /submits - 提交代码
  - [ ] GET /submits/:id - 获取提交详情
  - [ ] POST /code/execute - 执行代码测试

- [ ] **第三批**：课程和班级模块（2-3 小时）
  - [ ] POST /courses - 创建课程
  - [ ] GET /courses - 获取课程列表
  - [ ] POST /clazzes - 创建班级
  - [ ] POST /clazzes/:clazzId/join - 加入班级
  - [ ] POST /clazzes/:clazzId/tasks - 创建任务

- [ ] **第四批**：管理功能模块（2-3 小时）
  - [ ] GET /user - 获取用户列表（管理员）
  - [ ] PUT /user/:id - 更新用户（管理员）
  - [ ] DELETE /problems/:id - 删除题目
  - [ ] GET /statistics/system - 系统统计

- [ ] **第五批**：其他功能（2-3 小时）
  - [ ] 测试用例CRUD接口（8个）
  - [ ] 任务管理接口（6个）
  - [ ] 班级成员管理接口（4个）
  - [ ] 统计接口（3个）

**实施步骤**:
1. 在 `cmd/main.go` 添加全局 API 信息注释
2. 在 `server.go` 注册 Swagger 路由：`s.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))`
3. 逐个 handler 添加注释（参考 `SWAGGER_集成指南.md`）
4. 运行 `swag init -g cmd/main.go` 生成文档
5. 访问 `http://localhost:8080/swagger/index.html` 验证

**验收标准**:
- 所有公开 REST API 都有完整的 Swagger 注释
- 参数类型、默认值、枚举值描述准确
- 响应结构包含成功和失败场景
- 需要认证的接口标记 `@Security BearerAuth`

---

### 2. 自动化测试用例补充 ⚠️
**预估工作量**: 5-8 小时  
**紧急程度**: 高  
**建议时间**: 1 周内完成

**详细任务**:
- [x] ✅ 用户模块测试（已完成）
  - [x] 用户注册和登录
  - [x] 用户资料获取和更新
  - [x] 权限控制测试

- [ ] **题目模块测试**（1.5-2 小时）
  - [ ] 题目CRUD测试
  - [ ] 题目列表筛选和分页
  - [ ] 权限控制（公开/私有题目）
  - [ ] 代码执行测试

- [ ] **提交模块测试**（1.5-2 小时）
  - [ ] 代码提交流程
  - [ ] 判题结果查询
  - [ ] 提交历史记录
  - [ ] 用户提交统计

- [ ] **课程模块测试**（2-3 小时）
  - [ ] 课程CRUD测试
  - [ ] 班级管理测试
  - [ ] 邀请码加入班级
  - [ ] 任务创建和完成

- [ ] **测试用例模块测试**（0.5-1 小时）
  - [ ] 测试用例CRUD
  - [ ] 与题目关联测试

- [ ] **统计模块测试**（0.5-1 小时）
  - [ ] 用户统计数据
  - [ ] 系统统计数据
  - [ ] 权限控制

**实施步骤**:
1. 参考 `server_user_test.go` 的测试模式
2. 为每个模块创建对应的 `_test.go` 文件
3. 使用 `testutils` 包中的工具函数
4. 覆盖正常流程、边界条件、错误处理、权限控制
5. 运行 `go test ./pkg/app/api-server/server/... -v`

**验收标准**:
- 每个模块都有对应的测试文件
- 覆盖核心业务流程
- 权限控制测试完整
- 所有测试通过

---

### 3. 密码加密存储 🔒
**预估工作量**: 1-2 小时  
**紧急程度**: 高（安全问题）  
**建议时间**: 尽快完成

**当前状态**:
- `service_user.go:118-125` 有 TODO 标注
- 密码验证逻辑被注释，使用明文比对

**修复步骤**:
1. 启用 `service_user.go` 中的 `bcrypt.CompareHashAndPassword`
2. 取消注释第 119-122 行
3. 删除明文密码比对（第 123-125 行）
4. 验证注册和登录流程

**代码位置**: `pkg/app/api-server/service/service_user.go`

**验收标准**:
- 注册时密码使用 bcrypt 加密
- 登录时使用 bcrypt 验证
- 数据库中不存储明文密码
- 注册登录功能正常

---

## 中优先级任务

### 4. 生成 Swagger 文档并验证 📄
**预估工作量**: 0.5 小时  
**依赖**: 任务 #1（Swagger 注释补充）

**步骤**:
```bash
# 安装 swag CLI（如果未安装）
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/main.go

# 启动服务
go run cmd/main.go run

# 访问文档
http://localhost:8080/swagger/index.html
```

**验收标准**:
- `swag init` 运行无错误
- 生成 `docs/` 目录（docs.go, swagger.json, swagger.yaml）
- 访问 Swagger UI 能看到完整的 API 文档
- 可以在 UI 中交互测试接口

---

### 5. 配置 CI/CD 自动化测试 🤖
**预估工作量**: 2-3 小时  
**建议时间**: 测试补充完成后

**任务**:
- [ ] 创建 GitHub Actions workflow
- [ ] 配置测试数据库（MongoDB、Redis）
- [ ] 运行自动化测试
- [ ] 生成测试覆盖率报告
- [ ] 上传到 Codecov 或类似服务

**参考配置**: 见 `T7_自动化测试指南.md` 中的 GitHub Actions 示例

---

## 低优先级任务

### 6. 功能扩展（未来版本）
**预估工作量**: 根据具体功能评估

- [ ] **密码找回功能** (5-8 小时)
  - 手机验证码集成
  - 密码重置接口
  - 验证码有效期控制

- [ ] **教师审核机制** (3-5 小时)
  - 添加用户审核状态字段
  - 管理员审核接口
  - 审核通知机制

- [ ] **头像上传功能** (2-3 小时)
  - 文件上传接口
  - 图片存储（本地/OSS）
  - 头像更新接口

- [ ] **注册时关联班级邀请码** (2-3 小时)
  - 注册接口添加 `invite_code` 参数
  - 自动加入对应班级
  - 验证邀请码有效性

---

### 7. 性能优化
**预估工作量**: 根据评估结果决定

- [ ] **数据库索引优化**
  - 分析慢查询
  - 添加必要的索引
  - 优化复杂查询

- [ ] **缓存策略调整**
  - Redis 缓存热点数据
  - 缓存失效策略
  - 缓存预热

- [ ] **并发控制优化**
  - 乐观锁优化
  - 事务边界调整
  - 连接池配置

---

### 8. 安全审计
**预估工作量**: 3-5 小时

- [ ] SQL/NoSQL 注入检查
- [ ] XSS 防护验证
- [ ] CSRF 防护
- [ ] 敏感数据加密
- [ ] API 访问频率限制
- [ ] 渗透测试

---

## 配置和环境

### 需要配置的环境变量 / 配置项

#### `.env` 文件 (需要创建)
```
# MongoDB配置（生产环境）
MONGO_URI=mongodb://username:password@host:port
MONGO_DATABASE=zk_code_arena_prod

# Redis配置（生产环境）
REDIS_ADDR=host:port
REDIS_PASSWORD=password
REDIS_DB=0

# Sandbox API（沙箱服务）
SANDBOX_API_URL=http://sandbox-api:5050
SANDBOX_API_TOKEN=your-token-here

# JWT Secret
JWT_SECRET=your-secret-key-here

# 服务器域名
SERVER_DOMAIN=codearena.zkau.edu.cn
```

#### 开发环境 vs 生产环境
- 开发环境使用 `conf/config.yaml`
- 生产环境使用环境变量覆盖
- 确保 `.env` 文件不提交到 Git（已在 `.gitignore`）

---

## 操作指引

### 如何补充 Swagger 注释
1. 参考 `docs/接口完备性巡检_v1/SWAGGER_集成指南.md`
2. 查看示例注释格式
3. 为每个 handler 函数添加注释
4. 运行 `swag init` 生成文档
5. 访问 Swagger UI 验证

### 如何编写测试用例
1. 参考 `docs/接口完备性巡检_v1/T7_自动化测试指南.md`
2. 查看 `server_user_test.go` 示例
3. 使用 `testutils` 包中的工具函数
4. 编写测试用例
5. 运行 `go test` 验证

### 如何部署到生产环境
1. 配置 `.env` 文件
2. 构建 Docker 镜像：`docker build -t zk-code-arena-server .`
3. 运行容器：`docker-compose up -d`
4. 验证服务：`curl http://localhost:8080/health`
5. 查看日志：`docker logs zk-code-arena-server`

---

## 问题反馈

如遇到问题，请记录以下信息：
1. 错误描述
2. 复现步骤
3. 错误日志
4. 环境信息（Go 版本、MongoDB 版本等）

提交 Issue 或联系开发团队。

---

**文档版本**: v1.0  
**最后更新**: 2025/10/9  
**维护人员**: 开发团队

