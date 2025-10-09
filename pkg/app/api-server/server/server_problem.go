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
	problemGroup := g.Group("/problem")
	{
		// 公开路由
		problemGroup.GET("/", s.GetProblems)          // 获取题目列表
		problemGroup.GET("/search", s.SearchProblems) // 搜索题目
		problemGroup.GET("/:id", s.GetProblem)        // 获取题目详情
	}

	// 需要认证的路由
	securedGroup := problemGroup.Group("/").Use(middleware.JWTMiddleware())
	{
		securedGroup.POST("/", s.CreateProblem)      // 创建题目
		securedGroup.PUT("/:id", s.UpdateProblem)    // 更新题目
		securedGroup.DELETE("/:id", s.DeleteProblem) // 删除题目

		securedGroup.POST("/:id/run", middleware.CodeRunRateLimitMiddleware(), s.RunCode) // 运行代码测试
	}
}

// GetProblems godoc
// @Summary      获取题目列表
// @Description  分页获取题目列表，支持按难度、标签筛选
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Param        difficulty query string false "难度" Enums(easy, medium, hard)
// @Param        tags query []string false "标签列表"
// @Param        include_private query boolean false "包含私有题目（需教师或管理员权限）" default(false)
// @Success      200 {object} map[string]interface{} "题目列表"
// @Failure      500 {object} map[string]interface{} "获取失败"
// @Router       /problem [get]
func (s *Server) GetProblems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	difficulty := c.Query("difficulty")
	tags := c.QueryArray("tags")
	includePrivate := c.DefaultQuery("include_private", "false") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	roleVal, _ := c.Get("role")
	role := models.UserRole("")
	if roleVal != nil {
		role = models.UserRole(roleVal.(string))
	}

	if includePrivate && role != models.RoleAdmin && role != models.RoleTeacher {
		includePrivate = false
	}

	var userObjectID *primitive.ObjectID
	if includePrivate {
		if userIDVal, exists := c.Get("user_id"); exists {
			if objID, err := primitive.ObjectIDFromHex(userIDVal.(string)); err == nil {
				userObjectID = &objID
			}
		}
	}

	ctx := c.Request.Context()
	problems, total, err := s.svc.ProblemService.GetProblems(
		ctx,
		page,
		pageSize,
		models.ProblemDifficulty(difficulty),
		tags,
		includePrivate,
		role,
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

// CreateProblem godoc
// @Summary      创建题目（教师/管理员）
// @Description  创建新题目（仅管理员和教师）
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        request body models.Problem true "题目信息"
// @Success      200 {object} models.Problem "创建成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Failure      401 {object} map[string]interface{} "需要登录"
// @Failure      403 {object} map[string]interface{} "权限不足"
// @Failure      500 {object} map[string]interface{} "创建失败"
// @Security     BearerAuth
// @Router       /problem [post]
func (s *Server) CreateProblem(c *gin.Context) {
	// 检查权限：只有管理员和老师可以创建题目
	role, exists := c.Get("role")
	if !exists {
		utils.UnauthorizedResponse(c, "需要登录")
		return
	}

	userRole := role.(string)
	if userRole != string(models.RoleAdmin) && userRole != string(models.RoleTeacher) {
		utils.ForbiddenResponse(c, "权限不足")
		return
	}

	var problem models.Problem
	if err := c.ShouldBindJSON(&problem); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置创建者
	userID, _ := c.Get("user_id")
	problem.CreatedBy, _ = primitive.ObjectIDFromHex(userID.(string))

	ctx := c.Request.Context()
	if err := s.svc.ProblemService.CreateProblem(ctx, &problem); err != nil {
		utils.InternalServerErrorResponse(c, "创建题目失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, problem)
}

// UpdateProblem godoc
// @Summary      更新题目（教师/管理员）
// @Description  更新题目信息（管理员或题目创建者）
// @Tags         题目
// @Accept       json
// @Produce      json
// @Param        id path string true "题目ID"
// @Param        request body models.Problem true "更新的题目信息"
// @Success      200 {object} models.Problem "更新成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Failure      401 {object} map[string]interface{} "需要登录"
// @Failure      403 {object} map[string]interface{} "权限不足"
// @Failure      404 {object} map[string]interface{} "题目不存在"
// @Failure      500 {object} map[string]interface{} "更新失败"
// @Security     BearerAuth
// @Router       /problem/{id} [put]
func (s *Server) UpdateProblem(c *gin.Context) {
	// 检查权限：只有管理员和老师可以更新题目
	role, exists := c.Get("role")
	if !exists {
		utils.UnauthorizedResponse(c, "需要登录")
		return
	}

	userRole := role.(string)
	if userRole != string(models.RoleAdmin) && userRole != string(models.RoleTeacher) {
		utils.ForbiddenResponse(c, "权限不足")
		return
	}

	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "无效的题目ID")
		return
	}

	var updateReq models.Problem
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	problem, err := s.svc.ProblemService.GetProblemByID(ctx, id)
	if err != nil {
		utils.NotFoundResponse(c, "题目不存在")
		return
	}

	// 检查权限：只有创建者或管理员可以更新
	userID, _ := c.Get("user_id")
	currentUserID, _ := primitive.ObjectIDFromHex(userID.(string))
	if userRole != string(models.RoleAdmin) && problem.CreatedBy != currentUserID {
		utils.ForbiddenResponse(c, "只能更新自己创建的题目")
		return
	}

	// 更新题目信息
	problem.Title = updateReq.Title
	problem.Description = updateReq.Description
	problem.Input = updateReq.Input
	problem.Output = updateReq.Output
	problem.SampleInput = updateReq.SampleInput
	problem.SampleOutput = updateReq.SampleOutput
	problem.Hint = updateReq.Hint
	problem.Source = updateReq.Source
	problem.Author = updateReq.Author
	problem.Difficulty = updateReq.Difficulty
	problem.TimeLimit = updateReq.TimeLimit
	problem.MemoryLimit = updateReq.MemoryLimit
	problem.Tags = updateReq.Tags
	problem.Status = updateReq.Status
	problem.IsPublic = updateReq.IsPublic

	if err := s.svc.ProblemService.UpdateProblem(ctx, problem); err != nil {
		utils.InternalServerErrorResponse(c, "更新题目失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, problem)
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

	ctx := c.Request.Context()
	problems, total, err := s.svc.ProblemService.SearchProblems(
		ctx,
		keyword,
		models.ProblemDifficulty(difficulty),
		tags,
		page,
		pageSize,
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
