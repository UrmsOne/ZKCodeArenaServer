# 请求参数统一整理报告

**日期**: 2025/10/09  
**任务**: 统一整理项目中所有 API 请求参数定义

---

## 📋 背景

在接口完备性巡检过程中发现，项目中请求参数定义方式混乱，存在以下问题：

1. **定义分散**：部分在 `models/course.go`，部分在 handler 中使用匿名结构体
2. **不可复用**：匿名结构体无法在多处重用
3. **文档不友好**：Swagger 无法正确识别匿名结构体，需使用临时注释方式
4. **缺乏一致性**：不同模块的请求参数定义风格不统一

## ✅ 解决方案

### 1. 创建统一请求参数文件

**新建文件**: `pkg/models/requests.go`

所有 API 请求参数统一定义在此文件中，按模块分组：

```go
// ==================== 用户模块请求 ====================
type LoginRequest struct { ... }
type UpdateProfileRequest struct { ... }
type UpdateUserRequest struct { ... }

// ==================== 题目模块请求 ====================
// (RunCodeRequest 在 service 包中)

// ==================== 提交模块请求 ====================
type SubmitCodeRequest struct { ... }

// ==================== 测试用例模块请求 ====================
type CreateTestCaseRequest struct { ... }
type UpdateTestCaseRequest struct { ... }
type BatchCreateTestCasesRequest struct { ... }

// ==================== 课程模块请求 ====================
type CreateCourseRequest struct { ... }
type UpdateCourseRequest struct { ... }
type PageQueryCourseRequest struct { ... }

// ==================== 班级模块请求 ====================
type CreateClazzRequest struct { ... }
type UpdateClazzRequest struct { ... }
type JoinClazzRequest struct { ... }
type AddClazzMemberRequest struct { ... }
type RemoveClazzMembersRequest struct { ... }

// ==================== 任务模块请求 ====================
type CreateTaskRequest struct { ... }
type AddTaskRequest struct { ... }
type UpdateTaskRequest struct { ... }
type FinishTaskRequest struct { ... }
```

---

## 🔧 修改详情

### 用户模块 (`server_user.go`)

#### 1. LoginRequest
**修改前**:
```go
var loginReq struct {
    StudentID string `json:"student_id" binding:"required"`
    Password  string `json:"password" binding:"required"`
}
```

**修改后**:
```go
var loginReq models.LoginRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.LoginRequest true "登录信息"
```

#### 2. UpdateProfileRequest
**修改前**:
```go
var updateReq struct {
    RealName string `json:"real_name"`
    Email    string `json:"email"`
    Bio      string `json:"bio"`
    School   string `json:"school"`
    Major    string `json:"major"`
    Grade    string `json:"grade"`
    Class    string `json:"class"`
    Phone    string `json:"phone"`
}
```

**修改后**:
```go
var updateReq models.UpdateProfileRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.UpdateProfileRequest true "用户信息"
```

#### 3. UpdateUserRequest
**修改前**:
```go
var updateReq struct {
    Role     string `json:"role"`
    IsActive *bool  `json:"is_active"`
}
```

**修改后**:
```go
var updateReq models.UpdateUserRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.UpdateUserRequest true "更新信息"
```

---

### 提交模块 (`server_submit.go`)

#### SubmitCodeRequest
**修改前**:
```go
var submitReq struct {
    ProblemID primitive.ObjectID `json:"problem_id" binding:"required"`
    Code      string             `json:"code" binding:"required"`
    Language  string             `json:"language" binding:"required"`
}
```

**修改后**:
```go
var submitReq models.SubmitCodeRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.SubmitCodeRequest true "提交信息"
```

---

### 测试用例模块 (`server_testcase.go`)

#### 1. CreateTestCaseRequest
**修改前**:
```go
var req struct {
    ProblemID   string `json:"problem_id" binding:"required"`
    Input       string `json:"input" binding:"required"`
    Output      string `json:"output" binding:"required"`
    IsSample    bool   `json:"is_sample"`
    TimeLimit   *int   `json:"time_limit,omitempty"`
    MemoryLimit *int   `json:"memory_limit,omitempty"`
    Score       int    `json:"score,omitempty"`
}
```

**修改后**:
```go
var req models.CreateTestCaseRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.CreateTestCaseRequest true "测试用例信息"
```

#### 2. UpdateTestCaseRequest
**修改前**:
```go
var req struct {
    Input       string `json:"input"`
    Output      string `json:"output"`
    IsSample    *bool  `json:"is_sample,omitempty"`
    TimeLimit   *int   `json:"time_limit,omitempty"`
    MemoryLimit *int   `json:"memory_limit,omitempty"`
    Score       *int   `json:"score,omitempty"`
}
```

**修改后**:
```go
var req models.UpdateTestCaseRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.UpdateTestCaseRequest true "更新信息"
```

#### 3. BatchCreateTestCasesRequest
**修改前**:
```go
var req struct {
    ProblemID string            `json:"problem_id" binding:"required"`
    TestCases []models.TestCase `json:"test_cases" binding:"required,min=1"`
}
```

**修改后**:
```go
var req models.BatchCreateTestCasesRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.BatchCreateTestCasesRequest true "批量测试用例"
```

---

### 课程模块 (`server_course.go`)

#### 1. AddClazzMemberRequest
**修改前**:
```go
var req struct {
    MemberID string `json:"member_id" binding:"required"`
}
```

**修改后**:
```go
var req models.AddClazzMemberRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.AddClazzMemberRequest true "成员ID"
```

#### 2. RemoveClazzMembersRequest
**修改前**:
```go
var body struct {
    MemberIDs []string `json:"member_ids" binding:"required"`
    ClazzID   string   `json:"clazz_id"`
}
```

**修改后**:
```go
var body models.RemoveClazzMembersRequest
```

