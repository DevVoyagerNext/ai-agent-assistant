package controller

import (
	"admin/backend/internal/httpx"

	"github.com/gin-gonic/gin"
)

func (ctl *AdminController) Me(c *gin.Context) {
	res, err := ctl.adminService.Me(currentAdminID(c))
	if err != nil {
		httpx.Fail(c, 401, err.Error())
		return
	}

	httpx.OK(c, res)
}
