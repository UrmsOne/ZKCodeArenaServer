/*
@Author: sir
@Date: 2025/9/24
@Name: service_course.go
@Description: 课程服务层实现（支持多班级嵌入式架构）
*/

package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"
	"zk-code-arena-server/conf"
	taskstrategy "zk-code-arena-server/pkg/app/api-server/service/task-servies"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/utils"

	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CourseService 课程服务
type CourseService struct{}

// NewCourseService 创建课程服务实例
func NewCourseService() *CourseService {
	return &CourseService{}
}

// CreateCourse 创建课程
func (s *CourseService) CreateCourse(ctx context.Context, req *models.CreateCourseRequest, creatorID string) (string, error) {
	// 验证创建者权限
	creatorObjID, err := primitive.ObjectIDFromHex(creatorID)
	if err != nil {
		return "", errors.New("无效的用户ID")
	}

	// 创建课程对象
	now := time.Now()
	course := &models.Course{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   creatorObjID,
		Status:      models.CourseStatusActive,
		CTime:       now,
		MTime:       now,
	}

	// 插入数据库
	_, err = utils.GetCollection("courses").InsertOne(ctx, course)
	if err != nil {
		log.Printf("创建课程失败: %v", err)
		return "", errors.New("创建课程失败")
	}

	return course.ID.Hex(), nil
}

// CreateClass 为课程创建班级
func (s *CourseService) CreateClass(ctx context.Context, req *models.CreateClazzRequest, userId string) (*models.ClazzResponse, error) {
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
		log.Printf("查询失败: %v", err)
		return nil, err
	}

	//TODO 这里业务到底是只有课程创建者才能创建班级还是老师也可以
	if result.CreatorID.Hex() != userId {
		return nil, errors.New("权限不足")
	}

	var maxMembers int
	if req.MaxMembers == nil {
		maxMembers = 60
	} else if *req.MaxMembers > 100 {
		maxMembers = 100
	} else {
		maxMembers = *req.MaxMembers
	}

	// 创建班级对象，包含从课程中获取的教师信息
	now := time.Now()
	clazz := &models.Clazz{
		ID:            primitive.NewObjectID(),
		Name:          req.Name,
		CourseId:      courseID,
		Description:   req.Description,
		Schedule:      req.Schedule,
		MemberIDs:     []primitive.ObjectID{},
		TeacherIds:    result.TeacherIds,
		RequireInvite: req.RequireInvite,
		MaxMembers:    maxMembers,
		AddNums:       0,
		Status:        models.ClassStatusActive, // 默认为活跃状态，支持加入
		CTime:         now,
		MTime:         now,
	}

	insertResult, err := utils.GetCollection("clazzes").InsertOne(ctx, clazz)
	if err != nil {
		log.Printf("插入班级失败: %v", err)
		return nil, errors.New("创建班级失败")
	}

	// 生成邀请码和二维码
	inviteCode := s.generateInviteCode()

	qrCode, err := s.generateQRCode(insertResult.InsertedID.(primitive.ObjectID).Hex(), inviteCode)
	p := ""
	if err == nil {
		p = base64.StdEncoding.EncodeToString(qrCode)
		// 缓存邀请码
		s.cacheInviteCode(ctx, clazz.ID.Hex(), inviteCode, p)
	}
	return &models.ClazzResponse{
		ClazzId: clazz.ID.Hex(),
		QRCode:  p,
	}, nil
}

