package controller

import (
	"admin/backend/internal/httpx"

	"github.com/gin-gonic/gin"
)

func (ctl *AdminController) AuditRecords(c *gin.Context) {
	res, err := ctl.adminService.AuditRecords()
	if err != nil {
		httpx.Fail(c, 500, err.Error())
		return
	}

	httpx.OK(c, res)
}
