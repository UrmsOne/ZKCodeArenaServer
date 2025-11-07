/*
@Author:
@Date: 2025/10/25
@Name: service_clazz.go
@Description: 班级服务层实现
*/

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"

	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ClazzService 班级服务
type ClazzService struct{}

// NewClazzService 创建班级服务实例
func NewClazzService() *ClazzService {
	return &ClazzService{}
}

// 刷新二维码
func (s2 *ClazzService) RefreshQrcode(userId string, courseId string, clazzId string, ctx context.Context) (interface{}, error) {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errors.New("用户错误")
	}
	courseObjId, err := primitive.ObjectIDFromHex(courseId)
	if err != nil {
		return nil, errors.New("课程错误")
	}

	if isSuccess, err := authorized(ctx, utils.GetCollection("courses"), courseObjId, userObjId); !isSuccess || err != nil {
		if !isSuccess {
			return nil, errors.New("权限不足")
		}
		return nil, err
	}

	ran := fmt.Sprintf("%d,%s", rand.Int(), clazzId)
	encode, err := qrcode.Encode(ran, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}
	// 将二维码字节切片编码为 Base64 字符串
	base64QRCode := base64.StdEncoding.EncodeToString(encode)

	err = utils.RedisClient.HSet(ctx, "clazz_qrcode"+clazzId, "qrcode", base64QRCode, "ran", ran).Err()
	utils.RedisClient.Expire(ctx, "clazz_qrcode"+clazzId, 30*time.Minute)

	if err != nil {
		return nil, err
	}
	return base64QRCode, nil
}

// CreateClass 创建班级
func (s *ClazzService) CreateClass(ctx context.Context, req *models.CreateClazzRequest, userId string) (*models.ClazzResponse, error) {
	coll := utils.GetCollection("courses")
	courseID, err := primitive.ObjectIDFromHex(req.CourseId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": courseID}
	// 使用 projection 查询 created_by 和 teacher_ids 字段
	projection := bson.M{"created_by": 1, "teacher_ids": 1}

	var result struct {
		CreatorID  primitive.ObjectID   `bson:"created_by"`
		TeacherIds []primitive.ObjectID `bson:"teacher_ids,omitempty"`
	}
	// 执行查询
	err = coll.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("课程不存在")
		}
		return nil, err
	}

	// 验证权限 - 任何课程教师都可以创建班级
	isTeacher := false
	for _, teacherId := range result.TeacherIds {
		if teacherId.Hex() == userId {
			isTeacher = true
			break
		}
	}
	if !isTeacher {
		return nil, errors.New("权限不足，只有课程教师可以创建班级")
	}

	// 处理教师ID列表 - 现在是必须字段
	teacherIds := make([]primitive.ObjectID, 0, len(req.TeacherIds))

	// 验证请求中的教师ID是否有效
	for _, teacherId := range req.TeacherIds {
		teacherObjId, err := primitive.ObjectIDFromHex(teacherId)
		if err != nil {
			return nil, errors.New("无效的教师ID: " + teacherId)
		}
		teacherIds = append(teacherIds, teacherObjId)
	}

	// 验证这些教师是否属于该课程
	for _, reqTeacherId := range teacherIds {
		found := false
		for _, courseTeacherId := range result.TeacherIds {
			if reqTeacherId == courseTeacherId {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("教师 " + reqTeacherId.Hex() + " 不属于该课程")
		}
	}

	var maxMembers int
	if req.MaxMembers == nil {
		maxMembers = 60
	} else if *req.MaxMembers > 100 {
		maxMembers = 100
	} else {
		maxMembers = *req.MaxMembers
	}

	// 创建班级对象，使用请求中指定的教师信息
	now := time.Now()
	clazz := &models.Clazz{
		ID:            primitive.NewObjectID(),
		Name:          req.Name,
		CourseId:      courseID,
		Description:   req.Description,
		Schedule:      req.Schedule,
		TeacherIds:    teacherIds,
		RequireInvite: req.RequireInvite,
		MaxMembers:    maxMembers,
		AddNums:       0,
		Status:        models.ClassStatusActive, // 默认为活跃状态，支持加入
		CTime:         now,
		MTime:         now,
	}

	one, err := utils.GetCollection("clazzes").InsertOne(ctx, clazz)
	if err != nil {
		return nil, errors.New("创建班级失败")
	}

	if req.RequireInvite {
		id := one.InsertedID.(primitive.ObjectID)
		ran := fmt.Sprintf("%d,%s", rand.Int(), id.Hex())
		encode, err := qrcode.Encode(ran, qrcode.Medium, 256)
		if err != nil {
			return nil, err
		}
		base64QRCode := base64.StdEncoding.EncodeToString(encode)
		err = utils.RedisClient.HSet(ctx, "clazz_qrcode"+id.Hex(), "qrcode", base64QRCode, "ran", ran).Err()
		utils.RedisClient.Expire(ctx, "clazz_qrcode"+id.Hex(), 30*time.Minute)
	}

	return &models.ClazzResponse{
		ClazzId: clazz.ID.Hex(),
	}, nil
}

// GetClazzByID 获取班级详情
func (s *ClazzService) GetClazzByID(ctx context.Context, clazzID string, userID string) (*models.GetClazzResponse, error) {
	clazzObjID, err := primitive.ObjectIDFromHex(clazzID)
	if err != nil {
		return nil, errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	// 尝试从Redis缓存中获取班级详情
	cacheKey := "clazz_detail:" + clazzID + ":" + userID
	cachedData, err := utils.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		// 缓存命中，直接返回缓存数据
		var response models.GetClazzResponse
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			return &response, nil
		}
		// 如果反序列化失败，继续执行下面的逻辑
	}

	// 先查询班级基本信息
	clazzColl := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = clazzColl.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("班级不存在")
		}
		return nil, err
	}

	// 构造返回的响应对象
	response := &models.GetClazzResponse{
		ID:            clazz.ID,
		Name:          clazz.Name,
		Description:   clazz.Description,
		CourseId:      clazz.CourseId,
		Schedule:      clazz.Schedule,
		RequireInvite: clazz.RequireInvite,
		MaxMembers:    clazz.MaxMembers,
		AddNums:       clazz.AddNums,
		Status:        clazz.Status,
		CTime:         clazz.CTime,
	}

	courseObjID := clazz.CourseId
	courseColl := utils.GetCollection("courses")

	// 检查用户是否是班级成员（通过查询学生班级关联表）
	count, err := utils.GetCollection("student_classes").CountDocuments(ctx, bson.M{
		"student_id": userObjID,
		"class_id":   clazzObjID,
	})
	if err != nil {
		return nil, errors.New("检查班级成员失败: " + err.Error())
	}
	isMember := count > 0

	// 检查用户是否是课程创建者或教师
	isAuthorized, err := s.authorized(ctx, courseColl, courseObjID, userObjID)
	if err != nil {
		return nil, err
	}

	if !isAuthorized && !isMember {
		return nil, errors.New("权限不足")
	}

	// 获取班级教师信息
	var teachers []models.User
	if len(clazz.TeacherIds) > 0 {
		userColl := utils.GetCollection("users")
		teacherFilter := bson.M{
			"_id": bson.M{"$in": clazz.TeacherIds},
		}
		teacherCursor, err := userColl.Find(ctx, teacherFilter)
		if err != nil {
			return nil, errors.New("查询教师信息失败: " + err.Error())
		}
		defer teacherCursor.Close(ctx)

		if err = teacherCursor.All(ctx, &teachers); err != nil {
			return nil, errors.New("解析教师信息失败: " + err.Error())
		}
	}

	// 转换为用户资料数组
	response.Teachers = make([]models.UserProfile, 0, len(teachers))
	for _, teacher := range teachers {
		response.Teachers = append(response.Teachers, *teacher.ToProfile())
	}

	// 将结果缓存到Redis中，缓存10分钟
	cacheData, err := json.Marshal(response)
	if err == nil {
		utils.RedisClient.Set(ctx, cacheKey, cacheData, 10*time.Minute)
	}

	return response, nil
}

