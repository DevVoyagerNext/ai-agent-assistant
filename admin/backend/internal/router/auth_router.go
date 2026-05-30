package router

import (
	"admin/backend/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRouter(group *gin.RouterGroup, adminController *controller.AdminController) {
	group.POST("/admin/register", adminController.Register)
	group.POST("/admin/login", adminController.Login)
}
