# 题目相关接口 Postman 使用指南

本文档提供 `server_problem.go` 中所有接口的 Postman 调用示例。

## 前置说明

### 基础配置
- **Base URL**: `http://localhost:8080/api/v1`（根据实际环境调整）
- **需要认证的接口**: 在 Headers 中添加 `Authorization: Bearer {your_jwt_token}`

### 环境变量建议
在 Postman 中设置以下环境变量：
- `base_url`: `http://localhost:8080/api/v1`
- `token`: 登录后获取的 JWT Token
- `problem_id`: 测试用的题目ID

---

## 1. 获取每日推荐题目

### 请求配置
```
GET {{base_url}}/daily-problem
```

### Headers
```
Content-Type: application/json
```

### Query Parameters
无

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "daily_problem": {
      "id": "507f1f77bcf86cd799439011",
      "title": "两数之和",
      "difficulty": "easy",
      "tags": ["数组", "哈希表"],
      "acceptance_rate": 45.5,
      "total_submissions": 1000,
      "total_accepted": 455,
      "user_status": "accepted"
    },
    "message": "每日推荐题目获取成功"
  }
}
```

---

## 2. 获取题目列表（用户端）

### 请求配置
```
GET {{base_url}}/problem
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 可选，带token可获取用户状态
```

### Query Parameters
| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| page | int | 否 | 页码，默认1 | 1 |
| page_size | int | 否 | 每页数量，默认10，最大100 | 20 |
| difficulty | string | 否 | 难度：easy/medium/hard | easy |
| tags | array | 否 | 标签列表（可多个） | 数组,哈希表 |

### 完整示例
```
GET {{base_url}}/problem?page=1&page_size=20&difficulty=easy&tags=数组&tags=哈希表
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problems": [
      {
        "id": "507f1f77bcf86cd799439011",
        "title": "两数之和",
        "difficulty": "easy",
        "tags": ["数组", "哈希表"],
        "acceptance_rate": 45.5,
        "total_submissions": 1000,
        "total_accepted": 455,
        "user_status": "accepted"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20,
    "total_page": 3
  }
}
```

---

## 3. 获取题目列表（管理员端）

### 请求配置
```
GET {{base_url}}/admin/problems
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 必须，且需要管理员权限
```

### Query Parameters
同用户端题目列表接口

### 完整示例
```
GET {{base_url}}/admin/problems?page=1&page_size=10
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problems": [
      {
        "id": "507f1f77bcf86cd799439011",
        "title": "两数之和",
        "difficulty": "easy",
        "tags": ["数组"],
        "status": "published",
        "is_public": true,
        "created_by": "507f1f77bcf86cd799439012"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_page": 10
  }
}
```

---

## 4. 获取题目详情

### 请求配置
```
GET {{base_url}}/problem/{{problem_id}}
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 访问非公开题目时必须
```

### Path Variables
- `problem_id`: 题目ID（ObjectID格式）

### 完整示例
```
GET {{base_url}}/problem/507f1f77bcf86cd799439011
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "title": "两数之和",
    "description": "给定一个整数数组...",
    "input": "输入格式说明",
    "output": "输出格式说明",
    "sample_input": "样例输入",
    "sample_output": "样例输出",
    "hint": "提示信息",
    "difficulty": "easy",
    "tags": ["数组", "哈希表"],
    "time_limit": 1000,
    "memory_limit": 256,
    "is_public": true,
    "status": "published",
    "created_at": "2025-01-25T10:00:00Z",
    "updated_at": "2025-01-25T10:00:00Z"
  }
}
```

---

## 5. 获取题目详情聚合信息

### 请求配置
```
GET {{base_url}}/problem/{{problem_id}}/detail
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 访问非公开题目时必须
```

### Path Variables
- `problem_id`: 题目ID

### 完整示例
```
GET {{base_url}}/problem/507f1f77bcf86cd799439011/detail
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problem": {
      "id": "507f1f77bcf86cd799439011",
      "title": "两数之和",
      "description": "给定一个整数数组...",
      "difficulty": "easy",
      "tags": ["数组", "哈希表"]
    },
    "examples": [
      {
        "input": "示例输入1",
        "output": "示例输出1",
        "explanation": "说明"
      }
    ]
  }
}
```

---

## 6. 创建题目 ⭐

### 请求配置
```
POST {{base_url}}/problem
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 必须，且需要教师或管理员权限
```

### Request Body
```json
{
  "title": "两数之和",
  "description": "给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值的那两个整数，并返回它们的数组下标。",
  "input": "第一行包含一个整数 n，表示数组长度\n第二行包含 n 个整数，表示数组元素\n第三行包含一个整数 target",
  "output": "输出两个整数，表示满足条件的两个数的下标（从0开始）",
  "sample_input": "4\n2 7 11 15\n9",
  "sample_output": "0 1",
  "hint": "可以使用哈希表来优化时间复杂度",
  "source": "LeetCode",
  "author": "张三老师",
  "difficulty": "easy",
  "tags": ["数组", "哈希表"],
  "time_limit": 1000,
  "memory_limit": 256,
  "status": "draft",
  "is_public": false
}
```

### 字段说明
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 题目标题 |
| description | string | 是 | 题目描述 |
| input | string | 否 | 输入格式说明 |
| output | string | 否 | 输出格式说明 |
| sample_input | string | 否 | 样例输入 |
| sample_output | string | 否 | 样例输出 |
| hint | string | 否 | 提示信息 |
| source | string | 否 | 题目来源 |
| author | string | 否 | 题目作者 |
| difficulty | string | 是 | 难度：easy/medium/hard |
| tags | array | 否 | 标签数组 |
| time_limit | int | 否 | 时间限制（毫秒），默认1000 |
| memory_limit | int | 否 | 内存限制（MB），默认256 |
| status | string | 否 | 状态：draft/published，默认draft |
| is_public | bool | 否 | 是否公开，默认false |

### 业务规则
- ⚠️ 草稿状态（draft）不能设为公开（is_public: true）
- 默认状态为草稿（draft）
- 只有教师和管理员可以创建题目

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "title": "两数之和",
    "description": "给定一个整数数组...",
    "difficulty": "easy",
    "tags": ["数组", "哈希表"],
    "status": "draft",
    "is_public": false,
    "created_by": "507f1f77bcf86cd799439012",
    "created_at": "2025-01-25T10:00:00Z",
    "updated_at": "2025-01-25T10:00:00Z"
  }
}
```

