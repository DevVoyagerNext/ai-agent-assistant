package model

import "time"

// Subject 学科表：学习学科分类（数据结构、计网、Go语言等）
type Subject struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;comment:学科主键ID" json:"id"`
	Slug         string    `gorm:"unique;not null;type:varchar(50);comment:学科唯一标识，如data_structure" json:"slug"`
	Name         string    `gorm:"not null;type:varchar(100);comment:学科显示名称，如数据结构" json:"name"`
	Icon         string    `gorm:"type:varchar(255);comment:学科图标CSS类名/URL地址" json:"icon"`
	Description  string    `gorm:"type:text;comment:学科简介描述" json:"description"`
	CoverImageID int       `gorm:"default:0;comment:学科封面图片ID，关联images表" json:"coverImageId"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:学科创建时间" json:"createdAt"`
}

// TableName Subject 表名
func (Subject) TableName() string {
	return "subjects"
}
