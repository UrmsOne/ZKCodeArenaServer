package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetTagList 分页查询标签列表
func (s *Service) GetTagList(ctx context.Context, req *models.TagQueryRequest) ([]*models.Tag, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
		utils.Logger.Debugf("GetTagList: 页码无效，设置默认值 1")
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
		utils.Logger.Debugf("GetTagList: 每页条数无效，设置默认值 10")
	}

	tagCol := utils.GetCollection("tags")
	filter := bson.M{}
	if req.Keyword != "" {
		filter["name"] = bson.M{"$regex": req.Keyword, "$options": "i"}
	}

	total, err := tagCol.CountDocuments(ctx, filter)
	if err != nil {
		utils.Logger.Errorf("GetTagList: 统计标签总数失败, error=%v", err)
		return nil, 0, fmt.Errorf("统计标签总数失败: %w", err)
	}
	utils.Logger.Debugf("GetTagList: 标签总条数统计完成, total=%d", total)

	opts := options.Find().
		SetSkip(int64((req.Page - 1) * req.PageSize)).
		SetLimit(int64(req.PageSize)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := tagCol.Find(ctx, filter, opts)
	if err != nil {
		utils.Logger.Errorf("GetTagList: 查询标签列表失败, error=%v", err)
		return nil, 0, fmt.Errorf("查询标签列表失败: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			utils.Logger.Errorf("GetTagList: 关闭游标失败, error=%v", err)
		}
	}()

	var tags []*models.Tag
	for cursor.Next(ctx) {
		var t models.Tag
		if err := cursor.Decode(&t); err != nil {
			utils.Logger.Errorf("GetTagList: 解析标签数据失败, error=%v", err)
			return nil, 0, fmt.Errorf("解析标签失败: %w", err)
		}
		tags = append(tags, &t)
	}

	utils.Logger.Infof("GetTagList: 标签列表查询完成, 返回条数=%d", len(tags))
	return tags, total, cursor.Err()
}

// GetTagByID 获取标签详情
func (s *Service) GetTagByID(ctx context.Context, tagID primitive.ObjectID) (*models.Tag, error) {
	tagCol := utils.GetCollection("tags")

	var tag models.Tag
	err := tagCol.FindOne(ctx, bson.M{"_id": tagID}).Decode(&tag)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.Logger.Warnf("GetTagByID: 标签不存在, tagID=%s", tagID.Hex())
			return nil, fmt.Errorf("标签ID「%s」不存在", tagID.Hex())
		}
		utils.Logger.Errorf("GetTagByID: 查询标签失败, tagID=%s, error=%v", tagID.Hex(), err)
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	utils.Logger.Infof("GetTagByID: 标签查询成功, tagID=%s, tagName=%s", tagID.Hex(), tag.Name)
	return &tag, nil
}

// CreateTag 创建标签
func (s *Service) CreateTag(ctx context.Context, tag *models.Tag) error {
	tag.ID = primitive.NewObjectID()
	tag.CreatedAt = time.Now()
	tag.UpdatedAt = time.Now()

	tagCol := utils.GetCollection("tags")

	// 校验标签名唯一
	var existing models.Tag
	err := tagCol.FindOne(ctx, bson.M{"name": tag.Name}).Decode(&existing)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			utils.Logger.Errorf("CreateTag: 校验标签名唯一性失败, name=%s, error=%v", tag.Name, err)
			return fmt.Errorf("校验标签名失败: %w", err)
		}
	} else {
		utils.Logger.Warnf("CreateTag: 标签名称已存在, name=%s", tag.Name)
		return fmt.Errorf("标签名称「%s」已存在", tag.Name)
	}

	_, err = tagCol.InsertOne(ctx, tag)
	if err != nil {
		utils.Logger.Errorf("CreateTag: 插入标签数据失败, name=%s, id=%s, error=%v", tag.Name, tag.ID.Hex(), err)
		return fmt.Errorf("插入标签失败: %w", err)
	}
	utils.Logger.Infof("CreateTag: 标签创建成功, name=%s, id=%s", tag.Name, tag.ID.Hex())
	return nil
}

