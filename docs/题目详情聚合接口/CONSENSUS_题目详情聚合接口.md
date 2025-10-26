# CONSENSUS - 题目详情聚合接口

## 明确的需求描述

创建一个新的题目详情聚合接口，为前端提供完整的题目展示信息，包括：
- 题目基本信息（标题、描述、难度等）
- 所有示例测试用例（输入、输出）
- 不包含用户提交状态（保持接口简洁）

## 验收标准

### 功能性验收标准
- [ ] 新接口 `GET /problem/:id/detail` 能正确返回题目详情和示例测试用例
- [ ] 返回格式统一，包含完整的题目信息和示例用例数组
- [ ] 权限控制：公开题目无需认证，私有题目需要登录验证
- [ ] 错误处理：无效ID返回400，题目不存在返回404，权限不足返回401

### 性能验收标准
- [ ] 接口响应时间 < 200ms（正常情况下）
- [ ] 单次请求获取所有必要信息，避免前端多次请求

### 兼容性验收标准
- [ ] 不影响现有 `/problem/:id` 接口功能
- [ ] 遵循现有API响应格式规范
- [ ] 包含完整的Swagger文档注释

## 技术实现方案

### 架构设计
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│   Client        │    │   Server Layer   │    │   Service Layer     │
│                 │───▶│  GetProblemDetail │───▶│ GetProblemDetail    │
│ GET /problem/   │    │  - 参数验证      │    │ - 获取题目信息      │
│ :id/detail      │    │  - 权限检查      │    │ - 获取示例测试用例  │
│                 │    │  - 错误处理      │    │ - 数据组装返回      │
└─────────────────┘    └──────────────────┘    └─────────────────────┘
                                  │                        │
                                  │                        ▼
                         ┌────────────────┐    ┌─────────────────────┐
                         │  Response      │    │  Repository Layer   │
                         │ ProblemDetail  │    │ - ProblemRepository │
                         │ Response       │    │ - TestCaseRepository│
                         └────────────────┘    └─────────────────────┘
```

### 数据模型设计
```go
// ProblemDetailResponse 题目详情响应模型
type ProblemDetailResponse struct {
    Problem     *Problem   `json:"problem"`      // 题目基本信息
    SampleCases []TestCase `json:"sample_cases"` // 示例测试用例列表
}
```

### 接口规范
```
GET /problem/:id/detail

Path Parameters:
- id: string (required) - 题目ID（MongoDB ObjectID格式）

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "problem": {
      "id": "507f1f77bcf86cd799439011",
      "title": "两数之和",
      "description": "题目描述...",
      "difficulty": "easy",
      "time_limit": 1000,
      "memory_limit": 256,
      // ... 其他字段
    },
    "sample_cases": [
      {
        "id": "507f1f77bcf86cd799439012",
        "problem_id": "507f1f77bcf86cd799439011",
        "input": "nums = [2,7,11,15], target = 9",
        "output": "[0,1]",
        "is_sample": true
      }
      // ... 更多示例用例
    ]
  }
}
```

## 技术约束和集成方案

### 技术约束
1. **框架约束**: 使用现有的Gin框架和中间件
2. **数据库约束**: 复用现有的MongoDB连接和Repository层
3. **认证约束**: 使用现有的JWT中间件进行权限控制
4. **响应格式约束**: 遵循现有的统一响应格式

### 集成方案
1. **Repository层**: 复用现有的 `ProblemRepository.GetProblemByID` 和 `TestCaseRepository.GetSampleTestCases`
2. **Service层**: 在 `ProblemService` 中新增 `GetProblemDetail` 方法
3. **Server层**: 在 `/problem` 路由组下添加新的 `/:id/detail` 路由
4. **模型层**: 在 `models/requests.go` 中新增 `ProblemDetailResponse` 结构体

### 权限设计
- **公开题目**: 无需认证，直接返回
- **私有题目**: 检查JWT Token，验证用户登录状态
- **不存在的题目**: 返回404错误
- **权限不足**: 返回401错误

## 任务边界限制

### 包含范围
✅ 新增题目详情聚合接口  
✅ 创建响应数据模型  
✅ 实现Service层业务逻辑  
✅ 添加Server层路由处理  
✅ 编写Swagger文档  
✅ 权限控制实现  

### 排除范围  
❌ 不修改现有题目接口  
❌ 不返回用户提交状态  
❌ 不返回非示例测试用例  
❌ 不涉及缓存实现  
❌ 不包含提交记录接口  
❌ 不包含题解相关功能  

## 风险评估和缓解策略

### 主要风险
1. **性能风险**: 单次请求需要查询两个集合
   - **缓解**: 利用MongoDB的查询优化，考虑后续添加缓存
   
2. **数据一致性风险**: 题目和测试用例可能不匹配
   - **缓解**: 在Service层进行数据校验
   
3. **权限泄露风险**: 私有题目的示例用例被错误暴露
   - **缓解**: 严格按照题目的权限设置控制访问

### 监控指标
- API响应时间
- 错误率统计  
- 权限验证成功率

## 最终确认

所有不确定性已解决：
- ✅ 接口路径确定为 `/problem/:id/detail`
- ✅ 不包含用户提交状态
- ✅ 响应模型设计清晰
- ✅ 权限策略与现有系统保持一致
- ✅ 技术实现方案可行
