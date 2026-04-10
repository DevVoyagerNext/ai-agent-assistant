package router

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

type KnowledgeNodeRouter struct{}

func (s *KnowledgeNodeRouter) InitKnowledgeNodeRouter(Router *gin.RouterGroup) {
	nodeController := &controller.KnowledgeNodeController{}
	nodeRouter := Router.Group("/nodes")
	{
		// 1. 获取该教材下的顶级知识点 (如: /v1/nodes/top?subjectId=1)
		nodeRouter.GET("/top", nodeController.GetTopLevelNodes)
		// 2. 获取某个知识点下的最近一级子知识点 (如: /v1/nodes/12/children)
		nodeRouter.GET("/:nodeId/children", nodeController.GetChildNodes)
		// 3. 获取某个知识点的信息（标题内容层级等详细信息） (如: /v1/nodes/12/detail)
		nodeRouter.GET("/:nodeId/detail", nodeController.GetNodeDetail)
	}
}

func (s *KnowledgeNodeRouter) InitAuthKnowledgeNodeRouter(Router *gin.RouterGroup) {
	nodeController := &controller.KnowledgeNodeController{}
	authNodeRouter := Router.Group("/nodes")
	{
		// 4. 获取某个知识点的随堂笔记 (如: /v1/nodes/12/note)
		authNodeRouter.GET("/:nodeId/note", nodeController.GetUserStudyNote)
		// 5. 更新某个知识点的学习状态 (如: PUT /v1/nodes/12/status)
		authNodeRouter.PUT("/:nodeId/status", nodeController.UpdateNodeStatus)
	}
}
