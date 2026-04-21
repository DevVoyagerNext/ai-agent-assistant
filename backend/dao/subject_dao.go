package dao

import (
	"backend/global"
	"backend/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type SubjectDao struct{}

func (d *SubjectDao) GetCategories(ctx context.Context) ([]model.SubjectCategory, error) {
	var categories []model.SubjectCategory
	err := global.GVA_DB.WithContext(ctx).Where("is_active = ?", 1).Order("sort_order asc").Find(&categories).Error
	return categories, err
}

func (d *SubjectDao) GetSubjectsByCategoryID(ctx context.Context, categoryId int) ([]model.Subject, error) {
	var subjects []model.Subject
	err := global.GVA_DB.WithContext(ctx).
		Joins("JOIN subject_category_rel ON subject_category_rel.subject_id = subjects.id").
		Where("subject_category_rel.category_id = ? AND subjects.status = ?", categoryId, "published").
		Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) GetAllSubjects(ctx context.Context) ([]model.Subject, error) {
	var subjects []model.Subject
	err := global.GVA_DB.WithContext(ctx).Where("status = ?", "published").Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) SearchSubjectsByName(ctx context.Context, keyword string, page int, pageSize int) ([]model.Subject, int64, error) {
	query := global.GVA_DB.WithContext(ctx).Model(&model.Subject{}).Where("name LIKE ? AND status = ?", "%"+keyword+"%", "published")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var subjects []model.Subject
	err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&subjects).Error
	return subjects, total, err
}

func (d *SubjectDao) GetUserCreatedSubjects(ctx context.Context, userId uint, page, pageSize int) ([]model.Subject, int64, error) {
	query := global.GVA_DB.WithContext(ctx).Model(&model.Subject{}).Where("creator_id = ?", userId)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var subjects []model.Subject
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&subjects).Error
	return subjects, total, err
}

func (d *SubjectDao) CreateSubjectWithTx(tx *gorm.DB, subject *model.Subject) error {
	return tx.Create(subject).Error
}

func (d *SubjectDao) UpdateSubjectDraftWithTx(tx *gorm.DB, subjectId int, nameDraft, iconDraft, descDraft string) error {
	return tx.Model(&model.Subject{}).Where("id = ?", subjectId).Updates(map[string]interface{}{
		"name_draft":        nameDraft,
		"icon_draft":        iconDraft,
		"description_draft": descDraft,
		"has_draft":         1,
	}).Error
}

func (d *SubjectDao) GetSubjectWritingProgress(ctx context.Context, userId uint, subjectId int) (*model.SubjectWritingProgress, error) {
	var progress model.SubjectWritingProgress
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id = ?", userId, subjectId).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (d *SubjectDao) UpsertSubjectWritingProgress(ctx context.Context, userId uint, subjectId int, nodeId int) error {
	return global.GVA_DB.WithContext(ctx).Exec(`
		INSERT INTO subject_writing_progress (user_id, subject_id, last_node_id, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE last_node_id = VALUES(last_node_id), updated_at = CURRENT_TIMESTAMP
	`, userId, subjectId, nodeId).Error
}

func (d *SubjectDao) GetUserCollectedSubjects(ctx context.Context, userId uint, page, pageSize int) ([]model.Subject, int64, error) {
	query := global.GVA_DB.WithContext(ctx).
		Model(&model.Subject{}).
		Joins("JOIN user_collect_items ON user_collect_items.subject_id = subjects.id").
		Where("user_collect_items.user_id = ?", userId)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var subjects []model.Subject
	offset := (page - 1) * pageSize
	err := query.Order("user_collect_items.created_at desc").Offset(offset).Limit(pageSize).Find(&subjects).Error
	return subjects, total, err
}

func (d *SubjectDao) GetUserLikedSubjects(ctx context.Context, userId uint, page, pageSize int) ([]model.Subject, int64, error) {
	query := global.GVA_DB.WithContext(ctx).
		Model(&model.Subject{}).
		Joins("JOIN user_subject_likes ON user_subject_likes.subject_id = subjects.id").
		Where("user_subject_likes.user_id = ?", userId)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var subjects []model.Subject
	offset := (page - 1) * pageSize
	err := query.Order("user_subject_likes.created_at desc").Offset(offset).Limit(pageSize).Find(&subjects).Error
	return subjects, total, err
}

func (d *SubjectDao) GetUserCollectFolders(ctx context.Context, userId uint) ([]model.UserCollectFolder, error) {
	var folders []model.UserCollectFolder
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("created_at desc").
		Find(&folders).Error
	return folders, err
}

