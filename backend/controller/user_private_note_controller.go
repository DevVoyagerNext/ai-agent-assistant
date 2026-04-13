package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserPrivateNoteController struct {
	privateNoteService service.UserPrivateNoteService
	authService        service.AuthService
}

// GetPrivateNoteOrChildren 获取私人笔记内容或子文件夹列表
func (con *UserPrivateNoteController) GetPrivateNoteOrChildren(c *gin.Context) {
	noteIdStr := c.Param("noteId")
	noteId, err := strconv.Atoi(noteIdStr)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "笔记ID格式错误", c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	scopeStr := strings.TrimSpace(c.Query("scope"))
	scope := 2
	if scopeStr != "" {
		scope, err = strconv.Atoi(scopeStr)
		if err != nil {
			response.FailWithMsg(errmsg.CodeError, "scope 参数错误", c)
			return
		}
	}

	res, err := con.privateNoteService.GetNoteOrChildrenWithScope(c.Request.Context(), userId, noteId, scope)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(res, c)
}

// CreatePrivateNote 创建私人文件夹或笔记
func (con *UserPrivateNoteController) CreatePrivateNote(c *gin.Context) {
	var req dto.CreatePrivateNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	err = con.privateNoteService.CreatePrivateNote(c.Request.Context(), userId, req)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}

// DeletePrivateNote 删除私人文件/文件夹（文件夹递归删除）
func (con *UserPrivateNoteController) DeletePrivateNote(c *gin.Context) {
	noteIdStr := c.Param("noteId")
	noteId, err := strconv.Atoi(noteIdStr)
	if err != nil || noteId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "笔记ID格式错误", c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	if err := con.privateNoteService.DeletePrivateNote(c.Request.Context(), userId, noteId); err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}
