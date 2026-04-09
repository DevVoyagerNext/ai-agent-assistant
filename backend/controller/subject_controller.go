package controller

import (
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SubjectController struct {
	subjectService service.SubjectService
}

// GetCategories 获取教材分类接口（不需要登录）
func (con *SubjectController) GetCategories(c *gin.Context) {
	res, code := con.subjectService.GetCategories(c.Request.Context())
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetSubjectsByCategory 通过教材分类获取该分类的教材数据（不需要登录）
func (con *SubjectController) GetSubjectsByCategory(c *gin.Context) {
	categoryIdStr := c.Param("id")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	res, code := con.subjectService.GetSubjectsByCategoryID(c.Request.Context(), categoryId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetAllSubjects 获取所有的教材（不需要登录）
func (con *SubjectController) GetAllSubjects(c *gin.Context) {
	res, code := con.subjectService.GetAllSubjects(c.Request.Context())
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserCollectedSubjects 获取该用户收藏的教材（需要登录）
func (con *SubjectController) GetUserCollectedSubjects(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserCollectedSubjects(c.Request.Context(), userId.(uint))
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserLikedSubjects 获取该用户点赞的教材（需要登录）
func (con *SubjectController) GetUserLikedSubjects(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserLikedSubjects(c.Request.Context(), userId.(uint))
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserLearningSubjects 获取该用户正在学习的教材（需要登录）
func (con *SubjectController) GetUserLearningSubjects(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserSubjectsByStatus(c.Request.Context(), userId.(uint), "learning")
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserCompletedSubjects 获取该用户已经学习完成的教材（需要登录）
func (con *SubjectController) GetUserCompletedSubjects(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserSubjectsByStatus(c.Request.Context(), userId.(uint), "completed")
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserLastLearningSubject 获取该用户上次学习的教材及进度（需要登录）
func (con *SubjectController) GetUserLastLearningSubject(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserLastLearningSubject(c.Request.Context(), userId.(uint))
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}
