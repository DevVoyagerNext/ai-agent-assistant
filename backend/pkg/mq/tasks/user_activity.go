package tasks

import (
	"backend/dao"
	"backend/global"
	"backend/model"
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// UserActivityPayload 用户行为更新负载
type UserActivityPayload struct {
	UserID     uint   `json:"userId"`
	ActionType string `json:"actionType"`
	TargetType string `json:"targetType"`
	TargetID   int    `json:"targetId"`
	Score      int    `json:"score"`
}

// HandleUserActivity 处理用户行为更新任务
func HandleUserActivity(ctx context.Context, payloadStr string) error {
	var payload struct {
		Payload UserActivityPayload `json:"payload"`
	}

	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		global.GVA_LOG.Error("解析用户行为消息失败", zap.Error(err))
		return err
	}

	p := payload.Payload
	if p.UserID == 0 || p.ActionType == "" {
		return nil
	}

	description := ""
	if p.ActionType == "study_note" && p.TargetType == "knowledge_nodes" {
		// 1. 查询知识点名称和学科ID
		var nodeDao dao.KnowledgeNodeDao
		node, err := nodeDao.GetNodeByID(p.TargetID)
		if err == nil {
			// 2. 查询学科名称
			var subjectDao dao.SubjectDao
			subject, err := subjectDao.GetSubjectById(ctx, node.SubjectID)
			if err == nil {
				description = fmt.Sprintf("用户学习了知识点 %s (学科: %s)", node.Name, subject.Name)
			} else {
				description = fmt.Sprintf("用户学习了知识点 %s", node.Name)
			}
		}
	}

	// 1. 记录明细流水
	log := &model.UserActivityLog{
		UserID:     int(p.UserID),
		ActionType: p.ActionType,
		TargetType: p.TargetType,
		TargetID:   p.TargetID,
		ActionDesc: description,
		Score:      p.Score,
	}
	if err := dao.CreateUserActivityLog(ctx, log); err != nil {
		global.GVA_LOG.Error("创建用户行为日志失败", zap.Error(err))
		return err
	}

	// 2. 更新每日统计
	if err := dao.UpsertUserDailyActionStat(ctx, int(p.UserID), p.ActionType, p.TargetType, p.Score); err != nil {
		global.GVA_LOG.Error("更新用户每日行为统计失败", zap.Error(err))
		return err
	}

	return nil
}
