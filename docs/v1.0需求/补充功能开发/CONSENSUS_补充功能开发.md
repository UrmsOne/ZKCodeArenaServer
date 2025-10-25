# 补充功能开发 - 共识文档

## 📋 文档信息
- **任务名称**: v1.0补充功能开发
- **创建时间**: 2025-10-06
- **版本**: v1.0
- **状态**: ✅ 已确认

---

## 1. 需求描述

### 1.1 原始需求
基于 `SUMMARY_接口开发分析总结.md` 的缺失接口清单，本次开发以下6个接口：

**中优先级（2个）**：
1. 代码运行测试接口
2. 题目搜索接口

**低优先级（4个）**：
3. 批量导入测试用例
4. 收藏题目
5. 用户解题统计
6. 系统统计数据

### 1.2 业务价值
- **代码运行测试**: 提升用户调试体验，减少无效提交
- **题目搜索**: 提高题库可用性，快速定位题目
- **批量导入**: 提高出题效率
- **收藏功能**: 增强用户学习体验
- **统计数据**: 提供学习反馈和管理决策支持

---

## 2. 功能详细规格

### 2.1 代码运行测试

#### 需求描述
允许用户在提交前运行代码，仅使用示例测试用例，不计入提交记录。

#### 已确认决策（Q9）
- ✅ 实现运行测试接口
- ✅ 复用判题逻辑
- ✅ 仅运行示例用例
- ⚠️ 需要频率限制（防滥用）

#### 接口规范
```
POST /api/v1/problem/:id/run

请求体：
{
  "code": "用户代码",
  "language": "java|cpp|python|c|go"
}

响应：
{
  "code": 200,
  "data": {
    "results": [
      {
        "test_case_id": "xxx",
        "status": "accepted|wrong_answer|runtime_error|...",
        "time_used": 100,     // ms
        "memory_used": 1024,  // KB
        "input": "示例输入",
        "output": "实际输出",
        "expected": "期望输出",
        "error": "错误信息"
      }
    ],
    "summary": {
      "total": 3,
      "passed": 2,
      "failed": 1
    }
  }
}
```

#### 业务规则
1. ✅ 只运行 `IsSample = true` 的测试用例
2. ✅ 返回详细的输入输出信息（示例用例可见）
3. ✅ 不创建 Submit 记录
4. ✅ 不更新题目统计（AC数、提交数）
5. ⚠️ 频率限制：每个用户每题每分钟最多运行3次
6. ✅ 需要登录认证

#### 技术约束
- 复用现有 `JudgeService` 的判题逻辑
- 使用 Redis 实现频率限制
- 超时时间：单次运行总时长不超过30秒

---

### 2.2 题目搜索

#### 需求描述
支持题目标题、描述的模糊搜索，以及难度、标签的精确筛选。

#### 已确认决策（Q10）
- ✅ 使用 MongoDB $regex 搜索
- ✅ 支持标题、描述模糊搜索
- ✅ 支持难度、标签筛选
- ❌ 不实现高亮显示（前端处理）
- ❌ 不实现搜索历史（延后）

#### 接口规范
```
GET /api/v1/problem/search?keyword=链表&difficulty=medium&tags=算法,数据结构&page=1&page_size=10

响应：
{
  "code": 200,
  "data": {
    "problems": [...],
    "total": 25,
    "page": 1,
    "page_size": 10,
    "total_page": 3
  }
}
```

#### 查询规则
1. `keyword`: 在 `title` 和 `description` 字段中模糊匹配（不区分大小写）
2. `difficulty`: 精确匹配（可选）
3. `tags`: 数组包含查询（可选，支持多个标签）
4. `is_public`: 默认只查公开题目
5. `status`: 默认只查已发布题目
6. 排序：按创建时间倒序

#### 技术实现
- 使用 MongoDB `$regex` + `$options: "i"` 实现不区分大小写搜索
- 使用 `$and` 组合多个条件
- 建议添加索引：`title`, `difficulty`, `tags`

---

### 2.3 批量导入测试用例

#### 需求描述
允许教师批量导入测试用例，提高出题效率。

#### 接口规范
```
POST /api/v1/testcase/batch

请求体：
{
  "problem_id": "题目ID",
  "test_cases": [
    {
      "input": "输入1",
      "output": "输出1",
      "is_sample": false,
      "score": 10
    },
    {
      "input": "输入2",
      "output": "输出2",
      "is_sample": false,
      "score": 10
    }
  ]
}

响应：
{
  "code": 200,
  "data": {
    "success_count": 10,
    "failed_count": 0,
    "failed_items": []
  }
}
```

