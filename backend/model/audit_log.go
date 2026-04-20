package model

import "time"

// AuditLog 审批行为历史流水表
type AuditLog struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;comment:流水主键ID" json:"id"`
	TargetType    string    `gorm:"type:enum('subject', 'node', 'content');not null;index:idx_target,priority:1;comment:审批对象类型" json:"targetType"`
	TargetID      int       `gorm:"not null;index:idx_target,priority:2;comment:对应主表的主键ID" json:"targetId"`
	AdminID       int       `gorm:"not null;index:idx_admin;comment:审批管理员ID (关联users.id)" json:"adminId"`
	Action        string    `gorm:"type:enum('approve', 'reject');not null;comment:审批动作：approve=通过, reject=驳回" json:"action"`
	Remark        string    `gorm:"type:varchar(500);default:'';comment:审批意见/驳回理由" json:"remark"`
	DraftSnapshot string    `gorm:"type:json;default:NULL;comment:审批那一刻的草稿数据快照(用于对账)" json:"draftSnapshot"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;comment:审批时间" json:"createdAt"`
}

// TableName 表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
