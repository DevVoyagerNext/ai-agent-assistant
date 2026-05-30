package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Data: data,
		Msg:  "操作成功",
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: nil,
		Msg:  msg,
	})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: 401,
		Data: nil,
		Msg:  msg,
	})
}
