# DESIGN - 每日一题与题目统计功能架构设计

## 整体架构图

```mermaid
graph TB
    subgraph "API Layer (Server)"
        A1[DailyProblemAPI]
        A2[DifficultyStatsAPI] 
        A3[ProblemAPI - Enhanced]
        A4[SubmitAPI - Enhanced]
        A5[TestCaseAPI - Enhanced]
    end
    
    subgraph "Service Layer"
        B1[DailyProblemService]
        B2[StatisticsService - Enhanced]
        B3[ProblemService - Enhanced]
        B4[SubmitService - Enhanced]
        B5[TestCaseService - Enhanced]
    end
    
    subgraph "Repository Layer - NEW"
        R1[BaseRepository - 通用CRUD]
        R2[ProblemRepository]
        R3[SubmitRepository]
        R4[TestCaseRepository]
    end
    
    subgraph "Cache Layer"
        C1[DailyRecommendationCache]
        C2[DifficultyStatsCache]
    end
    
    subgraph "Database Layer"
        D1[(problems)]
        D2[(submits)]
        D3[(testcases)]
    end
    
    %% API to Service connections
    A1 --> B1
    A2 --> B2
    A3 --> B3
    A4 --> B4
    A5 --> B5
    
    %% Service to Repository connections
    B1 --> R2
    B2 --> R2
    B3 --> R2
    B4 --> R3
    B5 --> R4
    
    %% Repository inheritance
    R2 --> R1
    R3 --> R1
    R4 --> R1
    
    %% Service to Cache connections  
    B1 --> C1
    B2 --> C2
    
    %% Repository to Database connections
    R1 --> D1
    R1 --> D2
    R1 --> D3
    
    %% Cache to Database fallback
    C1 -.-> D1
    C2 -.-> D1
```

## 核心组件设计

### 1. 每日一题模块

#### DailyProblemService
```go
type DailyProblemService struct {
    problemService *ProblemService
    cache         map[string]*models.ProblemList // 简单内存缓存
    lastUpdate    time.Time
}

// 核心方法
func (s *DailyProblemService) GetDailyProblem(ctx context.Context, userID *primitive.ObjectID) (*models.ProblemList, error)
func (s *DailyProblemService) refreshDailyRecommendation(ctx context.Context) error
```

#### 推荐策略
```mermaid
flowchart TD
    A[开始] --> B{检查缓存}
    B -->|命中且未过期| C[返回缓存结果]
    B -->|未命中或已过期| D[查询数据库]
    D --> E[筛选已发布公开题目]
    E --> F[随机选择一道题目]
    F --> G[更新缓存]
    G --> H[返回结果]
    C --> I[结束]
    H --> I
```

### 2. 难度统计模块

#### 轻量级统计API
```go
type DifficultyStats struct {
    Easy   int64 `json:"easy"`
    Medium int64 `json:"medium"`
    Hard   int64 `json:"hard"`
    Total  int64 `json:"total"`
}

// StatisticsService 新增方法
func (s *StatisticsService) GetDifficultyStats(ctx context.Context) (*DifficultyStats, error)
```

### 3. Repository层设计 (核心架构优化)

#### BaseRepository (通用数据访问层)
```go
type BaseRepository struct {
    db *mongo.Database
}

// 轻量通用更新函数
func (r *BaseRepository) BuildUpdateSet(obj interface{}) (bson.M, error) {
    data, err := bson.Marshal(obj)
    if err != nil {
        return nil, err
    }
    var update bson.M
    if err := bson.Unmarshal(data, &update); err != nil {
        return nil, err
    }
    if len(update) == 0 {
        return nil, nil
    }
    return bson.M{"$set": update}, nil
}

// 带时间戳的通用更新
func (r *BaseRepository) BuildUpdateSetWithTime(obj interface{}) (bson.M, error) {
    updateDoc, err := r.BuildUpdateSet(obj)
    if err != nil {
        return nil, err
    }
    if updateDoc == nil {
        return bson.M{"$set": bson.M{"updatedAt": time.Now()}}, nil
    }
    updateDoc["$set"].(bson.M)["updatedAt"] = time.Now()
    return updateDoc, nil
}

// 通用更新方法
func (r *BaseRepository) UpdateOne(ctx context.Context, collection string, filter bson.M, updateData interface{}) (*mongo.UpdateResult, error) {
    updateDoc, err := r.BuildUpdateSetWithTime(updateData)
    if err != nil {
        return nil, err
    }
    if updateDoc == nil {
        return nil, fmt.Errorf("no fields to update")
    }
    
    coll := r.db.Collection(collection)
    return coll.UpdateOne(ctx, filter, updateDoc)
}
```

