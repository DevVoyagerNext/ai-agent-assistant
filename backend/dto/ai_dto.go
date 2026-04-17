package dto

// AIChatReq AI 聊天请求参数
type AIChatReq struct {
	Prompt    string `json:"prompt" form:"prompt" binding:"required,max=1000"`
	SessionID int64  `json:"sessionId" form:"sessionId"` // 可选，不传或为0表示新会话
	ParentID  *int64 `json:"parentId" form:"parentId"`   // 可选，用于分支对话
}

// AIChatRes AI 聊天返回结果
type AIChatRes struct {
	Reply     string `json:"reply"`
	SessionID int64  `json:"sessionId"`
	MessageID int64  `json:"messageId"`
}

// UpdateSessionTitleReq 修改会话标题请求参数
type UpdateSessionTitleReq struct {
	Title string `json:"title" binding:"required,max=100"`
}

// SessionItemRes 会话列表项响应
type SessionItemRes struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ModelID   string `json:"modelId"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
}

// SessionListRes 会话列表响应（游标分页）
type SessionListRes struct {
	List    []SessionItemRes `json:"list"`
	HasMore bool             `json:"hasMore"`
}

// MessageItemRes 消息列表项响应
type MessageItemRes struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"sessionId"`
	ParentID  *int64 `json:"parentId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// MessageListRes 消息列表响应
type MessageListRes struct {
	Total int64            `json:"total"`
	List  []MessageItemRes `json:"list"`
}
