package model

import "time"

// UserDailyActionStat 用户每日分类行为统计表
type UserDailyActionStat struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID       int       `gorm:"not null;uniqueIndex:uk_user_date_action_target,priority:1;index:idx_user_date,priority:1;comment:用户ID" json:"userId"`
	ActivityDate time.Time `gorm:"type:date;not null;uniqueIndex:uk_user_date_action_target,priority:2;index:idx_user_date,priority:2;index:idx_date_search,priority:1;comment:统计日期" json:"activityDate"`
	ActionType   string    `gorm:"type:varchar(50);not null;uniqueIndex:uk_user_date_action_target,priority:3;index:idx_date_search,priority:2;comment:行为类型标识" json:"actionType"`
	TargetType   string    `gorm:"type:varchar(50);not null;default:'';uniqueIndex:uk_user_date_action_target,priority:4;index:idx_date_search,priority:3;comment:对象类型标识" json:"targetType"`
	ActionCount  int       `gorm:"default:0;comment:该用户当日在该类型下的触发次数" json:"actionCount"`
	ActionScore  int       `gorm:"default:0;comment:该用户当日该项累积的总得分" json:"actionScore"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:统计更新时间" json:"updatedAt"`
}

// TableName 表名
func (UserDailyActionStat) TableName() string {
	return "user_daily_action_stats"
}
