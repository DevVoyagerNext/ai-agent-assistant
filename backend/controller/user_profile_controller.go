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

// UpdateSharedNoteStatus 更新分享状态接口
func (u *UserController) UpdateSharedNoteStatus(c *gin.Context) {
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

	var req dto.UpdateSharedNoteStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	errCode := u.userService.UpdateSharedNoteStatus(c.Request.Context(), userId, shareId, req.IsActive)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(nil, c)
}

// UpdateSharedNoteExpire 更新分享过期时间接口
func (u *UserController) UpdateSharedNoteExpire(c *gin.Context) {
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

	var req dto.UpdateSharedNoteExpireReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	if req.ExpireMinutes <= 0 && req.ExpireAt == "" {
		response.FailWithMsg(errmsg.CodeError, "至少需要传递延长分钟数或具体过期时间之一", c)
		return
	}

	errCode := u.userService.UpdateSharedNoteExpire(c.Request.Context(), userId, shareId, req)
	if errCode != errmsg.CodeSuccess {
		// 为了给前端更友好的提示，如果是因为时间解析或过去时间导致的错误，这里也可以细化
		// 但基于你之前的代码结构，我们统一抛出 CodeError
		response.FailWithMsg(errCode, "更新过期时间失败，请检查时间格式是否为 2006-01-02 15:04:05 且时间必须在未来", c)
		return
	}

	response.Ok(nil, c)
}