// JoinClazz 加入班级
func (s *CourseService) JoinClazz(ctx context.Context, req *models.JoinClazzRequest, userID string) error {
	c := utils.GetCollection("clazzes")
	classObjID, err := primitive.ObjectIDFromHex(req.ClazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	filter := bson.M{"_id": classObjID}
	var clazz models.Clazz
	if err = c.FindOne(ctx, filter).Decode(&clazz); err != nil {
		return errors.New("班级不存在")
	}

	// 检查是否可以加入
	if !clazz.CanJoin() {
		return errors.New("班级已满或已结束")
	}

	memberId, _ := primitive.ObjectIDFromHex(userID)
	// 检查是否已是成员
	if clazz.IsMember(memberId) {
		return errors.New("您已经是该班级成员")
	}

	// 验证邀请码（如果需要）
	if clazz.RequireInvite {
		if req.InviteCode == nil {
			return errors.New("加入该班级需要邀请码")
		}
		if res, err := s.validateInviteCode(ctx, req.ClazzID, *req.InviteCode); err != nil || !res {
			if err != nil {
				return err
			}
			return errors.New("邀请码错误或已过期")
		}
	}

	filter = bson.M{
		"_id":   clazz.ID,
		"$expr": bson.M{"$lt": []interface{}{"$add_nums", "$max_members"}}, //乐观锁
	}

	update := bson.M{
		"$addToSet": bson.M{
			"member_ids": memberId,
		},
		"$inc": bson.M{
			"add_nums": 1,
		},
		"$set": bson.M{
			"mtime": time.Now(),
		},
	}

	result, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("加入班级失败: " + err.Error())
	}

	if result.MatchedCount == 0 {
		return errors.New("班级不存在")
	}

	return nil
}

// GetCourseByID 根据ID获取课程详情
func (s *CourseService) GetCourseByID(ctx context.Context, courseID string, userId string) (*models.CourseResponse, error) {
	courseObjID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		return nil, errors.New("无效的课程ID")
	}

	course, err := s.getCourseByID(ctx, courseObjID)
	if err != nil {
		return nil, errors.New("课程不存在")
	}

	//校验权限
	exist := contains(course.TeacherIds, userId)
	if !exist && course.CreatedBy.Hex() != userId {
		return nil, errors.New("权限不足")
	}

	// 获取创建者信息
	creator, err := s.getUserByID(ctx, course.CreatedBy)
	if err != nil {
		return nil, errors.New("获取创建者信息失败: " + err.Error())
	}

	var creatorProfile *models.UserProfile
	if creator != nil {
		creatorProfile = creator.ToProfile()
	}

	return &models.CourseResponse{
		ID:            course.ID,
		Name:          course.Name,
		Avatar:        course.Avatar,
		Description:   course.Description,
		TeacherIds:    course.TeacherIds,
		Status:        course.Status,
		CTime:         course.CTime,
		CreatedByUser: creatorProfile,
	}, nil
}

func (s *CourseService) UpdateCourseAvatar(ctx context.Context, file multipart.File, header *multipart.FileHeader, userId string, courseId string) error {
	courseObjId, err := primitive.ObjectIDFromHex(courseId)
	if err != nil {
		return err
	}

	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	coll := utils.GetCollection("courses")
	if auth, err := authorized(ctx, coll, courseObjId, userObjId); err != nil || auth != true {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("查不到课程该课程")
		}
		return errors.New("权限不足")
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	// 将文件内容转为Base64字符串
	base64Str := base64.StdEncoding.EncodeToString(fileBytes)

	filter := bson.M{"_id": courseObjId}
	update := bson.M{"$set": bson.M{"avatar": base64Str}}
	if err = coll.FindOneAndUpdate(ctx, filter, update).Err(); err != nil {
		return err
	}
	return nil
}

func (s *CourseService) AddCourseTask(ctx context.Context, req *models.AddTaskRequest, userId string) error {
	courseObjId, err := primitive.ObjectIDFromHex(req.CourseId)
	if err != nil {
		return err
	}
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	clazzId, err := primitive.ObjectIDFromHex(req.ClazzId)
	if err != nil {
		return err
	}

	if flag, err := authorized(ctx, utils.GetCollection("courses"), courseObjId, userObjId); !flag || err != nil {
		if err != nil {
			return err
		}
		return errors.New("权限不足")
	}

	// StartTime 是必需的，不使用指针
	startTime := req.StartTime
	status := models.TaskStatusNotStarted
	// 如果开始时间小于等于当前时间，设置为活跃状态TODO,写一个定时任务将状态更改
	if startTime.Before(time.Now()) || startTime.Equal(time.Now()) {
		status = models.TaskStatusActive
	}

	if req.EndTime != nil && req.EndTime.Before(startTime) {
		return errors.New("无效的时间选择")
	}

	ids := make([]primitive.ObjectID, len(req.RelationIDs))
	for i, id := range req.RelationIDs {
		hex, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		ids[i] = hex
	}

	now := time.Now()
	task := &models.Task{
		ID:          primitive.NewObjectID(),
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		StartTime:   startTime,
		EndTime:     req.EndTime,
		RelationIDs: ids,
		Status:      status,
		CourseId:    courseObjId,
		ClazzId:     clazzId,
		CTime:       now,
		MTime:       now,
		CID:         userObjId,
	}

	// 插入到独立的tasks集合中
	coll := utils.GetCollection("tasks")
	_, err = coll.InsertOne(ctx, task)
	if err != nil {
		return err
	}

	return nil
}

