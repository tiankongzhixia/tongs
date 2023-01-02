package initialize

import (
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"tongs/global"
	"tongs/utils"
)

func InitLogger() {
	isExist := fileutil.IsExist(global.CONFIG.Server.Zap.Director) // 判断是否有Director文件夹
	if !isExist {
		fmt.Printf("创建 %v 文件夹\n", global.CONFIG.Server.Zap.Director)
		_ = os.Mkdir(global.CONFIG.Server.Zap.Director, os.ModePerm)
	}

	cores := utils.Zap.GetZapCores()
	logger := zap.New(zapcore.NewTee(cores...))

	if global.CONFIG.Server.Zap.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	global.Log = logger
}