// UpdateClazzInfo 更新班级信息（不包括教师和成员）
func (s *ClazzService) UpdateClazzInfo(ctx context.Context, clazzID string, userId string, req *models.UpdateClazzRequest) error {
	clazzObjID, err := primitive.ObjectIDFromHex(clazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	coll := utils.GetCollection("clazzes")

	// 先查询班级信息以验证权限
	var clazz models.Clazz
	if err = coll.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 验证操作者是否有权限修改班级
	// 只有课程创建者或课程教师可以修改班级
	courseObjID := clazz.CourseId
	courseColl := utils.GetCollection("courses")

	isAuthorized, err := s.authorized(ctx, courseColl, courseObjID, userObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或教师可以修改班级")
	}

	// 构建更新字段
	updateFields := bson.M{}
	// 检查每个字段，如果字段非空，则加入更新内容
	if req.Name != "" {
		updateFields["name"] = req.Name
	}
	if req.Description != "" {
		updateFields["description"] = req.Description
	}
	if req.Schedule != "" {
		updateFields["schedule"] = req.Schedule
	}
	if req.RequireInvite != nil {
		updateFields["require_invite"] = *req.RequireInvite
	}
	if req.MaxMembers != nil {
		// 确保最大成员数不超过100
		maxMembers := *req.MaxMembers
		if maxMembers > 100 {
			maxMembers = 100
		}
		updateFields["max_members"] = maxMembers
	}

	// 如果没有任何字段需要更新，返回早期退出
	if len(updateFields) == 0 {
		return errors.New("无更改内容")
	}

	updateFields["mtime"] = time.Now()

	filter := bson.M{"_id": clazzObjID}
	update := bson.M{"$set": updateFields}
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("找不到该班级")
	}

	return nil
}

// DeleteClazz 删除班级
func (s *ClazzService) DeleteClazz(ctx context.Context, clazzID string, userID string) error {
	clazzObjID, err := primitive.ObjectIDFromHex(clazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	coll := utils.GetCollection("clazzes")

	// 先查询班级信息
	var clazz models.Clazz

	filter := bson.M{
		"_id": clazzObjID,
	}

	if err = coll.FindOne(ctx, filter).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 验证操作者是否有权限删除班级
	// 只有课程创建者或课程教师可以删除班级
	courseObjID := clazz.CourseId
	courseColl := utils.GetCollection("courses")

	isAuthorized, err := s.authorized(ctx, courseColl, courseObjID, userObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或教师可以删除班级")
	}

	// 检查班级是否还有成员
	if clazz.AddNums > 0 {
		return errors.New("班级还有成员，无法删除")
	}

	// 删除班级
	if _, err = coll.DeleteOne(ctx, filter); err != nil {
		return errors.New("删除班级失败: " + err.Error())
	}

	return nil
}

// AddClazzMember 添加班级成员
func (s *ClazzService) AddClazzMember(ctx context.Context, clazzID string, memberID string, operatorID string) error {
	clazzObjID, err := primitive.ObjectIDFromHex(clazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	memberObjID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		return errors.New("无效的成员ID")
	}

	operatorObjID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("无效的操作者ID")
	}

	coll := utils.GetCollection("clazzes")

	// 先查询班级信息以验证权限
	var clazz models.Clazz
	if err = coll.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 验证操作者是否有权限添加成员
	// 只有课程创建者或课程教师可以添加成员
	courseObjID := clazz.CourseId
	courseColl := utils.GetCollection("courses")

	isAuthorized, err := s.authorized(ctx, courseColl, courseObjID, operatorObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或教师可以添加成员")
	}

	// 检查班级是否已满
	if clazz.AddNums >= clazz.MaxMembers {
		return errors.New("班级已满，无法添加更多成员")
	}

	// 检查是否已是成员（通过查询学生班级关联表）
	count, err := utils.GetCollection("student_classes").CountDocuments(ctx, bson.M{
		"student_id": memberObjID,
		"class_id":   clazzObjID,
	})
	if err != nil {
		return errors.New("检查班级成员失败: " + err.Error())
	}

	if count > 0 {
		return errors.New("该用户已经是班级成员")
	}

	// 添加成员
	filter := bson.M{
		"_id":   clazzObjID,
		"$expr": bson.M{"$lt": []interface{}{"$add_nums", "$max_members"}}, //乐观锁
	}

	update := bson.M{
		"$inc": bson.M{
			"add_nums": 1,
		},
		"$set": bson.M{
			"mtime": time.Now(),
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("添加成员失败: " + err.Error())
	}

	if result.MatchedCount == 0 {
		return errors.New("班级已满，无法添加更多成员")
	}

	// 同时创建学生班级关联记录
	studentClass := &models.StudentClass{
		ID:        primitive.NewObjectID(),
		StudentID: memberObjID,
		ClassID:   clazzObjID,
		CourseID:  clazz.CourseId,
		JoinTime:  time.Now(),
		Status:    "active",
		CTime:     time.Now(),
		MTime:     time.Now(),
	}

	_, err = utils.GetCollection("student_classes").InsertOne(ctx, studentClass)
	if err != nil {
		// 如果创建关联记录失败，应该回滚之前的班级成员添加操作
		// 这里简化处理，实际应该使用事务
		_, _ = coll.UpdateOne(ctx, bson.M{"_id": clazzObjID}, bson.M{"$inc": bson.M{"add_nums": -1}})
	}

	return nil
}

// RemoveClazzMembers 批量移除班级成员
func (s *ClazzService) RemoveClazzMembers(ctx context.Context, req models.RemoveClazzMembersRequest, operatorID string) error {
	clazzObjID, err := primitive.ObjectIDFromHex(req.ClazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	operatorObjID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("无效的操作者ID")
	}

	var null *bool
	// 去重成员ID
	uniqueMemberIDs := make(map[string]*bool)
	for _, memberID := range req.MemberIDs {
		uniqueMemberIDs[memberID] = null
	}

	// 转换成员ID
	memberObjIDs := make([]primitive.ObjectID, 0, len(uniqueMemberIDs))
	for memberID := range uniqueMemberIDs {
		objID, err := primitive.ObjectIDFromHex(memberID)
		if err != nil {
			return errors.New("无效的成员ID: " + memberID)
		}
		memberObjIDs = append(memberObjIDs, objID)
	}

	coll := utils.GetCollection("clazzes")

	// 先查询班级信息以验证权限
	var clazz models.Clazz
	if err = coll.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 验证操作者是否有权限移除成员
	// 只有课程创建者或课程教师可以移除成员
	courseObjID := clazz.CourseId
	courseColl := utils.GetCollection("courses")

	isAuthorized, err := s.authorized(ctx, courseColl, courseObjID, operatorObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或教师可以移除成员")
	}

	// 检查成员是否是班级成员（通过查询学生班级关联表）
	validMembers := make([]primitive.ObjectID, 0)
	for _, memberObjID := range memberObjIDs {
		count, err := utils.GetCollection("student_classes").CountDocuments(ctx, bson.M{
			"student_id": memberObjID,
			"class_id":   clazzObjID,
		})
		if err != nil {
			return errors.New("检查班级成员失败: " + err.Error())
		}

		if count > 0 {
			validMembers = append(validMembers, memberObjID)
		}
	}

	if len(validMembers) == 0 {
		return errors.New("没有有效的班级成员需要移除")
	}

	// 移除成员
	filter := bson.M{
		"_id": clazzObjID,
	}

	update := bson.M{
		"$inc": bson.M{
			"add_nums": -len(validMembers),
		},
		"$set": bson.M{
			"mtime": time.Now(),
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("移除成员失败: " + err.Error())
	}

	if result.MatchedCount == 0 {
		return errors.New("班级不存在")
	}

	// 同时删除学生班级关联记录
	_, err = utils.GetCollection("student_classes").DeleteMany(ctx, bson.M{
		"class_id":   clazzObjID,
		"student_id": bson.M{"$in": validMembers},
	})
	if err != nil {
		// 如果删除关联记录失败，记录日志但不返回错误
	}

	return nil
}

// AddClazzTeacher 为班级添加教师
func (s *ClazzService) AddClazzTeacher(ctx context.Context, clazzId string, teacherId string, userId string) error {
	// 验证用户权限（只有课程创建者或课程教师才能添加班级教师）
	clazzObjID, err := primitive.ObjectIDFromHex(clazzId)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	teacherObjID, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return errors.New("无效的教师ID")
	}

	// 检查班级是否存在并验证权限
	coll := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = coll.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 验证操作者是否有权限添加教师
	courseColl := utils.GetCollection("courses")
	isAuthorized, err := authorized(ctx, courseColl, clazz.CourseId, userObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或课程教师可以添加班级教师")
	}

	// 检查教师是否已经是班级的教师
	for _, id := range clazz.TeacherIds {
		if id == teacherObjID {
			return errors.New("该教师已经是班级的教师")
		}
	}

	// 添加教师到班级
	filter := bson.M{"_id": clazzObjID}
	update := bson.M{"$addToSet": bson.M{"teacher_ids": teacherObjID}, "$set": bson.M{"mtime": time.Now()}}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("添加班级教师失败: " + err.Error())
	}

	return nil
}

