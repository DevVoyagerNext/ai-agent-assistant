package course

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SubjectsRouter struct {}

// InitSubjectsRouter 初始化 教材审批 路由信息
func (s *SubjectsRouter) InitSubjectsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	subjectsRouter := Router.Group("subjects").Use(middleware.OperationRecord())
	subjectsRouterWithoutRecord := Router.Group("subjects")
	subjectsRouterWithoutAuth := PublicRouter.Group("subjects")
	{
		subjectsRouter.POST("createSubjects", subjectsApi.CreateSubjects)   // 新建教材审批
		subjectsRouter.DELETE("deleteSubjects", subjectsApi.DeleteSubjects) // 删除教材审批
		subjectsRouter.DELETE("deleteSubjectsByIds", subjectsApi.DeleteSubjectsByIds) // 批量删除教材审批
		subjectsRouter.PUT("updateSubjects", subjectsApi.UpdateSubjects)    // 更新教材审批
	}
	{
		subjectsRouterWithoutRecord.GET("findSubjects", subjectsApi.FindSubjects)        // 根据ID获取教材审批
		subjectsRouterWithoutRecord.GET("getSubjectsList", subjectsApi.GetSubjectsList)  // 获取教材审批列表
	}
	{
	    subjectsRouterWithoutAuth.GET("getSubjectsPublic", subjectsApi.GetSubjectsPublic)  // 教材审批开放接口
	}
}
