package model

import "time"

// SubjectCategoryRel 教材与分类的多对多关联表
type SubjectCategoryRel struct {
	ID         int       `gorm:"primaryKey;autoIncrement;comment:自增主键" json:"id"`
	SubjectID  int       `gorm:"not null;uniqueIndex:uk_subject_category,priority:1;comment:关联教材ID (subjects.id)" json:"subjectId"`
	CategoryID int       `gorm:"not null;uniqueIndex:uk_subject_category,priority:2;comment:关联分类ID (subject_categories.id)" json:"categoryId"`
	CreatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;comment:关联时间" json:"createdAt"`
}

// TableName SubjectCategoryRel 表名
func (SubjectCategoryRel) TableName() string {
	return "subject_category_rel"
}
