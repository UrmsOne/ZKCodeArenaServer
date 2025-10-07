/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service.go
@Description: 服务层入口
*/

package service

import (
	"context"
	"fmt"

	"zk-code-arena-server/conf"
	"zk-code-arena-server/pkg/queue"
	"zk-code-arena-server/pkg/sandbox"
	"zk-code-arena-server/pkg/utils"
)

// Service 服务层结构
type Service struct {
	UserService       *UserService
	ProblemService    *ProblemService
	SubmitService     *SubmitService
	JudgeService      *JudgeService
	TestCaseService   *TestCaseService
	CourseService     *CourseService
	StatisticsService *StatisticsService
}

// NewService 创建服务实例
func NewService() *Service {
	// 1. 创建基础服务（无依赖）
	userService := NewUserService()
	submitService := NewSubmitService()
	testCaseService := NewTestCaseService()
	courseService := NewCourseService()

	// 2. 创建沙箱客户端
	// 将 conf.LanguageConfig 转换为 sandbox.LanguageConfig
	sandboxLanguages := make(map[string]*sandbox.LanguageConfig)
	for lang, confLang := range conf.Config.Judge.Languages {
		sandboxLang := &sandbox.LanguageConfig{
			Name: confLang.Name,
		}
		if confLang.Compile != nil {
			// 调试：打印配置
			fmt.Printf("=== 加载语言配置: %s ===\nSourceFile: '%s'\nExecutableFile: '%s'\n",
				lang, confLang.Compile.SourceFile, confLang.Compile.ExecutableFile)

			sandboxLang.Compile = &sandbox.StageConfig{
				Args:           confLang.Compile.Args,
				Env:            confLang.Compile.Env,
				TimeLimit:      confLang.Compile.TimeLimit,
				MemoryLimit:    confLang.Compile.MemoryLimit,
				ProcLimit:      confLang.Compile.ProcLimit,
				SourceFile:     confLang.Compile.SourceFile,
				ExecutableFile: confLang.Compile.ExecutableFile,
			}
		}
		if confLang.Run != nil {
			sandboxLang.Run = &sandbox.StageConfig{
				Args:           confLang.Run.Args,
				Env:            confLang.Run.Env,
				TimeLimit:      confLang.Run.TimeLimit,
				MemoryLimit:    confLang.Run.MemoryLimit,
				ProcLimit:      confLang.Run.ProcLimit,
				SourceFile:     confLang.Run.SourceFile,
				ExecutableFile: confLang.Run.ExecutableFile,
			}
		}
		sandboxLanguages[lang] = sandboxLang
	}

	sandboxClient := sandbox.NewGoJudgeClient(
		conf.Config.Judge.SandboxURL,
		sandboxLanguages,
	)

	// 3. 创建消息队列
	queueSize := conf.Config.Judge.QueueSize
	if queueSize <= 0 {
		queueSize = 100
	}
	messageQueue := queue.NewChannelQueue(queueSize, submitService)

	// 4. 创建需要依赖注入的服务
	// ProblemService 需要 sandboxClient 和 testCaseService
	problemService := NewProblemService(sandboxClient, testCaseService)
	
	// JudgeService 需要所有基础服务的引用
	judgeService := NewJudgeService(
		sandboxClient,
		submitService,
		testCaseService,
		problemService,
		messageQueue,
	)
	
	// StatisticsService 需要其他服务的引用
	statisticsService := NewStatisticsService(submitService, problemService, userService)

	return &Service{
		UserService:       userService,
		ProblemService:    problemService,
		SubmitService:     submitService,
		JudgeService:      judgeService,
		TestCaseService:   testCaseService,
		CourseService:     courseService,
		StatisticsService: statisticsService,
	}
}

// Close 关闭服务
func (s *Service) Close(ctx context.Context) error {
	// 关闭数据库连接
	return utils.CloseMongoDB(ctx)
}
