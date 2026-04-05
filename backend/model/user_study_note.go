package model

import "time"

// UserStudyNote 用户学习数据表：存储用户学习进度、个人笔记
type UserStudyNote struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;comment:笔记主键ID" json:"id"`
	UserID      int       `gorm:"not null;uniqueIndex:idx_user_node,priority:1;comment:所属用户ID" json:"userId"`
	NodeID      int       `gorm:"not null;uniqueIndex:idx_user_node,priority:2;comment:关联知识节点ID" json:"nodeId"`
	NoteContent string    `gorm:"type:mediumtext;comment:用户笔记内容" json:"noteContent"`
	IsImportant int8      `gorm:"default:0;comment:笔记是否标记为重要" json:"isImportant"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

// TableName UserStudyNote 表名
func (UserStudyNote) TableName() string {
	return "user_study_notes"
}
