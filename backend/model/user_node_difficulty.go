package model

import "time"

// UserNodeDifficulty 用户标记知识点简单困难表
type UserNodeDifficulty struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID     int       `gorm:"not null;uniqueIndex:idx_user_node,priority:1;comment:评价用户ID" json:"userId"`
	NodeID     int       `gorm:"not null;uniqueIndex:idx_user_node,priority:2;index:idx_node_difficulty,priority:1;comment:知识节点ID" json:"nodeId"`
	Difficulty string    `gorm:"not null;type:enum('easy','medium','hard');index:idx_node_difficulty,priority:2;comment:难度评价：简单、中等、困难" json:"difficulty"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:评价时间" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:修改时间" json:"updatedAt"`
}

// TableName UserNodeDifficulty 表名
func (UserNodeDifficulty) TableName() string {
	return "user_node_difficulty"
}
