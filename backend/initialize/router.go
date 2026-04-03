package initialize

import (
	"backend/controller"
	"backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.Default()
	Router.Use(corsMiddleware())

	userController := &controller.UserController{}

	// 无需认证的路由
	v1 := Router.Group("/v1")
	{
		v1.POST("/user/send-email", userController.SendRegisterEmail)
		v1.POST("/user/register", userController.Register)
		v1.POST("/user/login", userController.Login)
	}

	// 需要 JWT 认证的路由
	authV1 := Router.Group("/v1")
	authV1.Use(middleware.JWTAuth())
	{
		authV1.GET("/user/info", userController.GetUserInfo)
	}

	return Router
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173": {},
		"http://127.0.0.1:5173": {},
		"http://localhost:8080": {},
		"http://127.0.0.1:8080": {},
		"http://localhost:3000": {},
		"http://127.0.0.1:3000": {},
		"http://localhost:4173": {},
		"http://127.0.0.1:4173": {},
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length,Content-Type")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
