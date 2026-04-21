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

// KnowledgeNodeSimpleRes 知识点简易信息
type KnowledgeNodeSimpleRes struct {
	ID        uint   `json:"id"`
	ParentID  int    `json:"parentId"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	SortOrder int    `json:"sortOrder"`
	ImageUrl  string `json:"imageUrl"`
	IsLeaf    int8   `json:"isLeaf"`
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

// UpsertUserStudyNoteReq 创建或修改随堂笔记请求体
type UpsertUserStudyNoteReq struct {
	NoteContent string `json:"noteContent" binding:"max=1000"`
	IsImportant int8   `json:"isImportant"`
}

// UpdateNodeStatusReq 修改知识点学习状态请求体
type UpdateNodeStatusReq struct {
	Status string `json:"status" binding:"required,oneof=unstarted learning completed"`
}

// MarkNodeDifficultyReq 标记知识点难度请求体
type MarkNodeDifficultyReq struct {
	Difficulty string `json:"difficulty" binding:"required,oneof=easy medium hard"`
}

// CreateKnowledgeNodeReq 创建知识节点请求体
type CreateKnowledgeNodeReq struct {
	SubjectID int    `json:"subjectId" binding:"required,gt=0"`
	ParentID  int    `json:"parentId" binding:"min=0"`
	NameDraft string `json:"nameDraft" binding:"required,min=1,max=150"`
}

// UpdateKnowledgeNodeDraftReq 更新知识节点名称请求体
type UpdateKnowledgeNodeDraftReq struct {
	SubjectID int    `json:"subjectId" binding:"required,gt=0"`
	NameDraft string `json:"nameDraft" binding:"required,min=1,max=150"`
}

// UpsertKnowledgeContentReq 创建或更新知识点正文内容请求体
type UpsertKnowledgeContentReq struct {
	ContentDraft string `json:"contentDraft" binding:"required"`
}

// AuthorChildNodeRes 创作者视角的子节点信息
type AuthorChildNodeRes struct {
	ID          uint   `json:"id"`
	SubjectID   int    `json:"subjectId"`
	ParentID    int    `json:"parentId"`
	Name        string `json:"name"`
	NameDraft   string `json:"nameDraft"`
	Status      string `json:"status"`
	AuditStatus int8   `json:"auditStatus"`
	HasDraft    int8   `json:"hasDraft"`
	Path        string `json:"path"`
	IsLeaf      int8   `json:"isLeaf"`
}

// AuthorNodeContentRes 创作者视角的节点内容信息
type AuthorNodeContentRes struct {
	Content      string `json:"content"`
	ContentDraft string `json:"contentDraft"`
	AuditStatus  int8   `json:"auditStatus"`
	HasDraft     int8   `json:"hasDraft"`
	IsLeaf       int8   `json:"isLeaf"`
}

// AuthorInitEditRes 创作者进入编辑页面的初始响应信息
type AuthorInitEditRes struct {
	LastNodeID int                  `json:"lastNodeId"` // 最后编辑或默认的节点ID
	NodeList   []AuthorChildNodeRes `json:"nodeList"`   // 需要展开的节点树扁平列表
}
