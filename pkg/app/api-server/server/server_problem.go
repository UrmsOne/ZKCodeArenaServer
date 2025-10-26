/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/25 13:09
@Name: server_problem.go
@Description: 题目相关路由处理
*/

package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zk-code-arena-server/pkg/app/api-server/service"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"
	"zk-code-arena-server/pkg/utils/middleware"
)

// RegisterProblem 注册题目相关路由
func (s *Server) RegisterProblem(g *gin.RouterGroup) {
	// 每日一题（独立路由，不在problem组下）
	g.GET("/daily-problem", s.GetDailyProblem) // 获取每日推荐题目
	
	// 用户端题目路由组
	problemGroup := g.Group("/problem")
	{
		// 公开路由
		problemGroup.GET("/", s.GetProblems)          // 获取题目列表（用户端）
		problemGroup.GET("/search", s.SearchProblems) // 搜索题目
		problemGroup.GET("/:id", s.GetProblem)        // 获取题目详情
		problemGroup.GET("/:id/detail", s.GetProblemDetail) // 获取题目详情聚合信息
	}

	// 需要认证的路由
	securedGroup := problemGroup.Group("/").Use(middleware.JWTMiddleware())
	{
		securedGroup.POST("/", s.CreateProblem)      // 创建题目
		securedGroup.PUT("/:id", s.UpdateProblem)    // 更新题目
		securedGroup.DELETE("/:id", s.DeleteProblem) // 删除题目

		securedGroup.POST("/:id/run", middleware.CodeRunRateLimitMiddleware(), s.RunCode) // 运行代码测试
	}

	// 管理员专用路由组
	adminGroup := g.Group("/admin/problems").Use(middleware.JWTMiddleware(), middleware.RequireRole(models.RoleAdmin))
	{
		adminGroup.GET("/", s.GetProblemsForAdmin) // 获取题目列表（管理员端）
	}
}

// GetProblems godoc
// @Summary      获取题目列表（用户端）
// @Description  用户端获取题目列表，只显示公开已发布的题目，支持按难度、标签筛选
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Param        difficulty query string false "难度" Enums(easy, medium, hard)
// @Param        tags query []string false "标签列表"
// @Success      200 {object} map[string]interface{} "题目列表"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Failure      500 {object} map[string]interface{} "获取失败"
// @Router       /problem [get]
func (s *Server) GetProblems(c *gin.Context) {
	// 参数解析和验证
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	difficulty := c.Query("difficulty")
	tags := c.QueryArray("tags")

	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取用户ID（可选，用于获取用户提交状态）
	var userObjectID *primitive.ObjectID
	if userIDVal, exists := c.Get("user_id"); exists {
		if objID, err := primitive.ObjectIDFromHex(userIDVal.(string)); err == nil {
			userObjectID = &objID
		}
	}

	// 调用新的用户端Service方法
	ctx := c.Request.Context()
	problems, total, err := s.svc.ProblemService.GetProblemsForUser(
		ctx,
		page,
		pageSize,
		models.ProblemDifficulty(difficulty),
		tags,
		userObjectID,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取题目列表失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"problems":   problems,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetProblemsForAdmin godoc
// @Summary      获取题目列表（管理员端）
// @Description  管理员端获取题目列表，可查看所有题目包括私有和草稿状态，支持按难度、标签筛选
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Param        difficulty query string false "难度" Enums(easy, medium, hard)
// @Param        tags query []string false "标签列表"
// @Success      200 {object} map[string]interface{} "题目列表"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Failure      401 {object} map[string]interface{} "未授权访问"
// @Failure      403 {object} map[string]interface{} "权限不足"
// @Failure      500 {object} map[string]interface{} "获取失败"
// @Router       /admin/problems [get]
// @Security     BearerAuth
func (s *Server) GetProblemsForAdmin(c *gin.Context) {
	// 参数解析和验证
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	difficulty := c.Query("difficulty")
	tags := c.QueryArray("tags")

	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 调用管理员端Service方法
	ctx := c.Request.Context()
	problems, total, err := s.svc.ProblemService.GetProblemsForAdmin(
		ctx,
		page,
		pageSize,
		models.ProblemDifficulty(difficulty),
		tags,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取题目列表失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"problems":   problems,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetProblem godoc
// @Summary      获取题目详情
// @Description  根据题目ID获取完整题目信息
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        id path string true "题目ID"
// @Success      200 {object} models.Problem "题目详情"
// @Failure      400 {object} map[string]interface{} "无效的题目ID"
// @Failure      401 {object} map[string]interface{} "需要登录"
// @Failure      404 {object} map[string]interface{} "题目不存在"
// @Router       /problem/{id} [get]
func (s *Server) GetProblem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "无效的题目ID")
		return
	}

	ctx := c.Request.Context()
	problem, err := s.svc.ProblemService.GetProblemByID(ctx, id)
	if err != nil {
		utils.NotFoundResponse(c, "题目不存在")
		return
	}

	// 检查权限：非公开题目需要认证
	if !problem.IsPublic {
		_, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "需要登录才能查看此题目")
			return
		}
	}

	utils.SuccessResponse(c, problem)
}

