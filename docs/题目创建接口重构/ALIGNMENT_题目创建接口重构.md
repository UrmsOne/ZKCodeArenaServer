# 题目创建接口重构 - 对齐文档

## 📋 原始需求

对添加题目逻辑进行重构：
1. 添加status的状态变量
2. 按最规范化开发该接口
3. status默认为草稿

---

## 🎯 项目特性规范

### 当前项目架构

**技术栈：**
- 语言：Go
- Web框架：Gin
- 数据库：MongoDB
- ORM：mongo-driver
- 文档：Swagger

**项目结构：**
```
pkg/
├── models/          # 数据模型
│   ├── problem.go   # Problem结构体，已有ProblemStatus定义
│   └── requests.go  # 请求/响应模型
├── app/api-server/
│   ├── server/      # HTTP Handler层
│   │   └── server_problem.go  # 题目相关接口
│   └── service/     # 业务逻辑层
│       └── service_problem.go # 题目服务逻辑
└── utils/           # 工具函数
    └── response.go  # 统一响应格式
```

**现有代码规范：**
1. 使用Swagger注释
2. 权限检查在Handler层
3. 业务逻辑在Service层
4. 统一错误响应格式
5. 使用binding标签验证

---

## 📊 现状分析

### 当前实现问题

#### 1. **Handler层（server_problem.go:175-206）**
```go
func (s *Server) CreateProblem(c *gin.Context) {
    // ❌ 直接绑定整个Problem结构体
    var problem models.Problem
    if err := c.ShouldBindJSON(&problem); err != nil {
        utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
        return
    }
    
    // ❌ 缺少字段验证和默认值设置
    problem.CreatedBy, _ = primitive.ObjectIDFromHex(userID.(string))
    
    // ❌ 直接调用service，未做输入规范化
    if err := s.svc.ProblemService.CreateProblem(ctx, &problem); err != nil {
        utils.InternalServerErrorResponse(c, "创建题目失败: "+err.Error())
        return
    }
}
```

**问题清单：**
- ❌ 缺少专门的CreateProblemRequest结构
- ❌ Status、IsPublic等字段由前端直接传入，不安全
- ❌ 缺少字段合法性验证
- ❌ 错误处理不够细致
- ❌ 缺少业务规则校验

#### 2. **Service层（service_problem.go:37-56）**
```go
func (s *ProblemService) CreateProblem(ctx context.Context, problem *models.Problem) error {
    problem.ID = primitive.NewObjectID()
    problem.CreatedAt = time.Now()
    problem.UpdatedAt = time.Now()
    problem.ACCount = 0
    problem.SubmitCount = 0
    
    // ✅ 新增了默认值设置
    if problem.Status == "" {
        problem.Status = models.StatusPublished // ❌ 当前是Published，需改为Draft
    }
    if problem.Status == models.StatusPublished && !problem.IsPublic {
        problem.IsPublic = true
    }
    
    collection := utils.GetCollection("problems")
    _, err := collection.InsertOne(ctx, problem)
    return err
}
```

**问题清单：**
- ❌ Status默认值是Published而非Draft
- ❌ 缺少业务规则验证（如必填字段）
- ❌ 缺少权限相关逻辑（创建者与IsPublic关系）
- ❌ 未返回创建结果的详细信息

#### 3. **Model层（problem.go:33-56）**
```go
type Problem struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Title        string             `bson:"title" json:"title" binding:"required"`
    Description  string             `bson:"description" json:"description" binding:"required"`
    // ... 其他字段
    Status       ProblemStatus      `bson:"status" json:"status"`
    IsPublic     bool               `bson:"is_public" json:"is_public"`
    // ...
}
```

**问题清单：**
- ❌ Status字段没有binding标签，缺少验证
- ❌ 缺少CreateProblemRequest专用结构
- ❌ TimeLimit、MemoryLimit等没有默认值和范围限制

---

## 🔍 需求理解

### 核心需求拆解

1. **状态管理规范化**
   - Status默认值：Draft（草稿）
   - Status状态流转：Draft → Published → Archived
   - 状态变更需要权限控制

2. **接口规范化**
   - 请求模型：专用CreateProblemRequest
   - 响应模型：统一格式
   - 验证规则：完整的字段验证
   - 错误处理：细化错误类型

