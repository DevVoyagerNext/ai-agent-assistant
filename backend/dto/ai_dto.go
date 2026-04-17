package dto

// AIChatReq AI 聊天请求参数
type AIChatReq struct {
	Prompt string `json:"prompt" binding:"required,max=2000"`
}

// AIChatRes AI 聊天返回结果
type AIChatRes struct {
	Reply string `json:"reply"`
}