// GetProblemDetail godoc
// @Summary      获取题目详情聚合信息
// @Description  获取题目基本信息和示例测试用例的聚合数据，用于前端题目详情页面展示
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "题目ID"
// @Success      200  {object}  utils.Response{data=models.ProblemDetailResponse}  "成功"
// @Failure      400  {object}  utils.Response  "请求参数错误"
// @Failure      401  {object}  utils.Response  "未授权访问"
// @Failure      404  {object}  utils.Response  "题目不存在"
// @Failure      500  {object}  utils.Response  "服务器内部错误"
// @Router       /problem/{id}/detail [get]
func (s *Server) GetProblemDetail(c *gin.Context) {
	// 参数验证：解析题目ID
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "无效的题目ID")
		return
	}

	// 数据操作：获取题目详情聚合信息
	ctx := c.Request.Context()
	problemDetail, err := s.svc.ProblemService.GetProblemDetail(ctx, id)
	if err != nil {
		// 业务逻辑：根据错误类型返回不同响应
		if err.Error() == "题目不存在" || err.Error() == "获取题目信息失败: mongo: no documents in result" {
			utils.NotFoundResponse(c, "题目不存在")
			return
		}
		if err.Error() == "获取示例测试用例失败: mongo: no documents in result" {
			utils.NotFoundResponse(c, "题目测试用例不存在")
			return
		}
		utils.InternalServerErrorResponse(c, "获取题目详情失败")
		return
	}

	// 权限检查：非公开题目需要认证
	if !problemDetail.Problem.IsPublic {
		_, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "需要登录才能查看此题目")
			return
		}
	}

	// 成功响应
	utils.SuccessResponse(c, problemDetail)
}

