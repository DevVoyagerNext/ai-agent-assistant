package model

import "time"

// NoteShare 笔记分享表：管理用户笔记的分享、访问权限
type NoteShare struct {
	ID            uint      `gorm:"primaryKey;autoIncrement;comment:分享记录主键ID" json:"id"`
	UserID        int       `gorm:"not null;index:idx_user_private_note,priority:1;comment:分享者用户ID" json:"userId"`
	PrivateNoteID int       `gorm:"not null;index:idx_user_private_note,priority:2;comment:关联 user_private_notes 表的 ID" json:"privateNoteId"`
	NoteType      string    `gorm:"not null;type:enum('folder','markdown');comment:分享时的类型：文件夹或文件" json:"noteType"`
	ShareToken    string    `gorm:"unique;not null;type:varchar(64);index:idx_share_token;comment:分享链接唯一标识" json:"shareToken"`
	ShareCode     string    `gorm:"not null;type:char(4);comment:4位访问提取码" json:"shareCode"`
	ExpiresAt     time.Time `gorm:"not null;comment:失效时间" json:"expiresAt"`
	ViewCount     int       `gorm:"default:0;comment:访问次数" json:"viewCount"`
	IsActive      int8      `gorm:"default:1;comment:状态：1=有效，0=已取消" json:"isActive"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:分享创建时间" json:"createdAt"`
}

// TableName NoteShare 表名
func (NoteShare) TableName() string {
	return "note_shares"
}
