package dto

import "time"

// CreatePrivateNoteReq 创建私人笔记或文件夹请求
type CreatePrivateNoteReq struct {
	ParentID int    `json:"parentId"`
	Type     string `json:"type" binding:"required,oneof=folder markdown"`
	Title    string `json:"title" binding:"required,max=255"`
	Content  string `json:"content"` // type 为 markdown 时不能为空，后面在 service 层做更复杂的校验
}

// PrivateNoteItemRes 私人笔记列表项
type PrivateNoteItemRes struct {
	ID        uint      `json:"id"`
	ParentID  int       `json:"parentId"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// PrivateNoteDetailRes 私人笔记详情
type PrivateNoteDetailRes struct {
	ID        uint      `json:"id"`
	ParentID  int       `json:"parentId"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// PrivateNoteResponse 获取内容的响应结构，根据类型可能是列表或详情
type PrivateNoteResponse struct {
	Type     string      `json:"type"` // folder 或 markdown
	Children []PrivateNoteItemRes `json:"children,omitempty"` // 当 type 为 folder 时有值
	Content  *PrivateNoteDetailRes `json:"content,omitempty"`  // 当 type 为 markdown 时有值
}