// CreateProblem godoc
// @Summary      创建题目（教师/管理员）
// @Description  创建新题目，默认状态为草稿，草稿状态不能公开
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        request body models.CreateProblemRequest true "题目信息"
// @Success      200 {object} models.Problem "创建成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误或业务规则错误"
// @Failure      401 {object} map[string]interface{} "需要登录"
// @Failure      403 {object} map[string]interface{} "权限不足"
// @Failure      500 {object} map[string]interface{} "创建失败"
// @Security     BearerAuth
// @Router       /problem [post]
func (s *Server) CreateProblem(c *gin.Context) {
	//检查权限 只有管理员和老师可以创建题目
	role, exists := c.Get("role")
	if !exists {
		utils.UnauthorizedResponse(c, "需要登录")
		return
	}
	userRole := role.(string)
	if userRole != string(models.RoleAdmin) && userRole != string(models.RoleTeacher) {
		utils.ForbiddenResponse(c, "权限不足，只有管理员和教师可以创建题目")
		return
	}
	
	// 2. 绑定和验证请求参数
	var req models.CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 业务规则预检：草稿状态不能公开
	if req.Status != nil && *req.Status == models.StatusDraft {
		if req.IsPublic != nil && *req.IsPublic == true {
			utils.BadRequestResponse(c, "草稿状态的题目不能设为公开")
			return
		}
	}

	// 4. 构建Problem对象
	problem := models.Problem{
		Title:        req.Title,
		Description:  req.Description,
		Input:        req.Input,
		Output:       req.Output,
		SampleInput:  req.SampleInput,
		SampleOutput: req.SampleOutput,
		Hint:         req.Hint,
		Source:       req.Source,
		Author:       req.Author,
		Difficulty:   req.Difficulty,
		Tags:         req.Tags,
	}

	// 5. 设置可选字段（使用指针判断是否传入）
	if req.TimeLimit != nil {
		problem.TimeLimit = *req.TimeLimit
	}
	if req.MemoryLimit != nil {
		problem.MemoryLimit = *req.MemoryLimit
	}
	if req.Status != nil {
		problem.Status = *req.Status
	}
	if req.IsPublic != nil {
		problem.IsPublic = *req.IsPublic
	}

	// 6. 设置创建者
	userID, _ := c.Get("user_id")
	problem.CreatedBy, _ = primitive.ObjectIDFromHex(userID.(string))

	// 7. 调用Service层
	ctx := c.Request.Context()
	if err := s.svc.ProblemService.CreateProblem(ctx, &problem); err != nil {
		utils.InternalServerErrorResponse(c, "创建题目失败: "+err.Error())
		return
	}

	// 8. 返回成功响应
	utils.SuccessResponse(c, problem)
}

// UpdateProblem godoc
// @Summary      更新题目（教师/管理员）
// @Description  更新题目信息（管理员或题目创建者），草稿状态不能设为公开
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        id path string true "题目ID"
// @Param        request body models.UpdateProblemRequest true "更新的题目信息"
// @Success      200 {object} models.Problem "更新成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误或业务规则错误"
// @Failure      401 {object} map[string]interface{} "需要登录"
// @Failure      403 {object} map[string]interface{} "权限不足"
// @Failure      404 {object} map[string]interface{} "题目不存在"
// @Failure      500 {object} map[string]interface{} "更新失败"
// @Security     BearerAuth
// @Router       /problem/{id} [put]
func (s *Server) UpdateProblem(c *gin.Context) {
	// 1. 检查权限：只有管理员和老师可以更新题目
	role, exists := c.Get("role")
	if !exists {
		utils.UnauthorizedResponse(c, "需要登录")
		return
	}

	userRole := role.(string)
	if userRole != string(models.RoleAdmin) && userRole != string(models.RoleTeacher) {
		utils.ForbiddenResponse(c, "权限不足，只有管理员和教师可以更新题目")
		return
	}

	// 2. 解析题目ID
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "无效的题目ID")
		return
	}

	// 3. 绑定和验证请求参数
	var req models.UpdateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 4. 业务规则预检：草稿状态不能公开
	if req.Status != nil && *req.Status == models.StatusDraft {
		if req.IsPublic != nil && *req.IsPublic == true {
			utils.BadRequestResponse(c, "草稿状态的题目不能设为公开")
			return
		}
	}

	// 5. 获取现有题目（用于权限检查）
	ctx := c.Request.Context()
	problem, err := s.svc.ProblemService.GetProblemByID(ctx, id)
	if err != nil {
		utils.NotFoundResponse(c, "题目不存在")
		return
	}

	// 6. 检查权限：只有创建者或管理员可以更新
	userID, _ := c.Get("user_id")
	currentUserID, _ := primitive.ObjectIDFromHex(userID.(string))
	if userRole != string(models.RoleAdmin) && problem.CreatedBy != currentUserID {
		utils.ForbiddenResponse(c, "只能更新自己创建的题目")
		return
	}

	// 7. 调用 Service 层更新（直接传递请求对象）
	if err := s.svc.ProblemService.UpdateProblem(ctx, id, &req); err != nil {
		utils.InternalServerErrorResponse(c, "更新题目失败: "+err.Error())
		return
	}

	// 8. 重新获取更新后的题目（用于返回）
	updatedProblem, err := s.svc.ProblemService.GetProblemByID(ctx, id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取更新后的题目失败: "+err.Error())
		return
	}

	// 9. 返回成功响应
	utils.SuccessResponse(c, updatedProblem)
}