3. **业务规则**
   - 草稿状态的题目自动设为非公开
   - 只有已发布的题目才能设为公开
   - 创建者自动设置
   - 默认值合理设置

---

## ❓ 疑问澄清（需要确认）

### 关键决策点

#### 1. **状态流转规则**
**问题：** Status的状态变更规则是否需要限制？

**选项：**
- A. 自由变更（Draft ↔ Published ↔ Archived）
- B. 单向流转（Draft → Published → Archived）
- C. 有条件变更（如Published需满足测试用例等条件）

**建议：** 选项A（先实现简单规则，后续可扩展）

#### 2. **IsPublic与Status的关系**
**问题：** IsPublic和Status的关系如何定义？

**选项：**
- A. 独立控制（Status和IsPublic互不影响）
- B. 关联控制（只有Published状态才能设为Public）
- C. 自动关联（Published自动Public，Draft自动Private）

**建议：** 选项B（已发布才能公开，更合理）

**规则：**
```
Draft     → IsPublic = false（强制）
Published → IsPublic = 可选（由创建者决定）
Archived  → IsPublic = 可选（保持原状态）
```

#### 3. **默认值设置**
**问题：** 创建题目时哪些字段应该有默认值？

**建议默认值：**
```go
Status:       StatusDraft      // 默认草稿
IsPublic:     false            // 默认私有
ACCount:      0                // 通过次数
SubmitCount:  0                // 提交次数
TimeLimit:    1000             // 1秒（如果未传）
MemoryLimit:  256              // 256MB（如果未传）
Tags:         []               // 空数组
```

#### 4. **验证规则**
**问题：** 哪些字段需要验证？验证规则是什么？

**建议验证规则：**
```go
Title:        required, min=1, max=200
Description:  required, min=10
Difficulty:   required, oneof=easy medium hard
TimeLimit:    min=100, max=10000  (0.1s-10s)
MemoryLimit:  min=32, max=1024    (32MB-1GB)
Tags:         max=10              (最多10个标签)
```

#### 5. **权限控制**
**问题：** 创建题目的权限如何控制？

**当前逻辑：** Admin和Teacher可以创建
**是否需要：**
- 限制Teacher创建的题目数量？
- 区分Admin和Teacher创建的题目权限？

**建议：** 保持当前逻辑，Admin和Teacher都可以创建，暂不限制数量

---

## 📝 边界确认

### 任务范围

**包含：**
✅ 创建CreateProblemRequest/Response模型
✅ 重构CreateProblem接口（Handler + Service）
✅ 添加完整的字段验证
✅ 实现Status默认为Draft
✅ 规范化错误处理
✅ 更新Swagger文档
✅ 添加详细日志

**不包含：**
❌ UpdateProblem接口重构（单独任务）
❌ 题目发布工作流（状态变更接口）
❌ 测试用例管理
❌ 数据库迁移脚本
❌ 前端代码修改

### 兼容性要求

- ⚠️ **接口变更影响：** 需要前端配合修改（请求结构变化）
- ✅ **数据库兼容：** 新代码需兼容现有数据
- ✅ **向后兼容：** 现有题目数据不受影响

---

## 🎯 验收标准

1. **功能验收**
   - [ ] 创建题目时Status默认为Draft
   - [ ] 创建题目时IsPublic默认为false
   - [ ] Draft状态的题目不能设为公开
   - [ ] 所有必填字段验证通过
   - [ ] 字段范围验证正确
   - [ ] 错误信息清晰准确

2. **代码质量**
   - [ ] 遵循项目现有代码规范
   - [ ] 添加完整的Swagger注释
   - [ ] 关键逻辑有日志记录
   - [ ] 错误处理规范

3. **测试验收**
   - [ ] 正常创建题目（草稿）
   - [ ] 验证规则测试（必填、范围）
   - [ ] 权限测试（Teacher、Admin）
   - [ ] 默认值测试

---

## 🔄 下一步行动

等待确认以下关键决策：
1. Status状态流转规则（建议选项A）
2. IsPublic与Status关系（建议选项B）
3. 默认值设置（见建议）
4. 验证规则（见建议）

**确认后进入：** Architect阶段（架构设计）

