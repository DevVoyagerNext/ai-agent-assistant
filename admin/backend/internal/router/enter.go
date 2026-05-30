package router

import (
	"admin/backend/internal/config"
	"admin/backend/internal/controller"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterRouter(r *gin.Engine, redisClient *redis.Client, cfg config.Config, adminController *controller.AdminController) {
	v1 := r.Group("/v1")
	RegisterAuthRouter(v1, adminController)
	RegisterAdminRouter(v1, redisClient, cfg, adminController)
}
