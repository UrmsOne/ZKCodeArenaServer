/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service_user.go
@Description: 用户服务
*/

package service

import (
	"context"
	"errors"
	"time"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	// 检查用户名是否已存在
	existingUser, _ := s.GetUserByUsername(ctx, user.Username)
	if existingUser != nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, _ = s.GetUserByEmail(ctx, user.Email)
	if existingUser != nil {
		return errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// 设置默认值
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true
	if user.Role == "" {
		user.Role = models.RoleStudent
	}

	// 插入数据库
	collection := utils.GetCollection("users")
	_, err = collection.InsertOne(ctx, user)
	return err
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	collection := utils.GetCollection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	collection := utils.GetCollection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByStudentID(ctx context.Context, stdId string) (*models.User, error) {
	collection := utils.GetCollection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"student_id": stdId}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	collection := utils.GetCollection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidateUser 验证用户登录
func (s *UserService) ValidateUser(ctx context.Context, stdId, password string) (*models.User, error) {
	user, err := s.GetUserByStudentID(ctx, stdId)
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}

	if !user.IsActive {
		return nil, errors.New("账户已被禁用")
	}

	// 验证密码（使用 bcrypt）
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err = s.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	collection := utils.GetCollection("users")
	_, err := collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	return err
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(ctx context.Context, page, pageSize int, role models.UserRole) ([]*models.User, int64, error) {
	collection := utils.GetCollection("users")

	// 构建查询条件
	filter := bson.M{}
	if role != "" {
		filter["role"] = role
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	return users, total, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	collection := utils.GetCollection("users")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
