# ZK Code Arena Server 静态API文档

## 1. 概述

ZK Code Arena Server 是为仲恺农业工程学院学生打造的在线编程练习平台，提供完整的刷题、作业、竞赛和代码评测功能。

### 1.1 基本信息
- **基础URL**: `http://localhost:8080/api/v1`
- **API版本**: v1.0
- **协议**: HTTP/HTTPS
- **认证方式**: JWT Bearer Token

### 1.2 认证
大多数API端点需要认证。在请求头中添加：
```
Authorization: Bearer <your-jwt-token>
```

### 1.3 通用响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

## 2. 题目模块

### 2.1 获取题目列表
**GET** `/problem`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| page | integer | 否 | 页码，默认1 |
| page_size | integer | 否 | 每页数量，默认10 |
| difficulty | string | 否 | 难度筛选 (easy, medium, hard) |
| tags | array[string] | 否 | 标签筛选 |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problems": [
      {
        "id": "5f9d4b4b4b4b4b4b4b4b4b4b",
        "title": "A+B Problem",
        "difficulty": "easy",
        "tags": ["数学", "入门"],
        "ac_count": 100,
        "submit_count": 200,
        "status": "published",
        "is_public": true,
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_page": 1
  }
}
```

#### 错误响应
- **400**: 请求参数错误
- **500**: 服务器内部错误

### 2.2 获取题目详情
**GET** `/problem/{id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 题目ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4b",
    "title": "A+B Problem",
    "description": "计算两个整数的和",
    "input": "一行两个整数 a 和 b，用空格分隔。",
    "output": "一行一个整数，表示 a + b 的值。",
    "sample_input": "1 2",
    "sample_output": "3",
    "hint": "注意数据范围",
    "source": "经典题目",
    "author": "管理员",
    "difficulty": "easy",
    "time_limit": 1000,
    "memory_limit": 256,
    "tags": ["数学", "入门"],
    "status": "published",
    "is_public": true,
    "ac_count": 100,
    "submit_count": 200,
    "created_by": "5f9d4b4b4b4b4b4b4b4b4b4a",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 错误响应
- **400**: 无效的题目ID
- **401**: 需要登录才能查看私有题目
- **404**: 题目不存在

### 2.3 获取题目详情聚合信息
**GET** `/problem/{id}/detail`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 题目ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problem": {
      "id": "5f9d4b4b4b4b4b4b4b4b4b4b",
      "title": "A+B Problem",
      "description": "计算两个整数的和",
      "input": "一行两个整数 a 和 b，用空格分隔。",
      "output": "一行一个整数，表示 a + b 的值。",
      "sample_input": "1 2",
      "sample_output": "3",
      "hint": "注意数据范围",
      "source": "经典题目",
      "author": "管理员",
      "difficulty": "easy",
      "time_limit": 1000,
      "memory_limit": 256,
      "tags": ["数学", "入门"],
      "status": "published",
      "is_public": true,
      "ac_count": 100,
      "submit_count": 200,
      "created_by": "5f9d4b4b4b4b4b4b4b4b4b4a",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    "sample_cases": [
      {
        "id": "5f9d4b4b4b4b4b4b4b4b4b4c",
        "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
        "input": "1 2",
        "output": "3",
        "is_sample": true,
        "score": 0,
        "created_at": "2025-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### 错误响应
- **400**: 无效的题目ID
- **401**: 需要登录才能查看私有题目
- **404**: 题目不存在或测试用例不存在
- **500**: 服务器内部错误

### 2.4 创建题目
**POST** `/problem`

#### 请求参数
```json
{
  "title": "新题目",
  "description": "题目描述",
  "input": "输入描述",
  "output": "输出描述",
  "sample_input": "样例输入",
  "sample_output": "样例输出",
  "hint": "提示",
  "source": "来源",
  "author": "作者",
  "difficulty": "easy",
  "tags": ["标签1", "标签2"],
  "time_limit": 1000,
  "memory_limit": 256,
  "status": "draft",
  "is_public": false
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4b",
    "title": "新题目",
    "description": "题目描述",
    "input": "输入描述",
    "output": "输出描述",
    "sample_input": "样例输入",
    "sample_output": "样例输出",
    "hint": "提示",
    "source": "来源",
    "author": "作者",
    "difficulty": "easy",
    "time_limit": 1000,
    "memory_limit": 256,
    "tags": ["标签1", "标签2"],
    "status": "draft",
    "is_public": false,
    "ac_count": 0,
    "submit_count": 0,
    "created_by": "5f9d4b4b4b4b4b4b4b4b4b4a",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 错误响应
- **400**: 请求参数错误或业务规则错误
- **401**: 需要登录
- **403**: 权限不足
- **500**: 创建失败

### 2.5 更新题目
**PUT** `/problem/{id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 题目ID |

```json
{
  "title": "更新后的题目",
  "description": "更新后的描述",
  "input": "更新后的输入描述",
  "output": "更新后的输出描述",
  "sample_input": "更新后的样例输入",
  "sample_output": "更新后的样例输出",
  "hint": "更新后的提示",
  "source": "更新后的来源",
  "author": "更新后的作者",
  "difficulty": "medium",
  "tags": ["更新标签1", "更新标签2"],
  "time_limit": 2000,
  "memory_limit": 512,
  "status": "published",
  "is_public": true
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4b",
    "title": "更新后的题目",
    "description": "更新后的描述",
    "input": "更新后的输入描述",
    "output": "更新后的输出描述",
    "sample_input": "更新后的样例输入",
    "sample_output": "更新后的样例输出",
    "hint": "更新后的提示",
    "source": "更新后的来源",
    "author": "更新后的作者",
    "difficulty": "medium",
    "time_limit": 2000,
    "memory_limit": 512,
    "tags": ["更新标签1", "更新标签2"],
    "status": "published",
    "is_public": true,
    "ac_count": 0,
    "submit_count": 0,
    "created_by": "5f9d4b4b4b4b4b4b4b4b4b4a",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 错误响应
- **400**: 请求参数错误或业务规则错误
- **401**: 需要登录
- **403**: 权限不足
- **404**: 题目不存在
- **500**: 更新失败

### 2.6 删除题目
**DELETE** `/problem/{id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 题目ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "删除成功"
  }
}
```

#### 错误响应
- **400**: 无效的题目ID
- **403**: 权限不足
- **500**: 删除失败

### 2.7 运行代码测试
**POST** `/problem/{id}/run`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 题目ID |

```json
{
  "code": "print('Hello World')",
  "language": "python"
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success": true,
    "status": "accepted",
    "output": "Hello World\n",
    "error": "",
    "time_used": 10,
    "memory_used": 1024
  }
}
```

#### 错误响应
- **400**: 请求参数错误
- **404**: 题目不存在
- **500**: 运行失败

### 2.8 搜索题目
**GET** `/problem/search`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| keyword | string | 否 | 关键词 |
| difficulty | string | 否 | 难度筛选 (easy, medium, hard) |
| tags | array[string] | 否 | 标签筛选 |
| page | integer | 否 | 页码，默认1 |
| page_size | integer | 否 | 每页数量，默认10 |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "problems": [
      {
        "id": "5f9d4b4b4b4b4b4b4b4b4b4b",
        "title": "A+B Problem",
        "difficulty": "easy",
        "tags": ["数学", "入门"],
        "ac_count": 100,
        "submit_count": 200,
        "status": "published",
        "is_public": true,
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_page": 1
  }
}
```

#### 错误响应
- **500**: 搜索失败

## 3. 测试用例模块

### 3.1 获取题目的测试用例列表
**GET** `/testcase/problem/{problem_id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| problem_id | string | 是 | 题目ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "test_cases": [
      {
        "id": "5f9d4b4b4b4b4b4b4b4b4b4c",
        "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
        "input": "1 2",
        "output": "3",
        "is_sample": true,
        "time_limit": 1000,
        "memory_limit": 256,
        "score": 20,
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 1
  }
}
```

#### 错误响应
- **400**: 无效的题目ID
- **500**: 获取失败

### 3.2 获取测试用例详情
**GET** `/testcase/{id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 测试用例ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4c",
    "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
    "input": "1 2",
    "output": "3",
    "is_sample": true,
    "time_limit": 1000,
    "memory_limit": 256,
    "score": 20,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 错误响应
- **400**: 无效的测试用例ID
- **500**: 获取失败

### 3.3 创建测试用例
**POST** `/testcase`

#### 请求参数
```json
{
  "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
  "input": "测试输入",
  "output": "测试输出",
  "is_sample": false,
  "time_limit": 1000,
  "memory_limit": 256,
  "score": 20
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4d",
    "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
    "input": "测试输入",
    "output": "测试输出",
    "is_sample": false,
    "time_limit": 1000,
    "memory_limit": 256,
    "score": 20,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 错误响应
- **400**: 请求参数错误
- **500**: 创建失败

### 3.4 更新测试用例
**PUT** `/testcase/{id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 测试用例ID |

```json
{
  "input": "更新后的测试输入",
  "output": "更新后的测试输出",
  "is_sample": true,
  "time_limit": 2000,
  "memory_limit": 512,
  "score": 30
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4d",
    "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
    "input": "更新后的测试输入",
    "output": "更新后的测试输出",
    "is_sample": true,
    "time_limit": 2000,
    "memory_limit": 512,
    "score": 30,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 错误响应
- **400**: 请求参数错误
- **404**: 测试用例不存在
- **500**: 更新失败

### 3.5 删除测试用例
**DELETE** `/testcase/{id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 测试用例ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "测试用例删除成功"
  }
}
```

#### 错误响应
- **400**: 无效的测试用例ID
- **500**: 删除失败

### 3.6 批量创建测试用例
**POST** `/testcase/batch`

#### 请求参数
```json
{
  "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
  "test_cases": [
    {
      "input": "测试输入1",
      "output": "测试输出1",
      "is_sample": false,
      "time_limit": 1000,
      "memory_limit": 256,
      "score": 20
    },
    {
      "input": "测试输入2",
      "output": "测试输出2",
      "is_sample": true,
      "time_limit": 1000,
      "memory_limit": 256,
      "score": 20
    }
  ]
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "批量创建成功"
  }
}
```

#### 错误响应
- **400**: 请求参数错误
- **401**: 需要登录
- **403**: 权限不足
- **404**: 题目不存在
- **500**: 创建失败

## 4. 提交模块

### 4.1 提交代码
**POST** `/submit`

#### 请求参数
```json
{
  "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
  "code": "print('Hello World')",
  "language": "python"
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "提交成功，正在评测中",
    "submit_id": "5f9d4b4b4b4b4b4b4b4b4b4e"
  }
}
```

#### 错误响应
- **400**: 请求参数错误或不支持的编程语言
- **401**: 未认证用户
- **403**: 无法提交私有题目
- **404**: 题目不存在
- **500**: 提交失败

### 4.2 获取提交列表
**GET** `/submit`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| page | integer | 否 | 页码，默认1 |
| page_size | integer | 否 | 每页数量，默认10 |
| problem_id | string | 否 | 题目ID筛选 |
| user_id | string | 否 | 用户ID筛选（管理员可用） |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "submits": [
      {
        "id": "5f9d4b4b4b4b4b4b4b4b4b4e",
        "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
        "user_id": "5f9d4b4b4b4b4b4b4b4b4b4a",
        "language": "python",
        "status": "accepted",
        "time_used": 10,
        "memory_used": 1024,
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_page": 1
  }
}
```

#### 错误响应
- **400**: 请求参数错误
- **401**: 需要登录
- **500**: 获取失败

### 4.3 获取提交详情
**GET** `/submit/{id}`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 提交ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4e",
    "problem_id": "5f9d4b4b4b4b4b4b4b4b4b4b",
    "user_id": "5f9d4b4b4b4b4b4b4b4b4b4a",
    "code": "print('Hello World')",
    "language": "python",
    "status": "accepted",
    "result": {
      "status": "accepted",
      "time_used": 10,
      "memory_used": 1024,
      "compile_error": "",
      "runtime_error": "",
      "test_results": [
        {
          "test_case_id": "5f9d4b4b4b4b4b4b4b4b4b4c",
          "status": "accepted",
          "time_used": 10,
          "memory_used": 1024,
          "is_sample": true,
          "output": "Hello World\n",
          "expected": "Hello World\n",
          "error": ""
        }
      ]
    },
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 错误响应
- **400**: 无效的提交ID
- **401**: 需要登录
- **403**: 权限不足
- **404**: 提交不存在

### 4.4 获取提交状态（轻量级）
**GET** `/submit/{id}/status`

#### 请求参数
| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | string | 是 | 提交ID |

#### 响应示例
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "5f9d4b4b4b4b4b4b4b4b4b4e",
    "status": "accepted",
    "progress": null,
    "message": "通过",
    "updated_at": "2025-01-01T00:00:00Z",
    "time_used": 10,
    "memory_used": 1024
  }
}
```

#### 错误响应
- **400**: 无效的提交ID
- **401**: 需要登录
- **403**: 权限不足
- **404**: 提交不存在

## 5. 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 6. 数据模型

### 6.1 题目模型 (Problem)
| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | string | 题目ID |
| title | string | 题目标题 |
| description | string | 题目描述 |
| input | string | 输入描述 |
| output | string | 输出描述 |
| sample_input | string | 样例输入 |
| sample_output | string | 样例输出 |
| hint | string | 提示 |
| source | string | 来源 |
| author | string | 作者 |
| difficulty | string | 难度 (easy, medium, hard) |
| time_limit | integer | 时间限制(ms) |
| memory_limit | integer | 内存限制(MB) |
| tags | array[string] | 标签 |
| status | string | 状态 (draft, published, archived) |
| is_public | boolean | 是否公开 |
| ac_count | integer | AC次数 |
| submit_count | integer | 提交次数 |
| created_by | string | 创建者ID |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

### 6.2 测试用例模型 (TestCase)
| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | string | 测试用例ID |
| problem_id | string | 题目ID |
| input | string | 输入数据 |
| output | string | 期望输出 |
| is_sample | boolean | 是否为示例用例 |
| time_limit | integer | 时间限制(ms) |
| memory_limit | integer | 内存限制(MB) |
| score | integer | 用例分数 |
| created_at | string | 创建时间 |

### 6.3 提交模型 (Submit)
| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | string | 提交ID |
| problem_id | string | 题目ID |
| user_id | string | 用户ID |
| code | string | 代码 |
| language | string | 编程语言 |
| status | string | 状态 |
| result | object | 评测结果 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

### 6.4 评测结果模型 (JudgeResult)
| 字段名 | 类型 | 描述 |
|--------|------|------|
| status | string | 状态 |
| time_used | integer | 时间使用(ms) |
| memory_used | integer | 内存使用(KB) |
| compile_error | string | 编译错误信息 |
| runtime_error | string | 运行时错误信息 |
| test_results | array[object] | 测试用例结果 |

## 7. 状态码说明

### 7.1 题目状态
| 状态 | 说明 |
|------|------|
| draft | 草稿 |
| published | 已发布 |
| archived | 已归档 |

### 7.2 提交状态
| 状态 | 说明 |
|------|------|
| pending | 等待中 |
| running | 运行中 |
| accepted | 通过 |
| wrong_answer | 答案错误 |
| time_limit | 超时 |
| memory_limit | 内存超限 |
| runtime_error | 运行时错误 |
| compile_error | 编译错误 |
| system_error | 系统错误 |

### 7.3 编程语言
| 语言 | 说明 |
|------|------|
| c | C语言 |
| cpp | C++ |
| java | Java |
| python | Python |
| go | Go |