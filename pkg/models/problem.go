/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: problem.go
@Description: 题目数据模型
*/

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ProblemDifficulty 题目难度
type ProblemDifficulty string

const (
	DifficultyEasy   ProblemDifficulty = "easy"   // 简单
	DifficultyMedium ProblemDifficulty = "medium" // 中等
	DifficultyHard   ProblemDifficulty = "hard"   // 困难
)

// ProblemStatus 题目状态
type ProblemStatus string

const (
	StatusDraft     ProblemStatus = "draft"     // 草稿
	StatusPublished ProblemStatus = "published" // 已发布
	StatusArchived  ProblemStatus = "archived"  // 已归档
)

// Problem 题目模型
type Problem struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title        string             `bson:"title" json:"title" binding:"required"`
	Description  string             `bson:"description" json:"description" binding:"required"`
	Input        string             `bson:"input" json:"input"`
	Output       string             `bson:"output" json:"output"`
	SampleInput  string             `bson:"sample_input" json:"sample_input"`
	SampleOutput string             `bson:"sample_output" json:"sample_output"`
	Hint         string             `bson:"hint" json:"hint"`
	Source       string             `bson:"source" json:"source"`
	Author       string             `bson:"author" json:"author"`
	Difficulty   ProblemDifficulty  `bson:"difficulty" json:"difficulty"`
	TimeLimit    int                `bson:"time_limit" json:"time_limit"`     // 时间限制(ms)
	MemoryLimit  int                `bson:"memory_limit" json:"memory_limit"` // 内存限制(MB)
	Tags         []string           `bson:"tags" json:"tags"`
	Status       ProblemStatus      `bson:"status" json:"status"`
	IsPublic     bool               `bson:"is_public" json:"is_public"`
	ACCount      int                `bson:"ac_count" json:"ac_count"`         // AC 次数
	SubmitCount  int                `bson:"submit_count" json:"submit_count"` // 提交次数
	CreatedBy    primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// TestCase 测试用例
type TestCase struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProblemID   primitive.ObjectID `bson:"problem_id" json:"problem_id"`
	Input       string             `bson:"input" json:"input"`
	Output      string             `bson:"output" json:"output"`
	IsSample    bool               `bson:"is_sample" json:"is_sample"`
	
	// 可选的超时配置（优先级高于题目默认配置）
	TimeLimit   *int               `bson:"time_limit,omitempty" json:"time_limit,omitempty"`     // 时间限制(ms)，可选
	MemoryLimit *int               `bson:"memory_limit,omitempty" json:"memory_limit,omitempty"` // 内存限制(MB)，可选
	
	Score       int                `bson:"score" json:"score"`           // 用例分数（可选，用于部分分）
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// ProblemList 题目列表项
type ProblemList struct {
	ID          primitive.ObjectID `json:"id"`
	Title       string             `json:"title"`
	Difficulty  ProblemDifficulty  `json:"difficulty"`
	Tags        []string           `json:"tags"`
	ACCount     int                `json:"ac_count"`
	SubmitCount int                `json:"submit_count"`
	Status      ProblemStatus      `json:"status"`
	IsPublic    bool               `json:"is_public"`
	CreatedAt   time.Time          `json:"created_at"`
}