#### ProblemRepository
```go
type ProblemRepository struct {
    *BaseRepository
}

func (r *ProblemRepository) UpdateProblem(ctx context.Context, problemID primitive.ObjectID, req *models.UpdateProblemRequest) error {
    // 业务规则验证
    if err := r.validateProblemUpdate(req); err != nil {
        return err
    }
    
    // 使用通用更新方法
    _, err := r.UpdateOne(ctx, "problems", bson.M{"_id": problemID}, req)
    return err
}

func (r *ProblemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Problem, error) {
    // 具体的查询实现
}

func (r *ProblemRepository) GetPublicProblems(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*models.Problem, error) {
    // 具体的查询实现
}
```

#### SubmitRepository & TestCaseRepository
```go
type SubmitRepository struct {
    *BaseRepository
}

type TestCaseRepository struct {
    *BaseRepository
}

// 都继承BaseRepository的通用CRUD方法
// 各自实现特定的业务查询方法
```

---

## 分层设计和核心组件

### API层增强
```go
// 新增每日一题API
func (s *Server) GetDailyProblem(c *gin.Context)

// 新增轻量级统计API  
func (s *Server) GetProblemDifficultyStats(c *gin.Context)

// 现有API保持不变，内部逻辑优化
func (s *Server) UpdateProblem(c *gin.Context)    // 使用新的更新逻辑
func (s *Server) UpdateSubmit(c *gin.Context)     // 使用新的更新逻辑  
func (s *Server) UpdateTestCase(c *gin.Context)   // 使用新的更新逻辑
```

### Service层重构
```go
// Problem Service 优化
func (s *ProblemService) UpdateProblem(ctx context.Context, problemID primitive.ObjectID, req *models.UpdateProblemRequest) error {
    // 使用 CommonUpdateService.BuildUpdateDocument
    updateDoc, err := s.commonUpdateService.BuildUpdateDocument(req)
    if err != nil {
        return err
    }
    
    return s.commonUpdateService.UpdateDocument(ctx, "problems", 
        bson.M{"_id": problemID}, updateDoc)
}
```

### 数据流向图
```mermaid
sequenceDiagram
    participant C as Client
    participant A as API Layer
    participant S as Service Layer
    participant Cache as Cache
    participant DB as Database
    participant U as UpdateUtils
    
    Note over C,U: 每日一题流程
    C->>A: GET /daily-problem
    A->>S: GetDailyProblem()
    S->>Cache: 检查缓存
    alt 缓存命中
        Cache-->>S: 返回缓存数据
    else 缓存未命中
        S->>DB: 查询题目列表
        DB-->>S: 返回题目数据
        S->>S: 随机选择题目
        S->>Cache: 更新缓存
    end
    S-->>A: 返回题目信息
    A-->>C: JSON响应
    
    Note over C,U: 更新操作流程
    C->>A: PUT /problem/{id}
    A->>S: UpdateProblem()
    S->>U: BuildUpdateDocument()
    U->>U: BSON序列化处理
    U-->>S: 返回$set文档
    S->>DB: UpdateOne操作
    DB-->>S: 更新结果
    S-->>A: 操作结果
    A-->>C: JSON响应
```

---

## 模块依赖关系图

```mermaid
graph LR
    subgraph "新增模块"
        DailyProblem[DailyProblemService]
        CommonUpdate[CommonUpdateService]
        UpdateBuilder[UpdateDocumentBuilder]
    end
    
    subgraph "增强模块"
        Problem[ProblemService]
        Submit[SubmitService] 
        TestCase[TestCaseService]
        Statistics[StatisticsService]
    end
    
    subgraph "现有模块"
        Auth[AuthService]
        Utils[UtilsPackage]
        Middleware[Middleware]
    end
    
    %% 依赖关系
    DailyProblem --> Problem
    DailyProblem --> Statistics
    
    Problem --> CommonUpdate
    Submit --> CommonUpdate
    TestCase --> CommonUpdate
    
    CommonUpdate --> UpdateBuilder
    CommonUpdate --> Utils
    
    %% 现有依赖保持不变
    Problem --> Auth
    Submit --> Auth
    TestCase --> Auth
    
    DailyProblem --> Middleware
```

---

## 接口契约定义

### 1. 每日一题API

```yaml
GET /api/daily-problem
Summary: 获取每日推荐题目
Parameters:
  - None (用户信息从JWT获取)
Response:
  200:
    content:
      application/json:
        schema:
          type: object
          properties:
            code:
              type: integer
              example: 200
            message:
              type: string
              example: "success"
            data:
              $ref: '#/components/schemas/ProblemList'
            extra:
              type: object
              properties:
                date:
                  type: string
                  format: date
                  example: "2025-10-25"
                cache_hit:
                  type: boolean
                  example: true
```

