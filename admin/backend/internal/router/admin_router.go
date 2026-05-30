package router

import (
	"admin/backend/internal/config"
	"admin/backend/internal/controller"
	"admin/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterAdminRouter(group *gin.RouterGroup, redisClient *redis.Client, cfg config.Config, adminController *controller.AdminController) {
	admin := group.Group("/admin")
	admin.Use(middleware.AdminAuth(redisClient, cfg))

	RegisterProfileRouter(admin, adminController)
	RegisterSubjectReviewRouter(admin, adminController)
	RegisterNodeReviewRouter(admin, adminController)
	RegisterAuditRecordRouter(admin, adminController)
}
