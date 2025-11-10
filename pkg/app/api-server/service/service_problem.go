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

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/app/api-server/repository"
	"zk-code-arena-server/pkg/sandbox"
	"zk-code-arena-server/pkg/utils"
)

type ProblemService struct {
	sandboxClient   sandbox.Client
	testCaseService *TestCaseService
	repo           *repository.ProblemRepository
}

func NewProblemService(sandboxClient sandbox.Client, testCaseService *TestCaseService, repo *repository.ProblemRepository) *ProblemService {
	return &ProblemService{
		sandboxClient:   sandboxClient,
		testCaseService: testCaseService,
		repo:           repo,
	}
}

// CreateProblem 创建题目
func (s *ProblemService) CreateProblem(ctx context.Context, problem *models.Problem) error {
	// 1. 设置系统字段
	problem.ID = primitive.NewObjectID()
	problem.CreatedAt = time.Now()
	problem.UpdatedAt = time.Now()

	// 2. 设置默认Status（如果未传）
	if problem.Status == "" {
		problem.Status = models.StatusDraft
		utils.Logger.Infof("CreateProblem: 未指定状态，设置默认状态为草稿")
	}

	// 3. 应用Status与IsPublic关联规则
	if problem.Status == models.StatusDraft {
		// 草稿状态强制私有
		if problem.IsPublic {
			utils.Logger.Warnf("CreateProblem: 草稿状态不能公开，强制设为私有")
		}
		problem.IsPublic = false
	}
	// Published/Archived状态，保持用户设置或默认私有
	// （IsPublic由Handler层传入，这里不修改）

	// 4. 设置默认计数器
	problem.ACCount = 0
	problem.SubmitCount = 0

	// 5. 设置默认限制（如果为0）
	if problem.TimeLimit == 0 {
		problem.TimeLimit = 1000
		utils.Logger.Debugf("CreateProblem: 使用默认时间限制 1000ms")
	}
	if problem.MemoryLimit == 0 {
		problem.MemoryLimit = 256
		utils.Logger.Debugf("CreateProblem: 使用默认内存限制 256MB")
	}

	// 6. 确保Tags不为nil
	if problem.Tags == nil {
		problem.Tags = []string{}
	}

	// 7. 记录详细日志
	utils.Logger.Infof("CreateProblem: title=%s, difficulty=%s, status=%s, isPublic=%v, createdBy=%s",
		problem.Title, problem.Difficulty, problem.Status, problem.IsPublic, problem.CreatedBy.Hex())

	// 8. 持久化到数据库 - 委托给Repository层
	err := s.repo.CreateProblem(ctx, problem)
	if err != nil {
		utils.Logger.Errorf("CreateProblem: 数据库插入失败, error=%v", err)
		return fmt.Errorf("数据库操作失败: %w", err)
	}

	utils.Logger.Infof("CreateProblem: 题目创建成功, id=%s", problem.ID.Hex())
	return nil
}
// GetProblemByID 根据ID获取题目
func (s *ProblemService) GetProblemByID(ctx context.Context, id primitive.ObjectID) (*models.Problem, error) {
	// 数据操作：直接委托给Repository层
	return s.repo.GetByID(ctx, id)
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
	// 业务逻辑：记录日志信息
	if !includePrivate {
		utils.Logger.Debugf("GetProblems: 公开模式，过滤条件: is_public=true, status=published")
	} else {
		switch role {
		case models.RoleAdmin:
			utils.Logger.Debugf("GetProblems: 管理员模式，查看所有题目")
		case models.RoleTeacher:
			if userID != nil {
				utils.Logger.Debugf("GetProblems: 教师模式，查看自己创建的题目, userID=%s", userID.Hex())
			} else {
				utils.Logger.Debugf("GetProblems: 教师模式但无userID，退回公开模式")
			}
		default:
			utils.Logger.Debugf("GetProblems: 学生/游客模式，只看公开已发布题目")
		}
	}

	// 数据操作：委托给Repository层
	return s.repo.GetProblemsWithUserStatus(ctx, page, pageSize, difficulty, tags, includePrivate, role, userID)
}

// getProblemsByCondition 核心查询方法（私有，用于内部复用）
func (s *ProblemService) getProblemsByCondition(
	ctx context.Context,
	condition *models.ProblemQueryCondition,
) ([]*models.ProblemList, int64, error) {
	// 统一的核心查询逻辑
	return s.repo.GetProblemsWithUserStatus(
		ctx,
		condition.Page,
		condition.PageSize,
		condition.Difficulty,
		condition.Tags,
		condition.IncludePrivate,
		condition.Role,
		condition.UserID,
	)
}

