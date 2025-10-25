package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CourseStatus 课程状态
type CourseStatus int8

const (
	CourseStatusDraft    CourseStatus = iota // 草稿
	CourseStatusActive                       // 活跃
	CourseStatusArchived                     // 已归档
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
	CID           primitive.ObjectID   `bson:"c_id" json:"c_id"`
	MTime         time.Time            `bson:"mtime" json:"mtime"`
}

// Course 课程模型
type Course struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Avatar      string               `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Name        string               `bson:"name" json:"name" binding:"required"`
	TeacherIds  []primitive.ObjectID `bson:"teacher_ids,omitempty" json:"teacher_ids,omitempty"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	CreatedBy   primitive.ObjectID   `bson:"created_by" json:"created_by"`
	Status      CourseStatus         `bson:"status" json:"status"`
	CTime       time.Time            `bson:"ctime" json:"ctime"`
	MTime       time.Time            `bson:"mtime" json:"mtime"`
}

// Task 课程任务模型
type Task struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title       string               `bson:"title" json:"title" binding:"required"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	Type        TaskType             `bson:"type" json:"type" binding:"required"`
	StartTime   time.Time            `bson:"start_time" json:"start_time" binding:"required"`
	EndTime     *time.Time           `bson:"end_time" json:"end_time"`
	RelationIDs []primitive.ObjectID `bson:"relation_ids,omitempty" json:"relation_ids,omitempty"`
	Status      TaskStatus           `bson:"status" json:"status"`
	FinishIds   []primitive.ObjectID `bson:"finish_ids" json:"finish_ids"` // 完成任务的成员ID集合
	CourseId    primitive.ObjectID   `bson:"course_id" json:"course_id"`
	ClazzId     primitive.ObjectID   `bson:"clazz_id" json:"clazz_id"`
	CTime       time.Time            `bson:"ctime" json:"ctime"`
	CID         primitive.ObjectID   `bson:"c_id" json:"c_id"`
	MTime       time.Time            `bson:"mtime" json:"mtime"`
}
type RelationsUsers struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID     primitive.ObjectID `bson:"task_id" json:"task_id"`
	RelationID primitive.ObjectID `bson:"relation_id,omitempty" json:"relation_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	CTime      time.Time          `bson:"ctime" json:"ctime"`
}

// 请求和响应结构体

// 注意：所有请求参数定义已迁移到 pkg/models/requests.go
// 这里保留的响应类型将来可能迁移到 pkg/models/responses.go

// CourseResponse 课程响应（包含班级信息）
type CourseResponse struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Avatar        string               `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Name          string               `bson:"name" json:"name" binding:"required"`
	TeacherIds    []primitive.ObjectID `bson:"teacher_ids,omitempty" json:"teacher_ids,omitempty"`
	Description   string               `bson:"description,omitempty" json:"description,omitempty"`
	Status        CourseStatus         `bson:"status" json:"status"`
	CTime         time.Time            `bson:"ctime" json:"ctime"`
	CreatedByUser *UserProfile         `json:"created_by_user,omitempty"`
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

// 成员方法
// AddMember 添加成员到班级
func (cl *Clazz) AddMember(userID primitive.ObjectID) {
	// 检查是否已存在
	for _, memberID := range cl.MemberIDs {
		if memberID == userID {
			return
		}
	}
	cl.MemberIDs = append(cl.MemberIDs, userID)
	cl.AddNums = len(cl.MemberIDs)
	cl.MTime = time.Now()
}

// RemoveMember 从班级移除成员
func (cl *Clazz) RemoveMember(userID primitive.ObjectID) {
	for i, memberID := range cl.MemberIDs {
		if memberID == userID {
			cl.MemberIDs = append(cl.MemberIDs[:i], cl.MemberIDs[i+1:]...)
			cl.AddNums = len(cl.MemberIDs)
			cl.MTime = time.Now()
			return
		}
	}
}

// IsMember 检查用户是否是班级成员
func (cl *Clazz) IsMember(userID primitive.ObjectID) bool {
	for _, memberID := range cl.MemberIDs {
		if memberID == userID {
			return true
		}
	}
	return false
}

// CanJoin 检查是否可以加入班级
func (cl *Clazz) CanJoin() bool {
	if cl.AddNums >= cl.MaxMembers {
		return false
	}
	return cl.Status == ClassStatusActive
}
