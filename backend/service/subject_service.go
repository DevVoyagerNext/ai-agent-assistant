package service

import (
	"backend/dao"
	"backend/dto"
	"backend/global"
	"backend/model"
	"backend/pkg/errmsg"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const subjectNameMaxLength = 100

type SubjectService struct {
	subjectDao       dao.SubjectDao
	knowledgeNodeDao dao.KnowledgeNodeDao
}

func subjectAuditStatusText(status int8) string {
	switch status {
	case 1:
		return "待审核"
	case 2:
		return "已通过"
	case 3:
		return "已驳回"
	default:
		return "编辑中"
	}
}

func subjectAuditActionText(action string) string {
	switch action {
	case "approve":
		return "通过"
	case "reject":
		return "驳回"
	default:
		return ""
	}
}

func subjectPublishState(subject model.Subject) (bool, string) {
	if subject.Status != "draft" {
		return false, "只有草稿状态的教材可以发布"
	}
	if subject.AuditStatus == 1 {
		return false, "教材正在审核中，请勿重复提交"
	}
	if subject.AuditStatus == 2 {
		return false, "教材已通过审核，无需再次发布"
	}
	return true, ""
}

func buildUserCreatedSubjectRes(subject model.Subject, likeCount, collectCount int64, auditLog *model.AuditLog) dto.UserCreatedSubjectRes {
	canPublish, disabledReason := subjectPublishState(subject)
	res := dto.UserCreatedSubjectRes{
		ID:                    subject.ID,
		Slug:                  subject.Slug,
		Name:                  subject.Name,
		NameDraft:             subject.NameDraft,
		Icon:                  subject.Icon,
		IconDraft:             subject.IconDraft,
		Description:           subject.Description,
		DescriptionDraft:      subject.DescriptionDraft,
		CoverImageID:          subject.CoverImageID,
		CoverImageIDDraft:     subject.CoverImageIDDraft,
		Status:                subject.Status,
		AuditStatus:           subject.AuditStatus,
		AuditStatusText:       subjectAuditStatusText(subject.AuditStatus),
		LastLogID:             subject.LastLogID,
		HasDraft:              subject.HasDraft,
		CanPublish:            canPublish,
		PublishDisabledReason: disabledReason,
		CreatedAt:             subject.CreatedAt,
		LikeCount:             likeCount,
		CollectCount:          collectCount,
	}
	if auditLog != nil {
		res.LastAuditAction = auditLog.Action
		res.LastAuditActionText = subjectAuditActionText(auditLog.Action)
		res.LastAuditRemark = auditLog.Remark
		res.LastAuditAt = &auditLog.CreatedAt
	}
	return res
}

func normalizeSubjectName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("教材名称不能为空")
	}
	if utf8.RuneCountInString(name) > subjectNameMaxLength {
		return "", fmt.Errorf("教材名称不能超过 %d 个字符", subjectNameMaxLength)
	}
	return name, nil
}

func (s *SubjectService) enrichSubjectList(ctx context.Context, userId uint, subjects []model.Subject) ([]dto.SubjectRes, int) {
	var res []dto.SubjectRes
	if len(subjects) == 0 {
		return res, errmsg.CodeSuccess
	}

	var subjectIds []uint
	for _, sub := range subjects {
		subjectIds = append(subjectIds, sub.ID)
	}

	likedMap, collectedMap, progressMap, err := s.subjectDao.GetUserSubjectInteractions(ctx, userId, subjectIds)
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
	categories, err := s.subjectDao.GetCategories(ctx)
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
	subjects, err := s.subjectDao.GetSubjectsByCategoryID(ctx, categoryId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	return s.enrichSubjectList(ctx, userId, subjects)
}

func (s *SubjectService) GetAllSubjects(ctx context.Context, userId uint) ([]dto.SubjectRes, int) {
	subjects, err := s.subjectDao.GetAllSubjects(ctx)
	if err != nil {
		return nil, errmsg.CodeError
	}

	return s.enrichSubjectList(ctx, userId, subjects)
}

func (s *SubjectService) GetSubjectByID(ctx context.Context, subjectId int, userId uint) (*dto.SubjectRes, int) {
	subject, err := s.subjectDao.GetSubjectById(ctx, subjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errmsg.CodeSuccess
		}
		return nil, errmsg.CodeError
	}

	enriched, code := s.enrichSubjectList(ctx, userId, []model.Subject{*subject})
	if code != errmsg.CodeSuccess || len(enriched) == 0 {
		return nil, code
	}

	return &enriched[0], errmsg.CodeSuccess
}

