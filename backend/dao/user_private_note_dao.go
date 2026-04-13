package dao

import (
	"backend/global"
	"backend/model"
	"context"
)

type UserPrivateNoteDao struct{}

// GetNoteByID 获取笔记详情
func (dao *UserPrivateNoteDao) GetNoteByID(ctx context.Context, userID uint, noteID int) (model.UserPrivateNote, error) {
	var note model.UserPrivateNote
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND id = ?", userID, noteID).First(&note).Error
	return note, err
}

// GetNotesByParent 获取子文件夹/文件列表
func (dao *UserPrivateNoteDao) GetNotesByParent(ctx context.Context, userID uint, parentID int) ([]model.UserPrivateNote, error) {
	var notes []model.UserPrivateNote
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND parent_id = ?", userID, parentID).
		Order("sort_order asc, updated_at desc").
		Find(&notes).Error
	return notes, err
}

// CreateNote 创建文件夹或文件
func (dao *UserPrivateNoteDao) CreateNote(ctx context.Context, note *model.UserPrivateNote) error {
	return global.GVA_DB.WithContext(ctx).Create(note).Error
}