func (d *SubjectDao) GetUserCollectedSubjectsByFolder(ctx context.Context, userId uint, folderId int, page, pageSize int) ([]model.Subject, int64, error) {
	var total int64
	if err := global.GVA_DB.WithContext(ctx).
		Model(&model.UserCollectItem{}).
		Where("user_id = ? AND folder_id = ?", userId, folderId).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var items []model.UserCollectItem
	if err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND folder_id = ?", userId, folderId).
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	if len(items) == 0 {
		return []model.Subject{}, total, nil
	}

	var subjectIds []int
	for _, item := range items {
		subjectIds = append(subjectIds, item.SubjectID)
	}

	subjects, err := d.GetSubjectsByIds(ctx, subjectIds)
	if err != nil {
		return nil, 0, err
	}

	subjectMap := make(map[int]model.Subject)
	for _, sub := range subjects {
		subjectMap[int(sub.ID)] = sub
	}

	var orderedSubjects []model.Subject
	for _, id := range subjectIds {
		if sub, ok := subjectMap[id]; ok {
			orderedSubjects = append(orderedSubjects, sub)
		}
	}

	return orderedSubjects, total, nil
}

func (d *SubjectDao) GetUserRecentSubjectProgress(ctx context.Context, userId uint, page int, pageSize int) ([]model.UserSubjectProgress, int64, error) {
	var total int64
	if err := global.GVA_DB.WithContext(ctx).Model(&model.UserSubjectProgress{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var progresses []model.UserSubjectProgress
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("last_study_time desc").
		Offset(offset).
		Limit(pageSize).
		Find(&progresses).Error
	return progresses, total, err
}

func (d *SubjectDao) GetUserSubjectsByStatus(ctx context.Context, userId uint, status string) ([]model.UserSubjectProgress, error) {
	var progresses []model.UserSubjectProgress
	query := global.GVA_DB.WithContext(ctx).Where("user_id = ?", userId)
	if status == "completed" {
		query = query.Where("progress_percent = ?", 100)
	} else {
		query = query.Where("progress_percent < ?", 100)
	}
	err := query.Order("last_study_time desc").Find(&progresses).Error
	return progresses, err
}

func (d *SubjectDao) UpsertUserSubjectProgress(ctx context.Context, userId uint, subjectId int, nodeId int, progressPercent float64) error {
	var progress model.UserSubjectProgress
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id = ?", userId, subjectId).First(&progress).Error

	if err == nil {
		// 更新
		return global.GVA_DB.WithContext(ctx).Model(&progress).Updates(map[string]interface{}{
			"last_node_id":     nodeId,
			"last_study_time":  time.Now(),
			"progress_percent": progressPercent,
		}).Error
	} else if err == gorm.ErrRecordNotFound {
		// 创建
		progress = model.UserSubjectProgress{
			UserID:          int(userId),
			SubjectID:       subjectId,
			LastNodeID:      nodeId,
			LastStudyTime:   time.Now(),
			ProgressPercent: progressPercent,
		}
		return global.GVA_DB.WithContext(ctx).Create(&progress).Error
	}
	return err
}

func (d *SubjectDao) GetUserLastLearningSubject(ctx context.Context, userId uint) (*model.UserSubjectProgress, error) {
	var progress model.UserSubjectProgress
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND progress_percent < ?", userId, 100).
		Order("last_study_time desc").
		First(&progress).Error
	return &progress, err
}

func (d *SubjectDao) GetSubjectsByIds(ctx context.Context, ids []int) ([]model.Subject, error) {
	var subjects []model.Subject
	if len(ids) == 0 {
		return subjects, nil
	}
	err := global.GVA_DB.WithContext(ctx).Where("id IN ? AND status = ?", ids, "published").Find(&subjects).Error
	return subjects, err
}

func (d *SubjectDao) GetSubjectById(ctx context.Context, id int) (*model.Subject, error) {
	var subject model.Subject
	err := global.GVA_DB.WithContext(ctx).Where("id = ? AND status = ?", id, "published").First(&subject).Error
	return &subject, err
}

func (d *SubjectDao) GetUserSubjectInteractions(ctx context.Context, userId uint, subjectIds []uint) (map[uint]bool, map[uint]bool, map[uint]model.UserSubjectProgress, error) {
	likedMap := make(map[uint]bool)
	collectedMap := make(map[uint]bool)
	progressMap := make(map[uint]model.UserSubjectProgress)

	if userId == 0 || len(subjectIds) == 0 {
		return likedMap, collectedMap, progressMap, nil
	}

	var likes []model.UserSubjectLike
	if err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id IN ?", userId, subjectIds).Find(&likes).Error; err != nil {
		return nil, nil, nil, err
	}
	for _, l := range likes {
		likedMap[uint(l.SubjectID)] = true
	}

	var collects []model.UserCollectItem
	if err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id IN ?", userId, subjectIds).Find(&collects).Error; err != nil {
		return nil, nil, nil, err
	}
	for _, c := range collects {
		collectedMap[uint(c.SubjectID)] = true
	}

	var progresses []model.UserSubjectProgress
	if err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id IN ?", userId, subjectIds).Find(&progresses).Error; err != nil {
		return nil, nil, nil, err
	}
	for _, p := range progresses {
		progressMap[uint(p.SubjectID)] = p
	}

	return likedMap, collectedMap, progressMap, nil
}

