package model

import "time"

// Image 图片管理表：统一管理系统所有图片资源
type Image struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:图片主键ID" json:"id"`
	FileName  string    `gorm:"not null;type:varchar(255);comment:图片原始文件名" json:"fileName"`
	FilePath  string    `gorm:"not null;type:varchar(512);comment:图片存储路径/OSS链接" json:"filePath"`
	FileType  string    `gorm:"not null;type:varchar(20);comment:图片格式：png/jpg/jpeg/webp/gif" json:"fileType"`
	FileSize  int       `gorm:"not null;comment:图片大小，单位：字节" json:"fileSize"`
	Alt       string    `gorm:"type:varchar(255);comment:图片替代文本（SEO/加载失败显示）" json:"alt"`
	SubjectID int       `gorm:"default:0;comment:关联学科ID，0=不关联" json:"subjectId"`
	NodeID    int       `gorm:"default:0;comment:关联知识节点ID，0=不关联" json:"nodeId"`
	UserID    int       `gorm:"default:0;comment:关联用户ID，0=不关联" json:"userId"`
	SortOrder int       `gorm:"default:0;comment:图片排序序号" json:"sortOrder"`
	Status    int8      `gorm:"default:1;comment:图片状态：1=启用，0=禁用" json:"status"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:图片上传时间" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:图片更新时间" json:"updatedAt"`
}

// TableName Image 表名
func (Image) TableName() string {
	return "images"
}