// RemoveClazzTeacher 为班级移除教师
func (s *ClazzService) RemoveClazzTeacher(ctx context.Context, clazzId string, teacherId string, userId string) error {
	// 验证用户权限（只有课程创建者或课程教师才能移除班级教师）
	clazzObjID, err := primitive.ObjectIDFromHex(clazzId)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	teacherObjID, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return errors.New("无效的教师ID")
	}

	// 检查班级是否存在并验证权限
	coll := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = coll.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 验证操作者是否有权限移除教师
	courseColl := utils.GetCollection("courses")
	isAuthorized, err := authorized(ctx, courseColl, clazz.CourseId, userObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或教师可以移除班级教师")
	}

	// 检查教师是否是班级的教师
	found := false
	for _, id := range clazz.TeacherIds {
		if id == teacherObjID {
			found = true
			break
		}
	}

	if !found {
		return errors.New("该教师不是班级的教师")
	}

	// 不能移除最后一个教师
	if len(clazz.TeacherIds) <= 1 {
		return errors.New("不能移除最后一个教师")
	}

	// 从班级移除教师
	filter := bson.M{"_id": clazzObjID}
	update := bson.M{"$pull": bson.M{"teacher_ids": teacherObjID}, "$set": bson.M{"mtime": time.Now()}}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("移除班级教师失败: " + err.Error())
	}

	return nil
}

// GetClazzesByCourseId 通过课程ID查询所有班级
func (s *ClazzService) GetClazzesByCourseId(ctx context.Context, courseId string, userId string) ([]models.Clazz, error) {
	courseObjId, err := primitive.ObjectIDFromHex(courseId)
	if err != nil {
		return nil, errors.New("无效的课程ID")
	}

	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	// 验证用户是否有权限查看该课程的班级
	// 用户必须是课程创建者或课程教师
	coll := utils.GetCollection("courses")
	isAuthorized, err := s.authorized(ctx, coll, courseObjId, userObjId)
	if err != nil {
		return nil, err
	}

	if !isAuthorized {
		return nil, errors.New("权限不足")
	}

	// 查询该课程的所有班级
	clazzColl := utils.GetCollection("clazzes")
	filter := bson.M{"course_id": courseObjId}
	cursor, err := clazzColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Clazz
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// AddStudentToClass 将学生添加到班级
func (s *ClazzService) AddStudentToClass(ctx context.Context, studentID, classID, courseID string) error {
	studentObjID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		return errors.New("无效的学生ID")
	}

	classObjID, err := primitive.ObjectIDFromHex(classID)
	if err != nil {
		return errors.New("无效的班级ID")
	}
	courseObjID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		return errors.New("无效的班级ID")
	}
	// 检查学生是否已经在这个班级中
	coll := utils.GetCollection("student_classes")
	count, err := coll.CountDocuments(ctx, bson.M{
		"student_id": studentObjID,
		"class_id":   classObjID,
	})
	if err != nil {
		return errors.New("检查学生班级关系失败: " + err.Error())
	}

	if count > 0 {
		return errors.New("学生已经在这个班级中")
	}

	// 创建学生班级关联记录
	now := time.Now()
	studentClass := &models.StudentClass{
		ID:        primitive.NewObjectID(),
		StudentID: studentObjID,
		ClassID:   classObjID,
		CourseID:  courseObjID,
		JoinTime:  now,
		Status:    "active",
		CTime:     now,
		MTime:     now,
	}

	_, err = coll.InsertOne(ctx, studentClass)
	if err != nil {
		return errors.New("添加学生到班级失败: " + err.Error())
	}

	// 同时更新班级的成员列表
	clazzColl := utils.GetCollection("clazzes")
	filter := bson.M{
		"_id":   classObjID,
		"$expr": bson.M{"$lt": []interface{}{"$add_nums", "$max_members"}}, //乐观锁
	}

	update := bson.M{
		"$addToSet": bson.M{
			"member_ids": studentObjID,
		},
		"$inc": bson.M{
			"add_nums": 1,
		},
		"$set": bson.M{
			"mtime": now,
		},
	}

	result, err := clazzColl.UpdateOne(ctx, filter, update)
	if err != nil {
		// 如果更新班级失败，需要回滚之前的学生班级关联记录
		_, _ = coll.DeleteOne(ctx, bson.M{"_id": studentClass.ID})
		return errors.New("更新班级成员失败: " + err.Error())
	}

	if result.MatchedCount == 0 {
		// 如果没有匹配到班级记录，需要回滚之前的学生班级关联记录
		_, _ = coll.DeleteOne(ctx, bson.M{"_id": studentClass.ID})
		return errors.New("班级不存在或已满")
	}

	return nil
}