func (s *CourseService) PageQueryCourse(ctx context.Context, request *models.PageQueryCourseRequest, userId string) (*models.PageQueryCourseResponse, error) {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}

	// 设置默认值
	pageNum := int64(1)
	pageSize := int64(3)
	if request.PageNum != nil {
		pageNum = *request.PageNum
	}
	if request.PageSize != nil {
		pageSize = *request.PageSize
	}
	// 设置合理的限制
	if pageSize > 100 {
		pageSize = 100
	}

	filter := bson.D{bson.E{Key: "created_by", Value: userObjId}}
	// 只有当 name 不为 nil 且不为空时才添加 name 查询条件
	if request.Name != nil && *request.Name != "" {
		filter = append(filter, bson.E{"name", bson.D{
			{"$regex", *request.Name},
			{"$options", "i"},
		}})
	}
	// 只有当 status 不为 nil 时才添加 status 查询条件
	if request.Status != nil {
		filter = append(filter, bson.E{"status", *request.Status})
	}
	coll := utils.GetCollection("courses")
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	findOptions := options.Find().SetSkip((pageNum - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.D{
			{"status", -1},
			{"ctime", -1},
		})
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []models.Course
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	res := &models.PageQueryCourseResponse{
		Total:    total,
		PageNum:  pageNum,
		PageSize: pageSize,
		Courses:  results,
	}
	return res, nil
}

// PageQueryTeacherCourses 分页查询老师加入的课程
func (s *CourseService) PageQueryTeacherCourses(ctx context.Context, request *models.PageQueryTeacherCoursesRequest, userId string) (*models.PageQueryCourseResponse, error) {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}

	// 设置默认值
	pageNum := int64(1)
	pageSize := int64(10)
	if request.PageNum != nil {
		pageNum = *request.PageNum
	}
	if request.PageSize != nil {
		pageSize = *request.PageSize
	}
	// 设置合理的限制
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询条件：老师ID在teacher_ids数组中
	filter := bson.M{"teacher_ids": bson.M{"$in": []primitive.ObjectID{userObjId}}}

	// 只有当 name 不为 nil 且不为空时才添加 name 查询条件
	if request.Name != nil && *request.Name != "" {
		filter["name"] = bson.M{"$regex": *request.Name, "$options": "i"}
	}

	// 只有当 status 不为 nil 时才添加 status 查询条件
	if request.Status != nil {
		filter["status"] = *request.Status
	}

	coll := utils.GetCollection("courses")
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	findOptions := options.Find().SetSkip((pageNum - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.D{
			{"status", -1},
			{"ctime", -1},
		})
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []models.Course
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	res := &models.PageQueryCourseResponse{
		Total:    total,
		PageNum:  pageNum,
		PageSize: pageSize,
		Courses:  results,
	}
	return res, nil
}

// GetClazzesByCourseId 通过课程ID查询所有班级
func (s *CourseService) GetClazzesByCourseId(ctx context.Context, courseId string, userId string) ([]models.Clazz, error) {
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
	isAuthorized, err := authorized(ctx, coll, courseObjId, userObjId)
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

// UpdateClazzInfo 更新班级信息（不包括教师和成员）
func (s *CourseService) UpdateClazzInfo(ctx context.Context, clazzID string, userId string, req *models.UpdateClazzRequest) error {
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

	isAuthorized, err := authorized(ctx, courseColl, courseObjID, userObjID)
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

func (s *CourseService) DeleteTask(ctx context.Context, userId string, taskId string) error {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	taskObjId, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		return err
	}

	// 获取任务信息以验证权限
	collTasks := utils.GetCollection("tasks")
	var task models.Task
	taskFilter := bson.M{"_id": taskObjId}
	if err = collTasks.FindOne(ctx, taskFilter).Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("任务不存在")
		}
		return err
	}

	// 验证用户是否有权限删除任务
	collCourse := utils.GetCollection("courses")
	if flag, err := authorized(ctx, collCourse, task.CourseId, userObjId); !flag || err != nil {
		if err != nil {
			return err
		}
		return errors.New("权限不足")
	}

	// 从独立的tasks集合中删除任务
	_, err = collTasks.DeleteOne(ctx, taskFilter)
	if err != nil {
		return err
	}

	return nil
}

