package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ClassStatus 班级状态
type ClassStatus int8

const (
	ClassStatusNotStarted ClassStatus = iota // 未开始
	ClassStatusActive                        // 进行中
	ClassStatusEnded                         // 已结束
)

// TaskStatus 任务状态
type TaskStatus int8

const (
	TaskStatusNotStarted TaskStatus = iota // 未开始
	TaskStatusActive                       // 进行中
	TaskStatusEnded                        // 已结束
)

// TaskType 任务类型
type TaskType int8

const (
	TaskTypeProblem TaskType = 1 // 题单
	TaskTypeVideo   TaskType = 2 // 视频
)

// Clazz 班级嵌入文档,TODO 根据业务需求判断是否一个班级要有对应的教师，且教师必须是被邀请加入了课程，是否有多个教师
type Clazz struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id" swaggertype:"string" example:"507f1f77bcf86cd799439011"`
	Name          string               `bson:"name" json:"name" binding:"required" example:"算法与数据结构 - 第1班"`
	Description   string               `bson:"description,omitempty" json:"description,omitempty" example:"算法与数据结构课程的第1个教学班"`
	CourseId      primitive.ObjectID   `bson:"course_id" json:"course_id" binding:"required" swaggertype:"string" example:"507f1f77bcf86cd799439012"`
	Schedule      string               `bson:"schedule,omitempty" json:"schedule,omitempty" example:"周二 14:00-16:00"`
	MemberIDs     []primitive.ObjectID `bson:"member_ids,omitempty" json:"member_ids,omitempty" swaggertype:"array,string" example:"507f1f77bcf86cd799439013,507f1f77bcf86cd799439014"`
	TeacherIds    []primitive.ObjectID `bson:"teacher_ids,omitempty" json:"teacher_ids,omitempty" swaggertype:"array,string" example:"507f1f77bcf86cd799439015"`
	RequireInvite bool                 `bson:"require_invite" json:"require_invite" example:"false"`
	MaxMembers    int                  `bson:"max_members,omitempty" json:"max_members,omitempty" example:"50"`
	AddNums       int                  `bson:"add_nums" json:"add_nums" example:"25"`
	Status        ClassStatus          `bson:"status" json:"status" example:"1"`
	CTime         time.Time            `bson:"ctime" json:"ctime" example:"2024-10-26T10:00:00Z"`
	CID           primitive.ObjectID   `bson:"c_id" json:"c_id" swaggertype:"string" example:"507f1f77bcf86cd799439016"`
	MTime         time.Time            `bson:"mtime" json:"mtime" example:"2024-10-26T10:00:00Z"`
}

// Task 课程任务模型
type Task struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id" swaggertype:"string" example:"507f1f77bcf86cd799439011"`
	Title       string               `bson:"title" json:"title" binding:"required" example:"第1周编程作业"`
	Description string               `bson:"description,omitempty" json:"description,omitempty" example:"本周需要完成2道编程题目，请认真阅读题目要求并提交代码。"`
	Type        TaskType             `bson:"type" json:"type" binding:"required" example:"1"`
	StartTime   time.Time            `bson:"start_time" json:"start_time" binding:"required" example:"2024-10-26T10:00:00Z"`
	EndTime     *time.Time           `bson:"end_time" json:"end_time" example:"2024-11-02T23:59:59Z"`
	RelationIDs []primitive.ObjectID `bson:"relation_ids,omitempty" json:"relation_ids,omitempty" swaggertype:"array,string" example:"507f1f77bcf86cd799439011,507f1f77bcf86cd799439012"`
	Status      TaskStatus           `bson:"status" json:"status" example:"1"`
	CourseId    primitive.ObjectID   `bson:"course_id" json:"course_id" swaggertype:"string" example:"507f1f77bcf86cd799439015"`
	ClazzId     primitive.ObjectID   `bson:"clazz_id" json:"clazz_id" swaggertype:"string" example:"507f1f77bcf86cd799439016"`
	CTime       time.Time            `bson:"ctime" json:"ctime" example:"2024-10-26T10:00:00Z"`
	CID         primitive.ObjectID   `bson:"c_id" json:"c_id" swaggertype:"string" example:"507f1f77bcf86cd799439017"`
	MTime       time.Time            `bson:"mtime" json:"mtime" example:"2024-10-26T10:00:00Z"`
}

