package dao

import (
	"backend/global"
	"backend/model"
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateUserActivityLog 记录用户行为流水
func CreateUserActivityLog(ctx context.Context, log *model.UserActivityLog) error {
	return global.GVA_DB.WithContext(ctx).Create(log).Error
}

// UpsertUserDailyActionStat 更新或创建用户每日分类行为统计
func UpsertUserDailyActionStat(ctx context.Context, userID int, actionType, targetType string, score int) error {
	today := time.Now().Truncate(24 * time.Hour)

	// 使用 MySQL 的 ON DUPLICATE KEY UPDATE 语法实现原子的 Upsert
	// 这样可以彻底解决在高并发（如 MQ 快速消费）时，“先查后改”导致的 1062 Duplicate Entry 错误
	return global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "activity_date"},
			{Name: "action_type"},
			{Name: "target_type"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"action_count": gorm.Expr("user_daily_action_stats.action_count + ?", 1),
			"action_score": gorm.Expr("user_daily_action_stats.action_score + ?", score),
			"updated_at":   time.Now(),
		}),
	}).Create(&model.UserDailyActionStat{
		UserID:       userID,
		ActivityDate: today,
		ActionType:   actionType,
		TargetType:   targetType,
		ActionCount:  1,
		ActionScore:  score,
	}).Error
}

// GetUserFollowersCount 获取用户粉丝数 (被关注数)
func GetUserFollowersCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.UserFollow{}).Where("following_id = ?", userID).Count(&count).Error
	return count, err
}

// GetUserFollowingCount 获取用户关注数
func GetUserFollowingCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.UserFollow{}).Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}

// GetUserLearnedSubjectsCount 获取用户已学/在学教材总数
func GetUserLearnedSubjectsCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	// 查询用户有学习状态(learning/completed)或有笔记的节点所属的独特学科数量
	// 为简单起见，可以查询 user_study_status 表关联 knowledge_nodes 表去重统计 subject_id
	err := global.GVA_DB.WithContext(ctx).
		Table("user_study_status uss").
		Joins("JOIN knowledge_nodes kn ON uss.node_id = kn.id").
		Where("uss.user_id = ? AND uss.status IN ('learning', 'completed')", userID).
		Distinct("kn.subject_id").
		Count(&count).Error
	return count, err
}

// GetUserSharedNotesCount 获取用户分享的笔记总数
func GetUserSharedNotesCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.NoteShare{}).Where("user_id = ? AND is_active = 1", userID).Count(&count).Error
	return count, err
}
