package utils

import (
	"flag"
	"fmt"
)

var ConfigPathDefault = "."
var ConfigNameDefault = "config"

func GetConfigPath() string {
	var configPath string
	flag.StringVar(&configPath, "cp", "", "选择配置文件路径.")
	flag.Parse()
	if configPath != "" {
		fmt.Printf("您正在使用命令行的-cp参数传递的值,配置文件路径为%s\n", configPath)
		return configPath
	}
	return ConfigPathDefault
}

func GetEnv() string {
	var env string
	flag.StringVar(&env, "e", "", "选择环境.")
	flag.Parse()
	if env != "" {
		fmt.Printf("您正在使用命令行的-e参数传递的值,环境为%s\n", env)
		return env
	}
	return ""
}
