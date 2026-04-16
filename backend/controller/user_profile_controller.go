package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserActivitiesCalendar 获取用户活跃度日历接口
func (u *UserController) GetUserActivitiesCalendar(c *gin.Context) {
	userId, err := u.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenInvalid, c)
		return
	}

	errCode, res := u.userService.GetUserActivitiesCalendar(c.Request.Context(), userId)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(res, c)
}

// GetPublicPrivateNotes 获取用户公开的私人笔记列表接口
func (u *UserController) GetPublicPrivateNotes(c *gin.Context) {
	userId, err := u.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenInvalid, c)
		return
	}

	var req dto.PublicPrivateNoteListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	errCode, res := u.userService.GetPublicPrivateNotes(c.Request.Context(), userId, req)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(res, c)
}

// GetLearnedSubjects 获取已学/在学教材列表接口
func (u *UserController) GetLearnedSubjects(c *gin.Context) {
	userId, err := u.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenInvalid, c)
		return
	}

	errCode, res := u.userService.GetLearnedSubjects(c.Request.Context(), userId)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(res, c)
}

// GetSharedNotes 获取已分享笔记列表接口
func (u *UserController) GetSharedNotes(c *gin.Context) {
	userId, err := u.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenInvalid, c)
		return
	}

	var req dto.SharedNoteListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	errCode, res := u.userService.GetSharedNotes(c.Request.Context(), userId, req)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(res, c)
}

// CancelSharedNote 取消分享接口
func (u *UserController) CancelSharedNote(c *gin.Context) {
	userId, err := u.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenInvalid, c)
		return
	}

	shareIdStr := c.Param("id")
	shareId, err := strconv.Atoi(shareIdStr)
	if err != nil || shareId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "分享ID格式错误", c)
		return
	}

	errCode := u.userService.CancelSharedNote(c.Request.Context(), userId, shareId)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(nil, c)
}
