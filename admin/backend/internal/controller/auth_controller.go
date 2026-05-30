package controller

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/httpx"
	adminsession "admin/backend/internal/session"

	"github.com/gin-gonic/gin"
)

func (ctl *AdminController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, 400, "请求参数错误")
		return
	}

	res, err := ctl.adminService.Register(c.Request.Context(), req, adminsession.DeviceFromRequest(c))
	if err != nil {
		httpx.Fail(c, 400, err.Error())
		return
	}

	httpx.OK(c, res)
}

func (ctl *AdminController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, 400, "请求参数错误")
		return
	}

	res, err := ctl.adminService.Login(c.Request.Context(), req, adminsession.DeviceFromRequest(c))
	if err != nil {
		httpx.Fail(c, 400, err.Error())
		return
	}

	httpx.OK(c, res)
}
