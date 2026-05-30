package model

import "time"

type Subject struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatorID         int       `gorm:"not null;default:0;index:idx_creator" json:"creatorId"`
	Slug              string    `gorm:"unique;not null;type:varchar(50)" json:"slug"`
	Name              string    `gorm:"not null;type:varchar(100)" json:"name"`
	NameDraft         string    `gorm:"type:varchar(100);default:''" json:"nameDraft"`
	Icon              string    `gorm:"type:varchar(255)" json:"icon"`
	IconDraft         string    `gorm:"type:varchar(255);default:NULL" json:"iconDraft"`
	Description       string    `gorm:"type:text" json:"description"`
	DescriptionDraft  string    `gorm:"type:text;default:NULL" json:"descriptionDraft"`
	CoverImageID      int       `gorm:"column:cover_image_id;default:0" json:"coverImageId"`
	CoverImageIDDraft int       `gorm:"column:cover_image_id_draft;default:0" json:"coverImageIdDraft"`
	Status            string    `gorm:"type:enum('draft','published','archived');default:'draft'" json:"status"`
	AuditStatus       int8      `gorm:"default:0;index:idx_audit_status" json:"auditStatus"`
	LastLogID         int64     `gorm:"column:last_log_id;default:0" json:"lastLogId"`
	HasDraft          int8      `gorm:"type:tinyint(1);default:0" json:"hasDraft"`
	CreatedAt         time.Time `json:"createdAt"`
}

func (Subject) TableName() string {
	return "subjects"
}
