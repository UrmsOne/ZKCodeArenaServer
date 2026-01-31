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
			ID:            problem.ID,
			UniqueID:      problem.UniqueID,
			Title:         problem.Title,
			Difficulty:    problem.Difficulty,
			ACCount:       problem.ACCount,
			SubmitCount:   problem.SubmitCount,
			CreatedAt:     problem.CreatedAt,
			FavoriteCount: problem.FavoriteCount, // 新增
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
			ID:            problem.ID,
			UniqueID:      problem.UniqueID,
			Title:         problem.Title,
			Difficulty:    problem.Difficulty,
			Tags:          problem.Tags,
			ACCount:       problem.ACCount,
			SubmitCount:   problem.SubmitCount,
			Status:        problem.Status,
			IsPublic:      problem.IsPublic,
			CreatedAt:     problem.CreatedAt,
			FavoriteCount: problem.FavoriteCount, // 新增
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
	inFavorite bool,
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

	// 如果需要在收藏中搜索，且用户已登录
	if inFavorite && userID != nil {
		// 获取用户收藏的题目ID
		favoriteFilter := bson.M{"user_id": userID}
		cursor, err := p.Find(ctx, "user_favorite_problems", favoriteFilter, nil)
		if err != nil {
			return nil, 0, fmt.Errorf("获取收藏题目失败: %w", err)
		}
		defer cursor.Close(ctx)

		var favorites []models.UserFavoriteProblem
		if err := cursor.All(ctx, &favorites); err != nil {
			return nil, 0, fmt.Errorf("解析收藏题目失败: %w", err)
		}

		if len(favorites) > 0 {
			// 提取收藏的题目ID
			problemIDs := make([]primitive.ObjectID, len(favorites))
			for i, fav := range favorites {
				problemIDs[i] = fav.ProblemID
			}

			// 添加收藏题目ID过滤
			filters = append(filters, bson.M{"_id": bson.M{"$in": problemIDs}})
		} else {
			// 用户没有收藏任何题目，返回空结果
			return []*models.ProblemList{}, 0, nil
		}
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
			ID:            problem.ID,
			UniqueID:      problem.UniqueID,
			Title:         problem.Title,
			Difficulty:    problem.Difficulty,
			Tags:          problem.Tags,
			ACCount:       problem.ACCount,
			SubmitCount:   problem.SubmitCount,
			Status:        problem.Status,
			IsPublic:      problem.IsPublic,
			CreatedAt:     problem.CreatedAt,
			FavoriteCount: problem.FavoriteCount, // 新增
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

	// 新增：填充收藏状态
	if len(problems) > 0 {
		problemIDs := make([]primitive.ObjectID, len(problems))
		for i, problem := range problems {
			problemIDs[i] = problem.ID
		}

		favoriteMap, err := p.GetUserFavoriteProblemIDs(ctx, userID, problemIDs)
		if err != nil {
			return err
		}

		for _, problem := range problems {
			if favoriteMap[problem.ID] {
				trueValue := true
				problem.IsFavorite = &trueValue
			} else {
				falseValue := false
				problem.IsFavorite = &falseValue
			}
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

// AddFavorite 添加收藏
func (p *ProblemRepository) AddFavorite(ctx context.Context, userID, problemID primitive.ObjectID) error {
	// 1. 检查是否已经收藏
	count, err := p.CountDocuments(ctx, "user_favorite_problems", bson.M{
		"user_id":    userID,
		"problem_id": problemID,
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // 已经收藏，直接返回
	}

	// 2. 添加收藏记录
	favorite := &models.UserFavoriteProblem{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		ProblemID: problemID,
		CreatedAt: time.Now(),
	}
	if _, err := p.InsertOne(ctx, "user_favorite_problems", favorite); err != nil {
		return err
	}

	// 3. 增加题目收藏数
	coll := p.db.Collection("problems")
	_, err = coll.UpdateOne(ctx,
		bson.M{"_id": problemID},
		bson.M{"$inc": bson.M{"favorite_count": 1}, "$set": bson.M{"updated_at": time.Now()}},
	)
	if err != nil {
		utils.Logger.Errorf("AddFavorite: 增加收藏数失败, problemID=%v, error=%v", problemID, err)
		return err
	}
	return nil
}

// RemoveFavorite 取消收藏
func (p *ProblemRepository) RemoveFavorite(ctx context.Context, userID, problemID primitive.ObjectID) error {
	// 1. 删除收藏记录并检查是否成功删除
	result, err := p.DeleteOne(ctx, "user_favorite_problems", bson.M{
		"user_id":    userID,
		"problem_id": problemID,
	})
	if err != nil {
		return err
	}

	// 2. 只有当确实删除了一条收藏记录时，才减少题目收藏数
	if result.DeletedCount > 0 {
		// 3. 减少题目收藏数 - 使用原生 MongoDB 驱动方法
		coll := p.db.Collection("problems")
		_, err = coll.UpdateOne(ctx,
			bson.M{"_id": problemID},
			bson.M{
				"$inc": bson.M{"favorite_count": -1},
				"$set": bson.M{"updated_at": time.Now()},
			},
		)
		if err != nil {
			utils.Logger.Errorf("RemoveFavorite: 更新收藏数失败, problemID=%v, error=%v", problemID, err)
			return err
		}
	}

	return nil

}

// GetUserFavoriteProblems 获取用户收藏的题目列表
func (p *ProblemRepository) GetUserFavoriteProblems(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*models.ProblemList, int64, error) {
	// 1. 查找用户收藏的题目ID
	favoriteFilter := bson.M{"user_id": userID}
	favoriteSort := bson.D{{"created_at", -1}}
	favoriteOpts := p.BuildFindOptionsWithPagination(int64(page), int64(pageSize), favoriteSort, nil)

	// 使用 Find 方法获取游标
	cursor, err := p.Find(ctx, "user_favorite_problems", favoriteFilter, favoriteOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解码游标到结构体切片
	var favorites []models.UserFavoriteProblem
	if err := cursor.All(ctx, &favorites); err != nil {
		return nil, 0, err
	}

	// 2. 获取总收藏数
	total, err := p.CountDocuments(ctx, "user_favorite_problems", favoriteFilter)
	if err != nil {
		return nil, 0, err
	}

	if len(favorites) == 0 {
		return []*models.ProblemList{}, total, nil
	}

	// 3. 获取收藏的题目详情
	problemIDs := make([]primitive.ObjectID, len(favorites))
	for i, fav := range favorites {
		problemIDs[i] = fav.ProblemID
	}

	// 4. 查询题目详情
	problemFilter := bson.M{"_id": bson.M{"$in": problemIDs}}
	// 不需要分页，获取所有匹配的题目
	problemOpts := options.Find()
	// 按ID排序（与problemIDs顺序一致）
	problemOpts.SetSort(bson.D{{"_id", 1}})

	// 使用 FindWithPaginationTyped 获取题目列表
	problems, _, err := FindWithPaginationTyped[models.Problem](ctx, p.BaseRepository, "problems", problemFilter, problemOpts)
	if err != nil {
		return nil, 0, err
	}

	// 5. 转换为 ProblemList 并设置收藏状态
	problemList := make([]*models.ProblemList, len(problems))
	trueValue := true
	for i, problem := range problems {
		problemList[i] = &models.ProblemList{
			ID:            problem.ID,
			UniqueID:      problem.UniqueID,
			Title:         problem.Title,
			Difficulty:    problem.Difficulty,
			Tags:          problem.Tags,
			ACCount:       problem.ACCount,
			SubmitCount:   problem.SubmitCount,
			Status:        problem.Status,
			IsPublic:      problem.IsPublic,
			CreatedAt:     problem.CreatedAt,
			FavoriteCount: problem.FavoriteCount, // 新增
			IsFavorite:    &trueValue,            // 已收藏
		}
	}

	return problemList, total, nil
}

// GetUserFavoriteProblemIDs 获取用户收藏的题目ID列表
func (p *ProblemRepository) GetUserFavoriteProblemIDs(ctx context.Context, userID primitive.ObjectID, problemIDs []primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
	filter := bson.M{
		"user_id":    userID,
		"problem_id": bson.M{"$in": problemIDs},
	}

	// 使用 Find 方法获取游标
	cursor, err := p.Find(ctx, "user_favorite_problems", filter, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解码游标到结构体切片
	var favorites []models.UserFavoriteProblem
	if err := cursor.All(ctx, &favorites); err != nil {
		return nil, err
	}

	result := make(map[primitive.ObjectID]bool)
	for _, fav := range favorites {
		result[fav.ProblemID] = true
	}

	return result, nil
}

// GetNextProblem 获取当前题目的下一题
func (r *ProblemRepository) GetNextProblem(ctx context.Context, currentUniqueID int64) (*models.Problem, error) {
	// 筛选条件：unique_id大于当前ID，且题目为公开且已发布
	filter := bson.M{
		"unique_id": bson.M{"$gt": currentUniqueID},
		"is_public": true,
		"status":    models.StatusPublished,
	}

	var problem models.Problem
	// 直接使用MongoDB的Collection.FindOne方法，并设置排序选项
	coll := r.db.Collection("problems")
	findOneOptions := options.FindOne().SetSort(bson.M{"unique_id": 1})
	err := coll.FindOne(ctx, filter, findOneOptions).Decode(&problem)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("未找到下一题")
		}
		return nil, err
	}

	return &problem, nil
}
