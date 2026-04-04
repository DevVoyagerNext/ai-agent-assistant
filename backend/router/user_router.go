package router

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

// InitUserRouter 初始化用户路由(无需认证)
func (r *UserRouter) InitUserRouter(Router *gin.RouterGroup) {
	userController := &controller.UserController{}
	userRouter := Router.Group("/user")
	{
		userRouter.POST("/send-email", userController.SendRegisterEmail)
		userRouter.POST("/register", userController.Register)
		userRouter.POST("/login", userController.Login)
	}
}

// InitAuthUserRouter 初始化用户路由(需要认证)
func (r *UserRouter) InitAuthUserRouter(Router *gin.RouterGroup) {
	userController := &controller.UserController{}
	userRouter := Router.Group("/user")
	{
		userRouter.GET("/info", userController.GetUserInfo)
	}
}
