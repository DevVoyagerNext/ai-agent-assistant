package aidao

import (
	"backend/global"
	"backend/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type AIDao struct{}

// CreateSession 创建新的对话会话
func (d *AIDao) CreateSession(ctx context.Context, session *model.Session) error {
	return global.GVA_DB.WithContext(ctx).Create(session).Error
}

// GetSessionByID 查询指定的对话会话
func (d *AIDao) GetSessionByID(ctx context.Context, sessionID int64, userID uint) (*model.Session, error) {
	var session model.Session
	err := global.GVA_DB.WithContext(ctx).
		Where("id = ? AND user_id = ? AND is_deleted = false", sessionID, userID).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会话不存在或无权访问")
		}
		return nil, errors.New("查询会话失败")
	}
	return &session, nil
}

// CreateMessage 创建对话消息
func (d *AIDao) CreateMessage(ctx context.Context, msg *model.Message) error {
	return global.GVA_DB.WithContext(ctx).Create(msg).Error
}

// UpdateMessageContent 更新消息内容
func (d *AIDao) UpdateMessageContent(ctx context.Context, msgID int64, content string) error {
	return global.GVA_DB.WithContext(ctx).
		Model(&model.Message{}).
		Where("id = ?", msgID).
		Update("content", content).Error
}

// GetRecentMessages 获取会话的最近历史消息
func (d *AIDao) GetRecentMessages(ctx context.Context, sessionID int64, limit int) ([]model.Message, error) {
	var historyMsgs []model.Message
	err := global.GVA_DB.WithContext(ctx).
		Where("session_id = ? AND status = 'active'", sessionID).
		Order("created_at desc").
		Limit(limit).
		Find(&historyMsgs).Error
	if err != nil {
		return nil, err
	}

	// 反转顺序
	for i := len(historyMsgs)/2 - 1; i >= 0; i-- {
		opp := len(historyMsgs) - 1 - i
		historyMsgs[i], historyMsgs[opp] = historyMsgs[opp], historyMsgs[i]
	}

	return historyMsgs, nil
}
