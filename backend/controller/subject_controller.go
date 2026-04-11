package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SubjectController struct {
	subjectService service.SubjectService
	authService    service.AuthService
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

// ToggleSubjectLike 点赞或取消点赞教材（需要登录）
func (con *SubjectController) ToggleSubjectLike(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	subjectIdStr := c.Param("id")
	subjectId, err := strconv.Atoi(subjectIdStr)
	if err != nil || subjectId <= 0 {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	isLiked, code := con.subjectService.ToggleSubjectLike(c.Request.Context(), userId, subjectId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}

	// 返回当前的点赞状态
	response.Ok(gin.H{"isLiked": isLiked}, c)
}

// CreateCollectFolder 创建用户收藏夹
func (con *SubjectController) CreateCollectFolder(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	var req dto.CreateCollectFolderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数格式错误", c)
		return
	}

	folder, code := con.subjectService.CreateCollectFolder(c.Request.Context(), userId, req)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}

	response.Ok(folder, c)
}

// AddSubjectToFolder 将教材添加到指定收藏夹
func (con *SubjectController) AddSubjectToFolder(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	folderIdStr := c.Param("folderId")
	folderId, err := strconv.Atoi(folderIdStr)
	if err != nil || folderId <= 0 {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	var req dto.AddSubjectToFolderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数格式错误", c)
		return
	}

	code := con.subjectService.AddSubjectToFolder(c.Request.Context(), userId, folderId, req.SubjectID)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}

	response.Ok(nil, c)
}

// RemoveSubjectFromFolder 从指定收藏夹移除教材
func (con *SubjectController) RemoveSubjectFromFolder(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	folderIdStr := c.Param("folderId")
	folderId, err := strconv.Atoi(folderIdStr)
	if err != nil || folderId <= 0 {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	subjectIdStr := c.Param("subjectId")
	subjectId, err := strconv.Atoi(subjectIdStr)
	if err != nil || subjectId <= 0 {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	code := con.subjectService.RemoveSubjectFromFolder(c.Request.Context(), userId, folderId, subjectId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}

	response.Ok(nil, c)
}

// GetSubjectsByCategory 通过教材分类获取该分类的教材数据（不需要登录）
func (con *SubjectController) GetSubjectsByCategory(c *gin.Context) {
	categoryIdStr := c.Param("id")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	userId, _ := con.authService.GetUserID(c)
	res, code := con.subjectService.GetSubjectsByCategoryID(c.Request.Context(), categoryId, userId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetAllSubjects 获取所有的教材（不需要登录）
func (con *SubjectController) GetAllSubjects(c *gin.Context) {
	userId, _ := con.authService.GetUserID(c)
	res, code := con.subjectService.GetAllSubjects(c.Request.Context(), userId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// SearchSubjects 通过教材名称模糊搜索教材（不需要登录）
func (con *SubjectController) SearchSubjects(c *gin.Context) {
	var req dto.SubjectSearchReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	userId, _ := con.authService.GetUserID(c)
	res, code := con.subjectService.SearchSubjects(c.Request.Context(), req.Keyword, userId, req.Page, req.PageSize)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserCollectedSubjects 获取该用户收藏的教材（需要登录）
func (con *SubjectController) GetUserCollectedSubjects(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserCollectedSubjects(c.Request.Context(), userId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserLikedSubjects 获取该用户点赞的教材（需要登录）
func (con *SubjectController) GetUserLikedSubjects(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserLikedSubjects(c.Request.Context(), userId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserLearningSubjects 获取该用户正在学习的教材（需要登录）
func (con *SubjectController) GetUserLearningSubjects(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserSubjectsByStatus(c.Request.Context(), userId, "learning")
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserCompletedSubjects 获取该用户已经学习完成的教材（需要登录）
func (con *SubjectController) GetUserCompletedSubjects(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserSubjectsByStatus(c.Request.Context(), userId, "completed")
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserLastLearningSubject 分页获取最近学习的教材（需要登录）
func (con *SubjectController) GetUserLastLearningSubject(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	var req dto.PaginationReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	res, code := con.subjectService.GetUserRecentSubjects(c.Request.Context(), userId, req.Page, req.PageSize)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserCollectFolders 获取该用户的收藏夹（需要登录）
func (con *SubjectController) GetUserCollectFolders(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	res, code := con.subjectService.GetUserCollectFolders(c.Request.Context(), userId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}

// GetUserCollectedSubjectsByFolder 获取该用户收藏夹下的教材（需要登录）
func (con *SubjectController) GetUserCollectedSubjectsByFolder(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserNotExist, c)
		return
	}

	folderIdStr := c.Param("folderId")
	folderId, err := strconv.Atoi(folderIdStr)
	if err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	res, code := con.subjectService.GetUserCollectedSubjectsByFolder(c.Request.Context(), userId, folderId)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}
	response.Ok(res, c)
}
