package router

import (
	"backend/controller"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

// InitFileRouter 初始化文件相关路由
func InitFileRouter(Router *gin.RouterGroup) {
	fileRouter := Router.Group("files").Use(middleware.JWTAuth())
	fileController := controller.FileController{}
	{
		fileRouter.POST("upload", fileController.UploadFile)
		fileRouter.GET(":id", fileController.GetFileInfo)
	}
}
