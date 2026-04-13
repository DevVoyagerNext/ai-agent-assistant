package dao

import (
	"backend/global"
	"backend/model"
	"context"
	"time"
)

// DailyActivityCount 用于接收聚合后的每日活跃度
type DailyActivityCount struct {
	ActivityDate  time.Time `gorm:"column:activity_date"`
	ActivityCount int       `gorm:"column:activity_count"`
	ActivityScore int       `gorm:"column:activity_score"`
}

// GetUserActivities 获取用户一年的活跃度
func GetUserActivities(ctx context.Context, userID uint, startTime, endTime time.Time) ([]DailyActivityCount, error) {
	var activities []DailyActivityCount
	err := global.GVA_DB.WithContext(ctx).
		Table("user_daily_action_stats").
		Select("activity_date, sum(action_count) as activity_count, sum(action_score) as activity_score").
		Where("user_id = ? AND activity_date >= ? AND activity_date <= ?", userID, startTime, endTime).
		Group("activity_date").
		Find(&activities).Error
	return activities, err
}

// GetPublicPrivateNotes 获取公开私人笔记列表
func GetPublicPrivateNotes(ctx context.Context, userID uint, offset, limit int) (int64, []model.UserPrivateNote, error) {
	var notes []model.UserPrivateNote
	var total int64

	db := global.GVA_DB.WithContext(ctx).Model(&model.UserPrivateNote{}).
		Where("user_id = ? AND is_public = 1 AND type = 'markdown' AND is_deleted = 0", userID)

	err := db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}

	err = db.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&notes).Error
	return total, notes, err
}

// GetLearnedSubjects 获取已学/在学教材
// 包含关联 subject 基本信息
func GetLearnedSubjects(ctx context.Context, userID uint) ([]model.Subject, error) {
	var subjects []model.Subject
	// 查找用户正在学习或已完成的学习状态关联的所有 Subject
	err := global.GVA_DB.WithContext(ctx).
		Distinct("subjects.*").
		Table("subjects").
		Joins("JOIN knowledge_nodes kn ON kn.subject_id = subjects.id").
		Joins("JOIN user_study_status uss ON kn.id = uss.node_id").
		Where("uss.user_id = ? AND uss.status IN ('learning', 'completed')", userID).
		Find(&subjects).Error
	return subjects, err
}

// GetLearnedNodeCountBySubject 统计该教材下已学完(completed)的节点数
func GetLearnedNodeCountBySubject(ctx context.Context, userID uint, subjectID uint) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).
		Table("user_study_status uss").
		Joins("JOIN knowledge_nodes kn ON uss.node_id = kn.id").
		Where("uss.user_id = ? AND kn.subject_id = ? AND uss.status = 'completed'", userID, subjectID).
		Count(&count).Error
	return count, err
}

// GetTotalNodeCountBySubject 统计该教材下的总叶子节点(内容页)数
func GetTotalNodeCountBySubject(ctx context.Context, subjectID uint) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.KnowledgeNode{}).
		Where("subject_id = ? AND is_leaf = 1", subjectID).
		Count(&count).Error
	return count, err
}

// GetSharedNotes 获取已分享笔记列表
type SharedNoteResult struct {
	model.NoteShare
	NodeName string `gorm:"column:node_name"`
}

func GetSharedNotes(ctx context.Context, userID uint, offset, limit int) (int64, []SharedNoteResult, error) {
	var notes []SharedNoteResult
	var total int64

	db := global.GVA_DB.WithContext(ctx).
		Table("note_shares ns").
		Joins("JOIN knowledge_nodes kn ON ns.node_id = kn.id").
		Where("ns.user_id = ?", userID)

	err := db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}

	err = db.Select("ns.*, kn.name as node_name").
		Order("ns.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&notes).Error
	return total, notes, err
}
