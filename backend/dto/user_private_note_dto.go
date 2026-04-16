package dto

import "time"

// UpsertUserStudyNoteReq 创建或修改随堂笔记请求体 (注意：这里的笔记是指知识点随堂笔记)
// 我们这里是私人笔记 UserPrivateNote

// CreatePrivateNoteReq 创建私人笔记或文件夹请求
type CreatePrivateNoteReq struct {
	ParentID int    `json:"parentId"`
	Type     string `json:"type" binding:"required,oneof=folder markdown"`
	Title    string `json:"title" binding:"required,max=255"`
	Content  string `json:"content"`  // type 为 markdown 时可为空
	IsPublic int8   `json:"isPublic"` // 0=不公开, 1=公开
}

// UpdatePrivateNoteContentReq 修改私人笔记内容请求
type UpdatePrivateNoteContentReq struct {
	Content string `json:"content" binding:"max=1000"`
}

// UpdatePrivateNoteTitleReq 修改标题请求
type UpdatePrivateNoteTitleReq struct {
	Title string `json:"title" binding:"required,max=255"`
}

// UpdatePrivateNotePublicReq 修改公开状态请求
type UpdatePrivateNotePublicReq struct {
	IsPublic int8 `json:"isPublic" binding:"oneof=0 1"`
}

// PrivateNoteItemRes 私人笔记列表项
type PrivateNoteItemRes struct {
	ID        uint      `json:"id"`
	ParentID  int       `json:"parentId"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	IsPublic  int8      `json:"isPublic"`
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
	IsPublic  int8      `json:"isPublic"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// PrivateNoteResponse 获取内容的响应结构，根据类型可能是列表或详情
type PrivateNoteResponse struct {
	Type     string                `json:"type"`               // folder 或 markdown
	Total    int64                 `json:"total,omitempty"`    // 当 type 为 folder 时有值
	Children []PrivateNoteItemRes  `json:"children,omitempty"` // 当 type 为 folder 时有值
	Content  *PrivateNoteDetailRes `json:"content,omitempty"`  // 当 type 为 markdown 时有值
}

// SharePrivateNoteReq 分享私人笔记请求
type SharePrivateNoteReq struct {
	ExpiresAt string `json:"expiresAt" binding:"required"` // 过期时间，格式如 "2006-01-02 15:04:05"
}

// SharePrivateNoteRes 分享私人笔记返回
type SharePrivateNoteRes struct {
	ShareToken string `json:"shareToken"`
	ShareCode  string `json:"shareCode"`
	ExpiresAt  string `json:"expiresAt"`
}