// RefreshInviteCode 刷新邀请码
func (s *CourseService) RefreshInviteCode(ctx context.Context, courseID, classID, userID string) (*models.QrCodeResponse, error) {
	// 验证权限
	courseObjID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		return nil, errors.New("无效的课程ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	if auth, err := authorized(ctx, utils.GetCollection("courses"), courseObjID, userObjID); err != nil || !auth {
		if err != nil {
			return nil, errors.New("课程不存在")
		}
		return nil, errors.New("权限不足")
	}

	// 生成新的邀请码和二维码
	inviteCode := s.generateInviteCode()
	qrCode, err := s.generateQRCode(classID, inviteCode)
	if err != nil {
		return nil, errors.New("生成二维码失败")
	}

	// 更新缓存
	s.cacheInviteCode(ctx, classID, inviteCode, base64.StdEncoding.EncodeToString(qrCode))
	var response = &models.QrCodeResponse{
		QrCode:     qrCode,
		InviteCode: inviteCode,
	}

	return response, nil
}

func (s *CourseService) RemoveCourse(ctx context.Context, userId string, courseId string) error {
	userobjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	courseObjId, err := primitive.ObjectIDFromHex(courseId)
	if err != nil {
		return err
	}
	courseFilter := bson.D{
		{"_id", courseObjId},
		{"created_by", userobjId},
	}
	coll := utils.GetCollection("courses")

	// 不再使用事务，直接进行操作
	var course models.Course
	if err = coll.FindOne(ctx, courseFilter).Decode(&course); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("找不到该课程")
		}
		return err
	}

	if course.ID.Hex() == "" {
		return errors.New("权限不足")
	}

	//检查是否课程还有学生
	collClazz := utils.GetCollection("clazzes")
	projection := bson.D{
		{"add_nums", 1},
	}
	clazzFilter := bson.M{"course_id": courseObjId}

	cursor, err := collClazz.Find(ctx, clazzFilter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	//遍历结果集
	for cursor.Next(ctx) {
		var result struct {
			AddNums int32 `bson:"add_nums"`
		}
		if err := cursor.Decode(&result); err != nil {
			return err
		}
		if result.AddNums != 0 {
			return errors.New("课程学生未清空，不允许删除课程")
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	// 删除课程和相关班级，使用独立的操作而不是事务
	if _, err := coll.DeleteOne(ctx, bson.M{"_id": courseObjId}); err != nil {
		return err
	}

	if _, err := collClazz.DeleteMany(ctx, bson.M{"course_id": courseObjId}); err != nil {
		return err
	}

	return nil
}

func (s *CourseService) UpdateCourseInfo(ctx context.Context, userId string, req *models.UpdateCourseRequest) error {
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	courseObjId, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return err
	}
	update := bson.M{}

	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}
	update["mtime"] = time.Now()

	filter := bson.M{
		"_id":        courseObjId,
		"created_by": userObjId, // 只允许创建者修改
	}

	res, err := utils.GetCollection("courses").UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("课程不存在或权限不够")
	}
	return nil
}

