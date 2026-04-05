package model

import "time"

// KnowledgeContent 知识内容表：存储知识点的Markdown原文内容
type KnowledgeContent struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:内容主键ID" json:"id"`
	NodeID    int       `gorm:"not null;comment:关联知识节点ID" json:"nodeId"`
	Content   string    `gorm:"type:longtext;comment:知识点Markdown格式正文" json:"content"`
	VectorID  string    `gorm:"type:varchar(100);comment:向量数据库唯一ID，用于AI检索" json:"vectorId"`
	Source    string    `gorm:"default:'Hello-Algo';type:varchar(50);comment:内容来源标记" json:"source"`
	ImageID   int       `gorm:"default:0;comment:内容配图ID，关联images表" json:"imageId"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:内容更新时间" json:"updatedAt"`
}

// TableName KnowledgeContent 表名
func (KnowledgeContent) TableName() string {
	return "knowledge_contents"
}
