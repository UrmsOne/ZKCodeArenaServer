# 数据库字段问题分析报告

## 📋 概览

结合 `deploy/mongo/init/init.js` 数据库初始化脚本，发现代码中存在多个数据库字段不一致的问题。

---

## 🔴 严重问题（需立即修复）

### 1. TestCase 模型缺少 `updated_at` 字段

**问题描述：**
- **init.js 定义**：测试用例应包含 `created_at` 和 `updated_at` 两个时间字段
- **代码实现**：`pkg/models/problem.go` 中的 `TestCase` 结构体**只有 `CreatedAt`，缺少 `UpdatedAt`**

**影响范围：**
- 创建测试用例时不会设置 `updated_at`
- 更新测试用例时虽然 `BaseRepository.UpdateOne` 会自动添加 `updated_at`，但模型无法接收和返回该字段
- 前端无法获取测试用例的最后更新时间
- 数据库中可能存在 `updated_at` 字段，但代码无法操作

**受影响文件：**
```
pkg/models/problem.go
pkg/app/api-server/repository/testcase_repository.go
```

**修复建议：**
```go
// pkg/models/problem.go
type TestCase struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id" swaggertype:"string"`
    ProblemID   primitive.ObjectID `bson:"problem_id" json:"problem_id" swaggertype:"string"`
    Input       string             `bson:"input" json:"input"`
    Output      string             `bson:"output" json:"output"`
    IsSample    bool               `bson:"is_sample" json:"is_sample"`
    TimeLimit   *int               `bson:"time_limit,omitempty" json:"time_limit,omitempty"`
    MemoryLimit *int               `bson:"memory_limit,omitempty" json:"memory_limit,omitempty"`
    Score       int                `bson:"score" json:"score"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"` // ✅ 添加此字段
}
```

```go
// pkg/app/api-server/repository/testcase_repository.go
func (r *TestCaseRepository) CreateTestCase(ctx context.Context, testCase *models.TestCase) error {
    testCase.ID = primitive.NewObjectID()
    testCase.CreatedAt = time.Now()
    testCase.UpdatedAt = time.Now() // ✅ 添加此行
    
    // ... 其余代码
}

func (r *TestCaseRepository) BatchCreateTestCases(ctx context.Context, testCases []*models.TestCase) (*BatchImportResult, error) {
    // ...
    for i, testCase := range testCases {
        testCase.ID = primitive.NewObjectID()
        testCase.CreatedAt = time.Now()
        testCase.UpdatedAt = time.Now() // ✅ 添加此行
        
        // ... 其余代码
    }
}
```

---

## ⚠️ 潜在问题

### 2. 用户集合密码字段处理

**问题描述：**
- `init.js` 中插入的默认管理员用户**没有设置密码字段**
- 代码中 `User` 模型要求 `Password` 是必填字段（`binding:"required"`）

**当前状态：**
```javascript
// deploy/mongo/init/init.js (第55-63行)
db.users.insertOne({
    "username": "admin",
    "email": "admin@zk.edu.cn",
    "real_name": "系统管理员",
    "role": "admin",
    "is_active": true,
    "created_at": new Date(),
    "updated_at": new Date()
    // ❌ 缺少 password 字段
});
```

**影响：**
- 默认管理员账号无法登录（没有密码）
- 需要手动为该账号设置密码

**修复建议：**
```javascript
// 添加默认密码（建议使用 bcrypt 加密后的密码）
db.users.insertOne({
    "username": "admin",
    "password": "$2a$10$vI8aWBnW3fID.ZQ4/zo1G.q1lRps.9cGLcZEiGDMVr5yUP1KUOYTa", // 示例：admin123
    "email": "admin@zk.edu.cn",
    "real_name": "系统管理员",
    "role": "admin",
    "is_active": true,
    "created_at": new Date(),
    "updated_at": new Date()
});
```

### 3. 索引覆盖查询优化建议

**当前索引：**
```javascript
// submits 集合
db.submits.createIndex({ "user_id": 1, "problem_id": 1, "status": 1 }, 
    { name: "idx_user_problem_status" });
