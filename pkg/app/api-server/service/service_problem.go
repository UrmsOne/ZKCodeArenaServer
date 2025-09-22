/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service_problem.go
@Description: 题目服务
*/

package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"
)

type ProblemService struct{}

func NewProblemService() *ProblemService {
	return &ProblemService{}
}

// CreateProblem 创建题目
func (s *ProblemService) CreateProblem(ctx context.Context, problem *models.Problem) error {
	problem.ID = primitive.NewObjectID()
	problem.CreatedAt = time.Now()
	problem.UpdatedAt = time.Now()
	problem.ACCount = 0
	problem.SubmitCount = 0

	collection := utils.GetCollection("problems")
	_, err := collection.InsertOne(ctx, problem)
	return err
}

// GetProblemByID 根据ID获取题目
func (s *ProblemService) GetProblemByID(ctx context.Context, id primitive.ObjectID) (*models.Problem, error) {
	collection := utils.GetCollection("problems")
	var problem models.Problem
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&problem)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

// GetProblems 获取题目列表
func (s *ProblemService) GetProblems(ctx context.Context, page, pageSize int, difficulty models.ProblemDifficulty, tags []string, isPublic bool) ([]*models.ProblemList, int64, error) {
	collection := utils.GetCollection("problems")
	
	// 构建查询条件
	filter := bson.M{}
	if difficulty != "" {
		filter["difficulty"] = difficulty
	}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}
	if isPublic {
		filter["is_public"] = true
		filter["status"] = models.StatusPublished
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

	var problems []*models.ProblemList
	for cursor.Next(ctx) {
		var problem models.Problem
		if err := cursor.Decode(&problem); err != nil {
			return nil, 0, err
		}
		
		problems = append(problems, &models.ProblemList{
			ID:          problem.ID,
			Title:       problem.Title,
			Difficulty:  problem.Difficulty,
			Tags:        problem.Tags,
			ACCount:     problem.ACCount,
			SubmitCount: problem.SubmitCount,
			Status:      problem.Status,
			IsPublic:    problem.IsPublic,
			CreatedAt:   problem.CreatedAt,
		})
	}

	return problems, total, nil
}

// UpdateProblem 更新题目
func (s *ProblemService) UpdateProblem(ctx context.Context, problem *models.Problem) error {
	problem.UpdatedAt = time.Now()
	collection := utils.GetCollection("problems")
	_, err := collection.ReplaceOne(ctx, bson.M{"_id": problem.ID}, problem)
	return err
}

// DeleteProblem 删除题目
func (s *ProblemService) DeleteProblem(ctx context.Context, id primitive.ObjectID) error {
	collection := utils.GetCollection("problems")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// UpdateProblemStats 更新题目统计信息
func (s *ProblemService) UpdateProblemStats(ctx context.Context, problemID primitive.ObjectID, isAC bool) error {
	collection := utils.GetCollection("problems")
	
	update := bson.M{
		"$inc": bson.M{
			"submit_count": 1,
		},
	}
	
	if isAC {
		update["$inc"].(bson.M)["ac_count"] = 1
	}
	
	_, err := collection.UpdateOne(ctx, bson.M{"_id": problemID}, update)
	return err
}
