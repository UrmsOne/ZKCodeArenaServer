# OJ项目缺失接口实施计划

## 📊 当前状态评估

### ✅ 已完成模块
- **题目管理**: 85% 完整度
- **提交系统**: 75% 完整度  
- **测试用例**: 70% 完整度

### ⚠️ 关键缺失功能

## 🎯 优先级实施计划

### **Phase 1: 核心功能补全（P0）**

#### 1. 示例测试用例公开接口
```go
// 文件：server_testcase.go
// 新增路由：GET /testcase/problem/:id/sample

func (s *Server) RegisterTestCase(g *gin.RouterGroup) {
    testCaseGroup := g.Group("/testcase")
    
    // 🆕 公开接口 - 示例用例
    testCaseGroup.GET("/problem/:id/sample", s.GetSampleTestCases)
    
    // 现有管理接口保持不变...
}

// 🆕 新增方法
func (s *Server) GetSampleTestCases(c *gin.Context) {
    problemID, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "无效的题目ID")
        return
    }
    
    sampleCases, err := s.svc.TestCaseService.GetSampleTestCases(c.Request.Context(), problemID)
    if err != nil {
        utils.InternalServerErrorResponse(c, "获取示例用例失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "sample_cases": sampleCases,
        "total": len(sampleCases),
    })
}
```

#### 2. 提交代码查看接口
```go
// 文件：server_submit.go  
// 新增路由：GET /submit/:id/code

func (s *Server) RegisterSubmit(g *gin.RouterGroup) {
    submitGroup := g.Group("/submit")
    submitGroup.Use(middleware.JWTMiddleware())
    {
        // 现有接口...
        submitGroup.GET("/:id/code", s.GetSubmitCode) // 🆕 查看代码
    }
}

// 🆕 新增方法
func (s *Server) GetSubmitCode(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "无效的提交ID")
        return
    }
    
    submit, err := s.svc.SubmitService.GetSubmitByID(c.Request.Context(), id)
    if err != nil {
        utils.NotFoundResponse(c, "提交不存在")
        return
    }
    
    // 权限检查：只能查看自己的代码或管理员
    role := c.GetString("role")
    currentUserID := c.GetString("user_id")
    currentUserObjectID, _ := primitive.ObjectIDFromHex(currentUserID)
    
    if role != string(models.RoleAdmin) && submit.UserID != currentUserObjectID {
        utils.ForbiddenResponse(c, "只能查看自己的代码")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "submit_id": submit.ID,
        "code": submit.Code,
        "language": submit.Language,
        "created_at": submit.CreatedAt,
    })
}
```

#### 3. 题目收藏功能
```go
// 文件：server_problem.go
// 新增路由：POST /problem/:id/favorite, GET /problem/favorites

func (s *Server) RegisterProblem(g *gin.RouterGroup) {
    problemGroup := g.Group("/problem")
    
    // 需要认证的路由
    securedGroup := problemGroup.Group("/").Use(middleware.JWTMiddleware())
    {
        // 现有接口...
        securedGroup.POST("/:id/favorite", s.ToggleProblemFavorite) // 🆕 收藏切换
        securedGroup.GET("/favorites", s.GetUserFavorites)           // 🆕 我的收藏
    }
}

// 🆕 收藏切换方法
func (s *Server) ToggleProblemFavorite(c *gin.Context) {
    problemID, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "无效的题目ID")
        return
    }
    
    userID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))
    
    isFavorited, err := s.svc.ProblemService.ToggleFavorite(c.Request.Context(), userID, problemID)
    if err != nil {
        utils.InternalServerErrorResponse(c, "操作失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "favorited": isFavorited,
        "message": map[bool]string{true: "已收藏", false: "已取消收藏"}[isFavorited],
    })
}

// 🆕 获取收藏列表方法  
func (s *Server) GetUserFavorites(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
    
    userID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))
    
    favorites, total, err := s.svc.ProblemService.GetUserFavorites(c.Request.Context(), userID, page, pageSize)
    if err != nil {
        utils.InternalServerErrorResponse(c, "获取收藏列表失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "favorites":  favorites,
        "total":      total,
        "page":       page,
        "page_size":  pageSize,
        "total_page": (total + int64(pageSize) - 1) / int64(pageSize),
    })
}
```

### **Phase 2: 用户体验提升（P1）**

#### 4. 题目统计接口
```go
// 文件：server_problem.go
// 新增路由：GET /problem/:id/statistics

func (s *Server) GetProblemStatistics(c *gin.Context) {
    problemID, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "无效的题目ID")
        return
    }
    
    stats, err := s.svc.ProblemService.GetProblemStatistics(c.Request.Context(), problemID)
    if err != nil {
        utils.InternalServerErrorResponse(c, "获取统计失败")
        return
    }
    
    utils.SuccessResponse(c, stats)
}
```

