package service

import (
	"backend/dao"
	"backend/dto"
	"backend/model"
	"backend/pkg/errmsg"
	"context"
	"errors"

	"gorm.io/gorm"
)

type SubjectService struct {
	subjectDao dao.SubjectDao
}

func (s *SubjectService) enrichSubjectList(userId uint, subjects []model.Subject) ([]dto.SubjectRes, int) {
	var res []dto.SubjectRes
	if len(subjects) == 0 {
		return res, errmsg.CodeSuccess
	}

	var subjectIds []uint
	for _, sub := range subjects {
		subjectIds = append(subjectIds, sub.ID)
	}

	likedMap, collectedMap, progressMap, err := s.subjectDao.GetUserSubjectInteractions(userId, subjectIds)
	if err != nil {
		return nil, errmsg.CodeError
	}

	for _, sub := range subjects {
		sr := dto.ConvertSubjectToRes(&sub)
		sr.IsLiked = likedMap[sub.ID]
		sr.IsCollected = collectedMap[sub.ID]
		if p, ok := progressMap[sub.ID]; ok {
			sr.ProgressPercent = p.ProgressPercent
			sr.LastNodeID = p.LastNodeID
		}
		res = append(res, sr)
	}
	return res, errmsg.CodeSuccess
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

func (s *SubjectService) GetSubjectsByCategoryID(ctx context.Context, categoryId int, userId uint) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetSubjectsByCategoryID(categoryId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	return s.enrichSubjectList(userId, subjects)
}

func (s *SubjectService) GetAllSubjects(ctx context.Context, userId uint) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetAllSubjects()
	if err != nil {
		return nil, errmsg.CodeError
	}

	return s.enrichSubjectList(userId, subjects)
}

func (s *SubjectService) GetSubjectByID(ctx context.Context, subjectId int, userId uint) (*dto.SubjectRes, int) {
	subject, err := s.subjectDao.GetSubjectById(subjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errmsg.CodeSuccess // 或者返回特定的“未找到”错误码
		}
		return nil, errmsg.CodeError
	}

	enriched, code := s.enrichSubjectList(userId, []model.Subject{*subject})
	if code != errmsg.CodeSuccess || len(enriched) == 0 {
		return nil, code
	}

	return &enriched[0], errmsg.CodeSuccess
}

