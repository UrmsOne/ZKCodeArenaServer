# ZK Code Arena API 接口文档

本文档提供了题目创建和测试用例生成接口的详细信息，包括请求结构体、访问路径和示例，以便开发者可以使用这些信息编写Python脚本进行批量导入。

## 基础信息

- **API基础URL**: `http://8.138.184.24:8080/api/v1`
- **认证方式**: Bearer Token
- **请求格式**: JSON
- **响应格式**: JSON

## 通用响应格式

所有API响应都遵循以下格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 具体数据内容
  }
}
```

成功响应的`code`为200，`message`为"success"，具体数据在`data`字段中。

---

## 1. 题目创建接口

### 访问路径
```
POST /api/v1/problem
```

### 请求头
```
Authorization: Bearer <your_token>
Content-Type: application/json
```

### 权限要求
- 需要登录
- 角色要求：管理员(Admin)或教师(Teacher)

### 请求结构体 (CreateProblemRequest)

```json
{
  "title": "题目标题",                    // 必填，字符串，长度1-200
  "description": "题目描述",              // 必填，字符串，最少10个字符
  "input": "输入格式说明",                // 可选，字符串
  "output": "输出格式说明",               // 可选，字符串
  "sample_input": "示例输入",             // 可选，字符串
  "sample_output": "示例输出",            // 可选，字符串
  "hint": "解题提示",                     // 可选，字符串
  "source": "题目来源",                   // 可选，字符串
  "author": "题目作者",                   // 可选，字符串
  "difficulty": "easy|medium|hard",      // 必填，枚举值
  "tags": ["标签1", "标签2"],            // 必填，字符串数组，最多10个，每个标签长度1-20
  "time_limit": 1000,                    // 可选，整数，范围100-10000，单位毫秒，默认1000
  "memory_limit": 256,                   // 可选，整数，范围32-1024，单位MB，默认256
  "status": "draft|published|archived",  // 可选，枚举值，默认draft
  "is_public": true                      // 可选，布尔值，默认false
}
```

### 字段说明

#### 必填字段
- `title`: 题目标题，1-200个字符
- `description`: 题目描述，至少10个字符
- `difficulty`: 题目难度，可选值：
  - `"easy"`: 简单
  - `"medium"`: 中等
  - `"hard"`: 困难
- `tags`: 题目标签，字符串数组，最多10个标签，每个标签长度1-20

#### 可选字段
- `input`: 输入格式说明
- `output`: 输出格式说明
- `sample_input`: 示例输入
- `sample_output`: 示例输出
- `hint`: 解题提示
- `source`: 题目来源
- `author`: 题目作者
- `time_limit`: 时间限制，单位毫秒，范围100-10000，默认1000
- `memory_limit`: 内存限制，单位MB，范围32-1024，默认256
- `status`: 题目状态，可选值：
  - `"draft"`: 草稿（默认）
  - `"published"`: 已发布
  - `"archived"`: 已归档
- `is_public`: 是否公开，布尔值，默认false

### 业务规则
1. 草稿状态的题目不能设为公开
2. 只有管理员和教师可以创建题目
3. 创建者ID会自动从JWT令牌中获取

### 响应示例

成功响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "title": "两数之和",
    "description": "给定一个整数数组 nums 和一个整数目标值 target...",
    "difficulty": "easy",
    "tags": ["数组", "哈希表"],
    "time_limit": 1000,
    "memory_limit": 256,
    "status": "published",
    "is_public": true,
    "created_by": "507f1f77bcf86cd799439012",
    "created_at": "2024-10-26T10:00:00Z",
    "updated_at": "2024-10-26T10:00:00Z"
  }
}
```

错误响应：
```json
{
  "code": 400,
  "message": "请求参数错误: title是必填字段"
}
```

---

## 2. 测试用例生成接口

### 2.1 单个测试用例创建

#### 访问路径
```
POST /api/v1/testcase
```

#### 请求头
```
Authorization: Bearer <your_token>
Content-Type: application/json
```

#### 权限要求
- 需要登录

#### 请求结构体 (CreateTestCaseRequest)

```json
{
  "problem_id": "题目ID",               // 必填，字符串
  "input": "测试输入",                  // 必填，字符串
  "output": "期望输出",                 // 必填，字符串
  "is_sample": false,                  // 可选，布尔值，是否为示例用例
  "time_limit": 1000,                  // 可选，整数，时间限制(ms)
  "memory_limit": 256,                 // 可选，整数，内存限制(MB)
  "score": 10                          // 可选，整数，用例分数
}
```

#### 字段说明

##### 必填字段
- `problem_id`: 题目ID，字符串格式
- `input`: 测试输入数据
- `output`: 期望输出数据

##### 可选字段
- `is_sample`: 是否为示例用例，默认false
- `time_limit`: 时间限制，单位毫秒，会覆盖题目默认设置
- `memory_limit`: 内存限制，单位MB，会覆盖题目默认设置
- `score`: 用例分数，用于部分分计算

#### 响应示例

成功响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "507f1f77bcf86cd799439013",
    "problem_id": "507f1f77bcf86cd799439011",
    "input": "4\n2 7 11 15\n9",
    "output": "0 1",
    "is_sample": false,
    "time_limit": 1000,
    "memory_limit": 256,
    "score": 10,
    "created_at": "2024-10-26T10:00:00Z",
    "updated_at": "2024-10-26T10:00:00Z"
  }
}
```

### 2.2 批量测试用例创建

#### 访问路径
```
POST /api/v1/testcase/batch
```

#### 请求头
```
Authorization: Bearer <your_token>
Content-Type: application/json
```

#### 权限要求
- 需要登录
- 角色要求：管理员(Admin)或教师(Teacher)

#### 请求结构体 (BatchCreateTestCasesRequest)

```json
{
  "problem_id": "题目ID",               // 必填，字符串
  "test_cases": [                      // 必填，测试用例数组
    {
      "input": "测试输入1",
      "output": "期望输出1",
      "is_sample": true,
      "time_limit": 1000,
      "memory_limit": 256,
      "score": 10
    },
    {
      "input": "测试输入2",
      "output": "期望输出2",
      "is_sample": false,
      "time_limit": 1000,
      "memory_limit": 256,
      "score": 10
    }
  ]
}
```

#### 字段说明

##### 必填字段
- `problem_id`: 题目ID，字符串格式
- `test_cases`: 测试用例数组，至少包含一个测试用例

##### test_cases数组中每个对象的字段
- `input`: 测试输入数据，必填
- `output`: 期望输出数据，必填
- `is_sample`: 是否为示例用例，可选，默认false
- `time_limit`: 时间限制，单位毫秒，可选
- `memory_limit`: 内存限制，单位MB，可选
- `score`: 用例分数，可选

#### 响应示例

成功响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success_count": 8,
    "fail_count": 0,
    "errors": []
  }
}
```

---





### 获取题目详情
```
GET /api/v1/problem/{id}
```

### 获取题目测试用例列表
```
GET /api/v1/testcase/problem/{problem_id}
```

###  更新题目
```
PUT /api/v1/problem/{id}
```

### 更新测试用例
```
PUT /api/v1/testcase/{id}
```
