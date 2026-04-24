package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"
	"encoding/base64"
	"io"
	"strconv"
	"strings"

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

	userId, err := con.authService.GetUserID(c)
	if err != nil || userId == 0 {
		response.FailWithCode(errmsg.UserTokenNotExist, c) // 必须登录
		return
	}

	streamChan, sessionId, messageId, err := con.aiService.Chat(c.Request.Context(), userId, req)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	// 设置 Server-Sent Events 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 首先发送包含会话和消息ID的元数据事件
	c.SSEvent("meta", gin.H{
		"sessionId": sessionId,
		"messageId": messageId,
	})
	c.Writer.Flush()

	// 流式监听并向客户端发送 chunk 数据
	c.Stream(func(w io.Writer) bool {
		if chunk, ok := <-streamChan; ok {
			// 将字符串转换为 Base64 编码发送，既能完美保留空格和换行，又无需 JSON 序列化的开销
			encodedMsg := base64.StdEncoding.EncodeToString([]byte(chunk.Content))
			c.SSEvent(chunk.Type, encodedMsg)
			c.Writer.Flush()
			return true
		}
		// 通道关闭，发送结束标记
		c.SSEvent("done", "[DONE]")
		c.Writer.Flush()
		return false
	})
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

// GetSessionMessages 获取具体会话的消息列表（游标分页向上拉取）
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

	lastIdStr := c.Query("lastId")
	var lastId int64 = 0
	if lastIdStr != "" {
		parsedId, err := strconv.ParseInt(lastIdStr, 10, 64)
		if err == nil && parsedId > 0 {
			lastId = parsedId
		}
	}

	res, err := con.aiService.GetSessionMessages(c.Request.Context(), userId, sessionId, lastId)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(res, c)
}

// DownloadExportByTicket 通过一次性下载凭证下载 AI 生成的 PDF 导出文件
func (con *AIController) DownloadExportByTicket(c *gin.Context) {
	ticket := strings.TrimSpace(c.Param("ticket"))
	if ticket == "" {
		response.FailWithMsg(errmsg.CodeError, "下载凭证不能为空", c)
		return
	}

	userId, fileName, err := con.aiService.ConsumeExportDownloadTicket(c.Request.Context(), ticket)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	filePath, err := con.aiService.GetExportFilePath(c.Request.Context(), userId, fileName)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.FileAttachment(filePath, fileName)
}
