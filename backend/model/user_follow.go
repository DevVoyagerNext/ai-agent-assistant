package model

import "time"

// UserFollow 关注列表
type UserFollow struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	FollowerID  int       `gorm:"not null;uniqueIndex:uk_follower_following,priority:1;comment:关注者ID（动作发起者）" json:"followerId"`
	FollowingID int       `gorm:"not null;uniqueIndex:uk_follower_following,priority:2;index:idx_following_id;comment:被关注者ID（目标人物）" json:"followingId"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:关注时间" json:"createdAt"`
}

// TableName UserFollow 表名
func (UserFollow) TableName() string {
	return "user_follows"
}
