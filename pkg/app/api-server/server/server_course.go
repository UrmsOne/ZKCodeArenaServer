/*
@Author:
@Date: 2025/10/25
@Name: server_course.go
@Description: 课程服务器路由处理
*/

package server

import (
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"
	"zk-code-arena-server/pkg/utils/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterCourse(g *gin.RouterGroup) {
	courseGroup := g.Group("/courses")
	{
		// 需要认证的路由
		jwtGroup := courseGroup.Use(middleware.JWTMiddleware())
		{
			jwtGroup.GET("/:courseId", s.GetCourseById)
			jwtGroup.POST("/query", s.PageQueryCourse)
			jwtGroup.POST("/teacher/query", s.PageQueryTeacherCourses)
			jwtGroup.PUT("", s.UpdateCourse)
			jwtGroup.PUT("/:courseId/avatar", s.UpdateCourseAvatar)
			jwtGroup.POST("/teachers", s.addCourseTeacher)
			jwtGroup.DELETE("/teachers", s.removeCourseTeacher)

			//课程任务相关
			jwtGroup.POST("/task", s.AddTask)
			jwtGroup.PUT("/task", s.UpdateTask)
			jwtGroup.DELETE("/task/:taskId", s.DeleteTask)
			jwtGroup.GET("/task/:clazzId", s.GetTasksByClazzId)
			jwtGroup.POST("/finishtask", s.FinishTask)

			//班级相关
			jwtGroup.GET("/:courseId/clazzes", s.GetClazzesByCourseId)
			jwtGroup.POST("/clazzes", s.CreateClassForCourse)
			jwtGroup.GET("/clazzes/join", s.JoinClass)
			jwtGroup.GET("/clazzes/:clazzId", s.GetClazzById)
			jwtGroup.PUT("/clazzes", s.UpdateClazzInfo)
			jwtGroup.DELETE("/clazzes/:clazzId", s.DeleteClazz)
			jwtGroup.POST("/clazzes/members", s.AddClazzMember)
			jwtGroup.POST("/clazzes/members/remove", s.RemoveClazzMembers)
			jwtGroup.POST("/clazzes/teachers", s.addClazzTeacher)
			jwtGroup.DELETE("/clazzes/teachers", s.removeClazzTeacher)
		}

		// 需要老师权限的路由
		teacherGroup := jwtGroup.Use(middleware.RequireRole(models.RoleTeacher))
		{
			teacherGroup.POST("", s.CreateCourse) // 创建课程需要老师权限
		}
	}
}