// RemoveStudentFromClass 将学生从班级中移除
func (s *ClazzService) RemoveStudentFromClass(ctx context.Context, studentID, classID, courseID string) error {
	studentObjID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		return errors.New("无效的学生ID")
	}

	classObjID, err := primitive.ObjectIDFromHex(classID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	courseObjID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		return errors.New("无效的课程ID")
	}

	// 删除学生班级关联记录
	coll := utils.GetCollection("student_classes")
	filter := bson.M{
		"student_id": studentObjID,
		"class_id":   classObjID,
		"course_id":  courseObjID,
	}
	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("从班级移除学生失败: " + err.Error())
	}

	if result.DeletedCount == 0 {
		return errors.New("学生不在该班级中")
	}

	// 同时更新班级的成员列表
	clazzColl := utils.GetCollection("clazzes")
	update := bson.M{
		"$inc": bson.M{
			"add_nums": -1,
		},
		"$set": bson.M{
			"mtime": time.Now(),
		},
	}

	_, err = clazzColl.UpdateOne(ctx, bson.M{"_id": classObjID}, update)
	if err != nil {
		// 注意：这里如果更新失败，学生班级关联记录已经被删除，数据会不一致
		// 在生产环境中应该使用事务来保证一致性
		return errors.New("更新班级成员失败: " + err.Error())
	}

	return nil
}

// GetStudentClasses 获取学生的所有班级
func (s *ClazzService) GetStudentClasses(ctx context.Context, studentID string) ([]*models.StudentClassResponse, error) {
	studentObjID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		return nil, errors.New("无效的学生ID")
	}

	// 查询学生的所有班级关联记录
	coll := utils.GetCollection("student_classes")
	cursor, err := coll.Find(ctx, bson.M{"student_id": studentObjID, "status": "active"})
	if err != nil {
		return nil, errors.New("查询学生班级失败: " + err.Error())
	}
	defer cursor.Close(ctx)

	var studentClasses []models.StudentClass
	if err = cursor.All(ctx, &studentClasses); err != nil {
		return nil, errors.New("解析学生班级数据失败: " + err.Error())
	}

	// 构建响应数据
	var responses []*models.StudentClassResponse
	for _, sc := range studentClasses {
		response := &models.StudentClassResponse{
			ID:        sc.ID,
			StudentID: sc.StudentID,
			ClassID:   sc.ClassID,
			CourseID:  sc.CourseID,
			JoinTime:  sc.JoinTime,
			Status:    sc.Status,
			CTime:     sc.CTime,
			MTime:     sc.MTime,
		}

		// 获取学生信息
		user, err := s.getUserByID(ctx, sc.StudentID)
		if err == nil {
			response.Student = user.ToProfile()
		}

		// 获取班级信息
		clazzColl := utils.GetCollection("clazzes")
		var clazz models.Clazz
		if err = clazzColl.FindOne(ctx, bson.M{"_id": sc.ClassID}).Decode(&clazz); err == nil {
			response.Class = &clazz
		}

		// 获取课程信息
		course, err := s.getCourseByID(ctx, sc.CourseID)
		if err == nil {
			response.Course = course
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetClassStudents 获取班级的所有学生
func (s *ClazzService) GetClassStudents(ctx context.Context, classID string) ([]*models.UserProfile, error) {
	classObjID, err := primitive.ObjectIDFromHex(classID)
	if err != nil {
		return nil, errors.New("无效的班级ID")
	}

	// 查询班级的所有学生关联记录
	coll := utils.GetCollection("student_classes")
	cursor, err := coll.Find(ctx, bson.M{"class_id": classObjID, "status": "active"})
	if err != nil {
		return nil, errors.New("查询班级学生失败: " + err.Error())
	}
	defer cursor.Close(ctx)

	var studentClasses []models.StudentClass
	if err = cursor.All(ctx, &studentClasses); err != nil {
		return nil, errors.New("解析班级学生数据失败: " + err.Error())
	}

	// 获取所有学生信息
	var students []*models.UserProfile
	for _, sc := range studentClasses {
		user, err := s.getUserByID(ctx, sc.StudentID)
		if err != nil {
			continue
		}
		students = append(students, user.ToProfile())
	}

	return students, nil
}

// GetTasksByClazzID 根据班级ID获取任务列表
func (s *ClazzService) GetTasksByClazzID(ctx context.Context, clazzID string, userID string) ([]models.Task, error) {
	clazzObjID, err := primitive.ObjectIDFromHex(clazzID)
	if err != nil {
		return nil, errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	// 先验证用户是否有权限查看该班级的任务
	// 获取班级信息
	collClazz := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = collClazz.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("班级不存在")
		}
		return nil, err
	}

	// 验证用户权限：必须是班级成员、课程创建者或课程教师
	// 检查用户是否是班级成员（通过查询学生班级关联表）
	count, err := utils.GetCollection("student_classes").CountDocuments(ctx, bson.M{
		"student_id": userObjID,
		"class_id":   clazzObjID,
	})
	if err != nil {
		return nil, errors.New("检查班级成员失败: " + err.Error())
	}
	isMember := count > 0

	collCourse := utils.GetCollection("courses")
	isAuthorized, err := s.authorized(ctx, collCourse, clazz.CourseId, userObjID)
	if err != nil {
		return nil, err
	}

	if !isMember && !isAuthorized {
		return nil, errors.New("权限不足")
	}

	// 从独立的tasks集合中查询任务
	collTasks := utils.GetCollection("tasks")
	filter := bson.M{"clazz_id": clazzObjID}
	cursor, err := collTasks.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTaskByID 根据任务ID获取任务详情
func (s *ClazzService) GetTaskByID(ctx context.Context, taskID string, userID string) (*models.TaskResponse, error) {
	taskObjID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, errors.New("无效的任务ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	// 获取任务信息
	collTasks := utils.GetCollection("tasks")
	var task models.Task
	if err = collTasks.FindOne(ctx, bson.M{"_id": taskObjID}).Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("任务不存在")
		}
		return nil, err
	}

	// 验证用户权限：必须是班级成员、课程创建者或课程教师
	// 获取班级信息
	collClazz := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = collClazz.FindOne(ctx, bson.M{"_id": task.ClazzId}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("班级不存在")
		}
		return nil, err
	}

	// 检查用户是否是班级成员（通过查询学生班级关联表）
	count, err := utils.GetCollection("student_classes").CountDocuments(ctx, bson.M{
		"student_id": userObjID,
		"class_id":   task.ClazzId,
	})
	if err != nil {
		return nil, errors.New("检查班级成员失败: " + err.Error())
	}
	isMember := count > 0

	collCourse := utils.GetCollection("courses")
	isAuthorized, err := s.authorized(ctx, collCourse, clazz.CourseId, userObjID)
	if err != nil {
		return nil, err
	}

	if !isMember && !isAuthorized {
		return nil, errors.New("权限不足")
	}

	// 查询用户的任务完成状态
	state := 0

	// 通过查询user_task表来判断任务是否已完成
	userTaskColl := utils.GetCollection("user_task")
	var userTask models.UserTask
	err = userTaskColl.FindOne(ctx, bson.M{
		"task_id": taskObjID,
		"user_id": userObjID,
	}).Decode(&userTask)

	if err == nil && userTask.State == 1 {
		state = 1
	}

	// 查询用户在每道题上的完成情况
	completedQuestions := make(map[primitive.ObjectID]bool)
	userRelationColl := utils.GetCollection("user_relation_question")
	cursor, err := userRelationColl.Find(ctx, bson.M{
		"task_id": taskObjID,
		"user_id": userObjID,
		"state":   1, // 已完成的题目
	})
	if err != nil {
		return nil, errors.New("查询用户题目完成情况失败: " + err.Error())
	}
	defer cursor.Close(ctx)

	var userRelations []models.UserRelationQuestion
	if err = cursor.All(ctx, &userRelations); err != nil {
		return nil, errors.New("解析用户题目完成情况失败: " + err.Error())
	}

	for _, ur := range userRelations {
		completedQuestions[ur.RelationID] = true
	}

	// 查询题目详情
	questions := make([]models.QuestionDetail, 0, len(task.RelationIDs))
	if len(task.RelationIDs) > 0 {
		problemColl := utils.GetCollection("problems")
		problemCursor, err := problemColl.Find(ctx, bson.M{
			"_id": bson.M{"$in": task.RelationIDs},
		})
		if err != nil {
			return nil, errors.New("查询题目详情失败: " + err.Error())
		}
		defer problemCursor.Close(ctx)

		var problems []models.Problem
		if err = problemCursor.All(ctx, &problems); err != nil {
			return nil, errors.New("解析题目详情失败: " + err.Error())
		}

		// 创建题目ID到题目的映射
		problemMap := make(map[primitive.ObjectID]models.Problem)
		for _, p := range problems {
			problemMap[p.ID] = p
		}

		// 按照RelationIDs的顺序构建题目详情列表
		for _, relationID := range task.RelationIDs {
			if problem, exists := problemMap[relationID]; exists {
				question := models.QuestionDetail{
					ID:         problem.ID,
					Title:      problem.Title,
					Difficulty: string(problem.Difficulty),
					Completed:  completedQuestions[relationID],
				}
				questions = append(questions, question)
			}
		}
	}

	// 构建 TaskResponse
	taskResponse := &models.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Type:        task.Type,
		StartTime:   task.StartTime,
		EndTime:     task.EndTime,
		Status:      task.Status,
		CourseId:    task.CourseId,
		ClazzId:     task.ClazzId,
		CTime:       task.CTime,
		CID:         task.CID,
		MTime:       task.MTime,
		State:       state,
		Questions:   questions,
	}

	return taskResponse, nil
}

