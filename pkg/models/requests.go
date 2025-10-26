/*
@Author: omenkk7
@Date: 2025/10/09
@Description: 统一请求参数定义
*/

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==================== 用户模块请求 ====================

// LoginRequest 用户登录请求
type LoginRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新用户个人信息请求
type UpdateProfileRequest struct {
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
	School   string `json:"school"`
	Major    string `json:"major"`
	Grade    string `json:"grade"`
	Class    string `json:"class"`
	Phone    string `json:"phone"`
}

// UpdateUserRequest 更新用户信息请求（管理员）
type UpdateUserRequest struct {
	Role     string `json:"role"`
	IsActive *bool  `json:"is_active"`
}

// ==================== 题目模块请求 ====================

// CreateProblemRequest 创建题目请求
type CreateProblemRequest struct {
	// 基本信息（必填）
	Title       string            `json:"title" binding:"required,min=1,max=200"`
	Description string            `json:"description" binding:"required,min=10"`

	// 题目详情（可选）
	Input        string `json:"input"`
	Output       string `json:"output"`
	SampleInput  string `json:"sample_input"`
	SampleOutput string `json:"sample_output"`
	Hint         string `json:"hint"`
	Source       string `json:"source"`
	Author       string `json:"author"`

	// 难度和标签（必填）
	Difficulty ProblemDifficulty `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Tags       []string          `json:"tags" binding:"max=10,dive,min=1,max=20"`

	// 限制条件（可选，有默认值）
	TimeLimit   *int `json:"time_limit" binding:"omitempty,min=100,max=10000"`  // 默认1000ms
	MemoryLimit *int `json:"memory_limit" binding:"omitempty,min=32,max=1024"` // 默认256MB

	// 状态控制（可选）
	Status   *ProblemStatus `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsPublic *bool          `json:"is_public"` // 可选，默认false
}

// UpdateProblemRequest 更新题目请求
type UpdateProblemRequest struct {
	// 基本信息（可选）
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,min=10"`

	// 题目详情（可选）
	Input        *string `json:"input"`
	Output       *string `json:"output"`
	SampleInput  *string `json:"sample_input"`
	SampleOutput *string `json:"sample_output"`
	Hint         *string `json:"hint"`
	Source       *string `json:"source"`
	Author       *string `json:"author"`

	// 难度和标签（可选）
	Difficulty *ProblemDifficulty `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	Tags       *[]string          `json:"tags" binding:"omitempty,max=10,dive,min=1,max=20"`

	// 限制条件（可选）
	TimeLimit   *int `json:"time_limit" binding:"omitempty,min=100,max=10000"`
	MemoryLimit *int `json:"memory_limit" binding:"omitempty,min=32,max=1024"`

	// 状态控制（可选）
	Status   *ProblemStatus `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsPublic *bool          `json:"is_public"`
}

// RunCodeRequest 运行代码请求
// 注意：这个已在 service 包中定义，这里只是引用说明

// ==================== 提交模块请求 ====================

// SubmitCodeRequest 提交代码请求
type SubmitCodeRequest struct {
	ProblemID primitive.ObjectID `json:"problem_id" binding:"required"`
	Code      string             `json:"code" binding:"required"`
	Language  string             `json:"language" binding:"required"`
}

// ==================== 测试用例模块请求 ====================

// CreateTestCaseRequest 创建测试用例请求
type CreateTestCaseRequest struct {
	ProblemID   string `json:"problem_id" binding:"required"`
	Input       string `json:"input" binding:"required"`
	Output      string `json:"output" binding:"required"`
	IsSample    bool   `json:"is_sample"`
	TimeLimit   *int   `json:"time_limit,omitempty"`   // 可选的超时配置（ms）
	MemoryLimit *int   `json:"memory_limit,omitempty"` // 可选的内存限制（MB）
	Score       int    `json:"score,omitempty"`        // 用例分数
}

// UpdateTestCaseRequest 更新测试用例请求
type UpdateTestCaseRequest struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	IsSample    *bool  `json:"is_sample,omitempty"`
	TimeLimit   *int   `json:"time_limit,omitempty"`
	MemoryLimit *int   `json:"memory_limit,omitempty"`
	Score       *int   `json:"score,omitempty"`
}

// BatchCreateTestCasesRequest 批量创建测试用例请求
type BatchCreateTestCasesRequest struct {
	ProblemID string      `json:"problem_id" binding:"required"`
	TestCases []TestCase  `json:"test_cases" binding:"required,min=1"`
}

// ==================== 课程模块请求 ====================

// CreateCourseRequest 创建课程请求
type CreateCourseRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// UpdateCourseRequest 更新课程请求
type UpdateCourseRequest struct {
	ID          string        `bson:"_id,omitempty" json:"id" binding:"required"`
	Name        *string       `json:"name,omitempty"`
	Description *string       `json:"description,omitempty"`
	Status      *CourseStatus `json:"status,omitempty"`
}

