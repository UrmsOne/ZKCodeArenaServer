# 设计文档

## 概述

本设计文档描述如何将 `pkg/common/mongo/query.go` 中的通用查询功能集成到 `BaseRepository` 中，提供轻量、通用的数据访问方法。

设计原则：
1. **轻量简洁**: 只添加必要的 CRUD 方法和 query.go 的核心功能
2. **直接集成**: 直接调用 `pkg/common/mongo` 包的函数，不创建额外的包装层
3. **保持简单**: 不引入复杂的 Builder 模式或工具类

## 架构

### 整体架构

```
pkg/common/mongo/query.go (通用查询工具)
         ↓ (直接调用)
BaseRepository (基础仓储层 - 轻量封装)
         ↓ (继承)
具体 Repository (ProblemRepository, SubmitRepository 等)
```

## 组件和接口

### 1. BaseRepository 新增方法

#### 1.1 批量插入

```go
// InsertMany 批量插入文档
func (r *BaseRepository) InsertMany(
    ctx context.Context,
    collection string,
    documents []interface{},
) (*mongo.InsertManyResult, error)
```

**说明**: 已存在，保持不变

#### 1.2 批量删除

```go
// DeleteMany 批量删除文档
func (r *BaseRepository) DeleteMany(
    ctx context.Context,
    collection string,
    filter bson.M,
) (*mongo.DeleteResult, error)
```

**说明**: 新增方法，支持批量删除

#### 1.3 批量更新

```go
// UpdateMany 批量更新文档
func (r *BaseRepository) UpdateMany(
    ctx context.Context,
    collection string,
    filter bson.M,
    updateData interface{},
) (*mongo.UpdateResult, error)
```

**说明**: 新增方法，支持批量更新，自动添加 updated_at

#### 1.4 集成 query.go 的分页查询

```go
// FindWithPaginationTyped 泛型分页查询（直接调用 common.FindWithPagination）
func (r *BaseRepository) FindWithPaginationTyped[T any](
    ctx context.Context,
    collection string,
    filter interface{},
    opts *options.FindOptions,
) ([]T, int64, error)
```

**说明**: 
- 直接调用 `common.FindWithPagination[T]`
- 使用泛型支持类型安全
- 返回数据列表和总数

#### 1.5 集成 query.go 的辅助函数

直接在 BaseRepository 中暴露 `pkg/common/mongo` 的函数：

```go
// BuildSort 构建排序（直接调用 common.BuildSort）
func (r *BaseRepository) BuildSort(sortExpr string) bson.D {
    return common.BuildSort(sortExpr)
}

// BuildKeywordFilter 构建关键字过滤器（直接调用 common.BuildKeywordFilter）
func (r *BaseRepository) BuildKeywordFilter(keyword string, fields []string) bson.M {
    return common.BuildKeywordFilter(keyword, fields)
}

// MergeFilters 合并过滤器（直接调用 common.MergeFilters）
func (r *BaseRepository) MergeFilters(filters ...bson.M) bson.M {
    return common.MergeFilters(filters...)
}

// BuildFindOptionsWithPagination 构建查询选项（直接调用 common.BuildFindOptions）
func (r *BaseRepository) BuildFindOptionsWithPagination(
    page, pageSize int64,
    sort bson.D,
    projection bson.M,
) *options.FindOptions {
    return common.BuildFindOptions(page, pageSize, sort, projection)
}
```

**说明**: 这些方法只是简单的转发调用，不添加额外逻辑

## 实现细节

### 1. 导入 common 包

```go
import (
    common "zk-code-arena-server/pkg/common/mongo"
)
```

### 2. 批量删除实现

```go
func (r *BaseRepository) DeleteMany(ctx context.Context, collection string, filter bson.M) (*mongo.DeleteResult, error) {
    coll := r.db.Collection(collection)
    result, err := coll.DeleteMany(ctx, filter)
    
    if err != nil {
        utils.Logger.Errorf("DeleteMany: 批量删除失败, collection=%s, error=%v", collection, err)
        return nil, fmt.Errorf("批量删除失败: %w", err)
    }
    
    utils.Logger.Infof("DeleteMany: 批量删除成功, collection=%s, deleted=%d", collection, result.DeletedCount)
    return result, nil
}
```

### 3. 批量更新实现

