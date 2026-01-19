/*
@Author: sir
@Date: 2025/10/25
@Name: server_clazz.go
@Description: 班级服务器路由处理
*/

package server

import (
	"errors"

	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"
	"zk-code-arena-server/pkg/utils/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) RegisterClazz(g *gin.RouterGroup) {
	clazzGroup := g.Group("/clazzes").Use(middleware.JWTMiddleware())
	{
		// 班级
		clazzGroup.POST("", s.CreateClass)
		clazzGroup.GET("", s.GetClazzesByCourseId)
		clazzGroup.GET("/:clazzId", s.GetClazzById)
		clazzGroup.PUT("/:clazzId", s.UpdateClazzInfo)
		clazzGroup.DELETE("/:clazzId", s.DeleteClazz)
		clazzGroup.POST("/:clazzId/join", s.JoinClass)

		// 二维码管理
		clazzGroup.GET("/:clazzId/qrcode", s.GetQrcodeClazzById)
		clazzGroup.PUT("/:clazzId/qrcode", s.refreshQrcode)

		// 成员管理
		clazzGroup.POST("/:clazzId/members", s.AddClazzMember)
		clazzGroup.DELETE("/:clazzId/members/:memberId", s.RemoveClazzMembers)
		clazzGroup.POST("/:clazzId/teachers", s.addClazzTeacher)
		clazzGroup.DELETE("/:clazzId/teachers/:teacherId", s.removeClazzTeacher)

		// 任务管理
		clazzGroup.POST("/:clazzId/tasks", s.AddTask)
		clazzGroup.GET("/:clazzId/tasks", s.GetTasksByClazzId)
		clazzGroup.GET("/:clazzId/tasks/:taskId", s.GetTaskById)
		clazzGroup.PUT("/:clazzId/tasks/:taskId", s.UpdateTask)
		clazzGroup.DELETE("/:clazzId/tasks/:taskId", s.DeleteTask)
		clazzGroup.POST("/:clazzId/tasks/:taskId/finish", s.FinishTask)
		clazzGroup.POST("/:clazzId/tasks/:taskId/relations", s.AddTaskRelationIds)
		clazzGroup.DELETE("/:clazzId/tasks/:taskId/relations", s.RemoveTaskRelationIds)
		clazzGroup.GET("/:clazzId/tasks/:taskId/completion", s.PageQueryTaskCompletion)
		clazzGroup.POST("/:clazzId/tasks/:taskId/copy", s.CopyTaskToClass)

		// 学生班级关系
		clazzGroup.POST("/:clazzId/student/:studentId", s.AddStudentToClass)
		clazzGroup.DELETE("/:clazzId/student/:studentId", s.RemoveStudentFromClass)
		clazzGroup.GET("/:clazzId/students", s.GetClassStudents) //

	}
	stuGroup := g.Group("/students").Use(middleware.JWTMiddleware())
	{
		stuGroup.GET("/:studentId/classes", s.GetStudentClasses)
	}
}

// GetQrcodeClazzById godoc
// @Summary      获得班级二维码
// @Description  获得班级二维码
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Success      200 {object} models.SuccessResponse "获得成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Failure      403 {object} models.ErrorResponse "权限不足"
// @Failure      404 {object} models.ErrorResponse "二维码已过期"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/qrcode [get]
func (s *Server) GetQrcodeClazzById(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "参数为空")
		return
	}

	userID, _ := c.Get("user_id")
	qrcodeBase64, err := s.svc.CourseService.GetQrcode(userID.(string), clazzId, c.Request.Context())
	if err != nil {
		// 根据错误类型返回不同的响应
		if errors.Is(err, models.ErrPermissionDenied) {
			utils.ForbiddenResponse(c, err.Error())
			return
		} else if errors.Is(err, models.ErrQRCodeExpired) {
			utils.NotFoundResponse(c, err.Error())
			return
		} else {
			utils.BadRequestResponse(c, err.Error())
			return
		}
	}
	utils.SuccessResponse(c, qrcodeBase64)
}

// refreshQrcode godoc
// @Summary      更新班级二维码
// @Description  更新班级二维码
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Success      200 {object} models.SuccessResponse "更新成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/qrcode [put]
func (s *Server) refreshQrcode(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}
	userID, _ := c.Get("user_id")
	qrcodeBase64, err := s.svc.ClazzService.RefreshQrcodeByClazzId(userID.(string), clazzId, c.Request.Context())
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, qrcodeBase64)
}

