package initialize

import (
	"backend/controller"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.Default()

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
