# 数据库更新操作审计报告

**审计日期**: 2025年10月27日  
**审计范围**: 所有Repository层和Service层的数据库更新操作  
**审计目的**: 检查是否存在双重 `$set` 嵌套问题

---

## 问题背景

在更新题目接口中发现了一个 MongoDB 更新错误：

```json
{
  "code": 500,
  "message": "更新题目失败: 数据库操作失败: 数据库更新失败: write exception: write errors: [The dollar ($) prefixed field '$set' in '$set' is not allowed in the context of an update's replacement document. Consider using an aggregation pipeline with $replaceWith.]"
}
```

**原因**: 在调用 MongoDB 的 `UpdateOne` 方法时，手动包装了 `bson.M{"$set": ...}`，但后续又通过 `BaseRepository.UpdateOne` 方法再次添加 `$set`，导致双重嵌套。

---

## 审计结果

### ✅ 已修复的问题

#### 1. `problem_repository.go` - UpdateProblemFromRequest (第409行)

**问题代码**:
```go
// ❌ 错误：手动添加 $set 后，又传给会添加 $set 的方法
result, err := p.UpdateOne(ctx, "problems", bson.M{"_id": problemID}, bson.M{"$set": updateFields})
```

**修复后**:
```go
// ✅ 正确：直接使用 MongoDB 的 UpdateOne，手动控制 $set
filter := bson.M{"_id": problemID}
update := bson.M{"$set": updateFields}

coll := p.db.Collection("problems")
result, err := coll.UpdateOne(ctx, filter, update)
```

---

### ✅ 确认无问题的更新操作

#### 1. BaseRepository.UpdateOne 方法

**位置**: `base_repository.go:76-101`

**实现方式**:
```go
func (r *BaseRepository) UpdateOne(ctx context.Context, collection string, filter bson.M, updateData interface{}) (*mongo.UpdateResult, error) {
    // 自动调用 BuildUpdateSetWithTime，添加 $set
    updateDoc, err := r.BuildUpdateSetWithTime(updateData)
    // ...
    result, err := coll.UpdateOne(ctx, filter, updateDoc)
}
```

**使用原则**: 
- ✅ 传入普通对象或结构体
- ❌ 不要传入已包装 `$set` 的 bson.M

---

#### 2. ProblemRepository 更新操作

| 方法 | 位置 | 状态 | 说明 |
|------|------|------|------|
| `UpdateProblem` | 第66行 | ⚠️ 未使用 | 已被 UpdateProblemFromRequest 取代 |
| `UpdateProblemFromRequest` | 第342行 | ✅ 已修复 | 直接使用 MongoDB UpdateOne |
| `UpdateProblemStats` | 第254行 | ✅ 正确 | 直接使用 MongoDB UpdateOne，使用 $inc |

---

#### 3. SubmitRepository 更新操作

| 方法 | 位置 | 调用方式 | 状态 |
|------|------|----------|------|
| `UpdateSubmit` | 第68行 | `r.UpdateOne(ctx, "submits", filter, submit)` | ✅ 正确 |
| `UpdateSubmitStatus` | 第85行 | `r.UpdateOne(ctx, "submits", filter, updateData)` | ✅ 正确 |
| `UpdateSubmitResult` | 第106行 | `r.UpdateOne(ctx, "submits", filter, updateData)` | ✅ 正确 |
| `BatchUpdateStatus` | 第212行 | `coll.UpdateMany(ctx, filter, bson.M{"$set": updateData})` | ✅ 正确 |

**说明**: 前3个方法传入结构体，由 BaseRepository.UpdateOne 自动添加 `$set`；最后一个直接使用 MongoDB API，手动添加 `$set`。

---

#### 4. TestCaseRepository 更新操作

| 方法 | 位置 | 调用方式 | 状态 |
|------|------|----------|------|
| `UpdateTestCase` | 第122行 | `r.UpdateOne(ctx, "test_cases", filter, req)` | ✅ 正确 |
| `UpdateTestCaseComplete` | 第139行 | `r.UpdateOne(ctx, "test_cases", filter, testCase)` | ✅ 正确 |
| `UpdateTestCaseOrder` | 第240行 | `r.UpdateOne(ctx, "test_cases", filter, updateData)` | ✅ 正确 |

**说明**: 所有方法都传入结构体，由 BaseRepository.UpdateOne 自动添加 `$set`。

---

#### 5. Service 层直接更新操作

##### CourseService (service_course.go)

| 方法 | 位置 | 更新操作 | 状态 |
|------|------|----------|------|
| `UpdateTask` | 第472行 | `bson.M{"$set": updateFields}` | ✅ 正确 |
| `AddCourseTeacher` | 第667行 | `bson.M{"$addToSet": ..., "$set": ...}` | ✅ 正确 |
| `AddCourseTeachers` | 第716行 | `bson.M{"$addToSet": ..., "$set": ...}` | ✅ 正确 |
| `RemoveCourseTeachers` | 第765行 | `bson.M{"$pullAll": ..., "$set": ...}` | ✅ 正确 |
| `AddClazzTeacher` | 第831行 | `bson.M{"$addToSet": ..., "$set": ...}` | ✅ 正确 |
| `RemoveClazzTeacher` | 第976行 | `bson.M{"$pull": ..., "$set": ...}` | ✅ 正确 |

