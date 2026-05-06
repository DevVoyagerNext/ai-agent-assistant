package model

import "time"

// File 文件管理表：统一管理系统所有文件资源
type File struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:文件主键ID" json:"id"`
	FileName  string    `gorm:"not null;type:varchar(255);comment:文件原始文件名" json:"fileName"`
	FilePath  string    `gorm:"not null;type:varchar(512);comment:文件存储路径/OSS链接" json:"filePath"`
	FileType  string    `gorm:"not null;type:varchar(20);comment:文件格式：png/jpg/jpeg/webp/gif等" json:"fileType"`
	FileSize  int       `gorm:"not null;comment:文件大小，单位：字节" json:"fileSize"`
	Alt       string    `gorm:"type:varchar(255);comment:文件替代文本（SEO/加载失败显示）" json:"alt"`
	SubjectID int       `gorm:"default:0;comment:关联学科ID，0=不关联" json:"subjectId"`
	NodeID    int       `gorm:"default:0;comment:关联知识节点ID，0=不关联" json:"nodeId"`
	UserID    int       `gorm:"default:0;comment:关联用户ID，0=不关联" json:"userId"`
	SortOrder int       `gorm:"default:0;comment:文件排序序号" json:"sortOrder"`
	Status    int8      `gorm:"default:1;comment:文件状态：1=启用，0=禁用" json:"status"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:文件上传时间" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:文件更新时间" json:"updatedAt"`
}

// TableName File 表名
func (File) TableName() string {
	return "ai_study_assistant.files"
}