// 私有方法
// getCourseByID 根据ID获取课程
func (s *ClazzService) getCourseByID(ctx context.Context, courseID primitive.ObjectID) (*models.Course, error) {
	var course models.Course
	err := utils.GetCollection("courses").FindOne(ctx, bson.M{"_id": courseID}).Decode(&course)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("课程不存在")
		}
		return nil, err
	}
	return &course, nil
}

// getUserByID 根据ID获取用户
func (s *ClazzService) getUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := utils.GetCollection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

func (s *ClazzService) contains(slice []primitive.ObjectID, v string) bool {
	for _, item := range slice {
		if item.Hex() == v {
			return true
		}
	}
	return false
}

func (s *ClazzService) authorized(ctx context.Context, coll *mongo.Collection, courseId primitive.ObjectID, userId primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": courseId}
	var course models.Course
	projection := bson.M{
		"created_by":  1,
		"teacher_ids": 1,
	}
	if err := coll.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&course); err != nil {
		return false, errors.New("课程不存在")
	}
	isTeacher := s.contains(course.TeacherIds, userId.Hex())
	isCreator := course.CreatedBy == userId
	return isTeacher || isCreator, nil
}

func (s *ClazzService) cheekIsMember(ids []primitive.ObjectID, userID primitive.ObjectID) bool {
	for _, memberID := range ids {
		if memberID == userID {
			return true
		}
	}
	return false
}

