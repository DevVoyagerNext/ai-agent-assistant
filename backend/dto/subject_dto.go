package dto

import (
	"backend/model"
	"time"
)

type CategoryRes struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sortOrder"`
}

type SubjectRes struct {
	ID              uint    `json:"id"`
	Slug            string  `json:"slug"`
	Name            string  `json:"name"`
	Icon            string  `json:"icon"`
	Description     string  `json:"description"`
	CoverImageID    int     `json:"coverImageId"`
	IsLiked         bool    `json:"isLiked"`
	IsCollected     bool    `json:"isCollected"`
	ProgressPercent float64 `json:"progressPercent"`
	LastNodeID      int     `json:"lastNodeId"`
}

type UserSubjectProgressRes struct {
	Subject         SubjectRes `json:"subject"`
	Status          string     `json:"status"`
	IsLiked         bool       `json:"isLiked"`
	IsCollected     bool       `json:"isCollected"`
	ProgressPercent float64    `json:"progressPercent"`
	LastNodeID      int        `json:"lastNodeId"`
	LastStudyTime   time.Time  `json:"lastStudyTime"`
}

type RecentSubjectListRes struct {
	Total int64                    `json:"total"`
	List  []UserSubjectProgressRes `json:"list"`
}

type SubjectSearchReq struct {
	Keyword  string `form:"keyword" binding:"required,min=1,max=50"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

type UserPaginationReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

type SubjectListRes struct {
	Total int64        `json:"total"`
	List  []SubjectRes `json:"list"`
}

// UserCreatedSubjectRes 用户创建的教材返回结构，包含草稿和统计信息
type UserCreatedSubjectRes struct {
	ID                uint      `json:"id"`
	Slug              string    `json:"slug"`
	Name              string    `json:"name"`
	NameDraft         string    `json:"nameDraft"`
	Icon              string    `json:"icon"`
	Description       string    `json:"description"`
	DescriptionDraft  string    `json:"descriptionDraft"`
	CoverImageID      int       `json:"coverImageId"`
	CoverImageIDDraft int       `json:"coverImageIdDraft"`
	Status            string    `json:"status"`
	AuditStatus       int8      `json:"auditStatus"`
	HasDraft          int8      `json:"hasDraft"`
	CreatedAt         time.Time `json:"createdAt"`
	LikeCount         int64     `json:"likeCount"`    // 点赞总数
	CollectCount      int64     `json:"collectCount"` // 收藏总数
}

type UserCreatedSubjectListRes struct {
	Total int64                   `json:"total"`
	List  []UserCreatedSubjectRes `json:"list"`
}

type CollectFolderRes struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    int8      `json:"isPublic"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateCollectFolderReq struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	IsPublic    *int8  `json:"isPublic" binding:"required,oneof=0 1"`
}

type AddSubjectToFolderReq struct {
	SubjectID int `json:"subjectId" binding:"required,gt=0"`
}

type UpdateCollectFolderPublicReq struct {
	IsPublic int8 `json:"isPublic" binding:"oneof=0 1"`
}

type RenameCollectFolderReq struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

func ConvertSubjectToRes(s *model.Subject) SubjectRes {
	return SubjectRes{
		ID:           s.ID,
		Slug:         s.Slug,
		Name:         s.Name,
		Icon:         s.Icon,
		Description:  s.Description,
		CoverImageID: s.CoverImageID,
	}
}

func ConvertCategoryToRes(c *model.SubjectCategory) CategoryRes {
	return CategoryRes{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		Icon:      c.Icon,
		SortOrder: c.SortOrder,
	}
}

func ConvertCollectFolderToRes(f *model.UserCollectFolder) CollectFolderRes {
	return CollectFolderRes{
		ID:          f.ID,
		Name:        f.Name,
		Description: f.Description,
		IsPublic:    f.IsPublic,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}
