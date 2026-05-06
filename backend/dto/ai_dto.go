package dto

import (
	"encoding/json"
	"strings"
)

// AIChatFile AI 聊天文件信息
type AIChatFile struct {
	FileID   uint   `json:"file_id" form:"file_id"`     // 文件ID
	FileURL  string `json:"file_url" form:"file_url"`   // 文件下载链接/存储路径
	FileName string `json:"file_name" form:"file_name"` // 文件名
	FileType string `json:"file_type" form:"file_type"` // 文件类型 (如: image/png, application/pdf等)
	FileSize int64  `json:"file_size" form:"file_size"` // 文件大小 (Byte)
}

// ParseAIChatFiles 兼容解析 files JSON 字符串
func ParseAIChatFiles(raw string) ([]AIChatFile, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var files []AIChatFile
	if err := json.Unmarshal([]byte(raw), &files); err == nil {
		return files, nil
	}

	var single AIChatFile
	if err := json.Unmarshal([]byte(raw), &single); err == nil {
		return []AIChatFile{single}, nil
	}
	return nil, json.Unmarshal([]byte(raw), &files)
}

// AIChatReq AI 聊天请求参数
type AIChatReq struct {
	SkillID   string       `json:"skill_id" form:"skill_id"`                                 // 核心：功能标识，比如"wrong_question"、"essay_correct"、"feynman"
	UserInput string       `json:"user_input" form:"user_input" binding:"required,max=5000"` // 用户输入的文本
	SessionID string       `json:"session_id" form:"session_id"`                             // 会话ID，用于上下文传递
	Files     []AIChatFile `json:"files" form:"files"`                                       // 用户上传的图片/文档/PDF详细信息

	// 以下为兼容旧版预留参数
	Prompt         string `json:"prompt" form:"prompt"`                                              // 旧版参数
	ParentID       *int64 `json:"parentId" form:"parentId"`                                          // 可选，用于分支对话
	CurrentPageURL string `json:"currentPageUrl" form:"currentPageUrl" binding:"omitempty,max=2048"` // 可选，当前网页链接
	SelectedText   string `json:"selectedText" form:"selectedText" binding:"omitempty,max=5000"`     // 可选，用户选中的文本
}

// ChatStreamChunk 用于流式返回时区分工具状态、思考过程和正式回复
type ChatStreamChunk struct {
	Type    string `json:"type"`    // "tool"、"reasoning" 或 "message"
	Content string `json:"content"` // 内容片段
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

// MessageListRes 消息列表响应（游标分页）
type MessageListRes struct {
	List    []MessageItemRes `json:"list"`
	HasMore bool             `json:"hasMore"`
}