```go
func (r *BaseRepository) UpdateMany(ctx context.Context, collection string, filter bson.M, updateData interface{}) (*mongo.UpdateResult, error) {
    updateDoc, err := r.BuildUpdateSetWithTime(updateData)
    if err != nil {
        return nil, err
    }
    
    if updateDoc == nil {
        return nil, fmt.Errorf("没有字段需要更新")
    }
    
    coll := r.db.Collection(collection)
    result, err := coll.UpdateMany(ctx, filter, updateDoc)
    
    if err != nil {
        utils.Logger.Errorf("UpdateMany: 批量更新失败, collection=%s, error=%v", collection, err)
        return nil, fmt.Errorf("批量更新失败: %w", err)
    }
    
    utils.Logger.Infof("UpdateMany: 批量更新成功, collection=%s, matched=%d, modified=%d",
        collection, result.MatchedCount, result.ModifiedCount)
    
    return result, nil
}
```

### 4. 泛型分页查询实现

```go
func (r *BaseRepository) FindWithPaginationTyped[T any](
    ctx context.Context,
    collection string,
    filter interface{},
    opts *options.FindOptions,
) ([]T, int64, error) {
    coll := r.GetCollection(collection)
    return common.FindWithPagination[T](ctx, coll, filter, opts)
}
```

### 5. 辅助函数实现（直接转发）

```go
func (r *BaseRepository) BuildSort(sortExpr string) bson.D {
    return common.BuildSort(sortExpr)
}

func (r *BaseRepository) BuildKeywordFilter(keyword string, fields []string) bson.M {
    return common.BuildKeywordFilter(keyword, fields)
}

func (r *BaseRepository) MergeFilters(filters ...bson.M) bson.M {
    return common.MergeFilters(filters...)
}

func (r *BaseRepository) BuildFindOptionsWithPagination(page, pageSize int64, sort bson.D, projection bson.M) *options.FindOptions {
    return common.BuildFindOptions(page, pageSize, sort, projection)
}
```

## 错误处理

保持与现有 BaseRepository 一致的错误处理策略：

1. **日志记录**: 所有错误都记录到日志
2. **错误包装**: 使用 `fmt.Errorf` 包装底层错误
3. **统一格式**: 错误信息包含操作类型、集合名称和原始错误

## 测试策略

由于这些方法主要是转发调用 `pkg/common/mongo` 包的函数，测试重点在于：

1. **集成测试**: 验证 BaseRepository 方法能正确调用 common 包的函数
2. **批量操作测试**: 测试 DeleteMany 和 UpdateMany 的正确性

## 使用示例

### 示例 1: 批量删除

```go
// 删除所有草稿状态的题目
filter := bson.M{"status": "draft"}
result, err := r.DeleteMany(ctx, "problems", filter)
```

### 示例 2: 批量更新

```go
// 批量更新题目状态
type StatusUpdate struct {
    Status string `bson:"status"`
}
filter := bson.M{"is_public": false}
update := &StatusUpdate{Status: "published"}
result, err := r.UpdateMany(ctx, "problems", filter, update)
```

### 示例 3: 分页查询

```go
// 查询公开题目列表
filter := bson.M{"is_public": true}
sort := r.BuildSort("-created_at")
opts := r.BuildFindOptionsWithPagination(1, 20, sort, nil)

problems, total, err := r.FindWithPaginationTyped[models.Problem](ctx, "problems", filter, opts)
```

### 示例 4: 关键字搜索

```go
// 搜索题目
keywordFilter := r.BuildKeywordFilter("算法", []string{"title", "description"})
statusFilter := bson.M{"is_public": true}
filter := r.MergeFilters(keywordFilter, statusFilter)

sort := r.BuildSort("-created_at")
opts := r.BuildFindOptionsWithPagination(1, 20, sort, nil)

problems, total, err := r.FindWithPaginationTyped[models.Problem](ctx, "problems", filter, opts)
```

## 总结

本设计保持轻量简洁，通过直接集成 `pkg/common/mongo/query.go` 的功能到 `BaseRepository`，提供了：

1. **完整的 CRUD**: InsertOne/Many, UpdateOne/Many, DeleteOne/Many, Find/FindOne
2. **分页查询**: 泛型支持的类型安全分页查询
3. **查询辅助**: 排序、关键字搜索、过滤器合并等实用工具

所有方法都保持简单直接，不引入额外的抽象层。