#### 业务规则
1. ✅ 需要教师或管理员权限
2. ✅ 验证题目存在且有权限
3. ✅ 单次最多导入100个用例
4. ✅ 事务处理：全部成功或全部失败
5. ✅ 验证必填字段：input, output

#### 可选增强
- 支持 CSV/JSON 文件上传
- 支持从模板下载

---

### 2.4 收藏题目

#### 需求描述
用户可以收藏喜欢的题目，方便后续复习。

#### 数据模型
```go
// 新增 Collection
type ProblemFavorite struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
    ProblemID  primitive.ObjectID `bson:"problem_id" json:"problem_id"`
    CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// 复合索引
// {"user_id": 1, "problem_id": 1} - 唯一索引
// {"user_id": 1, "created_at": -1}
```

#### 接口规范
```
POST /api/v1/problem/:id/favorite
功能：收藏/取消收藏（toggle）

GET /api/v1/user/favorites?page=1&page_size=10
功能：获取用户收藏列表

DELETE /api/v1/problem/:id/favorite
功能：取消收藏
```

#### 响应示例
```json
{
  "code": 200,
  "data": {
    "is_favorited": true,
    "message": "收藏成功"
  }
}
```

#### 业务规则
1. ✅ 需要登录
2. ✅ 重复收藏自动忽略
3. ✅ 取消不存在的收藏返回成功
4. ✅ 题目详情页返回是否已收藏

---

### 2.5 用户解题统计

#### 需求描述
展示用户的解题情况统计，包括总提交数、AC数、各难度题目完成情况等。

#### 接口规范
```
GET /api/v1/user/statistics
或
GET /api/v1/user/:id/statistics

响应：
{
  "code": 200,
  "data": {
    "user_id": "xxx",
    "total_submit": 150,
    "total_ac": 80,
    "ac_rate": 53.3,
    "difficulty_stats": {
      "easy": {
        "total": 50,
        "solved": 30,
        "ac_rate": 60.0
      },
      "medium": {
        "total": 30,
        "solved": 15,
        "ac_rate": 50.0
      },
      "hard": {
        "total": 10,
        "solved": 2,
        "ac_rate": 20.0
      }
    },
    "recent_submits": [
      {
        "problem_id": "xxx",
        "problem_title": "两数之和",
        "status": "accepted",
        "language": "java",
        "submitted_at": "2025-10-06T10:00:00Z"
      }
    ],
    "solved_problems": 47  // 去重的AC题目数
  }
}
```

#### 计算逻辑
1. 从 `submits` 集合聚合统计
2. 统计维度：
   - 总提交数：所有提交
   - 总AC数：status = "accepted" 的提交
   - 解决题目数：AC的去重题目数
   - 按难度分组统计
3. 缓存策略：Redis缓存，TTL 5分钟

---

### 2.6 系统统计数据

#### 需求描述
管理员查看系统运行概况，包括用户数、题目数、提交数等核心指标。

#### 接口规范
```
GET /api/v1/admin/system/stats

响应：
{
  "code": 200,
  "data": {
    "users": {
      "total": 1000,
      "active_today": 50,
      "new_this_week": 30,
      "by_role": {
        "student": 900,
        "teacher": 90,
        "admin": 10
      }
    },
    "problems": {
      "total": 200,
      "published": 180,
      "draft": 20,
      "by_difficulty": {
        "easy": 80,
        "medium": 90,
        "hard": 30
      }
    },
    "submits": {
      "total": 50000,
      "today": 500,
      "this_week": 3000,
      "ac_count": 25000,
      "ac_rate": 50.0
    },
    "courses": {
      "total": 50,
      "active": 30
    },
    "system": {
      "version": "v1.0",
      "uptime": "15 days",
      "judge_queue_size": 5
    }
  }
}
```

#### 权限控制
- ✅ 只有 admin 角色可访问
- ❌ 其他角色返回 403

#### 性能优化
- 使用 MongoDB 聚合管道
- Redis 缓存，TTL 10分钟
- 异步更新缓存

---

## 3. 技术实现方案

### 3.1 整体架构

```
Client → Server (Controller) → Service → DB/Cache
                ↓
         JudgeService (代码运行)
         Redis (频率限制、缓存)
```

### 3.2 核心技术决策

