package model

import "time"

type KnowledgeContent struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NodeID       int       `gorm:"not null" json:"nodeId"`
	Content      string    `gorm:"type:longtext" json:"content"`
	ContentDraft string    `gorm:"type:longtext;default:NULL" json:"contentDraft"`
	AuditStatus  int8      `gorm:"default:0;index:idx_audit_status" json:"auditStatus"`
	LastLogID    int64     `gorm:"column:last_log_id;default:0" json:"lastLogId"`
	HasDraft     int8      `gorm:"type:tinyint(1);default:0" json:"hasDraft"`
	VectorID     string    `gorm:"column:vector_id;type:varchar(100)" json:"vectorId"`
	Source       string    `gorm:"default:'Hello-Algo';type:varchar(50)" json:"source"`
	ImageID      int       `gorm:"column:image_id;default:0" json:"imageId"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (KnowledgeContent) TableName() string {
	return "knowledge_contents"
}
