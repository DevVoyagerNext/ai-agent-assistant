package course

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SubjectsRouter struct {}

// InitSubjectsRouter 初始化 subjects表 路由信息
func (s *SubjectsRouter) InitSubjectsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	subjectsRouter := Router.Group("subjects").Use(middleware.OperationRecord())
	subjectsRouterWithoutRecord := Router.Group("subjects")
	subjectsRouterWithoutAuth := PublicRouter.Group("subjects")
	{
		subjectsRouter.POST("createSubjects", subjectsApi.CreateSubjects)   // 新建subjects表
		subjectsRouter.DELETE("deleteSubjects", subjectsApi.DeleteSubjects) // 删除subjects表
		subjectsRouter.DELETE("deleteSubjectsByIds", subjectsApi.DeleteSubjectsByIds) // 批量删除subjects表
		subjectsRouter.PUT("updateSubjects", subjectsApi.UpdateSubjects)    // 更新subjects表
	}
	{
		subjectsRouterWithoutRecord.GET("findSubjects", subjectsApi.FindSubjects)        // 根据ID获取subjects表
		subjectsRouterWithoutRecord.GET("getSubjectsList", subjectsApi.GetSubjectsList)  // 获取subjects表列表
	}
	{
	    subjectsRouterWithoutAuth.GET("getSubjectsPublic", subjectsApi.GetSubjectsPublic)  // subjects表开放接口
	}
}
