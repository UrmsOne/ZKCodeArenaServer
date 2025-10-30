/*
@Author: sir
@Date: 2025/10/25
@Name: server_clazz.go
@Description: 班级服务器路由处理
*/

package server

import (
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"
	"zk-code-arena-server/pkg/utils/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) RegisterClazz(g *gin.RouterGroup) {
	clazzGroup := g.Group("/clazzes")
	{
		// 需要认证的路由
		jwtGroup := clazzGroup.Use(middleware.JWTMiddleware())
		{
			jwtGroup.POST("", s.CreateClass)
			jwtGroup.POST("/join", s.JoinClass)
			jwtGroup.GET("/:clazzId", s.GetClazzById)
			jwtGroup.PUT("", s.UpdateClazzInfo)
			jwtGroup.DELETE("/:clazzId", s.DeleteClazz)
			jwtGroup.POST("/members", s.AddClazzMember)
			jwtGroup.POST("/members/remove", s.RemoveClazzMembers)
			jwtGroup.POST("/teachers", s.addClazzTeacher)
			jwtGroup.DELETE("/teachers", s.removeClazzTeacher)
			jwtGroup.GET("/course/:courseId", s.GetClazzesByCourseId)
			//二维码
			jwtGroup.PUT("/qrcode/:courseId/:clazzId", s.refreshQrcode)
			jwtGroup.GET("/qrcode/:clazzId", s.GetQrcodeClazzById)
			//课程任务相关
			jwtGroup.POST("/finishTask", s.FinishTask)
			jwtGroup.POST("/task", s.AddTask)
			jwtGroup.PUT("/task", s.UpdateTask)
			jwtGroup.DELETE("/task/:taskId", s.DeleteTask)
			jwtGroup.GET("/tasks/:clazzId", s.GetTasksByClazzId)
			jwtGroup.GET("/task/:taskId", s.GetTaskById)

			// 学生班级相关接口
			jwtGroup.POST("/student_classes", s.AddStudentToClass)
			jwtGroup.DELETE("/student_classes", s.RemoveStudentFromClass)
			jwtGroup.GET("/student_classes/:userId", s.GetStudentClasses)
			jwtGroup.GET("/class_students/:classId", s.GetClassStudents)
		}
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
// @Router       /clazzes/qrcode/{clazzId} [get]
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
		if err.Error() == "权限不足" {
			utils.ForbiddenResponse(c, err.Error())
			return
		} else if err.Error() == "二维码已过期" {
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
// @Success      200 {object} models.SuccessResponse "更新成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/qrcode/{courseId}/{clazzId} [put]
func (s *Server) refreshQrcode(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "参数为空")
		return
	}
	courseId := c.Param("courseId")
	if courseId == "" {
		utils.BadRequestResponse(c, "参数为空")
		return
	}

	userID, _ := c.Get("user_id")
	qrcodeBase64, err := s.svc.CourseService.RefreshQrcode(userID.(string), courseId, clazzId, c.Request.Context())
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
// @Param        ran query string true "随机值"
// @Param        clazzId query string true "班级ID"
// @Success      200 {object} utils.Response "加入成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/join [post]
func (s *Server) JoinClass(c *gin.Context) {
	var req models.JoinClazzRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.JoinClazz(c.Request.Context(), req, userID.(string)); err != nil {
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
// @Param        taskId path string true "任务ID"
// @Success      200 {object} utils.Response "删除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/task/{taskId} [delete]
func (s *Server) DeleteTask(c *gin.Context) {
	taskId := c.Param("taskId")
	if taskId == "" {
		utils.BadRequestResponse(c, "参数为空")
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
// @Param        request body models.UpdateTaskRequest true "更新的任务信息"
// @Success      200 {object} utils.Response "更新成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/task [put]
func (s *Server) UpdateTask(c *gin.Context) {
	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.UpdateTask(c.Request.Context(), userID.(string), req); err != nil {
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
// @Param        request body models.AddTaskRequest true "任务信息"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/task [post]
func (s *Server) AddTask(c *gin.Context) {
	var req models.AddTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.AddCourseTask(c.Request.Context(), &req, userID.(string)); err != nil {
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
// @Param        request body models.FinishTaskRequest true "完成任务信息"
// @Success      200 {object} utils.Response "完成成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/finishTask [post]
func (s *Server) FinishTask(c *gin.Context) {
	var req models.FinishTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.FinishTask(c.Request.Context(), req, userID.(string)); err != nil {
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
// @Router       /clazzes [put]
func (s *Server) UpdateClazzInfo(c *gin.Context) {
	var req models.UpdateClazzRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.UpdateClazzInfo(c.Request.Context(), req.ClazzID, userID.(string), &req); err != nil {
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
// @Router       /clazzes/members [post]
func (s *Server) AddClazzMember(c *gin.Context) {
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
	if err := s.svc.ClazzService.AddClazzMember(c.Request.Context(), req.ClazzID, req.MemberID, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// RemoveClazzMembers godoc
// @Summary      移除班级成员
// @Description  批量移除班级成员
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.RemoveClazzMembersRequest true "成员ID列表"
// @Success      200 {object} utils.Response "移除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/members/remove [post]
func (s *Server) RemoveClazzMembers(c *gin.Context) {
	var req models.RemoveClazzMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.RemoveClazzMembers(c.Request.Context(), req, userID.(string)); err != nil {
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
// @Param        request body models.AddClazzTeachersRequest true "添加教师请求"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/teachers [post]
func (s *Server) addClazzTeacher(c *gin.Context) {
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
	if err := s.svc.ClazzService.AddClazzTeacher(c.Request.Context(), req.ClazzId, req.TeacherIds[0], userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// @Summary      班级删除老师
// @Description  班级删除老师（支持批量删除，只有课程创建者可以操作）
// @Tags         班级
// @Accept       json
// @Produce      json
// @Param        request body models.RemoveClazzTeachersRequest true "移除教师请求"
// @Success      200 {object} utils.Response "移除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/teachers [delete]
func (s *Server) removeClazzTeacher(c *gin.Context) {
	var req models.RemoveClazzTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	if req.ClazzId == "" {
		utils.BadRequestResponse(c, "班级ID不能为空")
		return
	}

	if len(req.TeacherIds) == 0 {
		utils.BadRequestResponse(c, "教师ID列表不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.ClazzService.RemoveClazzTeacher(c.Request.Context(), req.ClazzId, req.TeacherIds[0], userID.(string)); err != nil {
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
// @Param        courseId path string true "课程ID"
// @Success      200 {object} utils.Response{data=[]models.Clazz} "班级列表"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/course/{courseId} [get]
func (s *Server) GetClazzesByCourseId(c *gin.Context) {
	courseId := c.Param("courseId")
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
// @Router       /clazzes/tasks/{clazzId} [get]
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
// @Param        taskId path string true "任务ID"
// @Success      200 {object} utils.Response{data=models.Task} "任务详情"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/task/{taskId} [get]
func (s *Server) GetTaskById(c *gin.Context) {
	taskId := c.Param("taskId")
	if taskId == "" {
		utils.BadRequestResponse(c, "任务ID不能为空")
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
// @Param        request body models.AddStudentToClassRequest true "添加学生到班级请求"
// @Success      200 {object} utils.Response "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/student_classes [post]
func (s *Server) AddStudentToClass(c *gin.Context) {
	var req models.AddStudentToClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := s.svc.ClazzService.AddStudentToClass(c.Request.Context(), req.StudentID, req.ClassID, req.CourseID); err != nil {
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
// @Param        request body models.RemoveStudentFromClassRequest true "从班级移除学生请求"
// @Success      200 {object} utils.Response "移除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
// @Security     BearerAuth
// @Router       /clazzes/student_classes [delete]
func (s *Server) RemoveStudentFromClass(c *gin.Context) {
	var req models.RemoveStudentFromClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := s.svc.ClazzService.RemoveStudentFromClass(c.Request.Context(), req.StudentID, req.ClassID, req.CourseID); err != nil {
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
// @Router       /clazzes/student_classes/{studentId} [get]
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
// @Router       /clazzes/class_students/{classId} [get]
func (s *Server) GetClassStudents(c *gin.Context) {
	classId := c.Param("classId")
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
