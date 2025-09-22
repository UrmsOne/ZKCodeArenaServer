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
	Judge struct {
		Type   string `yaml:"Type"`
		Java   struct {
			JDK11 struct {
				BaseArgs string `yaml:"BaseArgs"`
				Env      string `yaml:"Env"`
			} `yaml:"JDK11"`
		} `yaml:"Java"`
		Python struct {
			Python3 struct {
				BaseArgs string `yaml:"BaseArgs"`
				Env      string `yaml:"Env"`
			} `yaml:"Python3"`
		} `yaml:"Python"`
		Cpp struct {
			Gcc struct {
				BaseArgs string `yaml:"BaseArgs"`
				Env      string `yaml:"Env"`
			} `yaml:"Gcc"`
		} `yaml:"Cpp"`
	} `yaml:"Judge"`
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