func (s *CourseService) UpdateTask(ctx context.Context, userId string, req models.UpdateTaskRequest) error {
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
	if req.CourseId != "" {
		updateFields["course_id"] = req.CourseId
	}
	if req.ClazzId != "" {
		updateFields["clazz_id"] = req.ClazzId
	}
	if req.Title != "" {
		updateFields["title"] = req.Title
	}
	if req.Description != "" {
		updateFields["description"] = req.Description
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
	if req.RelationIDs != nil {
		// 在应用层去重 RelationIDs
		uniqueIDs := make(map[string]bool)
		uniqueRelationIDs := make([]string, 0)
		for _, id := range *req.RelationIDs {
			if !uniqueIDs[id] {
				uniqueIDs[id] = true
				uniqueRelationIDs = append(uniqueRelationIDs, id)
			}
		}
		updateFields["relation_ids"] = uniqueRelationIDs
	}

	// 如果没有任何字段需要更新，返回早期退出
	if len(updateFields) == 0 {
		return errors.New("无更改内容")
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

// GetClazzByID 获取班级详情
func (s *CourseService) GetClazzByID(ctx context.Context, clazzID string, userID string) (*models.GetClazzResponse, error) {
	clazzObjID, err := primitive.ObjectIDFromHex(clazzID)
	if err != nil {
		return nil, errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	coll := utils.GetCollection("clazzes")
	var clazz models.GetClazzResponse
	if err = coll.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("班级不存在")
		}
		return nil, err
	}

	// 验证用户是否有权限查看该班级
	// 用户必须是班级成员、课程创建者或课程教师
	courseObjID := clazz.CourseId
	courseColl := utils.GetCollection("courses")

	// 检查用户是否是班级成员
	isMember := cheekIsMember(clazz.MemberIDs, userObjID)

	// 检查用户是否是课程创建者或教师
	isAuthorized, err := authorized(ctx, courseColl, courseObjID, userObjID)
	if err != nil {
		return nil, err
	}

	if !isAuthorized && !isMember {
		return nil, errors.New("权限不足")
	}

	return &clazz, nil
}

func cheekIsMember(ids []primitive.ObjectID, userID primitive.ObjectID) bool {
	for _, memberID := range ids {
		if memberID == userID {
			return true
		}
	}
	return false
}

// DeleteClazz 删除班级
func (s *CourseService) DeleteClazz(ctx context.Context, clazzID string, userID string) error {
	clazzObjID, err := primitive.ObjectIDFromHex(clazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	coll := utils.GetCollection("clazzes")

	// 先查询班级信息以验证权限
	var clazz models.Clazz

	filter := bson.M{
		"_id":  clazzObjID,
		"c_id": userObjID,
	}

	if err = coll.FindOne(ctx, filter).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在或权限不够")
		}
		return err
	}

	// 检查班级是否还有成员
	if len(clazz.MemberIDs) > 0 {
		return errors.New("班级还有成员，无法删除")
	}

	// 删除班级
	if _, err = coll.DeleteOne(ctx, filter); err != nil {
		return errors.New("删除班级失败: " + err.Error())
	}

	return nil
}

// AddClazzMember 添加班级成员
func (s *CourseService) AddClazzMember(ctx context.Context, clazzID string, memberID string, operatorID string) error {
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

	isAuthorized, err := authorized(ctx, courseColl, courseObjID, operatorObjID)
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

	// 添加成员
	filter := bson.M{
		"_id":   clazzObjID,
		"$expr": bson.M{"$lt": []interface{}{"$add_nums", "$max_members"}}, //乐观锁
	}

	update := bson.M{
		"$addToSet": bson.M{
			"member_ids": memberObjID,
		},
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

	return nil
}

// RemoveClazzMembers 批量移除班级成员
func (s *CourseService) RemoveClazzMembers(ctx context.Context, req models.RemoveClazzMembersRequest, operatorID string) error {
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

	isAuthorized, err := authorized(ctx, courseColl, courseObjID, operatorObjID)
	if err != nil {
		return err
	}

	if !isAuthorized {
		return errors.New("权限不足，只有课程创建者或教师可以移除成员")
	}

	// 检查成员是否是班级成员
	validMembers := make([]primitive.ObjectID, 0)
	for _, memberObjID := range memberObjIDs {
		if clazz.IsMember(memberObjID) {
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
		"$pullAll": bson.M{
			"member_ids": validMembers,
		},
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

	return nil
}

// FinishTask 完成任务
func (s *CourseService) FinishTask(ctx context.Context, req models.FinishTaskRequest, userId string) error {
	clazzObjId, err := primitive.ObjectIDFromHex(req.ClazzID)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	taskObjId, err := primitive.ObjectIDFromHex(req.TaskID)
	if err != nil {
		return errors.New("无效的任务ID")
	}

	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}
	relationObjId, err := primitive.ObjectIDFromHex(req.RelationID)
	if err != nil {
		return err
	}
	// 验证用户是否有权限完成任务（必须是班级成员）
	collClazz := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = collClazz.FindOne(ctx, bson.M{"_id": clazzObjId}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 检查用户是否是班级成员
	if !clazz.IsMember(userObjId) {
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

	// 检查是否已经完成过该任务
	for _, id := range task.FinishIds {
		if id == userObjId {
			return errors.New("您已经完成过该任务")
		}
	}

	// 使用策略模式处理不同类型的任务
	factory := taskstrategy.NewTaskHandlerFactory()
	handler, err := factory.GetHandler(task.Type)
	if err != nil {
		return err
	}

	// 处理任务完成逻辑
	if err := handler.HandleTask(ctx, &task, relationObjId, userObjId); err != nil {
		return errors.New("处理任务失败: " + err.Error())
	}

	return nil
}

// GetTasksByClazzID 根据班级ID获取任务列表
func (s *CourseService) GetTasksByClazzID(ctx context.Context, clazzID string, userID string) ([]models.Task, error) {
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
	isMember := clazz.IsMember(userObjID)
	collCourse := utils.GetCollection("courses")
	isAuthorized, err := authorized(ctx, collCourse, clazz.CourseId, userObjID)

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

// 私有方法
// getCourseByID 根据ID获取课程
func (s *CourseService) getCourseByID(ctx context.Context, courseID primitive.ObjectID) (*models.Course, error) {
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
func (s *CourseService) getUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
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

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const length = 6

// generateInviteCode 生成邀请码
func (s *CourseService) generateInviteCode() string {

	// 使用加密安全的随机数生成器
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 回退到时间戳方案
		return fmt.Sprintf("%06d", time.Now().Unix()%1000000)
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes)
}

// generateQRCode 生成二维码
func (s *CourseService) generateQRCode(clazzID, inviteCode string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/api/v1/clazzes/join?clazzId=%s&invite_code=%s",
		conf.Config.Server.Domain, clazzID, inviteCode)
	qr, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	return qr, nil
}

// cacheInviteCode 缓存邀请码
func (s *CourseService) cacheInviteCode(ctx context.Context, classID, inviteCode string, p string) {
	client := utils.GetRedisClient()
	if client == nil {
		return
	}

	key := fmt.Sprintf("invite_code:%s", classID)
	pipe := client.Pipeline()
	pipe.HSet(ctx, key, "code", inviteCode)
	pipe.HSet(ctx, key, "qrcode", p)
	pipe.Expire(ctx, key, 24*time.Hour)

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("缓存邀请码失败: %v", err)
	}
}

// validateInviteCode 验证邀请码
func (s *CourseService) validateInviteCode(ctx context.Context, classID, inviteCode string) (bool, error) {
	client := utils.GetRedisClient()
	if client == nil {
		log.Printf("获取redis客户端失败")
		return false, errors.New("获取redis客户端失败")
	}

	key := fmt.Sprintf("invite_code:%s", classID)
	cachedCode, err := client.HGet(ctx, key, "code").Result()
	if err != nil {
		log.Printf("获取缓存邀请码失败: %v", err)
		return false, errors.New("获取缓存邀请码失败")
	}

	return cachedCode == inviteCode, nil
}

func contains(slice []primitive.ObjectID, v string) bool {
	for _, item := range slice {
		if item.Hex() == v {
			return true
		}
	}
	return false
}

func authorized(ctx context.Context, coll *mongo.Collection, courseId primitive.ObjectID, userId primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": courseId}
	var course models.Course
	projection := bson.M{
		"created_by":  1,
		"teacher_ids": 1,
	}
	if err := coll.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&course); err != nil {
		return false, errors.New("课程不存在")
	}
	isTeacher := contains(course.TeacherIds, userId.Hex())
	isCreator := course.CreatedBy == userId
	return isTeacher || isCreator, nil
}

// UpdateClazzTeachers 更新班级的教师信息
func (s *CourseService) UpdateClazzTeachers(ctx context.Context, courseID primitive.ObjectID, teacherIds []primitive.ObjectID) error {
	// 更新所有属于该课程的班级的教师信息
	filter := bson.M{"course_id": courseID}
	update := bson.M{"$set": bson.M{"teacher_ids": teacherIds, "mtime": time.Now()}}

	_, err := utils.GetCollection("clazzes").UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf("更新班级教师信息失败: %v", err)
		return errors.New("更新班级教师信息失败")
	}

	return nil
}

// AddCourseTeacher 为课程添加教师
func (s *CourseService) AddCourseTeacher(ctx context.Context, courseId string, teacherId string, userId string) error {
	// 验证用户权限（只有课程创建者才能添加教师）
	courseObjID, err := primitive.ObjectIDFromHex(courseId)
	if err != nil {
		return errors.New("无效的课程ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	teacherObjID, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return errors.New("无效的教师ID")
	}

	// 检查课程是否存在并验证权限
	coll := utils.GetCollection("courses")
	var course models.Course
	if err = coll.FindOne(ctx, bson.M{"_id": courseObjID}).Decode(&course); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("课程不存在")
		}
		return err
	}

	// 只有课程创建者可以添加教师
	if course.CreatedBy != userObjID {
		return errors.New("权限不足，只有课程创建者可以添加教师")
	}

	// 添加教师到课程
	filter := bson.M{"_id": courseObjID}
	update := bson.M{"$addToSet": bson.M{"teacher_ids": teacherObjID}, "$set": bson.M{"mtime": time.Now()}}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("添加教师失败: " + err.Error())
	}

	return nil
}

