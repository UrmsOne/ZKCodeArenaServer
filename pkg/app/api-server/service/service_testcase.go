/*
@Author: omenkk7
@Date: 2025/10/05
@Description: 测试用例服务
*/

package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"zk-code-arena-server/conf"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"
)

// TestCaseService 测试用例服务
type TestCaseService struct {
	collection *mongo.Collection
}

// NewTestCaseService 创建测试用例服务实例
func NewTestCaseService() *TestCaseService {
	return &TestCaseService{
		collection: utils.MongoClient.Database(conf.Config.Mongo.DbName).Collection("test_cases"),
	}
}

// GetTestCasesByProblemID 根据题目 ID 获取测试用例列表
func (s *TestCaseService) GetTestCasesByProblemID(ctx context.Context, problemID primitive.ObjectID) ([]*models.TestCase, error) {
	// 构建查询条件
	filter := bson.M{"problem_id": problemID}

	// 按创建时间排序
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	// 执行查询
	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var testCases []*models.TestCase
	if err = cursor.All(ctx, &testCases); err != nil {
		return nil, err
	}

	return testCases, nil
}

// CreateTestCase 创建测试用例
func (s *TestCaseService) CreateTestCase(ctx context.Context, testCase *models.TestCase) error {
	// 设置创建时间
	testCase.CreatedAt = time.Now()

	// 插入数据库
	result, err := s.collection.InsertOne(ctx, testCase)
	if err != nil {
		return err
	}

	// 设置 ID
	testCase.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

// UpdateTestCase 更新测试用例
func (s *TestCaseService) UpdateTestCase(ctx context.Context, testCase *models.TestCase) error {
	// 构建更新条件
	filter := bson.M{"_id": testCase.ID}

	// 构建更新内容
	update := bson.M{
		"$set": bson.M{
			"input":        testCase.Input,
			"output":       testCase.Output,
			"is_sample":    testCase.IsSample,
			"time_limit":   testCase.TimeLimit,
			"memory_limit": testCase.MemoryLimit,
			"score":        testCase.Score,
		},
	}

	// 执行更新
	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// 检查是否更新成功
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// DeleteTestCase 删除测试用例
func (s *TestCaseService) DeleteTestCase(ctx context.Context, id primitive.ObjectID) error {
	// 构建删除条件
	filter := bson.M{"_id": id}

	// 执行删除
	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	// 检查是否删除成功
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// GetTestCaseByID 根据 ID 获取测试用例
func (s *TestCaseService) GetTestCaseByID(ctx context.Context, id primitive.ObjectID) (*models.TestCase, error) {
	var testCase models.TestCase

	filter := bson.M{"_id": id}
	err := s.collection.FindOne(ctx, filter).Decode(&testCase)
	if err != nil {
		return nil, err
	}

	return &testCase, nil
}

// CountTestCasesByProblemID 统计题目的测试用例数量
func (s *TestCaseService) CountTestCasesByProblemID(ctx context.Context, problemID primitive.ObjectID) (int64, error) {
	filter := bson.M{"problem_id": problemID}
	return s.collection.CountDocuments(ctx, filter)
}
