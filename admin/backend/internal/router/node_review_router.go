package router

import (
	"admin/backend/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterNodeReviewRouter(group *gin.RouterGroup, adminController *controller.AdminController) {
	group.GET("/nodes/pending", adminController.PendingNodes)
	group.POST("/nodes/:id/approve", adminController.ApproveNode)
	group.POST("/nodes/:id/reject", adminController.RejectNode)
}
