/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/25 13:09
@Name: server_judge.go
@Description: 评测服务
*/

package server

import (
	"github.com/sirupsen/logrus"
)

// Judge 评测服务结构
type Judge struct {
	lg logrus.FieldLogger
}

// NewJudge 创建评测服务实例
func NewJudge() *Judge {
	return &Judge{
		lg: logrus.New(),
	}
}

// RunJudge 运行评测服务
func (j *Judge) RunJudge() error {
	j.lg.Info("评测服务已启动")

	// 这里可以添加评测服务的具体逻辑
	// 比如启动 go-judge 服务、处理评测队列等

	// 暂时返回 nil，实际实现中应该保持服务运行
	return nil
}
