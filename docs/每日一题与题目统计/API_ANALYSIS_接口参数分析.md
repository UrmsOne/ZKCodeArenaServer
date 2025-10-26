# 题目接口参数分析文档

> 文档版本：v1.0  
> 创建时间：2025/10/25  
> 作者：omenkk7

## 概述

本文档详细分析 `GetProblems` 和 `SearchProblems` 两个核心题目接口的参数结构、返回数据、业务逻辑和实现机制。

---

## 1. GetProblems 接口分析

### 1.1 接口基本信息

- **路径**: `GET /api/v1/problem/`
- **功能**: 获取题目列表（分页查询）
- **访问权限**: 公开接口，支持游客和登录用户

### 1.2 请求参数详解

| 参数名 | 类型 | 必需 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `page` | int | 否 | 1 | 页码，从1开始 |
| `page_size` | int | 否 | 10 | 每页条数，范围 1-100 |
| `difficulty` | string | 否 | "" | 难度筛选：easy/medium/hard |
| `tags` | []string | 否 | [] | 标签筛选，支持多个标签 |

#### 参数验证逻辑

```go
// 页码验证
if page < 1 {
    page = 1
}

// 页大小验证
if pageSize < 1 || pageSize > 100 {
    pageSize = 10
}

// 难度验证
validDifficulties := []string{"easy", "medium", "hard"}
// 如果传入无效难度值，会被忽略
```

### 1.3 返回数据结构

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problems": [
      {
        "id": "670123456789abcdef012345",
        "title": "两数之和",
        "difficulty": "easy",
        "tags": ["数组", "哈希表"],
        "ac_count": 1234,
        "submit_count": 2345,
        "status": "published",
        "is_public": true,
        "created_at": "2024-10-01T10:00:00Z",
        "user_status": "accepted"  // 仅登录用户返回
      }
    ],
    "total": 156,
    "page": 1,
    "page_size": 10,
    "total_page": 16
  }
}
```

#### 返回字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `problems` | []ProblemList | 题目列表 |
| `total` | int64 | 符合条件的总题目数 |
| `page` | int | 当前页码 |
| `page_size` | int | 每页条数 |
| `total_page` | int64 | 总页数 |

#### ProblemList 字段详解

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | ObjectID | 题目唯一标识 |
| `title` | string | 题目标题 |
| `difficulty` | string | 题目难度 (easy/medium/hard) |
| `tags` | []string | 题目标签数组 |
| `ac_count` | int | 通过提交数 |
| `submit_count` | int | 总提交数 |
| `status` | string | 题目状态 (draft/published) |
| `is_public` | bool | 是否公开 |
| `created_at` | time.Time | 创建时间 |
| `user_status` | *UserProblemStatus | 用户状态（可选） |

### 1.4 用户状态逻辑

#### 用户状态类型

```go
type UserProblemStatus string

const (
    UserStatusNotAttempted UserProblemStatus = "not_attempted" // 未尝试
    UserStatusAttempted    UserProblemStatus = "attempted"     // 已尝试
    UserStatusAccepted     UserProblemStatus = "accepted"      // 已通过
)
```

#### 状态判定逻辑

1. **未登录用户**: `user_status` 字段不返回 (omitempty)
2. **登录用户**: 
   - 查询用户对每个题目的最新提交状态
   - 使用批量查询优化（避免N+1问题）
   - 状态判定规则：
     - 无提交记录 → `not_attempted`
     - 有提交但未AC → `attempted` 
     - 有AC记录 → `accepted`

#### 性能优化

- **批量状态查询**: 一次性获取用户对所有题目的状态
- **MongoDB索引**: `idx_user_problem_status` (user_id, problem_id, status)
- **聚合查询**: 使用 `$group` 获取每题最新状态

### 1.5 权限控制机制

#### 访问权限

- **游客用户**: 只能查看 `status=published` 且 `isPublic=true` 的题目
- **登录用户**: 根据角色权限查看对应题目
  - 学生：只能查看公开已发布题目
  - 教师：可查看自己创建的私有题目
  - 管理员：可查看所有题目

#### 数据过滤

```go
// 基础过滤条件
filter := bson.M{
    "status":   models.StatusPublished,
    "isPublic": true,
}