func (d *SubjectDao) GetSubjectLike(ctx context.Context, userId uint, subjectId int) (*model.UserSubjectLike, error) {
	var like model.UserSubjectLike
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id = ?", userId, subjectId).First(&like).Error
	return &like, err
}

func (d *SubjectDao) CreateSubjectLike(ctx context.Context, userId uint, subjectId int) error {
	like := model.UserSubjectLike{
		UserID:    int(userId),
		SubjectID: subjectId,
	}
	return global.GVA_DB.WithContext(ctx).Create(&like).Error
}

func (d *SubjectDao) DeleteSubjectLike(ctx context.Context, userId uint, subjectId int) error {
	return global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id = ?", userId, subjectId).Delete(&model.UserSubjectLike{}).Error
}

func (d *SubjectDao) CreateCollectFolder(ctx context.Context, userId uint, name, description string, isPublic int8) (model.UserCollectFolder, error) {
	folder := model.UserCollectFolder{
		UserID:      int(userId),
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
	}
	err := global.GVA_DB.WithContext(ctx).Create(&folder).Error
	return folder, err
}

func (d *SubjectDao) UpdateCollectFolderPublic(ctx context.Context, folderID int, isPublic int8) error {
	return global.GVA_DB.WithContext(ctx).Model(&model.UserCollectFolder{}).Where("id = ?", folderID).Update("is_public", isPublic).Error
}

func (d *SubjectDao) UpdateCollectFolderName(ctx context.Context, folderID int, name string) error {
	return global.GVA_DB.WithContext(ctx).Model(&model.UserCollectFolder{}).Where("id = ?", folderID).Update("name", name).Error
}

func (d *SubjectDao) CheckCollectFolderNameExists(ctx context.Context, userId uint, name string, excludeFolderId int) (bool, error) {
	var count int64
	query := global.GVA_DB.WithContext(ctx).Model(&model.UserCollectFolder{}).
		Where("user_id = ? AND name = ?", userId, name)
	if excludeFolderId > 0 {
		query = query.Where("id <> ?", excludeFolderId)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (d *SubjectDao) GetCollectFolderById(ctx context.Context, userId uint, folderId int) (*model.UserCollectFolder, error) {
	var folder model.UserCollectFolder
	err := global.GVA_DB.WithContext(ctx).Where("id = ? AND user_id = ?", folderId, userId).First(&folder).Error
	return &folder, err
}

func (d *SubjectDao) AddSubjectToFolder(ctx context.Context, userId uint, folderId int, subjectId int) error {
	item := model.UserCollectItem{
		UserID:    int(userId),
		FolderID:  folderId,
		SubjectID: subjectId,
	}
	return global.GVA_DB.WithContext(ctx).Create(&item).Error
}

func (d *SubjectDao) CheckSubjectInFolder(ctx context.Context, userId uint, folderId int, subjectId int) (bool, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.UserCollectItem{}).
		Where("user_id = ? AND folder_id = ? AND subject_id = ?", userId, folderId, subjectId).
		Count(&count).Error
	return count > 0, err
}

func (d *SubjectDao) UncollectSubject(ctx context.Context, userId uint, subjectId int) error {
	return global.GVA_DB.WithContext(ctx).Where("user_id = ? AND subject_id = ?", userId, subjectId).
		Delete(&model.UserCollectItem{}).Error
}

func (d *SubjectDao) GetSubjectsStats(ctx context.Context, subjectIds []uint) (map[uint]int64, map[uint]int64, error) {
	likeCountMap := make(map[uint]int64)
	collectCountMap := make(map[uint]int64)

	if len(subjectIds) == 0 {
		return likeCountMap, collectCountMap, nil
	}

	// 统计点赞数
	type CountResult struct {
		SubjectID int
		Total     int64
	}

	var likeResults []CountResult
	err := global.GVA_DB.WithContext(ctx).
		Model(&model.UserSubjectLike{}).
		Select("subject_id, count(id) as total").
		Where("subject_id IN ?", subjectIds).
		Group("subject_id").
		Scan(&likeResults).Error
	if err != nil {
		return nil, nil, err
	}
	for _, res := range likeResults {
		likeCountMap[uint(res.SubjectID)] = res.Total
	}

	// 统计收藏数
	var collectResults []CountResult
	err = global.GVA_DB.WithContext(ctx).
		Model(&model.UserCollectItem{}).
		Select("subject_id, count(id) as total").
		Where("subject_id IN ?", subjectIds).
		Group("subject_id").
		Scan(&collectResults).Error
	if err != nil {
		return nil, nil, err
	}
	for _, res := range collectResults {
		collectCountMap[uint(res.SubjectID)] = res.Total
	}

	return likeCountMap, collectCountMap, nil
}