### 错误示例
```json
// 草稿状态不能公开
{
  "status": "draft",
  "is_public": true  // ❌ 错误：草稿不能公开
}

// 响应
{
  "code": 400,
  "message": "草稿状态的题目不能设为公开"
}
```

---

## 7. 更新题目

### 请求配置
```
PUT {{base_url}}/problem/{{problem_id}}
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 必须，且需要是题目创建者或管理员
```

### Path Variables
- `problem_id`: 要更新的题目ID

### Request Body
```json
{
  "title": "两数之和（修改版）",
  "description": "更新后的描述...",
  "difficulty": "medium",
  "tags": ["数组", "哈希表", "双指针"],
  "status": "published",
  "is_public": true,
  "time_limit": 2000,
  "memory_limit": 512
}
```

### 字段说明
- 所有字段都是可选的
- 只传需要更新的字段
- 同样遵循"草稿状态不能公开"规则

### 完整示例
```
PUT {{base_url}}/problem/507f1f77bcf86cd799439011
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "title": "两数之和（修改版）",
    "description": "更新后的描述...",
    "difficulty": "medium",
    "status": "published",
    "is_public": true,
    "updated_at": "2025-01-25T11:00:00Z"
  }
}
```

---

## 8. 删除题目

### 请求配置
```
DELETE {{base_url}}/problem/{{problem_id}}
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 必须，且需要管理员权限
```

### Path Variables
- `problem_id`: 要删除的题目ID

### 完整示例
```
DELETE {{base_url}}/problem/507f1f77bcf86cd799439011
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "删除成功"
  }
}
```

---

## 9. 搜索题目

### 请求配置
```
GET {{base_url}}/problem/search
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 可选，带token可获取用户状态
```

### Query Parameters
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 搜索关键词 |
| difficulty | string | 否 | 难度筛选 |
| tags | array | 否 | 标签筛选 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

### 完整示例
```
GET {{base_url}}/problem/search?keyword=数组&difficulty=easy&tags=哈希表&page=1&page_size=10
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problems": [
      {
        "id": "507f1f77bcf86cd799439011",
        "title": "两数之和",
        "difficulty": "easy",
        "tags": ["数组", "哈希表"],
        "user_status": "accepted"
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 10,
    "total_page": 1
  }
}
```

