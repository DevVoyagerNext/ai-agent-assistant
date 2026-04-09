package model

import "time"

// UserCollectItem 收藏夹具体内容表
type UserCollectItem struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"not null;index:idx_user_id;comment:用户ID" json:"userId"`
	FolderID  int       `gorm:"not null;uniqueIndex:uk_folder_subject,priority:1;comment:关联收藏夹ID" json:"folderId"`
	SubjectID int       `gorm:"not null;uniqueIndex:uk_folder_subject,priority:2;comment:关联教材ID" json:"subjectId"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName UserCollectItem 表名
func (UserCollectItem) TableName() string {
	return "user_collect_items"
}
