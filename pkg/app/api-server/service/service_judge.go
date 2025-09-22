/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service_judge.go
@Description: 评测服务
*/

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"zk-code-arena-server/conf"
	"zk-code-arena-server/pkg/models"
)

type JudgeService struct{}

func NewJudgeService() *JudgeService {
	return &JudgeService{}
}

// JudgeRequest 评测请求
type JudgeRequest struct {
	Code        string `json:"code"`
	Language    string `json:"language"`
	Input       string `json:"input"`
	TimeLimit   int    `json:"time_limit"`
	MemoryLimit int    `json:"memory_limit"`
}

// JudgeResponse 评测响应
type JudgeResponse struct {
	Status     string `json:"status"`
	TimeUsed   int    `json:"time_used"`
	MemoryUsed int    `json:"memory_used"`
	Output     string `json:"output"`
	Error      string `json:"error"`
}

// JudgeCode 评测代码
func (s *JudgeService) JudgeCode(ctx context.Context, code, language, input string, timeLimit, memoryLimit int) (*JudgeResponse, error) {
	// 构建请求
	req := JudgeRequest{
		Code:        code,
		Language:    language,
		Input:       input,
		TimeLimit:   timeLimit,
		MemoryLimit: memoryLimit,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送HTTP请求到评测服务
	httpReq, err := http.NewRequestWithContext(ctx, conf.Config.Sandbox.Method, conf.Config.Sandbox.Url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: conf.Config.Sandbox.Timeout,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("评测服务返回错误: %s", string(respBody))
	}

	// 解析响应
	var judgeResp JudgeResponse
	err = json.Unmarshal(respBody, &judgeResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &judgeResp, nil
}

// JudgeSubmit 评测提交
func (s *JudgeService) JudgeSubmit(ctx context.Context, submit *models.Submit, problem *models.Problem) (*models.JudgeResult, error) {
	// 获取测试用例
	testCases, err := s.getTestCases(ctx, problem.ID)
	if err != nil {
		return nil, fmt.Errorf("获取测试用例失败: %v", err)
	}

	result := &models.JudgeResult{
		Status:      models.StatusRunning,
		TestResults: make([]models.TestResult, 0, len(testCases)),
	}

	// 评测每个测试用例
	for _, testCase := range testCases {
		testResult := models.TestResult{
			TestCaseID: testCase.ID,
		}

		// 调用评测服务
		judgeResp, err := s.JudgeCode(ctx, submit.Code, string(submit.Language), testCase.Input, problem.TimeLimit, problem.MemoryLimit)
		if err != nil {
			testResult.Status = models.StatusSystemError
			testResult.Error = err.Error()
		} else {
			testResult.Status = s.mapJudgeStatus(judgeResp.Status)
			testResult.TimeUsed = judgeResp.TimeUsed
			testResult.MemoryUsed = judgeResp.MemoryUsed
			testResult.Output = judgeResp.Output
			testResult.Error = judgeResp.Error
		}

		result.TestResults = append(result.TestResults, testResult)

		// 如果编译错误，直接返回
		if testResult.Status == models.StatusCompileError {
			result.Status = models.StatusCompileError
			result.CompileError = testResult.Error
			break
		}

		// 如果运行时错误，记录并继续
		if testResult.Status == models.StatusRuntimeError {
			result.Status = models.StatusRuntimeError
			result.RuntimeError = testResult.Error
		}
	}

	// 确定最终状态
	if result.Status == models.StatusRunning {
		result.Status = s.determineFinalStatus(result.TestResults)
	}

	// 计算总的时间和内存使用
	result.TimeUsed = s.calculateMaxTime(result.TestResults)
	result.MemoryUsed = s.calculateMaxMemory(result.TestResults)

	return result, nil
}

// getTestCases 获取测试用例
func (s *JudgeService) getTestCases(ctx context.Context, problemID primitive.ObjectID) ([]models.TestCase, error) {
	// TODO: 实现获取测试用例的逻辑
	// 这里应该从数据库获取测试用例
	return []models.TestCase{}, nil
}

// mapJudgeStatus 映射评测状态
func (s *JudgeService) mapJudgeStatus(status string) models.SubmitStatus {
	switch status {
	case "accepted":
		return models.StatusAccepted
	case "wrong_answer":
		return models.StatusWrongAnswer
	case "time_limit":
		return models.StatusTimeLimit
	case "memory_limit":
		return models.StatusMemoryLimit
	case "runtime_error":
		return models.StatusRuntimeError
	case "compile_error":
		return models.StatusCompileError
	default:
		return models.StatusSystemError
	}
}

// determineFinalStatus 确定最终状态
func (s *JudgeService) determineFinalStatus(testResults []models.TestResult) models.SubmitStatus {
	hasAccepted := false
	hasWrongAnswer := false
	hasTimeLimit := false
	hasMemoryLimit := false
	hasRuntimeError := false

	for _, result := range testResults {
		switch result.Status {
		case models.StatusAccepted:
			hasAccepted = true
		case models.StatusWrongAnswer:
			hasWrongAnswer = true
		case models.StatusTimeLimit:
			hasTimeLimit = true
		case models.StatusMemoryLimit:
			hasMemoryLimit = true
		case models.StatusRuntimeError:
			hasRuntimeError = true
		}
	}

	// 优先级：编译错误 > 运行时错误 > 时间超限 > 内存超限 > 答案错误 > 通过
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
	if hasAccepted {
		return models.StatusAccepted
	}

	return models.StatusSystemError
}

// calculateMaxTime 计算最大时间使用
func (s *JudgeService) calculateMaxTime(testResults []models.TestResult) int {
	maxTime := 0
	for _, result := range testResults {
		if result.TimeUsed > maxTime {
			maxTime = result.TimeUsed
		}
	}
	return maxTime
}

// calculateMaxMemory 计算最大内存使用
func (s *JudgeService) calculateMaxMemory(testResults []models.TestResult) int {
	maxMemory := 0
	for _, result := range testResults {
		if result.MemoryUsed > maxMemory {
			maxMemory = result.MemoryUsed
		}
	}
	return maxMemory
}
