package model

import "time"

// Session 会话表
type Session struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID               int64     `gorm:"index:idx_user_list,priority:1;comment:所属用户ID" json:"userId"`
	Title                string    `gorm:"type:varchar(100);default:'新对话';comment:会话标题" json:"title"`
	Summary              string    `gorm:"type:text;comment:会话摘要" json:"summary"`
	LastSummaryMessageID int64     `gorm:"default:0;comment:最后一次摘要的消息ID" json:"lastSummaryMessageId"`
	ModelID              string    `gorm:"type:varchar(50);default:'default-model';comment:使用的模型ID" json:"modelId"`
	IsDeleted            bool      `gorm:"default:false;index:idx_user_list,priority:2;comment:逻辑删除标记" json:"isDeleted"`
	UpdatedAt            time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;index:idx_user_list,priority:3;comment:最后更新时间" json:"updatedAt"`
	CreatedAt            time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
}

// TableName 表名
func (Session) TableName() string {
	return "sessions"
}