// UpdateTag 更新标签
func (s *Service) UpdateTag(ctx context.Context, tagID primitive.ObjectID, req *models.UpdateTagRequest) error {
	tagCol := utils.GetCollection("tags")

	// 校验标签存在
	var tag models.Tag
	err := tagCol.FindOne(ctx, bson.M{"_id": tagID}).Decode(&tag)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.Logger.Warnf("UpdateTag: 标签不存在, tagID=%s", tagID.Hex())
			return fmt.Errorf("标签ID「%s」不存在", tagID.Hex())
		}
		utils.Logger.Errorf("UpdateTag: 查询标签失败, tagID=%s, error=%v", tagID.Hex(), err)
		return fmt.Errorf("查询标签失败: %w", err)
	}

	// 校验新名称唯一
	if req.Name != "" {
		var existing models.Tag
		err = tagCol.FindOne(ctx, bson.M{
			"_id":  bson.M{"$ne": tagID},
			"name": req.Name,
		}).Decode(&existing)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				utils.Logger.Errorf("UpdateTag: 校验新标签名唯一性失败, newName=%s, tagID=%s, error=%v", req.Name, tagID.Hex(), err)
				return fmt.Errorf("校验新标签名失败: %w", err)
			}
		} else {
			utils.Logger.Warnf("UpdateTag: 新标签名已存在, newName=%s", req.Name)
			return fmt.Errorf("标签名称「%s」已存在", req.Name)
		}
	}

	update := bson.M{
		"$set": bson.M{
			"name":       req.Name,
			"desc":       req.Desc,
			"updated_at": time.Now(),
		},
	}
	utils.Logger.Debugf("UpdateTag: 构建更新条件完成, tagID=%s", tagID.Hex())

	result, err := tagCol.UpdateByID(ctx, tagID, update)
	if err != nil {
		utils.Logger.Errorf("UpdateTag: 更新标签失败, tagID=%s, error=%v", tagID.Hex(), err)
		return fmt.Errorf("更新标签失败: %w", err)
	}
	if result.ModifiedCount == 0 {
		utils.Logger.Warnf("UpdateTag: 标签无更新内容, tagID=%s", tagID.Hex())
		return fmt.Errorf("标签无更新内容")
	}
	utils.Logger.Infof("UpdateTag: 标签更新成功, tagID=%s", tagID.Hex())
	return nil
}

// DeleteTag 删除标签
func (s *Service) DeleteTag(ctx context.Context, tagID primitive.ObjectID) error {
	tagCol := utils.GetCollection("tags")
	problemCol := utils.GetCollection("problems")

	// 先获取标签详情
	var tag models.Tag
	err := tagCol.FindOne(ctx, bson.M{"_id": tagID}).Decode(&tag)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.Logger.Warnf("DeleteTag: 标签不存在, tagID=%s", tagID.Hex())
			return fmt.Errorf("标签ID「%s」不存在", tagID.Hex())
		}
		utils.Logger.Errorf("DeleteTag: 查询标签失败, tagID=%s, error=%v", tagID.Hex(), err)
		return fmt.Errorf("查询标签失败: %w", err)
	}

	// 清理题目中的该标签
	_, err = problemCol.UpdateMany(ctx,
		bson.M{"tags": tag.Name},
		bson.M{"$pull": bson.M{"tags": tag.Name}},
	)
	if err != nil {
		utils.Logger.Errorf("DeleteTag: 清理题目关联标签失败, tagName=%s, error=%v", tag.Name, err)
		return fmt.Errorf("清理题目关联标签失败: %w", err)
	}

	// 删除标签
	result, err := tagCol.DeleteOne(ctx, bson.M{"_id": tagID})
	if err != nil {
		utils.Logger.Errorf("DeleteTag: 删除标签失败, tagID=%s, error=%v", tagID.Hex(), err)
		return fmt.Errorf("删除标签失败: %w", err)
	}
	if result.DeletedCount == 0 {
		utils.Logger.Warnf("DeleteTag: 标签删除失败, tagID=%s", tagID.Hex())
		return fmt.Errorf("标签删除失败")
	}
	utils.Logger.Infof("DeleteTag: 标签删除成功, tagID=%s, tagName=%s", tagID.Hex(), tag.Name)
	return nil
}

// GetPublicTagList 获取所有标签
func (s *Service) GetPublicTagList(ctx context.Context) ([]*models.Tag, error) {
	tagCol := utils.GetCollection("tags")
	cursor, err := tagCol.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		utils.Logger.Errorf("GetPublicTagList: 查询标签列表失败, error=%v", err)
		return nil, fmt.Errorf("查询标签列表失败: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			utils.Logger.Errorf("GetPublicTagList: 关闭游标失败, error=%v", err)
		}
	}()

	var tags []*models.Tag
	for cursor.Next(ctx) {
		var t models.Tag
		if err := cursor.Decode(&t); err != nil {
			utils.Logger.Errorf("GetPublicTagList: 解析标签失败, error=%v", err)
			return nil, fmt.Errorf("解析标签失败: %w", err)
		}
		tags = append(tags, &t)
	}
	utils.Logger.Infof("GetPublicTagList: 标签查询完成, 返回条数=%d", len(tags))
	return tags, cursor.Err()
}
