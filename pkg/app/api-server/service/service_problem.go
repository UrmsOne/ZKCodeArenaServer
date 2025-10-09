/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service_problem.go
@Description: 题目服务
*/

package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/sandbox"
	"zk-code-arena-server/pkg/utils"
)

type ProblemService struct {
	sandboxClient   sandbox.Client
	testCaseService *TestCaseService
}

func NewProblemService(sandboxClient sandbox.Client, testCaseService *TestCaseService) *ProblemService {
	return &ProblemService{
		sandboxClient:   sandboxClient,
		testCaseService: testCaseService,
	}
}

// CreateProblem 创建题目
func (s *ProblemService) CreateProblem(ctx context.Context, problem *models.Problem) error {
	problem.ID = primitive.NewObjectID()
	problem.CreatedAt = time.Now()
	problem.UpdatedAt = time.Now()
	problem.ACCount = 0
	problem.SubmitCount = 0

	collection := utils.GetCollection("problems")
	_, err := collection.InsertOne(ctx, problem)
	return err
}

// GetProblemByID 根据ID获取题目
func (s *ProblemService) GetProblemByID(ctx context.Context, id primitive.ObjectID) (*models.Problem, error) {
	collection := utils.GetCollection("problems")
	var problem models.Problem
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&problem)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