##### ClazzService (service_clazz.go)

| 方法 | 位置 | 更新操作 | 状态 |
|------|------|----------|------|
| `UpdateClazz` | 第300行 | `bson.M{"$set": updateFields}` | ✅ 正确 |
| `JoinClazz` | 第428行 | `bson.M{"$inc": ..., "$set": ...}` | ✅ 正确 |
| `LeaveClazz` | 第547行 | `bson.M{"$pull": ..., "$set": ...}` | ✅ 正确 |
| `AddClazzTeacher` | 第617行 | `bson.M{"$addToSet": ..., "$set": ...}` | ✅ 正确 |
| `RemoveClazzTeacher` | 第685行 | `bson.M{"$pull": ..., "$set": ...}` | ✅ 正确 |

**说明**: Service 层的这些方法都直接使用 MongoDB API (`coll.UpdateOne`)，手动构造更新文档，这是正确的做法。

---

## 最佳实践建议

### 1. 使用 BaseRepository.UpdateOne

**✅ 推荐场景**: 简单的字段更新

```go
// 定义更新数据结构
updateData := struct {
    Status string `bson:"status"`
    Name   string `bson:"name"`
}{
    Status: "active",
    Name:   "新名称",
}

// 调用 BaseRepository.UpdateOne（会自动添加 $set 和 updated_at）
result, err := r.UpdateOne(ctx, "collection", filter, updateData)
```

**注意事项**:
- ✅ 传入普通结构体或对象
- ❌ 不要传入 `bson.M{"$set": ...}`
- ✅ 自动添加 `updated_at` 字段
- ✅ 自动过滤 `omitempty` 字段

---

### 2. 直接使用 MongoDB UpdateOne

**✅ 推荐场景**: 需要使用特殊操作符（$inc, $push, $pull, $addToSet 等）

```go
// 手动构造完整的更新文档
update := bson.M{
    "$inc": bson.M{"count": 1},
    "$set": bson.M{"updated_at": time.Now()},
}

// 直接调用 MongoDB API
coll := r.db.Collection("collection")
result, err := coll.UpdateOne(ctx, filter, update)
```

**注意事项**:
- ✅ 完全控制更新操作
- ✅ 可以组合多个操作符
- ⚠️ 需要手动添加 `updated_at`
- ⚠️ 需要手动处理错误日志

---

### 3. 构建动态更新字段

**✅ 推荐场景**: 根据请求参数动态更新部分字段

```go
// 动态构建更新字段
updateFields := bson.M{}

if req.Title != nil {
    updateFields["title"] = *req.Title
}
if req.Description != nil {
    updateFields["description"] = *req.Description
}

// 添加更新时间
updateFields["updated_at"] = time.Now()

// 直接使用 MongoDB UpdateOne
filter := bson.M{"_id": id}
update := bson.M{"$set": updateFields}
result, err := coll.UpdateOne(ctx, filter, update)
```

---

## 代码清理建议

### 1. 删除未使用的方法

`problem_repository.go` 中的 `UpdateProblem` 方法（第66-86行）已被 `UpdateProblemFromRequest` 取代，建议删除以避免混淆：

```go
// ⚠️ 未使用的旧方法，建议删除
func (r *ProblemRepository) UpdateProblem(ctx context.Context, problemID primitive.ObjectID, req *models.UpdateProblemRequest) error {
    // ...
}
```

---

## 总结

### 发现的问题
- ❌ 1个双重 `$set` 嵌套问题 (已修复)
- ⚠️ 1个未使用的旧方法 (建议清理)

### 审计覆盖范围
- ✅ BaseRepository: 1个通用更新方法
- ✅ ProblemRepository: 3个更新方法
- ✅ SubmitRepository: 4个更新方法  
- ✅ TestCaseRepository: 3个更新方法
- ✅ CourseService: 6个直接更新操作
- ✅ ClazzService: 5个直接更新操作

### 结论
✅ **除已修复的 UpdateProblemFromRequest 方法外，所有其他更新操作均正确实现，无双重 $set 嵌套问题。**

---

## 附录：常见错误模式

### ❌ 错误模式 1: 双重 $set 嵌套

```go
// 错误：手动包装 $set，然后传给会再次添加 $set 的方法
update := bson.M{"$set": updateFields}
result, err := r.UpdateOne(ctx, collection, filter, update)
// UpdateOne 会再包装一次 $set，导致 {"$set": {"$set": {...}}}
```

### ✅ 正确方式 1: 传入普通对象

```go
// 正确：直接传入对象，让 UpdateOne 自动添加 $set
result, err := r.UpdateOne(ctx, collection, filter, updateData)
```

### ✅ 正确方式 2: 直接使用 MongoDB API

```go
// 正确：绕过 BaseRepository.UpdateOne，自己控制 $set
update := bson.M{"$set": updateFields}
coll := r.db.Collection(collection)
result, err := coll.UpdateOne(ctx, filter, update)
```

---

**审计人员**: AI Assistant  
**审核状态**: ✅ 已完成  
**最后更新**: 2025年10月27日

