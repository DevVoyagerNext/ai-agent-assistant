package router

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

type UserPrivateNoteRouter struct{}

func (s *UserPrivateNoteRouter) InitAuthUserPrivateNoteRouter(Router *gin.RouterGroup) {
	privateNoteController := &controller.UserPrivateNoteController{}
	privateNoteRouter := Router.Group("/user/notes/private")
	{
		// 1. 获取私人笔记内容或子文件夹列表 (如: GET /v1/user/notes/private/12)
		privateNoteRouter.GET("/:noteId", privateNoteController.GetPrivateNoteOrChildren)
		// 2. 创建私人文件夹或笔记 (如: POST /v1/user/notes/private)
		privateNoteRouter.POST("", privateNoteController.CreatePrivateNote)
		privateNoteRouter.POST("/", privateNoteController.CreatePrivateNote)
	}
}