// JoinClazz 通过二维码扫描加入班级
func (s *ClazzService) JoinClazz(ctx context.Context, req models.JoinClazzRequest, userID string) error {
	// 验证用户ID和班级ID格式
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	log.Printf("解析班级Obj: %s", req.ClazzID)
	clazzObjID, err := primitive.ObjectIDFromHex(req.ClazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	// 获取班级信息
	clazzColl := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = clazzColl.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 检查是否可以加入
	if !clazz.CanJoin() {
		return errors.New("班级已满或已结束")
	}

	// 检查用户是否已经是班级成员
	count, err := utils.GetCollection("student_classes").CountDocuments(ctx, bson.M{
		"student_id": userObjID,
		"class_id":   clazzObjID,
	})
	if err != nil {
		return errors.New("检查班级成员失败: " + err.Error())
	}

	if count > 0 {
		return errors.New("您已经是该班级成员")
	}

	// 如果班级不需要邀请，则直接加入
	if !clazz.RequireInvite {
		return s.joinClazzDirectly(ctx, &clazz, userObjID)
	}

	// 如果班级需要邀请，则验证二维码
	key := "clazz_qrcode" + req.ClazzID
	log.Printf("尝试从Redis获取二维码key: %s", key)
	storedRan, err := utils.RedisClient.HGet(ctx, key, "ran").Result()
	if err != nil {
		log.Printf("从Redis获取二维码失败: %v", err)
		return errors.New("二维码已过期或不存在")
	}

	// 验证ran值是否匹配
	if req.InviteCode == nil {
		return errors.New("二维码无效: 请求中未提供邀请码")
	}

	log.Printf("Stored ran: %s, Request ran: %s", storedRan, *req.InviteCode)

	if storedRan != *req.InviteCode {
		return errors.New("二维码无效: 邀请码不匹配")
	}

	// 执行加入班级的逻辑
	return s.joinClazzDirectly(ctx, &clazz, userObjID)
}

// joinClazzDirectly 直接加入班级
func (s *ClazzService) joinClazzDirectly(ctx context.Context, clazz *models.Clazz, userObjID primitive.ObjectID) error {
	clazzObjID := clazz.ID

	// 更新班级成员数量
	filter := bson.M{
		"_id":   clazzObjID,
		"$expr": bson.M{"$lt": []interface{}{"$add_nums", "$max_members"}}, // 乐观锁
	}

	update := bson.M{
		"$inc": bson.M{
			"add_nums": 1,
		},
		"$set": bson.M{
			"mtime": time.Now(),
		},
	}

	result, err := utils.GetCollection("clazzes").UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("加入班级失败: " + err.Error())
	}

	if result.MatchedCount == 0 {
		return errors.New("班级不存在或已满")
	}

	// 创建学生班级关联记录
	studentClass := &models.StudentClass{
		ID:        primitive.NewObjectID(),
		StudentID: userObjID,
		ClassID:   clazzObjID,
		CourseID:  clazz.CourseId,
		JoinTime:  time.Now(),
		CTime:     time.Now(),
		MTime:     time.Now(),
	}

	_, err = utils.GetCollection("student_classes").InsertOne(ctx, studentClass)
	if err != nil {
		// 如果创建关联记录失败，回滚班级成员数量
		_, _ = utils.GetCollection("clazzes").UpdateOne(ctx, bson.M{"_id": clazzObjID}, bson.M{"$inc": bson.M{"add_nums": -1}})
		return errors.New("加入班级失败: " + err.Error())
	}

	return nil
}

// DeleteClazz 删除班级
func (s *CourseService) DeleteClazz(ctx context.Context, clazzID string, userID string) error {
	// 此方法已迁移到 clazz 服务中
	return errors.New("此方法已迁移到 clazz 服务中")
}

// AddClazzMember 添加班级成员
func (s *CourseService) AddClazzMember(ctx context.Context, clazzID string, memberID string, operatorID string) error {
	// 此方法已迁移到 clazz 服务中
	return errors.New("此方法已迁移到 clazz 服务中")
}

// RemoveClazzMembers 批量移除班级成员
func (s *CourseService) RemoveClazzMembers(ctx context.Context, req models.RemoveClazzMembersRequest, operatorID string) error {
	// 此方法已迁移到 clazz 服务中
	return errors.New("此方法已迁移到 clazz 服务中")
}

// FinishTask 完成任务
func (s *ClazzService) FinishTask(ctx context.Context, req models.FinishTaskRequest, userID string) error {
	// 解析ID
	taskObjId, err := primitive.ObjectIDFromHex(req.TaskID)
	if err != nil {
		return errors.New("无效的任务ID")
	}

	relationObjId, err := primitive.ObjectIDFromHex(req.RelationID)
	if err != nil {
		return errors.New("无效的关系ID")
	}

	clazzObjId, err := primitive.ObjectIDFromHex(req.ClazzID)
	if err != nil {
		return errors.New("无效班级")
	}

	userObjId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 获取班级信息
	coll := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = coll.FindOne(ctx, bson.M{"_id": clazzObjId}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 检查用户是否是班级成员（通过查询学生班级关联表）
	count, err := utils.GetCollection("student_classes").CountDocuments(ctx, bson.M{
		"student_id": userObjId,
		"class_id":   clazzObjId,
	})

	if err != nil {
		return errors.New("检查班级成员失败: " + err.Error())
	}

	if count == 0 {
		return errors.New("您不是该班级成员，无法完成任务")
	}

	// 获取任务信息
	collTasks := utils.GetCollection("tasks")
	var task models.Task
	if err = collTasks.FindOne(ctx, bson.M{"_id": taskObjId}).Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("任务不存在")
		}
		return err
	}

	// 检查任务状态
	if task.StartTime.After(time.Now()) {
		return errors.New("任务还未开始")
	}

	if task.EndTime != nil && task.EndTime.Before(time.Now()) {
		return errors.New("任务已经结束")
	}

	// 检查 user_relation_question 表中是否已经有完成记录
	userRelationColl := utils.GetCollection("user_relation_question")
	count, err = userRelationColl.CountDocuments(ctx, bson.M{
		"task_id":     taskObjId,
		"user_id":     userObjId,
		"relation_id": relationObjId,
		"state":       1, // 1 表示已完成
	})
	if err != nil {
		return errors.New("检查题目完成状态失败: " + err.Error())
	}

	// 如果已经有完成记录，直接返回
	if count > 0 {
		return nil
	}

	// 在 user_relation_question 表中添加完成记录
	now := time.Now()
	userRelation := &models.UserRelationQuestion{
		ID:         primitive.NewObjectID(),
		RelationID: relationObjId,
		TaskID:     taskObjId,
		UserID:     userObjId,
		State:      1, // 已完成
		CTime:      now,
		MTime:      now,
	}

	_, err = userRelationColl.InsertOne(ctx, userRelation)
	if err != nil {
		return errors.New("记录题目完成状态失败: " + err.Error())
	}

	// 检查是否所有题目都已完成，如果是，则更新任务完成状态
	finishedCount, err := userRelationColl.CountDocuments(ctx, bson.M{
		"task_id": taskObjId,
		"user_id": userObjId,
		"state":   1, // 1 表示已完成
	})

	if err != nil {
		// 记录错误但不中断响应
		log.Printf("查询用户完成题目数量失败: %v", err)
	}

	// 如果已完成题目数等于总题目数，表示任务已完成
	totalQuestions := len(task.RelationIDs)
	isTaskCompleted := int(finishedCount) == totalQuestions && totalQuestions > 0

	// 更新 user_task 表中的状态
	userTaskColl := utils.GetCollection("user_task")

	// 查找是否已存在记录
	var existingUserTask models.UserTask
	err = userTaskColl.FindOne(ctx, bson.M{
		"task_id": taskObjId,
		"user_id": userObjId,
	}).Decode(&existingUserTask)

	if err == nil {
		// 更新现有记录
		updateFields := bson.M{
			"finished_count": int(finishedCount),
			"mtime":          now,
		}

		// 如果任务已完成，更新完成状态和完成时间
		if isTaskCompleted {
			completedAt := now
			updateFields["state"] = 1
			updateFields["completed_at"] = &completedAt
		}

		_, err = userTaskColl.UpdateOne(ctx, bson.M{
			"_id": existingUserTask.ID,
		}, bson.M{
			"$set": updateFields,
		})
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		// 创建新记录
		newUserTask := &models.UserTask{
			ID:            primitive.NewObjectID(),
			TaskID:        taskObjId,
			UserID:        userObjId,
			State:         0, // 默认未完成
			FinishedCount: int(finishedCount),
			CTime:         now,
			MTime:         now,
		}

		// 如果任务已完成，更新完成状态和完成时间
		if isTaskCompleted {
			completedAt := now
			newUserTask.State = 1
			newUserTask.CompletedAt = &completedAt
		}

		_, err = userTaskColl.InsertOne(ctx, newUserTask)
	}

	if err != nil {
		// 记录错误但不中断响应
		log.Printf("更新 user_task 集合失败: %v", err)
	}

	// 如果任务已完成，更新任务的完成用户列表
	if isTaskCompleted {
		filter := bson.M{"_id": taskObjId}
		update := bson.M{
			"$addToSet": bson.M{
				"finish_ids": userObjId,
			},
			"$set": bson.M{
				"mtime": time.Now(),
			},
		}

		_, err = collTasks.UpdateOne(ctx, filter, update)
		if err != nil {
			// 记录错误但不中断响应
			log.Printf("更新任务完成状态失败: %v", err)
		}
	}

	return nil
}

