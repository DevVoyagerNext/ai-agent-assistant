package dao

import (
	"backend/global"
	"backend/model"
	"context"
	"time"
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

func (dao *UserPrivateNoteDao) GetNotesByParentWithScope(ctx context.Context, userID uint, parentID int, scope int) ([]model.UserPrivateNote, error) {
	var notes []model.UserPrivateNote

	tx := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND parent_id = ? AND is_deleted = 0", userID, parentID)
	if scope == 0 {
		tx = tx.Where("is_public = 0")
	} else if scope == 1 {
		tx = tx.Where("is_public = 1")
	}

	err := tx.Order("sort_order asc, updated_at desc").Find(&notes).Error
	return notes, err
}

// GetNoteByID 获取笔记详情 (只查询未删除的)
func (dao *UserPrivateNoteDao) GetNoteByID(ctx context.Context, userID uint, noteID int) (model.UserPrivateNote, error) {
	var note model.UserPrivateNote
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND id = ? AND is_deleted = 0", userID, noteID).First(&note).Error
	return note, err
}

// GetNotesByParent 获取子文件夹/文件列表 (只查询未删除的)
func (dao *UserPrivateNoteDao) GetNotesByParent(ctx context.Context, userID uint, parentID int) ([]model.UserPrivateNote, error) {
	var notes []model.UserPrivateNote
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND parent_id = ? AND is_deleted = 0", userID, parentID).
		Order("sort_order asc, updated_at desc").
		Find(&notes).Error
	return notes, err
}

// CreateNote 创建文件夹或文件
func (dao *UserPrivateNoteDao) CreateNote(ctx context.Context, note *model.UserPrivateNote) error {
	return global.GVA_DB.WithContext(ctx).Create(note).Error
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