func (s *SubjectService) SearchSubjects(ctx context.Context, keyword string, userId uint, page int, pageSize int) (dto.SubjectListRes, int) {
	subjects, total, err := s.subjectDao.SearchSubjectsByName(keyword, page, pageSize)
	if err != nil {
		return dto.SubjectListRes{}, errmsg.CodeError
	}

	enriched, code := s.enrichSubjectList(userId, subjects)
	if code != errmsg.CodeSuccess {
		return dto.SubjectListRes{}, code
	}

	return dto.SubjectListRes{Total: total, List: enriched}, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserCollectedSubjects(ctx context.Context, userId uint) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetUserCollectedSubjects(userId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	return s.enrichSubjectList(userId, subjects)
}

func (s *SubjectService) GetUserLikedSubjects(ctx context.Context, userId uint) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetUserLikedSubjects(userId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	return s.enrichSubjectList(userId, subjects)
}

func (s *SubjectService) GetUserCollectFolders(ctx context.Context, userId uint) ([]dto.CollectFolderRes, int) {
	folders, err := s.subjectDao.GetUserCollectFolders(userId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	var res []dto.CollectFolderRes
	for _, f := range folders {
		res = append(res, dto.ConvertCollectFolderToRes(&f))
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserCollectedSubjectsByFolder(ctx context.Context, userId uint, folderId int) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetUserCollectedSubjectsByFolder(userId, folderId)
	if err != nil {
		return nil, errmsg.CodeError
	}
	return s.enrichSubjectList(userId, subjects)
}

func (s *SubjectService) GetUserRecentSubjects(ctx context.Context, userId uint, page int, pageSize int) (dto.RecentSubjectListRes, int) {
	progresses, total, err := s.subjectDao.GetUserRecentSubjectProgress(userId, page, pageSize)
	if err != nil {
		return dto.RecentSubjectListRes{}, errmsg.CodeError
	}

	if len(progresses) == 0 {
		return dto.RecentSubjectListRes{Total: total, List: []dto.UserSubjectProgressRes{}}, errmsg.CodeSuccess
	}

	var subjectIds []int
	for _, p := range progresses {
		subjectIds = append(subjectIds, p.SubjectID)
	}

	subjects, err := s.subjectDao.GetSubjectsByIds(subjectIds)
	if err != nil {
		return dto.RecentSubjectListRes{}, errmsg.CodeError
	}

	enrichedSubjects, code := s.enrichSubjectList(userId, subjects)
	if code != errmsg.CodeSuccess {
		return dto.RecentSubjectListRes{}, code
	}

	subjectMap := make(map[uint]dto.SubjectRes)
	for _, sub := range enrichedSubjects {
		subjectMap[sub.ID] = sub
	}

	var list []dto.UserSubjectProgressRes
	for _, p := range progresses {
		if sub, ok := subjectMap[uint(p.SubjectID)]; ok {
			list = append(list, dto.UserSubjectProgressRes{
				Subject:         sub,
				Status:          p.Status,
				IsLiked:         sub.IsLiked,
				IsCollected:     sub.IsCollected,
				ProgressPercent: p.ProgressPercent,
				LastNodeID:      p.LastNodeID,
				LastStudyTime:   p.LastStudyTime,
			})
		}
	}

	return dto.RecentSubjectListRes{Total: total, List: list}, errmsg.CodeSuccess
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

	enrichedSubjects, code := s.enrichSubjectList(userId, subjects)
	if code != errmsg.CodeSuccess {
		return nil, code
	}

	subjectMap := make(map[uint]dto.SubjectRes)
	for _, sub := range enrichedSubjects {
		subjectMap[sub.ID] = sub
	}

	var res []dto.UserSubjectProgressRes
	for _, p := range progresses {
		if sub, ok := subjectMap[uint(p.SubjectID)]; ok {
			res = append(res, dto.UserSubjectProgressRes{
				Subject:         sub,
				Status:          p.Status,
				IsLiked:         sub.IsLiked,
				IsCollected:     sub.IsCollected,
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

	enrichedSubjects, code := s.enrichSubjectList(userId, []model.Subject{*subject})
	if code != errmsg.CodeSuccess || len(enrichedSubjects) == 0 {
		return nil, errmsg.CodeError
	}

	res := &dto.UserSubjectProgressRes{
		Subject:         enrichedSubjects[0],
		Status:          progress.Status,
		IsLiked:         enrichedSubjects[0].IsLiked,
		IsCollected:     enrichedSubjects[0].IsCollected,
		ProgressPercent: progress.ProgressPercent,
		LastNodeID:      progress.LastNodeID,
		LastStudyTime:   progress.LastStudyTime,
	}

	return res, errmsg.CodeSuccess
}

func (s *SubjectService) ToggleSubjectLike(ctx context.Context, userId uint, subjectId int) (bool, int) {
	// 1. 检查是否已经点赞
	_, err := s.subjectDao.GetSubjectLike(userId, subjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 2. 未点赞，执行点赞
			if err := s.subjectDao.CreateSubjectLike(userId, subjectId); err != nil {
				return false, errmsg.CodeError
			}
			return true, errmsg.CodeSuccess // true 表示当前状态为已点赞
		}
		return false, errmsg.CodeError
	}

	// 3. 已点赞，取消点赞
	if err := s.subjectDao.DeleteSubjectLike(userId, subjectId); err != nil {
		return false, errmsg.CodeError
	}
	return false, errmsg.CodeSuccess // false 表示当前状态为未点赞
}

func (s *SubjectService) CreateCollectFolder(ctx context.Context, userId uint, req dto.CreateCollectFolderReq) (*dto.CollectFolderRes, int) {
	folder, err := s.subjectDao.CreateCollectFolder(userId, req.Name, req.Description, *req.IsPublic)
	if err != nil {
		return nil, errmsg.CodeError
	}
	return &dto.CollectFolderRes{
		ID:          int(folder.ID),
		Name:        folder.Name,
		Description: folder.Description,
		IsPublic:    folder.IsPublic,
		CreatedAt:   folder.CreatedAt,
		UpdatedAt:   folder.UpdatedAt,
	}, errmsg.CodeSuccess
}

func (s *SubjectService) AddSubjectToFolder(ctx context.Context, userId uint, folderId int, subjectId int) int {
	// 1. 检查收藏夹是否存在且属于该用户
	_, err := s.subjectDao.GetCollectFolderById(userId, folderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errmsg.CodeError // 或者定义专门的收藏夹不存在错误码
		}
		return errmsg.CodeError
	}

	// 2. 检查是否已经收藏过该教材
	exists, err := s.subjectDao.CheckSubjectInFolder(userId, folderId, subjectId)
	if err != nil {
		return errmsg.CodeError
	}
	if exists {
		return errmsg.CodeSuccess // 已经存在，幂等处理
	}

	// 3. 添加到收藏夹
	if err := s.subjectDao.AddSubjectToFolder(userId, folderId, subjectId); err != nil {
		return errmsg.CodeError
	}
	return errmsg.CodeSuccess
}

func (s *SubjectService) RemoveSubjectFromFolder(ctx context.Context, userId uint, folderId int, subjectId int) int {
	// 1. 检查收藏夹是否存在且属于该用户
	_, err := s.subjectDao.GetCollectFolderById(userId, folderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errmsg.CodeError
		}
		return errmsg.CodeError
	}

	// 2. 删除收藏记录
	if err := s.subjectDao.DeleteSubjectFromFolder(userId, folderId, subjectId); err != nil {
		return errmsg.CodeError
	}
	return errmsg.CodeSuccess
}
