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
			jwtGroup.DELETE("/:courseId", s.RemoveCourse)
			jwtGroup.PUT("/:courseId/:clazzId", s.RefreshQrcode) //这里如果直传班级也行但是会多一次数据库查询故我认为有必要冗余
			jwtGroup.GET("/:courseId", s.GetCourseById)
			jwtGroup.POST("/query", s.PageQueryCourse)
			jwtGroup.PUT("", s.UpdateCourse)
			jwtGroup.PUT("/:courseId/avatar", s.UpdateCourseAvatar)

			//课程任务相关
			jwtGroup.POST("/task", s.AddTask)
			jwtGroup.PUT("/task", s.UpdateTask)
			jwtGroup.DELETE("/task/:taskId", s.DeleteTask)
			jwtGroup.GET("/task/:clazzId", s.GetTasksByClazzId)
			jwtGroup.POST("/finishtask", s.FinishTask)

			//班级相关
			jwtGroup.POST("/clazzes", s.CreateClassForCourse)
			jwtGroup.GET("/clazzes/join", s.JoinClass)
			jwtGroup.GET("/clazzes/:clazzId", s.GetClazzById)
			jwtGroup.DELETE("/clazzes/:clazzId", s.DeleteClazz)
			jwtGroup.POST("/clazzes/:clazzId/members", s.AddClazzMember)
			jwtGroup.POST("/clazzes/:clazzId/members/remove", s.RemoveClazzMembers)
		}

		// 需要老师权限的路由
		teacherGroup := jwtGroup.Use(middleware.RequireRole(models.RoleTeacher))
		{
			teacherGroup.POST("", s.CreateCourse) // 创建课程需要老师权限
		}
	}
}

func (s *Server) JoinClass(c *gin.Context) {
	var req models.JoinClazzRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if req.ClazzID == "" {
		utils.BadRequestResponse(c, "班级id为空")
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.JoinClazz(c.Request.Context(), &req, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

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

func (s *Server) CreateClassForCourse(c *gin.Context) {
	var req models.CreateClazzRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	response, err := s.svc.CourseService.CreateClass(c.Request.Context(), &req, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, response)
}

func (s *Server) RefreshQrcode(c *gin.Context) {
	courseId := c.Param("courseId")
	clazzId := c.Param("clazzId")
	if courseId == "" || clazzId == "" {
		utils.BadRequestResponse(c, "关键参数为空")
		return
	}
	userID, _ := c.Get("user_id")

	response, err := s.svc.CourseService.RefreshInviteCode(c.Request.Context(), courseId, clazzId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, response)
}

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

// 分页查询课程
func (s *Server) PageQueryCourse(c *gin.Context) {
	var PageQueryCourseRequest models.PageQueryCourseRequest
	if err := c.ShouldBindJSON(&PageQueryCourseRequest); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	res, err := s.svc.CourseService.PageQueryCourse(c.Request.Context(), &PageQueryCourseRequest, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, res)
}

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

func (s *Server) AddTask(c *gin.Context) {
	var req models.AddTaskRequest
	if err := c.ShouldBind(&req); err != nil {
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

func (s *Server) FinishTask(c *gin.Context) {
	var req models.FinishTaskRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.FinishTask(c.Request.Context(), req, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

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

func (s *Server) GetTasksByClazzId(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级id为空")
		return
	}

	userID, _ := c.Get("user_id")
	tasks, err := s.svc.CourseService.GetTasksByClazzID(c.Request.Context(), clazzId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, tasks)
}

func (s *Server) GetClazzById(c *gin.Context) {
	clazzId := c.Param("clazzId")
	if clazzId == "" {
		utils.BadRequestResponse(c, "班级id为空")
		return
	}
	userID, _ := c.Get("user_id")
	clazz, err := s.svc.CourseService.GetClazzByID(c.Request.Context(), clazzId, userID.(string))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, clazz)
}

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

func (s *Server) AddClazzMember(c *gin.Context) {
	clazzId := c.Param("clazzId")
	memberId := c.Param("memberId")
	if clazzId == "" || memberId == "" {
		utils.BadRequestResponse(c, "班级id或成员id为空")
		return
	}
	userID, _ := c.Get("user_id")
	if err := s.svc.CourseService.AddClazzMember(c.Request.Context(), clazzId, memberId, userID.(string)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

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
