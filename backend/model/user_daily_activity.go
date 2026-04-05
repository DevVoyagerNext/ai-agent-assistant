package model

import "time"

// UserDailyActivity 统计用户活跃度
type UserDailyActivity struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement;comment:自增主键ID" json:"id"`
	UserID        int       `gorm:"not null;uniqueIndex:uk_user_date,priority:1;comment:用户ID，关联users表" json:"userId"`
	ActivityDate  time.Time `gorm:"type:date;not null;uniqueIndex:uk_user_date,priority:2;index:idx_date;comment:活跃日期（如 2024-05-20）" json:"activityDate"`
	ActivityCount int       `gorm:"default:0;comment:当日操作总次数" json:"activityCount"`
	ActivityScore int       `gorm:"default:0;comment:当日活跃总分（用于区分行为权重）" json:"activityScore"`
	LoginCount    int8      `gorm:"default:0;comment:当日登录次数" json:"loginCount"`
	StudyCount    int8      `gorm:"default:0;comment:当日学习节点数" json:"studyCount"`
	NoteCount     int8      `gorm:"default:0;comment:当日新增笔记数" json:"noteCount"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:最后更新时间" json:"updatedAt"`
}

// TableName UserDailyActivity 表名
func (UserDailyActivity) TableName() string {
	return "user_daily_activities"
}
