# ALIGNMENT - 题目详情聚合接口

## 原始需求

前端需要一个接口，当用户打开一道题目时，需要展示：
- 题目信息
- 默认测试用例
- 默认测试用例的输入和结果等信息

要求将这些信息聚合成一个结果提供。

## 项目上下文分析

### 现有项目架构
- **技术栈**: Go + Gin + MongoDB  
- **架构模式**: Repository + Service + Server 三层架构
- **数据访问**: 已重构为Repository层处理数据操作
- **认证机制**: JWT中间件

### 现有相关接口
1. **`GET /problem/:id`** - 获取题目详情
   - 返回: `models.Problem` 对象
   - 权限: 公开题目无需认证，私有题目需要登录
   - 缺陷: 只包含单个 `SampleInput` 和 `SampleOutput` 字段

2. **`GET /testcase/problem/:problem_id`** - 获取题目测试用例
   - 返回: 所有测试用例列表
   - 权限: 需要认证（管理员权限）
   - 缺陷: 普通用户无法访问

### 现有数据模型
```go
type Problem struct {
    ID           primitive.ObjectID `json:"id"`
    Title        string             `json:"title"`
    Description  string             `json:"description"`
    SampleInput  string             `json:"sample_input"`   // 单个示例输入
    SampleOutput string             `json:"sample_output"`  // 单个示例输出
    // ... 其他字段
}

type TestCase struct {
    ID        primitive.ObjectID `json:"id"`
    ProblemID primitive.ObjectID `json:"problem_id"`
    Input     string             `json:"input"`
    Output    string             `json:"output"`
    IsSample  bool               `json:"is_sample"`  // 区分示例测试用例
    // ... 其他字段
}
```

### 现有服务层能力
- `ProblemService.GetProblemByID` - 获取题目详情
- `TestCaseRepository.GetSampleTestCases` - 获取示例测试用例
- Repository层已经实现了数据访问解耦

## 需求理解确认

### 边界确认（明确任务范围）
- **包含内容**:
  - 创建新的聚合接口 `GET /problem/:id/detail`
  - 返回题目基本信息 + 所有示例测试用例
  - 支持公开访问（示例测试用例应该是公开的）
  
- **不包含内容**:
  - 不修改现有 `GET /problem/:id` 接口
  - 不返回非示例测试用例
  - 不涉及权限管理的复杂变更

### 需求理解（对现有项目的理解）
1. **数据来源**: 
   - 题目基本信息来自 `Problem` 表
   - 示例测试用例来自 `TestCase` 表（`is_sample: true`）

2. **权限策略**:
   - 公开题目: 任何人都可以访问
   - 私有题目: 需要登录用户才能访问
   - 示例测试用例: 应该与题目权限保持一致

3. **架构对齐**:
   - 新接口应该遵循现有的 Repository -> Service -> Server 架构
   - 复用现有的 `ProblemRepository` 和 `TestCaseRepository`

### 疑问澄清（存在歧义的地方）
1. **接口路径**: 使用 `/problem/:id/detail` 还是其他路径？
2. **返回格式**: 是否需要新的响应模型，还是直接组合现有模型？
3. **缓存策略**: 是否需要考虑缓存提升性能？
4. **用户状态**: 是否需要包含用户对该题目的提交状态？

## 智能决策和回答

### 基于现有项目模式的决策

1. **接口路径决策**: 使用 `/problem/:id/detail`
   - 理由: 与现有 `/problem/:id` 形成清晰区分，语义明确

2. **响应模型设计**: 创建新的聚合响应模型
   ```go
   type ProblemDetailResponse struct {
       Problem      *Problem    `json:"problem"`
       SampleCases  []TestCase  `json:"sample_cases"`
   }
   ```

3. **权限策略**: 与现有题目权限保持一致
   - 公开题目: 无需认证
   - 私有题目: 需要登录验证

4. **架构设计**: 遵循现有分层架构
   - Repository层: 复用现有方法
   - Service层: 创建聚合方法
   - Server层: 新增路由处理

### 不确定点需要确认
- **用户状态**: 是否需要在聚合接口中包含用户提交状态？（建议不包含，保持接口简洁）

## 技术实现概要

### 核心组件
1. **ProblemService.GetProblemDetail()** - 聚合业务逻辑
2. **Server.GetProblemDetail()** - 路由处理
3. **新增响应模型** - 统一返回格式

### 实现步骤预览
1. 创建聚合响应模型
2. 在 ProblemService 中添加 GetProblemDetail 方法
3. 在 Server 中添加路由和处理函数
4. 更新 Swagger 文档

## 验收标准
- [ ] 新接口能正确返回题目详情和示例测试用例
- [ ] 权限控制与现有题目接口保持一致
- [ ] 响应格式清晰，便于前端使用
- [ ] 不影响现有接口功能
- [ ] 包含完整的Swagger文档
