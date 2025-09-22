/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/25 13:09
@Name: server_problem.go
@Description: 题目相关路由处理
*/

package server

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"
	"zk-code-arena-server/pkg/utils/middleware"
)

// RegisterProblem 注册题目相关路由
func (s *Server) RegisterProblem(g *gin.RouterGroup) {
	problemGroup := g.Group("/problem")
	{
		// 公开路由
		problemGroup.GET("/", s.GetProblems)   // 获取题目列表
		problemGroup.GET("/:id", s.GetProblem) // 获取题目详情
	}

	// 需要认证的路由
	securedGroup := problemGroup.Group("/").Use(middleware.JWTMiddleware())
	{
		securedGroup.POST("/", s.CreateProblem)      // 创建题目
		securedGroup.PUT("/:id", s.UpdateProblem)    // 更新题目
		securedGroup.DELETE("/:id", s.DeleteProblem) // 删除题目
	}
}

// GetProblems 获取题目列表
func (s *Server) GetProblems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	difficulty := c.Query("difficulty")
	tags := c.QueryArray("tags")
	isPublic := c.DefaultQuery("is_public", "true") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Request.Context()
	problems, total, err := s.svc.ProblemService.GetProblems(ctx, page, pageSize, models.ProblemDifficulty(difficulty), tags, isPublic)
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

// GetProblem 获取题目详情
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

// CreateProblem 创建题目
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

// UpdateProblem 更新题目
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

// DeleteProblem 删除题目
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
