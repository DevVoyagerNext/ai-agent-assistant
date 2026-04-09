package model

import "time"

// SubjectCategory 教材所属学科分类表
type SubjectCategory struct {
	ID        int       `gorm:"primaryKey;autoIncrement;comment:分类主键ID" json:"id"`
	Name      string    `gorm:"not null;type:varchar(50);comment:分类名称，如：计算机基础、后端开发" json:"name"`
	Slug      string    `gorm:"not null;type:varchar(50);uniqueIndex;comment:分类唯一标识，用于URL，如：cs-basics" json:"slug"`
	Icon      string    `gorm:"type:varchar(255);comment:分类图标（可以是CSS类名或图片URL）" json:"icon"`
	SortOrder int       `gorm:"default:0;comment:展示排序（数值越小越靠前）" json:"sortOrder"`
	IsActive  int8      `gorm:"default:1;type:tinyint(1);comment:是否启用：1=是, 0=否" json:"isActive"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

// TableName SubjectCategory 表名
func (SubjectCategory) TableName() string {
	return "subject_categories"
}
