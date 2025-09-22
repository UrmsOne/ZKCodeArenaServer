/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service.go
@Description: 服务层入口
*/

package service

import (
	"context"
	"zk-code-arena-server/pkg/utils"
)

// Service 服务层结构
type Service struct {
	UserService    *UserService
	ProblemService *ProblemService
	SubmitService  *SubmitService
	JudgeService   *JudgeService
}

// NewService 创建服务实例
func NewService() *Service {
	return &Service{
		UserService:    NewUserService(),
		ProblemService: NewProblemService(),
		SubmitService:  NewSubmitService(),
		JudgeService:   NewJudgeService(),
	}
}

// Close 关闭服务
func (s *Service) Close(ctx context.Context) error {
	// 关闭数据库连接
	return utils.CloseMongoDB(ctx)
}
