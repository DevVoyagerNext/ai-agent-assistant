package tasks

import (
	"backend/dao"
	"backend/global"
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

// SubjectProgressPayload 教材学习进度更新负载
type SubjectProgressPayload struct {
	UserID    uint `json:"userId"`
	SubjectID int  `json:"subjectId"`
	NodeID    int  `json:"nodeId"`
}

// HandleSubjectProgress 处理教材进度更新任务
func HandleSubjectProgress(ctx context.Context, payloadStr string) error {
	var payload struct {
		Payload SubjectProgressPayload `json:"payload"`
	}

	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		global.GVA_LOG.Error("解析进度更新消息失败", zap.Error(err))
		return err
	}

	p := payload.Payload
	if p.UserID == 0 || p.SubjectID == 0 || p.NodeID == 0 {
		return nil // 忽略非法负载
	}

	// 1. 统计教材叶子节点总数
	var nodeDao dao.KnowledgeNodeDao
	totalLeafNodes, err := nodeDao.CountTotalLeafNodes(ctx, p.SubjectID)
	if err != nil {
		global.GVA_LOG.Error("统计教材总叶子节点数失败", zap.Error(err), zap.Int("subjectId", p.SubjectID))
		return err
	}

	// 2. 统计用户已学完的叶子节点数
	learnedLeafNodes, err := nodeDao.CountLearnedLeafNodes(ctx, p.UserID, p.SubjectID)
	if err != nil {
		global.GVA_LOG.Error("统计用户已学完叶子节点数失败", zap.Error(err), zap.Uint("userId", p.UserID), zap.Int("subjectId", p.SubjectID))
		return err
	}

	// 3. 计算进度百分比
	var progressPercent float64
	if totalLeafNodes > 0 {
		progressPercent = float64(learnedLeafNodes) / float64(totalLeafNodes) * 100
		if progressPercent > 100 {
			progressPercent = 100
		}
	}

	// 4. 使用 DAO 更新进度
	var subjectDao dao.SubjectDao
	return subjectDao.UpsertUserSubjectProgress(ctx, p.UserID, p.SubjectID, p.NodeID, progressPercent)
}
