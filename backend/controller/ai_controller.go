package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils/response"
	"backend/service"

	"github.com/gin-gonic/gin"
)

type AIController struct {
	aiService service.AIService
}

// Chat AI 聊天接口
func (con *AIController) Chat(c *gin.Context) {
	var req dto.AIChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(errmsg.CodeError, "参数错误或内容过长", c)
		return
	}

	reply, err := con.aiService.Chat(c.Request.Context(), req.Prompt)
	if err != nil {
		response.FailWithMsg(errmsg.CodeError, err.Error(), c)
		return
	}

	response.Ok(dto.AIChatRes{
		Reply: reply,
	}, c)
}
