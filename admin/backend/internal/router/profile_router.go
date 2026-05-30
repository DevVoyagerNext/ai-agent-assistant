package router

import (
	"admin/backend/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterProfileRouter(group *gin.RouterGroup, adminController *controller.AdminController) {
	group.GET("/me", adminController.Me)
}