// AddCourseTeachers 为课程添加多个教师
func (s *CourseService) AddCourseTeachers(ctx context.Context, courseId string, teacherIds []string, userId string) error {
	// 验证用户权限（只有课程创建者才能添加教师）
	courseObjID, err := primitive.ObjectIDFromHex(courseId)
	if err != nil {
		return errors.New("无效的课程ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 转换教师ID
	teacherObjIDs := make([]primitive.ObjectID, len(teacherIds))
	for i, id := range teacherIds {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("无效的教师ID: " + id)
		}
		teacherObjIDs[i] = objID
	}

	// 检查课程是否存在并验证权限
	coll := utils.GetCollection("courses")
	var course models.Course
	if err = coll.FindOne(ctx, bson.M{"_id": courseObjID}).Decode(&course); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("课程不存在")
		}
		return err
	}

	// 只有课程创建者可以添加教师
	if course.CreatedBy != userObjID {
		return errors.New("权限不足，只有课程创建者可以添加教师")
	}

	// 添加教师到课程
	filter := bson.M{"_id": courseObjID}
	update := bson.M{"$addToSet": bson.M{"teacher_ids": bson.M{"$each": teacherObjIDs}}, "$set": bson.M{"mtime": time.Now()}}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("添加教师失败: " + err.Error())
	}

	return nil
}