// PageQueryCourseRequest 分页查询课程请求
type PageQueryCourseRequest struct {
	PageNum  *int64  `json:"page_num,omitempty"`
	PageSize *int64  `json:"page_size,omitempty"`
	Status   *int8   `json:"status,omitempty"`
	Name     *string `json:"name,omitempty"`
}

// PageQueryTeacherCoursesRequest 分页查询老师加入的课程请求
type PageQueryTeacherCoursesRequest struct {
	PageNum  *int64  `json:"page_num,omitempty"`
	PageSize *int64  `json:"page_size,omitempty"`
	Status   *int8   `json:"status,omitempty"`
	Name     *string `json:"name,omitempty"`
}

// PageQueryCourseResponse 分页查询课程响应
type PageQueryCourseResponse struct {
	Total    int64    `json:"total"`
	PageNum  int64    `json:"page_num"`
	PageSize int64    `json:"page_size"`
	Courses  []Course `json:"courses"`
}

// AddCourseTeachersRequest 课程添加教师请求
type AddCourseTeachersRequest struct {
	CourseId   string   `json:"course_id" binding:"required"`
	TeacherIds []string `json:"teacher_ids" binding:"required"`
}

// RemoveCourseTeachersRequest 课程删除教师请求
type RemoveCourseTeachersRequest struct {
	CourseId   string   `json:"course_id" binding:"required"`
	TeacherIds []string `json:"teacher_ids" binding:"required"`
}

// ==================== 班级模块请求 ====================

// CreateClazzRequest 创建班级请求
type CreateClazzRequest struct {
	Name          string `json:"name" binding:"required"`
	CourseId      string `json:"course_id" binding:"required"`
	Description   string `json:"description,omitempty"`
	Schedule      string `json:"schedule,omitempty"`
	RequireInvite bool   `json:"require_invite" binding:"required"`
	MaxMembers    *int   `json:"max_members,omitempty"`
}

// UpdateClazzRequest 更新班级请求
type UpdateClazzRequest struct {
	ClazzID       string `bson:"clazz_id,omitempty" json:"clazz_id" binding:"required"`
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	Schedule      string `json:"schedule,omitempty"`
	RequireInvite *bool  `json:"require_invite,omitempty"`
	MaxMembers    *int   `json:"max_members,omitempty"`
}

// JoinClazzRequest 加入班级请求
type JoinClazzRequest struct {
	ClazzID    string  `json:"clazz_id" form:"clazzId" binding:"required"`
	InviteCode *string `json:"invite_code,omitempty" form:"invite_code,omitempty"`
}

// AddClazzMemberRequest 添加班级成员请求
type AddClazzMemberRequest struct {
	ClazzID  string `json:"clazz_id" form:"clazzId" binding:"required"`
	MemberID string `json:"member_id" binding:"required"`
}

// RemoveClazzMembersRequest 批量移除班级成员请求
type RemoveClazzMembersRequest struct {
	ClazzID   string   `json:"clazz_id" binding:"required"`
	MemberIDs []string `json:"member_ids" binding:"required"`
}

// GetClazzResponse 获取班级响应
type GetClazzResponse struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name          string               `bson:"name" json:"name" binding:"required"`
	Description   string               `bson:"description,omitempty" json:"description,omitempty"`
	CourseId      primitive.ObjectID   `bson:"course_id" json:"course_id" binding:"required"`
	Schedule      string               `bson:"schedule,omitempty" json:"schedule,omitempty"`
	MemberIDs     []primitive.ObjectID `bson:"member_ids,omitempty" json:"member_ids,omitempty"`
	TeacherIds    []primitive.ObjectID `bson:"teacher_ids,omitempty" json:"teacher_ids,omitempty"`
	RequireInvite bool                 `bson:"require_invite" json:"require_invite"`
	MaxMembers    int                  `bson:"max_members,omitempty" json:"max_members,omitempty"`
	AddNums       int                  `bson:"add_nums" json:"add_nums"`
	Status        ClassStatus          `bson:"status" json:"status"`
	CTime         time.Time            `bson:"ctime" json:"ctime"`
}

// AddClazzTeachersRequest 班级添加教师请求
type AddClazzTeachersRequest struct {
	ClazzId    string   `json:"clazz_id" binding:"required"`
	TeacherIds []string `json:"teacher_ids" binding:"required"`
}

// RemoveClazzTeachersRequest 班级移除教师请求
type RemoveClazzTeachersRequest struct {
	ClazzId    string   `json:"clazz_id" binding:"required"`
	TeacherIds []string `json:"teacher_ids" binding:"required"`
}

// ==================== 任务模块请求 ====================

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Title       string               `json:"title" binding:"required"`
	Description string               `json:"description,omitempty"`
	Type        TaskType             `json:"type" binding:"required"`
	StartTime   time.Time            `json:"start_time" binding:"required"`
	EndTime     time.Time            `json:"end_time" binding:"required"`
	RelationIDs []primitive.ObjectID `json:"relation_ids,omitempty"`
}