---

## 10. 运行代码测试

### 请求配置
```
POST {{base_url}}/problem/{{problem_id}}/run
```

### Headers
```
Content-Type: application/json
Authorization: Bearer {{token}}  # 必须
```

### Path Variables
- `problem_id`: 题目ID

### Request Body
```json
{
  "code": "#include <stdio.h>\nint main() {\n    printf(\"Hello World\\n\");\n    return 0;\n}",
  "language": "c"
}
```

### 字段说明
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 代码内容 |
| language | string | 是 | 编程语言：c/cpp/java/python/go 等 |

### 完整示例
```
POST {{base_url}}/problem/507f1f77bcf86cd799439011/run
```

### 示例响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "success",
    "output": "Hello World\n",
    "error": "",
    "time_used": 10,
    "memory_used": 1024
  }
}
```

---

## Postman Collection 导入

### 快速创建集合步骤

1. **创建新集合**
   - 打开 Postman
   - 点击 "New" → "Collection"
   - 命名为 "题目管理接口"

2. **设置环境变量**
   - 点击右上角齿轮图标
   - 新建环境 "Development"
   - 添加变量：
     - `base_url`: `http://localhost:8080/api/v1`
     - `token`: （登录后填入）
     - `problem_id`: （测试用题目ID）

3. **导入请求**
   - 按照上述每个接口的配置创建请求
   - 保存到对应的文件夹中

### 建议的文件夹结构
```
题目管理接口/
├── 公开接口/
│   ├── 获取每日推荐题目
│   ├── 获取题目列表
│   ├── 获取题目详情
│   ├── 获取题目详情聚合
│   └── 搜索题目
├── 教师/管理员接口/
│   ├── 创建题目
│   ├── 更新题目
│   └── 删除题目
└── 用户接口/
    └── 运行代码测试
```

---

## 常见错误码

| 状态码 | 说明 | 可能原因 |
|--------|------|----------|
| 200 | 成功 | 请求正常处理 |
| 400 | 请求参数错误 | JSON格式错误、参数类型错误、违反业务规则 |
| 401 | 未授权 | 未登录或token过期 |
| 403 | 权限不足 | 没有执行该操作的权限 |
| 404 | 资源不存在 | 题目ID不存在 |
| 500 | 服务器错误 | 服务器内部错误 |

---

## 测试流程建议

### 1. 基础流程
```
1. 用户登录 → 获取 token
2. 设置环境变量 token
3. 测试公开接口（获取题目列表、详情等）
```

### 2. 教师/管理员流程
```
1. 管理员/教师登录 → 获取 token
2. 创建题目（草稿状态）
3. 更新题目（发布并公开）
4. 测试获取题目详情
5. 删除题目（管理员）
```

### 3. 完整测试流程
```
1. 创建草稿题目 (is_public: false, status: draft)
2. 验证草稿不能公开 (尝试 is_public: true 应失败)
3. 更新为已发布 (status: published)
4. 更新为公开 (is_public: true)
5. 用户端查询题目列表（应该能看到）
6. 测试代码运行
7. 搜索题目
8. 删除题目
```

---

## 注意事项

1. **认证Token**: 需要先调用登录接口获取 JWT Token
2. **权限控制**: 
   - 创建/更新题目：需要教师或管理员权限
   - 删除题目：仅管理员
3. **业务规则**: 
   - 草稿状态不能公开
   - 非公开题目需要登录才能查看
4. **频率限制**: 
   - 代码运行接口有频率限制
5. **ObjectID格式**: 
   - 题目ID必须是有效的MongoDB ObjectID格式（24位十六进制字符串）

---

## 附录：完整的创建题目示例

```bash
# 使用 curl 命令示例
curl -X POST "http://localhost:8080/api/v1/problem" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "title": "两数之和",
    "description": "给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值的那两个整数，并返回它们的数组下标。",
    "input": "第一行包含一个整数 n\n第二行包含 n 个整数\n第三行包含目标值 target",
    "output": "输出两个整数，表示满足条件的下标",
    "sample_input": "4\n2 7 11 15\n9",
    "sample_output": "0 1",
    "difficulty": "easy",
    "tags": ["数组", "哈希表"],
    "time_limit": 1000,
    "memory_limit": 256,
    "status": "draft",
    "is_public": false
  }'
```