// RemoveCourseTeachers 从课程中删除多个教师
func (s *CourseService) RemoveCourseTeachers(ctx context.Context, courseId string, teacherIds []string, userId string) error {
	// 验证用户权限（只有课程创建者才能删除教师）
	courseObjID, err := primitive.ObjectIDFromHex(courseId)
	if err != nil {
		return errors.New("无效的课程ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 转换教师ID
	teacherObjIDs := make([]primitive.ObjectID, len(teacherIds))
	for i, id := range teacherIds {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("无效的教师ID: " + id)
		}
		teacherObjIDs[i] = objID
	}

	// 检查课程是否存在并验证权限
	coll := utils.GetCollection("courses")
	var course models.Course
	if err = coll.FindOne(ctx, bson.M{"_id": courseObjID}).Decode(&course); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("课程不存在")
		}
		return err
	}

	// 只有课程创建者可以删除教师
	if course.CreatedBy != userObjID {
		return errors.New("权限不足，只有课程创建者可以删除教师")
	}

	// 从课程中删除教师
	filter := bson.M{"_id": courseObjID}
	update := bson.M{"$pullAll": bson.M{"teacher_ids": teacherObjIDs}, "$set": bson.M{"mtime": time.Now()}}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("删除教师失败: " + err.Error())
	}

	// 同时从该课程的所有班级中删除这些教师
	clazzFilter := bson.M{"course_id": courseObjID}
	clazzUpdate := bson.M{"$pullAll": bson.M{"teacher_ids": teacherObjIDs}, "$set": bson.M{"mtime": time.Now()}}
	_, err = utils.GetCollection("clazzes").UpdateMany(ctx, clazzFilter, clazzUpdate)
	if err != nil {
		// 这里我们记录日志但不返回错误，因为主要目标是删除课程中的教师
		log.Printf("从班级中删除教师时出错: %v", err)
	}

	return nil
}

