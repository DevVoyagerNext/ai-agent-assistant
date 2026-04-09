package model

import "time"

// UserCollectFolder 用户收藏夹名录表
type UserCollectFolder struct {
	ID          int       `gorm:"primaryKey;autoIncrement;comment:收藏夹ID" json:"id"`
	UserID      int       `gorm:"not null;index:idx_user_id;comment:所属用户ID" json:"userId"`
	Name        string    `gorm:"not null;type:varchar(100);comment:收藏夹名称" json:"name"`
	Description string    `gorm:"default:'';type:varchar(255);comment:收藏夹简介" json:"description"`
	IsPublic    int8      `gorm:"default:0;type:tinyint(1);comment:是否公开：1=公开, 0=私密" json:"isPublic"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

// TableName UserCollectFolder 表名
func (UserCollectFolder) TableName() string {
	return "user_collect_folders"
}
