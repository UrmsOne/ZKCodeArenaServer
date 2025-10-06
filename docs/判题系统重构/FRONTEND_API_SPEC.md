# 判题系统 API 接口规范

> **文档版本**: v1.0  
> **更新日期**: 2025-10-05  
> **目标读者**: 前端开发人员  
> **Base URL**: `http://localhost:8080/api/v1`

---

## 📋 目录

- [测试用例管理 API](#测试用例管理-api)

---

## 🔐 认证说明

所有需要认证的接口都需要在请求头中携带 JWT Token：

```
Authorization: Bearer <your_jwt_token>
```

---

## 测试用例管理 API

### 1. 获取题目的测试用例列表

**接口地址**: `GET /testcase/problem/:problem_id`

**需要认证**: ✅ 是

**请求参数**:
| 参数名 | 位置 | 类型 | 必填 | 说明 |
|--------|------|------|------|------|
| problem_id | path | string | 是 | 题目 ID (ObjectID) |

**请求示例**:
```bash
curl -X GET "http://localhost:8080/api/v1/testcase/problem/507f1f77bcf86cd799439012" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "test_cases": [
      {
        "id": "507f1f77bcf86cd799439011",
        "problem_id": "507f1f77bcf86cd799439012",
        "input": "1 2\n",
        "output": "3\n",
        "is_sample": true,
        "time_limit": 2000,
        "memory_limit": 512,
        "score": 10,
        "created_at": "2025-10-05T10:00:00Z"
      },
      {
        "id": "507f1f77bcf86cd799439013",
        "problem_id": "507f1f77bcf86cd799439012",
        "input": "100 200\n",
        "output": "300\n",
        "is_sample": false,
        "created_at": "2025-10-05T10:00:00Z"
      }
    ],
    "total": 2
  }
}
```

---

### 2. 获取单个测试用例

**接口地址**: `GET /testcase/:id`

**需要认证**: ✅ 是

**请求参数**:
| 参数名 | 位置 | 类型 | 必填 | 说明 |
|--------|------|------|------|------|
| id | path | string | 是 | 测试用例 ID (ObjectID) |

**请求示例**:
```bash
curl -X GET "http://localhost:8080/api/v1/testcase/507f1f77bcf86cd799439011" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "problem_id": "507f1f77bcf86cd799439012",
    "input": "1 2\n",
    "output": "3\n",
    "is_sample": true,
    "time_limit": 2000,
    "memory_limit": 512,
    "score": 10,
    "created_at": "2025-10-05T10:00:00Z"
  }
}
```

---

### 3. 创建测试用例

**接口地址**: `POST /testcase/`

**需要认证**: ✅ 是

**请求参数**:
| 参数名 | 位置 | 类型 | 必填 | 说明 |
|--------|------|------|------|------|
| problem_id | body | string | 是 | 题目 ID (ObjectID) |
| input | body | string | 是 | 输入数据 |
| output | body | string | 是 | 期望输出 |
| is_sample | body | boolean | 否 | 是否为示例用例，默认 false |
| time_limit | body | int | 否 | 时间限制（ms），不填则使用题目默认值 |
| memory_limit | body | int | 否 | 内存限制（MB），不填则使用题目默认值 |
| score | body | int | 否 | 用例分数，默认 0 |

**请求示例**:
```bash
curl -X POST "http://localhost:8080/api/v1/testcase/" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "problem_id": "507f1f77bcf86cd799439012",
    "input": "1 2\n",
    "output": "3\n",
    "is_sample": true,
    "time_limit": 2000,
    "memory_limit": 512,
    "score": 10
  }'
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "problem_id": "507f1f77bcf86cd799439012",
    "input": "1 2\n",
    "output": "3\n",
    "is_sample": true,
    "time_limit": 2000,
    "memory_limit": 512,
    "score": 10,
    "created_at": "2025-10-05T10:00:00Z"
  }
}
```

---

### 4. 更新测试用例

**接口地址**: `PUT /testcase/:id`

**需要认证**: ✅ 是

**请求参数**:
| 参数名 | 位置 | 类型 | 必填 | 说明 |
|--------|------|------|------|------|
| id | path | string | 是 | 测试用例 ID (ObjectID) |
| input | body | string | 否 | 输入数据 |
| output | body | string | 否 | 期望输出 |
| is_sample | body | boolean | 否 | 是否为示例用例 |
| time_limit | body | int | 否 | 时间限制（ms） |
| memory_limit | body | int | 否 | 内存限制（MB） |
| score | body | int | 否 | 用例分数 |

**请求示例**:
```bash
curl -X PUT "http://localhost:8080/api/v1/testcase/507f1f77bcf86cd799439011" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "input": "2 3\n",
    "output": "5\n",
    "time_limit": 3000
  }'
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "problem_id": "507f1f77bcf86cd799439012",
    "input": "2 3\n",
    "output": "5\n",
    "is_sample": true,
    "time_limit": 3000,
    "memory_limit": 512,
    "score": 10,
    "created_at": "2025-10-05T10:00:00Z"
  }
}
```

---

### 5. 删除测试用例

**接口地址**: `DELETE /testcase/:id`

**需要认证**: ✅ 是

**请求参数**:
| 参数名 | 位置 | 类型 | 必填 | 说明 |
|--------|------|------|------|------|
| id | path | string | 是 | 测试用例 ID (ObjectID) |

**请求示例**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/testcase/507f1f77bcf86cd799439011" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "测试用例删除成功"
  }
}
```

---

## 📊 状态码说明

### HTTP 状态码
| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（Token 无效或过期） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |


---

## 🔗 相关文档

- [API_CHANGES_前端开发.md](./API_CHANGES_前端开发.md) - API 变更说明
- [MOCK_CONFIG.json](./MOCK_CONFIG.json) - Mock 数据配置
- [FRONTEND_MOCK_GUIDE.md](./FRONTEND_MOCK_GUIDE.md) - Mock 使用指南
- [FRONTEND_QUICKSTART.md](./FRONTEND_QUICKSTART.md) - 快速开始指南

---

## 📞 技术支持

如有疑问，请联系后端开发团队：
- **项目负责人**: omenkk7
- **更新日期**: 2025-10-05