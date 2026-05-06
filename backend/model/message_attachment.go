package model

import "time"

// MessageAttachment 消息附件表
type MessageAttachment struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:附件唯一ID" json:"id"`
	SessionID  int64     `gorm:"not null;index:idx_session_id;comment:所属会话ID" json:"sessionId"`
	MessageID  int64     `gorm:"not null;index:idx_msg_id;comment:所属消息ID" json:"messageId"`
	UserID     int64     `gorm:"not null;index:idx_user_file;comment:所属用户ID" json:"userId"`
	FileKey    string    `gorm:"type:varchar(255);not null;comment:七牛云存储路径对象名 (含 Agent/ 前缀)" json:"fileKey"`
	FileName   string    `gorm:"type:varchar(255);comment:原始文件名 (方便用户下载)" json:"fileName"`
	FileType   string    `gorm:"type:varchar(50);comment:文件类型 (image, pdf, doc等)" json:"fileType"`
	FileSize   int64     `gorm:"default:0;comment:文件大小 (Byte)" json:"fileSize"`
	SenderRole string    `gorm:"type:varchar(20);not null;comment:发送者角色: user 或 assistant" json:"senderRole"`
	CreatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
}

// TableName 表名
func (MessageAttachment) TableName() string {
	return "message_attachments"
}
