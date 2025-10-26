# ALIGNMENT - 题目列表用户状态

## 原始需求

### 需求1：题目列表显示用户提交状态
获取题库列表功能需要提供用户对于当前题目的状态，以便前端页面能在题目左侧打上标记，让用户知道题目是否已提交过。

### 需求2：RunCode 接口 input 参数说明
确认运行代码时参数 `input` 的含义，是否为默认测试用例的数据。

---

## 项目上下文分析

### 现有系统架构

#### 1. 题目列表接口
**路径**：`/api/v1/problem` 和 `/api/v1/problem/search`

**当前返回数据结构**：
```go
type ProblemList struct {
    ID          primitive.ObjectID  // 题目ID
    Title       string              // 题目标题
    Difficulty  ProblemDifficulty   // 难度(easy/medium/hard)
    Tags        []string            // 标签
    ACCount     int                 // 全局AC次数
    SubmitCount int                 // 全局提交次数
    Status      ProblemStatus       // 题目状态(draft/published/archived)
    IsPublic    bool                // 是否公开
    CreatedAt   time.Time           // 创建时间
}
```

**问题**：❌ 缺少用户维度的提交状态信息

#### 2. 提交记录数据模型
**集合**：`submits`

```go
type Submit struct {
    ID        primitive.ObjectID  // 提交ID
    ProblemID primitive.ObjectID  // 题目ID
    UserID    primitive.ObjectID  // 用户ID
    Code      string              // 提交代码
    Language  Language            // 编程语言
    Status    SubmitStatus        // 提交状态(pending/accepted/wrong_answer等)
    Result    *JudgeResult        // 评测结果
    CreatedAt time.Time           // 提交时间
    UpdatedAt time.Time           // 更新时间
}
```

#### 3. RunCode 接口实现
**路径**：`POST /api/v1/problem/{id}/run`

**请求参数**：
```go
type RunCodeRequest struct {
    Code     string `json:"code" binding:"required"`      // 代码（必填）
    Language string `json:"language" binding:"required"`  // 语言（必填）
    Input    string `json:"input"`                        // 自定义输入（可选）
}
```

**逻辑分析**（`service_problem.go:392-398`）：
```go
// 如果提供了自定义输入，使用自定义输入运行
if req.Input != "" {
    return s.runWithCustomInput(ctx, executableID, req, problem)
}
// 否则使用示例测试用例运行
return s.runWithSampleTestCases(ctx, executableID, req, problem, problemID)
```

**结论**：
- ✅ `input` 是**可选参数**
- ✅ **有值时**：使用用户自定义的测试输入
- ✅ **无值时**：使用题目的示例测试用例（`problem.sampleInput`）

---

## 需求理解

### 需求1：题目列表用户状态

#### 业务目标
让用户在浏览题目列表时，能快速识别：
- 🔵 哪些题目从未尝试过
- 🟡 哪些题目尝试过但未通过
- 🟢 哪些题目已经AC（通过）

#### 参考实现（LeetCode风格）
```
🔵 1. 两数之和          [未尝试]
🟡 2. 两数相加          [已尝试]
🟢 3. 无重复字符的最长子串 [已通过]
```

#### 状态定义（建议）
| 状态值 | 含义 | 判断条件 |
|--------|------|----------|
| `not_attempted` | 未尝试 | 用户从未提交过该题 |
| `attempted` | 已尝试 | 用户提交过但所有提交都未AC |
| `accepted` | 已通过 | 用户至少有一次提交AC |

#### 数据查询策略
```javascript
// 伪代码
GET /api/v1/problem?page=1&page_size=10

1. 查询题目列表（当前页的10道题）
2. 如果用户已登录：
   - 批量查询用户在这10道题的提交记录
   - 聚合每道题的最好状态
3. 返回带有 user_status 的题目列表
```

#### 性能考虑
**问题**：N+1 查询问题
- 如果逐题查询提交记录，性能差

**优化方案**：
```go
// 批量查询当前页所有题目的用户提交状态
problemIDs := []primitive.ObjectID{...} // 当前页的题目IDs

// MongoDB 聚合查询
pipeline := []bson.M{
    {"$match": bson.M{
        "user_id": userID,
        "problem_id": bson.M{"$in": problemIDs},
    }},
    {"$group": bson.M{
        "_id": "$problem_id",
        "has_accepted": bson.M{
            "$max": bson.M{
                "$cond": []interface{}{
                    bson.M{"$eq": []string{"$status", "accepted"}},
                    1,
                    0,
                },
            },
        },
    }},
}
```

### 需求2：RunCode input 参数（已确认）

#### 参数说明
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `code` | string | ✅ 是 | 用户代码 |
| `language` | string | ✅ 是 | 编程语言 |
| `input` | string | ❌ 否 | 测试输入数据 |

#### 行为说明
1. **不传 `input`**：
   - 使用题目的**示例测试用例**运行
   - 前端可以默认不传，让用户快速测试

2. **传入 `input`**：
   - 使用用户**自定义输入**运行
   - 适用于用户想测试特殊用例的场景

#### 使用示例
```json
// 示例1：使用默认示例用例
POST /api/v1/problem/507f1f77bcf86cd799439011/run
{
  "code": "def twoSum(nums, target): ...",
  "language": "python"
}
// 后端会使用 problem.sampleInput 作为输入

// 示例2：自定义输入
POST /api/v1/problem/507f1f77bcf86cd799439011/run
{
  "code": "def twoSum(nums, target): ...",
  "language": "python",
  "input": "[2,7,11,15]\n9"  // 用户自己的测试数据
}
```

