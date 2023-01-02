package model

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Result struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func Success(data interface{}) Result {
	return Result{
		Code: 0,
		Data: data,
		Msg:  "成功",
	}
}

func Fail(code int, msg string) Result {
	return Result{
		Code: code,
		Data: nil,
		Msg:  msg,
	}
}

// Ok 返回操作成功信息
func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, Success(nil))
	c.Abort()
}

// OkWithData 返回成功信息并且带返回参数
func OkWithData(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Success(data))
	c.Abort()
}

// Error 通用错误
func Error(code int, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Fail(code, msg))
	c.Abort()
}