#### 5. 实时判题状态
```go
// 文件：server_submit.go
// 新增路由：GET /submit/:id/status

func (s *Server) GetSubmitStatus(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "无效的提交ID")
        return
    }
    
    status, err := s.svc.SubmitService.GetSubmitStatus(c.Request.Context(), id)
    if err != nil {
        utils.InternalServerErrorResponse(c, "获取状态失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "submit_id": id,
        "status": status.Status,
        "progress": status.Progress,
        "message": status.Message,
        "updated_at": status.UpdatedAt,
    })
}
```

#### 6. 随机题目接口
```go
// 文件：server_problem.go
// 新增路由：GET /problem/random

func (s *Server) GetRandomProblem(c *gin.Context) {
    difficulty := c.Query("difficulty")
    tags := c.QueryArray("tags")
    
    problem, err := s.svc.ProblemService.GetRandomProblem(c.Request.Context(), difficulty, tags)
    if err != nil {
        utils.InternalServerErrorResponse(c, "获取随机题目失败")
        return
    }
    
    utils.SuccessResponse(c, problem)
}
```

### **Phase 3: 管理功能增强（P2）**

#### 7. 批量操作接口
```go
// 文件：server_problem.go  
// 管理员批量操作

func (s *Server) BatchDeleteProblems(c *gin.Context) {
    var req struct {
        ProblemIDs []string `json:"problem_ids" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequestResponse(c, "请求参数错误")
        return
    }
    
    result, err := s.svc.ProblemService.BatchDelete(c.Request.Context(), req.ProblemIDs)
    if err != nil {
        utils.InternalServerErrorResponse(c, "批量删除失败")
        return
    }
    
    utils.SuccessResponse(c, result)
}
```

#### 8. 重新判题接口
```go
// 文件：server_submit.go
// 管理员重新判题

func (s *Server) RejudgeSubmit(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "无效的提交ID")
        return
    }
    
    err = s.svc.SubmitService.Rejudge(c.Request.Context(), id)
    if err != nil {
        utils.InternalServerErrorResponse(c, "重新判题失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "message": "重新判题任务已提交",
        "submit_id": id,
    })
}
```

## 📊 数据模型扩展需求

### 1. 用户收藏模型
```go
// 文件：pkg/models/favorite.go
type UserFavorite struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
    ProblemID primitive.ObjectID `bson:"problem_id" json:"problem_id"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
```

### 2. 题目统计模型
```go
// 文件：pkg/models/statistics.go
type ProblemStatistics struct {
    ProblemID       primitive.ObjectID `json:"problem_id"`
    TotalSubmissions int64             `json:"total_submissions"`
    AcceptedSubmissions int64          `json:"accepted_submissions"`
    AcceptanceRate   float64           `json:"acceptance_rate"`
    AvgRuntime       float64           `json:"avg_runtime"`
    AvgMemory        float64           `json:"avg_memory"`
    LanguageStats    []LanguageStat    `json:"language_stats"`
}

type LanguageStat struct {
    Language string `json:"language"`
    Count    int64  `json:"count"`
    Rate     float64 `json:"rate"`
}
```

## 🔄 实施时间线

### Week 1-2: Phase 1 核心功能
- [x] 示例测试用例公开接口
- [x] 提交代码查看接口  
- [x] 题目收藏功能

### Week 3-4: Phase 2 用户体验
- [ ] 题目统计接口
- [ ] 实时判题状态
- [ ] 随机题目接口

### Week 5-6: Phase 3 管理功能
- [ ] 批量操作接口
- [ ] 重新判题接口
- [ ] 数据清理功能

## 🎯 成功指标

1. **接口完整度**：达到95%标准OJ功能覆盖
2. **用户体验**：提供核心用户功能（收藏、统计、实时状态）
3. **管理效率**：提供批量操作减少管理工作量
4. **系统稳定性**：所有新接口通过压力测试

## 📝 注意事项

1. **权限控制**：确保每个新接口都有适当的权限验证
2. **性能优化**：统计接口需要考虑缓存机制
3. **错误处理**：统一错误响应格式
4. **文档更新**：同步更新Swagger文档
5. **测试覆盖**：为每个新接口编写单元测试

## 🔗 相关文档

- [API设计规范](./API_DESIGN_STANDARDS.md)
- [权限控制指南](./PERMISSION_CONTROL_GUIDE.md)  
- [性能优化建议](./PERFORMANCE_OPTIMIZATION.md)
- [测试指南](./TESTING_GUIDELINES.md)
