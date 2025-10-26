# 题目创建接口重构 - TODO清单

## 🎯 必须完成的事项

### 1. 代码部署 ⭐⭐⭐

**操作：** 将修改的代码部署到服务器

```bash
# 1. 提交代码
git add pkg/models/requests.go
git add pkg/app/api-server/service/service_problem.go
git add pkg/app/api-server/server/server_problem.go
git add docs/题目创建接口重构/
git commit -m "refactor: 重构题目创建接口，Status默认为Draft"

# 2. 推送到远程
git push origin feature/problem-create-refactor

# 3. 在服务器上拉取代码
cd /path/to/ZKCodeArenaServer
git pull

# 4. 重新编译
go build -o bin/server cmd/*.go

# 5. 重启服务
docker-compose -f docker-compose.dev.yml restart app
# 或
docker-compose -f docker-compose.yml restart app
```

---

### 2. 更新Swagger文档 ⭐⭐⭐

**操作：** 重新生成Swagger文档

```bash
# 在项目根目录执行
swag init -g cmd/main.go

# 验证文档生成
# 访问 http://localhost:8080/swagger/index.html
```

**验证点：**
- [ ] POST /api/problem接口显示正确
- [ ] 请求参数显示为CreateProblemRequest
- [ ] 字段验证规则显示正确

---

### 3. 数据库数据修复 ⭐⭐

**问题：** 现有题目的status可能为空或为published

**操作：** 更新现有数据（可选，根据需要）

```bash
# 连接MongoDB
docker exec -it zk-arena-mongo mongosh zk_code_arena

# 检查现有数据
db.problems.find({}, {title: 1, status: 1, is_public: 1}).pretty()

# 如果需要，将现有题目设为draft（可选）
# 仅在需要时执行
db.problems.updateMany(
  {status: {$in: ["", "published"]}},
  {$set: {status: "draft"}}
)
```

**建议：** 保持现有数据不变，只对新创建的题目应用新规则

---

### 4. 集成测试 ⭐⭐⭐

**操作：** 执行测试用例验证功能

#### TC1: 创建题目（最小参数）

```bash
curl -X POST http://localhost:8080/api/problem \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "两数之和",
    "description": "给定一个整数数组和一个目标值，找出数组中和为目标值的两个数。",
    "difficulty": "easy",
    "tags": ["数组", "哈希表"]
  }'
```

**预期结果：**
```json
{
  "code": 0,
  "data": {
    "status": "draft",       // ✅ 默认draft
    "is_public": false,      // ✅ 默认false
    "time_limit": 1000,      // ✅ 默认1000
    "memory_limit": 256      // ✅ 默认256
  }
}
```

#### TC7: Draft+IsPublic=true（应该失败）

```bash
curl -X POST http://localhost:8080/api/problem \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试题目",
    "description": "这是一个测试题目，至少十个字符。",
    "difficulty": "easy",
    "tags": ["测试"],
    "status": "draft",
    "is_public": true
  }'
```

**预期结果：**
```json
{
  "code": 400,
  "message": "草稿状态的题目不能设为公开"
}
```

#### TC8: Published+IsPublic=true（应该成功）

```bash
curl -X POST http://localhost:8080/api/problem \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "公开题目",
    "description": "这是一个公开的已发布题目。",
    "difficulty": "medium",
    "tags": ["算法"],
    "status": "published",
    "is_public": true
  }'
```

**预期结果：**
```json
{
  "code": 0,
  "data": {
    "status": "published",   // ✅ published
    "is_public": true        // ✅ true
  }
}
```

**测试清单：**
- [ ] TC1: 最小参数测试通过
- [ ] TC3: Title为空返回400
- [ ] TC4: Description太短返回400
- [ ] TC7: Draft+IsPublic=true返回400
- [ ] TC8: Published+IsPublic=true成功

---

## 📋 前端配合事项

### 5. 前端接口调用修改 ⭐⭐⭐

**问题：** 前端请求结构需要调整

**操作指引：**

#### 旧代码（需要修改）

```javascript
// ❌ 旧代码
const createProblem = async (formData) => {
  await api.post('/api/problem', {
    title: formData.title,
    description: formData.description,
    // ... 其他字段
    status: 'published',  // ❌ 不再自动设为published
    is_public: true       // ❌ 需要根据status决定
  });
};
```

#### 新代码（推荐）

```javascript
// ✅ 新代码
const createProblem = async (formData) => {
  const request = {
    title: formData.title,
    description: formData.description,
    difficulty: formData.difficulty,
    tags: formData.tags || [],
    // 可选字段（不传则使用后端默认值）
    // status: 'draft',      // 默认draft，通常不传
    // is_public: false,     // 默认false，通常不传
    // time_limit: 1000,     // 默认1000，通常不传
    // memory_limit: 256,    // 默认256，通常不传
  };
  
  await api.post('/api/problem', request);
};
```

#### UI建议

```javascript
// ✅ 创建后默认为草稿，提供"发布"按钮
const publishProblem = async (problemId) => {
  await api.put(`/api/problem/${problemId}`, {
    status: 'published',
    is_public: true  // 或根据用户选择
  });
};
```

