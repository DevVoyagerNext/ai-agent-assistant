package router

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

type AIRouter struct{}

// InitAIRouter 初始化无需认证的 AI 路由
func (r *AIRouter) InitAIRouter(Router *gin.RouterGroup) {
	aiController := &controller.AIController{}
	aiRouter := Router.Group("/ai")
	{
		aiRouter.GET("/exports/tickets/:ticket", aiController.DownloadExportByTicket)
	}
}

// InitAuthAIRouter 初始化需要认证的 AI 路由
func (r *AIRouter) InitAuthAIRouter(Router *gin.RouterGroup) {
	aiController := &controller.AIController{}
	aiRouter := Router.Group("/ai")
	{
		aiRouter.POST("/chat", aiController.Chat)                               // AI 聊天接口
		aiRouter.GET("/sessions", aiController.GetUserSessions)                 // 获取用户的历史会话列表
		aiRouter.PUT("/sessions/:id/title", aiController.UpdateSessionTitle)    // 修改会话标题
		aiRouter.GET("/sessions/:id/messages", aiController.GetSessionMessages) // 获取具体会话的消息列表
	}
}