```

**分析：**
- 该复合索引用于用户题目状态查询，设计合理
- 但根据代码查询模式，建议添加更多优化

**建议新增索引：**
```javascript
// 提交列表查询优化（按时间倒序）
db.submits.createIndex({ "user_id": 1, "created_at": -1 });
db.submits.createIndex({ "problem_id": 1, "created_at": -1 });

// 统计查询优化
db.submits.createIndex({ "user_id": 1, "status": 1, "created_at": -1 });
```

---

## ✅ 正确实现的部分

### 1. Problem 模型字段完整
```go
// ✅ 包含所有 init.js 中定义的索引字段
type Problem struct {
    Title       string            `bson:"title"`       // text 索引
    Description string            `bson:"description"` // text 索引
    Difficulty  ProblemDifficulty `bson:"difficulty"`  // 普通索引
    Tags        []string          `bson:"tags"`        // 数组索引
    Status      ProblemStatus     `bson:"status"`      // 普通索引
    IsPublic    bool              `bson:"is_public"`   // 普通索引
    CreatedBy   primitive.ObjectID `bson:"created_by"` // 普通索引
    CreatedAt   time.Time         `bson:"created_at"`
    UpdatedAt   time.Time         `bson:"updated_at"`
}
```

### 2. User 模型字段完整
```go
// ✅ 包含所有唯一索引字段
type User struct {
    Username  string `bson:"username"`   // unique 索引
    Email     string `bson:"email"`      // unique 索引
    StudentID string `bson:"student_id"` // unique, sparse 索引
    CreatedAt time.Time `bson:"created_at"`
    UpdatedAt time.Time `bson:"updated_at"`
}
```

### 3. Submit 模型字段完整
```go
// ✅ 包含所有索引字段
type Submit struct {
    UserID    primitive.ObjectID `bson:"user_id"`    // 索引
    ProblemID primitive.ObjectID `bson:"problem_id"` // 索引
    Status    SubmitStatus       `bson:"status"`     // 索引
    CreatedAt time.Time          `bson:"created_at"` // 索引（倒序）
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

---

## 📝 修复优先级

| 优先级 | 问题 | 严重程度 | 修复难度 |
|--------|------|----------|----------|
| 🔴 P0 | TestCase 缺少 UpdatedAt 字段 | 高 | 低 |
| 🟡 P1 | 默认管理员账号缺少密码 | 中 | 低 |
| 🟢 P2 | 添加查询优化索引 | 低 | 低 |

---

## 🔧 修复检查清单

- [x] ✅ 修复 `TestCase` 模型，添加 `UpdatedAt` 字段
- [x] ✅ 更新 `testcase_repository.go` 中的创建逻辑，设置 `UpdatedAt`
- [x] ✅ 更新 `testcase_repository.go` 中的批量创建逻辑，设置 `UpdatedAt`
- [x] ✅ 修复 `init.js` 中默认管理员的密码字段
- [ ] 🔄 重新生成 Swagger 文档（需要手动执行）
- [ ] 🔄 运行数据库迁移（如果需要）

---

## 📊 数据库字段对比表

### TestCase 集合

| 字段 | init.js | models/problem.go | 状态 |
|------|---------|-------------------|------|
| _id | ✅ | ✅ | ✅ |
| problem_id | ✅ | ✅ | ✅ |
| input | ✅ | ✅ | ✅ |
| output | ✅ | ✅ | ✅ |
| is_sample | ✅ | ✅ (IsSample) | ✅ |
| time_limit | ✅ (可选) | ✅ (TimeLimit *int) | ✅ |
| memory_limit | ✅ (可选) | ✅ (MemoryLimit *int) | ✅ |
| score | ✅ (可选) | ✅ (Score int) | ✅ |
| created_at | ✅ | ✅ (CreatedAt) | ✅ |
| updated_at | ✅ | ❌ **缺失** | 🔴 |

---

## 🎯 总结

**核心问题：** `TestCase` 模型缺少 `updated_at` 字段，导致数据库定义与代码模型不一致。

**建议操作顺序：**
1. 修复 `TestCase` 模型定义
2. 更新 repository 层的创建逻辑
3. 修复 `init.js` 中的默认管理员密码
4. 重新生成 Swagger 文档
5. 测试验证修复效果

---

**生成时间：** 2025-10-27  
**检查范围：** 完整数据库结构与代码模型对比