// 根据用户角色调整过滤条件
if role == models.RoleTeacher {
    filter = bson.M{
        "$or": []bson.M{
            {"status": models.StatusPublished, "isPublic": true},
            {"creatorId": userID}, // 教师可见自己创建的题目
        },
    }
}
```

---

## 2. SearchProblems 接口分析

### 2.1 接口基本信息

- **路径**: `GET /api/v1/problem/search`
- **功能**: 搜索题目（关键词 + 高级筛选）
- **访问权限**: 公开接口，支持游客和登录用户

### 2.2 请求参数详解

| 参数名 | 类型 | 必需 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `q` | string | 是 | - | 搜索关键词 |
| `page` | int | 否 | 1 | 页码，从1开始 |
| `page_size` | int | 否 | 10 | 每页条数，范围 1-100 |
| `difficulty` | string | 否 | "" | 难度筛选：easy/medium/hard |
| `tags` | []string | 否 | [] | 标签筛选，支持多个标签 |

#### 关键参数说明

**搜索关键词 (`q`)**:
- **必需参数**: 不能为空
- **搜索范围**: 题目标题 (`title`)
- **搜索方式**: 模糊匹配，大小写不敏感
- **索引支持**: MongoDB文本索引 `title_text`

### 2.3 搜索算法说明

#### 搜索实现机制

```go
// 构建搜索条件
searchFilter := bson.M{
    "title": bson.M{
        "$regex":   keyword,
        "$options": "i", // 大小写不敏感
    },
}

// 结合基础过滤条件
finalFilter := bson.M{
    "$and": []bson.M{
        baseFilter,    // 权限和状态过滤
        searchFilter,  // 关键词搜索
        difficultyFilter, // 难度过滤
        tagsFilter,    // 标签过滤
    },
}
```

#### 搜索优化策略

1. **索引利用**:
   - 题目标题文本索引：`db.problems.createIndex({"title": "text"})`
   - 复合索引：`{status: 1, isPublic: 1, title: 1}`

2. **查询性能**:
   - 先应用权限过滤，减少搜索范围
   - 使用正则表达式进行模糊匹配
   - 支持分页，避免大结果集

3. **搜索准确性**:
   - 大小写不敏感匹配
   - 支持部分关键词匹配
   - 可扩展为全文搜索（未来优化）

### 2.4 结果排序逻辑

#### 排序规则

1. **相关性排序**: 
   - 标题完全匹配优先
   - 关键词位置权重（标题开头 > 中间）
   
2. **二级排序**:
   - 创建时间倒序 (`created_at DESC`)
   - 确保结果稳定性

#### 排序实现

```go
// MongoDB排序管道
sort := bson.M{
    "created_at": -1, // 按创建时间倒序
}

// 可扩展为复合排序
// sort := bson.M{
//     "_score": -1,     // 相关性得分
//     "created_at": -1, // 时间
// }
```

### 2.5 返回数据结构

SearchProblems 的返回结构与 GetProblems 完全一致：

```json
{
  "code": 200,
  "message": "success", 
  "data": {
    "problems": [...],  // 同 GetProblems
    "total": 23,
    "page": 1,
    "page_size": 10,
    "total_page": 3
  }
}
```

---

## 3. 接口对比分析

### 3.1 功能差异

| 方面 | GetProblems | SearchProblems |
|------|-------------|----------------|
| 主要用途 | 题目列表浏览 | 关键词搜索 |
| 关键词搜索 | ❌ 不支持 | ✅ 核心功能 |
| 分页查询 | ✅ 支持 | ✅ 支持 |
| 难度筛选 | ✅ 支持 | ✅ 支持 |
| 标签筛选 | ✅ 支持 | ✅ 支持 |
| 用户状态 | ✅ 支持 | ✅ 支持 |
| 默认排序 | 创建时间倒序 | 相关性+时间 |

### 3.2 性能对比

| 方面 | GetProblems | SearchProblems |
|------|-------------|----------------|
| 查询复杂度 | 低 | 中等 |
| 索引依赖 | 基础索引 | 文本索引 |
| 缓存适用性 | 高（结果稳定） | 低（动态搜索） |
| 并发性能 | 优秀 | 良好 |

### 3.3 使用场景

**GetProblems 适用场景**:
- 题目列表页面加载
- 按难度/标签筛选浏览
- 分页导航
- 缓存友好的场景

**SearchProblems 适用场景**:
- 用户主动搜索
- 题目名称模糊查找
- 搜索结果页面
- 实时搜索建议

---

## 4. 技术实现细节

### 4.1 数据库查询优化

#### 核心索引

```javascript
// 基础索引
db.problems.createIndex({"status": 1, "isPublic": 1})
db.problems.createIndex({"created_at": -1})
db.problems.createIndex({"difficulty": 1})

