/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: user.go
@Description: 用户数据模型
*/

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole 用户角色
type UserRole string

const (
	RoleAdmin   UserRole = "admin"   // 系统管理员
	RoleTeacher UserRole = "teacher" // 老师
	RoleStudent UserRole = "student" // 学生
	RoleContest UserRole = "contest" // 竞赛发起者
)

// User 用户模型
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username    string             `bson:"username" json:"username" binding:"required"`
	Password    string             `bson:"password" json:"password" binding:"required"`
	Email       string             `bson:"email" json:"email" binding:"required,email"`
	RealName    string             `bson:"real_name" json:"real_name"`
	StudentID   string             `bson:"student_id" json:"student_id"`
	Role        UserRole           `bson:"role" json:"role"`
	Avatar      string             `bson:"avatar" json:"avatar"`
	Bio         string             `bson:"bio" json:"bio"`
	School      string             `bson:"school" json:"school"`
	Major       string             `bson:"major" json:"major"`
	Grade       string             `bson:"grade" json:"grade"`
	Class       string             `bson:"class" json:"class"`
	Phone       string             `bson:"phone" json:"phone"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	LastLoginAt *time.Time         `bson:"last_login_at" json:"last_login_at"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserProfile 用户资料（不包含敏感信息）
type UserProfile struct {
	ID        primitive.ObjectID `json:"id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	RealName  string             `json:"real_name"`
	StudentID string             `json:"student_id"`
	Role      UserRole           `json:"role"`
	Avatar    string             `json:"avatar"`
	Bio       string             `json:"bio"`
	School    string             `json:"school"`
	Major     string             `json:"major"`
	Grade     string             `json:"grade"`
	Class     string             `json:"class"`
	IsActive  bool               `json:"is_active"`
	CreatedAt time.Time          `json:"created_at"`
}

// ToProfile 转换为用户资料
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		RealName:  u.RealName,
		StudentID: u.StudentID,
		Role:      u.Role,
		Avatar:    u.Avatar,
		Bio:       u.Bio,
		School:    u.School,
		Major:     u.Major,
		Grade:     u.Grade,
		Class:     u.Class,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}
}