func (s *ClazzService) UpdateTask(ctx context.Context, userId string, req models.UpdateTaskRequest) error {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	courseObjId, err := primitive.ObjectIDFromHex(req.CourseId)
	if err != nil {
		return err
	}
	_, err = primitive.ObjectIDFromHex(req.ClazzId)
	if err != nil {
		return err
	}
	taskObjId, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return err
	}
	if flag, err := authorized(ctx, utils.GetCollection("courses"), courseObjId, userObjId); err != nil || !flag {
		if err != nil {
			return err
		}
		return errors.New("权限不足")
	}

	// 构建更新字段
	updateFields := bson.M{}
	// 检查每个字段，如果字段非空，则加入更新内容
	if req.Title != nil {
		updateFields["title"] = *req.Title
	}
	if req.Description != nil {
		updateFields["description"] = *req.Description
	}
	if req.Type != nil {
		updateFields["type"] = *req.Type
	}
	if req.StartTime != nil {
		updateFields["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updateFields["end_time"] = *req.EndTime
	}
	updateFields["c_id"] = userObjId
	updateFields["mtime"] = time.Now()

	collTasks := utils.GetCollection("tasks")
	filter := bson.M{"_id": taskObjId}
	update := bson.M{"$set": updateFields}
	result, err := collTasks.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("找不到该任务")
	}
	return nil
}

// AddTaskRelationIds 为任务添加关系ID
func (s *ClazzService) AddTaskRelationIds(ctx context.Context, userId string, req models.AddTaskRelationIdsRequest) error {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	taskObjId, err := primitive.ObjectIDFromHex(req.TaskID)
	if err != nil {
		return err
	}

	// 获取任务信息以验证权限
	collTasks := utils.GetCollection("tasks")
	var task models.Task
	if err = collTasks.FindOne(ctx, bson.M{"_id": taskObjId}).Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("任务不存在")
		}
		return err
	}

	// 验证权限（只有课程创建者或班级教师可以修改任务）
	if flag, err := authorized(ctx, utils.GetCollection("courses"), task.CourseId, userObjId); err != nil || !flag {
		if err != nil {
			return err
		}
		return errors.New("权限不足")
	}

	// 转换RelationIDs
	relationObjIds := make([]primitive.ObjectID, len(req.RelationIDs))
	for i, id := range req.RelationIDs {
		hex, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		relationObjIds[i] = hex
	}

	// 使用$addToSet添加关系ID，避免重复
	filter := bson.M{"_id": taskObjId}
	update := bson.M{
		"$addToSet": bson.M{
			"relation_ids": bson.M{"$each": relationObjIds},
		},
		"$set": bson.M{
			"c_id":  userObjId,
			"mtime": time.Now(),
		},
	}

	result, err := collTasks.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("找不到该任务")
	}

	return nil
}

// RemoveTaskRelationIds 从任务中删除关系ID
func (s *ClazzService) RemoveTaskRelationIds(ctx context.Context, userId string, req models.RemoveTaskRelationIdsRequest) error {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	taskObjId, err := primitive.ObjectIDFromHex(req.TaskID)
	if err != nil {
		return err
	}

	// 获取任务信息以验证权限
	collTasks := utils.GetCollection("tasks")
	var task models.Task
	if err = collTasks.FindOne(ctx, bson.M{"_id": taskObjId}).Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("任务不存在")
		}
		return err
	}

	// 验证权限（只有课程创建者或班级教师可以修改任务）
	if flag, err := authorized(ctx, utils.GetCollection("courses"), task.CourseId, userObjId); err != nil || !flag {
		if err != nil {
			return err
		}
		return errors.New("权限不足")
	}

	// 转换RelationIDs
	relationObjIds := make([]primitive.ObjectID, len(req.RelationIDs))
	for i, id := range req.RelationIDs {
		hex, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		relationObjIds[i] = hex
	}

	// 使用$pull删除关系ID
	filter := bson.M{"_id": taskObjId}
	update := bson.M{
		"$pullAll": bson.M{
			"relation_ids": relationObjIds,
		},
		"$set": bson.M{
			"c_id":  userObjId,
			"mtime": time.Now(),
		},
	}

	result, err := collTasks.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("找不到该任务")
	}

	return nil
}

// CopyTaskToClass 将一个班级的任务复制到另一个班级
func (s *ClazzService) CopyTaskToClass(ctx context.Context, userID string, req models.CopyTaskToClassRequest) error {
	// 验证用户ID格式
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 验证源班级ID格式
	sourceClassObjID, err := primitive.ObjectIDFromHex(req.SourceClassID)
	if err != nil {
		return errors.New("无效的源班级ID")
	}

	// 验证目标班级ID格式
	targetClassObjID, err := primitive.ObjectIDFromHex(req.TargetClassID)
	if err != nil {
		return errors.New("无效的目标班级ID")
	}

	// 验证任务ID格式
	taskObjID, err := primitive.ObjectIDFromHex(req.TaskID)
	if err != nil {
		return errors.New("无效的任务ID")
	}

	// 获取源班级信息
	sourceClassColl := utils.GetCollection("clazzes")
	var sourceClass models.Clazz
	if err = sourceClassColl.FindOne(ctx, bson.M{"_id": sourceClassObjID}).Decode(&sourceClass); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("源班级不存在")
		}
		return err
	}

	// 获取目标班级信息
	targetClassColl := utils.GetCollection("clazzes")
	var targetClass models.Clazz
	if err = targetClassColl.FindOne(ctx, bson.M{"_id": targetClassObjID}).Decode(&targetClass); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("目标班级不存在")
		}
		return err
	}

	// 验证源班级和目标班级属于同一课程
	if sourceClass.CourseId != targetClass.CourseId {
		return errors.New("源班级和目标班级必须属于同一课程")
	}

	// 验证用户权限：只有课程创建者或课程教师可以复制任务
	courseColl := utils.GetCollection("courses")
	isAuthorized, err := authorized(ctx, courseColl, sourceClass.CourseId, userObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或课程教师可以复制任务")
	}

	// 获取源任务信息
	taskColl := utils.GetCollection("tasks")
	var sourceTask models.Task
	if err = taskColl.FindOne(ctx, bson.M{"_id": taskObjID, "clazz_id": sourceClassObjID}).Decode(&sourceTask); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("源任务不存在或不属于源班级")
		}
		return err
	}

	// 创建新任务（复制源任务的信息，但关联到目标班级）
	now := time.Now()
	newTask := &models.Task{
		ID:          primitive.NewObjectID(),
		Title:       sourceTask.Title,
		Description: sourceTask.Description,
		Type:        sourceTask.Type,
		StartTime:   sourceTask.StartTime,
		EndTime:     sourceTask.EndTime,
		RelationIDs: sourceTask.RelationIDs,
		Status:      sourceTask.Status,
		CourseId:    sourceTask.CourseId,
		ClazzId:     targetClassObjID, // 关联到目标班级
		CTime:       now,
		MTime:       now,
		CID:         userObjID,
	}

	// 插入新任务到数据库
	_, err = taskColl.InsertOne(ctx, newTask)
	if err != nil {
		return errors.New("复制任务失败: " + err.Error())
	}

	return nil
}

