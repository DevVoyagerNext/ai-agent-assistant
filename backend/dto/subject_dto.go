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

type SubjectListRes struct {
	Total int64        `json:"total"`
	List  []SubjectRes `json:"list"`
}

type CollectFolderRes struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    int8      `json:"isPublic"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
