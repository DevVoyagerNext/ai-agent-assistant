package model

import "time"

// UserStudyStatus 用户学习状态精简表
type UserStudyStatus struct {
	ID            uint       `gorm:"primaryKey;autoIncrement;comment:状态记录主键ID" json:"id"`
	UserID        int        `gorm:"not null;uniqueIndex:idx_user_node,priority:1;comment:所属用户ID" json:"userId"`
	NodeID        int        `gorm:"not null;uniqueIndex:idx_user_node,priority:2;comment:关联知识节点ID" json:"nodeId"`
	Status        string     `gorm:"default:'unstarted';type:enum('unstarted','learning','completed');comment:学习状态：unstarted=未开始, learning=学习中, completed=已学习" json:"status"`
	LastStudyTime *time.Time `gorm:"comment:最后一次点击或操作的时间" json:"lastStudyTime"`
	UpdatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:状态更新时间" json:"updatedAt"`
}

// TableName UserStudyStatus 表名
func (UserStudyStatus) TableName() string {
	return "user_study_status"
}
