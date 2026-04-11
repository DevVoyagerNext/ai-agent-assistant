package model

import "time"

// UserSubjectProgress 用户教材整体进度表
type UserSubjectProgress struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int       `gorm:"not null;uniqueIndex:uk_user_subject,priority:1;index:idx_user_time,priority:1;comment:用户ID" json:"userId"`
	SubjectID       int       `gorm:"not null;uniqueIndex:uk_user_subject,priority:2;comment:教材ID" json:"subjectId"`
	ProgressPercent float64   `gorm:"type:decimal(5,2);default:0.00;comment:进度百分比 (例如: 85.50)" json:"progressPercent"`
	LastNodeID      int       `gorm:"not null;comment:最后一次学习的节点ID" json:"lastNodeId"`
	LastStudyTime   time.Time `gorm:"type:datetime;not null;index:idx_user_time,priority:2,sort:desc;comment:最后一次学习时间" json:"lastStudyTime"`
	CreatedAt       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;comment:首次加入学习的时间" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

// TableName UserSubjectProgress 表名
func (UserSubjectProgress) TableName() string {
	return "user_subject_progress"
}