// 搜索索引
db.problems.createIndex({"title": "text"})

// 用户状态索引
db.submits.createIndex({"user_id": 1, "problem_id": 1, "status": 1})
```

#### 查询管道优化

```go
// GetProblems 查询管道
pipeline := []bson.M{
    {"$match": baseFilter},
    {"$sort": bson.M{"created_at": -1}},
    {"$skip": (page - 1) * pageSize},
    {"$limit": pageSize},
}

// SearchProblems 查询管道  
pipeline := []bson.M{
    {"$match": searchFilter},
    {"$sort": bson.M{"created_at": -1}},
    {"$skip": (page - 1) * pageSize}, 
    {"$limit": pageSize},
}
```

### 4.2 用户状态查询优化

#### 批量状态查询

```go
// 聚合管道获取用户状态
pipeline := []bson.M{
    {
        "$match": bson.M{
            "user_id":    userID,
            "problem_id": bson.M{"$in": problemIDs},
        },
    },
    {
        "$group": bson.M{
            "_id":           "$problem_id",
            "latest_status": bson.M{"$last": "$status"},
        },
    },
}
```

#### 状态映射逻辑

```go
func mapSubmitStatusToUserStatus(submitStatus string) models.UserProblemStatus {
    switch submitStatus {
    case string(models.StatusAccepted):
        return models.UserStatusAccepted
    default:
        return models.UserStatusAttempted
    }
}
```

### 4.3 错误处理机制

#### 参数验证错误

```json
{
  "code": 400,
  "message": "请求参数错误",
  "data": {
    "error": "page_size must be between 1 and 100"
  }
}
```

#### 搜索关键词错误

```json
{
  "code": 400, 
  "message": "搜索关键词不能为空",
  "data": null
}
```

#### 内部服务错误

```json
{
  "code": 500,
  "message": "获取题目列表失败: database connection timeout", 
  "data": null
}
```

---

## 5. 前端集成指南

### 5.1 接口调用示例

#### GetProblems 调用

```javascript
// 获取第一页题目（默认）
GET /api/v1/problem/

// 获取指定页面和筛选条件
GET /api/v1/problem/?page=2&page_size=20&difficulty=easy&tags=数组,哈希表
```

#### SearchProblems 调用

```javascript
// 搜索包含"数组"的题目
GET /api/v1/problem/search?q=数组

// 搜索+筛选
GET /api/v1/problem/search?q=两数&difficulty=easy&page=1&page_size=10
```

### 5.2 状态处理建议

#### 用户状态显示

```javascript
function renderProblemStatus(problem) {
    if (!problem.user_status) {
        return ''; // 未登录用户不显示状态
    }
    
    switch (problem.user_status) {
        case 'accepted':
            return '✅ 已通过';
        case 'attempted': 
            return '❌ 已尝试';
        case 'not_attempted':
            return '⚪ 未尝试';
        default:
            return '';
    }
}
```

### 5.3 分页组件集成

```javascript
function PaginationInfo({ total, page, pageSize, totalPage }) {
    return {
        current: page,
        pageSize: pageSize,
        total: total,
        totalPages: totalPage,
        showSizeChanger: true,
        pageSizeOptions: ['10', '20', '50', '100']
    };
}
```

---

## 6. 版本更新记录

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| v1.0 | 2025/10/25 | 初始版本，完整分析GetProblems和SearchProblems接口 |

---

## 7. 相关文档

- [API接口文档](../swagger.yaml)
- [数据模型设计](../../pkg/models/)
- [数据库索引规范](../数据库设计.md)
- [前端集成指南](../前端对接文档.md)

---

*本文档由6A工作流自动化流程生成，如有疑问请联系开发团队。*