// PageQueryTaskCompletion 分页查询班级任务完成情况
func (s *ClazzService) PageQueryTaskCompletion(ctx context.Context, req *models.PageQueryTaskCompletionRequest) (*models.PageQueryTaskCompletionResponse, error) {
	// 解析任务ID
	taskObjID, err := primitive.ObjectIDFromHex(req.TaskID)
	if err != nil {
		return nil, errors.New("无效的任务ID")
	}

	// 解析班级ID
	classObjID, err := primitive.ObjectIDFromHex(req.ClassID)
	if err != nil {
		return nil, errors.New("无效的班级ID")
	}

	// 设置默认分页参数
	pageNum := int64(1)
	pageSize := int64(10)
	if req.PageNum != nil {
		pageNum = *req.PageNum
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}
	// 设置合理的限制
	if pageSize > 50 {
		pageSize = 50
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageNum < 1 {
		pageNum = 1
	}

	// 验证班级是否存在
	clazzColl := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = clazzColl.FindOne(ctx, bson.M{"_id": classObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("班级不存在")
		}
		return nil, errors.New("查询班级信息失败: " + err.Error())
	}

	// 查询班级的所有学生关联记录
	studentClassColl := utils.GetCollection("student_classes")
	studentClassFilter := bson.M{
		"class_id": classObjID,
		"status":   "active",
	}

	// 查询班级学生总数
	total, err := studentClassColl.CountDocuments(ctx, studentClassFilter)
	if err != nil {
		return nil, errors.New("查询班级学生总数失败: " + err.Error())
	}

	// 分页查询班级学生关联记录
	findOptions := options.Find().
		SetSkip((pageNum - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.D{{"join_time", -1}})

	cursor, err := studentClassColl.Find(ctx, studentClassFilter, findOptions)
	if err != nil {
		return nil, errors.New("查询班级学生关联记录失败: " + err.Error())
	}
	defer cursor.Close(ctx)

	var studentClasses []models.StudentClass
	if err = cursor.All(ctx, &studentClasses); err != nil {
		return nil, errors.New("解析班级学生关联数据失败: " + err.Error())
	}

	// 收集学生ID
	studentIDs := make([]primitive.ObjectID, len(studentClasses))
	for i, sc := range studentClasses {
		studentIDs[i] = sc.StudentID
	}

	// 构建学生查询条件
	userFilter := bson.M{"_id": bson.M{"$in": studentIDs}}

	// 如果指定了学生ID进行精确查询
	if req.UserID != nil && *req.UserID != "" {
		userObjID, err := primitive.ObjectIDFromHex(*req.UserID)
		if err != nil {
			return nil, errors.New("无效的学生ID")
		}
		userFilter = bson.M{"_id": userObjID}
	}

	// 如果指定了学生姓名进行模糊查询
	if req.RealName != nil && *req.RealName != "" {
		userFilter["real_name"] = bson.M{"$regex": *req.RealName, "$options": "i"}
	}

	// 查询学生信息
	userColl := utils.GetCollection("users")
	userCursor, err := userColl.Find(ctx, userFilter)
	if err != nil {
		return nil, errors.New("查询学生信息失败: " + err.Error())
	}
	defer userCursor.Close(ctx)

	var students []*models.User
	if err = userCursor.All(ctx, &students); err != nil {
		return nil, errors.New("解析学生信息失败: " + err.Error())
	}

	// 构建学生ID到学生信息的映射
	studentMap := make(map[primitive.ObjectID]*models.User)
	for _, student := range students {
		studentMap[student.ID] = student
	}

	// 查询任务信息
	taskColl := utils.GetCollection("tasks")
	var task models.Task
	if err = taskColl.FindOne(ctx, bson.M{"_id": taskObjID}).Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("任务不存在")
		}
		return nil, errors.New("获取任务信息失败: " + err.Error())
	}

	totalQuestions := len(task.RelationIDs)

	// 查询这些学生的任务完成情况
	userTaskColl := utils.GetCollection("user_task")
	userTaskFilter := bson.M{
		"task_id": taskObjID,
		"user_id": bson.M{"$in": studentIDs},
	}
	userTaskCursor, err := userTaskColl.Find(ctx, userTaskFilter)
	if err != nil {
		return nil, errors.New("查询任务完成情况失败: " + err.Error())
	}
	defer userTaskCursor.Close(ctx)

	var userTasks []models.UserTask
	if err = userTaskCursor.All(ctx, &userTasks); err != nil {
		return nil, errors.New("解析任务完成情况数据失败: " + err.Error())
	}

	// 构建用户ID到任务完成情况的映射
	userTaskMap := make(map[primitive.ObjectID]*models.UserTask)
	for _, userTask := range userTasks {
		userTaskMap[userTask.UserID] = &userTask
	}

	// 构建响应数据
	completion := make([]models.TaskCompletionResponse, len(students))
	for i, student := range students {
		userTask, exists := userTaskMap[student.ID]
		completed := false
		var completedAt *time.Time
		finishedQuestions := 0

		if exists {
			completed = userTask.State == 1
			completedAt = userTask.CompletedAt
			finishedQuestions = userTask.FinishedCount
		}

		completion[i] = models.TaskCompletionResponse{
			UserID:            student.ID,
			RealName:          student.RealName,
			StudentID:         student.StudentID,
			Completed:         completed,
			CompletedAt:       completedAt,
			TotalQuestions:    totalQuestions,
			FinishedQuestions: finishedQuestions,
		}
	}

	res := &models.PageQueryTaskCompletionResponse{
		Total:      total,
		PageNum:    pageNum,
		PageSize:   pageSize,
		Completion: completion,
	}

	return res, nil
}
