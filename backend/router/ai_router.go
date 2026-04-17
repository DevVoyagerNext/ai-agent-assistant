package router

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

type AIRouter struct{}

// InitAuthAIRouter 初始化需要认证的 AI 路由
func (r *AIRouter) InitAuthAIRouter(Router *gin.RouterGroup) {
	aiController := &controller.AIController{}
	aiRouter := Router.Group("/ai")
	{
		aiRouter.POST("/chat", aiController.Chat) // AI 聊天接口
	}
}