// DeleteProblem godoc
// @Summary      删除题目（管理员）
// @Description  删除指定题目（仅管理员）
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        id path string true "题目ID"
// @Success      200 {object} map[string]interface{} "删除成功"
// @Failure      400 {object} map[string]interface{} "无效的题目ID"
// @Failure      403 {object} map[string]interface{} "权限不足"
// @Failure      500 {object} map[string]interface{} "删除失败"
// @Security     BearerAuth
// @Router       /problem/{id} [delete]
func (s *Server) DeleteProblem(c *gin.Context) {
	// 检查权限：只有管理员可以删除题目
	role, exists := c.Get("role")
	if !exists || role.(string) != string(models.RoleAdmin) {
		utils.ForbiddenResponse(c, "权限不足")
		return
	}

	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "无效的题目ID")
		return
	}

	ctx := c.Request.Context()
	if err := s.svc.ProblemService.DeleteProblem(ctx, id); err != nil {
		utils.InternalServerErrorResponse(c, "删除题目失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}

// RunCode godoc
// @Summary      运行代码测试
// @Description  在线运行代码进行测试（非提交），有频率限制
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        id path string true "题目ID"
// @Param        request body service.RunCodeRequest true "代码和语言"
// @Success      200 {object} map[string]interface{} "运行结果"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Failure      404 {object} map[string]interface{} "题目不存在"
// @Failure      500 {object} map[string]interface{} "运行失败"
// @Security     BearerAuth
// @Router       /problem/{id}/run [post]
func (s *Server) RunCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "无效的题目ID")
		return
	}

	var req service.RunCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	resp, err := s.svc.ProblemService.RunCode(ctx, id, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "运行代码失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, resp)
}

// SearchProblems godoc
// @Summary      搜索题目
// @Description  根据关键词、难度、标签搜索题目
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        keyword query string false "关键词"
// @Param        difficulty query string false "难度" Enums(easy, medium, hard)
// @Param        tags query []string false "标签列表"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Success      200 {object} map[string]interface{} "搜索结果"
// @Failure      500 {object} map[string]interface{} "搜索失败"
// @Router       /problem/search [get]
func (s *Server) SearchProblems(c *gin.Context) {
	keyword := c.Query("keyword")
	difficulty := c.Query("difficulty")
	tags := c.QueryArray("tags")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取用户ID（用于用户状态查询）
	var userObjectID *primitive.ObjectID
	if userIDVal, exists := c.Get("user_id"); exists {
		if objID, err := primitive.ObjectIDFromHex(userIDVal.(string)); err == nil {
			userObjectID = &objID
		}
	}

	ctx := c.Request.Context()
	problems, total, err := s.svc.ProblemService.SearchProblems(
		ctx,
		keyword,
		models.ProblemDifficulty(difficulty),
		tags,
		page,
		pageSize,
		userObjectID,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "搜索题目失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"problems":   problems,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetDailyProblem godoc
// @Summary      获取每日推荐题目
// @Description  获取当日推荐的题目，全局统一推荐
// @Tags         题目
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "每日推荐题目"
// @Failure      500 {object} map[string]interface{} "获取失败"
// @Router       /daily-problem [get]
func (s *Server) GetDailyProblem(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. 检查是否有登录用户
	var userObjectID *primitive.ObjectID
	if userIDInterface, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userIDInterface.(string); ok {
			if id, err := primitive.ObjectIDFromHex(userIDStr); err == nil {
				userObjectID = &id
			}
		}
	}

	// 2. 获取每日推荐题目（已经包含用户状态）
	dailyProblem, err := s.svc.DailyProblemService.GetDailyProblem(ctx, userObjectID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取每日推荐题目失败: "+err.Error())
		return
	}

	// 3. 返回结果
	utils.SuccessResponse(c, gin.H{
		"daily_problem": dailyProblem,
		"message":       "每日推荐题目获取成功",
	})
}