// GetProblems 获取题目列表
func (s *ProblemService) GetProblems(
	ctx context.Context,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	includePrivate bool,
	role models.UserRole,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	collection := utils.GetCollection("problems")

	// 构建查询条件
	filter := bson.M{}

	if !includePrivate {
		filter["is_public"] = true
		filter["status"] = models.StatusPublished
	} else {
		switch role {
		case models.RoleAdmin:
			// 管理员可以查看全部题目
		case models.RoleTeacher:
			if userID != nil {
				filter["created_by"] = *userID
			} else {
				filter["is_public"] = true
				filter["status"] = models.StatusPublished
			}
		default:
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

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var problems []*models.ProblemList
	for cursor.Next(ctx) {
		var problem models.Problem
		if err := cursor.Decode(&problem); err != nil {
			return nil, 0, err
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

// UpdateProblem 更新题目
func (s *ProblemService) UpdateProblem(ctx context.Context, problem *models.Problem) error {
	problem.UpdatedAt = time.Now()
	collection := utils.GetCollection("problems")
	_, err := collection.ReplaceOne(ctx, bson.M{"_id": problem.ID}, problem)
	return err
}

// DeleteProblem 删除题目
func (s *ProblemService) DeleteProblem(ctx context.Context, id primitive.ObjectID) error {
	collection := utils.GetCollection("problems")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// UpdateProblemStats 更新题目统计信息
func (s *ProblemService) UpdateProblemStats(ctx context.Context, problemID primitive.ObjectID, isAC bool) error {
	collection := utils.GetCollection("problems")

	update := bson.M{
		"$inc": bson.M{
			"submit_count": 1,
		},
	}

	if isAC {
		update["$inc"].(bson.M)["ac_count"] = 1
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": problemID}, update)
	return err
}

// RunCodeRequest 代码运行请求
type RunCodeRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
	Input    string `json:"input"` // 自定义输入（可选）
}

// RunCodeResponse 代码运行响应
type RunCodeResponse struct {
	Success    bool   `json:"success"`
	Status     string `json:"status"` // accepted, runtime_error, time_limit_exceeded, etc.
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	TimeUsed   int    `json:"time_used"`   // ms
	MemoryUsed int    `json:"memory_used"` // KB

	// 如果使用示例用例
	TestResults []models.TestResult `json:"test_results,omitempty"`
}

// RunCode 运行代码测试（非提交评测）
func (s *ProblemService) RunCode(ctx context.Context, problemID primitive.ObjectID, req *RunCodeRequest) (*RunCodeResponse, error) {
	logger := utils.GetLogger(ctx)

	// 1. 获取题目信息
	problem, err := s.GetProblemByID(ctx, problemID)
	if err != nil {
		return nil, fmt.Errorf("获取题目信息失败: %w", err)
	}

	// 2. 验证语言是否支持
	langConfig := s.sandboxClient.GetLanguageConfig(req.Language)
	if langConfig == nil {
		return nil, fmt.Errorf("不支持的语言: %s", req.Language)
	}

	// 3. 编译（如果需要）
	var executableID string
	if langConfig.Compile != nil {
		logger.Info(fmt.Sprintf("开始编译: Language=%s", req.Language))

		compileReq := &sandbox.CompileRequest{
			Language:   req.Language,
			SourceCode: req.Code,
		}

		compileResp, err := s.sandboxClient.CompileCode(ctx, compileReq)
		if err != nil {
			return nil, fmt.Errorf("编译失败: %w", err)
		}

		if !compileResp.Success {
			return &RunCodeResponse{
				Success: false,
				Status:  "compile_error",
				Error:   compileResp.CompileError,
			}, nil
		}

		executableID = compileResp.ExecutableID
		logger.Info(fmt.Sprintf("编译成功: ExecutableID=%s", executableID))
	} else {
		// 解释型语言，使用源代码
		executableID = req.Code
	}

	// 4. 如果提供了自定义输入，使用自定义输入运行
	if req.Input != "" {
		return s.runWithCustomInput(ctx, executableID, req, problem)
	}

	// 5. 否则使用示例测试用例运行
	return s.runWithSampleTestCases(ctx, executableID, req, problem, problemID)
}

// runWithCustomInput 使用自定义输入运行
func (s *ProblemService) runWithCustomInput(
	ctx context.Context,
	executableID string,
	req *RunCodeRequest,
	problem *models.Problem,
) (*RunCodeResponse, error) {
	runReq := &sandbox.RunRequest{
		Language:     req.Language,
		ExecutableID: executableID,
		Input:        req.Input,
		TimeLimit:    int64(problem.TimeLimit) * 1_000_000,   // ms -> ns
		MemoryLimit:  int64(problem.MemoryLimit) * 1_048_576, // MB -> bytes
	}

	runResp, err := s.sandboxClient.RunCode(ctx, runReq)
	if err != nil {
		return nil, fmt.Errorf("运行失败: %w", err)
	}

	resp := &RunCodeResponse{
		Success:    runResp.Status == sandbox.RunStatusAccepted,
		Status:     string(runResp.Status),
		Output:     runResp.Output,
		Error:      runResp.Error,
		TimeUsed:   int(runResp.Time / 1_000_000), // ns -> ms
		MemoryUsed: int(runResp.Memory / 1024),    // bytes -> KB
	}

	return resp, nil
}

// runWithSampleTestCases 使用示例测试用例运行
func (s *ProblemService) runWithSampleTestCases(
	ctx context.Context,
	executableID string,
	req *RunCodeRequest,
	problem *models.Problem,
	problemID primitive.ObjectID,
) (*RunCodeResponse, error) {
	// 获取示例测试用例
	testCases, err := s.testCaseService.GetSampleTestCases(ctx, problemID)
	if err != nil {
		return nil, fmt.Errorf("获取示例测试用例失败: %w", err)
	}

	if len(testCases) == 0 {
		return nil, fmt.Errorf("该题目没有示例测试用例")
	}

	// 运行所有示例测试用例
	var testResults []models.TestResult
	allPassed := true

	for _, testCase := range testCases {
		runReq := &sandbox.RunRequest{
			Language:     req.Language,
			ExecutableID: executableID,
			Input:        testCase.Input,
			TimeLimit:    int64(problem.TimeLimit) * 1_000_000,   // ms -> ns
			MemoryLimit:  int64(problem.MemoryLimit) * 1_048_576, // MB -> bytes
		}

		runResp, err := s.sandboxClient.RunCode(ctx, runReq)
		if err != nil {
			return nil, fmt.Errorf("运行测试用例失败: %w", err)
		}

		// 构建测试结果
		result := models.TestResult{
			TestCaseID: testCase.ID,
			TimeUsed:   int(runResp.Time / 1_000_000), // ns -> ms
			MemoryUsed: int(runResp.Memory / 1024),    // bytes -> KB
			IsSample:   true,
			Output:     runResp.Output,
			Expected:   testCase.Output,
			Error:      runResp.Error,
		}

		// 判断运行状态
		switch runResp.Status {
		case sandbox.RunStatusAccepted:
			if compareOutput(runResp.Output, testCase.Output) {
				result.Status = models.StatusAccepted
			} else {
				result.Status = models.StatusWrongAnswer
				allPassed = false
			}
		case sandbox.RunStatusTimeLimitExceeded:
			result.Status = models.StatusTimeLimit
			allPassed = false
		case sandbox.RunStatusMemoryLimitExceeded:
			result.Status = models.StatusMemoryLimit
			allPassed = false
		case sandbox.RunStatusRuntimeError:
			result.Status = models.StatusRuntimeError
			allPassed = false
		default:
			result.Status = models.StatusSystemError
			allPassed = false
		}

		testResults = append(testResults, result)
	}

	// 构建响应
	status := "accepted"
	if !allPassed {
		// 使用第一个失败的测试用例的状态
		for _, result := range testResults {
			if result.Status != models.StatusAccepted {
				status = string(result.Status)
				break
			}
		}
	}

	resp := &RunCodeResponse{
		Success:     allPassed,
		Status:      status,
		TestResults: testResults,
	}

	// 计算总时间和内存
	for _, result := range testResults {
		if result.TimeUsed > resp.TimeUsed {
			resp.TimeUsed = result.TimeUsed
		}
		if result.MemoryUsed > resp.MemoryUsed {
			resp.MemoryUsed = result.MemoryUsed
		}
	}

	return resp, nil
}

// compareOutput 比对输出
func compareOutput(actual, expected string) bool {
	// 去除首尾空白字符
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)

	// 逐行比对，忽略行尾空格
	actualLines := strings.Split(actual, "\n")
	expectedLines := strings.Split(expected, "\n")

	if len(actualLines) != len(expectedLines) {
		return false
	}

	for i := range actualLines {
		if strings.TrimRight(actualLines[i], " \t\r") != strings.TrimRight(expectedLines[i], " \t\r") {
			return false
		}
	}

	return true
}

// SearchProblems 搜索题目
func (s *ProblemService) SearchProblems(
	ctx context.Context,
	keyword string,
	difficulty models.ProblemDifficulty,
	tags []string,
	page, pageSize int,
) ([]*models.ProblemList, int64, error) {
	collection := utils.GetCollection("problems")

	// 构建查询条件
	filter := bson.M{
		"is_public": true,
		"status":    models.StatusPublished,
	}

	// 关键词搜索（标题和描述）
	if keyword != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": keyword, "$options": "i"}},       // 标题模糊匹配，不区分大小写
			{"description": bson.M{"$regex": keyword, "$options": "i"}}, // 描述模糊匹配
		}
	}

	// 难度筛选
	if difficulty != "" {
		filter["difficulty"] = difficulty
	}

	// 标签筛选
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.M{"created_at": -1}) // 按创建时间倒序

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var problems []*models.ProblemList
	for cursor.Next(ctx) {
		var problem models.Problem
		if err := cursor.Decode(&problem); err != nil {
			return nil, 0, err
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