// GetProblemsForUser 获取用户端题目列表
func (s *ProblemService) GetProblemsForUser(
	ctx context.Context,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 构建用户端查询条件
	condition := &models.ProblemQueryCondition{
		Page:           page,
		PageSize:       pageSize,
		Difficulty:     difficulty,
		Tags:           tags,
		IncludePrivate: false,                // 用户端固定为false，只看公开题目
		Role:           models.RoleStudent,   // 固定为学生角色
		UserID:         userID,               // 用于获取用户提交状态
	}
	
	// 业务逻辑：记录用户端查询日志
	if userID != nil {
		utils.Logger.Debugf("GetProblemsForUser: 用户端查询, userID=%s, difficulty=%s, tags=%v", 
			userID.Hex(), difficulty, tags)
	} else {
		utils.Logger.Debugf("GetProblemsForUser: 未登录用户查询, difficulty=%s, tags=%v", 
			difficulty, tags)
	}
	
	return s.getProblemsByCondition(ctx, condition)
}

// GetProblemsForAdmin 获取管理员端题目列表
func (s *ProblemService) GetProblemsForAdmin(
	ctx context.Context,
	page, pageSize int,
	difficulty models.ProblemDifficulty,
	tags []string,
) ([]*models.ProblemList, int64, error) {
	// 构建管理员端查询条件
	condition := &models.ProblemQueryCondition{
		Page:           page,
		PageSize:       pageSize,
		Difficulty:     difficulty,
		Tags:           tags,
		IncludePrivate: true,                // 管理员可以看所有题目，包括私有
		Role:           models.RoleAdmin,    // 管理员角色
		UserID:         nil,                 // 管理员不需要获取用户提交状态
		// 注意：status和createdBy筛选将在后续版本中支持
	}
	
	// 业务逻辑：记录管理员查询日志
	utils.Logger.Debugf("GetProblemsForAdmin: 管理员查询, difficulty=%s, tags=%v", 
		difficulty, tags)
	utils.Logger.Infof("GetProblemsForAdmin: 管理员可查看所有题目（包括私有和草稿状态）")
	
	return s.getProblemsByCondition(ctx, condition)
}




func (s *ProblemService) UpdateProblem(ctx context.Context, problemID primitive.ObjectID, req *models.UpdateProblemRequest) error {
	// 业务逻辑：记录更新操作日志和处理特殊业务规则
	if req.Status != nil && *req.Status == models.StatusDraft {
		utils.Logger.Warnf("UpdateProblem: 草稿状态不能公开，强制设为私有, problemID=%s", problemID.Hex())
	}
	
	utils.Logger.Infof("UpdateProblem: 开始更新题目, problemID=%s", problemID.Hex())

	// 数据操作：委托给Repository层处理字段映射和更新
	err := s.repo.UpdateProblemFromRequest(ctx, problemID, req)
	if err != nil {
		utils.Logger.Errorf("UpdateProblem: 数据库更新失败, problemID=%s, error=%v", problemID.Hex(), err)
		return fmt.Errorf("数据库操作失败: %w", err)
	}

	utils.Logger.Infof("UpdateProblem: 题目更新成功, problemID=%s", problemID.Hex())
	return nil
}


// DeleteProblem 删除题目
func (s *ProblemService) DeleteProblem(ctx context.Context, id primitive.ObjectID) error {
	// 数据操作：直接委托给Repository层
	return s.repo.DeleteProblem(ctx, id)
}

// UpdateProblemStats 更新题目统计信息
func (s *ProblemService) UpdateProblemStats(ctx context.Context, problemID primitive.ObjectID, isAC bool) error {
	// 数据操作：直接委托给Repository层
	return s.repo.UpdateProblemStats(ctx, problemID, isAC)
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
	userID *primitive.ObjectID,
) ([]*models.ProblemList, int64, error) {
	// 业务逻辑：记录搜索相关日志
	if userID != nil {
		utils.Logger.Debugf("SearchProblems: 开始搜索, keyword=%s, userID=%s", keyword, userID.Hex())
	} else {
		utils.Logger.Debugf("SearchProblems: 未登录用户搜索, keyword=%s", keyword)
	}

	// 数据操作：委托给Repository层
	// 业务逻辑：根据用户登录状态设置默认参数
	includePrivate := userID != nil // 登录用户可以看到私有题目
	role := models.RoleStudent      // 默认为学生角色
	
	return s.repo.SearchProblems(ctx, keyword, page, pageSize, difficulty, tags, includePrivate, role, userID)
}

// GetUserProblemStatuses 批量查询用户对多个题目的状态
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//   - problemIDs: 题目ID列表
// 返回:
//   - map[string]models.UserProblemStatus: key为题目ID字符串，value为用户状态
//   - error: 错误信息
func (s *ProblemService) GetUserProblemStatuses(
	ctx context.Context,
	userID primitive.ObjectID,
	problemIDs []primitive.ObjectID,
) (map[string]models.UserProblemStatus, error) {
	// 业务逻辑：记录查询日志
	utils.Logger.Debugf("GetUserProblemStatuses: 查询用户状态, userID=%s, problemCount=%d", 
		userID.Hex(), len(problemIDs))
	
	// 数据操作：委托给Repository层
	statusMap, err := s.repo.GetUserProblemStatuses(ctx, userID, problemIDs)
	if err != nil {
		utils.Logger.Errorf("GetUserProblemStatuses: Repository查询失败, userID=%s, error=%v", 
			userID.Hex(), err)
		return nil, err
	}
	
	// 业务逻辑：转换map键类型从ObjectID到string
	result := make(map[string]models.UserProblemStatus, len(statusMap))
	for problemID, status := range statusMap {
		result[problemID.Hex()] = status
	}
	
	utils.Logger.Infof("GetUserProblemStatuses: 查询完成, userID=%s, 返回%d个题目状态", 
		userID.Hex(), len(result))
	
	return result, nil
}

