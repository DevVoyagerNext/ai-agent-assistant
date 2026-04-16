package dao

import (
	"backend/global"
	"backend/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type UserPrivateNoteDao struct{}

func (dao *UserPrivateNoteDao) GetNoteByIDWithScope(ctx context.Context, userID uint, noteID int, scope int) (model.UserPrivateNote, error) {
	var note model.UserPrivateNote

	tx := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND id = ? AND is_deleted = 0", userID, noteID)
	if scope == 0 {
		tx = tx.Where("is_public = 0")
	} else if scope == 1 {
		tx = tx.Where("is_public = 1")
	}

	err := tx.First(&note).Error
	return note, err
}

// GetNotesByParentWithScope 获取子文件夹/文件列表 (支持公开状态过滤)
func (dao *UserPrivateNoteDao) GetNotesByParentWithScope(ctx context.Context, userID uint, parentID int, scope int, page, pageSize int) ([]model.UserPrivateNote, int64, error) {
	query := global.GVA_DB.WithContext(ctx).
		Model(&model.UserPrivateNote{}).
		Where("user_id = ? AND parent_id = ? AND is_deleted = 0", userID, parentID)

	if scope == 0 {
		query = query.Where("is_public = 0")
	} else if scope == 1 {
		query = query.Where("is_public = 1")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notes []model.UserPrivateNote
	offset := (page - 1) * pageSize
	err := query.Order("sort_order asc, updated_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&notes).Error
	return notes, total, err
}

// GetNoteByID 获取笔记详情 (只查询未删除的)
func (dao *UserPrivateNoteDao) GetNoteByID(ctx context.Context, userID uint, noteID int) (model.UserPrivateNote, error) {
	var note model.UserPrivateNote
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND id = ? AND is_deleted = 0", userID, noteID).First(&note).Error
	return note, err
}

// GetNotesByParent 获取子文件夹/文件列表 (只查询未删除的)
func (dao *UserPrivateNoteDao) GetNotesByParent(ctx context.Context, userID uint, parentID int, page, pageSize int) ([]model.UserPrivateNote, int64, error) {
	query := global.GVA_DB.WithContext(ctx).
		Model(&model.UserPrivateNote{}).
		Where("user_id = ? AND parent_id = ? AND is_deleted = 0", userID, parentID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notes []model.UserPrivateNote
	offset := (page - 1) * pageSize
	err := query.Order("sort_order asc, updated_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&notes).Error
	return notes, total, err
}

// CheckNoteExists 判断同一个文件夹下是否存在同名同类型的文件/文件夹
func (dao *UserPrivateNoteDao) CheckNoteExists(ctx context.Context, userID uint, parentID int, title string, noteType string) (bool, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).
		Model(&model.UserPrivateNote{}).
		Where("user_id = ? AND parent_id = ? AND title = ? AND type = ? AND is_deleted = 0", userID, parentID, title, noteType).
		Count(&count).Error
	return count > 0, err
}

// CreateNote 创建文件夹或文件
func (dao *UserPrivateNoteDao) CreateNote(ctx context.Context, note *model.UserPrivateNote) error {
	return global.GVA_DB.WithContext(ctx).Create(note).Error
}

// CreateNoteShare 创建分享记录
func (dao *UserPrivateNoteDao) CreateNoteShare(ctx context.Context, share *model.NoteShare) error {
	return global.GVA_DB.WithContext(ctx).Create(share).Error
}

// GetNoteShareByToken 根据分享 token 查询分享记录
func (dao *UserPrivateNoteDao) GetNoteShareByToken(ctx context.Context, shareToken string) (model.NoteShare, error) {
	var share model.NoteShare
	err := global.GVA_DB.WithContext(ctx).
		Where("share_token = ?", shareToken).
		First(&share).Error
	return share, err
}

// GetShareInfoByToken 获取分享基础信息（关联用户表和笔记表）
type ShareInfoResult struct {
	model.NoteShare
	AuthorName   string `gorm:"column:author_name"`
	AuthorAvatar string `gorm:"column:author_avatar"`
	NoteTitle    string `gorm:"column:note_title"`
	NoteType     string `gorm:"column:note_type"`
}

func (dao *UserPrivateNoteDao) GetShareInfoByToken(ctx context.Context, shareToken string) (ShareInfoResult, error) {
	var result ShareInfoResult
	err := global.GVA_DB.WithContext(ctx).
		Table("note_shares ns").
		Select("ns.*, u.username as author_name, u.avatar_url as author_avatar, upn.title as note_title, upn.type as note_type").
		Joins("LEFT JOIN users u ON ns.user_id = u.id").
		Joins("LEFT JOIN user_private_notes upn ON ns.private_note_id = upn.id").
		Where("ns.share_token = ?", shareToken).
		First(&result).Error
	return result, err
}

// IncreaseNoteShareViewCount 增加分享访问次数
func (dao *UserPrivateNoteDao) IncreaseNoteShareViewCount(ctx context.Context, shareID uint) error {
	return global.GVA_DB.WithContext(ctx).
		Model(&model.NoteShare{}).
		Where("id = ?", shareID).
		Update("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// UpdateNote 更新笔记或文件夹
func (dao *UserPrivateNoteDao) UpdateNote(ctx context.Context, userID uint, noteID int, updates map[string]interface{}) error {
	return global.GVA_DB.WithContext(ctx).
		Model(&model.UserPrivateNote{}).
		Where("user_id = ? AND id = ?", userID, noteID).
		Updates(updates).Error
}

// DeleteNotesByIDs 按 ID 批量逻辑删除用户的私人笔记/文件夹
func (dao *UserPrivateNoteDao) DeleteNotesByIDs(ctx context.Context, userID uint, ids []int) error {
	if len(ids) == 0 || userID == 0 {
		return nil
	}
	now := time.Now()
	return global.GVA_DB.WithContext(ctx).
		Model(&model.UserPrivateNote{}).
		Where("user_id = ? AND id IN ?", userID, ids).
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"deleted_at": &now,
		}).Error
}