**Swagger 注释改进**:
```go
// @Param request body models.RemoveClazzMembersRequest true "成员ID列表"
```

---

## 📊 统计数据

### 修改统计
- **新建文件**: 1 个 (`pkg/models/requests.go`)
- **修改文件**: 5 个
  - `server_user.go` - 3 处
  - `server_submit.go` - 1 处
  - `server_testcase.go` - 3 处
  - `server_course.go` - 2 处
  - `course.go` - 清理重复定义

### 请求参数统计
- **用户模块**: 3 个请求类型
- **提交模块**: 1 个请求类型
- **测试用例模块**: 3 个请求类型
- **课程模块**: 3 个请求类型
- **班级模块**: 5 个请求类型
- **任务模块**: 4 个请求类型

**总计**: 19 个请求类型定义

---

## 🎯 改进效果

### 1. **统一管理**
所有请求参数集中在 `pkg/models/requests.go` 一个文件中，便于：
- 快速查找和定位
- 统一修改和维护
- 代码审查和重构

### 2. **类型安全**
使用命名类型替代匿名结构体，提供：
- 编译时类型检查
- IDE 智能提示和自动完成
- 更好的错误提示

### 3. **Swagger 文档友好**
Swagger 可以正确识别并生成完整的请求参数模型：

**生成的模型**（部分）:
```
Generating models.LoginRequest
Generating models.UpdateUserRequest
Generating models.SubmitCodeRequest
Generating models.CreateTestCaseRequest
Generating models.UpdateTestCaseRequest
Generating models.BatchCreateTestCasesRequest
Generating models.AddClazzMemberRequest
Generating models.RemoveClazzMembersRequest
Generating models.CreateCourseRequest
Generating models.UpdateCourseRequest
Generating models.AddTaskRequest
Generating models.UpdateTaskRequest
Generating models.FinishTaskRequest
```

### 4. **代码复用**
相同的请求结构可以在多处重复使用：
- 测试代码
- Mock 数据生成
- 前端接口类型定义

### 5. **规范一致**
与现有的响应格式形成统一的代码组织风格：
```
pkg/utils/
└── response.go     ✅ 统一的响应格式

pkg/models/
└── requests.go     ✅ 统一的请求参数定义
```

---

## 📂 项目结构

统一整理后的项目结构：

```
pkg/
├── models/
│   ├── requests.go      ✅ 统一的请求参数定义（新建）
│   ├── course.go        ✅ 课程相关模型（已清理）
│   ├── problem.go       - 题目相关模型
│   ├── submit.go        - 提交相关模型
│   └── user.go          - 用户相关模型
│
├── utils/
│   └── response.go      ✅ 统一的响应格式
│
└── app/api-server/
    └── server/
        ├── server_user.go        ✅ 已更新使用统一请求类型
        ├── server_submit.go      ✅ 已更新使用统一请求类型
        ├── server_testcase.go    ✅ 已更新使用统一请求类型
        └── server_course.go      ✅ 已更新使用统一请求类型
```

---

## 🔍 验证结果

### 1. 编译检查
```bash
✅ No linter errors found.
```

### 2. Swagger 文档生成
```bash
✅ Successfully generated:
   - docs/docs.go
   - docs/swagger.json
   - docs/swagger.yaml

✅ All request models recognized and generated
```

### 3. 类型检查
所有修改的 handler 函数中：
- ✅ 请求参数正确绑定
- ✅ 字段访问无类型错误
- ✅ Swagger 注释正确引用

---

## 📝 最佳实践建议

### 1. **新增请求参数**
所有新增的 API 请求参数应在 `pkg/models/requests.go` 中定义：

```go
// 示例：新增获取用户列表的请求参数
type GetUsersRequest struct {
    Page     int    `json:"page" form:"page" binding:"required,min=1"`
    PageSize int    `json:"page_size" form:"page_size" binding:"required,min=1,max=100"`
    Role     string `json:"role" form:"role"`
    Keyword  string `json:"keyword" form:"keyword"`
}
```

### 2. **命名规范**
- 请求参数类型命名：`{操作}{模块}Request`
  - 例如：`CreateUserRequest`、`UpdateCourseRequest`
- 响应参数类型命名：`{模块}Response` 或 `{操作}{模块}Response`
  - 例如：`UserProfile`、`GetClazzResponse`

### 3. **注释规范**
每个请求类型应包含清晰的注释：
```go
// CreateUserRequest 用户注册请求
// 包含用户注册所需的基本信息
type CreateUserRequest struct {
    StudentID string `json:"student_id" binding:"required"`
    Username  string `json:"username" binding:"required"`
    Password  string `json:"password" binding:"required,min=6"`
    Email     string `json:"email" binding:"required,email"`
    Role      string `json:"role" binding:"required,oneof=student teacher admin"`
}
```

### 4. **Swagger 注释**
Handler 函数中使用完整的类型引用：
```go
// @Param request body models.CreateUserRequest true "用户信息"
```

### 5. **验证标签**
合理使用 binding 验证标签：
- `required` - 必填字段
- `min=N` / `max=N` - 数值范围
- `email` - 邮箱格式
- `oneof=A B C` - 枚举值
- `omitempty` - 可选字段

---

## 🎉 总结

通过本次请求参数统一整理：

1. ✅ **消除了代码混乱**：所有请求参数定义统一管理
2. ✅ **提升了代码质量**：类型安全、易于维护
3. ✅ **改善了文档生成**：Swagger 可以正确识别所有请求模型
4. ✅ **建立了规范标准**：为后续开发提供了最佳实践模板

**下一步建议**：
- 考虑创建 `pkg/models/responses.go` 统一管理响应类型
- 补充完整的接口文档和使用示例
- 为每个请求类型编写单元测试