// TaskResponse 任务响应模型
type TaskResponse struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id" swaggertype:"string" example:"507f1f77bcf86cd799439011"`
	Title       string               `bson:"title" json:"title" binding:"required" example:"第1周编程作业"`
	Description string               `bson:"description,omitempty" json:"description,omitempty" example:"本周需要完成2道编程题目，请认真阅读题目要求并提交代码。"`
	Type        TaskType             `bson:"type" json:"type" binding:"required" example:"1"`
	StartTime   time.Time            `bson:"start_time" json:"start_time" binding:"required" example:"2024-10-26T10:00:00Z"`
	EndTime     *time.Time           `bson:"end_time" json:"end_time" example:"2024-11-02T23:59:59Z"`
	Status      TaskStatus           `bson:"status" json:"status" example:"1"`
	CourseId    primitive.ObjectID   `bson:"course_id" json:"course_id" swaggertype:"string" example:"507f1f77bcf86cd799439015"`
	ClazzId     primitive.ObjectID   `bson:"clazz_id" json:"clazz_id" swaggertype:"string" example:"507f1f77bcf86cd799439016"`
	CTime       time.Time            `bson:"ctime" json:"ctime" example:"2024-10-26T10:00:00Z"`
	CID         primitive.ObjectID   `bson:"c_id" json:"c_id" swaggertype:"string" example:"507f1f77bcf86cd799439017"`
	MTime       time.Time            `bson:"mtime" json:"mtime" example:"2024-10-26T10:00:00Z"`
	State       int                  `json:"state" example:"0"`      // 0: 未完成, 1: 已完成
	RelationIDs []primitive.ObjectID `json:"relation_ids,omitempty"` // 关联ID列表
	Questions   []QuestionDetail     `json:"questions"`              // 题目详情列表
}

// QuestionDetail 题目详情
type QuestionDetail struct {
	ID         primitive.ObjectID `json:"id" swaggertype:"string"`
	UniqueID   int64              `json:"unique_id"` // 题目在任务中的唯一ID
	Title      string             `json:"title"`
	Difficulty string             `json:"difficulty"`
	Completed  bool               `json:"completed"` // 是否已完成
}

// UserTask 用户任务状态模型
type UserTask struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID        primitive.ObjectID `bson:"task_id" json:"task_id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	State         int                `bson:"state" json:"state"`                                   // 0: 未完成, 1: 已完成
	FinishedCount int                `bson:"finished_count" json:"finished_count"`                 // 完成题目数
	CompletedAt   *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"` // 完成时间
	CTime         time.Time          `bson:"ctime" json:"ctime"`
	MTime         time.Time          `bson:"mtime" json:"mtime"`
}

// UserRelationQuestion 用户关联题目状态模型
type UserRelationQuestion struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RelationID primitive.ObjectID `bson:"relation_id" json:"relation_id"`
	TaskID     primitive.ObjectID `bson:"task_id" json:"task_id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	State      int                `bson:"state" json:"state"` // 0: 未完成, 1: 已完成
	CTime      time.Time          `bson:"ctime" json:"ctime"`
	MTime      time.Time          `bson:"mtime" json:"mtime"`
}

type QrCodeResponse struct {
	QrCode     []byte `json:"qr_code"`
	InviteCode string `json:"invite_code"`
}

// ClazzResponse 班级响应（包含成员信息）
type ClazzResponse struct {
	ClazzId string `json:"clazz_id"`
	QRCode  string `json:"qrcode,omitempty"`
}

// CanJoin 检查是否可以加入班级
func (cl *Clazz) CanJoin() bool {
	if cl.AddNums >= cl.MaxMembers {
		return false
	}
	return cl.Status == ClassStatusActive
}
