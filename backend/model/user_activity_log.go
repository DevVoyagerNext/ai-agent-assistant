package model

import "time"

// UserActivityLog 用户行为明细流水表
type UserActivityLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:流水主键ID" json:"id"`
	UserID     int       `gorm:"not null;index:idx_user_time,priority:1;comment:用户ID" json:"userId"`
	ActionType string    `gorm:"type:varchar(50);not null;comment:行为类型（如：study, create, share, like）" json:"actionType"`
	TargetType string    `gorm:"type:varchar(50);not null;default:'';index:idx_target_resource,priority:1;comment:对象类型（对应业务表名，如：knowledge_nodes, user_notes, subjects）" json:"targetType"`
	TargetID   int       `gorm:"default:0;index:idx_target_resource,priority:2;comment:操作目标的主键ID" json:"targetId"`
	ActionDesc string    `gorm:"type:varchar(255);default:'';comment:行为的具体描述（如：学习了“单链表”章节）" json:"actionDesc"`
	Score      int       `gorm:"default:0;comment:单次行为产生的活跃分" json:"score"`
	ClientIP   string    `gorm:"type:varchar(45);default:'';comment:操作时的IP地址" json:"clientIp"`
	CreatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index:idx_user_time,priority:2;index:idx_created_at;comment:行为发生时间" json:"createdAt"`
}

// TableName 表名
func (UserActivityLog) TableName() string {
	return "user_activity_logs"
}