func (s *SubjectService) SearchSubjects(ctx context.Context, keyword string, userId uint, page int, pageSize int) (dto.SubjectListRes, int) {
	subjects, total, err := s.subjectDao.SearchSubjectsByName(ctx, keyword, page, pageSize)
	if err != nil {
		return dto.SubjectListRes{}, errmsg.CodeError
	}

	enriched, code := s.enrichSubjectList(ctx, userId, subjects)
	if code != errmsg.CodeSuccess {
		return dto.SubjectListRes{}, code
	}

	return dto.SubjectListRes{Total: total, List: enriched}, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserCreatedSubjects(ctx context.Context, userId uint, page int, pageSize int) (dto.UserCreatedSubjectListRes, int) {
	subjects, total, err := s.subjectDao.GetUserCreatedSubjects(ctx, userId, page, pageSize)
	if err != nil {
		return dto.UserCreatedSubjectListRes{}, errmsg.CodeError
	}

	if len(subjects) == 0 {
		return dto.UserCreatedSubjectListRes{Total: 0, List: []dto.UserCreatedSubjectRes{}}, errmsg.CodeSuccess
	}

	var subjectIds []uint
	var logIds []int64
	for _, sub := range subjects {
		subjectIds = append(subjectIds, sub.ID)
		if sub.LastLogID > 0 {
			logIds = append(logIds, sub.LastLogID)
		}
	}

	likeCountMap, collectCountMap, err := s.subjectDao.GetSubjectsStats(ctx, subjectIds)
	if err != nil {
		return dto.UserCreatedSubjectListRes{}, errmsg.CodeError
	}
	auditLogMap, err := s.subjectDao.GetAuditLogsByIDs(ctx, logIds)
	if err != nil {
		return dto.UserCreatedSubjectListRes{}, errmsg.CodeError
	}

	var resList []dto.UserCreatedSubjectRes
	for _, sub := range subjects {
		var auditLog *model.AuditLog
		if log, ok := auditLogMap[sub.LastLogID]; ok {
			auditLog = &log
		}
		resList = append(resList, buildUserCreatedSubjectRes(sub, likeCountMap[sub.ID], collectCountMap[sub.ID], auditLog))
	}

	return dto.UserCreatedSubjectListRes{Total: total, List: resList}, errmsg.CodeSuccess
}

// CreateSubject 创建新教材（同时生成顶级知识节点）
func (s *SubjectService) CreateSubject(ctx context.Context, userId uint, req dto.CreateSubjectReq) (uint, int) {
	name, err := normalizeSubjectName(req.NameDraft)
	if err != nil {
		return 0, errmsg.CodeError
	}

	// 生成一个简单的唯一 slug（UUID前8位）
	slug := fmt.Sprintf("%s-%s", uuid.New().String()[:8], fmt.Sprintf("%d", time.Now().Unix()))

	// 开启事务，保证教材和根节点同时创建
	var newSubjectId uint
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建教材记录
		subject := model.Subject{
			CreatorID:         int(userId),
			Slug:              slug,
			Name:              name, // 发布前默认占位
			NameDraft:         name,
			Description:       "", // 发布前默认占位
			DescriptionDraft:  req.DescriptionDraft,
			Icon:              "", // 发布前默认占位
			IconDraft:         req.IconDraft,
			CoverImageID:      0, // 发布前默认占位
			CoverImageIDDraft: req.CoverImageIdDraft,
			Status:            "draft",
			AuditStatus:       0,
			HasDraft:          1,
		}

		if err := s.subjectDao.CreateSubjectWithTx(tx, &subject); err != nil {
			return err
		}
		newSubjectId = subject.ID

		// 2. 创建关联的顶级“篇”节点 (parent_id = 0)
		topNode := model.KnowledgeNode{
			SubjectID:   int(newSubjectId),
			ParentID:    0,
			Path:        "0/",
			Name:        name,
			NameDraft:   name,
			Status:      "draft",
			AuditStatus: 0,
			HasDraft:    1,
			Level:       1,
			IsLeaf:      0,
			SortOrder:   1, // 第一个节点
		}

		if err := s.knowledgeNodeDao.CreateKnowledgeNodeWithTx(tx, &topNode); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, errmsg.CodeError
	}

	return newSubjectId, errmsg.CodeSuccess
}

