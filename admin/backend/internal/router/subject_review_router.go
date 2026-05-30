package router

import (
	"admin/backend/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterSubjectReviewRouter(group *gin.RouterGroup, adminController *controller.AdminController) {
	group.GET("/subjects/pending", adminController.PendingSubjects)
	group.POST("/subjects/:id/approve", adminController.ApproveSubject)
	group.POST("/subjects/:id/reject", adminController.RejectSubject)
}
