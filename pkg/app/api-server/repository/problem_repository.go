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
	"go.mongodb.org/mongo-driver/mongo"
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
func (p *ProblemRepository) CreateProblem(ctx context.Context, problem *models.Problem) error {
	// 设置系统字段
	problem.ID = primitive.NewObjectID()
	problem.CreatedAt = time.Now()
	problem.UpdatedAt = time.Now()

	// 应用业务规则
	if err := p.validateProblemCreation(problem); err != nil {
		return err
	}

	_, err := p.InsertOne(ctx, "problems", problem)
	if err != nil {
		utils.Logger.Errorf("CreateProblem: 创建题目失败, title=%s, error=%v", problem.Title, err)
		return fmt.Errorf("创建题目失败: %w", err)
	}

	utils.Logger.Infof("CreateProblem: 题目创建成功, id=%s, title=%s", problem.ID.Hex(), problem.Title)
	return nil
}

// GetByID 根据ID查询题目
func (p *ProblemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Problem, error) {
	var problem models.Problem
	err := p.FindOne(ctx, "problems", bson.M{"_id": id}, &problem)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

// DeleteProblem 删除题目
func (p *ProblemRepository) DeleteProblem(ctx context.Context, id primitive.ObjectID) error {
	result, err := p.DeleteOne(ctx, "problems", bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("题目不存在")
	}

	return nil
}

// GetProblems 查询题目列表 - 使用通用查询框架重构
func (p *ProblemRepository) GetProblems(
	ctx context.Context,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	includePrivate bool,
	role models.UserRole,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 构建查询条件
	filter := p.buildProblemFilter(difficulty, tags, includePrivate, role, userID)

	// 构建查询选项
	sort := p.BuildSort("-created_at")
	opts := p.BuildFindOptionsWithPagination(int64(page), int64(pageSize), sort, nil)

	// 执行泛型分页查询
	problems, total, err := FindWithPaginationTyped[models.Problem](ctx, p.BaseRepository, "problems", filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询题目列表失败: %w", err)
	}

	// 转换为 ProblemList
	var problemList []*models.ProblemList
	for _, problem := range problems {
		problemList = append(problemList, &models.ProblemList{
			ID:          problem.ID,
			UniqueID:    problem.UniqueID,
			Title:       problem.Title,
			Difficulty:  problem.Difficulty,
			ACCount:     problem.ACCount,
			SubmitCount: problem.SubmitCount,
			CreatedAt:   problem.CreatedAt,
		})
	}

	return problemList, total, nil
}

// GetProblemsWithUserStatus 查询题目列表（带用户状态）- 使用通用查询框架重构
func (p *ProblemRepository) GetProblemsWithUserStatus(
	ctx context.Context,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	includePrivate bool,
	role models.UserRole,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 构建查询条件
	filter := p.buildProblemFilter(difficulty, tags, includePrivate, role, userID)

	// 构建查询选项
	sort := p.BuildSort("-created_at")
	opts := p.BuildFindOptionsWithPagination(int64(page), int64(pageSize), sort, nil)

	// 执行泛型分页查询
	problems, total, err := FindWithPaginationTyped[models.Problem](ctx, p.BaseRepository, "problems", filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询题目列表失败: %w", err)
	}

	// 转换为 ProblemList
	var problemList []*models.ProblemList
	for _, problem := range problems {
		problemList = append(problemList, &models.ProblemList{
			ID:          problem.ID,
			UniqueID:    problem.UniqueID,
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

	// 填充用户状态（仅当用户已登录时）
	if userID != nil && len(problemList) > 0 {
		utils.Logger.Debugf("GetProblemsWithUserStatus: 开始填充用户状态, userID=%s, problemCount=%d", userID.Hex(), len(problemList))
		if err := p.fillUserProblemStatus(ctx, problemList, *userID); err != nil {
			utils.Logger.Errorf("GetProblemsWithUserStatus: 填充用户状态失败, userID=%s, error=%v", userID.Hex(), err)
		} else {
			utils.Logger.Debugf("GetProblemsWithUserStatus: 用户状态填充成功, userID=%s", userID.Hex())
		}
	} else if userID == nil {
		utils.Logger.Debugf("GetProblemsWithUserStatus: 未登录用户，跳过状态填充")
	}

	return problemList, total, nil
}

// SearchProblems 搜索题目 - 使用通用查询框架重构
func (p *ProblemRepository) SearchProblems(
	ctx context.Context,
	keyword string,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 使用 BuildKeywordFilter 构建关键字过滤器
	keywordFilter := p.BuildKeywordFilter(keyword, []string{"title", "description"})

	// 构建其他过滤器
	var filters []bson.M
	if len(keywordFilter) > 0 {
		filters = append(filters, keywordFilter)
	}

	// 添加公开和已发布的过滤条件
	filters = append(filters, bson.M{"is_public": true, "status": models.StatusPublished})

	// 添加难度过滤
	if difficulty != "" {
		filters = append(filters, bson.M{"difficulty": difficulty})
	}

	// 添加标签过滤
	if len(tags) > 0 {
		filters = append(filters, bson.M{"tags": bson.M{"$in": tags}})
	}

	// 使用 MergeFilters 合并所有过滤器
	filter := p.MergeFilters(filters...)

	// 构建查询选项
	sort := p.BuildSort("-created_at")
	opts := p.BuildFindOptionsWithPagination(int64(page), int64(pageSize), sort, nil)

	// 执行泛型分页查询
	problems, total, err := FindWithPaginationTyped[models.Problem](ctx, p.BaseRepository, "problems", filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索题目失败: %w", err)
	}

	// 转换为 ProblemList
	var problemList []*models.ProblemList
	for _, problem := range problems {
		problemList = append(problemList, &models.ProblemList{
			ID:          problem.ID,
			UniqueID:    problem.UniqueID,
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
	if userID != nil && len(problemList) > 0 {
		if err := p.fillUserProblemStatus(ctx, problemList, *userID); err != nil {
			utils.Logger.Errorf("SearchProblems: 填充用户状态失败, error=%v", err)
		}
	}

	return problemList, total, nil
}

// UpdateProblemStats 更新题目统计信息
func (p *ProblemRepository) UpdateProblemStats(ctx context.Context, problemID primitive.ObjectID, isAC bool) error {
	update := bson.M{
		"$inc": bson.M{
			"submit_count": 1,
		},
	}

	if isAC {
		update["$inc"].(bson.M)["ac_count"] = 1
	}

	coll := p.db.Collection("problems")
	_, err := coll.UpdateOne(ctx, bson.M{"_id": problemID}, update)

	if err != nil {
		utils.Logger.Errorf("UpdateProblemStats: 更新题目统计失败, problemID=%s, error=%v", problemID.Hex(), err)
		return fmt.Errorf("更新题目统计失败: %w", err)
	}

	return nil
}

// GetUserProblemStatuses 批量获取用户对一组题目的最新状态
func (p *ProblemRepository) GetUserProblemStatuses(ctx context.Context, userID primitive.ObjectID, problemIDs []primitive.ObjectID) (map[primitive.ObjectID]models.UserProblemStatus, error) {
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

	cursor, err := p.Aggregate(ctx, "submits", pipeline)
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
func (p *ProblemRepository) buildProblemFilter(
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
func (p *ProblemRepository) fillUserProblemStatus(ctx context.Context, problems []*models.ProblemList, userID primitive.ObjectID) error {
	// 收集所有题目ID
	problemIDs := make([]primitive.ObjectID, len(problems))
	for i, problem := range problems {
		problemIDs[i] = problem.ID
	}

	// 批量查询用户状态
	statusMap, err := p.GetUserProblemStatuses(ctx, userID, problemIDs)
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

// validateProblemCreation 验证题目创建参
func (p *ProblemRepository) validateProblemCreation(problem *models.Problem) error {
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
func (p *ProblemRepository) validateProblemUpdate(req *models.UpdateProblemRequest) error {
	// 业务规则：草稿状态强制私有
	if req.Status != nil && *req.Status == models.StatusDraft {
		// 通过修改请求对象来应用业务规则
		falseValue := false
		req.IsPublic = &falseValue
		utils.Logger.Warnf("validateProblemUpdate: 草稿状态不能公开，强制设为私有")
	}

	return nil
}

// GetNextSequenceValue 原子性地获取并递增指定序列的值
func (r *ProblemRepository) GetNextSequenceValue(ctx context.Context, sequenceName string) (int64, error) {
	countersCollection := r.db.Collection("counters")

	filter := bson.M{"name": sequenceName}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result struct {
		SequenceValue int64 `bson:"sequence_value"`
	}

	err := countersCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 如果计数器不存在，可以尝试初始化或直接报错
			utils.Logger.Errorf("GetNextSequenceValue: 计数器 '%s' 不存在", sequenceName)
			return 0, fmt.Errorf("计数器不存在，请先初始化")
		}
		utils.Logger.Errorf("GetNextSequenceValue: 数据库操作失败, sequence_name=%s, error=%v", sequenceName, err)
		return 0, err
	}

	return result.SequenceValue, nil
}

// GetByUniqueID 根据唯一编号查询题目
func (r *ProblemRepository) GetByUniqueID(ctx context.Context, uniqueID int64) (*models.Problem, error) {
	filter := bson.M{"unique_id": uniqueID}

	var problem models.Problem
	// 使用BaseRepository提供的FindOne方法，与GetByID保持一致
	err := r.FindOne(ctx, "problems", filter, &problem)
	if err != nil {
		return nil, err
	}

	return &problem, nil
}
