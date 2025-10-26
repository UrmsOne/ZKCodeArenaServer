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
			jwtGroup.PUT("/:courseId", s.UpdateCourse)
			jwtGroup.DELETE("/:courseId", s.RemoveCourse)
			jwtGroup.PUT("/:courseId/avatar", s.UpdateCourseAvatar)
			jwtGroup.GET("/:courseId/clazzes", s.GetClazzesByCourseId)
			jwtGroup.POST("/teacher/query", s.PageQueryTeacherCourses)
			jwtGroup.POST("/teachers", s.addCourseTeacher)
			jwtGroup.DELETE("/teachers", s.removeCourseTeacher)
			// 需要老师权限的路由
			teacherGroup := jwtGroup.Use(middleware.RequireRole(models.RoleTeacher))
			{
				teacherGroup.POST("", s.CreateCourse) // 创建课程需要老师权限
			}
		}
	}
}

// @Description  课程创建者添加老师（支持批量添加，只有课程创建者可以操作）
// @Summary      课程创建者添加老师
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        request body models.AddCourseTeachersRequest true "添加教师请求"
// @Success      200 {object} models.SuccessResponse "添加成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
// @Success      200 {object} models.DeleteResponse "删除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
// @Success      200 {object} models.CreateResponse "创建成功，返回课程ID"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
// @Success      200 {object} models.Course "课程详情"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
// @Success      200 {object} models.UpdateResponse "更新成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
// @Success      200 {object} models.UpdateResponse "更新成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
// @Success      200 {object} models.DeleteResponse "删除成功"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
// @Success      200 {object} models.CourseListResponse "课程列表"
// @Failure      400 {object} models.ErrorResponse "请求参数错误"
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
