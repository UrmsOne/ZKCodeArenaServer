/*
@Author: omenkk7
@Date: 2025/10/05
@Description: 判题服务 - 业务逻辑编排
*/

package service

import (
	"context"
	"fmt"
	"strings"

	"zk-code-arena-server/conf"
	"zk-code-arena-server/pkg/models"
	"zk-code-arena-server/pkg/queue"
	"zk-code-arena-server/pkg/sandbox"
	"zk-code-arena-server/pkg/utils"
)

// JudgeService 判题服务
type JudgeService struct {
	sandboxClient   sandbox.Client
	submitService   *SubmitService
	testCaseService *TestCaseService
	problemService  *ProblemService
	messageQueue    queue.MessageQueue
}

// JudgeContext 判题上下文
type JudgeContext struct {
	Task         *queue.JudgeTask
	Problem      *models.Problem
	TestCases    []*models.TestCase
	ExecutableID string
	Results      []models.TestResult
}

// NewJudgeService 创建判题服务实例
func NewJudgeService(
	sandboxClient sandbox.Client,
	submitService *SubmitService,
	testCaseService *TestCaseService,
	problemService *ProblemService,
	messageQueue queue.MessageQueue,
) *JudgeService {
	return &JudgeService{
		sandboxClient:   sandboxClient,
		submitService:   submitService,
		testCaseService: testCaseService,
		problemService:  problemService,
		messageQueue:    messageQueue,
	}
}

// Start 启动判题服务（启动消费者）
func (js *JudgeService) Start(ctx context.Context) error {
	logger := utils.GetLogger(ctx)
	logger.Info("启动判题服务...")

	// 1. 启动消息队列
	if err := js.messageQueue.Start(ctx); err != nil {
		return fmt.Errorf("启动消息队列失败: %w", err)
	}

	// 2. 恢复未完成的任务
	if err := js.RecoverTasks(ctx); err != nil {
		logger.Error(fmt.Sprintf("恢复任务失败: %v", err))
		// 不阻止服务启动
	}

	// 3. 启动消费者
	workers := conf.Config.Judge.Workers
	if workers <= 0 {
		workers = 4 // 默认值
	}
	if err := js.messageQueue.Consume(ctx, js.judgeTaskHandler, workers); err != nil {
		return fmt.Errorf("启动消费者失败: %w", err)
	}

	logger.Info("判题服务启动成功")
	return nil
}

// Stop 停止判题服务（优雅关闭）
func (js *JudgeService) Stop(ctx context.Context) error {
	logger := utils.GetLogger(ctx)
	logger.Info("正在停止判题服务...")

	if err := js.messageQueue.Stop(ctx); err != nil {
		return fmt.Errorf("停止消息队列失败: %w", err)
	}

	logger.Info("判题服务已停止")
	return nil
}

// SubmitTask 提交判题任务（生产者接口）
func (js *JudgeService) SubmitTask(ctx context.Context, task *queue.JudgeTask) error {
	return js.messageQueue.Push(ctx, task)
}

// RecoverTasks 恢复未完成的任务
func (js *JudgeService) RecoverTasks(ctx context.Context) error {
	logger := utils.GetLogger(ctx)
	logger.Info("开始恢复未完成的任务...")

	// 1. 将 "running" 状态的提交标记为 "system_error"
	if err := js.submitService.BatchUpdateStatus(ctx, models.StatusRunning, models.StatusSystemError, "服务重启，任务中断"); err != nil {
		logger.Error(fmt.Sprintf("批量更新 running 状态失败: %v", err))
	} else {
		logger.Info("已将 running 状态的提交标记为 system_error")
	}

	// 2. 获取待处理任务并重新入队
	pendingTasks, err := js.messageQueue.GetPendingTasks(ctx)
	if err != nil {
		return fmt.Errorf("获取待处理任务失败: %w", err)
	}

	logger.Info(fmt.Sprintf("找到 %d 个待处理任务，准备重新入队", len(pendingTasks)))

	for _, task := range pendingTasks {
		if err := js.messageQueue.Push(ctx, task); err != nil {
			logger.Error(fmt.Sprintf("重新入队任务失败: SubmitID=%s, Error=%v", task.SubmitID.Hex(), err))
		}
	}

	logger.Info(fmt.Sprintf("任务恢复完成，共恢复 %d 个任务", len(pendingTasks)))
	return nil
}