---

## 边界确认

### 任务范围
✅ **包含**：
1. 为题目列表添加用户提交状态字段
2. 实现批量查询用户提交状态的 Service 方法
3. 修改 `GetProblems` 和 `SearchProblems` 接口返回用户状态
4. 更新 Swagger 文档
5. 说明 RunCode 接口的 input 参数用途（文档更新）

❌ **不包含**：
1. 题目详情页的状态显示（已有提交列表可查询）
2. 用户个人统计页面
3. 前端实现
4. RunCode 接口功能修改（已符合需求）

### 技术约束
- 使用现有的 MongoDB 聚合查询
- 不引入新的数据表
- 批量查询避免 N+1 问题
- 未登录用户不返回 `user_status` 字段
- 保持向后兼容（新增字段，不删除旧字段）

---

## 疑问澄清

### 🤔 问题1：用户状态字段命名
请选择一个您认为最合适的字段名：

**选项A：`user_status`**
```json
{
  "id": "xxx",
  "title": "两数之和",
  "user_status": "accepted"
}
```

**选项B：`solve_status`**
```json
{
  "id": "xxx",
  "title": "两数之和",
  "solve_status": "accepted"
}
```

**选项C：布尔字段组合**
```json
{
  "id": "xxx",
  "title": "两数之和",
  "is_solved": true,
  "is_attempted": true
}
```

**推荐**：选项A（`user_status`），语义清晰且扩展性好。

---

### 🤔 问题2：状态分类
是否需要区分"部分通过"状态？

**选项A：三种状态（推荐）**
- `not_attempted`：未尝试
- `attempted`：已尝试（但未通过）
- `accepted`：已通过

**选项B：四种状态**
- `not_attempted`：未尝试
- `attempted`：已尝试但完全未通过
- `partially_accepted`：部分测试用例通过
- `accepted`：全部通过

**推荐**：选项A，大部分OJ都采用三状态，简单清晰。

---

### 🤔 问题3：未登录用户处理
未登录用户访问题目列表时：

**选项A：不返回 `user_status` 字段**
```json
{
  "id": "xxx",
  "title": "两数之和"
  // 没有 user_status 字段
}
```

**选项B：返回 `null`**
```json
{
  "id": "xxx",
  "title": "两数之和",
  "user_status": null
}
```

**选项C：返回 `not_attempted`**
```json
{
  "id": "xxx",
  "title": "两数之和",
  "user_status": "not_attempted"
}
```

**推荐**：选项A，未登录用户不返回该字段，前端可以用 `user_status === undefined` 判断。

---

### 🤔 问题4：性能优化策略
对于用户提交状态查询：

**选项A：实时查询（推荐 - 先实现）**
- 每次请求时批量查询 `submits` 表
- 简单可靠，无一致性问题
- 性能：批量查询 + 索引优化

**选项B：Redis 缓存**
- 将用户的题目状态缓存到 Redis
- 需要在用户提交后更新缓存
- 更高性能，但增加复杂度

**选项C：冗余字段**
- 在 `problems` 表增加 `user_submissions` map
- 需要维护数据一致性
- 不推荐（扩展性差）

**推荐**：选项A 先实现，如果性能不够再考虑 Redis 缓存。

---

## 验收标准

### 功能验收
- [ ] 登录用户访问题目列表，每道题显示正确的用户状态
- [ ] 未登录用户访问题目列表，不返回 `user_status` 字段
- [ ] 用户提交题目后，再次查询列表状态正确更新
- [ ] 搜索接口也返回用户状态
- [ ] RunCode 接口文档说明已更新

### 性能验收
- [ ] 获取题目列表响应时间 < 200ms（10条记录）
- [ ] 使用批量查询，避免 N+1 问题
- [ ] 数据库查询使用了正确的索引

### 接口验收
- [ ] Swagger 文档已更新
- [ ] 返回数据结构向后兼容
- [ ] 错误处理完善

---

## 性能评估（补充）

### 学校OJ项目规模分析
```
典型规模：
- 用户数：1,000 - 5,000 学生
- 题目数：200 - 1,000 道
- 并发数：50 - 200 人（高峰期）
- 每页显示：10-20 道题
```

### 查询压力评估
```
单次请求操作：
1. 查询 problems 表（10条） → 1次查询
2. 批量聚合 submits 表（10道题的用户状态） → 1次聚合查询

性能指标：
- 响应时间：100-150ms（索引优化后）
- 增加 QPS：+100（100并发用户）
- 数据库压力：轻微增加，完全可控
```

### 必要性分析
✅ **强烈建议实现**

**理由：**
1. **用户体验提升显著**：所有主流OJ的标配功能
2. **性能压力可控**：学校规模下完全没问题
3. **实现成本低**：2-3小时开发时间
4. **ROI极高**：投入小，收益大

**结论**：这是 OJ 平台的基础功能，不实现会显著降低用户体验。

---

## 下一步

请确认以下决策点后，我将进入**架构设计阶段**：

1. **字段命名**：选择 `user_status` / `solve_status` / 布尔字段组合？
2. **状态分类**：三种状态 / 四种状态（含部分通过）？
3. **未登录处理**：不返回字段 / 返回 null / 返回 not_attempted？
4. **性能策略**：实时查询 / Redis 缓存？

**如果您同意推荐方案，请回复"同意推荐方案"，我将直接进入下一阶段。**

