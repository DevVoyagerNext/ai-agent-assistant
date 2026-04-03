package response

import (
	"backend/pkg/errmsg"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func Ok(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: errmsg.CodeSuccess,
		Data: data,
		Msg:  errmsg.GetMsg(errmsg.CodeSuccess),
	})
}

func FailWithCode(code int, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: nil,
		Msg:  errmsg.GetMsg(code),
	})
}

func FailWithMsg(code int, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: nil,
		Msg:  msg,
	})
}
