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
	Title        string            `json:"title" binding:"required,min=1,max=200"`
	Description  string            `json:"description" binding:"required,min=10"`
	
	// 题目详情（可选）
	Input        string            `json:"input"`
	Output       string            `json:"output"`
	SampleInput  string            `json:"sample_input"`
	SampleOutput string            `json:"sample_output"`
	Hint         string            `json:"hint"`
	Source       string            `json:"source"`
	Author       string            `json:"author"`
	
	// 难度和标签（必填）
	Difficulty   ProblemDifficulty `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Tags         []string          `json:"tags" binding:"max=10,dive,min=1,max=20"`
	
	// 限制条件（可选，有默认值）
	TimeLimit    *int              `json:"time_limit" binding:"omitempty,min=100,max=10000"`    // 默认1000ms
	MemoryLimit  *int              `json:"memory_limit" binding:"omitempty,min=32,max=1024"`    // 默认256MB
	
	// 状态控制（可选）
	Status       *ProblemStatus    `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsPublic     *bool             `json:"is_public"` // 可选，默认false
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

// PageQueryCourseResponse 分页查询课程响应
type PageQueryCourseResponse struct {
	Total    int64    `json:"total"`
	PageNum  int64    `json:"page_num"`
	PageSize int64    `json:"page_size"`
	Courses  []Course `json:"courses"`
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
	Name          string               `json:"name,omitempty"`
	Description   string               `json:"description,omitempty"`
	Schedule      string               `json:"schedule,omitempty"`
	RequireInvite *bool                `json:"require_invite,omitempty"`
	MaxMembers    *int                 `json:"max_members,omitempty"`
	MemberIDs     []primitive.ObjectID `json:"member_ids,omitempty"`
}

// JoinClazzRequest 加入班级请求
type JoinClazzRequest struct {
	ClazzID    string  `json:"clazz_id" form:"clazzId" binding:"required"`
	InviteCode *string `json:"invite_code,omitempty" form:"invite_code,omitempty"`
}

// AddClazzMemberRequest 添加班级成员请求
type AddClazzMemberRequest struct {
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
	RequireInvite bool                 `bson:"require_invite" json:"require_invite"`
	MaxMembers    int                  `bson:"max_members,omitempty" json:"max_members,omitempty"`
	AddNums       int                  `bson:"add_nums" json:"add_nums"`
	Status        ClassStatus          `bson:"status" json:"status"`
	CTime         time.Time            `bson:"ctime" json:"ctime"`
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