### 2. 难度统计API

```yaml
GET /api/problems/difficulty-stats
Summary: 获取题目难度分布统计
Response:
  200:
    content:
      application/json:
        schema:
          type: object
          properties:
            code:
              type: integer
              example: 200
            message:
              type: string
              example: "success"
            data:
              type: object
              properties:
                easy:
                  type: integer
                  example: 45
                medium:
                  type: integer
                  example: 38
                hard:
                  type: integer
                  example: 17
                total:
                  type: integer
                  example: 100
```

### 3. 通用更新工具接口

```go
// UpdateDocumentBuilder 接口定义
type UpdateDocumentBuilder interface {
    Build(data interface{}) (bson.M, error)
    BuildWithCustomFields(data interface{}, customFields bson.M) (bson.M, error)
}

// CommonUpdateService 接口定义
type CommonUpdateService interface {
    UpdateDocument(ctx context.Context, collection string, filter bson.M, updateData interface{}) (*mongo.UpdateResult, error)
    UpdateDocumentWithValidation(ctx context.Context, collection string, filter bson.M, updateData interface{}, validator func(interface{}) error) (*mongo.UpdateResult, error)
}
```

---

## 异常处理策略

### 1. 每日一题异常处理
```go
// 分层异常处理
func (s *DailyProblemService) GetDailyProblem(ctx context.Context, userID *primitive.ObjectID) (*models.ProblemList, error) {
    // 1. 缓存异常 - 降级到数据库查询
    // 2. 数据库异常 - 返回默认题目或错误
    // 3. 无可用题目 - 返回特定错误码
    // 4. 超时处理 - 设置合理超时时间
}

// 错误类型定义
var (
    ErrNoDailyProblem = errors.New("no available daily problem")
    ErrCacheFailure   = errors.New("cache operation failed")
    ErrDatabaseFailure= errors.New("database operation failed")
)
```

### 2. 更新操作异常处理
```go
// 更新文档构建异常
func (u *UpdateDocumentBuilder) Build(data interface{}) (bson.M, error) {
    // 1. 类型检查异常
    // 2. BSON序列化异常  
    // 3. 字段解析异常
    // 4. 空文档异常 (所有字段都是空值)
}

// 数据库更新异常
func (s *CommonUpdateService) UpdateDocument(...) (*mongo.UpdateResult, error) {
    // 1. 连接异常
    // 2. 权限异常
    // 3. 文档不存在异常
    // 4. 并发更新异常
}
```

### 3. 统计API异常处理
```go
// 统计查询异常
func (s *StatisticsService) GetDifficultyStats(ctx context.Context) (*DifficultyStats, error) {
    // 1. 数据库连接异常 - 返回缓存数据
    // 2. 查询超时异常 - 设置超时限制
    // 3. 数据不一致异常 - 记录日志并返回近似值
}
```

---

## 性能优化策略

### 1. 缓存策略
```go
// 分层缓存设计
type CacheLayer struct {
    // L1: 内存缓存 (热点数据)
    memoryCache map[string]interface{}
    
    // L2: 可扩展到Redis (可选)
    // redisClient *redis.Client
}

// 缓存更新策略
// - 每日一题: 24小时TTL，定时刷新
// - 难度统计: 1小时TTL，写入时失效
```

### 2. 数据库优化
```go
// 查询优化
// 1. 复用现有索引
// 2. 使用聚合查询减少网络传输
// 3. 限制返回字段大小

// 更新优化  
// 1. 批量更新支持
// 2. 乐观锁防止并发冲突
// 3. 部分更新减少网络传输
```

### 3. 并发控制
```go
// 每日推荐生成的并发控制
var dailyRecommendationMutex sync.RWMutex

// 更新操作的并发安全
// 使用MongoDB的原子更新操作
// 避免读-修改-写的竞态条件
```

---

## 设计原则验证

### ✅ 架构一致性
- 保持现有Server→Service→Database三层架构
- 新增组件遵循相同的依赖注入模式
- 接口设计符合现有RESTful规范

### ✅ 接口完整性
- 每日一题API提供完整的题目信息
- 难度统计API返回结构化数据
- 通用更新工具支持扩展和定制

### ✅ 系统无冲突
- 新增缓存不影响现有数据流
- 通用更新工具可选择性使用
- 保持现有API接口签名不变

### ✅ 可行性验证
- 基于现有Go和MongoDB技术栈
- 利用BSON反射机制实现通用更新
- 缓存策略简单可靠，易于实现

**架构设计完成** ✅  
**模块依赖关系清晰** ✅  
**接口契约定义完整** ✅  
**异常处理策略完善** ✅
