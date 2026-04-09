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
	ID           uint   `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	CoverImageID int    `json:"coverImageId"`
}

type UserSubjectProgressRes struct {
	Subject         SubjectRes `json:"subject"`
	Status          string     `json:"status"`
	ProgressPercent float64    `json:"progressPercent"`
	LastNodeID      int        `json:"lastNodeId"`
	LastStudyTime   time.Time  `json:"lastStudyTime"`
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
