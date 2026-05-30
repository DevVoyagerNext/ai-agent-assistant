package model

import "time"

type AuditLog struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType    string    `gorm:"type:enum('subject','node','content');not null;index:idx_target,priority:1" json:"targetType"`
	TargetID      int       `gorm:"not null;index:idx_target,priority:2" json:"targetId"`
	AdminID       int       `gorm:"not null;index:idx_admin" json:"adminId"`
	Action        string    `gorm:"type:enum('approve','reject');not null" json:"action"`
	Remark        string    `gorm:"type:varchar(500);default:''" json:"remark"`
	DraftSnapshot string    `gorm:"type:json;default:NULL" json:"draftSnapshot"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
