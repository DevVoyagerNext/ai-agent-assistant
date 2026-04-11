package router

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

type SubjectRouter struct{}

// InitSubjectRouter 初始化教材路由(无需认证)
func (r *SubjectRouter) InitSubjectRouter(Router *gin.RouterGroup) {
	subjectController := &controller.SubjectController{}
	subjectRouter := Router.Group("/subjects")
	{
		subjectRouter.GET("", subjectController.GetAllSubjects)                     // 获取所有的教材
		subjectRouter.GET("/search", subjectController.SearchSubjects)              // 通过教材名称模糊搜索教材
		subjectRouter.GET("/categories", subjectController.GetCategories)           // 获取教材分类
		subjectRouter.GET("/category/:id", subjectController.GetSubjectsByCategory) // 通过分类获取教材
	}
}

// InitAuthSubjectRouter 初始化教材路由(需要认证)
func (r *SubjectRouter) InitAuthSubjectRouter(Router *gin.RouterGroup) {
	subjectController := &controller.SubjectController{}
	userSubjectRouter := Router.Group("/user/subjects")
	{
		userSubjectRouter.GET("/folders", subjectController.GetUserCollectFolders)                      // 获取用户收藏夹
		userSubjectRouter.GET("/folders/:folderId", subjectController.GetUserCollectedSubjectsByFolder) // 获取用户收藏夹下的教材
		userSubjectRouter.GET("/collected", subjectController.GetUserCollectedSubjects)                 // 获取用户收藏的教材
		userSubjectRouter.GET("/liked", subjectController.GetUserLikedSubjects)                         // 获取用户点赞的教材
		userSubjectRouter.GET("/learning", subjectController.GetUserLearningSubjects)                   // 获取用户正在学习的教材
		userSubjectRouter.GET("/completed", subjectController.GetUserCompletedSubjects)                 // 获取用户已经学习完成的教材
		userSubjectRouter.GET("/last-learning", subjectController.GetUserLastLearningSubject)           // 分页获取最近学习的教材
		userSubjectRouter.POST("/:id/like", subjectController.ToggleSubjectLike)                        // 点赞或取消点赞教材
	}
}