// JoinClass godoc
// @Summary      通过二维码加入班级
// @Description  学生通过扫描二维码加入班级
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        invite_code query string false "邀请码"
// @Success      200 {object} utils.Response "加入成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/join [post]
func (s *Server) JoinClass(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}

	var req models.JoinClazzRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.JoinClazz(c.Request.Context(), clazzId, req, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// DeleteTask godoc
// @Summary      删除任务
// @Description  删除指定任务
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Success      200 {object} utils.Response "删除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId} [delete]
func (s *Server) DeleteTask(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.DeleteTask(c.Request.Context(), userID.(string), taskId); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// UpdateTask godoc
// @Summary      更新任务
// @Description  更新任务信息
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Param        request body models.UpdateTaskRequest true "更新的任务信息"
// @Success      200 {object} utils.Response "更新成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId} [put]
func (s *Server) UpdateTask(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.UpdateTask(c.Request.Context(), userID.(string), clazzId, taskId, req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// AddTaskRelationIds godoc
// @Summary      添加任务关系ID
// @Description  为任务添加关系ID
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Param        request body models.AddTaskRelationIdsRequest true "添加的关系ID信息"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId}/relations [post]
func (s *Server) AddTaskRelationIds(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}

	var req models.AddTaskRelationIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.AddTaskRelationIds(c.Request.Context(), userID.(string), taskId, req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// RemoveTaskRelationIds godoc
// @Summary      删除任务关系ID
// @Description  从任务中删除关系ID
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Param        request body models.RemoveTaskRelationIdsRequest true "删除的关系ID信息"
// @Success      200 {object} utils.Response "删除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId}/relations [delete]
func (s *Server) RemoveTaskRelationIds(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}

	var req models.RemoveTaskRelationIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.RemoveTaskRelationIds(c.Request.Context(), userID.(string), taskId, req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// AddTask godoc
// @Summary      添加任务
// @Description  为班级添加新任务
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.AddTaskRequest true "任务信息"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks [post]
func (s *Server) AddTask(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}
	var req models.AddTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.AddCourseTask(c.Request.Context(), clazzId, &req, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// FinishTask godoc
// @Summary      完成任务
// @Description  标记任务为已完成
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Param        request body models.FinishTaskRequest true "完成任务信息"
// @Success      200 {object} utils.Response "完成成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId}/finish [post]
func (s *Server) FinishTask(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}
	var req models.FinishTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.FinishTask(c.Request.Context(), clazzId, taskId, req, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// CreateClass godoc
// @Summary      创建班级
// @Description  创建新班级
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        request body models.CreateClazzRequest true "班级信息"
// @Success      200 {object} utils.Response{data=models.GetClazzResponse} "创建成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes [post]
func (s *Server) CreateClass(c *gin.Context) {
	var req models.CreateClazzRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	response, err := s.svc.ClazzService.CreateClass(c.Request.Context(), &req, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, response)
}

//// JoinClass godoc
//// @Summary      加入班级
//// @Description  学生通过邀请码加入班级
//// @Tags         班级
//// @Accept       json
//// @Produce      json
//// @Param        clazzId query string true "班级ID"
//// @Param        invite_code query string true "邀请码"
//// @Success      200 {object} utils.Response "加入成功"
//// @Failure      400 {object} models.ErrorResponse "请求参数错误"
//// @Security     BearerAuth
//// @Router       /clazzes/join [get]
//func (s *Server) JoinClass(c *gin.Context) {
//	var req models.JoinClazzRequest
//	if err := c.ShouldBindQuery(&req); err != nil {
//		utils.BadRequestResponse(c, err.Error())
//		return
//	}
//
//	if req.ClazzID == "" {
//		utils.BadRequestResponse(c, "班级id为空")
//		return
//	}
//	userID, _ := c.Get("user_id")
//	if err := s.svc.ClazzService.joinClazz(c.Request.Context(), &req, userID.(string)); err != nil {
//		utils.BadRequestResponse(c, err.Error())
//		return
//	}
//	utils.SuccessResponse(c, nil)
//}

// GetClazzById godoc
// @Summary      获取班级详情
// @Description  根据班级ID获取班级详细信息
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Success      200 {object} utils.Response{data=models.Clazz} "班级详情"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId} [get]
func (s *Server) GetClazzById(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级id为空")
		return
	}
	userID, _ := c.Get("user_id")
	clazz, err := s.svc.ClazzService.GetClazzByID(c.Request.Context(), clazzId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, clazz)
}

// UpdateClazzInfo godoc
// @Summary      更新班级信息
// @Description  更新班级的基本信息（不包括教师和成员）
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.UpdateClazzRequest true "更新的班级信息"
// @Success      200 {object} utils.Response "更新成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId} [put]
func (s *Server) UpdateClazzInfo(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}

	var req models.UpdateClazzRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.UpdateClazzInfo(c.Request.Context(), clazzId, userID.(string), &req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// DeleteClazz godoc
// @Summary      删除班级
// @Description  删除指定班级
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Success      200 {object} utils.Response "删除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId} [delete]
func (s *Server) DeleteClazz(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级id为空")
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.DeleteClazz(c.Request.Context(), clazzId, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// AddClazzMember godoc
// @Summary      添加班级成员
// @Description  手动添加成员到班级
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.AddClazzMemberRequest true "成员ID"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/members [post]
func (s *Server) AddClazzMember(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}

	var req models.AddClazzMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if req.MemberID == "" {
		utils.BadRequestResponse(c, "成员id为空")
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.AddClazzMember(c.Request.Context(), clazzId, req.MemberID, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// RemoveClazzMembers godoc
// @Summary      移除班级成员
// @Description  移除班级成员
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        memberId path string true "成员ID"
// @Success      200 {object} utils.Response "移除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/members/{memberId} [delete]
func (s *Server) RemoveClazzMembers(c *gin.Context) {
	clazzId := c.Param("clazzId")
	memberId := c.Param("memberId")

	if clazzId == "" || memberId == "" {
		utils.BadRequestResponse(c, "班级ID或成员ID不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.RemoveClazzMember(c.Request.Context(), clazzId, memberId, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// @Summary      班级添加老师
// @Description  班级添加老师（支持批量添加，只有课程创建者可以操作，且教师必须已加入课程）
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.AddClazzTeachersRequest true "添加教师请求"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/teachers [post]
func (s *Server) addClazzTeacher(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}

	var req models.AddClazzTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if len(req.TeacherIds) == 0 {
		utils.BadRequestResponse(c, "教师ID列表不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.AddClazzTeachers(c.Request.Context(), clazzId, req.TeacherIds, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// @Summary      班级删除老师
// @Description  班级删除老师（只有课程创建者可以操作）
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        teacherId path string true "教师ID"
// @Success      200 {object} utils.Response "移除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/teachers/{teacherId} [delete]
func (s *Server) removeClazzTeacher(c *gin.Context) {
	clazzId := c.Param("clazzId")
	teacherId := c.Param("teacherId")

	if clazzId == "" || teacherId == "" {
		utils.BadRequestResponse(c, "班级ID或教师ID不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.RemoveClazzTeacher(c.Request.Context(), clazzId, teacherId, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// GetClazzesByCourseId godoc
// @Summary      获取课程的所有班级
// @Description  根据课程ID获取该课程的所有班级列表
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        course_id query string true "课程ID"
// @Success      200 {object} utils.Response{data=[]models.Clazz} "班级列表"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes [get]
func (s *Server) GetClazzesByCourseId(c *gin.Context) {
	courseId := c.Query("course_id")
	if courseId == "" {
		utils.BadRequestResponse(c, "课程ID不能为空")
		return
	}
	userID, _ := c.Get("user_id")
	clazzes, err := s.svc.ClazzService.GetClazzesByCourseId(c.Request.Context(), courseId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, clazzes)
}

// GetTasksByClazzId godoc
// @Summary      获取班级任务列表
// @Description  获取指定班级的任务列表
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Success      200 {object} []models.Task "任务列表"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks [get]
func (s *Server) GetTasksByClazzId(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	tasks, err := s.svc.ClazzService.GetTasksByClazzID(c.Request.Context(), clazzId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, tasks)
}

// GetTaskById godoc
// @Summary      获取任务详情
// @Description  根据任务ID获取任务详细信息
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Success      200 {object} utils.Response{data=models.TaskResponse} "任务详情"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId} [get]
func (s *Server) GetTaskById(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	task, err := s.svc.ClazzService.GetTaskByID(c.Request.Context(), taskId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, task)
}

// AddStudentToClass godoc
// @Summary      添加学生到班级
// @Description  将学生添加到指定班级
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        studentId path string true "学生ID"
// @Param        request body models.AddStudentToClassRequest true "添加学生到班级请求"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/student/{studentId} [post]
func (s *Server) AddStudentToClass(c *gin.Context) {
	clazzId := c.Param("clazzId")
	studentId := c.Param("studentId")

	if clazzId == "" || studentId == "" {
		utils.BadRequestResponse(c, "班级ID或学生ID不能为空")
		return
	}

	var req models.AddStudentToClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := s.svc.ClazzService.AddStudentToClass(c.Request.Context(), studentId, clazzId, req.CourseID); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// RemoveStudentFromClass godoc
// @Summary      从班级移除学生
// @Description  将学生从指定班级移除
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        studentId path string true "学生ID"
// @Param        request body models.RemoveStudentFromClassRequest true "从班级移除学生请求"
// @Success      200 {object} utils.Response "移除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/student/{studentId} [delete]
func (s *Server) RemoveStudentFromClass(c *gin.Context) {
	clazzId := c.Param("clazzId")
	studentId := c.Param("studentId")

	if clazzId == "" || studentId == "" {
		utils.BadRequestResponse(c, "班级ID或学生ID不能为空")
		return
	}

	var req models.RemoveStudentFromClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := s.svc.ClazzService.RemoveStudentFromClass(c.Request.Context(), studentId, clazzId, req.CourseID); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// GetStudentClasses godoc
// @Summary      获取学生的所有班级
// @Description  获取指定学生加入的所有班级
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        studentId path string true "学生ID"
// @Success      200 {object} []models.StudentClassResponse "学生班级列表"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /students/{studentId}/classes [get]
func (s *Server) GetStudentClasses(c *gin.Context) {
	studentId := c.Param("studentId")
	if studentId == "" {
		utils.BadRequestResponse(c, "学生ID不能为空")
		return
	}

	// 验证学生ID格式
	_, err := primitive.ObjectIDFromHex(studentId)
	if err != nil {
		utils.BadRequestResponse(c, "无效的学生ID")
		return
	}

	classes, err := s.svc.ClazzService.GetStudentClasses(c.Request.Context(), studentId)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, classes)
}

// GetClassStudents godoc
// @Summary      获取班级的所有学生
// @Description  获取指定班级的所有学生
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        classId path string true "班级ID"
// @Success      200 {object} []models.UserProfile "班级学生列表"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/students [get]
func (s *Server) GetClassStudents(c *gin.Context) {
	classId := c.Param("clazzId")
	if classId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}

	// 验证班级ID格式
	_, err := primitive.ObjectIDFromHex(classId)
	if err != nil {
		utils.BadRequestResponse(c, "无效的班级ID")
		return
	}

	students, err := s.svc.ClazzService.GetClassStudents(c.Request.Context(), classId)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, students)
}

// CopyTaskToClass godoc
// @Summary      复制任务到班级
// @Description  将一个班级的任务复制到另一个班级
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Param        request body models.CopyTaskToClassRequest true "复制任务请求"
// @Success      200 {object} utils.Response "复制成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId}/copy [post]
func (s *Server) CopyTaskToClass(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}

	var req models.CopyTaskToClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.CopyTaskToClass(c.Request.Context(), userID.(string), taskId, req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// PageQueryTaskCompletion godoc
// @Summary      分页查询班级任务完成情况
// @Description  分页查询指定班级下所有学生的任务完成情况，支持按学生姓名和学号搜索
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        taskId path string true "任务ID"
// @Param        page_num query int false "页码"
// @Param        page_size query int false "每页数量"
// @Param        real_name query string false "真实姓名"
// @Param        user_id query string false "用户ID"
// @Success      200 {object} utils.Response{data=models.PageQueryTaskCompletionResponse} "任务完成情况列表"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Failure      403 {object} models.ErrorResponse "权限不足"
// @Security     BearerAuth
// @Router       /clazzes/{clazzId}/tasks/{taskId}/completion [get]
func (s *Server) PageQueryTaskCompletion(c *gin.Context) {
	clazzId := c.Param("clazzId")
	taskId := c.Param("taskId")
	if clazzId == "" || taskId == "" {
		utils.BadRequestResponse(c, "班级ID或任务ID不能为空")
		return
	}

	var req models.PageQueryTaskCompletionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	response, err := s.svc.ClazzService.PageQueryTaskCompletion(c.Request.Context(), clazzId, taskId, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, response)
}
