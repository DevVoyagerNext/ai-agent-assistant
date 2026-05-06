package controller

import (
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	fileService service.FileService
}

// UploadFile 上传文件
func (con *FileController) UploadFile(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("userId")
	if !exists {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "未获取到上传文件", c)
		return
	}

	src, err := file.Open()
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "读取文件失败", c)
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "读取文件内容失败", c)
		return
	}

	res, err := con.fileService.UploadFile(c.Request.Context(), fileBytes, file.Filename, userID.(uint))
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(res, c)
}

// GetFileInfo 读取文件信息
func (con *FileController) GetFileInfo(c *gin.Context) {
	idStr := c.Param("id")
	fileID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || fileID == 0 {
		response.FailWithMsg(errmsg.CodeError, "文件ID不合法", c)
		return
	}

	res, err := con.fileService.GetFileInfo(c.Request.Context(), uint(fileID))
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(res, c)
}
