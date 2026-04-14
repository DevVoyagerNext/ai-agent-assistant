package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type KnowledgeNodeController struct {
	nodeService service.KnowledgeNodeService
	authService service.AuthService
}

// GetTopLevelNodes 获取教材顶级知识点
func (con *KnowledgeNodeController) GetTopLevelNodes(c *gin.Context) {
	subjectIdStr := c.Query("subjectId")
	subjectId, err := strconv.Atoi(subjectIdStr)
	if err != nil || subjectId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "教材ID格式错误", c)
		return
	}

	userId, _ := con.authService.GetUserID(c) // 允许游客

	nodes, err := con.nodeService.GetTopLevelNodes(c.Request.Context(), subjectId, userId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "获取顶级知识点失败", c)
		return
	}

	response.Ok(nodes, c)
}

// GetChildNodes 获取某个知识点的直属子节点
func (con *KnowledgeNodeController) GetChildNodes(c *gin.Context) {
	parentNodeIdStr := c.Param("nodeId")
	parentNodeId, err := strconv.Atoi(parentNodeIdStr)
	if err != nil || parentNodeId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "知识点ID格式错误", c)
		return
	}

	userId, _ := con.authService.GetUserID(c) // 允许游客

	nodes, err := con.nodeService.GetChildNodes(c.Request.Context(), parentNodeId, userId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "获取子节点失败", c)
		return
	}

	response.Ok(nodes, c)
}

// GetPathNodes 获取某个知识点的路径节点列表（所有祖先节点的同级节点列表）
func (con *KnowledgeNodeController) GetPathNodes(c *gin.Context) {
	nodeIdStr := c.Query("nodeId")
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil || nodeId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "知识点ID格式错误", c)
		return
	}

	userId, _ := con.authService.GetUserID(c) // 允许游客，如果是游客，则学习状态默认为 unstarted

	nodes, err := con.nodeService.GetPathNodes(c.Request.Context(), nodeId, userId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "获取路径节点列表失败", c)
		return
	}

	response.Ok(nodes, c)
}

// GetNodeDetail 获取知识点详情（包含正文、难度评价、用户进度）
func (con *KnowledgeNodeController) GetNodeDetail(c *gin.Context) {
	nodeIdStr := c.Param("nodeId")
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil || nodeId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "知识点ID格式错误", c)
		return
	}

	userId, _ := con.authService.GetUserID(c) // 允许游客

	detail, err := con.nodeService.GetNodeDetail(c.Request.Context(), nodeId, userId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "获取知识点详情失败", c)
		return
	}

	response.Ok(detail, c)
}

// GetUserStudyNote 获取用户对某个知识点的随堂笔记
func (con *KnowledgeNodeController) GetUserStudyNote(c *gin.Context) {
	nodeIdStr := c.Param("nodeId")
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil || nodeId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "知识点ID格式错误", c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	note, err := con.nodeService.GetUserStudyNote(c.Request.Context(), nodeId, userId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "获取随堂笔记失败", c)
		return
	}

	response.Ok(note, c)
}

// UpsertUserStudyNote 创建或修改随堂笔记
func (con *KnowledgeNodeController) UpsertUserStudyNote(c *gin.Context) {
	nodeIdStr := c.Param("nodeId")
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil || nodeId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "知识点ID格式错误", c)
		return
	}

	var req dto.UpsertUserStudyNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	err = con.nodeService.UpsertUserStudyNote(c.Request.Context(), userId, nodeId, req)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}

// UpdateNodeStatus 更新知识点学习状态
func (con *KnowledgeNodeController) UpdateNodeStatus(c *gin.Context) {
	nodeIdStr := c.Param("nodeId")
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil || nodeId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "知识点ID格式错误", c)
		return
	}

	var req dto.UpdateNodeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	err = con.nodeService.UpdateNodeStatus(c.Request.Context(), userId, nodeId, req.Status)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "更新状态失败", c)
		return
	}

	response.Ok(nil, c)
}

// MarkNodeDifficulty 标记知识点难度
func (con *KnowledgeNodeController) MarkNodeDifficulty(c *gin.Context) {
	nodeIdStr := c.Param("nodeId")
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil || nodeId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "知识点ID格式错误", c)
		return
	}

	var req dto.MarkNodeDifficultyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误: "+err.Error(), c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	err = con.nodeService.MarkNodeDifficulty(c.Request.Context(), userId, nodeId, req.Difficulty)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "标记难度失败", c)
		return
	}

	response.Ok(nil, c)
}
