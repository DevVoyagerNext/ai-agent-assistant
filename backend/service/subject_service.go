package service

import (
	"backend/dao"
	"backend/dto"
	"backend/pkg/errmsg"
	"context"
	"errors"
	"gorm.io/gorm"
)

type SubjectService struct {
	subjectDao dao.SubjectDao
}

func (s *SubjectService) GetCategories(ctx context.Context) ([]dto.CategoryRes, int) {
	categories, err := s.subjectDao.GetCategories()
	if err != nil {
		return nil, errmsg.CodeError
	}

	var res []dto.CategoryRes
	for _, c := range categories {
		res = append(res, dto.ConvertCategoryToRes(&c))
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetSubjectsByCategoryID(ctx context.Context, categoryId int) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetSubjectsByCategoryID(categoryId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	var res []dto.SubjectRes
	for _, sub := range subjects {
		res = append(res, dto.ConvertSubjectToRes(&sub))
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetAllSubjects(ctx context.Context) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetAllSubjects()
	if err != nil {
		return nil, errmsg.CodeError
	}

	var res []dto.SubjectRes
	for _, sub := range subjects {
		res = append(res, dto.ConvertSubjectToRes(&sub))
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserCollectedSubjects(ctx context.Context, userId uint) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetUserCollectedSubjects(userId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	var res []dto.SubjectRes
	for _, sub := range subjects {
		res = append(res, dto.ConvertSubjectToRes(&sub))
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserLikedSubjects(ctx context.Context, userId uint) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetUserLikedSubjects(userId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	var res []dto.SubjectRes
	for _, sub := range subjects {
		res = append(res, dto.ConvertSubjectToRes(&sub))
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserSubjectsByStatus(ctx context.Context, userId uint, status string) ([]dto.UserSubjectProgressRes, int) {
	progresses, err := s.subjectDao.GetUserSubjectsByStatus(userId, status)
	if err != nil {
		return nil, errmsg.CodeError
	}

	if len(progresses) == 0 {
		return []dto.UserSubjectProgressRes{}, errmsg.CodeSuccess
	}

	var subjectIds []int
	for _, p := range progresses {
		subjectIds = append(subjectIds, p.SubjectID)
	}

	subjects, err := s.subjectDao.GetSubjectsByIds(subjectIds)
	if err != nil {
		return nil, errmsg.CodeError
	}

	subjectMap := make(map[int]dto.SubjectRes)
	for _, sub := range subjects {
		subjectMap[int(sub.ID)] = dto.ConvertSubjectToRes(&sub)
	}

	var res []dto.UserSubjectProgressRes
	for _, p := range progresses {
		if sub, ok := subjectMap[p.SubjectID]; ok {
			res = append(res, dto.UserSubjectProgressRes{
				Subject:         sub,
				Status:          p.Status,
				ProgressPercent: p.ProgressPercent,
				LastNodeID:      p.LastNodeID,
				LastStudyTime:   p.LastStudyTime,
			})
		}
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserLastLearningSubject(ctx context.Context, userId uint) (*dto.UserSubjectProgressRes, int) {
	progress, err := s.subjectDao.GetUserLastLearningSubject(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errmsg.CodeSuccess
		}
		return nil, errmsg.CodeError
	}

	subject, err := s.subjectDao.GetSubjectById(progress.SubjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errmsg.CodeSuccess
		}
		return nil, errmsg.CodeError
	}

	res := &dto.UserSubjectProgressRes{
		Subject:         dto.ConvertSubjectToRes(subject),
		Status:          progress.Status,
		ProgressPercent: progress.ProgressPercent,
		LastNodeID:      progress.LastNodeID,
		LastStudyTime:   progress.LastStudyTime,
	}

	return res, errmsg.CodeSuccess
}
