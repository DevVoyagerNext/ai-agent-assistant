package router

import (
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.ZapRecovery(),
		middleware.RequestID(),
		middleware.ZapLogger(),
		middleware.CorsMiddleware(),
	)

	// 注册子路由
	userRouter := &UserRouter{}
	subjectRouter := &SubjectRouter{}
	knowledgeNodeRouter := &KnowledgeNodeRouter{}

	// v1 路由组
	v1 := r.Group("/v1")
	userRouter.InitUserRouter(v1)
	subjectRouter.InitSubjectRouter(v1)
	knowledgeNodeRouter.InitKnowledgeNodeRouter(v1)

	// auth v1 路由组
	authV1 := r.Group("/v1")
	authV1.Use(middleware.JWTAuth())
	userRouter.InitAuthUserRouter(authV1)
	subjectRouter.InitAuthSubjectRouter(authV1)
	knowledgeNodeRouter.InitAuthKnowledgeNodeRouter(authV1)

	return r
}
