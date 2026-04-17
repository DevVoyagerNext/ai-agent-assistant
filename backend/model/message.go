package model

import "time"

// Message 消息明细表
type Message struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	SessionID  int64     `gorm:"index:idx_session_time,priority:1;comment:所属会话ID" json:"sessionId"`
	ParentID   *int64    `gorm:"default:NULL;comment:父消息ID(可选)" json:"parentId"`
	Role       string    `gorm:"type:varchar(20);comment:消息角色(如user/assistant)" json:"role"`
	Content    string    `gorm:"type:longtext;comment:消息内容(支持长文本)" json:"content"`
	Status     string    `gorm:"type:varchar(20);default:'active';comment:消息状态" json:"status"`
	TokensUsed int       `gorm:"default:0;comment:使用的Token数量" json:"tokensUsed"`
	CreatedAt  time.Time `gorm:"type:timestamp(3);default:CURRENT_TIMESTAMP(3);index:idx_session_time,priority:2;comment:创建时间" json:"createdAt"`
}

// TableName 表名
func (Message) TableName() string {
	return "messages"
}
