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
		// 3. 修改私人笔记内容 (仅限文件)
		privateNoteRouter.PUT("/:noteId/content", privateNoteController.UpdatePrivateNoteContent)
		// 4. 修改文件/文件夹标题
		privateNoteRouter.PUT("/:noteId/title", privateNoteController.UpdatePrivateNoteTitle)
		// 5. 修改公开状态
		privateNoteRouter.PUT("/:noteId/public", privateNoteController.UpdatePrivateNotePublic)
		// 6. 删除私人文件/文件夹 (如: DELETE /v1/user/notes/private/12)
		privateNoteRouter.DELETE("/:noteId", privateNoteController.DeletePrivateNote)
		// 7. 分享私人笔记 (如: POST /v1/user/notes/private/12/share)
		privateNoteRouter.POST("/:noteId/share", privateNoteController.SharePrivateNote)
	}
}
