package initialize

import (
	"github.com/gin-gonic/gin"
	"tongs/api"
)

// InitRouter 初始化路由
func InitRouter() (e *gin.Engine) {
	http := gin.Default()

	http.GET("tongs", api.GetTongs)
	http.GET("tongs/detail", api.GetTongs)
	http.POST("tongs/run", api.RunTongs)
	http.POST("tongs/stop", api.StopTongs)

	http.GET("task", api.GetTasks)
	http.GET("task/detail", api.GetTasks)
	http.POST("task/run", api.RunTask)
	http.POST("task/stop", api.StopTask)
	http.POST("task/addUrl", api.AddUrl)
	return http
}