// UpdateClazzInfo godoc
// @Summary      更新班级信息
// @Description  更新班级的基本信息（不包括教师和成员）
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.UpdateClazzRequest true "更新的班级信息"
// @Success      200 {object} map[string]interface{} "更新成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/clazzes [put]
func (s *Server) UpdateClazzInfo(c *gin.Context) {
	var req models.UpdateClazzRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.UpdateClazzInfo(c.Request.Context(), req.ClazzID, userID.(string), &req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// addCourseTeacher godoc
// @Summary      课程创建者添加老师
// @Description  课程创建者添加老师（支持批量添加，只有课程创建者可以操作）
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.AddCourseTeachersRequest true "添加教师请求"
// @Success      200 {object} map[string]interface{} "添加成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/teachers [post]
func (s *Server) addCourseTeacher(c *gin.Context) {
	var req models.AddCourseTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if req.CourseId == "" {
		utils.BadRequestResponse(c, "课程ID不能为空")
		return
	}

	if len(req.TeacherIds) == 0 {
		utils.BadRequestResponse(c, "教师ID列表不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.AddCourseTeachers(c.Request.Context(), req.CourseId, req.TeacherIds, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// removeCourseTeacher godoc
// @Summary      课程创建者删除课程老师
// @Description  课程创建者删除课程老师（支持批量删除，只有课程创建者可以操作）
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.RemoveCourseTeachersRequest true "删除教师请求"
// @Success      200 {object} map[string]interface{} "删除成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/teachers [delete]
func (s *Server) removeCourseTeacher(c *gin.Context) {
	var req models.RemoveCourseTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if req.CourseId == "" {
		utils.BadRequestResponse(c, "课程ID不能为空")
		return
	}

	if len(req.TeacherIds) == 0 {
		utils.BadRequestResponse(c, "教师ID列表不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.RemoveCourseTeachers(c.Request.Context(), req.CourseId, req.TeacherIds, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// addClazzTeacher godoc
// @Summary      班级添加老师
// @Description  班级添加老师（支持批量添加，只有课程创建者可以操作，且教师必须已加入课程）
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.AddClazzTeachersRequest true "添加教师请求"
// @Success      200 {object} map[string]interface{} "添加成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/clazzes/teachers [post]
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
	if err := s.svc.CourseService.AddClazzTeachers(c.Request.Context(), req.ClazzId, req.TeacherIds, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// removeClazzTeacher godoc
// @Summary      班级删除老师
// @Description  班级删除老师（支持批量删除，只有课程创建者可以操作）
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.RemoveClazzTeachersRequest true "移除教师请求"
// @Success      200 {object} map[string]interface{} "移除成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/clazzes/teachers [delete]
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
	if err := s.svc.CourseService.RemoveClazzTeachers(c.Request.Context(), req.ClazzId, req.TeacherIds, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// JoinClass godoc
// @Summary      加入班级
// @Description  学生通过邀请码加入班级
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.AddCourseTeachersRequest true "添加教师请求"
// @Success      200 {object} map[string]interface{} "添加成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/teachers [post]
func (s *Server) addCourseTeacher(c *gin.Context) {
	var req models.AddCourseTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if req.CourseId == "" {
		utils.BadRequestResponse(c, "课程ID不能为空")
		return
	}

	if len(req.TeacherIds) == 0 {
		utils.BadRequestResponse(c, "教师ID列表不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.AddCourseTeachers(c.Request.Context(), req.CourseId, req.TeacherIds, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// @Description  课程创建者删除课程老师（支持批量删除，只有课程创建者可以操作）
// @Summary      课程创建者删除课程老师
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.RemoveCourseTeachersRequest true "删除教师请求"
// @Success      200 {object} map[string]interface{} "删除成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/teachers [delete]
func (s *Server) removeCourseTeacher(c *gin.Context) {
	var req models.RemoveCourseTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if req.CourseId == "" {
		utils.BadRequestResponse(c, "课程ID不能为空")
		return
	}

	if len(req.TeacherIds) == 0 {
		utils.BadRequestResponse(c, "教师ID列表不能为空")
		return
	}

	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.RemoveCourseTeachers(c.Request.Context(), req.CourseId, req.TeacherIds, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// CreateCourse godoc
// @Summary      创建课程（教师）
// @Description  教师创建新课程
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.CreateCourseRequest true "课程信息"
// @Success      200 {object} map[string]interface{} "创建成功，返回课程ID"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses [post]
func (s *Server) CreateCourse(c *gin.Context) {
	var req models.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	courseID, err := s.svc.CourseService.CreateCourse(c.Request.Context(), &req, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, courseID)
}

// GetCourseById godoc
// @Summary      获取课程详情
// @Description  根据课程ID获取课程详细信息
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        courseId path string true "课程ID"
// @Success      200 {object} map[string]interface{} "课程详情"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/{courseId} [get]
func (s *Server) GetCourseById(c *gin.Context) {
	courseId := c.Param("courseId")
	if courseId == "" {
		utils.BadRequestResponse(c, "课程id不存在")
		return
	}
	userID, _ := c.Get("user_id")
	response, err := s.svc.CourseService.GetCourseByID(c.Request.Context(), courseId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, response)
}

// UpdateCourse godoc
// @Summary      更新课程信息
// @Description  更新课程的基本信息
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.UpdateCourseRequest true "更新的课程信息"
// @Success      200 {object} map[string]interface{} "更新成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses [put]
func (s *Server) UpdateCourse(c *gin.Context) {
	var req models.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "课程id为空")
		return
	}
	userID, _ := c.Get("user_id")

	if err := s.svc.CourseService.UpdateCourseInfo(c.Request.Context(), userID.(string), &req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// UpdateCourseAvatar godoc
// @Summary      更新课程头像
// @Description  上传并更新课程头像图片
// @Tags         课程
// @Accept       multipart/form-data
// @Produce      json
// @Param        courseId path string true "课程ID"
// @Param        file formData file true "头像文件"
// @Success      200 {object} map[string]interface{} "更新成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/{courseId}/avatar [put]
func (s *Server) UpdateCourseAvatar(c *gin.Context) {
	// 解析 multipart 请求
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 限制10MB
		utils.BadRequestResponse(c, "图片内存过大")
		return
	}
	courseId := c.Param("courseId")
	if courseId == "" {
		utils.BadRequestResponse(c, "课程id为空")
		return
	}
	// 获取文件头
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	id, _ := c.Get("user_id")
	if err = s.svc.CourseService.UpdateCourseAvatar(c.Request.Context(), file, header, id.(string), courseId); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// RemoveCourse godoc
// @Summary      删除课程
// @Description  删除指定课程
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        courseId path string true "课程ID"
// @Success      200 {object} map[string]interface{} "删除成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/{courseId} [delete]
func (s *Server) RemoveCourse(c *gin.Context) {
	courseId := c.Param("courseId")
	if courseId == "" {
		utils.BadRequestResponse(c, "课程id为空")
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.RemoveCourse(c.Request.Context(), userID.(string), courseId); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// PageQueryTeacherCourses godoc
// @Summary      分页查询老师加入的课程
// @Description  分页查询当前老师加入的课程列表
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.PageQueryTeacherCoursesRequest true "分页查询参数"
// @Success      200 {object} map[string]interface{} "课程列表"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/teacher/query [post]
func (s *Server) PageQueryTeacherCourses(c *gin.Context) {
	var req models.PageQueryTeacherCoursesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	res, err := s.svc.CourseService.PageQueryTeacherCourses(c.Request.Context(), &req, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, clazz)
}

// DeleteClazz godoc
// @Summary      删除班级
// @Description  删除指定班级
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Success      200 {object} map[string]interface{} "删除成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/clazzes/{clazzId} [delete]
func (s *Server) DeleteClazz(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级id为空")
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.DeleteClazz(c.Request.Context(), clazzId, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// AddClazzMember godoc
// @Summary      添加班级成员
// @Description  手动添加成员到班级
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.AddClazzMemberRequest true "成员ID"
// @Success      200 {object} map[string]interface{} "添加成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/clazzes/members [post]
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
	if err := s.svc.CourseService.AddClazzMember(c.Request.Context(), req.ClazzID, req.MemberID, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// RemoveClazzMembers godoc
// @Summary      移除班级成员
// @Description  批量移除班级成员
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        clazzId path string true "班级ID"
// @Param        request body models.RemoveClazzMembersRequest true "成员ID列表"
// @Success      200 {object} map[string]interface{} "移除成功"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/clazzes/members/remove [post]
func (s *Server) RemoveClazzMembers(c *gin.Context) {

	var req models.RemoveClazzMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.RemoveClazzMembers(c.Request.Context(), req, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// PageQueryTeacherCourses godoc
// @Summary      分页查询老师加入的课程
// @Description  分页查询当前老师加入的课程列表
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.PageQueryTeacherCoursesRequest true "分页查询参数"
// @Success      200 {object} map[string]interface{} "课程列表"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/teacher/query [post]
func (s *Server) PageQueryTeacherCourses(c *gin.Context) {
	var req models.PageQueryTeacherCoursesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	res, err := s.svc.CourseService.PageQueryTeacherCourses(c.Request.Context(), &req, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, res)
}

// GetClazzesByCourseId godoc
// @Summary      获取课程的所有班级
// @Description  根据课程ID获取该课程的所有班级列表
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        courseId path string true "课程ID"
// @Success      200 {object} map[string]interface{} "班级列表"
// @Failure      400 {object} map[string]interface{} "请求参数错误"
// @Security     BearerAuth
// @Router       /courses/{courseId}/clazzes [get]
func (s *Server) GetClazzesByCourseId(c *gin.Context) {
	courseId := c.Param("courseId")
	if courseId == "" {
		utils.BadRequestResponse(c, "课程ID不能为空")
		return
	}
	userID, _ := c.Get("user_id")
	clazzes, err := s.svc.CourseService.GetClazzesByCourseId(c.Request.Context(), courseId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, clazzes)
}
