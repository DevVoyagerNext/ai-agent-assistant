package model

import "time"

// UserSubjectLike 用户教材点赞表
type UserSubjectLike struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"not null;uniqueIndex:uk_user_subject,priority:1" json:"userId"`
	SubjectID int       `gorm:"not null;uniqueIndex:uk_user_subject,priority:2" json:"subjectId"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName UserSubjectLike 表名
func (UserSubjectLike) TableName() string {
	return "user_subject_likes"
}
