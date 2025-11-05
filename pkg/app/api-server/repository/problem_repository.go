/*
@Author: omenkk7
@Date: 2025/10/25
@Description: Problem数据访问层 - 题目相关的数据操作
*/

package repository

import (
	"context"
	"fmt"
	"time"

	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProblemRepository Problem数据访问层
type ProblemRepository struct {
	*BaseRepository
}

// NewProblemRepository 创建ProblemRepository实例
func NewProblemRepository() *ProblemRepository {
	return &ProblemRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// CreateProblem 创建题目
func (r *ProblemRepository) CreateProblem(ctx context.Context, problem *models.Problem) error {
	// 设置系统字段
	problem.ID = primitive.NewObjectID()
	problem.CreatedAt = time.Now()
	problem.UpdatedAt = time.Now()

	// 应用业务规则
	if err := r.validateProblemCreation(problem); err != nil {
		return err
	}

	_, err := r.InsertOne(ctx, "problems", problem)
	if err != nil {
		utils.Logger.Errorf("CreateProblem: 创建题目失败, title=%s, error=%v", problem.Title, err)
		return fmt.Errorf("创建题目失败: %w", err)
	}

	utils.Logger.Infof("CreateProblem: 题目创建成功, id=%s, title=%s", problem.ID.Hex(), problem.Title)
	return nil
}

// GetByID 根据ID查询题目
func (r *ProblemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Problem, error) {
	var problem models.Problem
	err := r.FindOne(ctx, "problems", bson.M{"_id": id}, &problem)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

// DeleteProblem 删除题目
func (r *ProblemRepository) DeleteProblem(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.DeleteOne(ctx, "problems", bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("题目不存在")
	}

	return nil
}

// GetProblems 查询题目列表（带权限控制）
func (r *ProblemRepository) GetProblems(
	ctx context.Context,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	includePrivate bool,
	role models.UserRole,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 构建查询条件
	filter := r.buildProblemFilter(difficulty, tags, includePrivate, role, userID)

	// 获取总数
	total, err := r.CountDocuments(ctx, "problems", filter)
	if err != nil {
		return nil, 0, fmt.Errorf("统计题目数量失败: %w", err)
	}

	// 分页查询选项
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.Find(ctx, "problems", filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询题目列表失败: %w", err)
	}
	defer cursor.Close(ctx)

	var problems []*models.ProblemList
	for cursor.Next(ctx) {
		var problem models.Problem
		if err := cursor.Decode(&problem); err != nil {
			return nil, 0, fmt.Errorf("解析题目数据失败: %w", err)
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

// GetProblemsWithUserStatus 查询题目列表并填充用户状态
func (r *ProblemRepository) GetProblemsWithUserStatus(
	ctx context.Context,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	includePrivate bool,
	role models.UserRole,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 先获取题目列表
	problems, total, err := r.GetProblems(ctx, page, pageSize, difficulty, tags, includePrivate, role, userID)
	if err != nil {
		return nil, 0, err
	}

	// 填充用户状态（仅当用户已登录时）
	if userID != nil && len(problems) > 0 {
		utils.Logger.Debugf("GetProblemsWithUserStatus: 开始填充用户状态, userID=%s, problemCount=%d", userID.Hex(), len(problems))
		if err := r.fillUserProblemStatus(ctx, problems, *userID); err != nil {
			utils.Logger.Errorf("GetProblemsWithUserStatus: 填充用户状态失败, userID=%s, error=%v", userID.Hex(), err)
			// 不影响主要功能，只记录错误，但确保用户知道状态查询失败
		} else {
			utils.Logger.Debugf("GetProblemsWithUserStatus: 用户状态填充成功, userID=%s", userID.Hex())
		}
	} else if userID == nil {
		utils.Logger.Debugf("GetProblemsWithUserStatus: 未登录用户，跳过状态填充")
	}

	return problems, total, nil
}

// SearchProblems 搜索题目
func (r *ProblemRepository) SearchProblems(
	ctx context.Context,
	keyword string,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	includePrivate bool,
	role models.UserRole,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 构建搜索过滤条件
	filter := r.buildProblemFilter(difficulty, tags, includePrivate, role, userID)

	// 添加关键词搜索
	if keyword != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": keyword, "$options": "i"}},
			{"description": bson.M{"$regex": keyword, "$options": "i"}},
			{"tags": bson.M{"$regex": keyword, "$options": "i"}},
		}
	}

	// 获取总数
	total, err := r.CountDocuments(ctx, "problems", filter)
	if err != nil {
		return nil, 0, fmt.Errorf("统计搜索结果数量失败: %w", err)
	}

	// 分页查询选项
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.Find(ctx, "problems", filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索题目失败: %w", err)
	}
	defer cursor.Close(ctx)

	var problems []*models.ProblemList
	for cursor.Next(ctx) {
		var problem models.Problem
		if err := cursor.Decode(&problem); err != nil {
			return nil, 0, fmt.Errorf("解析搜索结果失败: %w", err)
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

	// 填充用户状态
	if userID != nil && len(problems) > 0 {
		if err := r.fillUserProblemStatus(ctx, problems, *userID); err != nil {
			utils.Logger.Errorf("SearchProblems: 填充用户状态失败, error=%v", err)
		}
	}

	return problems, total, nil
}

// UpdateProblemStats 更新题目统计信息
func (r *ProblemRepository) UpdateProblemStats(ctx context.Context, problemID primitive.ObjectID, isAC bool) error {
	update := bson.M{
		"$inc": bson.M{
			"submit_count": 1,
		},
	}

	if isAC {
		update["$inc"].(bson.M)["ac_count"] = 1
	}

	coll := r.db.Collection("problems")
	_, err := coll.UpdateOne(ctx, bson.M{"_id": problemID}, update)

	if err != nil {
		utils.Logger.Errorf("UpdateProblemStats: 更新题目统计失败, problemID=%s, error=%v", problemID.Hex(), err)
		return fmt.Errorf("更新题目统计失败: %w", err)
	}

	return nil
}

// GetUserProblemStatuses 批量获取用户对一组题目的最新状态
func (r *ProblemRepository) GetUserProblemStatuses(ctx context.Context, userID primitive.ObjectID, problemIDs []primitive.ObjectID) (map[primitive.ObjectID]models.UserProblemStatus, error) {
	if len(problemIDs) == 0 {
		return make(map[primitive.ObjectID]models.UserProblemStatus), nil
	}

	pipeline := []bson.M{
		// 1. 匹配用户和题目
		{
			"$match": bson.M{
				"user_id":    userID,
				"problem_id": bson.M{"$in": problemIDs},
			},
		},
		// 2. 按题目分组，获取最新状态
		{
			"$group": bson.M{
				"_id":           "$problem_id",
				"latest_status": bson.M{"$last": "$status"},
				"created_at":    bson.M{"$max": "$created_at"},
			},
		},
	}

	cursor, err := r.Aggregate(ctx, "submits", pipeline)
	if err != nil {
		utils.Logger.Errorf("GetUserProblemStatuses: 聚合查询失败, userID=%s, error=%v", userID.Hex(), err)
		return nil, fmt.Errorf("查询用户题目状态失败: %w", err)
	}
	defer cursor.Close(ctx)

	statusMap := make(map[primitive.ObjectID]models.UserProblemStatus)

	for cursor.Next(ctx) {
		var result struct {
			ID           primitive.ObjectID `bson:"_id"`
			LatestStatus string             `bson:"latest_status"`
		}

		if err := cursor.Decode(&result); err != nil {
			utils.Logger.Warnf("GetUserProblemStatuses: 解析结果失败, error=%v", err)
			continue
		}

		// 转换状态
		var userStatus models.UserProblemStatus
		if result.LatestStatus == string(models.StatusAccepted) {
			userStatus = models.UserStatusAccepted
		} else {
			userStatus = models.UserStatusAttempted
		}

		statusMap[result.ID] = userStatus
	}

	// 对于没有提交记录的题目，状态为未尝试
	for _, problemID := range problemIDs {
		if _, exists := statusMap[problemID]; !exists {
			statusMap[problemID] = models.UserStatusNotAttempted
		}
	}

	utils.Logger.Debugf("GetUserProblemStatuses: 查询完成, userID=%s, 返回%d个状态", userID.Hex(), len(statusMap))
	return statusMap, nil
}

// UpdateProblemFromRequest 根据UpdateProblemRequest更新题目
func (p *ProblemRepository) UpdateProblemFromRequest(ctx context.Context, problemID primitive.ObjectID, req *models.UpdateProblemRequest) error {
	// 构建更新字段map
	updateFields := bson.M{}

	// 添加请求中传入的字段
	if req.Title != nil {
		updateFields["title"] = *req.Title
	}
	if req.Description != nil {
		updateFields["description"] = *req.Description
	}
	if req.Input != nil {
		updateFields["input"] = *req.Input
	}
	if req.Output != nil {
		updateFields["output"] = *req.Output
	}
	if req.SampleInput != nil {
		updateFields["sample_input"] = *req.SampleInput
	}
	if req.SampleOutput != nil {
		updateFields["sample_output"] = *req.SampleOutput
	}
	if req.Hint != nil {
		updateFields["hint"] = *req.Hint
	}
	if req.Source != nil {
		updateFields["source"] = *req.Source
	}
	if req.Author != nil {
		updateFields["author"] = *req.Author
	}
	if req.Difficulty != nil {
		updateFields["difficulty"] = *req.Difficulty
	}
	if req.TimeLimit != nil {
		updateFields["time_limit"] = *req.TimeLimit
	}
	if req.MemoryLimit != nil {
		updateFields["memory_limit"] = *req.MemoryLimit
	}
	if req.Tags != nil {
		tags := *req.Tags
		if tags == nil {
			tags = []string{}
		}
		updateFields["tags"] = tags
	}
	if req.Status != nil {
		updateFields["status"] = *req.Status

		// 业务规则：草稿状态强制私有
		if *req.Status == models.StatusDraft {
			updateFields["is_public"] = false
		}
	}
	if req.IsPublic != nil {
		// 如果同时设置了 status 为 draft，则 isPublic 已经在上面被设置为 false
		if req.Status == nil || *req.Status != models.StatusDraft {
			updateFields["is_public"] = *req.IsPublic
		}
	}

	// 必须更新的字段：更新时间
	updateFields["updated_at"] = time.Now()

	// 执行更新操作（直接传入 updateFields，UpdateOne 会自动添加 $set）
	// 注意：不要再手动包装 bson.M{"$set": ...}，会导致双重 $set 错误
	filter := bson.M{"_id": problemID}
	update := bson.M{"$set": updateFields}

	coll := p.db.Collection("problems")
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.Logger.Errorf("UpdateProblemFromRequest: 更新失败, error=%v", err)
		return fmt.Errorf("数据库更新失败: %w", err)
	}

	// 检查是否找到文档
	if result.MatchedCount == 0 {
		return fmt.Errorf("题目不存在")
	}

	return nil
}

// buildProblemFilter 构建题目查询过滤条件
func (r *ProblemRepository) buildProblemFilter(
	difficulty models.ProblemDifficulty,
	tags []string,
	includePrivate bool,
	role models.UserRole,
	userID *primitive.ObjectID,
) bson.M {
	filter := bson.M{}

	if !includePrivate {
		// 公开模式：只显示公开且已发布的题目
		filter["is_public"] = true
		filter["status"] = models.StatusPublished
	} else {
		// 私有模式：根据角色决定可见范围
		switch role {
		case models.RoleAdmin:
			// 管理员可以查看全部题目
		case models.RoleTeacher:
			if userID != nil {
				// 教师查看自己创建的所有题目
				filter["created_by"] = *userID
			} else {
				// 无用户ID，退回到公开模式
				filter["is_public"] = true
				filter["status"] = models.StatusPublished
			}
		default:
			// 学生或未认证用户，只能看公开已发布的题目
			filter["is_public"] = true
			filter["status"] = models.StatusPublished
		}
	}

	if difficulty != "" {
		filter["difficulty"] = difficulty
	}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	return filter
}

// fillUserProblemStatus 填充题目列表的用户状态
func (r *ProblemRepository) fillUserProblemStatus(ctx context.Context, problems []*models.ProblemList, userID primitive.ObjectID) error {
	// 收集所有题目ID
	problemIDs := make([]primitive.ObjectID, len(problems))
	for i, problem := range problems {
		problemIDs[i] = problem.ID
	}

	// 批量查询用户状态
	statusMap, err := r.GetUserProblemStatuses(ctx, userID, problemIDs)
	if err != nil {
		return err
	}

	// 填充状态到题目列表
	for _, problem := range problems {
		if status, exists := statusMap[problem.ID]; exists {
			problem.UserStatus = &status
		}
	}

	return nil
}

// validateProblemCreation 验证题目创建参数
func (r *ProblemRepository) validateProblemCreation(problem *models.Problem) error {
	if problem.Title == "" {
		return fmt.Errorf("题目标题不能为空")
	}

	// 设置默认Status（如果未传）
	if problem.Status == "" {
		problem.Status = models.StatusDraft
	}

	// 应用Status与IsPublic关联规则
	if problem.Status == models.StatusDraft {
		problem.IsPublic = false
		utils.Logger.Infof("validateProblemCreation: 草稿状态强制设为私有, title=%s", problem.Title)
	}

	return nil
}

// validateProblemUpdate 验证题目更新参数
func (r *ProblemRepository) validateProblemUpdate(req *models.UpdateProblemRequest) error {
	// 业务规则：草稿状态强制私有
	if req.Status != nil && *req.Status == models.StatusDraft {
		// 通过修改请求对象来应用业务规则
		falseValue := false
		req.IsPublic = &falseValue
		utils.Logger.Warnf("validateProblemUpdate: 草稿状态不能公开，强制设为私有")
	}

	return nil
}
