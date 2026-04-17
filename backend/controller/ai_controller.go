package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AIController struct {
	aiService   service.AIService
	authService service.AuthService
}

// Chat AI 聊天接口
func (con *AIController) Chat(c *gin.Context) {
	var req dto.AIChatReq
	// ShouldBind 会根据 Content-Type 自动选择 JSON 绑定或 FormData 绑定
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误或内容过长", c)
		return
	}

	var fileContents []string
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files := form.File["files"]
		if len(files) > 3 {
			response.FailWithMsg(errmsg.CodeError, "最多只能上传3个文件", c)
			return
		}

		for _, file := range files {
			if file.Size > 5*1024*1024 { // 限制单文件 5MB
				response.FailWithMsg(errmsg.CodeError, "单个文件大小不能超过5MB: "+file.Filename, c)
				return
			}

			f, err := file.Open()
			if err != nil {
				response.FailWithMsg(errmsg.CodeError, "无法读取文件: "+file.Filename, c)
				return
			}

			contentBytes, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				response.FailWithMsg(errmsg.CodeError, "读取文件内容失败: "+file.Filename, c)
				return
			}

			fileContents = append(fileContents, fmt.Sprintf("【文件: %s】\n%s\n", file.Filename, string(contentBytes)))
		}
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	res, err := con.aiService.Chat(c.Request.Context(), userId, req, fileContents)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(res, c)
}

// UpdateSessionTitle 修改用户会话标题
func (con *AIController) UpdateSessionTitle(c *gin.Context) {
	sessionIdStr := c.Param("id")
	sessionId, err := strconv.ParseInt(sessionIdStr, 10, 64)
	if err != nil || sessionId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "会话ID格式错误", c)
		return
	}

	var req dto.UpdateSessionTitleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "标题格式错误或过长", c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	if err := con.aiService.UpdateSessionTitle(c.Request.Context(), userId, sessionId, req.Title); err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(nil, c)
}

// GetUserSessions 获取用户的历史会话列表（游标分页）
func (con *AIController) GetUserSessions(c *gin.Context) {
	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	lastIdStr := c.Query("lastId")
	var lastId int64 = 0
	if lastIdStr != "" {
		parsedId, err := strconv.ParseInt(lastIdStr, 10, 64)
		if err == nil && parsedId > 0 {
			lastId = parsedId
		}
	}

	res, err := con.aiService.GetUserSessions(c.Request.Context(), userId, lastId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, "获取会话列表失败", c)
		return
	}

	response.Ok(res, c)
}

// GetSessionMessages 获取具体会话的消息列表
func (con *AIController) GetSessionMessages(c *gin.Context) {
	sessionIdStr := c.Param("id")
	sessionId, err := strconv.ParseInt(sessionIdStr, 10, 64)
	if err != nil || sessionId <= 0 {
		response.FailWithMsg(errmsg.CodeError, "会话ID格式错误", c)
		return
	}

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c)
		return
	}

	res, err := con.aiService.GetSessionMessages(c.Request.Context(), userId, sessionId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(res, c)
}