// judgeTaskHandler 判题任务处理函数（传递给消息队列的 handler）
func (js *JudgeService) judgeTaskHandler(ctx context.Context, task *queue.JudgeTask) error {
	logger := utils.GetLogger(ctx)
	logger.Info(fmt.Sprintf("开始处理判题任务: SubmitID=%s", task.SubmitID.Hex()))

	if err := js.judgeTask(ctx, task); err != nil {
		logger.Error(fmt.Sprintf("判题失败: SubmitID=%s, Error=%v", task.SubmitID.Hex(), err))
		return err
	}

	logger.Info(fmt.Sprintf("判题完成: SubmitID=%s", task.SubmitID.Hex()))
	return nil
}

// judgeTask 判题主流程
func (js *JudgeService) judgeTask(ctx context.Context, task *queue.JudgeTask) error {
	// 1. 更新状态为 running
	if err := js.submitService.UpdateSubmitStatus(ctx, task.SubmitID, models.StatusRunning); err != nil {
		return fmt.Errorf("更新状态为 running 失败: %w", err)
	}

	// 2. 获取题目信息
	problem, err := js.problemService.GetProblemByID(ctx, task.ProblemID)
	if err != nil {
		js.submitService.UpdateSubmitStatus(ctx, task.SubmitID, models.StatusSystemError)
		return fmt.Errorf("获取题目信息失败: %w", err)
	}

	// 3. 获取测试用例
	testCases, err := js.testCaseService.GetTestCasesByProblemID(ctx, task.ProblemID)
	if err != nil {
		js.submitService.UpdateSubmitStatus(ctx, task.SubmitID, models.StatusSystemError)
		return fmt.Errorf("获取测试用例失败: %w", err)
	}

	if len(testCases) == 0 {
		js.submitService.UpdateSubmitStatus(ctx, task.SubmitID, models.StatusSystemError)
		return fmt.Errorf("题目没有测试用例")
	}

	// 4. 初始化判题上下文
	judgeCtx := &JudgeContext{
		Task:      task,
		Problem:   problem,
		TestCases: testCases,
		Results:   make([]models.TestResult, 0, len(testCases)),
	}

	// 5. 编译阶段
	if err := js.compileStage(ctx, judgeCtx); err != nil {
		// 编译错误
		result := &models.JudgeResult{
			Status:       models.StatusCompileError,
			TimeUsed:     0,
			MemoryUsed:   0,
			TestResults:  []models.TestResult{},
			CompileError: err.Error(),
		}
		return js.submitService.UpdateSubmitResult(ctx, task.SubmitID, result)
	}

	// 6. 运行测试用例
	if err := js.runTestCases(ctx, judgeCtx); err != nil {
		js.submitService.UpdateSubmitStatus(ctx, task.SubmitID, models.StatusSystemError)
		return fmt.Errorf("运行测试用例失败: %w", err)
	}

	// 7. 确定最终状态
	finalStatus := js.determineFinalStatus(judgeCtx.Results)
	maxTime := js.calculateMaxTime(judgeCtx.Results)
	maxMemory := js.calculateMaxMemory(judgeCtx.Results)

	// 8. 更新判题结果
	result := &models.JudgeResult{
		Status:      finalStatus,
		TimeUsed:    maxTime,
		MemoryUsed:  maxMemory,
		TestResults: judgeCtx.Results,
	}

	return js.submitService.UpdateSubmitResult(ctx, task.SubmitID, result)
}

