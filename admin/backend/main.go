package main

import (
	"log"

	"admin/backend/internal/config"
	"admin/backend/internal/controller"
	"admin/backend/internal/dao"
	"admin/backend/internal/database"
	"admin/backend/internal/httpx"
	"admin/backend/internal/middleware"
	"admin/backend/internal/router"
	"admin/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	redisClient, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	adminDao := dao.NewAdminDao(db)
	adminService := service.NewAdminService(adminDao, redisClient, cfg)
	if err := adminService.EnsureRootAdmin(); err != nil {
		log.Fatalf("ensure root admin: %v", err)
	}
	adminController := controller.NewAdminController(adminService)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.Cors())
	router.RegisterRouter(r, redisClient, cfg, adminController)

	r.NoRoute(func(c *gin.Context) {
		httpx.Fail(c, 404, "接口不存在")
	})

	addr := ":" + cfg.Server.Port
	log.Printf("admin backend listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
