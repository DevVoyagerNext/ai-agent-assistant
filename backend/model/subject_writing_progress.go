package model

import "time"

// SubjectWritingProgress 教材编写断点记录表
type SubjectWritingProgress struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int       `gorm:"not null;uniqueIndex:uk_user_subject,priority:1;index:idx_user_recent,priority:1;comment:作者ID" json:"userId"`
	SubjectID  int       `gorm:"not null;uniqueIndex:uk_user_subject,priority:2;comment:教材ID" json:"subjectId"`
	LastNodeID int       `gorm:"not null;comment:最后一次查看/编辑的知识节点ID" json:"lastNodeId"`
	UpdatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;index:idx_user_recent,priority:2,sort:desc;comment:最后操作时间" json:"updatedAt"`
}

// TableName 表名
func (SubjectWritingProgress) TableName() string {
	return "subject_writing_progress"
}