**前端TODO：**
- [ ] 修改创建题目的API调用
- [ ] 移除创建时的status和is_public选项
- [ ] 添加"发布"功能按钮
- [ ] 添加题目状态显示（草稿/已发布）
- [ ] UI上禁止Draft+IsPublic=true组合

---

## 🔧 可选优化事项

### 6. 添加单元测试（可选）⭐

**文件：** `pkg/app/api-server/server/server_problem_test.go`

```go
func TestCreateProblem_DefaultValues(t *testing.T) {
    // 测试默认值设置
}

func TestCreateProblem_DraftCannotBePublic(t *testing.T) {
    // 测试草稿不能公开
}
```

---

### 7. 性能监控（可选）⭐

添加创建题目的性能指标：

```go
// 在Service层添加
start := time.Now()
defer func() {
    duration := time.Since(start)
    utils.Logger.Infof("CreateProblem duration: %v", duration)
}()
```

---

### 8. 错误信息优化（可选）⭐

添加更详细的字段级错误信息：

```go
// 解析binding错误，返回友好提示
if errs, ok := err.(validator.ValidationErrors); ok {
    for _, e := range errs {
        // 返回具体哪个字段错误
    }
}
```

---

## 📝 文档相关

### 9. API文档更新

**位置：** 项目README或API文档

**添加内容：**

#### 创建题目 API

**端点：** `POST /api/problem`

**权限：** Teacher/Admin

**请求体：**
```json
{
  "title": "题目标题",          // 必填，1-200字符
  "description": "题目描述",    // 必填，至少10字符
  "difficulty": "easy",        // 必填，easy/medium/hard
  "tags": ["标签1"],          // 必填，最多10个
  "time_limit": 1000,         // 可选，100-10000ms，默认1000
  "memory_limit": 256,        // 可选，32-1024MB，默认256
  "status": "draft",          // 可选，draft/published/archived，默认draft
  "is_public": false          // 可选，默认false
}
```

**注意事项：**
- 创建的题目默认为草稿状态（draft）
- 草稿状态的题目强制私有（is_public=false）
- 已发布状态可以设置公开或私有

---

## ⚠️ 注意事项总结

### 部署前检查

- [ ] 代码已提交到Git
- [ ] Swagger文档已重新生成
- [ ] 已通知前端开发人员接口变更

### 部署后验证

- [ ] 服务正常启动，无错误日志
- [ ] Swagger文档可访问
- [ ] TC1测试通过（最小参数）
- [ ] TC7测试通过（Draft+IsPublic校验）
- [ ] TC8测试通过（Published可公开）

### 前端对接

- [ ] 前端已收到接口变更通知
- [ ] 前端已修改请求结构
- [ ] 前端已测试通过

---

## 🆘 问题排查

### 如果遇到问题：

#### 1. 编译错误

```bash
# 检查依赖
go mod tidy

# 重新编译
go build -o bin/server cmd/*.go
```

#### 2. Swagger不显示新接口

```bash
# 重新生成
swag init -g cmd/main.go

# 重启服务
docker-compose restart app
```

#### 3. 测试失败

```bash
# 查看日志
docker logs -f zk-code-arena-server-dev

# 检查MongoDB数据
docker exec -it zk-arena-mongo mongosh
use zk_code_arena
db.problems.find().sort({created_at: -1}).limit(1).pretty()
```

#### 4. 前端调用失败

- 检查请求体格式是否正确
- 检查必填字段是否都传了
- 查看后端日志的详细错误信息

---

## 📞 需要支持的事项

### 配置相关

- ✅ 无需额外配置
- ✅ 无需修改环境变量
- ✅ 无需修改数据库schema

### 依赖相关

- ✅ 无需安装新依赖
- ✅ go.mod无变更

### 第三方服务

- ✅ 无需对接新服务

---

## ✅ 完成标志

当以下所有项都完成时，表示重构完全完成：

- [ ] 代码已部署到服务器
- [ ] Swagger文档已更新
- [ ] TC1/TC7/TC8测试通过
- [ ] 前端已完成对接
- [ ] 前端测试通过
- [ ] 生产环境运行稳定（无错误日志）

---

## 🎉 快速开始

**最快的验证方式（5分钟）：**

```bash
# 1. 部署代码
git pull && go build -o bin/server cmd/*.go && docker-compose restart app

# 2. 生成Swagger
swag init -g cmd/main.go

# 3. 获取Token（用你的账号）
TOKEN=$(curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"admin","password":"password"}' \
  | jq -r '.data.token')

# 4. 测试创建题目
curl -X POST http://localhost:8080/api/problem \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"测试","description":"这是一个测试题目描述","difficulty":"easy","tags":["测试"]}'

# 5. 查看结果（应该看到status=draft, is_public=false）
```

---

**需要帮助？** 参考文档：
- `docs/题目创建接口重构/FINAL_题目创建接口重构.md` - 总结报告
- `docs/题目创建接口重构/ACCEPTANCE_题目创建接口重构.md` - 验收文档
- `docs/题目创建接口重构/DESIGN_题目创建接口重构.md` - 设计文档

