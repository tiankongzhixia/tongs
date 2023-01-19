package tongs

import (
	"tongs/global"
	"tongs/initialize"
)

type M map[string]interface{}

func Run() {
	http := initialize.InitRouter()
	port := global.CONFIG.Server.Port
	if port == "" {
		port = "8080"
	}
	http.Run(":" + port)
}

func Init() {
	initialize.InitConfig()
	initialize.InitLogger()
	initialize.InitDatabase()
	initialize.InitRedis()
	initialize.InitTongs()
}
