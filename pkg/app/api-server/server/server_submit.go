/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/25 13:09
@Name: server_submit.go
@Description: 提交相关路由处理
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

// RegisterSubmit 注册提交相关路由
func (s *Server) RegisterSubmit(g *gin.RouterGroup) {
	submitGroup := g.Group("/submit")
	{
		// 需要认证的路由
		submitGroup.Use(middleware.JWTMiddleware())
		{
			submitGroup.POST("/", s.SubmitCode)  // 提交代码
			submitGroup.GET("/", s.GetSubmits)   // 获取提交列表
			submitGroup.GET("/:id", s.GetSubmit) // 获取提交详情
		}
	}
}

// SubmitCode 提交代码
func (s *Server) SubmitCode(c *gin.Context) {
	var submitReq struct {
		ProblemID primitive.ObjectID `json:"problem_id" binding:"required"`
		Code      string             `json:"code" binding:"required"`
		Language  string             `json:"language" binding:"required"`
	}

	if err := c.ShouldBindJSON(&submitReq); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证编程语言
	language := models.Language(submitReq.Language)
	switch language {
	case models.LangC, models.LangCpp, models.LangJava, models.LangPython, models.LangGo:
		// 支持的语言
	default:
		utils.BadRequestResponse(c, "不支持的编程语言")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "未认证用户")
		return
	}

	submitUserID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, "无效的用户ID")
		return
	}

	// 检查题目是否存在
	ctx := c.Request.Context()
	problem, err := s.svc.ProblemService.GetProblemByID(ctx, submitReq.ProblemID)
	if err != nil {
		utils.NotFoundResponse(c, "题目不存在")
		return
	}

	// 检查题目是否公开
	if !problem.IsPublic {
		utils.ForbiddenResponse(c, "无法提交私有题目")
		return
	}

	// 创建提交记录
	submit := &models.Submit{
		ProblemID: submitReq.ProblemID,
		UserID:    submitUserID,
		Code:      submitReq.Code,
		Language:  language,
	}

	if err := s.svc.SubmitService.CreateSubmit(ctx, submit); err != nil {
		utils.InternalServerErrorResponse(c, "创建提交失败: "+err.Error())
		return
	}

	// 异步评测
	go func() {
		// 调用评测服务
		result, err := s.svc.JudgeService.JudgeSubmit(ctx, submit, problem)
		if err != nil {
			// 评测失败，更新状态为系统错误
			submit.Status = models.StatusSystemError
			submit.Result = &models.JudgeResult{
				Status:       models.StatusSystemError,
				RuntimeError: err.Error(),
			}
		} else {
			submit.Status = result.Status
			submit.Result = result
		}

		// 更新提交记录
		s.svc.SubmitService.UpdateSubmit(ctx, submit)

		// 更新题目统计
		isAC := result != nil && result.Status == models.StatusAccepted
		s.svc.ProblemService.UpdateProblemStats(ctx, submitReq.ProblemID, isAC)
	}()

	utils.SuccessResponse(c, gin.H{
		"message":   "提交成功，正在评测中",
		"submit_id": submit.ID,
	})
}

// GetSubmits 获取提交列表
func (s *Server) GetSubmits(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	problemIDStr := c.Query("problem_id")
	userIDStr := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var problemID, userID primitive.ObjectID
	var err error

	if problemIDStr != "" {
		problemID, err = primitive.ObjectIDFromHex(problemIDStr)
		if err != nil {
			utils.BadRequestResponse(c, "无效的题目ID")
			return
		}
	}

	if userIDStr != "" {
		userID, err = primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			utils.BadRequestResponse(c, "无效的用户ID")
			return
		}
	}

	// 检查权限：非管理员只能查看自己的提交
	role, exists := c.Get("role")
	if !exists {
		utils.UnauthorizedResponse(c, "需要登录")
		return
	}

	if role.(string) != string(models.RoleAdmin) {
		// 非管理员只能查看自己的提交
		currentUserID, _ := c.Get("user_id")
		currentUserObjectID, _ := primitive.ObjectIDFromHex(currentUserID.(string))
		userID = currentUserObjectID
	}

	ctx := c.Request.Context()
	submits, total, err := s.svc.SubmitService.GetSubmits(ctx, page, pageSize, userID, problemID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取提交列表失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"submits":    submits,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetSubmit 获取提交详情
func (s *Server) GetSubmit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "无效的提交ID")
		return
	}

	ctx := c.Request.Context()
	submit, err := s.svc.SubmitService.GetSubmitByID(ctx, id)
	if err != nil {
		utils.NotFoundResponse(c, "提交不存在")
		return
	}

	// 检查权限：非管理员只能查看自己的提交
	role, exists := c.Get("role")
	if !exists {
		utils.UnauthorizedResponse(c, "需要登录")
		return
	}

	if role.(string) != string(models.RoleAdmin) {
		currentUserID, _ := c.Get("user_id")
		currentUserObjectID, _ := primitive.ObjectIDFromHex(currentUserID.(string))
		if submit.UserID != currentUserObjectID {
			utils.ForbiddenResponse(c, "只能查看自己的提交")
			return
		}
	}

	utils.SuccessResponse(c, submit)
}
