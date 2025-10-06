/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 09:54
@Name: runServer.go
@Description: 服务器启动命令
*/

package main

import (
	"context"
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"zk-code-arena-server/conf"
	"zk-code-arena-server/pkg/app/api-server/server"
	"zk-code-arena-server/pkg/app/api-server/service"
	"zk-code-arena-server/pkg/utils"
)

var runServerCfg = struct {
	serverHost string
	serverPort string
	logDir     string
	configDir  string
}{}

func init() {
	rootCmd.AddCommand(runServerCmd)
}

var runServerCmd = &cobra.Command{
	Use:   "run",
	Short: "启动 ZK Code Arena 服务器",
	Long:  "启动 ZK Code Arena 服务器，监听指定端口并提供在线刷题平台服务",
	RunE:  runServe,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func runServe(cmd *cobra.Command, args []string) error {
	if err := cmd.Flags().Parse(args); err != nil {
		return err
	}
	
	ctx := context.Background()
	
	// 初始化日志系统
	lg := utils.GetLogger(ctx)
	if lg == nil {
		return errors.New("failed to initialize logger")
	}
	
	lg.Info("Starting ZK Code Arena Server...")
	
	// 加载配置文件
	conf.Init()
	
	// 使用配置文件重新初始化日志器
	utils.InitLogger()
	
	conf.ConfigUtils().OnConfigChange(func(in fsnotify.Event) {
		lg.Infof("配置文件已更改: %s, 操作: %s", in.Name, in.Op)
		// TODO: 实现配置文件热加载和服务优雅重启
	})
	conf.ConfigUtils().WatchConfig()
	
	// 创建停止信号通道
	stopCh := make(chan struct{})
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	
	// 使用 errgroup 管理多个 goroutine
	g := errgroup.Group{}
	
	// 初始化数据库连接
	if err := utils.InitMongoDB(); err != nil {
		lg.Errorf("数据库连接失败: %v", err)
		return err
	}
	lg.Info("数据库连接成功")

	// 初始化服务
	svc := service.NewService()
	opts := server.NewCmdOptions(conf.Config.App.Host, conf.Config.App.Port)
	s := server.NewServer(lg, svc, opts, stopCh)

	// 启动判题服务（消费者）
	lg.Info("启动判题服务...")
	if err := svc.JudgeService.Start(ctx); err != nil {
		lg.Errorf("判题服务启动失败: %v", err)
		return err
	}
	lg.Info("判题服务启动成功")

	// 启动主服务器
	g.Go(func() error {
		s.Init()
		lg.Infof("HTTP 服务器启动在 %s:%s", conf.Config.App.Host, conf.Config.App.Port)
		return s.Run()
	})

	// 监听系统信号
	g.Go(func() error {
		return waitForSignal(lg, signalCh, stopCh)
	})

	defer s.Shutdown(ctx)
	return g.Wait()
}

func waitForSignal(lg logrus.FieldLogger, signalCh <-chan os.Signal, stopCh chan struct{}) error {
	for {
		select {
		case s := <-signalCh:
			lg.Infof("接收到信号 %s，正在关闭服务器...", s)
			close(stopCh)
			lg.Info("服务器正在停止...")
			return nil
		}
	}
}
