package model

import "time"

// KnowledgeContent 知识内容表：存储知识点的Markdown原文内容
type KnowledgeContent struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;comment:内容主键ID" json:"id"`
	NodeID       int       `gorm:"not null;comment:关联知识节点ID" json:"nodeId"`
	Content      string    `gorm:"type:longtext;comment:知识点Markdown格式正文" json:"content"`
	ContentDraft string    `gorm:"type:longtext;default:NULL;comment:正文内容草稿" json:"contentDraft"`
	AuditStatus  int8      `gorm:"default:0;index:idx_audit_status;comment:审核状态：0=编辑中, 1=待审核, 2=已通过, 3=被驳回" json:"auditStatus"`
	LastLogID    int64     `gorm:"default:0;comment:关联最新一条审批流水ID" json:"lastLogId"`
	HasDraft     int8      `gorm:"type:tinyint(1);default:0;comment:标记是否有未处理的草稿：1=是, 0=否" json:"hasDraft"`
	VectorID     string    `gorm:"type:varchar(100);comment:向量数据库唯一ID，用于AI检索" json:"vectorId"`
	Source       string    `gorm:"default:'Hello-Algo';type:varchar(50);comment:内容来源标记" json:"source"`
	ImageID      int       `gorm:"default:0;comment:内容配图ID，关联images表" json:"imageId"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:内容更新时间" json:"updatedAt"`
}

// TableName KnowledgeContent 表名
func (KnowledgeContent) TableName() string {
	return "knowledge_contents"
}
