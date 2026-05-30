package dto

import "time"

type ReviewRequest struct {
	Remark string `json:"remark"`
}

type SubjectReviewItem struct {
	ID                uint      `json:"id" gorm:"column:id"`
	CreatorID         int       `json:"creatorId" gorm:"column:creator_id"`
	CreatorName       string    `json:"creatorName" gorm:"column:creator_name"`
	CreatorEmail      string    `json:"creatorEmail" gorm:"column:creator_email"`
	Slug              string    `json:"slug" gorm:"column:slug"`
	Name              string    `json:"name" gorm:"column:name"`
	NameDraft         string    `json:"nameDraft" gorm:"column:name_draft"`
	Icon              string    `json:"icon" gorm:"column:icon"`
	IconDraft         string    `json:"iconDraft" gorm:"column:icon_draft"`
	Description       string    `json:"description" gorm:"column:description"`
	DescriptionDraft  string    `json:"descriptionDraft" gorm:"column:description_draft"`
	CoverImageID      int       `json:"coverImageId" gorm:"column:cover_image_id"`
	CoverImageIDDraft int       `json:"coverImageIdDraft" gorm:"column:cover_image_id_draft"`
	Status            string    `json:"status" gorm:"column:status"`
	AuditStatus       int8      `json:"auditStatus" gorm:"column:audit_status"`
	HasDraft          int8      `json:"hasDraft" gorm:"column:has_draft"`
	CreatedAt         time.Time `json:"createdAt" gorm:"column:created_at"`
}

type NodeReviewItem struct {
	ID                   uint      `json:"id" gorm:"column:id"`
	SubjectID            int       `json:"subjectId" gorm:"column:subject_id"`
	SubjectName          string    `json:"subjectName" gorm:"column:subject_name"`
	CreatorID            int       `json:"creatorId" gorm:"column:creator_id"`
	CreatorName          string    `json:"creatorName" gorm:"column:creator_name"`
	CreatorEmail         string    `json:"creatorEmail" gorm:"column:creator_email"`
	ParentID             int       `json:"parentId" gorm:"column:parent_id"`
	Path                 string    `json:"path" gorm:"column:path"`
	Name                 string    `json:"name" gorm:"column:name"`
	NameDraft            string    `json:"nameDraft" gorm:"column:name_draft"`
	Status               string    `json:"status" gorm:"column:status"`
	AuditStatus          int8      `json:"auditStatus" gorm:"column:audit_status"`
	HasDraft             int8      `json:"hasDraft" gorm:"column:has_draft"`
	DraftDescendantCount int       `json:"draftDescendantCount" gorm:"column:draft_descendant_count"`
	Level                int8      `json:"level" gorm:"column:level"`
	IsLeaf               int8      `json:"isLeaf" gorm:"column:is_leaf"`
	Content              string    `json:"content" gorm:"column:content"`
	ContentDraft         string    `json:"contentDraft" gorm:"column:content_draft"`
	UpdatedAt            time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

type AuditRecord struct {
	ID               int64     `json:"id" gorm:"column:id"`
	TargetType       string    `json:"targetType" gorm:"column:target_type"`
	TargetID         int       `json:"targetId" gorm:"column:target_id"`
	TargetName       string    `json:"targetName" gorm:"column:target_name"`
	SubjectName      string    `json:"subjectName" gorm:"column:subject_name"`
	SubjectDraftName string    `json:"subjectDraftName" gorm:"column:subject_draft_name"`
	AdminID          int       `json:"adminId" gorm:"column:admin_id"`
	AdminName        string    `json:"adminName" gorm:"column:admin_name"`
	Action           string    `json:"action" gorm:"column:action"`
	Remark           string    `json:"remark" gorm:"column:remark"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at"`
}