| 功能 | 技术方案 | 理由 |
|-----|---------|------|
| 代码运行 | 复用JudgeService | 代码重用，逻辑一致 |
| 频率限制 | Redis + 滑动窗口 | 高性能，分布式友好 |
| 题目搜索 | MongoDB $regex | 简单有效，无需额外组件 |
| 统计查询 | MongoDB聚合 + Redis缓存 | 性能优化 |
| 收藏存储 | 独立Collection | 解耦，易扩展 |

### 3.3 依赖关系

```
代码运行 → JudgeService + TestCaseService
题目搜索 → ProblemService（新增方法）
批量导入 → TestCaseService（新增方法）
收藏功能 → 新增 FavoriteService
统计功能 → 新增 StatisticsService
```

---

## 4. 验收标准

### 4.1 功能验收

- [ ] **代码运行测试**
  - [ ] 只运行示例用例
  - [ ] 不创建提交记录
  - [ ] 频率限制生效
  - [ ] 返回详细执行结果
  
- [ ] **题目搜索**
  - [ ] 关键词搜索标题和描述
  - [ ] 难度筛选
  - [ ] 标签筛选
  - [ ] 分页正常
  
- [ ] **批量导入**
  - [ ] 单次最多100个
  - [ ] 权限校验
  - [ ] 事务处理
  
- [ ] **收藏功能**
  - [ ] 收藏/取消收藏
  - [ ] 查询收藏列表
  - [ ] 题目详情显示是否收藏
  
- [ ] **用户统计**
  - [ ] 统计数据准确
  - [ ] 按难度分组
  - [ ] 最近提交展示
  
- [ ] **系统统计**
  - [ ] 仅管理员可访问
  - [ ] 统计数据准确
  - [ ] 性能可接受

### 4.2 非功能验收

- [ ] 所有接口响应时间 < 1秒（正常情况）
- [ ] 代码运行接口单次执行 < 30秒
- [ ] 统计接口使用缓存，减少DB压力
- [ ] 频率限制准确，无法绕过
- [ ] 错误处理完善，返回友好错误信息

### 4.3 安全验收

- [ ] 代码运行接口有频率限制
- [ ] 批量导入有权限控制
- [ ] 系统统计只有管理员可访问
- [ ] 输入验证完整（防注入）

---

## 5. 技术约束

### 5.1 现有架构约束
- ✅ 使用现有的 Gin 路由框架
- ✅ 使用现有的 JWT 认证中间件
- ✅ 使用现有的 MongoDB 和 Redis
- ✅ 复用现有的 JudgeService

### 5.2 性能约束
- 代码运行：单次总时长 ≤ 30秒
- 搜索接口：响应时间 ≤ 500ms
- 统计接口：响应时间 ≤ 1秒（有缓存）
- 批量导入：单次 ≤ 100个用例

### 5.3 兼容性约束
- ✅ 不修改现有接口签名
- ✅ 不影响现有功能
- ✅ 向后兼容

---

## 6. 风险评估

### 6.1 技术风险

| 风险 | 影响 | 可能性 | 应对措施 |
|-----|------|--------|---------|
| 代码运行滥用 | 高 | 中 | 严格频率限制 + 监控 |
| 统计查询慢 | 中 | 中 | 缓存 + 索引优化 |
| 批量导入超时 | 低 | 低 | 限制数量 + 异步处理 |

### 6.2 业务风险

| 风险 | 影响 | 应对措施 |
|-----|------|---------|
| 用户频繁运行耗尽资源 | 高 | 频率限制 + 队列管理 |
| 搜索结果不准确 | 低 | 优化索引 + 分词（延后） |

---

## 7. 里程碑计划

### Week 1: 核心功能
- Day 1-2: 代码运行测试接口
- Day 3: 题目搜索接口
- Day 4-5: 测试和优化

### Week 2: 扩展功能
- Day 1-2: 批量导入 + 收藏功能
- Day 3-4: 统计功能
- Day 5: 整体测试

**总工作量**: 约 **6-7人日**

---

## 8. 确认事项

### 8.1 已确认
- ✅ 代码运行测试实现方式（复用判题逻辑）
- ✅ 题目搜索使用 MongoDB regex
- ✅ 频率限制使用 Redis
- ✅ 统计数据使用缓存优化

### 8.2 待确认
- ⏳ 代码运行的频率限制具体值（当前：每题每分钟3次）
- ⏳ 统计数据的缓存时间（当前：5-10分钟）
- ⏳ 是否需要收藏夹功能（目前只支持单纯收藏）

---

**文档状态**: ✅ 已确认，可进入架构设计阶段

**下一步**: 生成 `DESIGN_补充功能开发.md`

