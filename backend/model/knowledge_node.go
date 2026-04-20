package model

// KnowledgeNode 知识节点表：学习内容的树状骨架（篇-章-节-知识点）
type KnowledgeNode struct {
	ID          uint   `gorm:"primaryKey;autoIncrement;comment:知识节点主键ID" json:"id"`
	SubjectID   int    `gorm:"not null;index:idx_subject_path,priority:1;index:idx_subject_parent,priority:1;comment:所属学科ID，关联subjects表" json:"subjectId"`
	ParentID    int    `gorm:"default:0;index:idx_subject_parent,priority:2;comment:父节点ID，0表示顶级“篇”节点" json:"parentId"`
	Path        string `gorm:"default:'0/';type:varchar(255);index:idx_subject_path,priority:2,length:128;comment:层级路径，用于快速检索和排序" json:"path"`
	Name        string `gorm:"not null;type:varchar(150);comment:节点名称（如：栈、单链表、第一章）" json:"name"`
	NameDraft   string `gorm:"type:varchar(150);default:'';comment:节点标题草稿" json:"nameDraft"`
	AuditStatus int8   `gorm:"default:0;index:idx_audit_status;comment:审核状态：0=编辑中, 1=待审核, 2=已通过, 3=被驳回" json:"auditStatus"`
	LastLogID   int64  `gorm:"default:0;comment:关联最新一条审批流水ID" json:"lastLogId"`
	Level       int8   `gorm:"default:1;comment:节点层级：1=篇, 2=章, 3=节, 4=知识点" json:"level"`
	IsLeaf      int8   `gorm:"default:0;index:idx_is_leaf;comment:是否为叶子节点（内容页）：1=是, 0=否" json:"isLeaf"`
	SortOrder   int    `gorm:"default:0;comment:同层级下的显示排序序号" json:"sortOrder"`
	ImageID     int    `gorm:"default:0;comment:节点封面/配图ID" json:"imageId"`
	ImageUrl    string `gorm:"type:varchar(512);default:'';comment:节点封面图外部URL地址" json:"imageUrl"`
}

// TableName KnowledgeNode 表名
func (KnowledgeNode) TableName() string {
	return "knowledge_nodes"
}
