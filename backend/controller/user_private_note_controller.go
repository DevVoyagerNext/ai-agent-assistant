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

	tokenUserId, err := con.authService.GetUserID(c)
	if err != nil || tokenUserId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	targetUserId := tokenUserId
	targetUserIdStr := strings.TrimSpace(c.Query("userId"))
	if targetUserIdStr != "" {
		id, err := strconv.Atoi(targetUserIdStr)
		if err != nil || id <= 0 {
			response.FailWithMsg(errmsg.CodeError, "userId 参数错误", c)
			return
		}
		targetUserId = uint(id)
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	if scope == 0 || scope == 2 {
		if targetUserId != tokenUserId {
			response.FailWithCode(errmsg.UserPermissionDenied, c)
			return
		}
	}

	res, err := con.privateNoteService.GetNoteOrChildrenWithScope(c.Request.Context(), targetUserId, noteId, scope, page, pageSize)
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

// UpdatePrivateNoteContent 修改笔记内容
func (con *UserPrivateNoteController) UpdatePrivateNoteContent(c *gin.Context) {
	noteIdStr := c.Param("noteId")
	noteId, err := strconv.Atoi(noteIdStr)
	if err != nil || noteId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "笔记ID格式错误", c)
		return
	}

	var req dto.UpdatePrivateNoteContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	err = con.privateNoteService.UpdatePrivateNoteContent(c.Request.Context(), userId, noteId, req)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}

// UpdatePrivateNoteTitle 修改笔记/文件夹标题
func (con *UserPrivateNoteController) UpdatePrivateNoteTitle(c *gin.Context) {
	noteIdStr := c.Param("noteId")
	noteId, err := strconv.Atoi(noteIdStr)
	if err != nil || noteId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "笔记ID格式错误", c)
		return
	}

	var req dto.UpdatePrivateNoteTitleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	err = con.privateNoteService.UpdatePrivateNoteTitle(c.Request.Context(), userId, noteId, req)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}

// UpdatePrivateNotePublic 修改笔记/文件夹公开状态
func (con *UserPrivateNoteController) UpdatePrivateNotePublic(c *gin.Context) {
	noteIdStr := c.Param("noteId")
	noteId, err := strconv.Atoi(noteIdStr)
	if err != nil || noteId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "笔记ID格式错误", c)
		return
	}

	var req dto.UpdatePrivateNotePublicReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	err = con.privateNoteService.UpdatePrivateNotePublic(c.Request.Context(), userId, noteId, req)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}

// DeletePrivateNote 删除私人笔记或文件夹
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

	err = con.privateNoteService.DeletePrivateNote(c.Request.Context(), userId, noteId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}

// SharePrivateNote 分享私人笔记接口
func (con *UserPrivateNoteController) SharePrivateNote(c *gin.Context) {
	noteIdStr := c.Param("noteId")
	noteId, err := strconv.Atoi(noteIdStr)
	if err != nil || noteId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "笔记ID格式错误", c)
		return
	}

	var req dto.SharePrivateNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	res, err := con.privateNoteService.SharePrivateNote(c.Request.Context(), userId, noteId, req)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(res, c)
}
