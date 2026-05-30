package model

type KnowledgeNode struct {
	ID                   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	SubjectID            int    `gorm:"not null;index:idx_subject_path,priority:1;index:idx_subject_parent,priority:1" json:"subjectId"`
	ParentID             int    `gorm:"default:0;index:idx_subject_parent,priority:2" json:"parentId"`
	Path                 string `gorm:"default:'0/';type:varchar(255);index:idx_subject_path,priority:2,length:128" json:"path"`
	Name                 string `gorm:"not null;type:varchar(150)" json:"name"`
	NameDraft            string `gorm:"type:varchar(150);default:''" json:"nameDraft"`
	Status               string `gorm:"type:enum('draft','published','hidden');default:'draft'" json:"status"`
	AuditStatus          int8   `gorm:"default:0;index:idx_audit_status" json:"auditStatus"`
	LastLogID            int64  `gorm:"column:last_log_id;default:0" json:"lastLogId"`
	HasDraft             int8   `gorm:"type:tinyint(1);default:0" json:"hasDraft"`
	DraftDescendantCount int    `gorm:"not null;default:0" json:"draftDescendantCount"`
	Level                int8   `gorm:"default:1" json:"level"`
	IsLeaf               int8   `gorm:"default:0;index:idx_is_leaf" json:"isLeaf"`
	SortOrder            int    `gorm:"default:0" json:"sortOrder"`
	ImageID              int    `gorm:"column:image_id;default:0" json:"imageId"`
	ImageUrl             string `gorm:"column:image_url;type:varchar(512);default:''" json:"imageUrl"`
}

func (KnowledgeNode) TableName() string {
	return "knowledge_nodes"
}
