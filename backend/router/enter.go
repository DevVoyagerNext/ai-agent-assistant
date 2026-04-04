package router

import (
	"backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	// 注册子路由
	userRouter := &UserRouter{}
	
	// v1 路由组
	v1 := r.Group("/v1")
	userRouter.InitUserRouter(v1)

	// auth v1 路由组
	authV1 := r.Group("/v1")
	authV1.Use(middleware.JWTAuth())
	userRouter.InitAuthUserRouter(authV1)

	return r
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