// UpdateSubjectDraft 修改教材名称或简介草稿
func (s *SubjectService) UpdateSubjectDraft(ctx context.Context, userId uint, req dto.UpdateSubjectDraftReq) error {
	name, err := normalizeSubjectName(req.NameDraft)
	if err != nil {
		return err
	}

	// 1. 校验教材是否存在且属于该用户
	var subject model.Subject
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND creator_id = ?", req.SubjectID, userId).First(&subject).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("无权操作该教材或教材不存在")
		}
		return err
	}

	// 2. 开启事务更新教材草稿及对应的顶级节点草稿
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新教材的 draft 字段
		if err := s.subjectDao.UpdateSubjectDraftWithTx(tx, req.SubjectID, name, req.IconDraft, req.DescriptionDraft); err != nil {
			return err
		}
		// 同步更新 parent_id=0 的顶级知识点
		if err := s.knowledgeNodeDao.UpdateSubjectTopNodeDraftWithTx(tx, req.SubjectID, name); err != nil {
			return err
		}
		return nil
	})

	return err
}

// UpdateSubjectName 修改教材正式名称，并同步顶级知识节点名称。
func (s *SubjectService) UpdateSubjectName(ctx context.Context, userId uint, subjectId int, name string) error {
	name, err := normalizeSubjectName(name)
	if err != nil {
		return err
	}

	var subject model.Subject
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND creator_id = ?", subjectId, userId).First(&subject).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("无权操作该教材或教材不存在")
		}
		return err
	}

	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.subjectDao.UpdateSubjectNameWithTx(tx, subjectId, userId, name); err != nil {
			return err
		}
		if err := s.knowledgeNodeDao.UpdateSubjectTopNodeNameWithTx(tx, subjectId, name); err != nil {
			return err
		}
		return nil
	})
}

// PublishSubject 发布教材
func (s *SubjectService) PublishSubject(ctx context.Context, userId uint, subjectId int) (dto.UserCreatedSubjectRes, error) {
	// 1. 校验教材是否存在且属于该用户
	var subject model.Subject
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND creator_id = ?", subjectId, userId).First(&subject).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserCreatedSubjectRes{}, errors.New("无权操作该教材或教材不存在")
		}
		return dto.UserCreatedSubjectRes{}, err
	}

	if subject.Status != "draft" {
		return dto.UserCreatedSubjectRes{}, errors.New("只有草稿状态的教材可以发布")
	}
	if subject.AuditStatus == 1 {
		return dto.UserCreatedSubjectRes{}, errors.New("教材正在审核中，请勿重复提交")
	}

	// 2. 更新教材审核状态为待审核 (audit_status=1)，草稿在后台审批通过后再转正
	if err := s.subjectDao.PublishSubject(ctx, subjectId, userId); err != nil {
		return dto.UserCreatedSubjectRes{}, err
	}

	subject.AuditStatus = 1
	canPublish, disabledReason := subjectPublishState(subject)
	return dto.UserCreatedSubjectRes{
		ID:                    subject.ID,
		Slug:                  subject.Slug,
		Name:                  subject.Name,
		NameDraft:             subject.NameDraft,
		Icon:                  subject.Icon,
		IconDraft:             subject.IconDraft,
		Description:           subject.Description,
		DescriptionDraft:      subject.DescriptionDraft,
		CoverImageID:          subject.CoverImageID,
		CoverImageIDDraft:     subject.CoverImageIDDraft,
		Status:                subject.Status,
		AuditStatus:           subject.AuditStatus,
		AuditStatusText:       subjectAuditStatusText(subject.AuditStatus),
		LastLogID:             subject.LastLogID,
		HasDraft:              subject.HasDraft,
		CanPublish:            canPublish,
		PublishDisabledReason: disabledReason,
		CreatedAt:             subject.CreatedAt,
	}, nil
}

