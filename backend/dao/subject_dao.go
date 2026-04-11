package dao

import (
	"backend/global"
	"backend/model"
)

type SubjectDao struct{}

func (d *SubjectDao) GetCategories() ([]model.SubjectCategory, error) {
	var categories []model.SubjectCategory
	err := global.GVA_DB.Where("is_active = ?", 1).Order("sort_order asc").Find(&categories).Error
	return categories, err
}

func (d *SubjectDao) GetSubjectsByCategoryID(categoryId int) ([]model.Subject, error) {
	var subjects []model.Subject
	err := global.GVA_DB.
		Joins("JOIN subject_category_rel ON subject_category_rel.subject_id = subjects.id").
		Where("subject_category_rel.category_id = ?", categoryId).
		Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) GetAllSubjects() ([]model.Subject, error) {
	var subjects []model.Subject
	err := global.GVA_DB.Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) SearchSubjectsByName(keyword string, page int, pageSize int) ([]model.Subject, int64, error) {
	query := global.GVA_DB.Model(&model.Subject{}).Where("name LIKE ?", "%"+keyword+"%")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var subjects []model.Subject
	err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&subjects).Error
	return subjects, total, err
}

func (d *SubjectDao) GetUserCollectedSubjects(userId uint) ([]model.Subject, error) {
	var subjects []model.Subject
	err := global.GVA_DB.
		Joins("JOIN user_collect_items ON user_collect_items.subject_id = subjects.id").
		Where("user_collect_items.user_id = ?", userId).
		Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) GetUserLikedSubjects(userId uint) ([]model.Subject, error) {
	var subjects []model.Subject
	err := global.GVA_DB.
		Joins("JOIN user_subject_likes ON user_subject_likes.subject_id = subjects.id").
		Where("user_subject_likes.user_id = ?", userId).
		Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) GetUserCollectFolders(userId uint) ([]model.UserCollectFolder, error) {
	var folders []model.UserCollectFolder
	err := global.GVA_DB.
		Where("user_id = ?", userId).
		Order("created_at desc").
		Find(&folders).Error
	return folders, err
}

func (d *SubjectDao) GetUserCollectedSubjectsByFolder(userId uint, folderId int) ([]model.Subject, error) {
	var subjects []model.Subject
	err := global.GVA_DB.
		Distinct("subjects.*").
		Joins("JOIN user_collect_items ON user_collect_items.subject_id = subjects.id").
		Where("user_collect_items.user_id = ? AND user_collect_items.folder_id = ?", userId, folderId).
		Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) GetUserRecentSubjectProgress(userId uint, page int, pageSize int) ([]model.UserSubjectProgress, int64, error) {
	var total int64
	if err := global.GVA_DB.Model(&model.UserSubjectProgress{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var progresses []model.UserSubjectProgress
	err := global.GVA_DB.
		Where("user_id = ?", userId).
		Order("last_study_time desc").
		Offset(offset).
		Limit(pageSize).
		Find(&progresses).Error
	return progresses, total, err
}

func (d *SubjectDao) GetUserSubjectsByStatus(userId uint, status string) ([]model.UserSubjectProgress, error) {
	var progresses []model.UserSubjectProgress
	err := global.GVA_DB.
		Where("user_id = ? AND status = ?", userId, status).
		Order("last_study_time desc").
		Find(&progresses).Error
	return progresses, err
}

func (d *SubjectDao) GetUserLastLearningSubject(userId uint) (*model.UserSubjectProgress, error) {
	var progress model.UserSubjectProgress
	err := global.GVA_DB.
		Where("user_id = ? AND status = ?", userId, "learning").
		Order("last_study_time desc").
		First(&progress).Error
	return &progress, err
}

func (d *SubjectDao) GetSubjectsByIds(ids []int) ([]model.Subject, error) {
	var subjects []model.Subject
	if len(ids) == 0 {
		return subjects, nil
	}
	err := global.GVA_DB.Where("id IN ?", ids).Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) GetSubjectById(id int) (*model.Subject, error) {
	var subject model.Subject
	err := global.GVA_DB.Where("id = ?", id).First(&subject).Error
	return &subject, err
}

func (d *SubjectDao) GetUserSubjectInteractions(userId uint, subjectIds []uint) (map[uint]bool, map[uint]bool, map[uint]model.UserSubjectProgress, error) {
	likedMap := make(map[uint]bool)
	collectedMap := make(map[uint]bool)
	progressMap := make(map[uint]model.UserSubjectProgress)

	if userId == 0 || len(subjectIds) == 0 {
		return likedMap, collectedMap, progressMap, nil
	}

	var likes []model.UserSubjectLike
	if err := global.GVA_DB.Where("user_id = ? AND subject_id IN ?", userId, subjectIds).Find(&likes).Error; err != nil {
		return nil, nil, nil, err
	}
	for _, l := range likes {
		likedMap[uint(l.SubjectID)] = true
	}

	var collects []model.UserCollectItem
	if err := global.GVA_DB.Where("user_id = ? AND subject_id IN ?", userId, subjectIds).Find(&collects).Error; err != nil {
		return nil, nil, nil, err
	}
	for _, c := range collects {
		collectedMap[uint(c.SubjectID)] = true
	}

	var progresses []model.UserSubjectProgress
	if err := global.GVA_DB.Where("user_id = ? AND subject_id IN ?", userId, subjectIds).Find(&progresses).Error; err != nil {
		return nil, nil, nil, err
	}
	for _, p := range progresses {
		progressMap[uint(p.SubjectID)] = p
	}

	return likedMap, collectedMap, progressMap, nil
}

func (d *SubjectDao) GetSubjectLike(userId uint, subjectId int) (*model.UserSubjectLike, error) {
	var like model.UserSubjectLike
	err := global.GVA_DB.Where("user_id = ? AND subject_id = ?", userId, subjectId).First(&like).Error
	return &like, err
}

func (d *SubjectDao) CreateSubjectLike(userId uint, subjectId int) error {
	like := model.UserSubjectLike{
		UserID:    int(userId),
		SubjectID: subjectId,
	}
	return global.GVA_DB.Create(&like).Error
}

func (d *SubjectDao) DeleteSubjectLike(userId uint, subjectId int) error {
	return global.GVA_DB.Where("user_id = ? AND subject_id = ?", userId, subjectId).Delete(&model.UserSubjectLike{}).Error
}
