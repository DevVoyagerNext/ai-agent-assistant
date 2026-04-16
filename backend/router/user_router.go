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
		userRouter.POST("/refresh-token", userController.RefreshToken)
	}
}

// InitAuthUserRouter 初始化用户路由(需要认证)
func (r *UserRouter) InitAuthUserRouter(Router *gin.RouterGroup) {
	userController := &controller.UserController{}
	userRouter := Router.Group("/user")
	{
		userRouter.GET("/info", userController.GetUserInfo)                                // 核心聚合接口 (GetProfileSummary)
		userRouter.GET("/activities/calendar", userController.GetUserActivitiesCalendar)   // 异步接口：活跃度日历
		userRouter.GET("/notes/private/public-list", userController.GetPublicPrivateNotes) // 异步接口：公开的私人笔记
		userRouter.GET("/notes/shares", userController.GetSharedNotes)                     // 异步接口：已分享笔记
		userRouter.PUT("/notes/shares/:id/cancel", userController.CancelSharedNote)        // 取消分享笔记
		userRouter.GET("/subjects/learned", userController.GetLearnedSubjects)             // 异步接口：已学/在学教材
	}
}
