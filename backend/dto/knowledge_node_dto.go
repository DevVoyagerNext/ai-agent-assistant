package dto

// KnowledgeNodeItemRes 知识点基础信息（含用户进度及难度评价）
type KnowledgeNodeItemRes struct {
	ID        uint   `json:"id"`
	SubjectID int    `json:"subjectId"`
	ParentID  int    `json:"parentId"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Level     int8   `json:"level"`
	IsLeaf    int8   `json:"isLeaf"`
	SortOrder int    `json:"sortOrder"`
	ImageID   int    `json:"imageId"`

	// 难度评价信息（来自 node_metrics 表）
	EasyCount   int `json:"easyCount"`
	MediumCount int `json:"mediumCount"`
	HardCount   int `json:"hardCount"`

	// 用户的知识点学习进度 (unstarted, learning, completed)
	UserProgressStatus string `json:"userProgressStatus"`
}

// KnowledgeNodeDetailRes 知识点详细信息（带内容正文）
type KnowledgeNodeDetailRes struct {
	KnowledgeNodeItemRes
	Content string `json:"content"` // 对应 knowledge_contents 表正文
}

// UserStudyNoteRes 用户随堂笔记返回体
type UserStudyNoteRes struct {
	ID          uint   `json:"id"`
	NodeID      int    `json:"nodeId"`
	NoteContent string `json:"noteContent"`
	IsImportant int8   `json:"isImportant"`
	UpdatedAt   string `json:"updatedAt"`
}

// UpdateNodeStatusReq 修改知识点学习状态请求体
type UpdateNodeStatusReq struct {
	Status string `json:"status" binding:"required,oneof=unstarted learning completed"`
}

// MarkNodeDifficultyReq 标记知识点难度请求体
type MarkNodeDifficultyReq struct {
	Difficulty string `json:"difficulty" binding:"required,oneof=easy medium hard"`
}