func (s *SubjectService) GetUserCollectedSubjects(ctx context.Context, userId uint, page, pageSize int) ([]dto.SubjectRes, int64, error) {
	subjects, total, err := s.subjectDao.GetUserCollectedSubjects(ctx, userId, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var subjectIds []uint
	for _, sub := range subjects {
		subjectIds = append(subjectIds, sub.ID)
	}

	likedMap, collectedMap, progressMap, err := s.subjectDao.GetUserSubjectInteractions(ctx, userId, subjectIds)
	if err != nil {
		return nil, 0, err
	}

	var res []dto.SubjectRes
	for _, sub := range subjects {
		item := dto.ConvertSubjectToRes(&sub)
		item.IsLiked = likedMap[sub.ID]
		item.IsCollected = collectedMap[sub.ID]
		if p, ok := progressMap[sub.ID]; ok {
			item.ProgressPercent = p.ProgressPercent
			item.LastNodeID = p.LastNodeID
		}
		res = append(res, item)
	}
	return res, total, nil
}

func (s *SubjectService) GetUserLikedSubjects(ctx context.Context, userId uint, page, pageSize int) ([]dto.SubjectRes, int64, error) {
	subjects, total, err := s.subjectDao.GetUserLikedSubjects(ctx, userId, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var subjectIds []uint
	for _, sub := range subjects {
		subjectIds = append(subjectIds, sub.ID)
	}

	likedMap, collectedMap, progressMap, err := s.subjectDao.GetUserSubjectInteractions(ctx, userId, subjectIds)
	if err != nil {
		return nil, 0, err
	}

	var res []dto.SubjectRes
	for _, sub := range subjects {
		item := dto.ConvertSubjectToRes(&sub)
		item.IsLiked = likedMap[sub.ID]
		item.IsCollected = collectedMap[sub.ID]
		if p, ok := progressMap[sub.ID]; ok {
			item.ProgressPercent = p.ProgressPercent
			item.LastNodeID = p.LastNodeID
		}
		res = append(res, item)
	}
	return res, total, nil
}

func (s *SubjectService) GetUserCollectFolders(ctx context.Context, userId uint) ([]dto.CollectFolderRes, int) {
	folders, err := s.subjectDao.GetUserCollectFolders(ctx, userId)
	if err != nil {
		return nil, errmsg.CodeError
	}

	var res []dto.CollectFolderRes
	for _, f := range folders {
		res = append(res, dto.ConvertCollectFolderToRes(&f))
	}
	return res, errmsg.CodeSuccess
}

func (s *SubjectService) GetUserCollectedSubjectsByFolder(ctx context.Context, userId uint, folderId int, page, pageSize int) ([]dto.SubjectRes, int64, error) {
	subjects, total, err := s.subjectDao.GetUserCollectedSubjectsByFolder(ctx, userId, folderId, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var subjectIds []uint
	for _, sub := range subjects {
		subjectIds = append(subjectIds, sub.ID)
	}

	likedMap, collectedMap, progressMap, err := s.subjectDao.GetUserSubjectInteractions(ctx, userId, subjectIds)
	if err != nil {
		return nil, 0, err
	}

	var res []dto.SubjectRes
	for _, sub := range subjects {
		item := dto.ConvertSubjectToRes(&sub)
		item.IsLiked = likedMap[sub.ID]
		item.IsCollected = collectedMap[sub.ID]
		if p, ok := progressMap[sub.ID]; ok {
			item.ProgressPercent = p.ProgressPercent
			item.LastNodeID = p.LastNodeID
		}
		res = append(res, item)
	}
	return res, total, nil
}

func (s *SubjectService) UpdateCollectFolderPublic(ctx context.Context, userId uint, folderId int, isPublic int8) int {
	// 1. 检查收藏夹是否存在且属于该用户
	_, err := s.subjectDao.GetCollectFolderById(ctx, userId, folderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errmsg.CodeError
		}
		return errmsg.CodeError
	}

	// 2. 更新状态
	if err := s.subjectDao.UpdateCollectFolderPublic(ctx, folderId, isPublic); err != nil {
		return errmsg.CodeError
	}

	return errmsg.CodeSuccess
}

func (s *SubjectService) RenameCollectFolder(ctx context.Context, userId uint, folderId int, name string) int {
	// 1. 检查收藏夹是否存在且属于该用户
	_, err := s.subjectDao.GetCollectFolderById(ctx, userId, folderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errmsg.CodeError
		}
		return errmsg.CodeError
	}

	// 2. 检查新名字是否已存在（排除当前收藏夹）
	exists, err := s.subjectDao.CheckCollectFolderNameExists(ctx, userId, name, folderId)
	if err != nil {
		return errmsg.CodeError
	}
	if exists {
		return errmsg.CodeError // 之后可以定义专门的错误码，目前先用通用错误
	}

	// 3. 更新名称
	if err := s.subjectDao.UpdateCollectFolderName(ctx, folderId, name); err != nil {
		return errmsg.CodeError
	}

	return errmsg.CodeSuccess
}

func (s *SubjectService) GetUserRecentSubjects(ctx context.Context, userId uint, page int, pageSize int) (dto.RecentSubjectListRes, int) {
	progresses, total, err := s.subjectDao.GetUserRecentSubjectProgress(ctx, userId, page, pageSize)
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

	subjects, err := s.subjectDao.GetSubjectsByIds(ctx, subjectIds)
	if err != nil {
		return dto.RecentSubjectListRes{}, errmsg.CodeError
	}

	enrichedSubjects, code := s.enrichSubjectList(ctx, userId, subjects)
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
			status := "learning"
			if p.ProgressPercent == 100 {
				status = "completed"
			}
			list = append(list, dto.UserSubjectProgressRes{
				Subject:         sub,
				Status:          status,
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
	progresses, err := s.subjectDao.GetUserSubjectsByStatus(ctx, userId, status)
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

	subjects, err := s.subjectDao.GetSubjectsByIds(ctx, subjectIds)
	if err != nil {
		return nil, errmsg.CodeError
	}

	enrichedSubjects, code := s.enrichSubjectList(ctx, userId, subjects)
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
			derivedStatus := "learning"
			if p.ProgressPercent == 100 {
				derivedStatus = "completed"
			}
			res = append(res, dto.UserSubjectProgressRes{
				Subject:         sub,
				Status:          derivedStatus,
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
	progress, err := s.subjectDao.GetUserLastLearningSubject(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errmsg.CodeSuccess
		}
		return nil, errmsg.CodeError
	}

	subject, err := s.subjectDao.GetSubjectById(ctx, progress.SubjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errmsg.CodeSuccess
		}
		return nil, errmsg.CodeError
	}

	enrichedSubjects, code := s.enrichSubjectList(ctx, userId, []model.Subject{*subject})
	if code != errmsg.CodeSuccess || len(enrichedSubjects) == 0 {
		return nil, errmsg.CodeError
	}

	status := "learning"
	if progress.ProgressPercent == 100 {
		status = "completed"
	}

	res := &dto.UserSubjectProgressRes{
		Subject:         enrichedSubjects[0],
		Status:          status,
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
	_, err := s.subjectDao.GetSubjectLike(ctx, userId, subjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 2. 未点赞，执行点赞
			if err := s.subjectDao.CreateSubjectLike(ctx, userId, subjectId); err != nil {
				return false, errmsg.CodeError
			}
			return true, errmsg.CodeSuccess // true 表示当前状态为已点赞
		}
		return false, errmsg.CodeError
	}

	// 3. 已点赞，取消点赞
	if err := s.subjectDao.DeleteSubjectLike(ctx, userId, subjectId); err != nil {
		return false, errmsg.CodeError
	}
	return false, errmsg.CodeSuccess // false 表示当前状态为未点赞
}

func (s *SubjectService) CreateCollectFolder(ctx context.Context, userId uint, req dto.CreateCollectFolderReq) (*dto.CollectFolderRes, int) {
	folder, err := s.subjectDao.CreateCollectFolder(ctx, userId, req.Name, req.Description, *req.IsPublic)
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
	_, err := s.subjectDao.GetCollectFolderById(ctx, userId, folderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errmsg.CodeError // 或者定义专门的收藏夹不存在错误码
		}
		return errmsg.CodeError
	}

	// 2. 检查是否已经收藏过该教材
	exists, err := s.subjectDao.CheckSubjectInFolder(ctx, userId, folderId, subjectId)
	if err != nil {
		return errmsg.CodeError
	}
	if exists {
		return errmsg.CodeSuccess // 已经存在，幂等处理
	}

	// 3. 添加到收藏夹
	if err := s.subjectDao.AddSubjectToFolder(ctx, userId, folderId, subjectId); err != nil {
		return errmsg.CodeError
	}
	return errmsg.CodeSuccess
}

func (s *SubjectService) UncollectSubject(ctx context.Context, userId uint, subjectId int) int {
	// 直接根据 userId 和 subjectId 取消收藏（从所有收藏夹中移除）
	if err := s.subjectDao.UncollectSubject(ctx, userId, subjectId); err != nil {
		return errmsg.CodeError
	}
	return errmsg.CodeSuccess
}