// compileStage 编译阶段
func (js *JudgeService) compileStage(ctx context.Context, judgeCtx *JudgeContext) error {
	logger := utils.GetLogger(ctx)
	
	// 检查语言是否需要编译
	langConfig := js.sandboxClient.GetLanguageConfig(judgeCtx.Task.Language)
	if langConfig == nil {
		return fmt.Errorf("不支持的语言: %s", judgeCtx.Task.Language)
	}

	if langConfig.Compile == nil {
		// 解释型语言，无需编译
		logger.Info(fmt.Sprintf("语言 %s 无需编译", judgeCtx.Task.Language))
		return nil
	}

	logger.Info(fmt.Sprintf("开始编译: Language=%s", judgeCtx.Task.Language))

	compileReq := &sandbox.CompileRequest{
		Language:   judgeCtx.Task.Language,
		SourceCode: judgeCtx.Task.Code,
	}

	compileResp, err := js.sandboxClient.CompileCode(ctx, compileReq)
	if err != nil {
		return fmt.Errorf("沙箱编译失败: %w", err)
	}

	if !compileResp.Success {
		return fmt.Errorf("编译失败: %s", compileResp.CompileError)
	}

	judgeCtx.ExecutableID = compileResp.ExecutableID
	logger.Info(fmt.Sprintf("编译成功: ExecutableID=%s", compileResp.ExecutableID))
	return nil
}

// runTestCases 运行测试用例
func (js *JudgeService) runTestCases(ctx context.Context, judgeCtx *JudgeContext) error {
	logger := utils.GetLogger(ctx)
	logger.Info(fmt.Sprintf("开始运行测试用例，共 %d 个", len(judgeCtx.TestCases)))

	for i, testCase := range judgeCtx.TestCases {
		logger.Info(fmt.Sprintf("运行测试用例 %d/%d: TestCaseID=%s", i+1, len(judgeCtx.TestCases), testCase.ID.Hex()))

		result, err := js.runSingleTestCase(ctx, judgeCtx, testCase)
		if err != nil {
			return fmt.Errorf("运行测试用例失败: %w", err)
		}

		judgeCtx.Results = append(judgeCtx.Results, *result)
	}

	return nil
}

// runSingleTestCase 运行单个测试用例
func (js *JudgeService) runSingleTestCase(
	ctx context.Context,
	judgeCtx *JudgeContext,
	testCase *models.TestCase,
) (*models.TestResult, error) {
	// 获取时间和内存限制
	timeLimit := js.getTimeLimit(testCase, judgeCtx.Problem)
	memoryLimit := js.getMemoryLimit(testCase, judgeCtx.Problem)

	// 构建运行请求
	// 对于编译型语言，ExecutableID 是编译产物的文件ID
	// 对于解释型语言，ExecutableID 是源代码本身
	executableID := judgeCtx.ExecutableID
	if executableID == "" {
		// 解释型语言，使用源代码
		executableID = judgeCtx.Task.Code
	}
	
	runReq := &sandbox.RunRequest{
		Language:     judgeCtx.Task.Language,
		ExecutableID: executableID,
		Input:        testCase.Input,
		TimeLimit:    int64(timeLimit) * 1_000_000,   // ms -> ns
		MemoryLimit:  int64(memoryLimit) * 1_048_576, // MB -> bytes
	}

	// 调用沙箱运行
	runResp, err := js.sandboxClient.RunCode(ctx, runReq)
	if err != nil {
		return nil, fmt.Errorf("沙箱运行失败: %w", err)
	}

	// 构建测试结果
	result := &models.TestResult{
		TestCaseID: testCase.ID,
		TimeUsed:   int(runResp.Time / 1_000_000), // ns -> ms
		MemoryUsed: int(runResp.Memory / 1024),    // bytes -> KB
		IsSample:   testCase.IsSample,
	}

	// 判断运行状态
	switch runResp.Status {
	case sandbox.RunStatusAccepted:
		// 比对输出
		if js.compareOutput(runResp.Output, testCase.Output) {
			result.Status = models.StatusAccepted
		} else {
			result.Status = models.StatusWrongAnswer
		}
	case sandbox.RunStatusTimeLimitExceeded:
		result.Status = models.StatusTimeLimit
	case sandbox.RunStatusMemoryLimitExceeded:
		result.Status = models.StatusMemoryLimit
	case sandbox.RunStatusRuntimeError:
		result.Status = models.StatusRuntimeError
	default:
		result.Status = models.StatusSystemError
	}

	// 仅示例用例返回详细信息
	if testCase.IsSample {
		result.Output = runResp.Output
		result.Expected = testCase.Output
		result.Error = runResp.Error
	}

	return result, nil
}