// GetProblemDetail 获取题目详情聚合信息
func (s *ProblemService) GetProblemDetail(ctx context.Context, problemID primitive.ObjectID) (*models.ProblemDetailResponse, error) {
	// 业务逻辑：记录查询日志
	utils.Logger.Infof("GetProblemDetail: 开始获取题目详情聚合信息, problemID=%s", problemID.Hex())

	// 数据操作1：获取题目基本信息
	problem, err := s.repo.GetByID(ctx, problemID)
	if err != nil {
		utils.Logger.Errorf("GetProblemDetail: 获取题目信息失败, problemID=%s, error=%v", problemID.Hex(), err)
		return nil, fmt.Errorf("获取题目信息失败: %w", err)
	}

	// 数据操作2：获取示例测试用例
	sampleCases, err := s.testCaseService.GetSampleTestCases(ctx, problemID)
	if err != nil {
		utils.Logger.Errorf("GetProblemDetail: 获取示例测试用例失败, problemID=%s, error=%v", problemID.Hex(), err)
		return nil, fmt.Errorf("获取示例测试用例失败: %w", err)
	}

	// 业务逻辑：数据聚合和组装
	response := &models.ProblemDetailResponse{
		Problem:     problem,
		SampleCases: make([]models.TestCase, len(sampleCases)),
	}

	// 转换指针切片为值切片
	for i, testCase := range sampleCases {
		response.SampleCases[i] = *testCase
	}

	utils.Logger.Infof("GetProblemDetail: 题目详情聚合完成, problemID=%s, sampleCasesCount=%d", 
		problemID.Hex(), len(sampleCases))

	return response, nil
}

// BatchCreateProblems 批量创建题目
func (s *ProblemService) BatchCreateProblems(ctx context.Context, req *models.BatchCreateProblemsRequest) (*models.BatchCreateProblemsResponse, error) {
	// 业务逻辑：记录批量创建日志
	utils.Logger.Infof("BatchCreateProblems: 开始批量创建题目, count=%d", len(req.Problems))
	
	// 初始化响应
	response := &models.BatchCreateProblemsResponse{
		SuccessCount: 0,
		FailCount:     0,
		TotalCount:    len(req.Problems),
		Results:       make([]models.BatchCreateProblemResult, 0, len(req.Problems)),
	}
	
	// 遍历创建每个题目
	for i, problemReq := range req.Problems {
		// 创建Problem模型
		problem := &models.Problem{
			Title:        problemReq.Title,
			Description:  problemReq.Description,
			Input:        problemReq.Input,
			Output:       problemReq.Output,
			SampleInput:  problemReq.SampleInput,
			SampleOutput: problemReq.SampleOutput,
			Hint:         problemReq.Hint,
			Source:       problemReq.Source,
			Author:       problemReq.Author,
			Difficulty:   problemReq.Difficulty,
			Tags:         problemReq.Tags,
			CreatedBy:    problemReq.CreatedBy,
		}
		
		// 设置可选字段
		if problemReq.TimeLimit != nil {
			problem.TimeLimit = *problemReq.TimeLimit
		}
		if problemReq.MemoryLimit != nil {
			problem.MemoryLimit = *problemReq.MemoryLimit
		}
		if problemReq.Status != nil {
			problem.Status = *problemReq.Status
		}
		if problemReq.IsPublic != nil {
			problem.IsPublic = *problemReq.IsPublic
		}
		
		// 调用单个创建方法
		err := s.CreateProblem(ctx, problem)
		
		// 构建结果
		result := models.BatchCreateProblemResult{
			Index:     i,
			Title:     problemReq.Title,
			Success:   err == nil,
			ProblemID: problem.ID.Hex(),
		}
		
		if err != nil {
			result.Error = err.Error()
			response.FailCount++
			utils.Logger.Errorf("BatchCreateProblems: 题目创建失败, index=%d, title=%s, error=%v", 
				i, problemReq.Title, err)
		} else {
			response.SuccessCount++
			utils.Logger.Infof("BatchCreateProblems: 题目创建成功, index=%d, title=%s, id=%s", 
				i, problemReq.Title, problem.ID.Hex())
		}
		
		response.Results = append(response.Results, result)
	}
	
	utils.Logger.Infof("BatchCreateProblems: 批量创建完成, total=%d, success=%d, fail=%d", 
		response.TotalCount, response.SuccessCount, response.FailCount)
	
	return response, nil
}
