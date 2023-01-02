package api

import (
	"github.com/gin-gonic/gin"
	"strings"
	"tongs/global"
	"tongs/model"
)

func GetTongs(c *gin.Context) {
	model.OkWithData(global.TongsManager.GetTongsName(), c)
}

func RunTongs(c *gin.Context) {
	var param model.Param
	c.BindJSON(&param)
	t, err := global.TongsManager.FindTongs(param.Tongs)
	if err != nil {
		model.Error(-1, "Tongs不存在", c)
		return
	}
	err = t.Run(strings.Split(param.Url, ",")...)
	if err != nil {
		model.Error(-1, err.Error(), c)
		return
	}
	model.Ok(c)
}

func StopTongs(c *gin.Context) {
	var param model.Param
	c.BindJSON(&param)
	t, err := global.TongsManager.FindTongs(param.Tongs)
	if err != nil {
		model.Error(-1, "Tongs不存在", c)
		return
	}
	t.Stop()
	model.Ok(c)
}

func GetTasks(c *gin.Context) {
	tongs := c.Query("tongs")
	t, err := global.TongsManager.FindTongs(tongs)
	if err != nil {
		model.Error(-1, "Tongs不存在", c)
		return
	}

	model.OkWithData(t.Tasks, c)
}

func RunTask(c *gin.Context) {
	var param model.Param
	c.BindJSON(&param)

	t, err := global.TongsManager.FindTongs(param.Tongs)
	if err != nil {
		model.Error(-1, "Tongs不存在", c)
		return
	}
	err = t.RunTask(param.Task, param.Url)
	if err != nil {
		model.Error(-1, err.Error(), c)
		return
	}
	model.Ok(c)
}

func StopTask(c *gin.Context) {
	var param model.Param
	c.BindJSON(&param)

	t, err := global.TongsManager.FindTongs(param.Tongs)
	if err != nil {
		model.Error(-1, "Tongs不存在", c)
		return
	}
	t.StopTask(param.Task)
	model.Ok(c)
}

func AddUrl(c *gin.Context) {
	var param model.Param
	c.BindJSON(&param)
	t, err := global.TongsManager.FindTongs(param.Tongs)
	if err != nil {
		model.Error(-1, "Tongs不存在", c)
		return
	}
	err = t.AddURL(param.Task, param.Url)
	if err != nil {
		model.Error(-1, "Tongs不存在", c)
		return
	}
	model.Ok(c)
}
