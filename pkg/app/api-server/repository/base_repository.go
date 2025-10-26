/*
@Author: omenkk7
@Date: 2025/10/25
@Description: Repository基础层 - 通用数据访问功能
*/

package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zk-code-arena-server/pkg/utils"
)

// BaseRepository 基础Repository，提供通用的数据访问方法
type BaseRepository struct {
	db *mongo.Database
}

// NewBaseRepository 创建BaseRepository实例
func NewBaseRepository() *BaseRepository {
	return &BaseRepository{
		db: utils.MongoDB,
	}
}

// BuildUpdateSet 构建MongoDB更新文档 (基于BSON序列化)
// 这是轻量级的通用更新函数，避免了反射的性能开销
func (r *BaseRepository) BuildUpdateSet(obj interface{}) (bson.M, error) {
	// 使用BSON序列化自动处理omitempty标签
	data, err := bson.Marshal(obj)
	if err != nil {
		utils.Logger.Errorf("BuildUpdateSet: BSON序列化失败, error=%v", err)
		return nil, fmt.Errorf("BSON序列化失败: %w", err)
	}
	
	// 反序列化为bson.M，自动过滤omitempty字段
	var update bson.M
	if err := bson.Unmarshal(data, &update); err != nil {
		utils.Logger.Errorf("BuildUpdateSet: BSON反序列化失败, error=%v", err)
		return nil, fmt.Errorf("BSON反序列化失败: %w", err)
	}
	
	// 如果没有字段需要更新，返回nil
	if len(update) == 0 {
		utils.Logger.Debugf("BuildUpdateSet: 没有字段需要更新")
		return nil, nil
	}
	
	// 构造$set更新文档
	return bson.M{"$set": update}, nil
}

// BuildUpdateSetWithTime 构建带时间戳的更新文档
func (r *BaseRepository) BuildUpdateSetWithTime(obj interface{}) (bson.M, error) {
	updateDoc, err := r.BuildUpdateSet(obj)
	if err != nil {
		return nil, err
	}
	
	// 如果没有字段需要更新，只更新时间戳
	if updateDoc == nil {
		return bson.M{"$set": bson.M{"updatedAt": time.Now()}}, nil
	}
	
	// 添加updatedAt字段
	updateDoc["$set"].(bson.M)["updatedAt"] = time.Now()
	return updateDoc, nil
}

// UpdateOne 通用更新方法
func (r *BaseRepository) UpdateOne(ctx context.Context, collection string, filter bson.M, updateData interface{}) (*mongo.UpdateResult, error) {
	// 构建更新文档
	updateDoc, err := r.BuildUpdateSetWithTime(updateData)
	if err != nil {
		return nil, err
	}
	
	if updateDoc == nil {
		return nil, fmt.Errorf("没有字段需要更新")
	}
	
	// 执行更新操作
	coll := r.db.Collection(collection)
	result, err := coll.UpdateOne(ctx, filter, updateDoc)
	
	if err != nil {
		utils.Logger.Errorf("UpdateOne: 数据库更新失败, collection=%s, error=%v", collection, err)
		return nil, fmt.Errorf("数据库更新失败: %w", err)
	}
	
	utils.Logger.Infof("UpdateOne: 更新成功, collection=%s, matched=%d, modified=%d", 
		collection, result.MatchedCount, result.ModifiedCount)
	
	return result, nil
}

// FindOne 通用查询单个文档方法
func (r *BaseRepository) FindOne(ctx context.Context, collection string, filter bson.M, result interface{}) error {
	coll := r.db.Collection(collection)
	err := coll.FindOne(ctx, filter).Decode(result)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.Logger.Debugf("FindOne: 文档不存在, collection=%s", collection)
			return fmt.Errorf("文档不存在")
		}
		utils.Logger.Errorf("FindOne: 查询失败, collection=%s, error=%v", collection, err)
		return fmt.Errorf("查询失败: %w", err)
	}
	
	return nil
}

// Find 通用查询多个文档方法
func (r *BaseRepository) Find(ctx context.Context, collection string, filter bson.M, opts *options.FindOptions) (*mongo.Cursor, error) {
	coll := r.db.Collection(collection)
	cursor, err := coll.Find(ctx, filter, opts)
	
	if err != nil {
		utils.Logger.Errorf("Find: 查询失败, collection=%s, error=%v", collection, err)
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	
	return cursor, nil
}

// InsertOne 通用插入单个文档方法
func (r *BaseRepository) InsertOne(ctx context.Context, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	coll := r.db.Collection(collection)
	result, err := coll.InsertOne(ctx, document)
	
	if err != nil {
		utils.Logger.Errorf("InsertOne: 插入失败, collection=%s, error=%v", collection, err)
		return nil, fmt.Errorf("插入失败: %w", err)
	}
	
	utils.Logger.Infof("InsertOne: 插入成功, collection=%s, id=%v", collection, result.InsertedID)
	return result, nil
}

// InsertMany 通用批量插入方法
func (r *BaseRepository) InsertMany(ctx context.Context, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	coll := r.db.Collection(collection)
	result, err := coll.InsertMany(ctx, documents)
	
	if err != nil {
		utils.Logger.Errorf("InsertMany: 批量插入失败, collection=%s, error=%v", collection, err)
		return nil, fmt.Errorf("批量插入失败: %w", err)
	}
	
	utils.Logger.Infof("InsertMany: 批量插入成功, collection=%s, count=%d", collection, len(result.InsertedIDs))
	return result, nil
}

// DeleteOne 通用删除单个文档方法
func (r *BaseRepository) DeleteOne(ctx context.Context, collection string, filter bson.M) (*mongo.DeleteResult, error) {
	coll := r.db.Collection(collection)
	result, err := coll.DeleteOne(ctx, filter)
	
	if err != nil {
		utils.Logger.Errorf("DeleteOne: 删除失败, collection=%s, error=%v", collection, err)
		return nil, fmt.Errorf("删除失败: %w", err)
	}
	
	utils.Logger.Infof("DeleteOne: 删除成功, collection=%s, deleted=%d", collection, result.DeletedCount)
	return result, nil
}

// CountDocuments 通用文档计数方法
func (r *BaseRepository) CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error) {
	coll := r.db.Collection(collection)
	count, err := coll.CountDocuments(ctx, filter)
	
	if err != nil {
		utils.Logger.Errorf("CountDocuments: 计数失败, collection=%s, error=%v", collection, err)
		return 0, fmt.Errorf("计数失败: %w", err)
	}
	
	return count, nil
}

// Aggregate 通用聚合查询方法
func (r *BaseRepository) Aggregate(ctx context.Context, collection string, pipeline []bson.M) (*mongo.Cursor, error) {
	coll := r.db.Collection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline)
	
	if err != nil {
		utils.Logger.Errorf("Aggregate: 聚合查询失败, collection=%s, error=%v", collection, err)
		return nil, fmt.Errorf("聚合查询失败: %w", err)
	}
	
	return cursor, nil
}
