/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: config.go
@Description: 配置管理模块
*/

package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

var (
	AppMode      string        // 服务器启动模式，默认debug模式
	AppEnv       string        // 服务器部署环境，默认为开发环境模式
	Port         string        // 服务启动端口
	GracefulTime time.Duration // 服务优雅重启的时间，单位为秒，默认5s
	Config       *configYaml
	v            *viper.Viper
)

type configYaml struct {
	App struct {
		Mode         string        `yaml:"Mode"`
		Host         string        `yaml:"Host"`
		Port         string        `yaml:"Port"`
		Env          string        `yaml:"Env"`
		GracefulTime time.Duration `yaml:"GracefulTime"`
	} `yaml:"App"`
	Mongo struct {
		Uri    string `yaml:"Uri"`
		DbName string `yaml:"DbName"`
	} `yaml:"Mongo"`
	Judge JudgeConfig `yaml:"Judge"`
	Sandbox struct {
		Url     string        `yaml:"Url"`
		Method  string        `yaml:"Method"`
		Timeout time.Duration `yaml:"Timeout"`
	} `yaml:"Sandbox"`
	Log struct {
		Level  string `yaml:"Level"`
		Format string `yaml:"Format"`
		Output string `yaml:"Output"`
		File   struct {
			Path       string `yaml:"Path"`
			MaxSize    int    `yaml:"MaxSize"`
			MaxBackups int    `yaml:"MaxBackups"`
			MaxAge     int    `yaml:"MaxAge"`
			Compress   bool   `yaml:"Compress"`
		} `yaml:"File"`
	} `yaml:"Log"`
	JWT struct {
		Secret     string        `yaml:"Secret"`
		ExpireTime time.Duration `yaml:"ExpireTime"`
	} `yaml:"JWT"`
	Redis struct {
		Host     string `yaml:"Host"`
		Port     int    `yaml:"Port"`
		Password string `yaml:"Password"`
		DB       int    `yaml:"DB"`
		PoolSize int    `yaml:"PoolSize"`
	} `yaml:"Redis"`
}

// JudgeConfig 判题配置
type JudgeConfig struct {
	Workers   int                           `yaml:"Workers"`   // 消费者数量
	QueueSize int                           `yaml:"QueueSize"` // 队列容量
	Languages map[string]*LanguageConfig    `yaml:"Languages"` // 语言配置
}

// LanguageConfig 语言配置
type LanguageConfig struct {
	Name    string       `yaml:"name"`              // 语言名称
	Compile *StageConfig `yaml:"compile,omitempty"` // 编译配置（可选，解释型语言无需编译）
	Run     *StageConfig `yaml:"run"`               // 运行配置
}

// StageConfig 阶段配置（编译或运行）
type StageConfig struct {
	Args           []string `yaml:"args"`                       // 执行命令参数
	Env            []string `yaml:"env"`                        // 环境变量
	TimeLimit      int64    `yaml:"time_limit"`                 // 时间限制（纳秒）
	MemoryLimit    int64    `yaml:"memory_limit"`               // 内存限制（字节）
	ProcLimit      int      `yaml:"proc_limit"`                 // 进程数限制
	SourceFile     string   `yaml:"source_file,omitempty"`      // 源文件名
	ExecutableFile string   `yaml:"executable_file,omitempty"`  // 可执行文件名
}

func Init() {
	// viper mapstructure 配置读取
	v = viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("../../../conf")
	v.AddConfigPath("../../conf")
	v.AddConfigPath("../conf")
	v.AddConfigPath("./conf")
	
	// 设置环境变量替换器
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	v.AutomaticEnv()
	
	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("配置文件读取失败: %s", err))
	}
	
	// 解析配置到结构体
	c := &configYaml{}
	err = v.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("配置文件解析失败: %s", err))
	}
	
	Config = c
	
	// 开发环境打印配置信息
	if Config.App.Env == "development" {
		fmt.Printf("配置加载成功: %+v\n", c)
	}
}

func ConfigUtils() *viper.Viper {
	return v
}