// AddClazzTeacher 为班级添加教师
func (s *CourseService) AddClazzTeacher(ctx context.Context, clazzId string, teacherId string, userId string) error {
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

// AddClazzTeachers 为班级添加多个教师
func (s *CourseService) AddClazzTeachers(ctx context.Context, clazzId string, teacherIds []string, userId string) error {
	// 验证用户权限（只有课程创建者才能添加班级教师）
	clazzObjID, err := primitive.ObjectIDFromHex(clazzId)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 转换教师ID
	teacherObjIDs := make([]primitive.ObjectID, len(teacherIds))
	for i, id := range teacherIds {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("无效的教师ID: " + id)
		}
		teacherObjIDs[i] = objID
	}

	// 检查班级是否存在并验证权限
	clazzColl := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = clazzColl.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 只有课程创建者可以添加班级教师
	courseColl := utils.GetCollection("courses")
	var course models.Course
	if err = courseColl.FindOne(ctx, bson.M{"_id": clazz.CourseId}).Decode(&course); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("课程不存在")
		}
		return err
	}

	if course.CreatedBy != userObjID {
		return errors.New("权限不足，只有课程创建者可以添加班级教师")
	}

	// 检查这些教师是否已经是课程的教师
	invalidTeachers := make([]string, 0)
	for _, teacherObjID := range teacherObjIDs {
		found := false
		for _, courseTeacherID := range course.TeacherIds {
			if courseTeacherID == teacherObjID {
				found = true
				break
			}
		}
		if !found {
			invalidTeachers = append(invalidTeachers, teacherObjID.Hex())
		}
	}

	if len(invalidTeachers) > 0 {
		return errors.New("以下教师尚未加入课程，无法添加到班级: " + strings.Join(invalidTeachers, ", "))
	}

	// 添加教师到班级
	filter := bson.M{"_id": clazzObjID}
	update := bson.M{"$addToSet": bson.M{"teacher_ids": bson.M{"$each": teacherObjIDs}}, "$set": bson.M{"mtime": time.Now()}}
	_, err = clazzColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("添加班级教师失败: " + err.Error())
	}

	return nil
}

// RemoveClazzTeacher 为班级移除教师
func (s *CourseService) RemoveClazzTeacher(ctx context.Context, clazzId string, teacherId string, userId string) error {
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
		return errors.New("权限不足，只有课程创建者或课程教师可以移除班级教师")
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

// RemoveClazzTeachers 为班级移除多个教师
func (s *CourseService) RemoveClazzTeachers(ctx context.Context, clazzId string, teacherIds []string, userId string) error {
	// 验证用户权限（只有课程创建者才能移除班级教师）
	clazzObjID, err := primitive.ObjectIDFromHex(clazzId)
	if err != nil {
		return errors.New("无效的班级ID")
	}

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 转换教师ID
	teacherObjIDs := make([]primitive.ObjectID, len(teacherIds))
	for i, id := range teacherIds {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("无效的教师ID: " + id)
		}
		teacherObjIDs[i] = objID
	}

	// 检查班级是否存在并验证权限
	clazzColl := utils.GetCollection("clazzes")
	var clazz models.Clazz
	if err = clazzColl.FindOne(ctx, bson.M{"_id": clazzObjID}).Decode(&clazz); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("班级不存在")
		}
		return err
	}

	// 只有课程创建者可以移除班级教师
	courseColl := utils.GetCollection("courses")
	var course models.Course
	if err = courseColl.FindOne(ctx, bson.M{"_id": clazz.CourseId}).Decode(&course); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("课程不存在")
		}
		return err
	}

	if course.CreatedBy != userObjID {
		return errors.New("权限不足，只有课程创建者可以移除班级教师")
	}

	// 检查是否尝试移除所有教师
	if len(clazz.TeacherIds) <= len(teacherObjIDs) {
		return errors.New("不能移除所有教师，至少需要保留一个教师")
	}

	// 从班级移除教师
	filter := bson.M{"_id": clazzObjID}
	update := bson.M{"$pullAll": bson.M{"teacher_ids": teacherObjIDs}, "$set": bson.M{"mtime": time.Now()}}
	_, err = clazzColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("移除班级教师失败: " + err.Error())
	}

	return nil
}