// compareOutput 比对输出
func (js *JudgeService) compareOutput(actual, expected string) bool {
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

// determineFinalStatus 确定最终状态
func (js *JudgeService) determineFinalStatus(testResults []models.TestResult) models.SubmitStatus {
	if len(testResults) == 0 {
		return models.StatusSystemError
	}

	hasWrongAnswer := false
	hasTimeLimit := false
	hasMemoryLimit := false
	hasRuntimeError := false
	hasSystemError := false
	allAccepted := true

	for _, result := range testResults {
		switch result.Status {
		case models.StatusAccepted:
			// 继续
		case models.StatusWrongAnswer:
			hasWrongAnswer = true
			allAccepted = false
		case models.StatusTimeLimit:
			hasTimeLimit = true
			allAccepted = false
		case models.StatusMemoryLimit:
			hasMemoryLimit = true
			allAccepted = false
		case models.StatusRuntimeError:
			hasRuntimeError = true
			allAccepted = false
		case models.StatusSystemError:
			hasSystemError = true
			allAccepted = false
		default:
			allAccepted = false
		}
	}

	// 优先级：运行时错误 > 时间超限 > 内存超限 > 答案错误 > 系统错误 > 通过
	if hasRuntimeError {
		return models.StatusRuntimeError
	}
	if hasTimeLimit {
		return models.StatusTimeLimit
	}
	if hasMemoryLimit {
		return models.StatusMemoryLimit
	}
	if hasWrongAnswer {
		return models.StatusWrongAnswer
	}
	if hasSystemError {
		return models.StatusSystemError
	}
	if allAccepted {
		return models.StatusAccepted
	}

	return models.StatusSystemError
}

// getTimeLimit 获取时间限制（ms）
func (js *JudgeService) getTimeLimit(testCase *models.TestCase, problem *models.Problem) int {
	// 优先使用测试用例的配置
	if testCase.TimeLimit != nil && *testCase.TimeLimit > 0 {
		return *testCase.TimeLimit
	}
	// 使用题目默认配置
	return problem.TimeLimit
}

// getMemoryLimit 获取内存限制（MB）
func (js *JudgeService) getMemoryLimit(testCase *models.TestCase, problem *models.Problem) int {
	// 优先使用测试用例的配置
	if testCase.MemoryLimit != nil && *testCase.MemoryLimit > 0 {
		return *testCase.MemoryLimit
	}
	// 使用题目默认配置
	return problem.MemoryLimit
}

// calculateMaxTime 计算最大时间使用
func (js *JudgeService) calculateMaxTime(testResults []models.TestResult) int {
	maxTime := 0
	for _, result := range testResults {
		if result.TimeUsed > maxTime {
			maxTime = result.TimeUsed
		}
	}
	return maxTime
}

// calculateMaxMemory 计算最大内存使用
func (js *JudgeService) calculateMaxMemory(testResults []models.TestResult) int {
	maxMemory := 0
	for _, result := range testResults {
		if result.MemoryUsed > maxMemory {
			maxMemory = result.MemoryUsed
		}
	}
	return maxMemory
}