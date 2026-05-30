package router

import (
	"admin/backend/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterAuditRecordRouter(group *gin.RouterGroup, adminController *controller.AdminController) {
	group.GET("/audit-records", adminController.AuditRecords)
}