// AddTaskRequest 添加任务请求
type AddTaskRequest struct {
	CourseId    string     `json:"course_id" binding:"required"`
	ClazzId     string     `bson:"clazz_id" json:"clazz_id" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description,omitempty"`
	Type        TaskType   `json:"type" binding:"required"`
	StartTime   time.Time  `json:"start_time" binding:"required"` // 必须设置开始时间
	EndTime     *time.Time `json:"end_time"`                      // 如果不设置则没有结束时间
	RelationIDs []string   `bson:"relation_ids,omitempty" json:"relation_ids,omitempty"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	ID          string     `bson:"_id,omitempty" json:"task_id" binding:"required"`
	CourseId    string     `bson:"course_id" json:"course_id,omitempty" binding:"required"`
	ClazzId     string     `bson:"clazz_id" json:"clazz_id,omitempty" binding:"required"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Type        *TaskType  `json:"type,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	RelationIDs *[]string  `bson:"relation_ids,omitempty" json:"relation_ids,omitempty"`
}

// FinishTaskRequest 完成任务请求
type FinishTaskRequest struct {
	RelationID string `bson:"relation_id" json:"relation_id" binding:"required"`
	TaskID     string `bson:"task_id" json:"task_id" binding:"required"`
	ClazzID    string `bson:"clazz_id" json:"clazz_id" binding:"required"`
}

// ==================== 题目详情响应模型 ====================

// ProblemDetailResponse 题目详情聚合响应
type ProblemDetailResponse struct {
	Problem     *Problem   `json:"problem"`      // 题目基本信息
	SampleCases []TestCase `json:"sample_cases"` // 示例测试用例列表
}

// ==================== 题目查询条件模型 ====================

// ProblemQueryCondition 题目查询条件 (内部使用)
type ProblemQueryCondition struct {
	Page           int                `json:"page"`
	PageSize       int                `json:"page_size"`
	Difficulty     ProblemDifficulty  `json:"difficulty"`
	Tags           []string           `json:"tags"`
	IncludePrivate bool               `json:"include_private"`
	Role           UserRole           `json:"role"`
	UserID         *primitive.ObjectID `json:"user_id"`
	// 注意：Status和CreatedBy筛选将在后续版本中支持
	// Status         ProblemStatus      `json:"status"`      // 管理员端专用（待实现）
	// CreatedBy      *primitive.ObjectID `json:"created_by"`  // 管理员端专用（待实现）
}

// ==================== 提交状态响应模型 ====================

// SubmitStatusResponse 轻量级提交状态响应
type SubmitStatusResponse struct {
	ID        primitive.ObjectID `json:"id"`         // 提交ID
	Status    SubmitStatus      `json:"status"`     // 当前状态
	Progress  *JudgeProgress    `json:"progress"`   // 判题进度（可选）
	Message   string            `json:"message"`    // 状态描述信息
	UpdatedAt time.Time         `json:"updated_at"` // 最后更新时间
	
	// 完成后的基本结果信息（避免返回完整详细结果）
	TimeUsed   *int `json:"time_used,omitempty"`   // 时间使用(ms)
	MemoryUsed *int `json:"memory_used,omitempty"` // 内存使用(KB)
}

// JudgeProgress 判题进度信息
type JudgeProgress struct {
	CurrentTestCase int `json:"current_test_case"` // 当前测试用例索引（从1开始）
	TotalTestCases  int `json:"total_test_cases"`  // 总测试用例数
	Percentage      int `json:"percentage"`        // 完成百分比 (0-100)
}

// ==================== WebSocket消息模型 ====================

// WSMessage WebSocket消息基础结构
type WSMessage struct {
	Type    string      `json:"type"`              // 消息类型
	SubmitID string     `json:"submit_id"`         // 提交ID
	Data    interface{} `json:"data"`              // 消息数据
	Timestamp time.Time `json:"timestamp"`         // 消息时间戳
}

// WSStatusUpdate WebSocket状态更新消息
type WSStatusUpdate struct {
	Status   SubmitStatus   `json:"status"`           // 新状态
	Progress *JudgeProgress `json:"progress"`         // 进度信息（可选）
	Message  string         `json:"message"`          // 状态描述
	Result   *WSJudgeResult `json:"result,omitempty"` // 完成时的结果摘要
}

// WSJudgeResult WebSocket判题结果摘要（不包含详细测试用例）
type WSJudgeResult struct {
	Status       SubmitStatus `json:"status"`
	TimeUsed     int          `json:"time_used"`     // 时间使用(ms)
	MemoryUsed   int          `json:"memory_used"`   // 内存使用(KB)
	PassedCases  int          `json:"passed_cases"`  // 通过的测试用例数
	TotalCases   int          `json:"total_cases"`   // 总测试用例数
	CompileError string       `json:"compile_error,omitempty"` // 编译错误（如果有）
}