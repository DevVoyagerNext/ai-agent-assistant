package model

import "time"

// UserPrivateNote 私人笔记
type UserPrivateNote struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;comment:笔记主键ID" json:"id"`
	UserID      int       `gorm:"not null;index:idx_user_parent,priority:1;index:idx_user_path,priority:1;index:idx_public_lookup,priority:2;comment:所属用户ID" json:"userId"`
	ParentID    int       `gorm:"default:0;index:idx_user_parent,priority:2;comment:父节点ID，0表示根目录" json:"parentId"`
	Path        string    `gorm:"default:'0/';type:varchar(512);index:idx_user_path,priority:2,length:128;comment:层级路径，用于递归查询和权限继承" json:"path"`
	Type        string    `gorm:"not null;type:enum('folder','markdown');index:idx_type;comment:类型：folder=文件夹, markdown=文件" json:"type"`
	Title       string    `gorm:"not null;type:varchar(255);comment:文件夹名或文件名" json:"title"`
	Content     string    `gorm:"type:longtext;comment:Markdown正文内容（仅当type为markdown时有效）" json:"content"`
	IsPublic    int8       `gorm:"default:0;index:idx_public_lookup,priority:1;comment:是否公开：1=公开, 0=私密" json:"isPublic"`
	IsImportant int8       `gorm:"default:0;comment:是否标记为重要/收藏" json:"isImportant"`
	IsDeleted   int8       `gorm:"default:0;comment:逻辑删除标记：1=已删除（回收站）, 0=正常" json:"isDeleted"`
	SortOrder   int        `gorm:"default:0;comment:同层级下的显示排序" json:"sortOrder"`
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:最后修改时间" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"comment:删除时间，非空表示在回收站" json:"deletedAt"`
}

// TableName UserPrivateNote 表名
func (UserPrivateNote) TableName() string {
	return "user_private_notes"
}
