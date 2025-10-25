/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service_submit.go
@Description: 提交服务
*/

package service

import (
	"context"
	"time"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubmitService struct{}

func NewSubmitService() *SubmitService {
	return &SubmitService{}
}

// CreateSubmit 创建提交
func (s *SubmitService) CreateSubmit(ctx context.Context, submit *models.Submit) error {
	submit.ID = primitive.NewObjectID()
	submit.Status = models.StatusPending
	submit.CreatedAt = time.Now()
	submit.UpdatedAt = time.Now()

	collection := utils.GetCollection("submits")
	_, err := collection.InsertOne(ctx, submit)
	return err
}

// GetSubmitByID 根据ID获取提交
func (s *SubmitService) GetSubmitByID(ctx context.Context, id primitive.ObjectID) (*models.Submit, error) {
	collection := utils.GetCollection("submits")
	var submit models.Submit
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&submit)
	if err != nil {
		return nil, err
	}
	return &submit, nil
}

// GetSubmits 获取提交列表
func (s *SubmitService) GetSubmits(ctx context.Context, page, pageSize int, userID, problemID primitive.ObjectID) ([]*models.SubmitList, int64, error) {
	collection := utils.GetCollection("submits")

	// 构建查询条件
	filter := bson.M{}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}
	if !problemID.IsZero() {
		filter["problem_id"] = problemID
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

	var submits []*models.SubmitList
	for cursor.Next(ctx) {
		var submit models.Submit
		if err := cursor.Decode(&submit); err != nil {
			return nil, 0, err
		}

		submits = append(submits, &models.SubmitList{
			ID:         submit.ID,
			ProblemID:  submit.ProblemID,
			UserID:     submit.UserID,
			Language:   submit.Language,
			Status:     submit.Status,
			TimeUsed:   s.getTimeUsed(submit.Result),
			MemoryUsed: s.getMemoryUsed(submit.Result),
			CreatedAt:  submit.CreatedAt,
		})
	}

	return submits, total, nil
}

// UpdateSubmit 更新提交
func (s *SubmitService) UpdateSubmit(ctx context.Context, submit *models.Submit) error {
	submit.UpdatedAt = time.Now()
	collection := utils.GetCollection("submits")
	_, err := collection.ReplaceOne(ctx, bson.M{"_id": submit.ID}, submit)
	return err
}

// getTimeUsed 获取时间使用
func (s *SubmitService) getTimeUsed(result *models.JudgeResult) int {
	if result == nil {
		return 0
	}
	return result.TimeUsed
}

// getMemoryUsed 获取内存使用
func (s *SubmitService) getMemoryUsed(result *models.JudgeResult) int {
	if result == nil {
		return 0
	}
	return result.MemoryUsed
}

// GetSubmitsByStatus 根据状态获取提交列表
func (s *SubmitService) GetSubmitsByStatus(ctx context.Context, status models.SubmitStatus) ([]*models.Submit, error) {
	collection := utils.GetCollection("submits")

	filter := bson.M{"status": status}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var submits []*models.Submit
	if err := cursor.All(ctx, &submits); err != nil {
		return nil, err
	}

	return submits, nil
}

// UpdateSubmitStatus 更新提交状态
func (s *SubmitService) UpdateSubmitStatus(ctx context.Context, submitID primitive.ObjectID, status models.SubmitStatus) error {
	collection := utils.GetCollection("submits")

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": submitID}, update)
	return err
}

// UpdateSubmitResult 更新提交结果
func (s *SubmitService) UpdateSubmitResult(ctx context.Context, submitID primitive.ObjectID, result *models.JudgeResult) error {
	collection := utils.GetCollection("submits")

	update := bson.M{
		"$set": bson.M{
			"status":     result.Status,
			"result":     result,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": submitID}, update)
	return err
}

// BatchUpdateStatus 批量更新状态（用于服务重启恢复）
func (s *SubmitService) BatchUpdateStatus(ctx context.Context, fromStatus, toStatus models.SubmitStatus, errorMsg string) error {
	collection := utils.GetCollection("submits")

	filter := bson.M{"status": fromStatus}

	update := bson.M{
		"$set": bson.M{
			"status":     toStatus,
			"updated_at": time.Now(),
		},
	}

	// 如果是标记为系统错误，添加错误信息到结果中
	if toStatus == models.StatusSystemError && errorMsg != "" {
		update["$set"].(bson.M)["result"] = &models.JudgeResult{
			Status:      toStatus,
			TimeUsed:    0,
			MemoryUsed:  0,
			TestResults: []models.TestResult{},
		}
	}

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}
