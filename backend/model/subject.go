package model

import "time"

// Subject 学科表：学习学科分类（数据结构、计网、Go语言等）
type Subject struct {
	ID                uint      `gorm:"primaryKey;autoIncrement;comment:学科主键ID" json:"id"`
	CreatorID         int       `gorm:"not null;default:0;index:idx_creator;comment:所属作者/所有者ID" json:"creatorId"`
	Slug              string    `gorm:"unique;not null;type:varchar(50);comment:学科唯一标识，如data_structure" json:"slug"`
	Name              string    `gorm:"not null;type:varchar(100);comment:学科显示名称，如数据结构" json:"name"`
	NameDraft         string    `gorm:"type:varchar(100);default:'';comment:教材名称草稿" json:"nameDraft"`
	Icon              string    `gorm:"type:varchar(255);comment:学科图标CSS类名/URL地址" json:"icon"`
	Description       string    `gorm:"type:text;comment:学科简介描述" json:"description"`
	DescriptionDraft  string    `gorm:"type:text;default:NULL;comment:教材简介草稿" json:"descriptionDraft"`
	CoverImageID      int       `gorm:"default:0;comment:学科封面图片ID，关联images表" json:"coverImageId"`
	CoverImageIDDraft int       `gorm:"default:0;comment:教材封面ID草稿" json:"coverImageIdDraft"`
	Status            string    `gorm:"type:enum('draft', 'published', 'archived');default:'draft';comment:教材整体状态" json:"status"`
	AuditStatus       int8      `gorm:"default:0;index:idx_audit_status;comment:审核状态：0=编辑中, 1=待审核, 2=已通过, 3=被驳回" json:"auditStatus"`
	LastLogID         int64     `gorm:"default:0;comment:关联最新一条审批流水ID" json:"lastLogId"`
	CreatedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:学科创建时间" json:"createdAt"`
}

// TableName Subject 表名
func (Subject) TableName() string {
	return "subjects"
}
